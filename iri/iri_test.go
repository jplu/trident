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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// fst is a generic helper function that returns the first of two arguments.
// Useful for unwrapping functions that return a value and a boolean, like component accessors.
func fst[T, U any](val T, _ U) T {
	return val
}

// snd is a generic helper function that returns the second of two arguments.
// Useful for unwrapping functions that return a value and a boolean.
func snd[T, U any](_ T, val U) U {
	return val
}

// TestParsing verifies that various valid absolute IRIs can be parsed correctly by both
// the checked (ParseIri) and unchecked (ParseIriUnchecked) functions. It also compares
// the results of the component accessor methods (Scheme, Authority, etc.) from both
// parsing methods to ensure they are consistent.
func TestParsing(t *testing.T) {
	t.Parallel()
	examples := []string{
		"file://foo",
		"ftp://ftp.is.co.za/rfc/rfc1808.txt",
		"http://www.ietf.org/rfc/rfc2396.txt",
		"ldap://[2001:db8::7]/c=GB?objectClass?one",
		"mailto:John.Doe@example.com",
		"news:comp.infosystems.www.servers.unix",
		"tel:+1-816-555-1212",
		"telnet://192.0.2.16:80/",
		"urn:oasis:names:specification:docbook:dtd:xml:4.1.2",
		"http://example.com",
		"http://example.com/",
		"http://example.com/foo",
		"http://example.com/foo/bar",
		"http://example.com/foo/bar/",
		"http://example.com/foo/bar?q=1&r=2",
		"http://example.com/foo/bar/?q=1&r=2",
		"http://example.com#toto",
		"http://example.com/#toto",
		"http://example.com/foo#toto",
		"http://example.com/foo/bar#toto",
		"http://example.com/foo/bar/#toto",
		"http://example.com/foo/bar?q=1&r=2#toto",
		"http://example.com/foo/bar/?q=1&r=2#toto",
		// Test with various iunreserved characters
		"http://a.example/AZaz\u00C0\u00D6\u00D8\u00F6\u00F8\u02FF\u0370\u037D\u037F\u1FFF\u200C\u200D\u2070\u218F\u2C00\u2FEF\u3001\uD7FF\uFA0E\uFDCF\uFDF0\uFFEF\U00010000\U000EFFFD",
		"http://a.example/?AZaz\uE000\uF8FF\U000F0000\U000FFFFD\U00100000\U0010FFFD\u00C0\u00D6\u00D8\u00F6\u00F8\u02FF\u0370\u037D\u037F\u1FFF\u200C\u200D\u2070\u218F\u2C00\u2FEF\u3001\uD7FF\uFA0E\uFDCF\uFDF0\uFFEF\U00010000\U000EFFFD",
		// Test IPvFuture literals
		"http://[va.12z]",
		"http://[vff.B]",
		"http://[V0.a]",
	}

	for _, e := range examples {
		t.Run(e, func(t *testing.T) {
			checkParsing(t, e)
		})
	}
}

// checkParsing is a helper function for TestParsing. It performs the actual checks
// for a single IRI string.
func checkParsing(t *testing.T, e string) {
	t.Helper()
	t.Parallel()
	unchecked := ParseIriUnchecked(e)
	if unchecked.String() != e {
		t.Errorf("ParseIriUnchecked returned %q, want %q", unchecked.String(), e)
	}

	iri, err := ParseIri(e)
	if err != nil {
		t.Fatalf("ParseIri failed for %q: %v", e, err)
	}
	if unchecked.String() != iri.String() {
		t.Errorf("Unchecked %q != Checked %q", unchecked.String(), iri.String())
	}

	// Compare components
	if sch1 := unchecked.Scheme(); sch1 != iri.Scheme() {
		t.Errorf("Scheme mismatch: unchecked %q, checked %q", sch1, iri.Scheme())
	}
	if auth1, ok1 := unchecked.Authority(); auth1 != fst(iri.Authority()) || ok1 != snd(iri.Authority()) {
		t.Errorf("Authority mismatch: unchecked %q, checked %q", auth1, fst(iri.Authority()))
	}
	if path1 := unchecked.Path(); path1 != iri.Path() {
		t.Errorf("Path mismatch: unchecked %q, checked %q", path1, iri.Path())
	}
	if q1, ok1 := unchecked.Query(); q1 != fst(iri.Query()) || ok1 != snd(iri.Query()) {
		t.Errorf("Query mismatch: unchecked %q, checked %q", q1, fst(iri.Query()))
	}
	if f1, ok1 := unchecked.Fragment(); f1 != fst(iri.Fragment()) || ok1 != snd(iri.Fragment()) {
		t.Errorf("Fragment mismatch: unchecked %q, checked %q", f1, fst(iri.Fragment()))
	}
}

// TestRelativeParsing checks the parsing and resolution of various relative IRI references.
// It ensures that valid relative references are parsed correctly by both checked and
// unchecked methods and that their resolution against a base IRI produces consistent results.
func TestRelativeParsing(t *testing.T) {
	t.Parallel()
	examples := []string{
		"file:///foo/bar",
		"mailto:user@host?subject=blah",
		"dav:",
		"about:",
		"http://www.yahoo.com",
		"http://www.yahoo.com/",
		"http://1.2.3.4/",
		"http://www.yahoo.com/stuff",
		"http://www.yahoo.com/stuff/",
		"http://www.yahoo.com/hello%20world/",
		"http://www.yahoo.com?name=obi",
		"http://www.yahoo.com?name=obi+wan&status=jedi",
		"http://www.yahoo.com?onery",
		"http://www.yahoo.com#bottom",
		"http://www.yahoo.com/yelp.html#bottom",
		"https://www.yahoo.com/",
		"ftp://www.yahoo.com/",
		"ftp://www.yahoo.com/hello",
		"demo.txt",
		"demo/hello.txt",
		"demo/hello.txt?query=hello#fragment",
		"/cgi-bin/query?query=hello#fragment",
		"/demo.txt",
		"/hello/demo.txt",
		"hello/demo.txt",
		"/",
		"",
		"#",
		"#here",
		"http://www.yahoo.com?name=%00%01",
		"http://www.yaho%6f.com",
		"http://www.yahoo.com/hello%00world/",
		"http://www.yahoo.com/hello+world/",
		"http://www.yahoo.com?name=obi&",
		"http://www.yahoo.com?name=obi&type=",
		"http://www.yahoo.com/yelp.html#",
		"//",
		"http://example.org/aaa/bbb#ccc",
		"mailto:local@domain.org",
		"mailto:local@domain.org#frag",
		"HTTP://EXAMPLE.ORG/AAA/BBB#CCC",
		"//example.org/aaa/bbb#ccc",
		"/aaa/bbb#ccc",
		"bbb#ccc",
		"#ccc",
		"#",
		"A'C",
		"http://example.org/aaa%2fbbb#ccc",
		"http://example.org/aaa%2Fbbb#ccc",
		"%2F",
		"?%2F",
		"#?%2F",
		"aaa%2Fbbb",
		"http://example.org:80/aaa/bbb#ccc",
		"http://example.org:/aaa/bbb#ccc",
		"http://example.org./aaa/bbb#ccc",
		"http://example.123./aaa/bbb#ccc",
		"http://example.org",
		"http://[FEDC:AA98:7654:3210:FEDC:AA98:7654:3210]:80/index.html",
		"http://[1080:0:0:0:8:800:200C:417A]/index.html",
		"http://[3ffe:2a00:100:7031::1]",
		"http://[1080::8:800:200C:417A]/foo",
		"http://[::192.9.5.5]/ipng",
		"http://[::FFFF:129.144.52.38]:80/index.html",
		"http://[2010:836B:4179::836B:4179]",
		"//[2010:836B:4179::836B:4179]",
		"http://example/Andrȷ",
		"file:///C:/DEV/Haskell/lib/HXmlToolbox-3.01/examples/",
		"http://a/?\uE000",
		"?\uE000",
	}

	base, baseErr := ParseIri("http://a/b/c/d;p?q")
	if baseErr != nil {
		t.Fatalf("Failed to parse base IRI: %v", baseErr)
	}

	for _, e := range examples {
		t.Run(e, func(t *testing.T) {
			t.Parallel()
			unchecked := ParseRefUnchecked(e)
			if unchecked.String() != e {
				t.Errorf("ParseRefUnchecked returned %q, want %q", unchecked.String(), e)
			}

			ref, err := ParseRef(e)
			if err != nil {
				t.Fatalf("ParseRef failed for %q: %v", e, err)
			}
			if unchecked.String() != ref.String() {
				t.Errorf("Unchecked %q != Checked %q", unchecked.String(), ref.String())
			}

			resolved, err := base.Resolve(e)
			if err != nil {
				t.Fatalf("base.Resolve failed for %q: %v", e, err)
			}
			resolvedUnchecked := base.ResolveUnchecked(e)
			if resolved.String() != resolvedUnchecked.String() {
				t.Errorf("Resolve gives %q, ResolveUnchecked gives %q", resolved.String(), resolvedUnchecked.String())
			}
		})
	}
}

