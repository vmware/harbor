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
	"net/url"

	"github.com/pkg/errors"
)

const (
	// SupportedType indicates the supported preheating type 'image'.
	SupportedType = "image"
)

// PreheatImage contains related information which can help providers to get/pull the images.
type PreheatImage struct {
	// The image content type, only support 'image' now
	Type string `json:"type"`

	// The access URL of the preheating image
	URL string `json:"url"`

	// The headers which will be sent to the above URL of preheating image
	Headers map[string]interface{} `json:"headers"`

	// The image name
	ImageName string `json:"image,omitempty"`

	// The tag
	Tag string `json:"tag,omitempty"`

	// Digest of the preheating image
	Digest string `json:"digest"`
}

// FromJSON build preheating image from the given data.
func (img *PreheatImage) FromJSON(data string) error {
	if len(data) == 0 {
		return errors.New("empty JSON data")
	}

	if err := json.Unmarshal([]byte(data), img); err != nil {
		return errors.Wrap(err, "construct preheating image error")
	}

	return nil
}

// ToJSON encodes the preheating image to JSON data.
func (img *PreheatImage) ToJSON() (string, error) {
	data, err := json.Marshal(img)
	if err != nil {
		return "", errors.Wrap(err, "encode preheating image error")
	}

	return string(data), nil
}

// Validate PreheatImage
func (img *PreheatImage) Validate() error {
	if img.Type != SupportedType {
		return errors.Errorf("unsupported type '%s'", img.Type)
	}

	if len(img.ImageName) == 0 || len(img.Tag) == 0 {
		return errors.New("missing image repository or tag")
	}

	if len(img.Headers) == 0 {
		return errors.New("missing required headers")
	}

	_, err := url.Parse(img.URL)
	if err != nil {
		return errors.Wrap(err, "malformed registry URL")
	}

	return nil
}
