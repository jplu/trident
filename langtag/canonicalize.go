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

// canonicalize applies all canonicalization rules from RFC 5646, Sec 4.5.
func (cpr *canonicalParseRun) canonicalize() {
	cpr.canonicalizeExtlangToPrimary()
	cpr.canonicalizeDeprecated()
	cpr.canonicalizeVariantOrder()
	cpr.canonicalizeScriptSuppression()
	cpr.canonicalizeExtensionOrder()
}

// canonicalizeExtlangToPrimary replaces an extlang with its preferred primary language subtag.
func (cpr *canonicalParseRun) canonicalizeExtlangToPrimary() {
	if len(cpr.extlangs) == 0 {
		return
	}
	lowerLang := strings.ToLower(cpr.language)
	lowerExtlang := strings.ToLower(cpr.extlangs[0])

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

// canonicalizeDeprecated replaces individual deprecated subtags with their 'Preferred-Value'.
func (cpr *canonicalParseRun) canonicalizeDeprecated() {
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

	cpr.language = replaceIfPreferred(cpr.language, "language")
	cpr.script = replaceIfPreferred(cpr.script, "script")
	cpr.region = replaceIfPreferred(cpr.region, "region")
	for i, v := range cpr.variants {
		cpr.variants[i] = replaceIfPreferred(v, "variant")
	}
}

// compareVariants is a helper for sorting variants based on prefix dependencies.
func (cpr *canonicalParseRun) compareVariants(variantI, variantJ string) bool {
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

	if okI && prefixContainsVariant(recI.Prefix, variantJ) {
		return false // J is in I's prefix, so I must come after J.
	}
	if okJ && prefixContainsVariant(recJ.Prefix, variantI) {
		return true // I is in J's prefix, so I must come before J.
	}

	hasPrefixI := okI && len(recI.Prefix) > 0
	hasPrefixJ := okJ && len(recJ.Prefix) > 0
	if hasPrefixI != hasPrefixJ {
		return hasPrefixI // A variant with a prefix is more specific and comes first.
	}

	return variantI < variantJ // Fallback to alphabetical order.
}

// canonicalizeVariantOrder reorders variant subtags based on prefix dependencies.
func (cpr *canonicalParseRun) canonicalizeVariantOrder() {
	if len(cpr.variants) <= 1 {
		return
	}
	sort.Slice(cpr.variants, func(i, j int) bool {
		return cpr.compareVariants(cpr.variants[i], cpr.variants[j])
	})
}

// canonicalizeScriptSuppression removes redundant script subtags.
func (cpr *canonicalParseRun) canonicalizeScriptSuppression() {
	if cpr.script == "" {
		return
	}
	key := "language:" + strings.ToLower(cpr.language)
	langRec, ok := cpr.parent.registry.Records[key]
	if ok && langRec.SuppressScript != "" && strings.EqualFold(cpr.script, langRec.SuppressScript) {
		cpr.script = ""
	}
}

// canonicalizeExtensionOrder sorts extensions by their singleton character.
func (cpr *canonicalParseRun) canonicalizeExtensionOrder() {
	if len(cpr.extensions) > 1 {
		sort.Slice(cpr.extensions, func(i, j int) bool {
			return cpr.extensions[i].Singleton < cpr.extensions[j].Singleton
		})
	}
}
