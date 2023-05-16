// Copyright 2023 Dolthub, Inc.
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
	_ "embed"
	"runtime"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Embedded data that will be loaded into our WASM runtime
var (
	//go:embed icu/wasm/icu.wasm
	icuWasm []byte // This is generated using the "build.sh" script in the "icu" folder
)

var r wazero.Runtime
var modulePool = sync.Pool{
	New: func() any {
		ctx := context.Background()

		// Load the ICU library
		mod, err := r.Instantiate(ctx, icuWasm)
		if err != nil {
			panic(err)
		}

		// We set a finalizer here, as the pool will periodically empty itself, and we need to close the module during
		// that time.
		runtime.SetFinalizer(mod, func(mod api.Module) {
			_ = mod.Close(context.Background())
		})
		return mod
	},
}

func init() {
	ctx := context.Background()

	// Create the WASM runtime
	r = wazero.NewRuntime(ctx)
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
}
