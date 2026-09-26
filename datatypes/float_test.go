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
	"math"
	"testing"
)

// TestFloatSpecialValues tests the creation and detection of special float values like Infinity, -Infinity, and NaN.
func TestFloatSpecialValues(t *testing.T) {
	inf := GetFloatInfinity()
	if !math.IsInf(float64(inf), 1) {
		t.Errorf("GetFloatInfinity() = %v, expected +Inf", inf)
	}

	negInf := GetFloatNegInfinity()
	if !math.IsInf(float64(negInf), -1) {
		t.Errorf("GetFloatNegInfinity() = %v, expected -Inf", negInf)
	}

	nan := GetFloatNaN()
	if !nan.IsNaN() {
		t.Errorf("GetFloatNaN().IsNaN() = false, expected true")
	}

	val := Float(3.14)
	if val.IsNaN() {
		t.Errorf("Float(3.14).IsNaN() = true, expected false")
	}
}

// TestFloatIsIdenticalWith tests strict value identity checks, properly distinguishing between signed zeros and
// matching identical NaNs.
func TestFloatIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name     string
		f        Float
		other    XSDValue
		expected bool
	}{
		{
			name:     "identical positive floats",
			f:        Float(1.5),
			other:    Float(1.5),
			expected: true,
		},
		{
			name:     "different positive floats",
			f:        Float(1.5),
			other:    Float(2.5),
			expected: false,
		},
		{
			name:     "positive zero and positive zero",
			f:        Float(0.0),
			other:    Float(0.0),
			expected: true,
		},
		{
			name:     "negative zero and negative zero",
			f:        Float(float32(math.Copysign(0, -1))),
			other:    Float(float32(math.Copysign(0, -1))),
			expected: true,
		},
		{
			name:     "positive zero and negative zero are not identical",
			f:        Float(0.0),
			other:    Float(float32(math.Copysign(0, -1))),
			expected: false,
		},
		{
			name:     "negative zero and positive zero are not identical",
			f:        Float(float32(math.Copysign(0, -1))),
			other:    Float(0.0),
			expected: false,
		},
		{
			name:     "NaN and NaN identical bitwise",
			f:        GetFloatNaN(),
			other:    GetFloatNaN(),
			expected: true,
		},
		{
			name:     "positive infinity and positive infinity",
			f:        GetFloatInfinity(),
			other:    GetFloatInfinity(),
			expected: true,
		},
		{
			name:     "positive infinity and negative infinity",
			f:        GetFloatInfinity(),
			other:    GetFloatNegInfinity(),
			expected: false,
		},
		{
			name:     "incompatible type",
			f:        Float(1.0),
			other:    String("1.0"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.f.IsIdenticalWith(tt.other)
			if result != tt.expected {
				t.Errorf("Float(%v).IsIdenticalWith(%v) = %v, expected %v", tt.f, tt.other, result, tt.expected)
			}
		})
	}
}

