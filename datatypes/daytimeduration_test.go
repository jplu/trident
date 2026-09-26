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
	"time"
)

// TestNewDayTimeDuration verifies the creation of a DayTimeDuration from a decimal.
func TestNewDayTimeDuration(t *testing.T) {
	dec := NewDecimalFromInt64(3600)
	dtd := NewDayTimeDuration(dec)
	if !dtd.AsSeconds().IsIdenticalWith(dec) {
		t.Errorf("NewDayTimeDuration() = %v, want %v", dtd.AsSeconds(), dec)
	}
}

// TestDefaultDayTimeDuration verifies the default values of a DayTimeDuration.
func TestDefaultDayTimeDuration(t *testing.T) {
	dtd := DefaultDayTimeDuration()
	zeroDec := DefaultDecimal()
	if !dtd.AsSeconds().IsIdenticalWith(zeroDec) {
		t.Errorf("DefaultDayTimeDuration() = %v, want %v", dtd.AsSeconds(), zeroDec)
	}
	if dtd.String() != "PT0S" {
		t.Errorf("DefaultDayTimeDuration().String() = %q, want %q", dtd.String(), "PT0S")
	}
}

// TestNewDayTimeDurationFromStd verifies creating a DayTimeDuration from a standard time.Duration.
func TestNewDayTimeDurationFromStd(t *testing.T) {
	tests := []struct {
		name    string
		stdDur  time.Duration
		wantSec string
	}{
		{
			name:    "positive duration",
			stdDur:  2*time.Hour + 30*time.Minute + 15*time.Second + 500*time.Millisecond,
			wantSec: "9015.5",
		},
		{
			name:    "zero duration",
			stdDur:  0,
			wantSec: "0",
		},
		{
			name:    "negative duration",
			stdDur:  -10 * time.Second,
			wantSec: "-10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dtd := NewDayTimeDurationFromStd(tt.stdDur)
			gotSec := dtd.AsSeconds().String()
			if gotSec != tt.wantSec {
				t.Errorf("NewDayTimeDurationFromStd(%v) seconds = %q, want %q", tt.stdDur, gotSec, tt.wantSec)
			}
		})
	}
}

// TestDayTimeDurationComponents verifies extracting component values from a DayTimeDuration.
func TestDayTimeDurationComponents(t *testing.T) {
	t.Run("positive components", func(t *testing.T) {
		dtd, err := ParseDayTimeDuration("P2DT3H4M5.5S")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		if got := dtd.Days(); got != 2 {
			t.Errorf("Days() = %d, want 2", got)
		}
		if got := dtd.Hours(); got != 3 {
			t.Errorf("Hours() = %d, want 3", got)
		}
		if got := dtd.Minutes(); got != 4 {
			t.Errorf("Minutes() = %d, want 4", got)
		}
		wantSec, _ := ParseDecimal("5.5")
		if got := dtd.Seconds(); !got.IsIdenticalWith(wantSec) {
			t.Errorf("Seconds() = %v, want %v", got, wantSec)
		}
	})

	t.Run("negative components", func(t *testing.T) {
		dtd, err := ParseDayTimeDuration("-P2DT3H4M5S")
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}

		if got := dtd.Days(); got != -2 {
			t.Errorf("Days() = %d, want -2", got)
		}
		if got := dtd.Hours(); got != -3 {
			t.Errorf("Hours() = %d, want -3", got)
		}
		if got := dtd.Minutes(); got != -4 {
			t.Errorf("Minutes() = %d, want -4", got)
		}
	})
}

// TestDayTimeDurationArithmetic verifies addition, subtraction, and negation of DayTimeDurations.
func TestDayTimeDurationArithmetic(t *testing.T) {
	d1, _ := ParseDayTimeDuration("PT1H")
	d2, _ := ParseDayTimeDuration("PT30M")

	t.Run("Add", func(t *testing.T) {
		sum := d1.Add(d2)
		want, _ := ParseDayTimeDuration("PT1H30M")
		if !sum.IsIdenticalWith(want) {
			t.Errorf("Add() = %v, want %v", sum, want)
		}
	})

	t.Run("Sub", func(t *testing.T) {
		diff := d1.Sub(d2)
		want, _ := ParseDayTimeDuration("PT30M")
		if !diff.IsIdenticalWith(want) {
			t.Errorf("Sub() = %v, want %v", diff, want)
		}
	})

	t.Run("Neg", func(t *testing.T) {
		neg := d1.Neg()
		want, _ := ParseDayTimeDuration("-PT1H")
		if !neg.IsIdenticalWith(want) {
			t.Errorf("Neg() = %v, want %v", neg, want)
		}
	})
}

