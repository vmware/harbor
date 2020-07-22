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
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/goharbor/harbor/src/common/dao/project"
	"github.com/goharbor/harbor/src/common/models"
)

func TestProjectMemberAPI_Get(t *testing.T) {
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/projects/1/members",
			},
			code: http.StatusUnauthorized,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/members",
				credential: admin,
			},
			code: http.StatusOK,
		},
		// 400
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/0/members",
				credential: admin,
			},
			code: http.StatusBadRequest,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        fmt.Sprintf("/api/projects/1/members/%d", projAdminPMID),
				credential: admin,
			},
			code: http.StatusOK,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/members/121",
				credential: admin,
			},
			code: http.StatusNotFound,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/99999/members/121",
				credential: admin,
			},
			code: http.StatusNotFound,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestProjectMemberAPI_Post(t *testing.T) {
	userID, err := dao.Register(models.User{
		Username: "restuser",
		Password: "Harbor12345",
		Email:    "restuser@example.com",
	})
	defer dao.DeleteUser(int(userID))
	if err != nil {
		t.Errorf("Error occurred when create user: %v", err)
	}

	ugList, err := group.QueryUserGroup(models.UserGroup{GroupType: 1, LdapGroupDN: "cn=harbor_users,ou=sample,ou=vmware,dc=harbor,dc=com"})
	if err != nil {
		t.Errorf("Failed to query the user group")
	}
	if len(ugList) <= 0 {
		t.Errorf("Failed to query the user group")
	}
	httpUgList, err := group.QueryUserGroup(models.UserGroup{GroupType: 2, GroupName: "vsphere.local\\administrators"})
	if err != nil {
		t.Errorf("Failed to query the user group")
	}
	if len(httpUgList) <= 0 {
		t.Errorf("Failed to query the user group")
	}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/projects/1/members",
				bodyJSON: &models.MemberReq{
					Role: 1,
					MemberUser: models.User{
						UserID: int(userID),
					},
				},
			},
			code: http.StatusUnauthorized,
		},
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/projects/1/members",
				bodyJSON: &models.MemberReq{
					Role: 1,
					MemberUser: models.User{
						UserID: int(userID),
					},
				},
				credential: admin,
			},
			code: http.StatusCreated,
		},
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/projects/1/members",
				bodyJSON: &models.MemberReq{
					Role: 1,
					MemberUser: models.User{
						Username: "notexistuser",
					},
				},
				credential: admin,
			},
			code: http.StatusBadRequest,
		},
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/projects/1/members",
				bodyJSON: &models.MemberReq{
					Role: 1,
					MemberUser: models.User{
						UserID: 0,
					},
				},
				credential: admin,
			},
			code: http.StatusInternalServerError,
		},
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/members?entityname=restuser",
				credential: admin,
			},
			code: http.StatusOK,
		},
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/projects/1/members",
				credential: admin,
			},
			code: http.StatusOK,
		},
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/members",
				credential: admin,
				bodyJSON: &models.MemberReq{
					Role: 1,
					MemberGroup: models.UserGroup{
						GroupType:   1,
						LdapGroupDN: "cn=harbor_users,ou=groups,dc=example,dc=com",
					},
				},
			},
			code: http.StatusInternalServerError,
		},
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/members",
				credential: admin,
				bodyJSON: &models.MemberReq{
					Role: 1,
					MemberGroup: models.UserGroup{
						GroupType: 2,
						ID:        httpUgList[0].ID,
					},
				},
			},
			code: http.StatusCreated,
		},
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/members",
				credential: admin,
				bodyJSON: &models.MemberReq{
					Role: 1,
					MemberGroup: models.UserGroup{
						GroupType: 1,
						ID:        ugList[0].ID,
					},
				},
			},
			code: http.StatusCreated,
		},
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/projects/1/members",
				credential: admin,
				bodyJSON: &models.MemberReq{
					Role: 1,
					MemberGroup: models.UserGroup{
						GroupType: 2,
						GroupName: "vsphere.local/users",
					},
				},
			},
			code: http.StatusInternalServerError,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestProjectMemberAPI_PutAndDelete(t *testing.T) {

	userID, err := dao.Register(models.User{
		Username: "restuser",
		Password: "Harbor12345",
		Email:    "restuser@example.com",
	})
	defer dao.DeleteUser(int(userID))
	if err != nil {
		t.Errorf("Error occurred when create user: %v", err)
	}

	ID, err := project.AddProjectMember(models.Member{
		ProjectID:  1,
		Role:       1,
		EntityID:   int(userID),
		EntityType: "u",
	})
	if err != nil {
		t.Errorf("Error occurred when add project member: %v", err)
	}

	projectID, err := dao.AddProject(models.Project{Name: "memberputanddelete", OwnerID: 1})
	if err != nil {
		t.Errorf("Error occurred when add project: %v", err)
	}
	defer dao.DeleteProject(projectID)

	memberID, err := project.AddProjectMember(models.Member{
		ProjectID:  projectID,
		Role:       1,
		EntityID:   int(userID),
		EntityType: "u",
	})
	if err != nil {
		t.Errorf("Error occurred when add project member: %v", err)
	}

	URL := fmt.Sprintf("/api/projects/1/members/%v", ID)
	badURL := fmt.Sprintf("/api/projects/1/members/%v", 0)
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    URL,
				bodyJSON: &models.Member{
					Role: 2,
				},
			},
			code: http.StatusUnauthorized,
		},
		// 200
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    URL,
				bodyJSON: &models.Member{
					Role: 2,
				},
				credential: admin,
			},
			code: http.StatusOK,
		},
		// 200
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    URL,
				bodyJSON: &models.Member{
					Role: 4,
				},
				credential: admin,
			},
			code: http.StatusOK,
		},
		// 400
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    badURL,
				bodyJSON: &models.Member{
					Role: 2,
				},
				credential: admin,
			},
			code: http.StatusBadRequest,
		},
		// 400
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    URL,
				bodyJSON: &models.Member{
					Role: -2,
				},
				credential: admin,
			},
			code: http.StatusBadRequest,
		},
		// 404
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    fmt.Sprintf("/api/projects/1/members/%v", memberID),
				bodyJSON: &models.Member{
					Role: 2,
				},
				credential: admin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        URL,
				credential: admin,
			},
			code: http.StatusOK,
		},
		// 404
		{
			request: &testingRequest{
				method: http.MethodDelete,
				url:    fmt.Sprintf("/api/projects/1/members/%v", memberID),
				bodyJSON: &models.Member{
					Role: 2,
				},
				credential: admin,
			},
			code: http.StatusNotFound,
		},
	}

	runCodeCheckingCases(t, cases...)

}

func Test_isValidRole(t *testing.T) {
	type args struct {
		role int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"project admin", args{1}, true},
		{"maintainer", args{4}, true},
		{"developer", args{2}, true},
		{"guest", args{3}, true},
		{"limited guest", args{5}, true},
		{"unknow", args{6}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidRole(tt.args.role); got != tt.want {
				t.Errorf("isValidRole() = %v, want %v", got, tt.want)
			}
		})
	}
}
