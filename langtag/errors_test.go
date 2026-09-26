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
	"errors"
	"fmt"
	"testing"
)

// TestErrorVariablesDefinitions verifies that all exported error variables are
// properly initialized with their defined error messages, are non-nil, and
// correctly participate in errors.Is comparisons both directly and wrapped.
func TestErrorVariablesDefinitions(t *testing.T) {
	testCases := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "ErrEmptyExtension",
			err:         ErrEmptyExtension,
			expectedMsg: "if an extension subtag is present, it must not be empty",
		},
		{
			name:        "ErrEmptyPrivateUse",
			err:         ErrEmptyPrivateUse,
			expectedMsg: "if the 'x' subtag is present, it must not be empty",
		},
		{
			name:        "ErrForbiddenChar",
			err:         ErrForbiddenChar,
			expectedMsg: "the langtag contains a char not allowed",
		},
		{
			name:        "ErrInvalidSubtag",
			err:         ErrInvalidSubtag,
			expectedMsg: "a subtag fails to parse or is not a valid IANA subtag",
		},
		{
			name:        "ErrInvalidLanguage",
			err:         ErrInvalidLanguage,
			expectedMsg: "the given language subtag is invalid",
		},
		{
			name:        "ErrSubtagTooLong",
			err:         ErrSubtagTooLong,
			expectedMsg: "a subtag may be eight characters in length at maximum",
		},
		{
			name:        "ErrEmptySubtag",
			err:         ErrEmptySubtag,
			expectedMsg: "a subtag should not be empty",
		},
		{
			name:        "ErrTooManyExtlangs",
			err:         ErrTooManyExtlangs,
			expectedMsg: "at maximum one extlang is allowed",
		},
		{
			name:        "ErrDuplicateVariant",
			err:         ErrDuplicateVariant,
			expectedMsg: "the same variant subtag appears more than once",
		},
		{
			name:        "ErrDuplicateSingleton",
			err:         ErrDuplicateSingleton,
			expectedMsg: "the same extension singleton appears more than once",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatalf("expected %s to be non-nil", tc.name)
			}

			if got := tc.err.Error(); got != tc.expectedMsg {
				t.Errorf("%s.Error() = %q, want %q", tc.name, got, tc.expectedMsg)
			}

			if !errors.Is(tc.err, tc.err) {
				t.Errorf("errors.Is(%s, %s) = false, want true", tc.name, tc.name)
			}

			wrapped := fmt.Errorf("context wrapper: %w", tc.err)
			if !errors.Is(wrapped, tc.err) {
				t.Errorf("errors.Is(wrapped, %s) = false, want true", tc.name)
			}
		})
	}
}

// TestErrorVariablesUniqueness verifies that every exported error is a distinct
// sentinel error instance and that no two errors share identical error messages.
func TestErrorVariablesUniqueness(t *testing.T) {
	allErrors := []struct {
		name string
		err  error
	}{
		{"ErrEmptyExtension", ErrEmptyExtension},
		{"ErrEmptyPrivateUse", ErrEmptyPrivateUse},
		{"ErrForbiddenChar", ErrForbiddenChar},
		{"ErrInvalidSubtag", ErrInvalidSubtag},
		{"ErrInvalidLanguage", ErrInvalidLanguage},
		{"ErrSubtagTooLong", ErrSubtagTooLong},
		{"ErrEmptySubtag", ErrEmptySubtag},
		{"ErrTooManyExtlangs", ErrTooManyExtlangs},
		{"ErrDuplicateVariant", ErrDuplicateVariant},
		{"ErrDuplicateSingleton", ErrDuplicateSingleton},
	}

	for i := range allErrors {
		for j := i + 1; j < len(allErrors); j++ {
			err1 := allErrors[i]
			err2 := allErrors[j]

			if errors.Is(err1.err, err2.err) {
				t.Errorf("expected %s and %s to be distinct sentinel errors", err1.name, err2.name)
			}

			if err1.err.Error() == err2.err.Error() {
				t.Errorf(
					"expected %s and %s to have distinct messages, both have %q",
					err1.name,
					err2.name,
					err1.err.Error(),
				)
			}
		}
	}
}

// TestTypeExtlangConstant verifies the value of the unexported typeExtlang constant.
func TestTypeExtlangConstant(t *testing.T) {
	const expected = "extlang"
	if typeExtlang != expected {
		t.Errorf("typeExtlang = %q, want %q", typeExtlang, expected)
	}
}

