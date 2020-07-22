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
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	quotatesting "github.com/goharbor/harbor/src/testing/pkg/quota"
	drivertesting "github.com/goharbor/harbor/src/testing/pkg/quota/driver"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite

	reference string
	driver    *drivertesting.Driver
	quotaMgr  *quotatesting.Manager
	ctl       Controller

	quota *quota.Quota
}

func (suite *ControllerTestSuite) SetupTest() {
	suite.reference = "mock"

	suite.driver = &drivertesting.Driver{}
	driver.Register(suite.reference, suite.driver)

	suite.quotaMgr = &quotatesting.Manager{}
	suite.ctl = &controller{quotaMgr: suite.quotaMgr, reservedExpiration: defaultReservedExpiration}

	hardLimits := types.ResourceList{types.ResourceStorage: 100}
	suite.quota = &quota.Quota{Hard: hardLimits.String(), Used: types.Zero(hardLimits).String()}
}

func (suite *ControllerTestSuite) PrepareForUpdate(q *quota.Quota, newUsage interface{}) {
	mock.OnAnything(suite.quotaMgr, "GetByRefForUpdate").Return(q, nil)

	mock.OnAnything(suite.driver, "CalculateUsage").Return(newUsage, nil)

	mock.OnAnything(suite.quotaMgr, "Update").Return(nil)
}

func (suite *ControllerTestSuite) TestGetReservedResources() {
	reservedExpiration := time.Second * 3
	ctl := &controller{reservedExpiration: reservedExpiration}

	reference, referenceID := "reference", uuid.New().String()

	{
		resources, err := ctl.getReservedResources(context.TODO(), reference, referenceID)
		suite.Nil(err)
		suite.Len(resources, 0)
	}

	suite.Nil(ctl.setReservedResources(context.TODO(), reference, referenceID, types.ResourceList{types.ResourceStorage: 100}))

	{
		resources, err := ctl.getReservedResources(context.TODO(), reference, referenceID)
		suite.Nil(err)
		suite.Len(resources, 1)
	}

	time.Sleep(reservedExpiration)

	{
		resources, err := ctl.getReservedResources(context.TODO(), reference, referenceID)
		suite.Nil(err)
		suite.Len(resources, 0)
	}
}

func (suite *ControllerTestSuite) TestReserveResources() {
	mock.OnAnything(suite.quotaMgr, "GetByRefForUpdate").Return(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()
	resources := types.ResourceList{types.ResourceStorage: 100}

	ctl := suite.ctl.(*controller)

	suite.Nil(ctl.reserveResources(ctx, suite.reference, referenceID, resources))

	suite.Error(ctl.reserveResources(ctx, suite.reference, referenceID, resources))
}

func (suite *ControllerTestSuite) TestUnreserveResources() {
	mock.OnAnything(suite.quotaMgr, "GetByRefForUpdate").Return(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()
	resources := types.ResourceList{types.ResourceStorage: 100}

	ctl := suite.ctl.(*controller)

	suite.Nil(ctl.reserveResources(ctx, suite.reference, referenceID, resources))

	suite.Error(ctl.reserveResources(ctx, suite.reference, referenceID, resources))

	suite.Nil(ctl.unreserveResources(ctx, suite.reference, referenceID, resources))

	suite.Nil(ctl.reserveResources(ctx, suite.reference, referenceID, resources))
}

func (suite *ControllerTestSuite) TestRefresh() {
	suite.PrepareForUpdate(suite.quota, types.ResourceList{types.ResourceStorage: 0})

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()

	suite.Nil(suite.ctl.Refresh(ctx, suite.reference, referenceID))
}

func (suite *ControllerTestSuite) TestRefreshDriverNotFound() {
	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})

	suite.Error(suite.ctl.Refresh(ctx, uuid.New().String(), uuid.New().String()))
}

func (suite *ControllerTestSuite) TestRefershNegativeUsage() {
	suite.PrepareForUpdate(suite.quota, types.ResourceList{types.ResourceStorage: -1})

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()

	suite.Error(suite.ctl.Refresh(ctx, suite.reference, referenceID))
}

func (suite *ControllerTestSuite) TestRefreshUsageExceed() {
	suite.PrepareForUpdate(suite.quota, types.ResourceList{types.ResourceStorage: 101})

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()

	suite.Error(suite.ctl.Refresh(ctx, suite.reference, referenceID))
}

func (suite *ControllerTestSuite) TestRefreshIgnoreLimitation() {
	suite.PrepareForUpdate(suite.quota, types.ResourceList{types.ResourceStorage: 101})

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()

	suite.Nil(suite.ctl.Refresh(ctx, suite.reference, referenceID, IgnoreLimitation(true)))
}

func (suite *ControllerTestSuite) TestNoResourcesRequest() {
	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()

	suite.Nil(suite.ctl.Request(ctx, suite.reference, referenceID, nil, func() error { return nil }))
}
func (suite *ControllerTestSuite) TestRequest() {
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()
	resources := types.ResourceList{types.ResourceStorage: 100}

	suite.Nil(suite.ctl.Request(ctx, suite.reference, referenceID, resources, func() error { return nil }))
}

func (suite *ControllerTestSuite) TestRequestExceed() {
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()
	resources := types.ResourceList{types.ResourceStorage: 101}

	suite.Error(suite.ctl.Request(ctx, suite.reference, referenceID, resources, func() error { return nil }))
}

func (suite *ControllerTestSuite) TestRequestFunctionFailed() {
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()
	resources := types.ResourceList{types.ResourceStorage: 100}

	suite.Error(suite.ctl.Request(ctx, suite.reference, referenceID, resources, func() error { return fmt.Errorf("error") }))
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}

func BenchmarkGetReservedResources(b *testing.B) {
	ctl := &controller{reservedExpiration: defaultReservedExpiration}

	ctx := context.TODO()
	reference, referenceID := "reference", uuid.New().String()
	ctl.setReservedResources(ctx, reference, referenceID, types.ResourceList{types.ResourceStorage: 100})

	for i := 0; i < b.N; i++ {
		ctl.getReservedResources(ctx, reference, referenceID)
	}
}

func BenchmarkSetReservedResources(b *testing.B) {
	ctl := &controller{reservedExpiration: defaultReservedExpiration}

	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		s := strconv.Itoa(i)
		ctl.setReservedResources(ctx, "reference"+s, s, types.ResourceList{types.ResourceStorage: 100})
	}
}
