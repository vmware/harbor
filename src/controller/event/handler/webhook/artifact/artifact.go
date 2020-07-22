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

package artifact

import (
	"context"
	"fmt"

	beegorm "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/handler/util"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	notifyModel "github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/pkg/repository"
)

// Handler preprocess artifact event data
type Handler struct {
}

// Handle preprocess artifact event data and then publish hook event
func (a *Handler) Handle(value interface{}) error {
	switch v := value.(type) {
	case *event.PushArtifactEvent:
		return a.handle(v.ArtifactEvent)
	case *event.PullArtifactEvent:
		return a.handle(v.ArtifactEvent)
	case *event.DeleteArtifactEvent:
		return a.handle(v.ArtifactEvent)
	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}
	return nil
}

// IsStateful ...
func (a *Handler) IsStateful() bool {
	return false
}

func (a *Handler) handle(event *event.ArtifactEvent) error {
	prj, err := project.Ctl.Get(orm.Context(), event.Artifact.ProjectID, project.Metadata(true))
	if err != nil {
		log.Errorf("failed to get project: %d, error: %v", event.Artifact.ProjectID, err)
		return err
	}

	policies, err := notification.PolicyMgr.GetRelatedPolices(prj.ProjectID, event.EventType)
	if err != nil {
		log.Errorf("failed to find policy for %s event: %v", event.EventType, err)
		return err
	}
	if len(policies) == 0 {
		log.Debugf("cannot find policy for %s event: %v", event.EventType, event)
		return nil
	}

	payload, err := a.constructArtifactPayload(event, prj)
	if err != nil {
		return err
	}

	err = util.SendHookWithPolicies(policies, payload, event.EventType)
	if err != nil {
		return err
	}
	return nil
}

func (a *Handler) constructArtifactPayload(event *event.ArtifactEvent, project *models.Project) (*model.Payload, error) {
	repoName := event.Repository
	if repoName == "" {
		return nil, fmt.Errorf("invalid %s event with empty repo name", event.EventType)
	}

	repoType := models.ProjectPrivate
	if project.IsPublic() {
		repoType = models.ProjectPublic
	}

	imageName := util.GetNameFromImgRepoFullName(repoName)

	payload := &notifyModel.Payload{
		Type:    event.EventType,
		OccurAt: event.OccurAt.Unix(),
		EventData: &notifyModel.EventData{
			Repository: &notifyModel.Repository{
				Name:         imageName,
				Namespace:    project.Name,
				RepoFullName: repoName,
				RepoType:     repoType,
			},
		},
		Operator: event.Operator,
	}

	ctx := orm.NewContext(context.Background(), beegorm.NewOrm())
	repoRecord, err := repository.Mgr.GetByName(ctx, repoName)
	if err != nil {
		log.Errorf("failed to get repository with name %s: %v", repoName, err)
	} else {
		// for the delete repository event, it cannot get the repo info here, just let the creation time be empty.
		payload.EventData.Repository.DateCreated = repoRecord.CreationTime.Unix()
	}

	var reference string
	if len(event.Tags) == 0 {
		reference = event.Artifact.Digest
	} else {
		reference = event.Tags[0]
	}
	resURL, err := util.BuildImageResourceURL(repoName, reference)
	if err != nil {
		log.Errorf("get resource URL failed: %v", err)
		return nil, err
	}

	resource := &notifyModel.Resource{
		Tag:         reference,
		Digest:      event.Artifact.Digest,
		ResourceURL: resURL,
	}
	payload.EventData.Resources = append(payload.EventData.Resources, resource)

	return payload, nil
}
