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
	"fmt"
	"github.com/distribution/distribution"
	"github.com/distribution/distribution/manifest/manifestlist"
	"github.com/distribution/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/lib"
	libCache "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"strings"
	"time"
)

const defaultHandler = "default"

// NewCacheHandlerRegistry ...
func NewCacheHandlerRegistry(local localInterface) map[string]ManifestCacheHandler {
	manListHandler := &ManifestListCache{
		local: local,
		cache: libCache.Default(),
	}
	manHandler := &ManifestCache{local}
	registry := map[string]ManifestCacheHandler{
		manifestlist.MediaTypeManifestList: manListHandler,
		v1.MediaTypeImageIndex:             manListHandler,
		schema2.MediaTypeManifest:          manHandler,
		defaultHandler:                     manHandler,
	}
	return registry
}

// ManifestCacheHandler define how to cache manifest content
type ManifestCacheHandler interface {
	// CacheContent - cache the content of the manifest
	CacheContent(ctx context.Context, remoteRepo string, man distribution.Manifest, art lib.ArtifactInfo, r RemoteInterface)
}

// ManifestListCache handle Manifest list type and index type
type ManifestListCache struct {
	cache libCache.Cache
	local localInterface
}

// CacheContent ...
func (m *ManifestListCache) CacheContent(ctx context.Context, remoteRepo string, man distribution.Manifest, art lib.ArtifactInfo, r RemoteInterface) {
	_, payload, err := man.Payload()
	if err != nil {
		log.Errorf("failed to get payload, error %v", err)
		return
	}
	key := getManifestListKey(art.Repository, art.Digest)
	log.Debugf("cache manifest list with key=cache:%v", key)
	err = m.cache.Save(key, payload, manifestListCacheInterval)
	if err != nil {
		log.Errorf("failed to cache payload, error %v", err)
	}
	err = m.push(ctx, art.Repository, getReference(art), man)
	if err != nil {
		log.Errorf("error when push manifest list to local :%v", err)
	}
}

// cacheTrimmedDigest - cache the change Trimmed Digest for controller.EnsureTag when digest is changed
func (m *ManifestListCache) cacheTrimmedDigest(ctx context.Context, newDig string) {
	if m.cache == nil {
		return
	}
	art := lib.GetArtifactInfo(ctx)
	key := TrimmedManifestlist + string(art.Digest)
	err := m.cache.Save(key, newDig)
	if err != nil {
		log.Warningf("failed to cache the trimmed manifest, err %v", err)
		return
	}
	log.Debugf("Saved key:%v, value:%v", key, newDig)

}

func (m *ManifestListCache) updateManifestList(ctx context.Context, repo string, manifest distribution.Manifest) (distribution.Manifest, error) {
	switch v := manifest.(type) {
	case *manifestlist.DeserializedManifestList:
		existMans := make([]manifestlist.ManifestDescriptor, 0)
		for _, ma := range v.Manifests {
			art := lib.ArtifactInfo{Repository: repo, Digest: string(ma.Digest)}
			a, err := m.local.GetManifest(ctx, art)
			if err != nil {
				return nil, err
			}
			if a != nil {
				existMans = append(existMans, ma)
			}
		}
		return manifestlist.FromDescriptors(existMans)
	}
	return nil, fmt.Errorf("current manifest list type is unknown, manifest type[%T], content [%+v]", manifest, manifest)
}

func (m *ManifestListCache) push(ctx context.Context, repo, reference string, man distribution.Manifest) error {
	// For manifest list, it might include some different manifest
	// it will wait and check for 30 mins, if all depend manifests are ready then push it
	// if time exceed, then push a updated manifest list which contains existing manifest
	var newMan distribution.Manifest
	var err error
	for n := 0; n < maxManifestListWait; n++ {
		log.Debugf("waiting for the manifest ready, repo %v, tag:%v", repo, reference)
		time.Sleep(sleepIntervalSec * time.Second)
		newMan, err = m.updateManifestList(ctx, repo, man)
		if err != nil {
			return err
		}
		if len(newMan.References()) == len(man.References()) {
			break
		}
	}
	if len(newMan.References()) == 0 {
		return errors.New("manifest list doesn't contain any pushed manifest")
	}
	_, pl, err := newMan.Payload()
	if err != nil {
		log.Errorf("failed to get payload, error %v", err)
		return err
	}
	log.Debugf("The manifest list payload: %v", string(pl))
	newDig := digest.FromBytes(pl)
	m.cacheTrimmedDigest(ctx, string(newDig))
	// Because the manifest list maybe updated, need to recheck if it is exist in local
	art := lib.ArtifactInfo{Repository: repo, Tag: reference}
	a, err := m.local.GetManifest(ctx, art)
	if err != nil {
		return err
	}
	if a != nil && a.Digest == string(newDig) {
		return nil
	}
	// when pushing with digest, should push to its actual digest
	if strings.HasPrefix(reference, "sha256:") {
		reference = string(newDig)
	}
	return m.local.PushManifest(repo, reference, newMan)
}

// ManifestCache default Manifest handler
type ManifestCache struct {
	local localInterface
}

// CacheContent ...
func (m *ManifestCache) CacheContent(ctx context.Context, remoteRepo string, man distribution.Manifest, art lib.ArtifactInfo, r RemoteInterface) {
	var waitBlobs []distribution.Descriptor
	for n := 0; n < maxManifestWait; n++ {
		time.Sleep(sleepIntervalSec * time.Second)
		waitBlobs = m.local.CheckDependencies(ctx, art.Repository, man)
		if len(waitBlobs) == 0 {
			break
		}
		log.Debugf("Current n=%v artifact: %v:%v", n, art.Repository, art.Tag)
	}
	if len(waitBlobs) > 0 {
		// docker client will skip to pull layers exist in local
		// these blobs are not exist in the proxy server
		// it will cause the manifest dependency check always fail
		// need to push these blobs before push manifest to avoid failure
		log.Debug("Waiting blobs not empty, push it to local repo directly")
		for _, desc := range waitBlobs {
			err := m.putBlobToLocal(remoteRepo, art.Repository, desc, r)
			if err != nil {
				log.Errorf("Failed to push blob to local repo, error: %v", err)
				return
			}
		}
	}
	err := m.local.PushManifest(art.Repository, getReference(art), man)
	if err != nil {
		log.Errorf("failed to push manifest, tag: %v, error %v", art.Tag, err)
	}
}

func (m *ManifestCache) putBlobToLocal(remoteRepo string, localRepo string, desc distribution.Descriptor, r RemoteInterface) error {
	log.Debugf("Put blob to local registry!, sourceRepo:%v, localRepo:%v, digest: %v", remoteRepo, localRepo, desc.Digest)
	_, bReader, err := r.BlobReader(remoteRepo, string(desc.Digest))
	if err != nil {
		log.Errorf("failed to create blob reader, error %v", err)
		return err
	}
	defer bReader.Close()
	err = m.local.PushBlob(localRepo, desc, bReader)
	return err
}
