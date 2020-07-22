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

package internal

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	scantesting "github.com/goharbor/harbor/src/testing/controller/scan"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type AutoScanTestSuite struct {
	suite.Suite

	originalProjectController project.Controller
	projectController         *projecttesting.Controller

	originalScanController scan.Controller
	scanController         *scantesting.Controller
}

func (suite *AutoScanTestSuite) SetupTest() {
	suite.originalProjectController = project.Ctl
	suite.projectController = &projecttesting.Controller{}
	project.Ctl = suite.projectController

	suite.originalScanController = scan.DefaultController
	suite.scanController = &scantesting.Controller{}
	scan.DefaultController = suite.scanController
}

func (suite *AutoScanTestSuite) TearDownTest() {
	project.Ctl = suite.originalProjectController
	scan.DefaultController = suite.originalScanController
}

func (suite *AutoScanTestSuite) TestGetProjectFailed() {
	mock.OnAnything(suite.projectController, "Get").Return(nil, errors.NotFoundError(nil))

	ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})
	art := &artifact.Artifact{}

	suite.Error(autoScan(ctx, art))
}

func (suite *AutoScanTestSuite) TestAutoScanDisabled() {
	mock.OnAnything(suite.projectController, "Get").Return(&models.Project{
		Metadata: map[string]string{
			models.ProMetaAutoScan: "false",
		},
	}, nil)

	ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})
	art := &artifact.Artifact{}

	suite.Nil(autoScan(ctx, art))
}

func (suite *AutoScanTestSuite) TestAutoScan() {
	mock.OnAnything(suite.projectController, "Get").Return(&models.Project{
		Metadata: map[string]string{
			models.ProMetaAutoScan: "true",
		},
	}, nil)

	mock.OnAnything(suite.scanController, "Scan").Return(nil)

	ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})
	art := &artifact.Artifact{}

	suite.Nil(autoScan(ctx, art))
}

func (suite *AutoScanTestSuite) TestAutoScanFailed() {
	mock.OnAnything(suite.projectController, "Get").Return(&models.Project{
		Metadata: map[string]string{
			models.ProMetaAutoScan: "true",
		},
	}, nil)

	mock.OnAnything(suite.scanController, "Scan").Return(errors.ConflictError(nil))

	ctx := orm.NewContext(nil, &ormtesting.FakeOrmer{})
	art := &artifact.Artifact{}

	suite.Error(autoScan(ctx, art))
}

func TestAutoScanTestSuite(t *testing.T) {
	suite.Run(t, &AutoScanTestSuite{})
}
