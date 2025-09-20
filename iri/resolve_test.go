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
	"strings"
	"testing"
)

// newTestParserWithBase is a helper to create an iriParser with a pre-parsed base IRI.
// This is used in tests that require a base IRI context for resolution.
func newTestParserWithBase(t *testing.T, baseIRI string) *iriParser {
	// We need to run a parse on the base to get its component positions.
	// We use an unchecked parse as the base is assumed to be valid for tests.
	pos, err := run(baseIRI, nil, true, &voidOutputBuffer{})
	if err != nil {
		t.Fatalf("Failed to parse base IRI for test setup: %s, error: %v", baseIRI, err)
	}

	b := &iriParserBase{
		iri:          baseIRI,
		schemeEnd:    pos.SchemeEnd,
		authorityEnd: pos.AuthorityEnd,
		pathEnd:      pos.PathEnd,
		queryEnd:     pos.QueryEnd,
		hasBase:      true,
	}

	return &iriParser{base: b}
}

// TestIsValidRefScheme tests the scheme validation logic.
// RFC 3986, Section 3.1: scheme = ALPHA *( ALPHA / DIGIT / "+" / "-" / "." ).
func TestIsValidRefScheme(t *testing.T) {
	tests := []struct {
		name       string
		schemePart string
		want       bool
	}{
		{"Valid lowercase", "http", true},
		{"Valid with digits", "h1", true},
		{"Valid with plus", "foo+bar", true},
		{"Valid with hyphen", "foo-bar", true},
		{"Valid with period", "foo.bar", true},
		{"Valid single letter", "a", true},
		{"Invalid empty", "", false},
		{"Invalid starts with digit", "1http", false},
		{"Invalid starts with hyphen", "-foo", false},
		{"Invalid contains illegal char", "ht:tp", false},
		{"Invalid contains space", "ht tp", false},
		{"Invalid UTF-8 char", "schème", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidRefScheme(tt.schemePart); got != tt.want {
				t.Errorf("isValidRefScheme(%q) = %v, want %v", tt.schemePart, got, tt.want)
			}
		})
	}
}

// TestExtractRefScheme tests the extraction of a scheme from a reference string.
// It relies on the scheme syntax defined in RFC 3986, Section 3.1.
func TestExtractRefScheme(t *testing.T) {
	tests := []struct {
		name       string
		ref        string
		wantScheme string
		wantRest   string
		wantOK     bool
	}{
		{"Valid scheme", "http://example.com", "http", "//example.com", true},
		{"URN scheme", "urn:isbn:12345", "urn", "isbn:12345", true},
		{"No scheme (network-path)", "//example.com", "", "//example.com", false},
		{"No scheme (absolute-path)", "/path/to/file", "", "/path/to/file", false},
		{"No scheme (relative-path)", "path/to/file", "", "path/to/file", false},
		{"Invalid scheme syntax", "1http:rest", "", "1http:rest", false},
		{"Empty scheme part", ":rest", "", ":rest", false},
		{"No colon", "example.com", "", "example.com", false},
		{"Empty string", "", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotScheme, gotRest, gotOK := extractRefScheme(tt.ref)
			if gotScheme != tt.wantScheme || gotRest != tt.wantRest || gotOK != tt.wantOK {
				t.Errorf("extractRefScheme(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.ref, gotScheme, gotRest, gotOK, tt.wantScheme, tt.wantRest, tt.wantOK)
			}
		})
	}
}

