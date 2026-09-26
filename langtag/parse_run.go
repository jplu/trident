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
	"strings"
	"unicode"
)

// BCP 47 constants for subtag validation.
const (
	maxSubtagLen          = 8 // Maximum length of any subtag.
	maxExtlangs           = 1 // Maximum number of extended language subtags allowed by the specification in validity checks.
	maxExtlangsWellFormed = 3 // Maximum number of extended language subtags allowed by the structural ABNF grammar in well-formedness checks.
	scriptLen             = 4 // A script subtag is always four letters.
	regionAlphaLen        = 2 // An alphabetic region subtag is always two letters.
	regionNumericLen      = 3 // A numeric region subtag is always three digits.
	extlangLen            = 3 // An extended language subtag is always three letters.
	shortPrimaryLangLen   = 3 // Maximum length of a primary language that can be followed by an extended language subtag.
	minVariantLenAlpha    = 5 // Minimum length of a variant starting with a letter.
	minVariantLenDigit    = 4 // Minimum length of a variant starting with a digit.
)

// parseState represents the parsing state of the language tag parser as defined in the governing specification.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Representation:
// An integer-based enumeration representing distinct stages of subtag validation (language, extlang, script, region,
// variant, extension, and private-use).
type parseState int

const (
	stateStart         parseState = iota // Expecting a primary language subtag.
	stateAfterLanguage                   // After a 2-3 letter primary language, expecting extlang, script, etc.
	stateAfterExtLang                    // After a primary language longer than 3 letters or an extlang, expecting script, region, etc.
	stateAfterScript                     // After a script, expecting region, variant, etc.
	stateAfterRegion                     // After a region, expecting variant, etc.
	stateInVariant                       // In a sequence of one or more variants.
	stateInExtension                     // In an extension sequence (after a singleton).
	stateInPrivateUse                    // In a private-use sequence (after 'x').
)

// canonicalParseRun represents the transient state of a single parsing run as defined in the governing specification.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Representation:
// A stateful tracking structure used to parse, validate, and temporarily house components of a BCP 47 language tag
// before final rendering.
type canonicalParseRun struct {
	parent            *Parser
	language          string
	extlangs          []string
	script            string
	region            string
	variants          []string
	extensions        []Extension
	privateuse        []string
	subtags           []string
	state             parseState
	checkValidity     bool
	seenVariants      map[string]struct{}
	seenSingletons    map[rune]struct{}
	extlangsCount     int
	extensionExpected bool
}

// newCanonicalParseRun initializes and returns a new transient parsing run.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - input: The raw language tag input string.
//   - checkValidity: A boolean flag indicating whether subtags must be validated against the registry.
//
// Returns:
//   - *canonicalParseRun: A pointer to the initialized transient parsing run context.
func (p *Parser) newCanonicalParseRun(input string, checkValidity bool) *canonicalParseRun {
	return &canonicalParseRun{
		parent:        p,
		subtags:       strings.Split(input, "-"),
		checkValidity: checkValidity,
	}
}

// validateSubtag performs basic syntactic length and nullity checks on a single subtag.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - subtag: The individual subtag string to evaluate.
//
// Returns:
//   - error: ErrEmptySubtag if the string is empty, ErrSubtagTooLong if it exceeds eight characters, otherwise nil.
func validateSubtag(subtag string) error {
	if len(subtag) == 0 {
		return ErrEmptySubtag
	}
	// Spec Rule: RFC 5646 (Section 2.1)
	// All subtags have a maximum length of eight characters.
	if len(subtag) > maxSubtagLen {
		return ErrSubtagTooLong
	}
	return nil
}

// prepareSubtags sanitizes the input subtags by identifying and trimming a trailing hyphen.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//
// Returns:
//   - []string: The slice of prepared and sanitized subtags.
//   - bool: True if an empty trailing subtag indicating a trailing hyphen was removed, false otherwise.
func (cpr *canonicalParseRun) prepareSubtags() ([]string, bool) {
	hasTrailingHyphen := len(cpr.subtags) > 1 && cpr.subtags[len(cpr.subtags)-1] == ""
	if hasTrailingHyphen {
		return cpr.subtags[:len(cpr.subtags)-1], true
	}
	return cpr.subtags, false
}

