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
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

type MetaDaoTestSuite struct {
	htesting.Suite
	dao          MetaDAO
	userID       int
	username     string
	deleteUserID int
}

func (suite *MetaDaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.ClearSQLs = []string{}
	suite.dao = NewMetaDao()
	suite.userID = 1234
	suite.deleteUserID = 2234
	suite.username = "oidc_meta_testuser"
	suite.ExecSQL("INSERT INTO harbor_user (user_id, username,password,realname) VALUES(?,?,'test','test')", suite.userID, suite.username)
	suite.ExecSQL("INSERT INTO harbor_user (user_id, username,password,realname) VALUES(?,?,'test','test')", suite.deleteUserID, suite.deleteUserID)
	ctx := orm.Context()
	_, err := suite.dao.Create(ctx, &models.OIDCUser{
		UserID: suite.userID,
		SubIss: `ca4bb144-4b5c-4d1b-9469-69cb3768af8fhttps://sso.andrea.muellerpublic.de/auth/realms/harbor`,
		Secret: `<enc-v1>7uBP9yqtdnVAhoA243GSv8nOXBWygqzaaEdq9Kqla+q4hOaBZmEMH9vUJi4Yjbh3`,
		Token:  `xxxx`,
	})
	suite.Nil(err)
	suite.appendClearSQL(suite.userID)
	_, err = suite.dao.Create(ctx, &models.OIDCUser{
		UserID: suite.deleteUserID,
		SubIss: `ca4bb144-4b5c-4d1b-9469-69cb3768af9fhttps://sso.andrea.muellerpublic.de/auth/realms/harbor`,
		Secret: `<enc-v1>7uBP9yqtdnVAhoA243GSv8nOXBWygqzaaEdq9Kqla+q4hOaBZmEMH9vUJi4Yjbh3`,
		Token:  `xxxx`,
	})
	suite.Nil(err)
	suite.appendClearSQL(suite.deleteUserID)
}

func (suite *MetaDaoTestSuite) TestList() {
	ctx := orm.Context()
	l, err := suite.dao.List(ctx, q.New(q.KeyWords{"user_id": suite.userID}))
	suite.Nil(err)
	suite.Equal(1, len(l))
	suite.Equal("xxxx", l[0].Token)
}

func (suite *MetaDaoTestSuite) TestGetByUsername() {
	ctx := orm.Context()
	ou, err := suite.dao.GetByUsername(ctx, suite.username)
	suite.Nil(err)
	suite.Equal(suite.userID, ou.UserID)
	suite.Equal("ca4bb144-4b5c-4d1b-9469-69cb3768af8fhttps://sso.andrea.muellerpublic.de/auth/realms/harbor", ou.SubIss)
	suite.Equal("xxxx", ou.Token)
}

func (suite *MetaDaoTestSuite) TestUpdate() {
	ctx := orm.Context()
	l, err := suite.dao.List(ctx, q.New(q.KeyWords{"user_id": suite.userID}))
	suite.Nil(err)
	id := l[0].ID
	ou := &models.OIDCUser{
		ID:     id,
		Secret: "newsecret",
	}
	err = suite.dao.Update(ctx, ou, "secret")
	suite.Nil(err)
	l, err = suite.dao.List(ctx, q.New(q.KeyWords{"user_id": suite.userID}))
	suite.Nil(err)
	suite.Equal("newsecret", l[0].Secret)
}

func (suite *MetaDaoTestSuite) TestDeleteByUserId() {
	ctx := orm.Context()
	err := suite.dao.DeleteByUserID(ctx, suite.deleteUserID)
	suite.Nil(err)
	l, err := suite.dao.List(ctx, q.New(q.KeyWords{"user_id": suite.deleteUserID}))
	suite.Nil(err)
	suite.True(len(l) == 0)
}

func (suite *MetaDaoTestSuite) appendClearSQL(uid int) {
	suite.ClearSQLs = append(suite.ClearSQLs, fmt.Sprintf("DELETE FROM oidc_user WHERE user_id = %d", uid))
	suite.ClearSQLs = append(suite.ClearSQLs, fmt.Sprintf("DELETE FROM harbor_user WHERE user_id = %d", uid))
}

func TestMetaDaoTestSuite(t *testing.T) {
	suite.Run(t, &MetaDaoTestSuite{})
}
