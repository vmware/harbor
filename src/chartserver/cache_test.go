package chartserver

import (
	"encoding/json"
	"testing"

	"helm.sh/helm/v3/pkg/chart"
	helm_repo "helm.sh/helm/v3/pkg/repo"
)

var (
	mockChart = &ChartVersionDetails{
		Metadata: &helm_repo.ChartVersion{
			Metadata: &chart.Metadata{
				Name:    "mock_chart",
				Version: "0.1.0",
			},
			Digest: "mock_digest",
		},
		Dependencies: make([]*chart.Dependency, 0),
	}
)

// Test the no cache set scenario
func TestNoCache(t *testing.T) {
	chartCache := NewChartCache(nil)
	if chartCache == nil {
		t.Fatalf("cache instance should not be nil")
	}

	if chartCache.IsEnabled() {
		t.Fatal("chart cache should not be enabled")
	}
}

// Test the in memory cache
func TestInMemoryCache(t *testing.T) {
	chartCache := NewChartCache(&ChartCacheConfig{
		DriverType: cacheDriverMem,
	})
	if chartCache == nil {
		t.Fatalf("cache instance should not be nil")
	}

	if !chartCache.IsEnabled() {
		t.Fatal("chart cache should be enabled")
	}

	if chartCache.driverType != cacheDriverMem {
		t.Fatalf("expect driver type %s but got %s", cacheDriverMem, chartCache.driverType)
	}

	chartCache.PutChart(mockChart)
	theCachedChart := chartCache.GetChart(mockChart.Metadata.Digest)
	if theCachedChart == nil || theCachedChart.Metadata.Name != mockChart.Metadata.Name {
		t.Fatal("In memory cache does work")
	}
}

// Test redis cache
// Failed to config redis cache and then use in memory instead
func TestRedisCache(t *testing.T) {
	redisConfigV := make(map[string]string)
	redisConfigV["key"] = cacheCollectionName
	redisConfigV["conn"] = ":6379"
	redisConfigV["dbNum"] = "0"
	redisConfigV["password"] = ""

	redisConfig, _ := json.Marshal(redisConfigV)

	chartCache := NewChartCache(&ChartCacheConfig{
		DriverType: cacheDriverRedis,
		Config:     string(redisConfig),
	})
	if chartCache == nil {
		t.Fatalf("cache instance should not be nil")
	}

	if !chartCache.IsEnabled() {
		t.Fatal("chart cache should be enabled")
	}

	if chartCache.driverType != cacheDriverRedis {
		t.Fatalf("expect driver type '%s' but got '%s'", cacheDriverRedis, chartCache.driverType)
	}

	chartCache.PutChart(mockChart)
	theCachedChart := chartCache.GetChart(mockChart.Metadata.Digest)
	if theCachedChart == nil || theCachedChart.Metadata.Name != mockChart.Metadata.Name {
		t.Fatal("In memory cache does work")
	}
}
