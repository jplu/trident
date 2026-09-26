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
	"math/big"
	"strings"
)

// Duration represents the duration primitive datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A six-property model split into two components: a yearMonth component (expressed as an integer number of months) and
// a dayTime component (expressed as a decimal number of seconds). Both components must share the same sign (either both
// non-negative or both non-positive).
type Duration struct {
	// YearMonth represents the year and month components of the duration.
	yearMonth YearMonthDuration
	// DayTime represents the day, hour, minute, and second components of the duration.
	dayTime DayTimeDuration
}

// NewDuration constructs a new Duration from the given number of months and seconds.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Parameters:
//   - months: The total number of months representing the year-month component.
//   - seconds: The arbitrary-precision decimal representing total seconds for the day-time component.
//
// Returns:
//   - Duration: The constructed Duration instance.
//   - error: An error if the year-month and day-time components have opposite signs.
func NewDuration(months int64, seconds Decimal) (Duration, error) {
	ym := NewYearMonthDuration(months)
	dt := NewDayTimeDuration(seconds)
	ymPositive := ym.months > 0
	dtPositive := dt.seconds.getValue().Sign() > 0

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6)
	// Both yearMonthDuration and dayTimeDuration components must have the same sign (either both non-negative or both
	// non-positive). A duration cannot have opposite signs (e.g., positive years with negative days).
	if (ym.months != 0 && dt.seconds.getValue().Sign() != 0) && (ymPositive != dtPositive) {
		return Duration{}, ErrOppositeSignInDurationComponents
	}
	return Duration{yearMonth: ym, dayTime: dt}, nil
}

// ParseDuration parses a string literal matching the lexical representation of xsd:duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.2)
//
// Parameters:
//   - s: The raw string literal to be parsed.
//
// Returns:
//   - Duration: The parsed Duration value if successful.
//   - error: An error of type ParseDurationError if the lexical format is violated.
func ParseDuration(s string) (Duration, error) {
	parts, _, err := durationParts(s)
	if err != nil {
		return Duration{}, err
	}
	if parts.yearMonth == nil && parts.dayTime == nil {
		return Duration{}, newParseDurationError("empty duration")
	}
	var months int64
	if parts.yearMonth != nil {
		months = *parts.yearMonth
	}
	seconds := DefaultDecimal()
	if parts.dayTime != nil {
		seconds = *parts.dayTime
	}
	return NewDuration(months, seconds)
}

// Years returns the year component of the Duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//
// Returns:
//   - int64: The number of whole years represented by the year-month component.
func (d Duration) Years() int64 {
	return d.yearMonth.Years()
}

// Months returns the month component of the Duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//
// Returns:
//   - int64: The remaining months of the year-month component after subtracting whole years.
func (d Duration) Months() int64 {
	return d.yearMonth.Months()
}

// Days returns the day component of the Duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//
// Returns:
//   - int64: The number of whole days represented by the day-time component.
func (d Duration) Days() int64 {
	return d.dayTime.Days()
}

// Hours returns the hour component of the Duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//
// Returns:
//   - int64: The number of whole hours represented by the day-time component after subtracting whole days.
func (d Duration) Hours() int64 {
	return d.dayTime.Hours()
}

// Minutes returns the minute component of the Duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//
// Returns:
//   - int64: The number of whole minutes represented by the day-time component after subtracting whole hours.
func (d Duration) Minutes() int64 {
	return d.dayTime.Minutes()
}

// Seconds returns the second component of the Duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//
// Returns:
// - Decimal: The high-precision decimal representing seconds and fractional seconds of the day-time component after
// subtracting whole minutes.
func (d Duration) Seconds() Decimal {
	return d.dayTime.Seconds()
}

// AllMonths returns the total number of months in the year-month component of the Duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//
// Returns:
//   - int64: The total number of months represented.
func (d Duration) AllMonths() int64 {
	return d.yearMonth.AllMonths()
}

// AllSeconds returns the total number of seconds in the day-time component of the Duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//
// Returns:
//   - Decimal: The total number of seconds represented as a high-precision decimal.
func (d Duration) AllSeconds() Decimal {
	return d.dayTime.AsSeconds()
}

// CheckedAdd performs an overflow-checked addition of two Duration values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Parameters:
//   - rhs: The Duration value to add.
//
// Returns:
//   - Duration: The resulting sum of the two Duration values.
//   - error: An error if the addition exceeds the bounds of either component.
func (d Duration) CheckedAdd(rhs Duration) (Duration, error) {
	ym, err := d.yearMonth.CheckedAdd(rhs.yearMonth)
	if err != nil {
		return Duration{}, err
	}
	dt := d.dayTime.Add(rhs.dayTime)
	return NewDuration(ym.months, dt.seconds)
}

// CheckedSub performs an overflow-checked subtraction of two Duration values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Parameters:
//   - rhs: The Duration value to subtract.
//
// Returns:
//   - Duration: The resulting difference of the two Duration values.
//   - error: An error if the subtraction exceeds the bounds of either component.
func (d Duration) CheckedSub(rhs Duration) (Duration, error) {
	ym, err := d.yearMonth.CheckedSub(rhs.yearMonth)
	if err != nil {
		return Duration{}, err
	}
	dt := d.dayTime.Sub(rhs.dayTime)
	return NewDuration(ym.months, dt.seconds)
}

