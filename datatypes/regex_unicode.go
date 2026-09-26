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

package datatypes

import "strings"

// xmlNameStartCharClass represents the NameStartChar production as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
//
// Representation:
// An RE2-compatible character class representing the NameStartChar range from the XML 1.0 specification.
// W3C XSD 1.1 Part 2 regular expressions utilize this character range to implement the \i and \I (NameStartChar and
// non-NameStartChar) escape categories.
const xmlNameStartCharClass = `[A-Za-z_:\x{C0}-\x{D6}\x{D8}-\x{F6}\x{F8}-\x{2FF}\x{370}-\x{37D}\x{37F}-\x{1FFF}\x{200C}-\x{200D}\x{2070}-\x{218F}\x{2C00}-\x{2FEF}\x{3001}-\x{D7FF}\x{F900}-\x{FDCF}\x{FDF0}-\x{FFFD}\x{10000}-\x{EFFFF}]`

// xmlNameCharClass represents the NameChar production as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
//
// Representation:
// An RE2-compatible character class representing the NameChar range from the XML 1.0 specification.
// W3C XSD 1.1 Part 2 regular expressions utilize this character range to implement the \c and \C (NameChar and
// non-NameChar) escape categories.
const xmlNameCharClass = `[A-Za-z_:\x{C0}-\x{D6}\x{D8}-\x{F6}\x{F8}-\x{2FF}\x{370}-\x{37D}\x{37F}-\x{1FFF}\x{200C}-\x{200D}\x{2070}-\x{218F}\x{2C00}-\x{2FEF}\x{3001}-\x{D7FF}\x{F900}-\x{FDCF}\x{FDF0}-\x{FFFD}\x{10000}-\x{EFFFF}\-\.0-9\x{B7}\x{0300}-\x{036F}\x{203F}-\x{2040}]`

// translateUnicodeEscape translates an XML Schema Unicode property block or category escape into a syntax recognized by
// Go's RE2 regular expression engine.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2.3)
//
// Parameters:
//   - escape: The raw Unicode property or block escape sequence to translate.
//
// Returns:
//   - string: The translated escape sequence compatible with Go's RE2 engine.
func translateUnicodeEscape(escape string) string {
	if len(escape) < 5 || (escape[0:2] != `\p` && escape[0:2] != `\P`) {
		return escape
	}

	prefix := escape[:3]
	inner := escape[3 : len(escape)-1]

	if !strings.HasPrefix(inner, "Is") {
		return escape
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.4.2.3)
	// Unicode block escapes are prefixed with "Is" (e.g., \p{IsBasicLatin}).
	//
	// Implementation Note: RE2 engine compatibility
	// Go's standard library RE2 engine does not recognize the "Is" prefix for Unicode blocks and expects the block name
	// directly or with a "Block_" prefix.
	blockName := inner[2:]

	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.4.2.3)
	// Regular expressions reference Unicode 3.1 as the base for character property definitions.
	//
	// Implementation Note: Go Unicode tables mapping
	// Several block names specified in Unicode 3.1 have been modified in the newer Unicode standards implemented by
	// Go's unicode tables. These cases are mapped here to maintain conformance with the W3C XSD 1.1 Part 2 standard
	// names.
	switch blockName {
	case "PrivateUse":
		blockName = "Private_Use_Area"
	case "Greek":
		blockName = "Greek_and_Coptic"
	case "CombiningMarksforSymbols":
		blockName = "Combining_Diacritical_Marks_for_Symbols"
	}

	return prefix + blockName + "}"
}
