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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package iri

import (
	"strings"
	"testing"
)

// TestParseURIToRef tests the conversion from a URI string to an IRI Ref.
// RFC Reference: RFC 3987, Section 3.2.
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
			expected: "a:/b",
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

// TestRef_ToURI tests the conversion from an IRI Ref to a URI string.
// Based on RFC 3987, Section 3.1.
func TestRef_ToURI(t *testing.T) {
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
			"IDNA failure fallback with percent-encoded space in host",
			"http://a%20b.com/",
			"http://a%20b.com/",
		},
		{
			"NFC normalization before encoding",
			"http://example.com/e\u0301",
			"http://example.com/%C3%A9",
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
