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
	"strings"
)

// xsdSigned represents the signed integer type constraint as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.16 - Section 3.4.19)
//
// Representation:
// A generic type constraint matching the underlying primitive types of signed integer derivative types.
type xsdSigned interface {
	~int8 | ~int16 | ~int32 | ~int64
}

// xsdUnsigned represents the unsigned integer type constraint as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.20 - Section 3.4.24)
//
// Representation:
// A generic type constraint matching the underlying primitive types of unsigned integer derivative types.
type xsdUnsigned interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

// parseSigned parses a string as a base-10 signed integer of a specified bit size.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.16 - Section 3.4.19)
//
// Parameters:
//   - s: The raw string literal to be parsed into a signed integer.
//   - bitSize: The target integer bit-width constraint (8, 16, 32, or 64 bits).
//   - rangeErr: The pre-allocated error returned if the value exceeds the target bounds.
//
// Returns:
//   - T: The parsed signed integer value of type T.
//   - error: An error if the parsing or bit-width validation fails.
func parseSigned[T xsdSigned](s string, bitSize int, rangeErr error) (T, error) {
	// Implementation Note: Signed integer parsing delegation
	// The Go standard library function strconv.ParseInt is used to parse the string with the specified bit-width. It
	// returns a range error under overflow conditions.
	val, err := strconv.ParseInt(s, 10, bitSize)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return 0, rangeErr
		}
		return 0, ErrParseInteger
	}
	return T(val), nil
}

// parseUnsigned parses a string as a base-10 unsigned integer of a specified bit size.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.20 - Section 3.4.24)
//
// Parameters:
//   - s: The raw string literal to be parsed into an unsigned integer.
//   - bitSize: The target integer bit-width constraint (8, 16, 32, or 64 bits).
//   - rangeErr: The pre-allocated error returned if the value exceeds the target bounds.
//
// Returns:
//   - T: The parsed unsigned integer value of type T.
//   - error: An error if the parsing, non-negativity constraint, or bit-width validation fails.
func parseUnsigned[T xsdUnsigned](s string, bitSize int, rangeErr error) (T, error) {
	sClean := strings.TrimPrefix(s, "+")

	if strings.HasPrefix(sClean, "-") {
		// Implementation Note: Zero negative sign validation
		// Check if the remaining characters after '-' are all '0' to comply with XSD rules.
		isZero := true
		for _, ch := range sClean[1:] {
			if ch != '0' {
				isZero = false
				break
			}
		}

		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.20.1)
		// The lexical representation of non-negative integers may include a
		// leading '-' sign if and only if the remaining digits are all '0'.
		if isZero && len(sClean) > 1 {
			return 0, nil
		}
		return 0, rangeErr
	}

	// Implementation Note: Unsigned integer parsing delegation
	// The Go standard library function strconv.ParseUint is used to parse the string after cleaning standard prefixes
	// and checking negative sign constraints.
	val, err := strconv.ParseUint(sClean, 10, bitSize)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return 0, rangeErr
		}
		return 0, ErrParseInteger
	}
	return T(val), nil
}

// isIdentical checks value space identity between two comparable values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.2)
//
// Parameters:
//   - a: The comparable value of type T to be evaluated.
//   - other: The other XSD value to compare against.
//
// Returns:
//   - bool: True if the values correspond to the exact same element in their value space, false otherwise.
func isIdentical[T comparable](a T, other XSDValue) bool {
	if o, ok := other.(T); ok {
		return a == o
	}
	return false
}

// compareSigned compares a signed integer value with another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2)
//
// Parameters:
//   - a: The signed integer value of type T to be compared.
//   - other: The other XSD value to compare against.
//
// Returns:
//   - int: Negative if a is less than other, zero if equal, positive if greater.
//   - error: An error if the types cannot be compared in their value spaces.
func compareSigned[T xsdSigned](a T, other XSDValue) (int, error) {
	if o, ok := other.(T); ok {
		if a < o {
			return -1, nil
		}
		if a > o {
			return 1, nil
		}
		return 0, nil
	}
	if dp, ok := other.(DecimalProvider); ok {
		// Implementation Note: Cross-type decimal mapping for comparison
		// Project the signed integer value to an arbitrary-precision rational number using the ToDecimal mapping. This
		// allows cross-datatype comparison in compliance with W3C RDF 1.2 (Section 5.1).
		return NewIntegerFromInt64(int64(a)).ToDecimal().getValue().Cmp(dp.ToDecimal().getValue()), nil
	}
	return 0, errors.New("incomparable types")
}

// compareUnsigned compares an unsigned integer value with another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2)
//
// Parameters:
//   - u: The unsigned integer value of type T to be compared.
//   - other: The other XSD value to compare against.
//
// Returns:
//   - int: Negative if u is less than other, zero if equal, positive if greater.
//   - error: An error if the types cannot be compared in their value spaces.
func compareUnsigned[T xsdUnsigned](u T, other XSDValue) (int, error) {
	if o, ok := other.(T); ok {
		if u < o {
			return -1, nil
		}
		if u > o {
			return 1, nil
		}
		return 0, nil
	}
	if dp, ok := other.(DecimalProvider); ok {
		// Implementation Note: Unsigned integer conversion to decimal
		// Project the unsigned integer value to an arbitrary-precision rational number using big.Rat. This allows
		// cross-datatype comparison in compliance with W3C RDF 1.2 (Section 5.1).
		r := new(big.Rat).SetInt(new(big.Int).SetUint64(uint64(u)))
		return r.Cmp(dp.ToDecimal().getValue()), nil
	}
	return 0, errors.New("incomparable types")
}
