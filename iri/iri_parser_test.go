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
	"errors"
	"reflect"
	"strings"
	"testing"
)

// setupTestParser is a helper to create and configure a parser instance for testing.
// It uses a stringOutputBuffer to allow inspection of the parsed output.
func setupTestParser(input string, unchecked bool) *iriParser {
	return &iriParser{
		iri:       input,
		base:      &iriParserBase{hasBase: false},
		input:     newParserInput(input),
		output:    &stringOutputBuffer{builder: &strings.Builder{}},
		unchecked: unchecked,
	}
}

// assertError checks if the received error matches the expected error.
// It handles nil errors, sentinel errors, and custom *kindError types by message prefix.
func assertError(t *testing.T, got, want error) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Errorf("unexpected error: %v", got)
		}
		return
	}

	if got == nil {
		t.Errorf("expected error %v, but got nil", want)
		return
	}

	// Check for sentinel errors first.
	if errors.Is(want, errNoScheme) || errors.Is(want, errPathStartingWithSlashes) {
		if !errors.Is(got, want) {
			t.Errorf("got error %v, want sentinel error %v", got, want)
		}
		return
	}

	// Check for kindError.
	var wantKE *kindError
	if errors.As(want, &wantKE) {
		var gotKE *kindError
		if !errors.As(got, &gotKE) {
			t.Errorf("got error of type %T, want *kindError", got)
			return
		}
		if !strings.HasPrefix(gotKE.message, wantKE.message) {
			t.Errorf("got error message %q, want prefix %q", gotKE.message, wantKE.message)
		}
		return // Handled this case.
	}

	// Fallback for any other error type.
	if !errors.Is(got, want) {
		t.Errorf("got error %v, want %v", got, want)
	}
}

