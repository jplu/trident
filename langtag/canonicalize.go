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
	"sort"
	"strings"
)

// canonicalize applies all canonicalization rules to the parse run.
//
// Specification Reference:
// RFC 5646 (Section 4.5).
func (cpr *canonicalParseRun) canonicalize() {
	// Step 1: Extlang Replacement
	// Replaces an extended language subtag with its preferred primary language subtag.
	cpr.canonicalizeExtlangToPrimary()

	// Step 2: Deprecated Subtag Substitution
	// Replaces any deprecated subtags with their preferred-value alternatives.
	cpr.canonicalizeDeprecated()

	// Step 3: Variant Reordering
	// Sorts variant subtags according to mutual prefix dependencies and alphabetical fallback.
	cpr.canonicalizeVariantOrder()

	// Step 4: Extension Ordering
	// Reorders extension segments alphabetically by their case-insensitive singleton.
	cpr.canonicalizeExtensionOrder()
}

// canonicalizeExtlangToPrimary replaces an extlang subtag with its preferred primary language subtag.
//
// Specification Reference:
// RFC 5646 (Section 4.5).
func (cpr *canonicalParseRun) canonicalizeExtlangToPrimary() {
	// Spec Rule: RFC 5646 (Section 4.5)
	// Subtags are replaced by their 'Preferred-Value', if there is one. For extlangs, the original primary language
	// subtag is also replaced if there is a primary language subtag in the 'Preferred-Value'.
	if len(cpr.extlangs) == 0 {
		return
	}
	lowerLang := strings.ToLower(cpr.language)
	lowerExtlang := strings.ToLower(cpr.extlangs[0])

	// Implementation Note: Case-insensitive registry lookup
	// Format the registry search key using the prefix and lowercased extlang name.
	key := "extlang:" + lowerExtlang
	rec, ok := cpr.parent.registry.Records[key]
	if !ok || rec.Type != typeExtlang {
		return
	}

	hasMatchingPrefix := false
	for _, pfx := range rec.Prefix {
		if strings.EqualFold(pfx, lowerLang) {
			hasMatchingPrefix = true
			break
		}
	}
	if hasMatchingPrefix && rec.PreferredValue != "" {
		cpr.language = rec.PreferredValue
		cpr.extlangs = cpr.extlangs[1:] // Remove the used extlang.
	}
}

// canonicalizeDeprecated replaces individual deprecated subtags with their preferred value.
//
// Specification Reference:
// RFC 5646 (Section 4.5).
func (cpr *canonicalParseRun) canonicalizeDeprecated() {
	// Spec Rule: RFC 5646 (Section 4.5)
	// Redundant or grandfathered tags and individual subtags are replaced by their 'Preferred-Value', if there is one.
	replaceIfPreferred := func(subtag, subtagType string) string {
		if subtag == "" {
			return ""
		}
		key := subtagType + ":" + strings.ToLower(subtag)
		if rec, ok := cpr.parent.registry.Records[key]; ok && rec.PreferredValue != "" {
			return rec.PreferredValue
		}
		return subtag
	}

	// Step 1: Replace Language Subtag
	// Evaluates and updates the primary language subtag.
	cpr.language = replaceIfPreferred(cpr.language, "language")

	// Step 2: Replace Script Subtag
	// Evaluates and updates the script subtag.
	cpr.script = replaceIfPreferred(cpr.script, "script")

	// Step 3: Replace Region Subtag
	// Evaluates and updates the region subtag.
	cpr.region = replaceIfPreferred(cpr.region, "region")

	// Step 4: Replace Variant Subtags
	// Iterates and updates each variant subtag individually.
	for i, v := range cpr.variants {
		cpr.variants[i] = replaceIfPreferred(v, "variant")
	}
}

