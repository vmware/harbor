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
	"testing"

	"github.com/goharbor/harbor/src/pkg/quota/types"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

type ManagerTestSuite struct {
	htesting.Suite
}

func (suite *ManagerTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearSQLs = []string{
		"DELETE FROM quota WHERE id > 1",
		"DELETE FROM quota_usage WHERE id > 1",
	}
}

func (suite *ManagerTestSuite) TestCreate() {
	ctx := suite.Context()

	{
		hardLimits := types.ResourceList{types.ResourceStorage: 100}
		id, err := Mgr.Create(ctx, "project", "1000", hardLimits)

		if suite.Nil(err) {
			q, err := Mgr.Get(ctx, id)
			suite.Nil(err)

			hard, _ := q.GetHard()
			used, _ := q.GetUsed()

			suite.Equal(hardLimits, hard)
			suite.Equal(types.Zero(hardLimits), used)
		}

		Mgr.Delete(ctx, id)
	}

	{
		hardLimits := types.ResourceList{types.ResourceStorage: 100}
		usage := types.ResourceList{types.ResourceStorage: 10}
		id, err := Mgr.Create(ctx, "project", "1000", hardLimits, usage)

		if suite.Nil(err) {
			q, err := Mgr.Get(ctx, id)
			suite.Nil(err)

			hard, _ := q.GetHard()
			used, _ := q.GetUsed()

			suite.Equal(hardLimits, hard)
			suite.Equal(usage, used)
		}

		Mgr.Delete(ctx, id)
	}
}

func (suite *ManagerTestSuite) TestUpdate() {
	ctx := suite.Context()

	{
		hardLimits := types.ResourceList{types.ResourceStorage: 100}
		id, err := Mgr.Create(ctx, "project", "1000", hardLimits)

		q, err := Mgr.Get(ctx, id)
		if suite.Nil(err) {
			hard, _ := q.GetHard()
			used, _ := q.GetUsed()
			suite.Equal(hardLimits, hard)
			suite.Equal(types.Zero(hardLimits), used)
		}

		{
			q.SetUsed(hardLimits)
			if suite.Nil(Mgr.Update(ctx, q)) {
				q1, err := Mgr.Get(ctx, id)
				suite.Nil(err)
				hard, _ := q1.GetHard()
				used, _ := q1.GetUsed()
				suite.Equal(hardLimits, hard)
				suite.Equal(hardLimits, used)
			}
		}

		Mgr.Delete(ctx, id)
	}
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}
