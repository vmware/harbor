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

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"

	"github.com/goharbor/harbor/src/common/dao"
	member "github.com/goharbor/harbor/src/common/dao/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/cmd/helm/search"
	helm_repo "helm.sh/helm/v3/pkg/repo"
)

func TestSearch(t *testing.T) {
	fmt.Println("Testing Search(SearchGet) API")
	// Use mock chart search handler
	searchHandler = func(string, []string) ([]*search.Result, error) {
		results := []*search.Result{}
		results = append(results, &search.Result{
			Name:  "library/harbor",
			Score: 0,
			Chart: &helm_repo.ChartVersion{},
		})

		return results, nil
	}
	// create a public project named "search"
	projectID1, err := dao.AddProject(models.Project{
		Name:    "search",
		OwnerID: int(nonSysAdminID),
	})
	require.Nil(t, err)
	defer dao.DeleteProject(projectID1)

	err = dao.AddProjectMetadata(&models.ProjectMetadata{
		ProjectID: projectID1,
		Name:      "public",
		Value:     "true",
	})
	require.Nil(t, err)

	memberID1, err := member.AddProjectMember(models.Member{
		ProjectID:  projectID1,
		EntityID:   int(nonSysAdminID),
		EntityType: common.UserMember,
		Role:       common.RoleGuest,
	})
	require.Nil(t, err)
	defer member.DeleteProjectMemberByID(memberID1)

	// create a private project named "search-2", the "-" is necessary
	// in the project name to test some corner cases
	projectID2, err := dao.AddProject(models.Project{
		Name:    "search-2",
		OwnerID: int(nonSysAdminID),
	})
	require.Nil(t, err)
	defer dao.DeleteProject(projectID2)

	memberID2, err := member.AddProjectMember(models.Member{
		ProjectID:  projectID2,
		EntityID:   int(nonSysAdminID),
		EntityType: common.UserMember,
		Role:       common.RoleGuest,
	})
	require.Nil(t, err)
	defer member.DeleteProjectMemberByID(memberID2)

	// add a repository in project "search"
	err = dao.AddRepository(models.RepoRecord{
		ProjectID: projectID1,
		Name:      "search/hello-world",
	})
	require.Nil(t, err)

	// add a repository in project "search-2"
	err = dao.AddRepository(models.RepoRecord{
		ProjectID: projectID2,
		Name:      "search-2/hello-world",
	})
	require.Nil(t, err)

	// search without login
	result := &searchResult{}
	err = handleAndParse(&testingRequest{
		method: http.MethodGet,
		url:    "/api/search",
		queryStruct: struct {
			Keyword string `url:"q"`
		}{
			Keyword: "search",
		},
	}, result)
	require.Nil(t, err)
	require.Equal(t, 1, len(result.Project))
	require.Equal(t, 1, len(result.Repository))
	assert.Equal(t, "search", result.Project[0].Name)
	assert.Equal(t, "search/hello-world", result.Repository[0]["repository_name"].(string))

	// search with user who is the member of the project
	err = handleAndParse(&testingRequest{
		method: http.MethodGet,
		url:    "/api/search",
		queryStruct: struct {
			Keyword string `url:"q"`
		}{
			Keyword: "search",
		},
		credential: nonSysAdmin,
	}, result)
	require.Nil(t, err)
	require.Equal(t, 2, len(result.Project))
	require.Equal(t, 2, len(result.Repository))
	projects := map[string]struct{}{}
	repositories := map[string]struct{}{}
	for _, project := range result.Project {
		projects[project.Name] = struct{}{}
	}
	for _, repository := range result.Repository {
		repositories[repository["repository_name"].(string)] = struct{}{}
	}

	_, exist := projects["search"]
	assert.True(t, exist)
	_, exist = projects["search-2"]
	assert.True(t, exist)
	_, exist = repositories["search/hello-world"]
	assert.True(t, exist)
	_, exist = repositories["search-2/hello-world"]
	assert.True(t, exist)

	// search with system admin
	err = handleAndParse(&testingRequest{
		method: http.MethodGet,
		url:    "/api/search",
		queryStruct: struct {
			Keyword string `url:"q"`
		}{
			Keyword: "search",
		},
		credential: sysAdmin,
	}, result)
	require.Nil(t, err)
	require.Equal(t, 2, len(result.Project))
	require.Equal(t, 2, len(result.Repository))
	projects = map[string]struct{}{}
	repositories = map[string]struct{}{}
	for _, project := range result.Project {
		projects[project.Name] = struct{}{}
	}
	for _, repository := range result.Repository {
		repositories[repository["repository_name"].(string)] = struct{}{}
	}
	_, exist = projects["search"]
	assert.True(t, exist)
	_, exist = projects["search-2"]
	assert.True(t, exist)
	_, exist = repositories["search/hello-world"]
	assert.True(t, exist)
	_, exist = repositories["search-2/hello-world"]
	assert.True(t, exist)

	chartSettings := map[string]interface{}{
		common.WithChartMuseum: true,
	}
	config.InitWithSettings(chartSettings)
	defer func() {
		// reset config
		config.Init()
	}()

	// Search chart
	err = handleAndParse(&testingRequest{
		method: http.MethodGet,
		url:    "/api/search",
		queryStruct: struct {
			Keyword string `url:"q"`
		}{
			Keyword: "harbor",
		},
		credential: sysAdmin,
	}, result)
	require.Nil(t, err)
	require.Equal(t, 1, len(*(result.Chart)))
	require.Equal(t, "library/harbor", (*result.Chart)[0].Name)

	// Restore chart search handler
	searchHandler = nil
}
