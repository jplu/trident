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
	"strings"
	"testing"
)

// TestParseURIToRef tests the conversion from a URI string to an IRI Ref.
func TestParseURIToRef(t *testing.T) {
	testCases := []struct {
		name     string
		uri      string
		expected string
		hasError bool
	}{
		{
			name:     "Valid UTF-8 sequence",
			uri:      "http://example.org/D%C3%BCrst",
			expected: "http://example.org/Dürst",
			hasError: false,
		},
		{
			name:     "Valid URI with non-UTF8 percent encoding",
			uri:      "http://example.org/%FCrst",
			expected: "http://example.org/%FCrst",
			hasError: false,
		},
		{
			name:     "Forbidden Bidi character",
			uri:      "http://example.com/%E2%80%AE",
			expected: "http://example.com/%E2%80%AE",
			hasError: false,
		},
		{
			name:     "Malformed URI with incomplete percent encoding",
			uri:      "http://example.com/%C",
			expected: "",
			hasError: true,
		},
		{
			name:     "Malformed URI with invalid percent encoding",
			uri:      "http://example.com/foo%GGbar",
			expected: "",
			hasError: true,
		},
		{
			name:     "Mixed valid and invalid sequences",
			uri:      "/a%C3%A9b%E9c/",
			expected: "/aéb%E9c/",
			hasError: false,
		},
		{
			name:     "Invalid decoded IRI",
			uri:      "a%3A/b",
			expected: "a%3A/b",
			hasError: false,
		},
		{
			name:     "Control character U+0080 in path remains percent-encoded",
			uri:      "http://example.org/%C2%80",
			expected: "http://example.org/%C2%80",
			hasError: false,
		},
		{
			name:     "Private-use character in path remains percent-encoded",
			uri:      "http://example.org/%EE%80%80",
			expected: "http://example.org/%EE%80%80",
			hasError: false,
		},
		{
			name:     "Private-use character in query is successfully decoded",
			uri:      "http://example.org/?q=%EE%80%80",
			expected: "http://example.org/?q=\uE000",
			hasError: false,
		},
		{
			name:     "Full URI with all components decoded",
			uri:      "http://user%C3%A9:pass%C3%A9@host%C3%A9:8080/path%C3%A9?query%C3%A9#frag%C3%A9",
			expected: "http://useré:passé@hosté:8080/pathé?queryé#fragé",
			hasError: false,
		},
		{
			name:     "URI with scheme but no authority, with query and fragment",
			uri:      "mailto:foo%C3%A9?query%C3%A9#frag%C3%A9",
			expected: "mailto:fooé?queryé#fragé",
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := ParseURIToRef(tc.uri)
			if tc.hasError {
				if err == nil {
					t.Fatal("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if ref.String() != tc.expected {
					t.Errorf("Expected converted IRI '%s', got '%s'", tc.expected, ref.String())
				}
			}
		})
	}
}

// TestRefToURI tests the conversion from an IRI Ref to a URI string.
func TestRefToURI(t *testing.T) {
	testCases := []struct {
		name     string
		iri      string
		expected string
	}{
		{
			"Simple ASCII IRI",
			"http://example.com/a/b",
			"http://example.com/a/b",
		},
		{
			"Non-ASCII path",
			"http://example.com/résumé",
			"http://example.com/r%C3%A9sum%C3%A9",
		},
		{
			"Non-ASCII query",
			"http://example.com/?p=résumé",
			"http://example.com/?p=r%C3%A9sum%C3%A9",
		},
		{
			"Non-ASCII fragment",
			"http://example.com/#résumé",
			"http://example.com/#r%C3%A9sum%C3%A9",
		},
		{
			"Non-ASCII userinfo",
			"ftp://résumé@example.com/",
			"ftp://r%C3%A9sum%C3%A9@example.com/",
		},
		{
			"IDNA host",
			"http://résumé.example.org/",
			"http://xn--rsum-bpad.example.org/",
		},
		{
			"Full IRI with all parts",
			"http://user:p@résumé.com:8080/p?q=v#f",
			"http://user:p@xn--rsum-bpad.com:8080/p?q=v#f",
		},
		{
			"IDNA handling of leading hyphen non-ASCII host",
			"http://-résumé.com/",
			"http://xn---rsum-csad.com/",
		},
		{
			"IDNA handling of long label",
			"http://" + strings.Repeat("a", 63) + ".com/",
			"http://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.com/",
		},
		{
			"IDNA failure fallback with overly long label",
			"http://" + strings.Repeat("a", 64) + ".com/",
			"http://" + strings.Repeat(
				"a",
				64,
			) + ".com/",
		},
		{
			"IPv6 literal in host",
			"http://[2001:db8::1]:8080/",
			"http://[2001:db8::1]:8080/",
		},
		{
			"IDNA failure fallback with percent-encoded space in host",
			"http://a%20b.com/",
			"http://a%20b.com/",
		},
		{
			"RFC 3987 Step 1c preservation of unnormalized characters",
			"http://example.com/e\u0301",
			"http://example.com/e%CC%81",
		},
		{
			"Relative IRI with non-ASCII path, query and fragment",
			"résumé?q=résumé#résumé",
			"r%C3%A9sum%C3%A9?q=r%C3%A9sum%C3%A9#r%C3%A9sum%C3%A9",
		},
		{
			"IDNA ToASCII error: truncated punycode digit sequence in xn-- label",
			"http://xn--b.com/",
			"http://xn--b.com/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref := mustParseRef(t, tc.iri)
			uri := ref.ToURI()
			if uri != tc.expected {
				t.Errorf("Expected URI '%s', got '%s'", tc.expected, uri)
			}
		})
	}
}

