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

import "fmt"

// compareValues compares two XSD values as defined by their value spaces.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2)
//
// Parameters:
//   - val: The XSD value to be compared.
//   - bound: The boundary value to compare against.
//
// Returns:
//   - int: The comparison result (-1 if val < bound, 0 if equal, 1 if greater).
//   - error: An error if the types cannot be compared.
func compareValues(val, bound XSDValue) (int, error) {
	cp, ok := val.(ComparableProvider)
	if !ok {
		dVal, ok1 := val.(DecimalProvider)
		dBound, ok2 := bound.(DecimalProvider)
		if ok1 && ok2 {
			vd := dVal.ToDecimal()
			bd := dBound.ToDecimal()

			// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.1)
			// Decimal and its derived types (including integers) share a single
			// mathematical value space, meaning they can be compared directly
			// through their decimal numeric representation.
			return vd.getValue().Cmp(bd.getValue()), nil
		}
		return 0, fmt.Errorf("datatype %T does not support bound comparisons", val)
	}
	return cp.Compare(bound)
}

// FacetMinInclusive represents the minInclusive constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.10)
//
// Representation:
// An XSDValue boundary representing the inclusive lower limit of a datatype's value space.
type FacetMinInclusive struct {
	// BaseFacet provides common facet properties such as the name and the 'fixed' flag.
	BaseFacet
	// Value is the lower bound (inclusive) against which values are compared.
	Value XSDValue
}

// NewFacetMinInclusive instantiates a new minInclusive constraining facet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.10)
//
// Parameters:
//   - val: The boundary XSDValue.
//   - fixed: A boolean indicating if the facet's value cannot be overridden in derived types.
//
// Returns:
//   - FacetMinInclusive: The instantiated minInclusive constraining facet.
func NewFacetMinInclusive(val XSDValue, fixed bool) FacetMinInclusive {
	// Implementation Note: Fixed overrides check
	// Under W3C XSD 1.1 Part 2 (Section 2.4.1.2), the 'fixed' parameter controls whether further derived types can
	// override the value of this facet.
	return FacetMinInclusive{BaseFacet: BaseFacet{Name: NameMinInclusive, Fixed: fixed}, Value: val}
}

// Check validates if the given XSDValue meets the minInclusive constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.10.4)
//
// Parameters:
//   - val: The XSDValue to validate against the lower bound.
//
// Returns:
//   - error: An error if validation fails or if comparing values returns an error.
func (f FacetMinInclusive) Check(val XSDValue) error {
	cmp, err := compareValues(val, f.Value)
	if err != nil {
		return err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.10.4)
	// A value must be mathematically greater than or equal to the minInclusive value.
	if cmp < 0 {
		return fmt.Errorf("minInclusive violation: value %s is less than bound %s", val, f.Value)
	}
	return nil
}

// FacetMinExclusive represents the minExclusive constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.9)
//
// Representation:
// An XSDValue boundary representing the exclusive lower limit of a datatype's value space.
type FacetMinExclusive struct {
	// BaseFacet provides common facet properties such as the name and the 'fixed' flag.
	BaseFacet
	// Value is the lower bound (exclusive) against which values are compared.
	Value XSDValue
}

// NewFacetMinExclusive instantiates a new minExclusive constraining facet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.9)
//
// Parameters:
//   - val: The boundary XSDValue.
//   - fixed: A boolean indicating if the facet's value cannot be overridden in derived types.
//
// Returns:
//   - FacetMinExclusive: The instantiated minExclusive constraining facet.
func NewFacetMinExclusive(val XSDValue, fixed bool) FacetMinExclusive {
	// Implementation Note: Fixed overrides check
	// Under W3C XSD 1.1 Part 2 (Section 2.4.1.2), the 'fixed' parameter controls whether further derived types can
	// override the value of this facet.
	return FacetMinExclusive{BaseFacet: BaseFacet{Name: NameMinExclusive, Fixed: fixed}, Value: val}
}

