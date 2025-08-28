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
package langtag

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// newTestParser creates a parser with a predefined registry for testing purposes.
// This allows for isolated testing of functions that rely on registry data
// without depending on the embedded registry.
func newTestParser(records map[string]Record) *Parser {
	return &Parser{
		registry: &Registry{
			Records:  records,
			FileDate: "2023-01-01",
		},
	}
}

// TestValidateSubtag tests the basic syntactic validation for a subtag.
func TestValidateSubtag(t *testing.T) {
	testCases := []struct {
		name    string
		subtag  string
		wantErr error
	}{
		{"Valid", "abc", nil},
		{"Valid max length", "12345678", nil},
		{"Empty", "", ErrEmptySubtag},
		{"Too long", "123456789", ErrSubtagTooLong},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateSubtag(tc.subtag); !errors.Is(err, tc.wantErr) {
				t.Errorf("validateSubtag() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestPrepareSubtags checks the helper that isolates subtags for parsing.
func TestPrepareSubtags(t *testing.T) {
	p := newTestParser(nil)
	testCases := []struct {
		name              string
		input             string
		expectedSubtags   []string
		expectedHasHyphen bool
	}{
		{"No trailing hyphen", "en-US", []string{"en", "US"}, false},
		{"With trailing hyphen", "en-US-", []string{"en", "US"}, true},
		{"Single tag", "en", []string{"en"}, false},
		{"Single tag with hyphen", "en-", []string{"en"}, true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun(tc.input, false)
			subtags, hasHyphen := cpr.prepareSubtags()
			if !reflect.DeepEqual(subtags, tc.expectedSubtags) {
				t.Errorf("prepareSubtags() subtags = %v, want %v", subtags, tc.expectedSubtags)
			}
			if hasHyphen != tc.expectedHasHyphen {
				t.Errorf("prepareSubtags() hasHyphen = %v, want %v", hasHyphen, tc.expectedHasHyphen)
			}
		})
	}
}

// TestParsePrivateUseOnly tests the specific parser for private-use-only tags.
func TestParsePrivateUseOnly(t *testing.T) {
	p := newTestParser(nil)
	testCases := []struct {
		name       string
		subtags    []string
		wantErr    error
		expectedPU []string
	}{
		{"Valid", []string{"x", "my", "tag"}, nil, []string{"my", "tag"}},
		{"Empty private use", []string{"x"}, ErrEmptyPrivateUse, nil},
		{"Subtag too long", []string{"x", "toolongtag"}, ErrSubtagTooLong, nil},
		{"Empty subtag", []string{"x", "a", "", "b"}, ErrEmptySubtag, nil},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("", false)
			err := cpr.parsePrivateUseOnly(tc.subtags)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("parsePrivateUseOnly() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil {
				if !reflect.DeepEqual(cpr.privateuse, tc.expectedPU) {
					t.Errorf("parsePrivateUseOnly() privateuse = %v, want %v", cpr.privateuse, tc.expectedPU)
				}
				if cpr.state != stateInPrivateUse {
					t.Errorf("Expected state to be stateInPrivateUse")
				}
			}
		})
	}
}

// TestCheckFinalState tests the validation checks that run after parsing.
func TestCheckFinalState(t *testing.T) {
	p := newTestParser(nil)
	testCases := []struct {
		name              string
		hasTrailingHyphen bool
		setup             func(*canonicalParseRun)
		wantErr           error
	}{
		{"OK, no trailing hyphen", false, nil, nil},
		{"OK, trailing hyphen", true, nil, nil},
		{"Empty extension with hyphen", true, func(cpr *canonicalParseRun) {
			cpr.extensionExpected = true
		}, ErrEmptyExtension},
		{"Empty private use with hyphen", true, func(cpr *canonicalParseRun) {
			cpr.state = stateInPrivateUse
			cpr.privateuse = []string{}
		}, ErrEmptyPrivateUse},
		{"Pending extension, no hyphen", false, func(cpr *canonicalParseRun) {
			cpr.extensionExpected = true
		}, ErrEmptyExtension},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("", false)
			if tc.setup != nil {
				tc.setup(cpr)
			}
			err := cpr.checkFinalState(tc.hasTrailingHyphen)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("checkFinalState() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestCheckForTooManyExtlangs tests the validation for multiple extlangs.
func TestCheckForTooManyExtlangs(t *testing.T) {
	p := newTestParser(map[string]Record{
		"extlang:gan": {Type: "extlang", Subtag: "gan", Prefix: []string{"zh"}},
		"extlang:yue": {Type: "extlang", Subtag: "yue", Prefix: []string{"zh"}},
	})
	testCases := []struct {
		name          string
		subtag        string
		extlangsCount int
		checkValidity bool
		wantErr       error
	}{
		{"OK, none so far", "gan", 0, true, nil},
		{"OK, one so far, not an extlang format", "Latn", 1, true, nil},
		{"Error, one so far, validating", "yue", 1, true, ErrTooManyExtlangs},
		{"Error, one so far, not validating", "abc", 1, false, ErrTooManyExtlangs},
		{"OK, one so far, not in registry", "zzz", 1, true, nil},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("", tc.checkValidity)
			cpr.extlangsCount = tc.extlangsCount
			err := cpr.checkForTooManyExtlangs(tc.subtag)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("checkForTooManyExtlangs() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

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

	// Expected positions:
	// zh                 (len=2) -> languageEnd = 2
	// -hak               (len=4) -> extlangEnd = 2 + 4 = 6
	// -Hans              (len=5) -> scriptEnd = 6 + 5 = 11
	// -CN                (len=3) -> regionEnd = 11 + 3 = 14
	// -variant           (len=8) -> variantEnd = 14 + 8 = 22
	// -u-co-phonebk      (len=13) -> extensionEnd = 22 + 13 = 35
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

	// Test case with only a language subtag.
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

	// Test case with empty extensions value.
	cprEmptyExt := &canonicalParseRun{
		language:   "en",
		extensions: []Extension{{Singleton: 'a'}},
	}
	// en (2) + -a (2) = 4
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
// formatted string according to RFC 5646, Section 2.1.1, which recommends specific
// casing for subtags (lowercase language, title case script, uppercase region).
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
				state:      stateInPrivateUse, // to trigger private use rendering
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
			name: "Private use only tag",
			cpr: &canonicalParseRun{
				privateuse: []string{"Whatever"},
			},
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

// TestCanonicalizeExtensionOrder verifies that extensions are sorted by singleton
// as required by RFC 5646, Section 4.5, for canonicalization.
func TestCanonicalizeExtensionOrder(t *testing.T) {
	cpr := &canonicalParseRun{
		extensions: []Extension{
			{Singleton: 'z', Value: "last"},
			{Singleton: 'a', Value: "first"},
			{Singleton: 'm', Value: "middle"},
		},
	}
	cpr.canonicalizeExtensionOrder()

	expectedOrder := []rune{'a', 'm', 'z'}
	for i, ext := range cpr.extensions {
		if ext.Singleton != expectedOrder[i] {
			t.Errorf("Expected singleton at index %d to be '%c', but got '%c'", i, expectedOrder[i], ext.Singleton)
		}
	}

	// Test with a single extension (should do nothing)
	cprSingle := &canonicalParseRun{extensions: []Extension{{Singleton: 'a', Value: "one"}}}
	cprSingle.canonicalizeExtensionOrder()
	if len(cprSingle.extensions) != 1 || cprSingle.extensions[0].Singleton != 'a' {
		t.Error("canonicalizeExtensionOrder modified a single-element slice")
	}
}

// TestCanonicalizeScriptSuppression ensures redundant script subtags are removed,
// as specified in RFC 5646, Section 3.1.9. This process simplifies tags by
// omitting the script when it is the default for a language (e.g., 'Latn' for English).
func TestCanonicalizeScriptSuppression(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:en": {Type: "language", Subtag: "en", SuppressScript: "Latn"},
		"language:zh": {Type: "language", Subtag: "zh", SuppressScript: "Hans"},
		"language:sr": {Type: "language", Subtag: "sr"},
	})
	cpr := p.newCanonicalParseRun("en-Latn", true)
	cpr.language = "en"
	cpr.script = "Latn"

	cpr.canonicalizeScriptSuppression()
	if cpr.script != "" {
		t.Errorf("Expected script 'Latn' to be suppressed for language 'en', but it was not.")
	}

	// Test non-matching script (should not be suppressed)
	cpr.language = "zh"
	cpr.script = "Hant" // Suppress-Script for 'zh' is 'Hans'
	cpr.canonicalizeScriptSuppression()
	if cpr.script != "Hant" {
		t.Errorf("Expected script 'Hant' not to be suppressed, but it was.")
	}

	// Test language without Suppress-Script
	cpr.language = "sr"
	cpr.script = "Latn"
	cpr.canonicalizeScriptSuppression()
	if cpr.script != "Latn" {
		t.Errorf("Script was suppressed for a language with no Suppress-Script field.")
	}

	// Test with no script (should do nothing)
	cpr.script = ""
	cpr.canonicalizeScriptSuppression()
	if cpr.script != "" {
		t.Error("Function modified an already empty script field.")
	}
}

// TestCompareVariants checks the logic for sorting variants based on their 'Prefix' fields,
// as described in RFC 5646, Section 4.1, point 6. This is a helper for ordering variants
// during canonicalization.
func TestCompareVariants(t *testing.T) {
	p := newTestParser(map[string]Record{
		"variant:a":        {Type: "variant", Subtag: "a", Prefix: []string{"de-b"}},
		"variant:b":        {Type: "variant", Subtag: "b", Prefix: []string{"de"}},
		"variant:c":        {Type: "variant", Subtag: "c", Prefix: []string{"de"}},
		"variant:d":        {Type: "variant", Subtag: "d"},
		"variant:e":        {Type: "variant", Subtag: "e", Prefix: []string{"fr"}},
		"variant:biske":    {Type: "variant", Subtag: "biske", Prefix: []string{"sl-rozaj"}},
		"variant:rozaj":    {Type: "variant", Subtag: "rozaj", Prefix: []string{"sl"}},
		"variant:nedis":    {Type: "variant", Subtag: "nedis", Prefix: []string{"sl"}},
		"variant:scotland": {Type: "variant", Subtag: "scotland", Prefix: []string{"en"}},
		"variant:fonipa":   {Type: "variant", Subtag: "fonipa"},
	})
	cpr := p.newCanonicalParseRun("", false)

	testCases := []struct {
		v1, v2   string
		expected bool
	}{
		// RFC example: 'rozaj' ("sl") vs 'biske' ("sl-rozaj") -> 'rozaj' comes first
		{"rozaj", "biske", true},
		{"biske", "rozaj", false},
		// RFC example: 'scotland' ("en") vs 'fonipa' (no prefix) -> 'scotland' comes first
		{"scotland", "fonipa", true},
		{"fonipa", "scotland", false},
		{"b", "a", true},  // 'b' must come before 'a' because 'a' depends on 'b'
		{"a", "b", false}, // 'a' must come after 'b'
		{"b", "d", true},  // 'b' has a prefix, 'd' does not
		{"d", "b", false}, // 'd' has no prefix, 'b' does
		{"b", "c", true},  // Same prefix status, fallback to alphabetical
		{"c", "b", false}, // Same prefix status, fallback to alphabetical
	}

	for _, tc := range testCases {
		t.Run(tc.v1+"_vs_"+tc.v2, func(t *testing.T) {
			if got := cpr.compareVariants(tc.v1, tc.v2); got != tc.expected {
				t.Errorf("compareVariants(%s, %s) = %v, want %v", tc.v1, tc.v2, got, tc.expected)
			}
		})
	}
}

// TestCanonicalizeVariantOrder checks that variants are reordered correctly
// according to the logic in compareVariants, fulfilling the canonicalization
// requirement from RFC 5646, Section 4.5.
func TestCanonicalizeVariantOrder(t *testing.T) {
	p := newTestParser(map[string]Record{
		"variant:1994":  {Type: "variant", Subtag: "1994", Prefix: []string{"sl-rozaj-biske"}},
		"variant:biske": {Type: "variant", Subtag: "biske", Prefix: []string{"sl-rozaj"}},
		"variant:rozaj": {Type: "variant", Subtag: "rozaj", Prefix: []string{"sl"}},
	})
	cpr := p.newCanonicalParseRun("", false)

	// Test case from RFC 5646, Appendix A: sl-rozaj-biske-1994
	cpr.variants = []string{"1994", "rozaj", "biske"}
	cpr.canonicalizeVariantOrder()
	expected := []string{"rozaj", "biske", "1994"}
	if !reflect.DeepEqual(cpr.variants, expected) {
		t.Errorf("canonicalizeVariantOrder() failed.\nGot:      %v\nExpected: %v", cpr.variants, expected)
	}

	// Test with one variant (should not change)
	cpr.variants = []string{"rozaj"}
	cpr.canonicalizeVariantOrder()
	if len(cpr.variants) != 1 || cpr.variants[0] != "rozaj" {
		t.Errorf("canonicalizeVariantOrder() modified a single-element slice.")
	}
}

// TestCanonicalizeDeprecated verifies that deprecated subtags are replaced by their
// 'Preferred-Value', a key step in canonicalization from RFC 5646, Section 4.5.
func TestCanonicalizeDeprecated(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:iw":    {Type: "language", Subtag: "iw", PreferredValue: "he"},
		"region:zr":      {Type: "region", Subtag: "zr", PreferredValue: "cd"},
		"variant:badvar": {Type: "variant", Subtag: "badvar", PreferredValue: "goodvar"},
	})
	cpr := p.newCanonicalParseRun("", true)
	cpr.language = "iw"
	cpr.script = "Latn" // No preferred value
	cpr.region = "zr"
	cpr.variants = []string{"badvar"}

	cpr.canonicalizeDeprecated()

	if cpr.language != "he" {
		t.Errorf("Expected deprecated language 'iw' to be replaced by 'he', got '%s'", cpr.language)
	}
	if cpr.region != "cd" {
		t.Errorf("Expected deprecated region 'zr' to be replaced by 'cd', got '%s'", cpr.region)
	}
	if len(cpr.variants) != 1 || cpr.variants[0] != "goodvar" {
		t.Errorf("Expected deprecated variant 'badvar' to be replaced by 'goodvar', got '%v'", cpr.variants)
	}
	if cpr.script != "Latn" {
		t.Errorf("Script without preferred value was changed.")
	}

	// Test empty subtag
	cpr.region = ""
	cpr.canonicalizeDeprecated()
	if cpr.region != "" {
		t.Error("Empty region was changed during deprecation check.")
	}
}

// TestCanonicalizeExtlangToPrimary checks the canonicalization rule from RFC 5646,
// Section 4.5, where a language-extlang combination is replaced by the extlang's
// primary language subtag equivalent. For example, "zh-cmn" becomes "cmn".
func TestCanonicalizeExtlangToPrimary(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:zh": {Type: "language", Subtag: "zh"},
		"extlang:cmn": {Type: "extlang", Subtag: "cmn", Prefix: []string{"zh"}, PreferredValue: "cmn"},
		"extlang:hak": {Type: "extlang", Subtag: "hak", Prefix: []string{"zh"}, PreferredValue: "hak"},
		"extlang:gan": {Type: "extlang", Subtag: "gan", Prefix: []string{"zh"}},
		"extlang:aao": {Type: "extlang", Subtag: "aao", Prefix: []string{"ar"}},
	})
	cpr := p.newCanonicalParseRun("zh-cmn", true)

	// Test case: zh-cmn -> cmn
	cpr.language = "zh"
	cpr.extlangs = []string{"cmn"}
	cpr.canonicalizeExtlangToPrimary()
	if cpr.language != "cmn" || len(cpr.extlangs) != 0 {
		t.Errorf("Expected 'zh-cmn' to canonicalize to 'cmn', got lang='%s', extlangs=%v", cpr.language, cpr.extlangs)
	}

	// Test case: no extlangs (should not change)
	cpr.language = "zh"
	cpr.extlangs = []string{}
	cpr.canonicalizeExtlangToPrimary()
	if cpr.language != "zh" || len(cpr.extlangs) != 0 {
		t.Errorf("Function incorrectly modified tag with no extlangs.")
	}

	// Test case: extlang without Preferred-Value (should not change)
	cpr.language = "zh"
	cpr.extlangs = []string{"gan"}
	cpr.canonicalizeExtlangToPrimary()
	if cpr.language != "zh" || !reflect.DeepEqual(cpr.extlangs, []string{"gan"}) {
		t.Errorf("Function incorrectly modified tag where extlang has no Preferred-Value.")
	}

	// Test case: mismatched prefix (should not change)
	cpr.language = "zh"
	cpr.extlangs = []string{"aao"} // aao's prefix is 'ar'
	cpr.canonicalizeExtlangToPrimary()
	if cpr.language != "zh" || !reflect.DeepEqual(cpr.extlangs, []string{"aao"}) {
		t.Errorf("Function incorrectly modified tag with mismatched prefix.")
	}

	// Test case: invalid extlang (not in registry)
	cpr.language = "zh"
	cpr.extlangs = []string{"zzz"} // 'zzz' is not a registered extlang in the test parser.
	cpr.canonicalizeExtlangToPrimary()

	if cpr.language != "zh" || !reflect.DeepEqual(cpr.extlangs, []string{"zzz"}) {
		t.Errorf(
			"Function incorrectly modified tag with an invalid extlang subtag. got lang='%s', extlangs=%v",
			cpr.language,
			cpr.extlangs,
		)
	}
}

// TestCanonicalize serves as an integration test for the entire canonicalization
// process, ensuring all individual canonicalization functions are called in order.
func TestCanonicalize(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:en":   {Type: "language", Subtag: "en", SuppressScript: "Latn"},
		"language:zh":   {Type: "language", Subtag: "zh"},
		"extlang:cmn":   {Type: "extlang", Subtag: "cmn", Prefix: []string{"zh"}, PreferredValue: "cmn"},
		"script:latn":   {Type: "script", Subtag: "Latn"},
		"region:bu":     {Type: "region", Subtag: "bu", PreferredValue: "mm"},
		"variant:biske": {Type: "variant", Subtag: "biske", Prefix: []string{"sl-rozaj"}},
		"variant:rozaj": {Type: "variant", Subtag: "rozaj", Prefix: []string{"sl"}},
		"i-klingon":     {Type: "grandfathered", Tag: "i-klingon", PreferredValue: "tlh"},
	})

	// This tag combines multiple non-canonical features:
	// - extlang that should be promoted to primary
	// - script that should be suppressed (but isn't for 'cmn')
	// - deprecated region
	// - out-of-order variants
	// - out-of-order extensions
	cpr := p.newCanonicalParseRun("zh-cmn-Latn-BU-biske-rozaj-b-ext2-a-ext1", true)
	err := cpr.parse()
	if err != nil {
		t.Fatalf("Initial parse failed: %v", err)
	}
	cpr.canonicalize()

	if cpr.language != "cmn" {
		t.Errorf("Expected language 'cmn', got '%s'", cpr.language)
	}
	if len(cpr.extlangs) != 0 {
		t.Errorf("Expected extlangs to be empty, got %v", cpr.extlangs)
	}
	// Note: SuppressScript check for 'en' doesn't apply as language becomes 'cmn'.
	// No SuppressScript is defined for 'cmn' in the test registry.
	if cpr.script != "Latn" {
		t.Errorf("Expected script 'Latn', got '%s'", cpr.script)
	}
	if cpr.region != "mm" {
		t.Errorf("Expected region 'mm', got '%s'", cpr.region)
	}
	expectedVariants := []string{"rozaj", "biske"}
	if !reflect.DeepEqual(cpr.variants, expectedVariants) {
		t.Errorf("Expected variants %v, got %v", expectedVariants, cpr.variants)
	}
	expectedExtensions := []Extension{{Singleton: 'a', Value: "ext1"}, {Singleton: 'b', Value: "ext2"}}
	if !reflect.DeepEqual(cpr.extensions, expectedExtensions) {
		t.Errorf("Expected extensions %v, got %v", expectedExtensions, cpr.extensions)
	}
}

