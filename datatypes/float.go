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
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Float represents the float primitive datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.4)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// Single-precision 32-bit floating-point numbers conforming to the IEEE 754 standard,
// including positive and negative infinity, and a Not-a-Number (NaN) value.
type Float float32

// GetFloatInfinity returns the positive infinity representation of Float.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.4.1)
//
// Parameters:
//
// Returns:
//   - Float: The positive infinity Float representation.
func GetFloatInfinity() Float {
	return Float(math.Inf(1))
}

// GetFloatNegInfinity returns the negative infinity representation of Float.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.4.1)
//
// Parameters:
//
// Returns:
//   - Float: The negative infinity Float representation.
func GetFloatNegInfinity() Float {
	return Float(math.Inf(-1))
}

// GetFloatNaN returns the Not-a-Number (NaN) representation of Float.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.4.1)
//
// Parameters:
//
// Returns:
//   - Float: The Not-a-Number Float representation.
func GetFloatNaN() Float {
	return Float(math.NaN())
}

// IsNaN reports whether the Float value is Not-a-Number (NaN).
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.4.1)
//
// Parameters:
//
// Returns:
//   - bool: True if the Float value is Not-a-Number, false otherwise.
func (f Float) IsNaN() bool {
	return math.IsNaN(float64(f))
}

// IsIdenticalWith checks if this Float is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.4.1)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if this Float is identical to the other value, false otherwise.
func (f Float) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Float); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.4.1)
		// Positive zero is identical to positive zero, and negative zero is identical to negative zero, but positive
		// zero is not identical to negative zero.
		//
		// Implementation Note: Bitwise identity mapping.
		// Go's standard == operator considers negative and positive zero equal. We use math.Float32bits to perform
		// bitwise comparison, ensuring they are distinguished as non-identical in accordance with the specification.
		return math.Float32bits(float32(f)) == math.Float32bits(float32(o))
	}
	return false
}

// Compare evaluates the order relation of this Float against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.4.1)
//
// Parameters:
//   - other: The other XSDValue to compare with this Float.
//
// Returns:
//   - int: Negative one if less, zero if equal, one if greater.
//   - error: An error if the types are incomparable, or if one of the values is NaN.
func (f Float) Compare(other XSDValue) (int, error) {
	o, ok := other.(Float)
	if !ok {
		return 0, errors.New("incomparable types")
	}
	v1, v2 := float64(f), float64(o)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.4.1)
	// The value space of float is partially ordered; Not-a-Number (NaN) is incomparable with any other float value,
	// including itself.
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return 0, errors.New("NaN is incomparable")
	}
	if v1 < v2 {
		return -1, nil
	}
	if v1 > v2 {
		return 1, nil
	}
	return 0, nil
}

// ParseFloat parses a string literal matching the lexical representation of xsd:float.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.4.2)
//
// Parameters:
//   - s: The lexical string representation to parse.
//
// Returns:
//   - Float: The parsed Float value.
//   - error: An error if the string is not a valid float literal.
func ParseFloat(s string) (Float, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.4.2)
	// Specific literals "INF", "+INF", "-INF", and "NaN" must be explicitly matched to map successfully to special IEEE
	// 754 states.
	switch s {
	case strINF, strPosINF:
		return GetFloatInfinity(), nil
	case strNegINF:
		return GetFloatNegInfinity(), nil
	case strNaN:
		return GetFloatNaN(), nil
	}

	// Implementation Note: rejection case 1
	// Reject Go-specific hexadecimal float syntax and underscores which are invalid in XSD.
	if strings.ContainsAny(s, "xXpP_") {
		return 0, errors.New("invalid float lexical representation")
	}

	val, err := strconv.ParseFloat(s, 32)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.4.1)
			// Values that are too large or too small to be represented map to infinity or zero.
			return Float(val), nil
		}
		return 0, err
	}

	// Implementation Note: rejection case 2
	// Catch other Go-specific valid strings like "Infinity" or "nan".
	if math.IsInf(val, 0) || math.IsNaN(val) {
		return 0, errors.New("invalid float lexical representation")
	}

	return Float(val), nil
}

// String returns the canonical lexical representation of the Float value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical string representation of the Float value.
func (f Float) String() string {
	return canonicalFloatString(float64(f), true)
}

// canonicalFloatString formats a float64 into its canonical representation.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.4.2)
//
// Parameters:
//   - f: The double precision float value to format.
//   - isFloat32: A boolean indicating whether to format with single-precision limits.
//
// Returns:
//   - string: The canonical string representation of the floating-point value.
func canonicalFloatString(f float64, isFloat32 bool) string {
	// Step 1: Special values processing.
	// Intercept infinity and NaN states to return their defined special representations.
	if math.IsInf(f, 1) {
		return strINF
	}
	if math.IsInf(f, -1) {
		return strNegINF
	}
	if math.IsNaN(f) {
		return strNaN
	}

	// Step 2: Zero values processing.
	// Map positive and negative zero to their specific canonical exponential representations.
	//
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.4.2)
	// If the value is positive or negative zero, the canonical representation must be "0.0E0" or "-0.0E0" respectively.
	if f == 0.0 {
		if math.Signbit(f) {
			return "-0.0E0"
		}
		return "0.0E0"
	}

	// Step 3: Base formatting.
	// Format the number using the scientific exponential layout adjusted for precision.
	prec := 64
	if isFloat32 {
		prec = 32
	}

	s := strconv.FormatFloat(f, 'E', -1, prec)
	parts := strings.Split(s, "E")
	mantissa, exponent := parts[0], parts[1]

	// Step 4: Mantissa normalization.
	// Ensure that there is a fractional component present in the final mantissa segment.
	//
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.4.2)
	// The mantissa must contain exactly one non-zero digit before the decimal point, and at least one digit in the
	// fractional part.
	if !strings.Contains(mantissa, ".") {
		mantissa += ".0"
	}

	exponent = strings.TrimPrefix(exponent, "+")

	isNeg := false
	if strings.HasPrefix(exponent, "-") {
		isNeg = true
		exponent = exponent[1:]
	}

	// Step 5: Exponent normalization.
	// Clean the exponent from redundant leading zeros to construct a canonical representation.
	//
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.4.2)
	// The exponent component must be a signed integer without leading zeros.
	exponent = strings.TrimLeft(exponent, "0")
	if exponent == "" {
		exponent = "0"
	}
	if isNeg && exponent != "0" {
		exponent = "-" + exponent
	}

	return fmt.Sprintf("%sE%s", mantissa, exponent)
}
