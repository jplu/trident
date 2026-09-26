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

import "strings"

// render reconstructs the language tag string from the parsed components.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - b: a strings.Builder used to assemble the formatted language tag.
func (cpr *canonicalParseRun) render(b *strings.Builder) {
	// Step 1: Render Primary Language or Private Use Tag
	// Reconstructs the primary language subtag or handles a tag consisting solely of private-use subtags.
	if cpr.language != "" {
		// Spec Rule: RFC 5646 (Section 2.1.1)
		// Primary language codes are recommended to be written in lowercase.
		b.WriteString(strings.ToLower(cpr.language))
	} else if len(cpr.privateuse) > 0 {
		// Spec Rule: RFC 5646 (Section 2.2.7)
		// A tag may consist entirely of private use subtags, separated from other subtags by 'x'.
		b.WriteByte('x')
		for _, subtag := range cpr.privateuse {
			b.WriteByte('-')
			b.WriteString(strings.ToLower(subtag))
		}
		return
	}

	// Step 2: Append Extended Language Subtags
	// Appends each extended language subtag following the primary language.
	for _, subtag := range cpr.extlangs {
		// Spec Rule: RFC 5646 (Section 2.1.1)
		// Extended language subtags are formatted and written in lowercase.
		b.WriteByte('-')
		b.WriteString(strings.ToLower(subtag))
	}

	// Step 3: Append Script Subtag
	// Appends the script subtag, if present, formatted using title case.
	if cpr.script != "" {
		// Spec Rule: RFC 5646 (Section 2.1.1)
		// Script codes use lowercase with the initial letter capitalized.
		b.WriteByte('-')
		writeTitleCase(b, cpr.script)
	}

	// Step 4: Append Region Subtag
	// Appends the region subtag, if present, formatted in uppercase.
	if cpr.region != "" {
		// Spec Rule: RFC 5646 (Section 2.1.1)
		// Country or region codes are recommended to be capitalized.
		b.WriteByte('-')
		b.WriteString(strings.ToUpper(cpr.region))
	}

	// Step 5: Append Variant Subtags
	// Appends each variant subtag in its parsed order.
	for _, subtag := range cpr.variants {
		// Spec Rule: RFC 5646 (Section 2.1.1)
		// Variant subtags are formatted and written in lowercase.
		b.WriteByte('-')
		b.WriteString(strings.ToLower(subtag))
	}

	// Step 6: Append Extension Subtags
	// Appends each extension block introduced by its unique singleton character.
	for _, ext := range cpr.extensions {
		b.WriteByte('-')
		b.WriteRune(ext.Singleton)
		if ext.Value != "" {
			b.WriteByte('-')
			// Spec Rule: RFC 5646 (Section 2.2.6)
			// Extension subtags are formatted and written in lowercase.
			b.WriteString(strings.ToLower(ext.Value))
		}
	}

	// Step 7: Append Private Use Subtags
	// Appends trailing private-use subtags separated by the singleton 'x'.
	if cpr.state == stateInPrivateUse && len(cpr.privateuse) > 0 {
		// Spec Rule: RFC 5646 (Section 2.2.7)
		// Private use subtags are written in lowercase and follow all other subtags.
		b.WriteByte('-')
		b.WriteByte('x')
		for _, subtag := range cpr.privateuse {
			b.WriteByte('-')
			b.WriteString(strings.ToLower(subtag))
		}
	}
}

// getPositions computes the final end positions of each component in the rendered tag string.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Returns:
//   - tagElementsPositions: a struct containing the logical component boundary indices.
func (cpr *canonicalParseRun) getPositions() tagElementsPositions {
	var pos tagElementsPositions
	cursor := 0

	// Implementation Note: Offset calculation using raw byte slices
	// This calculation assumes that all parsed subtags contain only US-ASCII characters,
	// allowing length calculations to use byte count directly instead of rune count.

	// Step 1: Compute Primary Language Offset
	// Calculates the end position of the primary language subtag.
	if cpr.language != "" {
		cursor = len(cpr.language)
	}
	pos.languageEnd = cursor

	// Step 2: Compute Extended Language Offset
	// Calculates the combined end position of any parsed extended language subtags.
	if len(cpr.extlangs) > 0 {
		for _, ext := range cpr.extlangs {
			cursor += 1 + len(ext)
		}
	}
	pos.extlangEnd = cursor

	// Step 3: Compute Script Offset
	// Calculates the end position of the script subtag if parsed.
	if cpr.script != "" {
		cursor += 1 + len(cpr.script)
	}
	pos.scriptEnd = cursor

	// Step 4: Compute Region Offset
	// Calculates the end position of the region subtag if parsed.
	if cpr.region != "" {
		cursor += 1 + len(cpr.region)
	}
	pos.regionEnd = cursor

	// Step 5: Compute Variant Offset
	// Calculates the combined end position of any variant subtags.
	if len(cpr.variants) > 0 {
		for _, v := range cpr.variants {
			cursor += 1 + len(v)
		}
	}
	pos.variantEnd = cursor

	// Step 6: Compute Extension Offset
	// Calculates the combined end position of all parsed extension blocks.
	if len(cpr.extensions) > 0 {
		for _, ext := range cpr.extensions {
			cursor += 1 + 1 // -s
			if ext.Value != "" {
				cursor += 1 + len(ext.Value)
			}
		}
	}
	pos.extensionEnd = cursor

	return pos
}