// TestHandleSingleton checks the parsing of single-character subtags, which,
// according to RFC 5646, Section 2.2.6 & 2.2.7, introduce extension or
// private-use sequences.
func TestHandleSingleton(t *testing.T) {
	p := newTestParser(nil)
	testCases := []struct {
		name          string
		subtag        string
		initialState  func(*canonicalParseRun)
		checkValidity bool
		expectedErr   error
		finalState    func(*testing.T, *canonicalParseRun)
	}{
		{
			name:        "Private use singleton 'x'",
			subtag:      "x",
			expectedErr: nil,
			finalState: func(t *testing.T, cpr *canonicalParseRun) {
				if cpr.state != stateInPrivateUse {
					t.Errorf("Expected stateInPrivateUse, got %v", cpr.state)
				}
			},
		},
		{
			name:        "Extension singleton 'a'",
			subtag:      "a",
			expectedErr: nil,
			finalState: func(t *testing.T, cpr *canonicalParseRun) {
				if cpr.state != stateInExtension {
					t.Errorf("Expected stateInExtension, got %v", cpr.state)
				}
				if !cpr.extensionExpected {
					t.Error("Expected extensionExpected to be true")
				}
				if len(cpr.extensions) != 1 || cpr.extensions[0].Singleton != 'a' {
					t.Errorf("Expected extension 'a' to be added, got %v", cpr.extensions)
				}
			},
		},
		{
			name:   "Error on empty extension",
			subtag: "b",
			initialState: func(cpr *canonicalParseRun) {
				cpr.extensionExpected = true
			},
			expectedErr: ErrEmptyExtension,
		},
		{
			name:          "Error on duplicate singleton",
			subtag:        "a",
			checkValidity: true,
			initialState: func(cpr *canonicalParseRun) {
				cpr.seenSingletons = map[rune]struct{}{'a': {}}
			},
			expectedErr: ErrDuplicateSingleton,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("", tc.checkValidity)
			if tc.initialState != nil {
				tc.initialState(cpr)
			}
			err := cpr.handleSingleton(tc.subtag)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("handleSingleton() error = %v, wantErr %v", err, tc.expectedErr)
			}
			if tc.finalState != nil {
				tc.finalState(t, cpr)
			}
		})
	}
}