// TestFloatCompare tests the relational comparison of Float values, including correct error reporting for NaNs and type
// mismatches.
func TestFloatCompare(t *testing.T) {
	tests := []struct {
		name        string
		f           Float
		other       XSDValue
		expectedCmp int
		expectError bool
	}{
		{
			name:        "less than",
			f:           Float(1.0),
			other:       Float(2.0),
			expectedCmp: -1,
			expectError: false,
		},
		{
			name:        "greater than",
			f:           Float(2.0),
			other:       Float(1.0),
			expectedCmp: 1,
			expectError: false,
		},
		{
			name:        "equal values",
			f:           Float(1.5),
			other:       Float(1.5),
			expectedCmp: 0,
			expectError: false,
		},
		{
			name:        "positive zero and negative zero are equal in order relation",
			f:           Float(0.0),
			other:       Float(float32(math.Copysign(0, -1))),
			expectedCmp: 0,
			expectError: false,
		},
		{
			name:        "receiver is NaN",
			f:           GetFloatNaN(),
			other:       Float(1.0),
			expectedCmp: 0,
			expectError: true,
		},
		{
			name:        "other is NaN",
			f:           Float(1.0),
			other:       GetFloatNaN(),
			expectedCmp: 0,
			expectError: true,
		},
		{
			name:        "both are NaN",
			f:           GetFloatNaN(),
			other:       GetFloatNaN(),
			expectedCmp: 0,
			expectError: true,
		},
		{
			name:        "incomparable type",
			f:           Float(1.0),
			other:       String("1.0"),
			expectedCmp: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmp, err := tt.f.Compare(tt.other)
			if tt.expectError {
				if err == nil {
					t.Errorf("Float(%v).Compare(%v) expected error, got nil", tt.f, tt.other)
				}
			} else {
				if err != nil {
					t.Errorf("Float(%v).Compare(%v) unexpected error: %v", tt.f, tt.other, err)
				}
				if cmp != tt.expectedCmp {
					t.Errorf("Float(%v).Compare(%v) = %d, expected %d", tt.f, tt.other, cmp, tt.expectedCmp)
				}
			}
		})
	}
}

// TestParseFloat tests parsing string literals into Float values, handling scientific notation, overflow, underflow,
// and invalid syntax.
func TestParseFloat(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Float
		expectError bool
		checkNaN    bool
	}{
		{
			name:     "INF literal",
			input:    "INF",
			expected: GetFloatInfinity(),
		},
		{
			name:     "+INF literal",
			input:    "+INF",
			expected: GetFloatInfinity(),
		},
		{
			name:     "-INF literal",
			input:    "-INF",
			expected: GetFloatNegInfinity(),
		},
		{
			name:     "NaN literal",
			input:    "NaN",
			checkNaN: true,
		},
		{
			name:        "invalid Go hex float syntax lowercase x",
			input:       "0x1.0p0",
			expectError: true,
		},
		{
			name:        "invalid Go hex float syntax uppercase X",
			input:       "0X1.0P0",
			expectError: true,
		},
		{
			name:        "invalid binary exponent p",
			input:       "1.0p2",
			expectError: true,
		},
		{
			name:        "invalid binary exponent P",
			input:       "1.0P2",
			expectError: true,
		},
		{
			name:        "invalid underscore in literal",
			input:       "1_000.0",
			expectError: true,
		},
		{
			name:        "syntax error non-numeric",
			input:       "abc",
			expectError: true,
		},
		{
			name:        "Go-specific Infinity rejected",
			input:       "Infinity",
			expectError: true,
		},
		{
			name:        "Go-specific -Infinity rejected",
			input:       "-Infinity",
			expectError: true,
		},
		{
			name:        "Go-specific lowercase nan rejected",
			input:       "nan",
			expectError: true,
		},
		{
			name:     "standard positive float",
			input:    "12.34",
			expected: Float(12.34),
		},
		{
			name:     "standard negative float",
			input:    "-12.34",
			expected: Float(-12.34),
		},
		{
			name:     "scientific notation positive exponent",
			input:    "1.23E2",
			expected: Float(123.0),
		},
		{
			name:     "scientific notation negative exponent",
			input:    "1.23e-2",
			expected: Float(0.0123),
		},
		{
			name:     "zero literal",
			input:    "0",
			expected: Float(0.0),
		},
		{
			name:     "negative zero literal",
			input:    "-0",
			expected: Float(float32(math.Copysign(0, -1))),
		},
		{
			name:     "overflow maps to infinity",
			input:    "1e50",
			expected: GetFloatInfinity(),
		},
		{
			name:     "negative overflow maps to negative infinity",
			input:    "-1e50",
			expected: GetFloatNegInfinity(),
		},
		{
			name:     "underflow maps to zero",
			input:    "1e-50",
			expected: Float(0.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runParseFloatTest(t, tt)
		})
	}
}

