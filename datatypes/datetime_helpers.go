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
	"math"
	"math/big"
)

// DateTimeSevenPropertyModel represents the seven-property model for date/time values as defined in the governing
// specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7 & Appendix D.2.1)
//
// Representation:
// An internal property model consisting of seven components: year, month, day,
// hour, minute, second, and timezoneOffset.
type DateTimeSevenPropertyModel struct {
	// Year represents the year component. It can be negative or positive.
	Year *int64

	// Month represents the month component, ranging from 1 to 12.
	Month *uint8

	// Day represents the day of the month component, ranging from 1 to 31.
	Day *uint8

	// Hour represents the hour component, ranging from 0 to 24.
	Hour *uint8

	// Minute represents the minute component, ranging from 0 to 59.
	Minute *uint8

	// Second represents the second component as a high-precision Decimal.
	Second *Decimal

	// TimezoneOffset represents the timezone offset from UTC.
	TimezoneOffset *TimezoneOffset
}

// validateMonth validates the month component of a DateTimeSevenPropertyModel.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - month: A pointer to the month value to validate.
//
// Returns:
//   - error: An error if the month is outside the range [1, 12].
func validateMonth(month *uint8) error {
	if month == nil {
		return nil
	}
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1)
	// ·month·: an integer between 1 and 12 inclusive.
	if *month < 1 || *month > monthsPerYear {
		return newParseDateTimeError(fmt.Sprintf("month must be between 1 and 12, got %d", *month))
	}
	return nil
}

// validateDay validates the day component of a DateTimeSevenPropertyModel.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - year: An optional pointer to the year value.
//   - month: An optional pointer to the month value.
//   - day: A pointer to the day value to validate.
//
// Returns:
//   - error: An error if the day is invalid for the given month and year.
func validateDay(year *int64, month, day *uint8) error {
	if day == nil {
		return nil
	}
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1)
	// ·day·: an integer between 1 and 31 inclusive, possibly restricted further depending on ·month· and ·year·.
	if *day < 1 {
		return newParseDateTimeError(fmt.Sprintf("day must be between 1 and 31, got %d", *day))
	}
	if month != nil {
		maxDays := daysInMonth(year, int64(*month))
		if *day > maxDays {
			return newParseDateTimeError(fmt.Sprintf("%d is not a valid day of month %d", *day, *month))
		}
		return nil
	}
	if *day > maxDaysInMonth {
		return newParseDateTimeError(fmt.Sprintf("day must be between 1 and 31, got %d", *day))
	}
	return nil
}

// validateHour validates the hour component of a DateTimeSevenPropertyModel.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1 & Section 3.2.7.1)
//
// Parameters:
//   - hour: A pointer to the hour value to validate.
//   - minute: An optional pointer to the minute value.
//   - second: An optional pointer to the second Decimal value.
//
// Returns:
//   - error: An error if the hour is invalid or exceeds 24:00:00.
func validateHour(hour, minute *uint8, second *Decimal) error {
	if hour == nil {
		return nil
	}
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1 & Section 3.2.7.1)
	// ·hour·: an integer between 0 and 23 inclusive (24 is only permitted at 24:00:00).
	if *hour > hoursPerDay {
		return newParseDateTimeError(fmt.Sprintf("hour must be between 0 and 24, got %d", *hour))
	}
	if *hour < hoursPerDay {
		return nil
	}
	var minuteVal uint8
	if minute != nil {
		minuteVal = *minute
	}
	var secVal Decimal
	if second != nil {
		secVal = *second
	}
	if minuteVal != 0 || secVal.getValue().Sign() != 0 {
		return newParseDateTimeError("times are not allowed to be after 24:00:00")
	}
	return nil
}

// validateMinute validates the minute component of a DateTimeSevenPropertyModel.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - minute: A pointer to the minute value to validate.
//
// Returns:
//   - error: An error if the minute is outside the range [0, 59].
func validateMinute(minute *uint8) error {
	if minute == nil {
		return nil
	}
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1)
	// ·minute·: an integer between 0 and 59 inclusive.
	if *minute > maxMinutesPerHour {
		return newParseDateTimeError(fmt.Sprintf("minute must be between 0 and 59, got %d", *minute))
	}
	return nil
}

// validateSecond validates the second component of a DateTimeSevenPropertyModel.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - second: A pointer to the second Decimal value to validate.
//
// Returns:
//   - error: An error if the second is negative or greater than or equal to 60.
func validateSecond(second *Decimal) error {
	if second == nil {
		return nil
	}
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1)
	// ·second·: a decimal number greater than or equal to 0 and less than 60.
	sixty := NewDecimalFromInt64(secondsPerMinute)
	if second.getValue().Sign() < 0 || second.getValue().Cmp(sixty.getValue()) >= 0 {
		return newParseDateTimeError("seconds must be between 00.0 and 60.0")
	}
	return nil
}

