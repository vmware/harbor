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
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/rbac/system"
	"github.com/goharbor/harbor/src/lib/config"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	authHeader = "Authorization"
)

type reqChecker struct {
	ctl project.Controller
}

func (rc *reqChecker) check(req *http.Request) (string, error) {
	securityCtx, ok := security.FromContext(req.Context())
	if !ok {
		return "", fmt.Errorf("the security context got from request is nil")
	}
	al := accessList(req)
	if len(al) == 0 {
		return "", fmt.Errorf("un-recognized request: %s %s", req.Method, req.URL.Path)
	}
	for _, a := range al {
		if a.target == login && !securityCtx.IsAuthenticated() {
			return getChallenge(req, al), errors.New("unauthorized")
		}
		if a.target == catalog {
			resource := system.NewNamespace().Resource(rbac.ResourceCatalog)
			if !securityCtx.Can(req.Context(), rbac.ActionRead, resource) {
				return getChallenge(req, al), fmt.Errorf("unauthorized to list catalog")
			}
		}
		if a.target == repository && req.Header.Get(authHeader) == "" &&
			(req.Method == http.MethodHead || req.Method == http.MethodGet) { // make sure 401 is returned for CLI HEAD, see #11271
			return getChallenge(req, al), fmt.Errorf("authorize header needed to send HEAD to repository")
		} else if a.target == repository {
			pn := strings.Split(a.name, "/")[0]
			pid, err := rc.projectID(req.Context(), pn)
			if err != nil {
				return "", err
			}
			resource := rbac_project.NewNamespace(pid).Resource(rbac.ResourceRepository)
			if !securityCtx.Can(req.Context(), a.action, resource) {
				return getChallenge(req, al), fmt.Errorf("unauthorized to access repository: %s, action: %s", a.name, a.action)
			}
		}
	}
	return "", nil
}

func (rc *reqChecker) projectID(ctx context.Context, name string) (int64, error) {
	p, err := rc.ctl.Get(ctx, name)
	if err != nil {
		return 0, err
	}

	return p.ProjectID, nil
}

func getChallenge(req *http.Request, accessList []access) string {
	logger := log.G(req.Context())
	auth := req.Header.Get(authHeader)
	if len(auth) > 0 || lib.V2CatalogURLRe.MatchString(req.URL.Path) {
		// Return basic auth challenge by default, incl. request to '/v2/_catalog'
		return `Basic realm="harbor"`
	}
	// No auth header, treat it as CLI and redirect to token service
	tokenSvc, err := tokenSvcEndpointSvc(req)
	if err != nil {
		logger.Errorf("failed to get the token service url, error: %v", err)
	}
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

func tokenSvcEndpointSvc(req *http.Request) (string, error) {
	rawCoreURL := config.InternalCoreURL()
	if match(req.Context(), req.Host, rawCoreURL) {
		return formatServiceTokenUrl(rawCoreURL), nil
	}

	if serviceUrl, err := config.TokenServiceURL(); serviceUrl != "" {
		return serviceUrl, err
	}
	extEndpoint, err := config.ExtEndpoint()

	return formatServiceTokenUrl(extEndpoint), err
}

func formatServiceTokenUrl(url string) string {
	return fmt.Sprintf("%s/service/token", strings.TrimSuffix(url, "/"))
}

func match(ctx context.Context, reqHost, rawURL string) bool {
	logger := log.G(ctx)
	cfgURL, err := url.Parse(rawURL)
	if err != nil {
		logger.Errorf("Failed to parse url: %s, error: %v", rawURL, err)
		return false
	}
	if cfgURL.Scheme == "http" && cfgURL.Port() == "80" ||
		cfgURL.Scheme == "https" && cfgURL.Port() == "443" {
		cfgURL.Host = cfgURL.Hostname()
	}
	if cfgURL.Scheme == "http" && strings.HasSuffix(reqHost, ":80") {
		reqHost = strings.TrimSuffix(reqHost, ":80")
	}
	if cfgURL.Scheme == "https" && strings.HasSuffix(reqHost, ":443") {
		reqHost = strings.TrimSuffix(reqHost, ":443")
	}
	return reqHost == cfgURL.Host
}

var (
	once    sync.Once
	checker reqChecker
)

// Middleware checks the permission of the request to access the artifact
func Middleware() func(http.Handler) http.Handler {
	once.Do(func() {
		if checker.ctl == nil { // for UT, where ctl has been set to a mock value
			checker = reqChecker{
				ctl: project.Ctl,
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
