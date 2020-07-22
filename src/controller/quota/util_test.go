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

package quota

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	quotatesting "github.com/goharbor/harbor/src/testing/pkg/quota"
	drivertesting "github.com/goharbor/harbor/src/testing/pkg/quota/driver"
	"github.com/stretchr/testify/suite"
)

type RefreshForProjectsTestSuite struct {
	suite.Suite

	originalProjectCtl project.Controller
	projectCtl         *projecttesting.Controller

	originalQuotaCtl Controller
	quotaMgr         quota.Manager

	originalDriver driver.Driver
	driver         *drivertesting.Driver
}

func (suite *RefreshForProjectsTestSuite) SetupTest() {
	suite.originalDriver, _ = Driver(context.TODO(), ProjectReference)
	suite.driver = &drivertesting.Driver{}
	driver.Register(ProjectReference, suite.driver)

	suite.originalProjectCtl = project.Ctl
	suite.projectCtl = &projecttesting.Controller{}
	project.Ctl = suite.projectCtl

	suite.originalQuotaCtl = Ctl

	suite.quotaMgr = &quotatesting.Manager{}
	Ctl = &controller{
		quotaMgr: suite.quotaMgr,
	}
}

func (suite *RefreshForProjectsTestSuite) TearDownTest() {
	project.Ctl = suite.originalProjectCtl
	Ctl = suite.originalQuotaCtl

	driver.Register(ProjectReference, suite.originalDriver)
}

func (suite *RefreshForProjectsTestSuite) TestRefreshForProjects() {
	rand.Seed(time.Now().UnixNano())

	startProjectID := rand.Int63()
	var firstPageProjects, secondPageProjects []*models.Project
	for i := 0; i < 50; i++ {
		firstPageProjects = append(firstPageProjects, &models.Project{
			ProjectID: startProjectID + int64(i),
		})
	}

	for i := 0; i < 10; i++ {
		secondPageProjects = append(secondPageProjects, &models.Project{
			ProjectID: startProjectID + 50 + int64(i),
		})
	}

	page := 1
	mock.OnAnything(suite.projectCtl, "List").Return(func(context.Context, *models.ProjectQueryParam, ...project.Option) []*models.Project {
		defer func() {
			page++
		}()

		if page == 1 {
			return firstPageProjects
		} else if page == 2 {
			return secondPageProjects
		} else {
			return nil
		}
	}, nil)

	q := &quota.Quota{}
	q.SetHard(types.ResourceList{types.ResourceStorage: 10})
	q.SetUsed(types.ResourceList{types.ResourceStorage: 0})

	mock.OnAnything(suite.quotaMgr, "GetByRef").Return(q, nil)
	mock.OnAnything(suite.quotaMgr, "GetByRefForUpdate").Return(q, nil)
	mock.OnAnything(suite.quotaMgr, "Update").Return(nil)
	mock.OnAnything(suite.driver, "CalculateUsage").Return(types.ResourceList{types.ResourceStorage: 1}, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	RefreshForProjects(ctx)
	suite.Equal(3, page)
}

func TestRefreshForProjectsTestSuite(t *testing.T) {
	suite.Run(t, &RefreshForProjectsTestSuite{})
}
