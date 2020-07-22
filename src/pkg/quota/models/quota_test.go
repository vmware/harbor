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

package models

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/stretchr/testify/assert"
)

func TestGetWarningResources(t *testing.T) {
	assert := assert.New(t)

	q := Quota{}

	q.SetHard(types.ResourceList{types.ResourceStorage: 300})
	q.SetUsed(types.ResourceList{types.ResourceStorage: 300})

	resources, err := q.GetWarningResources(85)
	assert.Nil(err)
	assert.Len(resources, 1)
}
