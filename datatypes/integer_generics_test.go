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

// errCustomRange is a test error used to verify boundary overflow propagation.
var errCustomRange = errors.New("custom range error")

// TestParseSigned tests the parsing of signed integers across valid, out-of-range, and malformed inputs.
func TestParseSigned(t *testing.T) {
	val, err := parseSigned[int32]("42", 32, errCustomRange)
	if err != nil || val != 42 {
		t.Fatalf("expected 42 and nil error, got %d and %v", val, err)
	}

	valNeg, err := parseSigned[int32]("-42", 32, errCustomRange)
	if err != nil || valNeg != -42 {
		t.Fatalf("expected -42 and nil error, got %d and %v", valNeg, err)
	}

	_, errRange := parseSigned[int8]("128", 8, errCustomRange)
	if !errors.Is(errRange, errCustomRange) {
		t.Fatalf("expected errCustomRange, got %v", errRange)
	}

	_, errSyntax := parseSigned[int32]("invalid", 32, errCustomRange)
	if !errors.Is(errSyntax, ErrParseInteger) {
		t.Fatalf("expected ErrParseInteger, got %v", errSyntax)
	}
}

// TestParseUnsigned tests the parsing of unsigned integers including prefixes, zero representations, overflows, and
// syntax errors.
func TestParseUnsigned(t *testing.T) {
	valPos, err := parseUnsigned[uint32]("+123", 32, errCustomRange)
	if err != nil || valPos != 123 {
		t.Fatalf("expected 123 and nil error, got %d and %v", valPos, err)
	}

	valPlain, err := parseUnsigned[uint32]("456", 32, errCustomRange)
	if err != nil || valPlain != 456 {
		t.Fatalf("expected 456 and nil error, got %d and %v", valPlain, err)
	}

	valZero, err := parseUnsigned[uint32]("-0", 32, errCustomRange)
	if err != nil || valZero != 0 {
		t.Fatalf("expected 0 and nil error for -0, got %d and %v", valZero, err)
	}

	valMultiZero, err := parseUnsigned[uint32]("-000", 32, errCustomRange)
	if err != nil || valMultiZero != 0 {
		t.Fatalf("expected 0 and nil error for -000, got %d and %v", valMultiZero, err)
	}

	_, errDashOnly := parseUnsigned[uint32]("-", 32, errCustomRange)
	if !errors.Is(errDashOnly, errCustomRange) {
		t.Fatalf("expected errCustomRange for solitary dash, got %v", errDashOnly)
	}

	_, errNegNonZero := parseUnsigned[uint32]("-1", 32, errCustomRange)
	if !errors.Is(errNegNonZero, errCustomRange) {
		t.Fatalf("expected errCustomRange for negative number, got %v", errNegNonZero)
	}

	_, errNegMixedZero := parseUnsigned[uint32]("-05", 32, errCustomRange)
	if !errors.Is(errNegMixedZero, errCustomRange) {
		t.Fatalf("expected errCustomRange for -05, got %v", errNegMixedZero)
	}

	_, errRange := parseUnsigned[uint8]("256", 8, errCustomRange)
	if !errors.Is(errRange, errCustomRange) {
		t.Fatalf("expected errCustomRange for overflow, got %v", errRange)
	}

	_, errSyntax := parseUnsigned[uint32]("notanumber", 32, errCustomRange)
	if !errors.Is(errSyntax, ErrParseInteger) {
		t.Fatalf("expected ErrParseInteger for invalid syntax, got %v", errSyntax)
	}
}

// TestIsIdentical tests value space identity evaluation between comparable types and arbitrary XSD values.
func TestIsIdentical(t *testing.T) {
	if !isIdentical(Int(10), Int(10)) {
		t.Fatal("expected identical values of the same type to return true")
	}

	if isIdentical(Int(10), Int(20)) {
		t.Fatal("expected distinct values of the same type to return false")
	}

	if isIdentical(Int(10), String("10")) {
		t.Fatal("expected different types to return false")
	}
}

