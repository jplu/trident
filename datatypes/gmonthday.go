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

// GMonthDay represents the gMonthDay primitive datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.12)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A recurring calendar date consisting of a month and a day of the month, optionally
// accompanied by a timezone offset, modeled internally via a high-precision timeline coordinate.
type GMonthDay struct {
	// timestamp represents the internal linear timeline value
	// and timezone information of the GMonthDay value.
	timestamp Timestamp
}

// NewGMonthDay constructs a new GMonthDay instance from the specified month, day, and optional timezone offset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.12)
//
// Parameters:
//   - month: The 1-based month value (1 to 12).
//   - day: The 1-based day of the month (1 to 31).
//   - timezone: The optional timezone offset from UTC.
//
// Returns:
//   - GMonthDay: The constructed GMonthDay value.
//   - error: An error if the Gregorian calendar fields violate calendar constraints or boundary limits.
func NewGMonthDay(month, day uint8, timezone *TimezoneOffset) (GMonthDay, error) {
	ts, err := newTimestamp(&DateTimeSevenPropertyModel{Month: &month, Day: &day, TimezoneOffset: timezone})
	if err != nil {
		return GMonthDay{}, err
	}
	return GMonthDay{timestamp: ts}, nil
}

// ParseGMonthDay parses a string literal matching the lexical representation of xsd:gMonthDay.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.12.2)
//
// Parameters:
//   - input: The raw string literal to be parsed.
//
// Returns:
//   - GMonthDay: The parsed GMonthDay value.
//   - error: A ParseDateTimeError if parsing or validation fails.
func ParseGMonthDay(input string) (GMonthDay, error) {
	// Step 1: Validate and parse prefix character.
	// Confirm that the lexical string begins with the first hyphen character.
	rem, err := expectChar(input, '-', "gMonthDay values must start with '--'")
	if err != nil {
		return GMonthDay{}, err
	}

	// Step 2: Validate and parse second prefix character.
	// Confirm that the lexical string is immediately followed by a second hyphen character.
	rem, err = expectChar(rem, '-', "gMonthDay values must start with '--'")
	if err != nil {
		return GMonthDay{}, err
	}

	// Step 3: Parse the month component.
	// Extract exactly two digits representing the month and validate that they lie in the range [1, 12].
	month, rem, err := parseTwoDigit(rem, 1, monthsPerYear, "month")
	if err != nil {
		return GMonthDay{}, err
	}

	// Step 4: Validate and parse component separator.
	// Confirm that a hyphen character separates the parsed month and day components.
	rem, err = expectChar(rem, '-', "month and day must be separated by '-'")
	if err != nil {
		return GMonthDay{}, err
	}

	// Step 5: Parse the day component.
	// Extract exactly two digits representing the day of the month and validate that they lie in the range [1, 31].
	day, rem, err := parseTwoDigit(rem, 1, maxDaysInMonth, "day")
	if err != nil {
		return GMonthDay{}, err
	}

	// Step 6: Parse the timezone offset suffix.
	// Extract and process any optional timezone offset trailing characters.
	tz, rem, _ := parseTimezone(rem)

	if rem != "" {
		return GMonthDay{}, newParseDateTimeError("unrecognized value suffix")
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.12.1)
	// The year component is absent in the gMonthDay value space. Consequently,
	// February 29 is treated as valid to accommodate leap years. A nil year
	// is passed to the validation routine to enforce this leap year rule.
	if vErr := validateDayOfMonth(nil, month, day); vErr != nil {
		return GMonthDay{}, vErr
	}

	return NewGMonthDay(month, day, tz)
}

// Month returns the month component of the GMonthDay value, ranging from 1 to 12.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.12.1)
//
// Returns:
//   - uint8: The 1-based month value (1 to 12).
func (g GMonthDay) Month() uint8 {
	return g.timestamp.month()
}

// Day returns the day component of the GMonthDay value, ranging from 1 to 31.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.12.1)
//
// Returns:
//   - uint8: The 1-based day of the month (1 to 31).
func (g GMonthDay) Day() uint8 {
	return g.timestamp.day()
}

// TimezoneOffset returns the optional timezone offset of the GMonthDay value, or nil if unzoned.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.12.1)
//
// Returns:
//   - *TimezoneOffset: The timezone offset associated with the value, or nil if the value is unzoned.
func (g GMonthDay) TimezoneOffset() *TimezoneOffset {
	return g.timestamp.timezoneOffset
}

// Adjust adjusts the GMonthDay to a new timezone offset, returning an updated GMonthDay.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - newTz: The target timezone offset to apply.
//
// Returns:
//   - GMonthDay: The GMonthDay adjusted to the target timezone.
func (g GMonthDay) Adjust(newTz *TimezoneOffset) GMonthDay {
	ts := g.timestamp.adjust(newTz)
	return GMonthDay{timestamp: ts}
}

// IsIdenticalWith checks if this GMonthDay is identical to another XSDValue according to the identity criteria defined
// in the specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.12.4)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (g GMonthDay) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(GMonthDay); ok {
		return g.timestamp.value.IsIdenticalWith(o.timestamp.value) &&
			((g.timestamp.timezoneOffset == nil && o.timestamp.timezoneOffset == nil) ||
				(g.timestamp.timezoneOffset != nil && o.timestamp.timezoneOffset != nil &&
					*g.timestamp.timezoneOffset == *o.timestamp.timezoneOffset))
	}
	return false
}

// Compare evaluates the temporal order of this GMonthDay against another XSDValue according to the partial order
// defined in the specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An error if incomparable or if timeline uncertainty overlaps.
func (g GMonthDay) Compare(other XSDValue) (int, error) {
	if o, ok := other.(GMonthDay); ok {
		res, det := g.timestamp.compare(o.timestamp)
		if !det {
			return 0, ErrDateTimeOverflow
		}
		return res, nil
	}
	return 0, ErrDateTimeOverflow
}

// String returns the canonical lexical representation of the GMonthDay value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.12.2)
//
// Returns:
// - string: The formatted canonical lexical representation matching the pattern "--MM-DD" with a timezone suffix if
// present.
func (g GMonthDay) String() string {
	var tzStr string
	if g.TimezoneOffset() != nil {
		tzStr = g.TimezoneOffset().String()
	}
	return fmt.Sprintf("--%02d-%02d%s", g.Month(), g.Day(), tzStr)
}
