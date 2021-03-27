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
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/distribution/distribution"
	"github.com/distribution/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
	"github.com/stretchr/testify/suite"
)

var (
	manifest = `{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 1510,
      "digest": "sha256:fce289e99eb9bca977dae136fbe2a82b6b7d4c372474c9235adc1741675f587e"
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 977,
         "digest": "sha256:1b930d010525941c1d56ec53b97bd057a67ae1865eebf042686d2a2d18271ced"
      }
   ]
}`
	config = `{
  "architecture": "amd64",
  "config": {
    "Hostname": "",
    "Domainname": "",
    "User": "",
    "AttachStdin": false,
    "AttachStdout": false,
    "AttachStderr": false,
    "Tty": false,
    "OpenStdin": false,
    "StdinOnce": false,
    "Env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    ],
    "Cmd": [
      "/hello"
    ],
    "ArgsEscaped": true,
    "Image": "sha256:a6d1aaad8ca65655449a26146699fe9d61240071f6992975be7e720f1cd42440",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": {
		"maintainer": "tester@vmware.com"	
	}
  },
  "container": "8e2caa5a514bb6d8b4f2a2553e9067498d261a0fd83a96aeaaf303943dff6ff9",
  "container_config": {
    "Hostname": "8e2caa5a514b",
    "Domainname": "",
    "User": "",
    "AttachStdin": false,
    "AttachStdout": false,
    "AttachStderr": false,
    "Tty": false,
    "OpenStdin": false,
    "StdinOnce": false,
    "Env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    ],
    "Cmd": [
      "/bin/sh",
      "-c",
      "#(nop) ",
      "CMD [\"/hello\"]"
    ],
    "ArgsEscaped": true,
    "Image": "sha256:a6d1aaad8ca65655449a26146699fe9d61240071f6992975be7e720f1cd42440",
    "Volumes": null,
    "WorkingDir": "",
    "Entrypoint": null,
    "OnBuild": null,
    "Labels": {
    }
  },
  "created": "2019-01-01T01:29:27.650294696Z",
  "docker_version": "18.06.1-ce",
  "history": [
    {
      "created": "2019-01-01T01:29:27.416803627Z",
      "created_by": "/bin/sh -c #(nop) COPY file:f77490f70ce51da25bd21bfc30cb5e1a24b2b65eb37d4af0c327ddc24f0986a6 in / "
    },
    {
      "created": "2019-01-01T01:29:27.650294696Z",
      "created_by": "/bin/sh -c #(nop)  CMD [\"/hello\"]",
      "empty_layer": true
    }
  ],
  "os": "linux",
  "rootfs": {
    "type": "layers",
    "diff_ids": [
      "sha256:af0b15c8625bb1938f1d7b17081031f649fd14e6b233688eea3c5483994a66a3"
    ]
  }
}`
)

type manifestV2ProcessorTestSuite struct {
	suite.Suite
	processor *manifestV2Processor
	regCli    *registry.FakeClient
}

func (m *manifestV2ProcessorTestSuite) SetupTest() {
	m.regCli = &registry.FakeClient{}
	m.processor = &manifestV2Processor{}
	m.processor.ManifestProcessor = &base.ManifestProcessor{RegCli: m.regCli}
}

func (m *manifestV2ProcessorTestSuite) TestAbstractMetadata() {
	artifact := &artifact.Artifact{}
	m.regCli.On("PullBlob", mock.Anything, mock.Anything).Return(0, ioutil.NopCloser(bytes.NewReader([]byte(config))), nil)
	err := m.processor.AbstractMetadata(nil, artifact, []byte(manifest))
	m.Require().Nil(err)
	m.NotNil(artifact.ExtraAttrs["created"])
	m.Equal("amd64", artifact.ExtraAttrs["architecture"])
	m.Equal("linux", artifact.ExtraAttrs["os"])
	m.NotNil(artifact.ExtraAttrs["config"])
	m.Equal("tester@vmware.com", artifact.ExtraAttrs["author"])
	m.regCli.AssertExpectations(m.T())
}

func (m *manifestV2ProcessorTestSuite) TestAbstractAddition() {
	// unknown addition
	_, err := m.processor.AbstractAddition(nil, nil, "unknown_addition")
	m.True(errors.IsErr(err, errors.BadRequestCode))

	// build history
	artifact := &artifact.Artifact{}
	manifest, _, err := distribution.UnmarshalManifest(schema2.MediaTypeManifest, []byte(manifest))
	m.Require().Nil(err)
	m.regCli.On("PullManifest").Return(manifest, "", nil)
	m.regCli.On("PullBlob").Return(0, ioutil.NopCloser(strings.NewReader(config)), nil)
	addition, err := m.processor.AbstractAddition(nil, artifact, AdditionTypeBuildHistory)
	m.Require().Nil(err)
	m.Equal("application/json; charset=utf-8", addition.ContentType)
	m.Equal(`[{"created":"2019-01-01T01:29:27.416803627Z","created_by":"/bin/sh -c #(nop) COPY file:f77490f70ce51da25bd21bfc30cb5e1a24b2b65eb37d4af0c327ddc24f0986a6 in / "},{"created":"2019-01-01T01:29:27.650294696Z","created_by":"/bin/sh -c #(nop)  CMD [\"/hello\"]","empty_layer":true}]`, string(addition.Content))
}

func (m *manifestV2ProcessorTestSuite) TestGetArtifactType() {
	m.Assert().Equal(ArtifactTypeImage, m.processor.GetArtifactType(nil, nil))
}

func (m *manifestV2ProcessorTestSuite) TestListAdditionTypes() {
	additions := m.processor.ListAdditionTypes(nil, nil)
	m.EqualValues([]string{AdditionTypeBuildHistory}, additions)
}

func TestManifestV2ProcessorTestSuite(t *testing.T) {
	suite.Run(t, &manifestV2ProcessorTestSuite{})
}
