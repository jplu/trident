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

// Time represents the xsd:time primitive datatype as defined in the governing specifications.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.8)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// The value space of time is the set of recurring points in time (hour, minute,
// and second) during a single day, optionally accompanied by a timezone offset,
// modeled internally as a Timestamp.
type Time struct {
	// Timestamp is the internal temporal representation containing timeline and timezone information.
	timestamp Timestamp
}

// NewTime constructs a new Time value from hour, minute, second, and optional timezone offset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.8 & Appendix D.2.1)
//
// Parameters:
//   - hour: The hour component of the time value.
//   - minute: The minute component of the time value.
//   - second: The high-precision decimal second component.
//   - timezone: A pointer to the timezone offset, or nil if the value is unzoned.
//
// Returns:
//   - Time: The successfully constructed Time value.
//   - error: An error if the components do not define a valid point on the timeline or violate calendar bounds.
func NewTime(hour, minute uint8, second Decimal, timezone *TimezoneOffset) (Time, error) {
	h := hour

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.8)
	// The lexical representation '24:00:00' is permitted and corresponds to '00:00:00'
	// on the following day. For the time value space itself, it is normalized to
	// '00:00:00' of the same day.
	if h == 24 && minute == 0 && second.getValue().Sign() == 0 {
		h = 0
	}
	ts, err := newTimestamp(&DateTimeSevenPropertyModel{
		Hour:           &h,
		Minute:         &minute,
		Second:         &second,
		TimezoneOffset: timezone,
	})
	if err != nil {
		return Time{}, err
	}
	return Time{timestamp: ts}, nil
}

// ParseTime parses a lexical string matching the representation of xsd:time.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.8.2)
//
// Parameters:
//   - input: The raw string literal to be parsed.
//
// Returns:
//   - Time: The parsed Time value.
//   - error: An error if the lexical format or range constraints are violated.
func ParseTime(input string) (Time, error) {
	hour, rem, err := parseTwoDigit(input, 0, hoursPerDay, "hour")
	if err != nil {
		return Time{}, err
	}
	rem, err = expectChar(rem, ':', "hours and minutes must be separated by ':'")
	if err != nil {
		return Time{}, err
	}
	minute, rem, err := parseTwoDigit(rem, 0, maxMinutesPerHour, "minute")
	if err != nil {
		return Time{}, err
	}
	rem, err = expectChar(rem, ':', "minutes and seconds must be separated by ':'")
	if err != nil {
		return Time{}, err
	}
	second, rem, err := parseSecond(rem)
	if err != nil {
		return Time{}, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.8)
	// Hour 24 is restricted; any time after 24:00:00 (such as 24:00:00.001) is strictly invalid.
	if hour == 24 && (minute != 0 || second.getValue().Sign() != 0) {
		return Time{}, newParseDateTimeError("times are not allowed to be after 24:00:00")
	}

	tz, rem, _ := parseTimezone(rem)
	if rem != "" {
		return Time{}, newParseDateTimeError("unrecognized value suffix")
	}
	return NewTime(hour, minute, second, tz)
}

// Hour returns the hour component of the Time value, bounded between 0 and 23.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.8.1)
//
// Parameters:
//
//	None.
//
// Returns:
//   - uint8: The hour component of the Time.
func (t Time) Hour() uint8 {
	return t.timestamp.hour()
}

// Minute returns the minute component of the Time value, bounded between 0 and 59.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.8.1)
//
// Parameters:
//
//	None.
//
// Returns:
//   - uint8: The minute component of the Time.
func (t Time) Minute() uint8 {
	return t.timestamp.minute()
}

// Second returns the arbitrary-precision second component of the Time value, bounded between 0 and 60 (exclusive).
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.8.1)
//
// Parameters:
//
//	None.
//
// Returns:
//   - Decimal: The high-precision decimal second component of the Time.
func (t Time) Second() Decimal {
	return t.timestamp.second()
}

// TimezoneOffset returns the associated timezone offset of the Time value, or nil if unzoned.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.8.1)
//
// Parameters:
//
//	None.
//
// Returns:
//   - *TimezoneOffset: A pointer to the timezone offset, or nil if unzoned.
func (t Time) TimezoneOffset() *TimezoneOffset {
	return t.timestamp.timezoneOffset
}

// IsIdenticalWith checks if this Time is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.8.4)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - bool: True if both values are identical, false otherwise.
func (t Time) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Time); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.1)
		// Identity for temporal datatypes requires both equivalent timeline coordinates
		// and matching timezone offset components (both must either have equivalent offsets
		// or both be absent).
		return t.timestamp.value.IsIdenticalWith(o.timestamp.value) &&
			((t.timestamp.timezoneOffset == nil && o.timestamp.timezoneOffset == nil) ||
				(t.timestamp.timezoneOffset != nil && o.timestamp.timezoneOffset != nil &&
					*t.timestamp.timezoneOffset == *o.timestamp.timezoneOffset))
	}
	return false
}

// Compare evaluates the order relation of this Time against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An error if the values are incomparable.
func (t Time) Compare(other XSDValue) (int, error) {
	if o, ok := other.(Time); ok {
		res, determinate := t.timestamp.compare(o.timestamp)
		if !determinate {
			return 0, ErrDateTimeOverflow
		}
		return res, nil
	}
	return 0, ErrDateTimeOverflow
}

// String returns the canonical lexical representation of the Time value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.2)
//
// Parameters:
//
//	None.
//
// Returns:
//   - string: The canonical string representation in the format "hh:mm:ss.sss" with optional timezone suffix.
func (t Time) String() string {
	var tzStr string
	if t.TimezoneOffset() != nil {
		tzStr = t.TimezoneOffset().String()
	}
	secStr := t.Second().String()
	if !strings.Contains(secStr, ".") && len(secStr) == 1 {
		secStr = "0" + secStr
	} else if strings.Contains(secStr, ".") {
		parts := strings.Split(secStr, ".")
		if len(parts[0]) == 1 {
			secStr = "0" + secStr
		}
	}
	return fmt.Sprintf("%02d:%02d:%s%s", t.Hour(), t.Minute(), secStr, tzStr)
}

// toDateTimeOnDefaultDate converts the Time value into a DateTime value on a fixed recovery date.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.8)
//
// Parameters:
//
//	None.
//
// Returns:
//   - DateTime: The constructed DateTime value mapped onto the recovery date 1972-12-31.
func (t Time) toDateTimeOnDefaultDate() DateTime {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.8)
	// For comparison or calculations requiring timeline translation, a 'time' value
	// is mapped onto a default date timeline (using the fixed recovery date 1972-12-31).
	dt, _ := NewDateTime(
		defaultTimeYear,
		defaultTimeMonth,
		defaultTimeDay,
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.TimezoneOffset(),
	)
	return dt
}

// Adjust adjusts the timezone of the Time value to the specified TimezoneOffset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - newTz: A pointer to the target TimezoneOffset to be applied, or nil to make it unzoned.
//
// Returns:
//   - Time: The adjusted Time value.
func (t Time) Adjust(newTz *TimezoneOffset) Time {
	dt := t.toDateTimeOnDefaultDate()
	newDt := dt.Adjust(newTz)
	newT, _ := NewTime(newDt.Hour(), newDt.Minute(), newDt.Second(), newDt.TimezoneOffset())

	return newT
}
