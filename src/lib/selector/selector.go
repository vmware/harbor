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

package selector

// Selector is used to filter the inputting list
type Selector interface {
	// Select the matched ones
	//
	//  Arguments:
	//    artifacts []*Candidate : candidates for matching
	//
	//  Returns:
	//    []*Candidate : matched candidates
	Select(artifacts []*Candidate) ([]*Candidate, error)
}

// Factory is factory method to return a selector implementation
// Pattern can be any type of data.
// TODO: 'extras' can also be an optional interface{} to accept more complicated data.
type Factory func(decoration string, pattern interface{}, extras string) Selector
