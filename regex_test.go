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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegexMatch(t *testing.T) {
	ctx := context.Background()
	regex := CreateRegex()
	require.NoError(t, regex.SetRegexString(ctx, `abc+.*this st`, RegexFlags_None))
	err := regex.SetMatchString(ctx, "Find the abc in this string")
	require.NoError(t, err)
	ok, err := regex.Matches(ctx, 0, 0)
	require.NoError(t, err)
	require.True(t, ok)
	err = regex.SetMatchString(ctx, "Find the abc in this here string")
	require.NoError(t, err)
	ok, err = regex.Matches(ctx, 0, 0)
	require.NoError(t, err)
	require.False(t, ok)
	require.NoError(t, regex.Close())

	regex = CreateRegex()
	require.NoError(t, regex.SetRegexString(ctx, `[a-zA-Z0-9]{5} \w{4} aab`, RegexFlags_None))
	err = regex.SetMatchString(ctx, "Words like aab don't exist")
	require.NoError(t, err)
	ok, err = regex.Matches(ctx, 0, 0)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, regex.Close())

	regex = CreateRegex()
	require.NoError(t, regex.SetRegexString(ctx, `^[aA]bcd[eE]$`, RegexFlags_None))
	err = regex.SetMatchString(ctx, "abcde")
	require.NoError(t, err)
	ok, err = regex.Matches(ctx, 0, 0)
	require.NoError(t, err)
	require.True(t, ok)
	err = regex.SetMatchString(ctx, "Abcde")
	require.NoError(t, err)
	ok, err = regex.Matches(ctx, 0, 0)
	require.NoError(t, err)
	require.True(t, ok)
	err = regex.SetMatchString(ctx, "AbcdE")
	require.NoError(t, err)
	ok, err = regex.Matches(ctx, 0, 0)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, regex.Close())
}

func TestRegexReplace(t *testing.T) {
	ctx := context.Background()
	regex := CreateRegex()
	require.NoError(t, regex.SetRegexString(ctx, `[a-z]+`, RegexFlags_None))
	err := regex.SetMatchString(ctx, "abc def ghi")
	require.NoError(t, err)
	replacedStr, err := regex.Replace(ctx, "X", 1, 2)
	require.NoError(t, err)
	require.Equal(t, "abc X ghi", replacedStr)
	replacedStr, err = regex.Replace(ctx, "X", 1, 3)
	require.NoError(t, err)
	require.Equal(t, "abc def X", replacedStr)
	replacedStr, err = regex.Replace(ctx, "X", 1, 0)
	require.NoError(t, err)
	require.Equal(t, "X X X", replacedStr)
	require.NoError(t, regex.Close())
}
