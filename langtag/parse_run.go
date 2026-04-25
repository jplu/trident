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
	maxSubtagLen        = 8 // Maximum length of any subtag.
	maxExtlangs         = 1 // Maximum number of extended language subtags allowed by spec.
	scriptLen           = 4 // A script subtag is always 4 letters.
	regionAlphaLen      = 2 // An alphabetic region subtag is always 2 letters.
	regionNumericLen    = 3 // A numeric region subtag is always 3 digits.
	extlangLen          = 3 // An extended language subtag is always 3 letters.
	shortPrimaryLangLen = 3 // Max length of a primary language that can be followed by an extlang.
	minVariantLenAlpha  = 5 // Min length of a variant starting with a letter.
	minVariantLenDigit  = 4 // Min length of a variant starting with a digit.
)

// parseState represents the current position in the state machine during parsing.
type parseState int

const (
	stateStart         parseState = iota // Expecting a primary language subtag.
	stateAfterLanguage                   // After a 2-3 letter primary language, expecting extlang, script, etc.
	stateAfterExtLang                    // After a >3 letter primary lang or an extlang, expecting script, region, etc.
	stateAfterScript                     // After a script, expecting region, variant, etc.
	stateAfterRegion                     // After a region, expecting variant, etc.
	stateInVariant                       // In a sequence of one or more variants.
	stateInExtension                     // In an extension sequence (after a singleton).
	stateInPrivateUse                    // In a private-use sequence (after 'x').
)

// canonicalParseRun holds the state for a single "valid" parsing and
// canonicalization operation. It is not exposed publicly.
type canonicalParseRun struct {
	parent *Parser
	// The fields below hold the parsed subtags.
	language   string
	extlangs   []string
	script     string
	region     string
	variants   []string
	extensions []Extension
	privateuse []string
	// Internal state for the parsing process.
	subtags           []string
	state             parseState
	checkValidity     bool
	seenVariants      map[string]struct{}
	seenSingletons    map[rune]struct{}
	extlangsCount     int
	extensionExpected bool
}

// newCanonicalParseRun creates a new parsing run for a given tag string.
func (p *Parser) newCanonicalParseRun(input string, checkValidity bool) *canonicalParseRun {
	return &canonicalParseRun{
		parent:        p,
		subtags:       strings.Split(input, "-"),
		checkValidity: checkValidity,
	}
}

// validateSubtag performs basic syntactic checks on a single subtag.
func validateSubtag(subtag string) error {
	if len(subtag) == 0 {
		return ErrEmptySubtag
	}
	if len(subtag) > maxSubtagLen {
		return ErrSubtagTooLong
	}
	return nil
}

// prepareSubtags trims a potential trailing hyphen and returns the subtags to parse.
func (cpr *canonicalParseRun) prepareSubtags() ([]string, bool) {
	hasTrailingHyphen := len(cpr.subtags) > 1 && cpr.subtags[len(cpr.subtags)-1] == ""
	if hasTrailingHyphen {
		return cpr.subtags[:len(cpr.subtags)-1], true
	}
	return cpr.subtags, false
}

// parsePrivateUseOnly handles tags that start with the private-use singleton "x".
func (cpr *canonicalParseRun) parsePrivateUseOnly(subtags []string) error {
	if len(subtags) == 1 {
		return ErrEmptyPrivateUse
	}
	for _, subtag := range subtags[1:] {
		if err := validateSubtag(subtag); err != nil {
			return err
		}
		cpr.privateuse = append(cpr.privateuse, subtag)
	}
	cpr.state = stateInPrivateUse
	return nil
}

// processSubtags loops through the subtags of a standard langtag and parses them.
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
		case stateStart, stateAfterLanguage, stateAfterExtLang, stateAfterScript, stateAfterRegion, stateInVariant:
			if err := cpr.handleLangtagSubtag(i, subtag); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkFinalState performs validation checks after all subtags have been processed.
func (cpr *canonicalParseRun) checkFinalState(hasTrailingHyphen bool) error {
	if hasTrailingHyphen {
		if cpr.extensionExpected {
			return ErrEmptyExtension
		}
		if cpr.state == stateInPrivateUse && len(cpr.privateuse) == 0 {
			return ErrEmptyPrivateUse
		}
	}

	// This final check is for cases where a singleton is the last subtag, e.g., "en-a-b-foo"
	if cpr.extensionExpected {
		return ErrEmptyExtension
	}
	return nil
}

// parse executes the parsing state machine over the subtags.
func (cpr *canonicalParseRun) parse() error {
	subtagsToParse, hasTrailingHyphen := cpr.prepareSubtags()

	if len(subtagsToParse) > 0 && strings.EqualFold(subtagsToParse[0], "x") {
		return cpr.parsePrivateUseOnly(subtagsToParse)
	}

	if err := cpr.processSubtags(subtagsToParse); err != nil {
		return err
	}

	return cpr.checkFinalState(hasTrailingHyphen)
}