// TestParserTriggersErrForbiddenChar verifies that ErrForbiddenChar is returned when
// language tags contain non-ASCII, punctuation, or whitespace characters.
func TestParserTriggersErrForbiddenChar(t *testing.T) {
	invalidTags := []string{
		"en_US",
		"en@US",
		"en US",
		"en.US",
		"en/US",
		"en\tUS",
		"en\nUS",
		"en-é",
		"en-日本語",
	}

	for _, tag := range invalidTags {
		t.Run(tag, func(t *testing.T) {
			_, err := ParseWellFormed(tag)
			if !errors.Is(err, ErrForbiddenChar) {
				t.Errorf("ParseWellFormed(%q) error = %v, want ErrForbiddenChar", tag, err)
			}

			p := &Parser{registry: &Registry{Records: make(map[string]Record)}}
			_, err = p.Parse(tag)
			if !errors.Is(err, ErrForbiddenChar) {
				t.Errorf("Parser.Parse(%q) error = %v, want ErrForbiddenChar", tag, err)
			}
		})
	}
}

// TestParserTriggersErrEmptySubtag verifies that ErrEmptySubtag is returned when
// consecutive hyphens, leading hyphens, or trailing hyphens are present.
func TestParserTriggersErrEmptySubtag(t *testing.T) {
	invalidTags := []string{
		"-en",
		"en-",
		"en--US",
		"zh-Hans--CN",
		"zh-Hans-CN-",
	}

	for _, tag := range invalidTags {
		t.Run(tag, func(t *testing.T) {
			_, err := ParseWellFormed(tag)
			if !errors.Is(err, ErrEmptySubtag) {
				t.Errorf("ParseWellFormed(%q) error = %v, want ErrEmptySubtag", tag, err)
			}
		})
	}
}

// TestParserTriggersErrSubtagTooLong verifies that ErrSubtagTooLong is returned
// when any subtag exceeds the maximum allowed length of eight characters.
func TestParserTriggersErrSubtagTooLong(t *testing.T) {
	invalidTags := []string{
		"toolonglanguage",
		"en-toolongsubtag",
		"en-Latn-toolongsubtag",
		"x-toolongprivatesubtag",
	}

	for _, tag := range invalidTags {
		t.Run(tag, func(t *testing.T) {
			_, err := ParseWellFormed(tag)
			if !errors.Is(err, ErrSubtagTooLong) {
				t.Errorf("ParseWellFormed(%q) error = %v, want ErrSubtagTooLong", tag, err)
			}
		})
	}
}

// TestParserTriggersErrInvalidLanguage verifies that ErrInvalidLanguage is returned
// when the primary language subtag is malformed or not found in the registry.
func TestParserTriggersErrInvalidLanguage(t *testing.T) {
	t.Run("syntactic invalidity", func(t *testing.T) {
		syntacticallyInvalid := []string{
			"a",
			"123",
			"1a",
		}

		for _, tag := range syntacticallyInvalid {
			_, err := ParseWellFormed(tag)
			if !errors.Is(err, ErrInvalidLanguage) {
				t.Errorf("ParseWellFormed(%q) error = %v, want ErrInvalidLanguage", tag, err)
			}
		}
	})

	t.Run("semantic registry missing", func(t *testing.T) {
		p := &Parser{
			registry: &Registry{
				Records: make(map[string]Record),
			},
		}

		_, err := p.ParseAndNormalize("zzzz")
		if !errors.Is(err, ErrInvalidLanguage) {
			t.Errorf("ParseAndNormalize(\"zzzz\") error = %v, want ErrInvalidLanguage", err)
		}
	})
}

// TestParserTriggersErrEmptyPrivateUse verifies that ErrEmptyPrivateUse is returned
// when the private-use singleton 'x' is not followed by any subtags.
func TestParserTriggersErrEmptyPrivateUse(t *testing.T) {
	invalidTags := []string{
		"x",
		"X",
		"en-x",
		"en-Latn-x",
	}

	for _, tag := range invalidTags {
		t.Run(tag, func(t *testing.T) {
			_, err := ParseWellFormed(tag)
			if !errors.Is(err, ErrEmptyPrivateUse) {
				t.Errorf("ParseWellFormed(%q) error = %v, want ErrEmptyPrivateUse", tag, err)
			}
		})
	}
}

// TestParserTriggersErrEmptyExtension verifies that ErrEmptyExtension is returned
// when an extension singleton is not followed by any extension subtags.
func TestParserTriggersErrEmptyExtension(t *testing.T) {
	invalidTags := []string{
		"en-a",
		"en-u",
		"en-a-b-val",
	}

	for _, tag := range invalidTags {
		t.Run(tag, func(t *testing.T) {
			_, err := ParseWellFormed(tag)
			if !errors.Is(err, ErrEmptyExtension) {
				t.Errorf("ParseWellFormed(%q) error = %v, want ErrEmptyExtension", tag, err)
			}
		})
	}
}

