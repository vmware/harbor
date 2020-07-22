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

// Package config provides functions to handle the configurations of job service.
package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	yaml "gopkg.in/yaml.v2"
)

const (
	jobServiceProtocol                   = "JOB_SERVICE_PROTOCOL"
	jobServicePort                       = "JOB_SERVICE_PORT"
	jobServiceHTTPCert                   = "JOB_SERVICE_HTTPS_CERT"
	jobServiceHTTPKey                    = "JOB_SERVICE_HTTPS_KEY"
	jobServiceWorkerPoolBackend          = "JOB_SERVICE_POOL_BACKEND"
	jobServiceWorkers                    = "JOB_SERVICE_POOL_WORKERS"
	jobServiceRedisURL                   = "JOB_SERVICE_POOL_REDIS_URL"
	jobServiceRedisNamespace             = "JOB_SERVICE_POOL_REDIS_NAMESPACE"
	jobServiceRedisIdleConnTimeoutSecond = "JOB_SERVICE_POOL_REDIS_CONN_IDLE_TIMEOUT_SECOND"
	jobServiceAuthSecret                 = "JOBSERVICE_SECRET"
	coreURL                              = "CORE_URL"

	// JobServiceProtocolHTTPS points to the 'https' protocol
	JobServiceProtocolHTTPS = "https"
	// JobServiceProtocolHTTP points to the 'http' protocol
	JobServiceProtocolHTTP = "http"

	// JobServicePoolBackendRedis represents redis backend
	JobServicePoolBackendRedis = "redis"

	// secret of UI
	uiAuthSecret = "CORE_SECRET"

	// redis protocol schema
	redisSchema = "redis://"
)

// DefaultConfig is the default configuration reference
var DefaultConfig = &Configuration{}

// Configuration loads and keeps the related configuration items of job service.
type Configuration struct {
	// Protocol server listening on: https/http
	Protocol string `yaml:"protocol"`

	// Server listening port
	Port uint `yaml:"port"`

	// Additional config when using https
	HTTPSConfig *HTTPSConfig `yaml:"https_config,omitempty"`

	// Configurations of worker worker
	PoolConfig *PoolConfig `yaml:"worker_pool,omitempty"`

	// Job logger configurations
	JobLoggerConfigs []*LoggerConfig `yaml:"job_loggers,omitempty"`

	// Logger configurations
	LoggerConfigs []*LoggerConfig `yaml:"loggers,omitempty"`
}

// HTTPSConfig keeps additional configurations when using https protocol
type HTTPSConfig struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

// RedisPoolConfig keeps redis worker info.
type RedisPoolConfig struct {
	RedisURL  string `yaml:"redis_url"`
	Namespace string `yaml:"namespace"`
	// IdleTimeoutSecond closes connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeoutSecond int64 `yaml:"idle_timeout_second"`
}

// PoolConfig keeps worker worker configurations.
type PoolConfig struct {
	// Worker concurrency
	WorkerCount  uint             `yaml:"workers"`
	Backend      string           `yaml:"backend"`
	RedisPoolCfg *RedisPoolConfig `yaml:"redis_pool,omitempty"`
}

// CustomizedSettings keeps the customized settings of logger
type CustomizedSettings map[string]interface{}

// LogSweeperConfig keeps settings of log sweeper
type LogSweeperConfig struct {
	Duration int                `yaml:"duration"`
	Settings CustomizedSettings `yaml:"settings"`
}

// LoggerConfig keeps logger basic configurations.
type LoggerConfig struct {
	Name     string             `yaml:"name"`
	Level    string             `yaml:"level"`
	Settings CustomizedSettings `yaml:"settings"`
	Sweeper  *LogSweeperConfig  `yaml:"sweeper"`
}

// Load the configuration options from the specified yaml file.
// If the yaml file is specified and existing, load configurations from yaml file first;
// If detecting env variables is specified, load configurations from env variables;
// Please pay attentions, the detected env variable will override the same configuration item loading from file.
//
// yamlFilePath	string: The path config yaml file
// readEnv       bool  : Whether detect the environment variables or not
func (c *Configuration) Load(yamlFilePath string, detectEnv bool) error {
	if !utils.IsEmptyStr(yamlFilePath) {
		// Try to load from file first
		data, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return err
		}
		if err = yaml.Unmarshal(data, c); err != nil {
			return err
		}
	}

	if detectEnv {
		// Load from env variables
		c.loadEnvs()
	}

	// translate redis url if needed
	if c.PoolConfig != nil && c.PoolConfig.RedisPoolCfg != nil {
		redisAddress := c.PoolConfig.RedisPoolCfg.RedisURL
		if !utils.IsEmptyStr(redisAddress) {
			if _, err := url.Parse(redisAddress); err != nil {
				return fmt.Errorf("bad redis url for jobservice, %s", redisAddress)
			}
			if !strings.Contains(redisAddress, "://") {
				c.PoolConfig.RedisPoolCfg.RedisURL = fmt.Sprintf("%s%s", redisSchema, redisAddress)
			}
		}
	}

	// Validate settings
	return c.validate()
}

// GetAuthSecret get the auth secret from the env
func GetAuthSecret() string {
	return utils.ReadEnv(jobServiceAuthSecret)
}

// GetCoreURL get the core url from the env
func GetCoreURL() string {
	return utils.ReadEnv(coreURL)
}

// GetUIAuthSecret get the auth secret of UI side
func GetUIAuthSecret() string {
	return utils.ReadEnv(uiAuthSecret)
}

