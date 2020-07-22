package job

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/jobservice/job"
)

var (
	// GlobalClient is an instance of the default client that can be used globally
	GlobalClient             Client = NewDefaultClient(config.InternalJobServiceURL(), config.CoreSecret())
	statusBehindErrorPattern        = "mismatch job status for stopping job: .*, job status (.*) is behind Running"
	statusBehindErrorReg            = regexp.MustCompile(statusBehindErrorPattern)

	// ErrJobNotFound indicates the job not found
	ErrJobNotFound = errors.New("job not found")
)

// Client wraps interface to access jobservice.
type Client interface {
	SubmitJob(*models.JobData) (string, error)
	GetJobLog(uuid string) ([]byte, error)
	PostAction(uuid, action string) error
	GetExecutions(uuid string) ([]job.Stats, error)
	// TODO Redirect joblog when we see there's memory issue.
}

// StatusBehindError represents the error got when trying to stop a success/failed job
type StatusBehindError struct {
	status string
}

// Error returns the detail message about the error
func (s *StatusBehindError) Error() string {
	return "status behind error"
}

// Status returns the current status of the job
func (s *StatusBehindError) Status() string {
	return s.status
}

// DefaultClient is the default implementation of Client interface
type DefaultClient struct {
	endpoint string
	client   *commonhttp.Client
}

// NewDefaultClient creates a default client based on endpoint and secret.
func NewDefaultClient(endpoint, secret string) *DefaultClient {
	var c *commonhttp.Client
	httpCli := &http.Client{
		Transport: commonhttp.GetHTTPTransport(commonhttp.SecureTransport),
	}
	if len(secret) > 0 {
		c = commonhttp.NewClient(httpCli, auth.NewSecretAuthorizer(secret))
	} else {
		c = commonhttp.NewClient(httpCli)
	}
	e := strings.TrimRight(endpoint, "/")
	return &DefaultClient{
		endpoint: e,
		client:   c,
	}
}

// NewReplicationClient used to create a client for replication
func NewReplicationClient(endpoint, secret string) *DefaultClient {
	var c *commonhttp.Client

	if len(secret) > 0 {
		c = commonhttp.NewClient(&http.Client{
			Transport: commonhttp.GetHTTPTransport(commonhttp.SecureTransport),
		},
			auth.NewSecretAuthorizer(secret))
	} else {
		c = commonhttp.NewClient(&http.Client{
			Transport: commonhttp.GetHTTPTransport(commonhttp.SecureTransport),
		})
	}

	e := strings.TrimRight(endpoint, "/")
	return &DefaultClient{
		endpoint: e,
		client:   c,
	}
}

// SubmitJob call jobservice API to submit a job and returns the job's UUID.
func (d *DefaultClient) SubmitJob(jd *models.JobData) (string, error) {
	url := d.endpoint + "/api/v1/jobs"
	jq := models.JobRequest{
		Job: jd,
	}
	b, err := json.Marshal(jq)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusAccepted {
		return "", &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
	}
	stats := &models.JobStats{}
	if err := json.Unmarshal(data, stats); err != nil {
		return "", err
	}
	return stats.Stats.JobID, nil
}

// GetJobLog call jobserivce API to get the log of a job.  It only accepts the UUID of the job
func (d *DefaultClient) GetJobLog(uuid string) ([]byte, error) {
	url := d.endpoint + "/api/v1/jobs/" + uuid + "/log"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
	}
	return data, nil
}

// GetExecutions ...
func (d *DefaultClient) GetExecutions(periodicJobID string) ([]job.Stats, error) {
	url := fmt.Sprintf("%s/api/v1/jobs/%s/executions?page_number=1&page_size=100", d.endpoint, periodicJobID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
	}
	var exes []job.Stats
	err = json.Unmarshal(data, &exes)
	if err != nil {
		return nil, err
	}
	return exes, nil
}

// PostAction call jobservice's API to operate action for job specified by uuid
func (d *DefaultClient) PostAction(uuid, action string) error {
	url := d.endpoint + "/api/v1/jobs/" + uuid
	req := struct {
		Action string `json:"action"`
	}{
		Action: action,
	}
	if err := d.client.Post(url, req); err != nil {
		status, flag := isStatusBehindError(err)
		if flag {
			return &StatusBehindError{
				status: status,
			}
		}
		if e, ok := err.(*commonhttp.Error); ok && e.Code == http.StatusNotFound {
			return ErrJobNotFound
		}
		return err
	}
	return nil
}

func isStatusBehindError(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	strs := statusBehindErrorReg.FindStringSubmatch(err.Error())
	if len(strs) != 2 {
		return "", false
	}
	return strs[1], true
}
