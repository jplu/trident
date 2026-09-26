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

// nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to
// test unexported functions.
package datatypes

import (
	"errors"
	"testing"
)

// TestNewGYear verifies constructor behavior with and without timezone offsets.
func TestNewGYear(t *testing.T) {
	tz, err := NewTimezoneOffset(120)
	if err != nil {
		t.Fatalf("unexpected timezone error: %v", err)
	}

	gy1 := NewGYear(2024, nil)
	if gy1.Year() != 2024 {
		t.Errorf("expected year 2024, got %d", gy1.Year())
	}
	if gy1.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone, got %v", gy1.TimezoneOffset())
	}

	gy2 := NewGYear(-45, &tz)
	if gy2.Year() != -45 {
		t.Errorf("expected year -45, got %d", gy2.Year())
	}
	if gy2.TimezoneOffset() == nil || gy2.TimezoneOffset().offset != 120 {
		t.Errorf("expected timezone offset 120, got %v", gy2.TimezoneOffset())
	}
}

// TestParseGYearValid verifies parsing valid gYear lexical literals.
func TestParseGYearValid(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantYear  int64
		hasTz     bool
		wantTzOff int16
	}{
		{
			name:     "valid positive 4-digit year",
			input:    "2024",
			wantYear: 2024,
			hasTz:    false,
		},
		{
			name:      "valid positive year with UTC timezone",
			input:     "2024Z",
			wantYear:  2024,
			hasTz:     true,
			wantTzOff: 0,
		},
		{
			name:      "valid positive year with positive offset",
			input:     "2024+05:30",
			wantYear:  2024,
			hasTz:     true,
			wantTzOff: 330,
		},
		{
			name:      "valid positive year with negative offset",
			input:     "2024-08:00",
			wantYear:  2024,
			hasTz:     true,
			wantTzOff: -480,
		},
		{
			name:     "valid negative year",
			input:    "-0005",
			wantYear: -5,
			hasTz:    false,
		},
		{
			name:      "valid negative year with timezone",
			input:     "-0005Z",
			wantYear:  -5,
			hasTz:     true,
			wantTzOff: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gy, err := ParseGYear(tc.input)
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if gy.Year() != tc.wantYear {
				t.Errorf("expected year %d, got %d", tc.wantYear, gy.Year())
			}
			if tc.hasTz {
				if gy.TimezoneOffset() == nil {
					t.Fatalf("expected timezone offset, got nil")
				}
				if gy.TimezoneOffset().offset != tc.wantTzOff {
					t.Errorf("expected timezone offset %d, got %d", tc.wantTzOff, gy.TimezoneOffset().offset)
				}
			} else if gy.TimezoneOffset() != nil {
				t.Errorf("expected nil timezone offset, got %v", gy.TimezoneOffset())
			}
		})
	}
}

// TestParseGYearInvalid verifies parsing invalid gYear lexical literals.
func TestParseGYearInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid year format non-numeric",
			input: "ABCD",
		},
		{
			name:  "invalid year less than 4 digits",
			input: "123",
		},
		{
			name:  "invalid year with leading zero more than 4 digits",
			input: "02024",
		},
		{
			name:  "invalid unrecognized trailing suffix",
			input: "2024Zextra",
		},
		{
			name:  "invalid suffix without timezone",
			input: "2024-invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseGYear(tc.input)
			if err == nil {
				t.Fatalf("expected error for input %q, got nil", tc.input)
			}
		})
	}
}

// TestGYearYear verifies the Year accessor method.
func TestGYearYear(t *testing.T) {
	gy1, err := ParseGYear("1999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gy1.Year() != 1999 {
		t.Errorf("expected 1999, got %d", gy1.Year())
	}

	gy2, err := ParseGYear("-0044")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gy2.Year() != -44 {
		t.Errorf("expected -44, got %d", gy2.Year())
	}
}

// TestGYearTimezoneOffset verifies the TimezoneOffset accessor method.
func TestGYearTimezoneOffset(t *testing.T) {
	gyUnzoned, err := ParseGYear("2020")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gyUnzoned.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", gyUnzoned.TimezoneOffset())
	}

	gyZoned, err := ParseGYear("2020+01:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gyZoned.TimezoneOffset() == nil || gyZoned.TimezoneOffset().offset != 60 {
		t.Errorf("expected timezone offset 60, got %v", gyZoned.TimezoneOffset())
	}
}

