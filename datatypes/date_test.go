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

// TestNewDate verifies the constructor for Date.
func TestNewDate(t *testing.T) {
	tz := GetUTC()
	date, err := NewDate(2023, 5, 15, tz)
	if err != nil {
		t.Fatalf("unexpected error creating Date: %v", err)
	}

	if date.Year() != 2023 {
		t.Errorf("expected year 2023, got %d", date.Year())
	}
	if date.Month() != 5 {
		t.Errorf("expected month 5, got %d", date.Month())
	}
	if date.Day() != 15 {
		t.Errorf("expected day 15, got %d", date.Day())
	}
	if date.TimezoneOffset() == nil || *date.TimezoneOffset() != *tz {
		t.Errorf("expected timezone %v, got %v", tz, date.TimezoneOffset())
	}
}

// TestNewDateInvalid verifies error handling for invalid date parameters in NewDate.
func TestNewDateInvalid(t *testing.T) {
	tests := []struct {
		name  string
		year  int64
		month uint8
		day   uint8
	}{
		{"invalid month zero", 2023, 0, 15},
		{"invalid month 13", 2023, 13, 15},
		{"invalid day zero", 2023, 5, 0},
		{"invalid day 32", 2023, 5, 32},
		{"non-leap year feb 29", 2023, 2, 29},
		{"april 31", 2023, 4, 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDate(tt.year, tt.month, tt.day, nil)
			if err == nil {
				t.Errorf("expected error for NewDate(%d, %d, %d), got nil", tt.year, tt.month, tt.day)
			}
		})
	}
}

type parseDateTestCase struct {
	input       string
	expectedY   int64
	expectedM   uint8
	expectedD   uint8
	hasTimezone bool
	tzOffset    int16
}

// assertParsedDate is a helper that checks parsed date fields against expected values.
func assertParsedDate(t *testing.T, date Date, tt parseDateTestCase) {
	t.Helper()
	if date.Year() != tt.expectedY {
		t.Errorf("expected year %d, got %d", tt.expectedY, date.Year())
	}
	if date.Month() != tt.expectedM {
		t.Errorf("expected month %d, got %d", tt.expectedM, date.Month())
	}
	if date.Day() != tt.expectedD {
		t.Errorf("expected day %d, got %d", tt.expectedD, date.Day())
	}
	if tt.hasTimezone {
		if date.TimezoneOffset() == nil {
			t.Fatalf("expected timezone offset, got nil")
		}
		if date.TimezoneOffset().offset != tt.tzOffset {
			t.Errorf("expected timezone offset %d, got %d", tt.tzOffset, date.TimezoneOffset().offset)
		}
	} else if date.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", date.TimezoneOffset())
	}
}

// TestParseDateValid verifies parsing of valid date strings.
func TestParseDateValid(t *testing.T) {
	tests := []parseDateTestCase{
		{
			input:       "2023-05-15",
			expectedY:   2023,
			expectedM:   5,
			expectedD:   15,
			hasTimezone: false,
		},
		{
			input:       "-0004-12-31Z",
			expectedY:   -4,
			expectedM:   12,
			expectedD:   31,
			hasTimezone: true,
			tzOffset:    0,
		},
		{
			input:       "2020-02-29+02:00",
			expectedY:   2020,
			expectedM:   2,
			expectedD:   29,
			hasTimezone: true,
			tzOffset:    120,
		},
		{
			input:       "0001-01-01-05:00",
			expectedY:   1,
			expectedM:   1,
			expectedD:   1,
			hasTimezone: true,
			tzOffset:    -300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			date, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("unexpected parsing error for %q: %v", tt.input, err)
			}
			assertParsedDate(t, date, tt)
		})
	}
}

// TestParseDateInvalid verifies error handling for invalid date strings.
func TestParseDateInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid year digits", "202-05-15"},
		{"year leading zero overflow", "02023-05-15"},
		{"missing first hyphen", "2023/05-15"},
		{"missing first hyphen concatenated", "202305-15"},
		{"invalid month number", "2023-13-15"},
		{"invalid month zero", "2023-00-15"},
		{"non-numeric month", "2023-XX-15"},
		{"missing second hyphen", "2023-05/15"},
		{"missing second hyphen concatenated", "2023-0515"},
		{"invalid day number", "2023-05-32"},
		{"invalid day zero", "2023-05-00"},
		{"non-numeric day", "2023-05-YY"},
		{"unrecognized suffix", "2023-05-15EXTRA"},
		{"non-leap year feb 29", "2021-02-29"},
		{"april 31st invalid", "2023-04-31"},
		{"invalid timezone offset", "2023-05-15+25:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDate(tt.input)
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
		})
	}
}

