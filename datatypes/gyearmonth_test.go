/*
Copyright 2025-2026 Trident Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//nolint:testpackage // White-box test in the same package to access unexported functions.
package datatypes

import (
	"errors"
	"testing"
)

// assertGYearMonthFields verifies the year, month, and timezone offset on a GYearMonth value.
func assertGYearMonthFields(t *testing.T, gym GYearMonth, expYr int64, expMo uint8, tzExpected *int16) {
	t.Helper()
	if gym.Year() != expYr {
		t.Errorf("expected year %d, got %d", expYr, gym.Year())
	}
	if gym.Month() != expMo {
		t.Errorf("expected month %d, got %d", expMo, gym.Month())
	}
	if tzExpected == nil {
		if gym.TimezoneOffset() != nil {
			t.Errorf("expected nil timezone offset, got %v", gym.TimezoneOffset())
		}
		return
	}
	if gym.TimezoneOffset() == nil {
		t.Fatalf("expected timezone offset but got nil")
	}
	if gym.TimezoneOffset().offset != *tzExpected {
		t.Errorf("expected timezone offset %d, got %d", *tzExpected, gym.TimezoneOffset().offset)
	}
}

// TestNewGYearMonth verifies construction and property accessors of GYearMonth.
func TestNewGYearMonth(t *testing.T) {
	tz := GetUTC()
	gym, err := NewGYearMonth(2024, 6, tz)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gym.Year() != 2024 {
		t.Errorf("expected year 2024, got %d", gym.Year())
	}
	if gym.Month() != 6 {
		t.Errorf("expected month 6, got %d", gym.Month())
	}
	if gym.TimezoneOffset() == nil || gym.TimezoneOffset().offset != 0 {
		t.Errorf("expected UTC timezone offset, got %v", gym.TimezoneOffset())
	}

	gymUnzoned, err := NewGYearMonth(-45, 11, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gymUnzoned.Year() != -45 {
		t.Errorf("expected year -45, got %d", gymUnzoned.Year())
	}
	if gymUnzoned.Month() != 11 {
		t.Errorf("expected month 11, got %d", gymUnzoned.Month())
	}
	if gymUnzoned.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", gymUnzoned.TimezoneOffset())
	}

	if _, err = NewGYearMonth(2024, 0, nil); err == nil {
		t.Errorf("expected error for month 0, got nil")
	}
	if _, err = NewGYearMonth(2024, 13, nil); err == nil {
		t.Errorf("expected error for month 13, got nil")
	}
}

// TestParseGYearMonthValid tests the parsing of valid gYearMonth lexical representations.
func TestParseGYearMonthValid(t *testing.T) {
	tz0 := int16(0)
	tz300 := int16(300)
	tzNeg150 := int16(-150)

	tests := []struct {
		input      string
		expectedYr int64
		expectedMo uint8
		expectedTz *int16
	}{
		{input: "2023-05", expectedYr: 2023, expectedMo: 5, expectedTz: nil},
		{input: "-0045-11", expectedYr: -45, expectedMo: 11, expectedTz: nil},
		{input: "2023-05Z", expectedYr: 2023, expectedMo: 5, expectedTz: &tz0},
		{input: "2023-05+05:00", expectedYr: 2023, expectedMo: 5, expectedTz: &tz300},
		{input: "2023-05-02:30", expectedYr: 2023, expectedMo: 5, expectedTz: &tzNeg150},
		{input: "12345-12", expectedYr: 12345, expectedMo: 12, expectedTz: nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gym, err := ParseGYearMonth(tt.input)
			if err != nil {
				t.Fatalf("unexpected error parsing %q: %v", tt.input, err)
			}
			assertGYearMonthFields(t, gym, tt.expectedYr, tt.expectedMo, tt.expectedTz)
		})
	}
}

// TestParseGYearMonthInvalid tests the error handling branches during parsing of invalid gYearMonth inputs.
func TestParseGYearMonthInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "invalid year digits", input: "23-05"},
		{name: "missing separator after year", input: "2023/05"},
		{name: "invalid month format", input: "2023-5"},
		{name: "month out of range upper", input: "2023-13"},
		{name: "month out of range lower", input: "2023-00"},
		{name: "unrecognized suffix without timezone", input: "2023-05extra"},
		{name: "unrecognized suffix after timezone", input: "2023-05Ztrailing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseGYearMonth(tt.input)
			if err == nil {
				t.Fatalf("expected error parsing %q, but got nil", tt.input)
			}
		})
	}
}

// TestGYearMonthAdjust verifies timezone adjustment across all offset combinations.
func TestGYearMonthAdjust(t *testing.T) {
	tzPlus2, err := NewTimezoneOffset(120)
	if err != nil {
		t.Fatalf("unexpected error creating timezone: %v", err)
	}

	unzoned, err := ParseGYearMonth("2023-05")
	if err != nil {
		t.Fatalf("unexpected error parsing: %v", err)
	}

	adjustedToZoned := unzoned.Adjust(&tzPlus2)
	if adjustedToZoned.TimezoneOffset() == nil || adjustedToZoned.TimezoneOffset().offset != 120 {
		t.Errorf("expected timezone +02:00, got %v", adjustedToZoned.TimezoneOffset())
	}

	adjustedToUnzoned := adjustedToZoned.Adjust(nil)
	if adjustedToUnzoned.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", adjustedToUnzoned.TimezoneOffset())
	}

	adjustedZonedToZoned := adjustedToZoned.Adjust(GetUTC())
	if adjustedZonedToZoned.TimezoneOffset() == nil || adjustedZonedToZoned.TimezoneOffset().offset != 0 {
		t.Errorf("expected UTC timezone offset, got %v", adjustedZonedToZoned.TimezoneOffset())
	}

	adjustedUnzonedToUnzoned := unzoned.Adjust(nil)
	if adjustedUnzonedToUnzoned.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", adjustedUnzonedToUnzoned.TimezoneOffset())
	}
}

// TestGYearMonthIsIdenticalWith tests identity comparison between GYearMonth values and other XSD types.
func TestGYearMonthIsIdenticalWith(t *testing.T) {
	val1, _ := ParseGYearMonth("2023-05")
	val2, _ := ParseGYearMonth("2023-05")
	valDiffYear, _ := ParseGYearMonth("2024-05")
	valDiffMonth, _ := ParseGYearMonth("2023-06")
	valZonedUTC1, _ := ParseGYearMonth("2023-05Z")
	valZonedUTC2, _ := ParseGYearMonth("2023-05+00:00")
	valZonedPlus, _ := ParseGYearMonth("2023-05+02:00")

	if !val1.IsIdenticalWith(val2) {
		t.Errorf("expected identical values for unzoned 2023-05")
	}
	if val1.IsIdenticalWith(valDiffYear) {
		t.Errorf("expected false for different years")
	}
	if val1.IsIdenticalWith(valDiffMonth) {
		t.Errorf("expected false for different months")
	}
	if val1.IsIdenticalWith(valZonedUTC1) {
		t.Errorf("expected false comparing unzoned with zoned")
	}
	if valZonedUTC1.IsIdenticalWith(val1) {
		t.Errorf("expected false comparing zoned with unzoned")
	}
	if !valZonedUTC1.IsIdenticalWith(valZonedUTC2) {
		t.Errorf("expected identical values for zoned UTC matches")
	}
	if valZonedUTC1.IsIdenticalWith(valZonedPlus) {
		t.Errorf("expected false for different timezone offsets")
	}
	if val1.IsIdenticalWith(String("2023-05")) {
		t.Errorf("expected false for incompatible type")
	}
}

// TestGYearMonthCompare tests order comparisons and partial order indeterminacy.
func TestGYearMonthCompare(t *testing.T) {
	val1, _ := ParseGYearMonth("2023-05")
	val2, _ := ParseGYearMonth("2023-06")
	valEqual, _ := ParseGYearMonth("2023-05")
	valZoned, _ := ParseGYearMonth("2023-05Z")

	cmp, err := val1.Compare(val2)
	if err != nil || cmp != -1 {
		t.Errorf("expected -1 and nil error, got %d, %v", cmp, err)
	}

	cmp, err = val2.Compare(val1)
	if err != nil || cmp != 1 {
		t.Errorf("expected 1 and nil error, got %d, %v", cmp, err)
	}

	cmp, err = val1.Compare(valEqual)
	if err != nil || cmp != 0 {
		t.Errorf("expected 0 and nil error, got %d, %v", cmp, err)
	}

	_, err = val1.Compare(String("invalid"))
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected ErrDateTimeOverflow for incomparable type, got %v", err)
	}

	_, err = val1.Compare(valZoned)
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected ErrDateTimeOverflow for indeterminate comparison, got %v", err)
	}
}

// TestGYearMonthString tests canonical serialization across negative, positive, zoned, and unzoned variants.
func TestGYearMonthString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{input: "2023-05", expected: "2023-05"},
		{input: "0005-02", expected: "0005-02"},
		{input: "-0045-11", expected: "-0045-11"},
		{input: "2023-05Z", expected: "2023-05Z"},
		{input: "2023-05+05:00", expected: "2023-05+05:00"},
		{input: "-0045-11-02:00", expected: "-0045-11-02:00"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gym, err := ParseGYearMonth(tt.input)
			if err != nil {
				t.Fatalf("unexpected error parsing %q: %v", tt.input, err)
			}
			if gym.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, gym.String())
			}
		})
	}
}
