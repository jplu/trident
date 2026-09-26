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

// XSDValue represents the value space element as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.1)
// W3C RDF 1.2 (Section 3.4)
//
// Representation:
// A common abstraction for all values within the XSD value space. It provides
// an interface for obtaining the canonical lexical representation and evaluating
// identity between values in their respective value spaces.
type XSDValue interface {
	// String returns the canonical lexical representation of the value.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Appendix E)
	//
	// Returns:
	//   - string: The canonical lexical representation of the value as defined by the datatype's canonical mapping.
	String() string

	// IsIdenticalWith checks if this value is identical to another value in its value space.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 2.3.1)
	//
	// Parameters:
	//   - other: The other XSD value to compare for identity.
	//
	// Returns:
	//   - bool: True if both values are identical within their shared value space, false otherwise.
	IsIdenticalWith(other XSDValue) bool
}