// TestHandleExtensionSubtag tests the parsing of subtags within an extension
// sequence, as defined by the 'extension' ABNF production in RFC 5646, Section 2.1.
func TestHandleExtensionSubtag(t *testing.T) {
	p := newTestParser(nil)
	cpr := p.newCanonicalParseRun("", false)
	cpr.extensions = append(cpr.extensions, Extension{Singleton: 'u'})

	// First subtag after singleton
	err := cpr.handleExtensionSubtag("val1")
	if err != nil {
		t.Fatalf("handleExtensionSubtag failed: %v", err)
	}
	if cpr.extensions[0].Value != "val1" {
		t.Errorf("Expected extension value 'val1', got '%s'", cpr.extensions[0].Value)
	}
	if cpr.extensionExpected {
		t.Error("extensionExpected should be false after processing a subtag")
	}

	// Second subtag, should be appended
	err = cpr.handleExtensionSubtag("val2")
	if err != nil {
		t.Fatalf("handleExtensionSubtag failed: %v", err)
	}
	if cpr.extensions[0].Value != "val1-val2" {
		t.Errorf("Expected extension value 'val1-val2', got '%s'", cpr.extensions[0].Value)
	}

	// Error case: call without a pending extension
	cprError := p.newCanonicalParseRun("", false)
	err = cprError.handleExtensionSubtag("fail")
	if !errors.Is(err, ErrInvalidSubtag) {
		t.Errorf("Expected ErrInvalidSubtag, got %v", err)
	}
}

