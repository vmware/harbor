package notification

import (
	"github.com/goharbor/harbor/src/jobservice/job"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestSlackJobMaxFails(t *testing.T) {
	rep := &SlackJob{}
	// test default max fails
	assert.Equal(t, uint(10), rep.MaxFails())

	// test user defined max fails
	_ = os.Setenv(maxFails, "15")
	assert.Equal(t, uint(15), rep.MaxFails())

	// test user defined wrong max fails
	_ = os.Setenv(maxFails, "abc")
	assert.Equal(t, uint(10), rep.MaxFails())
}

func TestSlackJobShouldRetry(t *testing.T) {
	rep := &SlackJob{}
	assert.True(t, rep.ShouldRetry())
}

func TestSlackJobValidate(t *testing.T) {
	rep := &SlackJob{}
	assert.NotNil(t, rep.Validate(nil))

	jp := job.Parameters{
		"address": "https://webhook.slack.com/hsdouihhsd988",
		"payload": "slack payload",
	}
	assert.Nil(t, rep.Validate(jp))
}

func TestSlackJobRun(t *testing.T) {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}

	ctx.On("GetLogger").Return(logger)

	rep := &SlackJob{}

	// test slack request
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)

			// test request method
			assert.Equal(t, http.MethodPost, r.Method)
			// test request body
			assert.Equal(t, string(body), `{"key": "value"}`)
		}))
	defer ts.Close()
	params := map[string]interface{}{
		"skip_cert_verify": true,
		"payload":          `{"key": "value"}`,
		"address":          ts.URL,
	}
	// test correct slack response
	assert.Nil(t, rep.Run(ctx, params))

	tsWrong := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
	defer tsWrong.Close()
	paramsWrong := map[string]interface{}{
		"skip_cert_verify": true,
		"payload":          `{"key": "value"}`,
		"address":          tsWrong.URL,
	}
	// test incorrect slack response
	assert.NotNil(t, rep.Run(ctx, paramsWrong))
}
