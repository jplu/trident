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
				t.Errorf(
					"prepareSubtags() hasHyphen = %v, want %v",
					hasHyphen,
					tc.expectedHasHyphen,
				)
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
					t.Errorf(
						"parsePrivateUseOnly() privateuse = %v, want %v",
						cpr.privateuse,
						tc.expectedPU,
					)
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
		{"Error, trailing hyphen", true, nil, ErrEmptySubtag},
		{"Empty extension with hyphen", true, func(cpr *canonicalParseRun) {
			cpr.extensionExpected = true
		}, ErrEmptySubtag},
		{"Empty private use with hyphen", true, func(cpr *canonicalParseRun) {
			cpr.state = stateInPrivateUse
			cpr.privateuse = []string{}
		}, ErrEmptySubtag},
		{"Pending extension, no hyphen", false, func(cpr *canonicalParseRun) {
			cpr.extensionExpected = true
		}, ErrEmptyExtension},
		{"Empty private use, no hyphen", false, func(cpr *canonicalParseRun) {
			cpr.state = stateInPrivateUse
			cpr.privateuse = []string{}
		}, ErrEmptyPrivateUse},
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
		{"OK, one so far, not validating", "abc", 1, false, nil},
		{"OK, two so far, not validating", "def", 2, false, nil},
		{"Error, three so far, not validating", "ghi", 3, false, ErrTooManyExtlangs},
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
func TestTryParseAsVariant(t *testing.T) {
	p := newTestParser(map[string]Record{
		"variant:boche":    {Type: "variant", Subtag: "boche"},
		"variant:1694":     {Type: "variant", Subtag: "1694"},
		"variant:scotland": {Type: "variant", Subtag: "scotland"},
		"en-gb-oed":        {Type: "grandfathered", Tag: "en-GB-oed"},
	})

	testCases := []struct {
		name          string
		subtag        string
		initialState  parseState
		checkValidity bool
		expectParse   bool
		expectedErr   error
		setup         func(*canonicalParseRun)
	}{
		{"Valid alpha variant", "scotland", stateAfterRegion, false, true, nil, nil},
		{"Valid digit variant", "1694", stateInVariant, false, true, nil, nil},
		{"Strict too short alpha", "scot", stateAfterRegion, false, false, nil, nil},
		{"Strict too short digit", "169", stateAfterRegion, false, false, nil, nil},
		{"Valid after script", "scotland", stateAfterScript, false, true, nil, nil},
		{"Valid format but not in registry", "invalid", stateAfterRegion, true, false, nil, nil},
		{"Valid and in registry", "boche", stateAfterRegion, true, true, nil, nil},
		{"Duplicate variant error", "boche", stateInVariant, true, false, ErrDuplicateVariant, nil},
		{
			"Valid grandfathered variant",
			"oed",
			stateAfterRegion,
			false,
			true,
			nil,
			func(cpr *canonicalParseRun) {
				cpr.subtags = []string{"en", "GB", "oed"}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("en-US", tc.checkValidity)
			cpr.state = tc.initialState
			if tc.setup != nil {
				tc.setup(cpr)
			}
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
		{"OK second extlang", "yue", stateAfterLanguage, 1, false, true},
		{"OK third extlang", "yue", stateAfterLanguage, 2, false, true},
		{"Too many extlangs well-formed", "yue", stateAfterLanguage, 3, false, false},
		{"Too many extlangs validating", "yue", stateAfterLanguage, 1, true, false},
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
func TestHandleLangtagSubtag(t *testing.T) {
	p := newTestParser(map[string]Record{
		"language:en": {Type: "language", Subtag: "en"},
		"script:latn": {Type: "script", Subtag: "Latn"},
	})

	cpr := p.newCanonicalParseRun("en-US", false)
	err := cpr.handleLangtagSubtag(0, "en")
	if err != nil || cpr.language != "en" {
		t.Errorf(
			"handleLangtagSubtag failed for primary language: err=%v, lang=%s",
			err,
			cpr.language,
		)
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
		t.Errorf(
			"handleLangtagSubtag failed to parse script: err=%v, script=%s, state=%v",
			err,
			cpr.script,
			cpr.state,
		)
	}

	cpr = p.newCanonicalParseRun("zh-yue-gan", false)
	cpr.language = "zh"
	cpr.state = stateAfterLanguage

	err = cpr.handleLangtagSubtag(1, "yue")
	if err != nil || len(cpr.extlangs) != 1 || cpr.extlangs[0] != "yue" {
		t.Errorf(
			"Expected first extlang to parse successfully: err=%v, extlangs=%v",
			err,
			cpr.extlangs,
		)
	}
	if cpr.state != stateAfterLanguage {
		t.Errorf(
			"Expected state to remain stateAfterLanguage to allow more extlangs, got %v",
			cpr.state,
		)
	}

	err = cpr.handleLangtagSubtag(2, "gan")
	if err != nil || len(cpr.extlangs) != 2 || cpr.extlangs[1] != "gan" {
		t.Errorf(
			"Expected second extlang to parse successfully: err=%v, extlangs=%v",
			err,
			cpr.extlangs,
		)
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
		setup         func(*canonicalParseRun)
	}{
		{"Valid 2-letter", "en", false, "en", stateAfterLanguage, nil, nil},
		{"Valid 3-letter", "deu", false, "deu", stateAfterLanguage, nil, nil},
		{"Valid 5-8 letter", "enochian", false, "enochian", stateAfterExtLang, nil, nil},
		{"Strict 1-letter error", "e", false, "", 0, ErrInvalidLanguage, nil},
		{"Invalid too long", "longlanguage", false, "", 0, ErrInvalidLanguage, nil},
		{"Invalid non-alpha", "en1", false, "", 0, ErrInvalidLanguage, nil},
		{"Valid but not in registry", "zz", true, "", 0, ErrInvalidLanguage, nil},
		{"Valid and in registry", "en", true, "en", stateAfterLanguage, nil, nil},
		{
			"Valid grandfathered 1-letter 'i'",
			"i",
			false,
			"i",
			stateAfterLanguage,
			nil,
			func(cpr *canonicalParseRun) {
				cpr.subtags = []string{"i", "ami"}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := p.newCanonicalParseRun("en-US", tc.checkValidity)
			if tc.setup != nil {
				tc.setup(cpr)
			}
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
		{
			"Private use only",
			"x-my-own-tag",
			nil,
			false,
			func(t *testing.T, cpr *canonicalParseRun) {
				if !reflect.DeepEqual(cpr.privateuse, []string{"my", "own", "tag"}) {
					t.Error("Private use was not parsed correctly")
				}
			},
		},
		{
			"Final subtag is extension",
			"en-a-foo",
			nil,
			false,
			func(t *testing.T, cpr *canonicalParseRun) {
				if cpr.extensionExpected {
					t.Error("extensionExpected should be false at end of parse")
				}
			},
		},
		{
			"Tag with private use subtags (well-formed)",
			"en-x-priv1-priv2",
			nil,
			false,
			func(t *testing.T, cpr *canonicalParseRun) {
				if !reflect.DeepEqual(cpr.privateuse, []string{"priv1", "priv2"}) {
					t.Errorf("Private use was not parsed correctly: got %v, want [priv1 priv2]", cpr.privateuse)
				}
				if cpr.state != stateInPrivateUse {
					t.Errorf("Expected stateInPrivateUse, got %v", cpr.state)
				}
			},
		},
		{
			"Tag with private use subtags (validating)",
			"en-US-x-custom",
			nil,
			true,
			func(t *testing.T, cpr *canonicalParseRun) {
				if !reflect.DeepEqual(cpr.privateuse, []string{"custom"}) {
					t.Errorf("Private use was not parsed correctly: got %v, want [custom]", cpr.privateuse)
				}
				if cpr.state != stateInPrivateUse {
					t.Errorf("Expected stateInPrivateUse, got %v", cpr.state)
				}
			},
		},
		{"ErrEmptySubtag", "en--US", ErrEmptySubtag, false, nil},
		{"ErrEmptySubtag in private use", "x-a--b", ErrEmptySubtag, false, nil},
		{"ErrSubtagTooLong", "en-abcdefghi", ErrSubtagTooLong, false, nil},
		{"ErrSubtagTooLong in private use", "x-abcdefghi", ErrSubtagTooLong, false, nil},
		{"ErrEmptyPrivateUse", "x", ErrEmptyPrivateUse, false, nil},
		{"ErrEmptyExtension at end", "en-a", ErrEmptyExtension, false, nil},
		{"ErrEmptyPrivateUse with trailing hyphen", "en-x-", ErrEmptySubtag, false, nil},
		{"ErrEmptyPrivateUse standard tag", "en-x", ErrEmptyPrivateUse, false, nil},
		{
			"ErrEmptySubtag private-use-only with trailing hyphen",
			"x-a-",
			ErrEmptySubtag,
			false,
			nil,
		},
		{"ErrTooManyExtlangs", "zh-gan-yue", ErrTooManyExtlangs, true, nil},
		{"ErrDuplicateVariant", "de-1901-1901", ErrDuplicateVariant, true, nil},
		{"ErrDuplicateSingleton", "en-a-foo-a-bar", ErrDuplicateSingleton, true, nil},
		{"ErrInvalidSubtag", "en-US-1234", ErrInvalidSubtag, true, nil},
		{"Valid non-validating multi-extlang", "en-abc-def", nil, false, nil},
		{"ErrTooManyExtlangs non-validating", "en-abc-def-ghi-jkl", ErrTooManyExtlangs, false, nil},
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

// TestIsTagGrandfathered verifies the detection of grandfathered tags with and without registry data.
func TestIsTagGrandfathered(t *testing.T) {
	pWithReg := newTestParser(map[string]Record{
		"en-gb-oed":            {Type: "grandfathered", Tag: "en-GB-oed"},
		"custom-grandfathered": {Type: "grandfathered", Tag: "custom-grandfathered"},
		"custom-redundant":     {Type: "redundant", Tag: "custom-redundant"},
	})
	pWithoutReg := newTestParser(nil)

	testCases := []struct {
		name     string
		parser   *Parser
		tag      string
		expected bool
	}{
		{"Static list match without registry", pWithoutReg, "en-GB-oed", true},
		{"Static list match with registry", pWithReg, "en-GB-oed", true},
		{"Registry grandfathered match not in static list", pWithReg, "custom-grandfathered", true},
		{"Registry redundant match not in static list", pWithReg, "custom-redundant", true},
		{"Irregular tag without registry", pWithoutReg, "i-default", true},
		{"Regular tag without registry", pWithoutReg, "zh-min-nan", true},
		{"Non-grandfathered tag with registry", pWithReg, "en-US", false},
		{"Non-grandfathered tag without registry", pWithoutReg, "en-US", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpr := tc.parser.newCanonicalParseRun(tc.tag, false)
			if got := cpr.isTagGrandfathered(); got != tc.expected {
				t.Errorf("isTagGrandfathered() = %v, want %v", got, tc.expected)
			}
		})
	}
}