// TestParseURIToRefBidiAndUTF8 validation tests the conversion of URIs to IRIs,
// ensuring invalid UTF-8 and forbidden bidirectional formatting characters are kept percent-encoded.
func TestParseURIToRefBidiAndUTF8(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid UTF-8 with allowed ASCII",
			input:    "http://example.org/hello-world",
			expected: "http://example.org/hello-world",
		},
		{
			name:     "Valid UTF-8 with allowed non-ASCII",
			input:    "http://example.org/r%C3%A9sum%C3%A9",
			expected: "http://example.org/résumé",
		},
		{
			name:     "Invalid UTF-8 sequence remains percent-encoded",
			input:    "http://example.org/%C3%28",
			expected: "http://example.org/%C3%28",
		},
		{
			name:     "Forbidden bidi character LRM (U+200E) remains percent-encoded",
			input:    "http://example.org/%E2%80%8E",
			expected: "http://example.org/%E2%80%8E",
		},
		{
			name:     "Forbidden bidi character RLM (U+200F) remains percent-encoded",
			input:    "http://example.org/%E2%80%8F",
			expected: "http://example.org/%E2%80%8F",
		},
		{
			name:     "Forbidden bidi character LRE (U+202A) remains percent-encoded",
			input:    "http://example.org/%E2%80%AA",
			expected: "http://example.org/%E2%80%AA",
		},
		{
			name:     "Empty URI path remains empty",
			input:    "http://example.org",
			expected: "http://example.org",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := ParseURIToRef(tc.input)
			if err != nil {
				t.Fatalf("ParseURIToRef(%q) returned unexpected error: %v", tc.input, err)
			}
			result := ref.String()
			if result != tc.expected {
				t.Errorf("ParseURIToRef(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestDecodeComponentDirect tests decodeComponent directly with various inputs to cover all branches.
func TestDecodeComponentDirect(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		isValid  func(rune) bool
		expected string
	}{
		{
			name:     "s[i] != '%' is true",
			s:        "abc",
			isValid:  func(_ rune) bool { return true },
			expected: "abc",
		},
		{
			name:     "s[i] == '%' but i+2 >= len(s)",
			s:        "ab%",
			isValid:  func(_ rune) bool { return true },
			expected: "ab%",
		},
		{
			name:     "s[i] == '%' but !isASCIIHexDigit",
			s:        "ab%G1",
			isValid:  func(_ rune) bool { return true },
			expected: "ab%G1",
		},
		{
			name:     "b[0] <= maxASCII, isUnreserved is true",
			s:        "%41",
			isValid:  func(_ rune) bool { return true },
			expected: "A",
		},
		{
			name:     "b[0] <= maxASCII, isUnreserved is false",
			s:        "%2F",
			isValid:  func(_ rune) bool { return true },
			expected: "%2F",
		},
		{
			name:     "b[0] > maxASCII (non-ASCII block)",
			s:        "%C3%BC",
			isValid:  func(_ rune) bool { return true },
			expected: "ü",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := decodeComponent(tc.s, tc.isValid)
			if res != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, res)
			}
		})
	}
}

// TestDecodeNonASCIIBlockDirect tests decodeNonASCIIBlock directly with various inputs to cover all branches.
func TestDecodeNonASCIIBlockDirect(t *testing.T) {
	testCases := []struct {
		name         string
		s            string
		start        int
		isValid      func(rune) bool
		expectedRes  string
		expectedNext int
	}{
		{
			name:         "i < len(s) is false (start >= len(s))",
			s:            "abc",
			start:        3,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "",
			expectedNext: 3,
		},
		{
			name:         "s[i] == '%' is false",
			s:            "abc",
			start:        0,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "",
			expectedNext: 0,
		},
		{
			name:         "i+2 >= len(s) is true",
			s:            "%C",
			start:        0,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "",
			expectedNext: 0,
		},
		{
			name:         "!isASCIIHexDigit is true",
			s:            "%CG",
			start:        0,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "",
			expectedNext: 0,
		},
		{
			name:         "octet[0] <= maxASCII is true",
			s:            "%41",
			start:        0,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "",
			expectedNext: 0,
		},
		{
			name:         "utf8.Valid(decodedBytes) is false",
			s:            "%C3",
			start:        0,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "%C3",
			expectedNext: 3,
		},
		{
			name:         "isForbiddenBidiFormatting(r) is true",
			s:            "%E2%80%8E",
			start:        0,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "%E2%80%8E",
			expectedNext: 9,
		},
		{
			name:         "!isValidRune(r) is true",
			s:            "%C3%BC",
			start:        0,
			isValid:      func(_ rune) bool { return false },
			expectedRes:  "%C3%BC",
			expectedNext: 6,
		},
		{
			name:         "allValid is true",
			s:            "%C3%BC",
			start:        0,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "ü",
			expectedNext: 6,
		},
		{
			name:         "Subsequent iteration triggers i+2 >= len(s)",
			s:            "%C3%BC%C",
			start:        0,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "ü",
			expectedNext: 6,
		},
		{
			name:         "Subsequent iteration triggers !isASCIIHexDigit",
			s:            "%C3%BC%CG",
			start:        0,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "ü",
			expectedNext: 6,
		},
		{
			name:         "Subsequent iteration triggers octet[0] <= maxASCII",
			s:            "%C3%BC%41",
			start:        0,
			isValid:      func(_ rune) bool { return true },
			expectedRes:  "ü",
			expectedNext: 6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, next := decodeNonASCIIBlock(tc.s, tc.start, tc.isValid)
			if res != tc.expectedRes || next != tc.expectedNext {
				t.Errorf("Expected %q and %d, got %q and %d", tc.expectedRes, tc.expectedNext, res, next)
			}
		})
	}
}
