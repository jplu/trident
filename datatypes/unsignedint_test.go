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
	"math/big"
	"testing"
)

// TestParseUnsignedInt tests the lexical parsing and range validation for xsd:unsignedInt values.
func TestParseUnsignedInt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    UnsignedInt
		expectedErr error
	}{
		{
			name:        "valid minimum bound",
			input:       "0",
			expected:    0,
			expectedErr: nil,
		},
		{
			name:        "valid maximum bound",
			input:       "4294967295",
			expected:    4294967295,
			expectedErr: nil,
		},
		{
			name:        "valid positive sign prefix",
			input:       "+12345",
			expected:    12345,
			expectedErr: nil,
		},
		{
			name:        "valid negative zero",
			input:       "-0",
			expected:    0,
			expectedErr: nil,
		},
		{
			name:        "valid multiple negative zeros",
			input:       "-000",
			expected:    0,
			expectedErr: nil,
		},
		{
			name:        "overflow upper bound",
			input:       "4294967296",
			expected:    0,
			expectedErr: ErrOutOfBoundsUnsignedInt,
		},
		{
			name:        "negative value violation",
			input:       "-1",
			expected:    0,
			expectedErr: ErrOutOfBoundsUnsignedInt,
		},
		{
			name:        "negative sign only",
			input:       "-",
			expected:    0,
			expectedErr: ErrOutOfBoundsUnsignedInt,
		},
		{
			name:        "invalid syntax characters",
			input:       "123abc",
			expected:    0,
			expectedErr: ErrParseInteger,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    0,
			expectedErr: ErrParseInteger,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseUnsignedInt(tc.input)
			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Fatalf("expected value %d, got %d", tc.expected, got)
			}
		})
	}
}

// TestUnsignedIntString tests the canonical string representation of UnsignedInt values.
func TestUnsignedIntString(t *testing.T) {
	tests := []struct {
		name     string
		val      UnsignedInt
		expected string
	}{
		{
			name:     "zero",
			val:      0,
			expected: "0",
		},
		{
			name:     "arbitrary value",
			val:      987654,
			expected: "987654",
		},
		{
			name:     "maximum bound",
			val:      4294967295,
			expected: "4294967295",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.val.String(); got != tc.expected {
				t.Fatalf("expected string %q, got %q", tc.expected, got)
			}
		})
	}
}

// TestUnsignedIntIsIdenticalWith tests the identity comparison between UnsignedInt and other XSD values.
func TestUnsignedIntIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name     string
		base     UnsignedInt
		other    XSDValue
		expected bool
	}{
		{
			name:     "identical same value",
			base:     42,
			other:    UnsignedInt(42),
			expected: true,
		},
		{
			name:     "non-identical different value",
			base:     42,
			other:    UnsignedInt(43),
			expected: false,
		},
		{
			name:     "non-identical different type same numeric value",
			base:     42,
			other:    UnsignedLong(42),
			expected: false,
		},
		{
			name:     "non-identical string type",
			base:     42,
			other:    String("42"),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.base.IsIdenticalWith(tc.other); got != tc.expected {
				t.Fatalf("expected identity %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestUnsignedIntCompare tests order evaluation of UnsignedInt against same and cross-datatype values.
func TestUnsignedIntCompare(t *testing.T) {
	tests := []struct {
		name        string
		base        UnsignedInt
		other       XSDValue
		expected    int
		expectError bool
	}{
		{
			name:        "same type less than",
			base:        10,
			other:       UnsignedInt(20),
			expected:    -1,
			expectError: false,
		},
		{
			name:        "same type greater than",
			base:        20,
			other:       UnsignedInt(10),
			expected:    1,
			expectError: false,
		},
		{
			name:        "same type equal",
			base:        15,
			other:       UnsignedInt(15),
			expected:    0,
			expectError: false,
		},
		{
			name:        "cross-type decimal provider less than",
			base:        10,
			other:       NewDecimalFromInt64(20),
			expected:    -1,
			expectError: false,
		},
		{
			name:        "cross-type decimal provider greater than",
			base:        20,
			other:       NewDecimalFromInt64(10),
			expected:    1,
			expectError: false,
		},
		{
			name:        "cross-type decimal provider equal",
			base:        15,
			other:       NewDecimalFromInt64(15),
			expected:    0,
			expectError: false,
		},
		{
			name:        "incomparable type",
			base:        10,
			other:       String("10"),
			expected:    0,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.base.Compare(tc.other)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error for comparison, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected comparison error: %v", err)
			}
			if got != tc.expected {
				t.Fatalf("expected comparison result %d, got %d", tc.expected, got)
			}
		})
	}
}

// TestUnsignedIntToDecimal tests conversion of UnsignedInt to arbitrary-precision Decimal representation.
func TestUnsignedIntToDecimal(t *testing.T) {
	val := UnsignedInt(4294967295)
	dec := val.ToDecimal()

	expectedRat := new(big.Rat).SetInt(new(big.Int).SetUint64(4294967295))
	if dec.getValue().Cmp(expectedRat) != 0 {
		t.Fatalf("expected decimal %v, got %v", expectedRat, dec.getValue())
	}

	zeroVal := UnsignedInt(0)
	zeroDec := zeroVal.ToDecimal()
	if zeroDec.getValue().Sign() != 0 {
		t.Fatalf("expected zero decimal, got %v", zeroDec.getValue())
	}
}
