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

package job

const (
	defaultPriority uint = 1000
)

// PrioritySampler define the job priority generation method
type PrioritySampler interface {
	// Priority for the given job.
	// Job with high priority has the more probabilities to execute.
	// e.g.:
	//   always process X jobs before Y jobs if priorityX > priority Y
	//
	// Arguments:
	//   job string: the job type
	//
	// Returns:
	//   uint: the priority value (between 1 and 10000)
	For(job string) uint
}

// defaultSampler is default implementation of PrioritySampler
type defaultSampler struct{}

// For the given job
func (ps *defaultSampler) For(job string) uint {
	switch job {
	// As an example, sample job has the lowest priority
	case SampleJob:
		return 1
	case SlackJob:
		return 1
		// add more cases here if specified job priority is required
	// case XXX:
	//	return 2000
	default:
		return defaultPriority
	}
}

// Priority returns the default job priority sampler implementation.
func Priority() PrioritySampler {
	return &defaultSampler{}
}
