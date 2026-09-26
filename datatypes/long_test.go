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
	"strconv"
	"testing"
)

// TestParseLong verifies the parsing of lexical string representations into Long values.
func TestParseLong(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Long
		expectedErr error
	}{
		{
			name:        "valid zero",
			input:       "0",
			expected:    0,
			expectedErr: nil,
		},
		{
			name:        "valid positive",
			input:       "123456789",
			expected:    123456789,
			expectedErr: nil,
		},
		{
			name:        "valid positive with plus sign",
			input:       "+987654321",
			expected:    987654321,
			expectedErr: nil,
		},
		{
			name:        "valid negative",
			input:       "-987654321",
			expected:    -987654321,
			expectedErr: nil,
		},
		{
			name:        "valid max int64",
			input:       strconv.FormatInt(math.MaxInt64, 10),
			expected:    Long(math.MaxInt64),
			expectedErr: nil,
		},
		{
			name:        "valid min int64",
			input:       strconv.FormatInt(math.MinInt64, 10),
			expected:    Long(math.MinInt64),
			expectedErr: nil,
		},
		{
			name:        "overflow upper bound",
			input:       "9223372036854775808",
			expected:    0,
			expectedErr: ErrOutOfBoundsLong,
		},
		{
			name:        "overflow lower bound",
			input:       "-9223372036854775809",
			expected:    0,
			expectedErr: ErrOutOfBoundsLong,
		},
		{
			name:        "invalid syntax letters",
			input:       "abc",
			expected:    0,
			expectedErr: ErrParseInteger,
		},
		{
			name:        "invalid syntax decimal point",
			input:       "123.45",
			expected:    0,
			expectedErr: ErrParseInteger,
		},
		{
			name:        "invalid syntax empty string",
			input:       "",
			expected:    0,
			expectedErr: ErrParseInteger,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := ParseLong(tt.input)
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

// TestLongString tests the canonical lexical serialization of Long values.
func TestLongString(t *testing.T) {
	tests := []struct {
		name     string
		val      Long
		expected string
	}{
		{
			name:     "zero",
			val:      0,
			expected: "0",
		},
		{
			name:     "positive value",
			val:      42,
			expected: "42",
		},
		{
			name:     "negative value",
			val:      -42,
			expected: "-42",
		},
		{
			name:     "max int64",
			val:      Long(math.MaxInt64),
			expected: "9223372036854775807",
		},
		{
			name:     "min int64",
			val:      Long(math.MinInt64),
			expected: "-9223372036854775808",
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

// TestLongIsIdenticalWith tests the identity check between a Long and other XSD values.
func TestLongIsIdenticalWith(t *testing.T) {
	val := Long(100)

	if !val.IsIdenticalWith(Long(100)) {
		t.Fatalf("expected identical values to return true")
	}

	if val.IsIdenticalWith(Long(200)) {
		t.Fatalf("expected different Long values to return false")
	}

	if val.IsIdenticalWith(Int(100)) {
		t.Fatalf("expected different datatype to return false")
	}

	if val.IsIdenticalWith(String("100")) {
		t.Fatalf("expected string datatype to return false")
	}
}

// TestLongCompare tests the order relation between Long and other comparable XSD values.
func TestLongCompare(t *testing.T) {
	val := Long(50)

	cmp, err := val.Compare(Long(100))
	if err != nil || cmp != -1 {
		t.Fatalf("expected -1 and nil error, got %d, %v", cmp, err)
	}

	cmp, err = val.Compare(Long(50))
	if err != nil || cmp != 0 {
		t.Fatalf("expected 0 and nil error, got %d, %v", cmp, err)
	}

	cmp, err = val.Compare(Long(20))
	if err != nil || cmp != 1 {
		t.Fatalf("expected 1 and nil error, got %d, %v", cmp, err)
	}

	decValLess := NewDecimalFromInt64(100)
	cmp, err = val.Compare(decValLess)
	if err != nil || cmp != -1 {
		t.Fatalf("expected -1 when compared to larger Decimal, got %d, %v", cmp, err)
	}

	decValEqual := NewDecimalFromInt64(50)
	cmp, err = val.Compare(decValEqual)
	if err != nil || cmp != 0 {
		t.Fatalf("expected 0 when compared to equal Decimal, got %d, %v", cmp, err)
	}

	decValGreater := NewDecimalFromInt64(20)
	cmp, err = val.Compare(decValGreater)
	if err != nil || cmp != 1 {
		t.Fatalf("expected 1 when compared to smaller Decimal, got %d, %v", cmp, err)
	}

	_, err = val.Compare(String("50"))
	if err == nil {
		t.Fatalf("expected error comparing Long with String, got nil")
	}
}

// TestLongToDecimal tests the conversion of a Long to its arbitrary-precision Decimal representation.
func TestLongToDecimal(t *testing.T) {
	tests := []struct {
		name     string
		val      Long
		expected string
	}{
		{
			name:     "zero",
			val:      0,
			expected: "0",
		},
		{
			name:     "positive",
			val:      123456789,
			expected: "123456789",
		},
		{
			name:     "negative",
			val:      -987654321,
			expected: "-987654321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := tt.val.ToDecimal()
			if dec.String() != tt.expected {
				t.Fatalf("expected decimal string %s, got %s", tt.expected, dec.String())
			}
		})
	}
}
