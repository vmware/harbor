package chartserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	contentTypeHeader = "content-type"
	contentTypeJSON   = "application/json"
)

// Extract error object '{"error": "****---***"}' from the content if existing
// nil error will be returned if it does exist
func extractError(content []byte) (text string, err error) {
	if len(content) == 0 {
		return "", nil
	}

	errorObj := make(map[string]string)
	err = json.Unmarshal(content, &errorObj)
	if err != nil {
		return "", err
	}

	if errText, ok := errorObj["error"]; ok {
		return errText, nil
	}

	return "", nil
}

// Parse the redis configuration to the beego cache pattern
// redis://:password@host:6379/1
// redis+sentinel://anonymous:password@host1:26379,host2:26379/mymaster/1
func parseRedisConfig(redisConfigV string) (map[string]string, error) {
	if len(redisConfigV) == 0 {
		return nil, errors.New("empty redis config")
	}

	redisConfig := make(map[string]string)
	redisConfig["key"] = cacheCollectionName

	if strings.Index(redisConfigV, "//") < 0 {
		redisConfigV = "redis://" + redisConfigV
	}
	u, err := url.Parse(redisConfigV)
	if err != nil {
		return nil, fmt.Errorf("bad _REDIS_URL:%s", redisConfigV)
	}
	if u.Scheme == "redis+sentinel" {
		ps := strings.Split(u.Path, "/")
		if len(ps) < 2 {
			return nil, fmt.Errorf("bad redis sentinel url: no master name, %s", redisConfigV)
		}
		if _, err := strconv.Atoi(ps[1]); err == nil {
			return nil, fmt.Errorf("bad redis sentinel url: master name should not be a number, %s", redisConfigV)
		}
		redisConfig["conn"] = u.Host

		if u.User != nil {
			password, isSet := u.User.Password()
			if isSet {
				redisConfig["password"] = password
			}
		}
		if len(ps) > 2 {
			if _, err := strconv.Atoi(ps[2]); err != nil {
				return nil, fmt.Errorf("bad redis sentinel url: bad db, %s", redisConfigV)
			}
			redisConfig["dbNum"] = ps[2]
		} else {
			redisConfig["dbNum"] = "0"
		}
		redisConfig["masterName"] = ps[1]
	} else if u.Scheme == "redis" {
		redisConfig["conn"] = u.Host // host
		if u.User != nil {
			password, isSet := u.User.Password()
			if isSet {
				redisConfig["password"] = password
			}
		}
		if len(u.Path) > 1 {
			if _, err := strconv.Atoi(u.Path[1:]); err != nil {
				return nil, fmt.Errorf("bad redis url: bad db, %s", redisConfigV)
			}
			redisConfig["dbNum"] = u.Path[1:]
		} else {
			redisConfig["dbNum"] = "0"
		}
	} else {
		return nil, fmt.Errorf("bad redis scheme, %s", redisConfigV)
	}

	return redisConfig, nil
}

// What's the cache driver if it is set
func parseCacheDriver() (string, bool) {
	driver, ok := os.LookupEnv(cacheDriverENVKey)
	return strings.ToLower(driver), ok
}

// Get and parse the configuration for the chart cache
func getCacheConfig() (*ChartCacheConfig, error) {
	driver, isSet := parseCacheDriver()
	if !isSet {
		return nil, nil
	}

	if driver != cacheDriverMem && driver != cacheDriverRedis {
		return nil, fmt.Errorf("cache driver '%s' is not supported, only support 'memory' and 'redis'", driver)
	}

	if driver == cacheDriverMem {
		return &ChartCacheConfig{
			DriverType: driver,
		}, nil
	}

	redisConfigV := os.Getenv(redisENVKey)
	redisCfg, err := parseRedisConfig(redisConfigV)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis configurations from '%s' with error: %s", redisCfg, err)
	}
	if _, isSet := redisCfg["masterName"]; isSet {
		driver = "redis_sentinel"
	}

	// Convert config map to string
	cfgData, err := json.Marshal(redisCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis configurations from '%s' with error: %s", redisCfg, err)
	}

	return &ChartCacheConfig{
		DriverType: driver,
		Config:     string(cfgData),
	}, nil
}
