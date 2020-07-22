package notification

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
)

// SlackJob implements the job interface, which send notification to slack by slack incoming webhooks.
type SlackJob struct {
	client *http.Client
	logger logger.Interface
}

// MaxFails returns that how many times this job can fail.
func (sj *SlackJob) MaxFails() (result uint) {
	// Default max fails count is 10, and its max retry interval is around 3h
	// Large enough to ensure most situations can notify successfully
	result = 10
	if maxFails, exist := os.LookupEnv(maxFails); exist {
		mf, err := strconv.ParseUint(maxFails, 10, 32)
		if err != nil {
			logger.Warningf("Fetch slack job maxFails error: %s", err.Error())
			return result
		}
		result = uint(mf)
	}
	return result
}

// MaxCurrency is implementation of same method in Interface.
func (sj *SlackJob) MaxCurrency() uint {
	return 1
}

// ShouldRetry ...
func (sj *SlackJob) ShouldRetry() bool {
	return true
}

// Validate implements the interface in job/Interface
func (sj *SlackJob) Validate(params job.Parameters) error {
	if params == nil {
		// Params are required
		return errors.New("missing parameter of slack job")
	}

	payload, ok := params["payload"]
	if !ok {
		return errors.Errorf("missing job parameter 'payload'")
	}
	_, ok = payload.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'payload', expecting string but got %s", reflect.TypeOf(payload).String())
	}

	address, ok := params["address"]
	if !ok {
		return errors.Errorf("missing job parameter 'address'")
	}
	_, ok = address.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'address', expecting string but got %s", reflect.TypeOf(address).String())
	}
	return nil
}

// Run implements the interface in job/Interface
func (sj *SlackJob) Run(ctx job.Context, params job.Parameters) error {
	if err := sj.init(ctx, params); err != nil {
		return err
	}

	err := sj.execute(params)
	if err != nil {
		sj.logger.Error(err)
	}

	// Wait a second for slack rate limit, refer to https://api.slack.com/docs/rate-limits
	time.Sleep(time.Second)
	return err
}

// init slack job
func (sj *SlackJob) init(ctx job.Context, params map[string]interface{}) error {
	sj.logger = ctx.GetLogger()

	// default use insecure transport
	sj.client = GetHTTPInstance(true)
	if v, ok := params["skip_cert_verify"]; ok {
		if insecure, ok := v.(bool); ok {
			if insecure {
				sj.client = GetHTTPInstance(false)
			}
		}
	}
	return nil
}

// execute slack job
func (sj *SlackJob) execute(params map[string]interface{}) error {
	payload := params["payload"].(string)
	address := params["address"].(string)

	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader([]byte(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sj.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack job(target: %s) response code is %d", address, resp.StatusCode)
	}
	return nil
}
