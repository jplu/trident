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
	"strconv"
	"strings"
)

// durationPartsInternal represents the raw parsed components of an XSD duration as defined in the governing
// specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Representation:
// A structure containing intermediate parsed components of a duration value, separated into year-month (months) and
// day-time (seconds) elements.
type durationPartsInternal struct {
	// yearMonth holds the total months computed from the years (Y) and months (M) components.
	yearMonth *int64
	// dayTime holds the total seconds computed from days (D), hours (H), minutes (M), and seconds (S) components.
	dayTime *Decimal
}

const (
	// stateStart represents the parser state after 'P' but before any unit or time separator.
	stateStart = iota
	// stateAfterYear represents the parser state after parsing the years ('Y') component.
	stateAfterYear
	// stateAfterMonth represents the parser state after parsing the months ('M') component (pre-T).
	stateAfterMonth
	// stateAfterDay represents the parser state after parsing the days ('D') component.
	stateAfterDay
	// stateAfterT represents the parser state immediately after encountering the time separator ('T').
	stateAfterT
	// stateAfterHour represents the parser state after parsing the hours ('H') component.
	stateAfterHour
	// stateAfterMinute represents the parser state after parsing the minutes ('M') component (post-T).
	stateAfterMinute
	// stateAfterSecond represents the parser state after parsing the seconds ('S') component.
	stateAfterSecond
)

// durationParts parses the input duration string into its year-month and day-time components.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//   - input: The raw string literal representing the duration to be parsed.
//
// Returns:
//   - durationPartsInternal: The parsed year-month and day-time components.
//   - string: The remaining unparsed suffix of the input string.
//   - error: An error if the duration string violates the lexical space rules.
func durationParts(input string) (durationPartsInternal, string, error) {
	isNegative := false
	if strings.HasPrefix(input, "-") {
		isNegative = true
		input = input[1:]
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6)
	// The duration lexical space requires a leading 'P' character, preceded by an optional minus sign to designate a
	// negative duration.
	if !strings.HasPrefix(input, "P") {
		return durationPartsInternal{}, input, newParseDurationError("durations must start with 'P'")
	}
	input = input[1:]

	var ym *int64
	var dt *Decimal
	state := stateStart

	for len(input) > 0 {
		if input[0] == 'T' {
			// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6)
			// The time designator 'T' must precede all time-related units (hours, minutes, seconds) and must appear
			// exactly once if any time components are present.
			if state >= stateAfterT {
				return durationPartsInternal{}, input, newParseDurationError("duplicated time separator 'T'")
			}
			state = stateAfterT
			input = input[1:]
			continue
		}

		numStr, rem := decimalPrefix(input)
		if numStr == "" || len(rem) == 0 {
			return durationPartsInternal{}, input, newParseDurationError("invalid duration component")
		}

		unit := rem[0]
		input = rem[1:]

		var err error
		ym, dt, state, err = processDurationUnit(unit, numStr, state, isNegative, ym, dt)
		if err != nil {
			return durationPartsInternal{}, input, err
		}
	}

	return durationPartsInternal{yearMonth: ym, dayTime: dt}, input, nil
}

// processDurationUnit parses and validates a single duration component.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Parameters:
//   - unit: The byte designator of the duration component (e.g., 'Y', 'M', 'D', 'H', 'S').
//   - numStr: The string representing the numerical value of the component.
//   - state: The current state of the parser state machine.
//   - isNeg: A boolean indicating whether the overall duration is negative.
//   - ym: A pointer to the accumulated year-month duration value in months.
//   - dt: A pointer to the accumulated day-time duration value in seconds.
//
// Returns:
//   - *int64: The updated year-month duration value in months.
//   - *Decimal: The updated day-time duration value in seconds.
//   - int: The next parser state.
//   - error: An error if the component is out of order or represents a numerical overflow.
func processDurationUnit(
	unit byte,
	numStr string,
	state int,
	isNeg bool,
	ym *int64,
	dt *Decimal,
) (*int64, *Decimal, int, error) {
	switch unit {
	case 'Y':
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6)
		// Component ordering must strictly be Y, M, D, T, H, M, S. Years ('Y') must be the absolute first component if
		// present.
		if state >= stateAfterYear {
			return ym, dt, state, newParseDurationError("year out of order")
		}
		val, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return ym, dt, state, errParseOverflow
		}
		newYm, err := addDurationYM(ym, val*monthsPerYear, isNeg)
		return newYm, dt, stateAfterYear, err

	case 'M':
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6)
		// The 'M' designator represents months if it occurs before the 'T' separator, and minutes if it occurs after
		// the 'T' separator.
		if state >= stateAfterT {
			return processTimeMinute(numStr, state, isNeg, ym, dt)
		}
		if state >= stateAfterMonth {
			return ym, dt, state, newParseDurationError("month out of order")
		}
		val, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return ym, dt, state, errParseOverflow
		}
		newYm, err := addDurationYM(ym, val, isNeg)
		return newYm, dt, stateAfterMonth, err

	case 'D':
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6)
		// Days ('D') must be parsed before the time separator 'T' is reached.
		if state >= stateAfterDay {
			return ym, dt, state, newParseDurationError("day out of order")
		}
		newDt, err := processDayTimeDecimal(numStr, isNeg, dt, secondsPerDay)
		return ym, newDt, stateAfterDay, err

	case 'H':
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6)
		// Hours ('H') can only be parsed after the 'T' time separator.
		switch {
		case state < stateAfterT:
			return ym, dt, state, newParseDurationError("hour must follow 'T'")
		case state >= stateAfterHour:
			return ym, dt, state, newParseDurationError("hour out of order")
		}
		newDt, err := processDayTimeDecimal(numStr, isNeg, dt, secondsPerHour)
		return ym, newDt, stateAfterHour, err

	case 'S':
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6)
		// Seconds ('S') can only be parsed after the 'T' time separator.
		switch {
		case state < stateAfterT:
			return ym, dt, state, newParseDurationError("second must follow 'T'")
		case state >= stateAfterSecond:
			return ym, dt, state, newParseDurationError("second out of order")
		}
		val, err := ParseDecimal(numStr)
		if err != nil {
			return ym, dt, state, errParseOverflow
		}
		newDt := addDurationDT(dt, val, isNeg)
		return ym, newDt, stateAfterSecond, err

	default:
		return ym, dt, state, newParseDurationError("unexpected type character")
	}
}

