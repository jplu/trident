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

// mustParseGMonth is a test helper that parses a gMonth literal and panics on failure.
func mustParseGMonth(tb testing.TB, s string) GMonth {
	tb.Helper()
	g, err := ParseGMonth(s)
	if err != nil {
		tb.Fatalf("mustParseGMonth(%q) unexpected error: %v", s, err)
	}
	return g
}

// mustTimezoneOffset is a test helper that creates a TimezoneOffset from minutes and panics on failure.
func mustTimezoneOffset(tb testing.TB, minutes int16) *TimezoneOffset {
	tb.Helper()
	tz, err := NewTimezoneOffset(minutes)
	if err != nil {
		tb.Fatalf("mustTimezoneOffset(%d) unexpected error: %v", minutes, err)
	}
	return &tz
}

// TestNewGMonth tests the instantiation of GMonth instances using NewGMonth.
func TestNewGMonth(t *testing.T) {
	tz := mustTimezoneOffset(t, 120)

	tests := []struct {
		name      string
		month     uint8
		tz        *TimezoneOffset
		wantMonth uint8
		wantTz    *TimezoneOffset
		wantErr   bool
	}{
		{
			name:      "unzoned month 1",
			month:     1,
			tz:        nil,
			wantMonth: 1,
			wantTz:    nil,
			wantErr:   false,
		},
		{
			name:      "zoned month 11 with +02:00",
			month:     11,
			tz:        tz,
			wantMonth: 11,
			wantTz:    tz,
			wantErr:   false,
		},
		{
			name:      "zoned month 7 with UTC",
			month:     7,
			tz:        GetUTC(),
			wantMonth: 7,
			wantTz:    GetUTC(),
			wantErr:   false,
		},
		{
			name:      "invalid month 0",
			month:     0,
			tz:        nil,
			wantMonth: 0,
			wantTz:    nil,
			wantErr:   true,
		},
		{
			name:      "invalid month 13",
			month:     13,
			tz:        nil,
			wantMonth: 0,
			wantTz:    nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm, err := NewGMonth(tt.month, tt.tz)
			assertNewGMonthResult(t, gm, err, tt.wantMonth, tt.wantTz, tt.wantErr)
		})
	}
}

// assertNewGMonthResult validates the result of NewGMonth against expected values and error expectations.
func assertNewGMonthResult(
	t *testing.T,
	gm GMonth,
	err error,
	wantMonth uint8,
	wantTz *TimezoneOffset,
	wantErr bool,
) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Fatal("NewGMonth expected error, got nil")
		}
		return
	}
	if err != nil {
		t.Fatalf("NewGMonth unexpected error: %v", err)
	}
	if gm.Month() != wantMonth {
		t.Errorf("Month() = %d, want %d", gm.Month(), wantMonth)
	}
	assertGMonthTimezone(t, gm.TimezoneOffset(), wantTz)
}

// assertGMonthTimezone asserts that the observed timezone offset matches the expected offset.
func assertGMonthTimezone(t *testing.T, gotTz, wantTz *TimezoneOffset) {
	t.Helper()
	if wantTz == nil {
		if gotTz != nil {
			t.Errorf("TimezoneOffset() = %v, want nil", gotTz)
		}
		return
	}
	if gotTz == nil || *gotTz != *wantTz {
		t.Errorf("TimezoneOffset() = %v, want %v", gotTz, wantTz)
	}
}