// handlePrimaryLanguage handles the parsing and validation of the first subtag.
func (cpr *canonicalParseRun) handlePrimaryLanguage(subtag string) error {
	minLen := 2
	if !cpr.checkValidity {
		minLen = 1
	}
	if len(subtag) < minLen || len(subtag) > 8 || !isAlphabetic(subtag) {
		return ErrInvalidLanguage
	}

	if cpr.checkValidity {
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

// checkForTooManyExtlangs ensures that no more than the maximum number of extlangs are present.
// RFC 5646 Sec 2.2.2 Rule 4 allows at most one extlang.
func (cpr *canonicalParseRun) checkForTooManyExtlangs(subtag string) error {
	if cpr.extlangsCount >= maxExtlangs && len(subtag) == extlangLen && isAlphabetic(subtag) {
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

// handleLangtagSubtag dispatches parsing for a subtag that is part of the main langtag.
func (cpr *canonicalParseRun) handleLangtagSubtag(i int, subtag string) error {
	if i == 0 {
		return cpr.handlePrimaryLanguage(subtag)
	}
	if len(subtag) == 1 {
		return cpr.handleSingleton(subtag)
	}

	if err := cpr.checkForTooManyExtlangs(subtag); err != nil {
		return err
	}

	// Attempt to parse the subtag in the order defined by the RFC:
	// extlang -> script -> region -> variant
	if cpr.tryParseAsExtlang(subtag) {
		cpr.state = stateAfterExtLang
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

// tryParseAsExtlang attempts to parse the subtag as an extended language.
func (cpr *canonicalParseRun) tryParseAsExtlang(subtag string) bool {
	if cpr.state != stateAfterLanguage || cpr.extlangsCount >= maxExtlangs ||
		len(subtag) != extlangLen || !isAlphabetic(subtag) {
		return false
	}
	if cpr.checkValidity {
		lowerSubtag := strings.ToLower(subtag)
		key := "extlang:" + lowerSubtag
		rec, ok := cpr.parent.registry.Records[key]
		if !ok || rec.Type != typeExtlang {
			return false // It's not a valid extlang, maybe it's a script or region.
		}
	}
	cpr.extlangsCount++
	cpr.extlangs = append(cpr.extlangs, subtag)
	return true
}

// tryParseAsScript attempts to parse the subtag as a script.
func (cpr *canonicalParseRun) tryParseAsScript(subtag string) bool {
	if cpr.state > stateAfterExtLang || len(subtag) != scriptLen || !isAlphabetic(subtag) {
		return false
	}
	if cpr.checkValidity {
		lowerSubtag := strings.ToLower(subtag)
		key := "script:" + lowerSubtag
		rec, ok := cpr.parent.registry.Records[key]
		if !ok || rec.Type != "script" {
			return false // It's not a valid script, maybe it's a region or variant.
		}
	}
	cpr.script = subtag
	return true
}

// tryParseAsRegion attempts to parse the subtag as a region.
func (cpr *canonicalParseRun) tryParseAsRegion(subtag string) bool {
	isRegionFmt := (len(subtag) == regionAlphaLen && isAlphabetic(subtag)) ||
		(len(subtag) == regionNumericLen && isNumeric(subtag))
	if cpr.state > stateAfterScript || !isRegionFmt {
		return false
	}
	if cpr.checkValidity {
		lowerSubtag := strings.ToLower(subtag)
		key := "region:" + lowerSubtag
		rec, ok := cpr.parent.registry.Records[key]
		if !ok || rec.Type != "region" {
			return false // It's not a valid region, maybe it's a variant.
		}
	}
	cpr.region = subtag
	return true
}

// tryParseAsVariant attempts to parse the subtag as a variant.
func (cpr *canonicalParseRun) tryParseAsVariant(subtag string) (bool, error) {
	startsWithLetter := len(subtag) >= minVariantLenAlpha && isAlpha(subtag[0])
	startsWithDigit := len(subtag) >= minVariantLenDigit && isDigit(subtag[0])
	isVariantFmt := (startsWithLetter || startsWithDigit) && isAlphanumeric(subtag)

	if !cpr.checkValidity {
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
			return false, nil // Not a valid variant. Could be an error for the whole tag.
		}
		if cpr.seenVariants == nil {
			cpr.seenVariants = make(map[string]struct{})
		}
		if _, seen := cpr.seenVariants[lowerSubtag]; seen {
			return false, ErrDuplicateVariant
		}
		cpr.seenVariants[lowerSubtag] = struct{}{}
	}
	cpr.variants = append(cpr.variants, subtag)
	return true, nil
}

// handleExtensionSubtag parses a subtag that is part of an extension sequence.
func (cpr *canonicalParseRun) handleExtensionSubtag(subtag string) error {
	if len(subtag) == 1 {
		return cpr.handleSingleton(subtag)
	}
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

// handleSingleton handles a single-character subtag, which starts an
// extension or a private-use sequence.
func (cpr *canonicalParseRun) handleSingleton(subtag string) error {
	if cpr.extensionExpected {
		return ErrEmptyExtension
	}
	s := unicode.ToLower(rune(subtag[0]))
	if cpr.checkValidity {
		if cpr.seenSingletons == nil {
			cpr.seenSingletons = make(map[rune]struct{})
		}
		if _, ok := cpr.seenSingletons[s]; ok {
			return ErrDuplicateSingleton
		}
		cpr.seenSingletons[s] = struct{}{}
	}
	if s == 'x' {
		cpr.state = stateInPrivateUse
		return nil
	}
	cpr.state = stateInExtension
	cpr.extensionExpected = true
	cpr.extensions = append(cpr.extensions, Extension{Singleton: s})
	return nil
}
