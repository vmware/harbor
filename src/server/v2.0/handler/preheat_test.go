package handler

import (
	"reflect"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/task"

	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	instanceModel "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/stretchr/testify/assert"
)

func Test_convertProvidersToFrontend(t *testing.T) {
	backend, _ := provider.ListProviders()
	tests := []struct {
		name         string
		backend      []*provider.Metadata
		wantFrontend []*models.Metadata
	}{
		{"",
			backend,
			[]*models.Metadata{
				{ID: "dragonfly", Icon: "https://raw.githubusercontent.com/alibaba/Dragonfly/master/docs/images/logo.png", Maintainers: []string{"Jin Zhang/taiyun.zj@alibaba-inc.com"}, Name: "Dragonfly", Source: "https://github.com/alibaba/Dragonfly", Version: "0.10.1"},
				{Icon: "https://github.com/uber/kraken/blob/master/assets/kraken-logo-color.svg", ID: "kraken", Maintainers: []string{"mmpei/peimingming@corp.netease.com"}, Name: "Kraken", Source: "https://github.com/uber/kraken", Version: "0.1.3"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFrontend := convertProvidersToFrontend(tt.backend); !reflect.DeepEqual(gotFrontend, tt.wantFrontend) {
				t.Errorf("convertProvidersToFrontend() = %#v, want %#v", gotFrontend, tt.wantFrontend)
			}
		})
	}
}

func Test_convertPolicyToPayload(t *testing.T) {
	tests := []struct {
		name      string
		input     *policy.Schema
		expect    *models.PreheatPolicy
		shouldErr bool
	}{
		{
			name:      "should error",
			input:     nil,
			expect:    nil,
			shouldErr: true,
		},
		{
			name: "should success",
			input: &policy.Schema{
				ID:          0,
				Name:        "abc",
				Description: "test case",
				ProjectID:   0,
				ProviderID:  0,
				Filters:     nil,
				FiltersStr:  "",
				Trigger:     nil,
				TriggerStr:  "",
				Enabled:     false,
				CreatedAt:   time.Time{},
				UpdatedTime: time.Time{},
			},
			expect: &models.PreheatPolicy{
				CreationTime: strfmt.DateTime{},
				Description:  "test case",
				Enabled:      false,
				Filters:      "",
				ID:           0,
				Name:         "abc",
				ProjectID:    0,
				ProviderID:   0,
				Trigger:      "",
				UpdateTime:   strfmt.DateTime{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := convertPolicyToPayload(tt.input)
			if !tt.shouldErr {
				if !assert.Equal(t, tt.expect, actual) {
					t.Errorf("convertPolicyToPayload() = %#v, want %#v", actual, tt.expect)
				}
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func Test_convertParamPolicyToModelPolicy(t *testing.T) {
	tests := []struct {
		name      string
		input     *models.PreheatPolicy
		expect    *policy.Schema
		shouldErr bool
	}{
		{
			name:      "should err",
			input:     nil,
			expect:    nil,
			shouldErr: true,
		},
		{
			name:      "invalid name",
			input:     &models.PreheatPolicy{Name: "abc/-.**"},
			expect:    nil,
			shouldErr: true,
		},
		{
			name: "should success",
			input: &models.PreheatPolicy{
				CreationTime: strfmt.DateTime{},
				Description:  "test case",
				Enabled:      false,
				Filters:      "",
				ID:           0,
				Name:         "abc",
				ProjectID:    0,
				ProviderID:   0,
				Trigger:      "",
				UpdateTime:   strfmt.DateTime{},
			},
			expect: &policy.Schema{
				ID:          0,
				Name:        "abc",
				Description: "test case",
				ProjectID:   0,
				ProviderID:  0,
				Filters:     nil,
				FiltersStr:  "",
				Trigger:     nil,
				TriggerStr:  "",
				Enabled:     false,
				CreatedAt:   time.Time{},
				UpdatedTime: time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := convertParamPolicyToModelPolicy(tt.input)
			if !tt.shouldErr {
				if !assert.Equal(t, tt.expect, actual) {
					t.Errorf("convertParamPolicyToModelPolicy() = %#v, want %#v", actual, tt.expect)
				}
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func Test_convertInstanceToPayload(t *testing.T) {
	tests := []struct {
		name    string
		input   *instanceModel.Instance
		want    *models.Instance
		wantErr bool
	}{
		{
			name:    "want err",
			input:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid authData",
			input:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name: "want ok",
			input: &instanceModel.Instance{
				ID:             1,
				Name:           "abc",
				Description:    "something",
				Vendor:         "kraken",
				Enabled:        true,
				Endpoint:       "https://example.com",
				AuthMode:       "none",
				AuthData:       `{"name":"harbor"}`,
				Status:         "",
				Default:        true,
				Insecure:       true,
				SetupTimestamp: 1234,
			},
			want: &models.Instance{
				ID:             1,
				Name:           "abc",
				Description:    "something",
				Vendor:         "kraken",
				Enabled:        true,
				Endpoint:       "https://example.com",
				AuthMode:       "none",
				AuthInfo:       map[string]string{"name": "harbor"},
				Status:         "Unknown",
				Default:        true,
				Insecure:       true,
				SetupTimestamp: 1234,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertInstanceToPayload(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertInstanceToPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertInstanceToPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertParamInstanceToModelInstance(t *testing.T) {
	tests := []struct {
		name    string
		input   *models.Instance
		want    *instanceModel.Instance
		wantErr bool
	}{
		{
			name:    "want err",
			input:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid name",
			input:   &models.Instance{Name: "_aa/*b"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "want ok",
			input: &models.Instance{
				ID:             1,
				Name:           "abc",
				Description:    "something",
				Vendor:         "kraken",
				Enabled:        true,
				Endpoint:       "https://example.com",
				AuthMode:       "none",
				AuthInfo:       map[string]string{"name": "harbor"},
				Status:         "Unknown",
				Default:        true,
				Insecure:       true,
				SetupTimestamp: 1234,
			},
			want: &instanceModel.Instance{
				ID:             1,
				Name:           "abc",
				Description:    "something",
				Vendor:         "kraken",
				Enabled:        true,
				Endpoint:       "https://example.com",
				AuthMode:       "none",
				AuthData:       `{"name":"harbor"}`,
				Status:         "Unknown",
				Default:        true,
				Insecure:       true,
				SetupTimestamp: 1234,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertParamInstanceToModelInstance(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertParamInstanceToModelInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertParamInstanceToModelInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertExecutionToPayload(t *testing.T) {
	tests := []struct {
		name      string
		input     *task.Execution
		expect    *models.Execution
		shouldErr bool
	}{
		{
			name:      "nil model",
			input:     nil,
			expect:    nil,
			shouldErr: true,
		},
		{
			name: "should ok",
			input: &task.Execution{
				ID:            1,
				VendorType:    "p2p",
				VendorID:      1,
				Status:        "",
				StatusMessage: "",
				Metrics:       nil,
				Trigger:       "",
				ExtraAttrs:    nil,
				StartTime:     time.Time{},
				EndTime:       time.Time{},
			},
			expect: &models.Execution{
				EndTime:       "0001-01-01 00:00:00 +0000 UTC",
				ExtraAttrs:    nil,
				ID:            1,
				Metrics:       nil,
				StartTime:     "0001-01-01 00:00:00 +0000 UTC",
				Status:        "",
				StatusMessage: "",
				Trigger:       "",
				VendorID:      1,
				VendorType:    "p2p",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertExecutionToPayload(tt.input)
			if (err != nil) != tt.shouldErr {
				t.Errorf("convertExecutionToPayload() error = %v, wantErr %v", err, tt.shouldErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("convertExecutionToPayload() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func Test_convertTaskToPayload(t *testing.T) {
	tests := []struct {
		name      string
		input     *task.Task
		expect    *models.Task
		shouldErr bool
	}{
		{
			name:      "nil model",
			input:     nil,
			expect:    nil,
			shouldErr: true,
		},
		{
			name: "should ok",
			input: &task.Task{
				ID:            0,
				ExecutionID:   0,
				Status:        "",
				StatusMessage: "",
				RunCount:      0,
				ExtraAttrs:    nil,
				CreationTime:  time.Time{},
				StartTime:     time.Time{},
				UpdateTime:    time.Time{},
				EndTime:       time.Time{},
			},
			expect: &models.Task{
				CreationTime:  "0001-01-01 00:00:00 +0000 UTC",
				EndTime:       "0001-01-01 00:00:00 +0000 UTC",
				ExecutionID:   0,
				ExtraAttrs:    nil,
				ID:            0,
				RunCount:      0,
				StartTime:     "0001-01-01 00:00:00 +0000 UTC",
				Status:        "",
				StatusMessage: "",
				UpdateTime:    "0001-01-01 00:00:00 +0000 UTC",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertTaskToPayload(tt.input)
			if (err != nil) != tt.shouldErr {
				t.Errorf("convertTaskToPayload() error = %v, wantErr %v", err, tt.shouldErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("convertTaskToPayload() = %v, want %v", got, tt.expect)
			}
		})
	}
}
