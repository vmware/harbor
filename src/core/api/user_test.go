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
	"fmt"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/testing/apitests/apilib"
	"github.com/stretchr/testify/assert"

	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/config"
)

var testUser0002ID, testUser0003ID int
var testUser0002, testUser0003 apilib.User
var testUser0002Auth, testUser0003Auth *usrInfo

func TestUsersPost(t *testing.T) {

	fmt.Println("Testing User Add")

	assert := assert.New(t)
	apiTest := newHarborAPI()
	config.Upload(map[string]interface{}{
		common.AUTHMode:         "db_auth",
		common.SelfRegistration: false,
	})

	// case 1: register a new user without authentication
	t.Log("case 1: Register user without authentication")
	code, err := apiTest.UsersPost(testUser0002)
	if err != nil {
		t.Error("Error occurred while add a test User", err.Error())
		t.Log(err)
	} else {
		assert.Equal(http.StatusUnauthorized, code, "case 1: Add user status should be 401 for unauthenticated request")
	}

	config.Upload(map[string]interface{}{
		common.SelfRegistration: true,
	})

	// case 2: register a new user without admin auth, expect 403, because self registration is on
	t.Log("case 2: Register user without admin auth")
	code, err = apiTest.UsersPost(testUser0002)
	if err != nil {
		t.Error("Error occurred while add a test User", err.Error())
		t.Log(err)
	} else {
		// Should be 403 as only admin can call this API, otherwise it has to be called from browser, with session id
		assert.Equal(http.StatusForbidden, code, "case 2: Add user status should be 403")
	}

	// case 3: register a new user with admin auth, but username is empty, expect 400
	t.Log("case 3: Register user with admin auth, but username is empty")
	code, err = apiTest.UsersPost(testUser0002, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(400, code, "case 3: Add user status should be 400")
	}

	// case 4: register a new user with admin auth, but bad username format, expect 400
	testUser0002.Username = "test@$"
	t.Log("case 4: Register user with admin auth, but bad username format")
	code, err = apiTest.UsersPost(testUser0002, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(400, code, "case 4: Add user status should be 400")
	}

	// case 5: register a new user with admin auth, but bad userpassword format, expect 400
	testUser0002.Username = "testUser0002"
	t.Log("case 5: Register user with admin auth, but empty password.")
	code, err = apiTest.UsersPost(testUser0002, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(400, code, "case 5: Add user status should be 400")
	}

	// case 6: register a new user with admin auth, but email is empty, expect 400
	testUser0002.Password = "testUser0002"
	t.Log("case 6: Register user with admin auth, but email is empty")
	code, err = apiTest.UsersPost(testUser0002, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(400, code, "case 6: Add user status should be 400")
	}

	// case 7: register a new user with admin auth, but bad email format, expect 400
	testUser0002.Email = "test..."
	t.Log("case 7: Register user with admin auth, but bad email format")
	code, err = apiTest.UsersPost(testUser0002, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(400, code, "case 7: Add user status should be 400")
	}

	// case 7: register a new user with admin auth, but userrealname is empty, expect 400
	/*
		testUser0002.Email = "testUser0002@mydomain.com"
		fmt.Println("Register user with admin auth, but user realname is empty")
		code, err = apiTest.UsersPost(testUser0002, *admin)
		if err != nil {
			t.Error("Error occurred while add a user", err.Error())
			t.Log(err)
		} else {
			assert.Equal(400, code, "Add user status should be 400")
		}
	*/
	// case 8: register a new user with admin auth, but bad userrealname format, expect 400
	testUser0002.Email = "testUser0002@mydomain.com"
	testUser0002.Realname = "test$com"
	t.Log("case 8: Register user with admin auth, but bad user realname format")
	code, err = apiTest.UsersPost(testUser0002, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)

	} else {
		assert.Equal(400, code, "case 8: Add user status should be 400")
	}

	// case 9: register a new user with admin auth, but bad user comment, expect 400
	testUser0002.Realname = "testUser0002"
	testUser0002.Comment = "vmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm"
	t.Log("case 9: Register user with admin auth, but user comment length is illegal")
	code, err = apiTest.UsersPost(testUser0002, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(400, code, "case 9: Add user status should be 400")
	}
	testUser0002.Comment = "test user"

	// case 10: register an admin using non-admin user, expect 403
	t.Log("case 10: Register admin user with non admin auth")
	testUser0002.HasAdminRole = true
	code, err = apiTest.UsersPost(testUser0002)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(http.StatusForbidden, code, "case 10: Add user status should be 403")
	}
	testUser0002.HasAdminRole = false

	// case 11: register a new user with admin auth, expect 201
	t.Log("case 11: Register user with admin auth, right parameters")
	code, err = apiTest.UsersPost(testUser0002, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(201, code, "case 11: Add user status should be 201")
	}

	// case 12: register duplicate user with admin auth, expect 409
	t.Log("case 12: Register duplicate user with admin auth")
	code, err = apiTest.UsersPost(testUser0002, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(409, code, "case 12: Add user status should be 409")
	}

	// case 13: register a new user with admin auth, but duplicate email, expect 409
	t.Log("case 13: Register user with admin auth, but duplicate email")
	testUser0002.Username = "testUsertest"
	testUser0002.Email = "testUser0002@mydomain.com"
	code, err = apiTest.UsersPost(testUser0002, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(409, code, "case 13: Add user status should be 409")
	}
}

