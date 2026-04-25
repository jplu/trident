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

package langtag

import (
	"encoding/json"
	"strings"
)

// tagElementsPositions stores the calculated end positions of each major
// component within the final language tag string.
type tagElementsPositions struct {
	languageEnd, extlangEnd, scriptEnd, regionEnd, variantEnd, extensionEnd int
	isGrandfathered                                                         bool
}

// Extension represents a single extension in a language tag, e.g., `-u-co-phonebk`.
type Extension struct {
	Singleton rune
	Value     string
}

// LanguageTag represents a well-formed RFC 5646 language tag.
type LanguageTag struct {
	tag        string
	positions  tagElementsPositions
	extensions []Extension
}

// String returns the underlying language tag string. It implements the fmt.Stringer interface.
func (lt *LanguageTag) String() string {
	return lt.tag
}

// PrimaryLanguage returns the primary language subtag.
func (lt *LanguageTag) PrimaryLanguage() string {
	return lt.tag[:lt.positions.languageEnd]
}

// ExtendedLanguage returns the extended language subtags as a single string.
func (lt *LanguageTag) ExtendedLanguage() (string, bool) {
	if lt.positions.languageEnd == lt.positions.extlangEnd {
		return "", false
	}
	return lt.tag[lt.positions.languageEnd+1 : lt.positions.extlangEnd], true
}

// ExtendedLanguageSubtags returns a slice of extended language subtags.
func (lt *LanguageTag) ExtendedLanguageSubtags() []string {
	ext, ok := lt.ExtendedLanguage()
	if !ok {
		return nil
	}
	return strings.Split(ext, "-")
}

// FullLanguage returns the primary language subtag and its extended language subtags.
func (lt *LanguageTag) FullLanguage() string {
	return lt.tag[:lt.positions.extlangEnd]
}

// Script returns the script subtag.
func (lt *LanguageTag) Script() (string, bool) {
	if lt.positions.extlangEnd == lt.positions.scriptEnd {
		return "", false
	}
	return lt.tag[lt.positions.extlangEnd+1 : lt.positions.scriptEnd], true
}

// Region returns the region subtag.
func (lt *LanguageTag) Region() (string, bool) {
	if lt.positions.scriptEnd == lt.positions.regionEnd {
		return "", false
	}
	return lt.tag[lt.positions.scriptEnd+1 : lt.positions.regionEnd], true
}

// Variant returns the variant subtags as a single string.
func (lt *LanguageTag) Variant() (string, bool) {
	if lt.positions.regionEnd == lt.positions.variantEnd {
		return "", false
	}
	return lt.tag[lt.positions.regionEnd+1 : lt.positions.variantEnd], true
}

// VariantSubtags returns a slice of variant subtags.
func (lt *LanguageTag) VariantSubtags() []string {
	v, ok := lt.Variant()
	if !ok {
		return nil
	}
	return strings.Split(v, "-")
}

// ExtensionSubtags returns a slice of parsed extensions.
func (lt *LanguageTag) ExtensionSubtags() []Extension {
	if len(lt.extensions) == 0 {
		return nil
	}
	exts := make([]Extension, len(lt.extensions))
	copy(exts, lt.extensions)
	return exts
}

// PrivateUse returns the private use subtags as a single string (e.g., `phonebk-sort`).
func (lt *LanguageTag) PrivateUse() (string, bool) {
	if strings.HasPrefix(lt.tag, "x-") || strings.HasPrefix(lt.tag, "X-") {
		return lt.tag[2:], true
	}
	privateUseStart := lt.positions.extensionEnd
	if privateUseStart < len(lt.tag) &&
		(lt.tag[privateUseStart] == '-' && (lt.tag[privateUseStart+1] == 'x' || lt.tag[privateUseStart+1] == 'X')) {
		// The private use value starts after the "-x-" part.
		// privateUseStart points to the first '-', so we need to slice from +3
		// to skip over '-x-'. The parser ensures a subtag follows.
		return lt.tag[privateUseStart+3:], true
	}
	return "", false
}

// PrivateUseSubtags returns a slice of private use subtags.
func (lt *LanguageTag) PrivateUseSubtags() []string {
	part, ok := lt.PrivateUse()
	if !ok {
		return nil
	}
	return strings.Split(part, "-")
}

// IsGrandfathered returns true if the tag is a grandfathered tag.
func (lt *LanguageTag) IsGrandfathered() bool {
	return lt.positions.isGrandfathered
}

// MarshalJSON implements the json.Marshaler interface. It marshals the language
// tag as a JSON string.
func (lt *LanguageTag) MarshalJSON() ([]byte, error) {
	return json.Marshal(lt.tag)
}

// UnmarshalJSON implements the json.Unmarshaler interface. It performs a full
// validity check on the tag from the JSON string.
//
// Performance Warning: This method creates a new parser by calling NewParser()
// on every invocation, which is an expensive operation. For performance-critical
// applications, it is highly recommended to unmarshal into a string and then use a
// pre-initialized, long-lived parser instance to parse the tag.
func (lt *LanguageTag) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" {
		*lt = LanguageTag{}
		return nil
	}

	p, err := NewParser()
	if err != nil {
		return err
	}

	parsed, err := p.ParseAndNormalize(s)
	if err != nil {
		return err
	}
	*lt = parsed
	return nil
}
