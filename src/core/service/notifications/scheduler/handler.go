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

package scheduler

import (
	"encoding/json"

	"github.com/goharbor/harbor/src/core/service/notifications"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scheduler"
)

// Handler handles the scheduler requests
type Handler struct {
	notifications.BaseHandler
}

// Handle ...
func (h *Handler) Handle() {
	log.Debugf("received scheduler hook event for schedule %s", h.GetStringFromPath(":id"))

	var data job.StatusChange
	if err := json.Unmarshal(h.Ctx.Input.CopyBody(1<<32), &data); err != nil {
		log.Errorf("failed to decode hook event: %v", err)
		return
	}

	schedulerID, err := h.GetInt64FromPath(":id")
	if err != nil {
		log.Errorf("failed to get the schedule ID: %v", err)
		return
	}

	if err = scheduler.HandleLegacyHook(h.Ctx.Request.Context(), schedulerID, &data); err != nil {
		log.Errorf("failed to handle the legacy hook: %v", err)
		return
	}
}
