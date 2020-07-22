// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestCVEAllowlist_All(t *testing.T) {
	future := int64(4411494000)
	now := time.Now().Unix()
	cases := []struct {
		input   CVEAllowlist
		cveset  map[string]struct{}
		expired bool
	}{
		{
			input: CVEAllowlist{
				ID:        1,
				ProjectID: 0,
				Items:     []CVEAllowlistItem{},
			},
			cveset:  map[string]struct{}{},
			expired: false,
		},
		{
			input: CVEAllowlist{
				ID:        1,
				ProjectID: 0,
				Items:     []CVEAllowlistItem{},
				ExpiresAt: &now,
			},
			cveset:  map[string]struct{}{},
			expired: true,
		},
		{
			input: CVEAllowlist{
				ID:        2,
				ProjectID: 3,
				Items: []CVEAllowlistItem{
					{CVEID: "CVE-1999-0067"},
					{CVEID: "CVE-2016-7654321"},
				},
				ExpiresAt: &future,
			},
			cveset: map[string]struct{}{
				"CVE-1999-0067":    {},
				"CVE-2016-7654321": {},
			},
			expired: false,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expired, c.input.IsExpired())
		assert.True(t, reflect.DeepEqual(c.cveset, c.input.CVESet()))
	}
}