// Check validates if the given XSDValue meets the minExclusive constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.9.4)
//
// Parameters:
//   - val: The XSDValue to validate against the exclusive lower bound.
//
// Returns:
//   - error: An error if validation fails or if comparing values returns an error.
func (f FacetMinExclusive) Check(val XSDValue) error {
	cmp, err := compareValues(val, f.Value)
	if err != nil {
		return err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.9.4)
	// A value must be mathematically strictly greater than the minExclusive value.
	if cmp <= 0 {
		return fmt.Errorf("minExclusive violation: value %s is <= bound %s", val, f.Value)
	}
	return nil
}

// FacetMaxInclusive represents the maxInclusive constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.7)
//
// Representation:
// An XSDValue boundary representing the inclusive upper limit of a datatype's value space.
type FacetMaxInclusive struct {
	// BaseFacet provides common facet properties such as the name and the 'fixed' flag.
	BaseFacet
	// Value is the upper bound (inclusive) against which values are compared.
	Value XSDValue
}

// NewFacetMaxInclusive instantiates a new maxInclusive constraining facet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.7)
//
// Parameters:
//   - val: The boundary XSDValue.
//   - fixed: A boolean indicating if the facet's value cannot be overridden in derived types.
//
// Returns:
//   - FacetMaxInclusive: The instantiated maxInclusive constraining facet.
func NewFacetMaxInclusive(val XSDValue, fixed bool) FacetMaxInclusive {
	// Implementation Note: Fixed overrides check
	// Under W3C XSD 1.1 Part 2 (Section 2.4.1.2), the 'fixed' parameter controls whether further derived types can
	// override the value of this facet.
	return FacetMaxInclusive{BaseFacet: BaseFacet{Name: NameMaxInclusive, Fixed: fixed}, Value: val}
}

// Check validates if the given XSDValue meets the maxInclusive constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.7.4)
//
// Parameters:
//   - val: The XSDValue to validate against the upper bound.
//
// Returns:
//   - error: An error if validation fails or if comparing values returns an error.
func (f FacetMaxInclusive) Check(val XSDValue) error {
	cmp, err := compareValues(val, f.Value)
	if err != nil {
		return err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.7.4)
	// A value must be mathematically less than or equal to the maxInclusive value.
	if cmp > 0 {
		return fmt.Errorf("maxInclusive violation: value %s is greater than bound %s", val, f.Value)
	}
	return nil
}

// FacetMaxExclusive represents the maxExclusive constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.8)
//
// Representation:
// An XSDValue boundary representing the exclusive upper limit of a datatype's value space.
type FacetMaxExclusive struct {
	// BaseFacet provides common facet properties such as the name and the 'fixed' flag.
	BaseFacet
	// Value is the upper bound (exclusive) against which values are compared.
	Value XSDValue
}

// NewFacetMaxExclusive instantiates a new maxExclusive constraining facet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.8)
//
// Parameters:
//   - val: The boundary XSDValue.
//   - fixed: A boolean indicating if the facet's value cannot be overridden in derived types.
//
// Returns:
//   - FacetMaxExclusive: The instantiated maxExclusive constraining facet.
func NewFacetMaxExclusive(val XSDValue, fixed bool) FacetMaxExclusive {
	// Implementation Note: Fixed overrides check
	// Under W3C XSD 1.1 Part 2 (Section 2.4.1.2), the 'fixed' parameter controls whether further derived types can
	// override the value of this facet.
	return FacetMaxExclusive{BaseFacet: BaseFacet{Name: NameMaxExclusive, Fixed: fixed}, Value: val}
}

// Check validates if the given XSDValue meets the maxExclusive constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.8.4)
//
// Parameters:
//   - val: The XSDValue to validate against the exclusive upper bound.
//
// Returns:
//   - error: An error if validation fails or if comparing values returns an error.
func (f FacetMaxExclusive) Check(val XSDValue) error {
	cmp, err := compareValues(val, f.Value)
	if err != nil {
		return err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.8.4)
	// A value must be mathematically strictly less than the maxExclusive value.
	if cmp >= 0 {
		return fmt.Errorf("maxExclusive violation: value %s is >= bound %s", val, f.Value)
	}
	return nil
}
