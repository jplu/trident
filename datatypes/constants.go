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
	"time"
)

const (
	// secondsPerMinute represents the number of seconds in one minute.
	secondsPerMinute = 60

	// minutesPerHour represents the number of minutes in one hour.
	minutesPerHour = 60

	// secondsPerHour represents the number of seconds in one hour.
	secondsPerHour = 3600

	// hoursPerDay represents the number of hours in a single day.
	hoursPerDay = 24

	// secondsPerDay represents the number of seconds in a single day.
	secondsPerDay = 86400

	// monthsPerYear represents the number of months in a single year.
	monthsPerYear = 12

	// daysInLeapYear represents the number of days in a Gregorian leap year.
	daysInLeapYear = 366

	// daysInNormalYear represents the number of days in a standard non-leap year.
	daysInNormalYear = 365

	// daysIn400Years represents the total number of days in a 400-year Gregorian cycle.
	daysIn400Years = 146097

	// daysIn100Years represents the total number of days in a 100-year Gregorian cycle.
	daysIn100Years = 36524

	// daysIn4Years represents the total number of days in a 4-year Gregorian leap cycle.
	daysIn4Years = 1461

	// bitSize8 represents the bit width of an 8-bit integer.
	// It is used for validating values of xsd:byte and xsd:unsignedByte in accordance with W3C XSD 1.1 Part 2 (Section
	// 3.4.19) and W3C XSD 1.1 Part 2 (Section 3.4.24).
	bitSize8 = 8

	// bitSize16 represents the bit width of a 16-bit integer.
	// It is used for validating values of xsd:short and xsd:unsignedShort in accordance with W3C XSD 1.1 Part 2
	// (Section 3.4.18) and W3C XSD 1.1 Part 2 (Section 3.4.23).
	bitSize16 = 16

	// bitSize32 represents the bit width of a 32-bit integer.
	// It is used for validating values of xsd:int and xsd:unsignedInt in accordance with W3C XSD 1.1 Part 2 (Section
	// 3.4.17) and W3C XSD 1.1 Part 2 (Section 3.4.22).
	bitSize32 = 32

	// bitSize64 represents the bit width of a 64-bit integer.
	// It is used for validating values of xsd:long and xsd:unsignedLong in accordance with W3C XSD 1.1 Part 2 (Section
	// 3.4.16) and W3C XSD 1.1 Part 2 (Section 3.4.21).
	bitSize64 = 64

	// maxDaysInMonth represents the maximum possible number of days in any month.
	// It is used for day-of-month validation in accordance with W3C XSD 1.1 Part 2 (Section 3.3.7.1).
	maxDaysInMonth = 31

	// maxMinutesPerHour represents the upper bound index for minutes within an hour.
	// It is used for time validation in accordance with W3C XSD 1.1 Part 2 (Section 3.3.8.1).
	maxMinutesPerHour = 59

	// secondsPerNormalYear represents the total number of seconds in a standard non-leap year.
	// It is used for timeline calculations in accordance with W3C XSD 1.1 Part 2 (Appendix E.3.4).
	secondsPerNormalYear = 31536000

	// leapYearFactor4 represents the 4-year leap year rule divisor in accordance with W3C XSD 1.1 Part 2 (Appendix
	// E.3).
	leapYearFactor4 = 4

	// leapYearFactor100 represents the 100-year leap year rule divisor in accordance with W3C XSD 1.1 Part 2 (Appendix
	// E.3).
	leapYearFactor100 = 100

	// leapYearFactor400 represents the 400-year leap year rule divisor in accordance with W3C XSD 1.1 Part 2 (Appendix
	// E.3).
	leapYearFactor400 = 400

	// nanosecondsScale represents the decimal exponent scaling factor for nanoseconds.
	nanosecondsScale = 9

	// base5 represents the number 5, used for base-5 prime factorization in decimal termination checks.
	// It is used to analyze terminating decimal values in accordance with W3C XSD 1.1 Part 2 (Appendix E.2.2).
	base5 = 5

	// builderGrowthFactor represents a pre-allocation growth factor for regex translations.
	builderGrowthFactor = 2

	// maxTimezoneHours represents the absolute timezone offset limit in hours.
	// This is bounded in accordance with W3C XSD 1.1 Part 2 (Section 3.2.7.3).
	maxTimezoneHours = 14

	// maxTimezoneOffsetMinutes represents the absolute maximum timezone offset limit in minutes.
	// This is bounded in accordance with W3C XSD 1.1 Part 2 (Section 3.2.7.3).
	maxTimezoneOffsetMinutes = maxTimezoneHours * minutesPerHour

	// minTimezoneOffsetMinutes represents the absolute minimum timezone offset limit in minutes.
	// This is bounded in accordance with W3C XSD 1.1 Part 2 (Section 3.2.7.3).
	minTimezoneOffsetMinutes = -maxTimezoneOffsetMinutes

	// xsdDefaultEpochYear represents the default epoch year used during relative timeline comparisons.
	// It is specified in the timeline origin algorithm in W3C XSD 1.1 Part 2 (Appendix E.3.4).
	xsdDefaultEpochYear = 1971

	// base10 represents the decimal base multiplier.
	base10 = 10

	// maxFloatPrecision represents the maximum decimal digit formatting precision.
	// It safeguards against infinite loops during terminating decimal division fallbacks.
	maxFloatPrecision = 340

	// monthFebruary represents the numeric representation of February.
	monthFebruary = 2

	// monthMarch represents the numeric representation of March.
	monthMarch = 3

	// monthApril represents the numeric representation of April.
	monthApril = 4

	// monthJune represents the numeric representation of June.
	monthJune = 6

	// monthJuly represents the numeric representation of July.
	monthJuly = 7

	// monthSeptember represents the numeric representation of September.
	monthSeptember = 9

	// monthNovember represents the numeric representation of November.
	monthNovember = 11

	// daysInFebruaryNormal represents the number of days in February in a standard year.
	daysInFebruaryNormal = 28

	// daysInFebruaryLeap represents the number of days in February in a leap year.
	daysInFebruaryLeap = 29

	// daysInShortMonth represents the number of days in standard 30-day months.
	daysInShortMonth = 30

	// daysInLongMonth represents the number of days in standard 31-day months.
	daysInLongMonth = 31

	// minYearDigits represents the minimum number of digits allowed for the year component.
	// It is enforced in accordance with W3C XSD 1.1 Part 2 (Section 3.2.7.1).
	minYearDigits = 4

	// twoDigits represents the exact number of digits required for month, day, and time components.
	twoDigits = 2

	// maxRune16 represents the maximum value of a 16-bit Unicode character.
	maxRune16 = 0xFFFF

	// defaultTimeYear represents the default year applied when parsing isolated xsd:time values.
	// This recovery date is defined in accordance with W3C XSD 1.1 Part 2 (Section 3.3.8.1).
	defaultTimeYear = 1972

	// defaultTimeMonth represents the default month applied when parsing isolated xsd:time values.
	// This recovery date is defined in accordance with W3C XSD 1.1 Part 2 (Section 3.3.8.1).
	defaultTimeMonth = 12

	// defaultTimeDay represents the default day of the month applied when parsing isolated xsd:time values.
	// This recovery date is defined in accordance with W3C XSD 1.1 Part 2 (Section 3.3.8.1).
	defaultTimeDay = 31

	// anchorYear1696 represents the year 1696, used as part of the first comparison anchor in accordance with W3C XSD
	// 1.1 Part 2 (Section 3.3.6.1).
	anchorYear1696 = 1696

	// anchorYear1697 represents the year 1697, used as part of the second comparison anchor in accordance with W3C XSD
	// 1.1 Part 2 (Section 3.3.6.1).
	anchorYear1697 = 1697

	// anchorYear1903 represents the year 1903, used as part of the third and fourth comparison anchors in accordance
	// with W3C XSD 1.1 Part 2 (Section 3.3.6.1).
	anchorYear1903 = 1903
)

