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

import (
	"unicode/utf8"

	"github.com/jplu/trident/iri"
)

// AnyURI represents the anyURI primitive datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.17)
// W3C RDF 1.2 (Section 5.1)
// RFC 3987 (Section 2.2)
//
// Representation:
// A finite-length sequence of characters that maps to an Internationalized Resource Identifier (IRI) Reference, with a
// default whiteSpace facet behavior of 'collapse'.
type AnyURI string

// ParseAnyURI parses a string literal matching the lexical representation of xsd:anyURI.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.17.1)
// RFC 3987 (Section 2.2)
//
// Parameters:
//   - s: The raw input string containing the candidate IRI reference.
//
// Returns:
//   - AnyURI: The parsed and validated AnyURI value.
//   - error: An error if the input string fails to conform to the IRI reference syntax.
func ParseAnyURI(s string) (AnyURI, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.17.1)
	// The whiteSpace facet for anyURI is fixed to 'collapse', requiring leading, trailing, and duplicate internal
	// whitespace to be normalized prior to processing.
	collapsed := collapseWhitespace(s)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.17.2)
	// Identity and comparison are based on exact character sequences, meaning the lexical representation should not
	// undergo implicit Unicode normalization (such as NFC) during validation.
	_, err := iri.ParseRef(collapsed)
	if err != nil {
		return "", err
	}

	return AnyURI(collapsed), nil
}

// String returns the canonical lexical representation of the AnyURI value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.17.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical string representation of the AnyURI value.
func (a AnyURI) String() string {
	return string(a)
}

// IsIdenticalWith checks if this AnyURI is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.17.2)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (a AnyURI) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(AnyURI); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.17.2)
		// Two anyURI values are identical if and only if they consist of exactly the same sequence of characters.
		return a == o
	}
	return false
}

// Length returns the length of the AnyURI value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1.3)
//
// Parameters:
//
// Returns:
//   - int: The number of XML characters (Unicode code points) contained in the AnyURI value.
func (a AnyURI) Length() int {
	// Implementation Note: Character-based length measurement
	// Go strings are represented as UTF-8 byte sequences. We use the utf8 standard library package to count runes
	// (Unicode code points) rather than raw bytes to satisfy the XML character-counting constraint.
	return utf8.RuneCountInString(string(a))
}

// ToRef parses the AnyURI and returns its representation as a compiled IRI reference wrapper.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Parameters:
//
// Returns:
//   - *iri.Ref: A pointer to the parsed and compiled IRI reference structure, or nil if parsing fails.
func (a AnyURI) ToRef() *iri.Ref {
	ref, _ := iri.ParseRef(string(a))
	return ref
}
