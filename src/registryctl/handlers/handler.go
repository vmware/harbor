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

package handlers

import (
	"net/http"
	"os"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/registryctl/auth"
	"github.com/goharbor/harbor/src/registryctl/config"
	gorilla_handlers "github.com/gorilla/handlers"
)

// NewHandlerChain returns a gorilla router which is wrapped by  authenticate handler
// and logging handler
func NewHandlerChain(conf config.Configuration) http.Handler {
	h := newRouter(conf)
	secrets := map[string]string{
		"jobSecret": os.Getenv("JOBSERVICE_SECRET"),
	}
	insecureAPIs := map[string]bool{
		"/api/health": true,
	}
	h = newAuthHandler(auth.NewSecretHandler(secrets), h, insecureAPIs)
	h = gorilla_handlers.LoggingHandler(os.Stdout, h)
	return h
}

type authHandler struct {
	authenticator auth.AuthenticationHandler
	handler       http.Handler
	insecureAPIs  map[string]bool
}

func newAuthHandler(authenticator auth.AuthenticationHandler, handler http.Handler, insecureAPIs map[string]bool) http.Handler {
	return &authHandler{
		authenticator: authenticator,
		handler:       handler,
		insecureAPIs:  insecureAPIs,
	}
}

func (a *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.authenticator == nil {
		log.Errorf("No authenticator found in registry controller.")
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	if a.insecureAPIs != nil && a.insecureAPIs[r.URL.Path] {
		if a.handler != nil {
			a.handler.ServeHTTP(w, r)
		}
		return
	}

	err := a.authenticator.AuthorizeRequest(r)
	if err != nil {
		log.Errorf("failed to authenticate request: %v", err)
		http.Error(w, http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
		return
	}

	if a.handler != nil {
		a.handler.ServeHTTP(w, r)
	}
	return
}
