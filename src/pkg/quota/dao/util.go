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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/quota/types"
)

var (
	quotaOrderMap = map[string]string{
		"creation_time":  "b.creation_time asc",
		"+creation_time": "b.creation_time asc",
		"-creation_time": "b.creation_time desc",
		"update_time":    "b.update_time asc",
		"+update_time":   "b.update_time asc",
		"-update_time":   "b.update_time desc",
	}
)

type listQuery struct {
	ID           int64    `json:"id"`
	Reference    string   `json:"reference"`
	ReferenceID  string   `json:"reference_id"`
	ReferenceIDs []string `json:"reference_ids"`
}

func listConditions(query *q.Query) (string, []interface{}) {
	params := []interface{}{}
	sql := ""
	if query == nil {
		return sql, params
	}

	sql += `WHERE 1=1 `

	var q listQuery

	bytes, err := json.Marshal(query.Keywords)
	if err == nil {
		json.Unmarshal(bytes, &q)
	}

	if q.ID != 0 {
		sql += `AND a.id = ? `
		params = append(params, q.ID)
	}
	if q.Reference != "" {
		sql += `AND a.reference = ? `
		params = append(params, q.Reference)
	}
	if q.ReferenceID != "" {
		sql += `AND a.reference_id = ? `
		params = append(params, q.ReferenceID)
	}

	if len(q.ReferenceIDs) != 0 {
		sql += fmt.Sprintf(`AND a.reference_id IN (%s) `, orm.ParamPlaceholderForIn(len(q.ReferenceIDs)))
		params = append(params, q.ReferenceIDs)
	}

	return sql, params
}

func castQuantity(field string) string {
	// cast -1 to max int64 when order by field
	return fmt.Sprintf("CAST( (CASE WHEN (%[1]s) IS NULL THEN '0' WHEN (%[1]s) = '-1' THEN '9223372036854775807' ELSE (%[1]s) END) AS BIGINT )", field)
}

func listOrderBy(query *q.Query) string {
	orderBy := "b.creation_time DESC"

	if query != nil && query.Sorting != "" {
		if val, ok := quotaOrderMap[query.Sorting]; ok {
			orderBy = val
		} else {
			sort := query.Sorting

			order := "ASC"
			if sort[0] == '-' {
				order = "DESC"
				sort = sort[1:]
			}

			prefixes := []string{"hard.", "used."}
			for _, prefix := range prefixes {
				if strings.HasPrefix(sort, prefix) {
					resource := strings.TrimPrefix(sort, prefix)
					if types.IsValidResource(types.ResourceName(resource)) {
						field := fmt.Sprintf("%s->>'%s'", strings.TrimSuffix(prefix, "."), resource)
						orderBy = fmt.Sprintf("(%s) %s", castQuantity(field), order)
						break
					}
				}
			}
		}
	}

	return orderBy
}
