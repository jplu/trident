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
	"errors"
	"reflect"
	"testing"
)

// newTestParser creates a parser with a predefined registry for testing purposes.
// This allows for isolated testing of functions that rely on registry data
// without depending on the embedded registry. It is shared across parse_run_test.go,
// canonicalize_test.go, and render_test.go.
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

// TestHandleSingleton checks the parsing of single-character subtags.
// RFC 5646, Sections 2.2.6 & 2.2.7 define extension and private-use singletons.
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

// TestHandleExtensionSubtag tests the parsing of subtags within an extension sequence.
// RFC 5646, Section 2.1, 'extension' ABNF production.
func TestHandleExtensionSubtag(t *testing.T) {
	p := newTestParser(nil)
	cpr := p.newCanonicalParseRun("", false)
	cpr.extensions = append(cpr.extensions, Extension{Singleton: 'u'})

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

	err = cpr.handleExtensionSubtag("val2")
	if err != nil {
		t.Fatalf("handleExtensionSubtag failed: %v", err)
	}
	if cpr.extensions[0].Value != "val1-val2" {
		t.Errorf("Expected extension value 'val1-val2', got '%s'", cpr.extensions[0].Value)
	}

	cprError := p.newCanonicalParseRun("", false)
	err = cprError.handleExtensionSubtag("fail")
	if !errors.Is(err, ErrInvalidSubtag) {
		t.Errorf("Expected ErrInvalidSubtag, got %v", err)
	}
}

// TestTryParseAsVariant verifies the parsing of variant subtags.
// RFC 5646, Section 2.2.5.
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

// TestTryParseAsRegion checks parsing of region subtags.
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

// TestTryParseAsScript checks parsing of script subtags.
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

// TestTryParseAsExtlang checks parsing of extended language subtags.
// RFC 5646, Section 2.1 (3-letter) and Section 2.2.2 (one-extlang limit).
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
// processes subtags based on their position, length, and content.
// RFC 5646, Section 2.1 state machine.
func TestHandleLangtagSubtag(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:en": {Type: "language", Subtag: "en"},
		"script:latn": {Type: "script", Subtag: "Latn"},
	})

	cpr := p.newCanonicalParseRun("en-US", false)
	err := cpr.handleLangtagSubtag(0, "en")
	if err != nil || cpr.language != "en" {
		t.Errorf("handleLangtagSubtag failed for primary language: err=%v, lang=%s", err, cpr.language)
	}

	cpr = p.newCanonicalParseRun("en-a-foo", false)
	cpr.language = "en"
	cpr.state = stateInVariant
	err = cpr.handleLangtagSubtag(1, "a")
	if err != nil || cpr.state != stateInExtension {
		t.Errorf("handleLangtagSubtag failed for singleton: err=%v, state=%v", err, cpr.state)
	}

	cpr = p.newCanonicalParseRun("en-Latn", false)
	cpr.language = "en"
	cpr.state = stateAfterLanguage
	err = cpr.handleLangtagSubtag(1, "Latn")
	if err != nil || cpr.script != "Latn" || cpr.state != stateAfterScript {
		t.Errorf("handleLangtagSubtag failed to parse script: err=%v, script=%s, state=%v", err, cpr.script, cpr.state)
	}

	cpr = p.newCanonicalParseRun("en-123", true)
	cpr.language = "en"
	cpr.state = stateAfterLanguage
	err = cpr.handleLangtagSubtag(1, "123")
	if !errors.Is(err, ErrInvalidSubtag) {
		t.Errorf("Expected ErrInvalidSubtag for unparseable subtag, got %v", err)
	}
}

// TestHandlePrimaryLanguage validates the parsing of the first subtag.
// RFC 5646, Section 2.2.1.
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

// TestParse is an integration test for the state machine.
// RFC 5646, Section 2.1 ABNF.
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
