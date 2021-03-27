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

	"github.com/distribution/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func init() {
	mediaTypes := []string{
		v1.MediaTypeImageIndex,
		manifestlist.MediaTypeManifestList,
	}
	pc := &indexProcessor{}
	pc.IndexProcessor = &base.IndexProcessor{}
	if err := processor.Register(pc, mediaTypes...); err != nil {
		log.Errorf("failed to register processor for media type %v: %v", mediaTypes, err)
		return
	}
}

// indexProcessor processes image with OCI index and docker manifest list
type indexProcessor struct {
	*base.IndexProcessor
}

func (i *indexProcessor) GetArtifactType(ctx context.Context, artifact *artifact.Artifact) string {
	return ArtifactTypeImage
}
