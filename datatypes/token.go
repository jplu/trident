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

// Token represents the token datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.2)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// The set of strings that do not contain carriage return (#xD), line feed (#xA),
// or tab (#x9) characters, nor leading or trailing spaces, nor internal sequences
// of two or more consecutive spaces. It is derived from normalizedString.
type Token string

// ParseToken parses a string literal matching the lexical representation of xsd:token.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.2)
//
// Parameters:
//   - s: The raw string literal to be parsed.
//
// Returns:
//   - Token: The parsed Token value.
func ParseToken(s string) Token {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.2.1)
	// The whiteSpace facet of token is fixed to 'collapse'. This normalization is performed before mapping the lexical
	// literal to the value space.
	return Token(collapseWhitespace(s))
}

// String returns the canonical lexical representation of the Token value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.2)
//
// Returns:
//   - string: The canonical lexical representation of the Token value.
func (t Token) String() string {
	return string(t)
}

// IsIdenticalWith checks if this Token is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if the current Token is identical to the other XSDValue, false otherwise.
func (t Token) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Token); ok {
		return t == o
	}
	return false
}

// Length returns the length of the Token value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1)
//
// Returns:
//   - int: The number of Unicode code points (runes) in the Token string.
func (t Token) Length() int {
	// Implementation Note: Character-based length measurement.
	// Go strings are represented as UTF-8 byte sequences. The utf8 standard library package is used to count runes
	// (Unicode code points) rather than raw bytes.
	return utf8.RuneCountInString(string(t))
}
