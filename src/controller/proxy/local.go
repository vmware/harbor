//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxy

import (
	"context"
	"github.com/distribution/distribution"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/proxy/secret"
	"github.com/goharbor/harbor/src/pkg/registry"
	"io"
)

// TrimmedManifestlist - key prefix for trimmed manifest
const TrimmedManifestlist = "trimmedmanifestlist:"

// localInterface defines operations related to local repo under proxy mode
type localInterface interface {
	// BlobExist check if the blob exist in local repo
	BlobExist(ctx context.Context, art lib.ArtifactInfo) (bool, error)
	// GetManifest get the manifest info
	GetManifest(ctx context.Context, art lib.ArtifactInfo) (*artifact.Artifact, error)
	// PushBlob push blob to local repo
	PushBlob(localRepo string, desc distribution.Descriptor, bReader io.ReadCloser) error
	// PushManifest push manifest to local repo, ref can be digest or tag
	PushManifest(repo string, ref string, manifest distribution.Manifest) error
	// CheckDependencies check if the manifest's dependency is ready
	CheckDependencies(ctx context.Context, repo string, man distribution.Manifest) []distribution.Descriptor
	// DeleteManifest cleanup delete tag from local cache
	DeleteManifest(repo, ref string)
}

func (l *localHelper) GetManifest(ctx context.Context, art lib.ArtifactInfo) (*artifact.Artifact, error) {
	ref := getReference(art)
	a, err := l.artifactCtl.GetByReference(ctx, art.Repository, ref, nil)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	return a, nil
}

// localHelper defines operations related to local repo under proxy mode
type localHelper struct {
	registry    registry.Client
	artifactCtl artifactController
	cache       cache.Cache
}

type artifactController interface {
	GetByReference(ctx context.Context, repository, reference string, option *artifact.Option) (artifact *artifact.Artifact, err error)
}

// newLocalHelper create the localInterface
func newLocalHelper() localInterface {
	l := &localHelper{artifactCtl: artifact.Ctl}
	l.init()
	return l
}

func (l *localHelper) BlobExist(ctx context.Context, art lib.ArtifactInfo) (bool, error) {
	return l.registry.BlobExist(art.Repository, art.Digest)
}

func (l *localHelper) init() {
	if l.registry != nil {
		return
	}
	log.Debugf("core url:%s, local core url: %v", config.GetCoreURL(), config.LocalCoreURL())
	// the traffic is internal only
	registryURL := config.LocalCoreURL()
	l.registry = registry.NewClientWithAuthorizer(registryURL, secret.NewAuthorizer(), true)
	l.cache = cache.Default()
}

func (l *localHelper) PushBlob(localRepo string, desc distribution.Descriptor, bReader io.ReadCloser) error {
	log.Debugf("Put blob to local registry, localRepo:%v, digest: %v", localRepo, desc.Digest)
	ref := string(desc.Digest)
	artName := localRepo + ":" + ref
	// use inflight checker to avoid multiple requests to push blob to local in same time
	if !inflightChecker.addRequest(artName) {
		return nil
	}
	defer inflightChecker.removeRequest(artName)
	err := l.registry.PushBlob(localRepo, ref, desc.Size, bReader)
	return err
}

func (l *localHelper) PushManifest(repo string, ref string, manifest distribution.Manifest) error {
	// Make sure there is only one go routing to push current artName to local repo
	artName := repo + ":" + ref
	// use inflight checker to avoid multiple requests to push manifest to local in same time
	if !inflightChecker.addRequest(artName) {
		return nil
	}
	defer inflightChecker.removeRequest(artName)

	mediaType, payload, err := manifest.Payload()
	if err != nil {
		return err
	}
	_, err = l.registry.PushManifest(repo, ref, mediaType, payload)
	return err
}

// DeleteManifest cleanup delete tag from local repo
func (l *localHelper) DeleteManifest(repo, ref string) {
	log.Debugf("Remove tag from repo if it is exist, repo: %v ref: %v", repo, ref)
	if err := l.registry.DeleteManifest(repo, ref); err != nil {
		// sometimes user pull a non-exist image
		log.Warningf("failed to remove artifact, error %v", err)
	}
}

func (l *localHelper) CheckDependencies(ctx context.Context, repo string, man distribution.Manifest) []distribution.Descriptor {
	descriptors := man.References()
	waitDesc := make([]distribution.Descriptor, 0)
	for _, desc := range descriptors {
		log.Debugf("checking the blob dependency: %v", desc.Digest)
		art := lib.ArtifactInfo{Repository: repo, Digest: string(desc.Digest)}
		exist, err := l.BlobExist(ctx, art)
		if err != nil || !exist {
			log.Debugf("Check dependency failed!")
			waitDesc = append(waitDesc, desc)
		}
	}
	log.Debugf("Check dependency result %v", waitDesc)
	return waitDesc
}

// SendPullEvent send a pull image event
func SendPullEvent(a *artifact.Artifact, tag, operator string) {
	e := &metadata.PullArtifactEventMetadata{
		Artifact: &a.Artifact,
		Tag:      tag,
		Operator: operator,
	}
	event.BuildAndPublish(e)
}
