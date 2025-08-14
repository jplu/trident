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
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"reflect"
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

// TestLanguageTag_String tests the String() method.
// Based on RFC 5646, a language tag is a sequence of subtags. This test
// ensures the string representation is correct after parsing.
func TestLanguageTag_String(t *testing.T) {
	lt := mustParseAndNormalize(t, "en-US")
	if got := lt.String(); got != "en-US" {
		t.Errorf("String() = %q, want %q", got, "en-US")
	}
}

// TestLanguageTag_AsStr tests the AsStr() method. This is an alias for String().
func TestLanguageTag_AsStr(t *testing.T) {
	lt := mustParseAndNormalize(t, "de-DE")
	if got := lt.AsStr(); got != "de-DE" {
		t.Errorf("AsStr() = %q, want %q", got, "de-DE")
	}
}

// TestLanguageTag_PrimaryLanguage tests the PrimaryLanguage() method.
// RFC 5646 Section 2.2.1 defines the primary language subtag as the first subtag.
func TestLanguageTag_PrimaryLanguage(t *testing.T) {
	lt := mustParseAndNormalize(t, "sr-Latn-RS")
	want := "sr"
	if got := lt.PrimaryLanguage(); got != want {
		t.Errorf("PrimaryLanguage() = %q, want %q", got, want)
	}
}

