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

package artifact

import (
	"testing"

	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/blob"
	"github.com/goharbor/harbor/src/testing/mock"
	tart "github.com/goharbor/harbor/src/testing/pkg/artifact"
	tblob "github.com/goharbor/harbor/src/testing/pkg/blob"
	tpro "github.com/goharbor/harbor/src/testing/pkg/processor"
	"github.com/goharbor/harbor/src/testing/pkg/registry"

	"github.com/distribution/distribution"
	"github.com/distribution/distribution/manifest/schema1"
	"github.com/distribution/distribution/manifest/schema2"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
)

var (
	v1Manifest = `{
   "schemaVersion": 1,
   "name": "library/node",
   "tag": "5.5-onbuild",
   "architecture": "amd64",
   "fsLayers": [
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:e2f0af7be4d7ec1946e55d4edddf90f768fd622573b8f1f0a19fa3a087b11936"
      },
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:bb0313a4938416446d43fb6fc25c73d4b495575ae0b537ad2ffa0bb081a99916"
      },
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:9e9f27c613944beb01ac418fef42a04eb021787a0eef0126b2c73604a57a1384"
      },
      {
         "blobSum": "sha256:7a0c192d4d2536499ef0c65fa1c60e27ad39b4c4dcb9c703114bb8dc67f8fa5c"
      },
      {
         "blobSum": "sha256:6ecee6444751349ab3731ee4e10f40b93e98af06a70349ca66962b2c80c5cce2"
      },
      {
         "blobSum": "sha256:9269ba3950bb316abe52dc7010b0758b760e887a0d41af177162a55b2722bab7"
      },
      {
         "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
      },
      {
         "blobSum": "sha256:03e1855d4f316edea9545408dcac38be93e9ea6aba6e85610edf76db7ccbbfa7"
      }
   ],
   "history": [
      {
         "v1Compatibility": "{\"id\":\"1520dbfa834708e58189bd7ad3ddfe5251fbdab020d274e2f2934b193fedce3e\",\"parent\":\"892e1bee0938dd0f1e6cbe4fda1f9d8efb8529c1c7a6469302b2a616541d5c74\",\"created\":\"2016-01-26T16:54:31.506284103Z\",\"container\":\"57db0fd5498375241dfce628a92a28df825f6c6a185119032760f79802477074\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) CMD [\\\"npm\\\" \\\"start\\\"]\"],\"Image\":\"892e1bee0938dd0f1e6cbe4fda1f9d8efb8529c1c7a6469302b2a616541d5c74\",\"Volumes\":null,\"WorkingDir\":\"/usr/src/app\",\"Entrypoint\":null,\"OnBuild\":[\"COPY package.json /usr/src/app/\",\"RUN npm install\",\"COPY . /usr/src/app\"],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"npm\",\"start\"],\"Image\":\"892e1bee0938dd0f1e6cbe4fda1f9d8efb8529c1c7a6469302b2a616541d5c74\",\"Volumes\":null,\"WorkingDir\":\"/usr/src/app\",\"Entrypoint\":null,\"OnBuild\":[\"COPY package.json /usr/src/app/\",\"RUN npm install\",\"COPY . /usr/src/app\"],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"892e1bee0938dd0f1e6cbe4fda1f9d8efb8529c1c7a6469302b2a616541d5c74\",\"parent\":\"c138a9cd4a0adb6c81597e39bcd0dba2d2c181b2ca9a1a6c521cfbd159d90d2a\",\"created\":\"2016-01-26T16:54:30.676634208Z\",\"container\":\"772e639526dff6564ea3922abbc05d63604be9fd1f068ab53684df7a949067be\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) ONBUILD COPY . /usr/src/app\"],\"Image\":\"c138a9cd4a0adb6c81597e39bcd0dba2d2c181b2ca9a1a6c521cfbd159d90d2a\",\"Volumes\":null,\"WorkingDir\":\"/usr/src/app\",\"Entrypoint\":null,\"OnBuild\":[\"COPY package.json /usr/src/app/\",\"RUN npm install\",\"COPY . /usr/src/app\"],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"node\"],\"Image\":\"c138a9cd4a0adb6c81597e39bcd0dba2d2c181b2ca9a1a6c521cfbd159d90d2a\",\"Volumes\":null,\"WorkingDir\":\"/usr/src/app\",\"Entrypoint\":null,\"OnBuild\":[\"COPY package.json /usr/src/app/\",\"RUN npm install\",\"COPY . /usr/src/app\"],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"c138a9cd4a0adb6c81597e39bcd0dba2d2c181b2ca9a1a6c521cfbd159d90d2a\",\"parent\":\"3484e461ee7551398ff2a5fe7d29a7d8f7c13830f0f7629bd6e0f4d7853f3686\",\"created\":\"2016-01-26T16:54:30.007571536Z\",\"container\":\"2b3d627f133121fb0bdd6e656eb4efa4444a8af832760c59d7942b4b59e3ea18\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) ONBUILD RUN npm install\"],\"Image\":\"3484e461ee7551398ff2a5fe7d29a7d8f7c13830f0f7629bd6e0f4d7853f3686\",\"Volumes\":null,\"WorkingDir\":\"/usr/src/app\",\"Entrypoint\":null,\"OnBuild\":[\"COPY package.json /usr/src/app/\",\"RUN npm install\"],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"node\"],\"Image\":\"3484e461ee7551398ff2a5fe7d29a7d8f7c13830f0f7629bd6e0f4d7853f3686\",\"Volumes\":null,\"WorkingDir\":\"/usr/src/app\",\"Entrypoint\":null,\"OnBuild\":[\"COPY package.json /usr/src/app/\",\"RUN npm install\"],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"3484e461ee7551398ff2a5fe7d29a7d8f7c13830f0f7629bd6e0f4d7853f3686\",\"parent\":\"f8858c27980847a58f59954e09a2a9588ca56f878f8a9d9e9ca7f908a0d4424a\",\"created\":\"2016-01-26T16:54:29.347805328Z\",\"container\":\"c11f79f8aa2a23230f9d44618e5a50230f30a585cfc81cc089ce4848ef1ea97b\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) ONBUILD COPY package.json /usr/src/app/\"],\"Image\":\"f8858c27980847a58f59954e09a2a9588ca56f878f8a9d9e9ca7f908a0d4424a\",\"Volumes\":null,\"WorkingDir\":\"/usr/src/app\",\"Entrypoint\":null,\"OnBuild\":[\"COPY package.json /usr/src/app/\"],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"node\"],\"Image\":\"f8858c27980847a58f59954e09a2a9588ca56f878f8a9d9e9ca7f908a0d4424a\",\"Volumes\":null,\"WorkingDir\":\"/usr/src/app\",\"Entrypoint\":null,\"OnBuild\":[\"COPY package.json /usr/src/app/\"],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"f8858c27980847a58f59954e09a2a9588ca56f878f8a9d9e9ca7f908a0d4424a\",\"parent\":\"f64ab7978e6acb948555df37760590f64ea93ad1d2c23fce5a5658266d24d432\",\"created\":\"2016-01-26T16:54:28.738290402Z\",\"container\":\"a408ee4aa153f75028302c2459f75e96f20da4b95f4cce7c50252f617c4fc215\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) WORKDIR /usr/src/app\"],\"Image\":\"f64ab7978e6acb948555df37760590f64ea93ad1d2c23fce5a5658266d24d432\",\"Volumes\":null,\"WorkingDir\":\"/usr/src/app\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"node\"],\"Image\":\"f64ab7978e6acb948555df37760590f64ea93ad1d2c23fce5a5658266d24d432\",\"Volumes\":null,\"WorkingDir\":\"/usr/src/app\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"f64ab7978e6acb948555df37760590f64ea93ad1d2c23fce5a5658266d24d432\",\"parent\":\"5f8d821c760f574dd96974b4d70bc442f79b9f56ffe23530d2523eef026b152f\",\"created\":\"2016-01-26T16:54:28.066325756Z\",\"container\":\"ec647c177e1391cfa1cd60fa0a58147af0cb4b0b932a4c42b0932cf04444d76c\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"mkdir -p /usr/src/app\"],\"Image\":\"5f8d821c760f574dd96974b4d70bc442f79b9f56ffe23530d2523eef026b152f\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"node\"],\"Image\":\"5f8d821c760f574dd96974b4d70bc442f79b9f56ffe23530d2523eef026b152f\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"5f8d821c760f574dd96974b4d70bc442f79b9f56ffe23530d2523eef026b152f\",\"parent\":\"ebcf22a55440806b2cf690333dc39f6321deda73a34a790468f3ef58e459eeb6\",\"created\":\"2016-01-26T16:52:50.89027915Z\",\"container\":\"db127a54c3e12c478d144a0e480ee7f4ac3909d1440e9372bf54be82507ec5d7\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) CMD [\\\"node\\\"]\"],\"Image\":\"ebcf22a55440806b2cf690333dc39f6321deda73a34a790468f3ef58e459eeb6\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"node\"],\"Image\":\"ebcf22a55440806b2cf690333dc39f6321deda73a34a790468f3ef58e459eeb6\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"ebcf22a55440806b2cf690333dc39f6321deda73a34a790468f3ef58e459eeb6\",\"parent\":\"20ed370cdb6e6e36c94a673270155b7edc669e583f870e48833125f828e89e65\",\"created\":\"2016-01-26T16:52:45.83954478Z\",\"container\":\"5cbb1dc61fe39c8f13e41bb72956f8764c828f79c18d5dfcc6b1ff111e88b997\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"curl -SLO \\\"https://nodejs.org/dist/v$NODE_VERSION/node-v$NODE_VERSION-linux-x64.tar.gz\\\"   \\u0026\\u0026 curl -SLO \\\"https://nodejs.org/dist/v$NODE_VERSION/SHASUMS256.txt.asc\\\"   \\u0026\\u0026 gpg --verify SHASUMS256.txt.asc   \\u0026\\u0026 grep \\\" node-v$NODE_VERSION-linux-x64.tar.gz\\\\$\\\" SHASUMS256.txt.asc | sha256sum -c -   \\u0026\\u0026 tar -xzf \\\"node-v$NODE_VERSION-linux-x64.tar.gz\\\" -C /usr/local --strip-components=1   \\u0026\\u0026 rm \\\"node-v$NODE_VERSION-linux-x64.tar.gz\\\" SHASUMS256.txt.asc\"],\"Image\":\"20ed370cdb6e6e36c94a673270155b7edc669e583f870e48833125f828e89e65\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/bash\"],\"Image\":\"20ed370cdb6e6e36c94a673270155b7edc669e583f870e48833125f828e89e65\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":36385291}"
      },
      {
         "v1Compatibility": "{\"id\":\"20ed370cdb6e6e36c94a673270155b7edc669e583f870e48833125f828e89e65\",\"parent\":\"8ab6f3fcbdb58860b302cb53d3c090d1e4a56cc5a4ac548724be143f292bbd08\",\"created\":\"2016-01-26T16:52:38.484938978Z\",\"container\":\"db88a6dff38e4b3fc8798f5fbe7cff5a9c3a1fa2627fcb7601e627a18ef1359d\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) ENV NODE_VERSION=5.5.0\"],\"Image\":\"8ab6f3fcbdb58860b302cb53d3c090d1e4a56cc5a4ac548724be143f292bbd08\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\",\"NODE_VERSION=5.5.0\"],\"Cmd\":[\"/bin/bash\"],\"Image\":\"8ab6f3fcbdb58860b302cb53d3c090d1e4a56cc5a4ac548724be143f292bbd08\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"8ab6f3fcbdb58860b302cb53d3c090d1e4a56cc5a4ac548724be143f292bbd08\",\"parent\":\"ddfb2360ce1e908d5ecb4b678ee10686ba28a4fdcc68d70177d8fbafcaf2da24\",\"created\":\"2016-01-26T16:44:56.182612087Z\",\"container\":\"c510a06b1df9a9989f02337b1c0cbe0c549381e8d6d2f8f6e660328944c0186e\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) ENV NPM_CONFIG_LOGLEVEL=info\"],\"Image\":\"ddfb2360ce1e908d5ecb4b678ee10686ba28a4fdcc68d70177d8fbafcaf2da24\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"NPM_CONFIG_LOGLEVEL=info\"],\"Cmd\":[\"/bin/bash\"],\"Image\":\"ddfb2360ce1e908d5ecb4b678ee10686ba28a4fdcc68d70177d8fbafcaf2da24\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"ddfb2360ce1e908d5ecb4b678ee10686ba28a4fdcc68d70177d8fbafcaf2da24\",\"parent\":\"9536cbaf1242bbc772d382c828ad4c8d317fcd63ef9cde05f9cb4cd4b6871236\",\"created\":\"2016-01-26T16:38:21.781529683Z\",\"container\":\"f2ef38095c6c3a99e64a14d3d433f0a7ed7ecef2264afe40ccbeb93b38bc77d9\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"set -ex   \\u0026\\u0026 for key in     9554F04D7259F04124DE6B476D5A82AC7E37093B     94AE36675C464D64BAFA68DD7434390BDBE9B9C5     0034A06D9D9B0064CE8ADF6BF1747F4AD2306D93     FD3A5288F042B6850C66B31F09FE44734EB7990E     71DCFD284A79C3B38668286BC97EC7A07EDE3FC1     DD8F2338BAE7501E3DD5AC78C273792F7D83545D     B9AE9905FFD7803F25714661B63B535A4C206CA9     C4F0DFFF4E8C1A8236409D08E73BC641CC11F4C8   ; do     gpg --keyserver ha.pool.sks-keyservers.net --recv-keys \\\"$key\\\";   done\"],\"Image\":\"9536cbaf1242bbc772d382c828ad4c8d317fcd63ef9cde05f9cb4cd4b6871236\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/bash\"],\"Image\":\"9536cbaf1242bbc772d382c828ad4c8d317fcd63ef9cde05f9cb4cd4b6871236\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":51753}"
      },
      {
         "v1Compatibility": "{\"id\":\"9536cbaf1242bbc772d382c828ad4c8d317fcd63ef9cde05f9cb4cd4b6871236\",\"parent\":\"0288ae931294ce04f5d69c60146faca7d9be8de4004421d650f4227fa60bd92b\",\"created\":\"2016-01-25T22:31:08.823570982Z\",\"container\":\"528b705b36c3a1ae37343eec7824283170b0bffe8b40f16f830eab723ac2f08d\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"apt-get update \\u0026\\u0026 apt-get install -y --no-install-recommends \\t\\tautoconf \\t\\tautomake \\t\\tbzip2 \\t\\tfile \\t\\tg++ \\t\\tgcc \\t\\timagemagick \\t\\tlibbz2-dev \\t\\tlibc6-dev \\t\\tlibcurl4-openssl-dev \\t\\tlibevent-dev \\t\\tlibffi-dev \\t\\tlibgeoip-dev \\t\\tlibglib2.0-dev \\t\\tlibjpeg-dev \\t\\tliblzma-dev \\t\\tlibmagickcore-dev \\t\\tlibmagickwand-dev \\t\\tlibmysqlclient-dev \\t\\tlibncurses-dev \\t\\tlibpng-dev \\t\\tlibpq-dev \\t\\tlibreadline-dev \\t\\tlibsqlite3-dev \\t\\tlibssl-dev \\t\\tlibtool \\t\\tlibwebp-dev \\t\\tlibxml2-dev \\t\\tlibxslt-dev \\t\\tlibyaml-dev \\t\\tmake \\t\\tpatch \\t\\txz-utils \\t\\tzlib1g-dev \\t\\u0026\\u0026 rm -rf /var/lib/apt/lists/*\"],\"Image\":\"0288ae931294ce04f5d69c60146faca7d9be8de4004421d650f4227fa60bd92b\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/bash\"],\"Image\":\"0288ae931294ce04f5d69c60146faca7d9be8de4004421d650f4227fa60bd92b\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":314656819}"
      },
      {
         "v1Compatibility": "{\"id\":\"0288ae931294ce04f5d69c60146faca7d9be8de4004421d650f4227fa60bd92b\",\"parent\":\"9287fae7a16e8788603ae069270aa825457065062247f4c04d4983f00eba37a6\",\"created\":\"2016-01-25T22:29:12.503492968Z\",\"container\":\"a0533596d15ff539859472684f7e700042f357d02fa0c1fb6c5d8a1feac6c574\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"apt-get update \\u0026\\u0026 apt-get install -y --no-install-recommends \\t\\tbzr \\t\\tgit \\t\\tmercurial \\t\\topenssh-client \\t\\tsubversion \\t\\t\\t\\tprocps \\t\\u0026\\u0026 rm -rf /var/lib/apt/lists/*\"],\"Image\":\"9287fae7a16e8788603ae069270aa825457065062247f4c04d4983f00eba37a6\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/bash\"],\"Image\":\"9287fae7a16e8788603ae069270aa825457065062247f4c04d4983f00eba37a6\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":122576525}"
      },
      {
         "v1Compatibility": "{\"id\":\"9287fae7a16e8788603ae069270aa825457065062247f4c04d4983f00eba37a6\",\"parent\":\"5eb1402f041415f4d72ec331c9388e4981420dfe88ef4e9bdf904d4687e4de09\",\"created\":\"2016-01-25T22:28:10.88750042Z\",\"container\":\"ce5ccec57f456f36a78b32dad3a696a215ff0201270d47ee1c2f64a52508297a\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"apt-get update \\u0026\\u0026 apt-get install -y --no-install-recommends \\t\\tca-certificates \\t\\tcurl \\t\\twget \\t\\u0026\\u0026 rm -rf /var/lib/apt/lists/*\"],\"Image\":\"5eb1402f041415f4d72ec331c9388e4981420dfe88ef4e9bdf904d4687e4de09\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/bash\"],\"Image\":\"5eb1402f041415f4d72ec331c9388e4981420dfe88ef4e9bdf904d4687e4de09\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":[],\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":44300304}"
      },
      {
         "v1Compatibility": "{\"id\":\"5eb1402f041415f4d72ec331c9388e4981420dfe88ef4e9bdf904d4687e4de09\",\"parent\":\"77e39ee8211729e81d1f83f0c64fdef97979b930a97ddc8194b8ea46d49f7b50\",\"created\":\"2016-01-25T22:24:37.914712562Z\",\"container\":\"c59024072143b04b79ac341c51571fc698636e01c13b49c523309c84af4b70fe\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) CMD [\\\"/bin/bash\\\"]\"],\"Image\":\"77e39ee8211729e81d1f83f0c64fdef97979b930a97ddc8194b8ea46d49f7b50\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":[\"/bin/bash\"],\"Image\":\"77e39ee8211729e81d1f83f0c64fdef97979b930a97ddc8194b8ea46d49f7b50\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\"}"
      },
      {
         "v1Compatibility": "{\"id\":\"77e39ee8211729e81d1f83f0c64fdef97979b930a97ddc8194b8ea46d49f7b50\",\"created\":\"2016-01-25T22:24:35.279128653Z\",\"container\":\"e06f5a03fe1f6755f98fb354799db823a95e6c141ae40a2cb7ad7a6b09d41208\",\"container_config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) ADD file:e5a3d20748c5d3dd5fa11542dfa4ef8b72a0bb78ce09f6dae30eff5d045c67aa in /\"],\"Image\":\"\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":null},\"docker_version\":\"1.8.3\",\"config\":{\"Hostname\":\"e06f5a03fe1f\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":null,\"Cmd\":null,\"Image\":\"\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":null},\"architecture\":\"amd64\",\"os\":\"linux\",\"Size\":125082947}"
      }
   ],
   "signatures": [
      {
         "header": {
            "jwk": {
               "crv": "P-256",
               "kid": "KG7S:QIPL:FTS3:YAKZ:AADA:4GML:ITLH:7APP:O4F7:2NBA:A4IN:CWVF",
               "kty": "EC",
               "x": "PUiT1kV7Xf-U8M54gpCzvPc5mDUX9BjvizdBgy3oTsI",
               "y": "BAgeQchl9QibzPP2Qp_-gJMWr682QVWoHy52hLRHZ04"
            },
            "alg": "ES256"
         },
         "signature": "mfUOI0pPzkdceAKAFRMkrQVgeE9X7if43LEtfs5XvdyxO7lCG0fiVxmdi-KGaQu4lRsIfRNq6m5agNTm8u5DrA",
         "protected": "eyJmb3JtYXRMZW5ndGgiOjI3NTE4LCJmb3JtYXRUYWlsIjoiQ24wIiwidGltZSI6IjIwMjAtMDMtMTdUMTA6NTk6MDhaIn0"
      }
   ]
}`
	v2Manifest = `{
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
  ],
  "annotations": {
    "com.example.key1": "value1"
  }
}`

	index = `{
  "schemaVersion": 2,
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 7143,
      "digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f",
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "size": 7682,
      "digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    }
  ],
  "annotations": {
    "com.example.key1": "value1"
  }
}`
)