// TestGYearAdjust verifies timezone adjustment for GYear values.
func TestGYearAdjust(t *testing.T) {
	gy, err := ParseGYear("2024Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	targetTz, err := NewTimezoneOffset(180)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	adjusted := gy.Adjust(&targetTz)
	if adjusted.TimezoneOffset() == nil || adjusted.TimezoneOffset().offset != 180 {
		t.Errorf("expected offset 180, got %v", adjusted.TimezoneOffset())
	}

	unzoneAdjusted := gy.Adjust(nil)
	if unzoneAdjusted.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", unzoneAdjusted.TimezoneOffset())
	}

	gyUnzoned, err := ParseGYear("2024")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	zoneAdjusted := gyUnzoned.Adjust(&targetTz)
	if zoneAdjusted.TimezoneOffset() == nil || zoneAdjusted.TimezoneOffset().offset != 180 {
		t.Errorf("expected offset 180, got %v", zoneAdjusted.TimezoneOffset())
	}
}

// TestGYearIsIdenticalWith verifies identity rules across value spaces and timezone presence.
func TestGYearIsIdenticalWith(t *testing.T) {
	gy2024Unzoned1, _ := ParseGYear("2024")
	gy2024Unzoned2, _ := ParseGYear("2024")
	gy2025Unzoned, _ := ParseGYear("2025")
	gy2024UTC, _ := ParseGYear("2024Z")
	gy2024PlusOne, _ := ParseGYear("2024+01:00")
	gy2024UTC2, _ := ParseGYear("2024+00:00")

	if !gy2024Unzoned1.IsIdenticalWith(gy2024Unzoned2) {
		t.Errorf("identical unzoned GYear instances evaluated to false")
	}
	if gy2024Unzoned1.IsIdenticalWith(gy2025Unzoned) {
		t.Errorf("different unzoned GYear instances evaluated to true")
	}
	if gy2024Unzoned1.IsIdenticalWith(gy2024UTC) {
		t.Errorf("unzoned and zoned GYear instances evaluated to true")
	}
	if gy2024UTC.IsIdenticalWith(gy2024Unzoned1) {
		t.Errorf("zoned and unzoned GYear instances evaluated to true")
	}
	if !gy2024UTC.IsIdenticalWith(gy2024UTC2) {
		t.Errorf("matching zoned GYear instances evaluated to false")
	}
	if gy2024UTC.IsIdenticalWith(gy2024PlusOne) {
		t.Errorf("different timezone offsets evaluated to true")
	}
	if gy2024Unzoned1.IsIdenticalWith(String("2024")) {
		t.Errorf("different datatype evaluated to true")
	}
}

// TestGYearCompare verifies temporal comparison, order relations, and indeterminate cases.
func TestGYearCompare(t *testing.T) {
	gy2023, _ := ParseGYear("2023")
	gy2024, _ := ParseGYear("2024")
	gy2025, _ := ParseGYear("2025")
	gy2024Same, _ := ParseGYear("2024")

	cmp, err := gy2023.Compare(gy2024)
	if err != nil || cmp != -1 {
		t.Errorf("expected -1, nil; got %d, %v", cmp, err)
	}

	cmp, err = gy2024.Compare(gy2024Same)
	if err != nil || cmp != 0 {
		t.Errorf("expected 0, nil; got %d, %v", cmp, err)
	}

	cmp, err = gy2025.Compare(gy2024)
	if err != nil || cmp != 1 {
		t.Errorf("expected 1, nil; got %d, %v", cmp, err)
	}

	cmp, err = gy2024.Compare(String("2024"))
	if !errors.Is(err, ErrDateTimeOverflow) || cmp != 0 {
		t.Errorf("expected 0, ErrDateTimeOverflow for incomparable types; got %d, %v", cmp, err)
	}

	gy2024Z, _ := ParseGYear("2024Z")
	cmp, err = gy2024Z.Compare(gy2024)
	if !errors.Is(err, ErrDateTimeOverflow) || cmp != 0 {
		t.Errorf("expected 0, ErrDateTimeOverflow for indeterminate comparison; got %d, %v", cmp, err)
	}
}

// TestGYearString verifies canonical lexical representation formatting.
func TestGYearString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "positive four-digit year unzoned",
			input:    "2024",
			expected: "2024",
		},
		{
			name:     "positive four-digit year with Z",
			input:    "2024Z",
			expected: "2024Z",
		},
		{
			name:     "positive four-digit year with offset",
			input:    "2024+02:00",
			expected: "2024+02:00",
		},
		{
			name:     "negative year less than 4 digits canonicalized",
			input:    "-0005",
			expected: "-0005",
		},
		{
			name:     "negative year with timezone",
			input:    "-0005Z",
			expected: "-0005Z",
		},
		{
			name:     "negative multi-digit year",
			input:    "-12345",
			expected: "-12345",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gy, err := ParseGYear(tc.input)
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}
			if gy.String() != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, gy.String())
			}
		})
	}

	gySmallPos := NewGYear(5, nil)
	if gySmallPos.String() != "0005" {
		t.Errorf("expected \"0005\", got %q", gySmallPos.String())
	}
}
