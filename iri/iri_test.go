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
	"errors"
	"strings"
	"testing"
)

// TestParseError_Error tests the Error method of the ParseError type.
func TestParseError_Error(t *testing.T) {
	err := &ParseError{Message: "test message"}
	expected := "IRI parse error: test message"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestParseError_Unwrap tests the Unwrap method of the ParseError type.
func TestParseError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := &ParseError{Message: "wrapper", Err: innerErr}
	if unwrapped := err.Unwrap(); !errors.Is(unwrapped, innerErr) {
		t.Errorf("Expected unwrapped error to be '%v', got '%v'", innerErr, unwrapped)
	}
	if unwrapped := (&ParseError{}).Unwrap(); unwrapped != nil {
		t.Errorf("Expected unwrapped error to be nil, got '%v'", unwrapped)
	}
}

// TestParseRef_Valid tests parsing of various valid IRI-references.
func TestParseRef_Valid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"Absolute IRI", "http://example.com/p?q#f"},
		{"Valid Absolute IRI with colon in path", "a:b/c"},
		{"Relative-path reference", "a/b/c"},
		{"Absolute-path reference", "/a/b/c"},
		{"Network-path reference", "//example.com/path"},
		{"Empty reference", ""},
		{"Fragment-only reference", "#fragment"},
		{"Query-only reference", "?query"},
		{"URN", "urn:isbn:0451450523"},
		{"IRI with non-ASCII chars", "http://例子.com/résumé"},
		{"Valid absolute IRI with single-letter scheme", "a:b"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := ParseRef(tc.input)
			if err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}
			if ref.String() != tc.input {
				t.Errorf("Expected ref string '%s', got '%s'", tc.input, ref.String())
			}
		})
	}
}

// TestParseRef_Invalid tests parsing of various invalid IRI-references.
func TestParseRef_Invalid(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		errMsg string
	}{
		{"Invalid scheme start", "1http://example.com", "Invalid IRI character in first path segment"},
		{"Invalid path with // no authority", "scheme:..//path", "An IRI path is not allowed to start with //"},
		{"Invalid percent encoding", "http://example.com/%GG", "Invalid IRI percent encoding"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := ParseRef(tc.input)
			if err == nil {
				t.Fatal("Expected an error, but got none")
			}
			if ref != nil {
				t.Fatal("Expected a nil Ref on error, but got a value")
			}
			if !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tc.errMsg, err.Error())
			}
		})
	}
}

// TestParseNormalizedRef tests that ParseNormalizedRef applies NFC normalization.
func TestParseNormalizedRef(t *testing.T) {
	// Error case: colon in first path segment of a relative reference.
	_, err := ParseNormalizedRef("1:b")
	if err == nil {
		t.Fatal("Expected an error for invalid IRI, but got none")
	}
}

// TestParseIri tests that ParseIri requires an absolute IRI.
func TestParseIri(t *testing.T) {
	t.Run("Valid Absolute", func(t *testing.T) {
		iri, err := ParseIri("http://example.com")
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		if !iri.IsAbsolute() {
			t.Error("Expected IRI to be absolute")
		}
	})
	t.Run("Relative", func(t *testing.T) {
		_, err := ParseIri("/relative/path")
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}
		if !strings.Contains(err.Error(), "No scheme found") {
			t.Errorf("Expected error message to contain 'No scheme found', got '%s'", err.Error())
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := ParseIri("http://[")
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}
	})
}

// TestParseNormalizedIri tests ParseNormalizedIri with NFC normalization.
func TestParseNormalizedIri(t *testing.T) {
	decomposed := "e\u0301"
	composed := "\u00e9"
	iriStr := "http://example.com/" + decomposed
	iri, err := ParseNormalizedIri(iriStr)
	if err != nil {
		t.Fatalf("ParseNormalizedIri failed: %v", err)
	}
	expectedStr := "http://example.com/" + composed
	if iri.String() != expectedStr {
		t.Errorf("Expected IRI string to be normalized to NFC '%s', got '%s'", expectedStr, iri.String())
	}
	_, err = ParseNormalizedIri("/relative")
	if err == nil {
		t.Fatal("Expected an error for relative IRI, but got none")
	}
	_, err = ParseNormalizedIri("1:b")
	if err == nil {
		t.Fatal("Expected an error for invalid IRI, but got none")
	}
}

// TestNewIriFromRef tests creating an Iri from a Ref.
func TestNewIriFromRef(t *testing.T) {
	t.Run("Absolute Ref", func(t *testing.T) {
		ref := mustParseRef(t, "http://example.com")
		iri, err := NewIriFromRef(ref)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		if iri.String() != "http://example.com" {
			t.Errorf("Expected iri string 'http://example.com', got '%s'", iri.String())
		}
	})

	t.Run("Relative Ref", func(t *testing.T) {
		ref := mustParseRef(t, "/path/to/resource")
		iri, err := NewIriFromRef(ref)
		if err == nil {
			t.Fatal("Expected an error, but got none")
		}
		if iri != nil {
			t.Fatal("Expected a nil Iri on error, but got a value")
		}
		if !strings.Contains(err.Error(), "No scheme found") {
			t.Errorf("Expected error message to contain 'No scheme found', got '%s'", err.Error())
		}
	})
}

