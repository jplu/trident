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
	"fmt"
)

// Date represents the date primitive datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// The type models a Gregorian calendar date (consisting of year, month, and day properties)
// that optionally carries an associated timezone offset. The value space comprises top-open
// intervals of exactly one day in length on the timelines of dateTime, while its lexical space
// is formatted as '-'? yyyy '-' mm '-' dd zzz? with a fixed whiteSpace collapse facet.
type Date struct {
	// Timestamp represents the high-precision timeline coordinate of the date.
	timestamp Timestamp
}

// NewDate instantiates a Date value from its year, month, day, and an optional timezone offset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10 & Appendix D.2.1)
//
// Parameters:
//   - year: The signed 64-bit integer representing the Gregorian calendar year.
//   - month: The 8-bit unsigned integer representing the calendar month (1 to 12).
//   - day: The 8-bit unsigned integer representing the day of the month (1 to 31, bounded by the month).
//   - timezone: A pointer to the TimezoneOffset struct representing the timezone offset component, or nil if unzoned.
//
// Returns:
//   - Date: The successfully initialized Date value.
//   - error: An error if the Gregorian calendar fields violate calendar constraints or boundary limits.
func NewDate(year int64, month, day uint8, timezone *TimezoneOffset) (Date, error) {
	ts, err := newTimestamp(&DateTimeSevenPropertyModel{
		Year:           &year,
		Month:          &month,
		Day:            &day,
		TimezoneOffset: timezone,
	})
	if err != nil {
		return Date{}, err
	}
	return Date{timestamp: ts}, nil
}

// ParseDate parses a string literal matching the lexical representation of xsd:date.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10.1)
//
// Parameters:
//   - input: The raw string literal to be parsed and normalized.
//
// Returns:
//   - Date: The successfully parsed Date value.
//
// - error: A ParseDateTimeError if the input literal does not conform to the expected format or violates calendar
// boundaries.
func ParseDate(input string) (Date, error) {
	// Implementation Note: Parse individual date components sequentially.
	// The implementation parses the year, separator, month, separator, day, and timezone
	// in sequence to avoid complex backtracking regular expressions.

	// Step 1: Parse the year component.
	// Extracts the year from the starting digits, checking for minus signs and leading zeros.
	year, rem, err := parseYear(input)
	if err != nil {
		return Date{}, err
	}

	// Step 2: Expect the hyphen separator.
	// Checks that a '-' separator immediately follows the year component.
	rem, err = expectChar(rem, '-', "year and month must be separated by '-'")
	if err != nil {
		return Date{}, err
	}

	// Step 3: Parse the two-digit month component.
	// Parses exactly two digits and ensures they fall within the valid range of 1 to 12.
	month, rem, err := parseTwoDigit(rem, 1, monthsPerYear, "month")
	if err != nil {
		return Date{}, err
	}

	// Step 4: Expect the hyphen separator.
	// Checks that a '-' separator immediately follows the month component.
	rem, err = expectChar(rem, '-', "month and day must be separated by '-'")
	if err != nil {
		return Date{}, err
	}

	// Step 5: Parse the two-digit day component.
	// Parses exactly two digits and ensures they fall within the valid range of 1 to 31.
	day, rem, err := parseTwoDigit(rem, 1, maxDaysInMonth, "day")
	if err != nil {
		return Date{}, err
	}

	// Step 6: Parse the optional timezone offset.
	// Processes timezone offset indicators like 'Z' or a signed offset '+hh:mm' or '-hh:mm'.
	tz, rem, _ := parseTimezone(rem)

	// Step 7: Validate remaining suffix.
	// Ensures that there are no trailing characters after the parsed timezone component.
	if rem != "" {
		return Date{}, newParseDateTimeError("unrecognized value suffix")
	}

	// Step 8: Validate the day of month.
	// Checks if the day of month is logically valid for the parsed month and year, including leap years.
	if vErr := validateDayOfMonth(&year, month, day); vErr != nil {
		return Date{}, vErr
	}

	return NewDate(year, month, day, tz)
}

// Year returns the year component of the Date value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10)
//
// Parameters:
//
// Returns:
//   - int64: The signed 64-bit integer representing the year component, which can be negative.
func (d Date) Year() int64 {
	return d.timestamp.year()
}

