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
	"fmt"
	"runtime"
	"unicode/utf16"

	"github.com/tetratelabs/wazero/api"
	"gopkg.in/src-d/go-errors.v1"
)

// Regex is an interface that wraps around the ICU library, exposing ICU's regular expression functionality. It is
// imperative that Regex is closed once it is finished.
type Regex interface {
	// SetRegexString sets the string that will later be matched against. This must be called at least once before any other
	// calls are made (except for Close).
	SetRegexString(ctx context.Context, regexStr string, flags RegexFlags) error
	// SetMatchString sets the string that we will either be matching against, or executing the replacements on. This
	// must be called after SetRegexString, but before any other calls.
	SetMatchString(ctx context.Context, matchStr string) error
	// Matches returns whether the previously-set regex matches the previously-set match string. Must call
	// SetRegexString and SetMatchString before this function.
	Matches(ctx context.Context, start int, occurrence int) (bool, error)
	// Replace returns a new string with the replacement string occupying the matched portions of the match string,
	// based on the regex. Position starts at 1, not 0. Must call SetRegexString and SetMatchString before this function.
	Replace(ctx context.Context, replacementStr string, position int, occurrence int) (string, error)
	// StringBufferSize returns the size of the string buffers, in bytes. If the string buffer is not being used, then
	// this returns zero.
	StringBufferSize() uint32
	// Close frees up the internal resources. This MUST be called, else a panic will occur at some non-deterministic time.
	Close() error
}

var (
	// ErrRegexNotYetSet is returned when attempting to use another function before the regex has been initialized.
	ErrRegexNotYetSet = errors.NewKind("SetRegexString must be called before any other function")
	// ErrMatchNotYetSet is returned when attempting to use another function before the match string has been set.
	ErrMatchNotYetSet = errors.NewKind("SetMatchString must be called as there is nothing to match against")
	// ErrInvalidRegex is returned when an invalid regex is given
	ErrInvalidRegex = errors.NewKind("the given regular expression is invalid")
)

// ShouldPanic determines whether the finalizer will panic if it finds a Regex that has not been closed.
var ShouldPanic bool = true

// RegexFlags are flags to define the behavior of the regular expression. Use OR (|) to combine flags. All flag values
// were taken directly from ICU.
type RegexFlags uint32

const (
	// Enable case insensitive matching.
	RegexFlags_None RegexFlags = 0

	// Enable case insensitive matching.
	RegexFlags_Case_Insensitive RegexFlags = 2

	// Allow white space and comments within patterns.
	RegexFlags_Comments RegexFlags = 4

	// If set, '.' matches line terminators,  otherwise '.' matching stops at line end.
	RegexFlags_Dot_All RegexFlags = 32

	// If set, treat the entire pattern as a literal string. Metacharacters or escape sequences in the input sequence
	// will be given no special meaning.
	//
	// The flag RegexFlags_Case_Insensitive retains its impact on matching when used in conjunction with this flag. The
	// other flags become superfluous.
	RegexFlags_Literal RegexFlags = 16

	// Control behavior of "$" and "^". If set, recognize line terminators within string, otherwise, match only at start
	// and end of input string.
	RegexFlags_Multiline RegexFlags = 8

	// Unix-only line endings. When this mode is enabled, only '\n' is recognized as a line ending in the behavior
	// of ., ^, and $.
	RegexFlags_Unix_Lines RegexFlags = 1

	// Unicode word boundaries. If set, \b uses the Unicode TR 29 definition of word boundaries. Warning: Unicode word
	// boundaries are quite different from traditional regular expression word boundaries.
	// See http://unicode.org/reports/tr29/#Word_Boundaries
	RegexFlags_Unicode_Word RegexFlags = 256

	// Error on Unrecognized backslash escapes. If set, fail with an error on patterns that contain backslash-escaped
	// ASCII letters without a known special meaning. If this flag is not set, these escaped letters represent
	// themselves.
	RegexFlags_Error_On_Unknown_Escapes RegexFlags = 512
)