// TestLanguageTag_ExtendedLanguage tests the ExtendedLanguage() method.
// RFC 5646 Section 2.2.2 defines extended language subtags.
// Example from RFC Appendix A: zh-cmn-Hans-CN.
func TestLanguageTag_ExtendedLanguage(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		want     string
		wantOK   bool
		useParse bool
	}{
		{
			name:   "With extlang",
			tag:    "zh-yue-HK",
			want:   "yue",
			wantOK: true,
			// Use ParseAndNormalize to get "yue-HK", then convert to extlang form
		},
		{
			name:   "Without extlang",
			tag:    "en-US",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lt LanguageTag
			if tt.name == "With extlang" {
				// We create one from its canonical form as per RFC 4.5.
				canonical := mustParseAndNormalize(t, "yue-HK")
				var err error
				lt, err = p.ToExtlangForm(canonical)
				if err != nil {
					t.Fatalf("ToExtlangForm failed: %v", err)
				}
				if lt.String() != "zh-yue-HK" {
					t.Fatalf("ToExtlangForm result was %s, want zh-yue-HK", lt.String())
				}
			} else {
				lt = mustParseAndNormalize(t, tt.tag)
			}

			got, gotOK := lt.ExtendedLanguage()
			if got != tt.want {
				t.Errorf("ExtendedLanguage() got = %v, want %v", got, tt.want)
			}
			if gotOK != tt.wantOK {
				t.Errorf("ExtendedLanguage() gotOK = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}

// TestLanguageTag_ExtendedLanguageSubtags tests the ExtendedLanguageSubtags() method.
// RFC 5646 Section 2.1 ABNF defines `extlang = 3ALPHA *2("-" 3ALPHA)`.
// However, Section 2.2.2 clarifies only one extlang is currently permitted.
func TestLanguageTag_ExtendedLanguageSubtags(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want []string
	}{
		{
			name: "With extlang",
			tag:  "zh-yue-HK",
			want: []string{"yue"},
		},
		{
			name: "Without extlang",
			tag:  "en-US",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lt LanguageTag
			if tt.name == "With extlang" {
				canonical := mustParseAndNormalize(t, "yue-HK")
				var err error
				lt, err = p.ToExtlangForm(canonical)
				if err != nil {
					t.Fatalf("ToExtlangForm failed: %v", err)
				}
			} else {
				lt = mustParseAndNormalize(t, tt.tag)
			}

			if got := lt.ExtendedLanguageSubtags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtendedLanguageSubtags() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLanguageTag_FullLanguage tests the FullLanguage() method.
// It should return the primary language and any extended language subtags.
func TestLanguageTag_FullLanguage(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want string
	}{
		{
			name: "With extlang",
			tag:  "zh-yue-HK",
			want: "zh-yue",
		},
		{
			name: "Without extlang",
			tag:  "sr-Latn-RS",
			want: "sr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lt LanguageTag
			if tt.name == "With extlang" {
				canonical := mustParseAndNormalize(t, "yue-HK")
				var err error
				lt, err = p.ToExtlangForm(canonical)
				if err != nil {
					t.Fatalf("ToExtlangForm failed: %v", err)
				}
			} else {
				lt = mustParseAndNormalize(t, tt.tag)
			}

			if got := lt.FullLanguage(); got != tt.want {
				t.Errorf("FullLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestLanguageTag_Script tests the Script() method.
// RFC 5646 Section 2.2.3 defines the script subtag.
// Example from RFC Appendix A: sr-Latn.
func TestLanguageTag_Script(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		want   string
		wantOK bool
	}{
		{
			name:   "With script",
			tag:    "sr-Latn-RS",
			want:   "Latn",
			wantOK: true,
		},
		{
			name:   "Without script",
			tag:    "en-US",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := mustParseAndNormalize(t, tt.tag)
			got, gotOK := lt.Script()
			if got != tt.want {
				t.Errorf("Script() got = %q, want %q", got, tt.want)
			}
			if gotOK != tt.wantOK {
				t.Errorf("Script() gotOK = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}

// TestLanguageTag_Region tests the Region() method.
// RFC 5646 Section 2.2.4 defines the region subtag.
// Examples from RFC Appendix A: en-US, es-419.
func TestLanguageTag_Region(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		want   string
		wantOK bool
	}{
		{
			name:   "With 2-letter region",
			tag:    "de-DE",
			want:   "DE",
			wantOK: true,
		},
		{
			name:   "With 3-digit region",
			tag:    "es-419",
			want:   "419",
			wantOK: true,
		},
		{
			name:   "Without region",
			tag:    "fr-Latn",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := mustParseAndNormalize(t, tt.tag)
			got, gotOK := lt.Region()
			if got != tt.want {
				t.Errorf("Region() got = %q, want %q", got, tt.want)
			}
			if gotOK != tt.wantOK {
				t.Errorf("Region() gotOK = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}

// TestLanguageTag_Variant tests the Variant() method.
// RFC 5646 Section 2.2.5 defines variant subtags.
// Example from RFC Appendix A: sl-rozaj-biske.
func TestLanguageTag_Variant(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		want   string
		wantOK bool
	}{
		{
			name:   "With one variant",
			tag:    "sl-nedis",
			want:   "nedis",
			wantOK: true,
		},
		{
			name:   "With multiple variants",
			tag:    "sl-rozaj-biske",
			want:   "rozaj-biske",
			wantOK: true,
		},
		{
			name:   "Without variant",
			tag:    "en-US",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := mustParseAndNormalize(t, tt.tag)
			got, gotOK := lt.Variant()
			if got != tt.want {
				t.Errorf("Variant() got = %q, want %q", got, tt.want)
			}
			if gotOK != tt.wantOK {
				t.Errorf("Variant() gotOK = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}

// TestLanguageTag_VariantSubtags tests the VariantSubtags() method.
// This method should split the variant string into a slice of subtags.
func TestLanguageTag_VariantSubtags(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want []string
	}{
		{
			name: "With multiple variants",
			tag:  "sl-rozaj-biske",
			want: []string{"rozaj", "biske"},
		},
		{
			name: "Without variant",
			tag:  "en-US",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := mustParseAndNormalize(t, tt.tag)
			if got := lt.VariantSubtags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VariantSubtags() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLanguageTag_ExtensionSubtags tests the ExtensionSubtags() method.
// RFC 5646 Section 2.2.6 defines extension subtags.
func TestLanguageTag_ExtensionSubtags(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want []Extension
	}{
		{
			name: "With one extension",
			tag:  "en-US-u-islamcal",
			want: []Extension{{Singleton: 'u', Value: "islamcal"}},
		},
		{
			name: "With multiple extensions",
			tag:  "zh-CN-a-myext-b-another",
			// Note: extensions are sorted by singleton in canonical form
			want: []Extension{
				{Singleton: 'a', Value: "myext"},
				{Singleton: 'b', Value: "another"},
			},
		},
		{
			name: "Without extensions",
			tag:  "en-US",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extensions are not validated against a registry, so Parse is sufficient
			lt := mustParse(t, tt.tag)
			got := lt.ExtensionSubtags()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtensionSubtags() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLanguageTag_PrivateUse tests the PrivateUse() method.
// RFC 5646 Section 2.2.7 defines private use subtags, starting with 'x'.
// Examples from RFC Appendix A: de-CH-x-phonebk, x-whatever.
func TestLanguageTag_PrivateUse(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		want   string
		wantOK bool
	}{
		{
			name:   "With private use section",
			tag:    "de-CH-x-phonebk",
			want:   "phonebk",
			wantOK: true,
		},
		{
			name:   "With complex private use section and case normalization",
			tag:    "az-Arab-x-AZE-derbend",
			want:   "aze-derbend",
			wantOK: true,
		},
		{
			name:   "Tag is only private use",
			tag:    "x-whatever",
			want:   "whatever",
			wantOK: true,
		},
		{
			name:   "Without private use section",
			tag:    "en-US",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Private use subtags are not validated, so Parse is sufficient
			lt := mustParse(t, tt.tag)
			got, gotOK := lt.PrivateUse()
			if got != tt.want {
				t.Errorf("PrivateUse() got = %q, want %q", got, tt.want)
			}
			if gotOK != tt.wantOK {
				t.Errorf("PrivateUse() gotOK = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}

// TestLanguageTag_PrivateUseSubtags tests the PrivateUseSubtags() method.
// This method should split the private use string into a slice of subtags.
func TestLanguageTag_PrivateUseSubtags(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want []string
	}{
		{
			name: "With multiple private use subtags",
			tag:  "az-Arab-x-AZE-derbend",
			want: []string{"aze", "derbend"},
		},
		{
			name: "Without private use subtags",
			tag:  "en-US",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := mustParse(t, tt.tag)
			if got := lt.PrivateUseSubtags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrivateUseSubtags() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLanguageTag_IsGrandfathered tests the IsGrandfathered() method.
// RFC 5646 Section 2.2.8 defines grandfathered tags. The 'Parse' method
// should identify them. 'ParseAndNormalize' may replace them.
func TestLanguageTag_IsGrandfathered(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{
			name: "Irregular grandfathered",
			tag:  "i-klingon",
			want: true,
		},
		{
			name: "Regular grandfathered",
			tag:  "en-GB-oed",
			want: true,
		},
		{
			name: "Not grandfathered",
			tag:  "en-US",
			want: false,
		},
		{
			name: "Redundant (treated as grandfathered by Parse)",
			tag:  "zh-hakka",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse should identify a tag as grandfathered without replacing it.
			lt := mustParse(t, tt.tag)
			if got := lt.IsGrandfathered(); got != tt.want {
				t.Errorf("IsGrandfathered() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLanguageTag_MarshalJSON tests the MarshalJSON method.
func TestLanguageTag_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		lt      *LanguageTag
		want    []byte
		wantErr bool
	}{
		{
			name: "Valid tag",
			lt:   func() *LanguageTag { l := mustParseAndNormalize(t, "de-CH"); return &l }(),
			want: []byte(`"de-CH"`),
		},
		{
			name: "Empty tag",
			lt:   &LanguageTag{},
			want: []byte(`""`),
		},
		{
			name: "Nil tag",
			lt:   nil,
			want: []byte("null"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.lt)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

// TestLanguageTag_UnmarshalJSON tests the UnmarshalJSON method.
// The method should parse and validate the tag from a JSON string.
// As per RFC 5646 Sec 2.2.9, this implies full validation and canonicalization.
func TestLanguageTag_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantTag string
		wantErr bool
	}{
		{
			name:    "Valid tag",
			data:    []byte(`"en-US"`),
			wantTag: "en-US",
		},
		{
			name:    "Canonicalization applied",
			data:    []byte(`"art-lojban"`), // RFC 2.2.8: superseded by 'jbo'
			wantTag: "jbo",
		},
		{
			name:    "Case normalization applied",
			data:    []byte(`"sR-lAtN-rs"`),
			wantTag: "sr-Latn-RS",
		},
		{
			name:    "Invalid tag in JSON",
			data:    []byte(`"123-bogus"`), // RFC 2.2.9: 'valid' requires registered subtags
			wantErr: true,
		},
		{
			name:    "Empty JSON string",
			data:    []byte(`""`),
			wantTag: "",
		},
		{
			name:    "JSON null",
			data:    []byte("null"),
			wantTag: "",
		},
		{
			name:    "Not a JSON string",
			data:    []byte("123"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lt LanguageTag
			err := json.Unmarshal(tt.data, &lt)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got := lt.String(); got != tt.wantTag {
					t.Errorf("UnmarshalJSON() got tag %q, want %q", got, tt.wantTag)
				}
			}
		})
	}

	t.Run("NewParser failure", func(t *testing.T) {
		originalData := embeddedRegistryData
		t.Cleanup(func() {
			embeddedRegistryData = originalData
		})

		embeddedRegistryData = []byte{}

		var lt LanguageTag
		jsonData := []byte(`"en-US"`)
		err := json.Unmarshal(jsonData, &lt)

		if err == nil {
			t.Fatal("UnmarshalJSON() did not return an error, but was expected to")
		}

		wantErrMsg := "embedded language-subtag-registry file is empty or not found"
		if err.Error() != wantErrMsg {
			t.Errorf("UnmarshalJSON() error = %q, want %q", err, wantErrMsg)
		}
	})
}

// TestParser_Parse tests the non-validating Parse method.
// RFC 5646 Section 2.2.9 defines "well-formed" as conforming to the ABNF.
// This test checks for well-formedness and case normalization, not validity.
func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		wantTag string
		wantErr error
	}{
		// Well-formed cases from RFC Appendix A
		{name: "Simple tag", tag: "de", wantTag: "de"},
		{name: "Language-Region", tag: "en-US", wantTag: "en-US"},
		{name: "Language-Script-Region", tag: "sr-Latn-RS", wantTag: "sr-Latn-RS"},
		{name: "Case normalization", tag: "MN-cYRL-mn", wantTag: "mn-Cyrl-MN"}, // RFC 2.1.1
		{name: "Private use", tag: "de-CH-x-phonebk", wantTag: "de-CH-x-phonebk"},
		{name: "Private use only", tag: "x-whatever", wantTag: "x-whatever"},
		{name: "Grandfathered irregular", tag: "i-klingon", wantTag: "i-klingon"},
		{name: "Grandfathered regular", tag: "art-lojban", wantTag: "art-lojban"},
		{name: "Extension", tag: "en-a-myext-b-another", wantTag: "en-a-myext-b-another"},

		// Well-formed but not valid (should pass Parse)
		{name: "Unregistered language", tag: "zz-US", wantTag: "zz-US"},
		{name: "Unregistered script", tag: "en-Zzzz-US", wantTag: "en-Zzzz-US"},
		{name: "Duplicate variant", tag: "de-DE-1901-1901", wantTag: "de-DE-1901-1901"},
		{name: "Duplicate singleton", tag: "en-a-foo-a-bar", wantTag: "en-a-foo-a-bar"},

		// Not well-formed cases from RFC Appendix A and general syntax
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
		// Canonicalization cases from RFC 4.5
		{name: "Redundant tag replacement", tag: "zh-min-nan", wantTag: "nan"},
		{name: "Grandfathered replacement (art-lojban)", tag: "art-lojban", wantTag: "jbo"},
		{name: "Grandfathered replacement (i-klingon)", tag: "i-klingon", wantTag: "tlh"},
		{name: "Grandfathered no-replacement", tag: "i-enochian", wantTag: "i-enochian"},
		{name: "Subtag replacement", tag: "en-BU", wantTag: "en-MM"},
		{name: "Extlang canonicalization", tag: "zh-gan", wantTag: "gan"},
		{name: "Extension reordering", tag: "en-b-ccc-a-aaa", wantTag: "en-a-aaa-b-ccc"},
		{name: "Script suppression", tag: "is-Latn", wantTag: "is"},
		{name: "Case canonicalization", tag: "SR-LATN-rs", wantTag: "sr-Latn-RS"},

		// Validity error cases from RFC Appendix A and 2.2.9
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
// RFC 5646 Section 4.5 defines the 'extlang form'.
func TestParser_ToExtlangForm(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		wantTag     string
		wantErr     bool
		expectNoop  bool
		isGrandfath bool
	}{
		{
			name:    "Canonical to extlang",
			tag:     "hak-CN", // RFC 4.5 example
			wantTag: "zh-hak-CN",
		},
		{
			name:    "Primary language is an extlang",
			tag:     "yue", // RFC 4.1.2
			wantTag: "zh-yue",
		},
		{
			name:       "Language is not an extlang",
			tag:        "en-US",
			expectNoop: true,
		},
		{
			name:       "Tag is already in extlang form",
			tag:        "zh-hak-CN",
			wantTag:    "zh-hak-CN",
			expectNoop: false,
		},
		{
			name:        "Grandfathered tag",
			tag:         "i-klingon",
			expectNoop:  true,
			isGrandfath: true,
		},
		{
			name:       "Private use only tag",
			tag:        "x-my-tag",
			expectNoop: true,
		},
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
// process itself produces a malformed tag (e.g., due to a bad registry entry),
// the second internal parse catches it and returns an error. This covers the
// defensive error check in ParseAndNormalize.
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

	// We expect the internal re-parse of the canonical tag "en--us" to fail.
	// The canonicalize step replaces "bad" with "en--US", which is rendered
	// as "en--us". The reparsing of this string fails with ErrEmptySubtag.
	if !errors.Is(err, ErrEmptySubtag) {
		t.Errorf("ParseAndNormalize() with malformed canonical value returned error %v, want %v", err, ErrEmptySubtag)
	}
}

// TestParser_ToExtlangForm_CorruptRegistry tests that ToExtlangForm can handle
// a malformed prefix from a corrupt registry, exercising a defensive error check.
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
