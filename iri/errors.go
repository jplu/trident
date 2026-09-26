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

package iri

import (
	"errors"
	"fmt"
)

// newParseError wraps an existing error into a standardized ParseError.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Parameters:
//   - err: The source error indicating a violation of the IRI generic syntax.
//
// Returns:
//   - *ParseError: The wrapped parse error structure, or nil if the input
//     error is nil.
func newParseError(err error) *ParseError {
	if err == nil {
		return nil
	}
	return &ParseError{Message: err.Error(), Err: errors.Unwrap(err)}
}

// kindError represents the parser syntax error as defined in the governing
// specification.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Representation:
// A structure carrying descriptive failure text, the specific rune where
// parsing failed, and additional contextual string details.
type kindError struct {
	message string
	char    rune
	details string
}

// Error returns the formatted error string.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Returns:
//   - string: The formatted string detailing the invalid parser state,
//     incorporating the character or details if available.
func (e *kindError) Error() string {
	msg := e.message
	if e.char != 0 {
		// Implementation Note: Character formatting.
		// Fall back to formatting the single invalid rune if present to
		// provide clear localized error feedback.
		msg = fmt.Sprintf("%s '%c'", msg, e.char)
	} else if e.details != "" {
		// Implementation Note: String formatting.
		// Use contextual details when no single character is responsible
		// for the validation failure.
		msg = fmt.Sprintf("%s '%s'", msg, e.details)
	}
	return msg
}

// BidiGuidelineError represents the bidirectional presentation guideline
// violation as defined in the governing specification.
//
// Specification Reference:
// RFC 3987 (Section 4.2)
//
// Representation:
// A structured error indicating a violation of the recommended
// bidirectional presentation rules (such as mixed directionality or invalid
// boundary characters).
type BidiGuidelineError struct {
	Rule      string
	Component string
	Message   string
}

// Error returns the formatted error string.
//
// Specification Reference:
// RFC 3987 (Section 4.2)
//
// Returns:
//   - string: The formatted error message detailing the bidirectional
//     presentation violation.
func (e *BidiGuidelineError) Error() string {
	return fmt.Sprintf(
		"bidirectional presentation violation (%s) in component '%s': %s",
		e.Rule,
		e.Component,
		e.Message,
	)
}
