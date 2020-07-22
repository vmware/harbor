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

package doublestar

import (
	"encoding/json"

	"github.com/bmatcuk/doublestar"
	iselector "github.com/goharbor/harbor/src/lib/selector"
)

const (
	// Kind ...
	Kind = "doublestar"
	// Matches [pattern] for tag (default)
	Matches = "matches"
	// Excludes [pattern] for tag (default)
	Excludes = "excludes"
	// RepoMatches represents repository matches [pattern]
	RepoMatches = "repoMatches"
	// RepoExcludes represents repository excludes [pattern]
	RepoExcludes = "repoExcludes"
	// NSMatches represents namespace matches [pattern]
	NSMatches = "nsMatches"
	// NSExcludes represents namespace excludes [pattern]
	NSExcludes = "nsExcludes"
)

// selector for regular expression
type selector struct {
	// Pre defined pattern declarator
	// "matches", "excludes", "repoMatches" or "repoExcludes"
	decoration string
	// The pattern expression
	pattern string
	// whether match untagged
	untagged bool
}

// Select candidates by regular expressions
func (s *selector) Select(artifacts []*iselector.Candidate) (selected []*iselector.Candidate, err error) {
	value := ""
	excludes := false

	for _, art := range artifacts {
		switch s.decoration {
		case Matches:
			s, err := s.tagSelectMatch(art)
			if err != nil {
				return nil, err
			}
			if s {
				selected = append(selected, art)
			}
		case Excludes:
			s, err := s.tagSelectExclude(art)
			if err != nil {
				return nil, err
			}
			if s {
				selected = append(selected, art)
			}
		case RepoMatches:
			value = art.Repository
		case RepoExcludes:
			value = art.Repository
			excludes = true
		case NSMatches:
			value = art.Namespace
		case NSExcludes:
			value = art.Namespace
			excludes = true
		}

		if len(value) > 0 {
			matched, err := match(s.pattern, value)
			if err != nil {
				// if error occurred, directly throw it out
				return nil, err
			}

			if (matched && !excludes) || (!matched && excludes) {
				selected = append(selected, art)
			}
		}
	}

	return selected, nil
}

func (s *selector) tagSelectMatch(artifact *iselector.Candidate) (selected bool, err error) {
	if len(artifact.Tags) > 0 {
		for _, t := range artifact.Tags {
			matched, err := match(s.pattern, t)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	}
	return s.untagged, nil
}

func (s *selector) tagSelectExclude(artifact *iselector.Candidate) (selected bool, err error) {
	if len(artifact.Tags) > 0 {
		for _, t := range artifact.Tags {
			matched, err := match(s.pattern, t)
			if err != nil {
				return false, err
			}
			if !matched {
				return true, nil
			}
		}
		return false, nil
	}
	return !s.untagged, nil
}

// New is factory method for doublestar selector
func New(decoration string, pattern interface{}, extras string) iselector.Selector {
	untagged := true // default behavior for upgrade, active keep the untagged images
	if decoration == Excludes {
		untagged = false
	}
	if extras != "" {
		var extraObj struct {
			Untagged bool `json:"untagged"`
		}
		if err := json.Unmarshal([]byte(extras), &extraObj); err == nil {
			untagged = extraObj.Untagged
		}
	}

	var p string
	if pattern != nil {
		p, _ = pattern.(string)
	}

	return &selector{
		decoration: decoration,
		pattern:    p,
		untagged:   untagged,
	}
}

// match returns whether the str matches the pattern
func match(pattern, str string) (bool, error) {
	if len(pattern) == 0 {
		return true, nil
	}
	return doublestar.Match(pattern, str)
}