// TestDayTimeDuration_IsIdenticalWith verifies the identity check between DayTimeDurations.
func TestDayTimeDurationIsIdenticalWith(t *testing.T) {
	d1, _ := ParseDayTimeDuration("PT1H")
	d2, _ := ParseDayTimeDuration("PT3600S")
	d3, _ := ParseDayTimeDuration("PT2H")

	if !d1.IsIdenticalWith(d2) {
		t.Errorf("IsIdenticalWith() expected true for PT1H and PT3600S")
	}

	if d1.IsIdenticalWith(d3) {
		t.Errorf("IsIdenticalWith() expected false for PT1H and PT2H")
	}

	if d1.IsIdenticalWith(String("PT1H")) {
		t.Errorf("IsIdenticalWith() expected false for different XSDValue type")
	}
}

// TestDayTimeDurationCompare verifies comparison methods for DayTimeDurations.
func TestDayTimeDurationCompare(t *testing.T) {
	d1, _ := ParseDayTimeDuration("PT1H")
	d2, _ := ParseDayTimeDuration("PT2H")
	d3, _ := ParseDayTimeDuration("PT3600S")

	t.Run("less than", func(t *testing.T) {
		cmp, err := d1.Compare(d2)
		if err != nil || cmp != -1 {
			t.Errorf("Compare() = (%d, %v), want (-1, nil)", cmp, err)
		}
	})

	t.Run("greater than", func(t *testing.T) {
		cmp, err := d2.Compare(d1)
		if err != nil || cmp != 1 {
			t.Errorf("Compare() = (%d, %v), want (1, nil)", cmp, err)
		}
	})

	t.Run("equal", func(t *testing.T) {
		cmp, err := d1.Compare(d3)
		if err != nil || cmp != 0 {
			t.Errorf("Compare() = (%d, %v), want (0, nil)", cmp, err)
		}
	})

	t.Run("incomparable type", func(t *testing.T) {
		_, err := d1.Compare(String("PT1H"))
		if !errors.Is(err, ErrDurationOverflow) {
			t.Errorf("Compare() error = %v, want %v", err, ErrDurationOverflow)
		}
	})
}

// TestDayTimeDurationString verifies string formatting for DayTimeDurations.
func TestDayTimeDurationString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"PT0S", "PT0S"},
		{"P1D", "P1D"},
		{"PT1H", "PT1H"},
		{"PT1M", "PT1M"},
		{"PT1S", "PT1S"},
		{"P1DT2H3M4S", "P1DT2H3M4S"},
		{"-P1DT2H3M4S", "-P1DT2H3M4S"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			dtd, err := ParseDayTimeDuration(tt.input)
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}
			if got := dtd.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestParseDayTimeDuration verifies parsing valid and invalid DayTimeDuration strings.
func TestParseDayTimeDuration(t *testing.T) {
	t.Run("valid inputs", func(t *testing.T) {
		validCases := []string{
			"P1D",
			"PT1H",
			"PT1M",
			"PT1S",
			"PT1.5S",
			"P1DT2H3M4.5S",
			"-P1DT2H3M4.5S",
			"PT0S",
		}
		for _, s := range validCases {
			if _, err := ParseDayTimeDuration(s); err != nil {
				t.Errorf("ParseDayTimeDuration(%q) unexpected error: %v", s, err)
			}
		}
	})

	t.Run("invalid inputs", func(t *testing.T) {
		invalidCases := []struct {
			input string
			msg   string
		}{
			{"INVALID", "durations must start with 'P'"},
			{"P1Y", "there must not be any year or month component in a dayTimeDuration"},
			{"P1M", "there must not be any year or month component in a dayTimeDuration"},
			{"P1Y1D", "there must not be any year or month component in a dayTimeDuration"},
			{"P", "no day or time values found"},
		}

		for _, tc := range invalidCases {
			_, err := ParseDayTimeDuration(tc.input)
			if err == nil {
				t.Errorf("ParseDayTimeDuration(%q) expected error containing %q, got nil", tc.input, tc.msg)
			}
		}
	})
}
