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

// ErrOutOfBoundsShort is returned when a parsed integer falls outside the
// value space defined for the xsd:short datatype.
var ErrOutOfBoundsShort = errors.New("value out of bounds for xsd:short (-32768 to 32767)")

// Short represents the short datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.18)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A 16-bit signed integer value space derived from int by restricting the value space to the range [-32768, 32767]
// inclusive.
type Short int16

// ParseShort parses a string literal matching the lexical representation of xsd:short.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.18.1)
//
// Parameters:
//   - s: The raw input string containing the lexical representation of a short.
//
// Returns:
//   - Short: The parsed 16-bit signed integer value.
//   - error: An ErrOutOfBoundsShort if bounds are violated, or ErrParseInteger if the lexical format is invalid.
func ParseShort(s string) (Short, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.18.1)
	// The lexical representation of short allows a finite sequence of decimal digits with an optional leading sign.
	return parseSigned[Short](s, bitSize16, ErrOutOfBoundsShort)
}

// String returns the canonical lexical representation of the Short value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.18.2)
//
// Returns:
//   - string: The canonical lexical string representation of the short.
func (s Short) String() string {
	return strconv.FormatInt(int64(s), 10)
}

// IsIdenticalWith checks if this Short value is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if both values are identical, false otherwise.
func (s Short) IsIdenticalWith(other XSDValue) bool {
	return isIdentical(s, other)
}

// Compare evaluates the order relation of this Short against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.3)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - int: Negative if this is less than other, zero if equal, positive if greater.
//   - error: An error if the types are incomparable.
func (s Short) Compare(other XSDValue) (int, error) {
	return compareSigned(s, other)
}

// ToDecimal converts the Short value into its equivalent Decimal representation.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.18)
//
// Returns:
//   - Decimal: The arbitrary-precision decimal representation of the Short value.
func (s Short) ToDecimal() Decimal {
	return NewIntegerFromInt64(int64(s)).ToDecimal()
}
