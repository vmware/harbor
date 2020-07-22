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

package dao

import (
	"fmt"
	"sync"
	"testing"

	"github.com/astaxie/beego/orm"
)

func TestMaxOpenConns(t *testing.T) {
	var wg sync.WaitGroup

	queryNum := 200
	results := make([]bool, queryNum)
	for i := 0; i < queryNum; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			o := orm.NewOrm()
			if _, err := o.Raw("SELECT pg_sleep(10)").Exec(); err != nil {
				fmt.Printf("failed to get the count of the projects, error: %v\n", err)
				results[i] = false
			} else {
				results[i] = true
			}
		}(i)
	}

	wg.Wait()

	for _, success := range results {
		if !success {
			t.Fatal("max open conns not work")
		}
	}
}
