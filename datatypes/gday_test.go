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

// TestNewGDay tests creating valid and invalid GDay instances with and without timezones.
func TestNewGDay(t *testing.T) {
	t.Run("valid unzoned", func(t *testing.T) {
		g, err := NewGDay(15, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g.Day() != 15 {
			t.Errorf("expected day 15, got %d", g.Day())
		}
		if g.TimezoneOffset() != nil {
			t.Errorf("expected nil timezone offset, got %v", g.TimezoneOffset())
		}
	})

	t.Run("valid zoned", func(t *testing.T) {
		tz := GetUTC()
		g, err := NewGDay(28, tz)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if g.Day() != 28 {
			t.Errorf("expected day 28, got %d", g.Day())
		}
		if g.TimezoneOffset() == nil || *g.TimezoneOffset() != *tz {
			t.Errorf("expected timezone offset %v, got %v", tz, g.TimezoneOffset())
		}
	})

	t.Run("invalid day zero", func(t *testing.T) {
		_, err := NewGDay(0, nil)
		if err == nil {
			t.Fatalf("expected error for day 0, got nil")
		}
	})

	t.Run("invalid day exceeds 31", func(t *testing.T) {
		_, err := NewGDay(32, nil)
		if err == nil {
			t.Fatalf("expected error for day 32, got nil")
		}
	})
}

// TestParseGDay tests parsing various valid and invalid GDay string representations.
func TestParseGDay(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantDay     uint8
		hasTz       bool
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid day unzoned",
			input:   "---01",
			wantDay: 1,
			hasTz:   false,
		},
		{
			name:    "valid day 30 unzoned",
			input:   "---30",
			wantDay: 30,
			hasTz:   false,
		},
		{
			name:    "valid day 31 unzoned",
			input:   "---31",
			wantDay: 31,
			hasTz:   false,
		},
		{
			name:    "valid day with UTC timezone",
			input:   "---15Z",
			wantDay: 15,
			hasTz:   true,
		},
		{
			name:    "valid day with positive timezone",
			input:   "---15+05:30",
			wantDay: 15,
			hasTz:   true,
		},
		{
			name:    "valid day with negative timezone",
			input:   "---15-08:00",
			wantDay: 15,
			hasTz:   true,
		},
		{
			name:        "missing first hyphen",
			input:       "--15",
			wantErr:     true,
			errContains: "gDay values must start with '---'",
		},
		{
			name:        "missing second hyphen",
			input:       "-15",
			wantErr:     true,
			errContains: "gDay values must start with '---'",
		},
		{
			name:        "missing third hyphen",
			input:       "--15Z",
			wantErr:     true,
			errContains: "gDay values must start with '---'",
		},
		{
			name:        "no hyphens",
			input:       "15",
			wantErr:     true,
			errContains: "gDay values must start with '---'",
		},
		{
			name:        "invalid day digits: too short",
			input:       "---1",
			wantErr:     true,
			errContains: "day must be encoded with two digits",
		},
		{
			name:        "invalid day digits: non-numeric",
			input:       "---XX",
			wantErr:     true,
			errContains: "day must be encoded with two digits",
		},
		{
			name:        "invalid day value: zero",
			input:       "---00",
			wantErr:     true,
			errContains: "day must be between 01 and 31",
		},
		{
			name:        "invalid day value: exceeds 31",
			input:       "---32",
			wantErr:     true,
			errContains: "day must be between 01 and 31",
		},
		{
			name:        "unrecognized trailing suffix",
			input:       "---15extra",
			wantErr:     true,
			errContains: "unrecognized value suffix",
		},
		{
			name:        "unrecognized suffix after timezone",
			input:       "---15Zinvalid",
			wantErr:     true,
			errContains: "unrecognized value suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGDay(tt.input)
			assertParseGDayResult(t, tt.input, tt.wantDay, tt.hasTz, tt.wantErr, tt.errContains, got, err)
		})
	}
}

