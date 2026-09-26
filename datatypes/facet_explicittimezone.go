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
)

// ExplicitTimezoneValue represents the permitted lexical values for the explicitTimezone constraining facet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.14)
//
// Representation:
// A string-based representation of the allowed constraint configurations: "required", "prohibited", or "optional".
type ExplicitTimezoneValue string

const (
	// TimezoneRequired specifies that the timezone offset component is mandatory.
	TimezoneRequired ExplicitTimezoneValue = "required"

	// TimezoneProhibited specifies that the timezone offset component must not be present.
	TimezoneProhibited ExplicitTimezoneValue = "prohibited"

	// TimezoneOptional specifies that the timezone offset component is permitted but not required.
	TimezoneOptional ExplicitTimezoneValue = "optional"
)

// FacetExplicitTimezone represents the explicitTimezone constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.14)
//
// Representation:
// A struct wrapping the baseline facet metadata and the configured ExplicitTimezoneValue constraint.
type FacetExplicitTimezone struct {
	// BaseFacet provides the common structural metadata for this facet.
	BaseFacet

	// Value holds the explicit timezone constraint configuration.
	Value ExplicitTimezoneValue
}

// NewFacetExplicitTimezone constructs a new FacetExplicitTimezone instance.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.14)
//
// Parameters:
//   - val: The ExplicitTimezoneValue configuration ("required", "prohibited", or "optional").
//   - fixed: A boolean indicating whether further derived types are prohibited from overriding this facet.
//
// Returns:
//   - FacetExplicitTimezone: An initialized FacetExplicitTimezone configuration.
func NewFacetExplicitTimezone(val ExplicitTimezoneValue, fixed bool) FacetExplicitTimezone {
	return FacetExplicitTimezone{
		BaseFacet: BaseFacet{Name: NameExplicitTimezone, Fixed: fixed},
		Value:     val,
	}
}

// Check validates whether the given XSDValue satisfies the explicitTimezone constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.14.3)
//
// Parameters:
//   - val: The XSDValue to be evaluated against the explicitTimezone facet rules.
//
// Returns:
// - error: An error if the value violates the explicitTimezone constraint, or if it does not support timezone
// verification.
func (f FacetExplicitTimezone) Check(val XSDValue) error {
	tp, ok := val.(TimezoneProvider)
	if !ok {
		return fmt.Errorf("datatype %T does not support explicitTimezone facet", val)
	}

	tz := tp.TimezoneOffset()

	switch f.Value {
	case TimezoneRequired:
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.14.3)
		// If explicitTimezone has the value 'required', the value space of the restricted datatype is limited to those
		// values with a timezone offset.
		if tz == nil {
			return errors.New("explicitTimezone violation: timezone is required but was absent")
		}
	case TimezoneProhibited:
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.14.3)
		// If explicitTimezone has the value 'prohibited', the value space of the restricted datatype is limited to
		// those values without a timezone offset.
		if tz != nil {
			return errors.New("explicitTimezone violation: timezone is prohibited but was present")
		}
	case TimezoneOptional:
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.14.3)
		// If explicitTimezone has the value 'optional', no restrictions are imposed.
	}

	return nil
}
