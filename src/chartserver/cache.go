package chartserver

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	beego_cache "github.com/astaxie/beego/cache"
	hlog "github.com/goharbor/harbor/src/lib/log"

	// Enable redis cache adaptor
	_ "github.com/astaxie/beego/cache/redis"
)

const (
	standardExpireTime       = 3600 * time.Second
	redisENVKey              = "_REDIS_URL_CORE"
	cacheDriverENVKey        = "CHART_CACHE_DRIVER" // "memory" or "redis"
	cacheDriverMem           = "memory"
	cacheDriverRedis         = "redis"
	cacheDriverRedisSentinel = "redis_sentinel"
	cacheCollectionName      = "helm_chart_cache"
	maxTry                   = 10
)

// ChartCache is designed to cache some processed data for repeated accessing
// to improve the performance
type ChartCache struct {
	// Cache driver
	cache beego_cache.Cache

	// Keep the driver type
	driverType string

	// To indicate if the chart cache is enabled
	isEnabled bool
}

// ChartCacheConfig keeps the configurations of ChartCache
type ChartCacheConfig struct {
	// Only support 'in-memory' and 'redis' now
	DriverType string

	// Align with config
	Config string
}

// NewChartCache is constructor of ChartCache
// If return nil, that means no cache is enabled for chart repository server
func NewChartCache(config *ChartCacheConfig) *ChartCache {
	// Never return nil object
	chartCache := &ChartCache{
		isEnabled: false,
	}

	// Double check the configurations are what we want
	if config == nil {
		return chartCache
	}

	if config.DriverType != cacheDriverMem && config.DriverType != cacheDriverRedis {
		return chartCache
	}

	if config.DriverType == cacheDriverRedis {
		if len(config.Config) == 0 {
			return chartCache
		}
	}

	// Try to create the upstream cache
	cache := initCacheDriver(config)
	if cache == nil {
		return chartCache
	}

	// Cache enabled
	chartCache.isEnabled = true
	chartCache.driverType = config.DriverType
	chartCache.cache = cache

	return chartCache
}

// IsEnabled to indicate if the chart cache is successfully enabled
// The cache may be disabled if
//  user does not set
//  wrong configurations
func (chc *ChartCache) IsEnabled() bool {
	return chc.isEnabled
}

// PutChart caches the detailed data of chart version
func (chc *ChartCache) PutChart(chart *ChartVersionDetails) {
	// If cache is not enabled, do nothing
	if !chc.IsEnabled() {
		return
	}

	// As it's a valid json data anymore when retrieving back from redis cache,
	// here we use separate methods to handle the data according to the driver type
	if chart != nil {
		var err error

		switch chc.driverType {
		case cacheDriverMem:
			// Directly put object in
			err = chc.cache.Put(chart.Metadata.Digest, chart, standardExpireTime)
		case cacheDriverRedis:
			// Marshal to json data before saving
			var jsonData []byte
			if jsonData, err = json.Marshal(chart); err == nil {
				err = chc.cache.Put(chart.Metadata.Digest, jsonData, standardExpireTime)
			}
		default:
			// Should not reach here, but still put guard code here
			err = errors.New("Meet invalid cache driver")
		}

		if err != nil {
			// Just logged
			hlog.Errorf("Failed to cache chart object with error: %s\n", err)
			hlog.Warningf("If cache driver is using 'redis', please check the related configurations or the network connection")
		}
	}
}

// GetChart trys to retrieve it from the cache
// If hit, return the cached item;
// otherwise, nil object is returned
func (chc *ChartCache) GetChart(chartDigest string) *ChartVersionDetails {
	// If cache is not enabled, do nothing
	if !chc.IsEnabled() {
		return nil
	}

	object := chc.cache.Get(chartDigest)
	if object != nil {
		// Try to convert data
		// First try the normal way
		if chartDetails, ok := object.(*ChartVersionDetails); ok {
			return chartDetails
		}

		// Maybe json bytes
		if bytes, yes := object.([]byte); yes {
			chartDetails := &ChartVersionDetails{}
			err := json.Unmarshal(bytes, chartDetails)
			if err == nil {
				return chartDetails
			}
			// Just logged the error
			hlog.Errorf("Failed to retrieve chart from cache with error: %s", err)
		}
	}

	return nil
}

// Initialize the cache driver based on the config
func initCacheDriver(cacheConfig *ChartCacheConfig) beego_cache.Cache {
	switch cacheConfig.DriverType {
	case cacheDriverMem:
		hlog.Info("Enable memory cache for chart caching")
		return beego_cache.NewMemoryCache()
	case cacheDriverRedis:
		// New with retry
		count := 0
		for {
			count++
			redisCache, err := beego_cache.NewCache(cacheDriverRedis, cacheConfig.Config)
			if err != nil {
				// Just logged
				hlog.Errorf("Failed to initialize redis cache: %s", err)

				if count < maxTry {
					<-time.After(time.Duration(backoff(count)) * time.Second)
					continue
				}

				return nil
			}

			hlog.Info("Enable redis cache for chart caching")
			return redisCache
		}
	case cacheDriverRedisSentinel:
		// New with retry
		count := 0
		for {
			count++
			redisCache, err := beego_cache.NewCache(cacheDriverRedisSentinel, cacheConfig.Config)
			if err != nil {
				// Just logged
				hlog.Errorf("Failed to initialize redis cache: %s", err)

				if count < maxTry {
					<-time.After(time.Duration(backoff(count)) * time.Second)
					continue
				}

				return nil
			}

			hlog.Info("Enable redis cache for chart caching")
			return redisCache
		}
	default:
		break
	}

	// Any other cases
	hlog.Info("No cache is enabled for chart caching")
	return nil
}

// backoff: fast->slow->fast
func backoff(count int) int {
	f := 5 - math.Abs((float64)(count)-5)
	return (int)(math.Pow(2, f))
}
