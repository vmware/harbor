package preheat

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/lib/q"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"

	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	providerModel "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/testing/pkg/p2p/preheat/instance"
	pmocks "github.com/goharbor/harbor/src/testing/pkg/p2p/preheat/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type preheatSuite struct {
	suite.Suite
	ctx                context.Context
	controller         Controller
	fakeInstanceMgr    *instance.FakeManager
	fakePolicyMgr      *pmocks.FakeManager
	mockInstanceServer *httptest.Server
}

func TestPreheatSuite(t *testing.T) {
	t.Log("Start TestPreheatSuite")
	fakeInstanceMgr := &instance.FakeManager{}
	fakePolicyMgr := &pmocks.FakeManager{}

	var c = &controller{
		iManager: fakeInstanceMgr,
		pManager: fakePolicyMgr,
	}
	assert.NotNil(t, c)

	suite.Run(t, &preheatSuite{
		ctx:             context.Background(),
		controller:      c,
		fakeInstanceMgr: fakeInstanceMgr,
		fakePolicyMgr:   fakePolicyMgr,
	})
}

func TestNewController(t *testing.T) {
	c := NewController()
	assert.NotNil(t, c)
}

func (s *preheatSuite) SetupSuite() {
	config.Init()

	s.fakeInstanceMgr.On("List", mock.Anything, mock.Anything).Return([]*providerModel.Instance{
		{
			ID:       1,
			Vendor:   "dragonfly",
			Endpoint: "http://localhost",
			Status:   provider.DriverStatusHealthy,
			Enabled:  true,
		},
	}, nil)
	s.fakeInstanceMgr.On("Save", mock.Anything, mock.Anything).Return(int64(1), nil)
	s.fakeInstanceMgr.On("Count", mock.Anything, &q.Query{Keywords: map[string]interface{}{
		"endpoint": "http://localhost",
	}}).Return(int64(1), nil)
	s.fakeInstanceMgr.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil)
	s.fakeInstanceMgr.On("Delete", mock.Anything, int64(1)).Return(nil)
	s.fakeInstanceMgr.On("Delete", mock.Anything, int64(0)).Return(errors.New("not found"))
	s.fakeInstanceMgr.On("Get", mock.Anything, int64(1)).Return(&providerModel.Instance{
		ID:       1,
		Endpoint: "http://localhost",
	}, nil)
	s.fakeInstanceMgr.On("Get", mock.Anything, int64(0)).Return(nil, errors.New("not found"))

	// mock server for check health
	s.mockInstanceServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/_ping":
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			w.WriteHeader(http.StatusOK)
		}
	}))
	s.mockInstanceServer.Start()
}

// TearDownSuite clears the env.
func (s *preheatSuite) TearDownSuite() {
	s.mockInstanceServer.Close()
}

func (s *preheatSuite) TestGetAvailableProviders() {
	providers, err := s.controller.GetAvailableProviders()
	s.Equal(2, len(providers))
	expectProviders := map[string]interface{}{}
	expectProviders["dragonfly"] = nil
	expectProviders["kraken"] = nil
	_, ok := expectProviders[providers[0].ID]
	s.True(ok)
	_, ok = expectProviders[providers[1].ID]
	s.True(ok)
	s.NoError(err)
}

func (s *preheatSuite) TestListInstance() {
	instances, err := s.controller.ListInstance(s.ctx, nil)
	s.NoError(err)
	s.Equal(1, len(instances))
	s.Equal(int64(1), instances[0].ID)
}

func (s *preheatSuite) TestCreateInstance() {
	// Case: nil instance, expect error.
	id, err := s.controller.CreateInstance(s.ctx, nil)
	s.Empty(id)
	s.Error(err)

	// Case: instance with already existed endpoint, expect conflict.
	id, err = s.controller.CreateInstance(s.ctx, &providerModel.Instance{
		Endpoint: "http://localhost",
	})
	s.Equal(ErrorConflict, err)
	s.Empty(id)

	// Case: instance with invalid provider, expect error.
	id, err = s.controller.CreateInstance(s.ctx, &providerModel.Instance{
		Endpoint: "http://foo.bar",
		Status:   "healthy",
		Vendor:   "none",
	})
	s.NoError(err)
	s.Equal(int64(1), id)

	// Case: instance with valid provider, expect ok.
	id, err = s.controller.CreateInstance(s.ctx, &providerModel.Instance{
		Endpoint: "http://foo.bar",
		Status:   "healthy",
		Vendor:   "dragonfly",
	})
	s.NoError(err)
	s.Equal(int64(1), id)

	id, err = s.controller.CreateInstance(s.ctx, &providerModel.Instance{
		Endpoint: "http://foo.bar2",
		Status:   "healthy",
		Vendor:   "kraken",
	})
	s.NoError(err)
	s.Equal(int64(1), id)
}

func (s *preheatSuite) TestDeleteInstance() {
	err := s.controller.DeleteInstance(s.ctx, 0)
	s.Error(err)

	err = s.controller.DeleteInstance(s.ctx, int64(1))
	s.NoError(err)
}

