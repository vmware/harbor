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

package provider

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	cm "github.com/goharbor/harbor/src/common/models"
)

// This is a package to provide mock utilities.

// MockDragonflyProvider mocks a Dragonfly server.
func MockDragonflyProvider() *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case healthCheckEndpoint:
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			w.WriteHeader(http.StatusOK)
		case preheatEndpoint:
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			image := &PreheatImage{}
			if err := json.Unmarshal(data, image); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			if image.Type == "image" &&
				image.URL == "https://harbor.com" &&
				image.ImageName == "busybox" &&
				image.Tag == "latest" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ID":"dragonfly-id"}`))
				return
			}

			w.WriteHeader(http.StatusBadRequest)
		case strings.Replace(preheatTaskEndpoint, "{task_id}", "dragonfly-id", 1):
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}
			status := &dragonflyPreheatInfo{
				ID:         "dragonfly-id",
				StartTime:  time.Now().UTC().String(),
				FinishTime: time.Now().Add(5 * time.Minute).UTC().String(),
				Status:     "SUCCESS",
			}
			bytes, _ := json.Marshal(status)
			_, _ = w.Write(bytes)
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
}

// MockKrakenProvider mocks a Kraken server.
func MockKrakenProvider() *httptest.Server {
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case krakenHealthPath:
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			w.WriteHeader(http.StatusOK)
		case krakenPreheatPath:
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			var payload = &cm.Notification{
				Events: []cm.Event{},
			}

			if err := json.Unmarshal(data, payload); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			if len(payload.Events) > 0 {
				w.WriteHeader(http.StatusOK)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
}
