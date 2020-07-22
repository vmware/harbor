package gitlab

import (
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/replication/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestProjects(t *testing.T) {
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodGet,
		Pattern: "/api/v4/projects",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			data := `[

{
        "id": 12312344,
        "description": "",
        "name": "dockers",
        "name_with_namespace": "Library / dockers",
        "path": "dockers",
        "path_with_namespace": "library/dockers",
        "created_at": "2019-01-17T09:47:07.504Z",
        "default_branch": "master",
        "tag_list": [],
    
        "avatar_url": null,
        "star_count": 0,
        "forks_count": 0,
        "last_activity_at": "2019-06-09T15:18:10.045Z",
        "empty_repo": false,
        "archived": false,
        "visibility": "private",
        "resolve_outdated_diff_discussions": false,
        "container_registry_enabled": true,
        "issues_enabled": true,
        "merge_requests_enabled": true,
        "wiki_enabled": true,
        "jobs_enabled": true,
        "snippets_enabled": true,
        "shared_runners_enabled": true,
        "lfs_enabled": true,
        "creator_id": 123412412,
        "forked_from_project": {},
        "import_status": "finished",
        "open_issues_count": 0,
        "ci_default_git_depth": null,
        "public_jobs": true,
        "ci_config_path": null,
        "shared_with_groups": [],
        "only_allow_merge_if_pipeline_succeeds": false,
        "request_access_enabled": false,
        "only_allow_merge_if_all_discussions_are_resolved": false,
        "printing_merge_request_link_enabled": true,
        "merge_method": "merge",
        "external_authorization_classification_label": "",
        "permissions": {
            "project_access": null,
            "group_access": null
        },
        "mirror": false
    }

]`
			w.Header().Set("X-Next-Page", "")
			w.Write([]byte(data))
		},
	})
	client := &Client{
		url:      server.URL,
		username: "test",
		token:    "test",
		client: common_http.NewClient(
			&http.Client{
				Transport: util.GetHTTPTransport(true),
			}),
	}
	projects, e := client.getProjects()
	require.Nil(t, e)
	assert.Equal(t, 1, len(projects))
}
