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

// Package parser_test contains the white-box tests for the parser package.
// Being in the same package allows testing of unexported functions and types,
// which is crucial for verifying internal logic like path normalization and
// component deconstruction.
package parser //nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.

import (
	"reflect"
	"strings"
	"testing"
)

// cmpPositions is a helper function to compare two Positions structs. It reports
// a test error if they are not identical, providing a clear diff in the output.
func cmpPositions(t *testing.T, got, want Positions, context string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: Positions mismatch\n got: %+v\nwant: %+v", context, got, want)
	}
}

// TestRunAbsolute tests the Run function with various absolute IRIs and relative
// references without a base. It verifies that the parser correctly identifies
// and delineates the components (scheme, authority, path, etc.) for a wide range
// of valid IRI structures. It also tests the VoidOutputBuffer for validation-only runs.
func TestRunAbsolute(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name              string
		iri               string
		useVoidBuffer     bool
		unchecked         bool
		expectedPositions Positions
	}{
		{
			name: "simple http with all components",
			iri:  "http://example.com/foo?q=1#bar",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 18,
				PathEnd:      22,
				QueryEnd:     26,
			},
		},
		{
			name: "mailto scheme with path",
			iri:  "mailto:John.Doe@example.com",
			expectedPositions: Positions{
				SchemeEnd:    7,
				AuthorityEnd: 7,
				PathEnd:      27,
				QueryEnd:     27,
			},
		},
		{
			name: "urn scheme with path",
			iri:  "urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
			expectedPositions: Positions{
				SchemeEnd:    4,
				AuthorityEnd: 4,
				PathEnd:      51,
				QueryEnd:     51,
			},
		},
		{
			name: "ipv6 host with port and path",
			iri:  "ldap://[2001:db8::7]:80/c=GB?one",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 23,
				PathEnd:      28,
				QueryEnd:     32,
			},
		},
		{
			name: "ipv4 host with empty port and no path",
			iri:  "telnet://192.0.2.16:",
			expectedPositions: Positions{
				SchemeEnd:    7,
				AuthorityEnd: 20,
				PathEnd:      20,
				QueryEnd:     20,
			},
		},
		{
			name: "scheme with authority but no path",
			iri:  "foo://bar",
			expectedPositions: Positions{
				SchemeEnd:    4,
				AuthorityEnd: 9,
				PathEnd:      9,
				QueryEnd:     9,
			},
		},
		{
			name:          "void buffer validation only",
			iri:           "http://example.com/",
			useVoidBuffer: true,
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 18,
				PathEnd:      19,
				QueryEnd:     19,
			},
		},
		{
			name: "network-path reference (no scheme)",
			iri:  "//example.com/path",
			expectedPositions: Positions{
				SchemeEnd:    0,
				AuthorityEnd: 13,
				PathEnd:      18,
				QueryEnd:     18,
			},
		},
		{
			name: "path-absolute reference (no scheme)",
			iri:  "/path/to/resource",
			expectedPositions: Positions{
				SchemeEnd:    0,
				AuthorityEnd: 0,
				PathEnd:      17,
				QueryEnd:     17,
			},
		},
		{
			name: "path starting with scheme-like segment",
			iri:  "path:to/resource",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 5,
				PathEnd:      16,
				QueryEnd:     16,
			},
		},
		{
			name: "empty reference",
			iri:  "",
			expectedPositions: Positions{
				SchemeEnd:    0,
				AuthorityEnd: 0,
				PathEnd:      0,
				QueryEnd:     0,
			},
		},
		{
			name:      "unchecked mode with leading colon (parsed as empty scheme)",
			iri:       ":foo",
			unchecked: true,
			expectedPositions: Positions{
				SchemeEnd:    1,
				AuthorityEnd: 1,
				PathEnd:      4,
				QueryEnd:     4,
			},
		},
		{
			name: "ipv6 host without port followed by EOF",
			iri:  "http://[::1]",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 12,
				PathEnd:      12,
				QueryEnd:     12,
			},
		},
		{
			name: "ipv6 host followed by path",
			iri:  "http://[::1]/path",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 12,
				PathEnd:      17,
				QueryEnd:     17,
			},
		},
		{
			name: "ipv6 host followed by query",
			iri:  "http://[::1]?q",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 12,
				PathEnd:      12,
				QueryEnd:     14,
			},
		},
		{
			name: "ipv6 host followed by fragment",
			iri:  "http://[::1]#f",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 12,
				PathEnd:      12,
				QueryEnd:     12,
			},
		},
		{
			name: "regular host followed by query",
			iri:  "http://example.com?q",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 18,
				PathEnd:      18,
				QueryEnd:     20,
			},
		},
		{
			name: "regular host followed by fragment",
			iri:  "http://example.com#f",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 18,
				PathEnd:      18,
				QueryEnd:     18,
			},
		},
		{
			name: "ipvfuture host followed by path",
			iri:  "http://[v1.addr]/path",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 16,
				PathEnd:      21,
				QueryEnd:     21,
			},
		},
		{
			name: "path with percent encoding",
			iri:  "http://example.com/foo%20bar",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 18,
				PathEnd:      28,
				QueryEnd:     28,
			},
		},
		{
			name: "ipvfuture host with valid characters",
			iri:  "http://[v1.ab-cd:ef]/",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 20,
				PathEnd:      21,
				QueryEnd:     21,
			},
		},
		{
			name: "iunreserved unicode char in path",
			iri:  "http://a/\U000E1234",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 8,
				PathEnd:      13,
				QueryEnd:     13,
			},
		},
		{
			name:      "unchecked path after ipv6 literal without slash",
			iri:       "http://[::1]path",
			unchecked: true,
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 12,
				PathEnd:      0,
				QueryEnd:     0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var output OutputBuffer
			var builder *strings.Builder
			if tc.useVoidBuffer {
				output = &VoidOutputBuffer{}
			} else {
				builder = &strings.Builder{}
				output = &StringOutputBuffer{Builder: builder}
			}

			positions, err := Run(tc.iri, nil, tc.unchecked, output)

			if err != nil {
				t.Fatalf("Did not expect an error, but got: %v", err)
			}
			if !tc.useVoidBuffer && !tc.unchecked {
				if builder.String() != tc.iri {
					t.Errorf("Output string mismatch\n got: %q\nwant: %q", builder.String(), tc.iri)
				}
			}
			cmpPositions(t, positions, tc.expectedPositions, "Run Absolute")
		})
	}
}

