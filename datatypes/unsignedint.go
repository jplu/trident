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

// ErrOutOfBoundsUnsignedInt is returned when a parsed integer value falls
// outside the range of the xsd:unsignedInt value space [0, 4294967295].
var ErrOutOfBoundsUnsignedInt = errors.New("value out of bounds for xsd:unsignedInt (0 to 4294967295)")

// UnsignedInt represents the unsignedInt datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.22)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// Mathematical integers in the range of 0 to 4294967295 (inclusive), represented
// as a Go 32-bit unsigned integer. It is derived from unsignedLong by restricting
// the value space.
type UnsignedInt uint32

// ParseUnsignedInt parses a string literal matching the lexical representation of xsd:unsignedInt.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.22.1)
//
// Parameters:
//   - s: The raw string literal to be parsed and validated.
//
// Returns:
//   - UnsignedInt: The parsed 32-bit unsigned integer value.
//   - error: ErrOutOfBoundsUnsignedInt if the value is out of bounds, or ErrParseInteger if the format is invalid.
func ParseUnsignedInt(s string) (UnsignedInt, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.22.1)
	// The value space of unsignedInt is restricted to the closed interval [0, 4294967295].

	// Implementation Note: Delegation to generic parser
	// The generic parseUnsigned utility handles the collapse whiteSpace normalization
	// and validates representation details such as a minus sign with zero values ("-0").
	return parseUnsigned[UnsignedInt](s, bitSize32, ErrOutOfBoundsUnsignedInt)
}

// String returns the canonical lexical representation of the UnsignedInt value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.22.2)
//
// Parameters:
//   - none
//
// Returns:
//   - string: The canonical decimal representation without leading plus signs or zeros.
func (u UnsignedInt) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

// IsIdenticalWith checks if this UnsignedInt is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if both values belong to the same datatype and represent the exact same mathematical integer.
func (u UnsignedInt) IsIdenticalWith(other XSDValue) bool {
	return isIdentical(u, other)
}

// Compare evaluates the order relation of this UnsignedInt against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.3)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An error if the types are incomparable.
func (u UnsignedInt) Compare(other XSDValue) (int, error) {
	return compareUnsigned(u, other)
}

// ToDecimal projects the underlying value space of UnsignedInt into an arbitrary-precision Decimal representation.
//
// Specification Reference:
// W3C RDF 1.2 (Section 5.1)
//
// Parameters:
//   - none
//
// Returns:
//   - Decimal: The arbitrary-precision Decimal representation of the unsigned integer.
func (u UnsignedInt) ToDecimal() Decimal {
	// Spec Rule: W3C RDF 1.2 (Section 5.1)
	// Integer subtypes are treated as subsets of the decimal value space, permitting representation as Decimal.

	// Implementation Note: Conversion to arbitrary precision
	// This mapping converts the 32-bit unsigned integer to an arbitrary-precision rational number.
	r := new(big.Rat).SetInt(new(big.Int).SetUint64(uint64(u)))
	return Decimal{value: r}
}
