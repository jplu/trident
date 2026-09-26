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
	"testing"
)

// TestParseNormalizedString tests the parsing and whitespace normalization behavior of ParseNormalizedString.
func TestParseNormalizedString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected NormalizedString
	}{
		{
			name:     "plain string",
			input:    "hello world",
			expected: NormalizedString("hello world"),
		},
		{
			name:     "string with tabs, line feeds, and carriage returns",
			input:    "hello\tworld\r\ntest",
			expected: NormalizedString("hello world  test"),
		},
		{
			name:     "empty string",
			input:    "",
			expected: NormalizedString(""),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseNormalizedString(tc.input)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

// TestNormalizedStringString tests the String method of NormalizedString.
func TestNormalizedStringString(t *testing.T) {
	ns := NormalizedString("valid normalized string")
	if got := ns.String(); got != "valid normalized string" {
		t.Fatalf("expected %q, got %q", "valid normalized string", got)
	}
}

// TestNormalizedStringIsIdenticalWith tests the value space identity comparison for NormalizedString.
func TestNormalizedStringIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name     string
		receiver NormalizedString
		other    XSDValue
		expected bool
	}{
		{
			name:     "identical NormalizedString values",
			receiver: NormalizedString("sample"),
			other:    NormalizedString("sample"),
			expected: true,
		},
		{
			name:     "different NormalizedString values",
			receiver: NormalizedString("sample"),
			other:    NormalizedString("other"),
			expected: false,
		},
		{
			name:     "incompatible XSDValue type",
			receiver: NormalizedString("sample"),
			other:    String("sample"),
			expected: false,
		},
		{
			name:     "nil XSDValue",
			receiver: NormalizedString("sample"),
			other:    nil,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.receiver.IsIdenticalWith(tc.other)
			if got != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

// TestNormalizedStringLength tests the rune-based length calculation of NormalizedString.
func TestNormalizedStringLength(t *testing.T) {
	tests := []struct {
		name     string
		input    NormalizedString
		expected int
	}{
		{
			name:     "empty string",
			input:    NormalizedString(""),
			expected: 0,
		},
		{
			name:     "ascii string",
			input:    NormalizedString("hello"),
			expected: 5,
		},
		{
			name:     "multi-byte unicode string",
			input:    NormalizedString("こんにちは"),
			expected: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.input.Length(); got != tc.expected {
				t.Fatalf("expected length %d, got %d", tc.expected, got)
			}
		})
	}
}