// TestRef_Resolve_NormalExamples tests resolution based on RFC 3986, Section 5.4.1.
func TestRef_Resolve_NormalExamples(t *testing.T) {
	base := mustParseRef(t, "http://a/b/c/d;p?q")
	testCases := map[string]string{
		"g:h":     "g:h",
		"g":       "http://a/b/c/g",
		"./g":     "http://a/b/c/g",
		"g/":      "http://a/b/c/g/",
		"/g":      "http://a/g",
		"//g":     "http://g",
		"?y":      "http://a/b/c/d;p?y",
		"g?y":     "http://a/b/c/g?y",
		"#s":      "http://a/b/c/d;p?q#s",
		"g#s":     "http://a/b/c/g#s",
		"g?y#s":   "http://a/b/c/g?y#s",
		";x":      "http://a/b/c/;x",
		"g;x":     "http://a/b/c/g;x",
		"g;x?y#s": "http://a/b/c/g;x?y#s",
		"":        "http://a/b/c/d;p?q",
		".":       "http://a/b/c/",
		"./":      "http://a/b/c/",
		"..":      "http://a/b/",
		"../":     "http://a/b/",
		"../g":    "http://a/b/g",
		"../..":   "http://a/",
		"../../":  "http://a/",
		"../../g": "http://a/g",
	}

	for rel, expected := range testCases {
		t.Run(rel, func(t *testing.T) {
			resolved, err := base.Resolve(rel)
			if err != nil {
				t.Fatalf("Resolve failed for '%s': %v", rel, err)
			}
			if resolved.String() != expected {
				t.Errorf("For relative '%s', expected resolved IRI '%s', got '%s'", rel, expected, resolved.String())
			}
		})
	}
}

// TestRef_Resolve_AbnormalExamples tests resolution based on RFC 3986, Section 5.4.2.
func TestRef_Resolve_AbnormalExamples(t *testing.T) {
	base := mustParseRef(t, "http://a/b/c/d;p?q")
	testCases := map[string]string{
		"../../../g":    "http://a/g",
		"../../../../g": "http://a/g",
		"/./g":          "http://a/g",
		"/../g":         "http://a/g",
		"g.":            "http://a/b/c/g.",
		".g":            "http://a/b/c/.g",
		"g..":           "http://a/b/c/g..",
		"..g":           "http://a/b/c/..g",
		"./../g":        "http://a/b/g",
		"./g/.":         "http://a/b/c/g/",
		"g/./h":         "http://a/b/c/g/h",
		"g/../h":        "http://a/b/c/h",
		"g;x=1/./y":     "http://a/b/c/g;x=1/y",
		"g;x=1/../y":    "http://a/b/c/y",
		"g?y/./x":       "http://a/b/c/g?y/./x",
		"g?y/../x":      "http://a/b/c/g?y/../x",
		"g#s/./x":       "http://a/b/c/g#s/./x",
		"g#s/../x":      "http://a/b/c/g#s/../x",
	}

	for rel, expected := range testCases {
		t.Run(rel, func(t *testing.T) {
			resolved, err := base.Resolve(rel)
			if err != nil {
				t.Fatalf("Resolve failed for '%s': %v", rel, err)
			}
			if resolved.String() != expected {
				t.Errorf("For relative '%s', expected resolved IRI '%s', got '%s'", rel, expected, resolved.String())
			}
		})
	}
}

// TestRef_Resolve_Error tests resolution with an invalid relative reference.
func TestRef_Resolve_Error(t *testing.T) {
	base := mustParseRef(t, "http://a/b/c/d;p?q")
	_, err := base.Resolve("1:b")
	if err == nil {
		t.Fatal("Expected an error, but got none")
	}
	expectedMsg := "Invalid IRI character in first path segment"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestRef_ResolveTo tests the optimized resolution to a strings.Builder.
func TestRef_ResolveTo(t *testing.T) {
	base := mustParseRef(t, "http://a/b/c/d;p?q")
	relativeIRI := "../g"
	expectedIRI := "http://a/b/g"

	var builder strings.Builder
	pos, err := base.ResolveTo(relativeIRI, &builder)
	if err != nil {
		t.Fatalf("ResolveTo failed: %v", err)
	}
	resolvedStr := builder.String()
	if resolvedStr != expectedIRI {
		t.Errorf("Expected resolved string '%s', got '%s'", expectedIRI, resolvedStr)
	}

	var scheme, authority, path string
	var hasScheme, hasAuthority bool

	if pos.SchemeEnd > 0 {
		scheme = resolvedStr[:pos.SchemeEnd-1]
		hasScheme = true
	}
	if pos.AuthorityEnd > pos.SchemeEnd {
		authorityComponent := resolvedStr[pos.SchemeEnd:pos.AuthorityEnd]
		authority = strings.TrimPrefix(authorityComponent, "//")
		hasAuthority = true
	}
	path = resolvedStr[pos.AuthorityEnd:pos.PathEnd]

	if !hasScheme || scheme != "http" {
		t.Errorf("Expected scheme 'http', got '%s' (present: %v)", scheme, hasScheme)
	}
	if !hasAuthority || authority != "a" {
		t.Errorf("Expected authority 'a', got '%s' (present: %v)", authority, hasAuthority)
	}
	if path != "/b/g" {
		t.Errorf("Expected path '/b/g', got '%s'", path)
	}

	var errBuilder strings.Builder
	_, err = base.ResolveTo("1:b", &errBuilder)
	if err == nil {
		t.Fatal("Expected an error, but got none")
	}
	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Errorf("Expected error of type *ParseError, got %T", err)
	}
}

