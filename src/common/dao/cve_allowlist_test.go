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

package dao

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUpdateAndGetCVEAllowlist(t *testing.T) {
	require.Nil(t, ClearTable("cve_allowlist"))
	l2, err := GetCVEAllowlist(5)
	assert.Nil(t, err)
	assert.Nil(t, l2)

	longList := []models.CVEAllowlistItem{}
	for i := 0; i < 50; i++ {
		longList = append(longList, models.CVEAllowlistItem{CVEID: "CVE-1999-0067"})
	}

	e := int64(1573254000)
	in1 := models.CVEAllowlist{ProjectID: 3, Items: longList, ExpiresAt: &e}
	_, err = UpdateCVEAllowlist(in1)
	require.Nil(t, err)
	// assert.Equal(t, int64(1), n)
	out1, err := GetCVEAllowlist(3)
	require.Nil(t, err)
	assert.Equal(t, int64(3), out1.ProjectID)
	assert.Equal(t, longList, out1.Items)
	assert.Equal(t, e, *out1.ExpiresAt)

	sysCVEs := []models.CVEAllowlistItem{
		{CVEID: "CVE-2019-10164"},
		{CVEID: "CVE-2017-12345"},
	}
	in3 := models.CVEAllowlist{Items: sysCVEs}
	_, err = UpdateCVEAllowlist(in3)
	require.Nil(t, err)

	require.Nil(t, ClearTable("cve_allowlist"))
}
