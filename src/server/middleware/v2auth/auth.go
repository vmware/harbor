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

package v2auth

import (
	"fmt"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	authHeader = "Authorization"
)

type reqChecker struct {
	pm promgr.ProjectManager
}

func (rc *reqChecker) check(req *http.Request) (string, error) {
	securityCtx, ok := security.FromContext(req.Context())
	if !ok {
		return "", fmt.Errorf("the security context got from request is nil")
	}
	al := accessList(req)

	for _, a := range al {
		if a.target == login && !securityCtx.IsAuthenticated() {
			return getChallenge(req, al), errors.New("unauthorized")
		}
		if a.target == catalog && !securityCtx.IsSysAdmin() {
			return getChallenge(req, al), fmt.Errorf("unauthorized to list catalog")
		}
		if a.target == repository && req.Header.Get(authHeader) == "" && req.Method == http.MethodHead { // make sure 401 is returned for CLI HEAD, see #11271
			return getChallenge(req, al), fmt.Errorf("authorize header needed to send HEAD to repository")
		} else if a.target == repository {
			pn := strings.Split(a.name, "/")[0]
			pid, err := rc.projectID(pn)
			if err != nil {
				return "", err
			}
			resource := rbac.NewProjectNamespace(pid).Resource(rbac.ResourceRepository)
			if !securityCtx.Can(a.action, resource) {
				return getChallenge(req, al), fmt.Errorf("unauthorized to access repository: %s, action: %s", a.name, a.action)
			}
		}
	}
	return "", nil
}

func (rc *reqChecker) projectID(name string) (int64, error) {
	p, err := rc.pm.Get(name)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, fmt.Errorf("project not found, name: %s", name)
	}
	return p.ProjectID, nil
}

func getChallenge(req *http.Request, accessList []access) string {
	logger := log.G(req.Context())
	auth := req.Header.Get(authHeader)
	if len(auth) > 0 {
		// Return basic auth challenge by default
		return `Basic realm="harbor"`
	}
	// No auth header, treat it as CLI and redirect to token service
	ep, err := tokenSvcEndpoint(req)
	if err != nil {
		logger.Errorf("failed to get the endpoint for token service, error: %v", err)
	}
	tokenSvc := fmt.Sprintf("%s/service/token", strings.TrimSuffix(ep, "/"))
	scope := ""
	for _, a := range accessList {
		if len(scope) > 0 {
			scope += " "
		}
		scope += a.scopeStr(req.Context())
	}
	challenge := fmt.Sprintf(`Bearer realm="%s",service="%s"`, tokenSvc, token.Registry)
	if len(scope) > 0 {
		challenge = fmt.Sprintf(`%s,scope="%s"`, challenge, scope)
	}
	return challenge
}

func tokenSvcEndpoint(req *http.Request) (string, error) {
	logger := log.G(req.Context())
	rawCoreURL := config.InternalCoreURL()
	if coreURL, err := url.Parse(rawCoreURL); err == nil {
		if req.Host == coreURL.Host {
			return rawCoreURL, nil
		}
	} else {
		logger.Errorf("Failed to parse core url, error: %v, fallback to external endpoint", err)
	}
	return config.ExtEndpoint()
}

var (
	once    sync.Once
	checker reqChecker
)

// Middleware checks the permission of the request to access the artifact
func Middleware() func(http.Handler) http.Handler {
	once.Do(func() {
		if checker.pm == nil { // for UT, where pm has been set to a mock value
			checker = reqChecker{
				pm: config.GlobalProjectMgr,
			}
		}
	})
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if challenge, err := checker.check(req); err != nil {
				// the header is needed for "docker manifest" commands: https://github.com/docker/cli/issues/989
				rw.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
				rw.Header().Set("Www-Authenticate", challenge)
				lib_http.SendError(rw, errors.UnauthorizedError(err).WithMessage(err.Error()))
				return
			}
			next.ServeHTTP(rw, req)
		})
	}
}
