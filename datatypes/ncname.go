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

// ncNameRegex is the compiled regular expression used to validate NCName lexical values.
// It matches the NCName production defined in W3C XSD 1.1 Part 2 (Section 3.4.7),
// which excludes the colon character (':') from NameStartChar and NameChar.
var ncNameRegex = regexp.MustCompile(`^[` + xmlNameStartCharRange + `][` + xmlNameCharRange + `]*$`)

// NCName represents the NCName datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.7)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// The set of all strings that conform to the NCName production defined in Namespaces in XML 1.0,
// representing XML non-colonized names. It is derived from the Name datatype by restriction,
// explicitly prohibiting colon characters.
type NCName Name

// ParseNCName parses a string literal matching the lexical representation of xsd:NCName.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.7)
//
// Parameters:
//   - s: The raw string literal to be parsed and validated.
//
// Returns:
//   - NCName: The parsed and validated NCName value.
//   - error: An error if the input literal does not conform to the NCName constraints.
func ParseNCName(s string) (NCName, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.7)
	// The NCName datatype inherits the 'collapse' whiteSpace normalization facet from the token datatype.
	tok := ParseToken(s)
	if !ncNameRegex.MatchString(string(tok)) {
		return "", errors.New("value is not a valid XML NCName (non-colonized name)")
	}
	return NCName(tok), nil
}

// String returns the canonical lexical representation of the NCName value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.7.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical string representation, which is the identity mapping on its value space.
func (n NCName) String() string {
	return string(n)
}

// IsIdenticalWith checks if this NCName is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.7)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - bool: True if this NCName value is identical to the other value; false otherwise.
func (n NCName) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(NCName); ok {
		return n == o
	}
	return false
}

// Length returns the length of the NCName value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1)
//
// Parameters:
//
// Returns:
//   - int: The character-based length of the value.
func (n NCName) Length() int {
	return Token(n).Length()
}
