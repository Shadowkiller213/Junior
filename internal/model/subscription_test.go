package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"awesomeProject2/internal/model"
)

func TestMonthYear_Validate(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"07-2025", false},
		{"01-2000", false},
		{"12-1999", false},
		{"00-2025", true},  // month 0
		{"13-2025", true},  // month 13
		{"7-2025", true},   // month without leading zero
		{"07-25", true},    // 2-digit year
		{"07/2025", true},  // wrong separator
		{"", true},         // empty
		{"ab-2025", true},  // non-numeric month
		{"07-abcd", true},  // non-numeric year
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := model.MonthYear(tt.input).Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestMonthYear_ToTime(t *testing.T) {
	my := model.MonthYear("07-2025")
	got := my.ToTime()
	want := time.Date(2025, time.July, 1, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("ToTime() = %v, want %v", got, want)
	}
}

func TestMonthYear_MarshalJSON(t *testing.T) {
	my := model.MonthYear("07-2025")
	data, err := json.Marshal(my)
	if err != nil {
		t.Fatalf("MarshalJSON() unexpected error: %v", err)
	}
	if string(data) != `"07-2025"` {
		t.Errorf("MarshalJSON() = %s, want %q", data, "07-2025")
	}
}

func TestMonthYear_MarshalJSON_Invalid(t *testing.T) {
	my := model.MonthYear("bad")
	_, err := json.Marshal(my)
	if err == nil {
		t.Error("MarshalJSON() expected error for invalid value, got nil")
	}
}

func TestMonthYear_UnmarshalJSON(t *testing.T) {
	var my model.MonthYear
	if err := json.Unmarshal([]byte(`"07-2025"`), &my); err != nil {
		t.Fatalf("UnmarshalJSON() unexpected error: %v", err)
	}
	if my != "07-2025" {
		t.Errorf("UnmarshalJSON() = %q, want %q", my, "07-2025")
	}
}

func TestMonthYear_UnmarshalJSON_Invalid(t *testing.T) {
	var my model.MonthYear
	if err := json.Unmarshal([]byte(`"13-2025"`), &my); err == nil {
		t.Error("UnmarshalJSON() expected error for invalid month, got nil")
	}
}