// TestWrongRelativeParsing confirms that the parser correctly rejects a list
// of malformed IRI references, returning an error for each.
func TestWrongRelativeParsing(t *testing.T) {
	t.Parallel()
	examples := []string{
		"beepbeep\x07\x07", // Control characters
		"\n",
		"http://www yahoo.com", // Space in host
		"http://www.yahoo.com/hello world/",
		"http://www.yahoo.com/yelp.html#\"",
		"[2010:836B:4179::836B:4179]", // Missing scheme
		" ",
		"%", // Incomplete percent encoding
		"A%Z",
		"%ZZ",
		"%AZ",
		"A C", // Invalid characters
		"A`C",
		"A<C",
		"A>C",
		"A^C",
		"A\\C",
		"A{C",
		"A|C",
		"A}C",
		"A[C",
		"A]C",
		"A[**]C",
		"http://[xyz]/",
		"http://]/",
		"http://example.org/[2010:836B:4179::836B:4179]",
		"http://example.org/abc#[2010:836B:4179::836B:4179]",
		"http://example.org/xxx/[qwerty]#a[b]",
		"http://w3c.org:80path1/path2", // Missing slash after port
		":a/b",
		"http://example.com/\uE000",
		"\uE000",
		"http://example.com/#\uE000",
		"#\uE000",
		"//\uFFFF",
		"?\uFFFF",
		"/\u0000", // Invalid character ranges
		"?\u0000",
		"#\u0000",
		"/\uE000",
		"/\uF8FF",
		"/\U000F0000",
		"/\U000FFFFD",
		"/\U00100000",
		"/\U0010FFFD",
		"?\uFDEF",
		"?\uFFFF",
		"/\uFDEF",
		"/\uFFFF",
		"/\U0001FFFF",
		"/\U0002FFFF",
		"/\U0003FFFF",
		"/\U0004FFFF",
		"/\U0005FFFF",
		"/\U0006FFFF",
		"/\U0007FFFF",
		"/\U0008FFFF",
		"/\U0009FFFF",
		"/\U000AFFFF",
		"/\U000BFFFF",
		"/\U000CFFFF",
		"/\U000DFFFF",
		"/\U000EFFFF",
		"/\U000FFFFF",
		"http://[/",
		"http://[::1]a/",
		"//\u034f@[]",
		"//@@",
		"$:",
		"-:",
		":",
		"http://[]", // Invalid IPvFuture
		"http://[a]",
		"http://[vz]",
		"http://[v11]",
		"http://[v1.]",
		"http://[v.a]",
		"http://[v1.@]",
		"http://[v1.%01]",
		"//[v1.\u0582]",
	}

	base, baseErr := ParseIri("http://a/b/c/d;p?q")
	if baseErr != nil {
		t.Fatalf("Failed to parse base IRI: %v", baseErr)
	}

	for _, e := range examples {
		t.Run(e, func(t *testing.T) {
			t.Parallel()
			if _, err := base.Resolve(e); err == nil {
				t.Errorf("Expected an error for %q but got none", e)
			}
		})
	}
}

// TestWrongRelativeParsingOnScheme tests an edge case where a malformed relative
// path could be misinterpreted in the context of a scheme-only base IRI.
func TestWrongRelativeParsingOnScheme(t *testing.T) {
	t.Parallel()
	examples := []string{".///C:::"}
	base, baseErr := ParseIri("x:")
	if baseErr != nil {
		t.Fatalf("Failed to parse base IRI: %v", baseErr)
	}
	for _, e := range examples {
		t.Run(e, func(t *testing.T) {
			t.Parallel()
			if _, err := base.Resolve(e); err == nil {
				t.Errorf("Expected an error for %q but got none", e)
			}
		})
	}
}

// resolveTest is a helper struct for defining resolution test cases.
type resolveTest struct {
	name     string
	relative string
	base     string
	expected string
}

