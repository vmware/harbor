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

package policy

import (
	"sort"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/lib/selector/selectors/doublestar"
	"github.com/goharbor/harbor/src/lib/selector/selectors/label"
	"github.com/goharbor/harbor/src/lib/selector/selectors/severity"
	"github.com/goharbor/harbor/src/lib/selector/selectors/signature"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
)

// Filter defines the filter operations of the preheat policy.
type Filter interface {
	// Build filter from the given policy schema
	BuildFrom(pl *policy.Schema) Filter
	// Filter the inputting candidates and return the matched ones.
	Filter(candidates []*selector.Candidate) ([]*selector.Candidate, error)
}

type defaultFilter struct {
	// all kinds of underlying selectors
	selectors []selector.Selector
	// keep internal error
	error error
}

// NewFilter constructs a filter
func NewFilter() Filter {
	return &defaultFilter{}
}

// Filter candidates
func (df *defaultFilter) Filter(candidates []*selector.Candidate) ([]*selector.Candidate, error) {
	if len(df.selectors) == 0 {
		return nil, errors.New("no underlying filters")
	}

	if df.error != nil {
		// Internal error occurred
		return nil, df.error
	}

	var (
		// At the beginning
		filtered = candidates
		err      error
	)

	// Do filters
	for _, sl := range df.selectors {
		filtered, err = sl.Select(filtered)
		if err != nil {
			return nil, errors.Wrap(err, "do filter error")
		}

		if len(filtered) == 0 {
			// Return earlier
			return filtered, nil
		}
	}

	// Final filtered ones
	return filtered, nil
}

// BuildFrom builds filter from the given policy schema
func (df *defaultFilter) BuildFrom(pl *policy.Schema) Filter {
	if pl != nil && len(pl.Filters) > 0 {
		filters := make([]*policy.Filter, 0)
		// Copy filters and sort the filter list
		for _, fl := range pl.Filters {
			filters = append(filters, fl)
		}
		// Sort
		sort.SliceStable(filters, func(i, j int) bool {
			return filterOrder(filters[i].Type) < filterOrder(filters[j].Type)
		})

		// Build executable selector based on the filter
		if df.selectors == nil || len(df.selectors) > 0 {
			// make or reset
			df.selectors = make([]selector.Selector, 0)
		}

		for _, fl := range filters {
			sl, err := buildFilter(fl)
			if err != nil {
				df.error = errors.Wrap(err, "build filter error")
				// Return earlier
				return df
			}

			df.selectors = append(df.selectors, sl)
		}
	}

	return df
}

// Assign the filter with different order weight and then do filters with fixed order.
func filterOrder(t policy.FilterType) uint {
	switch t {
	case policy.FilterTypeRepository:
		return 0
	case policy.FilterTypeTag:
		return 1
	case policy.FilterTypeLabel:
		return 2
	case policy.FilterTypeSignature:
		return 3
	case policy.FilterTypeVulnerability:
		return 4
	default:
		return 5
	}
}

// buildFilter constructs the selector with the given filter object.
// The filter function leverages the selector lib.
func buildFilter(f *policy.Filter) (selector.Selector, error) {
	if f == nil {
		return nil, errors.New("nil policy filter object")
	}

	// Value should not be nil as all the following filters need pattern data,
	// even the pattern is empty string or zero int (not nil object).
	if f.Value == nil {
		return nil, errors.Errorf("pattern value is missing for filter: %s", f.Type)
	}

	// Check value type
	switch f.Type {
	case policy.FilterTypeRepository,
		policy.FilterTypeTag,
		policy.FilterTypeLabel:
		if _, ok := f.Value.(string); !ok {
			return nil, errors.Errorf("invalid string pattern format for filter: %s", f.Type)
		}
	case policy.FilterTypeSignature:
		if _, ok := f.Value.(bool); !ok {
			return nil, errors.Errorf("invalid boolean pattern format for filter: %s", f.Type)
		}
	case policy.FilterTypeVulnerability:
		if _, ok := f.Value.(int); !ok {
			return nil, errors.Errorf("invalid integer pattern format for filter: %s", f.Type)
		}
	}

	// Build selectors
	switch f.Type {
	case policy.FilterTypeRepository:
		return doublestar.New(doublestar.RepoMatches, f.Value, ""), nil
	case policy.FilterTypeTag:
		return doublestar.New(doublestar.Matches, f.Value, ""), nil
	case policy.FilterTypeLabel:
		return label.New(label.With, f.Value, ""), nil
	case policy.FilterTypeSignature:
		return signature.New(signature.All, f.Value, ""), nil
	case policy.FilterTypeVulnerability:
		return severity.New(severity.Lt, f.Value, ""), nil
	default:
		return nil, errors.Errorf("unknown filter type: %s", f.Type)
	}
}
