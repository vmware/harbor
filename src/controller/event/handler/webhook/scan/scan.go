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

package scan

import (
	"context"
	"time"

	o "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/handler/util"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// Handler preprocess scan artifact event
type Handler struct {
}

// Handle preprocess chart event data and then publish hook event
func (si *Handler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("empty scan artifact event")
	}

	e, ok := value.(*event.ScanImageEvent)
	if !ok {
		return errors.New("invalid scan artifact event type")
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(e.Artifact.NamespaceID, e.EventType)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	// If we cannot find policy including event type in project, return directly
	if len(policies) == 0 {
		log.Debugf("Cannot find policy for %s event: %v", e.EventType, e)
		return nil
	}

	// Get project
	prj, err := project.Ctl.Get(orm.Context(), e.Artifact.NamespaceID, project.Metadata(true))
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	payload, err := constructScanImagePayload(e, prj)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	err = util.SendHookWithPolicies(policies, payload, e.EventType)
	if err != nil {
		return errors.Wrap(err, "scan preprocess handler")
	}

	return nil
}

// IsStateful ...
func (si *Handler) IsStateful() bool {
	return false
}

func constructScanImagePayload(event *event.ScanImageEvent, project *models.Project) (*model.Payload, error) {
	repoType := models.ProjectPrivate
	if project.IsPublic() {
		repoType = models.ProjectPublic
	}

	repoName := util.GetNameFromImgRepoFullName(event.Artifact.Repository)

	payload := &model.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &model.EventData{
			Repository: &model.Repository{
				Name:         repoName,
				Namespace:    project.Name,
				RepoFullName: event.Artifact.Repository,
				RepoType:     repoType,
			},
		},
		Operator: event.Operator,
	}

	resURL, err := util.BuildImageResourceURL(event.Artifact.Repository, event.Artifact.Tag)
	if err != nil {
		return nil, errors.Wrap(err, "construct scan payload")
	}

	ctx := orm.NewContext(context.TODO(), o.NewOrm())

	reference := event.Artifact.Digest
	if reference == "" {
		reference = event.Artifact.Tag
	}

	art, err := artifact.Ctl.GetByReference(ctx, event.Artifact.Repository, event.Artifact.Digest, nil)
	if err != nil {
		return nil, err
	}
	// Wait for reasonable time to make sure the report is ready
	// Interval=500ms and total time = 5s
	// If the report is still not ready in the total time, then failed at then
	for i := 0; i < 10; i++ {
		// First check in case it is ready
		if re, err := scan.DefaultController.GetReport(ctx, art, []string{v1.MimeTypeNativeReport}); err == nil {
			if len(re) > 0 && len(re[0].Report) > 0 {
				break
			}
		} else {
			log.Error(errors.Wrap(err, "construct scan payload: wait report ready loop"))
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Add scan overview
	summaries, err := scan.DefaultController.GetSummary(ctx, art, []string{v1.MimeTypeNativeReport})
	if err != nil {
		return nil, errors.Wrap(err, "construct scan payload")
	}

	resource := &model.Resource{
		Tag:          event.Artifact.Tag,
		Digest:       event.Artifact.Digest,
		ResourceURL:  resURL,
		ScanOverview: summaries,
	}
	payload.EventData.Resources = append(payload.EventData.Resources, resource)

	return payload, nil
}
