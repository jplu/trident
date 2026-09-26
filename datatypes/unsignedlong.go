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
	"math/big"
	"strconv"
)

// ErrOutOfBoundsUnsignedLong represents the error returned when a parsed value
// falls outside the permitted value space of the xsd:unsignedLong datatype.
var ErrOutOfBoundsUnsignedLong = errors.New("value out of bounds for xsd:unsignedLong (0 to 18446744073709551615)")

// UnsignedLong represents the xsd:unsignedLong datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.21)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A 64-bit unsigned integer representing a mathematical integer in the range [0, 18446744073709551615] inclusive,
// transitively derived from xsd:decimal.
type UnsignedLong uint64

// ParseUnsignedLong parses a string literal matching the lexical representation of xsd:unsignedLong.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.21.1)
//
// Parameters:
//   - s: A string literal representing a non-negative integer.
//
// Returns:
//   - UnsignedLong: The parsed 64-bit unsigned integer value.
//   - error: An error if the literal does not conform to the lexical space, or falls outside the permitted bounds.
func ParseUnsignedLong(s string) (UnsignedLong, error) {
	return parseUnsigned[UnsignedLong](s, bitSize64, ErrOutOfBoundsUnsignedLong)
}

// String returns the canonical lexical representation of the UnsignedLong value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.21.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical representation of the UnsignedLong value as a base-10 digit string.
func (u UnsignedLong) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

// IsIdenticalWith checks if this UnsignedLong value is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (u UnsignedLong) IsIdenticalWith(other XSDValue) bool {
	return isIdentical(u, other)
}

// Compare evaluates the order relation of this UnsignedLong against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.3)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - int: An integer indicating the order relationship (-1 if less, 0 if equal, 1 if greater).
//   - error: An error if the types are incomparable in the value space.
func (u UnsignedLong) Compare(other XSDValue) (int, error) {
	return compareUnsigned(u, other)
}

// ToDecimal converts the UnsignedLong value into its equivalent arbitrary-precision Decimal representation.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.21)
//
// Parameters:
//
// Returns:
//   - Decimal: The equivalent arbitrary-precision Decimal value.
func (u UnsignedLong) ToDecimal() Decimal {
	// Implementation Note: Conversion of Go uint64 integers to arbitrary-precision rational decimals
	// Since xsd:unsignedLong is derived from xsd:decimal, its value space elements map directly into the decimal value
	// space using math/big.Rat to facilitate exact cross-datatype comparisons and boundary checks.
	r := new(big.Rat).SetInt(new(big.Int).SetUint64(uint64(u)))
	return Decimal{value: r}
}
