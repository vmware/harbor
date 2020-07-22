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

package dao

import (
	"context"
	"sync"
	"testing"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearSQLs = []string{
		"DELETE FROM quota WHERE id > 1",
		"DELETE FROM quota_usage WHERE id > 1",
	}
	suite.dao = New()
}

func (suite *DaoTestSuite) TestCount() {
	suite.Suite.TearDownSuite() // Clean other quotas

	reference := uuid.New().String()
	hardLimits := types.ResourceList{types.ResourceStorage: 100}
	usage := types.ResourceList{types.ResourceStorage: 0}

	ctx := suite.Context()

	suite.dao.Create(ctx, reference, "1", types.ResourceList{types.ResourceStorage: 200}, usage)
	suite.dao.Create(ctx, reference, "2", hardLimits, usage)
	suite.dao.Create(ctx, reference, "3", hardLimits, usage)
	suite.dao.Create(ctx, uuid.New().String(), "4", types.ResourceList{types.ResourceStorage: 10}, usage)

	{
		// Count all the quotas
		count, err := suite.dao.Count(ctx, nil)
		suite.Nil(err)
		suite.Equal(int64(5), count) // 4 + library project quota
	}

	{
		// Count quotas filter by reference
		count, err := suite.dao.Count(ctx, q.New(q.KeyWords{"reference": reference}))
		suite.Nil(err)
		suite.Equal(int64(3), count)
	}

	{
		// Count quotas filter by reference ids
		count, err := suite.dao.Count(ctx, q.New(q.KeyWords{"reference": reference, "reference_ids": []string{"1", "2"}}))
		suite.Nil(err)
		suite.Equal(int64(2), count)
	}
}

func (suite *DaoTestSuite) TestCreate() {
	hardLimits := types.ResourceList{types.ResourceStorage: 100}
	usage := types.ResourceList{types.ResourceStorage: 0}
	id, err := suite.dao.Create(suite.Context(), "project", "2", hardLimits, usage)
	suite.Nil(err)

	q, err := suite.dao.Get(suite.Context(), id)
	if suite.Nil(err) {
		hard, _ := q.GetHard()
		used, _ := q.GetUsed()

		suite.Equal(hardLimits, hard)
		suite.Equal(usage, used)
	}
}

func (suite *DaoTestSuite) TestDelete() {
	hardLimits := types.ResourceList{types.ResourceStorage: 100}
	usage := types.ResourceList{types.ResourceStorage: 0}

	id, err := suite.dao.Create(suite.Context(), "project", "3", hardLimits, usage)
	suite.Nil(err)

	{
		q, err := suite.dao.Get(suite.Context(), id)
		suite.Nil(err)
		suite.NotNil(q)
	}

	suite.Nil(suite.dao.Delete(suite.Context(), id))

	{
		_, err := suite.dao.Get(suite.Context(), id)
		suite.Error(err)
	}
}

func (suite *DaoTestSuite) TestGetByRef() {
	hardLimits := types.ResourceList{types.ResourceStorage: 100}
	usage := types.ResourceList{types.ResourceStorage: 0}

	reference, referenceID := "project", "4"
	id, err := suite.dao.Create(suite.Context(), reference, referenceID, hardLimits, usage)
	suite.Nil(err)

	{
		q, err := suite.dao.GetByRef(suite.Context(), reference, referenceID)
		suite.Nil(err)
		suite.NotNil(q)
	}

	suite.Nil(suite.dao.Delete(suite.Context(), id))

	{
		_, err := suite.dao.GetByRef(suite.Context(), reference, referenceID)
		suite.Error(err)
	}
}

