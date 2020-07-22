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

	_ "github.com/docker/distribution/registry/storage/driver/filesystem"
)

func TestConfigDoesNotExists(t *testing.T) {
	cfg := &Configuration{}
	err := cfg.Load("./config.not-existing.yaml", false)
	assert.NotNil(t, err)
}

func TestConfigLoadingWithEnv(t *testing.T) {
	os.Setenv("REGISTRYCTL_PROTOCOL", "https")
	os.Setenv("PORT", "1000")
	os.Setenv("LOG_LEVEL", "DEBUG")

	cfg := &Configuration{}
	err := cfg.Load("../config_test.yml", true)
	assert.Nil(t, err)
	assert.Equal(t, "https", cfg.Protocol)
	assert.Equal(t, "1000", cfg.Port)
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, "../reg_conf_test.yml", cfg.RegistryConfig)
}

func TestConfigLoadingWithYml(t *testing.T) {
	cfg := &Configuration{}
	err := cfg.Load("../config_test.yml", false)
	assert.Nil(t, err)
	assert.Equal(t, "http", cfg.Protocol)
	assert.Equal(t, "1234", cfg.Port)
	assert.Equal(t, "ERROR", cfg.LogLevel)
	assert.Equal(t, "../reg_conf_test.yml", cfg.RegistryConfig)
	assert.True(t, cfg.StorageDriver.Name() == "filesystem")
}

func TestGetLogLevel(t *testing.T) {
	err := DefaultConfig.Load("../config_test.yml", false)
	assert.Nil(t, err)
	assert.Equal(t, "ERROR", GetLogLevel())
}

func TestGetJobAuthSecret(t *testing.T) {
	os.Setenv("JOBSERVICE_SECRET", "test_job_secret")
	assert.Equal(t, "test_job_secret", GetJobAuthSecret())
}

func TestGetUIAuthSecret(t *testing.T) {
	os.Setenv("CORE_SECRET", "test_core_secret")
	assert.Equal(t, "test_core_secret", GetUIAuthSecret())
}
