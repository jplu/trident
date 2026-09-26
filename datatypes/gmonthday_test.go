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

// TestNewGMonthDay validates constructors of GMonthDay instances with and without timezone offsets.
func TestNewGMonthDay(t *testing.T) {
	gmd, err := NewGMonthDay(11, 15, nil)
	if err != nil {
		t.Fatalf("unexpected error creating unzoned GMonthDay: %v", err)
	}
	if gmd.Month() != 11 {
		t.Errorf("expected month 11, got %d", gmd.Month())
	}
	if gmd.Day() != 15 {
		t.Errorf("expected day 15, got %d", gmd.Day())
	}
	if gmd.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", gmd.TimezoneOffset())
	}

	tz := GetUTC()
	gmdZoned, err := NewGMonthDay(2, 29, tz)
	if err != nil {
		t.Fatalf("unexpected error creating zoned GMonthDay: %v", err)
	}
	if gmdZoned.Month() != 2 {
		t.Errorf("expected month 2, got %d", gmdZoned.Month())
	}
	if gmdZoned.Day() != 29 {
		t.Errorf("expected day 29, got %d", gmdZoned.Day())
	}
	if gmdZoned.TimezoneOffset() == nil || gmdZoned.TimezoneOffset().offset != 0 {
		t.Errorf("expected UTC timezone offset, got %v", gmdZoned.TimezoneOffset())
	}

	if _, err = NewGMonthDay(2, 30, nil); err == nil {
		t.Errorf("expected error for invalid day in NewGMonthDay(2, 30, nil), got nil")
	}
	if _, err = NewGMonthDay(13, 1, nil); err == nil {
		t.Errorf("expected error for invalid month in NewGMonthDay(13, 1, nil), got nil")
	}
}

// TestParseGMonthDaySuccess tests valid parsing scenarios including leap year handling and timezones.
func TestParseGMonthDaySuccess(t *testing.T) {
	tests := []struct {
		input       string
		expectedMo  uint8
		expectedDay uint8
		hasTz       bool
		expectedTz  int16
	}{
		{
			input:       "--01-15",
			expectedMo:  1,
			expectedDay: 15,
			hasTz:       false,
		},
		{
			input:       "--11-30Z",
			expectedMo:  11,
			expectedDay: 30,
			hasTz:       true,
			expectedTz:  0,
		},
		{
			input:       "--02-29+05:30",
			expectedMo:  2,
			expectedDay: 29,
			hasTz:       true,
			expectedTz:  330,
		},
		{
			input:       "--07-04-08:00",
			expectedMo:  7,
			expectedDay: 4,
			hasTz:       true,
			expectedTz:  -480,
		},
	}

	for _, tt := range tests {
		gmd, err := ParseGMonthDay(tt.input)
		if err != nil {
			t.Fatalf("unexpected parse failure for %q: %v", tt.input, err)
		}
		if gmd.Month() != tt.expectedMo {
			t.Errorf("input %q: expected month %d, got %d", tt.input, tt.expectedMo, gmd.Month())
		}
		if gmd.Day() != tt.expectedDay {
			t.Errorf("input %q: expected day %d, got %d", tt.input, tt.expectedDay, gmd.Day())
		}
		if tt.hasTz {
			if gmd.TimezoneOffset() == nil {
				t.Fatalf("input %q: expected timezone offset, got nil", tt.input)
			}
			if gmd.TimezoneOffset().offset != tt.expectedTz {
				t.Errorf("input %q: expected timezone %d, got %d", tt.input, tt.expectedTz, gmd.TimezoneOffset().offset)
			}
		} else if gmd.TimezoneOffset() != nil {
			t.Errorf("input %q: expected nil timezone, got %v", tt.input, gmd.TimezoneOffset())
		}
	}

	_, err := ParseGMonthDay("--12-31")
	if err != nil {
		t.Fatalf("unexpected parse failure for '--12-31': %v", err)
	}
}

// TestParseGMonthDayErrors verifies lexical and logical rejection branches during GMonthDay parsing.
func TestParseGMonthDayErrors(t *testing.T) {
	invalidInputs := []string{
		"",
		"-05-15",
		"05-15",
		"--13-15",
		"--00-15",
		"--aa-15",
		"--0515",
		"--05:15",
		"--05-00",
		"--05-32",
		"--05-bb",
		"--05-15suffix",
		"--05-15Ztrailing",
		"--02-30",
		"--04-31",
		"--06-31",
		"--09-31",
		"--11-31",
	}

	for _, input := range invalidInputs {
		_, err := ParseGMonthDay(input)
		if err == nil {
			t.Errorf("expected error parsing invalid input %q, got nil", input)
		}
	}
}

