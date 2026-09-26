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

// TestNewTime tests the constructor for Time with various hour, minute, second, and timezone combinations.
func TestNewTime(t *testing.T) {
	tz, err := NewTimezoneOffset(120)
	if err != nil {
		t.Fatalf("unexpected error creating timezone: %v", err)
	}

	sec5 := NewDecimalFromInt64(5)
	tm, err := NewTime(10, 20, sec5, &tz)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tm.Hour() != 10 || tm.Minute() != 20 {
		t.Errorf("expected 10:20, got %02d:%02d", tm.Hour(), tm.Minute())
	}

	tm24, err := NewTime(24, 0, DefaultDecimal(), nil)
	if err != nil {
		t.Fatalf("unexpected error for 24:00:00: %v", err)
	}
	if tm24.Hour() != 0 {
		t.Errorf("expected hour 0 after 24:00:00 normalization, got %d", tm24.Hour())
	}

	_, err = NewTime(24, 1, DefaultDecimal(), nil)
	if err == nil {
		t.Errorf("expected error for 24:01:00, got nil")
	}

	sec1 := NewDecimalFromInt64(1)
	_, err = NewTime(24, 0, sec1, nil)
	if err == nil {
		t.Errorf("expected error for 24:00:01, got nil")
	}
}

// TestParseTimeValid tests parsing valid xsd:time lexical representations.
func TestParseTimeValid(t *testing.T) {
	tests := []struct {
		input       string
		expectedH   uint8
		expectedM   uint8
		expectedTz  bool
		expectedOff int16
	}{
		{"12:30:45", 12, 30, false, 0},
		{"00:00:00", 0, 0, false, 0},
		{"24:00:00", 0, 0, false, 0},
		{"24:00:00Z", 0, 0, true, 0},
		{"12:30:45Z", 12, 30, true, 0},
		{"12:30:45+05:00", 12, 30, true, 300},
		{"12:30:45-05:00", 12, 30, true, -300},
		{"12:30:45.12345", 12, 30, false, 0},
	}

	for _, tc := range tests {
		tm, err := ParseTime(tc.input)
		if err != nil {
			t.Errorf("ParseTime(%q) unexpected error: %v", tc.input, err)
			continue
		}
		if tm.Hour() != tc.expectedH {
			t.Errorf("ParseTime(%q) Hour = %d, expected %d", tc.input, tm.Hour(), tc.expectedH)
		}
		if tm.Minute() != tc.expectedM {
			t.Errorf("ParseTime(%q) Minute = %d, expected %d", tc.input, tm.Minute(), tc.expectedM)
		}
		if tc.expectedTz {
			if tm.TimezoneOffset() == nil {
				t.Errorf("ParseTime(%q) expected timezone, got nil", tc.input)
			} else if tm.TimezoneOffset().offset != tc.expectedOff {
				t.Errorf("ParseTime(%q) expected timezone offset %d, got %d", tc.input, tc.expectedOff, tm.TimezoneOffset().offset)
			}
		} else if tm.TimezoneOffset() != nil {
			t.Errorf("ParseTime(%q) expected no timezone, got %v", tc.input, tm.TimezoneOffset())
		}
	}
}

// TestParseTimeInvalid tests lexical parsing errors for invalid xsd:time strings.
func TestParseTimeInvalid(t *testing.T) {
	invalidInputs := []string{
		"1:00:00",
		"25:00:00",
		"ab:00:00",
		"12-00:00",
		"12:60:00",
		"12:1:00",
		"12:00-00",
		"12:00:60",
		"12:00:5",
		"12:00:05.",
		"24:01:00",
		"24:00:01",
		"24:00:00.001",
		"12:00:00extra",
		"12:00:00+99:00",
	}

	for _, input := range invalidInputs {
		_, err := ParseTime(input)
		if err == nil {
			t.Errorf("ParseTime(%q) expected error, got nil", input)
		}
	}
}

// TestTimeGetters tests the accessor methods of Time.
func TestTimeGetters(t *testing.T) {
	tm, err := ParseTime("14:25:36.500+02:00")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if tm.Hour() != 14 {
		t.Errorf("expected hour 14, got %d", tm.Hour())
	}
	if tm.Minute() != 25 {
		t.Errorf("expected minute 25, got %d", tm.Minute())
	}
	if tm.Second().String() != "36.5" {
		t.Errorf("expected second 36.5, got %s", tm.Second().String())
	}
	if tm.TimezoneOffset() == nil || tm.TimezoneOffset().offset != 120 {
		t.Errorf("expected timezone offset +120, got %v", tm.TimezoneOffset())
	}

	unzoned, err := ParseTime("14:25:36")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if unzoned.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", unzoned.TimezoneOffset())
	}
}

