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

// Package langtag provides a comprehensive implementation for parsing, validating,
// and manipulating IETF BCP 47 language tags, as specified in RFC 5646.
//
// The package offers a robust and efficient solution for applications requiring
// strict conformance to language tag standards. It includes the full IANA
// Language Subtag Registry, embedded at compile time, ensuring that the module
// works out of the box with no additional setup.
//
// # Key Features
//
//   - Strict Validation: Performs full syntactic and semantic validation against
//     the IANA registry, including checks for deprecated or invalid subtags.
//   - Canonicalization: Normalizes language tags to their canonical form as
//     per RFC 5646, simplifying comparisons and storage. It also supports
//     converting canonical tags to the alternative "extlang form" for
//     compatibility purposes.
//   - High Performance: The primary entry point, NewParser(), returns a reusable,
//     thread-safe parser instance that is initialized only once.
//   - Full Component Access: Provides methods to easily access all parts of a
//     tag, including language, script, region, variants, extensions, and
//     private-use subtags.
//   - Self-Contained: The required IANA registry data is embedded directly into
//     the library, so it has no external file dependencies at runtime.
package langtag

import (
	"encoding/json"
	"errors"
	"strings"
)

// Errors that can occur during language tag parsing.
var (
	ErrEmptyExtension     = errors.New("if an extension subtag is present, it must not be empty")
	ErrEmptyPrivateUse    = errors.New("if the 'x' subtag is present, it must not be empty")
	ErrForbiddenChar      = errors.New("the langtag contains a char not allowed")
	ErrInvalidSubtag      = errors.New("a subtag fails to parse or is not a valid IANA subtag")
	ErrInvalidLanguage    = errors.New("the given language subtag is invalid")
	ErrSubtagTooLong      = errors.New("a subtag may be eight characters in length at maximum")
	ErrEmptySubtag        = errors.New("a subtag should not be empty")
	ErrTooManyExtlangs    = errors.New("at maximum one extlang is allowed")
	ErrDuplicateVariant   = errors.New("the same variant subtag appears more than once")
	ErrDuplicateSingleton = errors.New("the same extension singleton appears more than once")
)

const typeExtlang = "extlang"

// Parser is a reusable BCP 47 parser. It contains the parsed IANA registry
// and should be created once and reused for efficiency.
type Parser struct {
	registry *Registry
}

// LanguageTag represents a well-formed RFC 5646 language tag.
type LanguageTag struct {
	tag        string
	positions  tagElementsPositions
	extensions []Extension
}

// Parse checks if a tag is "well-formed" according to RFC 5646 syntax.
// It parses the tag into its components but does not validate individual
// language, script, region, or variant subtags against the IANA registry.
//
// Because grandfathered tags (e.g., "i-klingon") are part of the ABNF syntax
// and cannot be parsed compositionally, this method will identify them as
// single, un-decomposed units.
//
// This method does not perform full canonicalization (such as replacing
// deprecated subtags). It does, however, normalize the case of the subtags
// for consistent output. For full validation and normalization, use
// ParseAndNormalize.
func (p *Parser) Parse(tag string) (LanguageTag, error) {
	for _, r := range tag {
		// As per RFC 5646 Sec 2.1, only US-ASCII alphanumeric chars and hyphens are allowed.
		if !isLangtagChar(r) {
			return LanguageTag{}, ErrForbiddenChar
		}
	}

	isGrandfathered := false
	lowerInput := strings.ToLower(tag)
	if record, ok := p.registry.Records[lowerInput]; ok && record.IsGrandfathered() {
		isGrandfathered = true
	}

	cpr := p.newCanonicalParseRun(tag, false)
	err := cpr.parse()
	if err != nil {
		return LanguageTag{}, err
	}

	var builder strings.Builder
	builder.Grow(len(tag))
	cpr.render(&builder)
	renderedTag := builder.String()

	positions := cpr.getPositions()
	positions.isGrandfathered = isGrandfathered

	return LanguageTag{tag: renderedTag, positions: positions, extensions: cpr.extensions}, nil
}

// ParseAndNormalize checks if a tag is "well-formed" and "valid", and then
// canonicalizes it according to RFC 5646 section 4.5. Canonicalization includes
// replacing deprecated tags/subtags, sorting extensions, and normalizing case.
func (p *Parser) ParseAndNormalize(tag string) (LanguageTag, error) {
	lowerInput := strings.ToLower(tag)
	isGrandfathered := false
	checkValidity := true

	if record, ok := p.registry.Records[lowerInput]; ok && record.IsGrandfathered() {
		if record.PreferredValue != "" {
			tag = record.PreferredValue
		} else if record.Type == "grandfathered" {
			isGrandfathered = true
			checkValidity = false
		}
	}

	cpr := p.newCanonicalParseRun(tag, checkValidity)
	err := cpr.parse()
	if err != nil {
		return LanguageTag{}, err
	}
	cpr.canonicalize()

	var builder strings.Builder
	builder.Grow(len(tag))
	cpr.render(&builder)
	canonicalTag := builder.String()

	cprFinal := p.newCanonicalParseRun(canonicalTag, false)
	err = cprFinal.parse()
	if err != nil {
		return LanguageTag{}, err
	}

	positions := cprFinal.getPositions()
	positions.isGrandfathered = isGrandfathered

	return LanguageTag{tag: canonicalTag, positions: positions, extensions: cprFinal.extensions}, nil
}

// ToExtlangForm converts a canonical language tag into its "extlang form"
// as described in RFC 5646, Section 4.5. If the tag's primary language
// subtag has a corresponding 'extlang' record in the IANA registry, this
// method prepends the extlang's prefix to the tag.
//
// For example, the canonical tag "hak-CN" (Hakka, China) would be converted
// to "zh-hak-CN".
//
// If the tag cannot be converted to an extlang form (i.e., its primary
// language is not an extlang), the original, unmodified LanguageTag is returned.
// The method expects a canonical LanguageTag as input, such as one returned
// by ParseAndNormalize.
func (p *Parser) ToExtlangForm(lt LanguageTag) (LanguageTag, error) {
	primaryLang := lt.PrimaryLanguage()
	if primaryLang == "" || lt.IsGrandfathered() {
		return lt, nil
	}

	lowerPrimaryLang := strings.ToLower(primaryLang)
	key := typeExtlang + ":" + lowerPrimaryLang
	rec, ok := p.registry.Records[key]
	if !ok || rec.Type != typeExtlang || len(rec.Prefix) == 0 {
		return lt, nil
	}

	prefix := rec.Prefix[0]
	newTagStr := prefix + "-" + lt.String()

	cpr := p.newCanonicalParseRun(newTagStr, false)
	err := cpr.parse()
	if err != nil {
		return LanguageTag{}, err
	}

	var builder strings.Builder
	builder.Grow(len(newTagStr))
	cpr.render(&builder)
	finalTagStr := builder.String()

	positions := cpr.getPositions()
	positions.isGrandfathered = false

	return LanguageTag{
		tag:        finalTagStr,
		positions:  positions,
		extensions: cpr.extensions,
	}, nil
}

// String returns the underlying language tag string. It implements the fmt.Stringer interface.
func (lt *LanguageTag) String() string {
	return lt.tag
}

// AsStr returns the underlying language tag representation.
func (lt *LanguageTag) AsStr() string {
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

// Extension represents a single extension in a language tag, e.g., `-u-co-phonebk`.
type Extension struct {
	Singleton rune
	Value     string
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
