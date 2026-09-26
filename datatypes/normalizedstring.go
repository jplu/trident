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

import "unicode/utf8"

// NormalizedString represents the normalizedString datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.1)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A string scalar value representing whitespace-normalized strings, derived from the string primitive datatype where
// carriage return, line feed, and tab characters are replaced with space characters.
type NormalizedString string

// ParseNormalizedString parses a raw string literal into a NormalizedString value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.1)
//
// Parameters:
//   - s: The raw string literal to be parsed and normalized.
//
// Returns:
//   - NormalizedString: The parsed and normalized string value.
func ParseNormalizedString(s string) NormalizedString {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.1.1)
	// The whiteSpace facet of normalizedString is fixed to 'replace'. All occurrences of #x9 (tab), #xA (line feed),
	// and #xD (carriage return) must be replaced with #x20 (space).
	return NormalizedString(replaceWhitespace(s))
}

// String returns the canonical lexical representation of the NormalizedString value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.1)
//
// Returns:
//   - string: The canonical string representation, which corresponds to the identity mapping.
func (n NormalizedString) String() string {
	return string(n)
}

// IsIdenticalWith checks if this NormalizedString is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (n NormalizedString) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(NormalizedString); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 2.2.1)
		// Two string-derived values are identical if and only if they represent the exact same sequence of Unicode code
		// points.
		return n == o
	}
	return false
}

// Length returns the length of the NormalizedString value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1)
//
// Returns:
//   - int: The length of the string measured in terms of XML characters (Unicode code points).
func (n NormalizedString) Length() int {
	// Implementation Note: Character-based length measurement
	// Since Go strings are represented as UTF-8 byte sequences, the standard library's utf8.RuneCountInString function
	// is utilized to accurately count Unicode code points rather than raw bytes, satisfying character-based length
	// measurement.
	return utf8.RuneCountInString(string(n))
}
