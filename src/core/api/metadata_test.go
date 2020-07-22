// Copyright 2018 Project Harbor Authors
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

package api

import (
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateProjectMetadata(t *testing.T) {
	var metas map[string]string

	// nil metas
	ms, err := validateProjectMetadata(metas)
	require.Nil(t, err)
	require.Nil(t, ms)

	// valid key, invalid value(bool)
	metas = map[string]string{
		models.ProMetaPublic: "invalid_value",
	}
	ms, err = validateProjectMetadata(metas)
	require.NotNil(t, err)

	// valid key/value(bool)
	metas = map[string]string{
		models.ProMetaPublic: "1",
	}
	ms, err = validateProjectMetadata(metas)
	require.Nil(t, err)
	assert.Equal(t, "true", ms[models.ProMetaPublic])

	// valid key, invalid value(string)
	metas = map[string]string{
		models.ProMetaSeverity: "invalid_value",
	}
	ms, err = validateProjectMetadata(metas)
	require.NotNil(t, err)

	// valid key, valid value(string)
	metas = map[string]string{
		models.ProMetaSeverity: "High",
	}
	ms, err = validateProjectMetadata(metas)
	require.Nil(t, err)
	assert.Equal(t, "high", ms[models.ProMetaSeverity])
}

func TestMetaAPI(t *testing.T) {
	client := newHarborAPI()

	// non-exist project, it should return 401 if user is not logged in.
	code, _, err := client.PostMeta(*unknownUsr, int64(1000), nil)
	require.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, code)

	// non-login
	code, _, err = client.PostMeta(*unknownUsr, int64(1), nil)
	require.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, code)

	// test post
	code, _, err = client.PostMeta(*admin, int64(1), map[string]string{
		models.ProMetaAutoScan: "true",
	})
	require.Nil(t, err)
	assert.Equal(t, http.StatusCreated, code)

	// test get
	code, metas, err := client.GetMeta(*admin, int64(1), models.ProMetaAutoScan)
	require.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "true", metas[models.ProMetaAutoScan])

	// test put
	code, _, err = client.PutMeta(*admin, int64(1), models.ProMetaAutoScan,
		map[string]string{
			models.ProMetaAutoScan: "false",
		})
	require.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)

	code, metas, err = client.GetMeta(*admin, int64(1), models.ProMetaAutoScan)
	require.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "false", metas[models.ProMetaAutoScan])

	// test delete
	code, _, err = client.DeleteMeta(*admin, int64(1), models.ProMetaAutoScan)
	require.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)

	code, metas, err = client.GetMeta(*admin, int64(1), models.ProMetaAutoScan)
	require.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, code)
}
