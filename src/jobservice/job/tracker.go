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
	"encoding/json"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/list"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/gomodule/redigo/redis"
)

const (
	// Check in data placeholder for saving data space
	redundantCheckInData = "[REDUNDANT]"
)

// Tracker is designed to track the life cycle of the job described by the stats
// The status change is linear and then has strict preorder and successor
// Check should be enforced before switching
//
// Pending is default status when creating job, so no need to switch
type Tracker interface {
	// Save the job stats which tracked by this tracker to the backend
	//
	// Return:
	//   none nil error returned if any issues happened
	Save() error

	// Load the job stats which tracked by this tracker with the backend data
	//
	// Return:
	//   none nil error returned if any issues happened
	Load() error

	// Get the job stats which tracked by this tracker
	//
	// Returns:
	//  *models.Info : job stats data
	Job() *Stats

	// Update the properties of the job stats
	//
	// fieldAndValues ...interface{} : One or more properties being updated
	//
	// Returns:
	//  error if update failed
	Update(fieldAndValues ...interface{}) error

	// NumericID returns the numeric ID of periodic job.
	// Please pay attention, this only for periodic job.
	NumericID() (int64, error)

	// Mark the periodic job execution to done by update the score
	// of the relation between its periodic policy and execution to -1.
	PeriodicExecutionDone() error

	// Check in message
	CheckIn(message string) error

	// Update status with retry enabled
	UpdateStatusWithRetry(targetStatus Status) error

	// The current status of job
	Status() (Status, error)

	// Switch status to running
	Run() error

	// Switch status to stopped
	Stop() error

	// Switch the status to error
	Fail() error

	// Switch the status to success
	Succeed() error

	// Reset the status to `pending`
	Reset() error

	// Fire status hook to report the current status
	FireHook() error
}

// basicTracker implements Tracker interface based on redis
type basicTracker struct {
	namespace string
	context   context.Context
	pool      *redis.Pool
	jobID     string
	jobStats  *Stats
	callback  HookCallback
	retryList *list.SyncList
}

// NewBasicTrackerWithID builds a tracker with the provided job ID
func NewBasicTrackerWithID(
	ctx context.Context,
	jobID string,
	ns string,
	pool *redis.Pool,
	callback HookCallback,
	retryList *list.SyncList,
) Tracker {
	return &basicTracker{
		namespace: ns,
		context:   ctx,
		pool:      pool,
		jobID:     jobID,
		callback:  callback,
		retryList: retryList,
	}
}

// NewBasicTrackerWithStats builds a tracker with the provided job stats
func NewBasicTrackerWithStats(
	ctx context.Context,
	stats *Stats,
	ns string,
	pool *redis.Pool,
	callback HookCallback,
	retryList *list.SyncList,
) Tracker {
	return &basicTracker{
		namespace: ns,
		context:   ctx,
		pool:      pool,
		jobStats:  stats,
		jobID:     stats.Info.JobID,
		callback:  callback,
		retryList: retryList,
	}
}

// Refresh the job stats which tracked by this tracker
func (bt *basicTracker) Load() error {
	return bt.retrieve()
}

// Job returns the job stats which tracked by this tracker
func (bt *basicTracker) Job() *Stats {
	return bt.jobStats
}

// Update the properties of the job stats
func (bt *basicTracker) Update(fieldAndValues ...interface{}) error {
	if len(fieldAndValues) == 0 {
		return errors.New("no properties specified to update")
	}

	conn := bt.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	key := rds.KeyJobStats(bt.namespace, bt.jobID)
	args := []interface{}{"update_time", time.Now().Unix()} // update timestamp
	args = append(args, fieldAndValues...)

	return rds.HmSet(conn, key, args...)
}

// Status returns the current status of job tracked by this tracker
func (bt *basicTracker) Status() (Status, error) {
	// Retrieve the latest status again in case get the outdated one.
	conn := bt.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	rootKey := rds.KeyJobStats(bt.namespace, bt.jobID)
	return getStatus(conn, rootKey)
}

