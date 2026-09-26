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

// AnyAtomicType represents the anyAtomicType datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.2)
//
// Representation:
// A structure wrapping a raw lexical string value that serves as the common base
// type for all primitive atomic datatypes, acting as a union of their value spaces.
type AnyAtomicType struct {
	// Value holds the raw lexical string value of the atomic type.
	Value string
}

// String returns the canonical lexical representation of the AnyAtomicType value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.2)
//
// Returns:
//   - string: The raw lexical string representing the atomic value.
func (a AnyAtomicType) String() string {
	return a.Value
}

// IsIdenticalWith checks if this AnyAtomicType is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.2)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (a AnyAtomicType) IsIdenticalWith(other XSDValue) bool {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.2)
	// Two anyAtomicType values are identical if and only if their
	// underlying string representations are exactly equal.
	if o, ok := other.(AnyAtomicType); ok {
		return a.Value == o.Value
	}
	return false
}
