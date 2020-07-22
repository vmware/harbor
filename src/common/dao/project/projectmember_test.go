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

package project

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	_ "github.com/goharbor/harbor/src/core/auth/db"
	_ "github.com/goharbor/harbor/src/core/auth/ldap"
	cfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
)

func TestMain(m *testing.M) {

	// databases := []string{"mysql", "sqlite"}
	databases := []string{"postgresql"}
	for _, database := range databases {
		log.Infof("run test cases for database: %s", database)

		result := 1
		switch database {
		case "postgresql":
			dao.PrepareTestForPostgresSQL()
		default:
			log.Fatalf("invalid database: %s", database)
		}

		// Extract to test utils
		initSqls := []string{
			"insert into harbor_user (username, email, password, realname)  values ('member_test_01', 'member_test_01@example.com', '123456', 'member_test_01')",
			"insert into project (name, owner_id) values ('member_test_01', 1)",
			"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_group_01', 1, 'CN=harbor_users,OU=sample,OU=vmware,DC=harbor,DC=com')",
			"update project set owner_id = (select user_id from harbor_user where username = 'member_test_01') where name = 'member_test_01'",
			"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select user_id from harbor_user where username = 'member_test_01'), 'u', 1)",
			"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_01') , (select id from user_group where group_name = 'test_group_01'), 'g', 1)",

			"insert into harbor_user (username, email, password, realname)  values ('member_test_02', 'member_test_02@example.com', '123456', 'member_test_02')",
			"insert into project (name, owner_id) values ('member_test_02', 1)",
			"insert into user_group (group_name, group_type, ldap_group_dn) values ('test_group_02', 1, 'CN=harbor_users,OU=sample,OU=vmware,DC=harbor,DC=com')",
			"update project set owner_id = (select user_id from harbor_user where username = 'member_test_02') where name = 'member_test_02'",
			"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_02') , (select user_id from harbor_user where username = 'member_test_02'), 'u', 1)",
			"insert into project_member (project_id, entity_id, entity_type, role) values ( (select project_id from project where name = 'member_test_02') , (select id from user_group where group_name = 'test_group_02'), 'g', 1)",
		}

		clearSqls := []string{
			"delete from project where name='member_test_01' or name='member_test_02'",
			"delete from harbor_user where username='member_test_01' or username='member_test_02' or username='pm_sample'",
			"delete from user_group",
			"delete from project_member",
		}
		dao.PrepareTestData(clearSqls, initSqls)
		cfg.Init()
		result = m.Run()

		if result != 0 {
			os.Exit(result)
		}
	}

}

