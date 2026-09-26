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
package iri

import (
	"reflect"
	"testing"
)

// TestParseScheme tests parsing of the scheme component.
func TestParseScheme(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedOut string
		expectedErr error
	}{
		{name: "Valid scheme", input: "http:", expectedOut: "http:", expectedErr: nil},
		{name: "Valid with path", input: "http:/path", expectedOut: "http:/path", expectedErr: nil},
		{
			name:        "Valid with authority",
			input:       "http://host",
			expectedOut: "http://host",
			expectedErr: nil,
		},
		{
			name:        "Valid without slash (opaque path)",
			input:       "urn:example",
			expectedOut: "urn:example",
			expectedErr: nil,
		},
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
		{
			name:        "Absolute IRI",
			input:       "http://example.com",
			expectedOut: "http://example.com",
			expectedErr: nil,
		},
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

	validBaseRef, errp := ParseRef("http://example.com/b/c")
	if errp != nil {
		t.Fatalf("ParseRef() failed to create base: %v", errp)
	}
	validBase := &base{
		IRI: validBaseRef.iri,
		Pos: validBaseRef.positions,
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
		{
			name:        "Absolute Reference with Dot Segments and Base",
			input:       "file:///C:/foo/../../../bar.html",
			base:        baseIRI,
			unchecked:   false,
			expectedPos: Positions{SchemeEnd: 5, AuthorityEnd: 7, PathEnd: 16, QueryEnd: 16},
			expectedErr: nil,
		},
		{
			name:        "Absolute Reference with First Segment Colon (g:h) and Base",
			input:       "g:h",
			base:        baseIRI,
			unchecked:   false,
			expectedPos: Positions{SchemeEnd: 2, AuthorityEnd: 2, PathEnd: 3, QueryEnd: 3},
			expectedErr: nil,
		},
		{
			name:        "Relative Path with Valid Base",
			input:       "d/e",
			base:        validBase,
			unchecked:   false,
			expectedPos: Positions{SchemeEnd: 5, AuthorityEnd: 18, PathEnd: 24, QueryEnd: 24},
			expectedErr: nil,
		},
		{
			name:        "Absolute Path with Valid Base",
			input:       "/d/e",
			base:        validBase,
			unchecked:   false,
			expectedPos: Positions{SchemeEnd: 5, AuthorityEnd: 18, PathEnd: 22, QueryEnd: 22},
			expectedErr: nil,
		},
		{
			name:        "Relative Reference with Base Validation Error",
			input:       "d[e",
			base:        validBase,
			unchecked:   false,
			expectedPos: Positions{},
			expectedErr: &kindError{message: "Invalid IRI character"},
		},
		{
			name:        "Relative Reference with Base Starting Digit Validation Error",
			input:       "1[2",
			base:        validBase,
			unchecked:   false,
			expectedPos: Positions{},
			expectedErr: &kindError{message: "Invalid IRI character"},
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

// TestValidateRelativeRef tests validation of relative reference syntax.
func TestValidateRelativeRef(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			name:        "Valid relative path",
			input:       "path/to/resource",
			expectedErr: nil,
		},
		{
			name:        "Valid relative path with leading slash",
			input:       "/path/to/resource",
			expectedErr: nil,
		},
		{
			name:        "Valid relative with scheme and leading slash path",
			input:       "http:/path",
			expectedErr: nil,
		},
		{
			name:        "Valid relative with scheme and authority",
			input:       "http://host/path",
			expectedErr: nil,
		},
		{
			name:        "Invalid colon in first path segment without leading slash",
			input:       "http:path",
			expectedErr: &kindError{message: "Invalid IRI character in first path segment"},
		},
		{
			name:        "Invalid character in relative path",
			input:       "path[to",
			expectedErr: &kindError{message: "Invalid IRI character"},
		},
		{
			name:        "Starts with colon",
			input:       ":foo",
			expectedErr: errNoScheme,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			err := p.validateRelativeRef(tc.input)
			assertError(t, err, tc.expectedErr)
		})
	}
}