// CreateRegex creates a Regex, with a region of memory that has been preallocated to support strings that are less than
// or equal to the given size. Such strings will skip the allocation and deallocation phases, which save time. A size of
// zero will force all strings to be allocated and deallocated. The buffer is defined for one string, therefore double
// the amount given will actually be consumed (regex and match strings). Once the Regex is done with, you must remember
// to call Close. This Regex is intended for single-threaded use only, therefore it is advised for each thread to use
// its own Regex when one is needed.
func CreateRegex(stringBufferInBytes uint32) Regex {
	mod := modulePool.Get()
	pr := &privateRegex{
		mod:             mod,
		regexPtr:        0,
		regexStrUPtr:    0,
		matchStrUPtr:    0,
		matchStrUPtrLen: 0,

		bufferSize:     stringBufferInBytes,
		regexStrBuffer: 0,
		matchStrBuffer: 0,

		g_globalStackVar: mod.ExportedGlobal("globalStackVar").(api.MutableGlobal),

		f_malloc:                   mod.ExportedFunction("malloc"),
		f_free:                     mod.ExportedFunction("free"),
		f_replace:                  mod.ExportedFunction("replace"),
		f_uregex_open:              mod.ExportedFunction("uregex_open_68"),
		f_uregex_close:             mod.ExportedFunction("uregex_close_68"),
		f_uregex_start:             mod.ExportedFunction("uregex_start_68"),
		f_uregex_end:               mod.ExportedFunction("uregex_end_68"),
		f_uregex_find:              mod.ExportedFunction("uregex_find_68"),
		f_uregex_findNext:          mod.ExportedFunction("uregex_findNext_68"),
		f_uregex_getText:           mod.ExportedFunction("uregex_getText_68"),
		f_uregex_setText:           mod.ExportedFunction("uregex_setText_68"),
		f_uregex_replaceFirst:      mod.ExportedFunction("uregex_replaceFirst_68"),
		f_uregex_replaceAll:        mod.ExportedFunction("uregex_replaceAll_68"),
		f_uregex_appendReplacement: mod.ExportedFunction("uregex_appendReplacement_68"),
		f_uregex_appendTail:        mod.ExportedFunction("uregex_appendTail_68"),
		f_u_strToUTF8:              mod.ExportedFunction("u_strToUTF8_68"),
		f_u_strFromUTF8:            mod.ExportedFunction("u_strFromUTF8_68"),
	}
	// If we're creating string buffers, then we'll preallocate them
	if stringBufferInBytes > 0 {
		ctx := context.Background()
		regexStrBuffer, err := pr.malloc(ctx, stringBufferInBytes)
		if err != nil || regexStrBuffer == 0 {
			// If we get an error or couldn't allocate the buffer, then we'll just disable the string buffer
			pr.bufferSize = 0
		} else {
			matchStrBuffer, err := pr.malloc(ctx, stringBufferInBytes)
			if err != nil || matchStrBuffer == 0 {
				pr.bufferSize = 0
				if err = pr.free(ctx, regexStrBuffer); err != nil {
					panic(err) // This shouldn't fail, so we'll panic here, because something is very wrong
				}
			} else {
				pr.regexStrBuffer = UCharPtr(regexStrBuffer)
				pr.matchStrBuffer = UCharPtr(matchStrBuffer)
			}
		}
	}
	// This finalizer will let us know if a user never called Close. Although the module would eventually be reclaimed
	// by GC, this finalizer ensures that regexes are being used as efficiently as possible by maximizing pool rotations.
	// Hopefully, this would be caught during development and not in production.
	runtime.SetFinalizer(pr, func(pr *privateRegex) {
		if pr.mod != nil && ShouldPanic {
			panic("Finalizer found a Regex that was never closed")
		}
	})
	return pr
}

// privateRegex is the private implementation of the Regex interface.
type privateRegex struct {
	mod             api.Module
	regexPtr        URegularExpressionPtr
	regexStrUPtr    UCharPtr
	matchStrUPtr    UCharPtr
	matchStrUPtrLen int
	callStack       [8]uint64

	// Buffer details
	bufferSize     uint32
	regexStrBuffer UCharPtr
	matchStrBuffer UCharPtr

	// Global Variables
	g_globalStackVar api.MutableGlobal
	// Functions
	f_malloc                   api.Function
	f_free                     api.Function
	f_replace                  api.Function
	f_uregex_open              api.Function
	f_uregex_close             api.Function
	f_uregex_start             api.Function
	f_uregex_end               api.Function
	f_uregex_find              api.Function
	f_uregex_findNext          api.Function
	f_uregex_getText           api.Function
	f_uregex_setText           api.Function
	f_uregex_replaceFirst      api.Function
	f_uregex_replaceAll        api.Function
	f_uregex_appendReplacement api.Function
	f_uregex_appendTail        api.Function
	f_u_strToUTF8              api.Function
	f_u_strFromUTF8            api.Function
}

