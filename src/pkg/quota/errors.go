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

package quota

import (
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/pkg/quota/types"
)

// Errors contains all happened errors
type Errors []error

// GetErrors gets all errors that have occurred and returns a slice of errors (Error type)
func (errs Errors) GetErrors() []error {
	return errs
}

// Add adds an error to a given slice of errors
func (errs Errors) Add(newErrors ...error) Errors {
	for _, err := range newErrors {
		if err == nil {
			continue
		}

		if errors, ok := err.(Errors); ok {
			errs = errs.Add(errors...)
		} else {
			ok = true
			for _, e := range errs {
				if err == e {
					ok = false
				}
			}
			if ok {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

// Error takes a slice of all errors that have occurred and returns it as a formatted string
func (errs Errors) Error() string {
	var errors = []string{}
	for _, e := range errs {
		errors = append(errors, e.Error())
	}
	return strings.Join(errors, "; ")
}

// Exceeded returns exceeded errors from errs
func (errs Errors) Exceeded() error {
	var exceeded Errors
	for _, err := range errs.GetErrors() {
		if _, ok := err.(*ResourceOverflow); ok {
			exceeded = exceeded.Add(err)
		}
	}

	if len(exceeded) == 0 {
		return nil
	}

	return exceeded
}

// ResourceOverflow ...
type ResourceOverflow struct {
	Resource    types.ResourceName
	HardLimit   int64
	CurrentUsed int64
	NewUsed     int64
}

func (e *ResourceOverflow) Error() string {
	resource := e.Resource
	var (
		op    string
		delta int64
	)

	if e.NewUsed > e.CurrentUsed {
		op = "adding"
		delta = e.NewUsed - e.CurrentUsed
	} else {
		op = "subtracting"
		delta = e.CurrentUsed - e.NewUsed
	}

	return fmt.Sprintf("%s %s of %s resource, which when updated to current usage of %s will exceed the configured upper limit of %s.",
		op, resource.FormatValue(delta), resource,
		resource.FormatValue(e.CurrentUsed), resource.FormatValue(e.HardLimit))
}

// NewResourceOverflowError ...
func NewResourceOverflowError(resource types.ResourceName, hardLimit, currentUsed, newUsed int64) error {
	return &ResourceOverflow{Resource: resource, HardLimit: hardLimit, CurrentUsed: currentUsed, NewUsed: newUsed}
}

// ResourceNotFound ...
type ResourceNotFound struct {
	Resource types.ResourceName
}

func (e *ResourceNotFound) Error() string {
	return fmt.Sprintf("resource %s not found", e.Resource)
}

// NewResourceNotFoundError ...
func NewResourceNotFoundError(resource types.ResourceName) error {
	return &ResourceNotFound{Resource: resource}
}
