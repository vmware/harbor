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

package signature

import (
	sl "github.com/goharbor/harbor/src/lib/selector"
)

const (
	// Kind of this selector
	Kind = "signature"
	// Any tag of the artifact candidate is signed
	Any = "any"
	// All the tags of the artifact candidate are signed
	All = "all"
)

// selector filters the candidates by signing status (signature)
type selector struct {
	// Pre defined pattern decorations
	// "any" or "all"
	decoration string

	// expected status of signing
	expected bool
}

// Select candidates by the signing status of the candidate
func (s *selector) Select(artifacts []*sl.Candidate) (selected []*sl.Candidate, err error) {
	for _, a := range artifacts {
		matched := 0
		for _, t := range a.Tags {
			if a.Signatures[t] == s.expected {
				matched++
				if s.decoration == Any {
					break
				}
			} else {
				if s.decoration == All {
					break
				}
			}
		}

		if (s.decoration == Any && matched > 0) ||
			(s.decoration == All && matched == len(a.Tags)) {
			selected = append(selected, a)
		}
	}

	return selected, nil
}

// New is factory method for signature selector
func New(decoration string, pattern interface{}, extras string) sl.Selector {
	var e bool
	if pattern != nil {
		e, _ = pattern.(bool)
	}

	return &selector{
		decoration: decoration,
		expected:   e,
	}
}
