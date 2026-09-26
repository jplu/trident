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
	"strings"

	// TODO: At some point implement my own NFC module.
	"golang.org/x/text/unicode/norm"
)

// ParseError represents the parsing failure details as defined in the governing specification.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Representation:
// An error wrapper holding the descriptive failure message and the wrapped internal error details.
type ParseError struct {
	Message string
	Err     error
}

// Error returns the string representation of the parse error.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Returns:
//   - string: The formatted error message string.
func (e *ParseError) Error() string {
	return fmt.Sprintf("IRI parse error: %s", e.Message)
}

// Unwrap provides compatibility with Go's standard errors package by returning the underlying error.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Returns:
//   - error: The wrapped internal error, or nil if none.
func (e *ParseError) Unwrap() error {
	return e.Err
}

// ErrIriRelativize is returned by the Relativize method when it's not possible
// to create a relative reference because the target IRI's path contains dot segments
// ("." or ".."). Such paths must be normalized before relativization.
var ErrIriRelativize = errors.New(
	"it is not possible to make this IRI relative because it contains '/..' or '/.'",
)

// Ref represents the IRI reference as defined in the governing specification.
//
// Specification Reference:
// RFC 3987 (Section 1.3)
//
// Representation:
// A Unicode-based string reference representing either an absolute IRI or a relative IRI reference.
// It provides accessors for parsing and identifying individual IRI components.
type Ref struct {
	iri       string
	positions Positions
}

// ParseRef parses and validates a string as an IRI reference.
//
// Specification Reference:
// RFC 3987 (Section 3.1)
//
// Parameters:
//   - s: The raw string representation of the IRI reference to be parsed and validated.
//
// Returns:
//   - *Ref: A pointer to the parsed Ref struct.
//   - error: A ParseError if the string violates syntactic rules.
func ParseRef(s string) (*Ref, error) {
	var target strings.Builder
	target.Grow(len(s))
	output := &stringOutputBuffer{builder: &target}

	pos, err := run(s, nil, false, output)
	if err != nil {
		return nil, newParseError(err)
	}

	return &Ref{iri: target.String(), positions: pos}, nil
}

// ParseNormalizedRef parses and validates a string as an IRI reference after applying NFC normalization.
//
// Specification Reference:
// RFC 3987 (Section 3.1)
//
// Parameters:
//   - s: The raw string representation to be normalized and parsed.
//
// Returns:
//   - *Ref: A pointer to the normalized and parsed Ref struct.
//   - error: A ParseError if the string violates syntactic rules.
func ParseNormalizedRef(s string) (*Ref, error) {
	normalizedIRI := norm.NFC.String(s)
	pos, err := run(normalizedIRI, nil, false, &voidOutputBuffer{})
	if err != nil {
		return nil, newParseError(err)
	}
	return &Ref{iri: normalizedIRI, positions: pos}, nil
}

// Resolve resolves a relative IRI reference against the current Ref (acting as the base IRI).
//
// Specification Reference:
// RFC 3987 (Section 6.5)
//
// Parameters:
//   - relativeIRI: The relative IRI reference string to be resolved against the base.
//
// Returns:
//   - *Ref: A pointer to the newly resolved absolute Ref struct.
//   - error: An error if the resolution or validation fails.
func (r *Ref) Resolve(relativeIRI string) (*Ref, error) {
	builder := &strings.Builder{}
	builder.Grow(len(r.iri) + len(relativeIRI))
	pos, err := r.ResolveTo(relativeIRI, builder)
	if err != nil {
		return nil, err
	}
	return &Ref{iri: builder.String(), positions: pos}, nil
}

