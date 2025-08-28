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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package iri

import (
	"reflect"
	"testing"
)

// Tests for `applyDotSegmentRules` are based on RFC 3986, Section 5.2.4, Step 2.
// This function is tested first as it is a core helper for `removeDotSegments`.
func TestApplyDotSegmentRules(t *testing.T) {
	testCases := []struct {
		name         string
		in           string
		output       []string
		expectedIn   string
		expectedOut  []string
		expectedOk   bool
		rfcReference string
	}{
		// Rule 2A: "../" or "./"
		{
			name:         "Rule 2A: ../ prefix",
			in:           "../a/b",
			output:       []string{},
			expectedIn:   "a/b",
			expectedOut:  []string{},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.A",
		},
		{
			name:         "Rule 2A: ./ prefix",
			in:           "./a/b",
			output:       []string{},
			expectedIn:   "a/b",
			expectedOut:  []string{},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.A",
		},
		// Rule 2B: "/./" or "/."
		{
			name:         "Rule 2B: /./ prefix",
			in:           "/./a/b",
			output:       []string{},
			expectedIn:   "/a/b",
			expectedOut:  []string{},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.B",
		},
		{
			name:         "Rule 2B: /. exact",
			in:           "/.",
			output:       []string{},
			expectedIn:   "/",
			expectedOut:  []string{},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.B",
		},
		// Rule 2C: "/../" or "/.."
		{
			name:         "Rule 2C: /../ prefix with non-empty output",
			in:           "/../a/b",
			output:       []string{"/parent"},
			expectedIn:   "/a/b",
			expectedOut:  []string{},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.C",
		},
		{
			name:         "Rule 2C: /.. exact with non-empty output",
			in:           "/..",
			output:       []string{"/parent"},
			expectedIn:   "/",
			expectedOut:  []string{},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.C",
		},
		{
			name:         "Rule 2C: /../ with empty output",
			in:           "/../a",
			output:       []string{},
			expectedIn:   "/a",
			expectedOut:  []string{},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.C",
		},
		{
			name:         "Rule 2C: /../ with relative last segment in output",
			in:           "/../c",
			output:       []string{"a", "b"},
			expectedIn:   "/c",
			expectedOut:  []string{"a"},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.C",
		},
		{
			name:         "Rule 2C: /../ with single relative segment in output",
			in:           "/../b",
			output:       []string{"a"},
			expectedIn:   "b",
			expectedOut:  []string{},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.C",
		},
		// Rule 2D: "." or ".."
		{
			name:         "Rule 2D: . exact",
			in:           ".",
			output:       []string{},
			expectedIn:   "",
			expectedOut:  []string{},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.D",
		},
		{
			name:         "Rule 2D: .. exact",
			in:           "..",
			output:       []string{},
			expectedIn:   "",
			expectedOut:  []string{},
			expectedOk:   true,
			rfcReference: "RFC 3986, 5.2.4, 2.D",
		},
		// No rule applied
		{
			name:         "No rule applies",
			in:           "/a/b/c",
			output:       []string{},
			expectedIn:   "/a/b/c",
			expectedOut:  []string{},
			expectedOk:   false,
			rfcReference: "RFC 3986, 5.2.4, 2.E",
		},
		{
			name:         "No rule applies, empty in",
			in:           "",
			output:       []string{},
			expectedIn:   "",
			expectedOut:  []string{},
			expectedOk:   false,
			rfcReference: "RFC 3986, 5.2.4, 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newIn, newOut, ok := applyDotSegmentRules(tc.in, tc.output)

			if newIn != tc.expectedIn {
				t.Errorf("expected in to be %q, got %q", tc.expectedIn, newIn)
			}
			if !reflect.DeepEqual(newOut, tc.expectedOut) {
				t.Errorf("expected out to be %v, got %v", tc.expectedOut, newOut)
			}
			if ok != tc.expectedOk {
				t.Errorf("expected ok to be %v, got %v", tc.expectedOk, ok)
			}
		})
	}
}

// Tests for `extractFirstSegment` are based on RFC 3986, Section 5.2.4, Step 2.E.
// This function is also a helper for `removeDotSegments`.
func TestExtractFirstSegment(t *testing.T) {
	testCases := []struct {
		name              string
		in                string
		expectedSegment   string
		expectedRemainder string
		rfcReference      string
	}{
		{
			name:              "Path with leading slash and multiple segments",
			in:                "/a/b/c",
			expectedSegment:   "/a",
			expectedRemainder: "/b/c",
			rfcReference:      "RFC 3986, 5.2.4, 2.E",
		},
		{
			name:              "Path with leading slash and single segment",
			in:                "/a",
			expectedSegment:   "/a",
			expectedRemainder: "",
			rfcReference:      "RFC 3986, 5.2.4, 2.E",
		},
		{
			name:              "Path with leading slash and empty segment",
			in:                "//a",
			expectedSegment:   "/",
			expectedRemainder: "/a",
			rfcReference:      "RFC 3986, 5.2.4, 2.E",
		},
		{
			name:              "Path without leading slash and multiple segments",
			in:                "a/b/c",
			expectedSegment:   "a",
			expectedRemainder: "/b/c",
			rfcReference:      "RFC 3986, 5.2.4, 2.E",
		},
		{
			name:              "Path without leading slash and single segment",
			in:                "a",
			expectedSegment:   "a",
			expectedRemainder: "",
			rfcReference:      "RFC 3986, 5.2.4, 2.E",
		},
		{
			name:              "Empty input",
			in:                "",
			expectedSegment:   "",
			expectedRemainder: "",
			rfcReference:      "RFC 3986, 5.2.4, 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			segment, remainder := extractFirstSegment(tc.in)
			if segment != tc.expectedSegment {
				t.Errorf("expected segment to be %q, got %q", tc.expectedSegment, segment)
			}
			if remainder != tc.expectedRemainder {
				t.Errorf("expected remainder to be %q, got %q", tc.expectedRemainder, remainder)
			}
		})
	}
}

