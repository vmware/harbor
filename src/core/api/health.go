// Copyright 2019 Project Harbor Authors
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

package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/distribution/distribution/health"
	"github.com/goharbor/harbor/src/common/dao"
	httputil "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/redis"
)

var (
	timeout = 60 * time.Second
	// HealthCheckerRegistry ...
	HealthCheckerRegistry = map[string]health.Checker{}
)

type overallHealthStatus struct {
	Status     string                   `json:"status"`
	Components []*componentHealthStatus `json:"components"`
}

type componentHealthStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type healthy bool

func (h healthy) String() string {
	if h {
		return "healthy"
	}
	return "unhealthy"
}

// HealthAPI handles the request for "/api/health"
type HealthAPI struct {
	BaseController
}

// CheckHealth checks the health of system
func (h *HealthAPI) CheckHealth() {
	var isHealthy healthy = true
	components := []*componentHealthStatus{}
	c := make(chan *componentHealthStatus, len(HealthCheckerRegistry))
	for name, checker := range HealthCheckerRegistry {
		go check(name, checker, timeout, c)
	}
	for i := 0; i < len(HealthCheckerRegistry); i++ {
		componentStatus := <-c
		if len(componentStatus.Error) != 0 {
			isHealthy = false
		}
		components = append(components, componentStatus)
	}

	sort.Slice(components, func(i, j int) bool { return components[i].Name < components[j].Name })

	status := &overallHealthStatus{}
	status.Status = isHealthy.String()
	status.Components = components
	if !isHealthy {
		log.Debugf("unhealthy system status: %v", status)
	}
	h.WriteJSONData(status)
}

func check(name string, checker health.Checker,
	timeout time.Duration, c chan *componentHealthStatus) {
	statusChan := make(chan *componentHealthStatus)
	go func() {
		err := checker.Check()
		var healthy healthy = err == nil
		status := &componentHealthStatus{
			Name:   name,
			Status: healthy.String(),
		}
		if !healthy {
			status.Error = err.Error()
		}
		statusChan <- status
	}()

	select {
	case status := <-statusChan:
		c <- status
	case <-time.After(timeout):
		var healthy healthy = false
		c <- &componentHealthStatus{
			Name:   name,
			Status: healthy.String(),
			Error:  "failed to check the health status: timeout",
		}
	}
}

// HTTPStatusCodeHealthChecker implements a Checker to check that the HTTP status code
// returned matches the expected one
func HTTPStatusCodeHealthChecker(method string, url string, header http.Header,
	timeout time.Duration, statusCode int) health.Checker {
	return health.CheckFunc(func() error {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %v", err)
		}
		for key, values := range header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		client := httputil.NewClient(&http.Client{
			Transport: httputil.GetHTTPTransport(httputil.SecureTransport),
			Timeout:   timeout,
		})
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to check health: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != statusCode {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Debugf("failed to read response body: %v", err)
			}
			return fmt.Errorf("received unexpected status code: %d %s", resp.StatusCode, string(data))
		}

		return nil
	})
}

type updater struct {
	sync.Mutex
	status error
}

func (u *updater) Check() error {
	u.Lock()
	defer u.Unlock()

	return u.status
}

func (u *updater) update(status error) {
	u.Lock()
	defer u.Unlock()

	u.status = status
}

// PeriodicHealthChecker implements a Checker to check status periodically
func PeriodicHealthChecker(checker health.Checker, period time.Duration) health.Checker {
	u := &updater{
		// init the "status" as "unknown status" error to avoid returning nil error(which means healthy)
		// before the first health check request finished
		status: errors.New("unknown status"),
	}

	go func() {
		ticker := time.NewTicker(period)
		for {
			u.update(checker.Check())
			<-ticker.C
		}
	}()

	return u
}

func coreHealthChecker() health.Checker {
	return health.CheckFunc(func() error {
		return nil
	})
}

func portalHealthChecker() health.Checker {
	url := config.GetPortalURL()
	timeout := 60 * time.Second
	period := 10 * time.Second
	checker := HTTPStatusCodeHealthChecker(http.MethodGet, url, nil, timeout, http.StatusOK)
	return PeriodicHealthChecker(checker, period)
}