// TestParseRelative tests resolving relative references against an optional base IRI.
func TestParseRelative(t *testing.T) {
	baseRef, errp := ParseRef("http://example.com/b/c")
	if errp != nil {
		t.Fatalf("ParseRef() failed for base: %v", errp)
	}
	baseWithAuth := &base{IRI: baseRef.iri, Pos: baseRef.positions}

	testCases := []struct {
		name        string
		input       string
		base        *base
		expectedOut string
		expectedErr error
	}{
		{
			name:        "Relative path with base",
			input:       "d/e",
			base:        baseWithAuth,
			expectedOut: "http://example.com/b/d/e",
			expectedErr: nil,
		},
		{
			name:        "Absolute path with base",
			input:       "/d/e",
			base:        baseWithAuth,
			expectedOut: "http://example.com/d/e",
			expectedErr: nil,
		},
		{
			name:        "Query with base",
			input:       "?q=1",
			base:        baseWithAuth,
			expectedOut: "http://example.com/b/c?q=1",
			expectedErr: nil,
		},
		{
			name:        "Fragment with base",
			input:       "#frag",
			base:        baseWithAuth,
			expectedOut: "http://example.com/b/c#frag",
			expectedErr: nil,
		},
		{
			name:        "Empty reference with base",
			input:       "",
			base:        baseWithAuth,
			expectedOut: "http://example.com/b/c",
			expectedErr: nil,
		},
		{
			name:        "Network path reference with base",
			input:       "//other.org/p",
			base:        baseWithAuth,
			expectedOut: "http://other.org/p",
			expectedErr: nil,
		},
		{
			name:        "Absolute reference with base",
			input:       "https://other.com/x",
			base:        baseWithAuth,
			expectedOut: "https://other.com/x",
			expectedErr: nil,
		},
		{
			name:        "Invalid character in relative reference with base",
			input:       "d[e",
			base:        baseWithAuth,
			expectedOut: "",
			expectedErr: &kindError{message: "Invalid IRI character"},
		},
		{
			name:        "No base absolute path",
			input:       "/path/to/file",
			base:        nil,
			expectedOut: "/path/to/file",
			expectedErr: nil,
		},
		{
			name:        "No base relative path",
			input:       "path/to/file",
			base:        nil,
			expectedOut: "path/to/file",
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			if tc.base != nil {
				p.base = &iriParserBase{
					iri:          tc.base.IRI,
					schemeEnd:    tc.base.Pos.SchemeEnd,
					authorityEnd: tc.base.Pos.AuthorityEnd,
					pathEnd:      tc.base.Pos.PathEnd,
					queryEnd:     tc.base.Pos.QueryEnd,
					hasBase:      true,
				}
			}
			err := p.parseRelative()
			assertError(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				if got := p.output.string(); got != tc.expectedOut {
					t.Errorf("parseRelative() output = %q, want %q", got, tc.expectedOut)
				}
			}
		})
	}
}

// TestParseRelativeNoBase tests parsing relative references without a base IRI.
func TestParseRelativeNoBase(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedOut string
		expectedErr error
	}{
		{
			name:        "Absolute path",
			input:       "/path/to/file",
			expectedOut: "/path/to/file",
			expectedErr: nil,
		},
		{
			name:        "Relative path",
			input:       "path/to/file",
			expectedOut: "path/to/file",
			expectedErr: nil,
		},
		{
			name:        "Relative path with colon in first segment",
			input:       "path:file",
			expectedOut: "path",
			expectedErr: &kindError{message: "Invalid IRI character in first path segment"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := setupTestParser(tc.input, false)
			err := p.parseRelativeNoBase()
			assertError(t, err, tc.expectedErr)
			if tc.expectedErr == nil {
				if got := p.output.string(); got != tc.expectedOut {
					t.Errorf("parseRelativeNoBase() output = %q, want %q", got, tc.expectedOut)
				}
			}
		})
	}
}

// TestAbsoluteResolutionWithDotSegments verifies that resolving an absolute target IRI
// against a base IRI correctly normalizes relative dot-segments in its path.
func TestAbsoluteResolutionWithDotSegments(t *testing.T) {
	baseRef, err := ParseRef("http://www.example.com/foo/bar")
	if err != nil {
		t.Fatalf("ParseRef() failed for base: %v", err)
	}

	resolved, err := baseRef.Resolve("file:///C:/foo/../../../bar.html")
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	expected := "file:///bar.html"
	if got := resolved.String(); got != expected {
		t.Errorf("Resolve() resolved string = %q, want %q", got, expected)
	}
}

// TestAbsoluteResolutionWithFirstSegmentColon verifies that resolving an absolute target IRI
// with a first-segment colon (like "g:h") against a base IRI correctly resolves
// without triggering "Invalid IRI character in first path segment" validation errors.
func TestAbsoluteResolutionWithFirstSegmentColon(t *testing.T) {
	baseRef, err := ParseRef("http://a/b/c/d;p?q")
	if err != nil {
		t.Fatalf("ParseRef() failed for base: %v", err)
	}

	resolved, err := baseRef.Resolve("g:h")
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	expected := "g:h"
	if got := resolved.String(); got != expected {
		t.Errorf("Resolve() resolved string = %q, want %q", got, expected)
	}
}
