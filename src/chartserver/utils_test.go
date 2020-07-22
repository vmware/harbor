package chartserver

import (
	"encoding/json"
	"os"
	"testing"
)

// Test the utility function parseRedisConfig
func TestParseRedisConfig(t *testing.T) {
	// Case 1: empty addr
	redisAddr := ""
	if _, err := parseRedisConfig(redisAddr); err == nil {
		t.Fatal("expect non nil error but got nil one if addr is empty")
	}

	// Case 2: short pattern, addr:port
	redisAddr = "redis:6379"
	if parsedConn, err := parseRedisConfig(redisAddr); err != nil {
		t.Fatalf("expect nil error but got non nil one if addr is short pattern: %s\n", parsedConn)
	}

	// Case 3: long pattern but miss some parts
	redisAddr = "redis:6379?idle_timeout_seconds=100"
	if parsedConn, err := parseRedisConfig(redisAddr); err != nil {
		t.Fatalf("expect nil error but got non nil one if addr is long pattern with some parts missing: %v\n", parsedConn)
	} else {
		if num, ok := parsedConn["dbNum"]; !ok || num != "0" {
			t.Fatalf("expect 'dbNum:0' in the parsed conn str: %v\n", parsedConn)
		}
	}

	// Case 4: long pattern
	redisAddr = ":Passw0rd@redis:6379/1?idle_timeout_seconds=100"
	if parsedConn, err := parseRedisConfig(redisAddr); err != nil {
		t.Fatal("expect nil error but got non nil one if addr is long pattern")
	} else {
		if num, ok := parsedConn["dbNum"]; !ok || num != "1" {
			t.Fatalf("expect 'dbNum:1' in the parsed conn str: %v", parsedConn)
		}
		if p, ok := parsedConn["password"]; !ok || p != "Passw0rd" {
			t.Fatalf("expect 'password:Passw0rd' in the parsed conn str: %v", parsedConn)
		}
	}

	// Case 5: sentinel but miss master name
	redisAddr = "redis+sentinel://:Passw0rd@redis1:26379,redis2:26379/1?idle_timeout_seconds=100"
	if _, err := parseRedisConfig(redisAddr); err == nil {
		t.Fatal("expect no master name error but got nil")
	}

	// Case 6: sentinel
	redisAddr = "redis+sentinel://:Passw0rd@redis1:26379,redis2:26379/mymaster/1?idle_timeout_seconds=100"
	if parsedConn, err := parseRedisConfig(redisAddr); err != nil {
		t.Fatal("expect nil error but got non nil one if addr is long pattern")
	} else {
		if num, ok := parsedConn["dbNum"]; !ok || num != "1" {
			t.Fatalf("expect 'dbNum:0' in the parsed conn str: %v", parsedConn)
		}
		if p, ok := parsedConn["password"]; !ok || p != "Passw0rd" {
			t.Fatalf("expect 'password:Passw0rd' in the parsed conn str: %v", parsedConn)
		}
		if v, ok := parsedConn["masterName"]; !ok || v != "mymaster" {
			t.Fatalf("expect 'masterName:mymaster' in the parsed conn str: %v", parsedConn)
		}
		if v, ok := parsedConn["conn"]; !ok || v != "redis1:26379,redis2:26379" {
			t.Fatalf("expect 'conn:redis1:26379,redis2:26379' in the parsed conn str: %v", parsedConn)
		}
	}
}

func TestGetCacheConfig(t *testing.T) {
	// case 1: no cache set
	cacheConf, err := getCacheConfig()
	if err != nil || cacheConf != nil {
		t.Fatal("expect nil cache config and nil error but got non-nil one when parsing empty cache settings")
	}

	// case 2: unknown cache type
	os.Setenv(cacheDriverENVKey, "unknown")
	_, err = getCacheConfig()
	if err == nil {
		t.Fatal("expect non-nil error but got nil one when parsing unknown cache type")
	}

	// case 3: in memory cache type
	os.Setenv(cacheDriverENVKey, cacheDriverMem)
	memCacheConf, err := getCacheConfig()
	if err != nil || memCacheConf == nil || memCacheConf.DriverType != cacheDriverMem {
		t.Fatal("expect in memory cache driver but got invalid one")
	}

	// case 4: wrong redis cache conf
	os.Setenv(cacheDriverENVKey, cacheDriverRedis)
	os.Setenv(redisENVKey, "")
	_, err = getCacheConfig()
	if err == nil {
		t.Fatal("expect non-nil error but got nil one when parsing a invalid redis cache conf")
	}

	// case 5: redis cache conf
	os.Setenv(redisENVKey, ":Passw0rd@redis:6379/1?idle_timeout_seconds=100")
	redisConf, err := getCacheConfig()
	if err != nil {
		t.Fatalf("expect nil error but got non-nil one when parsing valid redis conf")
	}

	if redisConf == nil || redisConf.DriverType != cacheDriverRedis {
		t.Fatal("expect redis cache driver but got invalid one")
	}

	conf := make(map[string]string)
	if err = json.Unmarshal([]byte(redisConf.Config), &conf); err != nil {
		t.Fatal(err)
	}

	if v, ok := conf["conn"]; !ok {
		t.Fatal("expect 'conn' filed in the parsed conf but got nothing")
	} else {
		if v != "redis:6379" {
			t.Fatalf("expect %s but got %s", "redis:6379", v)
		}
	}

	// clear
	os.Unsetenv(cacheDriverENVKey)
	os.Unsetenv(redisENVKey)
}
