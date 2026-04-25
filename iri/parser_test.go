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
	"reflect"
	"testing"
)

// TestParseScheme tests parsing of the scheme component.
// RFC Reference: RFC 3986, Section 3.1.
func TestParseScheme(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedOut string
		expectedErr error
	}{
		{name: "Valid scheme", input: "http:", expectedOut: "http:", expectedErr: nil},
		{name: "Valid with path", input: "http:/path", expectedOut: "http:/path", expectedErr: nil},
		{name: "Valid with authority", input: "http://host", expectedOut: "http://host", expectedErr: nil},
		{name: "Fallback on EOF", input: "http", expectedOut: "http", expectedErr: nil},
		{
			name:        "Fallback on invalid char (successful parse)",
			input:       "ht^tp",
			expectedOut: "ht%5Etp",
			expectedErr: nil,
		},
		{
			name:        "Fallback on invalid char (invalid relative ref)",
			input:       "ht^tp:",
			expectedOut: "ht%5Etp",
			expectedErr: &kindError{message: "Invalid IRI character in first path segment"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			err := p.parseScheme()

			assertError(t, err, tc.expectedErr)
			if got := p.output.string(); got != tc.expectedOut {
				t.Errorf("parseScheme() output = %q, want %q", got, tc.expectedOut)
			}
		})
	}
}

// TestParseSchemeStart tests the initial state of the parser.
func TestParseSchemeStart(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedOut string
		expectedErr error
	}{
		{
			name:        "Network-Path Ref",
			input:       "//example.com/path",
			expectedOut: "//example.com/path",
			expectedErr: nil,
		},
		{
			name:        "Network-Path Ref with Query",
			input:       "//example.com?q",
			expectedOut: "//example.com?q",
			expectedErr: nil,
		},
		{
			name:        "Network-Path Ref Invalid Authority",
			input:       "//[invalid/path",
			expectedOut: "//",
			expectedErr: &kindError{message: "Invalid host IP: unterminated IP literal"},
		},
		{name: "Empty Input", input: "", expectedOut: "", expectedErr: nil},
		{name: "Starts with Colon", input: ":foo", expectedOut: "", expectedErr: errNoScheme},
		{name: "Absolute IRI", input: "http://example.com", expectedOut: "http://example.com", expectedErr: nil},
		{name: "Absolute Path", input: "/path", expectedOut: "/path", expectedErr: nil},
		{name: "Relative Path", input: "path", expectedOut: "path", expectedErr: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			err := p.parseSchemeStart()

			assertError(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				if got := p.output.string(); got != tc.expectedOut {
					t.Errorf("parseSchemeStart() output = %q, want %q", got, tc.expectedOut)
				}
			}
		})
	}
}

// TestRun tests the main entry point for the parser.
func TestRun(t *testing.T) {
	baseIRI := &base{
		IRI: "http://a/b/c",
		Pos: Positions{SchemeEnd: 5, AuthorityEnd: 10, PathEnd: 14, QueryEnd: 14},
	}

	testCases := []struct {
		name        string
		input       string
		base        *base
		unchecked   bool
		expectedPos Positions
		expectedErr error
	}{
		{
			name:        "Valid Absolute IRI",
			input:       "https://example.com/p?q#f",
			base:        nil,
			unchecked:   false,
			expectedPos: Positions{SchemeEnd: 6, AuthorityEnd: 19, PathEnd: 21, QueryEnd: 23},
			expectedErr: nil,
		},
		{
			name:        "Valid with Base (base ignored)",
			input:       "ftp://host/path",
			base:        baseIRI,
			unchecked:   false,
			expectedPos: Positions{SchemeEnd: 4, AuthorityEnd: 10, PathEnd: 15, QueryEnd: 15},
			expectedErr: nil,
		},
		{
			name:        "Network Path Ref",
			input:       "//host/path",
			base:        nil,
			unchecked:   false,
			expectedPos: Positions{SchemeEnd: 0, AuthorityEnd: 6, PathEnd: 11, QueryEnd: 11},
			expectedErr: nil,
		},
		{
			name:        "Unchecked Mode",
			input:       "a[b",
			base:        nil,
			unchecked:   true,
			expectedPos: Positions{SchemeEnd: 0, AuthorityEnd: 0, PathEnd: 3, QueryEnd: 3},
			expectedErr: nil,
		},
		{
			name:        "Parse Error Invalid Char",
			input:       "a[b",
			base:        nil,
			unchecked:   false,
			expectedPos: Positions{},
			expectedErr: &kindError{message: "Invalid IRI character"},
		},
		{
			name:        "No Scheme Error",
			input:       ":foo",
			base:        nil,
			unchecked:   false,
			expectedPos: Positions{},
			expectedErr: errNoScheme,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &voidOutputBuffer{}
			pos, err := run(tc.input, tc.base, tc.unchecked, output)

			assertError(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				if !reflect.DeepEqual(pos, tc.expectedPos) {
					t.Errorf("run() positions = %+v, want %+v", pos, tc.expectedPos)
				}
			}
		})
	}
}
