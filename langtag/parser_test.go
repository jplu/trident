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
	"log/slog"
	"os"
	"strings"
	"testing"
)

//nolint:gochecknoglobals // p is a global parser instance, initialized once by TestMain to speed up tests.
var p *Parser

func TestMain(m *testing.M) {
	var err error
	p, err = NewParser()
	if err != nil {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		logger.Error("FATAL: Failed to create new parser for tests", "error", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// mustParse is a test helper that parses a tag using the non-validating
// p.Parse method and fails the test if an error occurs.
func mustParse(t *testing.T, tag string) LanguageTag {
	t.Helper()
	lt, err := p.Parse(tag)
	if err != nil {
		t.Fatalf("mustParse failed for tag '%s': %v", tag, err)
	}
	return lt
}

// mustParseAndNormalize is a test helper that parses a tag using the validating
// and normalizing p.ParseAndNormalize method and fails the test if an error occurs.
func mustParseAndNormalize(t *testing.T, tag string) LanguageTag {
	t.Helper()
	lt, err := p.ParseAndNormalize(tag)
	if err != nil {
		t.Fatalf("mustParseAndNormalize failed for tag '%s': %v", tag, err)
	}
	return lt
}

// TestParser_Parse tests the non-validating Parse method.
// RFC 5646 Section 2.2.9 defines "well-formed" as conforming to the ABNF.
func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		wantTag string
		wantErr error
	}{
		{name: "Simple tag", tag: "de", wantTag: "de"},
		{name: "Language-Region", tag: "en-US", wantTag: "en-US"},
		{name: "Language-Script-Region", tag: "sr-Latn-RS", wantTag: "sr-Latn-RS"},
		{name: "Case normalization", tag: "MN-cYRL-mn", wantTag: "mn-Cyrl-MN"},
		{name: "Private use", tag: "de-CH-x-phonebk", wantTag: "de-CH-x-phonebk"},
		{name: "Private use only", tag: "x-whatever", wantTag: "x-whatever"},
		{name: "Grandfathered irregular", tag: "i-klingon", wantTag: "i-klingon"},
		{name: "Grandfathered regular", tag: "art-lojban", wantTag: "art-lojban"},
		{name: "Extension", tag: "en-a-myext-b-another", wantTag: "en-a-myext-b-another"},
		{name: "Unregistered language", tag: "zz-US", wantTag: "zz-US"},
		{name: "Unregistered script", tag: "en-Zzzz-US", wantTag: "en-Zzzz-US"},
		{name: "Duplicate variant", tag: "de-DE-1901-1901", wantTag: "de-DE-1901-1901"},
		{name: "Duplicate singleton", tag: "en-a-foo-a-bar", wantTag: "en-a-foo-a-bar"},
		{name: "Forbidden character", tag: "en_US", wantErr: ErrForbiddenChar},
		{name: "Empty subtag", tag: "en--US", wantErr: ErrEmptySubtag},
		{name: "Subtag too long", tag: "verylongsubtag-en", wantErr: ErrSubtagTooLong},
		{name: "Empty private use", tag: "x-", wantErr: ErrEmptyPrivateUse},
		{name: "Empty extension", tag: "en-a-", wantErr: ErrEmptyExtension},
		{name: "Empty extension sequence", tag: "en-a-b-foo", wantErr: ErrEmptyExtension},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.Parse(tt.tag)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.String() != tt.wantTag {
				t.Errorf("Parse() got = %q, want %q", got.String(), tt.wantTag)
			}
		})
	}
}

// TestParser_ParseAndNormalize tests the validating and canonicalizing ParseAndNormalize method.
// RFC 5646 Section 4.5 defines canonicalization. Section 2.2.9 defines validity.
func TestParser_ParseAndNormalize(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		wantTag string
		wantErr error
	}{
		{name: "Redundant tag replacement", tag: "zh-min-nan", wantTag: "nan"},
		{name: "Grandfathered replacement (art-lojban)", tag: "art-lojban", wantTag: "jbo"},
		{name: "Grandfathered replacement (i-klingon)", tag: "i-klingon", wantTag: "tlh"},
		{name: "Grandfathered no-replacement", tag: "i-enochian", wantTag: "i-enochian"},
		{name: "Subtag replacement", tag: "en-BU", wantTag: "en-MM"},
		{name: "Extlang canonicalization", tag: "zh-gan", wantTag: "gan"},
		{name: "Extension reordering", tag: "en-b-ccc-a-aaa", wantTag: "en-a-aaa-b-ccc"},
		{name: "Script suppression", tag: "is-Latn", wantTag: "is"},
		{name: "Case canonicalization", tag: "SR-LATN-rs", wantTag: "sr-Latn-RS"},
		{name: "Invalid language subtag", tag: "zz-US", wantErr: ErrInvalidLanguage},
		{name: "Invalid region subtag", tag: "en-BOGUS", wantErr: ErrInvalidSubtag},
		{name: "Two region tags", tag: "de-419-DE", wantErr: ErrInvalidSubtag},
		{name: "Duplicate variant", tag: "de-DE-1901-1901", wantErr: ErrDuplicateVariant},
		{name: "Duplicate singleton", tag: "ar-a-aaa-b-bbb-a-ccc", wantErr: ErrDuplicateSingleton},
		{name: "Too many extlangs", tag: "zh-gan-gan", wantErr: ErrTooManyExtlangs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.ParseAndNormalize(tt.tag)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseAndNormalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.String() != tt.wantTag {
				t.Errorf("ParseAndNormalize() got = %q, want %q", got.String(), tt.wantTag)
			}
		})
	}
}

