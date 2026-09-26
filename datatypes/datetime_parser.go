// Copyright 2025-2026 Trident Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datatypes

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// parseDateTime parses an XML Schema dateTime lexical string into a seven-property model.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Parameters:
//   - input: The raw string literal representing a dateTime value.
//
// Returns:
//   - DateTimeSevenPropertyModel: The parsed seven-property temporal model components.
//   - string: The remaining unparsed suffix of the input string.
//   - error: An error if the input violates the lexical constraints.
func parseDateTime(input string) (DateTimeSevenPropertyModel, string, error) {
	// Step 1: Parse the year component.
	// Extracts the year from the input string and validates its digit constraint.
	year, rem, err := parseYear(input)
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Step 2: Expect a hyphen separator.
	// Verifies that the year and month components are delimited by a hyphen.
	rem, err = expectChar(rem, '-', "year and month must be separated by '-'")
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Step 3: Parse the month component.
	// Extracts and validates the two-digit month value.
	month, rem, err := parseTwoDigit(rem, 1, monthsPerYear, "month")
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Step 4: Expect a hyphen separator.
	// Verifies that the month and day components are delimited by a hyphen.
	rem, err = expectChar(rem, '-', "month and day must be separated by '-'")
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Step 5: Parse the day component.
	// Extracts and validates the two-digit day value.
	day, rem, err := parseTwoDigit(rem, 1, maxDaysInMonth, "day")
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Step 6: Expect a 'T' date-time separator.
	// Verifies that the date and time segments are delimited by a capital 'T'.
	rem, err = expectChar(rem, 'T', "date and time must be separated by 'T'")
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Step 7: Parse the hour component.
	// Extracts and validates the two-digit hour value.
	hour, rem, err := parseTwoDigit(rem, 0, hoursPerDay, "hour")
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Step 8: Expect a colon separator.
	// Verifies that the hour and minute components are delimited by a colon.
	rem, err = expectChar(rem, ':', "hours and minutes must be separated by ':'")
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Step 9: Parse the minute component.
	// Extracts and validates the two-digit minute value.
	minute, rem, err := parseTwoDigit(rem, 0, maxMinutesPerHour, "minute")
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Step 10: Expect a colon separator.
	// Verifies that the minute and second components are delimited by a colon.
	rem, err = expectChar(rem, ':', "minutes and seconds must be separated by ':'")
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Step 11: Parse the second component.
	// Extracts and validates the high-precision decimal second value.
	second, rem, err := parseSecond(rem)
	if err != nil {
		return DateTimeSevenPropertyModel{}, input, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.7.1)
	// The value 24:00:00 is allowed to represent the end of the day, but any hour of 24 with a non-zero minute or
	// second is invalid.
	if hour == hoursPerDay && (minute != 0 || second.getValue().Sign() != 0) {
		return DateTimeSevenPropertyModel{}, input, newParseDateTimeError("times are not allowed to be after 24:00:00")
	}

	// Step 12: Parse timezone suffix.
	// Extracts the optional timezone offset component if present in the lexical representation.
	tz, rem, _ := parseTimezone(rem)

	// Step 13: Validate day of month.
	// Verifies that the day component does not exceed the maximum allowable days for the given month and year.
	if vErr := validateDayOfMonth(int64Ptr(year), month, day); vErr != nil {
		return DateTimeSevenPropertyModel{}, input, vErr
	}

	return DateTimeSevenPropertyModel{
		Year: &year, Month: &month, Day: &day,
		Hour: &hour, Minute: &minute, Second: &second, TimezoneOffset: tz,
	}, rem, nil
}

// parseYear extracts the year component from the beginning of the string.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.1)
//
// Parameters:
//   - s: The raw source string from which the year is extracted.
//
// Returns:
//   - int64: The parsed mathematical year value.
//   - string: The remaining suffix of the string after extraction.
//   - error: An error if the parsed representation violates the digit constraints.
func parseYear(s string) (int64, string, error) {
	sign := int64(1)
	rem := s
	if strings.HasPrefix(rem, "-") {
		sign = -1
		rem = rem[1:]
	}
	numStr, rem := integerPrefix(rem)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.7.1)
	// The year component must be represented by at least 4 digits.
	if len(numStr) < minYearDigits {
		return 0, s, newParseDateTimeError("year should be encoded on at least 4 digits")
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.7.1)
	// Leading zeros are prohibited for year representations exceeding 4 digits.
	if len(numStr) > minYearDigits && strings.HasPrefix(numStr, "0") {
		return 0, s, newParseDateTimeError("years must not start with 0 if more than 4 digits")
	}

	year, err := strconv.ParseInt(numStr, base10, 64)
	if err != nil {
		return 0, s, newParseDateTimeErrorWrap("invalid year", err)
	}
	return sign * year, rem, nil
}

// parseTwoDigit parses exactly two decimal digits from the start of the string and validates that the value falls
// within the specified closed range.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.1)
//
// Parameters:
//   - s: The source string starting with two digits.
//   - minVal: The inclusive lower limit of the valid range.
//   - maxVal: The inclusive upper limit of the valid range.
//   - name: The name of the field being parsed, used to construct error messages.
//
// Returns:
//   - uint8: The parsed unsigned 8-bit integer value.
//   - string: The remaining suffix of the string.
//   - error: An error if the format is invalid or bounds checks fail.
func parseTwoDigit(s string, minVal, maxVal int, name string) (uint8, string, error) {
	numStr, rem := integerPrefix(s)
	if len(numStr) != twoDigits {
		return 0, s, newParseDateTimeError(name + " must be encoded with two digits")
	}
	val, err := strconv.Atoi(numStr)

	// Implementation Note: Bounds validation and integer overflow prevention
	// Explicitly checks bounds against math.MaxUint8 and zero to eliminate potential G115 integer conversion overflows.
	if err != nil || val < minVal || val > maxVal || val < 0 || val > math.MaxUint8 {
		return 0, s, newParseDateTimeError(fmt.Sprintf("%s must be between %02d and %02d", name, minVal, maxVal))
	}
	return uint8(val), rem, nil
}

