package gc

import (
	"bytes"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/registryctl/api"
	"net/http"
	"os/exec"
	"time"
)

// NewHandler returns the handler to handler blob request
func NewHandler(registryConf string) http.Handler {
	return &handler{
		registryConf: registryConf,
	}
}

type handler struct {
	registryConf string
}

// ServeHTTP ...
func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		h.start(w, req)
	default:
		api.HandleNotMethodAllowed(w)
	}
}

// Result ...
type Result struct {
	Status    bool      `json:"status"`
	Msg       string    `json:"msg"`
	StartTime time.Time `json:"starttime"`
	EndTime   time.Time `json:"endtime"`
}

// start ...
func (h *handler) start(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("/bin/bash", "-c", "registry_DO_NOT_USE_GC garbage-collect --delete-untagged=false "+h.registryConf)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	start := time.Now()
	log.Debugf("Start to execute garbage collection...")
	if err := cmd.Run(); err != nil {
		log.Errorf("Fail to execute GC: %v, command err: %s", err, errBuf.String())
		api.HandleInternalServerError(w, err)
		return
	}

	gcr := Result{true, outBuf.String(), start, time.Now()}
	if err := api.WriteJSON(w, gcr); err != nil {
		log.Errorf("failed to write response: %v", err)
		return
	}
	log.Debugf("Successful to execute garbage collection...")
}
