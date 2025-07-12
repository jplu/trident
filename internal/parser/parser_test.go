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

// Package parser_test contains the white-box tests for the parser package.
// Being in the same package allows testing of unexported functions and types,
// which is crucial for verifying internal logic like path normalization and
// component deconstruction.
package parser //nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.

import (
	"fmt"
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
		useVoidBuffer     bool // If true, use VoidOutputBuffer to test validation-only mode.
		expectedPositions Positions
	}{
		{
			name: "simple http",
			iri:  "http://example.com/foo?q=1#bar",
			expectedPositions: Positions{
				SchemeEnd:    5,  // "http:"
				AuthorityEnd: 18, // "//example.com"
				PathEnd:      22, // "/foo"
				QueryEnd:     26, // "?q=1"
			},
		},
		{
			name: "simple mailto",
			iri:  "mailto:John.Doe@example.com",
			expectedPositions: Positions{
				SchemeEnd:    7,  // "mailto:"
				AuthorityEnd: 7,  // (empty)
				PathEnd:      27, // "John.Doe@example.com"
				QueryEnd:     27, // (empty)
			},
		},
		{
			name: "urn",
			iri:  "urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
			expectedPositions: Positions{
				SchemeEnd:    4,
				AuthorityEnd: 4,
				PathEnd:      51,
				QueryEnd:     51,
			},
		},
		{
			name: "ipv6 host with port",
			iri:  "ldap://[2001:db8::7]:80/c=GB?one",
			expectedPositions: Positions{
				SchemeEnd:    5,
				AuthorityEnd: 23,
				PathEnd:      28,
				QueryEnd:     32,
			},
		},
		{
			name: "ipv4 host with empty port and path",
			iri:  "telnet://192.0.2.16:",
			expectedPositions: Positions{
				SchemeEnd:    7,
				AuthorityEnd: 20,
				PathEnd:      20,
				QueryEnd:     20,
			},
		},
		{
			name: "valid IRI with authority (formerly 'path starting with slashes')",
			iri:  "foo://bar",
			expectedPositions: Positions{
				SchemeEnd:    4,
				AuthorityEnd: 9,
				PathEnd:      9,
				QueryEnd:     9,
			},
		},
		{
			name:          "void buffer validation",
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
			name: "network-path reference",
			iri:  "//example.com/path",
			expectedPositions: Positions{
				SchemeEnd:    0,
				AuthorityEnd: 13,
				PathEnd:      18,
				QueryEnd:     18,
			},
		},
		{
			name: "path-absolute reference",
			iri:  "/path/to/resource",
			expectedPositions: Positions{
				SchemeEnd:    0,
				AuthorityEnd: 0,
				PathEnd:      17,
				QueryEnd:     17,
			},
		},
		{
			name: "ambiguous path parsed as scheme (formerly 'path-rootless reference')",
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

			positions, err := Run(tc.iri, nil, false, output)

			if err != nil {
				t.Fatalf("Did not expect an error, but got: %v", err)
			}
			if !tc.useVoidBuffer {
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
func TestRunRelativeResolution(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		baseIRI        string
		relativeIRI    string
		expectedResult string
	}{
		// Normal examples from RFC 3986, Section 5.4.1
		{"http://a/b/c/d;p?q", "g:h", "g:h"},
		{"http://a/b/c/d;p?q", "g", "http://a/b/c/g"},
		{"http://a/b/c/d;p?q", "./g", "http://a/b/c/g"},
		{"http://a/b/c/d;p?q", "g/", "http://a/b/c/g/"},
		{"http://a/b/c/d;p?q", "/g", "http://a/g"},
		{"http://a/b/c/d;p?q", "//g", "http://g"},
		{"http://a/b/c/d;p?q", "?y", "http://a/b/c/d;p?y"},
		{"http://a/b/c/d;p?q", "g?y", "http://a/b/c/g?y"},
		{"http://a/b/c/d;p?q", "#s", "http://a/b/c/d;p?q#s"},
		{"http://a/b/c/d;p?q", "g#s", "http://a/b/c/g#s"},
		{"http://a/b/c/d;p?q", "g?y#s", "http://a/b/c/g?y#s"},
		{"http://a/b/c/d;p?q", ";x", "http://a/b/c/;x"},
		{"http://a/b/c/d;p?q", "g;x", "http://a/b/c/g;x"},
		{"http://a/b/c/d;p?q", "g;x?y#s", "http://a/b/c/g;x?y#s"},
		{"http://a/b/c/d;p?q", "", "http://a/b/c/d;p?q"},
		{"http://a/b/c/d;p?q", ".", "http://a/b/c/"},
		{"http://a/b/c/d;p?q", "./", "http://a/b/c/"},
		{"http://a/b/c/d;p?q", "..", "http://a/b/"},
		{"http://a/b/c/d;p?q", "../", "http://a/b/"},
		{"http://a/b/c/d;p?q", "../g", "http://a/b/g"},
		{"http://a/b/c/d;p?q", "../..", "http://a/"},
		{"http://a/b/c/d;p?q", "../../", "http://a/"},
		{"http://a/b/c/d;p?q", "../../g", "http://a/g"},

		// Abnormal examples from RFC 3986, Section 5.4.2
		{"http://a/b/c/d;p?q", "../../../g", "http://a/g"},
		{"http://a/b/c/d;p?q", "../../../../g", "http://a/g"},
		{"http://a/b/c/d;p?q", "/./g", "http://a/g"},
		{"http://a/b/c/d;p?q", "/../g", "http://a/g"},
		{"http://a/b/c/d;p?q", "g.", "http://a/b/c/g."},
		{"http://a/b/c/d;p?q", ".g", "http://a/b/c/.g"},
		{"http://a/b/c/d;p?q", "g..", "http://a/b/c/g.."},
		{"http://a/b/c/d;p?q", "..g", "http://a/b/c/..g"},
		{"http://a/b/c/d;p?q", "./../g", "http://a/b/g"},
		{"http://a/b/c/d;p?q", "./g/.", "http://a/b/c/g/"},
		{"http://a/b/c/d;p?q", "g/./h", "http://a/b/c/g/h"},
		{"http://a/b/c/d;p?q", "g/../h", "http://a/b/c/h"},
		{"http://a/b/c/d;p?q", "g;x=1/./y", "http://a/b/c/g;x=1/y"},
		{"http://a/b/c/d;p?q", "g;x=1/../y", "http://a/b/c/y"},
		{"http://a/b/c/d;p?q", "http:g", "http:g"}, // scheme is present in reference

		// Custom tests for edge cases
		{"file:foo", "?bar", "file:foo?bar"},
		{"file:foo", "#bar", "file:foo#bar"},
		{"file:foo", "/lv2.h", "file:/lv2.h"},
		{"file:foo", "///lv2.h", "file:///lv2.h"}, // has authority
		{"file:foo", "lv2.h", "file:lv2.h"},
		{"http://example.com", "s", "http://example.com/s"}, // Base without path

		// Path normalization with multiple slashes
		{"fred:///s//a/b/c", "../g", "fred:///s//a/g"},
		{"fred:///s//a/b/c", "../../g", "fred:///s//g"},
		{"fred:///s//a/b/c", "../../../g", "fred:///s/g"},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("%s_RESOLVE_%s", tc.baseIRI, tc.relativeIRI)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// First, parse the base IRI to get its positions.
			basePos, err := Run(tc.baseIRI, nil, false, &VoidOutputBuffer{})
			if err != nil {
				t.Fatalf("Setup failed: could not parse base IRI %q: %v", tc.baseIRI, err)
			}
			base := &Base{IRI: tc.baseIRI, Pos: basePos}

			// Now, run the resolution.
			builder := &strings.Builder{}
			output := &StringOutputBuffer{Builder: builder}
			resolvedPos, err := Run(tc.relativeIRI, base, false, output)
			if err != nil {
				t.Fatalf("Resolution failed for relative IRI %q: %v", tc.relativeIRI, err)
			}

			// Check if the resolved string is correct.
			if builder.String() != tc.expectedResult {
				t.Errorf("Resolved IRI string mismatch\n got: %q\nwant: %q", builder.String(), tc.expectedResult)
			}

			// As a final check, parse the expected result and compare its positions
			// with the positions from the resolved IRI.
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
		base        *string // Optional base for relative reference tests.
		expectedErr string
	}{
		{"no scheme", ":no-scheme", nil, "No scheme found in an absolute IRI"},
		{"invalid char in scheme", "+foo:bar", nil, "Invalid IRI character"},
		{"invalid char in userinfo", "s://u@ser@host", nil, "Invalid IRI character"},
		{"invalid port char", "s://h:port/", nil, "Invalid port character"},
		{"invalid percent encoding 1", "%", nil, "Invalid IRI percent encoding"},
		{"invalid percent encoding 2", "%1", nil, "Invalid IRI percent encoding"},
		{"invalid percent encoding 3", "%GG", nil, "Invalid IRI percent encoding"},
		{"unterminated ipv6", "http://[::1", nil, "unterminated IPv6 literal"},
		{"invalid char after ipv6", "http://[::1]a", nil, "Invalid character after IP literal"},
		{"invalid ipvfuture version", "http://[vG.addr]", nil, "Invalid IPvFuture version char"},
		{"invalid ipvfuture separator", "http://[v1addr]", nil, "Invalid IPvFuture format: no dot separator"},
		{"invalid ipvfuture empty addr", "http://[v1.]", nil, "Invalid IPvFuture: empty address part"},
		{"space in host", "http://exa mple.com", nil, "Invalid IRI character"},
		{"invalid path char", "a/b/c^d", nil, "Invalid IRI character"},
		{"invalid query char", "?a=\uFDEF", nil, "Invalid IRI character"},
		{"invalid fragment char", "#a\nb", nil, "Invalid IRI character"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var base *Base
			if tc.base != nil {
				// Parse base in unchecked mode for test setup simplicity.
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
		input    string
		expected string
	}{
		{"a/b/c/./../../g", "a/g"},
		{"mid/content=5/../6", "mid/6"},
		{".", ""},
		{"./", ""},
		{"..", ""},
		{"../", ""},
		{"../g", "g"},
		{"../..", ""},
		{"../../", ""},
		{"../../g", "g"},
		{"/.", "/"},
		{"/./", "/"},
		{"/..", "/"},
		{"/../", "/"},
		{"/../g", "/g"},
		{"/a/../g", "/g"},
		{"/a/b/../../g", "/g"},
		{"a/./b", "a/b"},
		{"a/../b", "b"},
		{"a/b/.", "a/b/"},
		{"a/b/./", "a/b/"},
		{"a/b/..", "a/"},
		{"a/b/../", "a/"},
		{"a/b/c/../..", "a/"},
		{"//a/b", "//a/b"},
		{"/a/b//c", "/a/b//c"},
		{"a//b/c", "a//b/c"},
		{"a/b/..//c", "a//c"},
		{"./a", "a"},
		{"a/.", "a/"},
		{"/a/.", "/a/"},
		{"//", "//"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
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
		{"full", "s://a/p?q#f", deconstructedRefResult{"s", "a", "/p", "q", "f", true, true, true}},
		{"no fragment", "s://a/p?q", deconstructedRefResult{"s", "a", "/p", "q", "", true, true, false}},
		{"no query", "s://a/p#f", deconstructedRefResult{"s", "a", "/p", "", "f", true, false, true}},
		{"no path", "s://a?q#f", deconstructedRefResult{"s", "a", "", "q", "f", true, true, true}},
		{"no authority", "s:p?q#f", deconstructedRefResult{"s", "", "p", "q", "f", false, true, true}},
		{"no scheme", "//a/p?q#f", deconstructedRefResult{"", "a", "/p", "q", "f", true, true, true}},
		{"path only", "p", deconstructedRefResult{"", "", "p", "", "", false, false, false}},
		{"empty", "", deconstructedRefResult{"", "", "", "", "", false, false, false}},
		{"fragment only", "#f", deconstructedRefResult{"", "", "", "", "f", false, false, true}},
		{"query only", "?q", deconstructedRefResult{"", "", "", "q", "", false, true, false}},
		{"ambiguous path", "a:b", deconstructedRefResult{"a", "", "b", "", "", false, false, false}},
		{"ambiguous path 2", "a:b/c", deconstructedRefResult{"a", "", "b/c", "", "", false, false, false}},
		{"unambiguous path", "a/b:c", deconstructedRefResult{"", "", "a/b:c", "", "", false, false, false}},
		{"network path", "//a/b", deconstructedRefResult{"", "a", "/b", "", "", true, false, false}},
		{"path absolute", "/a/b", deconstructedRefResult{"", "", "/a/b", "", "", false, false, false}},
		{"empty authority", "s://?q", deconstructedRefResult{"s", "", "", "q", "", true, true, false}},
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
		{"truncate middle", "abcdef", 3, "abc"},
		{"truncate to zero", "abcdef", 0, ""},
		{"truncate to full length", "abcdef", 6, "abcdef"},
		{"truncate with negative", "abcdef", -1, "abcdef"},
		{"truncate beyond length", "abcdef", 10, "abcdef"},
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
		// Run with unchecked=false to exercise the validation logic.
		_, _ = Run(iri, nil, false, &VoidOutputBuffer{})
	})
}

// FuzzResolve subjects the IRI resolution logic to a wide range of arbitrary
// base and relative IRI inputs. Its purpose is to uncover panics or other
// unexpected behavior in the resolution and normalization algorithms.
func FuzzResolve(f *testing.F) {
	testCases := []struct {
		base string
		rel  string
	}{
		{"http://a/b/c/d;p?q", "g:h"},
		{"http://a/b/c/d;p?q", "../g"},
		{"file:foo", "/lv2.h"},
		{"http://example.com", ""},
		{"a:b", "c:d"},
	}
	for _, tc := range testCases {
		f.Add(tc.base, tc.rel)
	}

	f.Fuzz(func(_ *testing.T, baseIRI, relIRI string) {
		// First, try to parse the base IRI. We use unchecked mode here because
		// the fuzzer might generate an invalid base, which is fine. If parsing
		// fails, we just skip this fuzz case.
		basePos, err := Run(baseIRI, nil, true, &VoidOutputBuffer{})
		if err != nil {
			return
		}
		base := &Base{IRI: baseIRI, Pos: basePos}

		// Now, attempt to resolve the relative IRI against the base.
		_, _ = Run(relIRI, base, false, &VoidOutputBuffer{})
	})
}