// CheckedNeg performs an overflow-checked negation of the Duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Parameters:
//
// Returns:
//   - Duration: The negated Duration value.
//   - error: An error if negating either component causes an integer or decimal overflow.
func (d Duration) CheckedNeg() (Duration, error) {
	ym, err := d.yearMonth.CheckedNeg()
	if err != nil {
		return Duration{}, err
	}
	dt := d.dayTime.Neg()
	return Duration{yearMonth: ym, dayTime: dt}, nil
}

// IsIdenticalWith checks if this Duration is strictly identical to another XSDValue according to the identity criteria
// defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.4)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - bool: True if the yearMonthDuration and dayTimeDuration components are respectively identical.
func (d Duration) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Duration); ok {
		return d.yearMonth.IsIdenticalWith(o.yearMonth) && d.dayTime.IsIdenticalWith(o.dayTime)
	}
	return false
}

// Compare evaluates the order relation of this Duration against another XSDValue according to the partial order defined
// in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - int: -1 if less, 0 if equal, 1 if greater.
//   - error: An error if the values are incomparable or if the comparison is indeterminate.
func (d Duration) Compare(other XSDValue) (int, error) {
	o, ok := other.(Duration)
	if !ok {
		return 0, ErrDurationOverflow
	}

	dates := getAnchorDates()

	var firstResult int
	for i, dt := range dates {
		props1, _ := dateTimePlusDuration(d, dt.ToProperties())
		d1, _ := newTimestamp(props1)

		props2, _ := dateTimePlusDuration(o, dt.ToProperties())
		d2, _ := newTimestamp(props2)

		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1)
		// The ordering of durations is a partial order. To determine relation determinacy, we must test the durations
		// against four specific anchor dateTimes. If comparisons across all four anchors yield the same result, the
		// relation is determinate; otherwise, the relation is indeterminate.
		res, _ := d1.compare(d2)

		if i == 0 {
			firstResult = res
		} else if res != firstResult {
			return 0, ErrDurationOverflow
		}
	}
	return firstResult, nil
}

// String returns the canonical lexical representation of the Duration value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical lexical representation string of the duration.
func (d Duration) String() string {
	ym := d.yearMonth.months
	ss := d.dayTime.seconds

	if (ym < 0 && ss.IsPositive()) || (ym > 0 && ss.IsNegative()) {
		return strInvalidDurationOppositeSigns
	}

	var sb strings.Builder
	isNegative := ym < 0 || ss.IsNegative()
	if isNegative {
		sb.WriteString("-")
	}
	sb.WriteString("P")

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6.2)
	// The canonical representation for a zero-length duration is "PT0S".
	if ym == 0 && ss.getValue().Sign() == 0 {
		return strZeroDuration
	}

	formatDurationYM(&sb, ym, isNegative)
	formatDurationDT(&sb, ss)

	return sb.String()
}

// formatDurationYM serializes the year and month components of the duration to a strings.Builder.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.2)
//
// Parameters:
//   - sb: The strings.Builder pointer to write the serialized representation to.
//   - ym: The absolute year-month months count.
//   - isNegative: A boolean indicating if the duration is negative.
//
// Returns:.
func formatDurationYM(sb *strings.Builder, ym int64, isNegative bool) {
	absYm := ym
	if isNegative && ym < 0 {
		absYm = -ym
	}
	y := absYm / monthsPerYear
	m := absYm % monthsPerYear
	if y != 0 {
		fmt.Fprintf(sb, "%dY", y)
	}
	if m != 0 {
		fmt.Fprintf(sb, "%dM", m)
	}
}

// formatDurationDT serializes the day, hour, minute, and second components of the duration to a strings.Builder.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.2)
//
// Parameters:
//   - sb: The strings.Builder pointer to write the serialized representation to.
//   - ss: The high-precision decimal representing total seconds for the day-time component.
//
// Returns:.
func formatDurationDT(sb *strings.Builder, ss Decimal) {
	absSs := ss.Abs()
	sInt := absSs.asI128()

	dayVal := new(big.Int).Quo(sInt, big.NewInt(secondsPerDay))
	sIntRemDay := new(big.Int).Rem(sInt, big.NewInt(secondsPerDay))
	hourVal := new(big.Int).Quo(sIntRemDay, big.NewInt(secondsPerHour))
	sIntRemHour := new(big.Int).Rem(sInt, big.NewInt(secondsPerHour))
	minVal := new(big.Int).Quo(sIntRemHour, big.NewInt(secondsPerMinute))

	if dayVal.Sign() != 0 {
		fmt.Fprintf(sb, "%sD", dayVal.String())
	}

	secRem, _ := absSs.CheckedRemEuclid(NewDecimalFromInt64(secondsPerMinute))
	hasTimePart := hourVal.Sign() != 0 || minVal.Sign() != 0 || secRem.getValue().Sign() != 0

	if hasTimePart {
		sb.WriteString("T")
		if hourVal.Sign() != 0 {
			fmt.Fprintf(sb, "%sH", hourVal.String())
		}
		if minVal.Sign() != 0 {
			fmt.Fprintf(sb, "%sM", minVal.String())
		}
		if secRem.getValue().Sign() != 0 {
			fmt.Fprintf(sb, "%sS", secRem.String())
		}
	}
}
