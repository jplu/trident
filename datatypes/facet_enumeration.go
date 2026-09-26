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

// FacetEnumeration represents the enumeration constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.5)
//
// Representation:
// It constrains the value space of a datatype to a specified set of values.
// A value satisfies the constraint if it is equal to one of the values
// in the enumeration's value space.
type FacetEnumeration struct {
	// BaseFacet provides the shared facet metadata, such as Name and Fixed status.
	BaseFacet

	// Values represents the set of allowed values defined in the enumeration facet.
	Values []XSDValue
}

// NewFacetEnumeration constructs and returns a new FacetEnumeration instance containing the provided slice of allowed
// values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.5)
//
// Parameters:
//   - vals: The slice of allowed XSDValue elements representing the permitted values.
//
// Returns:
//   - FacetEnumeration: The instantiated enumeration constraining facet.
func NewFacetEnumeration(vals []XSDValue) FacetEnumeration {
	return FacetEnumeration{
		BaseFacet: BaseFacet{
			Name:  NameEnumeration,
			Fixed: false,
		},
		Values: vals,
	}
}

// Check validates whether the given XSDValue satisfies the enumeration constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.5.4)
//
// Parameters:
//   - val: The XSDValue to validate against the enumeration's allowed values.
//
// Returns:
//   - error: An error if the value is not equal to any of the configured allowed values, or nil on success.
func (f FacetEnumeration) Check(val XSDValue) error {
	for _, ev := range f.Values {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.5.4)
		// A value satisfies an enumeration constraint if it is equal or identical to one of the values specified in the
		// value property.

		// Implementation Note: Identity optimization
		// Strict identity is evaluated first as a fast-path optimization.
		if val.IsIdenticalWith(ev) {
			return nil
		}

		// Implementation Note: Fallback value comparison
		// If the identity check fails, value-based comparison is performed to determine equivalence in the value space.
		cmp, err := compareValues(val, ev)
		if err == nil && cmp == 0 {
			return nil
		}
	}
	return fmt.Errorf("enumeration violation: value %s is not in the set of allowed values", val)
}
