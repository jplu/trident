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

// ErrOutOfBoundsByte represents the error returned when a parsed value falls outside the bounds of the byte value
// space.
var ErrOutOfBoundsByte = errors.New("value out of bounds for xsd:byte (-128 to 127)")

// Byte represents the byte datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.19)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A signed 8-bit integer representing mathematical integers in the range from -128 to 127 inclusive, derived from short
// by restriction.
type Byte int8

// ParseByte parses a string literal matching the lexical representation of xsd:byte.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.19)
//
// Parameters:
//   - s: The raw string literal to be parsed and normalized.
//
// Returns:
//   - Byte: The parsed Byte value if successful.
//
// - error: An ErrOutOfBoundsByte error if the value falls outside the permitted range, or ErrParseInteger if the syntax
// is invalid.
func ParseByte(s string) (Byte, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.19)
	// The value space of byte is strictly restricted to the signed 8-bit range from -128 to 127.
	return parseSigned[Byte](s, bitSize8, ErrOutOfBoundsByte)
}

// String returns the canonical lexical representation of the Byte value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.2)
//
// Returns:
//   - string: The canonical decimal string representation.
func (b Byte) String() string {
	return strconv.FormatInt(int64(b), 10)
}

// IsIdenticalWith checks if this Byte is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.19.4)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (b Byte) IsIdenticalWith(other XSDValue) bool {
	return isIdentical(b, other)
}

// Compare evaluates the order relation of this Byte against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.19.3)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - int: Negative one if less, zero if equal, one if greater.
//   - error: An error if the types are incomparable in their value spaces.
func (b Byte) Compare(other XSDValue) (int, error) {
	return compareSigned(b, other)
}

// ToDecimal converts the Byte value into its equivalent arbitrary-precision Decimal representation.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.19)
// W3C RDF 1.2 (Section 5.1)
//
// Returns:
//   - Decimal: The equivalent arbitrary-precision Decimal representation.
func (b Byte) ToDecimal() Decimal {
	return NewIntegerFromInt64(int64(b)).ToDecimal()
}
