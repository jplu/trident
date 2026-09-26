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

// Parser represents the language tag parser as defined in the governing specification.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Representation:
// A stateful parser containing a pointer to the parsed IANA Language Subtag Registry
// database used to validate and normalize language tags.
type Parser struct {
	registry *Registry
}

// NewParser creates a new parser instance from the embedded IANA registry.
//
// Specification Reference:
// RFC 5646 (Section 3)
//
// Parameters:
//
// Returns:
//   - *Parser: A pointer to the initialized Parser instance containing the parsed registry.
//   - error: An error if the embedded registry data is empty or if parsing the registry fails.
func NewParser() (*Parser, error) {
	// Implementation Note: Reusable Parser design
	// This function parses the entire IANA registry on every call, which is an expensive
	// operation. For performance, it is recommended to call this function once at startup
	// and reuse the returned parser instance.
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

// Parse checks if a language tag is well-formed according to RFC 5646 syntax.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - tag: A string representing the language tag to be checked for well-formedness.
//
// Returns:
//   - LanguageTag: The parsed LanguageTag structure containing components and position indices.
//   - error: An error if the tag contains forbidden characters or violates the structural syntax of the specification.
func (p *Parser) Parse(tag string) (LanguageTag, error) {
	// Step 1: Validate character set
	// Check that every character in the tag is a valid US-ASCII alphanumeric character or hyphen.
	for _, r := range tag {
		// Spec Rule: RFC 5646 (Section 2.1)
		// Language tags are sequences of characters from the US-ASCII repertoire consisting of letters, digits, and
		// hyphens.
		if !isLangtagChar(r) {
			return LanguageTag{}, ErrForbiddenChar
		}
	}

	// Step 2: Check for grandfathered tag
	// Determine if the input matches any grandfathered or redundant record in the static list or IANA registry.
	isGrandfathered := false
	lowerInput := strings.ToLower(tag)
	if isGrandfatheredTag(lowerInput) {
		isGrandfathered = true
	} else if record, ok := p.registry.Records[lowerInput]; ok && record.IsGrandfathered() {
		isGrandfathered = true
	}

	// Step 3: Run parsing state machine
	// Parse individual subtags structurally without checking their presence in the registry.
	cpr := p.newCanonicalParseRun(tag, false)
	err := cpr.parse()
	if err != nil {
		return LanguageTag{}, err
	}

	// Step 4: Construct LanguageTag representation
	// Render the parsed tag to a normalized casing and calculate component boundary positions.
	var builder strings.Builder
	builder.Grow(len(tag))
	cpr.render(&builder)
	renderedTag := builder.String()

	positions := cpr.getPositions()
	positions.isGrandfathered = isGrandfathered

	return LanguageTag{tag: renderedTag, positions: positions, extensions: cpr.extensions}, nil
}

// ParseAndNormalize parses, validates, and normalizes a language tag to its canonical form.
//
// Specification Reference:
// RFC 5646 (Section 2.2.9 and Section 4.5)
//
// Parameters:
//   - tag: A string representing the language tag to be parsed, validated, and normalized.
//
// Returns:
//   - LanguageTag: The fully canonicalized LanguageTag structure.
//   - error: An error if the tag is not well-formed, contains invalid subtags, or fails syntax and registry checks.
func (p *Parser) ParseAndNormalize(tag string) (LanguageTag, error) {
	// Step 1: Pre-process grandfathered tags
	// Check if the tag matches a grandfathered record and apply preferred replacement value if defined.
	lowerInput := strings.ToLower(tag)
	isGrandfathered := false
	checkValidity := true

	// Spec Rule: RFC 5646 (Section 2.2.8)
	// Grandfathered tags must be identified as they cannot be parsed compositionally, with preferred values applied if
	// available.
	if record, ok := p.registry.Records[lowerInput]; ok && record.IsGrandfathered() {
		if record.PreferredValue != "" {
			tag = record.PreferredValue
		} else if record.Type == "grandfathered" {
			isGrandfathered = true
			checkValidity = false
		}
	} else if isGrandfatheredTag(lowerInput) {
		isGrandfathered = true
		checkValidity = false
	}

	// Step 2: Validate and parse subtags
	// Perform structural parsing and validate subtags against the registry if checkValidity is true.
	cpr := p.newCanonicalParseRun(tag, checkValidity)
	err := cpr.parse()
	if err != nil {
		return LanguageTag{}, err
	}

	// Step 3: Apply canonicalization rules
	// Sort extensions, resolve deprecated subtags, and normalize casing as specified in the standard.
	cpr.canonicalize()

	var builder strings.Builder
	builder.Grow(len(tag))
	cpr.render(&builder)
	canonicalTag := builder.String()

	// Step 4: Re-parse the canonical tag
	// Run a final parse pass on the canonical tag string to update component end positions.
	cprFinal := p.newCanonicalParseRun(canonicalTag, false)
	err = cprFinal.parse()
	if err != nil {
		return LanguageTag{}, err
	}

	positions := cprFinal.getPositions()
	positions.isGrandfathered = isGrandfathered

	return LanguageTag{
		tag:        canonicalTag,
		positions:  positions,
		extensions: cprFinal.extensions,
	}, nil
}

// ParseWellFormed parses and validates a language tag for syntactic well-formedness according to the structural rules.
//
// Specification Reference:
// RFC 5646 (Section 2.1 and Section 2.2.9)
//
// Parameters:
//   - tag: A string representing the language tag to check for well-formedness.
//
// Returns:
//   - LanguageTag: The parsed LanguageTag structure containing components and position indices.
//   - error: An error if the tag contains forbidden characters or violates the structural syntax of the specification.
func ParseWellFormed(tag string) (LanguageTag, error) {
	// Step 1: Validate character set
	// Check that every character in the tag is a valid US-ASCII alphanumeric character or hyphen.
	for _, r := range tag {
		// Spec Rule: RFC 5646 (Section 2.1)
		// Language tags are sequences of characters from the US-ASCII repertoire consisting of letters, digits, and
		// hyphens.
		if !isLangtagChar(r) {
			return LanguageTag{}, ErrForbiddenChar
		}
	}

	// Step 2: Run syntax-only parsing state machine
	// Parse individual subtags structurally without checking their presence in the registry.
	cpr := &canonicalParseRun{
		subtags:       strings.Split(tag, "-"),
		checkValidity: false,
	}
	if err := cpr.parse(); err != nil {
		return LanguageTag{}, err
	}

	// Step 3: Construct LanguageTag representation
	// Render the parsed tag to a normalized casing and calculate component boundary positions.
	var builder strings.Builder
	builder.Grow(len(tag))
	cpr.render(&builder)
	renderedTag := builder.String()

	positions := cpr.getPositions()
	positions.isGrandfathered = isGrandfatheredTag(strings.ToLower(renderedTag))

	return LanguageTag{tag: renderedTag, positions: positions, extensions: cpr.extensions}, nil
}

// SuppressScript suppresses redundant script subtags within a language tag when they match the default script of the
// primary language.
//
// Specification Reference:
// RFC 5646 (Section 4.1)
//
// Parameters:
//   - lt: The parsed LanguageTag instance to evaluate for script suppression.
//
// Returns:
//   - LanguageTag: A new LanguageTag with the suppressed script removed if applicable, otherwise the original tag.
func (p *Parser) SuppressScript(lt LanguageTag) LanguageTag {
	// Step 1: Validate script presence
	// Verify if the language tag contains a script subtag before attempting suppression.
	if lt.IsGrandfathered() {
		return lt
	}
	script, ok := lt.Script()
	if !ok {
		return lt
	}

	// Step 2: Retrieve default script from registry
	// Look up the primary language in the registry to check for its Suppress-Script value.
	primaryLang := lt.PrimaryLanguage()
	key := "language:" + strings.ToLower(primaryLang)
	langRec, okRec := p.registry.Records[key]

	// Step 3: Suppress redundant script subtag
	// Evaluate and remove the redundant script subtag if it matches the registered default.
	//
	// Spec Rule: RFC 5646 (Section 4.1)
	// The script subtag SHOULD NOT be used to form language tags unless the script adds some distinguishing
	// information.
	if okRec && langRec.SuppressScript != "" && strings.EqualFold(script, langRec.SuppressScript) {
		cpr := p.newCanonicalParseRun(lt.String(), false)

		// Implementation Note: Safe error omission
		// Since the input LanguageTag is guaranteed to be syntactically well-formed, cpr.parse() will always succeed.
		_ = cpr.parse()
		cpr.script = ""

		var builder strings.Builder
		builder.Grow(len(lt.String()))
		cpr.render(&builder)
		canonicalTag := builder.String()

		cprFinal := p.newCanonicalParseRun(canonicalTag, false)

		// Implementation Note: Safe error omission
		// Rendering the tag after removing a valid script subtag always produces a well-formed BCP 47 string.
		_ = cprFinal.parse()

		positions := cprFinal.getPositions()
		positions.isGrandfathered = false

		return LanguageTag{
			tag:        canonicalTag,
			positions:  positions,
			extensions: cprFinal.extensions,
		}
	}

	return lt
}

// ToExtlangForm converts a canonical language tag into its extlang form.
//
// Specification Reference:
// RFC 5646 (Section 4.5)
//
// Parameters:
//   - lt: The LanguageTag instance in canonical form to be converted.
//
// Returns:
//   - LanguageTag: The converted LanguageTag in extlang form, or the original tag if no extlang mapping is applicable.
//   - error: An error if parsing the newly formed extlang tag string fails.
func (p *Parser) ToExtlangForm(lt LanguageTag) (LanguageTag, error) {
	// Step 1: Identify extlang prefix mapping
	// Look up the primary language subtag in the registry to find if it has an extlang record with a valid prefix.
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

	// Step 2: Form extlang tag string
	// Prepend the found prefix to the existing tag string.
	prefix := rec.Prefix[0]
	newTagStr := prefix + "-" + lt.String()

	// Step 3: Parse and render extlang form
	// Execute a parse run on the new tag string to correctly calculate updated subtag positions.
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
