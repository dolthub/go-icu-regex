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
)

type URegularExpressionPtr uint32
type UCharPtr uint32
type UErrorCode int32
type CharPtr int32

// void* malloc(size_t size)
func (pr *privateRegex) malloc(ctx context.Context, sz uint32) (uint32, error) {
	pr.callStack[0] = uint64(sz)
	err := pr.f_malloc.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return 0, err
	}
	return uint32(pr.callStack[0]), nil
}

// void free(void* ptr)
func (pr *privateRegex) free(ctx context.Context, ptr uint32) error {
	pr.callStack[0] = uint64(ptr)
	return pr.f_free.CallWithStack(ctx, pr.callStack[:])
}

// UChar* replace(URegularExpression* regexp, UChar* replacement, int replacementLen, UChar* original, int originalSize, int start, int occurrence, int* returnSize)
func (pr *privateRegex) replace(ctx context.Context, regex URegularExpressionPtr, replacement UCharPtr, replacementLen int, original UCharPtr, originalSize int, start int, occurrence int, returnSize *int) (returnStr UCharPtr, err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()
	returnSizeAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(returnSizeAddr), uint32(*returnSize))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(returnSizeAddr))
		if !ok {
			err = fmt.Errorf("could not read the return size")
		}
		*returnSize = int(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(regex), uint64(replacement), uint64(replacementLen), uint64(original), uint64(originalSize), uint64(start), uint64(occurrence), returnSizeAddr})
	err = pr.f_replace.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return 0, err
	}
	return UCharPtr(pr.callStack[0]), err
}

// URegularExpression* uregex_open(const UChar* pattern, int32_t patternLength, uint32_t flags, UErrorCode* status);
func (pr *privateRegex) uregex_open(ctx context.Context, str UCharPtr, strlen int, flags uint32, uerr *UErrorCode) (ptr URegularExpressionPtr, err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()
	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(str), uint64(strlen), uint64(flags), uint64(0), uerrAddr})
	err = pr.f_uregex_open.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return 0, err
	}
	return URegularExpressionPtr(pr.callStack[0]), nil
}

// void uregex_close(URegularExpression* regexp)
func (pr *privateRegex) uregex_close(ctx context.Context, p URegularExpressionPtr) error {
	pr.callStack[0] = uint64(p)
	return pr.f_uregex_close.CallWithStack(ctx, pr.callStack[:])
}

// int32_t uregex_start(URegularExpression *regexp, int32_t groupNum, UErrorCode* status)
func (pr *privateRegex) uregex_start(ctx context.Context, regex URegularExpressionPtr, group int, uerr *UErrorCode) (idx int32, err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()
	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(regex), uint64(group), uerrAddr})
	err = pr.f_uregex_start.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return 0, err
	}
	return int32(pr.callStack[0]), nil
}

// int32_t uregex_end(URegularExpression* regexp, int32_t groupNum, UErrorCode* status)
func (pr *privateRegex) uregex_end(ctx context.Context, regex URegularExpressionPtr, group int, uerr *UErrorCode) (idx int32, err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()
	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(regex), uint64(group), uerrAddr})
	err = pr.f_uregex_end.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return 0, err
	}
	return int32(pr.callStack[0]), nil
}

// UBool uregex_find(URegularExpression* regexp, int32_t startIndex, UErrorCode* status)
func (pr *privateRegex) uregex_find(ctx context.Context, regex URegularExpressionPtr, startIndex int, uerr *UErrorCode) (ok bool, err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()
	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(regex), uint64(startIndex), uerrAddr})
	err = pr.f_uregex_find.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return false, err
	}
	return pr.callStack[0] != 0, nil
}

// UBool uregex_findNext(URegularExpression* regexp, UErrorCode* status)
func (pr *privateRegex) uregex_findNext(ctx context.Context, regex URegularExpressionPtr, uerr *UErrorCode) (ok bool, err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()
	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(regex), uerrAddr})
	err = pr.f_uregex_findNext.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return false, err
	}
	return pr.callStack[0] != 0, nil
}

