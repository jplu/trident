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
	"strconv"
)

// ErrOutOfBoundsInt represents the error returned when a parsed integer value
// falls outside the permissible range defined for the xsd:int value space.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.17).
var ErrOutOfBoundsInt = errors.New("value out of bounds for xsd:int (-2147483648 to 2147483647)")

// Int represents the int datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.17)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A 32-bit signed integer representing mathematical integers in the range
// from -2147483648 to 2147483647 inclusive, derived from long by restriction.
type Int int32

// ParseInt parses a string literal matching the lexical representation of xsd:int.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.17.1)
//
// Parameters:
//   - s: A string containing the lexical representation of a 32-bit signed integer.
//
// Returns:
//   - Int: The parsed 32-bit signed integer value.
//
// - error: ErrParseInteger if the literal format is invalid, or ErrOutOfBoundsInt if the value falls outside the
// permitted range.
func ParseInt(s string) (Int, error) {
	// Implementation Note: Signed integer parsing delegation
	// This function delegates parsing to the parseSigned generic utility to enforce bit-width boundaries and return the
	// appropriate error if a range violation occurs.
	return parseSigned[Int](s, bitSize32, ErrOutOfBoundsInt)
}

// String returns the canonical lexical representation of the Int value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.17.2)
//
// Returns:
//   - string: The canonical decimal string representation of the integer.
func (i Int) String() string {
	return strconv.FormatInt(int64(i), 10)
}

// IsIdenticalWith checks if this Int is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.17)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (i Int) IsIdenticalWith(other XSDValue) bool {
	return isIdentical(i, other)
}

// Compare evaluates the order relation of this Int against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.17)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - int: Negative if less, zero if equal, positive if greater.
//   - error: An error if the types are incomparable.
func (i Int) Compare(other XSDValue) (int, error) {
	return compareSigned(i, other)
}

// ToDecimal converts the Int value to its equivalent Decimal representation.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.17)
//
// Returns:
//   - Decimal: The arbitrary-precision decimal representation of the integer.
func (i Int) ToDecimal() Decimal {
	return NewIntegerFromInt64(int64(i)).ToDecimal()
}