func (suite *DaoTestSuite) TestGetByRefForUpdate() {
	hardLimits := types.ResourceList{types.ResourceStorage: 100}
	usage := types.ResourceList{types.ResourceStorage: 0}

	reference, referenceID := "project", "5"
	id, err := suite.dao.Create(suite.Context(), reference, referenceID, hardLimits, usage)
	suite.Nil(err)

	var wg sync.WaitGroup

	count := int64(10)

	for i := int64(0); i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f := func(ctx context.Context) error {
				q, err := suite.dao.GetByRefForUpdate(ctx, reference, referenceID)
				suite.Nil(err)

				used, _ := q.GetUsed()
				used[types.ResourceStorage]++
				q.SetUsed(used)

				suite.dao.Update(ctx, q)

				return nil
			}

			orm.WithTransaction(f)(suite.Context())
		}()
	}
	wg.Wait()

	{
		q, err := suite.dao.Get(suite.Context(), id)
		suite.Nil(err)
		used, _ := q.GetUsed()
		suite.Equal(count, used[types.ResourceStorage])
	}
}

func (suite *DaoTestSuite) TestUpdate() {
	hardLimits := types.ResourceList{types.ResourceStorage: 100}
	usage := types.ResourceList{types.ResourceStorage: 0}

	id, err := suite.dao.Create(suite.Context(), "project", "6", hardLimits, usage)
	suite.Nil(err)

	newHardLimits := types.ResourceList{types.ResourceStorage: 200}
	newUsage := types.ResourceList{types.ResourceStorage: 1}

	{
		q, err := suite.dao.Get(suite.Context(), id)
		if suite.Nil(err) {
			q.SetHard(newHardLimits).SetUsed(newUsage)

			suite.Nil(suite.dao.Update(suite.Context(), q))
		}
	}

	{
		q, err := suite.dao.Get(suite.Context(), id)
		if suite.Nil(err) {
			hard, _ := q.GetHard()
			used, _ := q.GetUsed()

			suite.Equal(newHardLimits, hard)
			suite.Equal(newUsage, used)
		}
	}
}

func (suite *DaoTestSuite) TestList() {
	suite.Suite.TearDownSuite() // Clean other quotas

	reference := uuid.New().String()
	hardLimits := types.ResourceList{types.ResourceStorage: 100}
	usage := types.ResourceList{types.ResourceStorage: 0}

	ctx := suite.Context()

	suite.dao.Create(ctx, reference, "1", types.ResourceList{types.ResourceStorage: 200}, usage)
	suite.dao.Create(ctx, reference, "2", hardLimits, usage)
	suite.dao.Create(ctx, reference, "3", hardLimits, usage)
	suite.dao.Create(ctx, uuid.New().String(), "4", types.ResourceList{types.ResourceStorage: 10}, usage)

	{
		// List all the quotas
		quotas, err := suite.dao.List(ctx, nil)
		suite.Nil(err)
		suite.Equal(5, len(quotas)) // 4 + library project quota
		suite.NotEqual(reference, quotas[0].Reference)
		suite.Equal("4", quotas[0].ReferenceID)
	}

	{
		// List quotas filter by reference
		quotas, err := suite.dao.List(ctx, q.New(q.KeyWords{"reference": reference}))
		suite.Nil(err)
		suite.Equal(3, len(quotas))
	}

	{
		// List quotas filter by reference ids
		quotas, err := suite.dao.List(ctx, q.New(q.KeyWords{"reference": reference, "reference_ids": []string{"1", "2"}}))
		suite.Nil(err)
		suite.Equal(2, len(quotas))
	}

	{
		// List quotas by pagination
		quotas, err := suite.dao.List(ctx, &q.Query{PageSize: 2})
		suite.Nil(err)
		suite.Equal(2, len(quotas))
	}

	{
		// List quotas by sorting
		quotas, err := suite.dao.List(ctx, &q.Query{Keywords: q.KeyWords{"reference": reference}, Sorting: "-hard.storage"})
		suite.Nil(err)
		suite.Equal(reference, quotas[0].Reference)
		suite.Equal("1", quotas[0].ReferenceID)
	}

}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