func (s *preheatSuite) TestUpdateInstance() {
	s.fakeInstanceMgr.On("Update", s.ctx, mock.Anything).Return(errors.New("no properties provided to update"))
	err := s.controller.UpdateInstance(s.ctx, nil)
	s.Error(err)

	s.fakeInstanceMgr.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = s.controller.UpdateInstance(s.ctx, &providerModel.Instance{ID: 1}, "enabled")
	s.NoError(err)
}

func (s *preheatSuite) TestGetInstance() {
	inst, err := s.controller.GetInstance(s.ctx, 1)
	s.NoError(err)
	s.NotNil(inst)
}

func (s *preheatSuite) TestCountPolicy() {
	s.fakePolicyMgr.On("Count", s.ctx, mock.Anything).Return(int64(1), nil)
	id, err := s.controller.CountPolicy(s.ctx, nil)
	s.NoError(err)
	s.Equal(int64(1), id)
}

func (s *preheatSuite) TestCreatePolicy() {
	policy := &policy.Schema{Name: "test"}
	s.fakePolicyMgr.On("Create", s.ctx, policy).Return(int64(1), nil)
	id, err := s.controller.CreatePolicy(s.ctx, policy)
	s.NoError(err)
	s.Equal(int64(1), id)
	s.False(policy.CreatedAt.IsZero())
	s.False(policy.UpdatedTime.IsZero())
}

func (s *preheatSuite) TestGetPolicy() {
	s.fakePolicyMgr.On("Get", s.ctx, int64(1)).Return(&policy.Schema{Name: "test"}, nil)
	p, err := s.controller.GetPolicy(s.ctx, 1)
	s.NoError(err)
	s.Equal("test", p.Name)
}

func (s *preheatSuite) TestGetPolicyByName() {
	s.fakePolicyMgr.On("GetByName", s.ctx, int64(1), "test").Return(&policy.Schema{Name: "test"}, nil)
	p, err := s.controller.GetPolicyByName(s.ctx, 1, "test")
	s.NoError(err)
	s.Equal("test", p.Name)
}

func (s *preheatSuite) TestUpdatePolicy() {
	policy := &policy.Schema{Name: "test"}
	s.fakePolicyMgr.On("Update", s.ctx, policy, mock.Anything).Return(nil)
	err := s.controller.UpdatePolicy(s.ctx, policy, "")
	s.NoError(err)
	s.False(policy.UpdatedTime.IsZero())
}

func (s *preheatSuite) TestDeletePolicy() {
	s.fakePolicyMgr.On("Delete", s.ctx, int64(1)).Return(nil)
	err := s.controller.DeletePolicy(s.ctx, 1)
	s.NoError(err)
}

func (s *preheatSuite) TestListPolicies() {
	s.fakePolicyMgr.On("ListPolicies", s.ctx, mock.Anything).Return([]*policy.Schema{}, nil)
	p, err := s.controller.ListPolicies(s.ctx, nil)
	s.NoError(err)
	s.NotNil(p)
}

func (s *preheatSuite) TestListPoliciesByProject() {
	s.fakePolicyMgr.On("ListPoliciesByProject", s.ctx, int64(1), mock.Anything).Return([]*policy.Schema{}, nil)
	p, err := s.controller.ListPoliciesByProject(s.ctx, 1, nil)
	s.NoError(err)
	s.NotNil(p)
}

func (s *preheatSuite) TestCheckHealth() {
	// if instance is nil
	var instance *providerModel.Instance
	err := s.controller.CheckHealth(s.ctx, instance)
	s.Error(err)

	// unknown vendor
	instance = &providerModel.Instance{
		ID:       1,
		Name:     "test-instance",
		Vendor:   "unknown",
		Endpoint: "http://127.0.0.1",
		AuthMode: auth.AuthModeNone,
		Enabled:  true,
		Default:  true,
		Insecure: true,
		Status:   "Unknown",
	}
	err = s.controller.CheckHealth(s.ctx, instance)
	s.Error(err)

	// not health
	// health
	instance = &providerModel.Instance{
		ID:       1,
		Name:     "test-instance",
		Vendor:   provider.DriverDragonfly,
		Endpoint: "http://127.0.0.1",
		AuthMode: auth.AuthModeNone,
		Enabled:  true,
		Default:  true,
		Insecure: true,
		Status:   "Unknown",
	}
	err = s.controller.CheckHealth(s.ctx, instance)
	s.Error(err)

	// health
	instance = &providerModel.Instance{
		ID:       1,
		Name:     "test-instance",
		Vendor:   provider.DriverDragonfly,
		Endpoint: s.mockInstanceServer.URL,
		AuthMode: auth.AuthModeNone,
		Enabled:  true,
		Default:  true,
		Insecure: true,
		Status:   "Unknown",
	}
	err = s.controller.CheckHealth(s.ctx, instance)
	s.NoError(err)
}