var _ Regex = (*privateRegex)(nil)

// SetRegexString implements the interface Regex.
func (pr *privateRegex) SetRegexString(ctx context.Context, regexStr string, flags RegexFlags) (err error) {
	// Free any previously-set regex strings
	if err = pr.closeRegexPtrs(); err != nil {
		return err
	}

	// Convert regexStr to UTF16LE and then copy it to WASM memory
	utf16RegexStr, regexStrULen := toUTF16(regexStr)
	if uint32(regexStrULen*2) <= pr.bufferSize {
		pr.regexStrUPtr = pr.regexStrBuffer
	} else {
		regexStrUPtr, err := pr.malloc(ctx, uint32(regexStrULen*2))
		if err != nil {
			return err
		}
		pr.regexStrUPtr = UCharPtr(regexStrUPtr)
	}
	pr.mod.Memory().Write(uint32(pr.regexStrUPtr), utf16RegexStr)

	// Create the URegularExpression*
	errorCode := UErrorCode(0)
	regex, err := pr.uregex_open(ctx, pr.regexStrUPtr, regexStrULen, uint32(flags), &errorCode)
	if err != nil {
		return err
	}
	if errorCode > 0 {
		return ErrInvalidRegex.New()
	}
	pr.regexPtr = regex
	return nil
}

// SetMatchString implements the interface Regex.
func (pr *privateRegex) SetMatchString(ctx context.Context, matchStr string) (err error) {
	// Check for the regex pointer first
	if pr.regexPtr == 0 {
		return ErrRegexNotYetSet.New()
	}

	// Reset the match string pointer if necessary
	if err = pr.closeMatchPtr(); err != nil {
		return err
	}

	// Convert matchStr to UTF16LE and then copy it to WASM memory
	utf16MatchStr, matchStrULen := toUTF16(matchStr)
	if uint32(matchStrULen*2) <= pr.bufferSize {
		pr.matchStrUPtr = pr.matchStrBuffer
	} else {
		matchStrUPtr, err := pr.malloc(ctx, uint32(matchStrULen*2))
		if err != nil {
			return err
		}
		pr.matchStrUPtr = UCharPtr(matchStrUPtr)
	}
	pr.matchStrUPtrLen = matchStrULen
	pr.mod.Memory().Write(uint32(pr.matchStrUPtr), utf16MatchStr)

	// Set the text on the URegularExpression*
	errorCode := UErrorCode(0)
	err = pr.uregex_setText(ctx, pr.regexPtr, pr.matchStrUPtr, matchStrULen, &errorCode)
	if err != nil {
		return err
	}
	if errorCode > 0 {
		return fmt.Errorf("unexpected UErrorCode from uregex_setText: %d", errorCode)
	}
	return nil
}

// Matches implements the interface Regex.
func (pr *privateRegex) Matches(ctx context.Context, start int, occurrence int) (ok bool, err error) {
	// Check for the regex pointer first
	if pr.regexPtr == 0 {
		return false, ErrRegexNotYetSet.New()
	}

	// Check that the match string has been set
	if pr.matchStrUPtr == 0 {
		return false, ErrMatchNotYetSet.New()
	}

	// Return if we found a match
	var errorCode UErrorCode
	ok, err = pr.uregex_find(ctx, pr.regexPtr, start, &errorCode)
	if err != nil {
		return false, err
	}
	for i := 1; i < occurrence && ok; i++ {
		ok, err = pr.uregex_findNext(ctx, pr.regexPtr, &errorCode)
		if err != nil {
			return false, err
		}
	}
	if errorCode > 0 {
		return false, fmt.Errorf("unexpected UErrorCode from uregex_find/uregex_findNext: %d", errorCode)
	}
	return ok, err
}

