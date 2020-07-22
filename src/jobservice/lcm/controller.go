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

package lcm

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/list"
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/gomodule/redigo/redis"
)

const (
	// Waiting a short while if any errors occurred
	shortLoopInterval = 5 * time.Second
	// Waiting for long while if no retrying elements found
	longLoopInterval = 5 * time.Minute
	// loopInterval is the interval for the loop of restoring dead status
	loopInterval = 2 * time.Minute
	// shortInterval is initial interval and be as based to give random buffer to loopInterval
	shortInterval = 10
)

// Controller is designed to control the life cycle of the job
type Controller interface {
	// Run daemon process if needed
	Serve() error

	// New tracker from the new provided stats
	New(stats *job.Stats) (job.Tracker, error)

	// Track the life cycle of the specified existing job
	Track(jobID string) (job.Tracker, error)
}

// basicController is default implementation of Controller based on redis
type basicController struct {
	context   context.Context
	namespace string
	pool      *redis.Pool
	callback  job.HookCallback
	wg        *sync.WaitGroup
	retryList *list.SyncList
}

// NewController is the constructor of basic controller
func NewController(ctx *env.Context, ns string, pool *redis.Pool, callback job.HookCallback) Controller {
	return &basicController{
		context:   ctx.SystemContext,
		namespace: ns,
		pool:      pool,
		callback:  callback,
		wg:        ctx.WG,
		retryList: list.New(),
	}
}

// Serve ...
func (bc *basicController) Serve() error {
	bc.wg.Add(1)
	go bc.loopForRestoreDeadStatus()

	logger.Info("Status restoring loop is started")

	return nil
}

// New tracker
func (bc *basicController) New(stats *job.Stats) (job.Tracker, error) {
	if stats == nil {
		return nil, errors.New("nil stats when creating job tracker")
	}

	if err := stats.Validate(); err != nil {
		return nil, errors.Errorf("error occurred when creating job tracker: %s", err)
	}

	bt := job.NewBasicTrackerWithStats(bc.context, stats, bc.namespace, bc.pool, bc.callback, bc.retryList)
	if err := bt.Save(); err != nil {
		return nil, err
	}

	return bt, nil
}

// Track and attache with the job
func (bc *basicController) Track(jobID string) (job.Tracker, error) {
	bt := job.NewBasicTrackerWithID(bc.context, jobID, bc.namespace, bc.pool, bc.callback, bc.retryList)
	if err := bt.Load(); err != nil {
		return nil, err
	}

	return bt, nil
}

// loopForRestoreDeadStatus is a loop to restore the dead states of jobs
// Obviously,this retry is a try best action.
// The retry items are not persisted and they will be gone if the job service is restart.
func (bc *basicController) loopForRestoreDeadStatus() {
	// Generate random timer duration
	rd := func() time.Duration {
		return longLoopInterval + time.Duration(rand.Int31n(shortInterval))*time.Second
	}

	defer func() {
		logger.Info("Status restoring loop is stopped")
		bc.wg.Done()
	}()

	// Initialize the timer
	tm := time.NewTimer(shortInterval * time.Second)
	defer tm.Stop()

	for {
		select {
		case <-tm.C:
			// Reset timer
			tm.Reset(rd())

			// Retry the items in the list
			bc.retryLoop()
		case <-bc.context.Done():
			return // terminated
		}
	}
}

// retryLoop iterates the retry queue and do retrying
func (bc *basicController) retryLoop() {
	// Get connection
	conn := bc.pool.Get()
	defer func() {
		// Return redis connection
		if err := conn.Close(); err != nil {
			logger.Errorf("Failed to close redis connection: %v : %s", err, "retry loop: lcm")
		}
	}()

	// Check the list
	bc.retryList.Iterate(func(ele interface{}) bool {
		if change, ok := ele.(job.SimpleStatusChange); ok {
			err := retry(conn, bc.namespace, change)
			if err != nil {
				// Log the error
				logger.Errorf("Failed to retry the status update action: %v : %s", err, "retry loop: lcm")
			}

			if err == nil || errs.IsStatusMismatchError(err) {
				return true
			}
		}

		return false
	})
}

// retry status update action
func retry(conn redis.Conn, ns string, change job.SimpleStatusChange) error {
	// Debug
	logger.Debugf("Retry the status update action: %v", change)

	rootKey := rds.KeyJobStats(ns, change.JobID)
	trackKey := rds.KeyJobTrackInProgress(ns)

	reply, err := redis.String(rds.SetStatusScript.Do(
		conn,
		rootKey,
		trackKey,
		change.TargetStatus,
		change.Revision,
		time.Now().Unix(),
		change.JobID,
	))
	if err != nil {
		return errors.Wrap(err, "retry")
	}

	if reply != "ok" {
		return errs.StatusMismatchError(reply, change.TargetStatus)
	}

	return nil
}
