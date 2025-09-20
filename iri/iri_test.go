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
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"
)

// mustParseRef is a helper that parses a string as a Ref and fails the test if there's an error.
func mustParseRef(t *testing.T, s string) *Ref {
	t.Helper()
	r, err := ParseRef(s)
	if err != nil {
		t.Fatalf("mustParseRef failed for input '%s': %v", s, err)
	}
	return r
}

// mustParseIri is a helper that parses a string as an Iri and fails the test if there's an error.
func mustParseIri(t *testing.T, s string) *Iri {
	t.Helper()
	i, err := ParseIri(s)
	if err != nil {
		t.Fatalf("mustParseIri failed for input '%s': %v", s, err)
	}
	return i
}

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

// TestRef_String tests that the String method of a Ref returns the original parsed string.
func TestRef_String(t *testing.T) {
	// RFC 3987 Section 2: "an IRI is defined as a sequence of characters"
	// The String() method should return this original sequence.
	iriStr := "http://example.com/path?query#fragment"
	ref := mustParseRef(t, iriStr)
	if ref.String() != iriStr {
		t.Errorf("Expected String() to return '%s', got '%s'", iriStr, ref.String())
	}
}

type componentTestCase struct {
	name         string
	iri          string
	isAbsolute   bool
	scheme       string
	hasScheme    bool
	authority    string
	hasAuthority bool
	path         string
	query        string
	hasQuery     bool
	fragment     string
	hasFragment  bool
}

// assertComponents checks if the components of a Ref match the expectations in a test case.
func assertComponents(t *testing.T, ref *Ref, tc componentTestCase) {
	t.Helper()

	if got := ref.IsAbsolute(); got != tc.isAbsolute {
		t.Errorf("IsAbsolute() = %v, want %v", got, tc.isAbsolute)
	}

	s, ok := ref.Scheme()
	if ok != tc.hasScheme || s != tc.scheme {
		t.Errorf("Scheme() = (%q, %v), want (%q, %v)", s, ok, tc.scheme, tc.hasScheme)
	}

	a, ok := ref.Authority()
	if ok != tc.hasAuthority || a != tc.authority {
		t.Errorf("Authority() = (%q, %v), want (%q, %v)", a, ok, tc.authority, tc.hasAuthority)
	}

	if p := ref.Path(); p != tc.path {
		t.Errorf("Path() = %q, want %q", p, tc.path)
	}

	q, ok := ref.Query()
	if ok != tc.hasQuery || q != tc.query {
		t.Errorf("Query() = (%q, %v), want (%q, %v)", q, ok, tc.query, tc.hasQuery)
	}

	f, ok := ref.Fragment()
	if ok != tc.hasFragment || f != tc.fragment {
		t.Errorf("Fragment() = (%q, %v), want (%q, %v)", f, ok, tc.fragment, tc.hasFragment)
	}
}

// TestRef_ComponentAccessors tests the various methods for accessing IRI components on a Ref.
func TestRef_ComponentAccessors(t *testing.T) {
	testCases := []componentTestCase{
		{
			name:         "Full IRI",
			iri:          "foo://example.com:8042/over/there?name=ferret#nose",
			isAbsolute:   true,
			scheme:       "foo",
			hasScheme:    true,
			authority:    "example.com:8042",
			hasAuthority: true,
			path:         "/over/there",
			query:        "name=ferret",
			hasQuery:     true,
			fragment:     "nose",
			hasFragment:  true,
		},
		{
			name:         "Relative Reference",
			iri:          "/path/to/resource?key=val#frag",
			isAbsolute:   false,
			scheme:       "",
			hasScheme:    false,
			authority:    "",
			hasAuthority: false,
			path:         "/path/to/resource",
			query:        "key=val",
			hasQuery:     true,
			fragment:     "frag",
			hasFragment:  true,
		},
		{
			name:         "URN with no authority",
			iri:          "urn:example:animal:ferret:nose",
			isAbsolute:   true,
			scheme:       "urn",
			hasScheme:    true,
			authority:    "",
			hasAuthority: false,
			path:         "example:animal:ferret:nose",
			query:        "",
			hasQuery:     false,
			fragment:     "",
			hasFragment:  false,
		},
		{
			name:         "No Query or Fragment",
			iri:          "http://example.com/path",
			isAbsolute:   true,
			scheme:       "http",
			hasScheme:    true,
			authority:    "example.com",
			hasAuthority: true,
			path:         "/path",
			query:        "",
			hasQuery:     false,
			fragment:     "",
			hasFragment:  false,
		},
		{
			name:         "No Path",
			iri:          "mailto:user@example.com",
			isAbsolute:   true,
			scheme:       "mailto",
			hasScheme:    true,
			authority:    "",
			hasAuthority: false,
			path:         "user@example.com",
			query:        "",
			hasQuery:     false,
			fragment:     "",
			hasFragment:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref := mustParseRef(t, tc.iri)
			assertComponents(t, ref, tc)
		})
	}
}

