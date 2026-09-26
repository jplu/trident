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
	"reflect"
	"testing"
)

// TestIsPathChar tests the isPathChar function.
func TestIsPathChar(t *testing.T) {
	testCases := []struct {
		char     rune
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'5', true},
		{'~', true},
		{'_', true},
		{'é', true},
		{'€', true},
		{'!', true},
		{'$', true},
		{'*', true},
		{';', true},
		{':', true},
		{'@', true},
		{'/', true},
		{'?', false},
		{'#', false},
		{'[', false},
		{']', false},
	}

	for _, tc := range testCases {
		if got := isPathChar(tc.char); got != tc.expected {
			t.Errorf("isPathChar('%c') = %v, want %v", tc.char, got, tc.expected)
		}
	}
}

// TestIsQueryChar tests the isQueryChar function.
func TestIsQueryChar(t *testing.T) {
	testCases := []struct {
		char     rune
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{':', true},
		{'@', true},
		{'\uE000', true},
		{'\uF8FF', true},
		{'/', true},
		{'?', true},
		{'#', false},
		{'[', false},
		{']', false},
	}

	for _, tc := range testCases {
		if got := isQueryChar(tc.char); got != tc.expected {
			t.Errorf("isQueryChar('%c') = %v, want %v", tc.char, got, tc.expected)
		}
	}
}

// TestParseFragment tests parsing of the fragment component.
func TestParseFragment(t *testing.T) {
	rtlComponent := "\u05D0\u05D1\u05D2"

	testCases := []struct {
		name        string
		input       string
		unchecked   bool
		expected    string
		expectedErr error
	}{
		{name: "Empty", input: "", unchecked: false, expected: "", expectedErr: nil},
		{
			name:        "Valid ASCII",
			input:       "anchor",
			unchecked:   false,
			expected:    "anchor",
			expectedErr: nil,
		},
		{
			name:        "With Unreserved ucschar",
			input:       "ancre-é",
			unchecked:   false,
			expected:    "ancre-é",
			expectedErr: nil,
		},
		{
			name:        "With Allowed Delims",
			input:       "a/b?c",
			unchecked:   false,
			expected:    "a/b?c",
			expectedErr: nil,
		},
		{
			name:        "With Percent Encoding",
			input:       "a%20b",
			unchecked:   false,
			expected:    "a%20b",
			expectedErr: nil,
		},
		{
			name:        "With RTL Bidi (Allowed)",
			input:       "a" + rtlComponent,
			unchecked:   false,
			expected:    "a" + rtlComponent,
			expectedErr: nil,
		},
		{
			name:        "With Invalid Percent Encoding",
			input:       "%GG",
			unchecked:   false,
			expected:    "",
			expectedErr: &kindError{message: "Invalid percent-encoding sequence"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, tc.unchecked)
			err := p.parseFragment()

			assertError(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				if got := p.output.string(); got != tc.expected {
					t.Errorf("parseFragment() output = %q, want %q", got, tc.expected)
				}
			}
		})
	}
}

// TestHandleQueryEnd tests the helper for terminating query parsing.
func TestHandleQueryEnd(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		isFragment  bool
		queryPart   string
		expectedOut string
		wantErr     bool
	}{
		{
			name:        "End of Input",
			input:       "",
			isFragment:  false,
			queryPart:   "q=1",
			expectedOut: "q=1",
			wantErr:     false,
		},
		{
			name:        "Fragment Follows",
			input:       "#frag",
			isFragment:  true,
			queryPart:   "q=1",
			expectedOut: "q=1#frag",
			wantErr:     false,
		},
		{
			name:        "RTL Query (Allowed)",
			input:       "",
			isFragment:  false,
			queryPart:   "a\u05D0",
			expectedOut: "a\u05D0",
			wantErr:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			p.output.writeString(tc.queryPart)

			err := p.handleQueryEnd(tc.isFragment)

			if (err != nil) != tc.wantErr {
				t.Fatalf("handleQueryEnd() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got := p.output.string(); got != tc.expectedOut {
				t.Errorf("handleQueryEnd() output = %q, want %q", got, tc.expectedOut)
			}
			if p.outputPositions.QueryEnd != len(tc.queryPart) {
				t.Errorf("handleQueryEnd() did not set QueryEnd correctly. Got %d, want %d",
					p.outputPositions.QueryEnd, len(tc.queryPart))
			}
		})
	}
}