// TestResolveRelativeIRI contains a comprehensive suite of resolution tests,
// many drawn from RFC 3986 examples, to verify the correctness of the
// reference resolution algorithm.
func TestResolveRelativeIRI(t *testing.T) {
	t.Parallel()
	examples := []resolveTest{
		{
			name:     "RFC3986 Normal Example: path /.",
			relative: "/.",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/"},
		{
			name:     "RFC3986 Normal Example: path /.foo",
			relative: "/.foo",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/.foo",
		},
		{
			name:     "RFC3986 Normal Example: path .foo",
			relative: ".foo",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/.foo",
		},
		{
			name:     "RFC3986 Normal Example: new scheme",
			relative: "g:h",
			base:     "http://a/b/c/d;p?q",
			expected: "g:h",
		},
		{
			name:     "RFC3986 Normal Example: relative path",
			relative: "g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g",
		},
		{
			name:     "RFC3986 Normal Example: relative dot-slash",
			relative: "./g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g",
		},
		{
			name:     "RFC3986 Normal Example: relative path with slash",
			relative: "g/",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g/",
		},
		{
			name:     "RFC3986 Normal Example: path from root",
			relative: "/g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/g",
		},
		{
			name:     "RFC3986 Normal Example: scheme-relative",
			relative: "//g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://g",
		},
		{
			name:     "RFC3986 Normal Example: query only",
			relative: "?y",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/d;p?y",
		},
		{
			name:     "RFC3986 Normal Example: path and query",
			relative: "g?y",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g?y",
		},
		{
			name:     "RFC3986 Normal Example: fragment only",
			relative: "#s",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/d;p?q#s",
		},
		{
			name:     "RFC3986 Normal Example: path and fragment",
			relative: "g#s",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g#s",
		},
		{
			name:     "RFC3986 Normal Example: path, query, fragment",
			relative: "g?y#s",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g?y#s",
		},
		{
			name:     "RFC3986 Normal Example: path with semicolon",
			relative: ";x",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/;x",
		},
		{
			name:     "RFC3986 Normal Example: segment with semicolon",
			relative: "g;x",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g;x",
		},
		{
			name:     "RFC3986 Normal Example: all components",
			relative: "g;x?y#s",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g;x?y#s",
		},
		{
			name:     "RFC3986 Normal Example: empty reference",
			relative: "",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/d;p?q",
		},
		{
			name:     "RFC3986 Normal Example: single dot",
			relative: ".",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/",
		},
		{
			name:     "RFC3986 Normal Example: single dot-slash",
			relative: "./",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/",
		},
		{
			name:     "RFC3986 Normal Example: double dot",
			relative: "..",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/",
		},
		{
			name:     "RFC3986 Normal Example: double dot-slash",
			relative: "../",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/",
		},
		{
			name:     "RFC3986 Normal Example: path up one level",
			relative: "../g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/g",
		},
		{
			name:     "RFC3986 Normal Example: path up two levels",
			relative: "../..",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/",
		},
		{
			name:     "RFC3986 Normal Example: path up two levels slash",
			relative: "../../",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/",
		},
		{
			name:     "RFC3986 Normal Example: path up two levels with new segment",
			relative: "../../g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/g",
		},
		{
			name:     "RFC3986 Abnormal: path with dot-dot at root",
			relative: "/../g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/g",
		},
		{
			name:     "RFC3986 Abnormal: segment ending in dot",
			relative: "g.",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g.",
		},
		{
			name:     "RFC3986 Abnormal: segment starting with dot",
			relative: ".g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/.g",
		},
		{
			name:     "RFC3986 Abnormal: segment ending in double-dot",
			relative: "g..",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g..",
		},
		{
			name:     "RFC3986 Abnormal: segment starting with double-dot",
			relative: "..g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/..g",
		},
		{
			name:     "RFC3986 Abnormal: nonsensical path",
			relative: "./../g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/g",
		},
		{
			name:     "RFC3986 Abnormal: trailing dot segment",
			relative: "./g/.",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g/",
		},
		{
			name:     "RFC3986 Abnormal: middle dot segment",
			relative: "g/./h",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g/h",
		},
		{
			name:     "RFC3986 Abnormal: path normalization up and down",
			relative: "g/../h",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/h",
		},
		{
			name:     "RFC3986 Opaque URI",
			relative: "http:g",
			base:     "http://a/b/c/d;p?q",
			expected: "http:g",
		},
		{
			name:     "RFC3986 Opaque URI empty path",
			relative: "http:",
			base:     "http://a/b/c/d;p?q",
			expected: "http:",
		},
		{
			name:     "Path traversal up one level",
			relative: "../r",
			base:     "http://ex/x/y/z",
			expected: "http://ex/x/r",
		},
		{
			name:     "Simple relative path from directory",
			relative: "q/r",
			base:     "http://ex/x/y",
			expected: "http://ex/x/q/r",
		},
		{
			name:     "Simple relative path with fragment",
			relative: "q/r#s",
			base:     "http://ex/x/y",
			expected: "http://ex/x/q/r#s",
		},
		{
			name:     "Simple relative path from directory with trailing slash",
			relative: "z/",
			base:     "http://ex/x/y/",
			expected: "http://ex/x/y/z/",
		},
		{
			name:     "Fragment on file URI",
			relative: "#Animal",
			base:     "file:/swap/test/animal.rdf",
			expected: "file:/swap/test/animal.rdf#Animal",
		},
		{
			name:     "Absolute path on file URI",
			relative: "/r",
			base:     "file:/ex/x/y/z",
			expected: "file:/r"},
		{
			name:     "Relative path from authority-only IRI",
			relative: "s",
			base:     "http://example.com",
			expected: "http://example.com/s",
		},
		{
			name:     "Path normalization with params",
			relative: "g;x=1/./y",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g;x=1/y",
		},
		{
			name:     "Path normalization up and down with params",
			relative: "g;x=1/../y",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/y",
		},
		{
			name:     "Dot segment in query",
			relative: "g?y/./x",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g?y/./x",
		},
		{
			name:     "Dot-dot segment in query",
			relative: "g?y/../x",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g?y/../x",
		},
		{
			name:     "Dot segment in fragment",
			relative: "g#s/./x",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g#s/./x",
		},
		{
			name:     "Dot-dot segment in fragment",
			relative: "g#s/../x",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/b/c/g#s/../x",
		},
		{
			name:     "Abnormal path normalization from root",
			relative: "/a/b/c/./../../g",
			base:     "http://a/b/c/d;p?q",
			expected: "http://a/a/g",
		},
		{
			name:     "Base with path-like query: relative path",
			relative: "g",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/g",
		},
		{
			name:     "Base with path-like query: relative dot-slash",
			relative: "./g",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/g",
		},
		{
			name:     "Base with path-like query: relative path with slash",
			relative: "g/",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/g/",
		},
		{
			name:     "Base with path-like query: path from root",
			relative: "/g",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/g",
		},
		{
			name:     "Base with path-like query: scheme-relative",
			relative: "//g",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://g",
		},
		{
			name:     "Base with path-like query: query only",
			relative: "?y",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/d;p?y",
		},
		{
			name:     "Base with path-like query: path and query",
			relative: "g?y",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/g?y",
		},
		{
			name:     "Base with path-like query: dot in query",
			relative: "g?y/./x",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/g?y/./x",
		},
		{
			name:     "Base with path-like query: dot-dot in query",
			relative: "g?y/../x",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/g?y/../x",
		},
		{
			name:     "Base with path-like query: path and fragment",
			relative: "g#s",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/g#s",
		},
		{
			name:     "Base with path-like query: dot in fragment",
			relative: "g#s/./x",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/g#s/./x",
		},
		{
			name:     "Base with path-like query: dot-dot in fragment",
			relative: "g#s/../x",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/g#s/../x",
		},
		{
			name:     "Base with path-like query: dot-slash",
			relative: "./",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/c/",
		},
		{
			name:     "Base with path-like query: double dot-slash",
			relative: "../",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/",
		},
		{
			name:     "Base with path-like query: path up one level",
			relative: "../g",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/b/g",
		},
		{
			name:     "Base with path-like query: path up two levels slash",
			relative: "../../",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/",
		},
		{
			name:     "Base with path-like query: path up two levels",
			relative: "../../g",
			base:     "http://a/b/c/d;p?q=1/2",
			expected: "http://a/g",
		},
		{
			name:     "Base with path-like segment: relative path",
			relative: "g",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/d;p=1/g",
		},
		{
			name:     "Base with path-like segment: relative dot-slash",
			relative: "./g",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/d;p=1/g",
		},
		{
			name:     "Base with path-like segment: relative path with slash",
			relative: "g/",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/d;p=1/g/",
		},
		{
			name:     "Base with path-like segment: path and query",
			relative: "g?y",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/d;p=1/g?y",
		},
		{
			name:     "Base with path-like segment: semicolon path",
			relative: ";x",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/d;p=1/;x",
		},
		{
			name:     "Base with path-like segment: path with semicolon",
			relative: "g;x",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/d;p=1/g;x",
		},
		{
			name:     "Base with path-like segment: path norm with params",
			relative: "g;x=1/./y",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/d;p=1/g;x=1/y",
		},
		{
			name:     "Base with path-like segment: path norm up and down",
			relative: "g;x=1/../y",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/d;p=1/y",
		},
		{
			name:     "Base with path-like segment: dot-slash",
			relative: "./",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/d;p=1/",
		},
		{
			name:     "Base with path-like segment: double dot-slash",
			relative: "../",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/",
		},
		{
			name:     "Base with path-like segment: path up one level",
			relative: "../g",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/c/g",
		},
		{
			name:     "Base with path-like segment: path up two levels slash",
			relative: "../../",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/",
		},
		{
			name:     "Base with path-like segment: path up two levels",
			relative: "../../g",
			base:     "http://a/b/c/d;p=1/2?q",
			expected: "http://a/b/g"},
		{
			name:     "Base with empty authority: new scheme",
			relative: "g:h",
			base:     "fred:///s//a/b/c",
			expected: "g:h"},
		{
			name:     "Base with empty authority: relative path",
			relative: "g",
			base:     "fred:///s//a/b/c",
			expected: "fred:///s//a/b/g",
		},
		{
			name:     "Base with empty authority: relative dot-slash",
			relative: "./g",
			base:     "fred:///s//a/b/c",
			expected: "fred:///s//a/b/g",
		},
		{
			name:     "Base with empty authority: relative path with slash",
			relative: "g/",
			base:     "fred:///s//a/b/c",
			expected: "fred:///s//a/b/g/",
		},
		{
			name:     "Base with empty authority: path from root",
			relative: "/g",
			base:     "fred:///s//a/b/c",
			expected: "fred:///g",
		},
		{
			name:     "Base with empty authority: scheme-relative",
			relative: "//g",
			base:     "fred:///s//a/b/c",
			expected: "fred://g",
		},
		{
			name:     "Base with empty authority: scheme-relative with path",
			relative: "//g/x",
			base:     "fred:///s//a/b/c",
			expected: "fred://g/x",
		},
		{
			name:     "Base with empty authority: path from root with extra slash",
			relative: "///g",
			base:     "fred:///s//a/b/c",
			expected: "fred:///g",
		},
		{
			name:     "Base with empty authority: dot-slash",
			relative: "./",
			base:     "fred:///s//a/b/c",
			expected: "fred:///s//a/b/",
		},
		{
			name:     "Base with empty authority: double dot-slash",
			relative: "../",
			base:     "fred:///s//a/b/c",
			expected: "fred:///s//a/",
		},
		{
			name:     "Base with empty authority: path up one level",
			relative: "../g",
			base:     "fred:///s//a/b/c",
			expected: "fred:///s//a/g",
		},
		{
			name:     "Base with empty authority: path up two levels slash",
			relative: "../../",
			base:     "fred:///s//a/b/c",
			expected: "fred:///s//",
		},
		{
			name:     "Base with empty authority: path up two levels",
			relative: "../../g",
			base:     "fred:///s//a/b/c",
			expected: "fred:///s//g",
		},
		{
			name:     "Base with empty authority: path up three levels",
			relative: "../../../g",
			base:     "fred:///s//a/b/c",
			expected: "fred:///s/g",
		},
		{
			name:     "Base with empty authority: path up four levels",
			relative: "../../../../g",
			base:     "fred:///s//a/b/c",
			expected: "fred:///g",
		},
		{
			name:     "Absolute with different scheme",
			relative: "bar:abc",
			base:     "foo:xyz",
			expected: "bar:abc"},
		{
			name:     "Absolute with different authority",
			relative: "http://example/x/abc",
			base:     "http://example2/x/y/z",
			expected: "http://example/x/abc",
		},
		{
			name:     "Fragment containing slash",
			relative: "q/r#s/t",
			base:     "http://ex/x/y",
			expected: "http://ex/x/q/r#s/t",
		},
		{
			name:     "Absolute with ftp scheme",
			relative: "ftp://ex/x/q/r",
			base:     "http://ex/x/y",
			expected: "ftp://ex/x/q/r",
		},
		{
			name:     "Empty relative from file",
			relative: "",
			base:     "file:/ex/x/y/pdq",
			expected: "file:/ex/x/y/pdq",
		},
		{
			name:     "File path relative",
			relative: "z/",
			base:     "file:/ex/x/y/",
			expected: "file:/ex/x/y/z/",
		},
		{
			name:     "File path with authority",
			relative: "file://meetings.example.com/cal#m1",
			base:     "file:/devel/WWW/2000/10/swap/test/reluri-1.n3",
			expected: "file://meetings.example.com/cal#m1",
		},
		{
			name:     "File path with authority from different base",
			relative: "file://meetings.example.com/cal#m1",
			base:     "file:/home/connolly/w3ccvs/WWW/2000/10/swap/test/reluri-1.n3",
			expected: "file://meetings.example.com/cal#m1",
		},
		{
			name:     "Relative file path with fragment",
			relative: "./#blort",
			base:     "file:/some/dir/foo",
			expected: "file:/some/dir/#blort",
		},
		{
			name:     "Relative path to directory with trailing slash",
			relative: "./",
			base:     "http://example/x/abc.efg",
			expected: "http://example/x/",
		},
		{
			name:     "Relative path with colon",
			relative: "./q:r",
			base:     "http://ex/x/y",
			expected: "http://ex/x/q:r"},
		{
			name:     "Relative path with equals and colon",
			relative: "./p=q:r",
			base:     "http://ex/x/y",
			expected: "http://ex/x/p=q:r",
		},
		{
			name:     "Query with slashes",
			relative: "?pp/rr",
			base:     "http://ex/x/y?pp/qq",
			expected: "http://ex/x/y?pp/rr",
		},
		{
			name:     "Relative path from base with query",
			relative: "y/z",
			base:     "http://ex/x/y?pp/qq",
			expected: "http://ex/x/y/z",
		},
		{
			name:     "Relative path from authority and query",
			relative: "/x/y?q",
			base:     "http://ex?p",
			expected: "http://ex/x/y?q",
		},
		{
			name:     "Opaque with relative path",
			relative: "c/d",
			base:     "foo:a/b",
			expected: "foo:a/c/d",
		},
		{
			name:     "Opaque with absolute path",
			relative: "/c/d",
			base:     "foo:a/b",
			expected: "foo:/c/d",
		},
		{
			name:     "Empty relative from opaque with query and fragment",
			relative: "",
			base:     "foo:a/b?c#d",
			expected: "foo:a/b?c",
		},
		{
			name:     "Opaque with base path and new segment",
			relative: "b/c",
			base:     "foo:a",
			expected: "foo:b/c",
		},
		{
			name:     "Opaque path traversal up and down",
			relative: "../b/c",
			base:     "foo:/a/y/z",
			expected: "foo:/a/b/c",
		},
		{
			name:     "Opaque path traversal from root",
			relative: "../../d",
			base:     "foo://a//b/c",
			expected: "foo://a/d",
		},
		{
			name:     "Opaque dot",
			relative: ".",
			base:     "foo:a",
			expected: "foo:",
		},
		{
			name:     "Opaque double dot",
			relative: "..",
			base:     "foo:a",
			expected: "foo:",
		},
		{
			name:     "Path with encoded slash in base",
			relative: "abc",
			base:     "http://example/x/y%2Fz",
			expected: "http://example/x/abc",
		},
		{
			name:     "Path with encoded slash relative",
			relative: "../../x%2Fabc",
			base:     "http://example/a/x/y/z",
			expected: "http://example/a/x%2Fabc",
		},
		{
			name:     "Path with encoded colon relative",
			relative: "q%3Ar",
			base:     "http://ex/x/y",
			expected: "http://ex/x/q%3Ar",
		},
		{
			name:     "Dot-dot in query is not resolved",
			relative: "http://example/a/b?c/../d",
			base:     "foo:bar",
			expected: "http://example/a/b?c/../d",
		},
		{
			name:     "Dot-dot in fragment is not resolved",
			relative: "http://example/a/b#c/../d",
			base:     "foo:bar",
			expected: "http://example/a/b#c/../d",
		},
		{
			name:     "Opaque http relative",
			relative: "http:this",
			base:     "http://example.org/base/uri",
			expected: "http:this",
		},
		{
			name:     "Windows file path resolution",
			relative: "mini1.xml",
			base:     "file:///C:/DEV/Haskell/lib/HXmlToolbox-3.01/examples/",
			expected: "file:///C:/DEV/Haskell/lib/HXmlToolbox-3.01/examples/mini1.xml",
		},
		{
			name:     "Opaque file with query",
			relative: "?bar",
			base:     "file:foo",
			expected: "file:foo?bar",
		},
		{
			name:     "Opaque file with fragment",
			relative: "#bar",
			base:     "file:foo",
			expected: "file:foo#bar",
		},
		{
			name:     "Opaque file to absolute path",
			relative: "/lv2.h",
			base:     "file:foo",
			expected: "file:/lv2.h"},
		{
			name:     "Opaque file to absolute path with empty authority",
			relative: "///lv2.h",
			base:     "file:foo",
			expected: "file:///lv2.h",
		},
		{
			name:     "Opaque file with relative path",
			relative: "lv2.h",
			base:     "file:foo",
			expected: "file:lv2.h",
		},
		{
			name:     "Opaque file with empty path dot",
			relative: ".",
			base:     "file:",
			expected: "file:",
		},
	}

	for _, test := range examples {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			base, err := ParseIri(test.base)
			if err != nil {
				t.Fatalf("Failed to parse base IRI %q: %v", test.base, err)
			}
			expected, err := ParseRef(test.expected)
			if err != nil {
				t.Fatalf("Failed to parse expected IRI %q: %v", test.expected, err)
			}

			result, err := base.Resolve(test.relative)
			if err != nil {
				t.Fatalf("Resolving %q against %q failed: %v", test.relative, test.base, err)
			}
			if result.String() != expected.String() {
				t.Errorf("Resolve: got %q, want %q", result.String(), expected.String())
			}
		})
	}
}

