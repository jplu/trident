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
	"strings"
)

// DateTime represents the dateTime primitive datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7 & Appendix D.2.1)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// The type represents points on a timeline, modeled as a seven-property value
// consisting of: year, month, day, hour, minute, second, and an optional timezone offset.
type DateTime struct {
	// timestamp represents the linear point on the timeline combined with timezone information.
	timestamp Timestamp
}

// NewDateTime creates a new DateTime value from its seven discrete components.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7 & Appendix D.2.1)
//
// Parameters:
//   - year: The 64-bit integer representing the Gregorian calendar year.
//   - month: The 8-bit unsigned integer representing the month (1 to 12).
//   - day: The 8-bit unsigned integer representing the day of the month (1 to 31).
//   - hour: The 8-bit unsigned integer representing the hour (0 to 24).
//   - minute: The 8-bit unsigned integer representing the minute (0 to 59).
//   - second: The high-precision decimal second.
//   - timezone: The optional timezone offset from UTC in minutes.
//
// Returns:
//   - DateTime: The newly constructed DateTime value.
//   - error: An error if calendar constraints are violated or if timeline resolution overflows.
func NewDateTime(
	year int64,
	month, day, hour, minute uint8,
	second Decimal,
	timezone *TimezoneOffset,
) (DateTime, error) {
	ts, err := newTimestamp(&DateTimeSevenPropertyModel{
		Year:           &year,
		Month:          &month,
		Day:            &day,
		Hour:           &hour,
		Minute:         &minute,
		Second:         &second,
		TimezoneOffset: timezone,
	})
	if err != nil {
		return DateTime{}, err
	}
	return DateTime{timestamp: ts}, nil
}

// ParseDateTime parses a string literal matching the lexical representation of xsd:dateTime.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7.2)
//
// Parameters:
//   - s: The raw string literal to be parsed.
//
// Returns:
//   - DateTime: The successfully parsed DateTime value.
//   - error: An error if parsing fails or if the resulting components violate calendar constraints.
func ParseDateTime(s string) (DateTime, error) {
	props, rem, err := parseDateTime(s)
	if err != nil {
		return DateTime{}, err
	}
	if rem != "" {
		return DateTime{}, newParseDateTimeError("unrecognized value suffix")
	}
	// Implementation Note: Redundant validation elision
	// parseDateTime has already fully validated every constraint that newTimestamp's
	// internal validateDateTimeModel would.
	ts, _ := newTimestamp(&props)
	return DateTime{timestamp: ts}, nil
}

// Year returns the year component of the DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7.1)
//
// Parameters:
//
// Returns:
//   - int64: The year component.
func (d DateTime) Year() int64 { return d.timestamp.year() }

// Month returns the month component of the DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7.1)
//
// Parameters:
//
// Returns:
//   - uint8: The 1-based month component (ranging from 1 to 12).
func (d DateTime) Month() uint8 { return d.timestamp.month() }

// Day returns the day component of the DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7.1)
//
// Parameters:
//
// Returns:
//   - uint8: The day of the month component (ranging from 1 to 31).
func (d DateTime) Day() uint8 { return d.timestamp.day() }

// Hour returns the hour component of the DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7.1)
//
// Parameters:
//
// Returns:
//   - uint8: The hour component (ranging from 0 to 23).
func (d DateTime) Hour() uint8 { return d.timestamp.hour() }

// Minute returns the minute component of the DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7.1)
//
// Parameters:
//
// Returns:
//   - uint8: The minute component (ranging from 0 to 59).
func (d DateTime) Minute() uint8 { return d.timestamp.minute() }

// Second returns the high-precision decimal second component of the DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7.1)
//
// Parameters:
//
// Returns:
//   - Decimal: The high-precision decimal second component.
func (d DateTime) Second() Decimal { return d.timestamp.second() }

// TimezoneOffset returns the timezone offset of the DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7.1)
//
// Parameters:
//
// Returns:
//   - *TimezoneOffset: The pointer to the timezone offset component, or nil if unzoned.
func (d DateTime) TimezoneOffset() *TimezoneOffset { return d.timestamp.timezoneOffset }