// TestRunRelativeResolution tests the IRI reference resolution algorithm. It uses a
// set of base IRIs and relative references, many of which are drawn from the examples
// in RFC 3986, to ensure that the parser correctly resolves them to the expected absolute IRI.
// It also includes cases for testing the `unchecked` mode during resolution.
func TestRunRelativeResolution(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name           string
		baseIRI        string
		relativeIRI    string
		unchecked      bool
		expectedResult string
	}{
		// Normal examples from RFC 3986, Section 5.4.1
		{
			name:           "RFC3986 Normal: scheme",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g:h",
			expectedResult: "g:h",
		},
		{
			name:           "RFC3986 Normal: simple path",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g",
			expectedResult: "http://a/b/c/g",
		},
		{
			name:           "RFC3986 Normal: dot segment",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "./g",
			expectedResult: "http://a/b/c/g",
		},
		{
			name:           "RFC3986 Normal: path with slash",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g/",
			expectedResult: "http://a/b/c/g/",
		},
		{
			name:           "RFC3986 Normal: absolute path",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "/g",
			expectedResult: "http://a/g",
		},
		{
			name:           "RFC3986 Normal: network path",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "//g",
			expectedResult: "http://g",
		},
		{
			name:           "RFC3986 Normal: query only",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "?y",
			expectedResult: "http://a/b/c/d;p?y",
		},
		{
			name:           "RFC3986 Normal: path and query",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g?y",
			expectedResult: "http://a/b/c/g?y",
		},
		{
			name:           "RFC3986 Normal: fragment only",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "#s",
			expectedResult: "http://a/b/c/d;p?q#s",
		},
		{
			name:           "RFC3986 Normal: path and fragment",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g#s",
			expectedResult: "http://a/b/c/g#s",
		},
		{
			name:           "RFC3986 Normal: path, query, fragment",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g?y#s",
			expectedResult: "http://a/b/c/g?y#s",
		},
		{
			name:           "RFC3986 Normal: semicolon path",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    ";x",
			expectedResult: "http://a/b/c/;x",
		},
		{
			name:           "RFC3986 Normal: path and semicolon",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g;x",
			expectedResult: "http://a/b/c/g;x",
		},
		{
			name:           "RFC3986 Normal: full with semicolon",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g;x?y#s",
			expectedResult: "http://a/b/c/g;x?y#s",
		},
		{
			name:           "RFC3986 Normal: empty reference",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "",
			expectedResult: "http://a/b/c/d;p?q",
		},
		{
			name:           "RFC3986 Normal: single dot",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    ".",
			expectedResult: "http://a/b/c/",
		},
		{
			name:           "RFC3986 Normal: single dot slash",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "./",
			expectedResult: "http://a/b/c/",
		},
		{
			name:           "RFC3986 Normal: double dot",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "..",
			expectedResult: "http://a/b/",
		},
		{
			name:           "RFC3986 Normal: double dot slash",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "../",
			expectedResult: "http://a/b/",
		},
		{
			name:           "RFC3986 Normal: double dot path",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "../g",
			expectedResult: "http://a/b/g",
		},
		{
			name:           "RFC3986 Normal: two double dots",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "../..",
			expectedResult: "http://a/",
		},
		{
			name:           "RFC3986 Normal: two double dots slash",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "../../",
			expectedResult: "http://a/",
		},
		{
			name:           "RFC3986 Normal: two double dots path",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "../../g",
			expectedResult: "http://a/g",
		},
		{
			name:           "RFC3986 Normal: scheme with network path",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g://h",
			expectedResult: "g://h",
		},

		// Abnormal examples from RFC 3986, Section 5.4.2
		{
			name:           "RFC3986 Abnormal: three double dots",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "../../../g",
			expectedResult: "http://a/g",
		},
		{
			name:           "RFC3986 Abnormal: four double dots",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "../../../../g",
			expectedResult: "http://a/g",
		},
		{
			name:           "RFC3986 Abnormal: absolute dot",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "/./g",
			expectedResult: "http://a/g",
		},
		{
			name:           "RFC3986 Abnormal: absolute double dot",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "/../g",
			expectedResult: "http://a/g",
		},
		{
			name:           "RFC3986 Abnormal: path with trailing dot",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g.",
			expectedResult: "http://a/b/c/g.",
		},
		{
			name:           "RFC3986 Abnormal: path with leading dot",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    ".g",
			expectedResult: "http://a/b/c/.g",
		},
		{
			name:           "RFC3986 Abnormal: path with trailing double dot",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g..",
			expectedResult: "http://a/b/c/g..",
		},
		{
			name:           "RFC3986 Abnormal: path with leading double dot",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "..g",
			expectedResult: "http://a/b/c/..g",
		},
		{
			name:           "RFC3986 Abnormal: nested dots",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "./../g",
			expectedResult: "http://a/b/g",
		},
		{
			name:           "RFC3986 Abnormal: nested dots 2",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "./g/.",
			expectedResult: "http://a/b/c/g/",
		},
		{
			name:           "RFC3986 Abnormal: nested dots 3",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g/./h",
			expectedResult: "http://a/b/c/g/h",
		},
		{
			name:           "RFC3986 Abnormal: nested dots 4",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g/../h",
			expectedResult: "http://a/b/c/h",
		},
		{
			name:           "RFC3986 Abnormal: nested dots with semicolon",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g;x=1/./y",
			expectedResult: "http://a/b/c/g;x=1/y",
		},
		{
			name:           "RFC3986 Abnormal: nested dots with semicolon 2",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "g;x=1/../y",
			expectedResult: "http://a/b/c/y",
		},
		{
			name:           "RFC3986 Abnormal: scheme present in reference",
			baseIRI:        "http://a/b/c/d;p?q",
			relativeIRI:    "http:g",
			expectedResult: "http:g",
		},

		// Custom tests for edge cases
		{
			name:           "Custom: file scheme with query",
			baseIRI:        "file:foo",
			relativeIRI:    "?bar",
			expectedResult: "file:foo?bar",
		},
		{
			name:           "Custom: file scheme with fragment",
			baseIRI:        "file:foo",
			relativeIRI:    "#bar",
			expectedResult: "file:foo#bar",
		},
		{
			name:           "Custom: file scheme with absolute path",
			baseIRI:        "file:foo",
			relativeIRI:    "/lv2.h",
			expectedResult: "file:/lv2.h",
		},
		{
			name:           "Custom: file scheme with authority",
			baseIRI:        "file:foo",
			relativeIRI:    "///lv2.h",
			expectedResult: "file:///lv2.h",
		},
		{
			name:           "Custom: file scheme with relative path",
			baseIRI:        "file:foo",
			relativeIRI:    "lv2.h",
			expectedResult: "file:lv2.h",
		},
		{
			name:           "Custom: base without path",
			baseIRI:        "http://example.com",
			relativeIRI:    "s",
			expectedResult: "http://example.com/s",
		},

		// Path normalization with multiple slashes
		{
			name:           "Custom: path norm with multiple slashes 1",
			baseIRI:        "fred:///s//a/b/c",
			relativeIRI:    "../g",
			expectedResult: "fred:///s//a/g",
		},
		{
			name:           "Custom: path norm with multiple slashes 2",
			baseIRI:        "fred:///s//a/b/c",
			relativeIRI:    "../../g",
			expectedResult: "fred:///s//g",
		},
		{
			name:           "Custom: path norm with multiple slashes 3",
			baseIRI:        "fred:///s//a/b/c",
			relativeIRI:    "../../../g",
			expectedResult: "fred:///s/g",
		},

		// Test case to cover unchecked resolution path
		{
			name:           "Resolution with unchecked mode",
			baseIRI:        "http://example.com/a/b",
			relativeIRI:    "../c",
			unchecked:      true,
			expectedResult: "http://example.com/c",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			basePos, err := Run(tc.baseIRI, nil, true, &VoidOutputBuffer{})
			if err != nil {
				t.Fatalf("Setup failed: could not parse base IRI %q: %v", tc.baseIRI, err)
			}
			base := &Base{IRI: tc.baseIRI, Pos: basePos}

			builder := &strings.Builder{}
			output := &StringOutputBuffer{Builder: builder}
			resolvedPos, err := Run(tc.relativeIRI, base, tc.unchecked, output)
			if err != nil {
				t.Fatalf("Resolution failed for relative IRI %q with base %q: %v", tc.relativeIRI, tc.baseIRI, err)
			}

			if builder.String() != tc.expectedResult {
				t.Errorf("Resolved IRI string mismatch\n got: %q\nwant: %q", builder.String(), tc.expectedResult)
			}

			expectedPos, err := Run(tc.expectedResult, nil, false, &VoidOutputBuffer{})
			if err != nil {
				t.Fatalf("Post-check failed: could not parse expected result %q: %v", tc.expectedResult, err)
			}
			cmpPositions(t, resolvedPos, expectedPos, "Resolved Positions")
		})
	}
}

