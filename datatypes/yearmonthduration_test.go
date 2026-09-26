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

// TestNewYearMonthDuration tests the constructor NewYearMonthDuration.
func TestNewYearMonthDuration(t *testing.T) {
	ym := NewYearMonthDuration(25)
	if ym.months != 25 {
		t.Fatalf("expected 25 months, got %d", ym.months)
	}
}

// TestDefaultYearMonthDuration tests the zero-value constructor DefaultYearMonthDuration.
func TestDefaultYearMonthDuration(t *testing.T) {
	ym := DefaultYearMonthDuration()
	if ym.months != 0 {
		t.Fatalf("expected 0 months, got %d", ym.months)
	}
}

// TestYearMonthDurationAccessors tests the Years, Months, and AllMonths accessor methods.
func TestYearMonthDurationAccessors(t *testing.T) {
	tests := []struct {
		months     int64
		wantYears  int64
		wantMonths int64
		wantAll    int64
	}{
		{months: 25, wantYears: 2, wantMonths: 1, wantAll: 25},
		{months: -25, wantYears: -2, wantMonths: -1, wantAll: -25},
		{months: 0, wantYears: 0, wantMonths: 0, wantAll: 0},
		{months: 11, wantYears: 0, wantMonths: 11, wantAll: 11},
		{months: -11, wantYears: 0, wantMonths: -11, wantAll: -11},
	}

	for _, tc := range tests {
		ym := NewYearMonthDuration(tc.months)
		if got := ym.Years(); got != tc.wantYears {
			t.Errorf("Years() for %d months = %d, want %d", tc.months, got, tc.wantYears)
		}
		if got := ym.Months(); got != tc.wantMonths {
			t.Errorf("Months() for %d months = %d, want %d", tc.months, got, tc.wantMonths)
		}
		if got := ym.AllMonths(); got != tc.wantAll {
			t.Errorf("AllMonths() for %d months = %d, want %d", tc.months, got, tc.wantAll)
		}
	}
}

// TestYearMonthDurationCheckedAdd tests checked addition including overflow conditions.
func TestYearMonthDurationCheckedAdd(t *testing.T) {
	ym1 := NewYearMonthDuration(10)
	ym2 := NewYearMonthDuration(20)
	res, err := ym1.CheckedAdd(ym2)
	if err != nil || res.months != 30 {
		t.Fatalf("expected 30 months without error, got %d, err: %v", res.months, err)
	}

	maxYM := NewYearMonthDuration(math.MaxInt64)
	posYM := NewYearMonthDuration(1)
	if _, err = maxYM.CheckedAdd(posYM); !errors.Is(err, ErrDurationOverflow) {
		t.Fatalf("expected ErrDurationOverflow on positive overflow, got %v", err)
	}

	minYM := NewYearMonthDuration(math.MinInt64)
	negYM := NewYearMonthDuration(-1)
	if _, err = minYM.CheckedAdd(negYM); !errors.Is(err, ErrDurationOverflow) {
		t.Fatalf("expected ErrDurationOverflow on negative overflow, got %v", err)
	}
}

// TestYearMonthDurationCheckedSub tests checked subtraction including overflow conditions.
func TestYearMonthDurationCheckedSub(t *testing.T) {
	ym1 := NewYearMonthDuration(30)
	ym2 := NewYearMonthDuration(10)
	res, err := ym1.CheckedSub(ym2)
	if err != nil || res.months != 20 {
		t.Fatalf("expected 20 months without error, got %d, err: %v", res.months, err)
	}

	minYM := NewYearMonthDuration(math.MinInt64)
	posYM := NewYearMonthDuration(1)
	if _, err = minYM.CheckedSub(posYM); !errors.Is(err, ErrDurationOverflow) {
		t.Fatalf("expected ErrDurationOverflow on underflow, got %v", err)
	}

	maxYM := NewYearMonthDuration(math.MaxInt64)
	negYM := NewYearMonthDuration(-1)
	if _, err = maxYM.CheckedSub(negYM); !errors.Is(err, ErrDurationOverflow) {
		t.Fatalf("expected ErrDurationOverflow on overflow, got %v", err)
	}
}

