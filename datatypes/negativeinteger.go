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

// ErrNotNegativeInteger is returned when a parsed integer value is not strictly less than zero,
// violating the maxInclusive constraint of the xsd:negativeInteger datatype.
var ErrNotNegativeInteger = errors.New("value must be strictly less than 0")

// NegativeInteger represents the negativeInteger datatype as defined in the governing specifications.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.15)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A structure wrapping NonPositiveInteger, representing the infinite set of negative mathematical
// integers: {..., -2, -1}. It restricts the value space by setting the maxInclusive facet to -1.
type NegativeInteger struct {
	// NonPositiveInteger provides the underlying arbitrary-precision integer representation
	// constrained to values less than or equal to 0.
	NonPositiveInteger
}

// ParseNegativeInteger parses a string literal matching the lexical representation of xsd:negativeInteger.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.15.1)
//
// Parameters:
//   - s: The raw string literal to be parsed and validated.
//
// Returns:
//   - NegativeInteger: The parsed negative integer on success.
//   - error: An error of type ErrNotNegativeInteger or ErrParseInteger if validation fails.
func ParseNegativeInteger(s string) (NegativeInteger, error) {
	npi, err := ParseNonPositiveInteger(s)
	if err != nil {
		return NegativeInteger{}, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.15)
	// The value space of negativeInteger is restricted to negative integers {..., -2, -1},
	// meaning a value of exactly 0 is strictly prohibited.
	if npi.getValue().Sign() == 0 {
		return NegativeInteger{}, ErrNotNegativeInteger
	}
	return NegativeInteger{NonPositiveInteger: npi}, nil
}

// String returns the canonical lexical representation of the NegativeInteger value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.15.2)
//
// Returns:
//   - string: The canonical decimal representation of the negative integer.
func (n NegativeInteger) String() string {
	return n.Integer.String()
}

// IsIdenticalWith checks if this NegativeInteger is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.15)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (n NegativeInteger) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(NegativeInteger); ok {
		return n.Integer.IsIdenticalWith(o.Integer)
	}
	return false
}
