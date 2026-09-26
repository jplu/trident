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
	"fmt"
	"strings"
)

// FacetTotalDigits represents the totalDigits constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.11)
//
// Representation:
// A structured constraining facet containing a positive integer limit and a boolean flag
// indicating whether the facet value is fixed. It is used to restrict the magnitude and
// arithmetic precision of decimal values and their derived types in the value space.
type FacetTotalDigits struct {
	// BaseFacet provides the shared metadata properties for this constraining facet.
	BaseFacet
	// Value is the maximum number of allowed digits in the canonical representation.
	Value int
}

// NewFacetTotalDigits instantiates a new FacetTotalDigits with the specified value and fixed status.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.11)
//
// Parameters:
//   - val: The maximum number of allowed decimal digits as a positive integer.
//   - fixed: A boolean flag indicating whether further derived types are prohibited from overriding this facet.
//
// Returns:
//   - FacetTotalDigits: The constructed FacetTotalDigits facet instance.
func NewFacetTotalDigits(val int, fixed bool) FacetTotalDigits {
	return FacetTotalDigits{
		BaseFacet: BaseFacet{
			Name:  NameTotalDigits,
			Fixed: fixed,
		},
		Value: val,
	}
}

// Check asserts that the canonical representation of the mapped value satisfies the totalDigits constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.11.4)
//
// Parameters:
//   - val: The XSDValue instance to be evaluated.
//
// Returns:
//   - error: An error if the digit count exceeds the configured limit, or if the type does not support totalDigits.
func (f FacetTotalDigits) Check(val XSDValue) error {
	// Step 1: Resolve decimal value
	// Convert the input value to an arbitrary-precision Decimal representation.
	dp, ok := val.(DecimalProvider)
	if !ok {
		return fmt.Errorf("datatype %T does not support totalDigits facet", val)
	}
	dec := dp.ToDecimal()

	// Step 2: Retrieve canonical representation
	// Obtain the canonical string representation of the Decimal value.
	s := dec.String()

	// Step 3: Count decimal digits
	// Count the digits in the canonical string, accounting for signs, decimal points, and leading zeros.
	//
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.11.1)
	// The value of totalDigits must be a positiveInteger, constraining the total number of decimal digits in the
	// canonical representation of the value.
	//
	// Implementation Note: Normalizing decimal string for accurate digit counting
	// Signs, decimal points, and leading zeros are stripped from the canonical string representation of the decimal
	// value.
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.TrimLeft(s, "0")
	if s == "" {
		// Implementation Note: Special case handling for zero values
		// A mathematical zero canonicalizes to a string containing "0", representing exactly one digit. If totalDigits
		// is strictly less than one, a zero value violates the facet.
		if f.Value < 1 {
			return fmt.Errorf("totalDigits violation: value 0 exceeds totalDigits limit %d", f.Value)
		}
		return nil
	}

	// Step 4: Validate digit limit
	// Assert that the total digit count does not exceed the configured totalDigits limit.
	if len(s) > f.Value {
		return fmt.Errorf("totalDigits violation: value %s has %d digits, limit is %d", val, len(s), f.Value)
	}
	return nil
}

// FacetFractionDigits represents the fractionDigits constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.12)
//
// Representation:
// A structured constraining facet containing a non-negative integer limit and a boolean flag
// indicating whether the facet value is fixed. It is used to restrict the scale or number of
// fractional digits of decimal values in the value space.
type FacetFractionDigits struct {
	// BaseFacet provides the shared metadata properties for this constraining facet.
	BaseFacet
	// Value is the maximum number of allowed fractional digits.
	Value int
}

// NewFacetFractionDigits instantiates a new FacetFractionDigits with the specified value and fixed status.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.12)
//
// Parameters:
//   - val: The maximum number of allowed fractional digits as a non-negative integer.
//   - fixed: A boolean flag indicating whether further derived types are prohibited from overriding this facet.
//
// Returns:
//   - FacetFractionDigits: The constructed FacetFractionDigits facet instance.
func NewFacetFractionDigits(val int, fixed bool) FacetFractionDigits {
	return FacetFractionDigits{
		BaseFacet: BaseFacet{
			Name:  NameFractionDigits,
			Fixed: fixed,
		},
		Value: val,
	}
}

// Check asserts that the canonical representation of the mapped value satisfies the fractionDigits constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.12.4)
//
// Parameters:
//   - val: The XSDValue instance to be evaluated.
//
// Returns:
// - error: An error if the fractional digit count exceeds the configured limit, or if the type does not support
// fractionDigits.
func (f FacetFractionDigits) Check(val XSDValue) error {
	// Step 1: Resolve decimal value
	// Convert the input value to an arbitrary-precision Decimal representation.
	dp, ok := val.(DecimalProvider)
	if !ok {
		return fmt.Errorf("datatype %T does not support fractionDigits facet", val)
	}
	dec := dp.ToDecimal()

	// Step 2: Retrieve canonical representation
	// Extract the canonical string representation of the Decimal value.
	s := dec.String()

	// Step 3: Isolate fractional part
	// Separate the decimal string by the period character to identify the fractional substring.
	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		return nil
	}

	fracPart := parts[1]

	// Step 4: Validate fractional digit limit
	// Assert that the length of the fractional substring does not exceed the configured fractionDigits limit.
	//
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.12.1)
	// The number of decimal digits in the fractional part of the canonical representation of the value must be less
	// than or equal to the value of the fractionDigits facet.
	if len(fracPart) > f.Value {
		return fmt.Errorf(
			"fractionDigits violation: value %s has %d fraction digits, limit is %d",
			val,
			len(fracPart),
			f.Value,
		)
	}
	return nil
}
