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
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scan/allowlist"
	"net/http"
)

// SysCVEAllowlistAPI Handles the requests to manage system level CVE allowlist
type SysCVEAllowlistAPI struct {
	BaseController
	manager allowlist.Manager
}

// Prepare validates the request initially
func (sca *SysCVEAllowlistAPI) Prepare() {
	sca.BaseController.Prepare()
	if !sca.SecurityCtx.IsAuthenticated() {
		sca.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}
	if !sca.SecurityCtx.IsSysAdmin() && sca.Ctx.Request.Method != http.MethodGet {
		msg := fmt.Sprintf("only system admin has permission issue %s request to this API", sca.Ctx.Request.Method)
		log.Errorf(msg)
		sca.SendForbiddenError(errors.New(msg))
		return
	}
	sca.manager = allowlist.NewDefaultManager()
}

// Get handles the GET request to retrieve the system level CVE allowlist
func (sca *SysCVEAllowlistAPI) Get() {
	l, err := sca.manager.GetSys()
	if err != nil {
		sca.SendInternalServerError(err)
		return
	}
	sca.WriteJSONData(l)
}

// Put handles the PUT request to update the system level CVE allowlist
func (sca *SysCVEAllowlistAPI) Put() {
	var l models.CVEAllowlist
	if err := sca.DecodeJSONReq(&l); err != nil {
		log.Errorf("Failed to decode JSON array from request")
		sca.SendBadRequestError(err)
		return
	}
	if l.ProjectID != 0 {
		msg := fmt.Sprintf("Non-zero project ID for system CVE allowlist: %d.", l.ProjectID)
		log.Error(msg)
		sca.SendBadRequestError(errors.New(msg))
		return
	}
	if err := sca.manager.SetSys(l); err != nil {
		if allowlist.IsInvalidErr(err) {
			log.Errorf("Invalid CVE allowlist: %v", err)
			sca.SendBadRequestError(err)
			return
		}
		sca.SendInternalServerError(err)
		return
	}
}
