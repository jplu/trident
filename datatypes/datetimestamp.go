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

import "errors"

// ErrMissingTimezone represents the error returned when a dateTimeStamp lexical representation lacks an explicit
// timezone offset.
var ErrMissingTimezone = errors.New("dateTimeStamp requires an explicit timezone offset")

// DateTimeStamp represents the dateTimeStamp ordinary datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.28)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A structured date and time representation wrapping a DateTime value, requiring an explicit timezone offset as
// mandated by the explicitTimezone facet.
type DateTimeStamp struct {
	// DateTime represents the underlying dateTime value.
	DateTime
}

// ParseDateTimeStamp parses a string literal matching the lexical representation of xsd:dateTimeStamp.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.28.1)
//
// Parameters:
//   - s: The raw string literal representing the dateTimeStamp.
//
// Returns:
//   - DateTimeStamp: The successfully parsed DateTimeStamp value.
//
// - error: An error if the input does not conform to the lexical rules, or if the explicit timezone offset is missing.
func ParseDateTimeStamp(s string) (DateTimeStamp, error) {
	dt, err := ParseDateTime(s)
	if err != nil {
		return DateTimeStamp{}, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.28)
	// The dateTimeStamp datatype is derived from dateTime by restricting the explicitTimezone facet to 'required'.
	if dt.TimezoneOffset() == nil {
		return DateTimeStamp{}, ErrMissingTimezone
	}
	return DateTimeStamp{DateTime: dt}, nil
}

// String returns the canonical lexical representation of the DateTimeStamp value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.28.1)
//
// Returns:
//   - string: The canonical string representation always containing a timezone offset indicator.
func (d DateTimeStamp) String() string {
	return d.DateTime.String()
}

// IsIdenticalWith checks if this DateTimeStamp is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.28)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if both values are identical, false otherwise.
func (d DateTimeStamp) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(DateTimeStamp); ok {
		return d.DateTime.IsIdenticalWith(o.DateTime)
	}
	return false
}

// Compare evaluates the order relation of this DateTimeStamp against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.28)
//
// Parameters:
//   - other: The other XSDValue to evaluate.
//
// Returns:
//   - int: An integer indicating the relative order (-1 if less, 0 if equal, 1 if greater).
//   - error: An error under failure or if the types are incomparable.
func (d DateTimeStamp) Compare(other XSDValue) (int, error) {
	if o, ok := other.(DateTimeStamp); ok {
		return d.DateTime.Compare(o.DateTime)
	}
	return 0, ErrDateTimeOverflow
}
