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

package chart

import (
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/lib/orm"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/handler/util"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// Handler preprocess chart event data
type Handler struct {
}

// Handle preprocess chart event data and then publish hook event
func (cph *Handler) Handle(value interface{}) error {
	chartEvent, ok := value.(*event.ChartEvent)
	if !ok {
		return errors.New("invalid chart event type")
	}

	if chartEvent == nil || len(chartEvent.Versions) == 0 || len(chartEvent.ProjectName) == 0 || len(chartEvent.ChartName) == 0 {
		return fmt.Errorf("data miss in chart event: %v", chartEvent)
	}

	prj, err := project.Ctl.GetByName(orm.Context(), chartEvent.ProjectName, project.Metadata(true))
	if err != nil {
		log.Errorf("failed to find project[%s] for chart event: %v", chartEvent.ProjectName, err)
		return err
	}
	policies, err := notification.PolicyMgr.GetRelatedPolices(prj.ProjectID, chartEvent.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", chartEvent.EventType, err)
		return err
	}
	// if cannot find policy including event type in project, return directly
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", chartEvent.EventType, chartEvent)
		return nil
	}

	payload, err := constructChartPayload(chartEvent, prj)
	if err != nil {
		return err
	}

	err = util.SendHookWithPolicies(policies, payload, chartEvent.EventType)
	if err != nil {
		return err
	}

	return nil
}

// IsStateful ...
func (cph *Handler) IsStateful() bool {
	return false
}

func constructChartPayload(event *event.ChartEvent, project *models.Project) (*model.Payload, error) {
	repoType := models.ProjectPrivate
	if project.IsPublic() {
		repoType = models.ProjectPublic
	}

	payload := &model.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &model.EventData{
			Repository: &model.Repository{
				Name:         event.ChartName,
				Namespace:    event.ProjectName,
				RepoFullName: fmt.Sprintf("%s/%s", event.ProjectName, event.ChartName),
				RepoType:     repoType,
			},
		},
		Operator: event.Operator,
	}

	extURL, err := config.ExtEndpoint()
	if err != nil {
		return nil, fmt.Errorf("get external endpoint failed: %v", err)
	}

	resourcePrefix := fmt.Sprintf("%s/chartrepo/%s/charts/%s", extURL, event.ProjectName, event.ChartName)
	for _, v := range event.Versions {
		resURL := fmt.Sprintf("%s-%s.tgz", resourcePrefix, v)

		resource := &model.Resource{
			Tag:         v,
			ResourceURL: resURL,
		}
		payload.EventData.Resources = append(payload.EventData.Resources, resource)
	}
	return payload, nil
}
