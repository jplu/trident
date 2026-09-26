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
	"math"
	"strconv"
	"strings"
)

// Double represents the double primitive datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.5)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A 64-bit floating-point type patterned after the IEEE double-precision 64-bit
// floating-point datatype (IEEE 754), containing floating-point numbers,
// positive/negative zero, positive/negative infinity, and Not-a-Number (NaN).
type Double float64

// GetDoubleInfinity returns the positive infinity representation of Double.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.5.1)
//
// Returns:
//   - Double: A Double value representing positive infinity (+INF).
func GetDoubleInfinity() Double {
	return Double(math.Inf(1))
}

// GetDoubleNegInfinity returns the negative infinity representation of Double.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.5.1)
//
// Returns:
//   - Double: A Double value representing negative infinity (-INF).
func GetDoubleNegInfinity() Double {
	return Double(math.Inf(-1))
}

// GetDoubleNaN returns the Not-a-Number (NaN) representation of Double.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.5.1)
//
// Returns:
//   - Double: A Double value representing Not-a-Number (NaN).
func GetDoubleNaN() Double {
	return Double(math.NaN())
}

// IsNaN reports whether the Double value is Not-a-Number (NaN).
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.5.1)
//
// Returns:
//   - bool: True if the Double value is Not-a-Number (NaN), false otherwise.
func (d Double) IsNaN() bool {
	return math.IsNaN(float64(d))
}

// IsIdenticalWith checks if this Double is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.5.1)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if this Double is identical to the other XSDValue based on bitwise comparisons, false otherwise.
func (d Double) IsIdenticalWith(other XSDValue) bool {
	o, ok := other.(Double)
	if !ok {
		return false
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.5.1)
	// NaN values are identical to each other. Positive zero (0.0) and negative zero (-0.0)
	// are distinct and thus not identical. We enforce this identity model via bitwise comparison
	// of standard IEEE 754 representation.
	if d.IsNaN() && o.IsNaN() {
		return true
	}
	return math.Float64bits(float64(d)) == math.Float64bits(float64(o))
}

// Compare evaluates the order relation of this Double against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2 & Section 3.3.5.1)
//
// Parameters:
//   - other: The other XSDValue to evaluate against the receiver.
//
// Returns:
//   - int: Negative one (-1) if less than, zero (0) if equal, one (1) if greater than.
//   - error: An error if the types are incomparable, or if either value is NaN.
func (d Double) Compare(other XSDValue) (int, error) {
	o, ok := other.(Double)
	if !ok {
		return 0, errors.New("incomparable types")
	}
	v1, v2 := float64(d), float64(o)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.5.1)
	// NaN is incomparable with any value in the value space, including itself.
	if math.IsNaN(v1) || math.IsNaN(v2) {
		return 0, errors.New("NaN is incomparable")
	}
	if v1 < v2 {
		return -1, nil
	}
	if v1 > v2 {
		return 1, nil
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.5.1)
	// Positive zero (0) and negative zero (-0) are equal for comparison purposes
	// and bounds checking facets.
	return 0, nil
}

// ParseDouble parses a string literal matching the lexical representation of xsd:double.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.5.1)
//
// Parameters:
//   - s: The raw string literal to be parsed under whitespace collapse rules.
//
// Returns:
//   - Double: The parsed Double floating-point value.
//   - error: An error if the string is not a valid representation of a double-precision float.
func ParseDouble(s string) (Double, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.5.1)
	// Special literal cases must be explicitly captured. Note that "+INF" is
	// allowed in XSD 1.1, whereas only "INF" and "-INF" were allowed in XSD 1.0.
	switch s {
	case strINF, strPosINF:
		return GetDoubleInfinity(), nil
	case strNegINF:
		return GetDoubleNegInfinity(), nil
	case strNaN:
		return GetDoubleNaN(), nil
	}

	// Implementation Note: rejection case 1
	// Reject Go-specific hexadecimal float syntax and underscores which are invalid in XSD.
	if strings.ContainsAny(s, "xXpP_") {
		return 0, errors.New("invalid double lexical representation")
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.5)
			// Values that are too large or too small to be represented map to infinity or zero.
			return Double(val), nil
		}
		return 0, err
	}

	// Implementation Note: rejection case 2
	// Catch other Go-specific valid strings like "Infinity" or "nan".
	if math.IsInf(val, 0) || math.IsNaN(val) {
		return 0, errors.New("invalid double lexical representation")
	}

	return Double(val), nil
}

// String returns the canonical lexical representation of the Double value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.2)
//
// Returns:
//   - string: The canonical string representation of the Double value.
func (d Double) String() string {
	return canonicalFloatString(float64(d), false)
}