// TestRelativizeIRI verifies the `Relativize` method, which is the inverse of `Resolve`.
// For each test case, it computes a relative reference from a base and target IRI,
// and then performs a round-trip check by resolving that relative reference back
// against the base to ensure it matches the original target.
func TestRelativizeIRI(t *testing.T) {
	t.Parallel()
	examples := []resolveTest{
		{
			name:     "Identical opaque IRIs",
			relative: "",
			base:     "http:",
			expected: "http:",
		},
		{
			name:     "Identical hierarchical IRIs",
			relative: "",
			base:     "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "Identical with path",
			relative: "",
			base:     "http://example.com/foo",
			expected: "http://example.com/foo",
		},
		{
			name:     "Identical with longer path",
			relative: "",
			base:     "http://example.com/foo/bar",
			expected: "http://example.com/foo/bar",
		},
		{
			name:     "Identical with query",
			relative: "",
			base:     "http://example.com/foo/bar?bat",
			expected: "http://example.com/foo/bar?bat",
		},
		{
			name:     "Identical with query and fragment",
			relative: "#baz",
			base:     "http://example.com/foo/bar?bat#baz",
			expected: "http://example.com/foo/bar?bat#baz",
		},
		{
			name:     "Different schemes",
			relative: "http:",
			base:     "http:",
			expected: "https:",
		},
		{
			name:     "Different authorities",
			relative: "//example.com",
			base:     "http://example.com",
			expected: "http://example.org",
		},
		{
			name:     "Sibling path segments",
			relative: "foo",
			base:     "http://example.com/foo",
			expected: "http://example.com/bar",
		},
		{
			name:     "Different queries",
			relative: "?bat",
			base:     "http://example.com/foo?bat",
			expected: "http://example.com/foo?foo",
		},
		{
			name:     "Different fragments",
			relative: "#baz",
			base:     "http://example.com/foo?bat#baz",
			expected: "http://example.com/foo?bat#foo",
		},
		{
			name:     "Hierarchical from opaque",
			relative: "//example.com",
			base:     "http://example.com",
			expected: "http:",
		},
		{
			name:     "Hierarchical from authority-only",
			relative: "//example.com",
			base:     "http://example.com",
			expected: "http://",
		},
		{
			name:     "Child path from parent directory",
			relative: "foo",
			base:     "http://example.com/foo",
			expected: "http://example.com/",
		},
		{
			name:     "Path from different tree",
			relative: "/foo",
			base:     "http://example.com/foo",
			expected: "http://example.com/bar/baz",
		},
		{
			name:     "Sibling path",
			relative: "bar",
			base:     "http://example.com/foo/bar",
			expected: "http://example.com/foo/baz",
		},
		{
			name:     "Parent path from child",
			relative: "foo/bar",
			base:     "http://example.com/foo/bar",
			expected: "http://example.com/foo",
		},
		{
			name:     "Sibling query",
			relative: "?bar",
			base:     "http://example.com/foo?bar",
			expected: "http://example.com/foo?baz",
		},
		{
			name:     "Path to query",
			relative: "//example.com?bar",
			base:     "http://example.com?bar",
			expected: "http://example.com/a",
		},
		{
			name:     "No path from query",
			relative: "?bar",
			base:     "http://example.com?bar",
			expected: "http://example.com",
		},
		{
			name:     "Path with slash from query",
			relative: "//example.com?bar",
			base:     "http://example.com?bar",
			expected: "http://example.com/",
		},
		{
			name:     "Sibling fragment",
			relative: "#bar",
			base:     "http://example.com/foo#bar",
			expected: "http://example.com/foo#baz",
		},
		{
			name:     "Path from parent dir to file",
			relative: ".",
			base:     "http://example.com/foo/",
			expected: "http://example.com/foo/bar",
		},
		{
			name:     "Path with colon segment",
			relative: "/:",
			base:     "http://example.com/:",
			expected: "http://example.com/foo",
		},
		{
			name:     "Opaque from hierarchical",
			relative: "http:",
			base:     "http:",
			expected: "http://example.com",
		},
		{
			name:     "Opaque with query from hierarchical",
			relative: "http:?foo",
			base:     "http:?foo",
			expected: "http://example.com",
		},
		{
			name:     "No path from path",
			relative: "//example.com",
			base:     "http://example.com",
			expected: "http://example.com/foo",
		},
		{
			name:     "No path from path with query",
			relative: "//example.com",
			base:     "http://example.com",
			expected: "http://example.com?query",
		},
		{
			name:     "Path from path with query",
			relative: "foo",
			base:     "http://example.com/foo",
			expected: "http://example.com/foo?query",
		},
		{
			name:     "Opaque with query from hierarchical with query",
			relative: "http:?query",
			base:     "http:?query",
			expected: "http://example.com?query",
		},
		{
			name:     "Opaque path from hierarchical path",
			relative: "http:/path",
			base:     "http:/path",
			expected: "http://example.com/foo",
		},
		{
			name:     "Path with empty segment",
			relative: "//example.com//a",
			base:     "http://example.com//a",
			expected: "http://example.com/",
		},
		{
			name:     "URN child",
			relative: "ab",
			base:     "urn:ab",
			expected: "urn:",
		},
		{
			name:     "URN with path",
			relative: "urn:isbn:foo",
			base:     "urn:isbn:foo",
			expected: "urn:",
		},
		{
			name:     "URN with slash",
			relative: "is/bn:foo",
			base:     "urn:is/bn:foo",
			expected: "urn:",
		},
		{
			name:     "Opaque sibling path",
			relative: "e/p",
			base:     "t:e/e/p",
			expected: "t:e/s",
		},
		{
			name:     "Opaque child from parent with slash",
			relative: "gp",
			base:     "htt:/foo/gp",
			expected: "htt:/foo/",
		},
		{
			name:     "Opaque child from parent without slash",
			relative: "gp",
			base:     "htt:/gp",
			expected: "htt:/",
		},
		{
			name:     "Opaque from hierarchical with authority",
			relative: "x:",
			base:     "x:",
			expected: "x://foo",
		},
		{
			name:     "Opaque from opaque with path",
			relative: "x:",
			base:     "x:",
			expected: "x:02",
		},
		{
			name:     "Opaque from opaque with query",
			relative: "x:",
			base:     "x:",
			expected: "x:?foo"},
		{
			name:     "Fragment from no fragment",
			relative: "",
			base:     "http://example.com",
			expected: "http://example.com#foo",
		},
		{
			name:     "Same directory relative",
			relative: ".",
			base:     "http://example.com/a/",
			expected: "http://example.com/a/b",
		},
		{
			name:     "Same directory relative with query",
			relative: ".?c",
			base:     "http://example.com/a/?c",
			expected: "http://example.com/a/b",
		},
		{
			name:     "Opaque path with empty authority",
			relative: "t:o//",
			base:     "t:o//",
			expected: "t:o/",
		},
		{
			name:     "Opaque path with colon",
			relative: "t:a/c:d",
			base:     "t:a/c:d",
			expected: "t:a/b",
		},
		{
			name:     "Same dir with query",
			relative: ".",
			base:     "http://example.com/a/b/",
			expected: "http://example.com/a/b/?q=1",
		},
		{
			name:     "Root path from file",
			relative: "/foo",
			base:     "http://example.com/foo",
			expected: "http://example.com",
		},
		{
			name:     "Path traversal up and down",
			relative: "../c",
			base:     "http://example.com/a/c",
			expected: "http://example.com/a/b/",
		},
	}

	for _, test := range examples {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			original, err := ParseIri(test.base)
			if err != nil {
				t.Fatalf("Failed to parse original IRI %q: %v", test.base, err)
			}
			base, err := ParseIri(test.expected)
			if err != nil {
				t.Fatalf("Failed to parse base IRI %q: %v", test.expected, err)
			}

			actual, err := base.Relativize(original)
			if err != nil {
				t.Fatalf("Relativize failed: %v", err)
			}

			if actual.String() != test.relative {
				t.Errorf("Relativizing %q against %q gives %q, want %q",
					original.String(), base.String(), actual.String(), test.relative)
			}

			// Round-trip check
			resolved, err := base.Resolve(actual.String())
			if err != nil {
				t.Fatalf("Round-trip resolve failed: %v", err)
			}
			if resolved.String() != original.String() {
				t.Errorf("Round-trip failed: resolving %q against %q gives %q, want %q",
					actual.String(), base.String(), resolved.String(), original.String())
			}
		})
	}
}