// TestParseGMonth tests the lexical parsing of xsd:gMonth literals.
func TestParseGMonth(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMonth uint8
		wantTz    *TimezoneOffset
		wantErr   bool
	}{
		{
			name:      "valid unzoned month",
			input:     "--05",
			wantMonth: 5,
			wantTz:    nil,
			wantErr:   false,
		},
		{
			name:      "valid month with UTC",
			input:     "--11Z",
			wantMonth: 11,
			wantTz:    GetUTC(),
			wantErr:   false,
		},
		{
			name:      "valid month with positive offset",
			input:     "--01+05:30",
			wantMonth: 1,
			wantTz:    mustTimezoneOffset(t, 330),
			wantErr:   false,
		},
		{
			name:      "valid month with negative offset",
			input:     "--09-08:00",
			wantMonth: 9,
			wantTz:    mustTimezoneOffset(t, -480),
			wantErr:   false,
		},
		{
			name:      "valid month 12 parsed successfully and maps to timeline",
			input:     "--12",
			wantMonth: 12,
			wantTz:    nil,
			wantErr:   false,
		},
		{
			name:    "empty string fails first hyphen check",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing leading hyphen fails first hyphen check",
			input:   "05",
			wantErr: true,
		},
		{
			name:    "single leading hyphen fails second hyphen check",
			input:   "-05",
			wantErr: true,
		},
		{
			name:    "three leading hyphens fails two-digit month check",
			input:   "---05",
			wantErr: true,
		},
		{
			name:    "single digit month fails two-digit month check",
			input:   "--5",
			wantErr: true,
		},
		{
			name:    "three digit month fails two-digit month check",
			input:   "--005",
			wantErr: true,
		},
		{
			name:    "month zero is out of bounds",
			input:   "--00",
			wantErr: true,
		},
		{
			name:    "month thirteen is out of bounds",
			input:   "--13",
			wantErr: true,
		},
		{
			name:    "non-numeric month fails two-digit month check",
			input:   "--ab",
			wantErr: true,
		},
		{
			name:    "unrecognized trailing characters without timezone",
			input:   "--05trailing",
			wantErr: true,
		},
		{
			name:    "unrecognized trailing characters after timezone",
			input:   "--05Zextra",
			wantErr: true,
		},
		{
			name:    "unrecognized trailing characters after offset",
			input:   "--05+02:00extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm, err := ParseGMonth(tt.input)
			verifyParseResult(t, tt.input, gm, err, tt.wantMonth, tt.wantTz, tt.wantErr)
		})
	}
}

// verifyParseResult validates the result of ParseGMonth against expected month, timezone, and error expectations.
func verifyParseResult(
	t *testing.T,
	input string,
	gm GMonth,
	err error,
	wantMonth uint8,
	wantTz *TimezoneOffset,
	wantErr bool,
) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Fatalf("ParseGMonth(%q) expected error, got nil", input)
		}
		return
	}
	if err != nil {
		t.Fatalf("ParseGMonth(%q) unexpected error: %v", input, err)
	}
	if gm.Month() != wantMonth {
		t.Errorf("Month() = %d, want %d", gm.Month(), wantMonth)
	}
	assertGMonthTimezone(t, gm.TimezoneOffset(), wantTz)
}

// TestGMonthMonth tests the Month accessor method of GMonth across all calendar months.
func TestGMonthMonth(t *testing.T) {
	for m := uint8(1); m <= 12; m++ {
		gm, err := NewGMonth(m, nil)
		if err != nil {
			t.Fatalf("NewGMonth(%d, nil) error: %v", m, err)
		}
		if gm.Month() != m {
			t.Errorf("Month() = %d, want %d", gm.Month(), m)
		}
	}
}

// TestGMonthTimezoneOffset tests the TimezoneOffset accessor method of GMonth for both zoned and unzoned instances.
func TestGMonthTimezoneOffset(t *testing.T) {
	unzoned := mustParseGMonth(t, "--04")
	if unzoned.TimezoneOffset() != nil {
		t.Errorf("TimezoneOffset() for unzoned = %v, want nil", unzoned.TimezoneOffset())
	}

	zoned := mustParseGMonth(t, "--04+01:00")
	expectedTz := mustTimezoneOffset(t, 60)
	if zoned.TimezoneOffset() == nil || *zoned.TimezoneOffset() != *expectedTz {
		t.Errorf("TimezoneOffset() for zoned = %v, want %v", zoned.TimezoneOffset(), expectedTz)
	}
}