// TestDeconstructRef tests the pre-parsing of a reference string into its components.
// This decomposition follows the generic syntax defined in RFC 3986, Section 3.
func TestDeconstructRef(t *testing.T) {
	type result struct {
		scheme       string
		authority    string
		path         string
		query        string
		fragment     string
		hasAuthority bool
		hasQuery     bool
		hasFragment  bool
	}
	tests := []struct {
		name string
		ref  string
		want result
	}{
		{"Full reference", "http://a/b?c#d", result{"http", "a", "/b", "c", "d", true, true, true}},
		{"Network-path reference", "//a/b?c#d", result{"", "a", "/b", "c", "d", true, true, true}},
		{"Absolute-path reference", "/b?c#d", result{"", "", "/b", "c", "d", false, true, true}},
		{"Relative-path reference", "b?c#d", result{"", "", "b", "c", "d", false, true, true}},
		{"Scheme and path", "mailto:user@host", result{"mailto", "", "user@host", "", "", false, false, false}},
		{"Path only", "b", result{"", "", "b", "", "", false, false, false}},
		{"Query only", "?c", result{"", "", "", "c", "", false, true, false}},
		{"Fragment only", "#d", result{"", "", "", "", "d", false, false, true}},
		{"Empty reference", "", result{"", "", "", "", "", false, false, false}},
		{"Empty authority", "//?q", result{"", "", "", "q", "", true, true, false}},
		{"Authority, no path", "//a", result{"", "a", "", "", "", true, false, false}},
		{"Empty query", "path?", result{"", "", "path", "", "", false, true, false}},
		{"Empty fragment", "path#", result{"", "", "path", "", "", false, false, true}},
		{"Empty query and fragment", "path?#", result{"", "", "path", "", "", false, true, true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, a, p, q, f, ha, hq, hf := deconstructRef(tt.ref)
			got := result{s, a, p, q, f, ha, hq, hf}
			if got != tt.want {
				t.Errorf("deconstructRef(%q) incorrect component breakdown\ngot:  %+v\nwant: %+v", tt.ref, got, tt.want)
			}
		})
	}
}

// TestGetBaseComponents tests the extraction of components from a pre-parsed base IRI.
func TestGetBaseComponents(t *testing.T) {
	tests := []struct {
		name             string
		baseIRI          string
		wantScheme       string
		wantAuthority    string
		wantPath         string
		wantHasAuthority bool
		wantQuery        string
		wantHasQuery     bool
	}{
		{"Full base", "http://a/b?c", "http", "a", "/b", true, "c", true},
		{"No authority", "urn:foo:bar", "urn", "", "foo:bar", false, "", false},
		{"No query", "http://example.com/path", "http", "example.com", "/path", true, "", false},
		{"No path", "http://example.com", "http", "example.com", "", true, "", false},
		{"Scheme only", "foo:", "foo", "", "", false, "", false},
		{"Path is root", "http://a/", "http", "a", "/", true, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParserWithBase(t, tt.baseIRI)
			scheme, authority, path, hasAuthority, query, hasQuery := p.getBaseComponents()
			if scheme != tt.wantScheme {
				t.Errorf("getBaseComponents() scheme = %q, want %q", scheme, tt.wantScheme)
			}
			if authority != tt.wantAuthority {
				t.Errorf("getBaseComponents() authority = %q, want %q", authority, tt.wantAuthority)
			}
			if path != tt.wantPath {
				t.Errorf("getBaseComponents() path = %q, want %q", path, tt.wantPath)
			}
			if hasAuthority != tt.wantHasAuthority {
				t.Errorf("getBaseComponents() hasAuthority = %v, want %v", hasAuthority, tt.wantHasAuthority)
			}
			if query != tt.wantQuery {
				t.Errorf("getBaseComponents() query = %q, want %q", query, tt.wantQuery)
			}
			if hasQuery != tt.wantHasQuery {
				t.Errorf("getBaseComponents() hasQuery = %v, want %v", hasQuery, tt.wantHasQuery)
			}
		})
	}
}

