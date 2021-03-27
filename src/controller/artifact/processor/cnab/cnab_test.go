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

package cnab

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/distribution/distribution"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
)

type processorTestSuite struct {
	suite.Suite
	processor *processor
	regCli    *registry.FakeClient
}

func (p *processorTestSuite) SetupTest() {
	p.regCli = &registry.FakeClient{}
	p.processor = &processor{
		manifestProcessor: &base.ManifestProcessor{
			RegCli: p.regCli,
		},
	}
	p.processor.IndexProcessor = &base.IndexProcessor{RegCli: p.regCli}
}

func (p *processorTestSuite) TestAbstractMetadata() {
	manifest := `{
  "schemaVersion": 2,
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "digest": "sha256:e91b9dfcbbb3b88bac94726f276b89de46e4460b55f6e6d6f876e666b150ec5b",
    "size": 498
  },
  "layers": null
}`
	config := `{
  "description": "A short description of your bundle",
  "invocationImages": [
    {
      "contentDigest": "sha256:a59a4e74d9cc89e4e75dfb2cc7ea5c108e4236ba6231b53081a9e2506d1197b6",
      "image": "cnab/helloworld:0.1.1",
      "imageType": "docker",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "size": 942
    }
  ],
  "keywords": [
    "helloworld",
    "cnab",
    "tutorial"
  ],
  "maintainers": [
    {
      "email": "jane.doe@example.com",
      "name": "Jane Doe",
      "url": "https://example.com"
    }
  ],
  "name": "helloworld",
  "schemaVersion": "v1.0.0",
  "version": "0.1.1"
}`
	art := &artifact.Artifact{
		References: []*artifact.Reference{
			{
				ChildDigest: "sha256:b9616da7500f8c7c9a5e8d915714cd02d11bcc71ff5b4fd190bb77b1355c8549",
				Annotations: map[string]string{
					"io.cnab.manifest.type": "config",
				},
			},
		},
	}
	mani, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manifest))
	p.Require().Nil(err)
	p.regCli.On("PullManifest").Return(mani, "", nil)
	p.regCli.On("PullBlob").Return(0, ioutil.NopCloser(strings.NewReader(config)), nil)
	err = p.processor.AbstractMetadata(nil, art, nil)
	p.Require().Nil(err)
	p.Len(art.ExtraAttrs, 7)
	p.Equal("0.1.1", art.ExtraAttrs["version"].(string))
	p.Equal("helloworld", art.ExtraAttrs["name"].(string))
}

func (p *processorTestSuite) TestGetArtifactType() {
	p.Assert().Equal(ArtifactTypeCNAB, p.processor.GetArtifactType(nil, nil))
}

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, &processorTestSuite{})
}
