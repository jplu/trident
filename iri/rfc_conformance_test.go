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
	"encoding/json"
	"os"
	"testing"
)

// ParsedComponents represents the expected optional deconstructed components of the tested URI reference.
type ParsedComponents struct {
	Scheme    *string `json:"scheme"`
	Authority *string `json:"authority"`
	Path      *string `json:"path"`
	Query     *string `json:"query"`
	Fragment  *string `json:"fragment"`
}

// Resolution represents a resolution test scenario mapping a base URI to the expected resolved output.
type Resolution struct {
	BaseURI     string `json:"base_uri"`
	ResolvedURI string `json:"resolved_uri"`
}

// TestCase represents the structured data mapping directly to the JSON conformance file schema.
type TestCase struct {
	ID               string            `json:"id"`
	Category         string            `json:"category"`
	URIReference     string            `json:"uri_reference"`
	ParsedComponents *ParsedComponents `json:"parsed_components"`
	Resolutions      []Resolution      `json:"resolutions"`
}

// loadTestCases searches for and deserializes the generated JSON conformance file.
func loadTestCases() ([]TestCase, error) {
	path := "../tests/uri_tests.json"

	var data []byte
	var err error
	data, err = os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cases []TestCase
	if errUnmarshal := json.Unmarshal(data, &cases); errUnmarshal != nil {
		return nil, errUnmarshal
	}
	return cases, nil
}

// verifyOptionalComponent evaluates an optional parsed component against its expected value.
func verifyOptionalComponent(
	t *testing.T,
	id, label, title string,
	expected *string,
	actual string,
	present bool,
) {
	if expected == nil {
		if present {
			t.Errorf("[%s] Expected no %s, got %q", id, label, actual)
		}
	} else {
		if !present {
			t.Errorf("[%s] Expected %s %q, got none", id, label, *expected)
		} else if actual != *expected {
			t.Errorf("[%s] %s mismatch: expected %q, got %q", id, title, *expected, actual)
		}
	}
}

// verifyComponents evaluates parsed components from the reference against the expected JSON definitions.
func verifyComponents(t *testing.T, tc TestCase, ref *Ref) {
	expected := tc.ParsedComponents
	if expected == nil {
		return
	}

	actualScheme, hasScheme := ref.Scheme()
	verifyOptionalComponent(t, tc.ID, "scheme", "Scheme", expected.Scheme, actualScheme, hasScheme)

	actualAuth, hasAuth := ref.Authority()
	verifyOptionalComponent(
		t,
		tc.ID,
		"authority",
		"Authority",
		expected.Authority,
		actualAuth,
		hasAuth,
	)

	actualPath := ref.Path()
	if expected.Path != nil {
		if actualPath != *expected.Path {
			t.Errorf("[%s] Path mismatch: expected %q, got %q", tc.ID, *expected.Path, actualPath)
		}
	}

	actualQuery, hasQuery := ref.Query()
	verifyOptionalComponent(t, tc.ID, "query", "Query", expected.Query, actualQuery, hasQuery)

	actualFrag, hasFrag := ref.Fragment()
	verifyOptionalComponent(
		t,
		tc.ID,
		"fragment",
		"Fragment",
		expected.Fragment,
		actualFrag,
		hasFrag,
	)
}

// resolveAndVerify evaluates resolution for a given base and reference against expectations.
func resolveAndVerify(t *testing.T, tc TestCase, res Resolution, strict bool) {
	baseRef, err := ParseRef(res.BaseURI)
	if err != nil {
		if strict {
			t.Errorf("[%s] Strict parse failed for base URI %q: %v", tc.ID, res.BaseURI, err)
		} else {
			t.Logf("[%s] Skipped resolution due to base URI parse failure: %q", tc.ID, res.BaseURI)
		}
		return
	}

	resolvedRef, err := baseRef.Resolve(tc.URIReference)
	if err != nil {
		if strict {
			t.Errorf(
				"[%s] Strict resolution failed for reference %q against %q: %v",
				tc.ID,
				tc.URIReference,
				res.BaseURI,
				err,
			)
		} else {
			t.Logf("[%s] Skipped resolution failure: %q against %q", tc.ID, tc.URIReference, res.BaseURI)
		}
		return
	}

	actualResolved := resolvedRef.String()
	if actualResolved != res.ResolvedURI {
		if strict {
			t.Errorf(
				"[%s] Resolution mismatch!\nBase:     %q\nRef:      %q\nExpected: %q\nActual:   %q",
				tc.ID,
				res.BaseURI,
				tc.URIReference,
				res.ResolvedURI,
				actualResolved,
			)
		} else {
			t.Logf(
				"[%s] Lenient/Browser-specific discrepancy ignored:\nBase:     %q\nRef:      %q\nExpected: %q\nActual:   %q",
				tc.ID,
				res.BaseURI,
				tc.URIReference,
				res.ResolvedURI,
				actualResolved,
			)
		}
	}
}

// runTestCase parses the reference and triggers validation for its components and resolutions.
func runTestCase(t *testing.T, tc TestCase, strict bool) {
	ref, err := ParseRef(tc.URIReference)
	if err != nil {
		if strict {
			t.Errorf("[%s] Strict parse failed for reference %q: %v", tc.ID, tc.URIReference, err)
		} else {
			t.Logf("[%s] Skipped lenient reference parsing: %q (Error: %v)", tc.ID, tc.URIReference, err)
		}
		return
	}

	if strict {
		verifyComponents(t, tc, ref)
	}

	for _, res := range tc.Resolutions {
		resolveAndVerify(t, tc, res, strict)
	}
}

// TestRFCSuite runs the complete set of RFC conformance test cases loaded from the JSON file.
func TestRFCSuite(t *testing.T) {
	cases, err := loadTestCases()
	if err != nil {
		t.Fatalf("Could not load conformance test cases: %v", err)
	}

	strictCategories := map[string]bool{
		"RFC3986":           true,
		"RFC2397":           true,
		"RFC3987":           true,
		"Some random tests": true,
	}

	for _, tc := range cases {
		strict := strictCategories[tc.Category]
		runTestCase(t, tc, strict)
	}
}