// TestResolvePathAndQuery tests the path and query resolution logic from RFC 3986, Section 5.2.2.
func TestResolvePathAndQuery(t *testing.T) {
	p := &iriParser{} // The method does not depend on parser state, only its arguments.

	tests := []struct {
		name             string
		rPath            string
		rQuery           string
		rHasQuery        bool
		basePath         string
		baseQuery        string
		hasBaseQuery     bool
		hasBaseAuthority bool
		wantPath         string
		wantQuery        string
		wantHasQuery     bool
	}{
		{"Ref path is absolute", "/g", "y", true, "/a/b", "x", true, true, "/g", "y", true},
		{"Ref path is relative", "g", "y", true, "/a/b", "x", true, true, "/a/g", "y", true},
		{"Base has authority, no path", "g", "", false, "", "x", true, true, "/g", "", false},
		{"Ref path is empty, ref has query", "", "y", true, "/a/b", "x", true, true, "/a/b", "y", true},
		{"Ref path is empty, ref has no query", "", "", false, "/a/b", "x", true, true, "/a/b", "x", true},
		{"Ref path empty, no queries", "", "", false, "/a/b", "", false, true, "/a/b", "", false},
		{"Ref path is empty, base has empty query", "", "", false, "/a/b", "", true, true, "/a/b", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := &resolvedIRI{}
			p.resolvePathAndQuery(
				target, tt.rPath, tt.rQuery, tt.rHasQuery,
				tt.basePath, tt.baseQuery, tt.hasBaseQuery, tt.hasBaseAuthority,
			)
			if target.Path != tt.wantPath {
				t.Errorf("resolvePathAndQuery() path = %q, want %q", target.Path, tt.wantPath)
			}
			if target.Query != tt.wantQuery {
				t.Errorf("resolvePathAndQuery() query = %q, want %q", target.Query, tt.wantQuery)
			}
			if target.HasQuery != tt.wantHasQuery {
				t.Errorf("resolvePathAndQuery() hasQuery = %v, want %v", target.HasQuery, tt.wantHasQuery)
			}
		})
	}
}

// TestResolveComponents tests the main resolution algorithm against examples from RFC 3986.
// The tests are based on Section 5.4 and 5.2, using base URI "http://a/b/c/d;p?q".
func TestResolveComponents(t *testing.T) {
	baseIRI := "http://a/b/c/d;p?q"
	p := newTestParserWithBase(t, baseIRI)

	tests := []struct {
		name         string
		relativeRef  string
		wantAuth     string
		wantPath     string
		wantQuery    string
		wantFragment string
	}{
		// Normal Examples from RFC 3986, Section 5.2.2
		{"Absolute IRI reference", "ftp://example.net/file?query#frag", "example.net", "/file", "query", "frag"},
		{"Absolute IRI with dot segments", "http://c/d/../e", "c", "/e", "", ""},
		// Normal Examples from RFC 3986, Section 5.4.1
		{"Normal: g", "g", "a", "/b/c/g", "", ""},
		{"Normal: ./g", "./g", "a", "/b/c/g", "", ""},
		{"Normal: g/", "g/", "a", "/b/c/g/", "", ""},
		{"Normal: /g", "/g", "a", "/g", "", ""},
		{"Normal: //g", "//g", "g", "", "", ""},
		{"Normal: ?y", "?y", "a", "/b/c/d;p", "y", ""},
		{"Normal: g?y", "g?y", "a", "/b/c/g", "y", ""},
		{"Normal: #s", "#s", "a", "/b/c/d;p", "q", "s"},
		{"Normal: g#s", "g#s", "a", "/b/c/g", "", "s"},
		{"Normal: g?y#s", "g?y#s", "a", "/b/c/g", "y", "s"},
		{"Normal: empty", "", "a", "/b/c/d;p", "q", ""},
		{"Normal: .", ".", "a", "/b/c/", "", ""},
		{"Normal: ./", "./", "a", "/b/c/", "", ""},
		{"Normal: ..", "..", "a", "/b/", "", ""},
		{"Normal: ../", "../", "a", "/b/", "", ""},
		{"Normal: ../g", "../g", "a", "/b/g", "", ""},
		{"Normal: ../..", "../..", "a", "/", "", ""},
		{"Normal: ../../", "../../", "a", "/", "", ""},
		{"Normal: ../../g", "../../g", "a", "/g", "", ""},
		// Abnormal Examples from RFC 3986, Section 5.4.2
		{"Abnormal: ../../../g", "../../../g", "a", "/g", "", ""},
		{"Abnormal: ../../../../g", "../../../../g", "a", "/g", "", ""},
		{"Abnormal: /./g", "/./g", "a", "/g", "", ""},
		{"Abnormal: /../g", "/../g", "a", "/g", "", ""},
		{"Abnormal: g.", "g.", "a", "/b/c/g.", "", ""},
		{"Abnormal: .g", ".g", "a", "/b/c/.g", "", ""},
		{"Abnormal: g..", "g..", "a", "/b/c/g..", "", ""},
		{"Abnormal: ..g", "..g", "a", "/b/c/..g", "", ""},
		{"Abnormal: ./../g", "./../g", "a", "/b/g", "", ""},
		{"Abnormal: ./g/.", "./g/.", "a", "/b/c/g/", "", ""},
		{"Abnormal: g/./h", "g/./h", "a", "/b/c/g/h", "", ""},
		{"Abnormal: g/../h", "g/../h", "a", "/b/c/h", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.resolveComponents(tt.relativeRef)
			if result.Authority != tt.wantAuth {
				t.Errorf("resolveComponents(%q) Authority = %q, want %q", tt.relativeRef, result.Authority, tt.wantAuth)
			}
			if result.Path != tt.wantPath {
				t.Errorf("resolveComponents(%q) Path = %q, want %q", tt.relativeRef, result.Path, tt.wantPath)
			}
			if result.Query != tt.wantQuery {
				t.Errorf("resolveComponents(%q) Query = %q, want %q", tt.relativeRef, result.Query, tt.wantQuery)
			}
			if result.Fragment != tt.wantFragment {
				t.Errorf(
					"resolveComponents(%q) Fragment = %q, want %q",
					tt.relativeRef, result.Fragment, tt.wantFragment,
				)
			}
		})
	}
}

