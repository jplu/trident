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
	"bytes"
	"errors"
	"strings"
)

// Parser is a reusable BCP 47 parser. It contains the parsed IANA registry
// and should be created once and reused for efficiency.
type Parser struct {
	registry *Registry
}

// NewParser creates a new parser instance from the embedded IANA registry.
//
// IMPORTANT: This function parses the entire IANA registry on every call and is
// therefore an expensive operation. For performance, it is strongly recommended
// to call this function only once at application startup and reuse the returned
// parser instance throughout your application.
func NewParser() (*Parser, error) {
	if len(embeddedRegistryData) == 0 {
		return nil, errors.New("embedded language-subtag-registry file is empty or not found")
	}

	reader := bytes.NewReader(embeddedRegistryData)
	registry, err := ParseRegistry(reader)
	if err != nil {
		return nil, err
	}

	return &Parser{
		registry: registry,
	}, nil
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