// processTimeMinute processes the minutes ('M') component when it appears in the time portion.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Parameters:
//   - numStr: The string representing the numerical value of the minute component.
//   - state: The current state of the parser state machine.
//   - isNeg: A boolean indicating whether the overall duration is negative.
//   - ym: A pointer to the accumulated year-month duration value in months.
//   - dt: A pointer to the accumulated day-time duration value in seconds.
//
// Returns:
//   - *int64: The unchanged year-month duration value.
//   - *Decimal: The updated day-time duration value with minutes added.
//   - int: The next parser state (stateAfterMinute).
//   - error: An error if parsing the minute component fails or if the component is out of order.
func processTimeMinute(numStr string, state int, isNeg bool, ym *int64, dt *Decimal) (*int64, *Decimal, int, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6)
	// Component ordering must strictly be Y, M, D, T, H, M, S. Minutes ('M') in the time section must not be parsed out
	// of order.
	if state >= stateAfterMinute {
		return ym, dt, state, newParseDurationError("minute out of order")
	}
	newDt, err := processDayTimeDecimal(numStr, isNeg, dt, secondsPerMinute)
	return ym, newDt, stateAfterMinute, err
}

// processDayTimeDecimal parses a numeric string value for a day-time duration component.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Parameters:
//   - numStr: The string representing the numerical value of the component.
//   - isNeg: A boolean indicating whether the overall duration is negative.
//   - dt: A pointer to the accumulated day-time duration value.
//   - multiplier: The scaling factor in seconds (e.g., seconds per day, hour, or minute).
//
// Returns:
//   - *Decimal: The updated day-time duration value with the scaled component added.
//   - error: An error if the numerical string contains a fractional part on non-second components, or on overflow.
func processDayTimeDecimal(numStr string, isNeg bool, dt *Decimal, multiplier int64) (*Decimal, error) {
	val, err := ParseDecimal(numStr)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.6)
	// Fractional parts are only permitted on the seconds component. Therefore, any other component (years, months,
	// days, hours, minutes) containing a decimal point is invalid.
	if err != nil || strings.Contains(numStr, ".") {
		return dt, errParseOverflow
	}
	addVal := val.Mul(NewDecimalFromInt64(multiplier))
	return addDurationDT(dt, addVal, isNeg), nil
}

// addDurationYM safe-adds a month-based offset to the year-month duration accumulator.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//   - current: A pointer to the current accumulated month value (can be nil).
//   - val: The new month value to be added.
//   - isNegative: A boolean indicating whether the overall duration is negative.
//
// Returns:
//   - *int64: A pointer to the newly accumulated month value.
//   - error: An error if the addition operation causes a 64-bit integer overflow.
func addDurationYM(current *int64, val int64, isNegative bool) (*int64, error) {
	if isNegative {
		val = -val
	}
	var res int64
	if current != nil {
		res = *current
	}

	// Implementation Note: Overflow prevention during integer arithmetic
	// We check if adding val to the accumulated result would exceed math.MaxInt64 or math.MinInt64 to prevent undefined
	// behavior in signed integer addition.
	if (val > 0 && res > math.MaxInt64-val) || (val < 0 && res < math.MinInt64-val) {
		return nil, errParseOverflow
	}
	res += val
	return &res, nil
}

// addDurationDT safe-adds a second-based offset to the day-time duration accumulator.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6.1)
//
// Parameters:
//   - current: A pointer to the current accumulated seconds Decimal value (can be nil).
//   - val: The new seconds Decimal value to be added.
//   - isNegative: A boolean indicating whether the overall duration is negative.
//
// Returns:
//   - *Decimal: A pointer to the newly accumulated seconds Decimal value.
func addDurationDT(current *Decimal, val Decimal, isNegative bool) *Decimal {
	if isNegative {
		v := val.Neg()
		val = v
	}
	res := DefaultDecimal()
	if current != nil {
		res = *current
	}

	// Implementation Note: Decimal safe addition wrapper
	// In accordance with XSD 1.1 and RDF 1.2, arbitrary-precision decimal addition does not overflow.
	res = res.Add(val)
	return &res
}
