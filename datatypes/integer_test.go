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

// TestNewInteger tests constructing an Integer from an existing *big.Int.
func TestNewInteger(t *testing.T) {
	val := big.NewInt(42)
	i := NewInteger(val)
	if i.getValue().Cmp(val) != 0 {
		t.Fatalf("expected %v, got %v", val, i.getValue())
	}

	val.SetInt64(99)
	if i.getValue().Int64() != 42 {
		t.Fatalf("expected independent copy with value 42, got %v", i.getValue())
	}
}

// TestNewIntegerFromInt64 tests constructing an Integer from an int64 value.
func TestNewIntegerFromInt64(t *testing.T) {
	i := NewIntegerFromInt64(-12345)
	if i.getValue().Int64() != -12345 {
		t.Fatalf("expected -12345, got %v", i.getValue())
	}
}

// TestDefaultInteger tests the default zero-valued Integer initialization.
func TestDefaultInteger(t *testing.T) {
	i := DefaultInteger()
	if i.getValue().Sign() != 0 {
		t.Fatalf("expected 0, got %v", i.getValue())
	}
}

// TestParseInteger tests parsing valid and invalid integer string representations.
func TestParseInteger(t *testing.T) {
	validTests := []struct {
		input    string
		expected int64
	}{
		{"0", 0},
		{"123456789", 123456789},
		{"-987654321", -987654321},
		{"+42", 42},
	}

	for _, tt := range validTests {
		res, err := ParseInteger(tt.input)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tt.input, err)
		}
		if res.getValue().Int64() != tt.expected {
			t.Fatalf("for %q expected %d, got %v", tt.input, tt.expected, res.getValue())
		}
	}

	invalidTests := []string{
		"",
		"abc",
		"12.34",
		"123a",
		"--10",
		"++10",
	}

	for _, input := range invalidTests {
		_, err := ParseInteger(input)
		if !errors.Is(err, ErrParseInteger) {
			t.Fatalf("expected ErrParseInteger for %q, got %v", input, err)
		}
	}
}

// TestIntegerGetValueNil tests calling getValue on an uninitialized Integer struct.
func TestIntegerGetValueNil(t *testing.T) {
	var i Integer
	val := i.getValue()
	if val == nil || val.Sign() != 0 {
		t.Fatalf("expected 0 from nil Integer, got %v", val)
	}
}

// TestIntegerArithmeticOperations tests addition, subtraction, and multiplication.
func TestIntegerArithmeticOperations(t *testing.T) {
	a := NewIntegerFromInt64(15)
	b := NewIntegerFromInt64(5)

	sum := a.Add(b)
	if sum.getValue().Int64() != 20 {
		t.Fatalf("addition failed: expected 20, got %v", sum)
	}

	diff := a.Sub(b)
	if diff.getValue().Int64() != 10 {
		t.Fatalf("subtraction failed: expected 10, got %v", diff)
	}

	prod := a.Mul(b)
	if prod.getValue().Int64() != 75 {
		t.Fatalf("multiplication failed: expected 75, got %v", prod)
	}
}

// TestIntegerCheckedDiv tests checked division including division by zero.
func TestIntegerCheckedDiv(t *testing.T) {
	a := NewIntegerFromInt64(20)
	b := NewIntegerFromInt64(4)
	zero := DefaultInteger()

	quo, err := a.CheckedDiv(b)
	if err != nil || quo.getValue().Int64() != 5 {
		t.Fatalf("division failed: expected 5, got %v (err: %v)", quo, err)
	}

	_, err = a.CheckedDiv(zero)
	if err == nil {
		t.Fatalf("expected error on division by zero")
	}
}

// TestIntegerCheckedRem tests checked remainder calculations including division by zero.
func TestIntegerCheckedRem(t *testing.T) {
	a := NewIntegerFromInt64(23)
	b := NewIntegerFromInt64(5)
	zero := DefaultInteger()

	rem, err := a.CheckedRem(b)
	if err != nil || rem.getValue().Int64() != 3 {
		t.Fatalf("remainder failed: expected 3, got %v (err: %v)", rem, err)
	}

	_, err = a.CheckedRem(zero)
	if err == nil {
		t.Fatalf("expected error on remainder by zero")
	}
}

// TestIntegerCheckedRemEuclid tests Euclidean remainder under positive, negative, and zero divisor cases.
func TestIntegerCheckedRemEuclid(t *testing.T) {
	zero := DefaultInteger()
	val := NewIntegerFromInt64(10)

	_, err := val.CheckedRemEuclid(zero)
	if err == nil {
		t.Fatalf("expected error on Euclidean remainder by zero")
	}

	posA := NewIntegerFromInt64(7)
	posB := NewIntegerFromInt64(3)
	rem1, err := posA.CheckedRemEuclid(posB)
	if err != nil || rem1.getValue().Int64() != 1 {
		t.Fatalf("expected 1, got %v (err: %v)", rem1, err)
	}

	negA := NewIntegerFromInt64(-7)
	rem2, err := negA.CheckedRemEuclid(posB)
	if err != nil || rem2.getValue().Int64() != 2 {
		t.Fatalf("expected 2 for -7 mod 3, got %v (err: %v)", rem2, err)
	}

	negB := NewIntegerFromInt64(-3)
	rem3, err := negA.CheckedRemEuclid(negB)
	if err != nil || rem3.getValue().Int64() != 2 {
		t.Fatalf("expected 2 for -7 mod -3, got %v (err: %v)", rem3, err)
	}
}

