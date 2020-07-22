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

package selector

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	// Image kind
	Image = "image"
	// Chart kind
	Chart = "chart"
)

// Repository of candidate
type Repository struct {
	// Namespace(project) ID
	NamespaceID int64
	// Namespace
	Namespace string `json:"namespace"`
	// Repository name
	Name string `json:"name"`
	// So far we need the kind of repository and retrieve candidates with different APIs
	// TODO: REMOVE IT IN THE FUTURE IF WE SUPPORT UNIFIED ARTIFACT MODEL
	Kind string `json:"kind"`
}

// ToJSON marshals repository to JSON string
func (r *Repository) ToJSON() (string, error) {
	jsonData, err := json.Marshal(r)
	if err != nil {
		return "", errors.Wrap(err, "marshal reporitory")
	}

	return string(jsonData), nil
}

// FromJSON constructs the repository from json data
func (r *Repository) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to construct repository")
	}

	return json.Unmarshal([]byte(jsonData), r)
}

// Candidate for retention processor to match
type Candidate struct {
	// Namespace(project) ID
	NamespaceID int64 `json:"namespace_id"`
	// Namespace
	Namespace string `json:"namespace"`
	// Repository name
	Repository string `json:"repository"`
	// Kind of the candidate
	// "image" or "chart"
	Kind string `json:"kind"`
	// Tags attached with the candidate
	Tags []string `json:"tags"`
	// Digest
	Digest string `json:"digest"`
	// Pushed time in seconds
	PushedTime int64 `json:"pushed_time_second"`
	// Pulled time in seconds
	PulledTime int64 `json:"pulled_time_second"`
	// Created time in seconds
	CreationTime int64 `json:"create_time_second"`
	// Labels attached with the candidate
	Labels []string `json:"labels"`
	// Overall severity of the candidate
	// Use severity code value here to avoid pkg dependency issue.
	VulnerabilitySeverity uint `json:"vulnerability_severity"`
	// Signatures of the above Tags
	// This is not technical correct, just for keeping compatibilities with the original definition.
	Signatures map[string]bool `json:"signatures"`
}

// Hash code based on the candidate info for differentiation
func (c *Candidate) Hash() string {
	if c.Digest == "" {
		log.Errorf("Lack Digest of Candidate for %s/%s", c.Namespace, c.Repository)
	}
	raw := fmt.Sprintf("%s:%s/%s:%s", c.Kind, c.Namespace, c.Repository, c.Digest)

	return base64.StdEncoding.EncodeToString([]byte(raw))
}