// TestRelativizeIRIFails ensures that the `Relativize` method fails as expected when
// the target IRI contains dot-segments, which is not allowed.
func TestRelativizeIRIFails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		iri  string
		base string
	}{
		{
			name: "Hierarchical with dot-dot segment",
			iri:  "http://example.com/a/../b",
			base: "http://example.com/s",
		},
		{
			name: "Hierarchical with trailing dot-dot segment",
			iri:  "http://example.com/a/..",
			base: "http://example.com/s",
		},
		{
			name: "Hierarchical with dot segment",
			iri:  "http://example.com/./b",
			base: "http://example.com/s",
		},
		{
			name: "Hierarchical with trailing dot segment",
			iri:  "http://example.com/.",
			base: "http://example.com/s",
		},
		{
			name: "Opaque with trailing dot segment",
			iri:  "urn:.",
			base: "urn:",
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_AGAINST_%s", test.iri, test.base), func(t *testing.T) {
			t.Parallel()
			iri, err := ParseIri(test.iri)
			if err != nil {
				t.Fatalf("Bad test case, could not parse IRI %q: %v", test.iri, err)
			}
			base, err := ParseIri(test.base)
			if err != nil {
				t.Fatalf("Bad test case, could not parse base %q: %v", test.base, err)
			}

			_, err = base.Relativize(iri)
			if err == nil {
				t.Errorf("Expected Relativize to fail, but it did not")
			} else if !errors.Is(err, ErrIriRelativize) {
				t.Errorf("Expected ErrIriRelativize, but got: %v", err)
			}
		})
	}
}

// TestEq checks the value equality of the Iri struct. It verifies that two IRIs
// parsed from the same string are considered equal and can be used as map keys.
func TestEq(t *testing.T) {
	t.Parallel()

	iri1, err := ParseIri("http://example.com")
	if err != nil {
		t.Fatalf("ParseIri failed for iri1: %v", err)
	}
	iri2, err := ParseIri("http://example.com")
	if err != nil {
		t.Fatalf("ParseIri failed for iri2: %v", err)
	}
	iri3, err := ParseIri("http://example.org")
	if err != nil {
		t.Fatalf("ParseIri failed for iri3: %v", err)
	}

	if iri1.String() != "http://example.com" {
		t.Errorf("Expected iri.String() to be 'http://example.com', got %q", iri1.String())
	}

	if *iri1 != *iri2 {
		t.Errorf("Expected two identical IRIs to have equal values, but they did not")
	}
	if *iri1 == *iri3 {
		t.Errorf("Expected two different IRIs to have unequal values, but they were equal")
	}

	m := make(map[Iri]bool)
	m[*iri1] = true

	if !m[*iri2] {
		t.Error("Expected to find IRI in map using an equal IRI value as key")
	}

	if m[*iri3] {
		t.Error("Expected not to find IRI in map using a different IRI value as key")
	}

	// This is a check on pointer equality, not value equality.
	if iri1 == iri2 {
		t.Error("Expected two separately parsed IRIs to have different memory addresses, but they were the same")
	}
}

// TestStr is a simple check to ensure the String() method returns a sensible value.
func TestStr(t *testing.T) {
	t.Parallel()
	iri, err := ParseIri("http://example.com")
	if err != nil {
		t.Fatalf("ParseIri failed: %v", err)
	}
	if !strings.HasPrefix(iri.String(), "http://") {
		t.Error("Expected iri to have prefix 'http://'")
	}
}

