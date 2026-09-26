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
	"math"
	"math/big"
	"testing"
)

// TestGetUTC verifies that GetUTC returns a non-nil TimezoneOffset representing UTC with zero offset.
func TestGetUTC(t *testing.T) {
	utc := GetUTC()
	if utc == nil {
		t.Fatal("expected non-nil TimezoneOffset")
	}
	if utc.offset != 0 {
		t.Fatalf("expected offset 0, got %d", utc.offset)
	}
	if got := utc.String(); got != "Z" {
		t.Fatalf("expected 'Z', got %s", got)
	}
}

// TestNewTimezoneOffsetValid verifies creation of valid TimezoneOffset values across boundaries and nominal ranges.
func TestNewTimezoneOffsetValid(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int16
		expected string
	}{
		{name: "Min Bound (-14:00)", minutes: minTimezoneOffsetMinutes, expected: "-14:00"},
		{name: "Negative Offset (-05:00)", minutes: -300, expected: "-05:00"},
		{name: "Negative Fractional Offset (-03:30)", minutes: -210, expected: "-03:30"},
		{name: "Zero Offset (UTC)", minutes: 0, expected: "Z"},
		{name: "Positive Offset (+05:30)", minutes: 330, expected: "+05:30"},
		{name: "Max Bound (+14:00)", minutes: maxTimezoneOffsetMinutes, expected: "+14:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tz, err := NewTimezoneOffset(tt.minutes)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tz.offset != tt.minutes {
				t.Fatalf("expected offset %d, got %d", tt.minutes, tz.offset)
			}
			if got := tz.String(); got != tt.expected {
				t.Fatalf("expected string %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestNewTimezoneOffsetInvalid verifies that values outside [-840, 840] fail with an InvalidTimezoneError.
func TestNewTimezoneOffsetInvalid(t *testing.T) {
	tests := []struct {
		name    string
		minutes int16
	}{
		{name: "Below Min Bound", minutes: minTimezoneOffsetMinutes - 1},
		{name: "Above Max Bound", minutes: maxTimezoneOffsetMinutes + 1},
		{name: "Extreme Negative", minutes: math.MinInt16},
		{name: "Extreme Positive", minutes: math.MaxInt16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTimezoneOffset(tt.minutes)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var tzErr InvalidTimezoneError
			if !errors.As(err, &tzErr) {
				t.Fatalf("expected InvalidTimezoneError, got %T", err)
			}
			if tzErr.OffsetInMinutes != int64(tt.minutes) {
				t.Fatalf("expected offset in error %d, got %d", tt.minutes, tzErr.OffsetInMinutes)
			}
		})
	}
}

// TestNewTimezoneOffsetFromDayTimeDurationValid tests converting valid DayTimeDuration instances into TimezoneOffset.
func TestNewTimezoneOffsetFromDayTimeDurationValid(t *testing.T) {
	tests := []struct {
		name            string
		seconds         int64
		expectedMinutes int16
	}{
		{name: "Zero duration", seconds: 0, expectedMinutes: 0},
		{name: "Positive 2 hours", seconds: 7200, expectedMinutes: 120},
		{name: "Negative 5 hours", seconds: -18000, expectedMinutes: -300},
		{
			name:            "Max bound (+14h)",
			seconds:         int64(maxTimezoneOffsetMinutes) * secondsPerMinute,
			expectedMinutes: maxTimezoneOffsetMinutes,
		},
		{
			name:            "Min bound (-14h)",
			seconds:         int64(minTimezoneOffsetMinutes) * secondsPerMinute,
			expectedMinutes: minTimezoneOffsetMinutes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur := NewDayTimeDuration(NewDecimalFromInt64(tt.seconds))
			tz, err := NewTimezoneOffsetFromDayTimeDuration(dur)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tz.offset != tt.expectedMinutes {
				t.Fatalf("expected offset %d, got %d", tt.expectedMinutes, tz.offset)
			}
		})
	}
}

// TestNewTimezoneOffsetFromDayTimeDurationNonZeroSeconds verifies that non-zero remainder seconds trigger an error.
func TestNewTimezoneOffsetFromDayTimeDurationNonZeroSeconds(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
	}{
		{name: "1 second remainder", seconds: 61},
		{name: "Negative remainder", seconds: -59},
		{name: "Odd seconds", seconds: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur := NewDayTimeDuration(NewDecimalFromInt64(tt.seconds))
			_, err := NewTimezoneOffsetFromDayTimeDuration(dur)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var tzErr InvalidTimezoneError
			if !errors.As(err, &tzErr) {
				t.Fatalf("expected InvalidTimezoneError, got %T", err)
			}
		})
	}
}

