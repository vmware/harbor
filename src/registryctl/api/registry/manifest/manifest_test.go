package manifest

import (
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/storage/driver/inmemory"
	"github.com/docker/distribution/testutil"
	"github.com/goharbor/harbor/src/registryctl/api/registry/test"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteManifest(t *testing.T) {
	ctx := context.Background()
	inmemoryDriver := inmemory.New()

	registry := test.CreateRegistry(t, inmemoryDriver)
	repo := test.MakeRepository(t, registry, "mftest")

	// Create random layers
	randomLayers, err := testutil.CreateRandomLayers(3)
	if err != nil {
		t.Fatalf("failed to make layers: %v", err)
	}

	// Upload all layers
	err = testutil.UploadBlobs(repo, randomLayers)
	if err != nil {
		t.Fatalf("failed to upload layers: %v", err)
	}

	sharedKey := test.GetAnyKey(randomLayers)
	manifest, err := testutil.MakeSchema2Manifest(repo, append(test.GetKeys(randomLayers), sharedKey))
	if err != nil {
		t.Fatalf("failed to make manifest: %v", err)
	}

	manifestService := test.MakeManifestService(t, repo)
	_, err = manifestService.Put(ctx, manifest)
	if err != nil {
		t.Fatalf("manifest upload failed: %v", err)
	}

	manifestDigest, err := manifestService.Put(ctx, manifest)
	if err != nil {
		t.Fatalf("manifest upload failed: %v", err)
	}

	req, err := http.NewRequest(http.MethodDelete, "http://api/registry/{name}/manifests/{reference}/?tags=1,2,3", nil)
	varMap := make(map[string]string, 1)
	varMap["reference"] = manifestDigest.String()
	varMap["name"] = fmt.Sprintf("%v", repo.Named())
	req = mux.SetURLVars(req, varMap)

	manifestHandler := NewHandler(inmemoryDriver)
	rec := httptest.NewRecorder()
	manifestHandler.ServeHTTP(rec, req)
	assert.True(t, rec.Result().StatusCode == 200)

	// check that all of the layers of manifest are deleted.
	blobs := test.AllBlobs(t, registry)
	for dgst := range randomLayers {
		if _, ok := blobs[dgst]; !ok {
			t.Fatalf("random layer blob missing: %v", dgst)
		}
	}
}
