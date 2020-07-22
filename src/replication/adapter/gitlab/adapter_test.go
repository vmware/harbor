package gitlab

import (
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"testing"
)

func mustWriteHTTPResponse(t *testing.T, w io.Writer, fixturePath string) {
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("error opening fixture file: %v", err)
	}

	if _, err = io.Copy(w, f); err != nil {
		t.Fatalf("error writing response: %v", err)
	}
}
func getServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Www-Authenticate", "Bearer realm=\"http://"+r.Host+"/jwt/auth\",service=\"container_registry\"")
		w.WriteHeader(http.StatusUnauthorized)

	})
	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		w.Header().Set("X-Next-Page", "")

		switch search {
		case "dev-docker":
			mustWriteHTTPResponse(t, w, "testdata/projects/dev-docker.json")
			break
		case "dev-":
			mustWriteHTTPResponse(t, w, "testdata/projects/dev-docker.json")
			break
		case "-docker":
			mustWriteHTTPResponse(t, w, "testdata/projects/-docker.json")
			break
		case "":
			mustWriteHTTPResponse(t, w, "testdata/projects/all.json")
			break
		default:
			w.Header().Set("X-Next-Page", "")
			w.Write([]byte(`[]`))
			break
		}

	})
	for projectID := 1; projectID <= 5; projectID++ {
		mux.HandleFunc("/api/v4/projects/"+strconv.Itoa(projectID)+"/registry/repositories", func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("X-Next-Page", "")
			re := regexp.MustCompile(`projects/(?P<id>\d+)/registry`)
			match := re.FindStringSubmatch(r.RequestURI)
			mustWriteHTTPResponse(t, w, "testdata/repositories/"+match[1]+".json")

		})
		for repositoryID := 1; repositoryID <= 5; repositoryID++ {
			mux.HandleFunc("/api/v4/projects/"+strconv.Itoa(projectID)+"/registry/repositories/"+strconv.Itoa(repositoryID)+"1/tags", func(w http.ResponseWriter, r *http.Request) {

				w.Header().Set("X-Next-Page", "")
				re := regexp.MustCompile(`repositories/(?P<id>\d+)/tags`)
				match := re.FindStringSubmatch(r.RequestURI)
				mustWriteHTTPResponse(t, w, "testdata/tags/"+match[1]+".json")

			})

		}
	}
	server := httptest.NewServer(mux)
	return server
}

func getAdapter(t *testing.T) adp.Adapter {
	assertions := assert.New(t)
	factory, err := adp.GetFactory(model.RegistryTypeGitLab)
	assertions.Nil(err)
	assertions.NotNil(factory)
	server := getServer(t)
	adapter := newAdapter(&model.Registry{
		Type: model.RegistryTypeGitLab,
		URL:  server.URL,
		Credential: &model.Credential{
			AccessKey:    "test",
			AccessSecret: "test",
		},
	})
	assertions.Nil(err)
	assertions.NotNil(adapter)

	return adapter
}
func TestFetchImages(t *testing.T) {
	assertions := assert.New(t)

	ad := getAdapter(t)
	adapter := ad.(*adapter)
	templates := map[string]int{
		"library/dev-docker": 1,
		"library/*-docker":   1,
		"library/dev-*":      2,
		"*/dev-docker":       1,
		"library/*":          2,
	}
	for k, v := range templates {
		resources, err := adapter.FetchArtifacts([]*model.Filter{
			{
				Type:  model.FilterTypeName,
				Value: k,
			},
		})
		require.Nil(t, err, k, v)
		assertions.Len(resources, v, k, v)
	}

}