// Month returns the month component of the Date value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10)
//
// Parameters:
//
// Returns:
//   - uint8: The 8-bit unsigned integer representing the month component (1 to 12).
func (d Date) Month() uint8 {
	return d.timestamp.month()
}

// Day returns the day component of the Date value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10)
//
// Parameters:
//
// Returns:
//   - uint8: The 8-bit unsigned integer representing the day component (1 to 31, bounded by the month).
func (d Date) Day() uint8 {
	return d.timestamp.day()
}

// TimezoneOffset returns the optional timezone offset of the Date.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10)
//
// Parameters:
//
// Returns:
//   - *TimezoneOffset: A pointer to the TimezoneOffset of the value, or nil if unzoned.
func (d Date) TimezoneOffset() *TimezoneOffset {
	return d.timestamp.timezoneOffset
}

// IsIdenticalWith checks if this Date is strictly identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10.1)
//
// Parameters:
//   - other: The other XSDValue component to compare for identity.
//
// Returns:
//   - bool: True if the values are strictly identical in the value space; false otherwise.
func (d Date) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Date); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.10.1)
		// Two date values are identical if and only if they represent the same
		// timeline point and their timezone offsets are identical (either both absent,
		// or both present with the same offset value in minutes).
		return d.timestamp.value.IsIdenticalWith(o.timestamp.value) &&
			((d.timestamp.timezoneOffset == nil && o.timestamp.timezoneOffset == nil) ||
				(d.timestamp.timezoneOffset != nil && o.timestamp.timezoneOffset != nil &&
					*d.timestamp.timezoneOffset == *o.timestamp.timezoneOffset))
	}
	return false
}

// Compare evaluates the order relation of this Date against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - other: The other XSDValue component to evaluate against.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An ErrDateTimeOverflow if the comparison is indeterminate due to timezone overlap.
func (d Date) Compare(other XSDValue) (int, error) {
	if o, ok := other.(Date); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1)
		// Comparisons between date/time values follow a partial order. If one value
		// has a timezone and the other does not, the comparison is indeterminate
		// when their timeline ranges overlap (within the +/- 14 hours offset window).
		res, determinate := d.timestamp.compare(o.timestamp)
		if !determinate {
			return 0, ErrDateTimeOverflow
		}
		return res, nil
	}
	return 0, ErrDateTimeOverflow
}

// String returns the canonical lexical representation of the Date value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10.2)
//
// Parameters:
//
// Returns:
//   - string: The formatted canonical lexical representation matching the date pattern.
func (d Date) String() string {
	var tzStr string
	if d.TimezoneOffset() != nil {
		tzStr = d.TimezoneOffset().String()
	}
	y := d.Year()
	yearStr := fmt.Sprintf("%04d", y)
	if y < 0 {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.10.2)
		// Negative years must be prefixed with a '-' sign, and the year 0000
		// is prohibited. The year part must contain at least 4 digits.
		yearStr = fmt.Sprintf("-%04d", -y)
	}
	return fmt.Sprintf("%s-%02d-%02d%s", yearStr, d.Month(), d.Day(), tzStr)
}

// ToDateTime casts the Date value into a DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.10)
//
// Parameters:
//
// Returns:
//   - DateTime: The casted DateTime representation.
func (d Date) ToDateTime() DateTime {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.10)
	// Casting date to dateTime initializes the hourly, minute, and second values to 00:00:00.
	dt, _ := NewDateTime(d.Year(), d.Month(), d.Day(), 0, 0, DefaultDecimal(), d.TimezoneOffset())
	return dt
}

// Adjust adjusts the timezone of the Date value to the specified TimezoneOffset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - newTz: A pointer to the target TimezoneOffset to apply.
//
// Returns:
//   - Date: The adjusted Date value preserving the represented timeline instant.
func (d Date) Adjust(newTz *TimezoneOffset) Date {
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix E.3)
	// Adjusting a date to a timezone projects it to dateTime at 00:00:00, adjusts the
	// dateTime to the target timezone, and extracts the resulting date components.
	dt := d.ToDateTime()
	newDt := dt.Adjust(newTz)
	newD, _ := NewDate(newDt.Year(), newDt.Month(), newDt.Day(), newDt.TimezoneOffset())

	return newD
}
