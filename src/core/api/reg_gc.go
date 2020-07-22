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
	"errors"
	"github.com/goharbor/harbor/src/core/config"
	"net/http"
	"os"
	"strconv"

	common_job "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/core/api/models"
)

// GCAPI handles request of harbor GC...
type GCAPI struct {
	AJAPI
}

// Prepare validates the URL and parms, it needs the system admin permission.
func (gc *GCAPI) Prepare() {
	gc.BaseController.Prepare()
	if !gc.SecurityCtx.IsAuthenticated() {
		gc.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}
	if !gc.SecurityCtx.IsSysAdmin() {
		gc.SendForbiddenError(errors.New(gc.SecurityCtx.GetUsername()))
		return
	}
}

// Post according to the request, it creates a cron schedule or a manual trigger for GC.
// create a daily schedule for GC
// 	{
//  "schedule": {
//    "type": "Daily",
//    "cron": "0 0 0 * * *"
//  },
//  "parameters": {
//    "delete_untagged": true
//  }
//	}
// create a manual trigger for GC
// 	{
//  "schedule": {
//    "type": "Manual"
//  },
//  "parameters": {
//    "delete_untagged": true
//  }
//	}
func (gc *GCAPI) Post() {
	parameters := make(map[string]interface{})
	ajr := models.AdminJobReq{
		Parameters: parameters,
	}
	isValid, err := gc.DecodeJSONReqAndValidate(&ajr)
	if !isValid {
		gc.SendBadRequestError(err)
		return
	}
	ajr.Name = common_job.ImageGC
	ajr.Parameters["redis_url_reg"] = os.Getenv("_REDIS_URL_REG")
	ajr.Parameters["time_window"] = config.GetGCTimeWindow()
	gc.submit(&ajr)
	gc.Redirect(http.StatusCreated, strconv.FormatInt(ajr.ID, 10))
}

// Put handles GC cron schedule update/delete.
// Request: delete the schedule of GC
// 	{
//  "schedule": {
//    "type": "None",
//    "cron": ""
//  },
//  "parameters": {
//    "delete_untagged": true
//  }
//	}
func (gc *GCAPI) Put() {
	parameters := make(map[string]interface{})
	ajr := models.AdminJobReq{
		Parameters: parameters,
	}
	isValid, err := gc.DecodeJSONReqAndValidate(&ajr)
	if !isValid {
		gc.SendBadRequestError(err)
		return
	}
	ajr.Name = common_job.ImageGC
	ajr.Parameters["redis_url_reg"] = os.Getenv("_REDIS_URL_REG")
	ajr.Parameters["time_window"] = config.GetGCTimeWindow()
	gc.updateSchedule(ajr)
}

// GetGC ...
func (gc *GCAPI) GetGC() {
	id, err := gc.GetInt64FromPath(":id")
	if err != nil {
		gc.SendInternalServerError(errors.New("need to specify gc id"))
		return
	}
	gc.get(id)
}

// List returns the top 10 executions of GC which includes manual and cron.
func (gc *GCAPI) List() {
	gc.list(common_job.ImageGC)
}

// Get gets GC schedule ...
func (gc *GCAPI) Get() {
	gc.getSchedule(common_job.ImageGC)
}

// GetLog ...
func (gc *GCAPI) GetLog() {
	id, err := gc.GetInt64FromPath(":id")
	if err != nil {
		gc.SendBadRequestError(errors.New("invalid ID"))
		return
	}
	gc.getLog(id)
}
