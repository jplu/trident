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
	"testing"
)

// TestCanonicalizeExtensionOrder verifies that extensions are sorted by singleton.
// RFC 5646, Section 4.5.
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

	cprSingle := &canonicalParseRun{extensions: []Extension{{Singleton: 'a', Value: "one"}}}
	cprSingle.canonicalizeExtensionOrder()
	if len(cprSingle.extensions) != 1 || cprSingle.extensions[0].Singleton != 'a' {
		t.Error("canonicalizeExtensionOrder modified a single-element slice")
	}
}

// TestCanonicalizeScriptSuppression ensures redundant script subtags are removed.
// RFC 5646, Section 3.1.9.
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

	cpr.language = "zh"
	cpr.script = "Hant"
	cpr.canonicalizeScriptSuppression()
	if cpr.script != "Hant" {
		t.Errorf("Expected script 'Hant' not to be suppressed, but it was.")
	}

	cpr.language = "sr"
	cpr.script = "Latn"
	cpr.canonicalizeScriptSuppression()
	if cpr.script != "Latn" {
		t.Errorf("Script was suppressed for a language with no Suppress-Script field.")
	}

	cpr.script = ""
	cpr.canonicalizeScriptSuppression()
	if cpr.script != "" {
		t.Error("Function modified an already empty script field.")
	}
}

// TestCompareVariants checks the logic for sorting variants based on their Prefix fields.
// RFC 5646, Section 4.1, point 6.
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
		{"rozaj", "biske", true},
		{"biske", "rozaj", false},
		{"scotland", "fonipa", true},
		{"fonipa", "scotland", false},
		{"b", "a", true},
		{"a", "b", false},
		{"b", "d", true},
		{"d", "b", false},
		{"b", "c", true},
		{"c", "b", false},
	}

	for _, tc := range testCases {
		t.Run(tc.v1+"_vs_"+tc.v2, func(t *testing.T) {
			if got := cpr.compareVariants(tc.v1, tc.v2); got != tc.expected {
				t.Errorf("compareVariants(%s, %s) = %v, want %v", tc.v1, tc.v2, got, tc.expected)
			}
		})
	}
}

// TestCanonicalizeVariantOrder checks that variants are reordered correctly.
// RFC 5646, Section 4.5.
func TestCanonicalizeVariantOrder(t *testing.T) {
	p := newTestParser(map[string]Record{
		"variant:1994":  {Type: "variant", Subtag: "1994", Prefix: []string{"sl-rozaj-biske"}},
		"variant:biske": {Type: "variant", Subtag: "biske", Prefix: []string{"sl-rozaj"}},
		"variant:rozaj": {Type: "variant", Subtag: "rozaj", Prefix: []string{"sl"}},
	})
	cpr := p.newCanonicalParseRun("", false)

	cpr.variants = []string{"1994", "rozaj", "biske"}
	cpr.canonicalizeVariantOrder()
	expected := []string{"rozaj", "biske", "1994"}
	if !reflect.DeepEqual(cpr.variants, expected) {
		t.Errorf("canonicalizeVariantOrder() failed.\nGot:      %v\nExpected: %v", cpr.variants, expected)
	}

	cpr.variants = []string{"rozaj"}
	cpr.canonicalizeVariantOrder()
	if len(cpr.variants) != 1 || cpr.variants[0] != "rozaj" {
		t.Errorf("canonicalizeVariantOrder() modified a single-element slice.")
	}
}

// TestCanonicalizeDeprecated verifies that deprecated subtags are replaced by their
// Preferred-Value. RFC 5646, Section 4.5.
func TestCanonicalizeDeprecated(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:iw":    {Type: "language", Subtag: "iw", PreferredValue: "he"},
		"region:zr":      {Type: "region", Subtag: "zr", PreferredValue: "cd"},
		"variant:badvar": {Type: "variant", Subtag: "badvar", PreferredValue: "goodvar"},
	})
	cpr := p.newCanonicalParseRun("", true)
	cpr.language = "iw"
	cpr.script = "Latn"
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

	cpr.region = ""
	cpr.canonicalizeDeprecated()
	if cpr.region != "" {
		t.Error("Empty region was changed during deprecation check.")
	}
}

// TestCanonicalizeExtlangToPrimary checks the canonicalization rule from RFC 5646,
// Section 4.5, where a language-extlang combination is replaced by the extlang's
// primary language subtag equivalent.
func TestCanonicalizeExtlangToPrimary(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:zh": {Type: "language", Subtag: "zh"},
		"extlang:cmn": {Type: "extlang", Subtag: "cmn", Prefix: []string{"zh"}, PreferredValue: "cmn"},
		"extlang:hak": {Type: "extlang", Subtag: "hak", Prefix: []string{"zh"}, PreferredValue: "hak"},
		"extlang:gan": {Type: "extlang", Subtag: "gan", Prefix: []string{"zh"}},
		"extlang:aao": {Type: "extlang", Subtag: "aao", Prefix: []string{"ar"}},
	})
	cpr := p.newCanonicalParseRun("zh-cmn", true)

	cpr.language = "zh"
	cpr.extlangs = []string{"cmn"}
	cpr.canonicalizeExtlangToPrimary()
	if cpr.language != "cmn" || len(cpr.extlangs) != 0 {
		t.Errorf("Expected 'zh-cmn' to canonicalize to 'cmn', got lang='%s', extlangs=%v", cpr.language, cpr.extlangs)
	}

	cpr.language = "zh"
	cpr.extlangs = []string{}
	cpr.canonicalizeExtlangToPrimary()
	if cpr.language != "zh" || len(cpr.extlangs) != 0 {
		t.Errorf("Function incorrectly modified tag with no extlangs.")
	}

	cpr.language = "zh"
	cpr.extlangs = []string{"gan"}
	cpr.canonicalizeExtlangToPrimary()
	if cpr.language != "zh" || !reflect.DeepEqual(cpr.extlangs, []string{"gan"}) {
		t.Errorf("Function incorrectly modified tag where extlang has no Preferred-Value.")
	}

	cpr.language = "zh"
	cpr.extlangs = []string{"aao"}
	cpr.canonicalizeExtlangToPrimary()
	if cpr.language != "zh" || !reflect.DeepEqual(cpr.extlangs, []string{"aao"}) {
		t.Errorf("Function incorrectly modified tag with mismatched prefix.")
	}

	cpr.language = "zh"
	cpr.extlangs = []string{"zzz"}
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
// process. RFC 5646, Section 4.5.
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
