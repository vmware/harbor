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

package scheduler

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type callbackTestSuite struct {
	suite.Suite
}

func (c *callbackTestSuite) SetupTest() {
	registry = map[string]CallbackFunc{}
	err := RegisterCallbackFunc("callback", func(interface{}) error { return nil })
	c.Require().Nil(err)
}

func (c *callbackTestSuite) TestRegisterCallbackFunc() {
	// empty name
	err := RegisterCallbackFunc("", nil)
	c.NotNil(err)

	// nil callback function
	err = RegisterCallbackFunc("test", nil)
	c.NotNil(err)

	// pass
	err = RegisterCallbackFunc("test", func(interface{}) error { return nil })
	c.Nil(err)

	// duplicate name
	err = RegisterCallbackFunc("test", func(interface{}) error { return nil })
	c.NotNil(err)
}

func (c *callbackTestSuite) TestGetCallbackFunc() {
	// not exist
	_, err := getCallbackFunc("not-exist")
	c.NotNil(err)

	// pass
	f, err := getCallbackFunc("callback")
	c.Require().Nil(err)
	c.NotNil(f)
}

func (c *callbackTestSuite) TestCallbackFuncExist() {
	// not exist
	c.False(callbackFuncExist("not-exist"))

	// exist
	c.True(callbackFuncExist("callback"))
}

func TestCallbackTestSuite(t *testing.T) {
	s := &callbackTestSuite{}
	suite.Run(t, s)
}