type abstractorTestSuite struct {
	suite.Suite
	argMgr     *tart.FakeManager
	blobMgr    *tblob.Manager
	regCli     *registry.FakeClient
	abstractor *abstractor
	processor  *tpro.Processor
}

func (a *abstractorTestSuite) SetupTest() {
	a.regCli = &registry.FakeClient{}
	a.argMgr = &tart.FakeManager{}
	a.blobMgr = &tblob.Manager{}
	a.abstractor = &abstractor{
		artMgr:  a.argMgr,
		blobMgr: a.blobMgr,
		regCli:  a.regCli,
	}
	a.processor = &tpro.Processor{}
	// clear all registered processors
	processor.Registry = map[string]processor.Processor{}
	processor.Registry[schema2.MediaTypeImageConfig] = a.processor
}

// docker manifest v1
func (a *abstractorTestSuite) TestAbstractMetadataOfV1Manifest() {
	manifest, _, err := distribution.UnmarshalManifest(schema1.MediaTypeSignedManifest, []byte(v1Manifest))
	a.Require().Nil(err)
	mock.OnAnything(a.blobMgr, "List").Return([]*blob.Blob{
		{Size: 10},
		{Size: 20},
	}, nil)
	a.regCli.On("PullManifest").Return(manifest, "", nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}
	err = a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().Nil(err)
	a.Assert().Equal(int64(1), artifact.ID)
	a.Assert().Equal(schema1.MediaTypeSignedManifest, artifact.ManifestMediaType)
	a.Assert().Equal(schema1.MediaTypeSignedManifest, artifact.MediaType)
	a.Assert().Equal(int64(30+len([]byte(v1Manifest))), artifact.Size)
}

