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
	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/lib/selector/selectors/doublestar"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/index"
)

const (
	// AlgorithmOR for OR algorithm
	AlgorithmOR = "or"

	// TriggerKindSchedule Schedule
	TriggerKindSchedule = "Schedule"

	// TriggerReferencesJobid job_id
	TriggerReferencesJobid = "job_id"
	// TriggerSettingsCron cron
	TriggerSettingsCron = "cron"

	// ScopeLevelProject project
	ScopeLevelProject = "project"
)

// Metadata of policy
type Metadata struct {
	// ID of the policy
	ID int64 `json:"id"`

	// Algorithm applied to the rules
	// "OR" / "AND"
	Algorithm string `json:"algorithm" valid:"Required;Match(or)"`

	// Rule collection
	Rules []rule.Metadata `json:"rules"`

	// Trigger about how to launch the policy
	Trigger *Trigger `json:"trigger" valid:"Required"`

	// Which scope the policy will be applied to
	Scope *Scope `json:"scope" valid:"Required"`
}

// Valid Valid
func (m *Metadata) Valid(v *validation.Validation) {
	if m.Trigger == nil {
		_ = v.SetError("Trigger", "Can not be empty")
		return
	}
	if m.Scope == nil {
		_ = v.SetError("Scope", "Can not be empty")
		return
	}
	if m.Trigger != nil && m.Trigger.Kind == TriggerKindSchedule {
		if m.Trigger.Settings == nil {
			_ = v.SetError("Trigger.Settings", "Can not be empty")
		} else {
			if _, ok := m.Trigger.Settings[TriggerSettingsCron]; !ok {
				_ = v.SetError("Trigger.Settings.cron", "Can not be empty")
			}
		}
	}
	if !v.HasErrors() {
		for _, r := range m.Rules {
			if err := index.Valid(r.Template, r.Parameters); err != nil {
				_ = v.SetError("Parameters", err.Error())
				return
			}
			if ok, _ := v.Valid(&r); !ok {
				return
			}
		}
	}
}

// Trigger of the policy
type Trigger struct {
	// Const string to declare the trigger type
	// 'Schedule'
	Kind string `json:"kind" valid:"Required"`

	// Settings for the specified trigger
	// '[cron]="* 22 11 * * *"' for the 'Schedule'
	Settings map[string]interface{} `json:"settings" valid:"Required"`

	// References of the trigger
	// e.g: schedule job ID
	References map[string]interface{} `json:"references"`
}

// Scope definition
type Scope struct {
	// Scope level declaration
	// 'system', 'project' and 'repository'
	Level string `json:"level" valid:"Required;Match(/^(project)$/)"`

	// The reference identity for the specified level
	// 0 for 'system', project ID for 'project' and repo ID for 'repository'
	Reference int64 `json:"ref" valid:"Required"`
}

// WithNDaysSinceLastPull build a retention rule to keep images n days to since last pull
func WithNDaysSinceLastPull(projID int64, n int) *Metadata {
	return &Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Action:   "retain",
				Template: "nDaysSinceLastPull",
				TagSelectors: []*rule.Selector{
					{
						Kind:       doublestar.Kind,
						Decoration: doublestar.Matches,
						Pattern:    "**",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       doublestar.Kind,
							Decoration: doublestar.RepoMatches,
							Pattern:    "**",
						},
					},
				},
				Parameters: rule.Parameters{
					"nDaysSinceLastPull": n,
				},
			},
		},
		Trigger: &Trigger{
			Kind: "Schedule",
			Settings: map[string]interface{}{
				"cron": "0 0 0 * * *",
			},
		},
		Scope: &Scope{
			Level:     "project",
			Reference: projID,
		},
	}
}