// isTagGrandfathered evaluates if the complete raw tag sequence matches a registered grandfathered tag record.
//
// Specification Reference:
// RFC 5646 (Section 2.2.8)
//
// Parameters:
//
// Returns:
// - bool: True if the complete raw tag is found in the grandfathered registry or hardcoded static list, false
// otherwise.
func (cpr *canonicalParseRun) isTagGrandfathered() bool {
	subtags := cpr.subtags
	if len(subtags) > 1 && subtags[len(subtags)-1] == "" {
		subtags = subtags[:len(subtags)-1]
	}
	fullTag := strings.ToLower(strings.Join(subtags, "-"))

	// Step 1: Static Fallback Check
	// Check against the static list of BCP 47 grandfathered tags to support syntax-only parses with empty registries.
	if isGrandfatheredTag(fullTag) {
		return true
	}

	// Step 2: Registry Database Check
	// Check the loaded IANA database if a full parser context has been initialized.
	if cpr.parent != nil && cpr.parent.registry != nil {
		if record, ok := cpr.parent.registry.Records[fullTag]; ok && record.IsGrandfathered() {
			return true
		}
	}
	return false
}

// parsePrivateUseOnly parses tags that start with the private-use singleton "x".
//
// Specification Reference:
// RFC 5646 (Section 2.2.7)
//
// Parameters:
//   - subtags: The complete slice of subtags representing the private-use tag.
//
// Returns:
//   - error: ErrEmptyPrivateUse if no subtags follow the "x" singleton, otherwise any syntactic validation error.
func (cpr *canonicalParseRun) parsePrivateUseOnly(subtags []string) error {
	if len(subtags) == 1 {
		return ErrEmptyPrivateUse
	}
	// Step 1: Validate private-use subtags
	// Ensure each subtag in the private-use section conforms to BCP 47 syntactic constraints.
	for _, subtag := range subtags[1:] {
		if err := validateSubtag(subtag); err != nil {
			return err
		}
		cpr.privateuse = append(cpr.privateuse, subtag)
	}
	cpr.state = stateInPrivateUse
	return nil
}

