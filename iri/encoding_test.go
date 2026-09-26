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
package iri

import (
	"fmt"
	"strings"
	"testing"
)

// newTestParser creates an iriParser with a string-based output buffer for testing.
func newTestParser(input string, unchecked bool) (*iriParser, *stringOutputBuffer) {
	output := &stringOutputBuffer{builder: &strings.Builder{}}
	parser := &iriParser{
		iri:       input,
		input:     newParserInput(input),
		output:    output,
		unchecked: unchecked,
	}
	return parser, output
}

// TestPercentEncode tests the percent-encoding of non-ASCII characters.
func TestPercentEncode(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "ASCII only", input: "abc-123", expected: "abc-123"},
		{name: "Non-ASCII only (résumé)", input: "résumé", expected: "r%C3%A9sum%C3%A9"},
		{
			name:     "Mixed ASCII and non-ASCII",
			input:    "hello-résumé",
			expected: "hello-r%C3%A9sum%C3%A9",
		},
		{name: "Empty string", input: "", expected: ""},
		{name: "Non-BMP character (Old Italic)", input: "\U00010300", expected: "%F0%90%8C%80"},
		{name: "RFC 3986 example Katakana A", input: "\u30A2", expected: "%E3%82%A2"},
		{name: "Lax ASCII space", input: "foo bar", expected: "foo%20bar"},
		{name: "Lax ASCII angle brackets", input: "foo<bar>", expected: "foo%3Cbar%3E"},
		{
			name:     "Lax ASCII braces and backslash",
			input:    "foo{bar}\\baz",
			expected: "foo%7Bbar%7D%5Cbaz",
		},
		{name: "Lax ASCII pipe, caret and backtick", input: "a|b^c`d", expected: "a%7Cb%5Ec%60d"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			percentEncode(tc.input, &b)
			result := b.String()
			if result != tc.expected {
				t.Errorf("percentEncode(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestPercentEncodeRune tests the percent-encoding of a single rune.
func TestPercentEncodeRune(t *testing.T) {
	t.Run("string output buffer", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    rune
			expected string
		}{
			{name: "ASCII rune", input: 'a', expected: "a"},
			{name: "Non-ASCII rune (é)", input: 'é', expected: "%C3%A9"},
			{name: "Non-BMP rune (𐌀)", input: '\U00010300', expected: "%F0%90%8C%80"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				output := &stringOutputBuffer{builder: &strings.Builder{}}
				percentEncodeRune(tc.input, output)
				result := output.string()
				if result != tc.expected {
					t.Errorf("percentEncodeRune(%q) = %q; want %q", tc.input, result, tc.expected)
				}
			})
		}
	})

	t.Run("void output buffer", func(t *testing.T) {
		testCases := []struct {
			name        string
			input       rune
			expectedLen int
		}{
			{name: "ASCII rune", input: 'a', expectedLen: 1},
			{name: "Non-ASCII rune (é)", input: 'é', expectedLen: 6},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				output := &voidOutputBuffer{}
				percentEncodeRune(tc.input, output)
				resultLen := output.len()
				if resultLen != tc.expectedLen {
					t.Errorf(
						"len(percentEncodeRune(%q)) = %d; want %d",
						tc.input,
						resultLen,
						tc.expectedLen,
					)
				}
			})
		}
	})
}

// testReadEcharSuccess tests successful reading of percent-encoded characters.
func testReadEcharSuccess(t *testing.T) {
	t.Helper()
	testCases := []struct {
		name          string
		input         string
		expectedStr   string
		expectedInput string
	}{
		{
			name:          "Valid encoding uppercase",
			input:         "20rest",
			expectedStr:   "%20",
			expectedInput: "rest",
		},
		{
			name:          "Valid encoding lowercase",
			input:         "3arest",
			expectedStr:   "%3a",
			expectedInput: "rest",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, output := newTestParser(tc.input, false)
			err := p.readEchar()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output.string() != tc.expectedStr {
				t.Errorf("output string mismatch: got %q, want %q", output.string(), tc.expectedStr)
			}
			if p.input.asStr() != tc.expectedInput {
				t.Errorf(
					"remaining input mismatch: got %q, want %q",
					p.input.asStr(),
					tc.expectedInput,
				)
			}
		})
	}
}

// testReadEcharError tests error handling when reading percent-encoded characters.
func testReadEcharError(t *testing.T) {
	t.Helper()
	testCases := []struct {
		name        string
		input       string
		expectedErr string
	}{
		{
			name:        "Incomplete encoding - one char",
			input:       "3",
			expectedErr: "Invalid IRI percent encoding '%3'",
		},
		{
			name:        "Incomplete encoding - no chars",
			input:       "",
			expectedErr: "Invalid IRI percent encoding '%'",
		},
		{
			name:        "Invalid encoding - non-hex first char",
			input:       "G0",
			expectedErr: "Invalid IRI percent encoding '%G0'",
		},
		{
			name:        "Invalid encoding - non-hex second char",
			input:       "0G",
			expectedErr: "Invalid IRI percent encoding '%0G'",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := newTestParser(tc.input, false)
			err := p.readEchar()
			if err == nil {
				t.Fatalf("expected error '%s', but got nil", tc.expectedErr)
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("expected error '%s', got '%s'", tc.expectedErr, err.Error())
			}
		})
	}
}

