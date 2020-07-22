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

package internal

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
)

// Handler preprocess artifact event data
type Handler struct {
	Context func() context.Context
}

// Handle ...
func (a *Handler) Handle(value interface{}) error {
	switch v := value.(type) {
	case *event.PullArtifactEvent:
		return a.onPull(a.Context(), v.ArtifactEvent)
	case *event.PushArtifactEvent:
		return a.onPush(a.Context(), v.ArtifactEvent)
	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}
	return nil
}

// IsStateful ...
func (a *Handler) IsStateful() bool {
	return false
}

func (a *Handler) onPull(ctx context.Context, event *event.ArtifactEvent) error {
	go func() { a.updatePullTime(ctx, event) }()
	go func() { a.addPullCount(ctx, event) }()
	return nil
}

func (a *Handler) updatePullTime(ctx context.Context, event *event.ArtifactEvent) {
	var tagID int64
	if len(event.Tags) != 0 {
		tags, err := tag.Ctl.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"ArtifactID": event.Artifact.ID,
				"Name":       event.Tags[0],
			},
		}, nil)
		if err != nil {
			log.Infof("failed to list tags when to update pull time, %v", err)
		} else {
			tagID = tags[0].ID
		}
	}
	if err := artifact.Ctl.UpdatePullTime(ctx, event.Artifact.ID, tagID, time.Now()); err != nil {
		log.Debugf("failed to update pull time form artifact %d, %v", event.Artifact.ID, err)
	}
	return
}

func (a *Handler) addPullCount(ctx context.Context, event *event.ArtifactEvent) {
	if err := repository.Ctl.AddPullCount(ctx, event.Artifact.RepositoryID); err != nil {
		log.Debugf("failed to add pull count repository %d, %v", event.Artifact.RepositoryID, err)
	}
	return
}

func (a *Handler) onPush(ctx context.Context, event *event.ArtifactEvent) error {
	go func() {
		if err := autoScan(ctx, &artifact.Artifact{Artifact: *event.Artifact}); err != nil {
			log.Errorf("scan artifact %s@%s failed, error: %v", event.Artifact.RepositoryName, event.Artifact.Digest, err)
		}
	}()

	return nil
}