// TestParseQuery tests parsing of the query component.
func TestParseQuery(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{name: "Empty", input: "", expected: "", expectedErr: nil},
		{name: "Valid ASCII", input: "a=1&b=2", expected: "a=1&b=2", expectedErr: nil},
		{
			name:        "With Unreserved ucschar",
			input:       "search=résumé",
			expected:    "search=résumé",
			expectedErr: nil,
		},
		{name: "With Private ucschar", input: "a=\uE000", expected: "a=\uE000", expectedErr: nil},
		{name: "With Allowed Delims", input: "a/b:c@d?e", expected: "a/b:c@d?e", expectedErr: nil},
		{name: "Terminated by Fragment", input: "a=1#frag", expected: "a=1#frag", expectedErr: nil},
		{name: "With Invalid Char", input: "a=<b>", expected: "a=%3Cb%3E", expectedErr: nil},
		{name: "RTL Query (Allowed)", input: "a\u05D0", expected: "a\u05D0", expectedErr: nil},
		{
			name:        "With Invalid Non-Lax Char",
			input:       "q=[",
			expected:    "q=",
			expectedErr: &kindError{message: "Invalid IRI character"},
		},
		{
			name:        "With Invalid Percent Encoding",
			input:       "q=%GG",
			expected:    "q=",
			expectedErr: &kindError{message: "Invalid percent-encoding sequence"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			err := p.parseQuery()

			assertError(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				if got := p.output.string(); got != tc.expected {
					t.Errorf("parseQuery() output = %q, want %q", got, tc.expected)
				}
			}
		})
	}
}

// TestHandlePathTerminator tests handlePathTerminator for '?' and '#'.
func TestHandlePathTerminator(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		pathPart        string
		expectedHandled bool
		expectedOut     string
		wantErr         bool
	}{
		{
			name:            "Not a Terminator",
			input:           "/b",
			pathPart:        "/a",
			expectedHandled: false,
			expectedOut:     "/a",
			wantErr:         false,
		},
		{
			name:            "Query Terminator",
			input:           "?q=1",
			pathPart:        "/a",
			expectedHandled: true,
			expectedOut:     "/a?q=1",
			wantErr:         false,
		},
		{
			name:            "Fragment Terminator",
			input:           "#frag",
			pathPart:        "/a",
			expectedHandled: true,
			expectedOut:     "/a#frag",
			wantErr:         false,
		},
		{
			name:            "RTL Path (Allowed)",
			input:           "?q=1",
			pathPart:        "/a\u05D0",
			expectedHandled: true,
			expectedOut:     "/a\u05D0?q=1",
			wantErr:         false,
		},
		{
			name:            "Query Terminator with Error",
			input:           "?q=[",
			pathPart:        "/a",
			expectedHandled: true,
			expectedOut:     "",
			wantErr:         true,
		},
		{
			name:            "Fragment Terminator with Error",
			input:           "#%GG",
			pathPart:        "/a",
			expectedHandled: true,
			expectedOut:     "",
			wantErr:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			p.output.writeString(tc.pathPart)
			peekedChar, _ := p.input.peek()

			handled, err := p.handlePathTerminator(peekedChar)

			if (err != nil) != tc.wantErr {
				t.Fatalf("Error mismatch: got err %v, wantErr %v", err, tc.wantErr)
			}
			if handled != tc.expectedHandled {
				t.Fatalf("Handled mismatch: got %v, want %v", handled, tc.expectedHandled)
			}
			if !tc.wantErr {
				if got := p.output.string(); got != tc.expectedOut {
					t.Errorf("Output mismatch: got %q, want %q", got, tc.expectedOut)
				}
				if handled && p.outputPositions.PathEnd != len(tc.pathPart) {
					t.Errorf("PathEnd was not set correctly: got %d, want %d",
						p.outputPositions.PathEnd, len(tc.pathPart))
				}
			}
		})
	}
}

