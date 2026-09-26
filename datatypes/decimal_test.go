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
	"strings"
	"testing"
)

// mockDecimalProvider implements DecimalProvider for testing.
type mockDecimalProvider struct {
	dec Decimal
}

// ToDecimal returns the underlying decimal.
func (m mockDecimalProvider) ToDecimal() Decimal {
	return m.dec
}

// String returns the string representation of the decimal.
func (m mockDecimalProvider) String() string {
	return m.dec.String()
}

// IsIdenticalWith checks if the mock decimal is identical to another value.
func (m mockDecimalProvider) IsIdenticalWith(other XSDValue) bool {
	return m.dec.IsIdenticalWith(other)
}

// mockOtherXSDValue is a dummy XSDValue implementation for testing incomparable types.
type mockOtherXSDValue struct{}

// String returns a mock string representation.
func (m mockOtherXSDValue) String() string {
	return "mock"
}

// IsIdenticalWith always returns false for the mock value.
func (m mockOtherXSDValue) IsIdenticalWith(_ XSDValue) bool {
	return false
}

// TestNewDecimal tests creating a new decimal with a big int and scale.
func TestNewDecimal(t *testing.T) {
	dec := NewDecimal(big.NewInt(1234), 2)
	if got := dec.String(); got != "12.34" {
		t.Errorf("expected 12.34, got %s", got)
	}

	decZero := NewDecimal(big.NewInt(0), 0)
	if got := decZero.String(); got != "0" {
		t.Errorf("expected 0, got %s", got)
	}
}

// TestNewDecimalFromInt64 tests creating a decimal from an int64 value.
func TestNewDecimalFromInt64(t *testing.T) {
	d1 := NewDecimalFromInt64(42)
	if got := d1.String(); got != "42" {
		t.Errorf("expected 42, got %s", got)
	}

	d2 := NewDecimalFromInt64(-100)
	if got := d2.String(); got != "-100" {
		t.Errorf("expected -100, got %s", got)
	}
}

// TestDefaultDecimalAndGetValue tests default decimal creation and internal value retrieval.
func TestDefaultDecimalAndGetValue(t *testing.T) {
	var uninit Decimal
	if uninit.getValue() == nil {
		t.Error("getValue on uninitialized Decimal should not return nil")
	}
	if uninit.getValue().Cmp(big.NewRat(0, 1)) != 0 {
		t.Error("getValue on uninitialized Decimal should be 0")
	}

	def := DefaultDecimal()
	if def.getValue() == nil {
		t.Error("DefaultDecimal should not have nil value")
	}
	if def.String() != "0" {
		t.Errorf("expected 0, got %s", def.String())
	}
}

// TestAdd tests addition of decimals.
func TestAdd(t *testing.T) {
	d1, _ := ParseDecimal("12.5")
	d2, _ := ParseDecimal("7.5")
	sum := d1.Add(d2)
	if got := sum.String(); got != "20" {
		t.Errorf("expected 20, got %s", got)
	}

	var uninit Decimal
	sum2 := uninit.Add(d1)
	if got := sum2.String(); got != "12.5" {
		t.Errorf("expected 12.5, got %s", got)
	}
}

// TestSub tests subtraction of decimals.
func TestSub(t *testing.T) {
	d1, _ := ParseDecimal("10.5")
	d2, _ := ParseDecimal("3.2")
	diff := d1.Sub(d2)
	if got := diff.String(); got != "7.3" {
		t.Errorf("expected 7.3, got %s", got)
	}

	diffNeg := d2.Sub(d1)
	if got := diffNeg.String(); got != "-7.3" {
		t.Errorf("expected -7.3, got %s", got)
	}
}

// TestMul tests multiplication of decimals.
func TestMul(t *testing.T) {
	d1, _ := ParseDecimal("2.5")
	d2, _ := ParseDecimal("4.0")
	prod := d1.Mul(d2)
	if got := prod.String(); got != "10" {
		t.Errorf("expected 10, got %s", got)
	}

	d3, _ := ParseDecimal("-3.0")
	prodNeg := d1.Mul(d3)
	if got := prodNeg.String(); got != "-7.5" {
		t.Errorf("expected -7.5, got %s", got)
	}
}

