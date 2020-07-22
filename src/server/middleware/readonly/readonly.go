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

package readonly

import (
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"net/http"

	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/server/middleware"
)

// Config defines the config for ReadOnly middleware.
type Config struct {
	// ReadOnly defines a function to check whether is readonly mode for request
	ReadOnly func(*http.Request) bool
}

var (
	// DefaultConfig default readonly config
	DefaultConfig = Config{
		ReadOnly: func(r *http.Request) bool {
			return config.ReadOnly()
		},
	}

	// See more for safe method at https://developer.mozilla.org/en-US/docs/Glossary/safe
	safeMethods = map[string]bool{
		http.MethodGet:     true,
		http.MethodHead:    true,
		http.MethodOptions: true,
	}
)

// safeMethodSkipper returns false when the request method is safe methods
func safeMethodSkipper(r *http.Request) bool {
	return safeMethods[r.Method]
}

// Middleware middleware reject request when harbor set to readonly with default config
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return MiddlewareWithConfig(DefaultConfig, skippers...)
}

// MiddlewareWithConfig middleware reject request when harbor set to readonly with config
func MiddlewareWithConfig(config Config, skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	if len(skippers) == 0 {
		skippers = []middleware.Skipper{safeMethodSkipper}
	} else {
		skippers = append(skippers, []middleware.Skipper{safeMethodSkipper}...)
	}

	if config.ReadOnly == nil {
		config.ReadOnly = DefaultConfig.ReadOnly
	}

	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if config.ReadOnly(r) {
			pkgE := errors.New(nil).WithCode(errors.DENIED).WithMessage("The system is in read only mode. Any modification is prohibited.")
			lib_http.SendError(w, pkgE)
			return
		}

		next.ServeHTTP(w, r)
	}, skippers...)
}