func runParseFloatTest(t *testing.T, tt struct {
	name        string
	input       string
	expected    Float
	expectError bool
	checkNaN    bool
}) {
	res, err := ParseFloat(tt.input)
	if tt.expectError {
		if err == nil {
			t.Errorf("ParseFloat(%q) expected error, got nil", tt.input)
		}
		return
	}
	if err != nil {
		t.Fatalf("ParseFloat(%q) unexpected error: %v", tt.input, err)
	}
	if tt.checkNaN {
		if !res.IsNaN() {
			t.Errorf("ParseFloat(%q) = %v, expected NaN", tt.input, res)
		}
		return
	}
	if !res.IsIdenticalWith(tt.expected) {
		t.Errorf("ParseFloat(%q) = %v, expected %v", tt.input, res, tt.expected)
	}
}

// TestFloatString tests the string formatting of Float values to ensure they match expected XSD canonical
// representations.
func TestFloatString(t *testing.T) {
	tests := []struct {
		name     string
		f        Float
		expected string
	}{
		{
			name:     "positive zero",
			f:        Float(0.0),
			expected: "0.0E0",
		},
		{
			name:     "negative zero",
			f:        Float(float32(math.Copysign(0, -1))),
			expected: "-0.0E0",
		},
		{
			name:     "positive infinity",
			f:        GetFloatInfinity(),
			expected: "INF",
		},
		{
			name:     "negative infinity",
			f:        GetFloatNegInfinity(),
			expected: "-INF",
		},
		{
			name:     "NaN",
			f:        GetFloatNaN(),
			expected: "NaN",
		},
		{
			name:     "whole number adding fractional zero",
			f:        Float(1.0),
			expected: "1.0E0",
		},
		{
			name:     "whole power of 10",
			f:        Float(100.0),
			expected: "1.0E2",
		},
		{
			name:     "decimal with positive exponent",
			f:        Float(1234.5),
			expected: "1.2345E3",
		},
		{
			name:     "decimal with negative exponent",
			f:        Float(0.0125),
			expected: "1.25E-2",
		},
		{
			name:     "negative decimal value",
			f:        Float(-0.0125),
			expected: "-1.25E-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.f.String()
			if str != tt.expected {
				t.Errorf("Float(%v).String() = %q, expected %q", tt.f, str, tt.expected)
			}
		})
	}
}

// TestCanonicalFloatStringDirect tests the underlying canonical string generation helper for both float32 and float64
// parameters.
func TestCanonicalFloatStringDirect(t *testing.T) {
	tests := []struct {
		name      string
		val       float64
		isFloat32 bool
		expected  string
	}{
		{
			name:      "isFloat32 false positive infinity",
			val:       math.Inf(1),
			isFloat32: false,
			expected:  "INF",
		},
		{
			name:      "isFloat32 false negative infinity",
			val:       math.Inf(-1),
			isFloat32: false,
			expected:  "-INF",
		},
		{
			name:      "isFloat32 false NaN",
			val:       math.NaN(),
			isFloat32: false,
			expected:  "NaN",
		},
		{
			name:      "isFloat32 false positive zero",
			val:       0.0,
			isFloat32: false,
			expected:  "0.0E0",
		},
		{
			name:      "isFloat32 false negative zero",
			val:       math.Copysign(0, -1),
			isFloat32: false,
			expected:  "-0.0E0",
		},
		{
			name:      "isFloat32 false whole power of 10 without dot in mantissa",
			val:       1000.0,
			isFloat32: false,
			expected:  "1.0E3",
		},
		{
			name:      "isFloat32 false decimal value",
			val:       1.23456789012345,
			isFloat32: false,
			expected:  "1.23456789012345E0",
		},
		{
			name:      "isFloat32 false negative exponent",
			val:       1.5e-10,
			isFloat32: false,
			expected:  "1.5E-10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := canonicalFloatString(tt.val, tt.isFloat32)
			if res != tt.expected {
				t.Errorf("canonicalFloatString(%v, %v) = %q, expected %q", tt.val, tt.isFloat32, res, tt.expected)
			}
		})
	}
}
