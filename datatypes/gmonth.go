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

// GMonth represents the gMonth primitive datatype as defined in the governing specifications.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.14)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// The value space of gMonth is a recurring calendar month of the year, represented
// by a month property (1 to 12) and an optional timezoneOffset property. The year,
// day, hour, minute, and second properties are required to be absent.
type GMonth struct {
	// timestamp stores the underlying representation used for canonicalization,
	// identity verification, and value space comparison logic.
	timestamp Timestamp
}

// NewGMonth constructs a new GMonth value from a given month and timezone offset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.14.1)
//
// Parameters:
//   - month: The 1-based month number, which must be between 1 and 12 inclusive.
//   - timezone: An optional pointer to a TimezoneOffset.
//
// Returns:
//   - GMonth: The initialized GMonth struct.
//   - error: An error if the Gregorian calendar fields violate calendar constraints or boundary limits.
func NewGMonth(month uint8, timezone *TimezoneOffset) (GMonth, error) {
	ts, err := newTimestamp(&DateTimeSevenPropertyModel{Month: &month, TimezoneOffset: timezone})
	if err != nil {
		return GMonth{}, err
	}
	return GMonth{timestamp: ts}, nil
}

// ParseGMonth parses a string literal matching the lexical representation of xsd:gMonth.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.14.2)
//
// Parameters:
//   - input: The raw string literal to be parsed, formatted as "--MM" optionally followed by a timezone offset.
//
// Returns:
//   - GMonth: The successfully parsed GMonth value.
//   - error: An error if the input does not conform to the expected lexical format or calendar boundaries.
func ParseGMonth(input string) (GMonth, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.14.2)
	// The lexical representation of a gMonth value must begin with a double hyphen prefix.
	rem, err := expectChar(input, '-', "gMonth values must start with '--'")
	if err != nil {
		return GMonth{}, err
	}
	rem, err = expectChar(rem, '-', "gMonth values must start with '--'")
	if err != nil {
		return GMonth{}, err
	}
	month, rem, err := parseTwoDigit(rem, 1, monthsPerYear, "month")
	if err != nil {
		return GMonth{}, err
	}
	tz, rem, _ := parseTimezone(rem)
	if rem != "" {
		return GMonth{}, newParseDateTimeError("unrecognized value suffix")
	}

	return NewGMonth(month, tz)
}

// Month returns the month property value as a 1-based integer.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.14.1)
//
// Parameters:
//
// Returns:
//   - uint8: The month value, bounded between 1 and 12.
func (g GMonth) Month() uint8 {
	return g.timestamp.month()
}

// TimezoneOffset returns the timezone offset associated with the GMonth value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.14.1)
//
// Parameters:
//
// Returns:
//   - *TimezoneOffset: The timezone offset associated with the value, or nil if the value is unzoned.
func (g GMonth) TimezoneOffset() *TimezoneOffset {
	return g.timestamp.timezoneOffset
}

// Adjust adjusts the timezone offset of the GMonth value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - newTz: The target timezone offset to apply to the GMonth value, or nil to make it unzoned.
//
// Returns:
//   - GMonth: The adjusted GMonth value.
func (g GMonth) Adjust(newTz *TimezoneOffset) GMonth {
	ts := g.timestamp.adjust(newTz)
	return GMonth{timestamp: ts}
}

// IsIdenticalWith checks if this GMonth is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - bool: True if both values are identical, false otherwise.
func (g GMonth) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(GMonth); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 2.2.1)
		// Two date/time values are identical if and only if their timeline values
		// are equal and their timezone offsets are either both absent, or both
		// present and equal.
		return g.timestamp.value.IsIdenticalWith(o.timestamp.value) &&
			((g.timestamp.timezoneOffset == nil && o.timestamp.timezoneOffset == nil) ||
				(g.timestamp.timezoneOffset != nil && o.timestamp.timezoneOffset != nil &&
					*g.timestamp.timezoneOffset == *o.timestamp.timezoneOffset))
	}
	return false
}

// Compare evaluates the temporal order of this GMonth against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare against, which must be a GMonth.
//
// Returns:
//   - int: Negative if less than, zero if equal, positive if greater than.
//   - error: An ErrDateTimeOverflow if the comparison is indeterminate, or an error if types are incomparable.
func (g GMonth) Compare(other XSDValue) (int, error) {
	if o, ok := other.(GMonth); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1)
		// Values of date/time types are ordered by their timeOnTimeline values, with
		// uncertainty added when comparing zoned and unzoned values.
		res, det := g.timestamp.compare(o.timestamp)
		if !det {
			return 0, ErrDateTimeOverflow
		}
		return res, nil
	}
	return 0, ErrDateTimeOverflow
}

// String returns the canonical lexical representation of the GMonth value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.14.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical representation formatted as "--MM" optionally followed by a timezone suffix.
func (g GMonth) String() string {
	var tzStr string
	if g.TimezoneOffset() != nil {
		tzStr = g.TimezoneOffset().String()
	}
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.14.2)
	// The canonical representation formats a gMonth value as "--MM" with a
	// trailing timezone suffix if a timezone is present.
	return fmt.Sprintf("--%02d%s", g.Month(), tzStr)
}