// TestParser_ToExtlangForm tests converting a canonical tag to its extlang form.
// RFC 5646 Section 4.5.
func TestParser_ToExtlangForm(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		wantTag     string
		wantErr     bool
		expectNoop  bool
		isGrandfath bool
	}{
		{name: "Canonical to extlang", tag: "hak-CN", wantTag: "zh-hak-CN"},
		{name: "Primary language is an extlang", tag: "yue", wantTag: "zh-yue"},
		{name: "Language is not an extlang", tag: "en-US", expectNoop: true},
		{name: "Tag is already in extlang form", tag: "zh-hak-CN", wantTag: "zh-hak-CN", expectNoop: false},
		{name: "Grandfathered tag", tag: "i-klingon", expectNoop: true, isGrandfath: true},
		{name: "Private use only tag", tag: "x-my-tag", expectNoop: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lt LanguageTag
			if tt.isGrandfath {
				lt = mustParse(t, tt.tag)
			} else {
				lt = mustParseAndNormalize(t, tt.tag)
			}

			got, err := p.ToExtlangForm(lt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToExtlangForm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var finalTag string
			if tt.expectNoop {
				finalTag = lt.String()
			} else {
				finalTag = tt.wantTag
			}

			if got.String() != finalTag {
				t.Errorf("ToExtlangForm() got = %q, want %q", got.String(), finalTag)
			}
		})
	}
}

// TestParseAndNormalize_MalformedCanonicalization verifies that if the canonicalization
// process itself produces a malformed tag, the second internal parse catches it.
func TestParseAndNormalize_MalformedCanonicalization(t *testing.T) {
	malformedRegistry := `
File-Date: 2024-01-01
%%
Type: language
Subtag: en
Description: English
Added: 2005-10-16
%%
Type: language
Subtag: bad
Description: A bad subtag for testing to be replaced by a malformed value
Added: 2024-01-01
Deprecated: 2024-01-01
Preferred-Value: en--US
`
	reg, err := ParseRegistry(strings.NewReader(malformedRegistry))
	if err != nil {
		t.Fatalf("Failed to parse custom registry: %v", err)
	}
	malformedParser := &Parser{registry: reg}

	_, err = malformedParser.ParseAndNormalize("bad")

	if !errors.Is(err, ErrEmptySubtag) {
		t.Errorf("ParseAndNormalize() with malformed canonical value returned error %v, want %v", err, ErrEmptySubtag)
	}
}

// TestParser_ToExtlangForm_CorruptRegistry tests that ToExtlangForm handles
// a malformed prefix from a corrupt registry.
func TestParser_ToExtlangForm_CorruptRegistry(t *testing.T) {
	malformedRegistry := &Registry{
		Records: map[string]Record{
			"extlang:hak": {
				Type:           "extlang",
				Subtag:         "hak",
				Description:    []string{"Hakka Chinese"},
				Added:          "2009-07-29",
				PreferredValue: "hak",
				Prefix:         []string{"zh--badprefix"},
				Macrolanguage:  "zh",
			},
			"language:hak": {
				Type:        "language",
				Subtag:      "hak",
				Description: []string{"Hakka Chinese"},
				Added:       "2009-07-29",
			},
		},
	}

	corruptParser := &Parser{registry: malformedRegistry}

	lt, err := corruptParser.ParseAndNormalize("hak")
	if err != nil {
		t.Fatalf("Initial ParseAndNormalize failed unexpectedly: %v", err)
	}

	_, err = corruptParser.ToExtlangForm(lt)

	if !errors.Is(err, ErrEmptySubtag) {
		t.Errorf("ToExtlangForm with corrupt registry did not return the expected error.\nGot: %v\nWant: %v",
			err, ErrEmptySubtag)
	}
}
