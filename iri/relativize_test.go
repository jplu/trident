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
	"testing"
)

// mustParseAbsoluteIri is a helper that parses a string into an Iri for tests,
// panicking if the string is invalid.
func mustParseAbsoluteIri(s string) *Iri {
	iri, err := ParseIri(s)
	if err != nil {
		panic("test setup failed: could not parse base IRI: " + s)
	}
	return iri
}

// TestBuildRelativeRef tests the construction of a relative reference from its parts.
// This is the most basic building block for the relativize functions, based on
// the component recomposition logic from RFC 3986, Section 5.3.
func TestBuildRelativeRef(t *testing.T) {
	testCases := []struct {
		name     string
		relPath  string
		abs      *Iri
		expected string
	}{
		{
			name:     "relPath only, no query or fragment",
			relPath:  "c/d",
			abs:      mustParseAbsoluteIri("http://example.com/a/b"),
			expected: "c/d",
		},
		{
			name:     "relPath with query",
			relPath:  "c/d",
			abs:      mustParseAbsoluteIri("http://example.com/a/b?q=1"),
			expected: "c/d?q=1",
		},
		{
			name:     "relPath with fragment",
			relPath:  "c/d",
			abs:      mustParseAbsoluteIri("http://example.com/a/b#frag"),
			expected: "c/d#frag",
		},
		{
			name:     "relPath with query and fragment",
			relPath:  "c/d",
			abs:      mustParseAbsoluteIri("http://example.com/a/b?q=1#frag"),
			expected: "c/d?q=1#frag",
		},
		{
			name:     "empty relPath with query and fragment",
			relPath:  "",
			abs:      mustParseAbsoluteIri("http://example.com/a/b?q=1#frag"),
			expected: "?q=1#frag",
		},
		{
			name:     "empty relPath with only query",
			relPath:  "",
			abs:      mustParseAbsoluteIri("http://example.com/a/b?q=1"),
			expected: "?q=1",
		},
		{
			name:     "empty relPath with only fragment",
			relPath:  "",
			abs:      mustParseAbsoluteIri("http://example.com/a/b#frag"),
			expected: "#frag",
		},
		{
			name:     "empty relPath with no query or fragment",
			relPath:  "",
			abs:      mustParseAbsoluteIri("http://example.com/a/b"),
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := buildRelativeRef(tc.relPath, tc.abs)
			if err != nil {
				t.Fatalf("buildRelativeRef failed: %v", err)
			}
			if ref.String() != tc.expected {
				t.Errorf("Expected relative ref '%s', got '%s'", tc.expected, ref.String())
			}
		})
	}
}

// TestRelativizeForSamePathWithEmptyTargetQuery tests the edge case where paths
// are identical, but the target IRI has no query while the base does.
func TestRelativizeForSamePathWithEmptyTargetQuery(t *testing.T) {
	// This function is an internal helper. Its behavior depends only on the `abs`
	// argument, not the receiver `i`.
	base := mustParseAbsoluteIri("http://a/b/c?q=base")

	testCases := []struct {
		name     string
		target   *Iri
		expected string
	}{
		{
			name:     "target path has segments, no query/fragment",
			target:   mustParseAbsoluteIri("http://a/b/c"),
			expected: "c",
		},
		{
			name:     "target path has segments, with fragment",
			target:   mustParseAbsoluteIri("http://a/b/c#frag"),
			expected: "c#frag",
		},
		{
			name:     "target path ends with slash",
			target:   mustParseAbsoluteIri("http://a/b/c/"),
			expected: ".",
		},
		{
			name:     "target path ends with slash and has fragment",
			target:   mustParseAbsoluteIri("http://a/b/c/#frag"),
			expected: ".#frag",
		},
		{
			name:     "target has empty path and no authority",
			target:   mustParseAbsoluteIri("mailto:user@example.com"),
			expected: "mailto:user@example.com", // Returns full IRI string
		},
		{
			name:     "target has empty path and authority",
			target:   mustParseAbsoluteIri("http://example.com"),
			expected: "//example.com", // Returns scheme-relative reference
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := base.relativizeForSamePathWithEmptyTargetQuery(tc.target)
			if err != nil {
				t.Fatalf("relativizeForSamePathWithEmptyTargetQuery failed: %v", err)
			}
			if ref.String() != tc.expected {
				t.Errorf("Expected relative ref '%s', got '%s'", tc.expected, ref.String())
			}
		})
	}
}

// TestRelativizeForSamePath tests relativization when base and target paths are identical.
// RFC 3986, Section 4.4 discusses same-document references, which are the primary output here.
func TestRelativizeForSamePath(t *testing.T) {
	base := mustParseAbsoluteIri("http://a/b/c?q=1")

	testCases := []struct {
		name     string
		target   *Iri
		expected string
	}{
		{
			name:     "Identical query, no fragment -> empty ref",
			target:   mustParseAbsoluteIri("http://a/b/c?q=1"),
			expected: "",
		},
		{
			name:     "Identical query, with fragment -> fragment ref",
			target:   mustParseAbsoluteIri("http://a/b/c?q=1#frag"),
			expected: "#frag",
		},
		{
			name:     "Different query -> query ref",
			target:   mustParseAbsoluteIri("http://a/b/c?q=2"),
			expected: "?q=2",
		},
		{
			name:     "Different query and fragment -> query+fragment ref",
			target:   mustParseAbsoluteIri("http://a/b/c?q=2#frag"),
			expected: "?q=2#frag",
		},
		{
			name:     "Base has query, target has none -> path-based ref",
			target:   mustParseAbsoluteIri("http://a/b/c"),
			expected: "c",
		},
		{
			name:     "Base has no query, target has query -> query ref",
			target:   mustParseAbsoluteIri("http://a/b/c?q=2"),
			expected: "?q=2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var testBase *Iri
			if tc.name == "Base has no query, target has query -> query ref" {
				testBase = mustParseAbsoluteIri("http://a/b/c")
			} else {
				testBase = base
			}
			ref, err := testBase.relativizeForSamePath(tc.target)
			if err != nil {
				t.Fatalf("relativizeForSamePath failed: %v", err)
			}
			if ref.String() != tc.expected {
				t.Errorf("Expected relative ref '%s', got '%s'", tc.expected, ref.String())
			}
		})
	}
}