// validateDateTimeModel validates the discrete components of a DateTimeSevenPropertyModel
// against the value-space constraints defined in W3C XSD 1.1 Part 2 (Appendix D.2.1).
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Parameters:
//   - props: A pointer to the DateTimeSevenPropertyModel to validate.
//
// Returns:
//   - error: An error if any component violates the Gregorian calendar rules or value space bounds.
func validateDateTimeModel(props *DateTimeSevenPropertyModel) error {
	if props == nil {
		return nil
	}
	if err := validateMonth(props.Month); err != nil {
		return err
	}
	if err := validateDay(props.Year, props.Month, props.Day); err != nil {
		return err
	}
	if err := validateHour(props.Hour, props.Minute, props.Second); err != nil {
		return err
	}
	if err := validateMinute(props.Minute); err != nil {
		return err
	}
	return validateSecond(props.Second)
}

// timeOnTimeline computes the exact linear timestamp from the seven-property model.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3.4)
//
// Parameters:
//   - props: A pointer to the DateTimeSevenPropertyModel containing the temporal components to map.
//
// Returns:
//   - Decimal: A high-precision decimal representing the timeline coordinate in seconds.
func timeOnTimeline(props *DateTimeSevenPropertyModel) Decimal {
	var yr, mo, da, hr, mi int64
	var se Decimal
	var tzOffset int16

	if props.Year != nil {
		yr = *props.Year - 1
	} else {
		yr = xsdDefaultEpochYear
	}
	if props.Month != nil {
		mo = int64(*props.Month)
	} else {
		mo = monthsPerYear
	}
	if props.Day != nil {
		da = int64(*props.Day) - 1
	} else {
		da = int64(daysInMonth(props.Year, mo)) - 1
	}
	if props.Hour != nil {
		hr = int64(*props.Hour)
	}
	if props.Minute != nil {
		mi = int64(*props.Minute)
	}
	if props.Second != nil {
		se = *props.Second
	} else {
		se = DefaultDecimal()
	}
	if props.TimezoneOffset != nil {
		tzOffset = props.TimezoneOffset.offset
	}

	mi -= int64(tzOffset)

	val := big.NewInt(0)

	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix E.3)
	// Leap year calculations are based on the Gregorian calendar cycles of 4, 100, and 400 years, accumulating month
	// days dynamically based on standard calendar rules.
	val.Add(val, new(big.Int).Mul(big.NewInt(yr), big.NewInt(secondsPerNormalYear)))
	leaps := big.NewInt(0)
	leaps.Sub(leaps, new(big.Int).Div(big.NewInt(yr), big.NewInt(leapYearFactor100)))
	leaps.Add(leaps, new(big.Int).Div(big.NewInt(yr), big.NewInt(leapYearFactor400)))
	leaps.Add(leaps, new(big.Int).Div(big.NewInt(yr), big.NewInt(leapYearFactor4)))
	val.Add(val, new(big.Int).Mul(leaps, big.NewInt(secondsPerDay)))

	var monthDays int64
	for m := int64(1); m < mo; m++ {
		monthDays += int64(daysInMonth(int64Ptr(yr+1), m))
	}
	val.Add(val, new(big.Int).Mul(big.NewInt(monthDays), big.NewInt(secondsPerDay)))

	val.Add(val, new(big.Int).Mul(big.NewInt(da), big.NewInt(secondsPerDay)))
	val.Add(val, new(big.Int).Mul(big.NewInt(hr), big.NewInt(secondsPerHour)))
	val.Add(val, new(big.Int).Mul(big.NewInt(mi), big.NewInt(secondsPerMinute)))

	decVal, _ := ParseDecimal(val.String())
	return decVal.Add(se)
}

// normalizeMonth normalizes a month value that may be out of bounds by carrying over the excess years.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3.1)
//
// Parameters:
//   - yr: The year component to adjust.
//   - mo: The month component to normalize.
//
// Returns:
//   - int64: The adjusted year component.
//   - uint8: The normalized month component.
//   - error: An error of type ErrDateTimeOverflow if the adjustments exceed integer bounds.
func normalizeMonth(yr, mo int64) (int64, uint8, error) {
	if mo > 0 {
		yr += (mo - 1) / monthsPerYear
		mo = (mo-1)%monthsPerYear + 1
	} else {
		yr += mo/monthsPerYear - 1
		mo = monthsPerYear + mo%monthsPerYear
	}
	// Implementation Note: Bounds validation and integer overflow prevention
	// Explicitly checks bounds against math.MaxUint8 and zero to eliminate potential G115 integer conversion overflows.
	if yr > math.MaxInt64/2 || yr < math.MinInt64/2 || mo < 0 || mo > math.MaxUint8 {
		return 0, 0, ErrDateTimeOverflow
	}
	return yr, uint8(mo), nil
}

