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

// AnySimpleType represents the anySimpleType datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.1)
//
// Representation:
// The value space of anySimpleType is the union of the value spaces of all
// primitive datatypes and all list datatypes, serving as the root of the simple
// type definition hierarchy. The lexical space encompasses all finite-length
// sequences of zero or more characters matching the XML Char production.
type AnySimpleType struct {
	// Value holds the raw lexical string representing the value.
	Value string
}

// String returns the canonical lexical representation of the AnySimpleType value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.1.2)
//
// Parameters:
//
// Returns:
//   - string: The unconstrained raw lexical representation of the value.
func (a AnySimpleType) String() string {
	return a.Value
}

// IsIdenticalWith checks if this AnySimpleType is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.1)
//
// Parameters:
//   - other: The other XSDValue being compared for identity.
//
// Returns:
//   - bool: True if this value is identical to the other value, otherwise false.
func (a AnySimpleType) IsIdenticalWith(other XSDValue) bool {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 2.2.1)
	// Two anySimpleType values are identical if and only if they are the
	// same value in their respective value spaces. Since anySimpleType wraps
	// any simple value, this comparison serves as an identity comparison
	// of their raw lexical values.
	if o, ok := other.(AnySimpleType); ok {
		return a.Value == o.Value
	}
	return false
}
