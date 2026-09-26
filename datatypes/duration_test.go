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
	"math"
	"testing"
)

// TestNewDuration tests creating new durations.
func TestNewDuration(t *testing.T) {
	dec10 := NewDecimalFromInt64(10)
	decNeg10 := NewDecimalFromInt64(-10)
	decZero := DefaultDecimal()

	d, err := NewDuration(12, dec10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmp, checkErr := d.AllSeconds().Compare(dec10)
	if checkErr != nil || cmp != 0 || d.AllMonths() != 12 {
		t.Errorf("unexpected duration value: %v", d)
	}

	d, err = NewDuration(-12, decNeg10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmp, checkErr = d.AllSeconds().Compare(decNeg10)
	if checkErr != nil || cmp != 0 || d.AllMonths() != -12 {
		t.Errorf("unexpected duration value: %v", d)
	}

	if _, err = NewDuration(0, dec10); err != nil {
		t.Errorf("unexpected error for zero months: %v", err)
	}
	if _, err = NewDuration(12, decZero); err != nil {
		t.Errorf("unexpected error for zero seconds: %v", err)
	}
	if _, err = NewDuration(0, decZero); err != nil {
		t.Errorf("unexpected error for all zeros: %v", err)
	}

	if _, err = NewDuration(12, decNeg10); !errors.Is(err, ErrOppositeSignInDurationComponents) {
		t.Errorf("expected ErrOppositeSignInDurationComponents, got %v", err)
	}
	if _, err = NewDuration(-12, dec10); !errors.Is(err, ErrOppositeSignInDurationComponents) {
		t.Errorf("expected ErrOppositeSignInDurationComponents, got %v", err)
	}
}

// TestParseDuration tests parsing duration strings.
func TestParseDuration(t *testing.T) {
	testCases := []struct {
		input    string
		wantErr  bool
		expected string
	}{
		{"P1Y2M3DT4H5M6S", false, "P1Y2M3DT4H5M6S"},
		{"-P1Y2M3DT4H5M6S", false, "-P1Y2M3DT4H5M6S"},
		{"PT0S", false, "PT0S"},
		{"P0Y", false, "PT0S"},
		{"P1Y", false, "P1Y"},
		{"P1M", false, "P1M"},
		{"P1D", false, "P1D"},
		{"PT1H", false, "PT1H"},
		{"PT1M", false, "PT1M"},
		{"PT1S", false, "PT1S"},
		{"PT0.5S", false, "PT0.5S"},
		{"-P1Y2M", false, "-P1Y2M"},

		{"1Y2M", true, ""},
		{"P1Y2M3DT4H5M6SX", true, ""},
		{"P", true, ""},
		{"PT", true, ""},
		{"P1Y2M3DT4H5M6ST7H", true, ""},
		{"P1K", true, ""},
		{"P2M1Y", true, ""},
		{"P1D1M", true, ""},
		{"PT1M1H", true, ""},
		{"PT1S1M", true, ""},
		{"P1H", true, ""},
		{"P1S", true, ""},
		{"P1.5Y", true, ""},
		{"P1.5M", true, ""},
		{"P1.5D", true, ""},
		{"PT1.5H", true, ""},
		{"PT1.5M", true, ""},
		{"P9999999999999999999999999999Y", true, ""},
		{"P9999999999999999999999999999M", true, ""},
	}

	for _, tc := range testCases {
		d, err := ParseDuration(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseDuration(%q) expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseDuration(%q) unexpected error: %v", tc.input, err)
			} else if d.String() != tc.expected {
				t.Errorf("ParseDuration(%q) = %q, expected %q", tc.input, d.String(), tc.expected)
			}
		}
	}
}

// TestDurationGetters tests duration component getters.
func TestDurationGetters(t *testing.T) {
	d, err := ParseDuration("P1Y2M3DT4H5M6.5S")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := d.Years(); got != 1 {
		t.Errorf("Years() = %d, want 1", got)
	}
	if got := d.Months(); got != 2 {
		t.Errorf("Months() = %d, want 2", got)
	}
	if got := d.Days(); got != 3 {
		t.Errorf("Days() = %d, want 3", got)
	}
	if got := d.Hours(); got != 4 {
		t.Errorf("Hours() = %d, want 4", got)
	}
	if got := d.Minutes(); got != 5 {
		t.Errorf("Minutes() = %d, want 5", got)
	}

	sec := d.Seconds()
	expectedSec, _ := ParseDecimal("6.5")
	cmp, checkErr := sec.Compare(expectedSec)
	if checkErr != nil || cmp != 0 {
		t.Errorf("Seconds() = %v, want 6.5", sec)
	}

	if got := d.AllMonths(); got != 14 {
		t.Errorf("AllMonths() = %d, want 14", got)
	}

	expectedAllSec, _ := ParseDecimal("273906.5")
	cmp, checkErr = d.AllSeconds().Compare(expectedAllSec)
	if checkErr != nil || cmp != 0 {
		t.Errorf("AllSeconds() = %v, want %v", d.AllSeconds(), expectedAllSec)
	}
}

// TestDurationCheckedAdd tests safe addition of durations.
func TestDurationCheckedAdd(t *testing.T) {
	d1, _ := ParseDuration("P1Y2M")
	d2, _ := ParseDuration("P2Y3M")

	res, err := d1.CheckedAdd(d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.String() != "P3Y5M" {
		t.Errorf("CheckedAdd result = %s, want P3Y5M", res.String())
	}

	overflowYM := Duration{yearMonth: NewYearMonthDuration(math.MaxInt64)}
	if _, err = overflowYM.CheckedAdd(d1); err == nil {
		t.Errorf("expected overflow error in CheckedAdd for yearMonth")
	}

	dPosYM, _ := ParseDuration("P2Y")
	dNegDT, _ := ParseDuration("-P1Y10D")
	if _, err = dPosYM.CheckedAdd(dNegDT); !errors.Is(err, ErrOppositeSignInDurationComponents) {
		t.Errorf("expected ErrOppositeSignInDurationComponents, got %v", err)
	}
}

// TestDurationCheckedSub tests safe subtraction of durations.
func TestDurationCheckedSub(t *testing.T) {
	d1, _ := ParseDuration("P3Y5M")
	d2, _ := ParseDuration("P1Y2M")

	res, err := d1.CheckedSub(d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.String() != "P2Y3M" {
		t.Errorf("CheckedSub result = %s, want P2Y3M", res.String())
	}

	underflowYM := Duration{yearMonth: NewYearMonthDuration(math.MinInt64)}
	if _, err = underflowYM.CheckedSub(d2); err == nil {
		t.Errorf("expected underflow error in CheckedSub for yearMonth")
	}

	dPosYM, _ := ParseDuration("P2Y")
	dPosBoth, _ := ParseDuration("P1Y10D")
	if _, err = dPosYM.CheckedSub(dPosBoth); !errors.Is(err, ErrOppositeSignInDurationComponents) {
		t.Errorf("expected ErrOppositeSignInDurationComponents, got %v", err)
	}
}

// TestDurationCheckedNeg tests safe negation of durations.
func TestDurationCheckedNeg(t *testing.T) {
	d1, _ := ParseDuration("P1Y2M3DT4H5M6S")
	neg, err := d1.CheckedNeg()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if neg.String() != "-P1Y2M3DT4H5M6S" {
		t.Errorf("CheckedNeg result = %s, want -P1Y2M3DT4H5M6S", neg.String())
	}

	minYM := Duration{yearMonth: NewYearMonthDuration(math.MinInt64)}
	if _, err = minYM.CheckedNeg(); err == nil {
		t.Errorf("expected error when negating MinInt64 months")
	}
}

// TestDurationIsIdenticalWith tests exact identity comparison for durations.
func TestDurationIsIdenticalWith(t *testing.T) {
	d1, _ := ParseDuration("P1Y2M")
	d2, _ := ParseDuration("P1Y2M")
	d3, _ := ParseDuration("P1Y3M")

	if !d1.IsIdenticalWith(d2) {
		t.Errorf("expected d1 to be identical with d2")
	}
	if d1.IsIdenticalWith(d3) {
		t.Errorf("expected d1 not to be identical with d3")
	}
	if d1.IsIdenticalWith(String("P1Y2M")) {
		t.Errorf("expected d1 not to be identical with non-Duration type")
	}
}

// TestDurationCompare tests comparison between durations.
func TestDurationCompare(t *testing.T) {
	d1, _ := ParseDuration("P1Y")
	d2, _ := ParseDuration("P2Y")
	d3, _ := ParseDuration("P1Y")

	cmp, err := d1.Compare(d2)
	if err != nil || cmp != -1 {
		t.Errorf("d1.Compare(d2) = %d, %v; want -1, nil", cmp, err)
	}

	cmp, err = d2.Compare(d1)
	if err != nil || cmp != 1 {
		t.Errorf("d2.Compare(d1) = %d, %v; want 1, nil", cmp, err)
	}

	cmp, err = d1.Compare(d3)
	if err != nil || cmp != 0 {
		t.Errorf("d1.Compare(d3) = %d, %v; want 0, nil", cmp, err)
	}

	if _, err = d1.Compare(String("P1Y")); !errors.Is(err, ErrDurationOverflow) {
		t.Errorf("expected ErrDurationOverflow for incomparable type, got %v", err)
	}

	dMonth, _ := ParseDuration("P1M")
	dDays, _ := ParseDuration("P30D")
	if _, err = dMonth.Compare(dDays); !errors.Is(err, ErrDurationOverflow) {
		t.Errorf("expected ErrDurationOverflow for indeterminate comparison, got %v", err)
	}
}

// TestDurationStringEdgeCases tests string representation for edge case durations.
func TestDurationStringEdgeCases(t *testing.T) {
	dInvalid := Duration{
		yearMonth: NewYearMonthDuration(-12),
		dayTime:   NewDayTimeDuration(NewDecimalFromInt64(10)),
	}
	if got := dInvalid.String(); got != "invalid-duration-opposite-signs" {
		t.Errorf("String() = %q, want %q", got, "invalid-duration-opposite-signs")
	}

	dInvalid2 := Duration{
		yearMonth: NewYearMonthDuration(12),
		dayTime:   NewDayTimeDuration(NewDecimalFromInt64(-10)),
	}
	if got := dInvalid2.String(); got != "invalid-duration-opposite-signs" {
		t.Errorf("String() = %q, want %q", got, "invalid-duration-opposite-signs")
	}
}

// TestParseDurationErrorFormatting tests parsing error formatting and unwrapping.
func TestParseDurationErrorFormatting(t *testing.T) {
	errSimple := newParseDurationError("invalid syntax")
	if got := errSimple.Error(); got != "parsing duration: invalid syntax" {
		t.Errorf("Error() = %q, want %q", got, "parsing duration: invalid syntax")
	}
	if errSimple.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", errSimple.Unwrap())
	}

	errWrapped := ParseDurationError{msg: "overflow", err: ErrDurationOverflow}
	expectedMsg := "parsing duration: overflow: overflow during xsd:duration computation"
	if got := errWrapped.Error(); got != expectedMsg {
		t.Errorf("Error() = %q, want %q", got, expectedMsg)
	}
	if !errors.Is(errWrapped.Unwrap(), ErrDurationOverflow) {
		t.Errorf("Unwrap() = %v, want ErrDurationOverflow", errWrapped.Unwrap())
	}
}
