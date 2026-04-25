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

// ParseError is the error type returned by parsing functions in this package.
// It contains a descriptive message and may wrap a more specific internal error.
type ParseError struct {
	Message string
	Err     error
}

// Error returns the string representation of the parse error.
func (e *ParseError) Error() string {
	return fmt.Sprintf("IRI parse error: %s", e.Message)
}

// Unwrap provides compatibility with Go's standard errors package.
func (e *ParseError) Unwrap() error {
	return e.Err
}

// ErrIriRelativize is returned by the Relativize method when it's not possible
// to create a relative reference because the target IRI's path contains dot segments
// ("." or ".."). Such paths must be normalized before relativization.
var ErrIriRelativize = errors.New("it is not possible to make this IRI relative because it contains '/..' or '/.'")

// Ref represents an IRI reference, which can be either absolute or relative.
// It is an immutable type; methods that modify the IRI, like Resolve, return a new Ref.
// The internal `iri` string is stored exactly as provided to the parsing function.
// For comparison purposes where canonical equivalence is desired, use ParseNormalizedRef
// or the Normalize() method.
type Ref struct {
	iri       string
	positions Positions
}

// ParseRef parses and validates a string as an IRI reference.
// This function is compliant with RFC 3987, Section 3.1, Step 1c.
// It parses the string as-is, without applying any Unicode normalization.
// This preserves the exact character sequence of the input, which is critical for
// applications that use IRIs as unique, opaque identifiers.
//
// For applications that require canonical equivalence for comparison or storage,
// use ParseNormalizedRef instead.
func ParseRef(s string) (*Ref, error) {
	pos, err := run(s, nil, false, &voidOutputBuffer{})
	if err != nil {
		return nil, newParseError(err)
	}
	return &Ref{iri: s, positions: pos}, nil
}

// ParseNormalizedRef provides the previous behavior of ParseRef for users
// who need it. It first normalizes the input string to Unicode Normalization Form C (NFC)
// and then parses it. This is useful for ensuring that canonically equivalent IRIs
// are treated as identical, which is important for caching, history, and other
// comparison-sensitive operations.
//
// In accordance with RFC 3987 sections 3.1 and 5.3.2.2, this function should
// be used when the source of the IRI string is not from a pre-normalized Unicode
// source (e.g., read from paper or converted from a legacy encoding).
func ParseNormalizedRef(s string) (*Ref, error) {
	normalizedIRI := norm.NFC.String(s)
	pos, err := run(normalizedIRI, nil, false, &voidOutputBuffer{})
	if err != nil {
		return nil, newParseError(err)
	}
	return &Ref{iri: normalizedIRI, positions: pos}, nil
}

// Resolve resolves a relative IRI reference against the current Ref (which acts as the base IRI).
// It returns a new, absolute Ref. This operation is equivalent to resolving a hyperlink.
func (r *Ref) Resolve(relativeIRI string) (*Ref, error) {
	builder := &strings.Builder{}
	builder.Grow(len(r.iri) + len(relativeIRI))
	pos, err := r.ResolveTo(relativeIRI, builder)
	if err != nil {
		return nil, err
	}
	return &Ref{iri: builder.String(), positions: pos}, nil
}

// ResolveTo resolves a relative IRI reference and writes the result directly into
// the provided strings.Builder, avoiding extra allocations. It returns the positions
// of the components in the resulting IRI. This is useful for performance-critical code.
// The relative IRI reference is normalized to NFC before resolution.
func (r *Ref) ResolveTo(relativeIRI string, target *strings.Builder) (Positions, error) {
	// Note: Normalizing the relative part here is a good practice for consistency
	// of the resolved output, even if the base might not be normalized.
	normalizedRelativeIRI := norm.NFC.String(relativeIRI)

	b := &base{IRI: r.iri, Pos: r.positions}
	output := &stringOutputBuffer{builder: target}

	pos, err := run(normalizedRelativeIRI, b, false, output)
	if err != nil {
		return Positions{}, newParseError(err)
	}
	return pos, nil
}

// Iri represents a guaranteed absolute IRI. It embeds a Ref and provides convenience
// methods for working with IRIs that must be absolute.
type Iri struct {
	Ref
}

// ParseIri parses and validates a string, ensuring it is an absolute IRI.
// If the string is a relative reference, it returns an error. The string is not
// NFC normalized; for that, use ParseNormalizedIri.
func ParseIri(s string) (*Iri, error) {
	ref, err := ParseRef(s)
	if err != nil {
		return nil, err
	}
	return NewIriFromRef(ref)
}

// ParseNormalizedIri parses a string as an absolute IRI, first applying NFC normalization.
func ParseNormalizedIri(s string) (*Iri, error) {
	ref, err := ParseNormalizedRef(s)
	if err != nil {
		return nil, err
	}
	return NewIriFromRef(ref)
}

// NewIriFromRef attempts to create an absolute Iri from an existing Ref.
// It returns an error if the provided Ref is not absolute.
func NewIriFromRef(ref *Ref) (*Iri, error) {
	if !ref.IsAbsolute() {
		return nil, newParseError(errNoScheme)
	}
	return &Iri{Ref: *ref}, nil
}

// Resolve resolves a relative IRI reference against the current Iri and returns
// a new, absolute Iri.
func (i *Iri) Resolve(relativeIRI string) (*Iri, error) {
	ref, err := i.Ref.Resolve(relativeIRI)
	if err != nil {
		return nil, err
	}
	// The result of a resolution is always absolute.
	return &Iri{Ref: *ref}, nil
}

// ResolveTo resolves a relative IRI and writes the resulting absolute IRI
// to the provided strings.Builder, avoiding allocations.
func (i *Iri) ResolveTo(relativeIRI string, target *strings.Builder) error {
	_, err := i.Ref.ResolveTo(relativeIRI, target)
	return err
}