// TestRunInvalid tests that the parser correctly identifies and rejects
// invalid IRI strings. Each test case provides a malformed IRI and the
// expected error message substring, ensuring the validation logic is working
// as intended.
func TestRunInvalid(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		iri         string
		base        *string
		expectedErr string
	}{
		{
			name:        "no scheme",
			iri:         ":no-scheme",
			base:        nil,
			expectedErr: "No scheme found in an absolute IRI",
		},
		{
			name:        "invalid char in scheme",
			iri:         "+foo:bar",
			base:        nil,
			expectedErr: "Invalid IRI character",
		},
		{
			name:        "invalid char in userinfo",
			iri:         "s://u@ser@host",
			base:        nil,
			expectedErr: "Invalid IRI character",
		},
		{
			name:        "invalid port character",
			iri:         "s://h:port/",
			base:        nil,
			expectedErr: "Invalid port character",
		},
		{
			name:        "invalid percent encoding (incomplete)",
			iri:         "%1",
			base:        nil,
			expectedErr: "Invalid IRI percent encoding",
		},
		{
			name:        "invalid percent encoding (non-hex)",
			iri:         "%GG",
			base:        nil,
			expectedErr: "Invalid IRI percent encoding",
		},
		{
			name:        "unterminated ipv6 literal",
			iri:         "http://[::1",
			base:        nil,
			expectedErr: "unterminated IPv6 literal",
		},
		{
			name:        "invalid ip in literal (non-hex)",
			iri:         "http://[::G]",
			base:        nil,
			expectedErr: "Invalid host IP",
		},
		{
			name:        "invalid char after ipv6",
			iri:         "http://[::1]a",
			base:        nil,
			expectedErr: "Invalid character after IP literal",
		},
		{
			name:        "invalid ipvfuture version char",
			iri:         "http://[vG.addr]",
			base:        nil,
			expectedErr: "Invalid IPvFuture version char",
		},
		{
			name:        "invalid ipvfuture no dot separator",
			iri:         "http://[v1addr]",
			base:        nil,
			expectedErr: "Invalid IPvFuture format: no dot separator",
		},
		{
			name:        "invalid ipvfuture missing version",
			iri:         "http://[v.addr]",
			base:        nil,
			expectedErr: "Invalid IPvFuture: missing version",
		},
		{
			name:        "invalid ipvfuture empty address part",
			iri:         "http://[v1.]",
			base:        nil,
			expectedErr: "Invalid IPvFuture: empty address part",
		},
		{
			name:        "invalid ipvfuture address char",
			iri:         "http://[v1.addr^]",
			base:        nil,
			expectedErr: "Invalid IPvFuture address char",
		},
		{
			name:        "space in host",
			iri:         "http://exa mple.com",
			base:        nil,
			expectedErr: "Invalid IRI character",
		},
		{
			name:        "invalid path char",
			iri:         "a/b/c^d",
			base:        nil,
			expectedErr: "Invalid IRI character",
		},
		{
			name:        "invalid query char (in private use area)",
			iri:         "?a=\uFDEF",
			base:        nil,
			expectedErr: "Invalid IRI character",
		},
		{
			name:        "invalid fragment char (newline)",
			iri:         "#a\nb",
			base:        nil,
			expectedErr: "Invalid IRI character",
		},
		{
			name:        "invalid character  in userinfo",
			iri:         "s://user^info@host",
			base:        nil,
			expectedErr: "Invalid IRI character '^'",
		},
		{
			name:        "invalid relative ref with base",
			iri:         "%GG",
			base:        func() *string { s := "http://example.com/"; return &s }(),
			expectedErr: "Invalid IRI percent encoding",
		},
		{
			name:        "path starts with slashes after resolution",
			iri:         "../..//c",
			base:        func() *string { s := "foo:a/b"; return &s }(),
			expectedErr: "An IRI path is not allowed to start with // if there is no authority",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var base *Base
			if tc.base != nil {
				pos, err := Run(*tc.base, nil, true, &VoidOutputBuffer{})
				if err != nil {
					t.Fatalf("Failed to parse base for test: %v", err)
				}
				base = &Base{IRI: *tc.base, Pos: pos}
			}

			_, err := Run(tc.iri, base, false, &VoidOutputBuffer{})

			if err == nil {
				t.Fatalf("Expected an error for %q but got none", tc.iri)
			}
			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Errorf("Expected error to contain %q, but got %q", tc.expectedErr, err.Error())
			}
		})
	}
}

