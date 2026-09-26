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

// ErrOutOfBoundsLong represents the error returned when a parsed integer value
// falls outside the permitted value space of the xsd:long datatype.
var ErrOutOfBoundsLong = errors.New("value out of bounds for xsd:long (-9223372036854775808 to 9223372036854775807)")

// Long represents the xsd:long datatype as defined in the governing specifications.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.16)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// The value space is the set of common 64-bit signed integers, bounded by a
// minInclusive value of -9223372036854775808 and a maxInclusive value of
// 9223372036854775807. It is derived from xsd:integer by restriction.
type Long int64

// ParseLong parses a string literal matching the lexical representation of xsd:long.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.16.1)
//
// Parameters:
//   - s: The raw string literal to be parsed.
//
// Returns:
//   - Long: The parsed 64-bit signed integer value.
//
// - error: ErrParseInteger under failure to conform to lexical rules, or ErrOutOfBoundsLong if the value falls outside
// the valid 64-bit signed range.
func ParseLong(s string) (Long, error) {
	return parseSigned[Long](s, bitSize64, ErrOutOfBoundsLong)
}

// String returns the canonical lexical representation of the Long value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.16.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical decimal string representation.
func (l Long) String() string {
	return strconv.FormatInt(int64(l), 10)
}

// IsIdenticalWith checks if this Long value is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.16)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - bool: True if the values are identical, or false otherwise.
func (l Long) IsIdenticalWith(other XSDValue) bool {
	return isIdentical(l, other)
}

// Compare evaluates the order relation of this Long against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.16)
//
// Parameters:
//   - other: The other XSDValue to evaluate.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An error if the types are incomparable.
func (l Long) Compare(other XSDValue) (int, error) {
	return compareSigned(l, other)
}

// ToDecimal converts this Long value into its equivalent arbitrary-precision Decimal representation.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.16)
//
// Parameters:
//
// Returns:
//   - Decimal: The arbitrary-precision decimal representation of the long.
func (l Long) ToDecimal() Decimal {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.16)
	// The xsd:long datatype is derived from xsd:integer, which is transitively derived
	// from xsd:decimal. Therefore, any value within the long value space maps directly to
	// a corresponding decimal value representation in the shared value space.
	return NewIntegerFromInt64(int64(l)).ToDecimal()
}
