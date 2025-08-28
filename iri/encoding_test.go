/*
Copyright 2025 Trident Authors

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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package iri

import (
	"fmt"
	"strings"
	"testing"
)

// newTestParser is a helper to create an iriParser with a string-based output
// buffer for testing purposes.
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

// TestValidateDecodedBytes tests the validation of UTF-8 byte sequences.
// RFC Reference: RFC 3987, Section 3.2 specifies that percent-decoded octets
// must form valid UTF-8 sequences. Section 4.1 forbids bidirectional
// formatting characters in IRIs.
func TestValidateDecodedBytes(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{
			name:     "Valid UTF-8 with allowed ASCII",
			input:    []byte("hello-world"),
			expected: true,
		},
		{
			name:     "Valid UTF-8 with allowed non-ASCII",
			input:    []byte("résumé"),
			expected: true,
		},
		{
			name:     "Invalid UTF-8 sequence",
			input:    []byte{0xC3, 0x28}, // Invalid second byte for a two-byte sequence.
			expected: false,
		},
		{
			name:     "Forbidden bidi character LRM (U+200E)",
			input:    []byte("\u200e"),
			expected: false,
		},
		{
			name:     "Forbidden bidi character RLM (U+200F)",
			input:    []byte("\u200f"),
			expected: false,
		},
		{
			name:     "Forbidden bidi character LRE (U+202A)",
			input:    []byte("\u202a"),
			expected: false,
		},
		{
			name:     "Empty byte slice",
			input:    []byte{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validateDecodedBytes(tc.input)
			if result != tc.expected {
				t.Errorf("validateDecodedBytes(%v) = %v; want %v", tc.input, result, tc.expected)
			}
		})
	}
}

// TestNormalizePercentEncoding tests the normalization of percent-encoded octets.
// RFC Reference: RFC 3986, Section 6.2.2.2. It specifies that any percent-encoded
// octet corresponding to an unreserved character should be decoded. Unreserved
// characters are defined in RFC 3986, Section 2.3.
func TestNormalizePercentEncoding(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Decode unreserved ALPHA",
			input:    "example.com/%41%42%43",
			expected: "example.com/ABC",
		},
		{
			name:     "Decode unreserved DIGIT",
			input:    "example.com/%31%32%33",
			expected: "example.com/123",
		},
		{
			name:     "Decode unreserved special chars (-._~)",
			input:    "%2D%2E%5F%7E",
			expected: "-._~",
		},
		{
			name:     "Do not decode reserved gen-delims",
			input:    "a%2Fb%3Ac", // a/b:c
			expected: "a%2Fb%3Ac",
		},
		{
			name:     "Do not decode reserved sub-delims",
			input:    "a%24b%26c", // a$b&c
			expected: "a%24b%26c",
		},
		{
			name:     "Do not decode non-ASCII UTF-8 sequence",
			input:    "r%C3%A9sum%C3%A9", // résumé
			expected: "r%C3%A9sum%C3%A9",
		},
		{
			name:     "Mixed reserved and unreserved",
			input:    "a%2Fb%2E%31", // a/b.1
			expected: "a%2Fb.1",
		},
		{
			name:     "Lowercase hex digits are preserved (normalization is only for decoding)",
			input:    "a%2fb%2e%31",
			expected: "a%2fb.1",
		},
		{
			name:     "Invalid encoding - short",
			input:    "a%2",
			expected: "a%2",
		},
		{
			name:     "Invalid encoding - non-hex",
			input:    "a%2G",
			expected: "a%2G",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "No percent encoding",
			input:    "abc-123",
			expected: "abc-123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizePercentEncoding(tc.input)
			if result != tc.expected {
				t.Errorf("normalizePercentEncoding(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestPercentEncode tests the percent-encoding of non-ASCII characters.
// RFC Reference: RFC 3987, Section 3.1, Step 2 defines the mapping from IRI
// characters to URI octets via UTF-8, then percent-encoding. RFC 3986, Section 2.5
// provides examples of this process.
func TestPercentEncode(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ASCII only",
			input:    "abc-123",
			expected: "abc-123",
		},
		{
			name:     "Non-ASCII only (résumé)",
			input:    "résumé",
			expected: "r%C3%A9sum%C3%A9",
		},
		{
			name:     "Mixed ASCII and non-ASCII",
			input:    "hello-résumé",
			expected: "hello-r%C3%A9sum%C3%A9",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Non-BMP character (Old Italic)",
			input:    "\U00010300", // "𐌀"
			expected: "%F0%90%8C%80",
		},
		{
			name:     "RFC 3986 example Katakana A",
			input:    "\u30A2", // "ア"
			expected: "%E3%82%A2",
		},
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
// RFC Reference: Same as TestPercentEncode, this function is a building block
// for the overall IRI-to-URI mapping process.
func TestPercentEncodeRune(t *testing.T) {
	t.Run("string output buffer", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    rune
			expected string
		}{
			{
				name:     "ASCII rune",
				input:    'a',
				expected: "a",
			},
			{
				name:     "Non-ASCII rune (é)",
				input:    'é',
				expected: "%C3%A9",
			},
			{
				name:     "Non-BMP rune (𐌀)",
				input:    '\U00010300',
				expected: "%F0%90%8C%80",
			},
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
			{
				name:        "ASCII rune",
				input:       'a',
				expectedLen: 1,
			},
			{
				name:        "Non-ASCII rune (é)",
				input:       'é',
				expectedLen: 6, // len("%C3%A9")
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				output := &voidOutputBuffer{}
				percentEncodeRune(tc.input, output)
				resultLen := output.len()
				if resultLen != tc.expectedLen {
					t.Errorf("len(percentEncodeRune(%q)) = %d; want %d", tc.input, resultLen, tc.expectedLen)
				}
			})
		}
	})
}

// testReadEcharSuccess is a helper for TestIriParser_readEchar success cases.
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
				t.Errorf("remaining input mismatch: got %q, want %q", p.input.asStr(), tc.expectedInput)
			}
		})
	}
}

// testReadEcharError is a helper for TestIriParser_readEchar error cases.
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

// TestIriParser_readEchar tests the parsing of a percent-encoded triplet.
// RFC Reference: RFC 3986, Section 2.1 defines a percent-encoded octet as a '%'
// followed by two hexadecimal digits.
func TestIriParser_readEchar(t *testing.T) {
	t.Run("success", testReadEcharSuccess)
	t.Run("error", testReadEcharError)
}

// testReadURLCodepointOrEcharSuccess is a helper for TestIriParser_readURLCodepointOrEchar success cases.
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
			validFunc:     func(_ rune) bool { return false }, // The validator rejects it.
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
				t.Errorf("remaining input mismatch: got %q, want %q", p.input.asStr(), tc.expectedInput)
			}
		})
	}
}

// testReadURLCodepointOrEcharError is a helper for TestIriParser_readURLCodepointOrEchar error cases.
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
			name:        "Percent encoding error",
			inputRune:   '%',
			parserInput: "2G",
			validFunc:   pathCharValidator,
			expectedErr: "Invalid IRI percent encoding '%2G'",
		},
		{
			name:        "Invalid character - newline",
			inputRune:   '\n',
			validFunc:   pathCharValidator,
			expectedErr: fmt.Sprintf("Invalid IRI character '%c'", '\n'),
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
				t.Errorf("expected error '%s', got '%s'", tc.expectedErr, err.Error())
			}
		})
	}
}

// TestIriParser_readURLCodepointOrEchar tests the dispatcher for reading either
// a normal character or a percent-encoded sequence.
// RFC Reference: This function implements logic from multiple sections.
// RFC 3987, Section 3.1 (note) allows lenient parsing of certain ASCII chars.
// RFC 3986, Section 2.1 is for percent-encoding.
// RFC 3987, Section 2.2 defines the character sets for IRIs.
func TestIriParser_readURLCodepointOrEchar(t *testing.T) {
	t.Run("success", testReadURLCodepointOrEcharSuccess)
	t.Run("error", testReadURLCodepointOrEcharError)
}
