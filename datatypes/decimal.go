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
	"strings"
)

var (
	// ErrParseDecimalEnd is returned when the parser encounters an unexpected end
	// of a string while parsing a decimal representation.
	ErrParseDecimalEnd = errors.New("parsing decimal: unexpected end of string")

	// ErrParseDecimalBadChar is returned when an invalid character (such as letters,
	// multiple signs, or multiple decimal points) is encountered during decimal parsing.
	ErrParseDecimalBadChar = errors.New("parsing decimal: unexpected character")
)

// Decimal represents the decimal primitive datatype as defined in the governing specifications.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// The value space of decimal is the set of all real numbers that can be represented
// as decimal fractions (i.e., numbers of the form n * 10^-k, where n is a signed
// arbitrary-precision integer and k is a non-negative integer). This implementation
// supports arbitrary-precision decimals by wrapping math/big.Rat, ensuring alignment
// with both XML Schema 1.1 and RDF 1.2.
type Decimal struct {
	value *big.Rat
}

// NewDecimal constructs a Decimal equivalent to i / 10^n.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//   - i: The numerator of the fraction represented as a pointer to big.Int.
//   - n: The scale factor exponent representing the power of ten divisor.
//
// Returns:
//   - Decimal: The constructed arbitrary-precision Decimal instance.
func NewDecimal(i *big.Int, n uint32) Decimal {
	denom := new(big.Int).Exp(big.NewInt(base10), big.NewInt(int64(n)), nil)
	r := new(big.Rat).SetFrac(i, denom)
	return Decimal{value: r}
}

// NewDecimalFromInt64 creates a Decimal value from a native 64-bit signed integer.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//   - v: The 64-bit signed integer value to be converted.
//
// Returns:
//   - Decimal: The resulting arbitrary-precision Decimal instance.
func NewDecimalFromInt64(v int64) Decimal {
	r := new(big.Rat).SetInt64(v)
	return Decimal{value: r}
}

// getValue retrieves the underlying arbitrary-precision big.Rat value.
//
// Parameters:
//
// Returns:
//   - *big.Rat: The underlying pointer to the math/big.Rat structure.
func (d Decimal) getValue() *big.Rat {
	// Implementation Note: Safe initialization of uninitialized structures
	// If the internal value is nil, it returns a rational representing zero (0/1) to prevent nil-pointer dereferences.
	if d.value == nil {
		return big.NewRat(0, 1)
	}
	return d.value
}

// DefaultDecimal returns a zero-valued Decimal instance.
//
// Parameters:
//
// Returns:
//   - Decimal: The default initialized Decimal representing the mathematical value zero.
func DefaultDecimal() Decimal {
	return Decimal{value: big.NewRat(0, 1)}
}

// Add returns the mathematical sum of the receiver and the right-hand side Decimal.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//   - rhs: The right-hand side Decimal value to add.
//
// Returns:
//   - Decimal: The resulting sum of the addition as an arbitrary-precision Decimal.
func (d Decimal) Add(rhs Decimal) Decimal {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.3)
	// In accordance with XSD 1.1 and RDF 1.2, arbitrary-precision decimal addition does not overflow.
	res := new(big.Rat).Add(d.getValue(), rhs.getValue())
	return Decimal{value: res}
}

// Sub returns the mathematical difference of the receiver minus the right-hand side Decimal.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//   - rhs: The right-hand side Decimal value to subtract.
//
// Returns:
//   - Decimal: The resulting difference of the subtraction as an arbitrary-precision Decimal.
func (d Decimal) Sub(rhs Decimal) Decimal {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.3)
	// In accordance with XSD 1.1 and RDF 1.2, arbitrary-precision decimal subtraction does not overflow.
	res := new(big.Rat).Sub(d.getValue(), rhs.getValue())
	return Decimal{value: res}
}

// Mul returns the mathematical product of the receiver and the right-hand side Decimal.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//   - rhs: The right-hand side Decimal value to multiply.
//
// Returns:
//   - Decimal: The resulting product of the multiplication as an arbitrary-precision Decimal.
func (d Decimal) Mul(rhs Decimal) Decimal {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.3)
	// In accordance with XSD 1.1 and RDF 1.2, arbitrary-precision decimal multiplication does not overflow.
	res := new(big.Rat).Mul(d.getValue(), rhs.getValue())
	return Decimal{value: res}
}

