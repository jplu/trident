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

// FacetLength represents the length constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1)
//
// Representation:
// An integer value representing the required exact length of the value. The unit of length measurement is determined by
// the variety of the datatype (e.g., characters for string-derived types, octets for binary types, or items for list
// types).
type FacetLength struct {
	// BaseFacet provides common metadata properties such as the name and the 'fixed' flag.
	BaseFacet

	// Value is the required exact length as a non-negative integer.
	Value int
}

// NewFacetLength instantiates a new 'length' constraining facet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1)
//
// Parameters:
//   - val: The required exact length as a non-negative integer.
//   - fixed: A boolean indicating whether further derived types can override the value of this facet.
//
// Returns:
//   - FacetLength: The initialized FacetLength instance.
func NewFacetLength(val int, fixed bool) FacetLength {
	return FacetLength{
		BaseFacet: BaseFacet{
			Name:  NameLength,
			Fixed: fixed,
		},
		Value: val,
	}
}

// Check validates that the provided XSDValue satisfies the 'length' facet constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1.4)
//
// Parameters:
//   - val: The XSDValue to be validated against the length constraint.
//
// Returns:
//   - error: An error if the value fails the validation check, or nil if successful.
func (f FacetLength) Check(val XSDValue) error {
	// Implementation Note: Type assertion to LengthProvider
	// The target value must satisfy the LengthProvider interface to retrieve its structured length measurement.
	lp, ok := val.(LengthProvider)
	if !ok {
		return fmt.Errorf("datatype %T does not support length facet", val)
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.1.4)
	// A value satisfies a length constraint if and only if the length of the value is equal to the value of the length
	// facet.
	if lp.Length() != f.Value {
		return fmt.Errorf("length violation: expected %d, got %d", f.Value, lp.Length())
	}
	return nil
}

// FacetMinLength represents the minLength constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.2)
//
// Representation:
// An integer value representing the minimum acceptable length of the value. The unit of length measurement is
// determined by the variety of the datatype (e.g., characters for string-derived types, octets for binary types, or
// items for list types).
type FacetMinLength struct {
	// BaseFacet provides common metadata properties such as the name and the 'fixed' flag.
	BaseFacet

	// Value is the minimum acceptable length as a non-negative integer.
	Value int
}

// NewFacetMinLength instantiates a new 'minLength' constraining facet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.2)
//
// Parameters:
//   - val: The minimum acceptable length as a non-negative integer.
//   - fixed: A boolean indicating whether further derived types can override the value of this facet.
//
// Returns:
//   - FacetMinLength: The initialized FacetMinLength instance.
func NewFacetMinLength(val int, fixed bool) FacetMinLength {
	return FacetMinLength{
		BaseFacet: BaseFacet{
			Name:  NameMinLength,
			Fixed: fixed,
		},
		Value: val,
	}
}

// Check validates that the provided XSDValue satisfies the 'minLength' facet constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.2.4)
//
// Parameters:
//   - val: The XSDValue to be validated against the minLength constraint.
//
// Returns:
//   - error: An error if the value fails the validation check, or nil if successful.
func (f FacetMinLength) Check(val XSDValue) error {
	// Implementation Note: Type assertion to LengthProvider
	// The target value must satisfy the LengthProvider interface to retrieve its structured length measurement.
	lp, ok := val.(LengthProvider)
	if !ok {
		return fmt.Errorf("datatype %T does not support minLength facet", val)
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.2.4)
	// A value satisfies a minLength constraint if and only if the length of the value is greater than or equal to the
	// value of the minLength facet.
	if lp.Length() < f.Value {
		return fmt.Errorf("minLength violation: expected at least %d, got %d", f.Value, lp.Length())
	}
	return nil
}

// FacetMaxLength represents the maxLength constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.3)
//
// Representation:
// An integer value representing the maximum acceptable length of the value. The unit of length measurement is
// determined by the variety of the datatype (e.g., characters for string-derived types, octets for binary types, or
// items for list types).
type FacetMaxLength struct {
	// BaseFacet provides common metadata properties such as the name and the 'fixed' flag.
	BaseFacet

	// Value is the maximum acceptable length as a non-negative integer.
	Value int
}

// NewFacetMaxLength instantiates a new 'maxLength' constraining facet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.3)
//
// Parameters:
//   - val: The maximum acceptable length as a non-negative integer.
//   - fixed: A boolean indicating whether further derived types can override the value of this facet.
//
// Returns:
//   - FacetMaxLength: The initialized FacetMaxLength instance.
func NewFacetMaxLength(val int, fixed bool) FacetMaxLength {
	return FacetMaxLength{
		BaseFacet: BaseFacet{
			Name:  NameMaxLength,
			Fixed: fixed,
		},
		Value: val,
	}
}

// Check validates that the provided XSDValue satisfies the 'maxLength' facet constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.3.4)
//
// Parameters:
//   - val: The XSDValue to be validated against the maxLength constraint.
//
// Returns:
//   - error: An error if the value fails the validation check, or nil if successful.
func (f FacetMaxLength) Check(val XSDValue) error {
	// Implementation Note: Type assertion to LengthProvider
	// The target value must satisfy the LengthProvider interface to retrieve its structured length measurement.
	lp, ok := val.(LengthProvider)
	if !ok {
		return fmt.Errorf("datatype %T does not support maxLength facet", val)
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.3.4)
	// A value satisfies a maxLength constraint if and only if the length of the value is less than or equal to the
	// value of the maxLength facet.
	if lp.Length() > f.Value {
		return fmt.Errorf("maxLength violation: expected at most %d, got %d", f.Value, lp.Length())
	}
	return nil
}
