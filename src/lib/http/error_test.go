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

package http

import (
	std_errors "errors"
	openapi "github.com/go-openapi/errors"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendError(t *testing.T) {
	// unauthorized error
	rw := httptest.NewRecorder()
	err := errors.New(nil).WithCode(errors.UnAuthorizedCode).WithMessage("unauthorized")
	SendError(rw, err)
	assert.Equal(t, http.StatusUnauthorized, rw.Code)
	assert.Equal(t, `{"errors":[{"code":"UNAUTHORIZED","message":"unauthorized"}]}`+"\n", rw.Body.String())

	// internal server error
	rw = httptest.NewRecorder()
	err = errors.New(nil).WithCode(errors.GeneralCode).WithMessage("unknown")
	SendError(rw, err)
	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Equal(t, `{"errors":[{"code":"UNKNOWN","message":"internal server error"}]}`+"\n", rw.Body.String())

	// not internal server error
	rw = httptest.NewRecorder()
	err = errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("object not found")
	SendError(rw, err)
	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, `{"errors":[{"code":"NOT_FOUND","message":"object not found"}]}`+"\n", rw.Body.String())
}

func TestAPIError(t *testing.T) {
	var err error
	// open API error: github.com/go-openapi/errors.Error
	err = openapi.New(400, "bad request")
	statusCode, payload, stacktrace := apiError(err)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, `{"errors":[{"code":"BAD_REQUEST","message":"bad request"}]}`, payload)
	assert.Contains(t, stacktrace, `http.apiError`)

	// legacy error
	err = &commonhttp.Error{
		Code:    http.StatusNotFound,
		Message: "not found",
	}
	statusCode, payload, stacktrace = apiError(err)
	assert.Equal(t, http.StatusNotFound, statusCode)
	assert.Equal(t, `{"errors":[{"code":"NOT_FOUND","message":"not found"}]}`, payload)
	assert.Contains(t, stacktrace, `http.apiError`)

	// errors.Error
	err = errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("resource not found")
	statusCode, payload, stacktrace = apiError(err)
	assert.Equal(t, http.StatusNotFound, statusCode)
	assert.Equal(t, `{"errors":[{"code":"NOT_FOUND","message":"resource not found"}]}`, payload)
	assert.Contains(t, stacktrace, `http.TestAPIError`)

	// common error, common error has no stacktrace
	e := std_errors.New("customized error")
	statusCode, payload, stacktrace = apiError(e)
	assert.Equal(t, http.StatusInternalServerError, statusCode)
	assert.Equal(t, `{"errors":[{"code":"UNKNOWN","message":"unknown: customized error"}]}`, payload)
	assert.Contains(t, stacktrace, ``)

}
