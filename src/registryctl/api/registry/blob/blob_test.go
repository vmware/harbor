package blob

import (
	"github.com/distribution/distribution/registry/storage/driver/inmemory"
	"github.com/distribution/distribution/testutil"
	"github.com/goharbor/harbor/src/registryctl/api/registry/test"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeletionBlob(t *testing.T) {
	inmemoryDriver := inmemory.New()

	registry := test.CreateRegistry(t, inmemoryDriver)
	repo := test.MakeRepository(t, registry, "blobdeletion")

	// Create random layers
	randomLayers1, err := testutil.CreateRandomLayers(3)
	if err != nil {
		t.Fatalf("failed to make layers: %v", err)
	}

	randomLayers2, err := testutil.CreateRandomLayers(3)
	if err != nil {
		t.Fatalf("failed to make layers: %v", err)
	}

	// Upload all layers
	err = testutil.UploadBlobs(repo, randomLayers1)
	if err != nil {
		t.Fatalf("failed to upload layers: %v", err)
	}

	err = testutil.UploadBlobs(repo, randomLayers2)
	if err != nil {
		t.Fatalf("failed to upload layers: %v", err)
	}

	req, err := http.NewRequest(http.MethodDelete, "", nil)
	varMap := make(map[string]string, 1)
	varMap["reference"] = test.GetKeys(randomLayers1)[0].String()
	req = mux.SetURLVars(req, varMap)

	blobHandler := NewHandler(inmemoryDriver)
	rec := httptest.NewRecorder()
	blobHandler.ServeHTTP(rec, req)
	assert.True(t, rec.Result().StatusCode == 200)

	// layer1 is deleted and layer2 is still there
	blobs := test.AllBlobs(t, registry)
	for dgst := range randomLayers1 {
		if _, ok := blobs[dgst]; !ok {
			t.Logf("random layer 1 blob missing is correct as it has been deleted: %v", dgst)
		}
	}
	for dgst := range randomLayers2 {
		if _, ok := blobs[dgst]; !ok {
			t.Fatalf("random layer 2 blob missing: %v", dgst)
		}
	}
}
