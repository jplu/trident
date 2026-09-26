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
	"math"
)

// YearMonthDuration represents the yearMonthDuration datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// An integer number of months representing a duration of time, where the seconds component is restricted to zero,
// resulting in a totally ordered duration datatype.
type YearMonthDuration struct {
	// Months represents the duration in total months.
	months int64
}

// NewYearMonthDuration constructs a new YearMonthDuration with the specified number of months.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26)
//
// Parameters:
//   - months: The total duration in months represented as a signed 64-bit integer.
//
// Returns:
//   - YearMonthDuration: The initialized YearMonthDuration value.
func NewYearMonthDuration(months int64) YearMonthDuration {
	return YearMonthDuration{months: months}
}

// DefaultYearMonthDuration returns a YearMonthDuration representing a period of zero months.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26)
//
// Returns:
//   - YearMonthDuration: A YearMonthDuration value initialized to zero.
func DefaultYearMonthDuration() YearMonthDuration {
	return YearMonthDuration{months: 0}
}

// Years calculates the year component of the duration by dividing the total months by 12.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26)
//
// Returns:
//   - int64: The integer number of years contained in the duration.
func (d YearMonthDuration) Years() int64 {
	return d.months / monthsPerYear
}

// Months calculates the remaining month component of the duration after extracting whole years.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26)
//
// Returns:
//   - int64: The remaining number of months, ranging from -11 to 11.
func (d YearMonthDuration) Months() int64 {
	return d.months % monthsPerYear
}

// AllMonths returns the total number of months represented by this duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26)
//
// Returns:
//   - int64: The total number of months as a signed 64-bit integer.
func (d YearMonthDuration) AllMonths() int64 {
	return d.months
}

// CheckedAdd performs checked addition of two YearMonthDuration values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26)
//
// Parameters:
//   - rhs: The right-hand side YearMonthDuration value to add.
//
// Returns:
//   - YearMonthDuration: The sum of the two durations.
//   - error: An ErrDurationOverflow if the addition overflows the 64-bit integer range.
func (d YearMonthDuration) CheckedAdd(rhs YearMonthDuration) (YearMonthDuration, error) {
	// Implementation Note: Integer overflow prevention
	// This check ensures that adding the two signed 64-bit integer values does not exceed math.MaxInt64 or
	// math.MinInt64 to prevent silent wrap-around.
	if (rhs.months > 0 && d.months > math.MaxInt64-rhs.months) ||
		(rhs.months < 0 && d.months < math.MinInt64-rhs.months) {
		return YearMonthDuration{}, ErrDurationOverflow
	}
	return YearMonthDuration{months: d.months + rhs.months}, nil
}

// CheckedSub performs checked subtraction of two YearMonthDuration values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26)
//
// Parameters:
//   - rhs: The right-hand side YearMonthDuration value to subtract.
//
// Returns:
//   - YearMonthDuration: The difference of the two durations.
//   - error: An ErrDurationOverflow if the subtraction underflows or overflows the 64-bit integer range.
func (d YearMonthDuration) CheckedSub(rhs YearMonthDuration) (YearMonthDuration, error) {
	// Implementation Note: Integer overflow prevention
	// This check ensures that subtracting the two signed 64-bit integer values does not exceed math.MaxInt64 or
	// math.MinInt64 to prevent silent wrap-around.
	if (rhs.months > 0 && d.months < math.MinInt64+rhs.months) ||
		(rhs.months < 0 && d.months > math.MaxInt64+rhs.months) {
		return YearMonthDuration{}, ErrDurationOverflow
	}
	return YearMonthDuration{months: d.months - rhs.months}, nil
}

// CheckedNeg negates the YearMonthDuration value with overflow checking.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26)
//
// Returns:
//   - YearMonthDuration: The negated duration value.
//   - error: An ErrDurationOverflow if the negation overflows (specifically for math.MinInt64).
func (d YearMonthDuration) CheckedNeg() (YearMonthDuration, error) {
	// Implementation Note: Integer overflow prevention
	// Negating math.MinInt64 causes an overflow in two's complement representation since the absolute value exceeds
	// math.MaxInt64.
	if d.months == math.MinInt64 {
		return YearMonthDuration{}, ErrDurationOverflow
	}
	return YearMonthDuration{months: -d.months}, nil
}

// IsIdenticalWith checks if this YearMonthDuration is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
// - bool: True if the other value is a YearMonthDuration and represents the exact same number of months, false
// otherwise.
func (d YearMonthDuration) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(YearMonthDuration); ok {
		return d.months == o.months
	}
	return false
}

// Compare evaluates the order relation of this YearMonthDuration against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26.2)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An ErrDurationOverflow if the other value is incomparable.
func (d YearMonthDuration) Compare(other XSDValue) (int, error) {
	if o, ok := other.(YearMonthDuration); ok {
		if d.months < o.months {
			return -1, nil
		}
		if d.months > o.months {
			return 1, nil
		}
		return 0, nil
	}
	return 0, ErrDurationOverflow
}

// String returns the canonical lexical representation of the YearMonthDuration value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26.1)
//
// Returns:
//   - string: The canonical representation string (e.g., "P0M" or matching the pattern PnYnM).
func (d YearMonthDuration) String() string {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.26.1)
	// The canonical representation of the yearMonthDuration of zero length is handled as "P0M" to align with duration
	// serialization.
	if d.months == 0 {
		return "P0M"
	}
	dur, _ := NewDuration(d.months, DefaultDecimal())
	return dur.String()
}

// ParseYearMonthDuration parses a string literal matching the lexical representation of xsd:yearMonthDuration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.26.1)
//
// Parameters:
//   - s: The raw string literal to be parsed.
//
// Returns:
//   - YearMonthDuration: The parsed YearMonthDuration value on success.
//   - error: A ParseDurationError if the lexical format is violated or contains invalid components.
func ParseYearMonthDuration(s string) (YearMonthDuration, error) {
	parts, _, err := durationParts(s)
	if err != nil {
		return YearMonthDuration{}, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.26.1)
	// The lexical representation of yearMonthDuration is restricted to year and month components only; days, hours,
	// minutes, or seconds are prohibited.
	if parts.dayTime != nil {
		return YearMonthDuration{}, newParseDurationError(
			"there must not be any day or time component in a yearMonthDuration",
		)
	}
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.26.1)
	// A yearMonthDuration lexical representation must contain at least one of the year or month components.
	if parts.yearMonth == nil {
		return YearMonthDuration{}, newParseDurationError("no year and month values found")
	}
	return NewYearMonthDuration(*parts.yearMonth), nil
}
