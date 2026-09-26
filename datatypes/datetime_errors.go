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
	"errors"
	"fmt"
)

// ErrDateTimeOverflow represents an error that occurs when a date, time, or dateTime
// computation results in a value that is outside the supported value space.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2).
var ErrDateTimeOverflow = errors.New("overflow during xsd:dateTime computation")

// InvalidTimezoneError represents the timezone validation error as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.3)
//
// Representation:
// An error structure indicating that a timezone offset is outside the permissible limits of -14:00 to +14:00 inclusive.
type InvalidTimezoneError struct {
	// OffsetInMinutes holds the invalid timezone offset value in minutes.
	OffsetInMinutes int64
}

// Error returns the formatted string representation of the InvalidTimezoneError.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.3)
//
// Parameters:
//
// Returns:
//   - string: The formatted error message representing the timezone offset.
func (e InvalidTimezoneError) Error() string {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.7.3)
	// Timezone offsets must be formatted using the +/-hh:mm pattern.
	return fmt.Sprintf(
		"invalid timezone offset %02d:%02d",
		e.OffsetInMinutes/minutesPerHour,
		absInt64(e.OffsetInMinutes)%minutesPerHour,
	)
}

// ParseDateTimeError represents the lexical date-time parsing error as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.1)
//
// Representation:
// An error structure holding details of a lexical parsing failure when an input string deviates from the mandated date,
// time, or dateTime formats.
type ParseDateTimeError struct {
	// msg contains the specific message describing why parsing failed.
	msg string
	// err optionally holds an underlying wrapped error that caused the parsing failure.
	err error
}

// Error returns the formatted error message for the ParseDateTimeError.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.1)
//
// Parameters:
//
// Returns:
//   - string: The formatted error message with optional wrapped error details.
func (e ParseDateTimeError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("parsing date/time: %s: %v", e.msg, e.err)
	}
	return "parsing date/time: " + e.msg
}

// Unwrap returns the underlying wrapped error, facilitating Go's error unwrapping patterns.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.1)
//
// Parameters:
//
// Returns:
//   - error: The underlying wrapped error instance, or nil if none.
func (e ParseDateTimeError) Unwrap() error {
	return e.err
}

// newParseDateTimeError instantiates a ParseDateTimeError with a simple error message.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.1)
//
// Parameters:
//   - msg: The error message detailing the parsing failure.
//
// Returns:
//   - ParseDateTimeError: The newly constructed error instance.
func newParseDateTimeError(msg string) ParseDateTimeError {
	return ParseDateTimeError{msg: msg}
}

// newParseDateTimeErrorWrap instantiates a ParseDateTimeError wrapping an underlying error.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.1)
//
// Parameters:
//   - msg: The error message detailing the parsing failure.
//   - err: The underlying cause of the parsing failure.
//
// Returns:
//   - ParseDateTimeError: The newly constructed error instance wrapping the cause.
func newParseDateTimeErrorWrap(msg string, err error) ParseDateTimeError {
	return ParseDateTimeError{msg: msg, err: err}
}

// absInt64 returns the absolute value of a 64-bit integer.
//
// Parameters:
//   - n: The 64-bit signed integer.
//
// Returns:
//   - int64: The absolute value of the input.
func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
