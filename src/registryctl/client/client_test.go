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

package client

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"net/http"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
)

type clientTestSuite struct {
	suite.Suite
	client Client
}

func (c *clientTestSuite) SetupTest() {
	server, err := test.NewRegistryCtl(nil)
	if err != nil {
		fmt.Printf("failed to create registry: %v", err)
		os.Exit(1)
	}
	c.client = NewClient(server.URL, &Config{})
}

func (c *clientTestSuite) TesHealth() {
	err := c.client.Health()
	c.Require().Nil(err)
}

func (c *clientTestSuite) TesStartGC() {
	gcr, err := c.client.StartGC()
	c.Require().Nil(err)
	c.Equal(gcr.Msg, "hello-world")
	c.Equal(gcr.Status, true)
}

func (c *clientTestSuite) TestDeleteManifest() {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "DELETE",
			Pattern: "/api/registry/library/hello-world/manifests/latest",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusAccepted,
			}),
		})
	defer server.Close()

	err := NewClient(server.URL, &Config{}).DeleteManifest("library/hello-world", "latest")
	c.Require().Nil(err)
}

func (c *clientTestSuite) TestDeleteBlob() {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "DELETE",
			Pattern: "/api/registry/blob/sha256:adfasa34r2sfadf234n23n4",
			Handler: test.Handler(&test.Response{
				StatusCode: http.StatusAccepted,
			}),
		})
	defer server.Close()

	err := NewClient(server.URL, &Config{}).DeleteBlob("sha256:adfasa34r2sfadf234n23n4")
	c.Require().Nil(err)
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, &clientTestSuite{})
}