// dateTimePlusDuration adds a Duration to a DateTimeSevenPropertyModel.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3.3)
//
// Parameters:
//   - du: The Duration value to be added.
//   - dt: A pointer to the DateTimeSevenPropertyModel representing the starting instant.
//
// Returns:
//   - *DateTimeSevenPropertyModel: A pointer to the new model representing the resulting instant.
//   - error: An error of type ErrDateTimeOverflow if the calculation overflows the supported timeline.
func dateTimePlusDuration(du Duration, dt *DateTimeSevenPropertyModel) (*DateTimeSevenPropertyModel, error) {
	yr := int64(1)
	if dt.Year != nil {
		yr = *dt.Year
	}
	mo := uint8(1)
	if dt.Month != nil {
		mo = *dt.Month
	}
	da := uint8(1)
	if dt.Day != nil {
		da = *dt.Day
	}
	hr := uint8(0)
	if dt.Hour != nil {
		hr = *dt.Hour
	}
	mi := uint8(0)
	if dt.Minute != nil {
		mi = *dt.Minute
	}
	se := DefaultDecimal()
	if dt.Second != nil {
		se = *dt.Second
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix E.3.3)
	// Add the yearMonth and dayTime components of a duration to a dateTime value following the specified algorithm.

	// Step 1: Add yearMonth components and normalize
	// Add the yearMonthDuration components to the year and month, adjusting the year accordingly.
	newYr, normMo, err := normalizeMonth(yr, int64(mo)+du.AllMonths())
	if err != nil {
		return nil, err
	}

	// Step 2: Pin the day
	// Ensure the day does not exceed the maximum valid day of the resulting year and month.
	newDa := minUint8(da, daysInMonth(int64Ptr(newYr), int64(normMo)))

	tempProps := DateTimeSevenPropertyModel{
		Year: int64Ptr(newYr), Month: uint8Ptr(normMo), Day: uint8Ptr(newDa),
		Hour: uint8Ptr(hr), Minute: uint8Ptr(mi), Second: decimalPtr(se),
		TimezoneOffset: dt.TimezoneOffset,
	}
	ts, _ := newTimestamp(&tempProps)

	// Step 3: Add dayTime components
	// Convert to a timeline coordinate and add the dayTimeDuration component.
	ts = ts.addSeconds(du.AllSeconds())

	finalYr, finalMo, finalDa := ts.yearMonthDay()
	return &DateTimeSevenPropertyModel{
		Year:           mapIf(dt.Year, finalYr),
		Month:          mapIf(dt.Month, finalMo),
		Day:            mapIf(dt.Day, finalDa),
		Hour:           mapIf(dt.Hour, ts.hour()),
		Minute:         mapIf(dt.Minute, ts.minute()),
		Second:         mapIf(dt.Second, ts.second()),
		TimezoneOffset: dt.TimezoneOffset,
	}, nil
}

// daysInMonth returns the number of days in the specified month of the given year.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3.2)
//
// Parameters:
//   - y: An optional pointer to the year component to evaluate for leap years.
//   - m: The month component represented as a 1-based integer.
//
// Returns:
//   - uint8: The number of days in the specified month.
func daysInMonth(y *int64, m int64) uint8 {
	switch m {
	case monthFebruary:
		if y != nil {
			year := *y
			if year%leapYearFactor4 != 0 || (year%leapYearFactor100 == 0 && year%leapYearFactor400 != 0) {
				return daysInFebruaryNormal
			}
			return daysInFebruaryLeap
		}
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.12)
		// When the year property is absent (such as in gMonthDay processing), February 29 is treated as valid to
		// accommodate leap years.
		return daysInFebruaryLeap
	case monthApril, monthJune, monthSeptember, monthNovember:
		return daysInShortMonth
	default:
		return daysInLongMonth
	}
}

// minUint8 returns the minimum of two uint8 values.
//
// Parameters:
//   - a: The first uint8 value to compare.
//   - b: The second uint8 value to compare.
//
// Returns:
//   - uint8: The smaller of the two input values.
func minUint8(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}

// int64Ptr is a helper that returns a pointer to the given int64 value.
//
// Parameters:
//   - i: The int64 value to wrap in a pointer.
//
// Returns:
//   - *int64: A pointer to the provided value.
func int64Ptr(i int64) *int64 {
	return &i
}

// uint8Ptr is a helper that returns a pointer to the given uint8 value.
//
// Parameters:
//   - u: The uint8 value to wrap in a pointer.
//
// Returns:
//   - *uint8: A pointer to the provided value.
func uint8Ptr(u uint8) *uint8 {
	return &u
}

// decimalPtr is a helper that returns a pointer to the given Decimal value.
//
// Parameters:
//   - d: The Decimal value to wrap in a pointer.
//
// Returns:
//   - *Decimal: A pointer to the provided value.
func decimalPtr(d Decimal) *Decimal {
	return &d
}

// mapIf returns a pointer to the given value if the original field is not nil.
//
// Parameters:
//   - originalField: An optional pointer representing the original field.
//   - value: The new value to be wrapped if originalField is present.
//
// Returns:
//   - *T: A pointer to the new value if the original field is not nil, otherwise nil.
func mapIf[T any, O any](originalField *O, value T) *T {
	if originalField == nil {
		return nil
	}
	return &value
}
