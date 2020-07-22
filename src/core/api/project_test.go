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
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/testing/apitests/apilib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"strconv"
	"testing"
)

var addProject *apilib.ProjectReq
var addPID int

func addProjectByName(apiTest *testapi, projectName string) (int32, error) {
	req := apilib.ProjectReq{ProjectName: projectName}
	code, err := apiTest.ProjectsPost(*admin, req)
	if err != nil {
		return 0, err
	}
	if code != http.StatusCreated {
		return 0, fmt.Errorf("created failed")
	}

	code, projects, err := apiTest.ProjectsGet(&apilib.ProjectQuery{Name: projectName}, *admin)
	if err != nil {
		return 0, err
	}
	if code != http.StatusOK {
		return 0, fmt.Errorf("get failed")
	}

	if len(projects) == 0 {
		return 0, fmt.Errorf("oops")
	}

	return projects[0].ProjectId, nil
}

func deleteProjectByIDs(apiTest *testapi, projectIDs ...int32) error {
	for _, projectID := range projectIDs {
		_, err := apiTest.ProjectsDelete(*admin, fmt.Sprintf("%d", projectID))
		if err != nil {
			return err
		}
	}

	return nil
}

func InitAddPro() {
	addProject = &apilib.ProjectReq{ProjectName: "add_project", Metadata: map[string]string{models.ProMetaPublic: "true"}}
}

func TestAddProject(t *testing.T) {

	fmt.Println("\nTesting Add Project(ProjectsPost) API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	// prepare for test
	InitAddPro()

	// case 1: admin not login, expect project creation fail.

	result, err := apiTest.ProjectsPost(*unknownUsr, *addProject)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), result, "Case 1: Project creation status should be 401")
		// t.Log(result)
	}

	// case 2: admin successful login, expect project creation success.
	fmt.Println("case 2: admin successful login, expect project creation success.")

	result, err = apiTest.ProjectsPost(*admin, *addProject)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(201), result, "Case 2: Project creation status should be 201")
		// t.Log(result)
	}

	// case 3: duplicate project name, create project fail
	fmt.Println("case 3: duplicate project name, create project fail")

	result, err = apiTest.ProjectsPost(*admin, *addProject)
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(409), result, "Case 3: Project creation status should be 409")
		// t.Log(result)
	}

	// case 4: response code = 400 : Project name is illegal in length
	fmt.Println("case 4 : response code = 400 : Project name is illegal in length ")

	result, err = apiTest.ProjectsPost(*admin, apilib.ProjectReq{ProjectName: "", Metadata: map[string]string{models.ProMetaPublic: "true"}})
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), result, "case 4 : response code = 400 : Project name is illegal in length ")
	}

	// case 5: response code = 201 : expect project creation with quota success.
	fmt.Println("case 5 : response code = 201 : expect project creation with quota success ")

	var countLimit, storageLimit int64
	countLimit, storageLimit = 100, 10
	result, err = apiTest.ProjectsPost(*admin, apilib.ProjectReq{ProjectName: "with_quota", CountLimit: &countLimit, StorageLimit: &storageLimit})
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(201), result, "case 5 : response code = 201 : expect project creation with quota success ")
	}

	// case 6: response code = 400 : bad quota value, create project fail
	fmt.Println("case 6: response code = 400 : bad quota value, create project fail")

	countLimit, storageLimit = 100, -2
	result, err = apiTest.ProjectsPost(*admin, apilib.ProjectReq{ProjectName: "with_quota", CountLimit: &countLimit, StorageLimit: &storageLimit})
	if err != nil {
		t.Error("Error while creat project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), result, "case 6: response code = 400 : bad quota value, create project fail")
	}

	fmt.Printf("\n")

}

