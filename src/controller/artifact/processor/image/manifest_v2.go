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

package image

import (
	"context"
	"encoding/json"

	"github.com/distribution/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

// const definitions
const (
	// ArtifactTypeImage is the artifact type for image
	ArtifactTypeImage        = "IMAGE"
	AdditionTypeBuildHistory = "BUILD_HISTORY"
)

func init() {
	pc := &manifestV2Processor{}
	pc.ManifestProcessor = base.NewManifestProcessor()
	mediaTypes := []string{
		v1.MediaTypeImageConfig,
		schema2.MediaTypeImageConfig,
	}
	if err := processor.Register(pc, mediaTypes...); err != nil {
		log.Errorf("failed to register processor for media type %v: %v", mediaTypes, err)
		return
	}
}

// manifestV2Processor processes image with OCI manifest and docker v2 manifest
type manifestV2Processor struct {
	*base.ManifestProcessor
}

func (m *manifestV2Processor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact, manifest []byte) error {
	config := &v1.Image{}
	if err := m.ManifestProcessor.UnmarshalConfig(ctx, artifact.RepositoryName, manifest, config); err != nil {
		return err
	}
	if artifact.ExtraAttrs == nil {
		artifact.ExtraAttrs = map[string]interface{}{}
	}
	artifact.ExtraAttrs["created"] = config.Created
	artifact.ExtraAttrs["architecture"] = config.Architecture
	artifact.ExtraAttrs["os"] = config.OS
	artifact.ExtraAttrs["config"] = config.Config
	// if the author is null, try to get it from labels:
	// https://docs.docker.com/engine/reference/builder/#maintainer-deprecated
	author := config.Author
	if len(author) == 0 && len(config.Config.Labels) > 0 {
		author = config.Config.Labels["maintainer"]
	}
	artifact.ExtraAttrs["author"] = author
	return nil
}

func (m *manifestV2Processor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*processor.Addition, error) {
	if addition != AdditionTypeBuildHistory {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("addition %s isn't supported for %s(manifest version 2)", addition, ArtifactTypeImage)
	}

	mani, _, err := m.RegCli.PullManifest(artifact.RepositoryName, artifact.Digest)
	if err != nil {
		return nil, err
	}
	_, content, err := mani.Payload()
	if err != nil {
		return nil, err
	}
	config := &v1.Image{}
	if err = m.ManifestProcessor.UnmarshalConfig(ctx, artifact.RepositoryName, content, config); err != nil {
		return nil, err
	}
	content, err = json.Marshal(config.History)
	if err != nil {
		return nil, err
	}
	return &processor.Addition{
		Content:     content,
		ContentType: "application/json; charset=utf-8",
	}, nil
}

func (m *manifestV2Processor) GetArtifactType(ctx context.Context, artifact *artifact.Artifact) string {
	return ArtifactTypeImage
}

func (m *manifestV2Processor) ListAdditionTypes(ctx context.Context, artifact *artifact.Artifact) []string {
	return []string{AdditionTypeBuildHistory}
}
