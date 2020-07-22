package test

import (
	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/storage"
	"github.com/docker/distribution/registry/storage/driver"
	"github.com/docker/libtrust"
	"github.com/opencontainers/go-digest"
	"io"
	"testing"
)

// CreateRegistry ...
func CreateRegistry(t *testing.T, driver driver.StorageDriver, options ...storage.RegistryOption) distribution.Namespace {
	ctx := context.Background()
	k, err := libtrust.GenerateECP256PrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	options = append([]storage.RegistryOption{storage.EnableDelete, storage.Schema1SigningKey(k), storage.EnableSchema1}, options...)
	registry, err := storage.NewRegistry(ctx, driver, options...)
	if err != nil {
		t.Fatalf("Failed to construct namespace")
	}
	return registry
}

// MakeRepository ...
func MakeRepository(t *testing.T, registry distribution.Namespace, name string) distribution.Repository {
	ctx := context.Background()

	// Initialize a dummy repository
	named, err := reference.WithName(name)
	if err != nil {
		t.Fatalf("Failed to parse name %s:  %v", name, err)
	}

	repo, err := registry.Repository(ctx, named)
	if err != nil {
		t.Fatalf("Failed to construct repository: %v", err)
	}
	return repo
}

// AllBlobs ...
func AllBlobs(t *testing.T, registry distribution.Namespace) map[digest.Digest]struct{} {
	ctx := context.Background()
	blobService := registry.Blobs()
	allBlobsMap := make(map[digest.Digest]struct{})
	err := blobService.Enumerate(ctx, func(dgst digest.Digest) error {
		allBlobsMap[dgst] = struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("Error getting all blobs: %v", err)
	}
	return allBlobsMap
}

// GetAnyKey ...
func GetAnyKey(digests map[digest.Digest]io.ReadSeeker) (d digest.Digest) {
	for d = range digests {
		break
	}
	return
}

// GetAnyKeys ...
func GetKeys(digests map[digest.Digest]io.ReadSeeker) (ds []digest.Digest) {
	for d := range digests {
		ds = append(ds, d)
	}
	return
}

// MakeManifestService ...
func MakeManifestService(t *testing.T, repository distribution.Repository) distribution.ManifestService {
	ctx := context.Background()

	manifestService, err := repository.Manifests(ctx)
	if err != nil {
		t.Fatalf("Failed to construct manifest store: %v", err)
	}
	return manifestService
}