func jobserviceHealthChecker() health.Checker {
	url := config.InternalJobServiceURL() + "/api/v1/stats"
	timeout := 60 * time.Second
	period := 10 * time.Second
	checker := HTTPStatusCodeHealthChecker(http.MethodGet, url, nil, timeout, http.StatusOK)
	return PeriodicHealthChecker(checker, period)
}

func registryHealthChecker() health.Checker {
	url := getRegistryURL() + "/"
	timeout := 60 * time.Second
	period := 10 * time.Second
	checker := HTTPStatusCodeHealthChecker(http.MethodGet, url, nil, timeout, http.StatusOK)
	return PeriodicHealthChecker(checker, period)
}

func registryCtlHealthChecker() health.Checker {
	url := config.GetRegistryCtlURL() + "/api/health"
	timeout := 60 * time.Second
	period := 10 * time.Second
	checker := HTTPStatusCodeHealthChecker(http.MethodGet, url, nil, timeout, http.StatusOK)
	return PeriodicHealthChecker(checker, period)
}

func chartmuseumHealthChecker() health.Checker {
	url, err := config.GetChartMuseumEndpoint()
	if err != nil {
		log.Errorf("failed to get the URL of chartmuseum: %v", err)
	}
	url = url + "/health"
	timeout := 60 * time.Second
	period := 10 * time.Second
	checker := HTTPStatusCodeHealthChecker(http.MethodGet, url, nil, timeout, http.StatusOK)
	return PeriodicHealthChecker(checker, period)
}

func notaryHealthChecker() health.Checker {
	url := config.InternalNotaryEndpoint() + "/_notary_server/health"
	timeout := 60 * time.Second
	period := 10 * time.Second
	checker := HTTPStatusCodeHealthChecker(http.MethodGet, url, nil, timeout, http.StatusOK)
	return PeriodicHealthChecker(checker, period)
}

func databaseHealthChecker() health.Checker {
	period := 10 * time.Second
	checker := health.CheckFunc(func() error {
		_, err := dao.GetOrmer().Raw("SELECT 1").Exec()
		if err != nil {
			return fmt.Errorf("failed to run SQL \"SELECT 1\": %v", err)
		}
		return nil
	})
	return PeriodicHealthChecker(checker, period)
}

func redisHealthChecker() health.Checker {
	period := 10 * time.Second
	checker := health.CheckFunc(func() error {
		conn := redis.DefaultPool().Get()
		defer conn.Close()
		return nil
	})
	return PeriodicHealthChecker(checker, period)
}

func trivyHealthChecker() health.Checker {
	url := strings.TrimSuffix(config.TrivyAdapterURL(), "/") + "/probe/healthy"
	timeout := 60 * time.Second
	period := 10 * time.Second
	checker := HTTPStatusCodeHealthChecker(http.MethodGet, url, nil, timeout, http.StatusOK)
	return PeriodicHealthChecker(checker, period)
}

func registerHealthCheckers() {
	HealthCheckerRegistry["core"] = coreHealthChecker()
	HealthCheckerRegistry["portal"] = portalHealthChecker()
	HealthCheckerRegistry["jobservice"] = jobserviceHealthChecker()
	HealthCheckerRegistry["registry"] = registryHealthChecker()
	HealthCheckerRegistry["registryctl"] = registryCtlHealthChecker()
	HealthCheckerRegistry["database"] = databaseHealthChecker()
	HealthCheckerRegistry["redis"] = redisHealthChecker()
	if config.WithChartMuseum() {
		HealthCheckerRegistry["chartmuseum"] = chartmuseumHealthChecker()
	}
	if config.WithNotary() {
		HealthCheckerRegistry["notary"] = notaryHealthChecker()
	}
	if config.WithTrivy() {
		HealthCheckerRegistry["trivy"] = trivyHealthChecker()
	}
}

func getRegistryURL() string {
	endpoint, err := config.RegistryURL()
	if err != nil {
		log.Errorf("failed to get the URL of registry: %v", err)
		return ""
	}
	url, err := utils.ParseEndpoint(endpoint)
	if err != nil {
		log.Errorf("failed to parse the URL of registry: %v", err)
		return ""
	}
	return url.String()
}
