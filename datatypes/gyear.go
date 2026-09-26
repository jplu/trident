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

// GYear represents the gYear primitive datatype as defined in the governing specifications.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.11)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A Gregorian calendar year, optionally accompanied by a timezone offset, modeled as a
// single-property subset of the seven-property date/time model where month, day, hour,
// minute, and second properties are absent.
type GYear struct {
	// Timestamp represents the exact linear coordinates on the timeline
	// mapped for the Gregorian year.
	timestamp Timestamp
}

// NewGYear constructs a new GYear value from a given year and an optional timezone offset.
// This function cannot return an error as Gregorian year coordinates and timezone are always valid.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.11)
//
// Parameters:
//   - year: The 64-bit signed integer representing the Gregorian calendar year.
//   - timezone: The optional timezone offset from UTC.
//
// Returns:
//   - GYear: The constructed GYear value.
func NewGYear(year int64, timezone *TimezoneOffset) GYear {
	ts, _ := newTimestamp(&DateTimeSevenPropertyModel{Year: &year, TimezoneOffset: timezone})
	return GYear{timestamp: ts}
}

// ParseGYear parses a string literal matching the lexical representation of xsd:gYear.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.11.2)
//
// Parameters:
//   - input: The raw lexical string representing the Gregorian year and optional timezone.
//
// Returns:
//   - GYear: The parsed GYear value.
//   - error: An error of type ParseDateTimeError if the input does not conform to the lexical rules.
func ParseGYear(input string) (GYear, error) {
	year, rem, err := parseYear(input)
	if err != nil {
		return GYear{}, err
	}
	tz, rem, _ := parseTimezone(rem)
	if rem != "" {
		return GYear{}, newParseDateTimeError("unrecognized value suffix")
	}
	return NewGYear(year, tz), nil
}

// Year returns the year component of the GYear value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.11.1)
//
// Parameters:
//
// Returns:
//   - int64: The signed 64-bit year value.
func (g GYear) Year() int64 {
	return g.timestamp.year()
}

// TimezoneOffset returns the timezone offset associated with the GYear value,
// or nil if the value is unzoned (coordinate-local).
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.11.1)
//
// Parameters:
//
// Returns:
//   - *TimezoneOffset: The timezone offset associated with this value, or nil if unzoned.
func (g GYear) TimezoneOffset() *TimezoneOffset {
	return g.timestamp.timezoneOffset
}

// Adjust adjusts the GYear value to match the newly specified timezone offset.
// If the target timezone is nil, the adjusted GYear becomes unzoned.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - newTz: The target timezone offset to adjust to, or nil to make it unzoned.
//
// Returns:
//   - GYear: The adjusted GYear value.
func (g GYear) Adjust(newTz *TimezoneOffset) GYear {
	ts := g.timestamp.adjust(newTz)
	return GYear{timestamp: ts}
}

// IsIdenticalWith checks if this GYear is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.11.4)
//
// Parameters:
//   - other: The other XSDValue to compare identity with.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (g GYear) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(GYear); ok {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.1)
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

// Compare evaluates the order relation of this GYear against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - int: Negative if less, zero if equal, positive if greater.
//   - error: An error if the values are incomparable or the result is indeterminate.
func (g GYear) Compare(other XSDValue) (int, error) {
	if o, ok := other.(GYear); ok {
		res, det := g.timestamp.compare(o.timestamp)
		if !det {
			// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1)
			// Order relations on date/time types are partial. Comparing zoned and unzoned
			// values can yield an indeterminate result if their possible timeline
			// intervals overlap, which returns a timeline overflow error here.
			return 0, ErrDateTimeOverflow
		}
		return res, nil
	}
	return 0, ErrDateTimeOverflow
}

// String returns the canonical lexical representation of the GYear value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.11.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical string representation formatted as yyyy[zzz].
func (g GYear) String() string {
	var tzStr string
	if g.TimezoneOffset() != nil {
		tzStr = g.TimezoneOffset().String()
	}
	y := g.Year()
	yearStr := fmt.Sprintf("%04d", y)
	if y < 0 {
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.11.2)
		// Negative years must be prefixed with a '-' sign, and the year
		// part must contain at least 4 digits.
		yearStr = fmt.Sprintf("-%04d", -y)
	}
	return fmt.Sprintf("%s%s", yearStr, tzStr)
}