// TestCheckedDiv tests division of decimals including division by zero.
func TestCheckedDiv(t *testing.T) {
	d1, _ := ParseDecimal("10")
	d2, _ := ParseDecimal("2")
	quot, err := d1.CheckedDiv(d2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := quot.String(); got != "5" {
		t.Errorf("expected 5, got %s", got)
	}

	zero := DefaultDecimal()
	_, err = d1.CheckedDiv(zero)
	if err == nil || err.Error() != "division by zero" {
		t.Errorf("expected division by zero error, got %v", err)
	}
}

// TestCheckedRemEuclid tests Euclidean remainder operation on decimals.
func TestCheckedRemEuclid(t *testing.T) {
	zero := DefaultDecimal()
	d7, _ := ParseDecimal("7")
	d3, _ := ParseDecimal("3")

	_, err := d7.CheckedRemEuclid(zero)
	if err == nil || err.Error() != "division by zero" {
		t.Errorf("expected division by zero error, got %v", err)
	}

	rem1, err := d7.CheckedRemEuclid(d3)
	if err != nil || rem1.String() != "1" {
		t.Errorf("7 rem 3 expected 1, got %s, err %v", rem1.String(), err)
	}

	dNeg7, _ := ParseDecimal("-7")
	rem2, err := dNeg7.CheckedRemEuclid(d3)
	if err != nil || rem2.String() != "2" {
		t.Errorf("-7 rem 3 expected 2, got %s, err %v", rem2.String(), err)
	}

	dNeg3, _ := ParseDecimal("-3")
	rem3, err := d7.CheckedRemEuclid(dNeg3)
	if err != nil || rem3.String() != "-2" {
		t.Errorf("7 rem -3 expected -2, got %s, err %v", rem3.String(), err)
	}

	rem4, err := dNeg7.CheckedRemEuclid(dNeg3)
	if err != nil || rem4.String() != "-1" {
		t.Errorf("-7 rem -3 expected -1, got %s, err %v", rem4.String(), err)
	}
}

// TestNegAndAbs tests negation and absolute value operations.
func TestNegAndAbs(t *testing.T) {
	d, _ := ParseDecimal("5.5")

	neg := d.Neg()
	if neg.String() != "-5.5" {
		t.Errorf("expected -5.5, got %s", neg.String())
	}

	abs := neg.Abs()
	if abs.String() != "5.5" {
		t.Errorf("expected 5.5, got %s", abs.String())
	}

	absPos := d.Abs()
	if absPos.String() != "5.5" {
		t.Errorf("expected 5.5, got %s", absPos.String())
	}
}

// TestIsNegativeAndIsPositive tests positivity and negativity checks.
func TestIsNegativeAndIsPositive(t *testing.T) {
	pos, _ := ParseDecimal("10.5")
	neg, _ := ParseDecimal("-10.5")
	zero := DefaultDecimal()

	if !pos.IsPositive() || pos.IsNegative() {
		t.Error("pos.IsPositive() should be true, IsNegative() false")
	}

	if !neg.IsNegative() || neg.IsPositive() {
		t.Error("neg.IsNegative() should be true, IsPositive() false")
	}

	if zero.IsPositive() || zero.IsNegative() {
		t.Error("zero should be neither positive nor negative")
	}
}

// TestIsIdenticalWith tests strict identity comparison between values.
func TestIsIdenticalWith(t *testing.T) {
	d1, _ := ParseDecimal("12.34")
	d2, _ := ParseDecimal("12.34")
	d3, _ := ParseDecimal("12.35")

	if !d1.IsIdenticalWith(d2) {
		t.Error("d1 and d2 should be identical")
	}

	if d1.IsIdenticalWith(d3) {
		t.Error("d1 and d3 should not be identical")
	}

	if d1.IsIdenticalWith(mockOtherXSDValue{}) {
		t.Error("d1 should not be identical with non-Decimal XSDValue")
	}
}

// TestCompare tests comparison operations between decimals and other types.
func TestCompare(t *testing.T) {
	d1, _ := ParseDecimal("10")
	d2, _ := ParseDecimal("20")
	d1Equal, _ := ParseDecimal("10")

	cmp, err := d1.Compare(d2)
	if err != nil || cmp != -1 {
		t.Errorf("expected -1, got %d, err %v", cmp, err)
	}

	cmp, err = d2.Compare(d1)
	if err != nil || cmp != 1 {
		t.Errorf("expected 1, got %d, err %v", cmp, err)
	}

	cmp, err = d1.Compare(d1Equal)
	if err != nil || cmp != 0 {
		t.Errorf("expected 0, got %d, err %v", cmp, err)
	}

	dp := mockDecimalProvider{dec: d2}
	cmp, err = d1.Compare(dp)
	if err != nil || cmp != -1 {
		t.Errorf("expected -1 for DecimalProvider comparison, got %d, err %v", cmp, err)
	}

	_, err = d1.Compare(mockOtherXSDValue{})
	if err == nil {
		t.Error("expected error comparing with incomparable type")
	}
}

// TestToDecimal tests conversion to a decimal.
func TestToDecimal(t *testing.T) {
	d, _ := ParseDecimal("123.45")
	if d.ToDecimal().String() != "123.45" {
		t.Errorf("ToDecimal failed, got %s", d.ToDecimal().String())
	}
}

// TestAsI128 tests conversion to an internal 128-bit integer representation.
func TestAsI128(t *testing.T) {
	d1, _ := ParseDecimal("12.34")
	if d1.asI128().Cmp(big.NewInt(12)) != 0 {
		t.Errorf("expected 12, got %s", d1.asI128().String())
	}

	d2, _ := ParseDecimal("-12.34")
	if d2.asI128().Cmp(big.NewInt(-12)) != 0 {
		t.Errorf("expected -12, got %s", d2.asI128().String())
	}

	var uninit Decimal
	if uninit.asI128().Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected 0, got %s", uninit.asI128().String())
	}
}

