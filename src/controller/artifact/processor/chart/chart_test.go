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

package chart

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/distribution/distribution"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	chartserver "github.com/goharbor/harbor/src/pkg/chart"
	"github.com/goharbor/harbor/src/testing/pkg/chart"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
	helm_chart "helm.sh/helm/v3/pkg/chart"
)

var (
	chartManifest = `{"schemaVersion":2,"config":{"mediaType":"application/vnd.cncf.helm.config.v1+json","digest":"sha256:76a59ebef39013bf7b57e411629b569a5175590024f31eeaaa577a0f8da9e523","size":528},"layers":[{"mediaType":"application/tar+gzip","digest":"sha256:0bd64cfb958b68c71b46597e22185a41e784dc96e04090bc7d2a480b704c3b65","size":12607}]}`
	chartYaml     = `{
  "name":"redis",
  "home": "http://redis.io/",
  "sources": [
    "https://github.com/bitnami/bitnami-docker-redis"
  ],
  "version": "3.2.5",
  "description": "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
  "keywords": [
    "redis",
    "keyvalue",
    "database"
  ],
  "maintainers": [
    {
      "name": "bitnami-bot",
      "email":"containers@bitnami.com"
    }
  ],
  "icon": "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
  "apiVersion": "v1",
  "appVersion": "4.0.9"
}`
)

type processorTestSuite struct {
	suite.Suite
	processor *processor
	regCli    *registry.FakeClient
	chartOptr *chart.FakeOpertaor
}

func (p *processorTestSuite) SetupTest() {
	p.regCli = &registry.FakeClient{}
	p.chartOptr = &chart.FakeOpertaor{}
	p.processor = &processor{
		chartOperator: p.chartOptr,
	}
	p.processor.ManifestProcessor = &base.ManifestProcessor{RegCli: p.regCli}
}

func (p *processorTestSuite) TestAbstractAddition() {
	// unknown addition
	_, err := p.processor.AbstractAddition(nil, nil, "unknown_addition")
	p.True(errors.IsErr(err, errors.BadRequestCode))

	chartDetails := &chartserver.VersionDetails{
		Dependencies: []*helm_chart.Dependency{
			{
				Name:       "harbor",
				Version:    "v1.10",
				Repository: "github.com/goharbor",
			},
		},
		Values: map[string]interface{}{
			"cluster.enable":                   true,
			"cluster.slaveCount":               1,
			"image.pullPolicy":                 "Always",
			"master.securityContext.runAsUser": 1001,
		},
		Files: map[string]string{
			"README.MD":   "This chart bootstraps a [Redis](https://github.com/bitnami/bitnami-docker-redis) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.",
			"VALUES.YAML": `image:\n  ## Bitnami MongoDB registry\n  ##\n  registry: docker.io\n  ## Bitnami MongoDB image name\n  ##\n  repository: bitnami/mongodb\n  ## Bitnami MongoDB image tag\n  ## ref: https://hub.docker.com/r/bitnami/mongodb/tags/\n`,
		},
	}

	artifact := &artifact.Artifact{}
	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(chartManifest))
	p.Require().Nil(err)
	p.regCli.On("PullManifest").Return(manifest, "", nil)
	p.regCli.On("PullBlob").Return(0, ioutil.NopCloser(strings.NewReader(chartYaml)), nil)
	p.chartOptr.On("GetDetails").Return(chartDetails, nil)

	// values.yaml
	addition, err := p.processor.AbstractAddition(nil, artifact, AdditionTypeValues)
	p.Require().Nil(err)
	p.Equal("text/plain; charset=utf-8", addition.ContentType)
	p.Equal(`image:\n  ## Bitnami MongoDB registry\n  ##\n  registry: docker.io\n  ## Bitnami MongoDB image name\n  ##\n  repository: bitnami/mongodb\n  ## Bitnami MongoDB image tag\n  ## ref: https://hub.docker.com/r/bitnami/mongodb/tags/\n`, string(addition.Content))

	// README.md
	addition, err = p.processor.AbstractAddition(nil, artifact, AdditionTypeReadme)
	p.Require().Nil(err)
	p.Equal("text/markdown; charset=utf-8", addition.ContentType)
	p.Equal(`This chart bootstraps a [Redis](https://github.com/bitnami/bitnami-docker-redis) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.`, string(addition.Content))

	// README.md
	addition, err = p.processor.AbstractAddition(nil, artifact, AdditionTypeDependencies)
	p.Require().Nil(err)
	p.Equal("application/json; charset=utf-8", addition.ContentType)
	p.Equal(`[{"name":"harbor","version":"v1.10","repository":"github.com/goharbor"}]`, string(addition.Content))
}

func (p *processorTestSuite) TestGetArtifactType() {
	p.Assert().Equal(ArtifactTypeChart, p.processor.GetArtifactType(nil, nil))
}

func (p *processorTestSuite) TestListAdditionTypes() {
	additions := p.processor.ListAdditionTypes(nil, nil)
	p.EqualValues([]string{AdditionTypeValues, AdditionTypeReadme, AdditionTypeDependencies}, additions)
}

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, &processorTestSuite{})
}
