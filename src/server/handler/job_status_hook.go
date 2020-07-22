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

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	libhttp "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/router"
)

// NewJobStatusHandler creates a handler to handle the job status changing
func NewJobStatusHandler() http.Handler {
	return &jobStatusHandler{
		handler: task.NewHookHandler(),
	}
}

type jobStatusHandler struct {
	handler *task.HookHandler
}

func (j *jobStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	taskIDParam := router.Param(r.Context(), ":id")
	taskID, err := strconv.ParseInt(taskIDParam, 10, 64)
	if err != nil {
		libhttp.SendError(w, err)
		return
	}

	sc := &job.StatusChange{}
	if err = json.NewDecoder(r.Body).Decode(sc); err != nil {
		libhttp.SendError(w, err)
		return
	}

	if err = j.handler.Handle(r.Context(), taskID, sc); err != nil {
		// ignore the not found error to avoid the jobservice re-sending the hook
		if errors.IsNotFoundErr(err) {
			log.Warningf("got the status change hook for a non existing task %d", taskID)
			return
		}
		libhttp.SendError(w, err)
		return
	}
}
