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

package chart

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/core/middlewares/util"
)

type chartHandler struct {
	matchers []interceptor.Matcher
	next     http.Handler
}

// New ...
func New(next http.Handler, matchers ...interceptor.Matcher) http.Handler {
	if len(matchers) == 0 {
		matchers = defaultMatchers
	}
	return &chartHandler{
		matchers: matchers,
		next:     next,
	}
}

// ServeHTTP manifest ...
func (h *chartHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	interceptor := h.getInterceptor(req)
	if interceptor == nil {
		h.next.ServeHTTP(rw, req)
		return
	}

	if err := interceptor.HandleRequest(req); err != nil {
		log.Warningf("Error occurred when to handle request in count quota handler: %v", err)
		http.Error(rw, util.MarshalError("InternalError", fmt.Sprintf("Error occurred when to handle request in count quota handler: %v", err)),
			http.StatusInternalServerError)
		return
	}

	w := util.NewCustomResponseWriter(rw)
	h.next.ServeHTTP(w, req)

	interceptor.HandleResponse(w, req)
}

func (h *chartHandler) getInterceptor(req *http.Request) interceptor.Interceptor {
	for _, matcher := range h.matchers {
		if matcher.Match(req) {
			return matcher.SetupInterceptor(req)
		}
	}

	return nil
}
