/*
Copyright 2025 Trident Authors

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

var (
	// errNoScheme is returned when an absolute IRI is expected but no scheme
	// (e.g., "http:") is found. This typically occurs when the IRI string
	// starts with a colon, which is invalid.
	errNoScheme = &kindError{message: "No scheme found in an absolute IRI"}
	// errPathStartingWithSlashes is returned when an IRI has a path that
	// starts with "//" but does not have an authority component. This is
	// disallowed by RFC 3987 to avoid ambiguity with network-path references.
	// For example, "scheme:////path" is valid, but "scheme:/path" where the path
	// starts with `//` is not.
	errPathStartingWithSlashes = &kindError{
		message: "An IRI path is not allowed to start with // if there is no authority",
	}
)

// newParseError creates a new ParseError, wrapping the original error.
// It returns nil if the input error is nil.
func newParseError(err error) *ParseError {
	if err == nil {
		return nil
	}
	return &ParseError{Message: err.Error(), Err: errors.Unwrap(err)}
}

// kindError is a specialized error type used by the parser to provide
// detailed context about a parsing failure.
type kindError struct {
	message string
	char    rune
	details string
}

// Error formats the error message with any available character, details, or
// wrapped error information.
func (e *kindError) Error() string {
	msg := e.message
	if e.char != 0 {
		msg = fmt.Sprintf("%s '%c'", msg, e.char)
	} else if e.details != "" {
		msg = fmt.Sprintf("%s '%s'", msg, e.details)
	}
	return msg
}
