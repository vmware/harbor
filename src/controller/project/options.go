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

package project

// Option option for `Get` and `Exist` method of `Controller`
type Option func(*Options)

// Options options used by `Get` method of `Controller`
type Options struct {
	CVEAllowlist bool // get project with cve allowlist
	Metadata     bool // get project with metadata
}

// CVEAllowlist set CVEAllowlist for the Options
func CVEAllowlist(allowlist bool) Option {
	return func(opts *Options) {
		opts.CVEAllowlist = allowlist
	}
}

// Metadata set Metadata for the Options
func Metadata(metadata bool) Option {
	return func(opts *Options) {
		opts.Metadata = metadata
	}
}

func newOptions(options ...Option) *Options {
	opts := &Options{
		Metadata: true, // default get project with metadata
	}

	for _, f := range options {
		f(opts)
	}

	return opts
}
