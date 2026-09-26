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
)

// ErrParseInteger is returned when a string cannot be successfully parsed
// into an arbitrary-size integer because it does not match the required lexical format.
var ErrParseInteger = errors.New("invalid integer format")

// Integer represents the integer datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// An arbitrary-precision mathematical integer value mapped to the xsd:integer value space,
// wrapping a math/big.Int pointer.
type Integer struct {
	// value holds the underlying arbitrary-precision integer.
	value *big.Int
}

// NewInteger constructs a new Integer value from an existing big integer.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//   - val: The math/big.Int value to copy.
//
// Returns:
//   - Integer: The initialized arbitrary-precision Integer.
func NewInteger(val *big.Int) Integer {
	return Integer{value: new(big.Int).Set(val)}
}

// NewIntegerFromInt64 constructs a new Integer value from a standard 64-bit signed integer.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//   - val: The 64-bit signed integer value.
//
// Returns:
//   - Integer: The initialized arbitrary-precision Integer.
func NewIntegerFromInt64(val int64) Integer {
	return Integer{value: big.NewInt(val)}
}

// DefaultInteger returns an initialized Integer representing the value zero.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//
// Returns:
//   - Integer: An Integer initialized to mathematical zero.
func DefaultInteger() Integer {
	return Integer{value: big.NewInt(0)}
}

// ParseInteger parses a string literal matching the lexical representation of xsd:integer.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13.1)
//
// Parameters:
//   - s: The raw string representation of the integer.
//
// Returns:
//   - Integer: The successfully parsed Integer value.
//   - error: ErrParseInteger if the literal does not conform to the lexical rules.
func ParseInteger(s string) (Integer, error) {
	// Implementation Note: Infinite integer representation using math/big
	// Go strings are parsed into math/big.Int to allow representation of infinite mathematical integers.
	val, ok := new(big.Int).SetString(s, base10)
	if !ok {
		return Integer{}, ErrParseInteger
	}
	return Integer{value: val}, nil
}

// getValue retrieves the underlying big.Int pointer.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//
// Returns:
//   - *big.Int: The raw pointer to big.Int, defaulting to zero if nil.
func (i Integer) getValue() *big.Int {
	if i.value == nil {
		return big.NewInt(0)
	}
	return i.value
}

// Add performs an addition of two Integer values, returning the sum.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//   - rhs: The right-hand side Integer value to add.
//
// Returns:
//   - Integer: The sum of the two integers.
func (i Integer) Add(rhs Integer) Integer {
	return Integer{value: new(big.Int).Add(i.getValue(), rhs.getValue())}
}

// Sub performs a subtraction of two Integer values, returning the difference.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//   - rhs: The right-hand side Integer value to subtract.
//
// Returns:
//   - Integer: The difference of the two integers.
func (i Integer) Sub(rhs Integer) Integer {
	return Integer{value: new(big.Int).Sub(i.getValue(), rhs.getValue())}
}

// Mul performs a multiplication of two Integer values, returning the product.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//   - rhs: The right-hand side Integer value to multiply by.
//
// Returns:
//   - Integer: The product of the two integers.
func (i Integer) Mul(rhs Integer) Integer {
	return Integer{value: new(big.Int).Mul(i.getValue(), rhs.getValue())}
}

// CheckedDiv performs integer division of two Integer values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//   - rhs: The right-hand side Integer divisor value.
//
// Returns:
//   - Integer: The quotient of the integer division.
//   - error: An error if division by zero is attempted.
func (i Integer) CheckedDiv(rhs Integer) (Integer, error) {
	rv := rhs.getValue()
	if rv.Sign() == 0 {
		return Integer{}, errors.New("division by zero")
	}
	return Integer{value: new(big.Int).Quo(i.getValue(), rv)}, nil
}

