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

package cworker

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/gocraft/work"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/period"
	"github.com/goharbor/harbor/src/jobservice/runner"
	"github.com/goharbor/harbor/src/jobservice/worker"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/gomodule/redigo/redis"
)

var (
	workerPoolDeadTime = 10 * time.Second
)

const (
	workerPoolStatusHealthy      = "Healthy"
	workerPoolStatusDead         = "Dead"
	pingRedisMaxTimes            = 10
	defaultWorkerCount      uint = 10
)

// basicWorker is the worker implementation based on gocraft/work powered by redis.
type basicWorker struct {
	namespace string
	redisPool *redis.Pool
	pool      *work.WorkerPool
	enqueuer  *work.Enqueuer
	client    *work.Client
	context   *env.Context
	scheduler period.Scheduler
	ctl       lcm.Controller
	reaper    *reaper

	// key is name of known job
	// value is the type of known job
	knownJobs *sync.Map
}

// workerContext ...
// We did not use this context to pass context info so far, just a placeholder.
type workerContext struct{}

// log the job
func (rpc *workerContext) logJob(job *work.Job, next work.NextMiddlewareFunc) error {
	jobCopy := *job
	// as the args may contain sensitive information, ignore them when logging the detail
	jobCopy.Args = nil
	jobInfo, _ := utils.SerializeJob(&jobCopy)
	logger.Infof("Job incoming: %s", jobInfo)

	return next()
}

// NewWorker is constructor of worker
func NewWorker(ctx *env.Context, namespace string, workerCount uint, redisPool *redis.Pool, ctl lcm.Controller) worker.Interface {
	wc := defaultWorkerCount
	if workerCount > 0 {
		wc = workerCount
	}

	return &basicWorker{
		namespace: namespace,
		redisPool: redisPool,
		pool:      work.NewWorkerPool(workerContext{}, wc, namespace, redisPool),
		enqueuer:  work.NewEnqueuer(namespace, redisPool),
		client:    work.NewClient(namespace, redisPool),
		scheduler: period.NewScheduler(ctx.SystemContext, namespace, redisPool, ctl),
		ctl:       ctl,
		context:   ctx,
		knownJobs: new(sync.Map),
		reaper: &reaper{
			context:   ctx.SystemContext,
			namespace: namespace,
			pool:      redisPool,
			lcmCtl:    ctl,
			jobTypes:  make([]string, 0), // Append data later (at the start step)
		},
	}
}

// Start to serve
// Unblock action
func (w *basicWorker) Start() error {
	if w.redisPool == nil {
		return errors.New("missing redis pool")
	}

	if utils.IsEmptyStr(w.namespace) {
		return errors.New("missing namespace")
	}

	if w.context == nil || w.context.SystemContext == nil {
		// report and exit
		return errors.New("missing context")
	}

	if w.ctl == nil {
		return errors.New("missing job life cycle controller")
	}

	// Test the redis connection
	if err := w.ping(); err != nil {
		return err
	}

	// Start the periodic scheduler
	w.scheduler.Start()

	// Listen to the system signal
	w.context.WG.Add(1)
	go func() {
		defer func() {
			w.context.WG.Done()
			logger.Infof("Basic worker is stopped")
		}()

		<-w.context.SystemContext.Done()
		w.pool.Stop()
	}()

	// Start the backend worker pool
	// Add middleware
	w.pool.Middleware((*workerContext).logJob)
	// Non blocking call
	w.pool.Start()
	logger.Infof("Basic worker is started")

	// Start the reaper
	w.knownJobs.Range(func(k interface{}, v interface{}) bool {
		w.reaper.jobTypes = append(w.reaper.jobTypes, k.(string))

		return true
	})
	w.reaper.start()

	return nil
}

// RegisterJobs is used to register multiple jobs to worker.
func (w *basicWorker) RegisterJobs(jobs map[string]interface{}) error {
	if jobs == nil || len(jobs) == 0 {
		// Do nothing
		return nil
	}

	for name, j := range jobs {
		if err := w.registerJob(name, j); err != nil {
			return err
		}
	}

	return nil
}

// Enqueue job
func (w *basicWorker) Enqueue(jobName string, params job.Parameters, isUnique bool, webHook string) (*job.Stats, error) {
	var (
		j   *work.Job
		err error
	)

	// As the job is declared to be unique,
	// check the uniqueness of the job,
	// Here we only need to make sure only 1 job with the same type and parameters in the queue
	// For the uniqueness of executing, it can be checked in the running stage
	if isUnique {
		if j, err = w.enqueuer.EnqueueUnique(jobName, params); err != nil {
			return nil, err
		}
	} else {
		// Enqueue job
		if j, err = w.enqueuer.Enqueue(jobName, params); err != nil {
			return nil, err
		}
	}

	// avoid backend worker bug
	if j == nil {
		return nil, fmt.Errorf("job '%s' can not be enqueued, please check the job metatdata", jobName)
	}

	return generateResult(j, job.KindGeneric, isUnique, params, webHook), nil
}

