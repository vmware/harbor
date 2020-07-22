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

package quota

import (
	"fmt"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"net/http"
	"strings"

	cq "github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/server/middleware"
)

var (
	errNonSuccess = errors.New("non success status code")
)

// RequestConfig request resources middleware config
type RequestConfig struct {
	// ReferenceObject returns reference object which resources will be requested
	ReferenceObject func(r *http.Request) (reference string, referenceID string, err error)

	// Resources returns request resources for the reference object
	Resources func(r *http.Request, reference, referenceID string) (types.ResourceList, error)

	// ResourcesWarningPercent value from 0 to 100
	ResourcesWarningPercent int

	// ResourcesWarning returns event which will be notified when resources usage exceeded the wanring percent
	ResourcesWarning func(r *http.Request, reference, referenceID string, message string) event.Metadata

	// ResourcesExceeded returns event which will be notified when resources exceeded the limitation
	ResourcesExceeded func(r *http.Request, reference, referenceID string, message string) event.Metadata
}

// RequestMiddleware middleware which request resources
func RequestMiddleware(config RequestConfig, skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	if config.ResourcesWarningPercent == 0 {
		config.ResourcesWarningPercent = 85 // default 85%
	}

	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		logger := log.G(r.Context()).WithFields(log.Fields{"middleware": "quota", "action": "request", "url": r.URL.Path})

		if config.ReferenceObject == nil || config.Resources == nil {
			lib_http.SendError(w, fmt.Errorf("invald config the for middleware"))
			return
		}

		reference, referenceID, err := config.ReferenceObject(r)
		if err != nil {
			logger.Errorf("get reference object failed, error: %v", err)

			lib_http.SendError(w, err)
			return
		}

		enabled, err := quotaController.IsEnabled(r.Context(), reference, referenceID)
		if err != nil {
			logger.Errorf("check whether quota enabled for %s %s failed, error: %v", reference, referenceID, err)
			lib_http.SendError(w, err)
			return
		}

		if !enabled {
			// quota is disabled for the reference object, so direct to next handler
			logger.Debugf("quota is disabled for %s %s, so direct to next handler", reference, referenceID)
			next.ServeHTTP(w, r)
			return
		}

		resources, err := config.Resources(r, reference, referenceID)
		if err != nil {
			logger.Errorf("get resources failed, error: %v", err)

			lib_http.SendError(w, err)
			return
		}

		if len(resources) == 0 {
			// no resources request for this http request, so direct to next handler
			logger.Debug("no resources request for this http request, so direct to next handler")
			next.ServeHTTP(w, r)
			return
		}

		res, ok := w.(*lib.ResponseBuffer)
		if !ok {
			res = lib.NewResponseBuffer(w)
			defer res.Flush()
		}

		err = quotaController.Request(r.Context(), reference, referenceID, resources, func() error {
			next.ServeHTTP(res, r)
			if !res.Success() {
				return errNonSuccess
			}

			return nil
		})

		if err == nil && config.ResourcesWarning != nil {
			tryWarningNotification := func() {
				q, err := quotaController.GetByRef(r.Context(), reference, referenceID)
				if err != nil {
					logger.Warningf("get quota of %s %s failed, error: %v", reference, referenceID, err)
					return
				}

				resources, err := q.GetWarningResources(config.ResourcesWarningPercent)
				if err != nil {
					logger.Warningf("get warning resources failed, error: %v", err)
					return
				}

				if len(resources) == 0 {
					logger.Warningf("not warning resources found")
					return
				}

				hardLimits, _ := q.GetHard()
				used, _ := q.GetUsed()

				var parts []string
				for _, resource := range resources {
					s := fmt.Sprintf("resource %s used %s of %s",
						resource, resource.FormatValue(used[resource]), resource.FormatValue(hardLimits[resource]))
					parts = append(parts, s)
				}

				message := fmt.Sprintf("quota usage reach %d%%: %s", config.ResourcesWarningPercent, strings.Join(parts, "; "))
				evt := config.ResourcesWarning(r, reference, referenceID, message)
				notification.AddEvent(r.Context(), evt, true)
			}

			tryWarningNotification()
		}

		if err != nil && err != errNonSuccess {
			if config.ResourcesExceeded != nil {
				var errs quota.Errors // NOTE: quota.Errors is slice, so we need var here not pointer
				if errors.As(err, &errs) {
					if exceeded := errs.Exceeded(); exceeded != nil {
						evt := config.ResourcesExceeded(r, reference, referenceID, exceeded.Error())
						notification.AddEvent(r.Context(), evt, true)
					}
				}
			}

			res.Reset()

			var errs quota.Errors
			if errors.As(err, &errs) {
				lib_http.SendError(res, errors.DeniedError(nil).WithMessage(errs.Error()))
			} else {
				lib_http.SendError(res, err)
			}
		}

	}, skippers...)
}

// RefreshConfig refresh quota usage middleware config
type RefreshConfig struct {
	// IgnoreLimitation allow quota usage exceed the limitation when it's true
	IgnoreLimitation bool

	// ReferenceObject returns reference object its quota usage will refresh by reference and reference id
	ReferenceObject func(*http.Request) (reference string, referenceID string, err error)
}

// RefreshMiddleware middleware which refresh the quota usage after the response success
func RefreshMiddleware(config RefreshConfig, skipers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		// skip to refresh quota usage when response is not success
		if !isSuccess(statusCode) {
			return nil
		}

		if config.ReferenceObject == nil {
			return fmt.Errorf("invald config the for middleware")
		}

		logger := log.G(r.Context()).WithFields(log.Fields{"middleware": "quota", "action": "refresh", "url": r.URL.Path})

		reference, referenceID, err := config.ReferenceObject(r)
		if err != nil {
			logger.Errorf("get reference object to refresh quota usage failed, error: %v", err)
			return err
		}

		enabled, err := quotaController.IsEnabled(r.Context(), reference, referenceID)
		if err != nil {
			logger.Errorf("check whether quota enabled for %s %s failed, error: %v", reference, referenceID, err)
			return err
		}

		if !enabled {
			logger.Debugf("quota is disabled for %s %s, so return directly", reference, referenceID)
			return nil
		}

		if err = quotaController.Refresh(r.Context(), reference, referenceID, cq.IgnoreLimitation(config.IgnoreLimitation)); err != nil {
			logger.Errorf("refresh quota for %s %s failed, error: %v", reference, referenceID, err)

			var errs quota.Errors
			if errors.As(err, &errs) {
				return errors.DeniedError(nil).WithMessage(errs.Error())
			}

			return err
		}

		return nil
	}, skipers...)
}

func isSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusBadRequest
}