// TestRef_MarshalJSON tests the JSON marshaling of a Ref.
func TestRef_MarshalJSON(t *testing.T) {
	ref := mustParseRef(t, "http://example.com/a?b#c")
	jsonData, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	expected := `"http://example.com/a?b#c"`
	if string(jsonData) != expected {
		t.Errorf("Expected JSON string '%s', got '%s'", expected, string(jsonData))
	}
}

// TestRef_UnmarshalJSON tests the JSON unmarshaling of a Ref.
func TestRef_UnmarshalJSON(t *testing.T) {
	t.Run("Valid IRI", func(t *testing.T) {
		var ref Ref
		jsonData := []byte(`"http://example.com/a?b#c"`)
		err := json.Unmarshal(jsonData, &ref)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
		expected := "http://example.com/a?b#c"
		if ref.String() != expected {
			t.Errorf("Expected unmarshaled string '%s', got '%s'", expected, ref.String())
		}
	})

	t.Run("Invalid IRI", func(t *testing.T) {
		var ref Ref
		// Use an unambiguously invalid IRI syntax.
		jsonData := []byte(`"http://example.com/["`)
		err := json.Unmarshal(jsonData, &ref)
		if err == nil {
			t.Fatal("Expected an error for invalid IRI, but got none")
		}
		if !strings.Contains(err.Error(), "Invalid IRI character") {
			t.Errorf("Expected error message to contain 'Invalid IRI character', got '%s'", err.Error())
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		var ref Ref
		jsonData := []byte(`not-a-string`)
		err := json.Unmarshal(jsonData, &ref)
		if err == nil {
			t.Fatal("Expected an error for invalid JSON, but got none")
		}
	})
}

// TestParseRef_Valid tests parsing of various valid IRI-references.
func TestParseRef_Valid(t *testing.T) {
	// RFC 3986 & 3987 define the generic syntax for URI-reference and IRI-reference.
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
			if ref == nil {
				t.Fatal("Expected a non-nil Ref, but got nil")
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

// TestParseNormalizedRef tests that parsing a reference with this function results in an NFC-normalized string.
func TestParseNormalizedRef(t *testing.T) {
	// RFC 3987, Section 5.3.2.2 discusses character normalization (NFC).
	// "when a resource is created, its IRI should be as character normalized as possible (i.e., NFC...)"
	decomposed := "e\u0301" // e + combining acute accent
	composed := "\u00e9"    // é (precomposed)

	if !norm.NFC.IsNormalString(composed) {
		t.Fatalf("Test setup error: composed string '%s' is not in NFC", composed)
	}
	if norm.NFC.IsNormalString(decomposed) {
		t.Fatalf("Test setup error: decomposed string '%s' is in NFC", decomposed)
	}

	iriStr := "http://example.com/" + decomposed
	ref, err := ParseNormalizedRef(iriStr)
	if err != nil {
		t.Fatalf("ParseNormalizedRef failed: %v", err)
	}

	expectedStr := "http://example.com/" + composed
	if ref.String() != expectedStr {
		t.Errorf("Expected IRI string to be normalized to NFC '%s', got '%s'", expectedStr, ref.String())
	}

	// Test error case
	// Per RFC 3986 Section 4.2, a colon in the first path segment of a relative reference is not allowed.
	// Since "1:b" cannot be a scheme, it is parsed as a relative path and fails.
	_, err = ParseNormalizedRef("1:b")
	if err == nil {
		t.Fatal("Expected an error for invalid IRI, but got none")
	}
}

// TestParseURIToRef tests the conversion from a URI string (with percent-encoding) to an IRI Ref.
func TestParseURIToRef(t *testing.T) {
	// RFC 3987, Section 3.2: Converting URIs to IRIs
	testCases := []struct {
		name     string
		uri      string
		expected string
		hasError bool
	}{
		{
			name:     "Valid UTF-8 sequence",
			uri:      "http://example.org/D%C3%BCrst", // Dürst
			expected: "http://example.org/Dürst",
			hasError: false,
		},
		{
			name:     "Valid URI with non-UTF8 percent encoding",
			uri:      "http://example.org/%FCrst", // ü in latin1. Preserved correctly.
			expected: "http://example.org/%FCrst",
			hasError: false,
		},
		{
			name:     "Forbidden Bidi character",
			uri:      "http://example.com/%E2%80%AE", // U+202E RTL OVERRIDE
			expected: "http://example.com/%E2%80%AE", // Should remain encoded
			hasError: false,
		},
		{
			name:     "Malformed URI with incomplete percent encoding",
			uri:      "http://example.com/%C", // Correctly malformed input
			expected: "",                      // Expected string is irrelevant on error
			hasError: true,
		},
		{
			name:     "Malformed URI with invalid percent encoding",
			uri:      "http://example.com/foo%GGbar",
			expected: "", // Expected string is irrelevant on error
			hasError: true,
		},
		{
			name:     "Mixed valid and invalid sequences",
			uri:      "/a%C3%A9b%E9c/",
			expected: "/aéb%E9c/",
			hasError: false,
		},
		{
			name:     "Invalid decoded IRI",
			uri:      "a%3A/b", // decodes to "a:/b", which could be parsed as scheme:path-absolute
			expected: "a:/b",   // but it is not a valid relative reference
			hasError: false,    // NOTE: This behavior depends on re-parsing. The decoded string is a valid IRI-reference.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref, err := ParseURIToRef(tc.uri)
			if tc.hasError {
				if err == nil {
					t.Fatal("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
				if ref.String() != tc.expected {
					t.Errorf("Expected converted IRI '%s', got '%s'", tc.expected, ref.String())
				}
			}
		})
	}
}

// TestRef_ToURI tests the conversion from an IRI Ref back to a URI string, including IDNA and percent-encoding.
func TestRef_ToURI(t *testing.T) {
	// Based on RFC 3987, Section 3.1: Mapping of IRIs to URIs.
	// It requires NFC normalization, percent-encoding of non-ASCII, and IDNA for hosts.
	testCases := []struct {
		name     string
		iri      string
		expected string
	}{
		{
			"Simple ASCII IRI",
			"http://example.com/a/b",
			"http://example.com/a/b",
		},
		{
			"Non-ASCII path",
			"http://example.com/résumé",
			"http://example.com/r%C3%A9sum%C3%A9",
		},
		{
			"Non-ASCII query",
			"http://example.com/?p=résumé",
			"http://example.com/?p=r%C3%A9sum%C3%A9",
		},
		{
			"Non-ASCII fragment",
			"http://example.com/#résumé",
			"http://example.com/#r%C3%A9sum%C3%A9",
		},
		{
			"Non-ASCII userinfo",
			"ftp://résumé@example.com/",
			"ftp://r%C3%A9sum%C3%A9@example.com/",
		},
		{
			"IDNA host",
			"http://résumé.example.org/",
			"http://xn--rsum-bpad.example.org/",
		},
		{
			"Full IRI with all parts",
			"http://user:p@résumé.com:8080/p?q=v#f",
			"http://user:p@xn--rsum-bpad.com:8080/p?q=v#f",
		},
		{
			"IDNA handling of leading hyphen non-ASCII host",
			"http://-résumé.com/", // Hyphen at start is valid for the Go IDNA library
			"http://xn---rsum-csad.com/",
		},
		{
			"IDNA handling of long label (produces valid but long punycode)",
			"http://" + strings.Repeat("a", 63) + ".com/",
			"http://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.com/",
		},
		{
			"IDNA failure fallback with percent-encoded space in host",
			"http://a%20b.com/", // A valid IRI, but host is invalid for IDNA
			"http://a%20b.com/", // idna.ToASCII fails, fallback is a no-op as there are no non-ASCII chars
		},
		{
			"NFC normalization before encoding",
			"http://example.com/e\u0301", // non-NFC 'é'
			"http://example.com/%C3%A9",  // NFC 'é' percent-encoded
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref := mustParseRef(t, tc.iri)
			uri := ref.ToURI()
			if uri != tc.expected {
				t.Errorf("Expected URI '%s', got '%s'", tc.expected, uri)
			}
		})
	}
}

// TestRef_Normalize tests the syntax-based and scheme-based normalization of a Ref.
func TestRef_Normalize(t *testing.T) {
	// Based on RFC 3986, Section 6.2.2 and 6.2.3: Syntax-Based and Scheme-Based Normalization.
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"Case normalization (scheme, host)",
			"HTTP://User@Example.COM/Path",
			"http://User@example.com/Path",
		},
		{
			"Percent-encoding normalization (decode unreserved)",
			"http://example.com/%7Euser",
			"http://example.com/~user",
		},
		{
			"Path segment normalization (remove dot segments)",
			"http://example.com/a/b/../c/./d",
			"http://example.com/a/c/d",
		},
		{
			"Scheme-based: add / for empty path with authority",
			"http://example.com",
			"http://example.com/",
		},
		{
			"Scheme-based: remove default port",
			"http://example.com:80/path",
			"http://example.com/path",
		},
		{
			"Scheme-based: keep non-default port",
			"http://example.com:8080/path",
			"http://example.com:8080/path",
		},
		{
			"NFC normalization",
			"http://example.com/re\u0301sume\u0301.html",
			"http://example.com/résumé.html",
		},
		{
			"Combination of normalizations",
			"HTTP://EXAMPLE.COM:80/a/../b/%7E",
			"http://example.com/b/~",
		},
		{"Empty IRI", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref := mustParseRef(t, tc.input)
			normalizedRef := ref.Normalize()
			if normalizedRef.String() != tc.expected {
				t.Errorf("Expected normalized IRI '%s', got '%s'", tc.expected, normalizedRef.String())
			}
		})
	}

	t.Run("No-op returns same instance", func(t *testing.T) {
		iriStr := "http://example.com/already/normalized"
		ref := mustParseRef(t, iriStr)
		normalizedRef := ref.Normalize()
		if ref != normalizedRef {
			t.Error("Should return same instance if already normalized")
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

// TestRef_ResolveTo tests the optimized resolution of a relative IRI reference to a strings.Builder.
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

	// Manually extract components using the Positions struct
	var scheme, authority, path, query, fragment string
	var hasScheme, hasAuthority, hasQuery, hasFragment bool

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

	if pos.PathEnd < pos.QueryEnd {
		query = resolvedStr[pos.PathEnd+1 : pos.QueryEnd]
		hasQuery = true
	}

	if pos.QueryEnd < len(resolvedStr) {
		fragment = resolvedStr[pos.QueryEnd+1:]
		hasFragment = true
	}

	// Assertions on extracted components
	if !hasScheme {
		t.Errorf("Expected scheme to be present")
	}
	if scheme != "http" {
		t.Errorf("Expected scheme 'http', got '%s'", scheme)
	}
	if !hasAuthority {
		t.Errorf("Expected authority to be present")
	}
	if authority != "a" {
		t.Errorf("Expected authority 'a', got '%s'", authority)
	}
	if path != "/b/g" {
		t.Errorf("Expected path '/b/g', got '%s'", path)
	}
	if hasQuery {
		t.Errorf("Expected query to be absent, but got '%s'", query)
	}
	if hasFragment {
		t.Errorf("Expected fragment to be absent, but got '%s'", fragment)
	}

	// Check error case
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

// TestNewIriFromRef tests the creation of an Iri from a Ref, ensuring it handles absolute and relative refs correctly.
func TestNewIriFromRef(t *testing.T) {
	t.Run("Absolute Ref", func(t *testing.T) {
		ref := mustParseRef(t, "http://example.com")
		iri, err := NewIriFromRef(ref)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		if iri == nil {
			t.Fatal("Expected a non-nil Iri, but got nil")
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

// TestParseIri tests that parsing requires an absolute IRI and fails for relative references.
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

// TestParseNormalizedIri tests parsing an absolute IRI with NFC normalization.
func TestParseNormalizedIri(t *testing.T) {
	decomposed := "e\u0301" // e + combining acute accent
	composed := "\u00e9"    // é (precomposed)
	iriStr := "http://example.com/" + decomposed
	iri, err := ParseNormalizedIri(iriStr)
	if err != nil {
		t.Fatalf("ParseNormalizedIri failed: %v", err)
	}

	expectedStr := "http://example.com/" + composed
	if iri.String() != expectedStr {
		t.Errorf("Expected IRI string to be normalized to NFC '%s', got '%s'", expectedStr, iri.String())
	}

	// Test error cases
	_, err = ParseNormalizedIri("/relative")
	if err == nil {
		t.Fatal("Expected an error for relative IRI, but got none")
	}
	_, err = ParseNormalizedIri("1:b")
	if err == nil {
		t.Fatal("Expected an error for invalid IRI, but got none")
	}
}

// TestIri_Scheme tests the Scheme accessor for the Iri type.
func TestIri_Scheme(t *testing.T) {
	iri := mustParseIri(t, "https://example.com")
	if s := iri.Scheme(); s != "https" {
		t.Errorf("Expected scheme 'https', got '%s'", s)
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
	// Error case
	_, err = iri.Resolve("1:b")
	if err == nil {
		t.Fatal("Expected an error for invalid relative ref, but got none")
	}
}

// TestIri_ResolveTo tests the optimized resolution of a relative IRI reference against a base Iri to a strings.Builder.
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
	// Error case
	var errBuilder strings.Builder
	err = iri.ResolveTo("1:b", &errBuilder)
	if err == nil {
		t.Fatal("Expected an error for invalid relative ref, but got none")
	}
}

// TestIri_MarshalJSON tests the JSON marshaling of an Iri.
func TestIri_MarshalJSON(t *testing.T) {
	iri := mustParseIri(t, "http://example.com/a")
	jsonData, err := json.Marshal(iri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	expected := `"http://example.com/a"`
	if string(jsonData) != expected {
		t.Errorf("Expected JSON string '%s', got '%s'", expected, string(jsonData))
	}
}

// TestIri_UnmarshalJSON tests the JSON unmarshaling of an Iri, including validation.
func TestIri_UnmarshalJSON(t *testing.T) {
	t.Run("Valid Absolute IRI", func(t *testing.T) {
		var iri Iri
		jsonData := []byte(`"http://example.com"`)
		err := json.Unmarshal(jsonData, &iri)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
		if iri.String() != "http://example.com" {
			t.Errorf("Expected unmarshaled string 'http://example.com', got '%s'", iri.String())
		}
	})

	t.Run("Relative IRI", func(t *testing.T) {
		var iri Iri
		jsonData := []byte(`"/relative/path"`)
		err := json.Unmarshal(jsonData, &iri)
		if err == nil {
			t.Fatal("Expected an error for relative IRI, but got none")
		}
		if !strings.Contains(err.Error(), "No scheme found") {
			t.Errorf("Expected error message to contain 'No scheme found', got '%s'", err.Error())
		}
	})

	t.Run("Invalid IRI", func(t *testing.T) {
		var iri Iri
		jsonData := []byte(`"http://["`)
		err := json.Unmarshal(jsonData, &iri)
		if err == nil {
			t.Fatal("Expected an error for invalid IRI, but got none")
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		var iri Iri
		err := iri.UnmarshalJSON([]byte("not-json"))
		if err == nil {
			t.Fatal("Expected an error for invalid JSON, but got none")
		}
	})
}

// TestIri_Relativize_Valid tests the process of creating a valid relative reference.
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