func TestListProjects(t *testing.T) {
	fmt.Println("\nTest for Project GET API by project name")
	assert := assert.New(t)

	apiTest := newHarborAPI()
	var result []apilib.Project

	cMockServer, oldCtrl, err := mockChartController()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cMockServer.Close()
		chartController = oldCtrl
	}()

	// ----------------------------case 1 : Response Code=200----------------------------//
	fmt.Println("case 1: response code:200")
	httpStatusCode, result, err := apiTest.ProjectsGet(
		&apilib.ProjectQuery{
			Name:   addProject.ProjectName,
			Owner:  admin.Name,
			Public: true,
		})
	assert.Nil(err)
	assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	assert.Equal(addProject.ProjectName, result[0].ProjectName, "Project name is wrong")
	assert.Equal("true", result[0].Metadata[models.ProMetaPublic], "Public is wrong")

	// find add projectID
	addPID = int(result[0].ProjectId)

	// -------------------case 3 :  check admin project role------------------------//
	httpStatusCode, result, err = apiTest.ProjectsGet(
		&apilib.ProjectQuery{
			Name:   addProject.ProjectName,
			Owner:  admin.Name,
			Public: true,
		}, *admin)
	if err != nil {
		t.Error("Error while search project by proName and isPublic", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal(addProject.ProjectName, result[0].ProjectName, "Project name is wrong")
		assert.Equal("true", result[0].Metadata[models.ProMetaPublic], "Public is wrong")
		assert.Equal(int32(1), result[0].CurrentUserRoleId, "User project role is wrong")
	}

	// -------------------case 4 : add project member and check his role ------------------------//
	CommonAddUser()
	member := &models.MemberReq{
		Role: 2,
		MemberUser: models.User{
			Username: TestUserName,
		},
	}
	projectID := strconv.Itoa(addPID)
	var memberID int
	httpStatusCode, memberID, err = apiTest.AddProjectMember(*admin, projectID, member)
	if err != nil {
		t.Error("Error whihle add project role member", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(201), httpStatusCode, "httpStatusCode should be 201")
	}
	httpStatusCode, result, err = apiTest.ProjectsGet(
		&apilib.ProjectQuery{
			Name: addProject.ProjectName,
		}, *testUser)
	if err != nil {
		t.Error("Error while search project by proName and isPublic", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal(addProject.ProjectName, result[0].ProjectName, "Project name is wrong")
		assert.Equal("true", result[0].Metadata[models.ProMetaPublic], "Public is wrong")
		assert.Equal(int32(2), result[0].CurrentUserRoleId, "User project role is wrong")
	}
	httpStatusCode, err = apiTest.DeleteProjectMember(*admin, projectID, strconv.Itoa(memberID))
	if err != nil {
		t.Error("Error whihle add project role member", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}
	CommonDelUser()
}

// Get project by proID
func TestProGetByID(t *testing.T) {
	fmt.Println("\nTest for Project GET API by project id")
	assert := assert.New(t)

	apiTest := newHarborAPI()
	var result apilib.Project
	projectID := strconv.Itoa(addPID)

	cMockServer, oldCtrl, err := mockChartController()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cMockServer.Close()
		chartController = oldCtrl
	}()

	// ----------------------------case 1 : Response Code=200----------------------------//
	fmt.Println("case 1: response code:200")
	httpStatusCode, result, err := apiTest.ProjectsGetByPID(projectID)
	if err != nil {
		t.Error("Error while search project by proID", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal(addProject.ProjectName, result.ProjectName, "ProjectName is wrong")
		assert.Equal("true", result.Metadata[models.ProMetaPublic], "Public is wrong")
	}
	fmt.Printf("\n")
}
func TestDeleteProject(t *testing.T) {

	fmt.Println("\nTesting Delete Project(ProjectsPost) API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	projectID := strconv.Itoa(addPID)

	// --------------------------case 1: Response Code=401,User need to log in first.-----------------------//
	fmt.Println("case 1: Response Code=401,User need to log in first.")
	httpStatusCode, err := apiTest.ProjectsDelete(*unknownUsr, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "Case 1: Project deletion status should be 401")
	}

	// --------------------------case 2: Response Code=200---------------------------------//
	fmt.Println("case2: response code:200")
	httpStatusCode, err = apiTest.ProjectsDelete(*admin, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "Case 2: Project deletion status should be 200")
	}

	// --------------------------case 3: Response Code=404,Project does not exist---------------------------------//
	fmt.Println("case 3: Response Code=404,Project does not exist")
	projectID = "11"
	httpStatusCode, err = apiTest.ProjectsDelete(*admin, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "Case 3: Project deletion status should be 404")
	}

	// --------------------------case 4: Response Code=400,Invalid project id.---------------------------------//
	fmt.Println("case 4: Response Code=400,Invalid project id.")
	projectID = "cc"
	httpStatusCode, err = apiTest.ProjectsDelete(*admin, projectID)
	if err != nil {
		t.Error("Error while delete project", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(400), httpStatusCode, "Case 4: Project deletion status should be 400")
	}
	fmt.Printf("\n")

}
func TestProHead(t *testing.T) {
	t.Log("\nTest for Project HEAD API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	// ----------------------------case 1 : Response Code=200----------------------------//
	t.Log("case 1: response code:200")
	httpStatusCode, err := apiTest.ProjectsHead(*admin, "library")
	if err != nil {
		t.Error("Error while search project by proName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
	}

	// ----------------------------case 2 : Response Code=404:Project name does not exist.----------------------------//
	t.Log("case 2: response code:404,Project name does not exist.")
	httpStatusCode, err = apiTest.ProjectsHead(*admin, "libra")
	if err != nil {
		t.Error("Error while search project by proName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(404), httpStatusCode, "httpStatusCode should be 404")
	}

	t.Log("case 3: response code:401. Project exist with unauthenticated user")
	httpStatusCode, err = apiTest.ProjectsHead(*unknownUsr, "library")
	if err != nil {
		t.Error("Error while search project by proName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "httpStatusCode should be 404")
	}

	t.Log("case 4: response code:401. Project name does not exist with unauthenticated user")
	httpStatusCode, err = apiTest.ProjectsHead(*unknownUsr, "libra")
	if err != nil {
		t.Error("Error while search project by proName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(401), httpStatusCode, "httpStatusCode should be 404")
	}

	fmt.Printf("\n")
}

func TestPut(t *testing.T) {
	fmt.Println("\nTest for Project PUT API: Update properties for a selected project")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	project := &models.Project{
		Metadata: map[string]string{
			models.ProMetaPublic: "true",
		},
	}

	fmt.Println("case 1: response code:200")
	code, err := apiTest.ProjectsPut(*admin, "1", project)
	require.Nil(t, err)
	assert.Equal(int(200), code)

	fmt.Println("case 2: response code:401, User need to log in first.")
	code, err = apiTest.ProjectsPut(*unknownUsr, "1", project)
	require.Nil(t, err)
	assert.Equal(int(401), code)

	fmt.Println("case 3: response code:400, Invalid project id")
	code, err = apiTest.ProjectsPut(*admin, "cc", project)
	require.Nil(t, err)
	assert.Equal(int(400), code)

	fmt.Println("case 4: response code:404, Not found the project")
	code, err = apiTest.ProjectsPut(*admin, "1234", project)
	require.Nil(t, err)
	assert.Equal(int(404), code)

	fmt.Printf("\n")
}

func TestDeletable(t *testing.T) {
	apiTest := newHarborAPI()
	chServer, oldController, err := mockChartController()
	require.Nil(t, err)
	require.NotNil(t, chServer)
	defer chServer.Close()
	defer func() {
		chartController = oldController
	}()

	project := models.Project{
		Name:    "project_for_test_deletable",
		OwnerID: 1,
	}
	id, err := dao.AddProject(project)
	require.Nil(t, err)

	// non-exist project
	code, del, err := apiTest.ProjectDeletable(*admin, 1000)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotFound, code)

	// unauthorized
	code, del, err = apiTest.ProjectDeletable(*unknownUsr, id)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, code)

	// can be deleted
	code, del, err = apiTest.ProjectDeletable(*admin, id)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.True(t, del)

	err = dao.AddRepository(models.RepoRecord{
		Name:      project.Name + "/golang",
		ProjectID: id,
	})
	require.Nil(t, err)

	// can not be deleted as contains repository
	code, del, err = apiTest.ProjectDeletable(*admin, id)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.False(t, del)
}

func TestProjectSummary(t *testing.T) {
	fmt.Println("\nTest for Project Summary API")
	assert := assert.New(t)

	apiTest := newHarborAPI()

	projectID, err := addProjectByName(apiTest, "project-summary")
	assert.Nil(err)
	defer func() {
		deleteProjectByIDs(apiTest, projectID)
	}()

	// ----------------------------case 1 : Response Code=200----------------------------//
	fmt.Println("case 1: response code:200")
	httpStatusCode, summary, err := apiTest.ProjectSummary(*admin, fmt.Sprintf("%d", projectID))
	if err != nil {
		t.Error("Error while search project by proName", err.Error())
		t.Log(err)
	} else {
		assert.Equal(int(200), httpStatusCode, "httpStatusCode should be 200")
		assert.Equal(int64(1), summary.ProjectAdminCount)
		assert.Equal(map[string]int64{"storage": -1}, summary.Quota.Hard)
	}

	fmt.Printf("\n")
}

func TestHighestRole(t *testing.T) {
	cases := []struct {
		input  []int
		expect int
	}{
		{
			[]int{},
			0,
		},
		{
			[]int{
				common.RoleDeveloper,
				common.RoleMaintainer,
				common.RoleLimitedGuest,
			},
			common.RoleMaintainer,
		},
		{
			[]int{
				common.RoleProjectAdmin,
				common.RoleMaintainer,
				common.RoleMaintainer,
			},
			common.RoleProjectAdmin,
		},
		{
			[]int{
				99,
				33,
				common.RoleLimitedGuest,
			},
			common.RoleLimitedGuest,
		},
		{
			[]int{
				99,
				99,
				99,
			},
			0,
		},
		{
			nil,
			0,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expect, highestRole(c.input))
	}
}
