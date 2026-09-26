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

// ErrDurationOverflow represents an overflow during duration computations.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6).
var ErrDurationOverflow = errors.New("overflow during xsd:duration computation")

// ErrOppositeSignInDurationComponents indicates that the year-month and day-time
// components of an xsd:duration have opposite signs.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6).
var ErrOppositeSignInDurationComponents = errors.New(
	"the xsd:yearMonthDuration and xsd:dayTimeDuration components of a xsd:duration can't have opposite sign",
)

// ParseDurationError represents the duration parsing error as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Representation:
// An error structure wrapping a specific lexical failure detail message and an underlying error cause.
type ParseDurationError struct {
	// msg holds the specific lexical failure detail message.
	msg string
	// err wraps an underlying error cause, such as a numerical overflow.
	err error
}

// Error returns a formatted string representation of the ParseDurationError.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Returns:
//   - string: The formatted error message string.
func (e ParseDurationError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("parsing duration: %s: %v", e.msg, e.err)
	}
	return "parsing duration: " + e.msg
}

// Unwrap returns the underlying wrapped error.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Returns:
//   - error: The underlying wrapped error.
func (e ParseDurationError) Unwrap() error {
	return e.err
}

// errParseOverflow is a pre-allocated unexported error indicating a capacity overflow
// during the processing of duration component numbers.
var errParseOverflow = ParseDurationError{msg: "overflow error", err: ErrDurationOverflow}

// newParseDurationError instantiates a ParseDurationError with the given message.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.6)
//
// Parameters:
//   - msg: The failure detail message.
//
// Returns:
//   - ParseDurationError: The newly constructed error instance.
func newParseDurationError(msg string) ParseDurationError {
	return ParseDurationError{msg: msg}
}