func TestDeleteProjectMemberByID(t *testing.T) {
	currentProject, err := dao.GetProjectByName("member_test_01")

	if currentProject == nil || err != nil {
		fmt.Println("Failed to load project!")
	} else {
		fmt.Printf("Load project %+v", currentProject)
	}
	var addMember = models.Member{
		ProjectID:  currentProject.ProjectID,
		EntityID:   1,
		EntityType: common.UserMember,
		Role:       common.RoleDeveloper,
	}

	pmid, err := AddProjectMember(addMember)

	if err != nil {
		t.Fatalf("Failed to add project member error: %v", err)
	}

	type args struct {
		pmid int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Delete created", args{pmid}, false},
		{"Delete non exist", args{-13}, false},
		{"Delete non exist", args{13}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteProjectMemberByID(tt.args.pmid); (err != nil) != tt.wantErr {
				t.Errorf("DeleteProjectMemberByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
func TestAddProjectMember(t *testing.T) {

	currentProject, err := dao.GetProjectByName("member_test_01")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	member := models.Member{
		ProjectID:  currentProject.ProjectID,
		EntityID:   1,
		EntityType: common.UserMember,
		Role:       common.RoleProjectAdmin,
	}

	log.Debugf("Current project id %v", currentProject.ProjectID)
	pmid, err := AddProjectMember(member)
	if err != nil {
		t.Errorf("Error occurred in AddProjectMember: %v", err)
	}
	if pmid == 0 {
		t.Errorf("Error add project member, pmid=0")
	}

	queryMember := models.Member{
		ProjectID: currentProject.ProjectID,
		ID:        pmid,
	}

	memberList, err := GetProjectMember(queryMember)
	if err != nil {
		t.Errorf("Failed to query project member, %v, error: %v", queryMember, err)
	}

	if len(memberList) == 0 {
		t.Errorf("Failed to query project member, %v", queryMember)
	}

	_, err = AddProjectMember(models.Member{
		ProjectID:  -1,
		EntityID:   1,
		EntityType: common.UserMember,
		Role:       common.RoleProjectAdmin,
	})
	if err == nil {
		t.Fatal("Should failed with negative projectID")
	}
	_, err = AddProjectMember(models.Member{
		ProjectID:  1,
		EntityID:   -1,
		EntityType: common.UserMember,
		Role:       common.RoleProjectAdmin,
	})
	if err == nil {
		t.Fatal("Should failed with negative entityID")
	}
}
func TestUpdateProjectMemberRole(t *testing.T) {
	currentProject, err := dao.GetProjectByName("member_test_01")
	user := models.User{
		Username: "pm_sample",
		Email:    "pm_sample@example.com",
		Realname: "pm_sample",
		Password: "1234567d",
	}
	o := dao.GetOrmer()
	userID, err := o.Insert(&user)
	if err != nil {
		t.Errorf("Error occurred when add user: %v", err)
	}
	member := models.Member{
		ProjectID:  currentProject.ProjectID,
		EntityID:   int(userID),
		EntityType: common.UserMember,
		Role:       common.RoleProjectAdmin,
	}

	pmid, err := AddProjectMember(member)
	if err != nil {
		t.Errorf("Error occurred in UpdateProjectMember: %v", err)
	}

	UpdateProjectMemberRole(pmid, common.RoleDeveloper)

	queryMember := models.Member{
		ProjectID:  currentProject.ProjectID,
		EntityID:   int(userID),
		EntityType: common.UserMember,
	}

	memberList, err := GetProjectMember(queryMember)
	if err != nil {
		t.Errorf("Error occurred in GetProjectMember: %v", err)
	}
	if len(memberList) != 1 {
		t.Errorf("Error occurred in Failed,  size: %d, condition:%+v", len(memberList), queryMember)
	}
	memberItem := memberList[0]
	if memberItem.Role != common.RoleDeveloper || memberItem.Entityname != user.Username {
		t.Errorf("member doesn't match!")
	}

	memberList2, err := SearchMemberByName(currentProject.ProjectID, "pm_sample")
	if err != nil {
		t.Errorf("Error occurred when SearchMemberByName: %v", err)
	}
	if len(memberList2) == 0 {
		t.Errorf("Failed to search user pm_sample, project_id:%v, entityname:%v",
			currentProject.ProjectID, "pm_sample")
	}

	memberList3, err := SearchMemberByName(currentProject.ProjectID, "")
	if err != nil {
		t.Errorf("Error occurred when SearchMemberByName: %v", err)
	}
	if len(memberList3) == 0 {
		t.Errorf("Failed to search user pm_sample, project_id:%v, entityname is empty",
			currentProject.ProjectID)
	}
}

func TestGetProjectMember(t *testing.T) {
	currentProject, err := dao.GetProjectByName("member_test_01")
	if err != nil {
		t.Errorf("Error occurred when GetProjectByName: %v", err)
	}
	var memberList1 = []*models.Member{
		{
			ID:         346,
			Entityname: "admin",
			Rolename:   "projectAdmin",
			Role:       1,
			EntityID:   1,
			EntityType: "u"},
	}
	var memberList2 = []*models.Member{
		{
			ID:         398,
			Entityname: "test_group_01",
			Rolename:   "projectAdmin",
			Role:       1,
			EntityType: "g"},
	}
	type args struct {
		queryMember models.Member
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.Member
		wantErr bool
	}{
		{"Query default project member", args{models.Member{ProjectID: currentProject.ProjectID, Entityname: "admin"}}, memberList1, false},
		{"Query default project member group", args{models.Member{ProjectID: currentProject.ProjectID, Entityname: "test_group_01"}}, memberList2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetProjectMember(tt.args.queryMember)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectMember() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 1 {
				t.Errorf("Error occurred when query project member")
			}
			itemGot := got[0]
			itemWant := tt.want[0]

			if itemGot.Entityname != itemWant.Entityname || itemGot.Role != itemWant.Role || itemGot.EntityType != itemWant.EntityType {
				t.Errorf("test failed, got:%+v, want:%+v", itemGot, itemWant)
			}
		})
	}

}

func TestGetTotalOfProjectMembers(t *testing.T) {
	currentProject, _ := dao.GetProjectByName("member_test_02")

	type args struct {
		projectID int64
		roles     []int
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{"Get total of project admin", args{currentProject.ProjectID, []int{common.RoleProjectAdmin}}, 2, false},
		{"Get total of maintainer", args{currentProject.ProjectID, []int{common.RoleMaintainer}}, 0, false},
		{"Get total of developer", args{currentProject.ProjectID, []int{common.RoleDeveloper}}, 0, false},
		{"Get total of guest", args{currentProject.ProjectID, []int{common.RoleGuest}}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTotalOfProjectMembers(tt.args.projectID, tt.args.roles...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTotalOfProjectMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTotalOfProjectMembers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListRoles(t *testing.T) {
	// nil user
	roles, err := ListRoles(nil, 1)
	require.Nil(t, err)
	assert.Len(t, roles, 0)

	user, err := dao.GetUser(models.User{Username: "member_test_01"})
	require.Nil(t, err)
	project, err := dao.GetProjectByName("member_test_01")
	require.Nil(t, err)

	// user with empty groups
	roles, err = ListRoles(user, project.ProjectID)
	require.Nil(t, err)
	assert.Len(t, roles, 1)

	// user with a group whose ID doesn't exist
	user.GroupIDs = []int{9999}
	roles, err = ListRoles(user, project.ProjectID)
	require.Nil(t, err)
	require.Len(t, roles, 1)
	assert.Equal(t, common.RoleProjectAdmin, roles[0])

	// user with a valid group
	groupID, err := group.AddUserGroup(models.UserGroup{
		GroupName:   "group_for_list_role",
		GroupType:   1,
		LdapGroupDN: "CN=list_role_users,OU=sample,OU=vmware,DC=harbor,DC=com",
	})
	require.Nil(t, err)
	defer group.DeleteUserGroup(groupID)

	memberID, err := AddProjectMember(models.Member{
		ProjectID:  project.ProjectID,
		Role:       common.RoleDeveloper,
		EntityID:   groupID,
		EntityType: "g",
	})
	require.Nil(t, err)
	defer DeleteProjectMemberByID(memberID)

	user.GroupIDs = []int{groupID}
	roles, err = ListRoles(user, project.ProjectID)
	require.Nil(t, err)
	require.Len(t, roles, 2)
	assert.Equal(t, common.RoleProjectAdmin, roles[0])
	assert.Equal(t, common.RoleDeveloper, roles[1])
}

func PrepareGroupTest() {
	initSqls := []string{
		`insert into user_group (group_name, group_type, ldap_group_dn) values ('harbor_group_01', 1, 'cn=harbor_user,dc=example,dc=com')`,
		`insert into harbor_user (username, email, password, realname) values ('sample01', 'sample01@example.com', 'harbor12345', 'sample01')`,
		`insert into project (name, owner_id) values ('group_project', 1)`,
		`insert into project (name, owner_id) values ('group_project_private', 1)`,
		`insert into project_metadata (project_id, name, value) values ((select project_id from project where name = 'group_project'), 'public', 'false')`,
		`insert into project_metadata (project_id, name, value) values ((select project_id from project where name = 'group_project_private'), 'public', 'false')`,
		`insert into project_member (project_id, entity_id, entity_type, role) values ((select project_id from project where name = 'group_project'), (select id from user_group where group_name = 'harbor_group_01'),'g', 2)`,
	}

	clearSqls := []string{
		`delete from project_metadata where project_id in (select project_id from project where name in ('group_project', 'group_project_private'))`,
		`delete from project where name in ('group_project', 'group_project_private')`,
		`delete from project_member where project_id in (select project_id from project where name in ('group_project', 'group_project_private'))`,
		`delete from user_group where group_name = 'harbor_group_01'`,
		`delete from harbor_user where username = 'sample01'`,
	}
	dao.PrepareTestData(clearSqls, initSqls)
}
