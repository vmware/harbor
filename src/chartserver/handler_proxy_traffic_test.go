package chartserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	htesting "github.com/goharbor/harbor/src/testing"
	helm_repo "helm.sh/helm/v3/pkg/repo"
)

// The frontend server
var frontServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	mockController.ProxyTraffic(w, r)
}))

var mockServer *httptest.Server
var mockController *Controller

// Prepare case
func TestStartMockServers(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	mockController = c
	mockServer = s

	frontServer.Start()
}

// Test /health
func TestGetHealthOfBaseHandler(t *testing.T) {
	content, err := httpClient.GetContent(fmt.Sprintf("%s/api/chartrepo/health", frontServer.URL))
	if err != nil {
		t.Fatal(err)
	}

	status := make(map[string]interface{})
	if err := json.Unmarshal(content, &status); err != nil {
		t.Fatalf("Unmarshal error: %s, %s", err, content)
	}
	healthy, ok := status["health"].(bool)
	if !ok || !healthy {
		t.Fatalf("Expect healthy of server to be 'true' but got %v", status["health"])
	}
}

// Get /repo1/index.yaml
func TestGetIndexYamlByRepo(t *testing.T) {
	indexFile, err := getIndexYaml("/chartrepo/repo1/index.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(indexFile.Entries) != 3 {
		t.Fatalf("Expect index file with 3 entries, but got %d", len(indexFile.Entries))
	}
}

// Test download /:repo/charts/chart.tar
// Use this case to test the proxy function
func TestDownloadChart(t *testing.T) {
	content, err := httpClient.GetContent(fmt.Sprintf("%s/chartrepo/repo1/charts/harbor-0.2.0.tgz", frontServer.URL))
	if err != nil {
		t.Fatal(err)
	}

	gotSize := len(content)
	expectSize := len(htesting.HelmChartContent)

	if gotSize != expectSize {
		t.Fatalf("Expect %d bytes data but got %d bytes", expectSize, gotSize)
	}
}

// Get /api/repo1/charts/harbor
// 401 will be rewritten to 500 with specified error
func TestResponseRewrite(t *testing.T) {
	response, err := http.Get(fmt.Sprintf("%s/chartrepo/repo3/charts/harbor-0.8.1.tgz", frontServer.URL))
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expect status code 500 but got %d", response.StatusCode)
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Read bytes from http response failed with error: %s", err)
	}
	defer response.Body.Close()

	errObj := make(map[string]interface{})
	if err = json.Unmarshal(bytes, &errObj); err != nil {
		t.Fatalf("Unmarshal error: %s", err)
	}

	if msg, ok := errObj["error"]; !ok {
		t.Fatal("Expect an error message from server but got nothing")
	} else {
		if !strings.Contains(msg.(string), "operation request from unauthorized source is rejected") {
			t.Fatal("Missing the required error message")
		}
	}
}

// Clear env
func TestStopMockServers(t *testing.T) {
	frontServer.Close()
	mockServer.Close()
}

// Utility method for getting index yaml file
func getIndexYaml(path string) (*helm_repo.IndexFile, error) {
	content, err := httpClient.GetContent(fmt.Sprintf("%s%s", frontServer.URL, path))
	if err != nil {
		return nil, err
	}

	indexFile := &helm_repo.IndexFile{}
	if err := yaml.Unmarshal(content, indexFile); err != nil {
		return nil, fmt.Errorf("Unmarshal error: %s", err)
	}

	if indexFile == nil {
		return nil, fmt.Errorf("Got nil index yaml file")
	}

	return indexFile, nil
}
