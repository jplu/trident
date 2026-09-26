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
	"errors"
	"regexp"
)

// The xmlNameStartCharRange constant defines the set of characters allowed at the beginning of an XML Name.
// This is specified in the W3C XML 1.0 (Fifth Edition) Recommendation under NameStartChar.
const xmlNameStartCharRange = `A-Z_a-z\x{C0}-\x{D6}\x{D8}-\x{F6}\x{F8}-\x{2FF}\x{370}-\x{37D}\x{37F}-\x{1FFF}\x{200C}-\x{200D}\x{2070}-\x{218F}\x{2C00}-\x{2FEF}\x{3001}-\x{D7FF}\x{F900}-\x{FDCF}\x{FDF0}-\x{FFFD}\x{10000}-\x{EFFFF}`

// The xmlNameCharRange constant defines the set of characters allowed in subsequent positions of an XML Name.
// This is specified in the W3C XML 1.0 (Fifth Edition) Recommendation under NameChar.
const xmlNameCharRange = xmlNameStartCharRange + `\-\.0-9\x{B7}\x{0300}-\x{036F}\x{203F}-\x{2040}`

// The nameRegex variable is the compiled regular expression used to validate an XML Name.
// Under XML 1.0, a Name can also contain colons (:).
var nameRegex = regexp.MustCompile(`^[:` + xmlNameStartCharRange + `][:` + xmlNameCharRange + `]*$`)

// Name represents the Name datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.6)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// An XML Name, derived from the token datatype by restriction, representing a string that conforms to the Name
// production in XML 1.0.
type Name Token

// ParseName parses a string literal matching the lexical representation of xsd:Name.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.6)
//
// Parameters:
//   - s: The raw string literal to be parsed and validated.
//
// Returns:
//   - Name: The validated Name value.
//   - error: An error if the literal does not conform to the XML Name syntax or lexical rules.
func ParseName(s string) (Name, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.6).
	// Normalization must occur first by collapsing all whitespace characters.
	tok := ParseToken(s)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.6).
	// Validate that the normalized string matches the XML 1.0 Name production.
	if !nameRegex.MatchString(string(tok)) {
		return "", errors.New("value is not a valid XML Name")
	}
	return Name(tok), nil
}

// String returns the canonical lexical representation of the Name value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.6)
//
// Parameters:
//
// Returns:
//   - string: The canonical string representation of the Name, which corresponds to the identity mapping.
func (n Name) String() string {
	return string(n)
}

// IsIdenticalWith checks if this Name is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.6)
//
// Parameters:
//   - other: The other XSDValue to compare with for identity.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (n Name) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Name); ok {
		return n == o
	}
	return false
}

// Length returns the length of the Name value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1)
//
// Parameters:
//
// Returns:
//   - int: The length of the Name value measured in terms of XML characters (Unicode code points).
func (n Name) Length() int {
	return Token(n).Length()
}
