package retention

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/testing/pkg/repository"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type ControllerTestSuite struct {
	suite.Suite

	oldClient dep.Client
}

// SetupSuite ...
func (s *ControllerTestSuite) SetupSuite() {

}

// TestController ...
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (s *ControllerTestSuite) TestPolicy() {
	projectMgr := &fakeProjectManager{}
	repositoryMgr := &repository.FakeManager{}
	retentionScheduler := &fakeRetentionScheduler{}
	retentionLauncher := &fakeLauncher{}
	retentionMgr := NewManager()
	c := NewAPIController(retentionMgr, projectMgr, repositoryMgr, retentionScheduler, retentionLauncher)

	p1 := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
				Parameters: rule.Parameters{
					"latestPushedK": 10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "doublestar",
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       "doublestar",
							Decoration: "matches",
							Pattern:    ".+",
						},
					},
				},
			},
			{
				ID:       2,
				Priority: 1,
				Template: "latestPushedK",
				Disabled: true,
				Parameters: rule.Parameters{
					"latestPushedK": 3,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "doublestar",
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       "doublestar",
							Decoration: "matches",
							Pattern:    ".+",
						},
					},
				},
			},
		},
		Trigger: &policy.Trigger{
			Kind: "Schedule",
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	id, err := c.CreateRetention(p1)
	s.Require().Nil(err)
	s.Require().True(id > 0)

	p1, err = c.GetRetention(id)
	s.Require().Nil(err)
	s.Require().EqualValues("project", p1.Scope.Level)
	s.Require().True(p1.ID > 0)

	p1.Scope.Level = "test"
	err = c.UpdateRetention(p1)
	s.Require().Nil(err)
	p1, err = c.GetRetention(id)
	s.Require().Nil(err)
	s.Require().EqualValues("test", p1.Scope.Level)

	err = c.DeleteRetention(id)
	s.Require().Nil(err)

	p1, err = c.GetRetention(id)
	s.Require().NotNil(err)
	s.Require().True(strings.Contains(err.Error(), "no such Retention policy"))
	s.Require().Nil(p1)
}

func (s *ControllerTestSuite) TestExecution() {
	projectMgr := &fakeProjectManager{}
	repositoryMgr := &repository.FakeManager{}
	retentionScheduler := &fakeRetentionScheduler{}
	retentionLauncher := &fakeLauncher{}
	retentionMgr := NewManager()
	m := NewAPIController(retentionMgr, projectMgr, repositoryMgr, retentionScheduler, retentionLauncher)

	p1 := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
				Parameters: rule.Parameters{
					"latestPushedK": 10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "doublestar",
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       "doublestar",
							Decoration: "matches",
							Pattern:    ".+",
						},
					},
				},
			},
		},
		Trigger: &policy.Trigger{
			Kind: "Schedule",
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	policyID, err := m.CreateRetention(p1)
	s.Require().Nil(err)
	s.Require().True(policyID > 0)

	id, err := m.TriggerRetentionExec(policyID, ExecutionTriggerManual, false)
	s.Require().Nil(err)
	s.Require().True(id > 0)

	e1, err := m.GetRetentionExec(id)
	s.Require().Nil(err)
	s.Require().NotNil(e1)
	s.Require().EqualValues(id, e1.ID)

	err = m.OperateRetentionExec(id, "stop")
	s.Require().Nil(err)

	es, err := m.ListRetentionExecs(policyID, nil)
	s.Require().Nil(err)
	s.Require().EqualValues(1, len(es))

	ts, err := m.ListRetentionExecTasks(id, nil)
	s.Require().Nil(err)
	s.Require().EqualValues(0, len(ts))

}

type fakeRetentionScheduler struct {
}

func (f *fakeRetentionScheduler) Schedule(ctx context.Context, cron string, callbackFuncName string, params interface{}) (int64, error) {
	return 111, nil
}

func (f *fakeRetentionScheduler) UnSchedule(ctx context.Context, id int64) error {
	return nil
}

func (f *fakeRetentionScheduler) GetSchedule(ctx context.Context, id int64) (*scheduler.Schedule, error) {
	return nil, nil
}

type fakeLauncher struct {
}

func (f *fakeLauncher) Stop(executionID int64) error {
	return nil
}

func (f *fakeLauncher) Launch(policy *policy.Metadata, executionID int64, isDryRun bool) (int64, error) {
	return 0, nil
}
