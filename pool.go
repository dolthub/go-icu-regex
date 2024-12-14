// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package regex

import (
	"context"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"reflect"
	"runtime"
	"sync"
)

// modulePool is the pool that is used internally by the project.
var modulePool = NewPool()

// RuntimeTracker tracks all relevant information that the Pool needs regarding a runtime.
type RuntimeTracker struct {
	id       uint64
	r        wazero.Runtime
	compiled wazero.CompiledModule
	modules  []api.Module
	max      uint64
	fetches  uint64
}

// Pool is a special pool object for handling ICU regex modules. The cause isn't quite clear, but runtimes continue to
// hold onto memory even when their owned modules are closed, so this special pool type will also recycle the runtimes
// once a certain number of modules have been fetched.
type Pool struct {
	mutex           *sync.Mutex
	runtimes        []*RuntimeTracker
	returnedModules map[uintptr]uint64
	nextId          uint64
	maxFetch        uint64
}

// NewPool creates a new *Pool.
func NewPool() *Pool {
	r, compiled := createRuntime(context.Background())
	pool := &Pool{
		mutex: &sync.Mutex{},
		runtimes: []*RuntimeTracker{{
			id:       1,
			r:        r,
			compiled: compiled,
			modules:  make([]api.Module, 0, 16),
			max:      0,
			fetches:  0,
		}},
		returnedModules: make(map[uintptr]uint64),
		nextId:          2,
		maxFetch:        128,
	}
	return pool
}

// Get returns a new module from the pool.
func (pool *Pool) Get() api.Module {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	ctx := context.Background()
	rtracker := pool.runtimes[len(pool.runtimes)-1]
	rtracker.fetches++
	// If we've used up the number of fetches allowed in this runtime, then we'll create a new one
	if rtracker.fetches >= pool.maxFetch {
		r, compiled := createRuntime(ctx)
		rtracker = &RuntimeTracker{
			id:       pool.nextId,
			r:        r,
			compiled: compiled,
			modules:  make([]api.Module, 0, 16),
			max:      0,
			fetches:  0,
		}
		pool.runtimes = append(pool.runtimes, rtracker)
		pool.nextId++
	}
	var module api.Module
	// If the runtime has no modules remaining, then we need to create a new module
	if len(rtracker.modules) == 0 {
		rtracker.max++
		var err error
		module, err = rtracker.r.InstantiateModule(ctx, rtracker.compiled, icuConfig)
		if err != nil {
			panic(err)
		}
	} else {
		// Pop the last module from the slice
		module = rtracker.modules[len(rtracker.modules)-1]
		rtracker.modules = rtracker.modules[:len(rtracker.modules)-1]
	}
	// Now we need to track that this module is being returned
	pool.returnedModules[reflect.ValueOf(module).Pointer()] = rtracker.id
	runtime.SetFinalizer(module, func(module api.Module) {
		pool.finalized(module)
	})
	return module
}

// Put returns the module to the pool.
func (pool *Pool) Put(module api.Module) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.receivedModule(module, true)
}

// finalized is called by the finalizer, and only exists to catch orphaned modules.
func (pool *Pool) finalized(module api.Module) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.receivedModule(module, false)
}

// receivedModule is called when either the module is returned through Put, or the finalizer catches an orphaned module
// through finalized.
func (pool *Pool) receivedModule(module api.Module, isPut bool) {
	// Remove the finalizer that was set when the object was fetched.
	// This is only called from Put, as the finalizer is being called so we don't want to remove it.
	if isPut {
		runtime.SetFinalizer(module, nil)
	}
	// Grab the runtime ID and remove the module from the tracking map
	ptr := reflect.ValueOf(module).Pointer()
	runtimeId := pool.returnedModules[ptr]
	delete(pool.returnedModules, ptr)
	for rtrackerIdx := 0; rtrackerIdx < len(pool.runtimes); rtrackerIdx++ {
		ctx := context.Background()
		rtracker := pool.runtimes[rtrackerIdx]
		// If this is a different runtime, then we still need to check whether it should be removed
		if rtracker.id != runtimeId {
			if rtracker.fetches >= pool.maxFetch && uint64(len(rtracker.modules)) >= rtracker.max {
				pool.closeRuntime(ctx, rtrackerIdx, rtracker)
				rtrackerIdx--
			}
			continue
		}
		if isPut {
			// Add the module back to the runtime when called from Put
			rtracker.modules = append(rtracker.modules, module)
		} else {
			// We remove the module from the runtime altogether when called from the finalizer
			rtracker.max--
			_ = module.Close(ctx)
		}
		// If this runtime has run out of fetches and all of its modules are back, then we need to close and remove it
		if rtracker.fetches >= pool.maxFetch && uint64(len(rtracker.modules)) >= rtracker.max {
			pool.closeRuntime(ctx, rtrackerIdx, rtracker)
		}
		return
	}
	// We could not find the runtime ID (or the module was not in the map), which should never happen
	panic("go-icu-regex pool found orphaned module")
}

// closeRuntime closes the given runtime, as well as removing it from the list of runtimes.
func (pool *Pool) closeRuntime(ctx context.Context, rtrackerIdx int, rtracker *RuntimeTracker) {
	// First we'll close all the modules, then we'll close the runtime itself
	for _, mod := range rtracker.modules {
		_ = mod.Close(ctx)
	}
	_ = rtracker.r.Close(ctx)
	// We then remove the runtime from the slice
	newSlice := make([]*RuntimeTracker, len(pool.runtimes)-1)
	copy(newSlice, pool.runtimes[:rtrackerIdx])
	copy(newSlice, pool.runtimes[rtrackerIdx+1:])
	pool.runtimes = newSlice
}

// createRuntime creates a new runtime, as well as compiling the ICU module. The compiled module is only valid with the
// runtime that compiled it.
func createRuntime(ctx context.Context) (wazero.Runtime, wazero.CompiledModule) {
	r := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)
	envBuilder := r.NewHostModuleBuilder("env")
	noop_two := func(int32, int32) int32 { return -1 }
	noop_four := func(int32, int32, int32, int32) int32 { return -1 }
	envBuilder.NewFunctionBuilder().WithFunc(noop_two).Export("__syscall_stat64")
	envBuilder.NewFunctionBuilder().WithFunc(noop_two).Export("__syscall_lstat64")
	envBuilder.NewFunctionBuilder().WithFunc(noop_two).Export("__syscall_fstat64")
	envBuilder.NewFunctionBuilder().WithFunc(noop_four).Export("__syscall_newfstatat")
	_, err := envBuilder.Instantiate(ctx)
	if err != nil {
		panic(err)
	}
	compiledIcuWasm, err := r.CompileModule(ctx, icuWasm)
	if err != nil {
		panic(err)
	}
	return r, compiledIcuWasm
}

// SetPoolFetchMax determines how many fetches are allowed from the internal Pool before a runtime is recycled.
func SetPoolFetchMax(maxFetch uint64) {
	modulePool.mutex.Lock()
	defer modulePool.mutex.Unlock()
	modulePool.maxFetch = maxFetch
}