// assertParseGDayResult asserts that the ParseGDay invocation returned the expected GDay value or error.
func assertParseGDayResult(
	t *testing.T,
	input string,
	wantDay uint8,
	hasTz bool,
	wantErr bool,
	errContains string,
	got GDay,
	err error,
) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Fatalf("expected error for input %q, got nil", input)
		}
		if errContains != "" && !errors.Is(err, ErrDateTimeOverflow) {
			if err.Error() == "" || (len(errContains) > 0 && !containsString(err.Error(), errContains)) {
				t.Errorf("error %q does not contain %q", err.Error(), errContains)
			}
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error for input %q: %v", input, err)
	}
	if got.Day() != wantDay {
		t.Errorf("expected day %d, got %d", wantDay, got.Day())
	}
	if (got.TimezoneOffset() != nil) != hasTz {
		t.Errorf("expected hasTz=%v, got TimezoneOffset=%v", hasTz, got.TimezoneOffset())
	}
}

// TestGDayDayAndTimezoneOffset verifies the Day and TimezoneOffset accessor methods.
func TestGDayDayAndTimezoneOffset(t *testing.T) {
	tz, _ := NewTimezoneOffset(120)
	g, err := NewGDay(25, &tz)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.Day() != 25 {
		t.Errorf("expected day 25, got %d", g.Day())
	}
	if g.TimezoneOffset() == nil || g.TimezoneOffset().offset != 120 {
		t.Errorf("expected timezone offset 120, got %v", g.TimezoneOffset())
	}

	gUnzoned, err := NewGDay(10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gUnzoned.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset for unzoned GDay, got %v", gUnzoned.TimezoneOffset())
	}
}

// TestGDayAdjust tests adjusting a GDay instance to different timezones, including error injection.
func TestGDayAdjust(t *testing.T) {
	tzUTC := GetUTC()
	tzPlus2, _ := NewTimezoneOffset(120)
	tzMinus5, _ := NewTimezoneOffset(-300)

	tests := []struct {
		name        string
		input       string
		targetTz    *TimezoneOffset
		expectedDay uint8
		expectedTz  *TimezoneOffset
	}{
		{
			name:       "zoned to different zone",
			input:      "---15Z",
			targetTz:   &tzPlus2,
			expectedTz: &tzPlus2,
		},
		{
			name:       "zoned to unzoned",
			input:      "---15Z",
			targetTz:   nil,
			expectedTz: nil,
		},
		{
			name:       "unzoned to zoned",
			input:      "---15",
			targetTz:   tzUTC,
			expectedTz: tzUTC,
		},
		{
			name:        "unzoned to unzoned",
			input:       "---15",
			targetTz:    nil,
			expectedDay: 15,
			expectedTz:  nil,
		},
		{
			name:       "zoned with negative offset to zoned with positive offset",
			input:      "---15-05:00",
			targetTz:   &tzPlus2,
			expectedTz: &tzPlus2,
		},
		{
			name:       "zoned with negative offset to unzoned",
			input:      "---15-05:00",
			targetTz:   nil,
			expectedTz: nil,
		},
		{
			name:       "unzoned to zoned with negative offset",
			input:      "---15",
			targetTz:   &tzMinus5,
			expectedTz: &tzMinus5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, _ := ParseGDay(tt.input)
			adj := g.Adjust(tt.targetTz)
			if tt.expectedTz == nil {
				if adj.TimezoneOffset() != nil {
					t.Errorf("expected nil timezone, got %v", adj.TimezoneOffset())
				}
			} else {
				if adj.TimezoneOffset() == nil || *adj.TimezoneOffset() != *tt.expectedTz {
					t.Errorf("expected timezone %v, got %v", *tt.expectedTz, adj.TimezoneOffset())
				}
			}
			if tt.expectedDay != 0 && adj.Day() != tt.expectedDay {
				t.Errorf("expected day %d, got %d", tt.expectedDay, adj.Day())
			}
		})
	}
}