// TestIntegerCheckedNegAndAbs tests negation and absolute value computations.
func TestIntegerCheckedNegAndAbs(t *testing.T) {
	a := NewIntegerFromInt64(42)
	negA := a.Neg()
	if negA.getValue().Int64() != -42 {
		t.Fatalf("expected -42, got %v", negA)
	}

	posAgain := negA.Neg()
	if posAgain.getValue().Int64() != 42 {
		t.Fatalf("expected 42, got %v", posAgain)
	}

	absA := negA.Abs()
	if absA.getValue().Int64() != 42 {
		t.Fatalf("expected 42, got %v", absA)
	}

	absPos := a.Abs()
	if absPos.getValue().Int64() != 42 {
		t.Fatalf("expected 42, got %v", absPos)
	}
}

// TestIntegerSignChecks tests IsNegative and IsPositive across negative, zero, and positive values.
func TestIntegerSignChecks(t *testing.T) {
	neg := NewIntegerFromInt64(-1)
	zero := DefaultInteger()
	pos := NewIntegerFromInt64(1)

	if !neg.IsNegative() || neg.IsPositive() {
		t.Fatalf("expected neg to be negative and not positive")
	}

	if zero.IsNegative() || zero.IsPositive() {
		t.Fatalf("expected zero to be neither negative nor positive")
	}

	if pos.IsNegative() || !pos.IsPositive() {
		t.Fatalf("expected pos to be positive and not negative")
	}
}

// TestIntegerIsIdenticalWith tests identity comparison with matching, non-matching, and different types.
func TestIntegerIsIdenticalWith(t *testing.T) {
	a := NewIntegerFromInt64(100)
	b := NewIntegerFromInt64(100)
	c := NewIntegerFromInt64(101)
	otherType := String("100")

	if !a.IsIdenticalWith(b) {
		t.Fatalf("expected identical integers to match")
	}

	if a.IsIdenticalWith(c) {
		t.Fatalf("expected different integers to not match")
	}

	if a.IsIdenticalWith(otherType) {
		t.Fatalf("expected non-Integer type to return false")
	}
}

// TestIntegerCompare tests order comparisons against Integer, DecimalProvider, and incompatible types.
func TestIntegerCompare(t *testing.T) {
	val := NewIntegerFromInt64(10)
	less := NewIntegerFromInt64(5)
	equal := NewIntegerFromInt64(10)
	greater := NewIntegerFromInt64(15)

	cmp, err := val.Compare(less)
	if err != nil || cmp != 1 {
		t.Fatalf("expected 1, got %d (err: %v)", cmp, err)
	}

	cmp, err = val.Compare(equal)
	if err != nil || cmp != 0 {
		t.Fatalf("expected 0, got %d (err: %v)", cmp, err)
	}

	cmp, err = val.Compare(greater)
	if err != nil || cmp != -1 {
		t.Fatalf("expected -1, got %d (err: %v)", cmp, err)
	}

	decLess := NewDecimalFromInt64(5)
	decEqual := NewDecimalFromInt64(10)
	decGreater := NewDecimalFromInt64(15)

	cmp, err = val.Compare(decLess)
	if err != nil || cmp != 1 {
		t.Fatalf("expected 1 against DecimalProvider, got %d (err: %v)", cmp, err)
	}

	cmp, err = val.Compare(decEqual)
	if err != nil || cmp != 0 {
		t.Fatalf("expected 0 against DecimalProvider, got %d (err: %v)", cmp, err)
	}

	cmp, err = val.Compare(decGreater)
	if err != nil || cmp != -1 {
		t.Fatalf("expected -1 against DecimalProvider, got %d (err: %v)", cmp, err)
	}

	_, err = val.Compare(String("invalid"))
	if err == nil {
		t.Fatalf("expected error when comparing with incompatible type")
	}
}

// TestIntegerToDecimal tests converting an Integer to an arbitrary-precision Decimal representation.
func TestIntegerToDecimal(t *testing.T) {
	i := NewIntegerFromInt64(54321)
	d := i.ToDecimal()

	expectedRat := new(big.Rat).SetInt64(54321)
	if d.getValue().Cmp(expectedRat) != 0 {
		t.Fatalf("expected decimal %v, got %v", expectedRat, d.getValue())
	}
}

// TestIntegerString tests the canonical string representation of Integer values.
func TestIntegerString(t *testing.T) {
	tests := []struct {
		val      Integer
		expected string
	}{
		{NewIntegerFromInt64(0), "0"},
		{NewIntegerFromInt64(12345), "12345"},
		{NewIntegerFromInt64(-9876), "-9876"},
		{Integer{}, "0"},
	}

	for _, tt := range tests {
		res := tt.val.String()
		if res != tt.expected {
			t.Fatalf("expected %q, got %q", tt.expected, res)
		}
	}
}
