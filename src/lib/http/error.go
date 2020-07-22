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
	"fmt"
	openapi "github.com/go-openapi/errors"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"net/http"
	"strings"
)

var (
	codeMap = map[string]int{
		errors.BadRequestCode:                  http.StatusBadRequest,
		errors.DIGESTINVALID:                   http.StatusBadRequest,
		errors.MANIFESTINVALID:                 http.StatusBadRequest,
		errors.UNSUPPORTED:                     http.StatusBadRequest,
		errors.UnAuthorizedCode:                http.StatusUnauthorized,
		errors.ForbiddenCode:                   http.StatusForbidden,
		errors.MethodNotAllowedCode:            http.StatusMethodNotAllowed,
		errors.DENIED:                          http.StatusForbidden,
		errors.NotFoundCode:                    http.StatusNotFound,
		errors.ConflictCode:                    http.StatusConflict,
		errors.PreconditionCode:                http.StatusPreconditionFailed,
		errors.ViolateForeignKeyConstraintCode: http.StatusPreconditionFailed,
		errors.PROJECTPOLICYVIOLATION:          http.StatusPreconditionFailed,
		errors.GeneralCode:                     http.StatusInternalServerError,
	}
)

// SendError tries to parse the HTTP status code from the specified error, envelops it into
// an error array as the error payload and returns the code and payload to the response.
// And the error is logged as well
func SendError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	statusCode, errPayload, stackTrace := apiError(err)
	// the error detail is logged only, and will not be sent to the client to avoid leaking server information
	if statusCode >= http.StatusInternalServerError {
		log.Errorf("%s %s", errPayload, stackTrace)
		err = errors.New(nil).WithCode(errors.GeneralCode).WithMessage("internal server error")
		errPayload = errors.NewErrs(err).Error()
	} else {
		// only log the error whose status code < 500 when debugging to avoid log flooding
		log.Debug(errPayload)
	}
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, errPayload)
}

// generates the HTTP status code based on the specified error,
// envelops the error into an error array as the payload and return them
func apiError(err error) (statusCode int, errPayload, stackTrace string) {
	code := 0
	var openAPIErr openapi.Error
	if errors.As(err, &openAPIErr) {
		// Before executing operation handler, go-swagger will bind a parameters object to a request and validate the request,
		// it will return directly when bind and validate failed.
		// The response format of the default ServeError implementation does not match the internal error response format.
		// So we needed to convert the format to the internal error response format.
		code = int(openAPIErr.Code())
		errCode := strings.Replace(strings.ToUpper(http.StatusText(code)), " ", "_", -1)
		err = errors.New(nil).WithCode(errCode).WithMessage(openAPIErr.Error())
	} else if legacyErr, ok := err.(*commonhttp.Error); ok {
		// make sure the legacy error format is align with the new one
		code = legacyErr.Code
		errCode := strings.Replace(strings.ToUpper(http.StatusText(code)), " ", "_", -1)
		err = errors.New(nil).WithCode(errCode).WithMessage(legacyErr.Message)
	} else {
		code = codeMap[errors.ErrCode(err)]
	}
	if code == 0 {
		code = http.StatusInternalServerError
	}
	fullStack := ""
	if _, ok := err.(*errors.Error); ok {
		fullStack = err.(*errors.Error).StackTrace()
	}
	return code, errors.NewErrs(err).Error(), fullStack
}