// TestRefJSON tests the JSON marshalling and unmarshalling of the Ref type.
func TestRefJSON(t *testing.T) {
	t.Parallel()
	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()
		ref := ParseRefUnchecked("//example.com")
		jsonData, err := json.Marshal(ref)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		expected := []byte(`"//example.com"`)
		if !bytes.Equal(jsonData, expected) {
			t.Errorf("got %s, want %s", jsonData, expected)
		}
	})

	t.Run("Unmarshal valid", func(t *testing.T) {
		t.Parallel()
		jsonData := []byte(`"//example.com"`)
		var ref Ref
		if err := json.Unmarshal(jsonData, &ref); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if ref.String() != "//example.com" {
			t.Errorf("got %q, want %q", ref.String(), "//example.com")
		}
	})

	t.Run("Unmarshal invalid", func(t *testing.T) {
		t.Parallel()
		jsonData := []byte(`":"`)
		var ref Ref
		err := json.Unmarshal(jsonData, &ref)
		if err == nil {
			t.Fatal("json.Unmarshal expected to fail, but did not")
		}
		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("Expected a ParseError, but got %T: %v", err, err)
		}
	})

	t.Run("Unmarshal non-string JSON", func(t *testing.T) {
		t.Parallel()
		jsonData := []byte(`123`)
		var ref Ref
		err := json.Unmarshal(jsonData, &ref)
		if err == nil {
			t.Fatal("json.Unmarshal expected to fail for non-string JSON, but did not")
		}
		var unmarshalTypeError *json.UnmarshalTypeError
		if !errors.As(err, &unmarshalTypeError) {
			t.Errorf("Expected a json.UnmarshalTypeError, but got %T: %v", err, err)
		}
	})
}

// TestIriJSON tests the JSON marshalling and unmarshalling of the Iri type,
// ensuring that it correctly handles both valid and invalid absolute IRIs.
func TestIriJSON(t *testing.T) {
	t.Parallel()
	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()
		iri := ParseIriUnchecked("http://example.com")
		jsonData, err := json.Marshal(iri)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		expected := []byte(`"http://example.com"`)
		if !bytes.Equal(jsonData, expected) {
			t.Errorf("got %s, want %s", jsonData, expected)
		}
	})

	t.Run("Unmarshal valid", func(t *testing.T) {
		t.Parallel()
		jsonData := []byte(`"http://example.com"`)
		var iri Iri
		if err := json.Unmarshal(jsonData, &iri); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if iri.String() != "http://example.com" {
			t.Errorf("got %q, want %q", iri.String(), "http://example.com")
		}
	})

	t.Run("Unmarshal invalid IRI", func(t *testing.T) {
		t.Parallel()
		jsonData := []byte(`":"`)
		var iri Iri
		err := json.Unmarshal(jsonData, &iri)
		if err == nil {
			t.Fatal("json.Unmarshal expected to fail, but did not")
		}

		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("Expected a ParseError, but got %T: %v", err, err)
		}
	})

	t.Run("Unmarshal relative", func(t *testing.T) {
		t.Parallel()
		jsonData := []byte(`"//example.com"`)
		var iri Iri
		err := json.Unmarshal(jsonData, &iri)
		if err == nil {
			t.Fatal("json.Unmarshal expected to fail, but did not")
		}

		if !strings.Contains(err.Error(), "No scheme found") {
			t.Errorf("Expected 'No scheme found' error, but got: %v", err)
		}
	})
}

