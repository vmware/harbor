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

package allowlist

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
)

type invalidErr struct {
	msg string
}

func (ie *invalidErr) Error() string {
	return ie.msg
}

// NewInvalidErr ...
func NewInvalidErr(s string) error {
	return &invalidErr{
		msg: s,
	}
}

// IsInvalidErr checks if the error is an invalidErr
func IsInvalidErr(err error) bool {
	_, ok := err.(*invalidErr)
	return ok
}

const cveIDPattern = `^CVE-\d{4}-\d+$`

// Validate help validates the CVE allowlist, to ensure the CVE ID is valid and there's no duplication
func Validate(wl models.CVEAllowlist) error {
	m := map[string]struct{}{}
	//	re := regexp.MustCompile(cveIDPattern)
	for _, it := range wl.Items {
		//	 Bypass the cve format checking
		//		if !re.MatchString(it.CVEID) {
		//			return &invalidErr{fmt.Sprintf("invalid CVE ID: %s", it.CVEID)}
		//		}
		if _, ok := m[it.CVEID]; ok {
			return &invalidErr{fmt.Sprintf("duplicate CVE ID in allowlist: %s", it.CVEID)}
		}
		m[it.CVEID] = struct{}{}
	}
	return nil
}
