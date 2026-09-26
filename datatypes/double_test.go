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
	"math"
	"testing"
)

// TestGetDoubleSpecialValues verifies that special double values are correctly created.
func TestGetDoubleSpecialValues(t *testing.T) {
	inf := GetDoubleInfinity()
	if !math.IsInf(float64(inf), 1) {
		t.Errorf("GetDoubleInfinity() = %v, expected +INF", inf)
	}

	negInf := GetDoubleNegInfinity()
	if !math.IsInf(float64(negInf), -1) {
		t.Errorf("GetDoubleNegInfinity() = %v, expected -INF", negInf)
	}

	nan := GetDoubleNaN()
	if !nan.IsNaN() {
		t.Errorf("GetDoubleNaN() = %v, expected NaN", nan)
	}
}

// TestDoubleIsNaN tests the IsNaN method on various Double values.
func TestDoubleIsNaN(t *testing.T) {
	tests := []struct {
		name     string
		d        Double
		expected bool
	}{
		{"NaN", GetDoubleNaN(), true},
		{"Zero", Double(0.0), false},
		{"Positive Number", Double(123.456), false},
		{"Infinity", GetDoubleInfinity(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.IsNaN(); got != tt.expected {
				t.Errorf("Double(%v).IsNaN() = %v, expected %v", tt.d, got, tt.expected)
			}
		})
	}
}

// TestDoubleIsIdenticalWith tests the strict identity comparison of Double values.
func TestDoubleIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name     string
		d        Double
		other    XSDValue
		expected bool
	}{
		{"Equal values", Double(1.5), Double(1.5), true},
		{"Unequal values", Double(1.5), Double(2.5), false},
		{"Both NaN", GetDoubleNaN(), GetDoubleNaN(), true},
		{"d is NaN, other is normal", GetDoubleNaN(), Double(1.0), false},
		{"d is normal, other is NaN", Double(1.0), GetDoubleNaN(), false},
		{"Positive zero and negative zero", Double(0.0), Double(math.Copysign(0.0, -1)), false},
		{"Positive zero and positive zero", Double(0.0), Double(0.0), true},
		{"Incompatible type", Double(1.5), String("1.5"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.IsIdenticalWith(tt.other); got != tt.expected {
				t.Errorf("Double(%v).IsIdenticalWith(%v) = %v, expected %v", tt.d, tt.other, got, tt.expected)
			}
		})
	}
}

// TestDoubleCompare tests the relational comparison between Double and other XSD values.
func TestDoubleCompare(t *testing.T) {
	tests := []struct {
		name        string
		d           Double
		other       XSDValue
		expectedCmp int
		expectError bool
	}{
		{"Less than", Double(1.0), Double(2.0), -1, false},
		{"Greater than", Double(2.0), Double(1.0), 1, false},
		{"Equal values", Double(1.5), Double(1.5), 0, false},
		{"Positive and negative zero equal", Double(0.0), Double(math.Copysign(0.0, -1)), 0, false},
		{"d is NaN", GetDoubleNaN(), Double(1.0), 0, true},
		{"other is NaN", Double(1.0), GetDoubleNaN(), 0, true},
		{"Both NaN", GetDoubleNaN(), GetDoubleNaN(), 0, true},
		{"Incompatible type", Double(1.0), String("1.0"), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.Compare(tt.other)
			if (err != nil) != tt.expectError {
				t.Errorf("Double(%v).Compare(%v) error = %v, expectError %v", tt.d, tt.other, err, tt.expectError)
				return
			}
			if !tt.expectError && got != tt.expectedCmp {
				t.Errorf("Double(%v).Compare(%v) = %v, expected %v", tt.d, tt.other, got, tt.expectedCmp)
			}
		})
	}
}

// TestParseDouble tests parsing various string inputs into Double values.
func TestParseDouble(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Double
		expectError bool
	}{
		{"INF", "INF", GetDoubleInfinity(), false},
		{"+INF", "+INF", GetDoubleInfinity(), false},
		{"-INF", "-INF", GetDoubleNegInfinity(), false},
		{"NaN", "NaN", GetDoubleNaN(), false},
		{"Valid positive float", "123.45", Double(123.45), false},
		{"Valid negative float", "-123.45", Double(-123.45), false},
		{"Valid zero", "0", Double(0.0), false},
		{"Valid negative zero", "-0.0", Double(math.Copysign(0.0, -1)), false},
		{"Valid scientific notation", "1.23e4", Double(12300.0), false},
		{"Contains lowercase x", "0x12", 0, true},
		{"Contains uppercase X", "0X12", 0, true},
		{"Contains lowercase p", "0x1p2", 0, true},
		{"Contains uppercase P", "0x1P2", 0, true},
		{"Contains underscore", "1_000", 0, true},
		{"Invalid format", "abc", 0, true},
		{"Go valid but XSD invalid Infinity", "Infinity", 0, true},
		{"Go valid but XSD invalid nan", "nan", 0, true},
		{"Range overflow positive", "1e999999999999", GetDoubleInfinity(), false},
		{"Range overflow negative", "-1e999999999999", GetDoubleNegInfinity(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDouble(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("ParseDouble(%q) error = %v, expectError %v", tt.input, err, tt.expectError)
				return
			}
			if !tt.expectError {
				if tt.expected.IsNaN() {
					if !got.IsNaN() {
						t.Errorf("ParseDouble(%q) = %v, expected NaN", tt.input, got)
					}
				} else if got != tt.expected && math.Float64bits(float64(got)) != math.Float64bits(float64(tt.expected)) {
					t.Errorf("ParseDouble(%q) = %v, expected %v", tt.input, got, tt.expected)
				}
			}
		})
	}
}

// TestDoubleString tests the string formatting of Double values.
func TestDoubleString(t *testing.T) {
	tests := []struct {
		name     string
		d        Double
		expected string
	}{
		{"INF", GetDoubleInfinity(), "INF"},
		{"-INF", GetDoubleNegInfinity(), "-INF"},
		{"NaN", GetDoubleNaN(), "NaN"},
		{"Zero", Double(0.0), "0.0E0"},
		{"Negative Zero", Double(math.Copysign(0.0, -1)), "-0.0E0"},
		{"Positive float", Double(123.45), "1.2345E2"},
		{"Negative float", Double(-0.00123), "-1.23E-3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.String(); got != tt.expected {
				t.Errorf("Double(%v).String() = %q, expected %q", tt.d, got, tt.expected)
			}
		})
	}
}
