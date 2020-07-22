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

package p2p

import (
	"context"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/artifact/processor/image"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/p2p/preheat"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

// Handler ...
type Handler struct {
	Context func() context.Context
}

// Handle ...
func (p *Handler) Handle(value interface{}) error {
	switch value.(type) {
	case *event.PushArtifactEvent:
		pushArtEvent, _ := value.(*event.PushArtifactEvent)
		return p.handlePushArtifact(pushArtEvent)
	case *event.ScanImageEvent:
		scanImageEvent, _ := value.(*event.ScanImageEvent)
		return p.handleImageScanned(scanImageEvent)
	case *event.ArtifactLabeledEvent:
		artifactLabeledEvent, _ := value.(*event.ArtifactLabeledEvent)
		return p.handleArtifactLabeled(artifactLabeledEvent)
	default:
		return errors.New("unsupported type")
	}
}

// IsStateful ...
func (p *Handler) IsStateful() bool {
	return false
}

func (p *Handler) handlePushArtifact(event *event.PushArtifactEvent) error {
	if event.Artifact.Type != image.ArtifactTypeImage {
		return nil
	}
	log.Debugf("preheat artifact event %s:%s", event.Artifact.RepositoryName, event.Artifact.Digest)

	art, err := artifact.Ctl.Get(p.Context(), event.Artifact.ID, &artifact.Option{
		WithTag:   true,
		WithLabel: true,
	})
	if err != nil {
		return err
	}
	_, err = preheat.Enf.PreheatArtifact(p.Context(), art)
	return err
}

func (p *Handler) handleImageScanned(event *event.ScanImageEvent) error {
	log.Debugf("preheat image scanned %s:%s", event.Artifact.Repository, event.Artifact.Tag)
	art, err := artifact.Ctl.GetByReference(p.Context(), event.Artifact.Repository, event.Artifact.Digest,
		&artifact.Option{
			WithTag:   true,
			WithLabel: true,
		})
	if err != nil {
		return err
	}
	_, err = preheat.Enf.PreheatArtifact(p.Context(), art)
	return err
}

func (p *Handler) handleArtifactLabeled(event *event.ArtifactLabeledEvent) error {
	art, err := artifact.Ctl.Get(p.Context(), event.ArtifactID, &artifact.Option{
		WithTag:   true,
		WithLabel: true,
	})
	if art.Type != image.ArtifactTypeImage {
		return nil
	}
	log.Debugf("preheat artifact labeled %s:%s", art.Artifact.RepositoryName, art.Artifact.Digest)

	_, err = preheat.Enf.PreheatArtifact(p.Context(), art)
	return err
}