// TestRemoveDotSegments provides unit tests for the unexported removeDotSegments
// function. This ensures the path normalization logic is correct in isolation,
// covering various edge cases of "." and ".." segments.
func TestRemoveDotSegments(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "complex relative path",
			input:    "a/b/c/./../../g",
			expected: "a/g",
		},
		{
			name:     "path with content and navigation",
			input:    "mid/content=5/../6",
			expected: "mid/6",
		},
		{
			name:     "single dot",
			input:    ".",
			expected: "",
		},
		{
			name:     "single dot with slash",
			input:    "./",
			expected: "",
		},
		{
			name:     "double dot",
			input:    "..",
			expected: "",
		},
		{
			name:     "double dot with slash",
			input:    "../",
			expected: "",
		},
		{
			name:     "double dot with path",
			input:    "../g",
			expected: "g",
		},
		{
			name:     "two double dots",
			input:    "../..",
			expected: "",
		},
		{
			name:     "two double dots with slash",
			input:    "../../",
			expected: "",
		},
		{
			name:     "two double dots with path",
			input:    "../../g",
			expected: "g",
		},
		{
			name:     "absolute single dot",
			input:    "/.",
			expected: "/",
		},
		{
			name:     "absolute single dot with slash",
			input:    "/./",
			expected: "/",
		},
		{
			name:     "absolute double dot",
			input:    "/..",
			expected: "/",
		},
		{
			name:     "absolute double dot with slash",
			input:    "/../",
			expected: "/",
		},
		{
			name:     "absolute double dot with path",
			input:    "/../g",
			expected: "/g",
		},
		{
			name:     "absolute path navigation",
			input:    "/a/../g",
			expected: "/g",
		},
		{
			name:     "absolute path deep navigation",
			input:    "/a/b/../../g",
			expected: "/g",
		},
		{
			name:     "path with internal dot segment",
			input:    "a/./b",
			expected: "a/b",
		},
		{
			name:     "path with internal double dot segment",
			input:    "a/../b",
			expected: "b",
		},
		{
			name:     "path with trailing dot segment",
			input:    "a/b/.",
			expected: "a/b/",
		},
		{
			name:     "path with trailing dot segment and slash",
			input:    "a/b/./",
			expected: "a/b/",
		},
		{
			name:     "path with trailing double dot segment",
			input:    "a/b/..",
			expected: "a/",
		},
		{
			name:     "path with trailing double dot segment and slash",
			input:    "a/b/../",
			expected: "a/",
		},
		{
			name:     "path with deep trailing navigation",
			input:    "a/b/c/../..",
			expected: "a/",
		},
		{
			name:     "path with double slash authority",
			input:    "//a/b",
			expected: "//a/b",
		},
		{
			name:     "path with internal double slash",
			input:    "/a/b//c",
			expected: "/a/b//c",
		},
		{
			name:     "path with internal double slash relative",
			input:    "a//b/c",
			expected: "a//b/c",
		},
		{
			name:     "path with internal double slash and navigation",
			input:    "a/b/..//c",
			expected: "a//c",
		},
		{
			name:     "path with leading dot segment",
			input:    "./a",
			expected: "a",
		},
		{
			name:     "path with trailing dot segment no slash",
			input:    "a/.",
			expected: "a/",
		},
		{
			name:     "absolute path with trailing dot",
			input:    "/a/.",
			expected: "/a/",
		},
		{
			name:     "path is just double slash",
			input:    "//",
			expected: "//",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := removeDotSegments(tc.input)
			if result != tc.expected {
				t.Errorf("removeDotSegments(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// deconstructedRefResult is a helper struct for testing deconstructRef. It holds
// the expected components of a deconstructed reference string.
type deconstructedRefResult struct {
	Scheme, Authority, Path, Query, Fragment string
	HasAuth, HasQuery, HasFrag               bool
}

// TestDeconstructRef tests the unexported deconstructRef function, which is
// the first step in reference resolution. This test ensures that a reference
// string is correctly broken down into its five main components before the
// merging and normalization logic is applied.
func TestDeconstructRef(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		ref  string
		want deconstructedRefResult
	}{
		{
			name: "full iri",
			ref:  "s://a/p?q#f",
			want: deconstructedRefResult{
				Scheme:    "s",
				Authority: "a",
				Path:      "/p",
				Query:     "q",
				Fragment:  "f",
				HasAuth:   true,
				HasQuery:  true,
				HasFrag:   true,
			},
		},
		{
			name: "no fragment",
			ref:  "s://a/p?q",
			want: deconstructedRefResult{
				Scheme:    "s",
				Authority: "a",
				Path:      "/p",
				Query:     "q",
				Fragment:  "",
				HasAuth:   true,
				HasQuery:  true,
				HasFrag:   false,
			},
		},
		{
			name: "no query",
			ref:  "s://a/p#f",
			want: deconstructedRefResult{
				Scheme:    "s",
				Authority: "a",
				Path:      "/p",
				Query:     "",
				Fragment:  "f",
				HasAuth:   true,
				HasQuery:  false,
				HasFrag:   true,
			},
		},
		{
			name: "no path",
			ref:  "s://a?q#f",
			want: deconstructedRefResult{
				Scheme:    "s",
				Authority: "a",
				Path:      "",
				Query:     "q",
				Fragment:  "f",
				HasAuth:   true,
				HasQuery:  true,
				HasFrag:   true,
			},
		},
		{
			name: "no authority",
			ref:  "s:p?q#f",
			want: deconstructedRefResult{
				Scheme:    "s",
				Authority: "",
				Path:      "p",
				Query:     "q",
				Fragment:  "f",
				HasAuth:   false,
				HasQuery:  true,
				HasFrag:   true,
			},
		},
		{
			name: "no scheme",
			ref:  "//a/p?q#f",
			want: deconstructedRefResult{
				Scheme:    "",
				Authority: "a",
				Path:      "/p",
				Query:     "q",
				Fragment:  "f",
				HasAuth:   true,
				HasQuery:  true,
				HasFrag:   true,
			},
		},
		{
			name: "path only",
			ref:  "p",
			want: deconstructedRefResult{
				Scheme:    "",
				Authority: "",
				Path:      "p",
				Query:     "",
				Fragment:  "",
				HasAuth:   false,
				HasQuery:  false,
				HasFrag:   false,
			},
		},
		{
			name: "empty",
			ref:  "",
			want: deconstructedRefResult{
				Scheme:    "",
				Authority: "",
				Path:      "",
				Query:     "",
				Fragment:  "",
				HasAuth:   false,
				HasQuery:  false,
				HasFrag:   false,
			},
		},
		{
			name: "fragment only",
			ref:  "#f",
			want: deconstructedRefResult{
				Scheme:    "",
				Authority: "",
				Path:      "",
				Query:     "",
				Fragment:  "f",
				HasAuth:   false,
				HasQuery:  false,
				HasFrag:   true,
			},
		},
		{
			name: "query only",
			ref:  "?q",
			want: deconstructedRefResult{
				Scheme:    "",
				Authority: "",
				Path:      "",
				Query:     "q",
				Fragment:  "",
				HasAuth:   false,
				HasQuery:  true,
				HasFrag:   false,
			},
		},
		{
			name: "ambiguous path with colon",
			ref:  "a:b",
			want: deconstructedRefResult{
				Scheme:    "a",
				Authority: "",
				Path:      "b",
				Query:     "",
				Fragment:  "",
				HasAuth:   false,
				HasQuery:  false,
				HasFrag:   false,
			},
		},
		{
			name: "ambiguous path with colon and slash",
			ref:  "a:b/c",
			want: deconstructedRefResult{
				Scheme:    "a",
				Authority: "",
				Path:      "b/c",
				Query:     "",
				Fragment:  "",
				HasAuth:   false,
				HasQuery:  false,
				HasFrag:   false,
			},
		},
		{
			name: "unambiguous path with colon",
			ref:  "a/b:c",
			want: deconstructedRefResult{
				Scheme:    "",
				Authority: "",
				Path:      "a/b:c",
				Query:     "",
				Fragment:  "",
				HasAuth:   false,
				HasQuery:  false,
				HasFrag:   false,
			},
		},
		{
			name: "network path",
			ref:  "//a/b",
			want: deconstructedRefResult{
				Scheme:    "",
				Authority: "a",
				Path:      "/b",
				Query:     "",
				Fragment:  "",
				HasAuth:   true,
				HasQuery:  false,
				HasFrag:   false,
			},
		},
		{
			name: "path absolute",
			ref:  "/a/b",
			want: deconstructedRefResult{
				Scheme:    "",
				Authority: "",
				Path:      "/a/b",
				Query:     "",
				Fragment:  "",
				HasAuth:   false,
				HasQuery:  false,
				HasFrag:   false,
			},
		},
		{
			name: "empty authority",
			ref:  "s://?q",
			want: deconstructedRefResult{
				Scheme:    "s",
				Authority: "",
				Path:      "",
				Query:     "q",
				Fragment:  "",
				HasAuth:   true,
				HasQuery:  true,
				HasFrag:   false,
			},
		},
		{
			name: "invalid scheme start char",
			ref:  "1a:b",
			want: deconstructedRefResult{
				Scheme:    "",
				Authority: "",
				Path:      "1a:b",
				Query:     "",
				Fragment:  "",
				HasAuth:   false,
				HasQuery:  false,
				HasFrag:   false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotScheme, gotAuthority, gotPath, gotQuery, gotFragment, gotHasAuth, gotHasQuery, gotHasFrag := deconstructRef(
				tc.ref,
			)

			got := deconstructedRefResult{
				Scheme:    gotScheme,
				Authority: gotAuthority,
				Path:      gotPath,
				Query:     gotQuery,
				Fragment:  gotFragment,
				HasAuth:   gotHasAuth,
				HasQuery:  gotHasQuery,
				HasFrag:   gotHasFrag,
			}

			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("deconstructRef(%q) mismatch:\n got: %+v\nwant: %+v", tc.ref, got, tc.want)
			}
		})
	}
}

// TestStringOutputBuffer_Truncate verifies the behavior of the Truncate method
// on the StringOutputBuffer, ensuring it correctly handles various scenarios
// like truncating to zero, to the middle, and with out-of-bounds values.
func TestStringOutputBuffer_Truncate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		initial  string
		truncate int
		expected string
	}{
		{
			name:     "truncate middle",
			initial:  "abcdef",
			truncate: 3,
			expected: "abc",
		},
		{
			name:     "truncate to zero",
			initial:  "abcdef",
			truncate: 0,
			expected: "",
		},
		{
			name:     "truncate to full length",
			initial:  "abcdef",
			truncate: 6,
			expected: "abcdef",
		},
		{
			name:     "truncate with negative",
			initial:  "abcdef",
			truncate: -1,
			expected: "abcdef",
		},
		{
			name:     "truncate beyond length",
			initial:  "abcdef",
			truncate: 10,
			expected: "abcdef",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			builder := &strings.Builder{}
			builder.WriteString(tc.initial)
			buffer := StringOutputBuffer{Builder: builder}

			buffer.Truncate(tc.truncate)

			if buffer.String() != tc.expected {
				t.Errorf("After Truncate(%d), got %q, want %q", tc.truncate, buffer.String(), tc.expected)
			}
			if buffer.Len() != len(tc.expected) {
				t.Errorf("After Truncate(%d), Len() is %d, want %d", tc.truncate, buffer.Len(), len(tc.expected))
			}
		})
	}
}

