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

// TestErrDateTimeOverflow tests the ErrDateTimeOverflow error.
func TestErrDateTimeOverflow(t *testing.T) {
	expected := "overflow during xsd:dateTime computation"
	if got := ErrDateTimeOverflow.Error(); got != expected {
		t.Errorf("ErrDateTimeOverflow.Error() = %q, want %q", got, expected)
	}
}

// TestInvalidTimezoneError tests the InvalidTimezoneError error structure and its methods.
func TestInvalidTimezoneError(t *testing.T) {
	tests := []struct {
		name     string
		offset   int64
		expected string
	}{
		{
			name:     "positive offset",
			offset:   300,
			expected: "invalid timezone offset 05:00",
		},
		{
			name:     "negative offset",
			offset:   -300,
			expected: "invalid timezone offset -5:00",
		},
		{
			name:     "zero offset",
			offset:   0,
			expected: "invalid timezone offset 00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InvalidTimezoneError{OffsetInMinutes: tt.offset}
			if got := err.Error(); got != tt.expected {
				t.Errorf("InvalidTimezoneError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestParseDateTimeError tests parsing datetime errors with and without wrapped causes.
func TestParseDateTimeError(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		err := newParseDateTimeError("invalid format")

		expectedMsg := "parsing date/time: invalid format"
		if got := err.Error(); got != expectedMsg {
			t.Errorf("ParseDateTimeError.Error() = %q, want %q", got, expectedMsg)
		}

		if got := err.Unwrap(); got != nil {
			t.Errorf("ParseDateTimeError.Unwrap() = %v, want nil", got)
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		cause := errors.New("underlying cause")
		err := newParseDateTimeErrorWrap("invalid year", cause)

		expectedMsg := "parsing date/time: invalid year: underlying cause"
		if got := err.Error(); got != expectedMsg {
			t.Errorf("ParseDateTimeError.Error() = %q, want %q", got, expectedMsg)
		}

		if got := err.Unwrap(); !errors.Is(got, cause) {
			t.Errorf("ParseDateTimeError.Unwrap() = %v, want %v", got, cause)
		}
	})
}

// TestAbsInt64 tests the absInt64 helper function.
func TestAbsInt64(t *testing.T) {
	tests := []struct {
		input    int64
		expected int64
	}{
		{input: 10, expected: 10},
		{input: -10, expected: 10},
		{input: 0, expected: 0},
	}

	for _, tt := range tests {
		if got := absInt64(tt.input); got != tt.expected {
			t.Errorf("absInt64(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}
