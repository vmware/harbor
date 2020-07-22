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

package types

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ResourcesSuite struct {
	suite.Suite
}

func (suite *ResourcesSuite) TestNewResourceList() {
	res1, err1 := NewResourceList("")
	suite.Error(err1)
	suite.Nil(res1)
	suite.Equal(0, len(res1))

	res2, err2 := NewResourceList("{}")
	suite.Nil(err2)
	suite.NotNil(res2)
}

func (suite *ResourcesSuite) TestEquals() {
	suite.True(Equals(ResourceList{}, ResourceList{}))
	suite.True(Equals(ResourceList{ResourceStorage: 100}, ResourceList{ResourceStorage: 100}))
	suite.False(Equals(ResourceList{ResourceStorage: 100}, ResourceList{ResourceStorage: 200}))
}

func (suite *ResourcesSuite) TestAdd() {
	res1 := ResourceList{ResourceStorage: 100}
	res2 := ResourceList{ResourceStorage: 100}
	res3 := ResourceList{ResourceStorage: 100}

	suite.Equal(res1, Add(ResourceList{}, res1))
	suite.Equal(ResourceList{ResourceStorage: 200}, Add(res1, res2))
	suite.Equal(ResourceList{ResourceStorage: 200}, Add(res1, res3))
}

func (suite *ResourcesSuite) TestSubtract() {
	res1 := ResourceList{ResourceStorage: 100}
	res2 := ResourceList{ResourceStorage: 100}
	res3 := ResourceList{ResourceStorage: 100}

	suite.Equal(res1, Subtract(res1, ResourceList{}))
	suite.Equal(ResourceList{ResourceStorage: 0}, Subtract(res1, res2))
	suite.Equal(ResourceList{ResourceStorage: 0}, Subtract(res1, res3))
}

func (suite *ResourcesSuite) TestZero() {
	res1 := ResourceList{ResourceStorage: 100}
	res2 := ResourceList{ResourceStorage: 100}

	suite.Equal(ResourceList{}, Zero(ResourceList{}))
	suite.Equal(ResourceList{ResourceStorage: 0}, Zero(res1))
	suite.Equal(ResourceList{ResourceStorage: 0}, Zero(res2))
}

func (suite *ResourcesSuite) TestIsNegative() {
	suite.Len(IsNegative(ResourceList{ResourceStorage: -100}), 1)
	suite.Contains(IsNegative(ResourceList{ResourceStorage: -100}), ResourceStorage)
}

func TestRunResourcesSuite(t *testing.T) {
	suite.Run(t, new(ResourcesSuite))
}
