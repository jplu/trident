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

// TestReplaceWhitespace tests the replacement of tabs, line feeds, and carriage returns with spaces.
func TestReplaceWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "tab character replacement",
			input:    "hello\tworld",
			expected: "hello world",
		},
		{
			name:     "newline character replacement",
			input:    "hello\nworld",
			expected: "hello world",
		},
		{
			name:     "carriage return replacement",
			input:    "hello\rworld",
			expected: "hello world",
		},
		{
			name:     "all whitespace characters mixed",
			input:    "\t\r\n",
			expected: "   ",
		},
		{
			name:     "standard spaces unchanged",
			input:    "a b  c",
			expected: "a b  c",
		},
		{
			name:     "multibyte unicode characters",
			input:    "héllo\t✓\nworld\r!",
			expected: "héllo ✓ world !",
		},
		{
			name:     "no whitespace characters",
			input:    "sampleText123",
			expected: "sampleText123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := replaceWhitespace(tt.input)
			if actual != tt.expected {
				t.Errorf("replaceWhitespace(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}

// TestCollapseWhitespace tests the full whitespace collapse normalization algorithm.
func TestCollapseWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "spaces only",
			input:    "    ",
			expected: "",
		},
		{
			name:     "mixed whitespace only",
			input:    " \t \r \n ",
			expected: "",
		},
		{
			name:     "leading and trailing spaces stripped",
			input:    "   test   ",
			expected: "test",
		},
		{
			name:     "multiple consecutive internal spaces collapsed",
			input:    "a    b",
			expected: "a b",
		},
		{
			name:     "mixed internal whitespace collapsed to single spaces",
			input:    "a\t\r\n \t b",
			expected: "a b",
		},
		{
			name:     "word without whitespace",
			input:    "word",
			expected: "word",
		},
		{
			name:     "multiple words with unicode and irregular whitespace",
			input:    "  \t café   \n\r au   lait  ",
			expected: "café au lait",
		},
		{
			name:     "alternating characters and spaces",
			input:    "a b c d",
			expected: "a b c d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := collapseWhitespace(tt.input)
			if actual != tt.expected {
				t.Errorf("collapseWhitespace(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}