// Load env variables
func (c *Configuration) loadEnvs() {
	prot := utils.ReadEnv(jobServiceProtocol)
	if !utils.IsEmptyStr(prot) {
		c.Protocol = prot
	}

	p := utils.ReadEnv(jobServicePort)
	if !utils.IsEmptyStr(p) {
		if po, err := strconv.Atoi(p); err == nil {
			c.Port = uint(po)
		}
	}

	// Only when protocol is https
	if c.Protocol == JobServiceProtocolHTTPS {
		cert := utils.ReadEnv(jobServiceHTTPCert)
		if !utils.IsEmptyStr(cert) {
			if c.HTTPSConfig != nil {
				c.HTTPSConfig.Cert = cert
			} else {
				c.HTTPSConfig = &HTTPSConfig{
					Cert: cert,
				}
			}
		}

		certKey := utils.ReadEnv(jobServiceHTTPKey)
		if !utils.IsEmptyStr(certKey) {
			if c.HTTPSConfig != nil {
				c.HTTPSConfig.Key = certKey
			} else {
				c.HTTPSConfig = &HTTPSConfig{
					Key: certKey,
				}
			}
		}
	}

	backend := utils.ReadEnv(jobServiceWorkerPoolBackend)
	if !utils.IsEmptyStr(backend) {
		if c.PoolConfig == nil {
			c.PoolConfig = &PoolConfig{}
		}
		c.PoolConfig.Backend = backend
	}

	workers := utils.ReadEnv(jobServiceWorkers)
	if !utils.IsEmptyStr(workers) {
		if count, err := strconv.Atoi(workers); err == nil {
			if c.PoolConfig == nil {
				c.PoolConfig = &PoolConfig{}
			}
			c.PoolConfig.WorkerCount = uint(count)
		}
	}

	if c.PoolConfig != nil && c.PoolConfig.Backend == JobServicePoolBackendRedis {
		redisURL := utils.ReadEnv(jobServiceRedisURL)
		if !utils.IsEmptyStr(redisURL) {
			if c.PoolConfig.RedisPoolCfg == nil {
				c.PoolConfig.RedisPoolCfg = &RedisPoolConfig{}
			}
			c.PoolConfig.RedisPoolCfg.RedisURL = redisURL
		}

		rn := utils.ReadEnv(jobServiceRedisNamespace)
		if !utils.IsEmptyStr(rn) {
			if c.PoolConfig.RedisPoolCfg == nil {
				c.PoolConfig.RedisPoolCfg = &RedisPoolConfig{}
			}
			c.PoolConfig.RedisPoolCfg.Namespace = rn
		}

		it := utils.ReadEnv(jobServiceRedisIdleConnTimeoutSecond)
		if !utils.IsEmptyStr(it) {
			if c.PoolConfig.RedisPoolCfg == nil {
				c.PoolConfig.RedisPoolCfg = &RedisPoolConfig{}
			}
			v, err := strconv.Atoi(it)
			if err != nil {
				log.Warningf("Invalid idle timeout second: %s, will use 0 instead", it)
			} else {
				c.PoolConfig.RedisPoolCfg.IdleTimeoutSecond = int64(v)
			}
		}
	}

}

// Check if the configurations are valid settings.
func (c *Configuration) validate() error {
	if c.Protocol != JobServiceProtocolHTTPS &&
		c.Protocol != JobServiceProtocolHTTP {
		return fmt.Errorf("protocol should be %s or %s, but current setting is %s",
			JobServiceProtocolHTTP,
			JobServiceProtocolHTTPS,
			c.Protocol)
	}

	if !utils.IsValidPort(c.Port) {
		return fmt.Errorf("port number should be a none zero integer and less or equal 65535, but current is %d", c.Port)
	}

	if c.Protocol == JobServiceProtocolHTTPS {
		if c.HTTPSConfig == nil {
			return fmt.Errorf("certificate must be configured if serve with protocol %s", c.Protocol)
		}

		if utils.IsEmptyStr(c.HTTPSConfig.Cert) ||
			!utils.FileExists(c.HTTPSConfig.Cert) ||
			utils.IsEmptyStr(c.HTTPSConfig.Key) ||
			!utils.FileExists(c.HTTPSConfig.Key) {
			return fmt.Errorf("certificate for protocol %s is not correctly configured", c.Protocol)
		}
	}

	if c.PoolConfig == nil {
		return errors.New("no worker worker is configured")
	}

	if c.PoolConfig.Backend != JobServicePoolBackendRedis {
		return fmt.Errorf("worker worker backend %s does not support", c.PoolConfig.Backend)
	}

	// When backend is redis
	if c.PoolConfig.Backend == JobServicePoolBackendRedis {
		if c.PoolConfig.RedisPoolCfg == nil {
			return fmt.Errorf("redis worker must be configured when backend is set to '%s'", c.PoolConfig.Backend)
		}
		if utils.IsEmptyStr(c.PoolConfig.RedisPoolCfg.RedisURL) {
			return errors.New("URL of redis worker is empty")
		}
		if !strings.Contains(c.PoolConfig.RedisPoolCfg.RedisURL, "://") {
			return errors.New("invalid redis URL")
		}

		if _, err := url.Parse(c.PoolConfig.RedisPoolCfg.RedisURL); err != nil {
			return fmt.Errorf("invalid redis URL: %s", err.Error())
		}

		if utils.IsEmptyStr(c.PoolConfig.RedisPoolCfg.Namespace) {
			return errors.New("namespace of redis worker is required")
		}
	}

	// Job service loggers
	if len(c.LoggerConfigs) == 0 {
		return errors.New("missing logger config of job service")
	}

	// Job loggers
	if len(c.JobLoggerConfigs) == 0 {
		return errors.New("missing logger config of job")
	}

	return nil // valid
}
