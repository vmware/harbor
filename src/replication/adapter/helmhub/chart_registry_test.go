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

func TestFetchCharts(t *testing.T) {
	registry := &model.Registry{
		Type:     model.RegistryTypeHelmHub,
		URL:      baseURL,
		Insecure: false,
	}
	adapter, err := newAdapter(registry)
	require.Nil(t, err)
	// filter 1
	filters := []*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "k*/*",
		},
	}
	resources, err := adapter.FetchCharts(filters)
	require.Nil(t, err)
	assert.NotZero(t, len(resources))
	assert.Equal(t, model.ResourceTypeChart, resources[0].Type)
	assert.Equal(t, 1, len(resources[0].Metadata.Artifacts))
	assert.NotNil(t, resources[0].Metadata.Artifacts[0].Tags[0])
	// filter 2
	filters = []*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "harbor/*",
		},
	}
	resources, err = adapter.FetchCharts(filters)
	require.Nil(t, err)
	assert.NotZero(t, len(resources))
	assert.Equal(t, model.ResourceTypeChart, resources[0].Type)
	assert.Equal(t, "harbor/harbor", resources[0].Metadata.Repository.Name)
	assert.Equal(t, 1, len(resources[0].Metadata.Artifacts))
	assert.NotNil(t, resources[0].Metadata.Artifacts[0].Tags[0])
}

func TestChartExist(t *testing.T) {
	registry := &model.Registry{
		Type:     model.RegistryTypeHelmHub,
		URL:      baseURL,
		Insecure: false,
	}
	adapter, err := newAdapter(registry)
	require.Nil(t, err)
	exist, err := adapter.ChartExist("harbor/harbor", "1.0.0")
	require.Nil(t, err)
	require.True(t, exist)
}

func TestChartExist2(t *testing.T) {
	registry := &model.Registry{
		Type:     model.RegistryTypeHelmHub,
		URL:      baseURL,
		Insecure: false,
	}
	adapter, err := newAdapter(registry)
	require.Nil(t, err)
	exist, err := adapter.ChartExist("goharbor/harbor", "1.0.0")
	require.Nil(t, err)
	require.False(t, exist)

	exist, err = adapter.ChartExist("harbor/harbor", "1.0.100")
	require.Nil(t, err)
	require.False(t, exist)
}

func TestDownloadChart(t *testing.T) {
	registry := &model.Registry{
		Type:     model.RegistryTypeHelmHub,
		URL:      baseURL,
		Insecure: false,
	}
	adapter, err := newAdapter(registry)
	require.Nil(t, err)
	_, err = adapter.DownloadChart("harbor/harbor", "1.0.0")
	require.Nil(t, err)
}

func TestUploadChart(t *testing.T) {
	adapter := &adapter{}
	err := adapter.UploadChart("library/harbor", "1.0", nil)
	require.NotNil(t, err)
}

func TestDeleteChart(t *testing.T) {
	adapter := &adapter{}
	err := adapter.DeleteChart("library/harbor", "1.0")
	require.NotNil(t, err)
}
