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
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const testingRedisHost = "REDIS_HOST"

func TestGetRedisPool(t *testing.T) {
	pool, err := GetRedisPool("test", fmt.Sprintf("redis://%s:%d", getRedisHost(), 6379), nil)
	require.Nil(t, err)
	conn := pool.Get()
	defer conn.Close()
}

func getRedisHost() string {
	redisHost := os.Getenv(testingRedisHost)
	if redisHost == "" {
		redisHost = "127.0.0.1" // for local test
	}

	return redisHost
}
