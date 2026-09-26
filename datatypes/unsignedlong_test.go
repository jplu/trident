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
	"testing"
)

// TestParseUnsignedLong verifies the parsing logic for xsd:unsignedLong lexical values.
func TestParseUnsignedLong(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    UnsignedLong
		expectedErr error
	}{
		{
			name:        "valid zero",
			input:       "0",
			expected:    0,
			expectedErr: nil,
		},
		{
			name:        "valid positive with plus sign",
			input:       "+42",
			expected:    42,
			expectedErr: nil,
		},
		{
			name:        "valid maximum value",
			input:       "18446744073709551615",
			expected:    math.MaxUint64,
			expectedErr: nil,
		},
		{
			name:        "valid negative zero single digit",
			input:       "-0",
			expected:    0,
			expectedErr: nil,
		},
		{
			name:        "valid negative zero multiple digits",
			input:       "-000",
			expected:    0,
			expectedErr: nil,
		},
		{
			name:        "overflow upper boundary",
			input:       "18446744073709551616",
			expected:    0,
			expectedErr: ErrOutOfBoundsUnsignedLong,
		},
		{
			name:        "strictly negative value",
			input:       "-1",
			expected:    0,
			expectedErr: ErrOutOfBoundsUnsignedLong,
		},
		{
			name:        "negative value with zero prefix",
			input:       "-01",
			expected:    0,
			expectedErr: ErrOutOfBoundsUnsignedLong,
		},
		{
			name:        "invalid character string",
			input:       "invalid",
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := ParseUnsignedLong(tt.input)
			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, val)
			}
		})
	}
}

// TestUnsignedLongString validates the canonical string representation of UnsignedLong.
func TestUnsignedLongString(t *testing.T) {
	tests := []struct {
		name     string
		val      UnsignedLong
		expected string
	}{
		{
			name:     "zero value",
			val:      0,
			expected: "0",
		},
		{
			name:     "arbitrary positive value",
			val:      123456789,
			expected: "123456789",
		},
		{
			name:     "maximum uint64 value",
			val:      math.MaxUint64,
			expected: "18446744073709551615",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.String(); got != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestUnsignedLongIsIdenticalWith validates the identity comparison for UnsignedLong.
func TestUnsignedLongIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name     string
		val      UnsignedLong
		other    XSDValue
		expected bool
	}{
		{
			name:     "identical values",
			val:      UnsignedLong(100),
			other:    UnsignedLong(100),
			expected: true,
		},
		{
			name:     "different values of same type",
			val:      UnsignedLong(100),
			other:    UnsignedLong(200),
			expected: false,
		},
		{
			name:     "different datatype with same numeric value",
			val:      UnsignedLong(100),
			other:    Long(100),
			expected: false,
		},
		{
			name:     "nil comparison",
			val:      UnsignedLong(100),
			other:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.IsIdenticalWith(tt.other); got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

// TestUnsignedLongCompare validates the order comparison behavior of UnsignedLong.
func TestUnsignedLongCompare(t *testing.T) {
	tests := []struct {
		name        string
		val         UnsignedLong
		other       XSDValue
		expected    int
		expectError bool
	}{
		{
			name:        "same type less than",
			val:         UnsignedLong(10),
			other:       UnsignedLong(20),
			expected:    -1,
			expectError: false,
		},
		{
			name:        "same type equal",
			val:         UnsignedLong(20),
			other:       UnsignedLong(20),
			expected:    0,
			expectError: false,
		},
		{
			name:        "same type greater than",
			val:         UnsignedLong(30),
			other:       UnsignedLong(20),
			expected:    1,
			expectError: false,
		},
		{
			name:        "compare with DecimalProvider less than",
			val:         UnsignedLong(10),
			other:       NewDecimalFromInt64(20),
			expected:    -1,
			expectError: false,
		},
		{
			name:        "compare with DecimalProvider equal",
			val:         UnsignedLong(20),
			other:       NewDecimalFromInt64(20),
			expected:    0,
			expectError: false,
		},
		{
			name:        "compare with DecimalProvider greater than",
			val:         UnsignedLong(30),
			other:       NewDecimalFromInt64(20),
			expected:    1,
			expectError: false,
		},
		{
			name:        "incomparable type",
			val:         UnsignedLong(10),
			other:       String("10"),
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.val.Compare(tt.other)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, res)
			}
		})
	}
}

// TestUnsignedLongToDecimal verifies the conversion of UnsignedLong to an arbitrary-precision Decimal.
func TestUnsignedLongToDecimal(t *testing.T) {
	tests := []struct {
		name     string
		val      UnsignedLong
		expected string
	}{
		{
			name:     "zero to decimal",
			val:      0,
			expected: "0",
		},
		{
			name:     "arbitrary value to decimal",
			val:      42,
			expected: "42",
		},
		{
			name:     "max uint64 to decimal",
			val:      math.MaxUint64,
			expected: "18446744073709551615",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := tt.val.ToDecimal()
			if got := dec.String(); got != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}
