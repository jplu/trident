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
	"testing"
)

// TestParseString verifies that ParseString correctly creates a String value and returns no error.
func TestParseString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected String
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple ASCII string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "unicode string",
			input:    "日本語",
			expected: "日本語",
		},
		{
			name:     "string with whitespace",
			input:    " \t\n\r ",
			expected: " \t\n\r ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := ParseString(tt.input)
			if val != tt.expected {
				t.Errorf("ParseString(%q) = %q, expected %q", tt.input, val, tt.expected)
			}
		})
	}
}

// TestStringString verifies that the String method returns the underlying string literal unchanged.
func TestStringString(t *testing.T) {
	tests := []struct {
		name     string
		input    String
		expected string
	}{
		{
			name:     "empty string",
			input:    String(""),
			expected: "",
		},
		{
			name:     "ASCII string",
			input:    String("sample text"),
			expected: "sample text",
		},
		{
			name:     "multi-byte unicode string",
			input:    String("こんにちは"),
			expected: "こんにちは",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.String(); got != tt.expected {
				t.Errorf("String.String() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

// TestStringIsIdenticalWith tests identity comparison between String instances and other XSD values.
func TestStringIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name     string
		base     String
		other    XSDValue
		expected bool
	}{
		{
			name:     "identical strings",
			base:     String("value"),
			other:    String("value"),
			expected: true,
		},
		{
			name:     "different strings",
			base:     String("value1"),
			other:    String("value2"),
			expected: false,
		},
		{
			name:     "identical empty strings",
			base:     String(""),
			other:    String(""),
			expected: true,
		},
		{
			name:     "different type implementing XSDValue",
			base:     String("true"),
			other:    Boolean(true),
			expected: false,
		},
		{
			name:     "nil comparison",
			base:     String("value"),
			other:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.base.IsIdenticalWith(tt.other)
			if got != tt.expected {
				t.Errorf("String(%q).IsIdenticalWith(%v) = %v, expected %v", tt.base, tt.other, got, tt.expected)
			}
		})
	}
}

// TestStringLength tests character-based length measurement according to XML Schema rules.
func TestStringLength(t *testing.T) {
	tests := []struct {
		name     string
		input    String
		expected int
	}{
		{
			name:     "empty string",
			input:    String(""),
			expected: 0,
		},
		{
			name:     "ASCII characters",
			input:    String("hello"),
			expected: 5,
		},
		{
			name:     "multi-byte unicode characters",
			input:    String("日本語"),
			expected: 3,
		},
		{
			name:     "surrogate pair / astral code points",
			input:    String("𝄞🎵"),
			expected: 2,
		},
		{
			name:     "mixed ASCII and unicode",
			input:    String("hello 世界!"),
			expected: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.Length(); got != tt.expected {
				t.Errorf("String(%q).Length() = %d, expected %d", tt.input, got, tt.expected)
			}
		})
	}
}
