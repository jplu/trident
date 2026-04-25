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

import "testing"

// componentTestCase holds the expected values for each IRI component.
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

// assertComponents checks if the components of a Ref match the expected values.
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

// TestRef_String tests that String() returns the original parsed string.
// RFC 3987 Section 2: "an IRI is defined as a sequence of characters".
func TestRef_String(t *testing.T) {
	iriStr := "http://example.com/path?query#fragment"
	ref := mustParseRef(t, iriStr)
	if ref.String() != iriStr {
		t.Errorf("Expected String() to return '%s', got '%s'", iriStr, ref.String())
	}
}

// TestRef_ComponentAccessors tests the accessor methods for IRI components on a Ref.
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

// TestIri_Scheme tests the Scheme accessor on the Iri type.
func TestIri_Scheme(t *testing.T) {
	iri := mustParseIri(t, "https://example.com")
	if s := iri.Scheme(); s != "https" {
		t.Errorf("Expected scheme 'https', got '%s'", s)
	}
}
