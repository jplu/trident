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

// GDay represents the gDay primitive datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.13)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A recurring Gregorian calendar day of the month (1 to 31), independent of any month
// or year, modeled internally using a timeline coordinate and an optional timezone offset.
type GDay struct {
	// Timestamp represents the internal temporal model of the GDay value,
	// mapping it to a specific timeline coordinate using the 7-property model.
	timestamp Timestamp
}

// NewGDay constructs a new GDay instance for the specified day and optional timezone offset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.13)
//
// Parameters:
//   - day: The day of the month represented as a uint8.
//   - timezone: The optional timezone offset from UTC.
//
// Returns:
//   - GDay: The instantiated GDay value.
//   - error: An error if the temporal components violate the Gregorian calendar constraints.
func NewGDay(day uint8, timezone *TimezoneOffset) (GDay, error) {
	ts, err := newTimestamp(&DateTimeSevenPropertyModel{Day: &day, TimezoneOffset: timezone})
	if err != nil {
		return GDay{}, err
	}
	return GDay{timestamp: ts}, nil
}

// ParseGDay parses a string literal matching the lexical representation of xsd:gDay.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.13.2)
//
// Parameters:
//   - input: The raw string literal to be parsed.
//
// Returns:
//   - GDay: The parsed GDay value on success.
//   - error: An error of type ParseDateTimeError if parsing fails.
func ParseGDay(input string) (GDay, error) {
	// Step 1: Parse the triple hyphen prefix.
	// Verify and strip the leading triple hyphen "---" characters.
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.13.2)
	// The lexical representation of gDay must begin with a leading triple hyphen prefix ("---").
	rem, err := expectChar(input, '-', "gDay values must start with '---'")
	if err != nil {
		return GDay{}, err
	}
	rem, err = expectChar(rem, '-', "gDay values must start with '---'")
	if err != nil {
		return GDay{}, err
	}
	rem, err = expectChar(rem, '-', "gDay values must start with '---'")
	if err != nil {
		return GDay{}, err
	}

	// Step 2: Parse the day component.
	// Extract the two-digit day of the month and validate that it falls within the range of 1 to 31.
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.13.2)
	// The day component must be represented as a two-digit integer between 1 and 31.
	day, rem, err := parseTwoDigit(rem, 1, maxDaysInMonth, "day")
	if err != nil {
		return GDay{}, err
	}

	// Step 3: Parse the optional timezone offset.
	// Parse the optional timezone suffix and ensure no unrecognized trailing characters remain in the input string.
	// Implementation Note: Parsing of the optional timezone component.
	// Retrieves and validates the timezone offset component, returning an error if there are unrecognized trailing
	// characters.
	tz, rem, _ := parseTimezone(rem)
	if rem != "" {
		return GDay{}, newParseDateTimeError("unrecognized value suffix")
	}

	return NewGDay(day, tz)
}

// Day returns the day component of the recurring day value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.13.1)
//
// Returns:
//   - uint8: The day of the month ranging from 1 to 31.
func (g GDay) Day() uint8 {
	return g.timestamp.day()
}

// TimezoneOffset returns the optional timezone offset of the GDay value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.13.1)
//
// Returns:
//   - *TimezoneOffset: The associated timezone offset, or nil if the timezone offset is absent.
func (g GDay) TimezoneOffset() *TimezoneOffset {
	return g.timestamp.timezoneOffset
}

// Adjust adjusts the timezone offset of the GDay value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - newTz: The target timezone offset to apply.
//
// Returns:
//   - GDay: The adjusted GDay value with the updated timezone offset.
func (g GDay) Adjust(newTz *TimezoneOffset) GDay {
	ts := g.timestamp.adjust(newTz)
	return GDay{timestamp: ts}
}

// IsIdenticalWith checks if this GDay value is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.13.4)
//
// Parameters:
//   - other: The other XSDValue to be compared.
//
// Returns:
//   - bool: True if both values are identical, false otherwise.
func (g GDay) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(GDay); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.13.4)
		// Two gDay values are identical if and only if their day properties are equal
		// and their timezoneOffset properties are either both absent, or both present and equal.
		return g.timestamp.value.IsIdenticalWith(o.timestamp.value) &&
			((g.timestamp.timezoneOffset == nil && o.timestamp.timezoneOffset == nil) ||
				(g.timestamp.timezoneOffset != nil && o.timestamp.timezoneOffset != nil &&
					*g.timestamp.timezoneOffset == *o.timestamp.timezoneOffset))
	}
	return false
}

// Compare evaluates the ordering relation of this GDay value against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - int: Negative if less, zero if equal, positive if greater.
//   - error: An error if the compared values are incomparable.
func (g GDay) Compare(other XSDValue) (int, error) {
	if o, ok := other.(GDay); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2)
		// Comparison is performed by projecting both values onto a timeline.
		// If one value has an explicit timezone and the other does not, the
		// comparison can be indeterminate within the +/- 14 hour window, returning an error.
		res, det := g.timestamp.compare(o.timestamp)
		if !det {
			return 0, ErrDateTimeOverflow
		}
		return res, nil
	}
	return 0, ErrDateTimeOverflow
}

// String returns the canonical lexical representation of the GDay value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.13.2)
//
// Returns:
//   - string: The canonical lexical representation of the GDay value.
func (g GDay) String() string {
	var tzStr string
	if g.TimezoneOffset() != nil {
		tzStr = g.TimezoneOffset().String()
	}
	return fmt.Sprintf("---%02d%s", g.Day(), tzStr)
}
