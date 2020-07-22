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

package runner

import (
	"fmt"
	"runtime"
	"time"

	"github.com/gocraft/work"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/period"
	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	maxTrackRetries = 6
)

// RedisJob is a job wrapper to wrap the job.Interface to the style which can be recognized by the redis worker.
type RedisJob struct {
	job     interface{}    // the real job implementation
	context *env.Context   // context
	ctl     lcm.Controller // life cycle controller
}

// NewRedisJob is constructor of RedisJob
func NewRedisJob(job interface{}, ctx *env.Context, ctl lcm.Controller) *RedisJob {
	return &RedisJob{
		job:     job,
		context: ctx,
		ctl:     ctl,
	}
}

// Run the job
func (rj *RedisJob) Run(j *work.Job) (err error) {
	var (
		runningJob  job.Interface
		execContext job.Context
		tracker     job.Tracker
	)

	// Track the running job now
	jID := j.ID

	// Check if the job is a periodic one as periodic job has its own ID format
	if eID, yes := isPeriodicJobExecution(j); yes {
		jID = eID
	}

	// As the job stats may not be ready when job executing sometimes (corner case),
	// the track call here may get NOT_FOUND error. For that case, let's do retry to recovery.
	for retried := 0; retried <= maxTrackRetries; retried++ {
		tracker, err = rj.ctl.Track(jID)
		if err == nil {
			break
		}

		if errs.IsObjectNotFoundError(err) {
			if retried < maxTrackRetries {
				// Still have chance to re-track the given job.
				// Hold for a while and retry
				b := backoff(retried)
				logger.Errorf("Track job %s: stats may not have been ready yet, hold for %d ms and retry again", jID, b)
				<-time.After(time.Duration(b) * time.Millisecond)
				continue
			} else {
				// Exit and never try.
				// Directly return without retry again as we have no way to restore the stats again.
				j.Fails = 10000000000 // never retry
			}
		}

		// Log error and exit
		logger.Errorf("Job '%s:%s' exit with error: failed to get job tracker: %s", j.Name, j.ID, err)

		// ELSE:
		// As tracker creation failed, there is no way to mark the job status change.
		// Also a non nil error return consumes a fail. If all retries are failed here,
		// it will cause the job to be zombie one (pending forever).
		// Those zombie ones will be reaped by the reaper later.

		return
	}

	// Defer to switch status
	defer func() {
		// Switch job status based on the returned error.
		// The err happened here should not override the job run error, just log it.
		if err != nil {
			// log error
			logger.Errorf("Job '%s:%s' exit with error: %s", j.Name, j.ID, err)

			if er := tracker.Fail(); er != nil {
				logger.Errorf("Error occurred when marking the status of job %s:%s to failure: %s", j.Name, j.ID, er)
			}

			return
		}

		// Nil error might be returned by the stopped job. Check the latest status here.
		// If refresh latest status failed, let the process to go on to void missing status updating.
		if latest, er := tracker.Status(); er != nil {
			logger.Errorf("Error occurred when getting the status of job %s:%s: %s", j.Name, j.ID, er)
		} else {
			if latest == job.StoppedStatus {
				// Logged
				logger.Infof("Job %s:%s is stopped", j.Name, j.ID)
				return
			}
		}

		// Mark job status to success.
		logger.Infof("Job '%s:%s' exit with success", j.Name, j.ID)
		if er := tracker.Succeed(); er != nil {
			logger.Errorf("Error occurred when marking the status of job %s:%s to success: %s", j.Name, j.ID, er)
		}
	}()

	// Defer to handle runtime error
	defer func() {
		if r := recover(); r != nil {
			// Log the stack
			buf := make([]byte, 1<<10)
			size := runtime.Stack(buf, false)
			err = errors.Errorf("runtime error: %s; stack: %s", r, buf[0:size])
			logger.Errorf("Run job %s:%s error: %s", j.Name, j.ID, err)
		}
	}()

	// Do operation based on the job status
	jStatus := job.Status(tracker.Job().Info.Status)
	switch jStatus {
	case job.PendingStatus, job.ScheduledStatus:
		// do nothing now
		break
	case job.StoppedStatus:
		// Probably jobs has been stopped by directly mark status to stopped.
		// Directly exit and no retry
		return nil
	case job.RunningStatus, job.ErrorStatus:
		// The failed jobs can be put into retry queue and the in progress jobs may be
		// interrupted by a sudden service crash event, all those jobs can be rescheduled.
		// Reset job info.
		if err = tracker.Reset(); err != nil {
			// Log error and return the original error if existing
			err = errors.Wrap(err, fmt.Sprintf("retrying %s job %s:%s failed", jStatus.String(), j.Name, j.ID))

			if len(j.LastErr) > 0 {
				err = errors.Wrap(err, j.LastErr)
			}

			return
		}

		logger.Infof("Retrying job %s:%s, revision: %d", j.Name, j.ID, tracker.Job().Info.Revision)
		break
	case job.SuccessStatus:
		// do nothing
		return nil
	default:
		return errors.Errorf("mismatch status for running job: expected %s/%s but got %s", job.PendingStatus, job.ScheduledStatus, jStatus.String())
	}

	// Build job context
	if execContext, err = rj.context.JobContext.Build(tracker); err != nil {
		return
	}

	// Defer to close logger stream
	defer func() {
		// Close open io stream first
		if closer, ok := execContext.GetLogger().(logger.Closer); ok {
			if er := closer.Close(); er != nil {
				logger.Errorf("Close job logger failed: %s", er)
			}
		}
	}()

	// Wrap job
	runningJob = Wrap(rj.job)
	// Set status to run
	if err = tracker.Run(); err != nil {
		return
	}
	// Run the job
	err = runningJob.Run(execContext, j.Args)
	// Add error context
	if err != nil {
		err = errors.Wrap(err, "run error")
	}

	// Handle retry
	rj.retry(runningJob, j)
	// Handle periodic job execution
	if _, yes := isPeriodicJobExecution(j); yes {
		if er := tracker.PeriodicExecutionDone(); er != nil {
			// Just log it
			logger.Error(er)
		}
	}

	return
}

func (rj *RedisJob) retry(j job.Interface, wj *work.Job) {
	if !j.ShouldRetry() {
		// Cancel retry immediately
		// Make it big enough to avoid retrying
		wj.Fails = 10000000000
		return
	}
}

func isPeriodicJobExecution(j *work.Job) (string, bool) {
	epoch, ok := j.Args[period.PeriodicExecutionMark]
	return fmt.Sprintf("%s@%s", j.ID, epoch), ok
}

func bp(b bool) *bool {
	return &b
}

func backoff(x int) int {
	// y=ax^2+bx+c
	var a, b, c = -111, 666, 500

	y := a*x*x + b*x + c
	if y < 0 {
		y = 0 - y
	}

	return y
}