// TestGMonthDayAdjust tests timezone adjustments across unzoned and zoned GMonthDay instances.
func TestGMonthDayAdjust(t *testing.T) {
	gmd, err := ParseGMonthDay("--05-15")
	if err != nil {
		t.Fatalf("unexpected parse failure: %v", err)
	}

	tz, err := NewTimezoneOffset(120)
	if err != nil {
		t.Fatalf("unexpected error creating timezone: %v", err)
	}

	adjusted := gmd.Adjust(&tz)
	if adjusted.TimezoneOffset() == nil || adjusted.TimezoneOffset().offset != 120 {
		t.Errorf("expected timezone offset 120, got %v", adjusted.TimezoneOffset())
	}

	unzoned := adjusted.Adjust(nil)
	if unzoned.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone after unzoned adjustment, got %v", unzoned.TimezoneOffset())
	}

	noop := gmd.Adjust(nil)
	if noop.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone, got %v", noop.TimezoneOffset())
	}
}

// TestGMonthDayIsIdenticalWith tests identity equivalence rules across identical and divergent values.
func TestGMonthDayIsIdenticalWith(t *testing.T) {
	gmd1, _ := ParseGMonthDay("--05-15")
	gmd2, _ := ParseGMonthDay("--05-15")
	gmdDiffDate, _ := ParseGMonthDay("--05-16")
	gmdZonedUTC, _ := ParseGMonthDay("--05-15Z")
	gmdZonedUTC2, _ := ParseGMonthDay("--05-15Z")
	gmdZonedOffset, _ := ParseGMonthDay("--05-15+00:00")
	gmdZonedOther, _ := ParseGMonthDay("--05-15+02:00")

	if !gmd1.IsIdenticalWith(gmd2) {
		t.Errorf("expected %v to be identical with %v", gmd1, gmd2)
	}
	if gmd1.IsIdenticalWith(gmdDiffDate) {
		t.Errorf("expected %v not to be identical with %v", gmd1, gmdDiffDate)
	}
	if gmd1.IsIdenticalWith(gmdZonedUTC) {
		t.Errorf("expected unzoned %v not to be identical with zoned %v", gmd1, gmdZonedUTC)
	}
	if !gmdZonedUTC.IsIdenticalWith(gmdZonedUTC2) {
		t.Errorf("expected %v to be identical with %v", gmd1, gmdZonedUTC2)
	}
	if !gmdZonedUTC.IsIdenticalWith(gmdZonedOffset) {
		t.Errorf("expected %v to be identical with %v", gmdZonedUTC, gmdZonedOffset)
	}
	if gmdZonedUTC.IsIdenticalWith(gmdZonedOther) {
		t.Errorf("expected %v not to be identical with %v", gmdZonedUTC, gmdZonedOther)
	}
	if gmd1.IsIdenticalWith(String("--05-15")) {
		t.Errorf("expected comparison with non-GMonthDay to return false")
	}
}

// TestGMonthDayCompare tests order comparison relations including determinate and indeterminate cases.
func TestGMonthDayCompare(t *testing.T) {
	gmdA, _ := ParseGMonthDay("--05-15Z")
	gmdB, _ := ParseGMonthDay("--05-16Z")
	gmdC, _ := ParseGMonthDay("--05-15Z")

	cmp, err := gmdA.Compare(gmdB)
	if err != nil || cmp != -1 {
		t.Errorf("expected -1, got %d (err: %v)", cmp, err)
	}

	cmp, err = gmdB.Compare(gmdA)
	if err != nil || cmp != 1 {
		t.Errorf("expected 1, got %d (err: %v)", cmp, err)
	}

	cmp, err = gmdA.Compare(gmdC)
	if err != nil || cmp != 0 {
		t.Errorf("expected 0, got %d (err: %v)", cmp, err)
	}

	_, err = gmdA.Compare(String("not-gmonthday"))
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected ErrDateTimeOverflow for incompatible comparison, got %v", err)
	}

	gmdUnzoned, _ := ParseGMonthDay("--05-15")
	_, err = gmdUnzoned.Compare(gmdA)
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected ErrDateTimeOverflow for indeterminate comparison, got %v", err)
	}

	gmdFarPast, _ := ParseGMonthDay("--01-01")
	gmdFarFuture, _ := ParseGMonthDay("--11-30Z")
	cmp, err = gmdFarFuture.Compare(gmdFarPast)
	if err != nil || cmp != -1 {
		t.Errorf("expected determinate comparison -1, got %d (err: %v)", cmp, err)
	}
}

// TestGMonthDayString tests the canonical lexical serialization of GMonthDay instances.
func TestGMonthDayString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"--05-15", "--05-15"},
		{"--05-15Z", "--05-15Z"},
		{"--05-15+05:00", "--05-15+05:00"},
		{"--05-15-08:30", "--05-15-08:30"},
	}

	for _, tt := range tests {
		gmd, err := ParseGMonthDay(tt.input)
		if err != nil {
			t.Fatalf("unexpected parse failure for %q: %v", tt.input, err)
		}
		if gmd.String() != tt.expected {
			t.Errorf("input %q: expected string %q, got %q", tt.input, tt.expected, gmd.String())
		}
	}
}
