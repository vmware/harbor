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

package metadata

import (
	"testing"

	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/stretchr/testify/suite"
)

type quotaEventTestSuite struct {
	suite.Suite
}

func (suite *quotaEventTestSuite) TestResolveOfDeleteRepositoryEventMetadata() {
	e := &event.Event{}
	metadata := &QuotaMetaData{
		RepoName: "library/hello-world",
		Tag:      "latest",
		Digest:   "sha256:469b2a896fbc1123f4894ac8023003f23588967aee5c2cbbce15d6b49dfe048e",
		Level:    1,
		Msg:      "quota exceed",
	}
	err := metadata.Resolve(e)
	suite.Nil(err)
	suite.Equal(event2.TopicQuotaExceed, e.Topic)
	suite.NotNil(e.Data)
	data, ok := e.Data.(*event2.QuotaEvent)
	suite.True(ok)
	suite.Equal("library/hello-world", data.RepoName)
	suite.NotNil(data.Resource)
	suite.Equal("latest", data.Resource.Tag)
	suite.Equal("sha256:469b2a896fbc1123f4894ac8023003f23588967aee5c2cbbce15d6b49dfe048e", data.Resource.Digest)
}

func (suite *quotaEventTestSuite) TestNoResource() {
	e := &event.Event{}
	metadata := &QuotaMetaData{
		RepoName: "library/hello-world",
		Level:    2,
		Msg:      "quota exceed",
	}
	err := metadata.Resolve(e)
	suite.Nil(err)
	suite.Equal(event2.TopicQuotaWarning, e.Topic)
	suite.NotNil(e.Data)
	data, ok := e.Data.(*event2.QuotaEvent)
	suite.True(ok)
	suite.Equal("library/hello-world", data.RepoName)
	suite.Nil(data.Resource)
}

func (suite *quotaEventTestSuite) TestUnsupportedStatus() {
	e := &event.Event{}
	metadata := &QuotaMetaData{
		RepoName: "library/hello-world",
		Level:    3,
		Msg:      "quota exceed",
	}
	err := metadata.Resolve(e)
	suite.Error(err)
}

func TestQuotaEventTestSuite(t *testing.T) {
	suite.Run(t, &quotaEventTestSuite{})
}
