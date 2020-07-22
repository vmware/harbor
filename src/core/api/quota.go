// Copyright 2018 Project Harbor Authors
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
	"fmt"

	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/quota/models"
	"github.com/goharbor/harbor/src/pkg/quota/types"
)

// QuotaUpdateRequest struct for the body of put quota API
type QuotaUpdateRequest struct {
	Hard types.ResourceList `json:"hard"`
}

// QuotaAPI handles request to /api/quotas/
type QuotaAPI struct {
	BaseController
	quota *models.Quota
}

// Prepare validates the URL and the user
func (qa *QuotaAPI) Prepare() {
	qa.BaseController.Prepare()

	if !qa.SecurityCtx.IsAuthenticated() {
		qa.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	if !qa.SecurityCtx.IsSysAdmin() {
		qa.SendForbiddenError(errors.New(qa.SecurityCtx.GetUsername()))
		return
	}

	if len(qa.GetStringFromPath(":id")) != 0 {
		id, err := qa.GetInt64FromPath(":id")
		if err != nil || id <= 0 {
			text := "invalid quota ID: "
			if err != nil {
				text += err.Error()
			} else {
				text += fmt.Sprintf("%d", id)
			}
			qa.SendBadRequestError(errors.New(text))
			return
		}

		quota, err := quota.Ctl.Get(qa.Ctx.Request.Context(), id)
		if err != nil {
			qa.SendError(err)
			return
		}

		qa.quota = quota
	}
}

// Get returns quota by id
func (qa *QuotaAPI) Get() {
	qa.Data["json"] = qa.quota
	qa.ServeJSON()
}

// Put update the quota
func (qa *QuotaAPI) Put() {
	var req *QuotaUpdateRequest
	if err := qa.DecodeJSONReq(&req); err != nil {
		qa.SendBadRequestError(err)
		return
	}

	ctx := qa.Ctx.Request.Context()
	if err := quota.Validate(ctx, qa.quota.Reference, req.Hard); err != nil {
		qa.SendBadRequestError(err)
		return
	}

	qa.quota.SetHard(req.Hard)

	if err := quota.Ctl.Update(ctx, qa.quota); err != nil {
		qa.SendInternalServerError(fmt.Errorf("failed to update hard limits of the quota, error: %v", err))
		return
	}
}

// List returns quotas by query
func (qa *QuotaAPI) List() {
	page, size, err := qa.GetPaginationParams()
	if err != nil {
		qa.SendBadRequestError(err)
		return
	}

	query := &q.Query{
		Keywords: q.KeyWords{
			"reference":    qa.GetString("reference"),
			"reference_id": qa.GetString("reference_id"),
		},
		PageNumber: page,
		PageSize:   size,
		Sorting:    qa.GetString("sort"),
	}

	ctx := qa.Ctx.Request.Context()

	total, err := quota.Ctl.Count(ctx, query)
	if err != nil {
		qa.SendInternalServerError(fmt.Errorf("failed to query database for total of quotas, error: %v", err))
		return
	}

	quotas, err := quota.Ctl.List(ctx, query, quota.WithReferenceObject())
	if err != nil {
		qa.SendInternalServerError(fmt.Errorf("failed to query database for quotas, error: %v", err))
		return
	}

	qa.SetPaginationHeader(total, page, size)
	qa.Data["json"] = quotas
	qa.ServeJSON()
}
