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

// TestParseNonPositiveInteger validates parsing of non-positive integer literals across valid, invalid, and positive
// inputs.
func TestParseNonPositiveInteger(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedStr string
		expectedErr error
	}{
		{
			name:        "valid zero",
			input:       "0",
			expectedStr: "0",
			expectedErr: nil,
		},
		{
			name:        "valid negative zero",
			input:       "-0",
			expectedStr: "0",
			expectedErr: nil,
		},
		{
			name:        "valid negative integer",
			input:       "-1",
			expectedStr: "-1",
			expectedErr: nil,
		},
		{
			name:        "valid large negative integer",
			input:       "-9223372036854775808999999",
			expectedStr: "-9223372036854775808999999",
			expectedErr: nil,
		},
		{
			name:        "invalid syntax non-numeric",
			input:       "invalid",
			expectedStr: "",
			expectedErr: ErrParseInteger,
		},
		{
			name:        "invalid positive integer 1",
			input:       "1",
			expectedStr: "",
			expectedErr: ErrNotNonPositiveInteger,
		},
		{
			name:        "invalid positive integer with plus sign",
			input:       "+42",
			expectedStr: "",
			expectedErr: ErrNotNonPositiveInteger,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ParseNonPositiveInteger(tt.input)
			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.String() != tt.expectedStr {
				t.Fatalf("expected string %q, got %q", tt.expectedStr, res.String())
			}
		})
	}
}

// TestNonPositiveIntegerString verifies the canonical string representation of NonPositiveInteger values.
func TestNonPositiveIntegerString(t *testing.T) {
	val, err := ParseNonPositiveInteger("-100")
	if err != nil {
		t.Fatalf("unexpected error parsing non-positive integer: %v", err)
	}

	if val.String() != "-100" {
		t.Fatalf("expected -100, got %s", val.String())
	}
}

// TestNonPositiveIntegerIsIdenticalWith tests identity comparison for NonPositiveInteger values.
func TestNonPositiveIntegerIsIdenticalWith(t *testing.T) {
	val1, err := ParseNonPositiveInteger("-50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val2, err := ParseNonPositiveInteger("-50")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val3, err := ParseNonPositiveInteger("-10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !val1.IsIdenticalWith(val2) {
		t.Fatalf("expected %v to be identical with %v", val1, val2)
	}
	if val1.IsIdenticalWith(val3) {
		t.Fatalf("expected %v to not be identical with %v", val1, val3)
	}
	if val1.IsIdenticalWith(String("-50")) {
		t.Fatalf("expected %v to not be identical with different type", val1)
	}
}