// TestRecomposeIRI tests the assembly of an IRI from its components.
// The recomposition algorithm is defined in RFC 3986, Section 5.3.
func TestRecomposeIRI(t *testing.T) {
	tests := []struct {
		name string
		t    *resolvedIRI
		want string
	}{
		{"Full IRI", &resolvedIRI{"http", "a", "/b", "c", "d", true, true, true}, "http://a/b?c#d"},
		{"No authority", &resolvedIRI{"urn", "", "foo:bar", "c", "d", false, true, true}, "urn:foo:bar?c#d"},
		{"No query", &resolvedIRI{"http", "a", "/b", "", "d", true, false, true}, "http://a/b#d"},
		{"No fragment", &resolvedIRI{"http", "a", "/b", "c", "", true, true, false}, "http://a/b?c"},
		{"Scheme relative", &resolvedIRI{"", "a", "/b", "c", "d", true, true, true}, "//a/b?c#d"},
		{"Empty query part", &resolvedIRI{"http", "a", "/b", "", "d", true, true, true}, "http://a/b?#d"},
		{"Empty fragment part", &resolvedIRI{"http", "a", "/b", "c", "", true, true, true}, "http://a/b?c#"},
		{"Empty path", &resolvedIRI{"http", "a", "", "", "", true, false, false}, "http://a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &iriParser{output: &stringOutputBuffer{builder: &strings.Builder{}}}
			p.recomposeIRI(tt.t)
			got := p.output.string()
			if got != tt.want {
				t.Errorf("recomposeIRI() got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestValidateRelativeRef tests the sub-parser validation for a relative reference.
// The test cases check for syntax that would be invalid in a URI-reference.
func TestValidateRelativeRef(t *testing.T) {
	// This function requires a full parser instance, but it's defined in resolve.go
	// and is a key part of the resolution process.
	p := &iriParser{
		base: &iriParserBase{hasBase: false},
	}

	tests := []struct {
		name        string
		relativeRef string
		expectError bool
	}{
		{"Valid relative-path", "a/b", false},
		{"Valid absolute-path", "/a/b", false},
		{"Valid network-path", "//a/b", false},
		{"Valid query", "?q", false},
		{"Valid fragment", "#f", false},
		{"Valid empty", "", false},
		{"Invalid char in path", "a[b", true},
		{"Invalid percent encoding", "%GG", true},
		{"Invalid: colon in first segment", "a:b", true},
		{"Valid: colon not in first segment", "a/b:c", false},
		{"Valid: dot-segment with colon", "./a:b", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.validateRelativeRef(tt.relativeRef)
			if (err != nil) != tt.expectError {
				t.Errorf("validateRelativeRef(%q) returned error %v, expectError=%v",
					tt.relativeRef, err, tt.expectError)
			}
		})
	}
}

// TestParseRelative tests the integrated resolution process.
func TestParseRelative(t *testing.T) {
	baseIRI := "http://a/b/c/d;p?q"
	tests := []struct {
		name        string
		relativeRef string
		wantIRI     string
		expectError bool
	}{
		{"Ref with path", "g", "http://a/b/c/g", false},
		{"Ref with absolute path", "/g", "http://a/g", false},
		{"Ref with network path", "//g", "http://g", false},
		{"Ref with query", "?y", "http://a/b/c/d;p?y", false},
		{"Ref with fragment", "#s", "http://a/b/c/d;p?q#s", false},
		{"Ref with path, query, fragment", "../g?y#s", "http://a/b/g?y#s", false},
		{"Invalid ref syntax", "a[b", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup parser for each run.
			p := newTestParserWithBase(t, baseIRI)
			p.input = newParserInput(tt.relativeRef)
			p.output = &stringOutputBuffer{builder: &strings.Builder{}}

			err := p.parseRelative()

			if (err != nil) != tt.expectError {
				t.Errorf("parseRelative() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError {
				gotIRI := p.output.string()
				if gotIRI != tt.wantIRI {
					t.Errorf("parseRelative() resolution failed\nBase: %s\nRef:  %q\nGot:  %q\nWant: %q",
						baseIRI, tt.relativeRef, gotIRI, tt.wantIRI)
				}
			}
		})
	}

	t.Run("No base", func(_ *testing.T) {
		p := &iriParser{
			base:   &iriParserBase{hasBase: false},
			input:  newParserInput("a/b"),
			output: &stringOutputBuffer{builder: &strings.Builder{}},
		}
		_ = p.parseRelativeNoBase()
	})
}

// TestRecomposeNormalizedIRI tests the recomposition of an IRI from normalized components.
// This function is similar to `recomposeIRI` but is a standalone utility.
func TestRecomposeNormalizedIRI(t *testing.T) {
	tests := []struct {
		name      string
		scheme    string
		hasScheme bool
		userinfo  string
		host      string
		port      string
		hasAuth   bool
		path      string
		query     string
		hasQuery  bool
		fragment  string
		hasFrag   bool
		want      string
	}{
		{"Full", "http", true, "user", "host", "80", true, "/p", "q", true, "f", true, "http://user@host:80/p?q#f"},
		{"No user/port", "http", true, "", "host", "", true, "/p", "q", true, "f", true, "http://host/p?q#f"},
		{"Scheme relative", "", false, "", "host", "", true, "/p", "", false, "", false, "//host/p"},
		{"No authority", "urn", true, "", "", "", false, "a:b", "", false, "", false, "urn:a:b"},
		{"Path only", "", false, "", "", "", false, "/p", "", false, "", false, "/p"},
		{"Empty query", "http", true, "", "h", "", true, "/p", "", true, "", false, "http://h/p?"},
		{"Empty fragment", "http", true, "", "h", "", true, "/p", "", false, "", true, "http://h/p#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := recomposeNormalizedIRI(
				tt.scheme, tt.hasScheme,
				tt.userinfo, tt.host, tt.port, tt.hasAuth,
				tt.path,
				tt.query, tt.hasQuery,
				tt.fragment, tt.hasFrag,
			)
			if got != tt.want {
				t.Errorf("recomposeNormalizedIRI() got %q, want %q", got, tt.want)
			}
		})
	}
}