// NumericID returns the numeric ID of the periodic job
func (bt *basicTracker) NumericID() (int64, error) {
	if bt.jobStats.Info.NumericPID > 0 {
		return bt.jobStats.Info.NumericPID, nil
	}

	return -1, errors.Errorf("numeric ID not found for job: %s", bt.jobID)
}

// PeriodicExecutionDone mark the execution done
func (bt *basicTracker) PeriodicExecutionDone() error {
	if utils.IsEmptyStr(bt.jobStats.Info.UpstreamJobID) {
		return errors.Errorf("%s is not periodic job execution", bt.jobID)
	}

	key := rds.KeyUpstreamJobAndExecutions(bt.namespace, bt.jobStats.Info.UpstreamJobID)

	conn := bt.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	args := []interface{}{key, "XX", -1, bt.jobID}
	_, err := conn.Do("ZADD", args...)

	return err
}

// Check in message
func (bt *basicTracker) CheckIn(message string) error {
	if utils.IsEmptyStr(message) {
		return errors.New("check in error: empty message")
	}

	now := time.Now().Unix()
	current := Status(bt.jobStats.Info.Status)

	bt.refresh(current, message)
	err := bt.fireHookEvent(current, message)
	err = bt.Update(
		// skip checkin data here
		"check_in_at", now,
		"update_time", now,
	)

	return err
}

// Run job
// Either one is failed, the final return will be marked as failed.
func (bt *basicTracker) Run() error {
	if err := bt.setStatus(RunningStatus); err != nil {
		return errors.Wrap(err, "run")
	}

	return nil
}

// Stop job
// Stop is final status, if failed to do, retry should be enforced.
// Either one is failed, the final return will be marked as failed.
func (bt *basicTracker) Stop() error {
	if err := bt.setStatus(StoppedStatus); err != nil {
		return errors.Wrap(err, "stop")
	}

	return nil
}

// Fail job
// Fail is final status, if failed to do, retry should be enforced.
// Either one is failed, the final return will be marked as failed.
func (bt *basicTracker) Fail() error {
	if err := bt.setStatus(ErrorStatus); err != nil {
		return errors.Wrap(err, "fail")
	}

	return nil
}

// Succeed job
// Succeed is final status, if failed to do, retry should be enforced.
// Either one is failed, the final return will be marked as failed.
func (bt *basicTracker) Succeed() error {
	if err := bt.setStatus(SuccessStatus); err != nil {
		return errors.Wrap(err, "succeed")
	}

	return nil
}

// Save the stats of job tracked by this tracker
func (bt *basicTracker) Save() (err error) {
	if bt.jobStats == nil {
		return errors.New("nil job stats to save")
	}

	conn := bt.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	// Alliance
	stats := bt.jobStats

	key := rds.KeyJobStats(bt.namespace, stats.Info.JobID)
	args := make([]interface{}, 0)
	args = append(args, key)
	args = append(args,
		"id", stats.Info.JobID,
		"name", stats.Info.JobName,
		"kind", stats.Info.JobKind,
		"unique", stats.Info.IsUnique,
		"status", stats.Info.Status,
		"ref_link", stats.Info.RefLink,
		"enqueue_time", stats.Info.EnqueueTime,
		"run_at", stats.Info.RunAt,
		"cron_spec", stats.Info.CronSpec,
		"web_hook_url", stats.Info.WebHookURL,
		"numeric_policy_id", stats.Info.NumericPID,
	)
	if stats.Info.CheckInAt > 0 && !utils.IsEmptyStr(stats.Info.CheckIn) {
		args = append(args,
			"check_in", redundantCheckInData, // use data placeholder for saving space
			"check_in_at", stats.Info.CheckInAt,
		)
	}
	if stats.Info.DieAt > 0 {
		args = append(args, "die_at", stats.Info.DieAt)
	}

	if !utils.IsEmptyStr(stats.Info.UpstreamJobID) {
		args = append(args, "upstream_job_id", stats.Info.UpstreamJobID)
	}

	if len(stats.Info.Parameters) > 0 {
		if bytes, err := json.Marshal(&stats.Info.Parameters); err == nil {
			args = append(args, "parameters", string(bytes))
		}
	}
	// Set update timestamp
	args = append(args, "update_time", time.Now().Unix())
	// Set the first revision
	args = append(args, "revision", time.Now().Unix())

	// ACK data is saved/updated not via tracker, so ignore the ACK saving

	// Do it in a transaction
	err = conn.Send("MULTI")
	err = conn.Send("HMSET", args...)
	// Set inprogress track lock
	err = conn.Send("HSET", rds.KeyJobTrackInProgress(bt.namespace), stats.Info.JobID, 2)

	// Link with its upstream job if upstream job ID exists for future querying
	if !utils.IsEmptyStr(stats.Info.UpstreamJobID) {
		k := rds.KeyUpstreamJobAndExecutions(bt.namespace, stats.Info.UpstreamJobID)
		zargs := []interface{}{k, "NX", stats.Info.RunAt, stats.Info.JobID}
		err = conn.Send("ZADD", zargs...)
	}

	// Check command send error only once here before executing
	if err != nil {
		return
	}

	_, err = conn.Do("EXEC")

	return
}