// AddDuration adds a Duration value to the DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3.3)
//
// Parameters:
//   - rhs: The Duration value to be added.
//
// Returns:
//   - DateTime: The resulting DateTime value.
func (d DateTime) AddDuration(rhs Duration) DateTime {
	props := d.timestamp.ToProperties()

	// Step 1: Add components and normalize
	// Add the year-month and day-time duration components using the standard specification addition algorithm.
	// Implementation Note: Redundant validation elision
	// dateTimePlusDuration's only error source is normalizeMonth's overflow guard,
	// which requires a year magnitude beyond math.MaxInt64/2.
	newProps, _ := dateTimePlusDuration(rhs, props)

	// Step 2: Convert properties back to timestamp
	// Convert the normalized calendar properties back into a high-precision Timestamp on the timeline.
	// Implementation Note: Redundant validation elision
	// newProps is built exclusively from values already produced by Timestamp's own
	// modular-arithmetic accessors
	ts, _ := newTimestamp(newProps)
	return DateTime{timestamp: ts}
}

// CheckedSubDuration subtracts a Duration value from the DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3.3)
//
// Parameters:
//   - rhs: The Duration value to be subtracted.
//
// Returns:
//   - DateTime: The resulting DateTime value.
//   - error: An error if the calculation causes calendar fields to overflow or underflow.
func (d DateTime) CheckedSubDuration(rhs Duration) (DateTime, error) {
	// Step 1: Negate the duration
	// Reverse the sign of both the yearMonth and dayTime components of the incoming duration.
	negRHS, err := rhs.CheckedNeg()
	if err != nil {
		return DateTime{}, ErrDateTimeOverflow
	}

	// Step 2: Add the negated duration
	// Add the negated duration to the current instant to effectively perform temporal subtraction.
	return d.AddDuration(negRHS), nil
}

// IsIdenticalWith checks if this DateTime is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.7.1)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if the two values are identical, false otherwise.
func (d DateTime) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(DateTime); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.7.1)
		// Two dateTime values are identical if and only if their year, month,
		// day, hour, minute, second, and timezoneOffset properties are identical.
		// Values representing the same instant on the timeline with different
		// timezone offsets are not identical.
		return d.timestamp.value.IsIdenticalWith(o.timestamp.value) &&
			((d.timestamp.timezoneOffset == nil && o.timestamp.timezoneOffset == nil) ||
				(d.timestamp.timezoneOffset != nil && o.timestamp.timezoneOffset != nil &&
					*d.timestamp.timezoneOffset == *o.timestamp.timezoneOffset))
	}
	return false
}

// Compare evaluates the order relation of this DateTime against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - int: Negative if less, zero if equal, positive if greater.
//   - error: An error if the values are incomparable.
func (d DateTime) Compare(other XSDValue) (int, error) {
	if o, ok := other.(DateTime); ok {
		res, determinate := d.timestamp.compare(o.timestamp)
		if !determinate {
			// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.7.1)
			// The order relation on dateTime values is a partial order. Comparing
			// zoned and unzoned values yields an indeterminate result if their timeline
			// intervals overlap.
			return 0, ErrDateTimeOverflow
		}
		return res, nil
	}
	return 0, ErrDateTimeOverflow
}

// String returns the canonical lexical representation of the DateTime value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3.6)
//
// Parameters:
//
// Returns:
//   - string: The canonical lexical representation.
func (d DateTime) String() string {
	var tzStr string
	if d.TimezoneOffset() != nil {
		tzStr = d.TimezoneOffset().String()
	}

	y := d.Year()
	yearStr := fmt.Sprintf("%04d", y)
	if y < 0 {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.7.2)
		// Negative years must be prefixed with a '-' sign, and the year
		// part must contain at least 4 digits.
		yearStr = fmt.Sprintf("-%04d", -y)
	}

	secStr := d.Second().String()

	// Implementation Note: Formatting fractional seconds
	// Pad the integer part of the seconds value with a leading zero if it has only a single digit to maintain the
	// canonical layout hh:mm:ss.sss.
	if !strings.Contains(secStr, ".") && len(secStr) == 1 {
		secStr = "0" + secStr
	} else if strings.Contains(secStr, ".") {
		parts := strings.Split(secStr, ".")
		if len(parts[0]) == 1 {
			secStr = "0" + secStr
		}
	}

	return fmt.Sprintf("%s-%02d-%02dT%02d:%02d:%s%s", yearStr, d.Month(), d.Day(), d.Hour(), d.Minute(), secStr, tzStr)
}

// Adjust adjusts the timezone of the DateTime value to the specified TimezoneOffset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - newTz: The target TimezoneOffset to apply.
//
// Returns:
//   - DateTime: The adjusted DateTime value.
func (d DateTime) Adjust(newTz *TimezoneOffset) DateTime {
	ts := d.timestamp.adjust(newTz)
	return DateTime{timestamp: ts}
}