// Replace implements the interface Regex.
func (pr *privateRegex) Replace(ctx context.Context, replacementStr string, start int, occurrence int) (replacedStr string, err error) {
	// Check for the regex pointer first
	if pr.regexPtr == 0 {
		return "", ErrRegexNotYetSet.New()
	}

	// Check that the match string has been set
	if pr.matchStrUPtr == 0 {
		return "", ErrMatchNotYetSet.New()
	}

	// Convert replacementStr to UTF16LE and then copy it to WASM memory
	//TODO: should probably copy this once and enforce that it has been set before running (also free & reset in Close)
	utf16ReplacementStr, replacementStrULen := toUTF16(replacementStr)
	replacementStrUPtr, err := pr.malloc(ctx, uint32(replacementStrULen*2))
	if err != nil {
		return "", err
	}
	defer func() {
		if fErr := pr.free(ctx, replacementStrUPtr); err == nil {
			err = fErr
		}
	}()
	pr.mod.Memory().Write(replacementStrUPtr, utf16ReplacementStr)

	// Move to the starting position
	var returnSize int
	returnStr, err := pr.replace(ctx, pr.regexPtr, UCharPtr(replacementStrUPtr), replacementStrULen, pr.matchStrUPtr, pr.matchStrUPtrLen, start-1, occurrence, &returnSize)
	if err != nil {
		return "", err
	}
	defer func() {
		if fErr := pr.free(ctx, uint32(returnStr)); err == nil {
			err = fErr
		}
	}()
	returnStrBytes, ok := pr.mod.Memory().Read(uint32(returnStr), uint32(returnSize*2))
	if !ok {
		return "", fmt.Errorf("somehow failed when retrieving the string with replacements")
	}
	return fromUTF16(returnStrBytes), nil
}

// StringBufferSize implements the interface Regex.
func (pr *privateRegex) StringBufferSize() uint32 {
	return pr.bufferSize
}

// Close implements the interface Regex.
func (pr *privateRegex) Close() (err error) {
	if pr == nil || pr.mod == nil {
		return nil
	}
	err = pr.closeRegexPtrs()
	if nErr := pr.closeMatchPtr(); err == nil {
		err = nErr
	}
	// As we do not free the buffers in the other close functions (since they may be called without intending to close
	// the regex as a whole), we take care of freeing them here.
	if pr.bufferSize > 0 {
		ctx := context.Background()
		if nErr := pr.free(ctx, uint32(pr.regexStrBuffer)); err == nil {
			err = nErr
		}
		if nErr := pr.free(ctx, uint32(pr.matchStrBuffer)); err == nil {
			err = nErr
		}
	}
	if pr.mod != nil {
		modulePool.Put(pr.mod)
		pr.mod = nil
		runtime.SetFinalizer(pr, nil)
	}
	return err
}

// closeRegexPtr closes the regex pointers if they exist. This will not free the string buffer if it is being used.
func (pr *privateRegex) closeRegexPtrs() (err error) {
	ctx := context.Background()
	if pr.regexPtr != 0 {
		err = pr.uregex_close(ctx, pr.regexPtr)
	}
	if pr.regexStrUPtr != pr.regexStrBuffer && pr.regexStrUPtr != 0 {
		if freeErr := pr.free(ctx, uint32(pr.regexStrUPtr)); err == nil {
			err = freeErr
		}
	}
	pr.regexPtr = 0
	pr.regexStrUPtr = 0
	return err
}

// closeMatchPtr closes the match string pointer if it exists. This will not free the string buffer if it is being used.
func (pr *privateRegex) closeMatchPtr() (err error) {
	if pr.matchStrUPtr != pr.matchStrBuffer && pr.matchStrUPtr != 0 {
		err = pr.free(context.Background(), uint32(pr.matchStrUPtr))
	}
	pr.matchStrUPtr = 0
	pr.matchStrUPtrLen = 0
	return err
}

// toUTF16 returns a byte slice that contains the given string converted to UTF16LE, which is required for use with the
// ICU library. The length returned is the length that should be passed to ICU functions.
func toUTF16(str string) (convertedString []byte, length int) {
	// ICU complains about the lack of NULL-termination, but it works anyway. Adding NULL at the end seems to
	// prevent some matches (both a single NULL and two NULLs since it's a char16_t array). We just ignore the warning.
	utf16Raw := utf16.Encode([]rune(str))
	length = len(utf16Raw)
	convertedString = make([]byte, length*2)
	for i := 0; i < length; i++ {
		convertedString[2*i] = byte(utf16Raw[i])
		convertedString[(2*i)+1] = byte(utf16Raw[i] >> 8)
	}
	return
}

// fromUTF16 returns a string from a byte slice that contains a string in the UTF16LE format, which is how strings will
// be returned from the ICU library.
func fromUTF16(convertedString []byte) string {
	utf16Raw := make([]uint16, len(convertedString)/2)
	for i := 0; i < len(utf16Raw); i++ {
		utf16Raw[i] = uint16(convertedString[2*i]) | (uint16(convertedString[(2*i)+1]) << 8)
	}
	return string(utf16.Decode(utf16Raw))
}