// compareVariants compares two variant subtags to determine their correct relative order.
//
// Specification Reference:
// RFC 5646 (Section 4.1)
//
// Parameters:
//   - variantI: The first variant subtag to compare.
//   - variantJ: The second variant subtag to compare.
//
// Returns:
//   - bool: True if variantI should be ordered before variantJ, false otherwise.
func (cpr *canonicalParseRun) compareVariants(variantI, variantJ string) bool {
	// Spec Rule: RFC 5646 (Section 4.1)
	// If a variant lists a second variant in one of its 'Prefix' fields, the first variant SHOULD appear directly after
	// the second variant in any language tag where both occur.
	keyI := "variant:" + strings.ToLower(variantI)
	keyJ := "variant:" + strings.ToLower(variantJ)
	recI, okI := cpr.parent.registry.Records[keyI]
	recJ, okJ := cpr.parent.registry.Records[keyJ]

	prefixContainsVariant := func(prefixes []string, variant string) bool {
		for _, p := range prefixes {
			for _, sub := range strings.Split(p, "-") {
				if strings.EqualFold(sub, variant) {
					return true
				}
			}
		}
		return false
	}

	// Implementation Note: Circular prefix checking
	// Check if variant J is a registered prefix dependency for variant I.
	if okI && prefixContainsVariant(recI.Prefix, variantJ) {
		return false // Variant J is in I's prefix, so I must come after J.
	}
	// Check if variant I is a registered prefix dependency for variant J.
	if okJ && prefixContainsVariant(recJ.Prefix, variantI) {
		return true // Variant I is in J's prefix, so I must come before J.
	}

	hasPrefixI := okI && len(recI.Prefix) > 0
	hasPrefixJ := okJ && len(recJ.Prefix) > 0
	if hasPrefixI != hasPrefixJ {
		return hasPrefixI // A variant with a prefix is more specific and comes first.
	}

	// Implementation Note: Case-insensitive sorting
	// Convert both variant strings to lowercase before checking lexicographical ordering to ensure consistent
	// canonicalization.
	return strings.ToLower(variantI) < strings.ToLower(variantJ)
}

// canonicalizeVariantOrder reorders variant subtags based on prefix dependencies and alphabetical fallback.
//
// Specification Reference:
// RFC 5646 (Section 4.1).
func (cpr *canonicalParseRun) canonicalizeVariantOrder() {
	// Spec Rule: RFC 5646 (Section 4.5)
	// If more than one variant appears within a tag, processors MAY reorder the variants to obtain better matching
	// behavior or more consistent presentation.
	if len(cpr.variants) <= 1 {
		return
	}
	sort.Slice(cpr.variants, func(i, j int) bool {
		return cpr.compareVariants(cpr.variants[i], cpr.variants[j])
	})
}

// canonicalizeScriptSuppression removes redundant script subtags when they match the suppressed script of the language.
//
// Specification Reference:
// RFC 5646 (Section 4.1).
func (cpr *canonicalParseRun) canonicalizeScriptSuppression() {
	// Spec Rule: RFC 5646 (Section 4.1)
	// The script subtag SHOULD NOT be used to form language tags unless the script adds some distinguishing information
	// to the tag. The 'Suppress-Script' field defines when users SHOULD NOT include a script subtag.
	if cpr.script == "" {
		return
	}
	key := "language:" + strings.ToLower(cpr.language)
	langRec, ok := cpr.parent.registry.Records[key]
	if ok && langRec.SuppressScript != "" && strings.EqualFold(cpr.script, langRec.SuppressScript) {
		cpr.script = ""
	}
}

// canonicalizeExtensionOrder sorts extension subtags by their singleton character in case-insensitive ASCII order.
//
// Specification Reference:
// RFC 5646 (Section 4.5).
func (cpr *canonicalParseRun) canonicalizeExtensionOrder() {
	// Spec Rule: RFC 5646 (Section 4.5)
	// Extension sequences are ordered into case-insensitive ASCII order by singleton subtag.
	if len(cpr.extensions) > 1 {
		sort.Slice(cpr.extensions, func(i, j int) bool {
			return cpr.extensions[i].Singleton < cpr.extensions[j].Singleton
		})
	}
}
