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

package severity

import (
	sl "github.com/goharbor/harbor/src/lib/selector"
)

const (
	// Kind of this selector
	Kind = "severity"
	// Gte decoration: Severity of candidate should be greater than or equal to the expected severity.
	Gte = "gte"
	// Gt decoration: Severity of candidate should be greater than the expected severity.
	Gt = "gt"
	// Equal decoration: Severity of candidate should be equal to the expected severity.
	Equal = "equal"
	// Lte decoration: Severity of candidate should be less than or equal to the expected severity.
	Lte = "lte"
	// Lt decoration: Severity of candidate should be less than the expected severity.
	Lt = "lt"
)

// selector filters the candidates by comparing the vulnerability severity
type selector struct {
	// Pre defined pattern decorations
	// "gte", "gt", "equal", "lte" or "lt"
	decoration string

	// expected severity value
	severity uint
}

// Select candidates by comparing the vulnerability severity of the candidate
func (s *selector) Select(artifacts []*sl.Candidate) (selected []*sl.Candidate, err error) {
	for _, a := range artifacts {
		matched := false

		switch s.decoration {
		case Gte:
			if a.VulnerabilitySeverity >= s.severity {
				matched = true
			}
		case Gt:
			if a.VulnerabilitySeverity > s.severity {
				matched = true
			}
		case Equal:
			if a.VulnerabilitySeverity == s.severity {
				matched = true
			}
		case Lte:
			if a.VulnerabilitySeverity <= s.severity {
				matched = true
			}
		case Lt:
			if a.VulnerabilitySeverity < s.severity {
				matched = true
			}
		default:
			break
		}

		if matched {
			selected = append(selected, a)
		}
	}

	return selected, nil
}

// New is factory method for vulnerability severity selector
func New(decoration string, pattern interface{}, extras string) sl.Selector {
	var sev int
	if pattern != nil {
		sev, _ = pattern.(int)
	}

	return &selector{
		decoration: decoration,
		severity:   (uint)(sev),
	}
}