// TestTryParseAsVariant verifies the parsing of variant subtags, which must match
// the length and content constraints from RFC 5646, Section 2.2.5.
func TestTryParseAsVariant(t *testing.T) {
	p := newTestParser(map[string]Record{
		"variant:boche":    {Type: "variant", Subtag: "boche"},
		"variant:1694":     {Type: "variant", Subtag: "1694"},
		"variant:scotland": {Type: "variant", Subtag: "scotland"},
	})

	testCases := []struct {
		name          string
		subtag        string
		initialState  parseState
		checkValidity bool
		expectParse   bool
		expectedErr   error
	}{
		{"Valid alpha variant", "scotland", stateAfterRegion, false, true, nil},
		{"Valid digit variant", "1694", stateInVariant, false, true, nil},
		{"Permissive too short alpha", "scot", stateAfterRegion, false, true, nil},
		{"Permissive too short digit", "169", stateAfterRegion, false, true, nil},
		{"Valid after script", "scotland", stateAfterScript, false, true, nil},
		{"Valid format but not in registry", "invalid", stateAfterRegion, true, false, nil},
		{"Valid and in registry", "boche", stateAfterRegion, true, true, nil},
		{"Duplicate variant error", "boche", stateInVariant, true, false, ErrDuplicateVariant},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("", tc.checkValidity)
			cpr.state = tc.initialState
			if tc.name == "Duplicate variant error" {
				cpr.variants = []string{"boche"}
				cpr.seenVariants = map[string]struct{}{"boche": {}}
			}
			parsed, err := cpr.tryParseAsVariant(tc.subtag)
			if parsed != tc.expectParse {
				t.Errorf("tryParseAsVariant() parsed = %v, want %v", parsed, tc.expectParse)
			}
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("tryParseAsVariant() error = %v, wantErr %v", err, tc.expectedErr)
			}
			if parsed && len(cpr.variants) == 0 {
				t.Error("Expected variant to be added to cpr.variants, but it was not.")
			}
		})
	}
}