// Schedule job
func (w *basicWorker) Schedule(jobName string, params job.Parameters, runAfterSeconds uint64, isUnique bool, webHook string) (*job.Stats, error) {
	var (
		j   *work.ScheduledJob
		err error
	)

	// As the job is declared to be unique,
	// check the uniqueness of the job,
	// Here we only need to make sure only 1 job with the same type and parameters in the queue
	// For the uniqueness of executing, it can be checked in the running stage
	if isUnique {
		if j, err = w.enqueuer.EnqueueUniqueIn(jobName, int64(runAfterSeconds), params); err != nil {
			return nil, err
		}
	} else {
		// Enqueue job in
		if j, err = w.enqueuer.EnqueueIn(jobName, int64(runAfterSeconds), params); err != nil {
			return nil, err
		}
	}

	// avoid backend worker bug
	if j == nil {
		return nil, fmt.Errorf("job '%s' can not be enqueued, please check the job metatdata", jobName)
	}

	res := generateResult(j.Job, job.KindScheduled, isUnique, params, webHook)
	res.Info.RunAt = j.RunAt
	res.Info.Status = job.ScheduledStatus.String()

	return res, nil
}

// PeriodicallyEnqueue job
func (w *basicWorker) PeriodicallyEnqueue(jobName string, params job.Parameters, cronSetting string, isUnique bool, webHook string) (*job.Stats, error) {
	p := &period.Policy{
		ID:            utils.MakeIdentifier(),
		JobName:       jobName,
		CronSpec:      cronSetting,
		JobParameters: params,
		WebHookURL:    webHook,
	}

	id, err := w.scheduler.Schedule(p)
	if err != nil {
		return nil, err
	}

	res := &job.Stats{
		Info: &job.StatsInfo{
			JobID:       p.ID,
			JobName:     jobName,
			Status:      job.ScheduledStatus.String(),
			JobKind:     job.KindPeriodic,
			CronSpec:    cronSetting,
			WebHookURL:  webHook,
			NumericPID:  id,
			EnqueueTime: time.Now().Unix(),
			UpdateTime:  time.Now().Unix(),
			RefLink:     fmt.Sprintf("/api/v1/jobs/%s", p.ID),
			Parameters:  params,
		},
	}

	return res, nil
}

// Info of worker
func (w *basicWorker) Stats() (*worker.Stats, error) {
	// Get the status of worker pool via client
	hbs, err := w.client.WorkerPoolHeartbeats()
	if err != nil {
		return nil, err
	}

	// Find the heartbeat of this worker via pid
	stats := make([]*worker.StatsData, 0)
	for _, hb := range hbs {
		if hb.HeartbeatAt == 0 {
			continue // invalid ones
		}

		wPoolStatus := workerPoolStatusHealthy
		if time.Unix(hb.HeartbeatAt, 0).Add(workerPoolDeadTime).Before(time.Now()) {
			wPoolStatus = workerPoolStatusDead
		}
		stat := &worker.StatsData{
			WorkerPoolID: hb.WorkerPoolID,
			StartedAt:    hb.StartedAt,
			HeartbeatAt:  hb.HeartbeatAt,
			JobNames:     hb.JobNames,
			Concurrency:  hb.Concurrency,
			Status:       wPoolStatus,
		}
		stats = append(stats, stat)
	}

	if len(stats) == 0 {
		return nil, errors.New("failed to get stats of worker pools")
	}

	return &worker.Stats{
		Pools: stats,
	}, nil
}

