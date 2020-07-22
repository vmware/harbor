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

package preheat

import (
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	pr "github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
)

const (
	// PreheatParamProvider is a parameter keeping the preheating provider instance info.
	PreheatParamProvider = "provider"
	// PreheatParamImage is a parameter keeping the preheating artifact (image) info.
	PreheatParamImage = "image"
	// checkInterval indicates the interval of loop check.
	checkInterval = 10 * time.Second
	// checkTimeout indicates the overall timeout of the loop check.
	checkTimeout = 1801 * time.Second
)

// Job preheats the given artifact(image) to the target preheat provider.
type Job struct{}

// MaxFails of preheat job. Don't need to retry.
func (j *Job) MaxFails() uint {
	return 1
}

// MaxCurrency indicates no limitation to the concurrency of preheat job.
func (j *Job) MaxCurrency() uint {
	return 0
}

// ShouldRetry indicates no need to retry preheat job as it's just for a cache purpose.
func (j *Job) ShouldRetry() bool {
	return false
}

// Validate the parameters of preheat job.
func (j *Job) Validate(params job.Parameters) error {
	_, err := parseParamProvider(params)
	if err != nil {
		return err
	}

	_, err = parseParamImage(params)

	return err
}

// Run the preheat process.
func (j *Job) Run(ctx job.Context, params job.Parameters) error {
	// Get logger
	myLogger := ctx.GetLogger()

	// Parse parameters, ignore errors as they have been validated already
	p, _ := parseParamProvider(params)
	pi, _ := parseParamImage(params)

	// Print related info to log first
	myLogger.Infof(
		"Preheating image '%s:%s' to the target preheat provider: %s %s:%s",
		pi.ImageName,
		pi.Tag,
		p.Vendor,
		p.Name,
		p.Endpoint,
	)

	// Get driver factory for the given provider
	fac, ok := pr.GetProvider(p.Vendor)
	if !ok {
		err := errors.Errorf("No driver registered for provider %s", p.Vendor)
		myLogger.Error(err)
		return preheatJobRunningError(err)
	}

	// Construct driver
	d, err := fac(p)
	if err != nil {
		myLogger.Error(err)
		return preheatJobRunningError(err)
	}

	myLogger.Infof("Get preheat provider driver: %s", p.Vendor)

	// Start the preheat process
	// First, check the health of the provider
	h, err := d.GetHealth()
	if err != nil {
		myLogger.Error(err)
		return preheatJobRunningError(err)
	}

	if h.Status != pr.DriverStatusHealthy {
		err = errors.Errorf("unhealthy target preheat provider: %s", p.Vendor)
		myLogger.Error(err)
		return preheatJobRunningError(err)
	}

	myLogger.Infof("Check health of preheat provider instance: %s", pr.DriverStatusHealthy)

	// Then send the preheat requests to the target provider.
	st, err := d.Preheat(pi)
	if err != nil {
		myLogger.Error(err)
		return preheatJobRunningError(err)
	}

	myLogger.Info("Sending preheat request is successfully done")

	// For some of the drivers, e.g: Kraken, the returned status of preheating request contains the
	// final status info. No need to loop check the status.
	switch st.Status {
	case provider.PreheatingStatusSuccess:
		myLogger.Info("Preheating is completed")
		return nil
	case provider.PreheatingStatusFail:
		err = errors.New("preheating is failed")
		myLogger.Error(err)
		return preheatJobRunningError(err)
	case provider.PreheatingStatusPending,
		provider.PreheatingStatusRunning:
	// do nothing
	default:
		// in case
		err = errors.Errorf("unknown status '%s' returned by the preheat provider %s-%s:%s", st.Status, p.Vendor, p.Name, p.Endpoint)
		myLogger.Error(err)
		return preheatJobRunningError(err)
	}

	myLogger.Info("Start to loop check the preheating status until it's ready or timeout(30m)")
	// If process is not completed, loop check the status until it's ready.
	tk := time.NewTicker(checkInterval)
	defer tk.Stop()

	tm := time.NewTimer(checkTimeout)
	defer tm.Stop()

	for {
		select {
		case <-tk.C:
			s, err := d.CheckProgress(st.TaskID)
			if err != nil {
				myLogger.Error(err)
				return preheatJobRunningError(err)
			}

			// Finished
			if s.Status == provider.PreheatingStatusSuccess {
				myLogger.Info("Preheating is completed")
				return nil
			}
		case <-tm.C:
			return preheatJobRunningError(errors.Errorf("status check timeout: %v", checkTimeout))
		}
	}
}

// preheatJobRunningError is an internal error format
func preheatJobRunningError(err error) error {
	return errors.Wrap(err, "preheat job running error")
}

// parseParamProvider parses the provider param.
func parseParamProvider(params job.Parameters) (*provider.Instance, error) {
	data, err := parseStrValue(params, PreheatParamProvider)
	if err != nil {
		return nil, err
	}

	ins := &provider.Instance{}
	if err := ins.FromJSON(data); err != nil {
		return nil, errors.Wrap(err, "parse job parameter error")
	}

	// Validate required info
	if len(ins.Vendor) == 0 {
		return nil, errors.New("missing vendor of preheat provider")
	}

	if ins.AuthMode != auth.AuthModeNone && len(ins.AuthInfo) == 0 {
		return nil, errors.Errorf("missing auth info for '%s' auth mode", ins.AuthMode)
	}

	if len(ins.Endpoint) == 0 {
		return nil, errors.Errorf("missing endpoint of preheat provider")
	}

	return ins, nil
}

// parseParamImage parses the preheating image param.
func parseParamImage(params job.Parameters) (*pr.PreheatImage, error) {
	data, err := parseStrValue(params, PreheatParamImage)
	if err != nil {
		return nil, err
	}

	img := &pr.PreheatImage{}
	if err := img.FromJSON(data); err != nil {
		return nil, errors.Wrap(err, "parse job parameter error")
	}

	if err := img.Validate(); err != nil {
		return nil, errors.Wrap(err, "parse job parameter error")
	}

	return img, nil
}

// parseStrValue parses the string data of the given parameter key from the job parameters.
func parseStrValue(params job.Parameters, key string) (string, error) {
	param, ok := params[key]
	if !ok || param == nil {
		return "", errors.Errorf("missing job parameter '%s'", key)
	}

	data, ok := param.(string)
	if !ok || len(data) == 0 {
		return "", errors.Errorf("bad job parameter '%s'", key)
	}

	return data, nil
}