// TestTryParseAsRegion checks parsing of region subtags against the ABNF from
// RFC 5646, Section 2.1 (2-letter or 3-digit).
func TestTryParseAsRegion(t *testing.T) {
	p := newTestParser(map[string]Record{
		"region:us":  {Type: "region", Subtag: "us"},
		"region:419": {Type: "region", Subtag: "419"},
	})

	testCases := []struct {
		name          string
		subtag        string
		initialState  parseState
		checkValidity bool
		expectParse   bool
	}{
		{"Valid alpha region", "us", stateAfterScript, false, true},
		{"Valid numeric region", "419", stateAfterScript, false, true},
		{"Invalid format alpha", "usa", stateAfterScript, false, false},
		{"Invalid format numeric", "41", stateAfterScript, false, false},
		{"Invalid state", "us", stateAfterRegion, false, false},
		{"Valid but not in registry", "zz", stateAfterScript, true, false},
		{"Valid and in registry", "us", stateAfterScript, true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("", tc.checkValidity)
			cpr.state = tc.initialState
			parsed := cpr.tryParseAsRegion(tc.subtag)
			if parsed != tc.expectParse {
				t.Errorf("tryParseAsRegion() parsed = %v, want %v", parsed, tc.expectParse)
			}
			if parsed && cpr.region == "" {
				t.Error("Expected region to be set in cpr.region, but it was not.")
			}
		})
	}
}

