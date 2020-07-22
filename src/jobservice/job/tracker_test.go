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

package job

import (
	"context"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/list"

	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TrackerTestSuite tests functions of tracker
type TrackerTestSuite struct {
	suite.Suite

	namespace string
	pool      *redis.Pool
}

// TestTrackerTestSuite is entry of go test
func TestTrackerTestSuite(t *testing.T) {
	suite.Run(t, new(TrackerTestSuite))
}

// SetupSuite prepares test suite
func (suite *TrackerTestSuite) SetupSuite() {
	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()
}

// TearDownSuite prepares test suites
func (suite *TrackerTestSuite) TearDownSuite() {
	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	_ = tests.ClearAll(suite.namespace, conn)
}

// TestTracker tests tracker
func (suite *TrackerTestSuite) TestTracker() {
	jobID := utils.MakeIdentifier()
	mockJobStats := &Stats{
		Info: &StatsInfo{
			JobID:    jobID,
			Status:   SuccessStatus.String(),
			JobKind:  KindGeneric,
			JobName:  SampleJob,
			IsUnique: false,
		},
	}

	tracker := NewBasicTrackerWithStats(
		context.TODO(),
		mockJobStats,
		suite.namespace,
		suite.pool,
		func(hookURL string, change *StatusChange) error {
			return nil
		},
		list.New(),
	)

	err := tracker.Save()
	require.Nil(suite.T(), err, "save: nil error expected but got %s", err)

	s, err := tracker.Status()
	assert.Nil(suite.T(), err, "get status: nil error expected but got %s", err)
	assert.Equal(suite.T(), SuccessStatus, s, "get status: expected pending but got %s", s)

	j := tracker.Job()
	assert.Equal(suite.T(), jobID, j.Info.JobID, "job: expect job ID %s but got %s", jobID, j.Info.JobID)

	err = tracker.Update("web_hook_url", "http://hook.url")
	assert.Nil(suite.T(), err, "update: nil error expected but got %s", err)

	err = tracker.Load()
	assert.Nil(suite.T(), err, "load: nil error expected but got %s", err)
	assert.Equal(
		suite.T(),
		"http://hook.url",
		tracker.Job().Info.WebHookURL,
		"web hook: expect %s but got %s",
		"http://hook.url",
		tracker.Job().Info.WebHookURL,
	)

	err = tracker.Run()
	assert.Error(suite.T(), err, "run: non nil error expected but got nil")
	err = tracker.CheckIn("check in")
	assert.Nil(suite.T(), err, "check in: nil error expected but got %s", err)
	// check in is allowed to be repeated
	err = tracker.CheckIn("check in2")
	assert.Nil(suite.T(), err, "check in2: nil error expected but got %s", err)

	err = tracker.Succeed()
	assert.Nil(suite.T(), err, "succeed: nil error expected but got %s", err)
	// same status is allowed to update
	err = tracker.Succeed()
	assert.Nil(suite.T(), err, "succeed again: nil error expected but got %s", err)

	// final status can be set only once
	err = tracker.Fail()
	assert.Error(suite.T(), err, "fail: error expected but got nil")

	t := NewBasicTrackerWithID(
		context.TODO(),
		jobID,
		suite.namespace,
		suite.pool,
		func(hookURL string, change *StatusChange) error {
			return nil
		},
		list.New(),
	)
	err = t.Load()
	assert.NoError(suite.T(), err)

	var st Status
	err = t.Reset()
	assert.NoError(suite.T(), err)

	st, err = t.Status()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), PendingStatus, st)

	err = t.Stop()
	assert.NoError(suite.T(), err)

	st, err = t.Status()
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), StoppedStatus, st)
}

// TestPeriodicTracker tests tracker of periodic
func (suite *TrackerTestSuite) TestPeriodicTracker() {
	jobID := utils.MakeIdentifier()
	nID := time.Now().Unix()
	mockJobStats := &Stats{
		Info: &StatsInfo{
			JobID:      jobID,
			Status:     ScheduledStatus.String(),
			JobKind:    KindPeriodic,
			JobName:    SampleJob,
			IsUnique:   false,
			CronSpec:   "0 0 * * * *",
			NumericPID: nID,
		},
	}

	t := NewBasicTrackerWithStats(context.TODO(), mockJobStats, suite.namespace, suite.pool, nil, nil)
	err := t.Save()
	require.NoError(suite.T(), err)

	executionID := utils.MakeIdentifier()
	runAt := time.Now().Add(1 * time.Hour).Unix()
	executionStats := &Stats{
		Info: &StatsInfo{
			JobID:         executionID,
			Status:        ScheduledStatus.String(),
			JobKind:       KindScheduled,
			JobName:       SampleJob,
			IsUnique:      false,
			CronSpec:      "0 0 * * * *",
			RunAt:         runAt,
			EnqueueTime:   runAt,
			UpstreamJobID: jobID,
		},
	}

	t2 := NewBasicTrackerWithStats(context.TODO(), executionStats, suite.namespace, suite.pool, nil, nil)
	err = t2.Save()
	require.NoError(suite.T(), err)

	id, err := t.NumericID()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), nID, id)

	err = t2.PeriodicExecutionDone()
	require.NoError(suite.T(), err)
}