// TestIriParserreadEchar tests the parsing of a percent-encoded triplet.
func TestIriParserreadEchar(t *testing.T) {
	t.Run("success", testReadEcharSuccess)
	t.Run("error", testReadEcharError)
}

// testReadURLCodepointOrEcharSuccess tests successful reading of URL codepoints or encoded characters.
func testReadURLCodepointOrEcharSuccess(t *testing.T) {
	t.Helper()
	pathCharValidator := func(c rune) bool {
		return isIUnreservedOrSubDelims(c) || c == ':' || c == '@'
	}
	testCases := []struct {
		name          string
		inputRune     rune
		parserInput   string
		unchecked     bool
		validFunc     func(rune) bool
		expectedStr   string
		expectedInput string
	}{
		{
			name:          "Valid unreserved char",
			inputRune:     'a',
			validFunc:     pathCharValidator,
			expectedStr:   "a",
			expectedInput: "",
		},
		{
			name:          "Valid sub-delim char",
			inputRune:     '!',
			validFunc:     pathCharValidator,
			expectedStr:   "!",
			expectedInput: "",
		},
		{
			name:          "Valid colon in path",
			inputRune:     ':',
			validFunc:     pathCharValidator,
			expectedStr:   ":",
			expectedInput: "",
		},
		{
			name:          "Percent encoding trigger",
			inputRune:     '%',
			parserInput:   "20rest",
			validFunc:     pathCharValidator,
			expectedStr:   "%20",
			expectedInput: "rest",
		},
		{
			name:          "Unchecked mode with invalid char",
			inputRune:     '<',
			unchecked:     true,
			validFunc:     func(_ rune) bool { return false },
			expectedStr:   "<",
			expectedInput: "",
		},
		{
			name:          "Lax ASCII character - space",
			inputRune:     ' ',
			validFunc:     func(_ rune) bool { return false },
			expectedStr:   "%20",
			expectedInput: "",
		},
		{
			name:          "Lax ASCII character - brace",
			inputRune:     '{',
			validFunc:     func(_ rune) bool { return false },
			expectedStr:   "%7B",
			expectedInput: "",
		},
		{
			name:          "Valid non-ASCII char",
			inputRune:     'é',
			validFunc:     pathCharValidator,
			expectedStr:   "é",
			expectedInput: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, output := newTestParser(tc.parserInput, tc.unchecked)
			err := p.readURLCodepointOrEchar(tc.inputRune, tc.validFunc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if output.string() != tc.expectedStr {
				t.Errorf("output string mismatch: got %q, want %q", output.string(), tc.expectedStr)
			}
			if p.input.asStr() != tc.expectedInput {
				t.Errorf(
					"remaining input mismatch: got %q, want %q",
					p.input.asStr(),
					tc.expectedInput,
				)
			}
		})
	}
}

// testReadURLCodepointOrEcharError tests error handling when reading URL codepoints or encoded characters.
func testReadURLCodepointOrEcharError(t *testing.T) {
	t.Helper()
	pathCharValidator := func(c rune) bool {
		return isIUnreservedOrSubDelims(c) || c == ':' || c == '@'
	}
	testCases := []struct {
		name        string
		inputRune   rune
		parserInput string
		validFunc   func(rune) bool
		expectedErr string
	}{
		{
			name:        "Invalid character - newline",
			inputRune:   '\n',
			validFunc:   pathCharValidator,
			expectedErr: fmt.Sprintf("Invalid IRI character '%c'", '\n'),
		},
		{
			name:        "Invalid percent-encoding sequence without hex digits",
			inputRune:   '%',
			parserInput: "",
			validFunc:   pathCharValidator,
			expectedErr: "Invalid percent-encoding sequence '%'",
		},
		{
			name:        "Invalid percent-encoding sequence with non-hex characters",
			inputRune:   '%',
			parserInput: "ZZ",
			validFunc:   pathCharValidator,
			expectedErr: "Invalid percent-encoding sequence '%'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := newTestParser(tc.parserInput, false)
			err := p.readURLCodepointOrEchar(tc.inputRune, tc.validFunc)
			if err == nil {
				t.Fatalf("expected an error, but got nil")
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("expected error %q, got %q", tc.expectedErr, err.Error())
			}
		})
	}
}

// TestIriParserreadURLCodepointOrEchar tests the dispatcher for reading a character or percent-encoded sequence.
func TestIriParserreadURLCodepointOrEchar(t *testing.T) {
	t.Run("success", testReadURLCodepointOrEcharSuccess)
	t.Run("error", testReadURLCodepointOrEcharError)
}
