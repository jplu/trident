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
)

// ErrNotNonNegativeInteger is returned when a parsed integer value is strictly less than zero,
// violating the minInclusive constraint of the xsd:nonNegativeInteger datatype.
var ErrNotNonNegativeInteger = errors.New("value must be greater than or equal to 0")

// NonNegativeInteger represents the nonNegativeInteger datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.20)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// An arbitrary-precision mathematical integer value that is greater than or equal to 0, wrapped within an Integer
// struct.
type NonNegativeInteger struct {
	// Integer represents the underlying arbitrary-precision integer value.
	Integer
}

// ParseNonNegativeInteger parses a string literal matching the lexical representation of xsd:nonNegativeInteger.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.20.1)
//
// Parameters:
//   - s: The raw string representation to parse.
//
// Returns:
//   - NonNegativeInteger: The parsed non-negative integer value.
//
// - error: An error of type ErrParseInteger if the literal format is invalid, or ErrNotNonNegativeInteger if the parsed
// value is strictly negative.
func ParseNonNegativeInteger(s string) (NonNegativeInteger, error) {
	i, err := ParseInteger(s)
	if err != nil {
		return NonNegativeInteger{}, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.20.1)
	// The value space of nonNegativeInteger is restricted by a minInclusive facet value of 0, meaning negative numbers
	// are strictly prohibited.
	if i.getValue().Sign() < 0 {
		return NonNegativeInteger{}, ErrNotNonNegativeInteger
	}
	return NonNegativeInteger{Integer: i}, nil
}

// String returns the canonical lexical representation of the NonNegativeInteger value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.2)
//
// Returns:
//   - string: The canonical string representation of the non-negative integer.
func (n NonNegativeInteger) String() string {
	return n.Integer.String()
}

// IsIdenticalWith checks if this NonNegativeInteger is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.20.1)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (n NonNegativeInteger) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(NonNegativeInteger); ok {
		return n.Integer.IsIdenticalWith(o.Integer)
	}
	return false
}