// CheckedDiv returns the mathematical quotient of the receiver divided by the right-hand side Decimal.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//   - rhs: The right-hand side Decimal divisor value.
//
// Returns:
//   - Decimal: The resulting quotient of the division as an arbitrary-precision Decimal.
//   - error: An error if division by zero is attempted.
func (d Decimal) CheckedDiv(rhs Decimal) (Decimal, error) {
	if rhs.getValue().Sign() == 0 {
		return Decimal{}, errors.New("division by zero")
	}
	res := new(big.Rat).Quo(d.getValue(), rhs.getValue())
	return Decimal{value: res}, nil
}

// CheckedRemEuclid calculates the Euclidean remainder mathematically.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//   - rhs: The right-hand side Decimal divisor value.
//
// Returns:
//   - Decimal: The resulting Euclidean remainder as an arbitrary-precision Decimal.
//   - error: An error if division by zero is attempted.
func (d Decimal) CheckedRemEuclid(rhs Decimal) (Decimal, error) {
	if rhs.getValue().Sign() == 0 {
		return Decimal{}, errors.New("division by zero")
	}

	a := d.getValue()
	b := rhs.getValue()

	qRat := new(big.Rat).Quo(a, b)

	num := qRat.Num()
	den := qRat.Denom()
	qInt := new(big.Int).Quo(num, den)

	rem := new(big.Int).Rem(num, den)
	if num.Sign() < 0 && rem.Sign() != 0 {
		qInt.Sub(qInt, big.NewInt(1))
	}

	// Implementation Note: Euclidean Remainder calculation details
	// This calculation uses the mathematical formula: r = a - b * floor(a/b).
	sub := new(big.Rat).Mul(b, new(big.Rat).SetInt(qInt))
	res := new(big.Rat).Sub(a, sub)

	return Decimal{value: res}, nil
}

// Neg returns the mathematical negation of the Decimal value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//
// Returns:
//   - Decimal: The mathematically negated Decimal value.
func (d Decimal) Neg() Decimal {
	res := new(big.Rat).Neg(d.getValue())
	return Decimal{value: res}
}

// Abs returns the absolute mathematical value of the Decimal.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//
// Returns:
//   - Decimal: The absolute value of the Decimal.
func (d Decimal) Abs() Decimal {
	res := new(big.Rat).Abs(d.getValue())
	return Decimal{value: res}
}

// IsNegative reports whether the Decimal value is strictly less than zero.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//
// Returns:
//   - bool: True if the Decimal value is negative, false otherwise.
func (d Decimal) IsNegative() bool {
	return d.getValue().Sign() < 0
}

// IsPositive reports whether the Decimal value is strictly greater than zero.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//
// Returns:
//   - bool: True if the Decimal value is positive, false otherwise.
func (d Decimal) IsPositive() bool {
	return d.getValue().Sign() > 0
}

// IsIdenticalWith checks if this Decimal is identical to another XSDValue according to the identity criteria.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//   - other: The other XSDValue instance to compare.
//
// Returns:
//   - bool: True if the two values are identical, false otherwise.
func (d Decimal) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Decimal); ok {
		return d.getValue().Cmp(o.getValue()) == 0
	}
	return false
}

// Compare evaluates the order relation of this Decimal against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//   - other: The other XSDValue instance to compare with.
//
// Returns:
//   - int: Negative one if less, zero if equal, one if greater.
//   - error: An error if the other value is of an incomparable type.
func (d Decimal) Compare(other XSDValue) (int, error) {
	if o, ok := other.(Decimal); ok {
		return d.getValue().Cmp(o.getValue()), nil
	}
	if dp, ok := other.(DecimalProvider); ok {
		return d.getValue().Cmp(dp.ToDecimal().getValue()), nil
	}
	return 0, errors.New("incomparable types")
}

// ToDecimal projects this Decimal value to satisfy the DecimalProvider interface.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3)
//
// Parameters:
//
// Returns:
//   - Decimal: The identical Decimal instance.
func (d Decimal) ToDecimal() Decimal {
	return d
}