// CheckedRem calculates the remainder of integer division of two Integer values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//   - rhs: The right-hand side Integer divisor value.
//
// Returns:
//   - Integer: The remainder of the division.
//   - error: An error if division by zero is attempted.
func (i Integer) CheckedRem(rhs Integer) (Integer, error) {
	rv := rhs.getValue()
	if rv.Sign() == 0 {
		return Integer{}, errors.New("division by zero")
	}
	return Integer{value: new(big.Int).Rem(i.getValue(), rv)}, nil
}

// CheckedRemEuclid calculates the Euclidean remainder of two Integer values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//   - rhs: The right-hand side Integer divisor value.
//
// Returns:
//   - Integer: The Euclidean remainder of the division.
//   - error: An error if division by zero is attempted.
func (i Integer) CheckedRemEuclid(rhs Integer) (Integer, error) {
	rv := rhs.getValue()
	if rv.Sign() == 0 {
		return Integer{}, errors.New("division by zero")
	}
	res := new(big.Int).Rem(i.getValue(), rv)
	if res.Sign() < 0 {
		if rv.Sign() > 0 {
			res.Add(res, rv)
		} else {
			res.Sub(res, rv)
		}
	}
	return Integer{value: res}, nil
}

// Neg returns the negated value of the Integer.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//
// Returns:
//   - Integer: The negated Integer value.
func (i Integer) Neg() Integer {
	return Integer{value: new(big.Int).Neg(i.getValue())}
}

// Abs returns the absolute value of the Integer.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//
// Returns:
//   - Integer: The absolute Integer value.
func (i Integer) Abs() Integer {
	return Integer{value: new(big.Int).Abs(i.getValue())}
}

// IsNegative reports whether the Integer value is strictly less than zero.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//
// Returns:
//   - bool: True if the value is negative, false otherwise.
func (i Integer) IsNegative() bool {
	return i.getValue().Sign() < 0
}

// IsPositive reports whether the Integer value is strictly greater than zero.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13)
//
// Parameters:
//
// Returns:
//   - bool: True if the value is positive, false otherwise.
func (i Integer) IsPositive() bool {
	return i.getValue().Sign() > 0
}

// IsIdenticalWith checks if this Integer is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.13.4)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - bool: True if both values are identical in their mathematical value space, false otherwise.
func (i Integer) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Integer); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.13)
		// Two integer values are identical if and only if they represent the exact same mathematical integer value.
		return i.getValue().Cmp(o.getValue()) == 0
	}
	return false
}

// Compare evaluates the order relation of this Integer against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An error if the types are incomparable in the value space.
func (i Integer) Compare(other XSDValue) (int, error) {
	if o, ok := other.(Integer); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2)
		// Order relation of integer values matches the traditional mathematical order.
		return i.getValue().Cmp(o.getValue()), nil
	}
	if dp, ok := other.(DecimalProvider); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.1)
		// Since decimal and integer share a single mathematical value space, cross-type numeric comparison is executed
		// by projecting integers to their decimal representation.
		return i.ToDecimal().getValue().Cmp(dp.ToDecimal().getValue()), nil
	}
	return 0, errors.New("incomparable types")
}

// ToDecimal converts the Integer to an arbitrary-precision Decimal representation.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.1)
// W3C RDF 1.2 (Section 5.1)
//
// Parameters:
//
// Returns:
//   - Decimal: The Decimal representation of the Integer.
func (i Integer) ToDecimal() Decimal {
	// Implementation Note: Type promotion to math/big.Rat
	// The mathematical integer value is mapped directly to a big rational representing i / 1.
	r := new(big.Rat).SetInt(i.getValue())
	return Decimal{value: r}
}

// String returns the canonical lexical representation of the Integer value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical string representation.
func (i Integer) String() string {
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix E.2)
	// The canonical representation of an integer does not permit a leading "+" sign, nor does it permit any leading
	// zeroes, unless the value represented is zero, in which case the canonical representation is "0".
	return i.getValue().String()
}