// TestTryParseAsScript checks parsing of script subtags against the ABNF from
// RFC 5646, Section 2.1 (4-letter).
func TestTryParseAsScript(t *testing.T) {
	p := newTestParser(map[string]Record{
		"script:latn": {Type: "script", Subtag: "Latn"},
		"script:cyrl": {Type: "script", Subtag: "Cyrl"},
	})

	testCases := []struct {
		name          string
		subtag        string
		initialState  parseState
		checkValidity bool
		expectParse   bool
	}{
		{"Valid script", "latn", stateAfterExtLang, false, true},
		{"Invalid format", "lat", stateAfterExtLang, false, false},
		{"Invalid state", "latn", stateAfterScript, false, false},
		{"Valid but not in registry", "zyyy", stateAfterLanguage, true, false},
		{"Valid and in registry", "cyrl", stateAfterLanguage, true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("", tc.checkValidity)
			cpr.state = tc.initialState
			parsed := cpr.tryParseAsScript(tc.subtag)
			if parsed != tc.expectParse {
				t.Errorf("tryParseAsScript() parsed = %v, want %v", parsed, tc.expectParse)
			}
			if parsed && cpr.script == "" {
				t.Error("Expected script to be set in cpr.script, but it was not.")
			}
		})
	}
}

// TestTryParseAsExtlang checks parsing of extended language subtags against the ABNF
// from RFC 5646, Section 2.1 (3-letter). It also tests the one-extlang limit
// specified in Section 2.2.2.
func TestTryParseAsExtlang(t *testing.T) {
	p := newTestParser(map[string]Record{
		"extlang:gan": {Type: "extlang", Subtag: "gan", Prefix: []string{"zh"}},
		"extlang:yue": {Type: "extlang", Subtag: "yue", Prefix: []string{"zh"}},
	})

	testCases := []struct {
		name          string
		subtag        string
		initialState  parseState
		initialCount  int
		checkValidity bool
		expectParse   bool
	}{
		{"Valid extlang", "gan", stateAfterLanguage, 0, false, true},
		{"Invalid format", "ga", stateAfterLanguage, 0, false, false},
		{"Invalid state", "gan", stateAfterExtLang, 0, false, false},
		{"Too many extlangs", "yue", stateAfterLanguage, 1, false, false},
		{"Valid but not in registry", "zzz", stateAfterLanguage, 0, true, false},
		{"Valid and in registry", "yue", stateAfterLanguage, 0, true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("", tc.checkValidity)
			cpr.state = tc.initialState
			cpr.extlangsCount = tc.initialCount
			parsed := cpr.tryParseAsExtlang(tc.subtag)
			if parsed != tc.expectParse {
				t.Errorf("tryParseAsExtlang() parsed = %v, want %v", parsed, tc.expectParse)
			}
			if parsed && len(cpr.extlangs) == 0 {
				t.Error("Expected extlang to be added to cpr.extlangs, but it was not.")
			}
		})
	}
}