// UChar* uregex_getText(URegularExpression *regexp, int32_t* textLength, UErrorCode* status)
func (pr *privateRegex) uregex_getText(ctx context.Context, p URegularExpressionPtr, textLength *int, uerr *UErrorCode) (text UCharPtr, err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()

	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	textLengthAddr := origSP - 8
	pr.mod.Memory().WriteUint32Le(uint32(textLengthAddr), uint32(*textLength))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(textLengthAddr))
		if !ok {
			err = fmt.Errorf("could not read textLength")
		}
		*textLength = int(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(p), textLengthAddr, uerrAddr})
	err = pr.f_uregex_getText.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return 0, err
	}
	return UCharPtr(pr.callStack[0]), nil
}

// void uregex_setText(URegularExpression* regexp, const UChar* text, int32_t textLength, UErrorCode* status)
func (pr *privateRegex) uregex_setText(ctx context.Context, p URegularExpressionPtr, str UCharPtr, strlen int, uerr *UErrorCode) (err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()
	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(p), uint64(str), uint64(strlen), uerrAddr})
	return pr.f_uregex_setText.CallWithStack(ctx, pr.callStack[:])
}

// int32_t uregex_replaceFirst(URegularExpression* regexp, const UChar* replacementText, int32_t replacementLength, UChar* destBuf, int32_t destCapacity, UErrorCode* status);
func (pr *privateRegex) uregex_replaceFirst(ctx context.Context, p URegularExpressionPtr, replacementText UCharPtr, replacementLength int, destBuf UCharPtr, destCapacity int, uerr *UErrorCode) (resultLength int, err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()

	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(p), uint64(replacementText), uint64(replacementLength), uint64(destBuf), uint64(destCapacity), uerrAddr})
	err = pr.f_uregex_replaceFirst.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return 0, err
	}
	return int(pr.callStack[0]), nil
}

// int32_t uregex_replaceAll(URegularExpression* regexp, const UChar* replacementText, int32_t replacementLength, UChar* destBuf, int32_t destCapacity, UErrorCode* status);
func (pr *privateRegex) uregex_replaceAll(ctx context.Context, p URegularExpressionPtr, replacementText UCharPtr, replacementLength int, destBuf UCharPtr, destCapacity int, uerr *UErrorCode) (resultLength int, err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()

	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(p), uint64(replacementText), uint64(replacementLength), uint64(destBuf), uint64(destCapacity), uerrAddr})
	err = pr.f_uregex_replaceAll.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return 0, err
	}
	return int(pr.callStack[0]), nil
}

// int32_t uregex_appendReplacement(URegularExpression* regexp, UChar* replacementText, int32_t replacementLength, UChar** destBuf, int32_t* destCapacity, UErrorCode* status)
func (pr *privateRegex) uregex_appendReplacement(ctx context.Context, p URegularExpressionPtr, replacementText UCharPtr, replacementLength int, destBuf *UCharPtr, destCapacity *int, uerr *UErrorCode) (resultLength int, err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()

	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	destBufAddr := origSP - 8
	pr.mod.Memory().WriteUint32Le(uint32(destBufAddr), uint32(*destBuf))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(destBufAddr))
		if !ok {
			err = fmt.Errorf("could not read destBuf")
		}
		*destBuf = UCharPtr(res)
	}()

	destCapacityAddr := origSP - 12
	pr.mod.Memory().WriteUint32Le(uint32(destCapacityAddr), uint32(*destCapacity))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(destCapacityAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*destCapacity = int(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(p), uint64(replacementText), uint64(replacementLength), destBufAddr, destCapacityAddr, uerrAddr})
	err = pr.f_uregex_appendReplacement.CallWithStack(ctx, pr.callStack[:])
	if err != nil {
		return 0, err
	}
	return int(pr.callStack[0]), nil
}

