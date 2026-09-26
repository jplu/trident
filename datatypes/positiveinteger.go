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

// ErrNotPositiveInteger is returned when a parsed integer value is not strictly greater than zero,
// violating the value space constraints of the XSD positiveInteger datatype.
var ErrNotPositiveInteger = errors.New("value must be strictly greater than 0")

// PositiveInteger represents the positiveInteger datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.25)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// The value space of positiveInteger is the infinite set of strictly positive mathematical
// integers: {1, 2, ...}. It is derived from nonNegativeInteger by setting the minInclusive facet to 1.
type PositiveInteger struct {
	// NonNegativeInteger is the embedded base type from which positiveInteger is derived.
	NonNegativeInteger
}

// ParsePositiveInteger parses a string literal matching the lexical representation of xsd:positiveInteger.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.25)
//
// Parameters:
//   - s: The raw string literal to be parsed.
//
// Returns:
//   - PositiveInteger: The parsed PositiveInteger value.
//   - error: An error if the lexical format is invalid or the mathematical value is not strictly positive.
func ParsePositiveInteger(s string) (PositiveInteger, error) {
	nni, err := ParseNonNegativeInteger(s)
	if err != nil {
		return PositiveInteger{}, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.25)
	// The positiveInteger datatype is derived from nonNegativeInteger by setting
	// the minInclusive facet to 1. Consequently, a value of 0 is strictly excluded.
	if nni.getValue().Sign() == 0 {
		return PositiveInteger{}, ErrNotPositiveInteger
	}
	return PositiveInteger{NonNegativeInteger: nni}, nil
}

// String returns the canonical lexical representation of the PositiveInteger value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.25.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical decimal representation of the positive integer.
func (p PositiveInteger) String() string {
	return p.Integer.String()
}

// IsIdenticalWith checks if this PositiveInteger is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.25)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (p PositiveInteger) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(PositiveInteger); ok {
		return p.Integer.IsIdenticalWith(o.Integer)
	}
	return false
}