// TestParsePath tests parsing of the path component.
func TestParsePath(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		hasAuthority bool
		expected     string
		expectedErr  error
	}{
		{name: "Empty", input: "", hasAuthority: false, expected: "", expectedErr: nil},
		{
			name:         "Simple Segments",
			input:        "/a/b/c",
			hasAuthority: true,
			expected:     "/a/b/c",
			expectedErr:  nil,
		},
		{
			name:         "Ends with Query",
			input:        "/a/b?q=1",
			hasAuthority: true,
			expected:     "/a/b?q=1",
			expectedErr:  nil,
		},
		{
			name:         "Ends with Fragment",
			input:        "/a/b#frag",
			hasAuthority: true,
			expected:     "/a/b#frag",
			expectedErr:  nil,
		},
		{
			name:         "With ucschar",
			input:        "/résumé",
			hasAuthority: true,
			expected:     "/résumé",
			expectedErr:  nil,
		},
		{
			name:         "Path with Colon",
			input:        "/a:b/c",
			hasAuthority: true,
			expected:     "/a:b/c",
			expectedErr:  nil,
		},
		{
			name:         "Double Slash without Authority",
			input:        "//a/b",
			hasAuthority: false,
			expected:     "",
			expectedErr:  errPathStartingWithSlashes,
		},
		{
			name:         "Double Slash with Authority",
			input:        "/a/b",
			hasAuthority: true,
			expected:     "/a/b",
			expectedErr:  nil,
		},
		{
			name:         "Invalid Char",
			input:        "/a<b>",
			hasAuthority: true,
			expected:     "/a%3Cb%3E",
			expectedErr:  nil,
		},
		{
			name:         "Invalid Non-Lax Char in Path",
			input:        "/a[b",
			hasAuthority: true,
			expected:     "",
			expectedErr:  &kindError{message: "Invalid IRI character"},
		},
		{
			name:         "Invalid Percent Encoding in Path",
			input:        "/a%GG",
			hasAuthority: true,
			expected:     "",
			expectedErr:  &kindError{message: "Invalid percent-encoding sequence"},
		},
		{
			name:         "Path Terminated with Invalid Query",
			input:        "/a?q=[",
			hasAuthority: true,
			expected:     "",
			expectedErr:  &kindError{message: "Invalid IRI character"},
		},
		{
			name:         "Path Terminated with Invalid Fragment",
			input:        "/a#%GG",
			hasAuthority: true,
			expected:     "",
			expectedErr:  &kindError{message: "Invalid percent-encoding sequence"},
		},
		{
			name:         "RTL Path (Allowed)",
			input:        "/a\u05D0",
			hasAuthority: true,
			expected:     "/a\u05D0",
			expectedErr:  nil,
		},
		{
			name:         "Segment RTL Path (Allowed)",
			input:        "/seg1/a\u05D0/seg2",
			hasAuthority: true,
			expected:     "/seg1/a\u05D0/seg2",
			expectedErr:  nil,
		},
		{
			name:         "Double Slash In Middle Absolute without Authority",
			input:        "/a//b",
			hasAuthority: false,
			expected:     "/a//b",
			expectedErr:  nil,
		},
		{
			name:         "Double Slash In Middle Relative without Authority",
			input:        "a//b",
			hasAuthority: false,
			expected:     "a//b",
			expectedErr:  nil,
		},
		{
			name:         "Consecutive Slashes in Data URI-like Path",
			input:        "image/gif;base64,R0l...///...yw==",
			hasAuthority: false,
			expected:     "image/gif;base64,R0l...///...yw==",
			expectedErr:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			if tc.hasAuthority {
				p.outputPositions.SchemeEnd = 1
				p.outputPositions.AuthorityEnd = 2
			}

			err := p.parsePath()

			assertError(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				if got := p.output.string(); got != tc.expected {
					t.Errorf("parsePath() output = %q, want %q", got, tc.expected)
				}
			}
		})
	}
}

// TestParsePathNoScheme tests parsing a relative-path that cannot start with a scheme.
func TestParsePathNoScheme(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{name: "Valid", input: "a/b/c", expected: "a/b/c", expectedErr: nil},
		{name: "Valid with @", input: "user@host", expected: "user@host", expectedErr: nil},
		{
			name:     "Invalid Colon in First Segment",
			input:    "a:b/c",
			expected: "",
			expectedErr: &kindError{
				message: "Invalid IRI character in first path segment",
			},
		},
		{
			name:        "Invalid Non-Lax Char in First Segment",
			input:       "a[b",
			expected:    "",
			expectedErr: &kindError{message: "Invalid IRI character"},
		},
		{
			name:        "Invalid Percent Encoding in First Segment",
			input:       "a%GG",
			expected:    "",
			expectedErr: &kindError{message: "Invalid percent-encoding sequence"},
		},
		{
			name:        "Valid Colon in Second Segment",
			input:       "a/b:c",
			expected:    "a/b:c",
			expectedErr: nil,
		},
		{name: "Ends with Query", input: "a/b?q", expected: "a/b?q", expectedErr: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			err := p.parsePathNoScheme()

			assertError(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				if got := p.output.string(); got != tc.expected {
					t.Errorf("parsePathNoScheme() output = %q, want %q", got, tc.expected)
				}
			}
		})
	}
}

