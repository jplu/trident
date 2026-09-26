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

// TestParseUnsignedShort validates parsing of valid and invalid lexical representations for xsd:unsignedShort.
func TestParseUnsignedShort(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    UnsignedShort
		expectedErr error
	}{
		{
			name:        "valid zero",
			input:       "0",
			expected:    0,
			expectedErr: nil,
		},
		{
			name:        "valid max unsigned short",
			input:       "65535",
			expected:    65535,
			expectedErr: nil,
		},
		{
			name:        "valid positive signed number",
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
			name:        "valid negative zero multiple digits",
			input:       "-000",
			expected:    0,
			expectedErr: nil,
		},
		{
			name:        "negative non-zero error",
			input:       "-1",
			expected:    0,
			expectedErr: ErrOutOfBoundsUnsignedShort,
		},
		{
			name:        "negative non-zero with leading zeros error",
			input:       "-001",
			expected:    0,
			expectedErr: ErrOutOfBoundsUnsignedShort,
		},
		{
			name:        "negative single dash error",
			input:       "-",
			expected:    0,
			expectedErr: ErrOutOfBoundsUnsignedShort,
		},
		{
			name:        "out of bounds upper limit error",
			input:       "65536",
			expected:    0,
			expectedErr: ErrOutOfBoundsUnsignedShort,
		},
		{
			name:        "invalid characters error",
			input:       "123abc",
			expected:    0,
			expectedErr: ErrParseInteger,
		},
		{
			name:        "empty string error",
			input:       "",
			expected:    0,
			expectedErr: ErrParseInteger,
		},
		{
			name:        "decimal format error",
			input:       "123.45",
			expected:    0,
			expectedErr: ErrParseInteger,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := ParseUnsignedShort(tt.input)
			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("ParseUnsignedShort(%q) expected error %v, got %v", tt.input, tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseUnsignedShort(%q) unexpected error: %v", tt.input, err)
			}

			if val != tt.expected {
				t.Fatalf("ParseUnsignedShort(%q) = %d, expected %d", tt.input, val, tt.expected)
			}
		})
	}
}

// TestUnsignedShortString validates the canonical string representation of an UnsignedShort value.
func TestUnsignedShortString(t *testing.T) {
	tests := []struct {
		name     string
		val      UnsignedShort
		expected string
	}{
		{
			name:     "zero value",
			val:      0,
			expected: "0",
		},
		{
			name:     "mid-range value",
			val:      1234,
			expected: "1234",
		},
		{
			name:     "maximum bound value",
			val:      65535,
			expected: "65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.String(); got != tt.expected {
				t.Fatalf("UnsignedShort(%d).String() = %q, expected %q", tt.val, got, tt.expected)
			}
		})
	}
}

// TestUnsignedShortIsIdenticalWith tests value-space identity checks for UnsignedShort.
func TestUnsignedShortIsIdenticalWith(t *testing.T) {
	u := UnsignedShort(100)

	if !u.IsIdenticalWith(UnsignedShort(100)) {
		t.Fatal("expected UnsignedShort(100) to be identical with UnsignedShort(100)")
	}

	if u.IsIdenticalWith(UnsignedShort(101)) {
		t.Fatal("expected UnsignedShort(100) not to be identical with UnsignedShort(101)")
	}

	if u.IsIdenticalWith(String("100")) {
		t.Fatal("expected UnsignedShort(100) not to be identical with String(\"100\")")
	}
}

// TestUnsignedShortCompare validates ordering comparisons of UnsignedShort against other XSD values.
func TestUnsignedShortCompare(t *testing.T) {
	u := UnsignedShort(100)

	res, err := u.Compare(UnsignedShort(200))
	if err != nil || res != -1 {
		t.Fatalf("expected 100 < 200, got res=%d, err=%v", res, err)
	}

	res, err = u.Compare(UnsignedShort(50))
	if err != nil || res != 1 {
		t.Fatalf("expected 100 > 50, got res=%d, err=%v", res, err)
	}

	res, err = u.Compare(UnsignedShort(100))
	if err != nil || res != 0 {
		t.Fatalf("expected 100 == 100, got res=%d, err=%v", res, err)
	}

	res, err = u.Compare(NewDecimalFromInt64(200))
	if err != nil || res != -1 {
		t.Fatalf("expected 100 < Decimal(200), got res=%d, err=%v", res, err)
	}

	res, err = u.Compare(NewDecimalFromInt64(50))
	if err != nil || res != 1 {
		t.Fatalf("expected 100 > Decimal(50), got res=%d, err=%v", res, err)
	}

	res, err = u.Compare(NewDecimalFromInt64(100))
	if err != nil || res != 0 {
		t.Fatalf("expected 100 == Decimal(100), got res=%d, err=%v", res, err)
	}

	_, err = u.Compare(String("100"))
	if err == nil {
		t.Fatal("expected error comparing UnsignedShort with incomparable type String")
	}
}

// TestUnsignedShortToDecimal validates conversion from UnsignedShort to Decimal.
func TestUnsignedShortToDecimal(t *testing.T) {
	tests := []struct {
		name     string
		val      UnsignedShort
		expected *big.Rat
	}{
		{
			name:     "zero",
			val:      0,
			expected: big.NewRat(0, 1),
		},
		{
			name:     "positive number",
			val:      42,
			expected: big.NewRat(42, 1),
		},
		{
			name:     "maximum bound",
			val:      65535,
			expected: big.NewRat(65535, 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := tt.val.ToDecimal()
			if dec.getValue().Cmp(tt.expected) != 0 {
				t.Fatalf("UnsignedShort(%d).ToDecimal() = %v, expected %v", tt.val, dec.getValue(), tt.expected)
			}
		})
	}
}
