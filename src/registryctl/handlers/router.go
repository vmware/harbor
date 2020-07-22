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

package handlers

import (
	"github.com/goharbor/harbor/src/registryctl/api/registry/manifest"
	"net/http"

	"github.com/goharbor/harbor/src/registryctl/api"
	"github.com/goharbor/harbor/src/registryctl/api/registry/blob"
	"github.com/goharbor/harbor/src/registryctl/api/registry/gc"
	"github.com/goharbor/harbor/src/registryctl/config"
	"github.com/gorilla/mux"
)

func newRouter(conf config.Configuration) http.Handler {
	// create the root rooter
	rootRouter := mux.NewRouter()
	rootRouter.StrictSlash(true)
	rootRouter.HandleFunc("/api/health", api.Health).Methods("GET")

	rootRouter.Path("/api/registry/gc").Methods(http.MethodPost).Handler(gc.NewHandler(conf.RegistryConfig))
	rootRouter.Path("/api/registry/blob/{reference}").Methods(http.MethodDelete).Handler(blob.NewHandler(conf.StorageDriver))
	rootRouter.Path("/api/registry/{name:.*}/manifests/{reference}").Methods(http.MethodDelete).Handler(manifest.NewHandler(conf.StorageDriver))
	return rootRouter
}
