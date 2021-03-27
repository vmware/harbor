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

package processor

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/distribution/distribution/manifest/schema2"
	// annotation parsers will be registered
	"github.com/goharbor/harbor/src/controller/artifact/annotation"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// ArtifactTypeUnknown defines the type for the unknown artifacts
	ArtifactTypeUnknown = "UNKNOWN"
)

var (
	// DefaultProcessor is to process artifact which has no specific processor
	DefaultProcessor = &defaultProcessor{regCli: registry.Cli}

	artifactTypeRegExp = regexp.MustCompile(`^application/vnd\.[^.]*\.(.*)\.config\.[^.]*\+json$`)
)

// the default processor to process artifact
type defaultProcessor struct {
	regCli registry.Client
}

func (d *defaultProcessor) GetArtifactType(ctx context.Context, artifact *artifact.Artifact) string {
	// try to parse the type from the media type
	strs := artifactTypeRegExp.FindStringSubmatch(artifact.MediaType)
	if len(strs) == 2 {
		return strings.ToUpper(strs[1])
	}
	// can not get the artifact type from the media type, return unknown
	return ArtifactTypeUnknown
}
func (d *defaultProcessor) ListAdditionTypes(ctx context.Context, artifact *artifact.Artifact) []string {
	return nil
}

// The default processor will process user-defined artifact.
// AbstractMetadata will abstract data in a specific way.
// Annotation keys in artifact annotation will decide which content will be processed in artifact.
// Here is a manifest example:
// {
//   "schemaVersion": 2,
//   "config": {
//       "mediaType": "application/vnd.caicloud.model.config.v1alpha1+json",
//       "digest": "sha256:be948daf0e22f264ea70b713ea0db35050ae659c185706aa2fad74834455fe8c",
//       "size": 187,
//       "annotations": {
//           "io.goharbor.artifact.v1alpha1.skip-list": "metrics,git"
//       }
//   },
//   "layers": [
//       {
//           "mediaType": "image/png",
//           "digest": "sha256:d923b93eadde0af5c639a972710a4d919066aba5d0dfbf4b9385099f70272da0",
//           "size": 166015,
//           "annotations": {
//               "io.goharbor.artifact.v1alpha1.icon": ""
//           }
//       },
//       {
//           "mediaType": "application/tar+gzip",
//           "digest": "sha256:d923b93eadde0af5c639a972710a4d919066aba5d0dfbf4b9385099f70272da0",
//           "size": 166015
//       }
//   ]
// }
func (d *defaultProcessor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact, manifest []byte) error {
	if artifact.ManifestMediaType != v1.MediaTypeImageManifest && artifact.ManifestMediaType != schema2.MediaTypeManifest {
		return nil
	}
	// get manifest
	mani := &v1.Manifest{}
	if err := json.Unmarshal(manifest, mani); err != nil {
		return err
	}
	// get config layer
	_, blob, err := d.regCli.PullBlob(artifact.RepositoryName, mani.Config.Digest.String())
	if err != nil {
		return err
	}
	defer blob.Close()
	// parse metadata from config layer
	metadata := map[string]interface{}{}
	// Some artifact may not have empty config layer.
	if mani.Config.Size != 0 {
		if err := json.NewDecoder(blob).Decode(&metadata); err != nil {
			return err
		}
	}
	// Populate all metadata into the ExtraAttrs first.
	artifact.ExtraAttrs = metadata
	annotationParser := annotation.NewParser()
	err = annotationParser.Parse(ctx, artifact, manifest)
	if err != nil {
		log.Errorf("the annotation parser parse annotation for artifact error: %v", err)
	}

	return nil
}

func (d *defaultProcessor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*Addition, error) {
	// Addition not support for user-defined artifact yet.
	// It will be support in the future.
	// return error directly
	return nil, errors.New(nil).WithCode(errors.BadRequestCode).
		WithMessage("the processor for artifact %s not found, cannot get the addition", artifact.Type)
}
