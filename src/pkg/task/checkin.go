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

package task

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/jobservice/job"
)

var (
	registry = map[string]CheckInProcessor{}
)

// CheckInProcessor is the processor to process the check in data which is sent by jobservice via webhook
type CheckInProcessor func(ctx context.Context, task *Task, change *job.StatusChange) (err error)

// RegisterCheckInProcessor registers check in processor for the specific vendor type
func RegisterCheckInProcessor(vendorType string, processor CheckInProcessor) error {
	if _, exist := registry[vendorType]; exist {
		return fmt.Errorf("check in processor for %s already exists", vendorType)
	}
	registry[vendorType] = processor
	return nil
}
