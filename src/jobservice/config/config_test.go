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
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ConfigurationTestSuite tests the configuration loading
type ConfigurationTestSuite struct {
	suite.Suite
}

// TestConfigurationTestSuite is suite entry for 'go test'
func TestConfigurationTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigurationTestSuite))
}

// TestConfigLoadingFailed ...
func (suite *ConfigurationTestSuite) TestConfigLoadingFailed() {
	cfg := &Configuration{}
	err := cfg.Load("./config.not-existing.yaml", false)
	assert.NotNil(suite.T(), err, "load config from none-existing document, expect none nil error but got nil")
}

// TestConfigLoadingSucceed ...
func (suite *ConfigurationTestSuite) TestConfigLoadingSucceed() {
	cfg := &Configuration{}
	err := cfg.Load("../config_test.yml", false)
	assert.Nil(suite.T(), err, "Load config from yaml file, expect nil error but got error '%s'", err)
}

// TestConfigLoadingWithEnv ...
func (suite *ConfigurationTestSuite) TestConfigLoadingWithEnv() {
	err := setENV()
	require.Nil(suite.T(), err, "set envs: expect nil error but got error '%s'", err)

	defer func() {
		err := unsetENV()
		require.Nil(suite.T(), err, "unset envs: expect nil error but got error '%s'", err)
	}()

	cfg := &Configuration{}
	err = cfg.Load("../config_test.yml", true)
	require.Nil(suite.T(), err, "load config from yaml file, expect nil error but got error '%s'", err)

	assert.Equal(suite.T(), "https", cfg.Protocol, "expect protocol 'https', but got '%s'", cfg.Protocol)
	assert.Equal(suite.T(), uint(8989), cfg.Port, "expect port 8989 but got '%d'", cfg.Port)
	assert.Equal(
		suite.T(),
		uint(8),
		cfg.PoolConfig.WorkerCount,
		"expect worker count 8 but go '%d'",
		cfg.PoolConfig.WorkerCount,
	)
	assert.Equal(
		suite.T(),
		"redis://:password@8.8.8.8:6379/2",
		cfg.PoolConfig.RedisPoolCfg.RedisURL,
		"expect redis URL 'localhost' but got '%s'",
		cfg.PoolConfig.RedisPoolCfg.RedisURL,
	)
	assert.Equal(
		suite.T(),
		"ut_namespace",
		cfg.PoolConfig.RedisPoolCfg.Namespace,
		"expect redis namespace 'ut_namespace' but got '%s'",
		cfg.PoolConfig.RedisPoolCfg.Namespace,
	)
	assert.Equal(suite.T(), "js_secret", GetAuthSecret(), "expect auth secret 'js_secret' but got '%s'", GetAuthSecret())
	assert.Equal(suite.T(), "core_secret", GetUIAuthSecret(), "expect auth secret 'core_secret' but got '%s'", GetUIAuthSecret())
	assert.Equal(suite.T(), "core_url", GetCoreURL(), "expect core url 'core_url' but got '%s'", GetCoreURL())
}

// TestDefaultConfig ...
func (suite *ConfigurationTestSuite) TestDefaultConfig() {
	err := DefaultConfig.Load("../config_test.yml", true)
	require.Nil(suite.T(), err, "load config from yaml file, expect nil error but got error '%s'", err)

	redisURL := DefaultConfig.PoolConfig.RedisPoolCfg.RedisURL
	assert.Equal(suite.T(), "redis://localhost:6379", redisURL, "expect redisURL '%s' but got '%s'", "redis://localhost:6379", redisURL)

	jLoggerCount := len(DefaultConfig.JobLoggerConfigs)
	assert.Equal(suite.T(), 2, jLoggerCount, "expect 2 job loggers configured but got %d", jLoggerCount)

	loggerCount := len(DefaultConfig.LoggerConfigs)
	assert.Equal(suite.T(), 1, loggerCount, "expect 1 loggers configured but got %d", loggerCount)

	// Only verify the complicated one
	theLogger := DefaultConfig.JobLoggerConfigs[1]
	assert.Equal(suite.T(), "FILE", theLogger.Name, "expect FILE logger but got %s", theLogger.Name)
	assert.Equal(suite.T(), "INFO", theLogger.Level, "expect INFO log level of FILE logger but got %s", theLogger.Level)
	assert.NotEqual(suite.T(), 0, len(theLogger.Settings), "expect extra settings but got nothing")
	assert.Equal(
		suite.T(),
		"/tmp/job_logs",
		theLogger.Settings["base_dir"],
		"expect extra setting base_dir to be '/tmp/job_logs' but got %s",
		theLogger.Settings["base_dir"],
	)
	assert.NotNil(suite.T(), theLogger.Sweeper, "expect non nil sweeper of FILE logger but got nil")
	assert.Equal(suite.T(), 5, theLogger.Sweeper.Duration, "expect sweep duration to be 5 but got %d", theLogger.Sweeper.Duration)
	assert.Equal(
		suite.T(),
		"/tmp/job_logs",
		theLogger.Sweeper.Settings["work_dir"],
		"expect work dir of sweeper of FILE logger to be '/tmp/job_logs' but got %s",
		theLogger.Sweeper.Settings["work_dir"],
	)
}

func setENV() error {
	err := os.Setenv("JOB_SERVICE_PROTOCOL", "https")
	err = os.Setenv("JOB_SERVICE_PORT", "8989")
	err = os.Setenv("JOB_SERVICE_HTTPS_CERT", "../server.crt")
	err = os.Setenv("JOB_SERVICE_HTTPS_KEY", "../server.key")
	err = os.Setenv("JOB_SERVICE_POOL_BACKEND", "redis")
	err = os.Setenv("JOB_SERVICE_POOL_WORKERS", "8")
	err = os.Setenv("JOB_SERVICE_POOL_REDIS_URL", "redis://:password@8.8.8.8:6379/2")
	err = os.Setenv("JOB_SERVICE_POOL_REDIS_NAMESPACE", "ut_namespace")
	err = os.Setenv("JOBSERVICE_SECRET", "js_secret")
	err = os.Setenv("CORE_SECRET", "core_secret")
	err = os.Setenv("CORE_URL", "core_url")

	return err
}

func unsetENV() error {
	err := os.Unsetenv("JOB_SERVICE_PROTOCOL")
	err = os.Unsetenv("JOB_SERVICE_PORT")
	err = os.Unsetenv("JOB_SERVICE_HTTPS_CERT")
	err = os.Unsetenv("JOB_SERVICE_HTTPS_KEY")
	err = os.Unsetenv("JOB_SERVICE_POOL_BACKEND")
	err = os.Unsetenv("JOB_SERVICE_POOL_WORKERS")
	err = os.Unsetenv("JOB_SERVICE_POOL_REDIS_URL")
	err = os.Unsetenv("JOB_SERVICE_POOL_REDIS_NAMESPACE")
	err = os.Unsetenv("JOBSERVICE_SECRET")
	err = os.Unsetenv("CORE_SECRET")

	return err
}