// TestParsePathStart tests the path/query/fragment dispatcher.
func TestParsePathStart(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedOut string
		wantErr     bool
	}{
		{name: "EOF", input: "", expectedOut: "", wantErr: false},
		{name: "Starts with Query", input: "?q=1", expectedOut: "?q=1", wantErr: false},
		{name: "Starts with Fragment", input: "#frag", expectedOut: "#frag", wantErr: false},
		{name: "Starts with Slash", input: "/a/b", expectedOut: "/a/b", wantErr: false},
		{name: "Starts with pchar", input: "a/b", expectedOut: "a/b", wantErr: false},
		{name: "Starts with Invalid Lax Char", input: "<", expectedOut: "%3C", wantErr: false},
		{name: "Starts with Invalid Non-Lax Char", input: "[foo", expectedOut: "", wantErr: true},
		{name: "Starts with Slash Leads to Path Error", input: "/[foo", expectedOut: "", wantErr: true},
		{name: "Starts with pchar Leads to Path Error", input: "a[b", expectedOut: "", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, ok := newParserInput(tc.input).peek()
			p := setupTestParser(tc.input, false)

			err := p.parsePathStart(r, ok)

			if (err != nil) != tc.wantErr {
				t.Fatalf("parsePathStart() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got := p.output.string(); got != tc.expectedOut {
				t.Errorf("parsePathStart() output = %q, want %q", got, tc.expectedOut)
			}
			if !ok {
				if p.outputPositions.PathEnd != 0 || p.outputPositions.QueryEnd != 0 {
					t.Error("parsePathStart() on EOF did not set positions correctly")
				}
			}
		})
	}
}

// TestParsePathOrAuthority tests the dispatcher after "scheme:/".
func TestParsePathOrAuthority(t *testing.T) {
	testCases := []struct {
		name                string
		input               string
		expectedFinalString string
		expectedPos         Positions
		expectedErr         error
	}{
		{
			name:                "Authority and Path",
			input:               "/host/path",
			expectedFinalString: "scheme://host/path",
			expectedPos: Positions{
				SchemeEnd:    7,
				AuthorityEnd: 13,
				PathEnd:      18,
				QueryEnd:     18,
			},
			expectedErr: nil,
		},
		{
			name:                "Authority and Query",
			input:               "/host?query",
			expectedFinalString: "scheme://host?query",
			expectedPos: Positions{
				SchemeEnd:    7,
				AuthorityEnd: 13,
				PathEnd:      13,
				QueryEnd:     19,
			},
			expectedErr: nil,
		},
		{
			name:                "Authority and Fragment",
			input:               "/host#frag",
			expectedFinalString: "scheme://host#frag",
			expectedPos: Positions{
				SchemeEnd:    7,
				AuthorityEnd: 13,
				PathEnd:      13,
				QueryEnd:     13,
			},
			expectedErr: nil,
		},
		{
			name:                "Only Path",
			input:               "path",
			expectedFinalString: "scheme:/path",
			expectedPos: Positions{
				SchemeEnd:    7,
				AuthorityEnd: 7,
				PathEnd:      12,
				QueryEnd:     12,
			},
			expectedErr: nil,
		},
		{
			name:                "Empty Path",
			input:               "",
			expectedFinalString: "scheme:/",
			expectedPos:         Positions{SchemeEnd: 7, AuthorityEnd: 7, PathEnd: 8, QueryEnd: 8},
			expectedErr:         nil,
		},
		{
			name:                "Invalid Authority",
			input:               "/[invalid/path",
			expectedFinalString: "",
			expectedPos:         Positions{},
			expectedErr:         &kindError{message: "Invalid host IP: unterminated IP literal"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initialBuffer := "scheme:"
			p := setupTestParser(tc.input, false)
			p.output.writeString(initialBuffer)
			p.outputPositions.SchemeEnd = len(initialBuffer)
			p.output.writeRune('/')

			err := p.parsePathOrAuthority()

			assertError(t, err, tc.expectedErr)
			if tc.expectedErr != nil {
				return
			}
			if got := p.output.string(); got != tc.expectedFinalString {
				t.Errorf("parsePathOrAuthority() output = %q, want %q", got, tc.expectedFinalString)
			}
			if !reflect.DeepEqual(p.outputPositions, tc.expectedPos) {
				t.Errorf(
					"parsePathOrAuthority() positions = %+v, want %+v",
					p.outputPositions,
					tc.expectedPos,
				)
			}
		})
	}
}