// TestParserTriggersErrTooManyExtlangs verifies that ErrTooManyExtlangs is returned
// when more extended language subtags are supplied than permitted.
func TestParserTriggersErrTooManyExtlangs(t *testing.T) {
	t.Run("syntactic limit exceeded", func(t *testing.T) {
		tag := "en-aaa-bbb-ccc-ddd"
		_, err := ParseWellFormed(tag)
		if !errors.Is(err, ErrTooManyExtlangs) {
			t.Errorf("ParseWellFormed(%q) error = %v, want ErrTooManyExtlangs", tag, err)
		}
	})

	t.Run("validity limit exceeded", func(t *testing.T) {
		p := &Parser{
			registry: &Registry{
				Records: map[string]Record{
					"language:zh": {Type: "language", Subtag: "zh"},
					"extlang:cmn": {Type: typeExtlang, Subtag: "cmn"},
					"extlang:hak": {Type: typeExtlang, Subtag: "hak"},
				},
			},
		}

		tag := "zh-cmn-hak"
		_, err := p.ParseAndNormalize(tag)
		if !errors.Is(err, ErrTooManyExtlangs) {
			t.Errorf("ParseAndNormalize(%q) error = %v, want ErrTooManyExtlangs", tag, err)
		}
	})
}

// TestParserTriggersErrDuplicateVariant verifies that ErrDuplicateVariant is returned
// during validation when the same variant subtag is repeated.
func TestParserTriggersErrDuplicateVariant(t *testing.T) {
	p := &Parser{
		registry: &Registry{
			Records: map[string]Record{
				"language:sl":   {Type: "language", Subtag: "sl"},
				"variant:nedis": {Type: "variant", Subtag: "nedis"},
			},
		},
	}

	tag := "sl-nedis-nedis"
	_, err := p.ParseAndNormalize(tag)
	if !errors.Is(err, ErrDuplicateVariant) {
		t.Errorf("ParseAndNormalize(%q) error = %v, want ErrDuplicateVariant", tag, err)
	}
}

// TestParserTriggersErrDuplicateSingleton verifies that ErrDuplicateSingleton is returned
// during validation when the same extension singleton character appears more than once.
func TestParserTriggersErrDuplicateSingleton(t *testing.T) {
	p := &Parser{
		registry: &Registry{
			Records: map[string]Record{
				"language:en": {Type: "language", Subtag: "en"},
			},
		},
	}

	tag := "en-a-val1-a-val2"
	_, err := p.ParseAndNormalize(tag)
	if !errors.Is(err, ErrDuplicateSingleton) {
		t.Errorf("ParseAndNormalize(%q) error = %v, want ErrDuplicateSingleton", tag, err)
	}
}

// TestParserTriggersErrInvalidSubtag verifies that ErrInvalidSubtag is returned
// when a subtag fails to match any valid category during validity checks.
func TestParserTriggersErrInvalidSubtag(t *testing.T) {
	p := &Parser{
		registry: &Registry{
			Records: map[string]Record{
				"language:en": {Type: "language", Subtag: "en"},
			},
		},
	}

	tag := "en-unknown"
	_, err := p.ParseAndNormalize(tag)
	if !errors.Is(err, ErrInvalidSubtag) {
		t.Errorf("ParseAndNormalize(%q) error = %v, want ErrInvalidSubtag", tag, err)
	}
}

// TestJSONUnmarshalPropagatesErrors verifies that LanguageTag.UnmarshalJSON properly
// propagates parsing errors defined in errors.go.
func TestJSONUnmarshalPropagatesErrors(t *testing.T) {
	testCases := []struct {
		name        string
		jsonInput   string
		expectedErr error
	}{
		{
			name:        "forbidden character in JSON",
			jsonInput:   `"en_US"`,
			expectedErr: ErrForbiddenChar,
		},
		{
			name:        "empty subtag in JSON",
			jsonInput:   `"en--US"`,
			expectedErr: ErrEmptySubtag,
		},
		{
			name:        "subtag too long in JSON",
			jsonInput:   `"toolonglanguage"`,
			expectedErr: ErrSubtagTooLong,
		},
		{
			name:        "empty private use in JSON",
			jsonInput:   `"x"`,
			expectedErr: ErrEmptyPrivateUse,
		},
		{
			name:        "empty extension in JSON",
			jsonInput:   `"en-a"`,
			expectedErr: ErrEmptyExtension,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var lt LanguageTag
			err := json.Unmarshal([]byte(tc.jsonInput), &lt)
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("json.Unmarshal(%s) error = %v, want %v", tc.jsonInput, err, tc.expectedErr)
			}
		})
	}
}
