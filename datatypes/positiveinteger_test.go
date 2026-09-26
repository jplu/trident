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
)

// TestParsePositiveInteger tests parsing valid and invalid strings into PositiveInteger.
func TestParsePositiveInteger(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:        "valid positive integer",
			input:       "1",
			expected:    "1",
			expectedErr: nil,
		},
		{
			name:        "valid positive integer with leading plus",
			input:       "+42",
			expected:    "42",
			expectedErr: nil,
		},
		{
			name:        "valid large positive integer",
			input:       "100000000000000000000",
			expected:    "100000000000000000000",
			expectedErr: nil,
		},
		{
			name:        "invalid zero value",
			input:       "0",
			expectedErr: ErrNotPositiveInteger,
		},
		{
			name:        "invalid zero value with leading plus",
			input:       "+0",
			expectedErr: ErrNotPositiveInteger,
		},
		{
			name:        "invalid negative integer",
			input:       "-1",
			expectedErr: ErrNotNonNegativeInteger,
		},
		{
			name:        "invalid non-numeric literal",
			input:       "invalid",
			expectedErr: ErrParseInteger,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pi, err := ParsePositiveInteger(tc.input)
			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.expectedErr)
				}
				if !errors.Is(err, tc.expectedErr) && err.Error() != tc.expectedErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pi.String() != tc.expected {
				t.Fatalf("expected string %q, got %q", tc.expected, pi.String())
			}
		})
	}
}

// TestPositiveIntegerString tests the canonical string representation of PositiveInteger.
func TestPositiveIntegerString(t *testing.T) {
	pi, err := ParsePositiveInteger("12345")
	if err != nil {
		t.Fatalf("unexpected error parsing positive integer: %v", err)
	}
	expected := "12345"
	if actual := pi.String(); actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

// TestPositiveIntegerIsIdenticalWith tests the identity comparison method of PositiveInteger.
func TestPositiveIntegerIsIdenticalWith(t *testing.T) {
	pi1, err := ParsePositiveInteger("100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pi2, err := ParsePositiveInteger("100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pi3, err := ParsePositiveInteger("200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !pi1.IsIdenticalWith(pi2) {
		t.Fatal("expected identical values to return true")
	}
	if pi1.IsIdenticalWith(pi3) {
		t.Fatal("expected non-identical values to return false")
	}
	if pi1.IsIdenticalWith(String("100")) {
		t.Fatal("expected comparison with non-PositiveInteger to return false")
	}
}
