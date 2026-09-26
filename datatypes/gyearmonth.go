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

// GYearMonth represents the gYearMonth datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A specific month in a specific Gregorian calendar year, optionally accompanied by
// a timezone offset, modeled internally via a high-precision Timestamp.
type GYearMonth struct {
	// timestamp represents the internal linear timeline value and timezone configuration
	// of the GYearMonth instance.
	timestamp Timestamp
}

// NewGYearMonth constructs a new GYearMonth instance from a year, month, and optional timezone offset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10)
//
// Parameters:
//   - year: The signed 64-bit calendar year.
//   - month: The 1-based month component (1 to 12).
//   - timezone: The optional timezone offset from UTC.
//
// Returns:
//   - GYearMonth: The constructed GYearMonth value.
//   - error: An error if the Gregorian calendar fields violate calendar constraints.
func NewGYearMonth(year int64, month uint8, timezone *TimezoneOffset) (GYearMonth, error) {
	ts, err := newTimestamp(&DateTimeSevenPropertyModel{Year: &year, Month: &month, TimezoneOffset: timezone})
	if err != nil {
		return GYearMonth{}, err
	}
	return GYearMonth{timestamp: ts}, nil
}

// ParseGYearMonth parses a string matching the lexical representation of xsd:gYearMonth.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10.1)
//
// Parameters:
//   - input: The raw string literal to be parsed and validated.
//
// Returns:
//   - GYearMonth: The parsed GYearMonth value.
//   - error: An error of type ParseDateTimeError if the lexical representation is invalid.
func ParseGYearMonth(input string) (GYearMonth, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.10.1)
	// The lexical representation of gYearMonth is of the form '-'? yyyy '-' mm zzz?, parsed sequentially.

	// Step 1: Parse Year Component
	// Extract the signed calendar year from the beginning of the string.
	year, rem, err := parseYear(input)
	if err != nil {
		return GYearMonth{}, err
	}

	// Step 2: Validate Separator
	// Ensure that a hyphen separates the year and month components.
	rem, err = expectChar(rem, '-', "year and month must be separated by '-'")
	if err != nil {
		return GYearMonth{}, err
	}

	// Step 3: Parse Month Component
	// Parse the two-digit month component and validate that it falls within the range [01, 12].
	month, rem, err := parseTwoDigit(rem, 1, monthsPerYear, "month")
	if err != nil {
		return GYearMonth{}, err
	}

	// Step 4: Parse Timezone Offset
	// Parse the optional timezone suffix, ensuring no trailing unrecognized characters remain.
	tz, rem, _ := parseTimezone(rem)
	if rem != "" {
		return GYearMonth{}, newParseDateTimeError("unrecognized value suffix")
	}
	return NewGYearMonth(year, month, tz)
}

// Year returns the year component of the GYearMonth.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10.1)
//
// Parameters:
//
// Returns:
//   - int64: The signed 64-bit calendar year.
func (g GYearMonth) Year() int64 {
	return g.timestamp.year()
}

// Month returns the month component of the GYearMonth.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10.1)
//
// Parameters:
//
// Returns:
//   - uint8: The 1-based month component (1 to 12).
func (g GYearMonth) Month() uint8 {
	return g.timestamp.month()
}

// TimezoneOffset returns the optional timezone offset of the GYearMonth.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10.1)
//
// Parameters:
//
// Returns:
//   - *TimezoneOffset: The timezone offset of the GYearMonth, or nil if unzoned.
func (g GYearMonth) TimezoneOffset() *TimezoneOffset {
	return g.timestamp.timezoneOffset
}

// Adjust adjusts the timezone of the GYearMonth instance to a new timezone.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - newTz: The target timezone offset to adjust to.
//
// Returns:
//   - GYearMonth: The adjusted GYearMonth value.
func (g GYearMonth) Adjust(newTz *TimezoneOffset) GYearMonth {
	ts := g.timestamp.adjust(newTz)
	return GYearMonth{timestamp: ts}
}

// IsIdenticalWith checks if this GYearMonth is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (g GYearMonth) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(GYearMonth); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.10)
		// Two gYearMonth values are identical if and only if their year, month, and timezoneOffset properties are
		// pairwise identical.
		return g.timestamp.value.IsIdenticalWith(o.timestamp.value) &&
			((g.timestamp.timezoneOffset == nil && o.timestamp.timezoneOffset == nil) ||
				(g.timestamp.timezoneOffset != nil && o.timestamp.timezoneOffset != nil &&
					*g.timestamp.timezoneOffset == *o.timestamp.timezoneOffset))
	}
	return false
}

// Compare evaluates the order relation of this GYearMonth against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An error if the values are incomparable.
func (g GYearMonth) Compare(other XSDValue) (int, error) {
	if o, ok := other.(GYearMonth); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2)
		// Comparison is performed by projecting the partial date properties onto a linear timeline. Zoned and unzoned
		// values can yield an indeterminate result.
		res, det := g.timestamp.compare(o.timestamp)
		if !det {
			return 0, ErrDateTimeOverflow
		}
		return res, nil
	}
	return 0, ErrDateTimeOverflow
}

// String returns the canonical lexical representation of the GYearMonth value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical lexical representation of the GYearMonth.
func (g GYearMonth) String() string {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.10.2)
	// The canonical representation formats the year to at least 4 digits (prefixed with '-' if negative) and the month
	// to 2 digits, followed by the canonical timezone.
	var tzStr string
	if g.TimezoneOffset() != nil {
		tzStr = g.TimezoneOffset().String()
	}
	y := g.Year()
	yearStr := fmt.Sprintf("%04d", y)
	if y < 0 {
		yearStr = fmt.Sprintf("-%04d", -y)
	}
	return fmt.Sprintf("%s-%02d%s", yearStr, g.Month(), tzStr)
}