// TestNewTimezoneOffsetFromDayTimeDurationOverflow tests integer capacity boundary failures during conversion.
func TestNewTimezoneOffsetFromDayTimeDurationOverflow(t *testing.T) {
	t.Run("Exceeds MaxInt16", func(t *testing.T) {
		sec := (int64(math.MaxInt16) + 1) * secondsPerMinute
		dur := NewDayTimeDuration(NewDecimalFromInt64(sec))
		_, err := NewTimezoneOffsetFromDayTimeDuration(dur)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var tzErr InvalidTimezoneError
		if !errors.As(err, &tzErr) {
			t.Fatalf("expected InvalidTimezoneError, got %T", err)
		}
		if tzErr.OffsetInMinutes != int64(math.MaxInt16)+1 {
			t.Fatalf("expected offset %d, got %d", int64(math.MaxInt16)+1, tzErr.OffsetInMinutes)
		}
	})

	t.Run("Below MinInt16", func(t *testing.T) {
		sec := (int64(math.MinInt16) - 1) * secondsPerMinute
		dur := NewDayTimeDuration(NewDecimalFromInt64(sec))
		_, err := NewTimezoneOffsetFromDayTimeDuration(dur)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var tzErr InvalidTimezoneError
		if !errors.As(err, &tzErr) {
			t.Fatalf("expected InvalidTimezoneError, got %T", err)
		}
		if tzErr.OffsetInMinutes != int64(math.MinInt16)-1 {
			t.Fatalf("expected offset %d, got %d", int64(math.MinInt16)-1, tzErr.OffsetInMinutes)
		}
	})

	t.Run("Exceeds Int64 Representation", func(t *testing.T) {
		bigSec := new(big.Int).Lsh(big.NewInt(1), 70)
		bigSec.Mul(bigSec, big.NewInt(secondsPerMinute))
		dec := NewDecimal(bigSec, 0)
		dur := NewDayTimeDuration(dec)
		_, err := NewTimezoneOffsetFromDayTimeDuration(dur)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var tzErr InvalidTimezoneError
		if !errors.As(err, &tzErr) {
			t.Fatalf("expected InvalidTimezoneError, got %T", err)
		}
	})
}

// TestNewTimezoneOffsetFromDayTimeDurationRangeValidation tests duration values that fit into int16 but exceed the
// timezone range.
func TestNewTimezoneOffsetFromDayTimeDurationRangeValidation(t *testing.T) {
	tests := []struct {
		name    string
		minutes int64
	}{
		{name: "Positive 15 hours", minutes: 900},
		{name: "Negative 15 hours", minutes: -900},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur := NewDayTimeDuration(NewDecimalFromInt64(tt.minutes * secondsPerMinute))
			_, err := NewTimezoneOffsetFromDayTimeDuration(dur)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var tzErr InvalidTimezoneError
			if !errors.As(err, &tzErr) {
				t.Fatalf("expected InvalidTimezoneError, got %T", err)
			}
			if tzErr.OffsetInMinutes != tt.minutes {
				t.Fatalf("expected offset %d, got %d", tt.minutes, tzErr.OffsetInMinutes)
			}
		})
	}
}

// TestTimezoneOffsetToDayTimeDuration verifies conversion from TimezoneOffset to DayTimeDuration.
func TestTimezoneOffsetToDayTimeDuration(t *testing.T) {
	tests := []struct {
		name            string
		minutes         int16
		expectedSeconds int64
	}{
		{name: "Zero", minutes: 0, expectedSeconds: 0},
		{name: "Positive Offset", minutes: 120, expectedSeconds: 7200},
		{name: "Negative Offset", minutes: -300, expectedSeconds: -18000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tz, err := NewTimezoneOffset(tt.minutes)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dur := tz.ToDayTimeDuration()
			if dur.AsSeconds().asI128().Int64() != tt.expectedSeconds {
				t.Fatalf("expected %d seconds, got %d", tt.expectedSeconds, dur.AsSeconds().asI128().Int64())
			}
		})
	}
}

// TestTimezoneOffsetBEBytes verifies big-endian byte serialization and deserialization.
func TestTimezoneOffsetBEBytes(t *testing.T) {
	tests := []struct {
		name    string
		offset  int16
		beBytes [2]byte
	}{
		{name: "Zero", offset: 0, beBytes: [2]byte{0x00, 0x00}},
		{name: "Positive offset (+04:18)", offset: 258, beBytes: [2]byte{0x01, 0x02}},
		{name: "Max bound (+14:00)", offset: 840, beBytes: [2]byte{0x03, 0x48}},
		{name: "Negative offset (-05:00)", offset: -300, beBytes: [2]byte{0xFE, 0xD4}},
		{name: "Min bound (-14:00)", offset: -840, beBytes: [2]byte{0xFC, 0xB8}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tz, err := NewTimezoneOffset(tt.offset)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			b := tz.ToBEBytes()
			if b != tt.beBytes {
				t.Fatalf("expected bytes %v, got %v", tt.beBytes, b)
			}
			deserialized := TimezoneOffsetFromBEBytes(b)
			if deserialized.offset != tt.offset {
				t.Fatalf("expected offset %d, got %d", tt.offset, deserialized.offset)
			}
		})
	}
}