// TestIsPathChar tests the isPathChar predicate against RFC 3987.
// RFC Reference: RFC 3987, Section 2.2, `ipchar`.
func TestIsPathChar(t *testing.T) {
	testCases := []struct {
		char     rune
		expected bool
	}{
		// iunreserved (subset)
		{'a', true},
		{'Z', true},
		{'5', true},
		{'~', true},
		{'_', true},
		// ucschar (subset)
		{'é', true}, // U+00E9
		{'€', true}, // U+20AC
		// sub-delims
		{'!', true},
		{'$', true},
		{'*', true},
		{';', true},
		// Allowed pchar-specific
		{':', true},
		{'@', true},
		// Allowed path-specific
		{'/', true},
		// Disallowed characters
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

// TestIsQueryChar tests the isQueryChar predicate against RFC 3987.
// RFC Reference: RFC 3987, Section 2.2, `iquery`.
func TestIsQueryChar(t *testing.T) {
	testCases := []struct {
		char     rune
		expected bool
	}{
		// ipchar characters
		{'a', true},
		{'Z', true},
		{':', true},
		{'@', true},
		// iprivate characters
		{'\uE000', true},
		{'\uF8FF', true},
		// query-specific characters
		{'/', true},
		{'?', true},
		// Disallowed characters
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

// TestValidateBidiPart tests the bidi validation logic for a component.
// RFC Reference: RFC 3987, Section 4.2.
func TestValidateBidiPart(t *testing.T) {
	ltr := "hello"
	rtl := "\u05D0\u05D1\u05D2" // Hebrew Alef, Bet, Gimel

	testCases := []struct {
		name      string
		component string
		unchecked bool
		wantErr   bool
	}{
		{name: "Unchecked", component: "anything", unchecked: true, wantErr: false},
		{name: "Empty Component", component: "", unchecked: false, wantErr: false},
		{name: "Valid LTR", component: ltr, unchecked: false, wantErr: false},
		{name: "Valid RTL", component: rtl, unchecked: false, wantErr: false},
		{name: "Invalid Mixed", component: ltr + rtl, unchecked: false, wantErr: true},
		{name: "Invalid RTL Start", component: "a" + rtl, unchecked: false, wantErr: true},
		{name: "Invalid RTL End", component: rtl + "a", unchecked: false, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.component, tc.unchecked)
			p.output.writeString(tc.component)
			err := p.validateBidiPart(0)

			if (err != nil) != tc.wantErr {
				t.Errorf("validateBidiPart() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestParseFragment tests parsing of the fragment component.
// RFC Reference: RFC 3987, Section 2.2, `ifragment`.
func TestParseFragment(t *testing.T) {
	rtlComponent := "\u05D0\u05D1\u05D2" // Hebrew Alef, Bet, Gimel

	testCases := []struct {
		name        string
		input       string
		unchecked   bool
		expected    string
		expectedErr error
	}{
		{name: "Empty", input: "", unchecked: false, expected: "", expectedErr: nil},
		{name: "Valid ASCII", input: "anchor", unchecked: false, expected: "anchor", expectedErr: nil},
		{name: "With Unreserved ucschar", input: "ancre-é", unchecked: false, expected: "ancre-é", expectedErr: nil},
		{name: "With Allowed Delims", input: "a/b?c", unchecked: false, expected: "a/b?c", expectedErr: nil},
		{name: "With Percent Encoding", input: "a%20b", unchecked: false, expected: "a%20b", expectedErr: nil},
		{
			name:        "With Invalid Bidi",
			input:       "a" + rtlComponent,
			unchecked:   false,
			expected:    "",
			expectedErr: &kindError{message: "Invalid IRI component"},
		},
		{
			name:        "With Invalid Bidi (Unchecked)",
			input:       "a" + rtlComponent,
			unchecked:   true,
			expected:    "a" + rtlComponent,
			expectedErr: nil,
		},
		{
			name:        "Invalid Percent Encoding",
			input:       "%GG",
			unchecked:   false,
			expected:    "",
			expectedErr: &kindError{message: "Invalid IRI percent encoding"},
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
		{name: "End of Input", input: "", isFragment: false, queryPart: "q=1", expectedOut: "q=1", wantErr: false},
		{
			name:        "Fragment Follows",
			input:       "#frag",
			isFragment:  true,
			queryPart:   "q=1",
			expectedOut: "q=1#frag",
			wantErr:     false,
		},
		{name: "Bidi Error", input: "", isFragment: false, queryPart: "a\u05D0", expectedOut: "a\u05D0", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			p.output.writeString(tc.queryPart) // Simulate that a query has been parsed
			queryStart := 0

			err := p.handleQueryEnd(tc.isFragment, queryStart)

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
// RFC Reference: RFC 3987, Section 2.2, `iquery`.
func TestParseQuery(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{name: "Empty", input: "", expected: "", expectedErr: nil},
		{name: "Valid ASCII", input: "a=1&b=2", expected: "a=1&b=2", expectedErr: nil},
		{name: "With Unreserved ucschar", input: "search=résumé", expected: "search=résumé", expectedErr: nil},
		{name: "With Private ucschar", input: "a=\uE000", expected: "a=\uE000", expectedErr: nil},
		{name: "With Allowed Delims", input: "a/b:c@d?e", expected: "a/b:c@d?e", expectedErr: nil},
		{name: "Terminated by Fragment", input: "a=1#frag", expected: "a=1#frag", expectedErr: nil},
		{name: "With Invalid Char", input: "a=<b>", expected: "a=%3Cb%3E", expectedErr: nil},
		{name: "Bidi Error", input: "a\u05D0", expected: "", expectedErr: &kindError{message: "Invalid IRI component"}},
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
			expectedErr: &kindError{message: "Invalid IRI percent encoding"},
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

// TestHandlePathTerminator has been refactored to reduce complexity.
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
			name:            "Bidi Error",
			input:           "?q=1",
			pathPart:        "/a\u05D0",
			expectedHandled: true,
			expectedOut:     "/a\u05D0",
			wantErr:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			p.output.writeString(tc.pathPart)
			peekedChar, _ := p.input.peek()

			handled, err := p.handlePathTerminator(peekedChar, 0)

			// Simple, flat checks
			if (err != nil) != tc.wantErr {
				t.Fatalf("Error mismatch: got err %v, wantErr %v", err, tc.wantErr)
			}
			if handled != tc.expectedHandled {
				t.Fatalf("Handled mismatch: got %v, want %v", handled, tc.expectedHandled)
			}

			// Success case assertions
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
// RFC Reference: RFC 3986, Section 3.3 and RFC 3987, Section 2.2.
func TestParsePath(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		hasAuthority bool
		expected     string
		expectedErr  error
	}{
		{name: "Empty", input: "", hasAuthority: false, expected: "", expectedErr: nil},
		{name: "Simple Segments", input: "/a/b/c", hasAuthority: true, expected: "/a/b/c", expectedErr: nil},
		{name: "Ends with Query", input: "/a/b?q=1", hasAuthority: true, expected: "/a/b?q=1", expectedErr: nil},
		{name: "Ends with Fragment", input: "/a/b#frag", hasAuthority: true, expected: "/a/b#frag", expectedErr: nil},
		{name: "With ucschar", input: "/résumé", hasAuthority: true, expected: "/résumé", expectedErr: nil},
		{name: "Path with Colon", input: "/a:b/c", hasAuthority: true, expected: "/a:b/c", expectedErr: nil},
		{
			name:         "Double Slash without Authority",
			input:        "//a/b",
			hasAuthority: false,
			expected:     "",
			expectedErr:  errPathStartingWithSlashes,
		},
		{name: "Double Slash with Authority", input: "/a/b", hasAuthority: true, expected: "/a/b", expectedErr: nil},
		{name: "Invalid Char", input: "/a<b>", hasAuthority: true, expected: "/a%3Cb%3E", expectedErr: nil},
		{
			name:         "Bidi Error",
			input:        "/a\u05D0",
			hasAuthority: true,
			expected:     "",
			expectedErr:  &kindError{message: "Invalid IRI component"},
		},
		{
			name:         "Segment Bidi Error",
			input:        "/seg1/a\u05D0/seg2",
			hasAuthority: true,
			expected:     "",
			expectedErr:  &kindError{message: "Invalid IRI component"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			if tc.hasAuthority {
				// Simulate presence of authority to allow leading slash tests
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
// RFC Reference: RFC 3986, Section 4.2 and 3.3 `path-noscheme`.
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
		{name: "Valid Colon in Second Segment", input: "a/b:c", expected: "a/b:c", expectedErr: nil},
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
			if !ok { // EOF case
				if p.outputPositions.PathEnd != 0 || p.outputPositions.QueryEnd != 0 {
					t.Error("parsePathStart() on EOF did not set positions correctly")
				}
			}
		})
	}
}

// TestParsePathOrAuthority tests the dispatcher after "scheme:/".
// RFC Reference: RFC 3986, Section 3, `hier-part`.
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
			expectedPos:         Positions{SchemeEnd: 7, AuthorityEnd: 13, PathEnd: 18, QueryEnd: 18},
			expectedErr:         nil,
		},
		{
			name:                "Authority and Query",
			input:               "/host?query",
			expectedFinalString: "scheme://host?query",
			expectedPos:         Positions{SchemeEnd: 7, AuthorityEnd: 13, PathEnd: 13, QueryEnd: 19},
			expectedErr:         nil,
		},
		{
			name:                "Authority and Fragment",
			input:               "/host#frag",
			expectedFinalString: "scheme://host#frag",
			expectedPos:         Positions{SchemeEnd: 7, AuthorityEnd: 13, PathEnd: 13, QueryEnd: 13},
			expectedErr:         nil,
		},
		{
			name:                "Only Path",
			input:               "path",
			expectedFinalString: "scheme:/path",
			expectedPos:         Positions{SchemeEnd: 7, AuthorityEnd: 7, PathEnd: 12, QueryEnd: 12},
			expectedErr:         nil,
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
				t.Errorf("parsePathOrAuthority() positions = %+v, want %+v", p.outputPositions, tc.expectedPos)
			}
		})
	}
}

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
