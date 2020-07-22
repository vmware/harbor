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

package vuln

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSeverityVersion3(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want Severity
	}{
		{"none", args{"none"}, None},
		{"None", args{"None"}, None},
		{"negligible", args{"negligible"}, None},
		{"Negligible", args{"Negligible"}, None},
		{"low", args{"low"}, Low},
		{"Low", args{"Low"}, Low},
		{"medium", args{"medium"}, Medium},
		{"Medium", args{"Medium"}, Medium},
		{"high", args{"high"}, High},
		{"High", args{"High"}, High},
		{"critical", args{"critical"}, Critical},
		{"Critical", args{"Critical"}, Critical},
		{"invalid", args{"invalid"}, Unknown},
		{"Invalid", args{"Invalid"}, Unknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseSeverityVersion3(tt.args.str); got != tt.want {
				t.Errorf("ParseSeverityVersion3() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCode(t *testing.T) {
	assert.True(t, Critical.Code() > High.Code())
	assert.True(t, High.Code() > Medium.Code())
	assert.True(t, Medium.Code() > Low.Code())
	assert.True(t, Low.Code() > Negligible.Code())
	assert.True(t, Negligible.Code() > Unknown.Code())
	assert.True(t, Unknown.Code() == None.Code())
}
