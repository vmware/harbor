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

package event

import (
	"github.com/goharbor/harbor/src/lib/q"
	"testing"

	"github.com/goharbor/harbor/src/replication/config"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakedOperationController struct{}

func (f *fakedOperationController) StartReplication(policy *model.Policy, resource *model.Resource, trigger model.TriggerType) (int64, error) {
	return 1, nil
}
func (f *fakedOperationController) StopReplication(int64) error {
	return nil
}
func (f *fakedOperationController) ListExecutions(...*models.ExecutionQuery) (int64, []*models.Execution, error) {
	return 0, nil, nil
}
func (f *fakedOperationController) GetExecution(id int64) (*models.Execution, error) {
	return nil, nil
}
func (f *fakedOperationController) ListTasks(...*models.TaskQuery) (int64, []*models.Task, error) {
	return 0, nil, nil
}
func (f *fakedOperationController) GetTask(id int64) (*models.Task, error) {
	return nil, nil
}
func (f *fakedOperationController) UpdateTaskStatus(id int64, status string, statusRevision int64, statusCondition ...string) error {
	return nil
}
func (f *fakedOperationController) GetTaskLog(int64) ([]byte, error) {
	return nil, nil
}

type fakedPolicyController struct{}

func (f *fakedPolicyController) Create(*model.Policy) (int64, error) {
	return 0, nil
}
func (f *fakedPolicyController) List(...*model.PolicyQuery) (int64, []*model.Policy, error) {
	polices := []*model.Policy{
		{
			ID:       1,
			Enabled:  true,
			Deletion: true,
			Trigger: &model.Trigger{
				Type: model.TriggerTypeEventBased,
			},
			Filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "test/*",
				},
			},
			DestRegistry: &model.Registry{
				ID: 1,
			},
		},
		// nil trigger
		{
			ID:       2,
			Enabled:  true,
			Deletion: true,
			Trigger:  nil,
			Filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "library/*",
				},
			},
			DestRegistry: &model.Registry{
				ID: 1,
			},
		},
		// doesn't replicate deletion
		{
			ID:       3,
			Enabled:  true,
			Deletion: false,
			Trigger: &model.Trigger{
				Type: model.TriggerTypeEventBased,
			},
			Filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "library/*",
				},
			},
			DestRegistry: &model.Registry{
				ID: 1,
			},
		},
		// replicate deletion
		{
			ID:       4,
			Enabled:  true,
			Deletion: true,
			Trigger: &model.Trigger{
				Type: model.TriggerTypeEventBased,
			},
			Filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "library/*",
				},
			},
			DestRegistry: &model.Registry{
				ID: 1,
			},
		},
		// disabled
		{
			ID:       5,
			Enabled:  false,
			Deletion: true,
			Trigger: &model.Trigger{
				Type: model.TriggerTypeEventBased,
			},
			Filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "library/*",
				},
			},
			DestRegistry: &model.Registry{
				ID: 1,
			},
		},
		// the source registry is not local Harbor
		{
			ID:       6,
			Enabled:  true,
			Deletion: true,
			Trigger: &model.Trigger{
				Type: model.TriggerTypeEventBased,
			},
			Filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "library/*",
				},
			},
			SrcRegistry: &model.Registry{
				ID: 1,
			},
		},
	}
	return int64(len(polices)), polices, nil
}
func (f *fakedPolicyController) Get(id int64) (*model.Policy, error) {
	return nil, nil
}
func (f *fakedPolicyController) GetByName(name string) (*model.Policy, error) {
	return nil, nil
}
func (f *fakedPolicyController) Update(*model.Policy) error {
	return nil
}
func (f *fakedPolicyController) Remove(int64) error {
	return nil
}

type fakedRegistryManager struct{}

func (f *fakedRegistryManager) Add(*model.Registry) (int64, error) {
	return 0, nil
}
func (f *fakedRegistryManager) List(query *q.Query) (int64, []*model.Registry, error) {
	return 0, nil, nil
}
func (f *fakedRegistryManager) Get(id int64) (*model.Registry, error) {
	return &model.Registry{
		ID:   1,
		Type: model.RegistryTypeHarbor,
	}, nil
}
func (f *fakedRegistryManager) GetByName(name string) (*model.Registry, error) {
	return nil, nil
}
func (f *fakedRegistryManager) Update(*model.Registry, ...string) error {
	return nil
}
func (f *fakedRegistryManager) Remove(int64) error {
	return nil
}
func (f *fakedRegistryManager) HealthCheck() error {
	return nil
}
func TestGetRelatedPolicies(t *testing.T) {
	handler := &handler{
		policyCtl: &fakedPolicyController{},
	}
	policies, err := handler.getRelatedPolicies(&model.Resource{
		Metadata: &model.ResourceMetadata{
			Repository: &model.Repository{
				Name: "library/hello-world",
			},
		},
	})
	require.Nil(t, err)
	assert.Equal(t, 2, len(policies))
	assert.Equal(t, int64(3), policies[0].ID)
	assert.Equal(t, int64(4), policies[1].ID)

	policies, err = handler.getRelatedPolicies(&model.Resource{
		Metadata: &model.ResourceMetadata{
			Repository: &model.Repository{
				Name: "library/hello-world",
			},
		},
		Deleted: true,
	})
	require.Nil(t, err)
	assert.Equal(t, 1, len(policies))
	assert.Equal(t, int64(4), policies[0].ID)
}

func TestHandle(t *testing.T) {
	config.Config = &config.Configuration{}
	handler := NewHandler(&fakedPolicyController{},
		&fakedRegistryManager{},
		&fakedOperationController{})
	// nil event
	err := handler.Handle(nil)
	require.NotNil(t, err)

	// nil vtags
	err = handler.Handle(&Event{
		Resource: &model.Resource{
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{},
			},
		},
		Type: EventTypeArtifactPush,
	})
	require.NotNil(t, err)

	// unsupported event type
	err = handler.Handle(&Event{
		Resource: &model.Resource{
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
			},
		},
		Type: "unsupported",
	})
	require.NotNil(t, err)

	// push image
	err = handler.Handle(&Event{
		Resource: &model.Resource{
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Artifacts: []*model.Artifact{
					{
						Tags: []string{"latest"},
					},
				},
			},
		},
		Type: EventTypeArtifactPush,
	})
	require.Nil(t, err)

	// delete image
	err = handler.Handle(&Event{
		Resource: &model.Resource{
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Artifacts: []*model.Artifact{
					{
						Tags: []string{"latest"},
					},
				},
			},
		},
		Type: EventTypeArtifactDelete,
	})
	require.Nil(t, err)
}