// TestDateIsIdenticalWith checks the exact equality of dates including timezones.
func TestDateIsIdenticalWith(t *testing.T) {
	tz1, _ := NewTimezoneOffset(120)
	tz2, _ := NewTimezoneOffset(180)

	d1, err := NewDate(2023, 5, 15, nil)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}
	d2, err := NewDate(2023, 5, 15, nil)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}
	d3, err := NewDate(2023, 5, 16, nil)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}
	d4, err := NewDate(2023, 5, 15, &tz1)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}
	d5, err := NewDate(2023, 5, 15, &tz1)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}
	d6, err := NewDate(2023, 5, 15, &tz2)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}

	if !d1.IsIdenticalWith(d2) {
		t.Errorf("expected unzoned dates to be identical")
	}
	if d1.IsIdenticalWith(d3) {
		t.Errorf("expected different dates not to be identical")
	}
	if d1.IsIdenticalWith(d4) {
		t.Errorf("expected unzoned and zoned dates not to be identical")
	}
	if !d4.IsIdenticalWith(d5) {
		t.Errorf("expected identical zoned dates to be identical")
	}
	if d4.IsIdenticalWith(d6) {
		t.Errorf("expected dates with different timezone offsets not to be identical")
	}
	if d1.IsIdenticalWith(String("2023-05-15")) {
		t.Errorf("expected comparison with different XSDValue type to return false")
	}
}

// TestDateCompare verifies chronological comparison between dates.
func TestDateCompare(t *testing.T) {
	d1, err := NewDate(2023, 5, 15, nil)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}
	d2, err := NewDate(2023, 5, 16, nil)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}
	d3, err := NewDate(2023, 5, 15, nil)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}

	cmp, err := d1.Compare(d3)
	if err != nil || cmp != 0 {
		t.Errorf("expected 0, got %d (err: %v)", cmp, err)
	}

	cmp, err = d1.Compare(d2)
	if err != nil || cmp != -1 {
		t.Errorf("expected -1, got %d (err: %v)", cmp, err)
	}

	cmp, err = d2.Compare(d1)
	if err != nil || cmp != 1 {
		t.Errorf("expected 1, got %d (err: %v)", cmp, err)
	}

	tz := GetUTC()
	dZoned, err := NewDate(2023, 5, 15, tz)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}
	_, err = d1.Compare(dZoned)
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected ErrDateTimeOverflow for indeterminate comparison, got %v", err)
	}

	_, err = d1.Compare(String("2023-05-15"))
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected ErrDateTimeOverflow for incompatible type, got %v", err)
	}
}

// TestDateString verifies the string representation of dates.
func TestDateString(t *testing.T) {
	tz, _ := NewTimezoneOffset(330)

	tests := []struct {
		year     int64
		month    uint8
		day      uint8
		tz       *TimezoneOffset
		expected string
	}{
		{2023, 5, 15, nil, "2023-05-15"},
		{-4, 5, 15, nil, "-0004-05-15"},
		{2023, 5, 15, GetUTC(), "2023-05-15Z"},
		{2023, 5, 15, &tz, "2023-05-15+05:30"},
	}

	for _, tt := range tests {
		date, err := NewDate(tt.year, tt.month, tt.day, tt.tz)
		if err != nil {
			t.Fatalf("unexpected error creating date for %+v: %v", tt, err)
		}
		if str := date.String(); str != tt.expected {
			t.Errorf("expected string %q, got %q", tt.expected, str)
		}
	}
}

// TestDateToDateTime verifies conversion from Date to DateTime.
func TestDateToDateTime(t *testing.T) {
	tz := GetUTC()
	date, err := NewDate(2023, 5, 15, tz)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}

	dt := date.ToDateTime()

	if dt.Year() != 2023 || dt.Month() != 5 || dt.Day() != 15 {
		t.Errorf("date components mismatch in DateTime: %v", dt)
	}
	if dt.Hour() != 0 || dt.Minute() != 0 || dt.Second().getValue().Sign() != 0 {
		t.Errorf(
			"expected time component to be 00:00:00, got %02d:%02d:%s",
			dt.Hour(),
			dt.Minute(),
			dt.Second().String(),
		)
	}
	if dt.TimezoneOffset() == nil || *dt.TimezoneOffset() != *tz {
		t.Errorf("timezone offset mismatch in DateTime: %v", dt.TimezoneOffset())
	}
}

// TestDateAdjust verifies timezone adjustment for dates.
func TestDateAdjust(t *testing.T) {
	tz1, _ := NewTimezoneOffset(120)
	tz2, _ := NewTimezoneOffset(-300)

	d, err := NewDate(2023, 5, 15, &tz1)
	if err != nil {
		t.Fatalf("unexpected error creating date: %v", err)
	}

	adjusted := d.Adjust(&tz2)
	if adjusted.TimezoneOffset() == nil || *adjusted.TimezoneOffset() != tz2 {
		t.Errorf("expected timezone offset %v, got %v", tz2, adjusted.TimezoneOffset())
	}

	unzoned := d.Adjust(nil)
	if unzoned.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", unzoned.TimezoneOffset())
	}
}
