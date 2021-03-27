package blob

import (
	"errors"
	"github.com/distribution/distribution/registry/storage"
	storagedriver "github.com/distribution/distribution/registry/storage/driver"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/registryctl/api"
	"github.com/gorilla/mux"
	"net/http"
)

// NewHandler returns the handler to handler blob request
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

// DeleteBlob ...
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	ref := mux.Vars(r)["reference"]
	if ref == "" {
		api.HandleBadRequest(w, errors.New("no reference specified"))
		return
	}
	// don't parse the reference here as RemoveBlob does.
	cleaner := storage.NewVacuum(r.Context(), h.storageDriver)
	if err := cleaner.RemoveBlob(ref); err != nil {
		log.Infof("failed to remove blob: %s, with error:%v", ref, err)
		api.HandleError(w, err)
		return
	}
}
