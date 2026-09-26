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
package datatypes

import (
	"testing"
)

// TestParseLanguage tests valid and invalid lexical inputs for xsd:language.
func TestParseLanguage(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Language
		wantError bool
	}{
		{
			name:      "valid primary language tag",
			input:     "en",
			want:      Language("en"),
			wantError: false,
		},
		{
			name:      "valid language with subtag",
			input:     "en-US",
			want:      Language("en-US"),
			wantError: false,
		},
		{
			name:      "valid multi-subtag language with numbers",
			input:     "zh-Hant-CN-1996",
			want:      Language("zh-Hant-CN-1996"),
			wantError: false,
		},
		{
			name:      "valid single letter primary subtag",
			input:     "i-klingon",
			want:      Language("i-klingon"),
			wantError: false,
		},
		{
			name:      "valid with whitespace collapse",
			input:     "   fr-FR \t\n ",
			want:      Language("fr-FR"),
			wantError: false,
		},
		{
			name:      "invalid empty string",
			input:     "",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid whitespace only",
			input:     "   \t\n ",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid primary tag too long",
			input:     "toolongprimarytag",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid subtag too long",
			input:     "en-toolongsubtag",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid leading hyphen",
			input:     "-en",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid trailing hyphen",
			input:     "en-",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid consecutive hyphens",
			input:     "en--US",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid underscore separator",
			input:     "en_US",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid characters",
			input:     "en@US",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid numeric primary tag",
			input:     "123",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLanguage(tt.input)
			if (err != nil) != tt.wantError {
				t.Fatalf("ParseLanguage(%q) error = %v, wantError = %v", tt.input, err, tt.wantError)
			}
			if got != tt.want {
				t.Errorf("ParseLanguage(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestLanguageString tests the string representation of Language values.
func TestLanguageString(t *testing.T) {
	tests := []struct {
		name string
		lang Language
		want string
	}{
		{
			name: "primary language string",
			lang: Language("en"),
			want: "en",
		},
		{
			name: "subtag language string",
			lang: Language("en-US"),
			want: "en-US",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lang.String(); got != tt.want {
				t.Errorf("Language.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestLanguageIsIdenticalWith tests identity comparisons for Language values.
func TestLanguageIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name  string
		lang  Language
		other XSDValue
		want  bool
	}{
		{
			name:  "identical language values",
			lang:  Language("en-US"),
			other: Language("en-US"),
			want:  true,
		},
		{
			name:  "different language values",
			lang:  Language("en-US"),
			other: Language("en-GB"),
			want:  false,
		},
		{
			name:  "case sensitive mismatch",
			lang:  Language("en-US"),
			other: Language("en-us"),
			want:  false,
		},
		{
			name:  "different xsd datatype same string value",
			lang:  Language("en-US"),
			other: String("en-US"),
			want:  false,
		},
		{
			name:  "token xsd datatype same string value",
			lang:  Language("en-US"),
			other: Token("en-US"),
			want:  false,
		},
		{
			name:  "nil other value",
			lang:  Language("en-US"),
			other: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lang.IsIdenticalWith(tt.other); got != tt.want {
				t.Errorf("Language.IsIdenticalWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLanguageLength tests the length computation of Language values.
func TestLanguageLength(t *testing.T) {
	tests := []struct {
		name string
		lang Language
		want int
	}{
		{
			name: "short tag length",
			lang: Language("en"),
			want: 2,
		},
		{
			name: "tag with hyphen length",
			lang: Language("en-US"),
			want: 5,
		},
		{
			name: "multi-subtag length",
			lang: Language("zh-Hant-CN"),
			want: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.lang.Length(); got != tt.want {
				t.Errorf("Language.Length() = %d, want %d", got, tt.want)
			}
		})
	}
}