// UpdateStatusWithRetry updates the status with retry enabled.
// If update status failed, then retry if permitted.
// Try best to do
func (bt *basicTracker) UpdateStatusWithRetry(targetStatus Status) error {
	err := bt.compareAndSet(targetStatus)
	if err != nil {
		// Status mismatching error will be directly ignored as the status has already been outdated
		if !errs.IsStatusMismatchError(err) {
			// Push to the retrying daemon
			bt.retryList.Push(SimpleStatusChange{
				JobID:        bt.jobID,
				TargetStatus: targetStatus.String(),
				Revision:     bt.jobStats.Info.Revision,
			})
		}
	}

	return err
}

// Reset the job status to `pending` and update the revision.
// Usually for the retry jobs
func (bt *basicTracker) Reset() error {
	conn := bt.pool.Get()
	defer func() {
		closeConn(conn)
	}()

	now := time.Now().Unix()
	if _, err := rds.StatusResetScript.Do(
		conn,
		rds.KeyJobStats(bt.namespace, bt.jobStats.Info.JobID),
		rds.KeyJobTrackInProgress(bt.namespace),
		bt.jobStats.Info.JobID,
		PendingStatus.String(),
		now,
	); err != nil {
		return errors.Wrap(err, "reset")
	}

	// Sync current tracker
	bt.jobStats.Info.Status = PendingStatus.String()
	bt.jobStats.Info.Revision = now
	bt.jobStats.Info.UpdateTime = now
	bt.jobStats.Info.CheckIn = ""
	bt.jobStats.Info.CheckInAt = 0

	return nil
}

// FireHook fires status hook event to report current status
func (bt *basicTracker) FireHook() error {
	return bt.fireHookEvent(
		Status(bt.jobStats.Info.Status),
		bt.jobStats.Info.CheckIn,
	)
}

// setStatus sets the job status to the target status and fire status change hook
func (bt *basicTracker) setStatus(status Status) error {
	err := bt.UpdateStatusWithRetry(status)
	if !errs.IsStatusMismatchError(err) {
		bt.refresh(status)
		if er := bt.fireHookEvent(status); er != nil {
			// Add more error context
			if err != nil {
				return errors.Wrap(er, err.Error())
			}

			return er
		}
	}

	return err
}

// Refresh the job stats in mem
func (bt *basicTracker) refresh(targetStatus Status, checkIn ...string) {
	now := time.Now().Unix()

	bt.jobStats.Info.Status = targetStatus.String()
	if len(checkIn) > 0 {
		bt.jobStats.Info.CheckIn = checkIn[0]
		bt.jobStats.Info.CheckInAt = now
	}
	bt.jobStats.Info.UpdateTime = now
}

// FireHookEvent fires the hook event
func (bt *basicTracker) fireHookEvent(status Status, checkIn ...string) error {
	// Check if hook URL is registered
	if utils.IsEmptyStr(bt.jobStats.Info.WebHookURL) {
		// Do nothing
		return nil
	}

	change := &StatusChange{
		JobID:    bt.jobID,
		Status:   status.String(),
		Metadata: bt.jobStats.Info,
	}

	if len(checkIn) > 0 {
		change.CheckIn = checkIn[0]
	}

	// If callback is registered, then trigger now
	if bt.callback != nil {
		return bt.callback(bt.jobStats.Info.WebHookURL, change)
	}

	return nil
}

