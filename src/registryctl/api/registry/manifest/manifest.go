package manifest

import (
	"github.com/docker/distribution/registry/storage"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/registryctl/api"
	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	"net/http"
)

// NewHandler returns the handler to handler manifest request
func NewHandler(storageDriver storagedriver.StorageDriver) http.Handler {
	return &handler{
		storageDriver: storageDriver,
	}
}

type handler struct {
	storageDriver storagedriver.StorageDriver
}

// ServeHTTP ...
func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		h.delete(w, req)
	default:
		api.HandleNotMethodAllowed(w)
	}
}

// delete deletes manifest ...
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	ref := mux.Vars(r)["reference"]
	if ref == "" {
		api.HandleBadRequest(w, errors.New("no reference specified"))
		return
	}
	dgst, err := digest.Parse(ref)
	if err != nil {
		api.HandleBadRequest(w, errors.Wrap(err, "not supported reference"))
		return
	}
	repoName := mux.Vars(r)["name"]
	if repoName == "" {
		api.HandleBadRequest(w, errors.New("no repository name specified"))
		return
	}
	// let the tags as empty here, as it non-blocking GC. The tags deletion will be handled via DELETE /v2/manifest
	var tags []string
	cleaner := storage.NewVacuum(r.Context(), h.storageDriver)
	if err := cleaner.RemoveManifest(repoName, dgst, tags); err != nil {
		log.Infof("failed to remove manifest: %s, with error:%v", ref, err)
		api.HandleInternalServerError(w, err)
		return
	}
}