// docker manifest v2
func (a *abstractorTestSuite) TestAbstractMetadataOfV2Manifest() {
	manifest, _, err := distribution.UnmarshalManifest(schema2.MediaTypeManifest, []byte(v2Manifest))
	a.Require().Nil(err)
	a.regCli.On("PullManifest").Return(manifest, "", nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}
	a.processor.On("AbstractMetadata", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().Nil(err)
	a.Assert().Equal(int64(1), artifact.ID)
	a.Assert().Equal(schema2.MediaTypeManifest, artifact.ManifestMediaType)
	a.Assert().Equal(schema2.MediaTypeImageConfig, artifact.MediaType)
	a.Assert().Equal(int64(3043), artifact.Size)
	a.Require().Len(artifact.Annotations, 1)
	a.Equal("value1", artifact.Annotations["com.example.key1"])
}

// OCI index
func (a *abstractorTestSuite) TestAbstractMetadataOfIndex() {
	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageIndex, []byte(index))
	a.Require().Nil(err)
	a.regCli.On("PullManifest").Return(manifest, "", nil)
	a.argMgr.On("GetByDigest").Return(&artifact.Artifact{
		ID:   2,
		Size: 10,
	}, nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}
	err = a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().Nil(err)
	a.Assert().Equal(int64(1), artifact.ID)
	a.Assert().Equal(v1.MediaTypeImageIndex, artifact.ManifestMediaType)
	a.Assert().Equal(v1.MediaTypeImageIndex, artifact.MediaType)
	a.Assert().Equal(int64(668), artifact.Size)
	a.Require().Len(artifact.Annotations, 1)
	a.Assert().Equal("value1", artifact.Annotations["com.example.key1"])
	a.Len(artifact.References, 2)
}

type unknownManifest struct{}

func (u *unknownManifest) References() []distribution.Descriptor {
	return nil
}
func (u *unknownManifest) Payload() (mediaType string, payload []byte, err error) {
	return "unknown-manifest", nil, nil
}

// unknown
func (a *abstractorTestSuite) TestAbstractMetadataOfUnsupported() {
	a.regCli.On("PullManifest").Return(&unknownManifest{}, "", nil)
	artifact := &artifact.Artifact{
		ID: 1,
	}
	err := a.abstractor.AbstractMetadata(nil, artifact)
	a.Require().NotNil(err)
}

func TestAbstractorTestSuite(t *testing.T) {
	suite.Run(t, &abstractorTestSuite{})
}
