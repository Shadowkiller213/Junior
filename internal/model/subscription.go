package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MonthYear is a custom string type representing a month+year in "MM-YYYY" format.
type MonthYear string

// MarshalJSON implements json.Marshaler.
func (m MonthYear) MarshalJSON() ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(string(m))
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *MonthYear) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	tmp := MonthYear(s)
	if err := tmp.Validate(); err != nil {
		return err
	}
	*m = tmp
	return nil
}

// Validate checks that the MonthYear value is in "MM-YYYY" format with valid month/year values.
func (m MonthYear) Validate() error {
	s := string(m)
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid month-year format %q: expected MM-YYYY", s)
	}
	month, err := strconv.Atoi(parts[0])
	if err != nil || month < 1 || month > 12 || len(parts[0]) != 2 {
		return fmt.Errorf("invalid month in %q: must be 01-12", s)
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil || year < 1 || len(parts[1]) != 4 {
		return fmt.Errorf("invalid year in %q: must be a 4-digit number", s)
	}
	return nil
}

// ToTime returns the first day of the month represented by MonthYear.
func (m MonthYear) ToTime() time.Time {
	parts := strings.Split(string(m), "-")
	month, _ := strconv.Atoi(parts[0])
	year, _ := strconv.Atoi(parts[1])
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
}

// Subscription represents an online subscription record.
type Subscription struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int64      `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   MonthYear  `json:"start_date"`
	EndDate     *MonthYear `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateRequest is the request body for creating a subscription.
type CreateRequest struct {
	ServiceName string     `json:"service_name"`
	Price       int64      `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   MonthYear  `json:"start_date"`
	EndDate     *MonthYear `json:"end_date,omitempty"`
}

// UpdateRequest is the request body for updating a subscription.
type UpdateRequest struct {
	ServiceName string     `json:"service_name"`
	Price       int64      `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   MonthYear  `json:"start_date"`
	EndDate     *MonthYear `json:"end_date,omitempty"`
}

// TotalResponse is the response body for the total cost endpoint.
type TotalResponse struct {
	Total int64 `json:"total"`
}
