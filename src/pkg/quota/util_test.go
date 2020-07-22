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
	"testing"

	"github.com/goharbor/harbor/src/pkg/quota/types"
)

func TestIsSafe(t *testing.T) {
	type args struct {
		hardLimits       types.ResourceList
		currentUsed      types.ResourceList
		newUsed          types.ResourceList
		ignoreLimitation bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"unlimited",
			args{
				types.ResourceList{types.ResourceStorage: types.UNLIMITED},
				types.ResourceList{types.ResourceStorage: 1000},
				types.ResourceList{types.ResourceStorage: 1000},
				false,
			},
			false,
		},
		{
			"ok",
			args{
				types.ResourceList{types.ResourceStorage: 100},
				types.ResourceList{types.ResourceStorage: 10},
				types.ResourceList{types.ResourceStorage: 1},
				false,
			},
			false,
		},
		{
			"over the hard limit",
			args{
				types.ResourceList{types.ResourceStorage: 100},
				types.ResourceList{types.ResourceStorage: 0},
				types.ResourceList{types.ResourceStorage: 200},
				false,
			},
			true,
		},
		{
			"ignore limitation",
			args{
				types.ResourceList{types.ResourceStorage: 100},
				types.ResourceList{types.ResourceStorage: 0},
				types.ResourceList{types.ResourceStorage: 200},
				true,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IsSafe(tt.args.hardLimits, tt.args.currentUsed, tt.args.newUsed, tt.args.ignoreLimitation); (err != nil) != tt.wantErr {
				t.Errorf("IsSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrettyPrintResourceNames(t *testing.T) {
	type args struct {
		a []types.ResourceName
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"count", args{[]types.ResourceName{"count"}}, "count"},
		{"count,storage", args{[]types.ResourceName{"count", "storage"}}, "count,storage"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrettyPrintResourceNames(tt.args.a); got != tt.want {
				t.Errorf("PrettyPrintResourceNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
