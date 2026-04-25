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
	"encoding/json"
	"reflect"
	"testing"
)

// TestLanguageTag_String tests the String() method.
// Based on RFC 5646: a language tag is a sequence of subtags.
func TestLanguageTag_String(t *testing.T) {
	lt := mustParseAndNormalize(t, "en-US")
	if got := lt.String(); got != "en-US" {
		t.Errorf("String() = %q, want %q", got, "en-US")
	}
}

// TestLanguageTag_PrimaryLanguage tests the PrimaryLanguage() method.
// RFC 5646 Section 2.2.1: primary language subtag is the first subtag.
func TestLanguageTag_PrimaryLanguage(t *testing.T) {
	lt := mustParseAndNormalize(t, "sr-Latn-RS")
	want := "sr"
	if got := lt.PrimaryLanguage(); got != want {
		t.Errorf("PrimaryLanguage() = %q, want %q", got, want)
	}
}

// TestLanguageTag_ExtendedLanguage tests the ExtendedLanguage() method.
// RFC 5646 Section 2.2.2.
func TestLanguageTag_ExtendedLanguage(t *testing.T) {
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

// TestLanguageTag_ExtendedLanguageSubtags tests the ExtendedLanguageSubtags() method.
// RFC 5646 Section 2.1 ABNF: extlang = 3ALPHA *2("-" 3ALPHA).
func TestLanguageTag_ExtendedLanguageSubtags(t *testing.T) {
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

// TestLanguageTag_FullLanguage tests the FullLanguage() method.
func TestLanguageTag_FullLanguage(t *testing.T) {
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

// TestLanguageTag_Script tests the Script() method.
// RFC 5646 Section 2.2.3. Example: sr-Latn.
func TestLanguageTag_Script(t *testing.T) {
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

// TestLanguageTag_Region tests the Region() method.
// RFC 5646 Section 2.2.4. Examples: en-US, es-419.
func TestLanguageTag_Region(t *testing.T) {
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

// TestLanguageTag_Variant tests the Variant() method.
// RFC 5646 Section 2.2.5. Example: sl-rozaj-biske.
func TestLanguageTag_Variant(t *testing.T) {
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

// TestLanguageTag_VariantSubtags tests the VariantSubtags() method.
func TestLanguageTag_VariantSubtags(t *testing.T) {
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

// TestLanguageTag_ExtensionSubtags tests the ExtensionSubtags() method.
// RFC 5646 Section 2.2.6.
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

// TestLanguageTag_PrivateUse tests the PrivateUse() method.
// RFC 5646 Section 2.2.7. Examples: de-CH-x-phonebk, x-whatever.
func TestLanguageTag_PrivateUse(t *testing.T) {
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

// TestLanguageTag_PrivateUseSubtags tests the PrivateUseSubtags() method.
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

// TestLanguageTag_IsGrandfathered tests the IsGrandfathered() method.
// RFC 5646 Section 2.2.8.
func TestLanguageTag_IsGrandfathered(t *testing.T) {
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
// Per RFC 5646 Sec 2.2.9, this implies full validation and canonicalization.
func TestLanguageTag_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantTag string
		wantErr bool
	}{
		{name: "Valid tag", data: []byte(`"en-US"`), wantTag: "en-US"},
		{name: "Canonicalization applied", data: []byte(`"art-lojban"`), wantTag: "jbo"},
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
