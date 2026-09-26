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

// nmTokenRegex is the compiled regular expression used to validate NMTOKEN lexical values.
// It matches the Nmtoken production defined in XML 1.0.
var nmTokenRegex = regexp.MustCompile(`^[:` + xmlNameCharRange + `]+$`)

// NMTOKEN represents the NMTOKEN datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.4)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// NMTOKEN is represented as a string type derived from token by restriction,
// containing a set of characters that conform to the Nmtoken production in XML 1.0.
type NMTOKEN Token

// ParseNMTOKEN parses a string literal matching the lexical representation of xsd:NMTOKEN.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.4)
//
// Parameters:
//   - s: The raw string literal to be parsed and validated.
//
// Returns:
//   - NMTOKEN: The successfully parsed NMTOKEN value on success.
//   - error: An error if the input literal does not conform to the NMTOKEN lexical constraints.
func ParseNMTOKEN(s string) (NMTOKEN, error) {
	// Step 1: Token Normalization
	// Parse and normalize the raw string by applying token-level collapse whitespace strategy.
	tok := ParseToken(s)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.4)
	// The lexical space of NMTOKEN consists of tokens that match the XML 1.0 Nmtoken
	// production (one or more NameChar characters). Unlike NCName, colons are permitted
	// as they are valid XML NameChar characters.
	if !nmTokenRegex.MatchString(string(tok)) {
		return "", errors.New("value is not a valid XML NMTOKEN")
	}
	return NMTOKEN(tok), nil
}

// String returns the canonical lexical representation of the NMTOKEN value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.4)
//
// Parameters:
//
// Returns:
//   - string: The canonical string representation of the NMTOKEN value.
func (n NMTOKEN) String() string {
	return string(n)
}

// IsIdenticalWith checks if this NMTOKEN is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.4)
//
// Parameters:
//   - other: The other XSDValue to compare with the receiver.
//
// Returns:
//   - bool: True if this NMTOKEN is identical to the other value, false otherwise.
func (n NMTOKEN) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(NMTOKEN); ok {
		return n == o
	}
	return false
}

// Length returns the length of the NMTOKEN value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1)
//
// Parameters:
//
// Returns:
//   - int: The number of XML characters (Unicode code points) contained in the NMTOKEN value.
func (n NMTOKEN) Length() int {
	return Token(n).Length()
}