// Tests for `removeDotSegments` are based on the examples from RFC 3986.
func TestRemoveDotSegments(t *testing.T) {
	// The base path directory used in many RFC 3986 examples for merging.
	basePathDir := "/a/b/c/"

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		// RFC 3986 Section 5.2.4 Examples
		{"RFC 5.2.4 Ex1", "/a/b/c/./../../g", "/a/g"},
		{"RFC 5.2.4 Ex2", "mid/content=5/../6", "mid/6"},

		// RFC 3986 Section 5.4.1 Normal Examples (Path part of transformation)
		{"RFC 5.4.1 g", basePathDir + "g", "/a/b/c/g"},
		{"RFC 5.4.1 ./g", basePathDir + "./g", "/a/b/c/g"},
		{"RFC 5.4.1 g/", basePathDir + "g/", "/a/b/c/g/"},
		{"RFC 5.4.1 /g", "/g", "/g"},
		{"RFC 5.4.1 ../g", basePathDir + "../g", "/a/b/g"},
		{"RFC 5.4.1 ../../g", basePathDir + "../../g", "/a/g"},
		{"RFC 5.4.1 .", basePathDir + ".", "/a/b/c/"},
		{"RFC 5.4.1 ./", basePathDir + "./", "/a/b/c/"},
		{"RFC 5.4.1 ..", basePathDir + "..", "/a/b/"},
		{"RFC 5.4.1 ../", basePathDir + "../", "/a/b/"},
		{"RFC 5.4.1 ../..", basePathDir + "../..", "/a/"},
		{"RFC 5.4.1 ../../", basePathDir + "../../", "/a/"},

		// RFC 3986 Section 5.4.2 Abnormal Examples
		{"RFC 5.4.2 ../../../g", basePathDir + "../../../g", "/g"},
		{"RFC 5.4.2 ../../../../g", basePathDir + "../../../../g", "/g"},
		{"RFC 5.4.2 /./g", "/./g", "/g"},
		{"RFC 5.4.2 /../g", "/../g", "/g"},
		{"RFC 5.4.2 g.", basePathDir + "g.", "/a/b/c/g."},
		{"RFC 5.4.2 .g", basePathDir + ".g", "/a/b/c/.g"},
		{"RFC 5.4.2 g..", basePathDir + "g..", "/a/b/c/g.."},
		{"RFC 5.4.2 ..g", basePathDir + "..g", "/a/b/c/..g"},
		{"RFC 5.4.2 ./../g", basePathDir + "./../g", "/a/b/g"},
		{"RFC 5.4.2 ./g/.", basePathDir + "./g/.", "/a/b/c/g/"},
		{"RFC 5.4.2 g/./h", basePathDir + "g/./h", "/a/b/c/g/h"},
		{"RFC 5.4.2 g/../h", basePathDir + "g/../h", "/a/b/c/h"},
		{"RFC 5.4.2 a/../b", "a/../b", "b"},

		// Edge cases
		{"Empty string", "", ""},
		{"Single slash", "/", "/"},
		{"Double slash", "//", "//"}, // removeDotSegments does not check for path validity
		{"Just a dot", ".", ""},
		{"Just two dots", "..", ""},
		{"Root traversal", "/..", "/"},
		{"Trailing dot", "/a/b/.", "/a/b/"},
		{"Trailing dots", "/a/b/..", "/a/"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := removeDotSegments(tc.input); got != tc.expected {
				t.Errorf("removeDotSegments(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

// Tests for `resolvePath` are based on RFC 3986, Section 5.2.3, "Merge Paths".
// `resolvePath` implements the second bullet point of this section.
func TestResolvePath(t *testing.T) {
	testCases := []struct {
		name     string
		basePath string
		relPath  string
		expected string
	}{
		// RFC 3986, 5.2.3: "return a string consisting of the reference's path component
		// appended to all but the last segment of the base URI's path"
		{
			name:     "RFC merge example 1",
			basePath: "/a/b/c/d;p",
			relPath:  "g",
			expected: "/a/b/c/g",
		},
		{
			name:     "RFC merge example 2",
			basePath: "/a/b/c/d;p",
			relPath:  "./g",
			expected: "/a/b/c/g",
		},
		{
			name:     "RFC merge with up-directory",
			basePath: "/a/b/c/d;p",
			relPath:  "../g",
			expected: "/a/b/g",
		},
		{
			name:     "Base path is a directory",
			basePath: "/a/b/c/",
			relPath:  "g",
			expected: "/a/b/c/g",
		},
		{
			name:     "Base path has no slashes",
			basePath: "a",
			relPath:  "b",
			expected: "b",
		},
		{
			name:     "Base path has no slashes with up-directory",
			basePath: "a",
			relPath:  "../g",
			expected: "g",
		},
		{
			name:     "Base path is empty",
			basePath: "",
			relPath:  "a/b",
			expected: "a/b",
		},
		{
			name:     "Relative path is empty",
			basePath: "/a/b/c",
			relPath:  "",
			expected: "/a/b/",
		},
		{
			name:     "Base path is root",
			basePath: "/",
			relPath:  "a",
			expected: "/a",
		},
		{
			name:     "Both paths are empty",
			basePath: "",
			relPath:  "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolvePath(tc.basePath, tc.relPath); got != tc.expected {
				t.Errorf("resolvePath(%q, %q) = %q, want %q", tc.basePath, tc.relPath, got, tc.expected)
			}
		})
	}
}
