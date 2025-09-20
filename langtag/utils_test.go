/*
Copyright 2025 Trident Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUTHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package langtag

import (
	"strings"
	"testing"
)

// Test_isAlpha verifies the isAlpha function according to RFC 5646, Section 2.1,
// which defines ALPHA as A-Z / a-z.
func Test_isAlpha(t *testing.T) {
	tests := []struct {
		name     string
		b        byte
		expected bool
	}{
		{name: "lowercase a", b: 'a', expected: true},
		{name: "lowercase z", b: 'z', expected: true},
		{name: "uppercase A", b: 'A', expected: true},
		{name: "uppercase Z", b: 'Z', expected: true},
		{name: "middle lowercase", b: 'm', expected: true},
		{name: "middle uppercase", b: 'M', expected: true},
		{name: "digit 0", b: '0', expected: false},
		{name: "digit 9", b: '9', expected: false},
		{name: "hyphen", b: '-', expected: false},
		{name: "space", b: ' ', expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlpha(tt.b); got != tt.expected {
				t.Errorf("isAlpha('%c') = %v, want %v", tt.b, got, tt.expected)
			}
		})
	}
}

// Test_isDigit verifies the isDigit function according to RFC 5646, Section 2.1,
// which defines DIGIT as 0-9.
func Test_isDigit(t *testing.T) {
	tests := []struct {
		name     string
		b        byte
		expected bool
	}{
		{name: "digit 0", b: '0', expected: true},
		{name: "digit 9", b: '9', expected: true},
		{name: "digit 5", b: '5', expected: true},
		{name: "lowercase a", b: 'a', expected: false},
		{name: "uppercase Z", b: 'Z', expected: false},
		{name: "hyphen", b: '-', expected: false},
		{name: "space", b: ' ', expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDigit(tt.b); got != tt.expected {
				t.Errorf("isDigit('%c') = %v, want %v", tt.b, got, tt.expected)
			}
		})
	}
}

// Test_isAlphanum verifies the isAlphanum function according to RFC 5646, Section 2.1,
// which defines alphanum as ALPHA / DIGIT.
func Test_isAlphanum(t *testing.T) {
	tests := []struct {
		name     string
		b        byte
		expected bool
	}{
		{name: "lowercase a", b: 'a', expected: true},
		{name: "uppercase Z", b: 'Z', expected: true},
		{name: "digit 0", b: '0', expected: true},
		{name: "digit 9", b: '9', expected: true},
		{name: "middle lowercase", b: 'm', expected: true},
		{name: "middle uppercase", b: 'M', expected: true},
		{name: "middle digit", b: '5', expected: true},
		{name: "hyphen", b: '-', expected: false},
		{name: "space", b: ' ', expected: false},
		{name: "at symbol", b: '@', expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlphanum(tt.b); got != tt.expected {
				t.Errorf("isAlphanum('%c') = %v, want %v", tt.b, got, tt.expected)
			}
		})
	}
}

// Test_isLangtagChar verifies the isLangtagChar function based on RFC 5646, Section 2.1,
// which specifies that language tags are composed of alphanumeric subtags separated by hyphens.
func Test_isLangtagChar(t *testing.T) {
	tests := []struct {
		name     string
		r        rune
		expected bool
	}{
		{name: "lowercase a", r: 'a', expected: true},
		{name: "uppercase Z", r: 'Z', expected: true},
		{name: "digit 0", r: '0', expected: true},
		{name: "digit 9", r: '9', expected: true},
		{name: "hyphen", r: '-', expected: true},
		{name: "space", r: ' ', expected: false},
		{name: "underscore", r: '_', expected: false},
		{name: "non-ASCII letter", r: 'é', expected: false},
		{name: "percent symbol", r: '%', expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLangtagChar(tt.r); got != tt.expected {
				t.Errorf("isLangtagChar('%c') = %v, want %v", tt.r, got, tt.expected)
			}
		})
	}
}

// Test_isAlphabetic verifies the isAlphabetic function against RFC 5646, Section 2.1,
// which defines purely alphabetic subtags like 'script' (4ALPHA).
func Test_isAlphabetic(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected bool
	}{
		{name: "ISO 639-1 language", s: "en", expected: true},
		{name: "ISO 15924 script", s: "Latn", expected: true},
		{name: "ISO 3166-1 region", s: "DE", expected: true},
		{name: "mixed case", s: "MiXeD", expected: true},
		{name: "contains digit", s: "en1", expected: false},
		{name: "contains hyphen", s: "en-gb", expected: false},
		{name: "contains space", s: "e n", expected: false},
		{name: "empty string", s: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlphabetic(tt.s); got != tt.expected {
				t.Errorf("isAlphabetic(%q) = %v, want %v", tt.s, got, tt.expected)
			}
		})
	}
}

// Test_isNumeric verifies the isNumeric function against RFC 5646, Section 2.1,
// which defines purely numeric subtags like some 'region' codes (3DIGIT).
func Test_isNumeric(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected bool
	}{
		{name: "UN M.49 region", s: "419", expected: true},
		{name: "leading zero", s: "001", expected: true},
		{name: "simple numeric", s: "123", expected: true},
		{name: "contains letter", s: "12a", expected: false},
		{name: "contains hyphen", s: "1-2", expected: false},
		{name: "contains space", s: "1 2", expected: false},
		{name: "empty string", s: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNumeric(tt.s); got != tt.expected {
				t.Errorf("isNumeric(%q) = %v, want %v", tt.s, got, tt.expected)
			}
		})
	}
}

// Test_isAlphanumeric verifies the isAlphanumeric function against RFC 5646, Section 2.1,
// which defines alphanumeric subtags like 'variant' (e.g., 5*8alphanum).
func Test_isAlphanumeric(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected bool
	}{
		{name: "variant year", s: "1996", expected: true},
		{name: "variant name", s: "rozaj", expected: true},
		{name: "alphanumeric mix", s: "a1b2c3d4", expected: true},
		{name: "purely alpha", s: "nedis", expected: true},
		{name: "purely numeric", s: "007", expected: true},
		{name: "contains hyphen", s: "variant-1", expected: false},
		{name: "contains space", s: "variant 1", expected: false},
		{name: "empty string", s: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlphanumeric(tt.s); got != tt.expected {
				t.Errorf("isAlphanumeric(%q) = %v, want %v", tt.s, got, tt.expected)
			}
		})
	}
}

// Test_writeTitleCase verifies the writeTitleCase function against RFC 5646, Section 2.1.1,
// which recommends title case for script subtags (e.g., 'Cyrl').
func Test_writeTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "lowercase script", input: "cyrl", expected: "Cyrl"},
		{name: "uppercase script", input: "LATN", expected: "Latn"},
		{name: "mixed case script", input: "cYrL", expected: "Cyrl"},
		{name: "already title case", input: "Arab", expected: "Arab"},
		{name: "empty string", input: "", expected: ""},
		{name: "single lowercase char", input: "s", expected: "S"},
		{name: "single uppercase char", input: "S", expected: "S"},
		{name: "lowercase 'i'", input: "i", expected: "I"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			writeTitleCase(&b, tt.input)
			got := b.String()
			if got != tt.expected {
				t.Errorf("writeTitleCase(%q) wrote %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
