package chartserver

import (
	"testing"
)

// Test the function GetCountOfCharts
func TestGetCountOfCharts(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	count, err := c.GetCountOfCharts([]string{})
	if err != nil {
		t.Fatalf("expect nil error but got %s", err)
	}
	if count != 0 {
		t.Fatalf("expect 0 but got %d", count)
	}

	namespaces := []string{"repo1", "repo2"}
	count, err = c.GetCountOfCharts(namespaces)
	if err != nil {
		t.Fatalf("expect nil error but got %s", err)
	}

	if count != 5 {
		t.Fatalf("expect 5 but got %d", count)
	}

	_, err = c.GetCountOfCharts([]string{"not-existing-ns"})
	if err == nil {
		t.Fatal("expect non-nil error but got nil one")
	}
}

// Test the function DeleteChart
func TestDeleteChart(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := c.DeleteChart("repo1", "harbor"); err != nil {
		t.Fatal(err)
	}
}

// Test get /api/:repo/charts/:chart_name/:version
func TestGetChartVersion(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	chartVersion, err := c.GetChartVersionDetails("repo1", "harbor", "0.2.0")
	if err != nil {
		t.Fatal(err)
	}

	if chartVersion.Metadata.Name != "harbor" {
		t.Fatalf("Expect harbor chart version but got %s", chartVersion.Metadata.Name)
	}

	if chartVersion.Metadata.Version != "0.2.0" {
		t.Fatalf("Expect version '0.2.0' but got version %s", chartVersion.Metadata.Version)
	}

	if len(chartVersion.Dependencies) != 1 {
		t.Fatalf("Expect 1 dependency but got %d ones", len(chartVersion.Dependencies))
	}

	if len(chartVersion.Values) != 99 {
		t.Fatalf("Expect 99 k-v values but got %d", len(chartVersion.Values))
	}
}

// Test get /api/:repo/charts/:chart_name/:version with none-existing version
func TestGetChartVersionWithError(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	_, err = c.GetChartVersionDetails("repo1", "harbor", "1.0.0")
	if err == nil {
		t.Fatal("Expect an error but got nil")
	}
}

// Test the chart searching
func TestChartSearching(t *testing.T) {
	s, c, err := createMockObjects()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	namespaces := []string{"repo1", "repo2"}
	q := "harbor"

	results, err := c.SearchChart(q, namespaces)
	if err != nil {
		t.Fatalf("expect nil error but got '%s'", err)
	}

	if len(results) != 2 {
		t.Fatalf("expect 2 results but got %d", len(results))
	}
}