// processSubtags processes a slice of subtag strings sequentially through the state machine.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - subtags: The sequence of subtag strings to process.
//
// Returns:
//   - error: An error under syntax violations or invalid state transitions, otherwise nil.
func (cpr *canonicalParseRun) processSubtags(subtags []string) error {
	for i, subtag := range subtags {
		if err := validateSubtag(subtag); err != nil {
			return err
		}

		switch cpr.state {
		case stateInPrivateUse:
			cpr.privateuse = append(cpr.privateuse, subtag)
		case stateInExtension:
			if err := cpr.handleExtensionSubtag(subtag); err != nil {
				return err
			}
		case stateStart,
			stateAfterLanguage,
			stateAfterExtLang,
			stateAfterScript,
			stateAfterRegion,
			stateInVariant:
			if err := cpr.handleLangtagSubtag(i, subtag); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkFinalState checks that the parser did not terminate in an incomplete state.
//
// Specification Reference:
// RFC 5646 (Section 2.1, Section 2.2.6, and Section 2.2.7)
//
// Parameters:
//   - hasTrailingHyphen: A boolean indicating whether the raw input ended with a trailing hyphen.
//
// Returns:
// - error: ErrEmptySubtag if a trailing hyphen was detected, ErrEmptyExtension if an expected extension subtag is
// missing, ErrEmptyPrivateUse if a private use block is empty, otherwise nil.
func (cpr *canonicalParseRun) checkFinalState(hasTrailingHyphen bool) error {
	// Spec Rule: RFC 5646 (Section 2.1)
	// Language tags must not end with a trailing hyphen.
	if hasTrailingHyphen {
		return ErrEmptySubtag
	}

	// Spec Rule: RFC 5646 (Section 2.2.6, Clause 6)
	// Each singleton MUST be followed by at least one extension subtag (it cannot be empty at the end of parsing).
	if cpr.extensionExpected {
		return ErrEmptyExtension
	}

	// Spec Rule: RFC 5646 (Section 2.2.7, Rule 3)
	// Private use subtags are introduced by the reserved single-character subtag 'x' and must not be empty.
	if cpr.state == stateInPrivateUse && len(cpr.privateuse) == 0 {
		return ErrEmptyPrivateUse
	}

	return nil
}

// parse executes the parsing state machine over the prepared subtags.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//
// Returns:
//   - error: Any syntactic, semantic, or state validation error encountered during the parse run, otherwise nil.
func (cpr *canonicalParseRun) parse() error {
	// Step 1: Prepare Subtags
	// Split and sanitize subtags by trimming trailing hyphens if present.
	subtagsToParse, hasTrailingHyphen := cpr.prepareSubtags()

	// Spec Rule: RFC 5646 (Section 2.2.7)
	// Private use subtags are introduced by the reserved single-character subtag 'x'.
	if len(subtagsToParse) > 0 && strings.EqualFold(subtagsToParse[0], "x") {
		if err := cpr.parsePrivateUseOnly(subtagsToParse); err != nil {
			return err
		}
		return cpr.checkFinalState(hasTrailingHyphen)
	}

	// Step 2: Process Subtags
	// Loop through and parse standard subtags using the state machine.
	if err := cpr.processSubtags(subtagsToParse); err != nil {
		return err
	}

	// Step 3: Validate Final Parser State
	// Check that we didn't end with unfinished extension singletons or empty private use blocks.
	return cpr.checkFinalState(hasTrailingHyphen)
}

// handlePrimaryLanguage parses and validates the first subtag as a primary language.
//
// Specification Reference:
// RFC 5646 (Section 2.2.1)
//
// Parameters:
//   - subtag: The primary language subtag candidate string.
//
// Returns:
//   - error: ErrInvalidLanguage if the subtag fails format or registry constraints, otherwise nil.
func (cpr *canonicalParseRun) handlePrimaryLanguage(subtag string) error {
	minLen := 2
	if cpr.isTagGrandfathered() {
		// Spec Rule: RFC 5646 (Section 2.2.8)
		// Grandfathered tags may use single-character language subtags like "i".
		minLen = 1
	}
	// Spec Rule: RFC 5646 (Section 2.2.1)
	// Primary language subtags are 2 to 8 characters long, or 1 character for grandfathered tags.
	if len(subtag) < minLen || len(subtag) > 8 || !isAlphabetic(subtag) {
		return ErrInvalidLanguage
	}

	if cpr.checkValidity {
		// Spec Rule: RFC 5646 (Section 2.2.9)
		// A tag is considered valid if its primary language subtag appears in the IANA Language Subtag Registry.
		//
		// Implementation Note: Case-insensitive registry lookups
		// Convert the lookup key to lowercase before performing map lookups since keys in the parsed registry map are
		// indexed in lowercase.
		lowerSubtag := strings.ToLower(subtag)
		key := "language:" + lowerSubtag
		rec, recordExists := cpr.parent.registry.Records[key]
		if !recordExists || rec.Type != "language" {
			return ErrInvalidLanguage
		}
	}
	cpr.language = subtag
	cpr.state = stateAfterExtLang
	if len(subtag) <= shortPrimaryLangLen {
		cpr.state = stateAfterLanguage
	}
	return nil
}

// checkForTooManyExtlangs evaluates if an extra extlang is present when prohibited by the specification.
//
// Specification Reference:
// RFC 5646 (Section 2.2.2)
//
// Parameters:
//   - subtag: The subtag currently being evaluated as a potential additional extlang.
//
// Returns:
//   - error: ErrTooManyExtlangs if an extra extlang is present, otherwise nil.
func (cpr *canonicalParseRun) checkForTooManyExtlangs(subtag string) error {
	// Spec Rule: RFC 5646 (Section 2.2.2)
	// At maximum, only one extlang is allowed in a valid tag. The ABNF structures up to 3 for well-formedness.
	limit := maxExtlangs
	if !cpr.checkValidity {
		limit = maxExtlangsWellFormed
	}

	if cpr.extlangsCount >= limit && len(subtag) == extlangLen && isAlphabetic(subtag) {
		if cpr.checkValidity {
			key := "extlang:" + strings.ToLower(subtag)
			if _, ok := cpr.parent.registry.Records[key]; ok {
				return ErrTooManyExtlangs
			}
		} else {
			return ErrTooManyExtlangs
		}
	}
	return nil
}

// handleLangtagSubtag delegates parsing for a subtag based on the current parser state.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - i: The index of the subtag within the split sequence.
//   - subtag: The string value of the subtag to handle.
//
// Returns:
//   - error: An error if validation fails or no matching subtag type is identified, otherwise nil.
func (cpr *canonicalParseRun) handleLangtagSubtag(i int, subtag string) error {
	if i == 0 {
		return cpr.handlePrimaryLanguage(subtag)
	}
	// Spec Rule: RFC 5646 (Section 2.1)
	// Single-character subtags (singletons) start extension or private-use sequences.
	if len(subtag) == 1 {
		return cpr.handleSingleton(subtag)
	}

	if err := cpr.checkForTooManyExtlangs(subtag); err != nil {
		return err
	}

	// Step 1: Sequential Subtag Matching
	// Attempt to parse the subtag in the order defined by the RFC: extlang -> script -> region -> variant.
	if cpr.tryParseAsExtlang(subtag) {
		// Implementation Note: Safe state transitions for multiple extlangs
		// We only transition the state to stateAfterExtLang when the maximum allowed
		// limit of parsed extlang subtags for the current validation mode is reached.
		limit := maxExtlangs
		if !cpr.checkValidity {
			limit = maxExtlangsWellFormed
		}
		if cpr.extlangsCount == limit {
			cpr.state = stateAfterExtLang
		}
		return nil
	}
	if cpr.tryParseAsScript(subtag) {
		cpr.state = stateAfterScript
		return nil
	}
	if cpr.tryParseAsRegion(subtag) {
		cpr.state = stateAfterRegion
		return nil
	}
	if parsed, err := cpr.tryParseAsVariant(subtag); parsed || err != nil {
		if parsed {
			cpr.state = stateInVariant
		}
		return err
	}

	return ErrInvalidSubtag
}

// tryParseAsExtlang attempts to parse a subtag as an extended language.
//
// Specification Reference:
// RFC 5646 (Section 2.2.2)
//
// Parameters:
//   - subtag: The subtag string to evaluate.
//
// Returns:
//   - bool: True if successfully identified and stored as an extlang, false otherwise.
func (cpr *canonicalParseRun) tryParseAsExtlang(subtag string) bool {
	// Spec Rule: RFC 5646 (Section 2.2.2)
	// Extended language subtags consist of three letters. Up to three are structurally allowed by the ABNF.
	limit := maxExtlangs
	if !cpr.checkValidity {
		limit = maxExtlangsWellFormed
	}

	if cpr.state != stateAfterLanguage || cpr.extlangsCount >= limit ||
		len(subtag) != extlangLen || !isAlphabetic(subtag) {
		return false
	}
	if cpr.checkValidity {
		// Implementation Note: Case-insensitive map key lookup
		// The IANA registry keys are stored in lowercase, so the subtag must be lowercased before lookup.
		lowerSubtag := strings.ToLower(subtag)
		key := "extlang:" + lowerSubtag
		rec, ok := cpr.parent.registry.Records[key]
		if !ok || rec.Type != typeExtlang {
			return false
		}
	}
	cpr.extlangsCount++
	cpr.extlangs = append(cpr.extlangs, subtag)
	return true
}

// tryParseAsScript attempts to parse a subtag as a script.
//
// Specification Reference:
// RFC 5646 (Section 2.2.3)
//
// Parameters:
//   - subtag: The subtag string to evaluate.
//
// Returns:
//   - bool: True if successfully identified and stored as a script, false otherwise.
func (cpr *canonicalParseRun) tryParseAsScript(subtag string) bool {
	// Spec Rule: RFC 5646 (Section 2.2.3)
	// Script subtags are four letters long.
	if cpr.state > stateAfterExtLang || len(subtag) != scriptLen || !isAlphabetic(subtag) {
		return false
	}
	if cpr.checkValidity {
		// Implementation Note: Registry key lookup
		// The registry script codes must be lowercased to match our normalized internal index keys.
		lowerSubtag := strings.ToLower(subtag)
		key := "script:" + lowerSubtag
		rec, ok := cpr.parent.registry.Records[key]
		if !ok || rec.Type != "script" {
			return false
		}
	}
	cpr.script = subtag
	return true
}

// tryParseAsRegion attempts to parse a subtag as a region.
//
// Specification Reference:
// RFC 5646 (Section 2.2.4)
//
// Parameters:
//   - subtag: The subtag string to evaluate.
//
// Returns:
//   - bool: True if successfully identified and stored as a region, false otherwise.
func (cpr *canonicalParseRun) tryParseAsRegion(subtag string) bool {
	// Spec Rule: RFC 5646 (Section 2.2.4)
	// Region subtags consist of either 2 alphabetic characters (ISO 3166-1) or 3 numeric digits (UN M.49).
	isRegionFmt := (len(subtag) == regionAlphaLen && isAlphabetic(subtag)) ||
		(len(subtag) == regionNumericLen && isNumeric(subtag))
	if cpr.state > stateAfterScript || !isRegionFmt {
		return false
	}
	if cpr.checkValidity {
		// Implementation Note: Region map key normalization
		// Ensure case-insensitive region registration checks by converting to lowercase.
		lowerSubtag := strings.ToLower(subtag)
		key := "region:" + lowerSubtag
		rec, ok := cpr.parent.registry.Records[key]
		if !ok || rec.Type != "region" {
			return false
		}
	}
	cpr.region = subtag
	return true
}

// tryParseAsVariant attempts to parse a subtag as a variant.
//
// Specification Reference:
// RFC 5646 (Section 2.2.5)
//
// Parameters:
//   - subtag: The subtag string to evaluate.
//
// Returns:
//   - bool: True if successfully parsed as a variant, false otherwise.
//   - error: ErrDuplicateVariant if the variant has already been encountered in validity mode, otherwise nil.
func (cpr *canonicalParseRun) tryParseAsVariant(subtag string) (bool, error) {
	// Spec Rule: RFC 5646 (Section 2.2.5)
	// Variant subtags that begin with a letter must be at least five characters,
	// and those starting with a digit must be at least four characters.
	startsWithLetter := len(subtag) >= minVariantLenAlpha && isAlpha(subtag[0])
	startsWithDigit := len(subtag) >= minVariantLenDigit && isDigit(subtag[0])
	isVariantFmt := (startsWithLetter || startsWithDigit) && isAlphanumeric(subtag)

	if cpr.isTagGrandfathered() {
		// Spec Rule: RFC 5646 (Section 2.2.8)
		// Grandfathered tags bypass structural ABNF checks for variants.
		isVariantFmt = isAlphanumeric(subtag)
	}

	if (cpr.state > stateAfterRegion && cpr.state != stateInVariant) || !isVariantFmt {
		return false, nil
	}
	if cpr.checkValidity {
		lowerSubtag := strings.ToLower(subtag)
		key := "variant:" + lowerSubtag
		rec, ok := cpr.parent.registry.Records[key]
		if !ok || rec.Type != "variant" {
			return false, nil
		}
		if cpr.seenVariants == nil {
			cpr.seenVariants = make(map[string]struct{})
		}
		// Spec Rule: RFC 5646 (Section 2.2.5, Clause 5)
		// The same variant subtag MUST NOT be used more than once within a language tag.
		if _, seen := cpr.seenVariants[lowerSubtag]; seen {
			return false, ErrDuplicateVariant
		}
		cpr.seenVariants[lowerSubtag] = struct{}{}
	}
	cpr.variants = append(cpr.variants, subtag)
	return true, nil
}

// handleExtensionSubtag parses a subtag associated with a registered extension block.
//
// Specification Reference:
// RFC 5646 (Section 2.2.6)
//
// Parameters:
//   - subtag: The subtag string to parse within the extension block.
//
// Returns:
//   - error: ErrInvalidSubtag if no singleton is active, otherwise any structural parsing error.
func (cpr *canonicalParseRun) handleExtensionSubtag(subtag string) error {
	if len(subtag) == 1 {
		return cpr.handleSingleton(subtag)
	}
	// Spec Rule: RFC 5646 (Section 2.2.6)
	// Extension subtags can only appear after a valid singleton prefix has been registered.
	if len(cpr.extensions) == 0 {
		return ErrInvalidSubtag
	}
	lastExt := &cpr.extensions[len(cpr.extensions)-1]
	if lastExt.Value == "" {
		lastExt.Value = subtag
	} else {
		lastExt.Value += "-" + subtag
	}
	cpr.extensionExpected = false
	return nil
}

// handleSingleton evaluates a single-character subtag to initiate extension or private-use blocks.
//
// Specification Reference:
// RFC 5646 (Section 2.1 and Section 2.2.6)
//
// Parameters:
//   - subtag: The singleton string containing the single character.
//
// Returns:
// - error: ErrEmptyExtension if an active extension was left without a value, ErrDuplicateSingleton if already seen in
// validity mode, otherwise nil.
func (cpr *canonicalParseRun) handleSingleton(subtag string) error {
	// Spec Rule: RFC 5646 (Section 2.2.6, Clause 6)
	// Each singleton MUST be followed by at least one extension subtag (cannot be empty).
	if cpr.extensionExpected {
		return ErrEmptyExtension
	}
	s := unicode.ToLower(rune(subtag[0]))
	if cpr.checkValidity {
		if cpr.seenSingletons == nil {
			cpr.seenSingletons = make(map[rune]struct{})
		}
		// Spec Rule: RFC 5646 (Section 2.2.6, Clause 3)
		// Each singleton subtag MUST appear at most one time in each tag.
		if _, ok := cpr.seenSingletons[s]; ok {
			return ErrDuplicateSingleton
		}
		cpr.seenSingletons[s] = struct{}{}
	}
	// Spec Rule: RFC 5646 (Section 2.2.7)
	// The single-character subtag 'x' is reserved to introduce private-use subtag sequences.
	if s == 'x' {
		cpr.state = stateInPrivateUse
		return nil
	}
	cpr.state = stateInExtension
	cpr.extensionExpected = true
	cpr.extensions = append(cpr.extensions, Extension{Singleton: s})
	return nil
}