// TestYearMonthDurationCheckedNeg tests checked negation including boundary conditions.
func TestYearMonthDurationCheckedNeg(t *testing.T) {
	ym := NewYearMonthDuration(15)
	neg, err := ym.CheckedNeg()
	if err != nil || neg.months != -15 {
		t.Fatalf("expected -15 months without error, got %d, err: %v", neg.months, err)
	}

	minYM := NewYearMonthDuration(math.MinInt64)
	if _, err = minYM.CheckedNeg(); !errors.Is(err, ErrDurationOverflow) {
		t.Fatalf("expected ErrDurationOverflow on MinInt64 negation, got %v", err)
	}
}

// TestYearMonthDurationIsIdenticalWith tests identity comparison with matching, mismatched, and other types.
func TestYearMonthDurationIsIdenticalWith(t *testing.T) {
	ym1 := NewYearMonthDuration(12)
	ym2 := NewYearMonthDuration(12)
	ym3 := NewYearMonthDuration(24)

	if !ym1.IsIdenticalWith(ym2) {
		t.Fatalf("expected identical YearMonthDurations to return true")
	}
	if ym1.IsIdenticalWith(ym3) {
		t.Fatalf("expected non-identical YearMonthDurations to return false")
	}
	if ym1.IsIdenticalWith(String("P1Y")) {
		t.Fatalf("expected comparison with different XSDValue type to return false")
	}
}

// TestYearMonthDurationCompare tests the partial ordering evaluation of YearMonthDuration.
func TestYearMonthDurationCompare(t *testing.T) {
	ymSmall := NewYearMonthDuration(10)
	ymLarge := NewYearMonthDuration(20)
	ymEqual := NewYearMonthDuration(10)

	cmp, err := ymSmall.Compare(ymLarge)
	if err != nil || cmp != -1 {
		t.Fatalf("expected -1 without error, got %d, err: %v", cmp, err)
	}

	cmp, err = ymLarge.Compare(ymSmall)
	if err != nil || cmp != 1 {
		t.Fatalf("expected 1 without error, got %d, err: %v", cmp, err)
	}

	cmp, err = ymSmall.Compare(ymEqual)
	if err != nil || cmp != 0 {
		t.Fatalf("expected 0 without error, got %d, err: %v", cmp, err)
	}

	if _, err = ymSmall.Compare(String("P10M")); !errors.Is(err, ErrDurationOverflow) {
		t.Fatalf("expected ErrDurationOverflow when comparing with another datatype, got %v", err)
	}
}

// TestYearMonthDurationString tests serialization to canonical string format.
func TestYearMonthDurationString(t *testing.T) {
	tests := []struct {
		months int64
		want   string
	}{
		{months: 0, want: "P0M"},
		{months: 12, want: "P1Y"},
		{months: 14, want: "P1Y2M"},
		{months: 2, want: "P2M"},
		{months: -12, want: "-P1Y"},
		{months: -14, want: "-P1Y2M"},
		{months: -2, want: "-P2M"},
	}

	for _, tc := range tests {
		ym := NewYearMonthDuration(tc.months)
		if got := ym.String(); got != tc.want {
			t.Errorf("String() for %d months = %s, want %s", tc.months, got, tc.want)
		}
	}
}

// TestParseYearMonthDuration tests parsing of valid and invalid lexical representations.
func TestParseYearMonthDuration(t *testing.T) {
	tests := []struct {
		input      string
		wantMonths int64
		wantErr    bool
	}{
		{input: "P1Y2M", wantMonths: 14, wantErr: false},
		{input: "-P1Y2M", wantMonths: -14, wantErr: false},
		{input: "P2Y", wantMonths: 24, wantErr: false},
		{input: "P6M", wantMonths: 6, wantErr: false},
		{input: "P0M", wantMonths: 0, wantErr: false},
		{input: "invalid", wantErr: true},
		{input: "P1Yextra", wantErr: true},
		{input: "P1D", wantErr: true},
		{input: "P1YT1S", wantErr: true},
		{input: "P", wantErr: true},
	}

	for _, tc := range tests {
		got, err := ParseYearMonthDuration(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseYearMonthDuration(%q) expected error, got nil", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseYearMonthDuration(%q) unexpected error: %v", tc.input, err)
			continue
		}
		if got.months != tc.wantMonths {
			t.Errorf("ParseYearMonthDuration(%q) = %d months, want %d", tc.input, got.months, tc.wantMonths)
		}
	}
}
