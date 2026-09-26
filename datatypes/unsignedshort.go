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

// ErrOutOfBoundsUnsignedShort represents the error returned when a parsed value
// falls outside the permitted value space of the xsd:unsignedShort datatype.
var ErrOutOfBoundsUnsignedShort = errors.New("value out of bounds for xsd:unsignedShort (0 to 65535)")

// UnsignedShort represents the unsignedShort datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.23)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A 16-bit unsigned integer value mapped from the value space of nonNegativeInteger,
// representing mathematical integers in the range of 0 to 65535 inclusive.
type UnsignedShort uint16

// ParseUnsignedShort parses a string literal matching the lexical representation of xsd:unsignedShort.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.23.1)
//
// Parameters:
//   - s: The string literal representing the unsigned short to be parsed.
//
// Returns:
//   - UnsignedShort: The parsed UnsignedShort value.
//
// - error: An ErrOutOfBoundsUnsignedShort if the value is out of bounds, or ErrParseInteger if the lexical format is
// invalid.
func ParseUnsignedShort(s string) (UnsignedShort, error) {
	return parseUnsigned[UnsignedShort](s, bitSize16, ErrOutOfBoundsUnsignedShort)
}

// String returns the canonical lexical representation of the UnsignedShort value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.23.2)
//
// Returns:
//   - string: The canonical string representation of the unsigned short value.
func (u UnsignedShort) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

// IsIdenticalWith checks if this UnsignedShort is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (u UnsignedShort) IsIdenticalWith(other XSDValue) bool {
	return isIdentical(u, other)
}

// Compare evaluates the order relation of this UnsignedShort against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.3)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An error if the types are incomparable.
func (u UnsignedShort) Compare(other XSDValue) (int, error) {
	return compareUnsigned(u, other)
}

// ToDecimal projects the underlying value space of UnsignedShort into an arbitrary-precision Decimal value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.1)
//
// Returns:
//   - Decimal: The projected Decimal representation.
func (u UnsignedShort) ToDecimal() Decimal {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.1)
	// Integer types share a value space that is a subset of the decimal value space, which permits conversions and
	// comparisons using their decimal equivalents in accordance with W3C RDF 1.2 (Section 5.1).
	r := new(big.Rat).SetInt(new(big.Int).SetUint64(uint64(u)))
	return Decimal{value: r}
}
