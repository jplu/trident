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
	"regexp"
	"testing"
)

// TestXMLNameCharClasses verifies that the XML Name character classes are valid compiled regular expressions.
func TestXMLNameCharClasses(_ *testing.T) {
	_ = regexp.MustCompile(xmlNameStartCharClass)
	_ = regexp.MustCompile(xmlNameCharClass)
}

// TestTranslateUnicodeEscapeShortOrInvalidPrefix tests escape inputs that are too short or lack the proper prefix.
func TestTranslateUnicodeEscapeShortOrInvalidPrefix(t *testing.T) {
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
			name:     "length less than 5 with p prefix",
			input:    `\p{`,
			expected: `\p{`,
		},
		{
			name:     "length less than 5 without p prefix",
			input:    "abc",
			expected: "abc",
		},
		{
			name:     "length >= 5 but neither p nor P escape prefix",
			input:    `\w{IsBasicLatin}`,
			expected: `\w{IsBasicLatin}`,
		},
		{
			name:     "length >= 5 starting with brackets",
			input:    "[12345]",
			expected: "[12345]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := translateUnicodeEscape(tt.input)
			if actual != tt.expected {
				t.Errorf("translateUnicodeEscape(%q) = %q, expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}

// TestTranslateUnicodeEscapeWithoutIsPrefix tests Unicode escapes that do not contain the "Is" prefix.
func TestTranslateUnicodeEscapeWithoutIsPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single letter category p escape",
			input:    `\p{L}`,
			expected: `\p{L}`,
		},
		{
			name:     "multi letter category P escape",
			input:    `\P{Nd}`,
			expected: `\P{Nd}`,
		},
		{
			name:     "property without Is prefix",
			input:    `\p{Letter}`,
			expected: `\p{Letter}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := translateUnicodeEscape(tt.input)
			if actual != tt.expected {
				t.Errorf("translateUnicodeEscape(%q) = %q, expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}

// TestTranslateUnicodeEscapeMappedBlocks tests Unicode block escapes that are mapped to Go-compatible names.
func TestTranslateUnicodeEscapeMappedBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "PrivateUse block mapping",
			input:    `\p{IsPrivateUse}`,
			expected: `\p{Private_Use_Area}`,
		},
		{
			name:     "Greek block mapping with P prefix",
			input:    `\P{IsGreek}`,
			expected: `\P{Greek_and_Coptic}`,
		},
		{
			name:     "CombiningMarksforSymbols block mapping",
			input:    `\p{IsCombiningMarksforSymbols}`,
			expected: `\p{Combining_Diacritical_Marks_for_Symbols}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := translateUnicodeEscape(tt.input)
			if actual != tt.expected {
				t.Errorf("translateUnicodeEscape(%q) = %q, expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}

// TestTranslateUnicodeEscapeDefaultBlocks tests Unicode block escapes that use default block names.
func TestTranslateUnicodeEscapeDefaultBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "BasicLatin block",
			input:    `\p{IsBasicLatin}`,
			expected: `\p{BasicLatin}`,
		},
		{
			name:     "Cyrillic block with negated P escape",
			input:    `\P{IsCyrillic}`,
			expected: `\P{Cyrillic}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := translateUnicodeEscape(tt.input)
			if actual != tt.expected {
				t.Errorf("translateUnicodeEscape(%q) = %q, expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}
