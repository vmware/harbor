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

package server

import (
	"encoding/json"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"net/http"

	"github.com/goharbor/harbor/src/server/v2.0/route"
)

var (
	version = route.APIVersion
)

// APIVersion model
type APIVersion struct {
	Version string `json:"version"`
}

// GetAPIVersion returns the current supported API version
func GetAPIVersion(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(&APIVersion{Version: version}); err != nil {
		lib_http.SendError(w, err)
	}
}
