// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redis

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/core/config"
	"github.com/gomodule/redigo/redis"
)

var (
	pool     *redis.Pool
	poolOnce sync.Once

	poolMaxIdle                 = 200
	poolMaxActive               = 1000
	poolIdleTimeout       int64 = 180
	dialConnectionTimeout       = 30 * time.Second
	dialReadTimeout             = 10 * time.Second
	dialWriteTimeout            = 10 * time.Second
)

// DefaultPool return default redis pool
func DefaultPool() *redis.Pool {
	poolOnce.Do(func() {
		maxIdle, err := strconv.Atoi(os.Getenv("REDIS_POOL_MAX_IDLE"))
		if err != nil || maxIdle < 0 {
			maxIdle = poolMaxIdle
		}

		maxActive, err := strconv.Atoi(os.Getenv("REDIS_POOL_MAX_ACTIVE"))
		if err != nil || maxActive < 0 {
			maxActive = poolMaxActive
		}

		idleTimeout, err := strconv.ParseInt(os.Getenv("REDIS_POOL_IDLE_TIMEOUT"), 10, 64)
		if err != nil || idleTimeout < 0 {
			idleTimeout = poolIdleTimeout
		}

		url := config.GetRedisOfRegURL()
		if url == "" {
			url = "redis://localhost:6379/1"
		}
		pool, err = GetRedisPool("CommonRedis", url, &PoolParam{
			PoolMaxIdle:           maxIdle,
			PoolMaxActive:         maxActive,
			PoolIdleTimeout:       time.Duration(idleTimeout) * time.Second,
			DialConnectionTimeout: dialConnectionTimeout,
			DialReadTimeout:       dialReadTimeout,
			DialWriteTimeout:      dialWriteTimeout,
		})
	})

	return pool
}
