// Copyright 2026 Dolthub, Inc.
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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRegexLargeMatchString exercises each operation on a large match string.
// Only Replace builds a native result buffer, which previously over-read past
// its end and crashed once the result grew large enough to leave the heap. The
// other operations return a scalar or a correctly sized copy and are included
// to confirm the bug is limited to Replace. The single character pattern keeps
// the match string from exhausting the matcher's backtracking stack.
// See https://github.com/dolthub/dolt/issues/11183
func TestRegexLargeMatchString(t *testing.T) {
	const largeN = 2000000
	ctx := context.Background()
	for _, tt := range []struct {
		name  string
		check func(t *testing.T, re Regex)
	}{
		{"replace", func(t *testing.T, re Regex) {
			out, err := re.Replace(ctx, "b", 1, 0)
			require.NoError(t, err)
			require.Equal(t, largeN, len(out))
		}},
		{"substring", func(t *testing.T, re Regex) {
			out, ok, err := re.Substring(ctx, 1, 1)
			require.NoError(t, err)
			require.True(t, ok)
			require.Equal(t, "a", out)
		}},
		{"matches", func(t *testing.T, re Regex) {
			ok, err := re.Matches(ctx, 0, 0)
			require.NoError(t, err)
			require.True(t, ok)
		}},
		{"index_of", func(t *testing.T, re Regex) {
			idx, err := re.IndexOf(ctx, 1, 1, false)
			require.NoError(t, err)
			require.Equal(t, 1, idx)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			re := CreateRegex(512 * 1024)
			require.NoError(t, re.SetRegexString(ctx, "a", RegexFlags_None))
			require.NoError(t, re.SetMatchString(ctx, strings.Repeat("a", largeN)))
			tt.check(t, re)
			require.NoError(t, re.Close())
		})
	}
}
