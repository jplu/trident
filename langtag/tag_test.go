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
	"encoding/json"
	"reflect"
	"testing"
)

// TestLanguageTagString checks the formatted string representation of a language tag.
func TestLanguageTagString(t *testing.T) {
	lt := mustParseAndNormalize(t, "en-US")
	if got := lt.String(); got != "en-US" {
		t.Errorf("String() = %q, want %q", got, "en-US")
	}
}

// TestLanguageTagPrimaryLanguage verifies extraction of the base language subtag.
func TestLanguageTagPrimaryLanguage(t *testing.T) {
	lt := mustParseAndNormalize(t, "sr-Latn-RS")
	want := "sr"
	if got := lt.PrimaryLanguage(); got != want {
		t.Errorf("PrimaryLanguage() = %q, want %q", got, want)
	}
}

// TestLanguageTagExtendedLanguage evaluates extended language components and their existence.
func TestLanguageTagExtendedLanguage(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		want     string
		wantOK   bool
		useParse bool
	}{
		{name: "With extlang", tag: "zh-yue-HK", want: "yue", wantOK: true},
		{name: "Without extlang", tag: "en-US", want: "", wantOK: false},
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

// TestLanguageTagExtendedLanguageSubtags retrieves all extended language elements as a slice.
func TestLanguageTagExtendedLanguageSubtags(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want []string
	}{
		{name: "With extlang", tag: "zh-yue-HK", want: []string{"yue"}},
		{name: "Without extlang", tag: "en-US", want: nil},
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

// TestLanguageTagFullLanguage combines primary and extended language components correctly.
func TestLanguageTagFullLanguage(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want string
	}{
		{name: "With extlang", tag: "zh-yue-HK", want: "zh-yue"},
		{name: "Without extlang", tag: "sr-Latn-RS", want: "sr"},
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

// TestLanguageTagScript extracts the script subtag when present.
func TestLanguageTagScript(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		want   string
		wantOK bool
	}{
		{name: "With script", tag: "sr-Latn-RS", want: "Latn", wantOK: true},
		{name: "Without script", tag: "en-US", want: "", wantOK: false},
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

// TestLanguageTagRegion extracts country or region codes (letters or digits).
func TestLanguageTagRegion(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		want   string
		wantOK bool
	}{
		{name: "With 2-letter region", tag: "de-DE", want: "DE", wantOK: true},
		{name: "With 3-digit region", tag: "es-419", want: "419", wantOK: true},
		{name: "Without region", tag: "fr-Latn", want: "", wantOK: false},
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

// TestLanguageTagVariant extracts dialect or variation subtags.
func TestLanguageTagVariant(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		want   string
		wantOK bool
	}{
		{name: "With one variant", tag: "sl-nedis", want: "nedis", wantOK: true},
		{name: "With multiple variants", tag: "sl-rozaj-biske", want: "rozaj-biske", wantOK: true},
		{name: "Without variant", tag: "en-US", want: "", wantOK: false},
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

// TestLanguageTagVariantSubtags returns all variation subtags as a slice.
func TestLanguageTagVariantSubtags(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want []string
	}{
		{name: "With multiple variants", tag: "sl-rozaj-biske", want: []string{"rozaj", "biske"}},
		{name: "Without variant", tag: "en-US", want: nil},
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

// TestLanguageTagExtensionSubtags extracts extension singletons and their values.
func TestLanguageTagExtensionSubtags(t *testing.T) {
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
			want: []Extension{
				{Singleton: 'a', Value: "myext"},
				{Singleton: 'b', Value: "another"},
			},
		},
		{name: "Without extensions", tag: "en-US", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := mustParse(t, tt.tag)
			got := lt.ExtensionSubtags()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtensionSubtags() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLanguageTagPrivateUse extracts and normalizes private-use subtag sequences.
func TestLanguageTagPrivateUse(t *testing.T) {
	tests := []struct {
		name   string
		tag    string
		want   string
		wantOK bool
	}{
		{name: "With private use section", tag: "de-CH-x-phonebk", want: "phonebk", wantOK: true},
		{
			name:   "With complex private use section and case normalization",
			tag:    "az-Arab-x-AZE-derbend",
			want:   "aze-derbend",
			wantOK: true,
		},
		{name: "Tag is only private use", tag: "x-whatever", want: "whatever", wantOK: true},
		{name: "Without private use section", tag: "en-US", want: "", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// TestLanguageTagPrivateUseSubtags splits private-use sections into individual elements.
func TestLanguageTagPrivateUseSubtags(t *testing.T) {
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
		{name: "Without private use subtags", tag: "en-US", want: nil},
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

// TestLanguageTagIsGrandfathered identifies irregular, regular, and redundant grandfathered tags.
func TestLanguageTagIsGrandfathered(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{name: "Irregular grandfathered", tag: "i-klingon", want: true},
		{name: "Regular grandfathered", tag: "en-GB-oed", want: true},
		{name: "Not grandfathered", tag: "en-US", want: false},
		{name: "Redundant (treated as grandfathered by Parse)", tag: "zh-hakka", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := mustParse(t, tt.tag)
			if got := lt.IsGrandfathered(); got != tt.want {
				t.Errorf("IsGrandfathered() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLanguageTagMarshalJSON serializes a LanguageTag to JSON across various states.
func TestLanguageTagMarshalJSON(t *testing.T) {
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

// TestLanguageTagUnmarshalJSON deserializes JSON into a LanguageTag with validation and normalization.
func TestLanguageTagUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantTag string
		wantErr bool
	}{
		{name: "Valid tag", data: []byte(`"en-US"`), wantTag: "en-US"},
		{name: "Canonicalization not applied", data: []byte(`"art-lojban"`), wantTag: "art-lojban"},
		{name: "Case normalization applied", data: []byte(`"sR-lAtN-rs"`), wantTag: "sr-Latn-RS"},
		{name: "Invalid tag in JSON", data: []byte(`"123-bogus"`), wantErr: true},
		{name: "Empty JSON string", data: []byte(`""`), wantTag: ""},
		{name: "JSON null", data: []byte("null"), wantTag: ""},
		{name: "Not a JSON string", data: []byte("123"), wantErr: true},
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
}