const (
	// strZeroDuration represents the canonical zero duration string in accordance with W3C XSD 1.1 Part 2 (Section
	// 3.3.6.2).
	strZeroDuration = "PT0S"

	// strInvalidDurationOppositeSigns represents the invalid duration string returned when components have opposite
	// signs in accordance with W3C XSD 1.1 Part 2 (Section 3.3.6).
	strInvalidDurationOppositeSigns = "invalid-duration-opposite-signs"

	// strINF represents the standard lexical representation of positive infinity in accordance with W3C XSD 1.1 Part 2
	// (Section 3.3.4.2).
	strINF = "INF"

	// strPosINF represents the alternative lexical representation of positive infinity with a leading sign in
	// accordance with W3C XSD 1.1 Part 2 (Section 3.3.4.2).
	strPosINF = "+INF"

	// strNegINF represents the standard lexical representation of negative infinity in accordance with W3C XSD 1.1 Part
	// 2 (Section 3.3.4.2).
	strNegINF = "-INF"

	// strNaN represents the standard lexical representation of the float "not-a-number" value in accordance with W3C
	// XSD 1.1 Part 2 (Section 3.3.4.2).
	strNaN = "NaN"

	// primitiveString represents the local string name identifier of the primitive string type.
	primitiveString = "string"
)

