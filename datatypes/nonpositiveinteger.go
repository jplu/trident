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

// ErrNotNonPositiveInteger is returned when a parsed integer is strictly greater than 0,
// violating the maxInclusive constraint of the xsd:nonPositiveInteger datatype.
var ErrNotNonPositiveInteger = errors.New("value must be less than or equal to 0")

// NonPositiveInteger represents the nonPositiveInteger datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.14)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// The value space of nonPositiveInteger is the infinite set of all mathematical
// integers less than or equal to 0. It is derived from integer by restricting
// the upper bound with a maxInclusive facet value of 0.
type NonPositiveInteger struct {
	// Integer represents the underlying arbitrary-precision integer value.
	Integer
}

// ParseNonPositiveInteger parses a string lexical representation into a NonPositiveInteger.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.14)
//
// Parameters:
//   - s: The string literal matching the lexical representation of xsd:nonPositiveInteger.
//
// Returns:
//   - NonPositiveInteger: The parsed nonPositiveInteger value.
//
// - error: An error of type ErrParseInteger if the literal format is invalid, or ErrNotNonPositiveInteger if the parsed
// value is strictly greater than 0.
func ParseNonPositiveInteger(s string) (NonPositiveInteger, error) {
	i, err := ParseInteger(s)
	if err != nil {
		return NonPositiveInteger{}, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.14)
	// The value space of nonPositiveInteger contains only mathematical integers less than or equal to 0. A value
	// strictly greater than 0 is invalid.
	if i.getValue().Sign() > 0 {
		return NonPositiveInteger{}, ErrNotNonPositiveInteger
	}
	return NonPositiveInteger{Integer: i}, nil
}

// String returns the canonical lexical representation of the NonPositiveInteger value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical lexical representation as defined in the specification.
func (n NonPositiveInteger) String() string {
	return n.Integer.String()
}

// IsIdenticalWith checks if this NonPositiveInteger is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.14.1)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - bool: True if the other value is identical to this value, false otherwise.
func (n NonPositiveInteger) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(NonPositiveInteger); ok {
		return n.Integer.IsIdenticalWith(o.Integer)
	}
	return false
}
