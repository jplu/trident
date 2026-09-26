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

// ErrOutOfBoundsUnsignedByte represents the error returned when a parsed integer value falls
// outside the permitted value space of the xsd:unsignedByte datatype.
var ErrOutOfBoundsUnsignedByte = errors.New("value out of bounds for xsd:unsignedByte (0 to 255)")

// UnsignedByte represents the unsignedByte datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.24)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// An 8-bit unsigned integer representing a mathematical integer in the range [0, 255] inclusive,
// derived from unsignedShort by restricting the maxInclusive facet to 255.
type UnsignedByte uint8

// ParseUnsignedByte parses a lexical string representation into an UnsignedByte.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.24.1)
//
// Parameters:
//   - s: The raw string literal to be parsed.
//
// Returns:
//   - UnsignedByte: The parsed 8-bit unsigned integer value.
//
// - error: ErrOutOfBoundsUnsignedByte if the value is out of range, or ErrParseInteger if the literal does not match
// the valid integer lexical rules.
func ParseUnsignedByte(s string) (UnsignedByte, error) {
	return parseUnsigned[UnsignedByte](s, bitSize8, ErrOutOfBoundsUnsignedByte)
}

// String returns the canonical lexical representation of the UnsignedByte value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.24.2)
//
// Returns:
//   - string: The canonical decimal string representation.
func (u UnsignedByte) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

// IsIdenticalWith checks if this UnsignedByte is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.24)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (u UnsignedByte) IsIdenticalWith(other XSDValue) bool {
	return isIdentical(u, other)
}

// Compare evaluates the order relation of this UnsignedByte against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.24)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An error if the types are incomparable.
func (u UnsignedByte) Compare(other XSDValue) (int, error) {
	return compareUnsigned(u, other)
}

// ToDecimal converts the UnsignedByte value to its equivalent Decimal representation.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.24)
//
// Returns:
//   - Decimal: The arbitrary-precision Decimal representation.
func (u UnsignedByte) ToDecimal() Decimal {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.24)
	// Since xsd:unsignedByte is transitively derived from xsd:decimal (via the integer
	// hierarchy), converting it to a Decimal allows it to participate in shared decimal
	// mathematical operations and digit-based facet restrictions.
	r := new(big.Rat).SetInt(new(big.Int).SetUint64(uint64(u)))
	return Decimal{value: r}
}