// TestHandleLangtagSubtag verifies the dispatching logic that identifies and
// processes subtags based on their position, length, and content, following the
// state machine logic derived from the ABNF in RFC 5646, Section 2.1.
func TestHandleLangtagSubtag(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:en": {Type: "language", Subtag: "en"},
		"script:latn": {Type: "script", Subtag: "Latn"},
	})

	// Test case for i=0, should call handlePrimaryLanguage
	cpr := p.newCanonicalParseRun("en-US", false)
	err := cpr.handleLangtagSubtag(0, "en")
	if err != nil || cpr.language != "en" {
		t.Errorf("handleLangtagSubtag failed for primary language: err=%v, lang=%s", err, cpr.language)
	}

	// Test case for singleton, should call handleSingleton
	cpr = p.newCanonicalParseRun("en-a-foo", false)
	cpr.language = "en"
	cpr.state = stateInVariant
	err = cpr.handleLangtagSubtag(1, "a")
	if err != nil || cpr.state != stateInExtension {
		t.Errorf("handleLangtagSubtag failed for singleton: err=%v, state=%v", err, cpr.state)
	}

	// Test case for dispatching to a tryParse function
	cpr = p.newCanonicalParseRun("en-Latn", false)
	cpr.language = "en"
	cpr.state = stateAfterLanguage
	err = cpr.handleLangtagSubtag(1, "Latn")
	if err != nil || cpr.script != "Latn" || cpr.state != stateAfterScript {
		t.Errorf("handleLangtagSubtag failed to parse script: err=%v, script=%s, state=%v", err, cpr.script, cpr.state)
	}

	// Test case for invalid subtag that doesn't match any type
	cpr = p.newCanonicalParseRun("en-123", true) // Use validity to ensure it's not a region
	cpr.language = "en"
	cpr.state = stateAfterLanguage
	err = cpr.handleLangtagSubtag(1, "123")
	if !errors.Is(err, ErrInvalidSubtag) {
		t.Errorf("Expected ErrInvalidSubtag for unparseable subtag, got %v", err)
	}
}

// TestHandlePrimaryLanguage validates the parsing of the first subtag in a tag,
// which must be a language subtag conforming to RFC 5646, Section 2.2.1.
func TestHandlePrimaryLanguage(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:en":       {Type: "language", Subtag: "en"},
		"language:deu":      {Type: "language", Subtag: "deu"},
		"language:enochian": {Type: "language", Subtag: "enochian"},
		"i-ami":             {Type: "grandfathered", Tag: "i-ami"},
	})
	testCases := []struct {
		name          string
		subtag        string
		checkValidity bool
		expectedLang  string
		expectedState parseState
		expectedErr   error
	}{
		{"Valid 2-letter", "en", false, "en", stateAfterLanguage, nil},
		{"Valid 3-letter", "deu", false, "deu", stateAfterLanguage, nil},
		{"Valid 5-8 letter", "enochian", false, "enochian", stateAfterExtLang, nil},
		{"Permissive 1-letter", "e", false, "e", stateAfterLanguage, nil},
		{"Invalid too long", "longlanguage", false, "", 0, ErrInvalidLanguage},
		{"Invalid non-alpha", "en1", false, "", 0, ErrInvalidLanguage},
		{"Valid but not in registry", "zz", true, "", 0, ErrInvalidLanguage},
		{"Valid and in registry", "en", true, "en", stateAfterLanguage, nil},
		{"Valid grandfathered 'i' subtag", "i", false, "i", stateAfterLanguage, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("", tc.checkValidity)
			err := cpr.handlePrimaryLanguage(tc.subtag)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("handlePrimaryLanguage() error = %v, wantErr %v", err, tc.expectedErr)
			}
			if err == nil {
				if cpr.language != tc.expectedLang {
					t.Errorf("Expected language '%s', got '%s'", tc.expectedLang, cpr.language)
				}
				if cpr.state != tc.expectedState {
					t.Errorf("Expected state %v, got %v", tc.expectedState, cpr.state)
				}
			}
		})
	}
}

