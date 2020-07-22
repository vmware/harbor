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

package retention

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib/selector"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/lwp"
	"github.com/olekukonko/tablewriter"
)

const (
	actionMarkRetain    = "RETAIN"
	actionMarkDeletion  = "DEL"
	actionMarkError     = "ERR"
	actionMarkImmutable = "IMMUTABLE"
)

// Job of running retention process
type Job struct{}

// MaxFails of the job
func (pj *Job) MaxFails() uint {
	return 3
}

// MaxCurrency is implementation of same method in Interface.
func (pj *Job) MaxCurrency() uint {
	return 0
}

// ShouldRetry indicates job can be retried if failed
func (pj *Job) ShouldRetry() bool {
	return false
}

// Validate the parameters
func (pj *Job) Validate(params job.Parameters) (err error) {
	if _, err = getParamRepo(params); err == nil {
		if _, err = getParamMeta(params); err == nil {
			_, err = getParamDryRun(params)
		}
	}

	return
}

// Run the job
func (pj *Job) Run(ctx job.Context, params job.Parameters) error {
	// logger for logging
	myLogger := ctx.GetLogger()

	// Parameters have been validated, ignore error checking
	repo, _ := getParamRepo(params)
	liteMeta, _ := getParamMeta(params)
	isDryRun, _ := getParamDryRun(params)

	// Log stage: start
	repoPath := fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)
	myLogger.Infof("Run retention process.\n Repository: %s \n Rule Algorithm: %s \n Dry Run: %v", repoPath, liteMeta.Algorithm, isDryRun)

	// Stop check point 1:
	if isStopped(ctx) {
		logStop(myLogger)
		return nil
	}

	// Retrieve all the candidates under the specified repository
	allCandidates, err := dep.DefaultClient.GetCandidates(repo)
	if err != nil {
		return logError(myLogger, err)
	}

	// Log stage: load candidates
	myLogger.Infof("Load %d candidates from repository %s", len(allCandidates), repoPath)

	// Build the processor
	builder := policy.NewBuilder(allCandidates)
	processor, err := builder.Build(liteMeta, isDryRun)
	if err != nil {
		return logError(myLogger, err)
	}

	// Stop check point 2:
	if isStopped(ctx) {
		logStop(myLogger)
		return nil
	}

	// Run the flow
	results, err := processor.Process(allCandidates)
	if err != nil {
		return logError(myLogger, err)
	}

	// Log stage: results with table view
	logResults(myLogger, allCandidates, results)

	// Save retain and total num in DB
	return saveRetainNum(ctx, results, allCandidates, isDryRun)
}

func saveRetainNum(ctx job.Context, results []*selector.Result, allCandidates []*selector.Candidate, isDryRun bool) error {
	var realDelete []*selector.Result
	for _, r := range results {
		if r.Error == nil {
			realDelete = append(realDelete, r)
		}
	}
	retainObj := struct {
		Total    int                `json:"total"`
		Retained int                `json:"retained"`
		DryRun   bool               `json:"dry_run"`
		Deleted  []*selector.Result `json:"deleted"`
	}{
		Total:    len(allCandidates),
		Retained: len(allCandidates) - len(realDelete),
		DryRun:   isDryRun,
		Deleted:  realDelete,
	}
	c, err := json.Marshal(retainObj)
	if err != nil {
		return err
	}
	_ = ctx.Checkin(string(c))
	return nil
}

func logResults(logger logger.Interface, all []*selector.Candidate, results []*selector.Result) {
	hash := make(map[string]error, len(results))
	for _, r := range results {
		if r.Target != nil {
			hash[r.Target.Hash()] = r.Error
		}
	}

	op := func(c *selector.Candidate) string {
		if e, exists := hash[c.Hash()]; exists {
			if e != nil {
				if _, ok := e.(*selector.ImmutableError); ok {
					return actionMarkImmutable
				}
				return actionMarkError
			}

			return actionMarkDeletion
		}

		return actionMarkRetain
	}

	var buf bytes.Buffer

	data := make([][]string, len(all))

	for _, c := range all {
		row := []string{
			c.Digest,
			strings.Join(c.Tags, ","),
			c.Kind,
			strings.Join(c.Labels, ","),
			t(c.PushedTime),
			t(c.PulledTime),
			t(c.CreationTime),
			op(c),
		}
		data = append(data, row)
	}

	table := tablewriter.NewWriter(&buf)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Digest", "Tag", "Kind", "Labels", "PushedTime", "PulledTime", "CreatedTime", "Retention"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()

	logger.Infof("\n%s", buf.String())

	// log all the concrete errors if have
	for _, r := range results {
		if r.Error != nil {
			logger.Infof("Retention error for artifact %s:%s : %s", r.Target.Kind, arn(r.Target), r.Error)
		}
	}
}

func arn(art *selector.Candidate) string {
	return fmt.Sprintf("%s/%s:%s", art.Namespace, art.Repository, art.Digest)
}

func t(tm int64) string {
	if tm <= 0 {
		return ""
	}
	return time.Unix(tm, 0).Format("2006/01/02 15:04:05")
}

func isStopped(ctx job.Context) (stopped bool) {
	cmd, ok := ctx.OPCommand()
	stopped = ok && cmd == job.StopCommand

	return
}

func logStop(logger logger.Interface) {
	logger.Info("Retention job is stopped")
}

func logError(logger logger.Interface, err error) error {
	wrappedErr := errors.Wrap(err, "retention job")
	logger.Error(wrappedErr)

	return wrappedErr
}

func getParamDryRun(params job.Parameters) (bool, error) {
	v, ok := params[ParamDryRun]
	if !ok {
		return false, errors.Errorf("missing parameter: %s", ParamDryRun)
	}

	dryRun, ok := v.(bool)
	if !ok {
		return false, errors.Errorf("invalid parameter: %s", ParamDryRun)
	}

	return dryRun, nil
}

func getParamRepo(params job.Parameters) (*selector.Repository, error) {
	v, ok := params[ParamRepo]
	if !ok {
		return nil, errors.Errorf("missing parameter: %s", ParamRepo)
	}

	repoJSON, ok := v.(string)
	if !ok {
		return nil, errors.Errorf("invalid parameter: %s", ParamRepo)
	}

	repo := &selector.Repository{}
	if err := repo.FromJSON(repoJSON); err != nil {
		return nil, errors.Wrap(err, "parse repository from JSON")
	}

	return repo, nil
}

func getParamMeta(params job.Parameters) (*lwp.Metadata, error) {
	v, ok := params[ParamMeta]
	if !ok {
		return nil, errors.Errorf("missing parameter: %s", ParamMeta)
	}

	metaJSON, ok := v.(string)
	if !ok {
		return nil, errors.Errorf("invalid parameter: %s", ParamMeta)
	}

	meta := &lwp.Metadata{}
	if err := meta.FromJSON(metaJSON); err != nil {
		return nil, errors.Wrap(err, "parse retention policy from JSON")
	}

	return meta, nil
}
