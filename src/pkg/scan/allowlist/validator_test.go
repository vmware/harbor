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

package allowlist

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsInvalidErr(t *testing.T) {
	cases := []struct {
		instance error
		expect   bool
	}{
		{
			instance: nil,
			expect:   false,
		},
		{
			instance: fmt.Errorf("whatever"),
			expect:   false,
		},
		{
			instance: NewInvalidErr("This is true"),
			expect:   true,
		},
	}

	for n, c := range cases {
		t.Logf("Executing TestIsInvalidErr case: %d\n", n)
		assert.Equal(t, c.expect, IsInvalidErr(c.instance))
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		l       models.CVEAllowlist
		noError bool
	}{
		{
			l: models.CVEAllowlist{
				Items: nil,
			},
			noError: true,
		},
		{
			l: models.CVEAllowlist{
				Items: []models.CVEAllowlistItem{},
			},
			noError: true,
		},
		{
			l: models.CVEAllowlist{
				Items: []models.CVEAllowlistItem{
					{CVEID: "breakit"},
					{CVEID: "breakit"},
				},
			},
			noError: false,
		},
		{
			l: models.CVEAllowlist{
				Items: []models.CVEAllowlistItem{
					{CVEID: "CVE-2014-456132"},
					{CVEID: "CVE-2014-7654321"},
				},
			},
			noError: true,
		},
		{
			l: models.CVEAllowlist{
				Items: []models.CVEAllowlistItem{
					{CVEID: "CVE-2014-456132"},
					{CVEID: "CVE-2014-456132"},
					{CVEID: "CVE-2014-7654321"},
				},
			},
			noError: false,
		},
	}
	for n, c := range cases {
		t.Logf("Executing TestValidate case: %d\n", n)
		e := Validate(c.l)
		assert.Equal(t, c.noError, e == nil)
		if e != nil {
			assert.True(t, IsInvalidErr(e))
		}
	}
}