// TestGMonthAdjust tests the timezone adjustment behavior of GMonth via the Adjust method.
func TestGMonthAdjust(t *testing.T) {
	tz1 := mustTimezoneOffset(t, 60)
	tz2 := mustTimezoneOffset(t, 120)

	t.Run("unzoned to zoned", func(t *testing.T) {
		gm := mustParseGMonth(t, "--06")
		adj := gm.Adjust(tz1)
		if adj.TimezoneOffset() == nil || *adj.TimezoneOffset() != *tz1 {
			t.Errorf("TimezoneOffset() = %v, want %v", adj.TimezoneOffset(), tz1)
		}
	})

	t.Run("zoned to unzoned", func(t *testing.T) {
		gm := mustParseGMonth(t, "--06+01:00")
		adj := gm.Adjust(nil)
		if adj.TimezoneOffset() != nil {
			t.Errorf("TimezoneOffset() = %v, want nil", adj.TimezoneOffset())
		}
	})

	t.Run("zoned to different timezone", func(t *testing.T) {
		gm := mustParseGMonth(t, "--06+01:00")
		adj := gm.Adjust(tz2)
		if adj.TimezoneOffset() == nil || *adj.TimezoneOffset() != *tz2 {
			t.Errorf("TimezoneOffset() = %v, want %v", adj.TimezoneOffset(), tz2)
		}
	})

	t.Run("unzoned to unzoned remains unzoned", func(t *testing.T) {
		gm := mustParseGMonth(t, "--06")
		adj := gm.Adjust(nil)
		if adj.TimezoneOffset() != nil {
			t.Errorf("TimezoneOffset() = %v, want nil", adj.TimezoneOffset())
		}
	})
}