// TestRelativizeForNoAuthority tests relativization for IRIs without an authority component,
// such as 'mailto:' or 'urn:'.
func TestRelativizeForNoAuthority(t *testing.T) {
	testCases := []struct {
		name     string
		base     *Iri
		target   *Iri
		expected string
	}{
		{
			name:     "simple sibling path",
			base:     mustParseAbsoluteIri("scheme:a/b/c"),
			target:   mustParseAbsoluteIri("scheme:a/b/d"),
			expected: "d",
		},
		{
			name:     "path goes up and down",
			base:     mustParseAbsoluteIri("scheme:a/b/c"),
			target:   mustParseAbsoluteIri("scheme:a/d/e"),
			expected: "../d/e",
		},
		{
			name:     "target is deeper",
			base:     mustParseAbsoluteIri("scheme:a/b/"),
			target:   mustParseAbsoluteIri("scheme:a/b/c/d"),
			expected: "c/d",
		},
		{
			name:     "target is parent directory",
			base:     mustParseAbsoluteIri("scheme:a/b/c"),
			target:   mustParseAbsoluteIri("scheme:a/b/"),
			expected: ".",
		},
		{
			name:     "relative path with colon requires ./ prefix (no slashes)",
			base:     mustParseAbsoluteIri("urn:foo:a"),
			target:   mustParseAbsoluteIri("urn:foo:b:c"),
			expected: "./foo:b:c",
		},
		{
			name:     "relative path with colon in first segment requires ./ prefix",
			base:     mustParseAbsoluteIri("urn:foo:a/b"),
			target:   mustParseAbsoluteIri("urn:foo:a/c:d"),
			expected: "./c:d",
		},
		{
			name:     "empty relpath becomes dot",
			base:     mustParseAbsoluteIri("scheme:a/b"),
			target:   mustParseAbsoluteIri("scheme:a/"),
			expected: ".",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := tc.base.relativizeForNoAuthority(tc.target)
			if err != nil {
				t.Fatalf("relativizeForNoAuthority failed: %v", err)
			}
			if ref.String() != tc.expected {
				t.Errorf("Expected relative ref '%s', got '%s'", tc.expected, ref.String())
			}
		})
	}
}

// TestRelativizeWithAuthority tests relativization for IRIs with an authority component.
// The logic is the inverse of the path resolution defined in RFC 3986, Section 5.2.
func TestRelativizeWithAuthority(t *testing.T) {
	base := mustParseAbsoluteIri("http://a/b/c/d;p")

	testCases := []struct {
		name     string
		base     *Iri
		target   string
		expected string
	}{
		// Cases derived from reversing RFC 3986 Section 5.4.1 examples.
		{name: "RFC Example: g", base: base, target: "http://a/b/c/g", expected: "g"},
		{name: "RFC Example: g/", base: base, target: "http://a/b/c/g/", expected: "g/"},
		{name: "RFC Example: /g", base: base, target: "http://a/g", expected: "../../g"},
		{name: "RFC Example: ../g", base: base, target: "http://a/b/g", expected: "../g"},
		{name: "RFC Example: ../../g", base: base, target: "http://a/g", expected: "../../g"},
		{name: "RFC Example: ../..", base: base, target: "http://a/", expected: "../../"},
		{
			name:     "target path is prefix of base path",
			base:     base,
			target:   "http://a/b/",
			expected: "../",
		},
		{
			name:     "target is a sibling file",
			base:     base,
			target:   "http://a/b/c/g",
			expected: "g",
		},
		{
			name:     "target is same directory as base file",
			base:     base,
			target:   "http://a/b/c/",
			expected: ".",
		},
		{
			name:     "target has query and fragment",
			base:     base,
			target:   "http://a/b/g?y#s",
			expected: "../g?y#s",
		},
		{
			name:     "base path is empty, treated as /",
			base:     mustParseAbsoluteIri("http://a"),
			target:   "http://a/g",
			expected: "g",
		},
		{
			name:     "target path is slash, treated as /",
			base:     base,
			target:   "http://a/",
			expected: "../../",
		},
		{
			name:     "target path is empty, treated as /",
			base:     base,
			target:   "http://a", // This target IRI has an empty path ""
			expected: "../../",
		},
		{
			name:     "base is directory, target is file in it",
			base:     mustParseAbsoluteIri("http://a/b/c/"),
			target:   "http://a/b/c/g",
			expected: "g",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetIRI := mustParseAbsoluteIri(tc.target)
			ref, err := tc.base.relativizeWithAuthority(targetIRI)
			if err != nil {
				t.Fatalf("relativizeWithAuthority failed: %v", err)
			}
			if ref.String() != tc.expected {
				t.Errorf("Expected relative ref '%s', got '%s'", tc.expected, ref.String())
			}
		})
	}
}
