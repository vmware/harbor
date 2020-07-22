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

// Option option for `Refresh` method of `Controller`
type Option func(*Options)

// Options options used by `Refresh`, `Get`, `List` methods of `Controller`
type Options struct {
	IgnoreLimitation    bool
	WithReferenceObject bool
}

// IgnoreLimitation set IgnoreLimitation for the Options
func IgnoreLimitation(ignoreLimitation bool) func(*Options) {
	return func(opts *Options) {
		opts.IgnoreLimitation = ignoreLimitation
	}
}

// WithReferenceObject set WithReferenceObject to true for the Options
func WithReferenceObject() func(*Options) {
	return func(opts *Options) {
		opts.WithReferenceObject = true
	}
}

func newOptions(options ...Option) *Options {
	opts := &Options{}
	for _, f := range options {
		f(opts)
	}
	return opts
}