// int32_t uregex_appendTail(URegularExpression* regexp, UChar** destBuf, int32_t* destCapacity, UErrorCode* status)
func (pr *privateRegex) uregex_appendTail(ctx context.Context, p URegularExpressionPtr, destBuf *UCharPtr, destCapacity *int, uerr *UErrorCode) (err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()

	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	destBufAddr := origSP - 8
	pr.mod.Memory().WriteUint32Le(uint32(destBufAddr), uint32(*destBuf))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(destBufAddr))
		if !ok {
			err = fmt.Errorf("could not read destBuf")
		}
		*destBuf = UCharPtr(res)
	}()

	destCapacityAddr := origSP - 12
	pr.mod.Memory().WriteUint32Le(uint32(destCapacityAddr), uint32(*destCapacity))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(destCapacityAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*destCapacity = int(res)
	}()

	copy(pr.callStack[:], []uint64{uint64(p), destBufAddr, destCapacityAddr, uerrAddr})
	return pr.f_uregex_appendTail.CallWithStack(ctx, pr.callStack[:])
}

// char* u_strToUTF8(char* dest, int32_t destCapacity, int32_t* pDestLength, const UChar* src, int32_t srcLength, UErrorCode* pErrorCode)
func (pr *privateRegex) u_strToUTF8(ctx context.Context, buff CharPtr, bufflen int, outlen *int, str UCharPtr, strlen int, uerr *UErrorCode) (err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()
	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	var outlenptr uint32 = 0
	if outlen != nil {
		outlenptr = uint32(origSP - 8)
		pr.mod.Memory().WriteUint32Le(outlenptr, uint32(*outlen))
		defer func() {
			res, ok := pr.mod.Memory().ReadUint32Le(outlenptr)
			if !ok {
				err = fmt.Errorf("could not read UErrorCode")
			}
			*outlen = int(res)
		}()
	}

	copy(pr.callStack[:], []uint64{uint64(buff), uint64(bufflen), uint64(outlenptr), uint64(str), uint64(strlen), uerrAddr})
	return pr.f_u_strToUTF8.CallWithStack(ctx, pr.callStack[:])
}

// UChar* u_strFromUTF8(UChar* dest, int32_t destCapacity, int32_t* pDestLength, const char* src, int32_t srcLength, UErrorCode* pErrorCode)
func (pr *privateRegex) u_strFromUTF8(ctx context.Context, buff UCharPtr, bufflen int, outlen *int, str CharPtr, strlen int, uerr *UErrorCode) (err error) {
	origSP := pr.g_globalStackVar.Get()
	pr.g_globalStackVar.Set(origSP - 16)
	defer func() { pr.g_globalStackVar.Set(origSP) }()

	uerrAddr := origSP - 4
	pr.mod.Memory().WriteUint32Le(uint32(uerrAddr), uint32(*uerr))
	defer func() {
		res, ok := pr.mod.Memory().ReadUint32Le(uint32(uerrAddr))
		if !ok {
			err = fmt.Errorf("could not read UErrorCode")
		}
		*uerr = UErrorCode(res)
	}()

	var outlenptr uint32 = 0
	if outlen != nil {
		outlenptr = uint32(origSP - 8)
		pr.mod.Memory().WriteUint32Le(outlenptr, uint32(*outlen))
		defer func() {
			res, ok := pr.mod.Memory().ReadUint32Le(outlenptr)
			if !ok {
				err = fmt.Errorf("could not read UErrorCode")
			}
			*outlen = int(res)
		}()
	}

	copy(pr.callStack[:], []uint64{uint64(buff), uint64(bufflen), uint64(outlenptr), uint64(str), uint64(strlen), uerrAddr})
	return pr.f_u_strFromUTF8.CallWithStack(ctx, pr.callStack[:])
}
