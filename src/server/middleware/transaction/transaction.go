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

package transaction

import (
	"context"
	"errors"
	"fmt"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"net/http"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/server/middleware"
)

var (
	errNonSuccess = errors.New("non success status code")
)

type committedKey struct{}

// MustCommit mark http.Request as committed so that transaction
// middleware ignore the status code of the response and commit transaction for this request
func MustCommit(r *http.Request) error {
	committed, ok := r.Context().Value(committedKey{}).(*bool)
	if !ok {
		return fmt.Errorf("%s URL %s is not committable, please enable transaction middleware for it", r.Method, r.URL.Path)
	}

	*committed = true

	return nil
}

// Middleware middleware which add transaction for the http request with default config
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		res, ok := w.(*lib.ResponseBuffer)
		if !ok {
			res = lib.NewResponseBuffer(w)
			defer res.Flush()
		}

		h := func(ctx context.Context) error {
			committed := new(bool) // default false, not must commit

			cc := context.WithValue(ctx, committedKey{}, committed)
			next.ServeHTTP(res, r.WithContext(cc))

			if !(*committed) && !res.Success() {
				return errNonSuccess
			}

			return nil
		}

		if err := orm.WithTransaction(h)(r.Context()); err != nil && err != errNonSuccess {
			err = fmt.Errorf("deal with %s request in transaction failed: %v", r.URL.Path, err)

			// begin, commit or rollback transaction db error happened,
			// reset the response and set status code to 500
			if err := res.Reset(); err != nil {
				log.Errorf("reset the response failed: %v", err)
				return
			}
			lib_http.SendError(res, err)
		}
	}, skippers...)
}
