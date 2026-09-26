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
)

// Boolean represents the boolean primitive datatype as defined in the governing specifications.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.2)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// This type represents the two-valued mathematical system of binary logic, containing
// the value space of true and false. The lexical space consists of "true", "false",
// "1", and "0" with a fixed whiteSpace normalization behavior of 'collapse'.
type Boolean bool

// ParseBoolean parses a string literal matching the lexical representation of xsd:boolean.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.2.2)
//
// Parameters:
//   - s: A string literal representing a boolean value ("true", "false", "1", or "0").
//
// Returns:
//   - Boolean: The successfully parsed Boolean value.
//   - error: An error if the string literal does not conform to the permitted lexical space.
func ParseBoolean(s string) (Boolean, error) {
	switch s {
	case "true", "1":
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.2.2)
		// The literals "true" and "1" map to the mathematical value true.
		return true, nil
	case "false", "0":
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.2.2)
		// The literals "false" and "0" map to the mathematical value false.
		return false, nil
	default:
		return false, errors.New("invalid boolean lexical representation")
	}
}

// String returns the canonical lexical representation of the Boolean value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.4)
//
// Returns:
//   - string: The canonical string representation ("true" or "false") of the Boolean.
func (b Boolean) String() string {
	if b {
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix E.4)
		// The canonical representation for the true value is "true".
		return "true"
	}
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix E.4)
	// The canonical representation for the false value is "false".
	return "false"
}

// IsIdenticalWith checks if this Boolean is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if the other value is identical to this Boolean, false otherwise.
func (b Boolean) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Boolean); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 2.2.1)
		// Two boolean values are identical if and only if they represent the same element in the logical value space.
		return b == o
	}
	return false
}