// TestGDayIsIdenticalWith tests exact equality checks between GDay values and other types.
func TestGDayIsIdenticalWith(t *testing.T) {
	g15, _ := ParseGDay("---15")
	g15Dup, _ := ParseGDay("---15")
	g16, _ := ParseGDay("---16")
	g15Z, _ := ParseGDay("---15Z")
	g15ZDup, _ := ParseGDay("---15Z")
	g15Plus1, _ := ParseGDay("---15+01:00")
	g15Plus2, _ := ParseGDay("---15+02:00")

	tests := []struct {
		name     string
		base     GDay
		other    XSDValue
		expected bool
	}{
		{
			name:     "identical unzoned",
			base:     g15,
			other:    g15Dup,
			expected: true,
		},
		{
			name:     "identical zoned",
			base:     g15Z,
			other:    g15ZDup,
			expected: true,
		},
		{
			name:     "different days unzoned",
			base:     g15,
			other:    g16,
			expected: false,
		},
		{
			name:     "unzoned vs zoned",
			base:     g15,
			other:    g15Z,
			expected: false,
		},
		{
			name:     "zoned vs unzoned",
			base:     g15Z,
			other:    g15,
			expected: false,
		},
		{
			name:     "same day different timezones",
			base:     g15Plus1,
			other:    g15Plus2,
			expected: false,
		},
		{
			name:     "different datatype",
			base:     g15,
			other:    String("---15"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.base.IsIdenticalWith(tt.other)
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

// TestGDayCompare tests the chronological comparison logic between GDay values.
func TestGDayCompare(t *testing.T) {
	g10Z, _ := ParseGDay("---10Z")
	g10ZDup, _ := ParseGDay("---10Z")
	g20Z, _ := ParseGDay("---20Z")

	g01, _ := ParseGDay("---01")
	g20, _ := ParseGDay("---20")

	g15Unzoned, _ := ParseGDay("---15")
	g15Zoned, _ := ParseGDay("---15Z")

	t.Run("equal zoned", func(t *testing.T) {
		assertCompareResult(t, g10Z, g10ZDup, 0)
	})

	t.Run("less than zoned", func(t *testing.T) {
		assertCompareResult(t, g10Z, g20Z, -1)
	})

	t.Run("greater than zoned", func(t *testing.T) {
		assertCompareResult(t, g20Z, g10Z, 1)
	})

	t.Run("determinate both unzoned", func(t *testing.T) {
		assertCompareResult(t, g01, g20, -1)
		assertCompareResult(t, g20, g01, 1)
	})

	t.Run("determinate unzoned vs zoned", func(t *testing.T) {
		assertCompareResult(t, g01, g20Z, 1)
		assertCompareResult(t, g20Z, g01, -1)
	})

	t.Run("indeterminate unzoned vs zoned (overlapping uncertainty)", func(t *testing.T) {
		assertCompareResultError(t, g15Unzoned, g15Zoned, ErrDateTimeOverflow)
	})

	t.Run("incomparable datatype", func(t *testing.T) {
		assertCompareResultError(t, g10Z, String("---10Z"), ErrDateTimeOverflow)
	})
}

// assertCompareResult asserts that comparing two GDay values yields the expected result.
func assertCompareResult(t *testing.T, g1 GDay, other XSDValue, expectedRes int) {
	t.Helper()
	res, err := g1.Compare(other)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != expectedRes {
		t.Errorf("expected %d, got %d", expectedRes, res)
	}
}

// assertCompareResultError asserts that comparing two GDay values returns the expected error.
func assertCompareResultError(t *testing.T, g1 GDay, other XSDValue, expectedErr error) {
	t.Helper()
	_, err := g1.Compare(other)
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

// TestGDayString tests string serialization of GDay values with various timezone formats.
func TestGDayString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unzoned single digit day formatted with leading zero",
			input:    "---05",
			expected: "---05",
		},
		{
			name:     "unzoned double digit day",
			input:    "---25",
			expected: "---25",
		},
		{
			name:     "zoned UTC",
			input:    "---15Z",
			expected: "---15Z",
		},
		{
			name:     "zoned positive offset",
			input:    "---15+05:00",
			expected: "---15+05:00",
		},
		{
			name:     "zoned negative offset",
			input:    "---15-08:00",
			expected: "---15-08:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := ParseGDay(tt.input)
			if err != nil {
				t.Fatalf("unexpected error parsing %q: %v", tt.input, err)
			}
			if got := g.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// containsString reports whether the substring sub is present within s.
func containsString(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (len(s) > 0 && len(sub) > 0 && searchSubstring(s, sub)))
}

// searchSubstring scans through s to determine if sub exists as a contiguous slice.
func searchSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
