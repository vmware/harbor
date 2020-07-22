package helper

import "testing"

func TestImageRepository_Valid(t *testing.T) {
	tests := []struct {
		name string
		ir   ImageRepository
		want bool
	}{
		{"empty", "", false},
		{"invalid", "abc", false},
		{"invalid", "abc/def", false},
		{"valid", "abc/def:tag", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ir.Valid(); got != tt.want {
				t.Errorf("ImageRepository.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageRepository_Name(t *testing.T) {
	tests := []struct {
		name string
		ir   ImageRepository
		want string
	}{
		{"empty", "", ""},
		{"invalid", "abc", "abc"},
		{"invalid", "abc/def", "abc/def"},
		{"valid", "abc/def:tag", "abc/def"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ir.Name(); got != tt.want {
				t.Errorf("ImageRepository.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageRepository_Tag(t *testing.T) {
	tests := []struct {
		name string
		ir   ImageRepository
		want string
	}{
		{"empty", "", ""},
		{"invalid", "abc", ""},
		{"invalid", "abc/def", ""},
		{"valid", "abc/def:tag", "tag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ir.Tag(); got != tt.want {
				t.Errorf("ImageRepository.Tag() = %v, want %v", got, tt.want)
			}
		})
	}
}
