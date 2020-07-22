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

package types

import "testing"

func Test_byteCountToDisplaySize(t *testing.T) {
	type args struct {
		value int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"100 B", args{100}, "100 B"},
		{"1.0 KiB", args{1024}, "1.0 KiB"},
		{"1.5 KiB", args{1024 * 3 / 2}, "1.5 KiB"},
		{"1.0 MiB", args{1024 * 1024}, "1.0 MiB"},
		{"1.5 MiB", args{1024 * 1024 * 3 / 2}, "1.5 MiB"},
		{"1.0 GiB", args{1024 * 1024 * 1024}, "1.0 GiB"},
		{"1.5 GiB", args{1024 * 1024 * 1024 * 3 / 2}, "1.5 GiB"},
		{"1.0 TiB", args{1024 * 1024 * 1024 * 1024}, "1.0 TiB"},
		{"1.5 TiB", args{1024 * 1024 * 1024 * 1024 * 3 / 2}, "1.5 TiB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := byteCountToDisplaySize(tt.args.value); got != tt.want {
				t.Errorf("byteCountToDisplaySize() = %v, want %v", got, tt.want)
			}
		})
	}
}
