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
	"fmt"
	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/registry"
	"io"
)

// remoteInterface defines operations related to remote repository under proxy
type remoteInterface interface {
	// BlobReader create a reader for remote blob
	BlobReader(repo, dig string) (int64, io.ReadCloser, error)
	// Manifest get manifest by reference
	Manifest(repo string, ref string) (distribution.Manifest, error)
}

// remoteHelper defines operations related to remote repository under proxy
type remoteHelper struct {
	regID    int64
	registry adapter.ArtifactRegistry
}

// newRemoteHelper create a remoteHelper interface
func newRemoteHelper(regID int64) (remoteInterface, error) {
	r := &remoteHelper{regID: regID}
	if err := r.init(); err != nil {
		log.Errorf("failed to create remoteHelper error %v", err)
		return nil, err
	}
	return r, nil
}

func (r *remoteHelper) init() error {

	if r.registry != nil {
		return nil
	}
	reg, err := registry.NewDefaultManager().Get(r.regID)
	if err != nil {
		return err
	}
	if reg == nil {
		return fmt.Errorf("failed to get registry, registryID: %v", r.regID)
	}
	factory, err := adapter.GetFactory(reg.Type)
	if err != nil {
		return err
	}
	adp, err := factory.Create(reg)
	if err != nil {
		return err
	}
	r.registry = adp.(adapter.ArtifactRegistry)
	return nil
}

func (r *remoteHelper) BlobReader(repo, dig string) (int64, io.ReadCloser, error) {
	return r.registry.PullBlob(repo, dig)
}

func (r *remoteHelper) Manifest(repo string, ref string) (distribution.Manifest, error) {
	man, _, err := r.registry.PullManifest(repo, ref)
	return man, err
}