// ResolveTo resolves a relative IRI reference and writes the result directly into the target strings.Builder.
//
// Specification Reference:
// RFC 3987 (Section 6.5)
//
// Parameters:
//   - relativeIRI: The relative IRI reference string to be resolved against the base.
//   - target: The strings.Builder buffer where the resolved IRI is written.
//
// Returns:
//   - Positions: The byte offsets of the resolved components in the target builder.
//   - error: An error if the relative reference parsing fails.
func (r *Ref) ResolveTo(relativeIRI string, target *strings.Builder) (Positions, error) {
	// Implementation Note: Sanitization and validation of the relative reference before resolving.
	// Converting the input to NFC and parsing it ensures that invalid characters are caught and
	// lenient characters are properly handled prior to running the resolution process.
	normalizedRelativeIRI := norm.NFC.String(relativeIRI)
	parsedRef, err := ParseRef(normalizedRelativeIRI)
	if err != nil {
		return Positions{}, err
	}

	b := &base{IRI: r.iri, Pos: r.positions}
	output := &stringOutputBuffer{builder: target}

	// Implementation Note: Execute reference resolution
	// Resolve the pre-parsed, sanitized, and percent-encoded reference against the base components.
	pos, _ := run(parsedRef.iri, b, false, output)
	return pos, nil
}

// Iri represents the absolute IRI as defined in the governing specification.
//
// Specification Reference:
// RFC 3987 (Section 1.3)
//
// Representation:
// A guaranteed absolute Internationalized Resource Identifier containing a scheme and embedding a Ref.
type Iri struct {
	Ref
}

// ParseIri parses and validates a string, ensuring it is an absolute IRI.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Parameters:
//   - s: The raw string representation to be parsed as an absolute IRI.
//
// Returns:
//   - *Iri: A pointer to the absolute Iri struct.
//   - error: An error if the string is relative, or if it violates syntactic rules.
func ParseIri(s string) (*Iri, error) {
	ref, err := ParseRef(s)
	if err != nil {
		return nil, err
	}
	return NewIriFromRef(ref)
}

// ParseNormalizedIri parses a string as an absolute IRI, first applying NFC normalization.
//
// Specification Reference:
// RFC 3987 (Section 3.1)
//
// Parameters:
//   - s: The raw string representation to be normalized and parsed as an absolute IRI.
//
// Returns:
//   - *Iri: A pointer to the normalized absolute Iri struct.
//   - error: An error if the string is relative, or if it violates syntactic rules.
func ParseNormalizedIri(s string) (*Iri, error) {
	ref, err := ParseNormalizedRef(s)
	if err != nil {
		return nil, err
	}
	return NewIriFromRef(ref)
}

// NewIriFromRef attempts to create an absolute Iri from an existing Ref.
//
// Specification Reference:
// RFC 3987 (Section 1.3)
//
// Parameters:
//   - ref: A pointer to the Ref struct.
//
// Returns:
//   - *Iri: A pointer to the absolute Iri struct.
//   - error: An error if the provided Ref is not absolute.
func NewIriFromRef(ref *Ref) (*Iri, error) {
	if !ref.IsAbsolute() {
		return nil, newParseError(errNoScheme)
	}
	return &Iri{Ref: *ref}, nil
}

// Resolve resolves a relative IRI reference against the current Iri and returns a new absolute Iri.
//
// Specification Reference:
// RFC 3987 (Section 6.5)
//
// Parameters:
//   - relativeIRI: The relative IRI reference string to be resolved against the absolute base.
//
// Returns:
//   - *Iri: A pointer to the resolved absolute Iri.
//   - error: An error if the resolution or validation fails.
func (i *Iri) Resolve(relativeIRI string) (*Iri, error) {
	ref, err := i.Ref.Resolve(relativeIRI)
	if err != nil {
		return nil, err
	}
	// Spec Rule: RFC 3987 (Section 6.5)
	// The result of a reference resolution against an absolute base IRI is always absolute.
	return &Iri{Ref: *ref}, nil
}

// ResolveTo resolves a relative IRI and writes the resulting absolute IRI to the provided strings.Builder.
//
// Specification Reference:
// RFC 3987 (Section 6.5)
//
// Parameters:
//   - relativeIRI: The relative IRI reference string to be resolved.
//   - target: The strings.Builder buffer where the resolved IRI is written.
//
// Returns:
//   - error: An error if the resolution or validation fails.
func (i *Iri) ResolveTo(relativeIRI string, target *strings.Builder) error {
	_, err := i.Ref.ResolveTo(relativeIRI, target)
	return err
}
