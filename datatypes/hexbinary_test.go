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
	"bytes"
	"testing"
)

// TestParseHexBinary verifies the parsing of valid and invalid hexBinary string literals.
func TestParseHexBinary(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    HexBinary
		expectError bool
	}{
		{
			name:        "valid empty string",
			input:       "",
			expected:    HexBinary{},
			expectError: false,
		},
		{
			name:        "valid lower-case hex digits",
			input:       "0fb7",
			expected:    HexBinary{0x0f, 0xb7},
			expectError: false,
		},
		{
			name:        "valid upper-case hex digits",
			input:       "0FB7",
			expected:    HexBinary{0x0f, 0xb7},
			expectError: false,
		},
		{
			name:        "odd number of characters error",
			input:       "0fb",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid hex characters error",
			input:       "0fzz",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHexBinary(tt.input)
			if (err != nil) != tt.expectError {
				t.Fatalf("ParseHexBinary(%q) error = %v, expectError = %v", tt.input, err, tt.expectError)
			}
			if !tt.expectError && !bytes.Equal(got, tt.expected) {
				t.Errorf("ParseHexBinary(%q) = %v, expected %v", tt.input, got, tt.expected)
			}
		})
	}
}

// TestHexBinaryString verifies canonical serialization of HexBinary values into uppercase hexadecimal strings.
func TestHexBinaryString(t *testing.T) {
	tests := []struct {
		name     string
		val      HexBinary
		expected string
	}{
		{
			name:     "empty hexBinary",
			val:      HexBinary{},
			expected: "",
		},
		{
			name:     "single byte",
			val:      HexBinary{0x0a},
			expected: "0A",
		},
		{
			name:     "multiple bytes with lowercase letters converted to uppercase",
			val:      HexBinary{0x0f, 0xb7, 0x1c},
			expected: "0FB71C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.val.String()
			if got != tt.expected {
				t.Errorf("HexBinary(%v).String() = %q, expected %q", tt.val, got, tt.expected)
			}
		})
	}
}

// TestHexBinaryIsIdenticalWith tests the identity comparison logic for HexBinary values.
func TestHexBinaryIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name     string
		val      HexBinary
		other    XSDValue
		expected bool
	}{
		{
			name:     "identical byte sequences",
			val:      HexBinary{0x01, 0x02, 0x03},
			other:    HexBinary{0x01, 0x02, 0x03},
			expected: true,
		},
		{
			name:     "identical empty slices",
			val:      HexBinary{},
			other:    HexBinary{},
			expected: true,
		},
		{
			name:     "different byte values",
			val:      HexBinary{0x01, 0x02},
			other:    HexBinary{0x01, 0x03},
			expected: false,
		},
		{
			name:     "different lengths",
			val:      HexBinary{0x01},
			other:    HexBinary{0x01, 0x02},
			expected: false,
		},
		{
			name:     "different XSDValue type",
			val:      HexBinary{0x01},
			other:    String("01"),
			expected: false,
		},
		{
			name:     "nil other value",
			val:      HexBinary{0x01},
			other:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.val.IsIdenticalWith(tt.other)
			if got != tt.expected {
				t.Errorf("HexBinary(%v).IsIdenticalWith(%v) = %v, expected %v", tt.val, tt.other, got, tt.expected)
			}
		})
	}
}

// TestHexBinaryLength verifies the length measurement in terms of octet count.
func TestHexBinaryLength(t *testing.T) {
	tests := []struct {
		name     string
		val      HexBinary
		expected int
	}{
		{
			name:     "empty byte slice",
			val:      HexBinary{},
			expected: 0,
		},
		{
			name:     "single octet",
			val:      HexBinary{0xff},
			expected: 1,
		},
		{
			name:     "three octets",
			val:      HexBinary{0x01, 0x02, 0x03},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.val.Length()
			if got != tt.expected {
				t.Errorf("HexBinary(%v).Length() = %d, expected %d", tt.val, got, tt.expected)
			}
		})
	}
}