// parseSecond parses the seconds component, which may include a fractional decimal part as permitted by the
// specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Parameters:
//   - s: The raw string representation of the seconds.
//
// Returns:
//   - Decimal: The parsed high-precision decimal second value.
//   - string: The remaining suffix of the string.
//   - error: An error if the parsing or validation constraints are violated.
func parseSecond(s string) (Decimal, string, error) {
	numStr, rem := decimalPrefix(s)
	if len(numStr) < twoDigits || (strings.Contains(numStr, ".") && len(strings.Split(numStr, ".")[0]) != twoDigits) {
		return DefaultDecimal(), s, newParseDateTimeError("seconds integer part must be two digits")
	}
	if strings.HasSuffix(numStr, ".") {
		return DefaultDecimal(), s, newParseDateTimeError("seconds cannot end with a dot")
	}

	dec, _ := ParseDecimal(numStr)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.7)
	// Seconds can be arbitrary-precision decimals but must be strictly greater than or equal to 0, and strictly less
	// than 60.
	sixty := NewDecimalFromInt64(secondsPerMinute)
	if dec.getValue().Sign() < 0 || dec.getValue().Cmp(sixty.getValue()) >= 0 {
		return DefaultDecimal(), s, newParseDateTimeError("seconds must be between 00.0 and 60.0")
	}
	return dec, rem, nil
}

// parseTimezone parses the timezone offset suffix, returning a TimezoneOffset or an error.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.3)
//
// Parameters:
//   - s: The raw string containing the potential timezone suffix.
//
// Returns:
//   - *TimezoneOffset: A pointer to the parsed timezone offset instance, or nil if unzoned.
//   - string: The remaining suffix of the string after parsing.
//   - error: An error if the parsed offset deviates from the allowed boundaries.
func parseTimezone(s string) (*TimezoneOffset, string, error) {
	if s == "" {
		return nil, s, nil
	}
	if strings.HasPrefix(s, "Z") {
		return GetUTC(), s[1:], nil
	}

	sign := int16(1)
	rem := s
	switch {
	case strings.HasPrefix(rem, "+"):
		rem = rem[1:]
	case strings.HasPrefix(rem, "-"):
		sign = -1
		rem = rem[1:]
	default:
		return nil, s, nil
	}

	h, rem, err := parseTwoDigit(rem, 0, maxTimezoneHours, "timezone hour")
	if err != nil {
		return nil, s, err
	}
	rem, err = expectChar(rem, ':', "timezone hours and minutes must be separated by ':'")
	if err != nil {
		return nil, s, err
	}
	m, rem, err := parseTwoDigit(rem, 0, maxMinutesPerHour, "timezone minute")
	if err != nil {
		return nil, s, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.7.3)
	// The timezone offset must be between -14:00 and +14:00 inclusive.
	if h == maxTimezoneHours && m != 0 {
		return nil, s, newParseDateTimeError("timezone offset cannot exceed 14:00")
	}

	offset := sign * (int16(h)*minutesPerHour + int16(m))
	tz, _ := NewTimezoneOffset(offset)
	return &tz, rem, nil
}

// integerPrefix splits the input string into a prefix consisting entirely of ASCII digits and the remaining suffix of
// the string.
//
// Parameters:
//   - s: The raw source string.
//
// Returns:
//   - string: The prefix consisting of digit characters only.
//   - string: The remaining suffix of the string.
func integerPrefix(s string) (string, string) {
	for i, r := range s {
		if r < '0' || r > '9' {
			return s[:i], s[i:]
		}
	}
	return s, ""
}

// decimalPrefix splits the input string into a prefix representing a valid decimal number and the remaining suffix.
//
// Parameters:
//   - s: The raw source string.
//
// Returns:
//   - string: The extracted valid decimal prefix substring.
//   - string: The remaining suffix of the string.
func decimalPrefix(s string) (string, string) {
	dotSeen := false
	for i, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '.' && !dotSeen {
			dotSeen = true
			continue
		}
		return s[:i], s[i:]
	}
	return s, ""
}

// expectChar verifies that the input string starts with the expected rune.
//
// Parameters:
//   - s: The raw source string.
//   - char: The expected rune delimiter.
//   - msg: The error message to report on mismatch.
//
// Returns:
//   - string: The string with the parsed rune removed.
//   - error: An error if the starting rune does not match the expected character.
func expectChar(s string, char rune, msg string) (string, error) {
	if strings.HasPrefix(s, string(char)) {
		return s[1:], nil
	}
	return s, newParseDateTimeError(msg)
}

// validateDayOfMonth checks if the given day is valid for the specified month and year, accounting for leap years.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Parameters:
//   - year: A pointer to the year value, or nil if the year is absent.
//   - month: The month number to validate (1 to 12).
//   - day: The day of the month to validate (1 to 31).
//
// Returns:
//   - error: An error of type ParseDateTimeError if the day capacity is exceeded for the target month.
func validateDayOfMonth(year *int64, month, day uint8) error {
	if day > daysInMonth(year, int64(month)) {
		return newParseDateTimeError(fmt.Sprintf("%d is not a valid day of month %d", day, month))
	}
	return nil
}