// TestResolveRelativeIRIUnchecked verifies the `ResolveUnchecked` method with a large
// number of test cases, ensuring it produces correct results for known-good inputs.
func TestResolveRelativeIRIUnchecked(t *testing.T) {
	t.Parallel()

	examples := []resolveTest{
		{
			name:     "Path traversal up from root",
			relative: "../foo",
			base:     "http://host/",
			expected: "http://host/foo",
		},
		{
			name:     "Path traversal up from file",
			relative: "../foo",
			base:     "http://host/xyz",
			expected: "http://host/foo",
		},
		{
			name:     "Relative path with query",
			relative: "d/z?x=a",
			base:     "http://www.example.org/a/b/c/d",
			expected: "http://www.example.org/a/b/c/d/z?x=a",
		},
		{
			name:     "Absolute IRI",
			relative: "http://example.com/A",
			base:     "http://www.example.org/a/b/c/d",
			expected: "http://example.com/A",
		},
		{
			name:     "Empty from directory",
			relative: "",
			base:     "http://www.example.org/a/b/c/d/",
			expected: "http://www.example.org/a/b/c/d/",
		},
		{
			name:     "Dot from directory",
			relative: ".",
			base:     "http://www.example.org/a/b/c/d/",
			expected: "http://www.example.org/a/b/c/d/",
		},
		{
			name:     "Path traversal up and down",
			relative: "../../C/D",
			base:     "http://www.example.org/a/b/c/d/",
			expected: "http://www.example.org/a/b/C/D",
		},
		{
			name:     "Path traversal up and down to same dir",
			relative: "../../c/d/",
			base:     "http://www.example.org/a/b/c/d/",
			expected: "http://www.example.org/a/b/c/d/",
		},
		{
			name:     "Path traversal and new segment with fragment",
			relative: "../../c/d/X#bar",
			base:     "http://www.example.org/a/b/c/d/",
			expected: "http://www.example.org/a/b/c/d/X#bar",
		},
		{
			name:     "Path traversal and new longer path",
			relative: "../../c/d/e/f/g/",
			base:     "http://www.example.org/a/b/c/d/",
			expected: "http://www.example.org/a/b/c/d/e/f/g/",
		},
		{
			name:     "Path traversal and new segment with query",
			relative: "../../c/d/z?x=a",
			base:     "http://www.example.org/a/b/c/d/",
			expected: "http://www.example.org/a/b/c/d/z?x=a",
		},
		{
			name:     "W3C charmod test with unicode",
			relative: "http://example.org/#André",
			base:     "http://www.w3.org/2000/10/rdf-tests/rdfcore/rdf-charmod-uris/test001.rdf",
			expected: "http://example.org/#André",
		},
		{
			name:     "W3C charmod test with percent encoding",
			relative: "http://example.org/#Andr%C3%A9",
			base:     "http://www.w3.org/2000/10/rdf-tests/rdfcore/rdf-charmod-uris/test002.rdf",
			expected: "http://example.org/#Andr%C3%A9",
		},
		{
			name:     "W3C ID difference test with unicode",
			relative: "#Dürst",
			base:     "http://www.w3.org/2000/10/rdf-tests/rdfcore/rdfms-difference-between-ID-and-about/test2.rdf",
			expected: "http://www.w3.org/2000/10/rdf-tests/rdfcore/rdfms-difference-between-ID-and-about/test2.rdf#Dürst",
		},
		{
			name:     "Empty fragment on opaque URI",
			relative: "#",
			base:     "base:x",
			expected: "base:x#",
		},
		{
			name:     "Windows path with spaces empty relative",
			relative: "",
			base:     "file:///C:/Documents and Settings/jjchplb/Local Settings/Temp/test-load-with-41.rdf",
			expected: "file:///C:/Documents and Settings/jjchplb/Local Settings/Temp/test-load-with-41.rdf",
		},
		{
			name:     "Absolute from windows path",
			relative: "eh:/a",
			base:     "file:///C:/Documents and Settings/jjchplb/Local Settings/Temp/test-load-with-41.rdf",
			expected: "eh:/a",
		},
		{
			name:     "Jena empty fragment",
			relative: "#",
			base:     "file:///C:/eclipse/workspace/jena2/",
			expected: "file:///C:/eclipse/workspace/jena2/#",
		},
		{
			name:     "Jena empty relative",
			relative: "",
			base:     "file:///C:/eclipse/workspace/jena2/",
			expected: "file:///C:/eclipse/workspace/jena2/",
		},
		{
			name:     "Jena relative path",
			relative: "base",
			base:     "file:///C:/eclipse/workspace/jena2/",
			expected: "file:///C:/eclipse/workspace/jena2/base",
		},
		{
			name:     "Jena absolute scheme",
			relative: "eh://R",
			base:     "file:///C:/eclipse/workspace/jena2/",
			expected: "eh://R",
		},
		{
			name:     "Jena absolute scheme with path",
			relative: "eh:/O",
			base:     "file:///C:/eclipse/workspace/jena2/",
			expected: "eh:/O",
		},
		{
			name:     "Jena absolute scheme with fragment",
			relative: "rdf://test.com#",
			base:     "file:///C:/eclipse/workspace/jena2/",
			expected: "rdf://test.com#",
		},
		{
			name:     "Jena relative path from file",
			relative: "z",
			base:     "file:///C:/eclipse/workspace/jena2/foo.n3",
			expected: "file:///C:/eclipse/workspace/jena2/z",
		},
		{
			name:     "Jena ARQ Ask empty relative",
			relative: "",
			base:     "file:///C:/eclipse/workspace/jena2/testing/ARQ/Ask/manifest.ttl",
			expected: "file:///C:/eclipse/workspace/jena2/testing/ARQ/Ask/manifest.ttl",
		},
		{
			name:     "Jena ARQ Basic relative ttl",
			relative: "r-base-prefix-3.ttl",
			base:     "file:///C:/eclipse/workspace/jena2/testing/ARQ/Basic/manifest.ttl",
			expected: "file:///C:/eclipse/workspace/jena2/testing/ARQ/Basic/r-base-prefix-3.ttl",
		},
		{
			name:     "Jena ARQ Basic relative ttl 2",
			relative: "r-base-prefix-4.ttl",
			base:     "file:///C:/eclipse/workspace/jena2/testing/ARQ/Basic/manifest.ttl",
			expected: "file:///C:/eclipse/workspace/jena2/testing/ARQ/Basic/r-base-prefix-4.ttl",
		},
		{
			name:     "Jena ARQ Optional mailto",
			relative: "mailto:bert@example.net",
			base:     "file:///C:/eclipse/workspace/jena2/testing/ARQ/Optional/result-opt-1.ttl",
			expected: "mailto:bert@example.net",
		},
		{
			name:     "Jena ARQ Bound manifest",
			relative: "Bound/manifest.n3",
			base:     "file:///C:/eclipse/workspace/jena2/testing/ARQ/manifest-arq.ttl",
			expected: "file:///C:/eclipse/workspace/jena2/testing/ARQ/Bound/manifest.n3",
		},
		{
			name:     "Jena ARQ Construct manifest",
			relative: "Construct/manifest.ttl",
			base:     "file:///C:/eclipse/workspace/jena2/testing/ARQ/manifest-arq.ttl",
			expected: "file:///C:/eclipse/workspace/jena2/testing/ARQ/Construct/manifest.ttl",
		},
		{
			name:     "Jena ARQ Dataset manifest",
			relative: "Dataset/manifest.n3",
			base:     "file:///C:/eclipse/workspace/jena2/testing/ARQ/manifest-arq.ttl",
			expected: "file:///C:/eclipse/workspace/jena2/testing/ARQ/Dataset/manifest.n3",
		},
		{
			name:     "Jena DAWG mailto",
			relative: "mailto:jlow@example.com",
			base:     "file:///C:/eclipse/workspace/jena2/testing/DAWG-Approved/examples/ex2-4a.n3",
			expected: "mailto:jlow@example.com",
		},
		{
			name:     "Jena DAWG examples manifest",
			relative: "ex11.2.3.2_0.rq",
			base:     "file:///C:/eclipse/workspace/jena2/testing/DAWG/examples/manifest.n3",
			expected: "file:///C:/eclipse/workspace/jena2/testing/DAWG/examples/ex11.2.3.2_0.rq",
		},
		{
			name:     "Jena RDQL URN with comment-like chars",
			relative: "urn:/*not_a_comment*/",
			base:     "file:///C:/eclipse/workspace/jena2/testing/RDQL-ARQ/result-0-01.n3",
			expected: "urn:/*not_a_comment*/",
		},
		{
			name:     "Jena ontology bug test fragment",
			relative: "#y1",
			base:     "file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/test_hk_06/b.owl",
			expected: "file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/test_hk_06/b.owl#y1",
		},
		{
			name:     "Jena ontology bug test empty",
			relative: "",
			base:     "file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/test_hk_06/b.owl",
			expected: "file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/test_hk_06/b.owl",
		},
		{
			name:     "Jena ontology bug test relative with fragment",
			relative: "foo#ClassAC",
			base:     "file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/test_hk_07A.owl",
			expected: "file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/foo#ClassAC",
		},
		{
			name:     "Jena reasoners bug test",
			relative: "jason6",
			base:     "file:///C:/eclipse/workspace/jena2/testing/reasoners/bugs/sbug.rdf",
			expected: "file:///C:/eclipse/workspace/jena2/testing/reasoners/bugs/jason6",
		},
		{
			name:     "Jena reasoners URN",
			relative: "urn:x-propNum100",
			base:     "file:///C:/eclipse/workspace/jena2/testing/reasoners/bugs/subpropertyModel.n3",
			expected: "urn:x-propNum100",
		},
		{
			name:     "File with DOS path",
			relative: "",
			base:     "file:C:\\DOCUME~1\\jjchplb\\LOCALS~1\\Temp\\hedgehog6739.rdf",
			expected: "file:C:\\DOCUME~1\\jjchplb\\LOCALS~1\\Temp\\hedgehog6739.rdf",
		},
		{
			name:     "File with DOS path to absolute",
			relative: "http://spoo.net/O",
			base:     "file:C:\\DOCUME~1\\jjchplb\\LOCALS~1\\Temp\\hedgehog6739.rdf",
			expected: "http://spoo.net/O",
		},
		{
			name:     "File with DOS path to absolute 2",
			relative: "http://spoo.net/S",
			base:     "file:C:\\DOCUME~1\\jjchplb\\LOCALS~1\\Temp\\hedgehog6739.rdf",
			expected: "http://spoo.net/S",
		},
		{
			name:     "File with URN",
			relative: "urn:x-hp:eg/",
			base:     "file:doc/inference/data/owlDemoSchema.xml",
			expected: "urn:x-hp:eg/",
		},
		{
			name:     "File with relative path empty",
			relative: "",
			base:     "file:testing/abbreviated/relative-uris.rdf",
			expected: "file:testing/abbreviated/relative-uris.rdf",
		},
		{
			name:     "File with relative path dot",
			relative: ".",
			base:     "file:testing/abbreviated/relative-uris.rdf",
			expected: "file:testing/abbreviated/",
		},
		{
			name:     "File with relative path up and down",
			relative: "../../C/D",
			base:     "file:testing/abbreviated/relative-uris.rdf",
			expected: "file:C/D",
		},
		{
			name:     "File with relative path to scheme-relative",
			relative: "//example.com/A",
			base:     "file:testing/abbreviated/relative-uris.rdf",
			expected: "file://example.com/A",
		},
		{
			name:     "File with relative path to absolute with fragment",
			relative: "/A/B#foo/",
			base:     "file:testing/abbreviated/relative-uris.rdf",
			expected: "file:/A/B#foo/",
		},
		{
			name:     "File with relative path and fragment",
			relative: "X#bar",
			base:     "file:testing/abbreviated/relative-uris.rdf",
			expected: "file:testing/abbreviated/X#bar",
		},
		{
			name:     "File with longer relative path",
			relative: "e/f/g/",
			base:     "file:testing/abbreviated/relative-uris.rdf",
			expected: "file:testing/abbreviated/e/f/g/",
		},
		{
			name:     "File to absolute http",
			relative: "http://www.example.org/a/b/c/d/",
			base:     "file:testing/abbreviated/relative-uris.rdf",
			expected: "http://www.example.org/a/b/c/d/",
		},
		{
			name:     "File with relative path and query",
			relative: "z?x=a",
			base:     "file:testing/abbreviated/relative-uris.rdf",
			expected: "file:testing/abbreviated/z?x=a",
		},
		{
			name:     "QName in ID fragment",
			relative: "#one",
			base:     "file:testing/arp/qname-in-ID/bug74_0.rdf",
			expected: "file:testing/arp/qname-in-ID/bug74_0.rdf#one",
		},
		{
			name:     "QName in ID fragment with colon",
			relative: "#sw:test",
			base:     "file:testing/arp/qname-in-ID/bug74_0.rdf",
			expected: "file:testing/arp/qname-in-ID/bug74_0.rdf#sw:test",
		},
		{
			name:     "Fragment with double hash",
			relative: "#__rest3",
			base:     "file:testing/ontology/bugs/test_oh_01.owl",
			expected: "file:testing/ontology/bugs/test_oh_01.owl#__rest3",
		},
		{
			name:     "Fragment on LDP test file",
			relative: "#Union2",
			base:     "file:testing/ontology/owl/list-syntax/test-ldp.rdf",
			expected: "file:testing/ontology/owl/list-syntax/test-ldp.rdf#Union2",
		},
		{
			name:     "URN from OWL file",
			relative: "urn:foo",
			base:     "file:testing/reasoners/bugs/cardFPTest.owl",
			expected: "urn:foo",
		},
		{
			name:     "Absolute from OWL file",
			relative: "http://decsai.ugr.es/~ontoserver/bacarex2.owl",
			base:     "file:testing/reasoners/bugs/deleteBug.owl",
			expected: "http://decsai.ugr.es/~ontoserver/bacarex2.owl",
		},
		{
			name:     "Fragment on OWL file",
			relative: "#A",
			base:     "file:testing/reasoners/bugs/equivalentClassTest.owl",
			expected: "file:testing/reasoners/bugs/equivalentClassTest.owl#A",
		},
		{
			name:     "Opaque with colon",
			relative: "NC:ispinfo",
			base:     "http://bar.com/irrelevant",
			expected: "NC:ispinfo",
		},
		{
			name:     "Opaque with colon 2",
			relative: "NC:trickMe",
			base:     "http://bar.com/irrelevant",
			expected: "NC:trickMe",
		},
		{
			name:     "Chrome protocol",
			relative: "chrome://messenger/content/mailPrefsOverlay.xul",
			base:     "http://bar.com/irrelevant",
			expected: "chrome://messenger/content/mailPrefsOverlay.xul",
		},
		{
			name:     "Domain protocol",
			relative: "domain:aol.com",
			base:     "http://bar.com/irrelevant",
			expected: "domain:aol.com",
		},
		{
			name:     "IRI with trailing spaces",
			relative: "http://foo.com/    ",
			base:     "http://bar.com/irrelevant",
			expected: "http://foo.com/    ",
		},
		{
			name:     "IRI with trailing tab",
			relative: "http://foo.com/\t",
			base:     "http://bar.com/irrelevant",
			expected: "http://foo.com/\t",
		},
		{
			name:     "IRI with trailing newline",
			relative: "http://foo.com/\n\n",
			base:     "http://bar.com/irrelevant",
			expected: "http://foo.com/\n\n",
		},
		{
			name:     "IRI with trailing CR",
			relative: "http://foo.com/\r",
			base:     "http://bar.com/irrelevant",
			expected: "http://foo.com/\r",
		},
		{
			name:     "IRI with single quote",
			relative: "http://foo.com/'",
			base:     "http://bar.com/irrelevant",
			expected: "http://foo.com/'",
		},
		{
			name:     "IRI with tag char",
			relative: "http://foo.com/<b>boo",
			base:     "http://bar.com/irrelevant",
			expected: "http://foo.com/<b>boo",
		},
		{
			name:     "IRI with double quote",
			relative: "http://foo.com/\"",
			base:     "http://bar.com/irrelevant",
			expected: "http://foo.com/\"",
		},
		{
			name:     "Simple absolute IRI",
			relative: "http://foo.com/",
			base:     "http://bar.com/irrelevant",
			expected: "http://foo.com/",
		},
	}

	for _, test := range examples {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			base := ParseIriUnchecked(test.base)
			result := base.ResolveUnchecked(test.relative)
			if result.String() != test.expected {
				t.Errorf(
					"Resolving of %q against %q is wrong.\nGot:      %q\nExpected: %q",
					test.relative,
					test.base,
					result.String(),
					test.expected,
				)
			}
		})
	}
}