func TestUsersGet(t *testing.T) {

	fmt.Println("Testing User Get")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	testUser0002.Username = "testUser0002"
	// case 1: Get user2 with common auth, but no userid in path, expect 403

	testUser0002Auth = &usrInfo{"testUser0002", "testUser0002"}
	code, users, err := apiTest.UsersGet(testUser0002.Username, *testUser0002Auth)
	if err != nil {
		t.Error("Error occurred while get users", err.Error())
		t.Log(err)
	} else {
		assert.Equal(403, code, "Get users status should be 403")
	}
	// case 2: Get user2 with admin auth, expect 200
	code, users, err = apiTest.UsersGet(testUser0002.Username, *admin)
	if err != nil {
		t.Error("Error occurred while get users", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get users status should be 200")
		assert.Equal(1, len(users), "Get users record should be 1 ")
		testUser0002ID = users[0].UserId
	}
}

func TestUsersSearch(t *testing.T) {

	fmt.Println("Testing User Search")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	testUser0002.Username = "testUser0002"
	// case 1: Search user2 without auth, expect 401

	testUser0002Auth = &usrInfo{"testUser0002", "testUser0002"}
	code, users, err := apiTest.UsersSearch(testUser0002.Username)
	if err != nil {
		t.Error("Error occurred while search users", err.Error())
		t.Log(err)
	} else {
		assert.Equal(401, code, "Search users status should be 401")
	}
	// case 2: Search user2 with with common auth, expect 200
	code, users, err = apiTest.UsersSearch(testUser0002.Username, *testUser0002Auth)
	if err != nil {
		t.Error("Error occurred while search users", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Search users status should be 200")
		assert.Equal(1, len(users), "Search users record should be 1 ")
		testUser0002ID = users[0].UserID
	}
}

func TestUsersGetByID(t *testing.T) {

	fmt.Println("Testing User GetByID")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	// case 1: Get user2 with userID and his own auth, expect 200
	code, user, err := apiTest.UsersGetByID(testUser0002ID, *testUser0002Auth)
	if err != nil {
		t.Error("Error occurred while get users", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get users status should be 200")
		assert.Equal(testUser0002.Username, user.Username, "Get users username should be equal")
		assert.Equal(testUser0002.Email, user.Email, "Get users email should be equal")
	}
	// case 2: Get user2 with user3 auth, expect 403
	testUser0003.Username = "testUser0003"
	testUser0003.Email = "testUser0003@mydomain.com"
	testUser0003.Password = "testUser0003"
	testUser0003.Realname = "testUser0003"
	code, err = apiTest.UsersPost(testUser0003, *admin)
	if err != nil {
		t.Error("Error occurred while add a user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(201, code, "Add user status should be 201")
	}

	testUser0003Auth = &usrInfo{"testUser0003", "testUser0003"}
	code, user, err = apiTest.UsersGetByID(testUser0002ID, *testUser0003Auth)
	if err != nil {
		t.Error("Error occurred while get users", err.Error())
		t.Log(err)
	} else {
		assert.Equal(403, code, "Get users status should be 403")
	}
	// case 3: Get user that does not exist with user2 auth, expect 404 not found.
	code, user, err = apiTest.UsersGetByID(1000, *testUser0002Auth)
	if err != nil {
		t.Error("Error occurred while get users", err.Error())
		t.Log(err)
	} else {
		assert.Equal(404, code, "Get users status should be 404")
	}
	// Get user3ID in order to delete at the last of the test
	code, users, err := apiTest.UsersGet(testUser0003.Username, *admin)
	if err != nil {
		t.Error("Error occurred while get users", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get users status should be 200")
		assert.Equal(1, len(users), "Get users record should be 1")
		testUser0003ID = users[0].UserId
	}
}

func TestUsersPut(t *testing.T) {
	fmt.Println("Testing User Put")
	assert := assert.New(t)
	apiTest := newHarborAPI()
	var profile apilib.UserProfile
	// case 1: change user2 profile with user3 auth
	code, err := apiTest.UsersPut(testUser0002ID, profile, *testUser0003Auth)
	if err != nil {
		t.Error("Error occurred while change user profile", err.Error())
		t.Log(err)
	} else {
		assert.Equal(403, code, "Change user profile status should be 403")
	}
	// case 2: change user3 profile with user2 auth
	realname := "new realname"
	email := "new_email@mydomain.com"
	comment := "new comment"
	profile.UserID = testUser0003ID
	profile.Realname = realname
	profile.Email = email
	profile.Comment = comment
	code, err = apiTest.UsersPut(testUser0002ID, profile, *testUser0002Auth)
	if err != nil {
		t.Error("Error occurred while change user profile", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Change user profile status should be 200")
	}
	_, user, err := apiTest.UsersGetByID(testUser0003ID, *testUser0003Auth)
	if err != nil {
		t.Error("Error occurred while get user", err.Error())
	} else {
		assert.NotEqual(realname, user.Realname)
		assert.NotEqual(email, user.Email)
		assert.NotEqual(comment, user.Comment)
	}
	// case 3: change user2 profile with user2 auth, but bad parameters format.
	profile = apilib.UserProfile{}
	code, err = apiTest.UsersPut(testUser0002ID, profile, *testUser0002Auth)
	if err != nil {
		t.Error("Error occurred while change user profile", err.Error())
		t.Log(err)
	} else {
		assert.Equal(400, code, "Change user profile status should be 400")
	}
	// case 4: change user2 profile with user2 auth, but duplicate email.
	profile.Realname = "test user"
	profile.Email = "testUser0003@mydomain.com"
	profile.Comment = "change profile"
	code, err = apiTest.UsersPut(testUser0002ID, profile, *testUser0002Auth)
	if err != nil {
		t.Error("Error occurred while change user profile", err.Error())
		t.Log(err)
	} else {
		assert.Equal(409, code, "Change user profile status should be 409")
	}
	// case 5: change user2 profile with user2 auth, right parameters format.
	profile.Realname = "test user"
	profile.Email = "testUser0002@vmware.com"
	profile.Comment = "change profile"
	code, err = apiTest.UsersPut(testUser0002ID, profile, *testUser0002Auth)
	if err != nil {
		t.Error("Error occurred while change user profile", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Change user profile status should be 200")
		testUser0002.Email = profile.Email
	}
}

func TestUsersToggleAdminRole(t *testing.T) {
	fmt.Println("Testing Toggle User Admin Role")
	assert := assert.New(t)
	apiTest := newHarborAPI()
	// case 1: toggle user2 admin role without admin auth
	code, err := apiTest.UsersToggleAdminRole(testUser0002ID, *testUser0002Auth, true)
	if err != nil {
		t.Error("Error occurred while toggle user admin role", err.Error())
		t.Log(err)
	} else {
		assert.Equal(403, code, "Toggle user admin role status should be 403")
	}
	// case 2: toggle user2 admin role with admin auth
	code, err = apiTest.UsersToggleAdminRole(testUser0002ID, *admin, true)
	if err != nil {
		t.Error("Error occurred while toggle user admin role", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Toggle user admin role status should be 200")
	}
}

func buildChangeUserPasswordURL(id int) string {
	return fmt.Sprintf("/api/users/%d/password", id)
}

func TestUsersUpdatePassword(t *testing.T) {
	fmt.Println("Testing Update User Password")
	oldPassword := "old_Passw0rd"
	newPassword := "new_Passw0rd"

	user01 := models.User{
		Username: "user01_for_testing_change_password",
		Email:    "user01_for_testing_change_password@test.com",
		Password: oldPassword,
	}
	id, err := dao.Register(user01)
	require.Nil(t, err)
	user01.UserID = int(id)
	defer dao.DeleteUser(user01.UserID)

	user02 := models.User{
		Username: "user02_for_testing_change_password",
		Email:    "user02_for_testing_change_password@test.com",
		Password: oldPassword,
	}
	id, err = dao.Register(user02)
	require.Nil(t, err)
	user02.UserID = int(id)
	defer dao.DeleteUser(user02.UserID)

	cases := []*codeCheckingCase{
		// unauthorized
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    buildChangeUserPasswordURL(user01.UserID),
			},
			code: http.StatusUnauthorized,
		},
		// 404
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    buildChangeUserPasswordURL(10000),
				credential: &usrInfo{
					Name:   user01.Username,
					Passwd: user01.Password,
				},
			},
			code: http.StatusNotFound,
		},
		// 403, a normal user tries to change password of others
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    buildChangeUserPasswordURL(user02.UserID),
				credential: &usrInfo{
					Name:   user01.Username,
					Passwd: user01.Password,
				},
			},
			code: http.StatusForbidden,
		},
		// 400, empty old password
		{
			request: &testingRequest{
				method:   http.MethodPut,
				url:      buildChangeUserPasswordURL(user01.UserID),
				bodyJSON: &passwordReq{},
				credential: &usrInfo{
					Name:   user01.Username,
					Passwd: user01.Password,
				},
			},
			code: http.StatusBadRequest,
		},
		// 400, empty new password
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    buildChangeUserPasswordURL(user01.UserID),
				bodyJSON: &passwordReq{
					OldPassword: oldPassword,
				},
				credential: &usrInfo{
					Name:   user01.Username,
					Passwd: user01.Password,
				},
			},
			code: http.StatusBadRequest,
		},
		// 403, incorrect old password
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    buildChangeUserPasswordURL(user01.UserID),
				bodyJSON: &passwordReq{
					OldPassword: "incorrect_old_password",
					NewPassword: newPassword,
				},
				credential: &usrInfo{
					Name:   user01.Username,
					Passwd: user01.Password,
				},
			},
			code: http.StatusForbidden,
		},
		// 200, normal user change own password
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    buildChangeUserPasswordURL(user01.UserID),
				bodyJSON: &passwordReq{
					OldPassword: oldPassword,
					NewPassword: newPassword,
				},
				credential: &usrInfo{
					Name:   user01.Username,
					Passwd: user01.Password,
				},
			},
			code: http.StatusOK,
		},
		// 400, admin user change password of others.
		// the new password is same with the old one
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    buildChangeUserPasswordURL(user01.UserID),
				bodyJSON: &passwordReq{
					NewPassword: newPassword,
				},
				credential: admin,
			},
			code: http.StatusBadRequest,
		},
		// 200, admin user change password of others
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    buildChangeUserPasswordURL(user01.UserID),
				bodyJSON: &passwordReq{
					NewPassword: "another_new_Passw0rd",
				},
				credential: admin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestUsersDelete(t *testing.T) {

	fmt.Println("Testing User Delete")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	t.Log("delete user-case 1")
	// case 1:delete user without admin auth
	code, err := apiTest.UsersDelete(testUser0002ID, *testUser0003Auth)
	if err != nil {
		t.Error("Error occurred while delete test user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(403, code, "Delete test user status should be 403")
	}
	// case 2: delete user with admin auth, user2 has already been toggled to admin, but can not delete himself
	t.Log("delete user-case 2")
	code, err = apiTest.UsersDelete(testUser0002ID, *testUser0002Auth)
	if err != nil {
		t.Error("Error occurred while delete test user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(403, code, "Delete test user status should be 403")
	}
	// case 3: delete user with admin auth
	t.Log("delete user-case 3")
	code, err = apiTest.UsersDelete(testUser0002ID, *admin)
	if err != nil {
		t.Error("Error occurred while delete test user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Delete test user status should be 200")
	}
	// delete user3 with admin auth
	code, err = apiTest.UsersDelete(testUser0003ID, *admin)
	if err != nil {
		t.Error("Error occurred while delete test user", err.Error())
		t.Log(err)
	} else {
		assert.Equal(200, code, "Delete test user status should be 200")
	}
}

func TestModifiable(t *testing.T) {
	t.Log("Test modifiable.")
	assert := assert.New(t)
	base := BaseController{
		BaseAPI: api.BaseAPI{
			Controller: beego.Controller{},
		},
		SecurityCtx: nil,
		ProjectMgr:  nil,
	}

	ua1 := &UserAPI{
		BaseController:   base,
		currentUserID:    3,
		userID:           4,
		SelfRegistration: false,
		IsAdmin:          false,
		AuthMode:         "db_auth",
	}
	assert.False(ua1.modifiable())
	ua2 := &UserAPI{
		BaseController:   base,
		currentUserID:    3,
		userID:           4,
		SelfRegistration: false,
		IsAdmin:          true,
		AuthMode:         "db_auth",
	}
	assert.True(ua2.modifiable())
	ua3 := &UserAPI{
		BaseController:   base,
		currentUserID:    3,
		userID:           4,
		SelfRegistration: false,
		IsAdmin:          true,
		AuthMode:         "ldap_auth",
	}
	assert.False(ua3.modifiable())
	ua4 := &UserAPI{
		BaseController:   base,
		currentUserID:    1,
		userID:           1,
		SelfRegistration: false,
		IsAdmin:          true,
		AuthMode:         "ldap_auth",
	}
	assert.True(ua4.modifiable())
}

func TestUsersCurrentPermissions(t *testing.T) {
	fmt.Println("Testing Get Users Current Permissions")

	assert := assert.New(t)
	apiTest := newHarborAPI()

	httpStatusCode, permissions, err := apiTest.UsersGetPermissions("current", "/project/1", *projAdmin)
	assert.Nil(err)
	assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	assert.NotEmpty(permissions, "permissions should not be empty")

	httpStatusCode, permissions, err = apiTest.UsersGetPermissions("current", "/unsupport-scope", *projAdmin)
	assert.Nil(err)
	assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	assert.Empty(permissions, "permissions should be empty")

	httpStatusCode, _, err = apiTest.UsersGetPermissions(projAdminID, "/project/1", *projAdmin)
	assert.Nil(err)
	assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")

	httpStatusCode, _, err = apiTest.UsersGetPermissions(projDeveloperID, "/project/1", *projAdmin)
	assert.Nil(err)
	assert.Equal(int(403), httpStatusCode, "httpStatusCode should be 403")
}

func TestValidateSecret(t *testing.T) {
	assert.NotNil(t, validateSecret(""))
	assert.NotNil(t, validateSecret("12345678"))
	assert.NotNil(t, validateSecret("passw0rd"))
	assert.NotNil(t, validateSecret("PASSW0RD"))
	assert.NotNil(t, validateSecret("Sh0rt"))
	assert.Nil(t, validateSecret("Passw0rd"))
	assert.Nil(t, validateSecret("Thisis1Valid_password"))
}