// FuzzParse subjects the main Run function to a wide range of arbitrary inputs.
// Its goal is to find edge cases that cause panics or other unexpected behavior
// when parsing a standalone IRI. It runs in validation-only mode for efficiency.
func FuzzParse(f *testing.F) {
	testCases := []string{
		"http://example.com/foo?q=1#bar",
		"//example.com/path",
		"/path/to/resource",
		"path:to/resource",
		"",
		"%",
		"http://[::1",
		"foo://bar",
		":no-scheme",
	}
	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(_ *testing.T, iri string) {
		_, _ = Run(iri, nil, false, &VoidOutputBuffer{})
	})
}

// FuzzResolve subjects the IRI resolution logic to a wide range of arbitrary
// base and relative IRI inputs. Its purpose is to uncover panics or other
// unexpected behavior in the resolution and normalization algorithms.
func FuzzResolve(f *testing.F) {
	testCases := []struct {
		name string
		base string
		rel  string
	}{
		{
			name: "RFC3986 example with new scheme",
			base: "http://a/b/c/d;p?q",
			rel:  "g:h",
		},
		{
			name: "RFC3986 example with path normalization",
			base: "http://a/b/c/d;p?q",
			rel:  "../g",
		},
		{
			name: "File scheme with absolute path",
			base: "file:foo",
			rel:  "/lv2.h",
		},
		{
			name: "Empty relative reference",
			base: "http://example.com",
			rel:  "",
		},
		{
			name: "Ambiguous colon in path",
			base: "a:b",
			rel:  "c:d",
		},
	}
	for _, tc := range testCases {
		f.Add(tc.base, tc.rel)
	}

	f.Fuzz(func(_ *testing.T, baseIRI, relIRI string) {
		basePos, err := Run(baseIRI, nil, true, &VoidOutputBuffer{})
		if err != nil {
			return
		}
		base := &Base{IRI: baseIRI, Pos: basePos}

		_, _ = Run(relIRI, base, false, &VoidOutputBuffer{})
	})
}

// TestVoidOutputBuffer checks the behavior of the VoidOutputBuffer, ensuring it
// correctly tracks length without storing data, and that its methods like
// Truncate and Reset work as expected.
func TestVoidOutputBuffer(t *testing.T) {
	t.Parallel()
	b := &VoidOutputBuffer{}

	if b.Len() != 0 {
		t.Errorf("Initial Len() got %d, want 0", b.Len())
	}
	if b.String() != "" {
		t.Errorf("Initial String() got %q, want \"\"", b.String())
	}

	b.WriteRune('a')
	b.WriteString("xyz")
	expectedLen := 4
	if b.Len() != expectedLen {
		t.Errorf("After writes, Len() got %d, want %d", b.Len(), expectedLen)
	}

	if b.String() != "" {
		t.Errorf("String() should always be empty, but got %q", b.String())
	}

	b.Truncate(2)
	if b.Len() != 2 {
		t.Errorf("After Truncate(2), Len() got %d, want 2", b.Len())
	}

	b.Reset()
	if b.Len() != 0 {
		t.Errorf("After Reset, Len() got %d, want 0", b.Len())
	}
}