func (bt *basicTracker) compareAndSet(targetStatus Status) error {
	conn := bt.pool.Get()
	defer func() {
		closeConn(conn)
	}()

	rootKey := rds.KeyJobStats(bt.namespace, bt.jobID)
	trackKey := rds.KeyJobTrackInProgress(bt.namespace)
	reply, err := redis.String(rds.SetStatusScript.Do(
		conn,
		rootKey,
		trackKey,
		targetStatus.String(),
		bt.jobStats.Info.Revision,
		time.Now().Unix(),
		bt.jobID,
	))
	if err != nil {
		return errors.Wrap(err, "compare and set status error")
	}

	if reply != "ok" {
		return errs.StatusMismatchError(reply, targetStatus.String())
	}

	return nil
}

// retrieve the stats of job tracked by this tracker from the backend data
func (bt *basicTracker) retrieve() error {
	conn := bt.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	key := rds.KeyJobStats(bt.namespace, bt.jobID)
	vals, err := redis.Strings(conn.Do("HGETALL", key))
	if err != nil {
		return err
	}

	if vals == nil || len(vals) == 0 {
		return errs.NoObjectFoundError(bt.jobID)
	}

	res := &Stats{
		Info: &StatsInfo{},
	}

	for i, l := 0, len(vals); i < l; i = i + 2 {
		prop := vals[i]
		value := vals[i+1]
		switch prop {
		case "id":
			res.Info.JobID = value
			break
		case "name":
			res.Info.JobName = value
			break
		case "kind":
			res.Info.JobKind = value
		case "unique":
			v, err := strconv.ParseBool(value)
			if err != nil {
				v = false
			}
			res.Info.IsUnique = v
		case "status":
			res.Info.Status = value
			break
		case "ref_link":
			res.Info.RefLink = value
			break
		case "enqueue_time":
			res.Info.EnqueueTime = parseInt64(value)
			break
		case "update_time":
			res.Info.UpdateTime = parseInt64(value)
			break
		case "run_at":
			res.Info.RunAt = parseInt64(value)
			break
		case "check_in_at":
			res.Info.CheckInAt = parseInt64(value)
			break
		case "check_in":
			res.Info.CheckIn = "" // never read checkin placeholder data
			break
		case "cron_spec":
			res.Info.CronSpec = value
			break
		case "web_hook_url":
			res.Info.WebHookURL = value
			break
		case "die_at":
			res.Info.DieAt = parseInt64(value)
		case "upstream_job_id":
			res.Info.UpstreamJobID = value
			break
		case "numeric_policy_id":
			res.Info.NumericPID = parseInt64(value)
			break
		case "parameters":
			params := make(Parameters)
			if err := json.Unmarshal([]byte(value), &params); err == nil {
				res.Info.Parameters = params
			} else {
				logger.Error(errors.Wrap(err, "retrieve: tracker"))
			}
			break
		case "revision":
			res.Info.Revision = parseInt64(value)
			break
		case "ack":
			ack := &ACK{}
			if err := json.Unmarshal([]byte(value), ack); err == nil {
				res.Info.HookAck = ack
			} else {
				logger.Error(errors.Wrap(err, "retrieve: tracker"))
			}
		default:
			break
		}
	}

	bt.jobStats = res

	return nil
}

func getStatus(conn redis.Conn, key string) (Status, error) {
	values, err := rds.HmGet(conn, key, "status")
	if err != nil {
		return "", errors.Wrap(err, "get status error")
	}

	if len(values) == 1 {
		st := Status(values[0].([]byte))
		if st.Validate() == nil {
			return st, nil
		}
	}

	return "", errors.New("malformed status data returned")
}

func closeConn(conn redis.Conn) {
	if conn != nil {
		if err := conn.Close(); err != nil {
			logger.Errorf("Close redis connection failed with error: %s", err)
		}
	}
}

func parseInt64(v string) int64 {
	intV, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		logger.Errorf("Parse int64 error: %s", err)
		return 0
	}

	return intV
}