// TestResolveTo checks that the `ResolveTo` and `ResolveUncheckedTo` methods,
// which write to a strings.Builder, work correctly.
func TestResolveTo(t *testing.T) {
	t.Parallel()
	base := ParseIriUnchecked("http://foo.com/bar/baz")
	var builder strings.Builder

	err := base.ResolveTo("bat#foo", &builder)
	if err != nil {
		t.Fatalf("ResolveTo failed: %v", err)
	}
	expected := "http://foo.com/bar/bat#foo"
	if builder.String() != expected {
		t.Errorf("ResolveTo: got %q, want %q", builder.String(), expected)
	}

	builder.Reset()
	base.ResolveUncheckedTo("bat#foo", &builder)
	if builder.String() != expected {
		t.Errorf("ResolveUncheckedTo: got %q, want %q", builder.String(), expected)
	}
}

// FuzzRelativizeRoundtrip fuzz tests the Relativize and Resolve methods by
// generating a relative IRI and then resolving it back, checking if the result
// matches the original absolute IRI. This helps find edge cases in both algorithms.
func FuzzRelativizeRoundtrip(f *testing.F) {
	f.Add("http://example.com/a/b", "http://example.com/a/c")
	f.Add("http://example.com/a/b", "http://example.com/d")
	f.Add("urn:foo:bar", "urn:foo:baz")
	f.Add("http://a/b", "https://a/b")
	f.Add("file:/a/b/c", "file:/a/d")

	f.Fuzz(func(t *testing.T, absIRIStr, baseIRIStr string) {
		base, err := ParseIri(baseIRIStr)
		if err != nil {
			return // Skip invalid base IRIs generated by fuzzer.
		}
		abs, err := ParseIri(absIRIStr)
		if err != nil {
			return // Skip invalid absolute IRIs generated by fuzzer.
		}

		rel, err := base.Relativize(abs)
		if err != nil {
			return // Skip cases where relativization is not possible.
		}

		resolved, err := base.Resolve(rel.String())
		if err != nil {
			// This should not happen if relativize produced a valid reference.
			t.Fatalf("Failed to resolve back relative IRI %q from base %q: %v", rel.String(), base.String(), err)
		}

		if resolved.String() != abs.String() {
			t.Errorf("Relativize round-trip failed:\n  base:     %s\n  absolute: %s\n  relative: %s\n  resolved: %s",
				base.String(), abs.String(), rel.String(), resolved.String())
		}
	})
}

// TestParseIriFails ensures that ParseIri returns an error for syntactically
// invalid IRI strings.
func TestParseIriFails(t *testing.T) {
	t.Parallel()
	// These inputs are invalid because they are not valid IRI-references.
	invalidIRIs := []string{
		":",
		"http://[/",
		"http://a b.com/",
	}

	for _, s := range invalidIRIs {
		t.Run(s, func(t *testing.T) {
			t.Parallel()
			_, err := ParseIri(s)
			if err == nil {
				t.Errorf("ParseIri(%q) was expected to fail, but it did not", s)
			}
		})
	}
}

// TestParseError_Unwrap verifies that the Unwrap method of a ParseError
// correctly returns the underlying wrapped error, making it compatible
// with functions like errors.Is.
func TestParseError_Unwrap(t *testing.T) {
	t.Parallel()

	underlyingErr := errors.New("this is the specific cause")
	wrappedErr := fmt.Errorf("some context: %w", underlyingErr)
	parseErr := newParseError(wrappedErr)

	if !errors.Is(parseErr, underlyingErr) {
		t.Errorf("errors.Is failed: expected the ParseError to wrap the underlying error, but it did not")
	}

	unwrapped := errors.Unwrap(parseErr)
	if !errors.Is(unwrapped, underlyingErr) {
		t.Errorf(
			"errors.Unwrap() returned <%v>, which is not the expected underlying error <%v>",
			unwrapped,
			underlyingErr,
		)
	}
}

// TestParseRefUnchecked_Panic verifies that ParseRefUnchecked panics
// when the underlying parser unexpectedly returns an error, even in unchecked mode.
// This is a test for the defensive panic logic.
func TestParseRefUnchecked_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ParseRefUnchecked was expected to panic on invalid input, but it did not")
		}
	}()

	_ = ParseRefUnchecked("http://example.com/%")
}

// TestRef_RelativeScheme verifies that calling Scheme() on a relative reference
// correctly indicates that no scheme is present.
func TestRef_RelativeScheme(t *testing.T) {
	t.Parallel()

	ref, err := ParseRef("/path/only?q=1")
	if err != nil {
		t.Fatalf("ParseRef() failed unexpectedly for a valid relative ref: %v", err)
	}

	scheme, ok := ref.Scheme()

	if ok {
		t.Error("Scheme() returned ok=true for a relative reference, want false")
	}
	if scheme != "" {
		t.Errorf("Scheme() returned scheme=%q for a relative reference, want empty string", scheme)
	}
}

// TestParseIriUnchecked_Panic verifies that ParseIriUnchecked panics when called
// with a relative IRI, as is its documented behavior.
func TestParseIriUnchecked_Panic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ParseIriUnchecked was expected to panic on a relative IRI, but it did not")
		}
	}()

	_ = ParseIriUnchecked("//a/b/c")
}

// TestRelativizeWithFragmentAndQuery tests the Relativize method for cases where a
// relative path must be constructed and the target IRI contains a query and/or a
// fragment.
func TestRelativizeWithFragmentAndQuery(t *testing.T) {
	t.Parallel()
	base, _ := ParseIri("http://example.com/foo/bar")
	target, _ := ParseIri("http://example.com/foo/baz#section")

	rel, err := base.Relativize(target)
	if err != nil {
		t.Fatalf("Relativize failed: %v", err)
	}

	expected := "baz#section"
	if rel.String() != expected {
		t.Errorf("Relativizing %q against %q gives %q, want %q",
			target.String(), base.String(), rel.String(), expected)
	}

	targetWithQuery, _ := ParseIri("http://example.com/foo/baz?q=1#section")
	relWithQuery, err := base.Relativize(targetWithQuery)
	if err != nil {
		t.Fatalf("Relativize with query failed: %v", err)
	}
	expectedWithQuery := "baz?q=1#section"
	if relWithQuery.String() != expectedWithQuery {
		t.Errorf("Relativizing %q against %q gives %q, want %q",
			targetWithQuery.String(), base.String(), relWithQuery.String(), expectedWithQuery)
	}
}

// TestResolveUncheckedTo_Panic verifies that ResolveUncheckedTo panics when the
// underlying parser returns an error, even in unchecked mode. This is a test
// for the defensive panic logic.
func TestResolveUncheckedTo_Panic(t *testing.T) {
	t.Parallel()
	base := ParseIriUnchecked("http://a/b")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ResolveUncheckedTo was expected to panic on invalid input, but it did not")
		}
	}()

	var builder strings.Builder
	invalidRelativeIRI := "http://example.com/%"
	base.ResolveUncheckedTo(invalidRelativeIRI, &builder)
}