// TestIri_Resolve tests the resolution of a relative IRI reference against a base Iri.
func TestIri_Resolve(t *testing.T) {
	iri := mustParseIri(t, "http://a/b/c/d;p?q")
	resolved, err := iri.Resolve("../g")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if resolved.String() != "http://a/b/g" {
		t.Errorf("Expected resolved IRI 'http://a/b/g', got '%s'", resolved.String())
	}
	_, err = iri.Resolve("1:b")
	if err == nil {
		t.Fatal("Expected an error for invalid relative ref, but got none")
	}
}

// TestIri_ResolveTo tests the optimized resolution against a base Iri to a strings.Builder.
func TestIri_ResolveTo(t *testing.T) {
	iri := mustParseIri(t, "http://a/b/c/d;p?q")
	var builder strings.Builder
	err := iri.ResolveTo("../g", &builder)
	if err != nil {
		t.Fatalf("ResolveTo failed: %v", err)
	}
	if builder.String() != "http://a/b/g" {
		t.Errorf("Expected resolved string 'http://a/b/g', got '%s'", builder.String())
	}
	var errBuilder strings.Builder
	err = iri.ResolveTo("1:b", &errBuilder)
	if err == nil {
		t.Fatal("Expected an error for invalid relative ref, but got none")
	}
}

// TestIri_Relativize_Valid tests creating valid relative references.
func TestIri_Relativize_Valid(t *testing.T) {
	testCases := []struct {
		name     string
		base     string
		target   string
		expected string
	}{
		{"Same document", "http://a/b/c", "http://a/b/c", ""},
		{"Same path, add fragment", "http://a/b/c", "http://a/b/c#frag", "#frag"},
		{"Same path, different query", "http://a/b/c?q1", "http://a/b/c?q2", "?q2"},
		{"Path is subdirectory", "http://a/b/c", "http://a/b/c/d/e", "c/d/e"},
		{"Path goes up one level", "http://a/b/c/d", "http://a/b/c/e", "e"},
		{"Path goes up multiple levels", "http://a/b/c/d", "http://a/e", "../../e"},
		{"Different authority", "http://a/b/c", "http://x/y/z", "//x/y/z"},
		{"Different authority (no path)", "http://a/b/c", "http://x", "//x"},
		{"Different scheme", "http://a/b/c", "https://x/y/z", "https://x/y/z"},
		{"Same path, no target query", "http://a/b/c?q", "http://a/b/c", "c"},
		{"Same authority, different root path", "http://a/b", "http://a/c", "c"},
		{"Base with empty path", "http://a", "http://a/b/c", "b/c"},
		{"Base path to root path", "http://a/b/c", "http://a/", "../"},
		{"Different authority, no target authority", "http://a/b", "mailto:user@b", "mailto:user@b"},
		{"Base has authority, target does not", "http://example.com/a", "http:/b/c", "http:/b/c"},
		{"Target path is empty (with authority)", "http://a/b", "http://a", "//a"},
		{"Target path is empty (no authority)", "mailto:user@example.com", "mailto:", "mailto:"},
		{"Target path is empty", "http://a/b", "http://a/", "."},
		{"Base has no authority", "mailto:a@b.com", "mailto:c@d.com", "c@d.com"},
		{"No authority, up and down path", "foo:a/b/c", "foo:a/d/e", "../d/e"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			base := mustParseIri(t, tc.base)
			target := mustParseIri(t, tc.target)
			relativeRef, err := base.Relativize(target)
			if err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}
			if relativeRef.String() != tc.expected {
				t.Errorf("Expected relative ref '%s', got '%s'", tc.expected, relativeRef.String())
			}
		})
	}
}

// TestIri_Relativize_Invalid tests cases where relativization should fail.
func TestIri_Relativize_Invalid(t *testing.T) {
	testCases := []struct {
		name   string
		base   string
		target string
	}{
		{"Target has dot segments", "http://a/b/c", "http://a/b/./d"},
		{"Target has .. segment", "http://a/b/c", "http://a/b/../d"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			base := mustParseIri(t, tc.base)
			target := mustParseIri(t, tc.target)
			_, err := base.Relativize(target)
			if err == nil {
				t.Fatal("Expected an error, but got none")
			}
			if !errors.Is(err, ErrIriRelativize) {
				t.Errorf("Expected error '%v', but got '%v'", ErrIriRelativize, err)
			}
		})
	}
}
