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

package metadata

import (
	"context"
	"github.com/goharbor/harbor/src/common/security"
	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"time"
)

// ArtifactLabeledMetadata is the metadata from which the artifact labeled event can be resolved
type ArtifactLabeledMetadata struct {
	Ctx        context.Context
	ArtifactID int64
	LabelID    int64
	Operator   string
}

// Resolve to the event from the metadata
func (al *ArtifactLabeledMetadata) Resolve(event *event.Event) error {
	data := &event2.ArtifactLabeledEvent{
		ArtifactID: al.ArtifactID,
		LabelID:    al.LabelID,
		OccurAt:    time.Now(),
	}
	ctx, exist := security.FromContext(al.Ctx)
	if exist {
		data.Operator = ctx.GetUsername()
	}
	event.Topic = event2.TopicArtifactLabeled
	event.Data = data
	return nil
}
