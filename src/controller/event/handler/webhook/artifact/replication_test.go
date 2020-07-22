package artifact

import (
	"context"
	"testing"
	"time"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/replication"
	daoModels "github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakedNotificationPolicyMgr struct {
}

type fakedReplicationPolicyMgr struct {
}

type fakedReplicationMgr struct {
}

type fakedReplicationRegistryMgr struct {
}

type fakedProjectCtl struct {
}

func (f *fakedNotificationPolicyMgr) Create(*models.NotificationPolicy) (int64, error) {
	return 0, nil
}

// List the policies, returns the policy list and error
func (f *fakedNotificationPolicyMgr) List(int64) ([]*models.NotificationPolicy, error) {
	return nil, nil
}

// Get policy with specified ID
func (f *fakedNotificationPolicyMgr) Get(int64) (*models.NotificationPolicy, error) {
	return nil, nil
}

// GetByNameAndProjectID get policy by the name and projectID
func (f *fakedNotificationPolicyMgr) GetByNameAndProjectID(string, int64) (*models.NotificationPolicy, error) {
	return nil, nil
}

// Update the specified policy
func (f *fakedNotificationPolicyMgr) Update(*models.NotificationPolicy) error {
	return nil
}

// Delete the specified policy
func (f *fakedNotificationPolicyMgr) Delete(int64) error {
	return nil
}

// Test the specified policy
func (f *fakedNotificationPolicyMgr) Test(*models.NotificationPolicy) error {
	return nil
}

// GetRelatedPolices get event type related policies in project
func (f *fakedNotificationPolicyMgr) GetRelatedPolices(int64, string) ([]*models.NotificationPolicy, error) {
	return []*models.NotificationPolicy{
		{
			ID: 0,
		},
	}, nil
}

func (f *fakedReplicationMgr) StartReplication(policy *model.Policy, resource *model.Resource, trigger model.TriggerType) (int64, error) {
	return 0, nil
}
func (f *fakedReplicationMgr) StopReplication(int64) error {
	return nil
}
func (f *fakedReplicationMgr) ListExecutions(...*daoModels.ExecutionQuery) (int64, []*daoModels.Execution, error) {
	return 0, nil, nil
}
func (f *fakedReplicationMgr) GetExecution(int64) (*daoModels.Execution, error) {
	return &daoModels.Execution{
		PolicyID: 1,
		Trigger:  "manual",
	}, nil
}
func (f *fakedReplicationMgr) ListTasks(...*daoModels.TaskQuery) (int64, []*daoModels.Task, error) {
	return 0, nil, nil
}
func (f *fakedReplicationMgr) GetTask(id int64) (*daoModels.Task, error) {
	if id == 1 {
		return &daoModels.Task{
			ExecutionID: 1,
			// project info not included when replicating with docker registry
			SrcResource: "alpine:[v1]",
			DstResource: "gxt/alpine:[v1] ",
		}, nil
	}
	return &daoModels.Task{
		ExecutionID: 1,
		SrcResource: "library/alpine:[v1]",
		DstResource: "gxt/alpine:[v1] ",
	}, nil
}
func (f *fakedReplicationMgr) UpdateTaskStatus(id int64, status string, statusRevision int64, statusCondition ...string) error {
	return nil
}
func (f *fakedReplicationMgr) GetTaskLog(int64) ([]byte, error) {
	return nil, nil
}

// Create new policy
func (f *fakedReplicationPolicyMgr) Create(*model.Policy) (int64, error) {
	return 0, nil
}

// List the policies, returns the total count, policy list and error
func (f *fakedReplicationPolicyMgr) List(...*model.PolicyQuery) (int64, []*model.Policy, error) {
	return 0, nil, nil
}

// Get policy with specified ID
func (f *fakedReplicationPolicyMgr) Get(int64) (*model.Policy, error) {
	return &model.Policy{
		ID: 1,
		SrcRegistry: &model.Registry{
			ID: 0,
		},
		DestRegistry: &model.Registry{
			ID: 0,
		},
	}, nil
}

// Get policy by the name
func (f *fakedReplicationPolicyMgr) GetByName(string) (*model.Policy, error) {
	return nil, nil
}

// Update the specified policy
func (f *fakedReplicationPolicyMgr) Update(policy *model.Policy) error {
	return nil
}

// Remove the specified policy
func (f *fakedReplicationPolicyMgr) Remove(int64) error {
	return nil
}

// Add new registry
func (f *fakedReplicationRegistryMgr) Add(*model.Registry) (int64, error) {
	return 0, nil
}

// List registries, returns total count, registry list and error
func (f *fakedReplicationRegistryMgr) List(query *q.Query) (int64, []*model.Registry, error) {
	return 0, nil, nil
}

// Get the specified registry
func (f *fakedReplicationRegistryMgr) Get(int64) (*model.Registry, error) {
	return &model.Registry{
		Type: "harbor",
		Credential: &model.Credential{
			Type: "local",
		},
	}, nil
}

// GetByName gets registry by name
func (f *fakedReplicationRegistryMgr) GetByName(name string) (*model.Registry, error) {
	return nil, nil
}

// Update the registry, the "props" are the properties of registry
// that need to be updated
func (f *fakedReplicationRegistryMgr) Update(registry *model.Registry, props ...string) error {
	return nil
}

// Remove the registry with the specified ID
func (f *fakedReplicationRegistryMgr) Remove(int64) error {
	return nil
}

// HealthCheck checks health status of all registries and update result in database
func (f *fakedReplicationRegistryMgr) HealthCheck() error {
	return nil
}

func (f *fakedProjectCtl) Get(ctx context.Context, projectID int64, options ...project.Option) (*models.Project, error) {
	return &models.Project{ProjectID: 1}, nil
}
func (f *fakedProjectCtl) GetByName(ctx context.Context, projectName string, options ...project.Option) (*models.Project, error) {
	return &models.Project{ProjectID: 1}, nil
}
func (f *fakedProjectCtl) List(ctx context.Context, query *models.ProjectQueryParam, options ...project.Option) ([]*models.Project, error) {
	return nil, nil
}

func TestReplicationHandler_Handle(t *testing.T) {
	common_dao.PrepareTestForPostgresSQL()
	config.Init()

	PolicyMgr := notification.PolicyMgr
	execution := replication.OperationCtl
	rpPolicy := replication.PolicyCtl
	rpRegistry := replication.RegistryMgr
	prj := project.Ctl

	defer func() {
		notification.PolicyMgr = PolicyMgr
		replication.OperationCtl = execution
		replication.PolicyCtl = rpPolicy
		replication.RegistryMgr = rpRegistry
		project.Ctl = prj
	}()
	notification.PolicyMgr = &fakedNotificationPolicyMgr{}
	replication.OperationCtl = &fakedReplicationMgr{}
	replication.PolicyCtl = &fakedReplicationPolicyMgr{}
	replication.RegistryMgr = &fakedReplicationRegistryMgr{}
	project.Ctl = &fakedProjectCtl{}

	handler := &ReplicationHandler{}

	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ReplicationHandler Want Error 1",
			args: args{
				data: "",
			},
			wantErr: true,
		},
		{
			name: "ReplicationHandler 1",
			args: args{
				data: &event.ReplicationEvent{
					OccurAt: time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "ReplicationHandler with docker registry",
			args: args{
				data: &event.ReplicationEvent{
					OccurAt:           time.Now(),
					ReplicationTaskID: 1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(tt.args.data)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Nil(t, err)
		})
	}

}

func TestReplicationHandler_IsStateful(t *testing.T) {
	handler := &ReplicationHandler{}
	assert.False(t, handler.IsStateful())
}
