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

// tagElementsPositions represents the subtag boundary indices as defined in the governing specification.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Representation:
// An internal structural tracking object holding integer index offsets that define the boundaries of each parsed
// component within the consolidated language tag string.
type tagElementsPositions struct {
	languageEnd, extlangEnd, scriptEnd, regionEnd, variantEnd, extensionEnd int
	isGrandfathered                                                         bool
}

// Extension represents the single extension component as defined in the governing specification.
//
// Specification Reference:
// RFC 5646 (Section 2.2.6)
//
// Representation:
// A structure pairing a single-character extension singleton identifier (excluding 'x') with its associated value
// subtags.
type Extension struct {
	Singleton rune
	Value     string
}

// LanguageTag represents the parsed and validated language tag as defined in the governing specification.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Representation:
// A logical abstraction of a well-formed language tag, encapsulating the underlying canonicalized string, structural
// boundaries of components, and parsed extension subtags.
type LanguageTag struct {
	tag        string
	positions  tagElementsPositions
	extensions []Extension
}

// String returns the underlying language tag string.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//
// Returns:
//   - string: The fully serialized and normalized language tag string.
func (lt *LanguageTag) String() string {
	return lt.tag
}

// PrimaryLanguage extracts the primary language subtag.
//
// Specification Reference:
// RFC 5646 (Section 2.2.1)
//
// Parameters:
//
// Returns:
//   - string: The primary language subtag, representing the first element of the language tag.
func (lt *LanguageTag) PrimaryLanguage() string {
	return lt.tag[:lt.positions.languageEnd]
}

// ExtendedLanguage extracts the extended language subtags as a single string.
//
// Specification Reference:
// RFC 5646 (Section 2.2.2)
//
// Parameters:
//
// Returns:
//   - string: The extended language subtags joined by hyphens, or an empty string if not present.
//   - bool: True if extended language subtags exist, false otherwise.
func (lt *LanguageTag) ExtendedLanguage() (string, bool) {
	if lt.positions.languageEnd == lt.positions.extlangEnd {
		return "", false
	}
	return lt.tag[lt.positions.languageEnd+1 : lt.positions.extlangEnd], true
}

// ExtendedLanguageSubtags extracts the individual extended language subtags.
//
// Specification Reference:
// RFC 5646 (Section 2.2.2)
//
// Parameters:
//
// Returns:
//   - []string: A slice containing each parsed extended language subtag, or nil if none exist.
func (lt *LanguageTag) ExtendedLanguageSubtags() []string {
	ext, ok := lt.ExtendedLanguage()
	if !ok {
		return nil
	}
	return strings.Split(ext, "-")
}

// FullLanguage extracts the full language portion, combining primary and extended language subtags.
//
// Specification Reference:
// RFC 5646 (Section 2.2.1) and RFC 5646 (Section 2.2.2)
//
// Parameters:
//
// Returns:
//   - string: The combined primary and extended language subtags.
func (lt *LanguageTag) FullLanguage() string {
	return lt.tag[:lt.positions.extlangEnd]
}

// Script extracts the script subtag.
//
// Specification Reference:
// RFC 5646 (Section 2.2.3)
//
// Parameters:
//
// Returns:
//   - string: The script subtag if present, or an empty string.
//   - bool: True if a script subtag exists, false otherwise.
func (lt *LanguageTag) Script() (string, bool) {
	if lt.positions.extlangEnd == lt.positions.scriptEnd {
		return "", false
	}
	return lt.tag[lt.positions.extlangEnd+1 : lt.positions.scriptEnd], true
}

// Region extracts the region subtag.
//
// Specification Reference:
// RFC 5646 (Section 2.2.4)
//
// Parameters:
//
// Returns:
//   - string: The region subtag if present, or an empty string.
//   - bool: True if a region subtag exists, false otherwise.
func (lt *LanguageTag) Region() (string, bool) {
	if lt.positions.scriptEnd == lt.positions.regionEnd {
		return "", false
	}
	return lt.tag[lt.positions.scriptEnd+1 : lt.positions.regionEnd], true
}

// Variant extracts the variant subtags as a single string.
//
// Specification Reference:
// RFC 5646 (Section 2.2.5)
//
// Parameters:
//
// Returns:
//   - string: The variant subtags joined by hyphens, or an empty string.
//   - bool: True if variant subtags exist, false otherwise.
func (lt *LanguageTag) Variant() (string, bool) {
	if lt.positions.regionEnd == lt.positions.variantEnd {
		return "", false
	}
	return lt.tag[lt.positions.regionEnd+1 : lt.positions.variantEnd], true
}

