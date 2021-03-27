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

	"github.com/distribution/distribution/manifest/schema1"
	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

func init() {
	pc := &manifestV1Processor{}
	if err := processor.Register(pc, schema1.MediaTypeSignedManifest); err != nil {
		log.Errorf("failed to register processor for media type %s: %v", schema1.MediaTypeSignedManifest, err)
		return
	}
}

// manifestV1Processor processes image with docker v1 manifest
type manifestV1Processor struct {
}

func (m *manifestV1Processor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact, manifest []byte) error {
	mani := &schema1.Manifest{}
	if err := json.Unmarshal(manifest, mani); err != nil {
		return err
	}
	if artifact.ExtraAttrs == nil {
		artifact.ExtraAttrs = map[string]interface{}{}
	}
	artifact.ExtraAttrs["architecture"] = mani.Architecture
	return nil
}

func (m *manifestV1Processor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*processor.Addition, error) {
	return nil, errors.New(nil).WithCode(errors.BadRequestCode).
		WithMessage("addition %s isn't supported for %s(manifest version 1)", addition, ArtifactTypeImage)
}

func (m *manifestV1Processor) GetArtifactType(ctx context.Context, artifact *artifact.Artifact) string {
	return ArtifactTypeImage
}

func (m *manifestV1Processor) ListAdditionTypes(ctx context.Context, artifact *artifact.Artifact) []string {
	return nil
}