// asI128 truncates the underlying big.Rat value into an arbitrary-precision big.Int.
//
// Parameters:
//
// Returns:
//   - *big.Int: The truncated integer portion of the Decimal.
func (d Decimal) asI128() *big.Int {
	// Implementation Note: Truncation of rational to integer representation
	// This helper truncates fractional parts and is primarily used for precise day, hour, and second extraction in
	// temporal calculations.
	num := new(big.Int).Set(d.getValue().Num())
	denom := d.getValue().Denom()
	return num.Quo(num, denom)
}

// ParseDecimal parses a string literal matching the lexical representation of xsd:decimal.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.3.1)
//
// Parameters:
//   - s: The raw string literal to be parsed.
//
// Returns:
//   - Decimal: The parsed arbitrary-precision Decimal value.
//   - error: ErrParseDecimalEnd if the string is empty, or ErrParseDecimalBadChar under syntax failure.
func ParseDecimal(s string) (Decimal, error) {
	if s == "" {
		return Decimal{}, ErrParseDecimalEnd
	}

	dotCount := 0

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.3.1)
	// Validate characters: allow an optional leading sign, at most one decimal point, and only decimal digits.
	for i, r := range s {
		switch {
		case r == '+' || r == '-':
			if i != 0 {
				return Decimal{}, ErrParseDecimalBadChar
			}
		case r == '.':
			dotCount++
			if dotCount > 1 {
				return Decimal{}, ErrParseDecimalBadChar
			}
		case r < '0' || r > '9':
			return Decimal{}, ErrParseDecimalBadChar
		}
	}

	val, ok := new(big.Rat).SetString(s)
	if !ok {
		return Decimal{}, ErrParseDecimalBadChar
	}

	return Decimal{value: val}, nil
}

// String returns the canonical lexical representation of the Decimal value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.2.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical string representation of the Decimal.
func (d Decimal) String() string {
	val := d.getValue()
	if val.IsInt() {
		return val.Num().String()
	}

	num := val.Num()
	den := val.Denom()

	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix E.2.2)
	// Valid lexical representations of decimal values map strictly to terminating decimal values (represented as
	// n/10^k). We analyze the prime factors of the denominator to determine the exact number of fractional digits
	// needed for canonical rendering.
	tempDen := new(big.Int).Set(den)
	twos := 0
	for tempDen.Bit(0) == 0 {
		twos++
		tempDen.Rsh(tempDen, 1)
	}

	fives := 0
	five := big.NewInt(base5)
	zero := big.NewInt(0)
	mod := new(big.Int)
	quo := new(big.Int)

	for {
		quo.QuoRem(tempDen, five, mod)
		if mod.Cmp(zero) == 0 {
			fives++
			tempDen.Set(quo)
		} else {
			break
		}
	}

	// Implementation Note: Non-terminating decimal fallback
	// If the denominator contains prime factors other than 2 and 5, it cannot be expressed as a terminating decimal
	// (which can result from division in temporal calculations). To conform to canonical output requirements without
	// infinite loops, we fall back to a high-precision float representation.
	if tempDen.Cmp(big.NewInt(1)) != 0 {
		s := val.FloatString(maxFloatPrecision)
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
		return s
	}

	k := twos
	if fives > twos {
		k = fives
	}

	absNum := new(big.Int).Abs(num)
	mult := big.NewInt(1)
	if shift := k - twos; shift > 0 {
		mult.Lsh(mult, uint(shift))
	}
	if k > fives {
		pow5 := new(big.Int).Exp(five, big.NewInt(int64(k-fives)), nil)
		mult.Mul(mult, pow5)
	}

	scaledNum := new(big.Int).Mul(absNum, mult)
	strDigits := scaledNum.String()

	if len(strDigits) <= k {
		strDigits = strings.Repeat("0", k-len(strDigits)+1) + strDigits
	}

	insertPos := len(strDigits) - k
	intPart := strDigits[:insertPos]
	fracPart := strDigits[insertPos:]

	sign := ""
	if num.Sign() < 0 {
		sign = "-"
	}

	return sign + intPart + "." + fracPart
}
