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
	"errors"
	"testing"
)

// TestRelativize tests the primary Relativize method on an absolute IRI,
// covering all dispatch branches: target paths with dot segments, mismatched schemes,
// differing authorities, empty paths with and without authority, same-path delegation,
// and delegation to authority-based or no-authority algorithms.
func TestRelativize(t *testing.T) {
	testCases := []struct {
		name        string
		base        *Iri
		target      *Iri
		expected    string
		expectError bool
		expectedErr error
	}{
		{
			name:        "target path contains single dot segment in middle",
			base:        mustParseAbsoluteIri("http://example.com/a/b"),
			target:      mustParseAbsoluteIri("http://example.com/a/./b"),
			expectError: true,
			expectedErr: ErrIriRelativize,
		},
		{
			name:        "target path contains double dot segment in middle",
			base:        mustParseAbsoluteIri("http://example.com/a/b"),
			target:      mustParseAbsoluteIri("http://example.com/a/../b"),
			expectError: true,
			expectedErr: ErrIriRelativize,
		},
		{
			name:        "target path is single dot",
			base:        mustParseAbsoluteIri("http://example.com/a/b"),
			target:      mustParseAbsoluteIri("http://example.com/."),
			expectError: true,
			expectedErr: ErrIriRelativize,
		},
		{
			name:        "target path is double dot",
			base:        mustParseAbsoluteIri("http://example.com/a/b"),
			target:      mustParseAbsoluteIri("http://example.com/.."),
			expectError: true,
			expectedErr: ErrIriRelativize,
		},
		{
			name:     "different schemes (http vs https)",
			base:     mustParseAbsoluteIri("http://example.com/a/b"),
			target:   mustParseAbsoluteIri("https://example.com/a/b"),
			expected: "https://example.com/a/b",
		},
		{
			name:     "different schemes (http vs ftp)",
			base:     mustParseAbsoluteIri("http://example.com/a/b"),
			target:   mustParseAbsoluteIri("ftp://example.com/a/b"),
			expected: "ftp://example.com/a/b",
		},
		{
			name:     "base has authority, target lacks authority",
			base:     mustParseAbsoluteIri("http://example.com/a/b"),
			target:   mustParseAbsoluteIri("http:a/b"),
			expected: "http:a/b",
		},
		{
			name:     "base lacks authority, target has authority",
			base:     mustParseAbsoluteIri("http:a/b"),
			target:   mustParseAbsoluteIri("http://example.com/a/b"),
			expected: "//example.com/a/b",
		},
		{
			name:     "both have authorities, but authorities differ",
			base:     mustParseAbsoluteIri("http://example.com/a/b"),
			target:   mustParseAbsoluteIri("http://other.com/a/b"),
			expected: "//other.com/a/b",
		},
		{
			name:     "different ports in authority",
			base:     mustParseAbsoluteIri("http://example.com:8080/a"),
			target:   mustParseAbsoluteIri("http://example.com:9090/a"),
			expected: "//example.com:9090/a",
		},
		{
			name:     "target has empty path and authority, base has path",
			base:     mustParseAbsoluteIri("http://example.com/a/b"),
			target:   mustParseAbsoluteIri("http://example.com"),
			expected: "//example.com",
		},
		{
			name:     "target has empty path and no authority, base has path",
			base:     mustParseAbsoluteIri("urn:foo"),
			target:   mustParseAbsoluteIri("urn:"),
			expected: "urn:",
		},
		{
			name:     "same path, different query",
			base:     mustParseAbsoluteIri("http://example.com/a/b?q=1"),
			target:   mustParseAbsoluteIri("http://example.com/a/b?q=2"),
			expected: "?q=2",
		},
		{
			name:     "same path, identical",
			base:     mustParseAbsoluteIri("http://example.com/a/b"),
			target:   mustParseAbsoluteIri("http://example.com/a/b"),
			expected: "",
		},
		{
			name:     "neither has authority, delegating to relativizeForNoAuthority",
			base:     mustParseAbsoluteIri("urn:foo:a/b/c"),
			target:   mustParseAbsoluteIri("urn:foo:a/b/d"),
			expected: "d",
		},
		{
			name:     "both have authority, delegating to relativizeWithAuthority",
			base:     mustParseAbsoluteIri("http://example.com/a/b/c"),
			target:   mustParseAbsoluteIri("http://example.com/a/b/d"),
			expected: "d",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := tc.base.Relativize(tc.target)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.expectedErr != nil && !errors.Is(err, tc.expectedErr) {
					t.Errorf("expected error %v, got %v", tc.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Relativize failed: %v", err)
			}
			if ref.String() != tc.expected {
				t.Errorf("expected relative ref %q, got %q", tc.expected, ref.String())
			}
		})
	}
}

// TestBuildRelativeRef tests the construction of a relative reference from its parts.
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
			expected: "mailto:user@example.com",
		},
		{
			name:     "target has empty path and authority",
			target:   mustParseAbsoluteIri("http://example.com"),
			expected: "//example.com",
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

// TestRelativizeForNoAuthority tests relativization for IRIs without an authority component.
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
			name:     "relative path with colon in subsequent segment does not require ./ prefix",
			base:     mustParseAbsoluteIri("urn:foo:a/b/c"),
			target:   mustParseAbsoluteIri("urn:foo:a/b/d/e:f"),
			expected: "d/e:f",
		},
		{
			name:     "relpath starting with slash",
			base:     mustParseAbsoluteIri("urn:a"),
			target:   mustParseAbsoluteIri("urn:/b/c"),
			expected: "/b/c",
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
func TestRelativizeWithAuthority(t *testing.T) {
	base := mustParseAbsoluteIri("http://a/b/c/d;p")

	testCases := []struct {
		name     string
		base     *Iri
		target   string
		expected string
	}{
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
			target:   "http://a",
			expected: "../../",
		},
		{
			name:     "base is directory, target is file in it",
			base:     mustParseAbsoluteIri("http://a/b/c/"),
			target:   "http://a/b/c/g",
			expected: "g",
		},
		{
			name:     "target is parent file of directory base",
			base:     mustParseAbsoluteIri("http://example.com/a/b/c"),
			target:   "http://example.com/a/b",
			expected: "../b",
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

// TestSplitPathSegments tests the internal helper function splitPathSegments
// to ensure correct behavior and full branch coverage.
func TestSplitPathSegments(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectedSegs []string
		expectedFile string
	}{
		{
			name:         "empty path",
			path:         "",
			expectedSegs: nil,
			expectedFile: "",
		},
		{
			name:         "path with trailing slash",
			path:         "a/b/",
			expectedSegs: []string{"a", "b"},
			expectedFile: "",
		},
		{
			name:         "normal path without trailing slash",
			path:         "a/b/c",
			expectedSegs: []string{"a", "b"},
			expectedFile: "c",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			segs, file := splitPathSegments(tc.path)

			if len(segs) != len(tc.expectedSegs) {
				t.Fatalf("expected %d segments, got %d", len(tc.expectedSegs), len(segs))
			}
			for i := range segs {
				if segs[i] != tc.expectedSegs[i] {
					t.Errorf("at index %d: expected segment %q, got %q", i, tc.expectedSegs[i], segs[i])
				}
			}
			if file != tc.expectedFile {
				t.Errorf("expected file %q, got %q", tc.expectedFile, file)
			}
		})
	}
}