// TestCompareSigned tests the comparison logic for signed integers against same-type values, decimal providers, and
// incomparable types.
func TestCompareSigned(t *testing.T) {
	cmpLess, err := compareSigned(Int(10), Int(20))
	if err != nil || cmpLess != -1 {
		t.Fatalf("expected -1 and nil error, got %d and %v", cmpLess, err)
	}

	cmpGreater, err := compareSigned(Int(20), Int(10))
	if err != nil || cmpGreater != 1 {
		t.Fatalf("expected 1 and nil error, got %d and %v", cmpGreater, err)
	}

	cmpEqual, err := compareSigned(Int(15), Int(15))
	if err != nil || cmpEqual != 0 {
		t.Fatalf("expected 0 and nil error, got %d and %v", cmpEqual, err)
	}

	cmpDecLess, err := compareSigned(Int(10), NewDecimalFromInt64(20))
	if err != nil || cmpDecLess != -1 {
		t.Fatalf("expected -1 against decimal 20, got %d and %v", cmpDecLess, err)
	}

	cmpDecGreater, err := compareSigned(Int(20), NewDecimalFromInt64(10))
	if err != nil || cmpDecGreater != 1 {
		t.Fatalf("expected 1 against decimal 10, got %d and %v", cmpDecGreater, err)
	}

	cmpDecEqual, err := compareSigned(Int(15), NewDecimalFromInt64(15))
	if err != nil || cmpDecEqual != 0 {
		t.Fatalf("expected 0 against decimal 15, got %d and %v", cmpDecEqual, err)
	}

	_, errIncomp := compareSigned(Int(10), String("10"))
	if errIncomp == nil || errIncomp.Error() != "incomparable types" {
		t.Fatalf("expected 'incomparable types' error, got %v", errIncomp)
	}
}

// TestCompareUnsigned tests the comparison logic for unsigned integers against same-type values, decimal providers, and
// incomparable types.
func TestCompareUnsigned(t *testing.T) {
	cmpLess, err := compareUnsigned(UnsignedInt(10), UnsignedInt(20))
	if err != nil || cmpLess != -1 {
		t.Fatalf("expected -1 and nil error, got %d and %v", cmpLess, err)
	}

	cmpGreater, err := compareUnsigned(UnsignedInt(20), UnsignedInt(10))
	if err != nil || cmpGreater != 1 {
		t.Fatalf("expected 1 and nil error, got %d and %v", cmpGreater, err)
	}

	cmpEqual, err := compareUnsigned(UnsignedInt(15), UnsignedInt(15))
	if err != nil || cmpEqual != 0 {
		t.Fatalf("expected 0 and nil error, got %d and %v", cmpEqual, err)
	}

	cmpDecLess, err := compareUnsigned(UnsignedInt(10), NewDecimalFromInt64(20))
	if err != nil || cmpDecLess != -1 {
		t.Fatalf("expected -1 against decimal 20, got %d and %v", cmpDecLess, err)
	}

	cmpDecGreater, err := compareUnsigned(UnsignedInt(20), NewDecimalFromInt64(10))
	if err != nil || cmpDecGreater != 1 {
		t.Fatalf("expected 1 against decimal 10, got %d and %v", cmpDecGreater, err)
	}

	cmpDecEqual, err := compareUnsigned(UnsignedInt(15), NewDecimalFromInt64(15))
	if err != nil || cmpDecEqual != 0 {
		t.Fatalf("expected 0 against decimal 15, got %d and %v", cmpDecEqual, err)
	}

	_, errIncomp := compareUnsigned(UnsignedInt(10), String("10"))
	if errIncomp == nil || errIncomp.Error() != "incomparable types" {
		t.Fatalf("expected 'incomparable types' error, got %v", errIncomp)
	}
}
