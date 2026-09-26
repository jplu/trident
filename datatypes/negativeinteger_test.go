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

// TestParseNegativeInteger tests parsing valid and invalid lexical values for NegativeInteger.
func TestParseNegativeInteger(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectErr   bool
		expectedErr error
	}{
		{
			name:      "valid negative integer",
			input:     "-1",
			expected:  "-1",
			expectErr: false,
		},
		{
			name:      "valid large negative integer",
			input:     "-9223372036854775809",
			expected:  "-9223372036854775809",
			expectErr: false,
		},
		{
			name:        "invalid zero",
			input:       "0",
			expectErr:   true,
			expectedErr: ErrNotNegativeInteger,
		},
		{
			name:        "invalid negative zero",
			input:       "-0",
			expectErr:   true,
			expectedErr: ErrNotNegativeInteger,
		},
		{
			name:        "invalid positive integer",
			input:       "1",
			expectErr:   true,
			expectedErr: ErrNotNonPositiveInteger,
		},
		{
			name:        "invalid positive with sign",
			input:       "+5",
			expectErr:   true,
			expectedErr: ErrNotNonPositiveInteger,
		},
		{
			name:        "invalid non-numeric literal",
			input:       "invalid",
			expectErr:   true,
			expectedErr: ErrParseInteger,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, err := ParseNegativeInteger(tc.input)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tc.input)
				}
				if tc.expectedErr != nil && !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if val.String() != tc.expected {
				t.Fatalf("expected string %q, got %q", tc.expected, val.String())
			}
		})
	}
}

// TestNegativeIntegerString tests canonical string formatting of NegativeInteger values.
func TestNegativeIntegerString(t *testing.T) {
	val, err := ParseNegativeInteger("-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val.String() != "-42" {
		t.Fatalf("expected -42, got %s", val.String())
	}
}

// TestNegativeIntegerIsIdenticalWith tests identity comparison for NegativeInteger values.
func TestNegativeIntegerIsIdenticalWith(t *testing.T) {
	val1, err := ParseNegativeInteger("-10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val2, err := ParseNegativeInteger("-10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val3, err := ParseNegativeInteger("-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !val1.IsIdenticalWith(val2) {
		t.Fatalf("expected %v to be identical with %v", val1, val2)
	}

	if val1.IsIdenticalWith(val3) {
		t.Fatalf("expected %v not to be identical with %v", val1, val3)
	}

	otherType := String("-10")
	if val1.IsIdenticalWith(otherType) {
		t.Fatalf("expected %v not to be identical with different type %v", val1, otherType)
	}
}
