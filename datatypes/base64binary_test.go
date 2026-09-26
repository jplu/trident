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

// TestParseBase64Binary tests parsing base64 binary strings.
func TestParseBase64Binary(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  Base64Binary
		expectErr bool
	}{
		{
			name:      "valid simple base64",
			input:     "SGVsbG8gV29ybGQ=",
			expected:  Base64Binary([]byte("Hello World")),
			expectErr: false,
		},
		{
			name:      "valid empty string",
			input:     "",
			expected:  Base64Binary([]byte{}),
			expectErr: false,
		},
		{
			name:      "valid base64 with allowed whitespace (spaces, tabs, newlines, CR)",
			input:     " S G\tV s b G 8 g\r\nV 2 9 y b G Q = ",
			expected:  Base64Binary([]byte("Hello World")),
			expectErr: false,
		},
		{
			name:      "invalid base64 characters",
			input:     "!!!NotBase64!!!",
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "invalid base64 missing padding length",
			input:     "SGVsbG8",
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBase64Binary(tt.input)
			if (err != nil) != tt.expectErr {
				t.Fatalf("ParseBase64Binary(%q) error = %v, expectErr %v", tt.input, err, tt.expectErr)
			}
			if !tt.expectErr {
				if !bytes.Equal(got, tt.expected) {
					t.Errorf("ParseBase64Binary(%q) = %v, expected %v", tt.input, got, tt.expected)
				}
			}
		})
	}
}

// TestBase64BinaryString tests the string representation of Base64Binary.
func TestBase64BinaryString(t *testing.T) {
	tests := []struct {
		name     string
		b        Base64Binary
		expected string
	}{
		{
			name:     "canonical string representation for non-empty bytes",
			b:        Base64Binary([]byte("Hello World")),
			expected: "SGVsbG8gV29ybGQ=",
		},
		{
			name:     "canonical string representation for empty bytes",
			b:        Base64Binary([]byte{}),
			expected: "",
		},
		{
			name:     "canonical string representation for nil bytes",
			b:        nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.String(); got != tt.expected {
				t.Errorf("Base64Binary.String() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

// TestBase64BinaryIsIdenticalWith tests the identity comparison of Base64Binary values.
func TestBase64BinaryIsIdenticalWith(t *testing.T) {
	b1 := Base64Binary([]byte("test data"))
	b2 := Base64Binary([]byte("test data"))
	b3 := Base64Binary([]byte("different data"))

	tests := []struct {
		name     string
		b        Base64Binary
		other    XSDValue
		expected bool
	}{
		{
			name:     "identical Base64Binary values",
			b:        b1,
			other:    b2,
			expected: true,
		},
		{
			name:     "non-identical Base64Binary values",
			b:        b1,
			other:    b3,
			expected: false,
		},
		{
			name:     "comparison with different XSDValue type",
			b:        b1,
			other:    String("test data"),
			expected: false,
		},
		{
			name:     "both nil Base64Binary",
			b:        Base64Binary(nil),
			other:    Base64Binary(nil),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.IsIdenticalWith(tt.other); got != tt.expected {
				t.Errorf("Base64Binary.IsIdenticalWith() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

// TestBase64BinaryLength tests the length calculation of Base64Binary values.
func TestBase64BinaryLength(t *testing.T) {
	tests := []struct {
		name     string
		b        Base64Binary
		expected int
	}{
		{
			name:     "length of non-empty byte slice",
			b:        Base64Binary([]byte{1, 2, 3, 4, 5}),
			expected: 5,
		},
		{
			name:     "length of empty byte slice",
			b:        Base64Binary([]byte{}),
			expected: 0,
		},
		{
			name:     "length of nil byte slice",
			b:        nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.Length(); got != tt.expected {
				t.Errorf("Base64Binary.Length() = %d, expected %d", got, tt.expected)
			}
		})
	}
}