// TestParseDecimal tests parsing strings into decimals.
func TestParseDecimal(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr error
	}{
		{"", "", ErrParseDecimalEnd},
		{"123", "123", nil},
		{"+123", "123", nil},
		{"-123.45", "-123.45", nil},
		{".5", "0.5", nil},
		{"+.5", "0.5", nil},
		{"-.5", "-0.5", nil},
		{"123.", "123", nil},
		{"12a3", "", ErrParseDecimalBadChar},
		{"123+", "", ErrParseDecimalBadChar},
		{"-12-3", "", ErrParseDecimalBadChar},
		{"1.2.3", "", ErrParseDecimalBadChar},
		{"++12", "", ErrParseDecimalBadChar},
		{"+", "", ErrParseDecimalBadChar},
		{"-", "", ErrParseDecimalBadChar},
	}

	for _, tt := range tests {
		got, err := ParseDecimal(tt.input)
		if tt.wantErr != nil {
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseDecimal(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		} else {
			if err != nil {
				t.Errorf("ParseDecimal(%q) unexpected error: %v", tt.input, err)
			} else if got.String() != tt.want {
				t.Errorf("ParseDecimal(%q) = %q, want %q", tt.input, got.String(), tt.want)
			}
		}
	}
}

// TestDecimalString tests the string formatting of decimals.
func TestDecimalString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0", "0"},
		{"123", "123"},
		{"-123", "-123"},
		{"0.5", "0.5"},
		{"0.2", "0.2"},
		{"0.25", "0.25"},
		{"0.01", "0.01"},
		{"-0.25", "-0.25"},
		{"12.34", "12.34"},
	}

	for _, tt := range tests {
		d, err := ParseDecimal(tt.input)
		if err != nil {
			t.Fatalf("ParseDecimal(%q) failed: %v", tt.input, err)
		}
		if got := d.String(); got != tt.want {
			t.Errorf("ParseDecimal(%q).String() = %q, want %q", tt.input, got, tt.want)
		}
	}

	d1 := NewDecimalFromInt64(1)
	d3 := NewDecimalFromInt64(3)
	third, err := d1.CheckedDiv(d3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	thirdStr := third.String()
	if !strings.HasPrefix(thirdStr, "0.333333") {
		t.Errorf("expected 0.33333..., got %s", thirdStr)
	}
}