// TestParse is an integration test for the state machine, ensuring it correctly
// parses well-formed tags and rejects ill-formed ones, as defined by the ABNF
// in RFC 5646, Section 2.1.
func TestParse(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:de":  {Type: "language", Subtag: "de"},
		"language:en":  {Type: "language", Subtag: "en"},
		"language:zh":  {Type: "language", Subtag: "zh"},
		"extlang:gan":  {Type: "extlang", Subtag: "gan", Prefix: []string{"zh"}},
		"extlang:yue":  {Type: "extlang", Subtag: "yue", Prefix: []string{"zh"}},
		"script:latn":  {Type: "script", Subtag: "Latn"},
		"region:us":    {Type: "region", Subtag: "us"},
		"region:de":    {Type: "region", Subtag: "DE"},
		"variant:1901": {Type: "variant", Subtag: "1901"},
	})
	testCases := []struct {
		name        string
		tag         string
		expectErr   error
		validity    bool
		finalChecks func(*testing.T, *canonicalParseRun)
	}{
		{"Valid full tag", "en-Latn-US", nil, true, nil},
		{"Private use only", "x-my-own-tag", nil, false, func(t *testing.T, cpr *canonicalParseRun) {
			if !reflect.DeepEqual(cpr.privateuse, []string{"my", "own", "tag"}) {
				t.Error("Private use was not parsed correctly")
			}
		}},
		{"Final subtag is extension", "en-a-foo", nil, false, func(t *testing.T, cpr *canonicalParseRun) {
			if cpr.extensionExpected {
				t.Error("extensionExpected should be false at end of parse")
			}
		}},
		{"ErrEmptySubtag", "en--US", ErrEmptySubtag, false, nil},
		{"ErrEmptySubtag in private use", "x-a--b", ErrEmptySubtag, false, nil},
		{"ErrSubtagTooLong", "en-abcdefghi", ErrSubtagTooLong, false, nil},
		{"ErrSubtagTooLong in private use", "x-abcdefghi", ErrSubtagTooLong, false, nil},
		{"ErrEmptyPrivateUse", "x", ErrEmptyPrivateUse, false, nil},
		{"ErrEmptyExtension at end", "en-a", ErrEmptyExtension, false, nil},
		{"ErrEmptyPrivateUse with trailing hyphen", "en-x-", ErrEmptyPrivateUse, false, nil},
		{"ErrTooManyExtlangs", "zh-gan-yue", ErrTooManyExtlangs, true, nil},
		{"ErrDuplicateVariant", "de-1901-1901", ErrDuplicateVariant, true, nil},
		{"ErrDuplicateSingleton", "en-a-foo-a-bar", ErrDuplicateSingleton, true, nil},
		{"ErrInvalidSubtag", "en-US-1234", ErrInvalidSubtag, true, nil},
		{"ErrTooManyExtlangs non-validating", "en-abc-def", ErrTooManyExtlangs, false, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun(tc.tag, tc.validity)
			err := cpr.parse()
			if !errors.Is(err, tc.expectErr) {
				t.Errorf("parse() error = %v, wantErr %v", err, tc.expectErr)
			}
			if tc.finalChecks != nil {
				tc.finalChecks(t, cpr)
			}
		})
	}
}

// TestNewCanonicalParseRun ensures the constructor for a parsing run correctly
// initializes the struct from an input tag string.
func TestNewCanonicalParseRun(t *testing.T) {
	p := newTestParser(nil)
	input := "en-US"
	cpr := p.newCanonicalParseRun(input, true)

	if cpr.parent != p {
		t.Error("Parent parser was not set correctly.")
	}
	expectedSubtags := []string{"en", "US"}
	if !reflect.DeepEqual(cpr.subtags, expectedSubtags) {
		t.Errorf("Subtags not split correctly. Got %v, expected %v", cpr.subtags, expectedSubtags)
	}
	if !cpr.checkValidity {
		t.Error("checkValidity flag not set correctly.")
	}
}
