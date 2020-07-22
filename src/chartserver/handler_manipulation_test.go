package chartserver

import (
	"testing"
)

// Test get /api/:repo/charts/harbor
func TestGetChart(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	versions, err := c.GetChart("repo1", "harbor")
	if err != nil {
		t.Fatal(err)
	}

	if len(versions) != 2 {
		t.Fatalf("expect 2 chart versions of harbor but got %d", len(versions))
	}
}

// Test delete /api/:repo/charts/harbor/0.2.0
func TestDeleteChartVersion(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := c.DeleteChartVersion("repo1", "harbor", "0.2.0"); err != nil {
		t.Fatal(err)
	}
}

// Test get /api/:repo/charts
func TestRetrieveChartList(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	chartList, err := c.ListCharts("repo1")
	if err != nil {
		t.Fatal(err)
	}

	if len(chartList) != 2 {
		t.Fatalf("Expect to get 2 charts in the list but got %d", len(chartList))
	}

	foundItem := false
	for _, chartInfo := range chartList {
		if chartInfo.Name == "hello-helm" && chartInfo.TotalVersions == 2 {
			foundItem = true
			break
		}
	}

	if !foundItem {
		t.Fatalf("Expect chart 'hello-helm' with 2 versions but got nothing")
	}
}

// Test the GetChartVersion in utility handler
func TestGetChartVersionSummary(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	chartV, err := c.GetChartVersion("repo1", "harbor", "0.2.0")
	if err != nil {
		t.Fatal(err)
	}

	if chartV.Name != "harbor" {
		t.Fatalf("expect chart name 'harbor' but got '%s'", chartV.Name)
	}

	if chartV.Version != "0.2.0" {
		t.Fatalf("expect chart version '0.2.0' but got '%s'", chartV.Version)
	}
}