// getAnchorDates returns the four comparison anchor dates required to evaluate partially ordered durations.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2.1)
//
// Returns:
// - []Timestamp: a slice containing the four specific anchor date timestamps (1696-09-01T00:00:00Z,
// 1697-02-01T00:00:00Z, 1903-03-01T00:00:00Z, and 1903-07-01T00:00:00Z) used to mathematically evaluate the partial
// order.
func getAnchorDates() []Timestamp {
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1)
	// The order relation of durations requires four specific anchor dates (1696-09-01T00:00:00Z, 1697-02-01T00:00:00Z,
	// 1903-03-01T00:00:00Z, and 1903-07-01T00:00:00Z) to mathematically evaluate the partial order.
	return []Timestamp{
		MustNewTimestamp(&DateTimeSevenPropertyModel{
			Year:   int64Ptr(anchorYear1696),
			Month:  uint8Ptr(monthSeptember),
			Day:    uint8Ptr(1),
			Hour:   uint8Ptr(0),
			Minute: uint8Ptr(0),
			Second: decimalPtr(DefaultDecimal()),
		}),
		MustNewTimestamp(&DateTimeSevenPropertyModel{
			Year:   int64Ptr(anchorYear1697),
			Month:  uint8Ptr(monthFebruary),
			Day:    uint8Ptr(1),
			Hour:   uint8Ptr(0),
			Minute: uint8Ptr(0),
			Second: decimalPtr(DefaultDecimal()),
		}),
		MustNewTimestamp(&DateTimeSevenPropertyModel{
			Year:   int64Ptr(anchorYear1903),
			Month:  uint8Ptr(monthMarch),
			Day:    uint8Ptr(1),
			Hour:   uint8Ptr(0),
			Minute: uint8Ptr(0),
			Second: decimalPtr(DefaultDecimal()),
		}),
		MustNewTimestamp(&DateTimeSevenPropertyModel{
			Year:   int64Ptr(anchorYear1903),
			Month:  uint8Ptr(monthJuly),
			Day:    uint8Ptr(1),
			Hour:   uint8Ptr(0),
			Minute: uint8Ptr(0),
			Second: decimalPtr(DefaultDecimal()),
		}),
	}
}

// getTimezoneMaxDuration returns the maximum allowed timezone offset representation as a DayTimeDuration instance.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.3)
//
// Returns:
//   - DayTimeDuration: a DayTimeDuration representing the maximum 14-hour timezone offset limit.
func getTimezoneMaxDuration() DayTimeDuration {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.7.3)
	// The timezone offset of a temporal value must be between -14:00 and +14:00 inclusive. This function constructs a
	// positive duration representation corresponding to that absolute boundary limit.
	d := NewDayTimeDurationFromStd(time.Duration(maxTimezoneHours) * time.Hour)
	return d
}
