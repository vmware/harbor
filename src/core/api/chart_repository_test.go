package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/promgr/metamgr"
)

var (
	crOldController *chartserver.Controller
	crMockServer    *httptest.Server
)

func TestIsMultipartFormData(t *testing.T) {
	req, err := createRequest(http.MethodPost, "/api/chartrepo/charts")
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(headerContentType, "application/json")
	if isMultipartFormData(req) {
		t.Fatal("expect false result but got true")
	}

	req.Header.Set(headerContentType, contentTypeMultipart)
	if !isMultipartFormData(req) {
		t.Fatalf("expect %s result but got %s", contentTypeMultipart, req.Header.Get(headerContentType))
	}
}

// Test namespace cheking
func TestRequireNamespace(t *testing.T) {
	chartAPI := &ChartRepositoryAPI{}
	chartAPI.ProjectMgr = &mockProjectManager{}

	if !chartAPI.requireNamespace("library") {
		t.Fatal("expect namespace 'library' existing but got false")
	}
}

// Prepare
func TestPrepareEnv(t *testing.T) {
	var err error
	crMockServer, crOldController, err = mockChartController()
	if err != nil {
		t.Fatalf("Failed to start mock chart service with error: %s", err)
	}
}

// Test get health
func TestGetHealthStatus(t *testing.T) {
	status := make(map[string]interface{})
	err := handleAndParse(&testingRequest{
		url:        "/api/chartrepo/health",
		method:     http.MethodGet,
		credential: sysAdmin,
	}, &status)

	if err != nil {
		t.Fatal(err)
	}

	if _, ok := status["health"]; !ok {
		t.Fatal("expect 'health' but got nil")
	}
}

// Test get index by repo
func TestGetIndexByRepo(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:        "/chartrepo/library/index.yaml",
			method:     http.MethodGet,
			credential: projDeveloper,
		},
		code: http.StatusOK,
	})
}

// Test get index
func TestGetIndex(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:        "/chartrepo/index.yaml",
			method:     http.MethodGet,
			credential: sysAdmin,
		},
		code: http.StatusOK,
	})
}

// Test download chart
func TestDownloadChart(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:        "/chartrepo/library/charts/harbor-0.2.0.tgz",
			method:     http.MethodGet,
			credential: projDeveloper,
		},
		code: http.StatusOK,
	})
}

// Test get charts
func TesListCharts(t *testing.T) {
	charts := make([]*chartserver.ChartInfo, 0)
	err := handleAndParse(&testingRequest{
		url:        "/api/chartrepo/library/charts",
		method:     http.MethodGet,
		credential: projAdmin,
	}, &charts)

	if err != nil {
		t.Fatal(err)
	}

	if len(charts) != 2 {
		t.Fatalf("expect 2 charts but got %d", len(charts))
	}
}

// Test get chart versions
func TestListChartVersions(t *testing.T) {
	chartVersions := make(chartserver.ChartVersions, 0)
	err := handleAndParse(&testingRequest{
		url:        "/api/chartrepo/library/charts/harbor",
		method:     http.MethodGet,
		credential: projAdmin,
	}, &chartVersions)

	if err != nil {
		t.Fatal(err)
	}

	if len(chartVersions) != 2 {
		t.Fatalf("expect 2 chart versions but got %d", len(chartVersions))
	}
}

// Test get chart version details
func TestGetChartVersion(t *testing.T) {
	chartV := &chartserver.ChartVersionDetails{}
	err := handleAndParse(&testingRequest{
		url:        "/api/chartrepo/library/charts/harbor/0.2.0",
		method:     http.MethodGet,
		credential: projAdmin,
	}, chartV)

	if err != nil {
		t.Fatal(err)
	}

	if chartV.Metadata.Name != "harbor" {
		t.Fatalf("expect get chart 'harbor' but got %s", chartV.Metadata.Name)
	}

	if chartV.Metadata.Version != "0.2.0" {
		t.Fatalf("expect get chart version '0.2.0' but got %s", chartV.Metadata.Version)
	}
}

// Test delete chart version
func TestDeleteChartVersion(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:        "/api/chartrepo/library/charts/harbor/0.2.1",
			method:     http.MethodDelete,
			credential: projAdmin,
		},
		code: http.StatusOK,
	})
}

// Test delete chart
func TestDeleteChart(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:        "/api/chartrepo/library/charts/harbor",
			method:     http.MethodDelete,
			credential: projAdmin,
		},
		code: http.StatusOK,
	})
}

// Clear
func TestClearEnv(t *testing.T) {
	crMockServer.Close()
	chartController = crOldController
}

func createRequest(method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.RequestURI = url

	return req, nil
}

// Mock project manager
type mockProjectManager struct{}

func (mpm *mockProjectManager) Get(projectIDOrName interface{}) (*models.Project, error) {
	return nil, errors.New("Not implemented")
}

func (mpm *mockProjectManager) Create(*models.Project) (int64, error) {
	return -1, errors.New("Not implemented")
}

func (mpm *mockProjectManager) Delete(projectIDOrName interface{}) error {
	return errors.New("Not implemented")
}

func (mpm *mockProjectManager) Update(projectIDOrName interface{}, project *models.Project) error {
	return errors.New("Not implemented")
}
func (mpm *mockProjectManager) List(query *models.ProjectQueryParam) (*models.ProjectQueryResult, error) {
	results := &models.ProjectQueryResult{
		Total:    2,
		Projects: make([]*models.Project, 0),
	}

	results.Projects = append(results.Projects, &models.Project{ProjectID: 0, Name: "library"})
	results.Projects = append(results.Projects, &models.Project{ProjectID: 1, Name: "repo2"})

	return results, nil
}

func (mpm *mockProjectManager) IsPublic(projectIDOrName interface{}) (bool, error) {
	return false, errors.New("Not implemented")
}

func (mpm *mockProjectManager) Exists(projectIDOrName interface{}) (bool, error) {
	if projectIDOrName == nil {
		return false, errors.New("nil projectIDOrName")
	}

	if ns, ok := projectIDOrName.(string); ok {
		if ns == "library" {
			return true, nil
		}

		return false, nil
	}

	return false, errors.New("unknown type of projectIDOrName")
}

// get all public project
func (mpm *mockProjectManager) GetPublic() ([]*models.Project, error) {
	return nil, errors.New("Not implemented")
}

func (mpm *mockProjectManager) GetAuthorized(user *models.User) ([]*models.Project, error) {
	return nil, nil
}

// if the project manager uses a metadata manager, return it, otherwise return nil
func (mpm *mockProjectManager) GetMetadataManager() metamgr.ProjectMetadataManager {
	return nil
}
