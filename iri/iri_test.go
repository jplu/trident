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
	relative, base, expected string
}

// TestResolveRelativeIRI contains a comprehensive suite of resolution tests,
// many drawn from RFC 3986 examples, to verify the correctness of the
// reference resolution algorithm.
func TestResolveRelativeIRI(t *testing.T) {
	t.Parallel()
	examples := []resolveTest{
		{"/.", "http://a/b/c/d;p?q", "http://a/"},
		{"/.foo", "http://a/b/c/d;p?q", "http://a/.foo"},
		{".foo", "http://a/b/c/d;p?q", "http://a/b/c/.foo"},
		{"g:h", "http://a/b/c/d;p?q", "g:h"},
		{"g", "http://a/b/c/d;p?q", "http://a/b/c/g"},
		{"./g", "http://a/b/c/d;p?q", "http://a/b/c/g"},
		{"g/", "http://a/b/c/d;p?q", "http://a/b/c/g/"},
		{"/g", "http://a/b/c/d;p?q", "http://a/g"},
		{"//g", "http://a/b/c/d;p?q", "http://g"},
		{"?y", "http://a/b/c/d;p?q", "http://a/b/c/d;p?y"},
		{"g?y", "http://a/b/c/d;p?q", "http://a/b/c/g?y"},
		{"#s", "http://a/b/c/d;p?q", "http://a/b/c/d;p?q#s"},
		{"g#s", "http://a/b/c/d;p?q", "http://a/b/c/g#s"},
		{"g?y#s", "http://a/b/c/d;p?q", "http://a/b/c/g?y#s"},
		{";x", "http://a/b/c/d;p?q", "http://a/b/c/;x"},
		{"g;x", "http://a/b/c/d;p?q", "http://a/b/c/g;x"},
		{"g;x?y#s", "http://a/b/c/d;p?q", "http://a/b/c/g;x?y#s"},
		{"", "http://a/b/c/d;p?q", "http://a/b/c/d;p?q"},
		{".", "http://a/b/c/d;p?q", "http://a/b/c/"},
		{"./", "http://a/b/c/d;p?q", "http://a/b/c/"},
		{"..", "http://a/b/c/d;p?q", "http://a/b/"},
		{"../", "http://a/b/c/d;p?q", "http://a/b/"},
		{"../g", "http://a/b/c/d;p?q", "http://a/b/g"},
		{"../..", "http://a/b/c/d;p?q", "http://a/"},
		{"../../", "http://a/b/c/d;p?q", "http://a/"},
		{"../../g", "http://a/b/c/d;p?q", "http://a/g"},
		{"/./g", "http://a/b/c/d;p?q", "http://a/g"},
		{"/../g", "http://a/b/c/d;p?q", "http://a/g"},
		{"g.", "http://a/b/c/d;p?q", "http://a/b/c/g."},
		{".g", "http://a/b/c/d;p?q", "http://a/b/c/.g"},
		{"g..", "http://a/b/c/d;p?q", "http://a/b/c/g.."},
		{"..g", "http://a/b/c/d;p?q", "http://a/b/c/..g"},
		{"./../g", "http://a/b/c/d;p?q", "http://a/b/g"},
		{"./g/.", "http://a/b/c/d;p?q", "http://a/b/c/g/"},
		{"g/./h", "http://a/b/c/d;p?q", "http://a/b/c/g/h"},
		{"g/../h", "http://a/b/c/d;p?q", "http://a/b/c/h"},
		{"http:g", "http://a/b/c/d;p?q", "http:g"},
		{"http:", "http://a/b/c/d;p?q", "http:"},
		{"../r", "http://ex/x/y/z", "http://ex/x/r"},
		{"q/r", "http://ex/x/y", "http://ex/x/q/r"},
		{"q/r#s", "http://ex/x/y", "http://ex/x/q/r#s"},
		{"z/", "http://ex/x/y/", "http://ex/x/y/z/"},
		{"#Animal", "file:/swap/test/animal.rdf", "file:/swap/test/animal.rdf#Animal"},
		{"/r", "file:/ex/x/y/z", "file:/r"},
		{"s", "http://example.com", "http://example.com/s"},
		{"g/./h", "http://a/b/c/d;p?q", "http://a/b/c/g/h"},
		{"g/../h", "http://a/b/c/d;p?q", "http://a/b/c/h"},
		{"g;x=1/./y", "http://a/b/c/d;p?q", "http://a/b/c/g;x=1/y"},
		{"g;x=1/../y", "http://a/b/c/d;p?q", "http://a/b/c/y"},
		{"g?y/./x", "http://a/b/c/d;p?q", "http://a/b/c/g?y/./x"},
		{"g?y/../x", "http://a/b/c/d;p?q", "http://a/b/c/g?y/../x"},
		{"g#s/./x", "http://a/b/c/d;p?q", "http://a/b/c/g#s/./x"},
		{"g#s/../x", "http://a/b/c/d;p?q", "http://a/b/c/g#s/../x"},
		{"/a/b/c/./../../g", "http://a/b/c/d;p?q", "http://a/a/g"},
		{"g", "http://a/b/c/d;p?q=1/2", "http://a/b/c/g"},
		{"./g", "http://a/b/c/d;p?q=1/2", "http://a/b/c/g"},
		{"g/", "http://a/b/c/d;p?q=1/2", "http://a/b/c/g/"},
		{"/g", "http://a/b/c/d;p?q=1/2", "http://a/g"},
		{"//g", "http://a/b/c/d;p?q=1/2", "http://g"},
		{"?y", "http://a/b/c/d;p?q=1/2", "http://a/b/c/d;p?y"},
		{"g?y", "http://a/b/c/d;p?q=1/2", "http://a/b/c/g?y"},
		{"g?y/./x", "http://a/b/c/d;p?q=1/2", "http://a/b/c/g?y/./x"},
		{"g?y/../x", "http://a/b/c/d;p?q=1/2", "http://a/b/c/g?y/../x"},
		{"g#s", "http://a/b/c/d;p?q=1/2", "http://a/b/c/g#s"},
		{"g#s/./x", "http://a/b/c/d;p?q=1/2", "http://a/b/c/g#s/./x"},
		{"g#s/../x", "http://a/b/c/d;p?q=1/2", "http://a/b/c/g#s/../x"},
		{"./", "http://a/b/c/d;p?q=1/2", "http://a/b/c/"},
		{"../", "http://a/b/c/d;p?q=1/2", "http://a/b/"},
		{"../g", "http://a/b/c/d;p?q=1/2", "http://a/b/g"},
		{"../../", "http://a/b/c/d;p?q=1/2", "http://a/"},
		{"../../g", "http://a/b/c/d;p?q=1/2", "http://a/g"},
		{"g", "http://a/b/c/d;p=1/2?q", "http://a/b/c/d;p=1/g"},
		{"./g", "http://a/b/c/d;p=1/2?q", "http://a/b/c/d;p=1/g"},
		{"g/", "http://a/b/c/d;p=1/2?q", "http://a/b/c/d;p=1/g/"},
		{"g?y", "http://a/b/c/d;p=1/2?q", "http://a/b/c/d;p=1/g?y"},
		{";x", "http://a/b/c/d;p=1/2?q", "http://a/b/c/d;p=1/;x"},
		{"g;x", "http://a/b/c/d;p=1/2?q", "http://a/b/c/d;p=1/g;x"},
		{"g;x=1/./y", "http://a/b/c/d;p=1/2?q", "http://a/b/c/d;p=1/g;x=1/y"},
		{"g;x=1/../y", "http://a/b/c/d;p=1/2?q", "http://a/b/c/d;p=1/y"},
		{"./", "http://a/b/c/d;p=1/2?q", "http://a/b/c/d;p=1/"},
		{"../", "http://a/b/c/d;p=1/2?q", "http://a/b/c/"},
		{"../g", "http://a/b/c/d;p=1/2?q", "http://a/b/c/g"},
		{"../../", "http://a/b/c/d;p=1/2?q", "http://a/b/"},
		{"../../g", "http://a/b/c/d;p=1/2?q", "http://a/b/g"},
		{"g:h", "fred:///s//a/b/c", "g:h"},
		{"g", "fred:///s//a/b/c", "fred:///s//a/b/g"},
		{"./g", "fred:///s//a/b/c", "fred:///s//a/b/g"},
		{"g/", "fred:///s//a/b/c", "fred:///s//a/b/g/"},
		{"/g", "fred:///s//a/b/c", "fred:///g"},
		{"//g", "fred:///s//a/b/c", "fred://g"},
		{"//g/x", "fred:///s//a/b/c", "fred://g/x"},
		{"///g", "fred:///s//a/b/c", "fred:///g"},
		{"./", "fred:///s//a/b/c", "fred:///s//a/b/"},
		{"../", "fred:///s//a/b/c", "fred:///s//a/"},
		{"../g", "fred:///s//a/b/c", "fred:///s//a/g"},
		{"../../", "fred:///s//a/b/c", "fred:///s//"},
		{"../../g", "fred:///s//a/b/c", "fred:///s//g"},
		{"../../../g", "fred:///s//a/b/c", "fred:///s/g"},
		{"../../../../g", "fred:///s//a/b/c", "fred:///g"},
		{"g:h", "http:///s//a/b/c", "g:h"},
		{"g", "http:///s//a/b/c", "http:///s//a/b/g"},
		{"./g", "http:///s//a/b/c", "http:///s//a/b/g"},
		{"g/", "http:///s//a/b/c", "http:///s//a/b/g/"},
		{"/g", "http:///s//a/b/c", "http:///g"},
		{"//g", "http:///s//a/b/c", "http://g"},
		{"//g/x", "http:///s//a/b/c", "http://g/x"},
		{"///g", "http:///s//a/b/c", "http:///g"},
		{"./", "http:///s//a/b/c", "http:///s//a/b/"},
		{"../", "http:///s//a/b/c", "http:///s//a/"},
		{"../g", "http:///s//a/b/c", "http:///s//a/g"},
		{"../../", "http:///s//a/b/c", "http:///s//"},
		{"../../g", "http:///s//a/b/c", "http:///s//g"},
		{"../../../g", "http:///s//a/b/c", "http:///s/g"},
		{"../../../../g", "http:///s//a/b/c", "http:///g"},
		{"bar:abc", "foo:xyz", "bar:abc"},
		{"../abc", "http://example/x/y/z", "http://example/x/abc"},
		{"http://example/x/abc", "http://example2/x/y/z", "http://example/x/abc"},
		{"q/r#s/t", "http://ex/x/y", "http://ex/x/q/r#s/t"},
		{"ftp://ex/x/q/r", "http://ex/x/y", "ftp://ex/x/q/r"},
		{"", "http://ex/x/y", "http://ex/x/y"},
		{"", "http://ex/x/y/", "http://ex/x/y/"},
		{"", "http://ex/x/y/pdq", "http://ex/x/y/pdq"},
		{"../abc", "file:/e/x/y/z", "file:/e/x/abc"},
		{"/example/x/abc", "file:/example2/x/y/z", "file:/example/x/abc"},
		{"../r", "file:/ex/x/y/z", "file:/ex/x/r"},
		{"q/r", "file:/ex/x/y", "file:/ex/x/q/r"},
		{"q/r#s", "file:/ex/x/y", "file:/ex/x/q/r#s"},
		{"q/r#", "file:/ex/x/y", "file:/ex/x/q/r#"},
		{"q/r#s/t", "file:/ex/x/y", "file:/ex/x/q/r#s/t"},
		{"ftp://ex/x/q/r", "file:/ex/x/y", "ftp://ex/x/q/r"},
		{"", "file:/ex/x/y", "file:/ex/x/y"},
		{"", "file:/ex/x/y/", "file:/ex/x/y/"},
		{"", "file:/ex/x/y/pdq", "file:/ex/x/y/pdq"},
		{"z/", "file:/ex/x/y/", "file:/ex/x/y/z/"},
		{
			"file://meetings.example.com/cal#m1",
			"file:/devel/WWW/2000/10/swap/test/reluri-1.n3",
			"file://meetings.example.com/cal#m1",
		},
		{
			"file://meetings.example.com/cal#m1",
			"file:/home/connolly/w3ccvs/WWW/2000/10/swap/test/reluri-1.n3",
			"file://meetings.example.com/cal#m1",
		},
		{"./#blort", "file:/some/dir/foo", "file:/some/dir/#blort"},
		{"./#", "file:/some/dir/foo", "file:/some/dir/#"},
		{"./", "http://example/x/abc.efg", "http://example/x/"},
		{"./q:r", "http://ex/x/y", "http://ex/x/q:r"},
		{"./p=q:r", "http://ex/x/y", "http://ex/x/p=q:r"},
		{"?pp/rr", "http://ex/x/y?pp/qq", "http://ex/x/y?pp/rr"},
		{"y/z", "http://ex/x/y?pp/qq", "http://ex/x/y/z"},
		{"y?q", "http://ex/x/y?q", "http://ex/x/y?q"},
		{"/x/y?q", "http://ex?p", "http://ex/x/y?q"},
		{"c/d", "foo:a/b", "foo:a/c/d"},
		{"/c/d", "foo:a/b", "foo:/c/d"},
		{"", "foo:a/b?c#d", "foo:a/b?c"},
		{"b/c", "foo:a", "foo:b/c"},
		{"../b/c", "foo:/a/y/z", "foo:/a/b/c"},
		{"./b/c", "foo:a", "foo:b/c"},
		{"/./b/c", "foo:a", "foo:/b/c"},
		{"../../d", "foo://a//b/c", "foo://a/d"},
		{".", "foo:a", "foo:"},
		{"..", "foo:a", "foo:"},
		{"abc", "http://example/x/y%2Fz", "http://example/x/abc"},
		{"../../x%2Fabc", "http://example/a/x/y/z", "http://example/a/x%2Fabc"},
		{"../x%2Fabc", "http://example/a/x/y%2Fz", "http://example/a/x%2Fabc"},
		{"abc", "http://example/x%2Fy/z", "http://example/x%2Fy/abc"},
		{"q%3Ar", "http://ex/x/y", "http://ex/x/q%3Ar"},
		{"/x%2Fabc", "http://example/x/y%2Fz", "http://example/x%2Fabc"},
		{"/x%2Fabc", "http://example/x/y/z", "http://example/x%2Fabc"},
		{"http://example/a/b?c/../d", "foo:bar", "http://example/a/b?c/../d"},
		{"http://example/a/b#c/../d", "foo:bar", "http://example/a/b#c/../d"},
		{"http:this", "http://example.org/base/uri", "http:this"},
		{"http:this", "http:base", "http:this"},
		{
			"mini1.xml",
			"file:///C:/DEV/Haskell/lib/HXmlToolbox-3.01/examples/",
			"file:///C:/DEV/Haskell/lib/HXmlToolbox-3.01/examples/mini1.xml",
		},
		{"?bar", "file:foo", "file:foo?bar"},
		{"#bar", "file:foo", "file:foo#bar"},
		{"/lv2.h", "file:foo", "file:/lv2.h"},
		{"///lv2.h", "file:foo", "file:///lv2.h"},
		{"lv2.h", "file:foo", "file:lv2.h"},
		{".", "file:", "file:"},
		{"..", "file:", "file:"},
		{"./", "file:", "file:"},
		{"../", "file:", "file:"},
		{"./.", "file:", "file:"},
		{"../..", "file:", "file:"},
		{
			"http:./examplxm+ns/Seq/exhttpwsa//DtaAccnss/tencile#frag",
			"http://foo",
			"http:./examplxm+ns/Seq/exhttpwsa//DtaAccnss/tencile#frag",
		},
	}

	for _, test := range examples {
		name := fmt.Sprintf("%s_RESOLVE_%s", test.base, test.relative)
		t.Run(name, func(t *testing.T) {
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
		{"", "http:", "http:"},
		{"", "http://example.com", "http://example.com"},
		{"", "http://example.com/foo", "http://example.com/foo"},
		{"", "http://example.com/foo/bar", "http://example.com/foo/bar"},
		{"", "http://example.com/foo/bar?bat", "http://example.com/foo/bar?bat"},
		{"#baz", "http://example.com/foo/bar?bat#baz", "http://example.com/foo/bar?bat#baz"},
		{"http:", "http:", "https:"},
		{"//example.com", "http://example.com", "http://example.org"},
		{"foo", "http://example.com/foo", "http://example.com/bar"},
		{"?bat", "http://example.com/foo?bat", "http://example.com/foo?foo"},
		{"#baz", "http://example.com/foo?bat#baz", "http://example.com/foo?bat#foo"},
		{"//example.com", "http://example.com", "http:"},
		{"//example.com", "http://example.com", "http://"},
		{"foo", "http://example.com/foo", "http://example.com/"},
		{"/foo", "http://example.com/foo", "http://example.com/bar/baz"},
		{"bar", "http://example.com/foo/bar", "http://example.com/foo/baz"},
		{"foo/bar", "http://example.com/foo/bar", "http://example.com/foo"},
		{"?bar", "http://example.com/foo?bar", "http://example.com/foo?baz"},
		{"//example.com?bar", "http://example.com?bar", "http://example.com/a"},
		{"?bar", "http://example.com?bar", "http://example.com"},
		{"//example.com?bar", "http://example.com?bar", "http://example.com/"},
		{"#bar", "http://example.com/foo#bar", "http://example.com/foo#baz"},
		{".", "http://example.com/foo/", "http://example.com/foo/bar"},
		{"/:", "http://example.com/:", "http://example.com/foo"},
		{"http:", "http:", "http://example.com"},
		{"http:?foo", "http:?foo", "http://example.com"},
		{"//example.com", "http://example.com", "http://example.com/foo"},
		{"//example.com", "http://example.com", "http://example.com?query"},
		{"foo", "http://example.com/foo", "http://example.com/foo?query"},
		{"http:?query", "http:?query", "http://example.com?query"},
		{"http:/path", "http:/path", "http://example.com/foo"},
		{"//example.com//a", "http://example.com//a", "http://example.com/"},
		{"ab", "urn:ab", "urn:"},
		{"urn:isbn:foo", "urn:isbn:foo", "urn:"},
		{"is/bn:foo", "urn:is/bn:foo", "urn:"},
		{"e/p", "t:e/e/p", "t:e/s"},
		{"gp", "htt:/foo/gp", "htt:/foo/"},
		{"gp", "htt:/gp", "htt:/"},
		{"x:", "x:", "x://foo"},
		{"x:", "x:", "x:02"},
		{"x:", "x:", "x:?foo"},
		{"", "http://example.com", "http://example.com#foo"},
		{".", "http://example.com/a/", "http://example.com/a/b"},
		{".?c", "http://example.com/a/?c", "http://example.com/a/b"},
		{"t:o//", "t:o//", "t:o/"},
	}

	for _, test := range examples {
		t.Run(fmt.Sprintf("%s_RELATIVIZE_%s", test.base, test.expected), func(t *testing.T) {
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
	tests := []struct{ iri, base string }{
		{"http://example.com/a/../b", "http://example.com/s"},
		{"http://example.com/a/..", "http://example.com/s"},
		{"http://example.com/./b", "http://example.com/s"},
		{"http://example.com/.", "http://example.com/s"},
		{"urn:.", "urn:"},
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
		{"../foo", "http://host/", "http://host/foo"},
		{"../foo", "http://host/xyz", "http://host/foo"},
		{"d/z?x=a", "http://www.example.org/a/b/c/d", "http://www.example.org/a/b/c/d/z?x=a"},
		{"http://example.com/A", "http://www.example.org/a/b/c/d", "http://example.com/A"},
		{"", "http://www.example.org/a/b/c/d/", "http://www.example.org/a/b/c/d/"},
		{".", "http://www.example.org/a/b/c/d/", "http://www.example.org/a/b/c/d/"},
		{"../../C/D", "http://www.example.org/a/b/c/d/", "http://www.example.org/a/b/C/D"},
		{"../../c/d/", "http://www.example.org/a/b/c/d/", "http://www.example.org/a/b/c/d/"},
		{"../../c/d/X#bar", "http://www.example.org/a/b/c/d/", "http://www.example.org/a/b/c/d/X#bar"},
		{"../../c/d/e/f/g/", "http://www.example.org/a/b/c/d/", "http://www.example.org/a/b/c/d/e/f/g/"},
		{"../../c/d/z?x=a", "http://www.example.org/a/b/c/d/", "http://www.example.org/a/b/c/d/z?x=a"},
		{
			"http://example.org/#André",
			"http://www.w3.org/2000/10/rdf-tests/rdfcore/rdf-charmod-uris/test001.rdf",
			"http://example.org/#André",
		},
		{
			"http://example.org/#Andr%C3%A9",
			"http://www.w3.org/2000/10/rdf-tests/rdfcore/rdf-charmod-uris/test002.rdf",
			"http://example.org/#Andr%C3%A9",
		},
		{
			"#Dürst",
			"http://www.w3.org/2000/10/rdf-tests/rdfcore/rdfms-difference-between-ID-and-about/test2.rdf",
			"http://www.w3.org/2000/10/rdf-tests/rdfcore/rdfms-difference-between-ID-and-about/test2.rdf#Dürst",
		},
		{"#", "base:x", "base:x#"},
		{
			"",
			"file:///C:/Documents and Settings/jjchplb/Local Settings/Temp/test-load-with-41.rdf",
			"file:///C:/Documents and Settings/jjchplb/Local Settings/Temp/test-load-with-41.rdf",
		},
		{"eh:/a", "file:///C:/Documents and Settings/jjchplb/Local Settings/Temp/test-load-with-41.rdf", "eh:/a"},
		{"#", "file:///C:/eclipse/workspace/jena2/", "file:///C:/eclipse/workspace/jena2/#"},
		{"", "file:///C:/eclipse/workspace/jena2/", "file:///C:/eclipse/workspace/jena2/"},
		{"base", "file:///C:/eclipse/workspace/jena2/", "file:///C:/eclipse/workspace/jena2/base"},
		{"eh://R", "file:///C:/eclipse/workspace/jena2/", "eh://R"},
		{"eh:/O", "file:///C:/eclipse/workspace/jena2/", "eh:/O"},
		{"rdf://test.com#", "file:///C:/eclipse/workspace/jena2/", "rdf://test.com#"},
		{"z", "file:///C:/eclipse/workspace/jena2/foo.n3", "file:///C:/eclipse/workspace/jena2/z"},
		{
			"",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/Ask/manifest.ttl",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/Ask/manifest.ttl",
		},
		{
			"r-base-prefix-3.ttl",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/Basic/manifest.ttl",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/Basic/r-base-prefix-3.ttl",
		},
		{
			"r-base-prefix-4.ttl",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/Basic/manifest.ttl",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/Basic/r-base-prefix-4.ttl",
		},
		{
			"mailto:bert@example.net",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/Optional/result-opt-1.ttl",
			"mailto:bert@example.net",
		},
		{
			"Bound/manifest.n3",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/manifest-arq.ttl",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/Bound/manifest.n3",
		},
		{
			"Construct/manifest.ttl",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/manifest-arq.ttl",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/Construct/manifest.ttl",
		},
		{
			"Dataset/manifest.n3",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/manifest-arq.ttl",
			"file:///C:/eclipse/workspace/jena2/testing/ARQ/Dataset/manifest.n3",
		},
		{
			"mailto:jlow@example.com",
			"file:///C:/eclipse/workspace/jena2/testing/DAWG-Approved/examples/ex2-4a.n3",
			"mailto:jlow@example.com",
		},
		{
			"ex11.2.3.2_0.rq",
			"file:///C:/eclipse/workspace/jena2/testing/DAWG/examples/manifest.n3",
			"file:///C:/eclipse/workspace/jena2/testing/DAWG/examples/ex11.2.3.2_0.rq",
		},
		{
			"urn:/*not_a_comment*/",
			"file:///C:/eclipse/workspace/jena2/testing/RDQL-ARQ/result-0-01.n3",
			"urn:/*not_a_comment*/",
		},
		{
			"#y1",
			"file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/test_hk_06/b.owl",
			"file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/test_hk_06/b.owl#y1",
		},
		{
			"",
			"file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/test_hk_06/b.owl",
			"file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/test_hk_06/b.owl",
		},
		{
			"foo#ClassAC",
			"file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/test_hk_07A.owl",
			"file:///C:/eclipse/workspace/jena2/testing/ontology/bugs/foo#ClassAC",
		},
		{
			"jason6",
			"file:///C:/eclipse/workspace/jena2/testing/reasoners/bugs/sbug.rdf",
			"file:///C:/eclipse/workspace/jena2/testing/reasoners/bugs/jason6",
		},
		{
			"urn:x-propNum100",
			"file:///C:/eclipse/workspace/jena2/testing/reasoners/bugs/subpropertyModel.n3",
			"urn:x-propNum100",
		},
		{"eh:/V", "file:///C:/eclipse/workspace/jena2/testing/reasoners/bugs/unbroken.n3", "eh:/V"},
		{"eh:/a", "file:///C:/eclipse/workspace/jena2/testing/reasoners/bugs/unbroken.n3", "eh:/a"},
		{
			"",
			"file:C:\\DOCUME~1\\jjchplb\\LOCALS~1\\Temp\\hedgehog6739.rdf",
			"file:C:\\DOCUME~1\\jjchplb\\LOCALS~1\\Temp\\hedgehog6739.rdf",
		},
		{"http://spoo.net/O", "file:C:\\DOCUME~1\\jjchplb\\LOCALS~1\\Temp\\hedgehog6739.rdf", "http://spoo.net/O"},
		{"http://spoo.net/S", "file:C:\\DOCUME~1\\jjchplb\\LOCALS~1\\Temp\\hedgehog6739.rdf", "http://spoo.net/S"},
		{"urn:x-hp:eg/", "file:doc/inference/data/owlDemoSchema.xml", "urn:x-hp:eg/"},
		{"", "file:testing/abbreviated/relative-uris.rdf", "file:testing/abbreviated/relative-uris.rdf"},
		{".", "file:testing/abbreviated/relative-uris.rdf", "file:testing/abbreviated/"},
		{"../../C/D", "file:testing/abbreviated/relative-uris.rdf", "file:C/D"},
		{"//example.com/A", "file:testing/abbreviated/relative-uris.rdf", "file://example.com/A"},
		{"/A/B#foo/", "file:testing/abbreviated/relative-uris.rdf", "file:/A/B#foo/"},
		{"X#bar", "file:testing/abbreviated/relative-uris.rdf", "file:testing/abbreviated/X#bar"},
		{"e/f/g/", "file:testing/abbreviated/relative-uris.rdf", "file:testing/abbreviated/e/f/g/"},
		{
			"http://www.example.org/a/b/c/d/",
			"file:testing/abbreviated/relative-uris.rdf",
			"http://www.example.org/a/b/c/d/",
		},
		{"z?x=a", "file:testing/abbreviated/relative-uris.rdf", "file:testing/abbreviated/z?x=a"},
		{"", "file:testing/arp/error-msgs/test06.rdf", "file:testing/arp/error-msgs/test06.rdf"},
		{"#one", "file:testing/arp/qname-in-ID/bug74_0.rdf", "file:testing/arp/qname-in-ID/bug74_0.rdf#one"},
		{"#sw:test", "file:testing/arp/qname-in-ID/bug74_0.rdf", "file:testing/arp/qname-in-ID/bug74_0.rdf#sw:test"},
		{
			"http://localhost:8080/Repository/QueryAgent/UserOntology/qgen-example-1#",
			"file:testing/ontology/bugs/test_dk_01.xml",
			"http://localhost:8080/Repository/QueryAgent/UserOntology/qgen-example-1#",
		},
		{"owl#Thing", "file:testing/ontology/bugs/test_dk_01.xml", "file:testing/ontology/bugs/owl#Thing"},
		{"#__rest3", "file:testing/ontology/bugs/test_oh_01.owl", "file:testing/ontology/bugs/test_oh_01.owl#__rest3"},
		{
			"#Union2",
			"file:testing/ontology/owl/list-syntax/test-ldp.rdf",
			"file:testing/ontology/owl/list-syntax/test-ldp.rdf#Union2",
		},
		{"urn:foo", "file:testing/reasoners/bugs/cardFPTest.owl", "urn:foo"},
		{
			"http://decsai.ugr.es/~ontoserver/bacarex2.owl",
			"file:testing/reasoners/bugs/deleteBug.owl",
			"http://decsai.ugr.es/~ontoserver/bacarex2.owl",
		},
		{
			"#A",
			"file:testing/reasoners/bugs/equivalentClassTest.owl",
			"file:testing/reasoners/bugs/equivalentClassTest.owl#A",
		},
		{"NC:ispinfo", "http://bar.com/irrelevant", "NC:ispinfo"},
		{"NC:trickMe", "http://bar.com/irrelevant", "NC:trickMe"},
		{
			"chrome://messenger/content/mailPrefsOverlay.xul",
			"http://bar.com/irrelevant",
			"chrome://messenger/content/mailPrefsOverlay.xul",
		},
		{"domain:aol.com", "http://bar.com/irrelevant", "domain:aol.com"},
		{"http://foo.com/    ", "http://bar.com/irrelevant", "http://foo.com/    "},
		{"http://foo.com/   ", "http://bar.com/irrelevant", "http://foo.com/   "},
		{"http://foo.com/  ", "http://bar.com/irrelevant", "http://foo.com/  "},
		{"http://foo.com/ ", "http://bar.com/irrelevant", "http://foo.com/ "},
		{"http://foo.com/\t", "http://bar.com/irrelevant", "http://foo.com/\t"},
		{"http://foo.com/\n\n", "http://bar.com/irrelevant", "http://foo.com/\n\n"},
		{"http://foo.com/\r", "http://bar.com/irrelevant", "http://foo.com/\r"},
		{"http://foo.com/'", "http://bar.com/irrelevant", "http://foo.com/'"},
		{"http://foo.com/<b>boo", "http://bar.com/irrelevant", "http://foo.com/<b>boo"},
		{"http://foo.com/\"", "http://bar.com/irrelevant", "http://foo.com/\""},
		{"http://foo.com/", "http://bar.com/irrelevant", "http://foo.com/"},
		{"http://foo.com/", "http://bar.com/irrelevant", "http://foo.com/"},
	}

	for _, test := range examples {
		t.Run(fmt.Sprintf("%s_AGAINST_%s", test.relative, test.base), func(t *testing.T) {
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