// TestTimeIsIdenticalWith tests the value space identity verification for Time.
func TestTimeIsIdenticalWith(t *testing.T) {
	t1, _ := ParseTime("12:00:00")
	t2, _ := ParseTime("12:00:00")
	t3, _ := ParseTime("13:00:00")
	tZ1, _ := ParseTime("12:00:00Z")
	tZ2, _ := ParseTime("12:00:00Z")
	tPlus1, _ := ParseTime("13:00:00+01:00")
	tPlus2, _ := ParseTime("12:00:00+02:00")

	if t1.IsIdenticalWith(String("12:00:00")) {
		t.Errorf("expected non-Time to return false")
	}
	if !t1.IsIdenticalWith(t2) {
		t.Errorf("expected %v and %v to be identical", t1, t2)
	}
	if t1.IsIdenticalWith(t3) {
		t.Errorf("expected %v and %v not to be identical", t1, t3)
	}
	if t1.IsIdenticalWith(tZ1) {
		t.Errorf("expected unzoned and zoned not to be identical")
	}
	if tZ1.IsIdenticalWith(t1) {
		t.Errorf("expected zoned and unzoned not to be identical")
	}
	if !tZ1.IsIdenticalWith(tZ2) {
		t.Errorf("expected identical zoned times to match")
	}
	if tZ1.IsIdenticalWith(tPlus1) {
		t.Errorf("expected same instant with different offsets not to be identical")
	}
	if tPlus1.IsIdenticalWith(tPlus2) {
		t.Errorf("expected different timezone offsets not to be identical")
	}
}

// TestTimeCompare tests partial and total order comparisons of Time values.
func TestTimeCompare(t *testing.T) {
	t10, _ := ParseTime("10:00:00")
	t12, _ := ParseTime("12:00:00")
	t12Same, _ := ParseTime("12:00:00")
	t14, _ := ParseTime("14:00:00")
	t12Z, _ := ParseTime("12:00:00Z")
	t14Z, _ := ParseTime("14:00:00Z")

	_, err := t12.Compare(String("12:00:00"))
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected ErrDateTimeOverflow for incomparable type, got %v", err)
	}

	cmp, err := t10.Compare(t12)
	if err != nil || cmp != -1 {
		t.Errorf("expected -1, got %d (err: %v)", cmp, err)
	}

	cmp, err = t12.Compare(t12Same)
	if err != nil || cmp != 0 {
		t.Errorf("expected 0, got %d (err: %v)", cmp, err)
	}

	cmp, err = t14.Compare(t12)
	if err != nil || cmp != 1 {
		t.Errorf("expected 1, got %d (err: %v)", cmp, err)
	}

	cmp, err = t12Z.Compare(t14Z)
	if err != nil || cmp != -1 {
		t.Errorf("expected -1, got %d (err: %v)", cmp, err)
	}

	_, err = t12.Compare(t12Z)
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected indeterminate error when comparing overlapping zoned and unzoned times, got %v", err)
	}
}

// TestTimeString tests canonical lexical formatting of Time values across all second formatting branches.
func TestTimeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"10:20:05", "10:20:05"},
		{"10:20:15", "10:20:15"},
		{"10:20:05.5", "10:20:05.5"},
		{"10:20:15.5+01:00", "10:20:15.5+01:00"},
		{"10:20:05Z", "10:20:05Z"},
	}

	for _, tc := range tests {
		tm, err := ParseTime(tc.input)
		if err != nil {
			t.Fatalf("ParseTime(%q) unexpected error: %v", tc.input, err)
		}
		if tm.String() != tc.expected {
			t.Errorf("ParseTime(%q).String() = %q, expected %q", tc.input, tm.String(), tc.expected)
		}
	}
}

// TestTimeToDateTimeOnDefaultDate tests converting Time to DateTime on the fixed recovery date.
func TestTimeToDateTimeOnDefaultDate(t *testing.T) {
	tm, err := ParseTime("15:30:45Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dt := tm.toDateTimeOnDefaultDate()

	if dt.Hour() != 15 || dt.Minute() != 30 {
		t.Errorf("expected 15:30, got %02d:%02d", dt.Hour(), dt.Minute())
	}
	if dt.TimezoneOffset() == nil || dt.TimezoneOffset().offset != 0 {
		t.Errorf("expected UTC timezone offset, got %v", dt.TimezoneOffset())
	}
}

// TestTimeAdjust tests timezone adjustments on Time instances.
func TestTimeAdjust(t *testing.T) {
	tzPlus2, _ := NewTimezoneOffset(120)
	tzPlus5, _ := NewTimezoneOffset(300)

	unzoned, _ := ParseTime("10:00:00")
	adjustedZoned := unzoned.Adjust(&tzPlus2)
	if adjustedZoned.TimezoneOffset() == nil || adjustedZoned.TimezoneOffset().offset != 120 {
		t.Errorf("expected timezone offset 120, got %v", adjustedZoned.TimezoneOffset())
	}

	zoned, _ := ParseTime("10:00:00+02:00")
	adjustedUnzoned := zoned.Adjust(nil)
	if adjustedUnzoned.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", adjustedUnzoned.TimezoneOffset())
	}

	adjustedDiffZoned := zoned.Adjust(&tzPlus5)
	if adjustedDiffZoned.TimezoneOffset() == nil || adjustedDiffZoned.TimezoneOffset().offset != 300 {
		t.Errorf("expected timezone offset 300, got %v", adjustedDiffZoned.TimezoneOffset())
	}
}
