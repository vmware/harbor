// Copyright 2018 Project Harbor Authors
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

package api

import (
	"context"
	"fmt"
	"testing"

	o "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/testing/apitests/apilib"
	"github.com/goharbor/harbor/src/testing/mock"
	drivertesting "github.com/goharbor/harbor/src/testing/pkg/quota/driver"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	reference  = uuid.New().String()
	hardLimits = types.ResourceList{types.ResourceStorage: -1}
)

func init() {
	mockDriver := &drivertesting.Driver{}

	mockHardLimitsFn := func() types.ResourceList {
		return hardLimits
	}

	mockLoadFn := func(ctx context.Context, key string) driver.RefObject {
		return driver.RefObject{"id": key}
	}

	mockValidateFn := func(hardLimits types.ResourceList) error {
		if len(hardLimits) == 0 {
			return fmt.Errorf("no resources found")
		}

		return nil
	}

	mockDriver.On("HardLimits").Return(mockHardLimitsFn)
	mock.OnAnything(mockDriver, "Load").Return(mockLoadFn, nil)
	mock.OnAnything(mockDriver, "Validate").Return(mockValidateFn)

	driver.Register(reference, mockDriver)
}

func TestQuotaAPIList(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()

	ctx := orm.NewContext(context.TODO(), o.NewOrm())
	var quotaIDs []int64
	defer func() {
		for _, quotaID := range quotaIDs {
			quota.Ctl.Delete(ctx, quotaID)
		}
	}()

	count := 10
	for i := 0; i < count; i++ {
		quotaID, err := quota.Ctl.Create(ctx, reference, uuid.New().String(), hardLimits)
		assert.Nil(err)
		quotaIDs = append(quotaIDs, quotaID)
	}

	code, quotas, err := apiTest.QuotasGet(&apilib.QuotaQuery{Reference: reference}, *admin)
	assert.Nil(err)
	assert.Equal(int(200), code)
	assert.Len(quotas, count, fmt.Sprintf("quotas len should be %d", count))

	code, quotas, err = apiTest.QuotasGet(&apilib.QuotaQuery{Reference: reference, PageSize: 1}, *admin)
	assert.Nil(err)
	assert.Equal(int(200), code)
	assert.Len(quotas, 1)
}

func TestQuotaAPIGet(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()

	ctx := orm.NewContext(context.TODO(), o.NewOrm())
	quotaID, err := quota.Ctl.Create(ctx, reference, uuid.New().String(), hardLimits)
	assert.Nil(err)
	defer quota.Ctl.Delete(ctx, quotaID)

	code, quota, err := apiTest.QuotasGetByID(*admin, fmt.Sprintf("%d", quotaID))
	assert.Nil(err)
	assert.Equal(int(200), code)
	assert.Equal(map[string]int64{"storage": -1}, quota.Hard)

	code, _, err = apiTest.QuotasGetByID(*admin, "100")
	assert.Nil(err)
	assert.Equal(int(404), code)
}

func TestQuotaPut(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()

	ctx := orm.NewContext(context.TODO(), o.NewOrm())
	quotaID, err := quota.Ctl.Create(ctx, reference, uuid.New().String(), hardLimits)
	assert.Nil(err)
	defer quota.Ctl.Delete(ctx, quotaID)

	code, quota, err := apiTest.QuotasGetByID(*admin, fmt.Sprintf("%d", quotaID))
	assert.Nil(err)
	assert.Equal(int(200), code)
	assert.Equal(map[string]int64{"storage": -1}, quota.Hard)

	code, err = apiTest.QuotasPut(*admin, fmt.Sprintf("%d", quotaID), QuotaUpdateRequest{})
	assert.Nil(err, err)
	assert.Equal(int(400), code)

	code, err = apiTest.QuotasPut(*admin, fmt.Sprintf("%d", quotaID), QuotaUpdateRequest{Hard: types.ResourceList{types.ResourceStorage: 100}})
	assert.Nil(err)
	assert.Equal(int(200), code)

	code, quota, err = apiTest.QuotasGetByID(*admin, fmt.Sprintf("%d", quotaID))
	assert.Nil(err)
	assert.Equal(int(200), code)
	assert.Equal(map[string]int64{"storage": 100}, quota.Hard)
}
