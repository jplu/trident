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

// String represents the xsd:string primitive datatype as defined in the governing specifications.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.1)
// W3C RDF 1.2 (Section 2.2)
//
// Representation:
// An ordered, finite sequence of zero or more Unicode characters matching the XML Char production.
type String string

// ParseString parses a raw string literal into a String value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.1.1)
//
// Parameters:
//   - s: The raw string literal to be parsed.
//
// Returns:
//   - String: The mapped string value.
func ParseString(s string) String {
	return String(s)
}

// String returns the canonical lexical representation of the String value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.1.2)
//
// Returns:
//   - string: The canonical lexical representation of the string.
func (s String) String() string {
	return string(s)
}

// IsIdenticalWith checks if this String is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.1.4)
//
// Parameters:
//   - other: The other XSDValue to compare.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (s String) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(String); ok {
		return s == o
	}
	return false
}

// Length returns the length of the String value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1.3)
//
// Returns:
//   - int: The length of the string measured in XML characters.
func (s String) Length() int {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.1.3)
	// The length of a string is measured in terms of the number of XML characters.

	// Implementation Note: Character-based length measurement
	// Go strings are represented as UTF-8 byte sequences. We use the standard utf8.RuneCountInString function to count
	// Unicode code points (runes) instead of raw bytes.
	return utf8.RuneCountInString(string(s))
}
