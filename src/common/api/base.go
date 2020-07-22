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

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"net/http"
	"strconv"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/validation"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	defaultPageSize int64 = 500
	maxPageSize     int64 = 500

	// APIVersion is the current core api version
	APIVersion = "v2.0"
)

// BaseAPI wraps common methods for controllers to host API
type BaseAPI struct {
	beego.Controller
}

// GetStringFromPath gets the param from path and returns it as string
func (b *BaseAPI) GetStringFromPath(key string) string {
	return b.Ctx.Input.Param(key)
}

// GetInt64FromPath gets the param from path and returns it as int64
func (b *BaseAPI) GetInt64FromPath(key string) (int64, error) {
	value := b.Ctx.Input.Param(key)
	return strconv.ParseInt(value, 10, 64)
}

// ParamExistsInPath returns true when param exists in the path
func (b *BaseAPI) ParamExistsInPath(key string) bool {
	return b.GetStringFromPath(key) != ""
}

// Render returns nil as it won't render template
func (b *BaseAPI) Render() error {
	return nil
}

// RenderError provides shortcut to render http error
func (b *BaseAPI) RenderError(code int, text string) {
	lib_http.SendError(b.Ctx.ResponseWriter, &commonhttp.Error{
		Code:    code,
		Message: text,
	})
}

// DecodeJSONReq decodes a json request
func (b *BaseAPI) DecodeJSONReq(v interface{}) error {
	err := json.Unmarshal(b.Ctx.Input.CopyBody(1<<32), v)
	if err != nil {
		log.Errorf("Error while decoding the json request, error: %v, %v",
			err, string(b.Ctx.Input.CopyBody(1 << 32)[:]))
		return errors.New("Invalid json request")
	}
	return nil
}

// Validate validates v if it implements interface validation.ValidFormer
func (b *BaseAPI) Validate(v interface{}) (bool, error) {
	validator := validation.Validation{}
	isValid, err := validator.Valid(v)
	if err != nil {
		log.Errorf("failed to validate: %v", err)
		return false, err
	}

	if !isValid {
		message := ""
		for _, e := range validator.Errors {
			message += fmt.Sprintf("%s %s \n", e.Field, e.Message)
		}
		return false, errors.New(message)
	}
	return true, nil
}

// DecodeJSONReqAndValidate does both decoding and validation
func (b *BaseAPI) DecodeJSONReqAndValidate(v interface{}) (bool, error) {
	if err := b.DecodeJSONReq(v); err != nil {
		return false, err
	}
	return b.Validate(v)
}

// Redirect does redirection to resource URI with http header status code.
func (b *BaseAPI) Redirect(statusCode int, resouceID string) {
	requestURI := b.Ctx.Request.RequestURI
	resourceURI := requestURI + "/" + resouceID

	b.Ctx.Redirect(statusCode, resourceURI)
}

// GetIDFromURL checks the ID in request URL
func (b *BaseAPI) GetIDFromURL() (int64, error) {
	idStr := b.Ctx.Input.Param(":id")
	if len(idStr) == 0 {
		return 0, errors.New("invalid ID in URL")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid ID in URL")
	}

	return id, nil
}

// SetPaginationHeader set"Link" and "X-Total-Count" header for pagination request
func (b *BaseAPI) SetPaginationHeader(total, page, pageSize int64) {
	b.Ctx.ResponseWriter.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))

	link := ""

	// SetPaginationHeader set previous link
	if page > 1 && (page-1)*pageSize <= total {
		u := *(b.Ctx.Request.URL)
		q := u.Query()
		q.Set("page", strconv.FormatInt(page-1, 10))
		u.RawQuery = q.Encode()
		if len(link) != 0 {
			link += ", "
		}
		link += fmt.Sprintf("<%s>; rel=\"prev\"", u.String())
	}

	// SetPaginationHeader set next link
	if pageSize*page < total {
		u := *(b.Ctx.Request.URL)
		q := u.Query()
		q.Set("page", strconv.FormatInt(page+1, 10))
		u.RawQuery = q.Encode()
		if len(link) != 0 {
			link += ", "
		}
		link += fmt.Sprintf("<%s>; rel=\"next\"", u.String())
	}

	if len(link) != 0 {
		b.Ctx.ResponseWriter.Header().Set("Link", link)
	}
}

// GetPaginationParams ...
func (b *BaseAPI) GetPaginationParams() (page, pageSize int64, err error) {
	page, err = b.GetInt64("page", 1)
	if err != nil || page <= 0 {
		return 0, 0, errors.New("invalid page")
	}

	pageSize, err = b.GetInt64("page_size", defaultPageSize)
	if err != nil || pageSize <= 0 {
		return 0, 0, errors.New("invalid page_size")
	}

	if pageSize > maxPageSize {
		pageSize = maxPageSize
		log.Debugf("the parameter page_size %d exceeds the max %d, set it to max", pageSize, maxPageSize)
	}

	return page, pageSize, nil
}

// ParseAndHandleError : if the err is an instance of utils/error.Error,
// return the status code and the detail message contained in err, otherwise
// return 500
func (b *BaseAPI) ParseAndHandleError(text string, err error) {
	if err == nil {
		return
	}
	if e, ok := err.(*commonhttp.Error); ok {
		b.RenderError(e.Code, fmt.Sprintf("%s: %s", text, e.Message))
		return
	}
	b.SendInternalServerError(fmt.Errorf("%s: %v", text, err))
}

// SendUnAuthorizedError sends unauthorized error to the client.
func (b *BaseAPI) SendUnAuthorizedError(err error) {
	b.RenderError(http.StatusUnauthorized, err.Error())
}

// SendConflictError sends conflict error to the client.
func (b *BaseAPI) SendConflictError(err error) {
	b.RenderError(http.StatusConflict, err.Error())
}

// SendNotFoundError sends not found error to the client.
func (b *BaseAPI) SendNotFoundError(err error) {
	b.RenderError(http.StatusNotFound, err.Error())
}

// SendBadRequestError sends bad request error to the client.
func (b *BaseAPI) SendBadRequestError(err error) {
	b.RenderError(http.StatusBadRequest, err.Error())
}

// SendInternalServerError sends internal server error to the client.
// Note the detail info of err will not include in the response body.
// When you send an internal server error  to the client, you expect user to check the log
// to find out the root cause.
func (b *BaseAPI) SendInternalServerError(err error) {
	b.RenderError(http.StatusInternalServerError, err.Error())
}

// SendForbiddenError sends forbidden error to the client.
func (b *BaseAPI) SendForbiddenError(err error) {
	b.RenderError(http.StatusForbidden, err.Error())
}

// SendPreconditionFailedError sends conflict error to the client.
func (b *BaseAPI) SendPreconditionFailedError(err error) {
	b.RenderError(http.StatusPreconditionFailed, err.Error())
}

// SendStatusServiceUnavailableError sends service unavailable error to the client.
func (b *BaseAPI) SendStatusServiceUnavailableError(err error) {
	b.RenderError(http.StatusServiceUnavailable, err.Error())
}

// SendError return the error defined in OCI spec: https://github.com/opencontainers/distribution-spec/blob/master/spec.md#errors
// {
//	"errors:" [{
//			"code": <error identifier>,
//			"message": <message describing condition>,
//			// optional
//			"detail": <unstructured>
//		},
//		...
//	]
// }
func (b *BaseAPI) SendError(err error) {
	lib_http.SendError(b.Ctx.ResponseWriter, err)
}