// VariantSubtags extracts the individual variant subtags.
//
// Specification Reference:
// RFC 5646 (Section 2.2.5)
//
// Parameters:
//
// Returns:
//   - []string: A slice containing each parsed variant subtag, or nil if none exist.
func (lt *LanguageTag) VariantSubtags() []string {
	v, ok := lt.Variant()
	if !ok {
		return nil
	}
	return strings.Split(v, "-")
}

// ExtensionSubtags extracts the parsed extension subtags.
//
// Specification Reference:
// RFC 5646 (Section 2.2.6)
//
// Parameters:
//
// Returns:
//   - []Extension: A copy of the parsed extension subtags, or nil if none exist.
func (lt *LanguageTag) ExtensionSubtags() []Extension {
	if len(lt.extensions) == 0 {
		return nil
	}
	exts := make([]Extension, len(lt.extensions))
	copy(exts, lt.extensions)
	return exts
}

// PrivateUse extracts the private use subtags as a single string.
//
// Specification Reference:
// RFC 5646 (Section 2.2.7)
//
// Parameters:
//
// Returns:
//   - string: The private use subtags joined by hyphens, or an empty string.
//   - bool: True if private use subtags exist, false otherwise.
func (lt *LanguageTag) PrivateUse() (string, bool) {
	// Spec Rule: RFC 5646 (Section 2.2.7)
	// Private use subtags are separated from other subtags by the reserved single-character subtag 'x' and can occupy
	// the primary position.
	if strings.HasPrefix(lt.tag, "x-") || strings.HasPrefix(lt.tag, "X-") {
		return lt.tag[2:], true
	}
	privateUseStart := lt.positions.extensionEnd
	if privateUseStart < len(lt.tag) &&
		(lt.tag[privateUseStart] == '-' && (lt.tag[privateUseStart+1] == 'x' || lt.tag[privateUseStart+1] == 'X')) {
		// Implementation Note: Skip the singleton prefix "x"
		// The private use value starts after the "-x-" part. The variable privateUseStart points to the first '-', so
		// we slice from +3 to skip over '-x-'.
		return lt.tag[privateUseStart+3:], true
	}
	return "", false
}

// PrivateUseSubtags extracts the individual private use subtags.
//
// Specification Reference:
// RFC 5646 (Section 2.2.7)
//
// Parameters:
//
// Returns:
//   - []string: A slice containing each parsed private use subtag, or nil if none exist.
func (lt *LanguageTag) PrivateUseSubtags() []string {
	part, ok := lt.PrivateUse()
	if !ok {
		return nil
	}
	return strings.Split(part, "-")
}

// IsGrandfathered checks if the language tag is classified as grandfathered.
//
// Specification Reference:
// RFC 5646 (Section 2.2.8)
//
// Parameters:
//
// Returns:
//   - bool: True if the tag is a registered grandfathered tag, false otherwise.
func (lt *LanguageTag) IsGrandfathered() bool {
	return lt.positions.isGrandfathered
}

// MarshalJSON serializes the language tag into its JSON string representation.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//
// Returns:
//   - []byte: The JSON-encoded string representation of the language tag.
//   - error: An error if JSON marshaling fails.
func (lt *LanguageTag) MarshalJSON() ([]byte, error) {
	return json.Marshal(lt.tag)
}

// UnmarshalJSON deserializes and validates a language tag from a JSON string.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - data: A JSON-encoded byte slice representing the language tag.
//
// Returns:
//   - error: An error if JSON unmarshaling or syntax validation fails, otherwise nil.
func (lt *LanguageTag) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" {
		*lt = LanguageTag{}
		return nil
	}

	// Implementation Note: Lightweight Syntax-Only Validation
	// To avoid parsing the 1MB embedded registry file on every JSON unmarshal call, and in strict
	// adherence to the codebase constraint forbidding global variables, syntax-only well-formedness
	// validation is executed using an empty local registry map. This validates characters, subtag
	// lengths, and structural properties without high loading overheads. Full semantic validation
	// and normalization can be manually applied post-unmarshaling using a persistent Parser instance.
	p := &Parser{
		registry: &Registry{
			Records: make(map[string]Record),
		},
	}

	parsed, err := p.Parse(s)
	if err != nil {
		return err
	}
	*lt = parsed
	return nil
}
