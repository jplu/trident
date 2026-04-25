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
package langtag

import (
	"reflect"
	"strings"
	"testing"
)

// TestGetPositions ensures that the end positions of tag components are calculated correctly.
// RFC 5646 defines a specific structure (langtag), and this function's correctness
// is critical for all accessor methods on the LanguageTag struct.
func TestGetPositions(t *testing.T) {
	cpr := &canonicalParseRun{
		language:   "zh",
		extlangs:   []string{"hak"},
		script:     "Hans",
		region:     "CN",
		variants:   []string{"variant"},
		extensions: []Extension{{Singleton: 'u', Value: "co-phonebk"}},
	}

	// zh(2) + -hak(4) + -Hans(5) + -CN(3) + -variant(8) + -u-co-phonebk(13)
	expected := tagElementsPositions{
		languageEnd:  2,
		extlangEnd:   6,
		scriptEnd:    11,
		regionEnd:    14,
		variantEnd:   22,
		extensionEnd: 35,
	}

	positions := cpr.getPositions()
	if !reflect.DeepEqual(positions, expected) {
		t.Errorf("getPositions() failed.\nGot:      %+v\nExpected: %+v", positions, expected)
	}

	cprOnlyLang := &canonicalParseRun{language: "en"}
	expectedOnlyLang := tagElementsPositions{
		languageEnd:  2,
		extlangEnd:   2,
		scriptEnd:    2,
		regionEnd:    2,
		variantEnd:   2,
		extensionEnd: 2,
	}
	positionsOnlyLang := cprOnlyLang.getPositions()
	if !reflect.DeepEqual(positionsOnlyLang, expectedOnlyLang) {
		t.Errorf(
			"getPositions() for only language failed.\nGot:      %+v\nExpected: %+v",
			positionsOnlyLang,
			expectedOnlyLang,
		)
	}

	cprEmptyExt := &canonicalParseRun{
		language:   "en",
		extensions: []Extension{{Singleton: 'a'}},
	}
	// en(2) + -a(2) = 4
	expectedEmptyExt := tagElementsPositions{
		languageEnd:  2,
		extlangEnd:   2,
		scriptEnd:    2,
		regionEnd:    2,
		variantEnd:   2,
		extensionEnd: 4,
	}
	positionsEmptyExt := cprEmptyExt.getPositions()
	if !reflect.DeepEqual(positionsEmptyExt, expectedEmptyExt) {
		t.Errorf(
			"getPositions() for empty extension failed.\nGot:      %+v\nExpected: %+v",
			positionsEmptyExt,
			expectedEmptyExt,
		)
	}
}

// TestRender ensures that the parsed components are reassembled into a correctly
// formatted string per RFC 5646, Section 2.1.1 (lowercase language, title case
// script, uppercase region).
func TestRender(t *testing.T) {
	testCases := []struct {
		name     string
		cpr      *canonicalParseRun
		expected string
	}{
		{
			name: "Full tag",
			cpr: &canonicalParseRun{
				language:   "EN",
				extlangs:   []string{"hAk"},
				script:     "hANS",
				region:     "cn",
				variants:   []string{"VaRiAnT"},
				extensions: []Extension{{Singleton: 'u', Value: "Co-PhOnEbK"}},
				privateuse: []string{"PRIV"},
				state:      stateInPrivateUse,
			},
			expected: "en-hak-Hans-CN-variant-u-co-phonebk-x-priv",
		},
		{
			name:     "Only language",
			cpr:      &canonicalParseRun{language: "de"},
			expected: "de",
		},
		{
			name:     "Language and script",
			cpr:      &canonicalParseRun{language: "sr", script: "cyrl"},
			expected: "sr-Cyrl",
		},
		{
			name:     "Language and region",
			cpr:      &canonicalParseRun{language: "en", region: "us"},
			expected: "en-US",
		},
		{
			name:     "Language and variants",
			cpr:      &canonicalParseRun{language: "sl", variants: []string{"ROZAJ", "BISKE"}},
			expected: "sl-rozaj-biske",
		},
		{
			name: "Language and extension",
			cpr: &canonicalParseRun{
				language:   "de",
				extensions: []Extension{{Singleton: 'u', Value: "co-phonebk"}},
			},
			expected: "de-u-co-phonebk",
		},
		{
			name:     "Private use only tag",
			cpr:      &canonicalParseRun{privateuse: []string{"Whatever"}},
			expected: "x-whatever",
		},
		{
			name:     "Empty CPR",
			cpr:      &canonicalParseRun{},
			expected: "",
		},
		{
			name: "Extension without value",
			cpr: &canonicalParseRun{
				language:   "fr",
				extensions: []Extension{{Singleton: 't'}},
			},
			expected: "fr-t",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			tc.cpr.render(&b)
			result := b.String()
			if result != tc.expected {
				t.Errorf("render() failed. Got '%s', expected '%s'", result, tc.expected)
			}
		})
	}
}
