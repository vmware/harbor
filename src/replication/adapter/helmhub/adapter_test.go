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

package helmhub

import (
	"testing"

	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	adapter := &adapter{}
	info, err := adapter.Info()
	require.Nil(t, err)
	require.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, model.ResourceTypeChart, info.SupportedResourceTypes[0])
}

func TestPrepareForPush(t *testing.T) {
	adapter := &adapter{}
	err := adapter.PrepareForPush(nil)
	require.NotNil(t, err)
}

func TestHealthCheck(t *testing.T) {
	registry := &model.Registry{
		Type:     model.RegistryTypeHelmHub,
		URL:      baseURL,
		Insecure: true,
	}
	adapter, _ := newAdapter(registry)
	status, err := adapter.HealthCheck()
	require.Equal(t, model.Healthy, string(status))
	require.Nil(t, err)
}
