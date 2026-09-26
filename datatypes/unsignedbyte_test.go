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
	"math/big"
	"testing"
)

// TestParseUnsignedByteValid tests valid lexical representations for xsd:unsignedByte.
func TestParseUnsignedByteValid(t *testing.T) {
	tests := []struct {
		input    string
		expected UnsignedByte
	}{
		{"0", UnsignedByte(0)},
		{"255", UnsignedByte(255)},
		{"+128", UnsignedByte(128)},
		{"-0", UnsignedByte(0)},
		{"-000", UnsignedByte(0)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, err := ParseUnsignedByte(tt.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if val != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, val)
			}
		})
	}
}

// TestParseUnsignedByteInvalid tests invalid lexical representations and out-of-bounds inputs for xsd:unsignedByte.
func TestParseUnsignedByteInvalid(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{"overflow", "256", ErrOutOfBoundsUnsignedByte},
		{"negative number", "-1", ErrOutOfBoundsUnsignedByte},
		{"negative with non-zero suffix", "-01", ErrOutOfBoundsUnsignedByte},
		{"invalid characters", "abc", ErrParseInteger},
		{"empty string", "", ErrParseInteger},
		{"single minus", "-", ErrOutOfBoundsUnsignedByte},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseUnsignedByte(tt.input)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tt.input)
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

// TestUnsignedByteString tests the canonical lexical serialization of UnsignedByte.
func TestUnsignedByteString(t *testing.T) {
	tests := []struct {
		val      UnsignedByte
		expected string
	}{
		{UnsignedByte(0), "0"},
		{UnsignedByte(1), "1"},
		{UnsignedByte(255), "255"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.val.String(); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

// TestUnsignedByteIsIdenticalWith tests identity comparison for UnsignedByte.
func TestUnsignedByteIsIdenticalWith(t *testing.T) {
	val := UnsignedByte(42)

	if !val.IsIdenticalWith(UnsignedByte(42)) {
		t.Fatal("expected identical values to return true")
	}

	if val.IsIdenticalWith(UnsignedByte(43)) {
		t.Fatal("expected different values to return false")
	}

	if val.IsIdenticalWith(String("42")) {
		t.Fatal("expected different types to return false")
	}
}

// TestUnsignedByteCompare tests order comparison with other values and types.
func TestUnsignedByteCompare(t *testing.T) {
	val := UnsignedByte(100)

	cmp, err := val.Compare(UnsignedByte(50))
	if err != nil || cmp != 1 {
		t.Fatalf("expected 1, got %d (err: %v)", cmp, err)
	}

	cmp, err = val.Compare(UnsignedByte(150))
	if err != nil || cmp != -1 {
		t.Fatalf("expected -1, got %d (err: %v)", cmp, err)
	}

	cmp, err = val.Compare(UnsignedByte(100))
	if err != nil || cmp != 0 {
		t.Fatalf("expected 0, got %d (err: %v)", cmp, err)
	}

	decLess := NewDecimalFromInt64(50)
	cmp, err = val.Compare(decLess)
	if err != nil || cmp != 1 {
		t.Fatalf("expected 1 against decimal, got %d (err: %v)", cmp, err)
	}

	decGreater := NewDecimalFromInt64(150)
	cmp, err = val.Compare(decGreater)
	if err != nil || cmp != -1 {
		t.Fatalf("expected -1 against decimal, got %d (err: %v)", cmp, err)
	}

	decEqual := NewDecimalFromInt64(100)
	cmp, err = val.Compare(decEqual)
	if err != nil || cmp != 0 {
		t.Fatalf("expected 0 against decimal, got %d (err: %v)", cmp, err)
	}

	_, err = val.Compare(String("100"))
	if err == nil {
		t.Fatal("expected error comparing with incomparable type, got nil")
	}
}

// TestUnsignedByteToDecimal tests projection of UnsignedByte into arbitrary-precision Decimal.
func TestUnsignedByteToDecimal(t *testing.T) {
	val := UnsignedByte(200)
	dec := val.ToDecimal()

	expectedRat := new(big.Rat).SetInt(new(big.Int).SetUint64(200))
	if dec.getValue().Cmp(expectedRat) != 0 {
		t.Fatalf("expected decimal %v, got %v", expectedRat, dec.getValue())
	}
}