// TestGMonthIsIdenticalWith tests the identity comparison rules.
func TestGMonthIsIdenticalWith(t *testing.T) {
	gm1 := mustParseGMonth(t, "--05")
	gm2 := mustParseGMonth(t, "--05")
	gmDiffMonth := mustParseGMonth(t, "--06")
	gmZoned1 := mustParseGMonth(t, "--05Z")
	gmZoned2 := mustParseGMonth(t, "--05Z")
	gmZonedDiffTz := mustParseGMonth(t, "--05+01:00")
	nonGMonth := String("--05")

	tests := []struct {
		name  string
		a     GMonth
		other XSDValue
		want  bool
	}{
		{
			name:  "identical unzoned months",
			a:     gm1,
			other: gm2,
			want:  true,
		},
		{
			name:  "different unzoned months",
			a:     gm1,
			other: gmDiffMonth,
			want:  false,
		},
		{
			name:  "unzoned compared to zoned",
			a:     gm1,
			other: gmZoned1,
			want:  false,
		},
		{
			name:  "zoned compared to unzoned",
			a:     gmZoned1,
			other: gm1,
			want:  false,
		},
		{
			name:  "identical zoned months",
			a:     gmZoned1,
			other: gmZoned2,
			want:  true,
		},
		{
			name:  "zoned months with different offsets",
			a:     gmZoned1,
			other: gmZonedDiffTz,
			want:  false,
		},
		{
			name:  "incompatible XSDValue type",
			a:     gm1,
			other: nonGMonth,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.IsIdenticalWith(tt.other)
			if got != tt.want {
				t.Errorf("IsIdenticalWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGMonthCompare tests the partial ordering relation of GMonth instances.
func TestGMonthCompare(t *testing.T) {
	gmJan := mustParseGMonth(t, "--01")
	gmFeb := mustParseGMonth(t, "--02")
	gmJanAgain := mustParseGMonth(t, "--01")

	gmJanZ := mustParseGMonth(t, "--01Z")
	gmFebZ := mustParseGMonth(t, "--02Z")
	gmJanZAgain := mustParseGMonth(t, "--01Z")

	nonGMonth := String("--01")

	t.Run("unzoned comparisons", func(t *testing.T) {
		testCompareUnzoned(t, gmJan, gmFeb, gmJanAgain)
	})

	t.Run("zoned comparisons", func(t *testing.T) {
		testCompareZoned(t, gmJanZ, gmFebZ, gmJanZAgain)
	})

	t.Run("indeterminate comparison between zoned and unzoned with overlap", func(t *testing.T) {
		testCompareOverlap(t)
	})

	t.Run("determinate comparison between zoned and unzoned without overlap", func(t *testing.T) {
		testCompareNoOverlap(t, gmJanZ)
	})

	t.Run("incomparable types return ErrDateTimeOverflow", func(t *testing.T) {
		testCompareIncomparable(t, gmJan, nonGMonth)
	})
}

// testCompareUnzoned tests the order comparison of unzoned GMonth instances.
func testCompareUnzoned(t *testing.T, gmJan, gmFeb, gmJanAgain GMonth) {
	t.Helper()
	res, err := gmJan.Compare(gmFeb)
	if err != nil || res != -1 {
		t.Errorf("Compare(gmJan, gmFeb) = (%d, %v), want (-1, nil)", res, err)
	}

	res, err = gmFeb.Compare(gmJan)
	if err != nil || res != 1 {
		t.Errorf("Compare(gmFeb, gmJan) = (%d, %v), want (1, nil)", res, err)
	}

	res, err = gmJan.Compare(gmJanAgain)
	if err != nil || res != 0 {
		t.Errorf("Compare(gmJan, gmJanAgain) = (%d, %v), want (0, nil)", res, err)
	}
}

// testCompareZoned tests the order comparison of zoned GMonth instances.
func testCompareZoned(t *testing.T, gmJanZ, gmFebZ, gmJanZAgain GMonth) {
	t.Helper()
	res, err := gmJanZ.Compare(gmFebZ)
	if err != nil || res != -1 {
		t.Errorf("Compare(gmJanZ, gmFebZ) = (%d, %v), want (-1, nil)", res, err)
	}

	res, err = gmFebZ.Compare(gmJanZ)
	if err != nil || res != 1 {
		t.Errorf("Compare(gmFebZ, gmJanZ) = (%d, %v), want (1, nil)", res, err)
	}

	res, err = gmJanZ.Compare(gmJanZAgain)
	if err != nil || res != 0 {
		t.Errorf("Compare(gmJanZ, gmJanZAgain) = (%d, %v), want (0, nil)", res, err)
	}
}

// testCompareOverlap tests the indeterminate comparison behavior between overlapping zoned and unzoned GMonth
// instances.
func testCompareOverlap(t *testing.T) {
	t.Helper()
	gmMayUnzoned := mustParseGMonth(t, "--05")
	gmMayZoned := mustParseGMonth(t, "--05Z")

	_, err := gmMayUnzoned.Compare(gmMayZoned)
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("Compare(gmMayUnzoned, gmMayZoned) error = %v, want ErrDateTimeOverflow", err)
	}

	_, err = gmMayZoned.Compare(gmMayUnzoned)
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("Compare(gmMayZoned, gmMayUnzoned) error = %v, want ErrDateTimeOverflow", err)
	}
}

// testCompareNoOverlap tests the determinate comparison behavior between non-overlapping zoned and unzoned GMonth
// instances.
func testCompareNoOverlap(t *testing.T, gmJanZ GMonth) {
	t.Helper()
	gmNov := mustParseGMonth(t, "--11")
	res, err := gmJanZ.Compare(gmNov)
	if err != nil || res != 1 {
		t.Errorf("Compare(gmJanZ, gmNov) = (%d, %v), want (1, nil)", res, err)
	}

	res, err = gmNov.Compare(gmJanZ)
	if err != nil || res != -1 {
		t.Errorf("Compare(gmNov, gmJanZ) = (%d, %v), want (-1, nil)", res, err)
	}
}

// testCompareIncomparable tests that comparing a GMonth instance with an incompatible XSDValue returns an error.
func testCompareIncomparable(t *testing.T, gmJan GMonth, nonGMonth XSDValue) {
	t.Helper()
	_, err := gmJan.Compare(nonGMonth)
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("Compare(gmJan, nonGMonth) error = %v, want ErrDateTimeOverflow", err)
	}
}

// TestGMonthString tests the canonical string formatting of GMonth instances.
func TestGMonthString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unzoned single digit month padded to 2 digits",
			input: "--05",
			want:  "--05",
		},
		{
			name:  "unzoned double digit month",
			input: "--11",
			want:  "--11",
		},
		{
			name:  "unzoned month 12 maps to canonical format",
			input: "--12",
			want:  "--12",
		},
		{
			name:  "zoned with UTC",
			input: "--07Z",
			want:  "--07Z",
		},
		{
			name:  "zoned with positive offset",
			input: "--03+02:00",
			want:  "--03+02:00",
		},
		{
			name:  "zoned with negative offset",
			input: "--10-05:00",
			want:  "--10-05:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := mustParseGMonth(t, tt.input)
			if got := gm.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