// StopJob will stop the job
func (w *basicWorker) StopJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID to stop")
	}

	t, err := w.ctl.Track(jobID)
	if err != nil {
		return err
	}

	if job.RunningStatus.Compare(job.Status(t.Job().Info.Status)) < 0 {
		// Job has been in the final states
		logger.Warningf("Trying to stop a(n) %s job: ID=%s, Kind=%s", t.Job().Info.Status, jobID, t.Job().Info.JobKind)
		// Under this situation, the non-periodic job we're trying to stop has already been in the "non-running(stopped)" status.
		// As the goal of stopping the job running has achieved, we directly return nil here.
		if t.Job().Info.JobKind != job.KindPeriodic {
			return nil
		}

		// For the periodic job, its status should always be "Scheduled".
		// This case should never happen under the current model. But there might be some legacy job stats data
		// to cause such inconsistent situation.
		// Under this situation, let the periodical scheduler to handle and fix the issue.
	}

	switch t.Job().Info.JobKind {
	case job.KindGeneric:
		return t.Stop()
	case job.KindScheduled:
		// we need to delete the scheduled job in the queue if it is not running yet
		// otherwise, stop it.
		if err := w.client.DeleteScheduledJob(t.Job().Info.RunAt, jobID); err != nil {
			// Job is already running?
			logger.Errorf("scheduled job %s (run at = %d) is not found in the queue, is it running?", jobID, t.Job().Info.RunAt)
		}
		// Anyway, mark job stopped
		return t.Stop()
	case job.KindPeriodic:
		return w.scheduler.UnSchedule(jobID)
	default:
		return errors.Errorf("job kind %s is not supported", t.Job().Info.JobKind)
	}
}

// RetryJob retry the job
func (w *basicWorker) RetryJob(jobID string) error {
	return errors.New("not implemented")
}

// IsKnownJob ...
func (w *basicWorker) IsKnownJob(name string) (interface{}, bool) {
	return w.knownJobs.Load(name)
}

// ValidateJobParameters ...
func (w *basicWorker) ValidateJobParameters(jobType interface{}, params job.Parameters) error {
	if jobType == nil {
		return errors.New("nil job type")
	}

	theJ := runner.Wrap(jobType)
	return theJ.Validate(params)
}

// RegisterJob is used to register the job to the worker.
// j is the type of job
func (w *basicWorker) registerJob(name string, j interface{}) (err error) {
	if utils.IsEmptyStr(name) || j == nil {
		return errors.New("job can not be registered with empty name or nil interface")
	}

	// j must be job.Interface
	if _, ok := j.(job.Interface); !ok {
		return errors.Errorf("job must implement the job.Interface: %s", reflect.TypeOf(j).String())
	}

	// 1:1 constraint
	if jInList, ok := w.knownJobs.Load(name); ok {
		return fmt.Errorf("job name %s has been already registered with %s", name, reflect.TypeOf(jInList).String())
	}

	// Same job implementation can be only registered with one name
	w.knownJobs.Range(func(jName interface{}, jInList interface{}) bool {
		jobImpl := reflect.TypeOf(j).String()
		if reflect.TypeOf(jInList).String() == jobImpl {
			err = errors.Errorf("job %s has been already registered with name %s", jobImpl, jName)
			return false
		}

		return true
	})

	// Something happened in the range
	if err != nil {
		return
	}

	// Wrap job
	redisJob := runner.NewRedisJob(j, w.context, w.ctl)
	// Get more info from j
	theJ := runner.Wrap(j)
	// Put into the pool
	w.pool.JobWithOptions(
		name,
		work.JobOptions{
			MaxFails:       theJ.MaxFails(),
			MaxConcurrency: theJ.MaxCurrency(),
			Priority:       job.Priority().For(name),
			SkipDead:       true,
		},
		// Use generic handler to handle as we do not accept context with this way.
		func(job *work.Job) error {
			return redisJob.Run(job)
		},
	)
	// Keep the name of registered jobs as known jobs for future validation
	w.knownJobs.Store(name, j)

	logger.Infof("Register job %s with name %s", reflect.TypeOf(j).String(), name)

	return nil
}

// Ping the redis server
func (w *basicWorker) ping() error {
	conn := w.redisPool.Get()
	defer func() {
		_ = conn.Close()
	}()

	var err error
	for count := 1; count <= pingRedisMaxTimes; count++ {
		if _, err = conn.Do("ping"); err == nil {
			return nil
		}

		time.Sleep(time.Duration(count+4) * time.Second)
	}

	return fmt.Errorf("connect to redis server timeout: %s", err.Error())
}

// generate the job stats data
func generateResult(
	j *work.Job,
	jobKind string,
	isUnique bool,
	jobParameters job.Parameters,
	webHook string,
) *job.Stats {
	return &job.Stats{
		Info: &job.StatsInfo{
			JobID:       j.ID,
			JobName:     j.Name,
			JobKind:     jobKind,
			IsUnique:    isUnique,
			Status:      job.PendingStatus.String(),
			EnqueueTime: j.EnqueuedAt,
			UpdateTime:  time.Now().Unix(),
			RefLink:     fmt.Sprintf("/api/v1/jobs/%s", j.ID),
			Parameters:  jobParameters,
			WebHookURL:  webHook,
		},
	}
}
