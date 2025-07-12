/*
Copyright 2025 Trident Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUTHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package iri provides types and functions for working with Internationalized
// Resource Identifiers (IRIs) and IRI references as defined by RFC 3987.
//
// The package offers two main types:
//   - Ref: Represents an IRI reference, which can be either absolute (e.g., "http://example.com/a")
//     or relative (e.g., "/a", "b", "#c").
//   - Iri: Represents a guaranteed absolute IRI, which always includes a scheme.
//
// Key features include:
//   - Strict parsing and validation against RFC 3987.
//   - High-performance "unchecked" parsing for known-valid inputs.
//   - Reference resolution (`Resolve`) to compute an absolute IRI from a base and a relative reference.
//   - Relativization (`Relativize`) to compute a relative reference between two absolute IRIs.
//   - Zero-allocation resolution variants (`ResolveTo`) for performance-critical applications.
//   - Support for JSON marshalling and unmarshalling.
package iri

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jplu/trident/internal/parser"
)

// ParseError is the error type returned by parsing functions in this package.
// It contains a descriptive message and may wrap a more specific internal error.
type ParseError struct {
	Message string
	Err     error
}

// newParseError creates a new ParseError, wrapping the original error.
func newParseError(err error) *ParseError {
	return &ParseError{Message: err.Error(), Err: errors.Unwrap(err)}
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
type Ref struct {
	iri       string
	positions parser.Positions
}

// ParseRef parses and validates a string as an IRI reference.
// It returns a new Ref on success or a ParseError on failure.
func ParseRef(s string) (*Ref, error) {
	// Use the internal parser in validation-only mode to get component positions.
	positions, err := parser.Run(s, nil, false, &parser.VoidOutputBuffer{})
	if err != nil {
		return nil, newParseError(err)
	}
	return &Ref{iri: s, positions: positions}, nil
}

// ParseRefUnchecked parses a string as an IRI reference without performing
// strict validation. It is significantly faster than ParseRef but should only
// be used with input that is *known* to be a valid IRI reference.
//
// Providing an invalid IRI may cause a panic.
func ParseRefUnchecked(s string) *Ref {
	positions, err := parser.Run(s, nil, true, &parser.VoidOutputBuffer{})
	if err != nil {
		// This should only happen if the internal parser has a bug,
		// as unchecked mode is not expected to return errors.
		panic(fmt.Sprintf("ParseRefUnchecked failed on known-valid IRI: %v", err))
	}
	return &Ref{iri: s, positions: positions}
}

// Resolve resolves a relative IRI reference against the current Ref (which acts as the base IRI).
// It returns a new, absolute Ref. This operation is equivalent to resolving a hyperlink.
func (r *Ref) Resolve(relativeIRI string) (*Ref, error) {
	builder := &strings.Builder{}
	builder.Grow(len(r.iri) + len(relativeIRI)) // Pre-allocate for efficiency.
	positions, err := r.ResolveTo(relativeIRI, builder)
	if err != nil {
		return nil, err
	}
	return &Ref{iri: builder.String(), positions: positions}, nil
}

// ResolveUnchecked resolves a relative IRI reference without validation. It is faster
// than Resolve but should only be used when both the base and the relative IRI are
// known to be valid.
//
// Providing invalid input may cause a panic.
func (r *Ref) ResolveUnchecked(relativeIRI string) *Ref {
	builder := &strings.Builder{}
	builder.Grow(len(r.iri) + len(relativeIRI))
	positions := r.ResolveUncheckedTo(relativeIRI, builder)
	return &Ref{iri: builder.String(), positions: positions}
}

// ResolveTo resolves a relative IRI reference and writes the result directly into
// the provided strings.Builder, avoiding extra allocations. This is useful for
// performance-critical code.
func (r *Ref) ResolveTo(relativeIRI string, target *strings.Builder) (parser.Positions, error) {
	base := &parser.Base{IRI: r.iri, Pos: r.positions}
	output := &parser.StringOutputBuffer{Builder: target}
	positions, err := parser.Run(relativeIRI, base, false, output)
	if err != nil {
		return parser.Positions{}, newParseError(err)
	}
	return positions, nil
}

// ResolveUncheckedTo resolves a relative IRI reference without validation and writes
// the result to the provided strings.Builder.
//
// Providing invalid input may cause a panic.
func (r *Ref) ResolveUncheckedTo(relativeIRI string, target *strings.Builder) parser.Positions {
	base := &parser.Base{IRI: r.iri, Pos: r.positions}
	output := &parser.StringOutputBuffer{Builder: target}
	positions, err := parser.Run(relativeIRI, base, true, output)
	if err != nil {
		panic(fmt.Sprintf("ResolveUncheckedTo failed on known-valid IRI: %v", err))
	}
	return positions
}

// String returns the underlying string representation of the IRI reference.
func (r *Ref) String() string {
	return r.iri
}

// IsAbsolute returns true if the IRI reference is absolute (i.e., it has a scheme).
func (r *Ref) IsAbsolute() bool {
	return r.positions.SchemeEnd != 0
}

// Scheme returns the scheme component of the IRI (e.g., "http") and a boolean
// indicating whether it was present.
func (r *Ref) Scheme() (string, bool) {
	if !r.IsAbsolute() {
		return "", false
	}
	// The scheme ends one character before the colon.
	return r.iri[:r.positions.SchemeEnd-1], true
}

// Authority returns the authority component of the IRI (e.g., "example.com:80")
// and a boolean indicating whether it was present. The leading "//" is not included.
func (r *Ref) Authority() (string, bool) {
	if r.positions.AuthorityEnd <= r.positions.SchemeEnd {
		return "", false
	}

	authorityComponent := r.iri[r.positions.SchemeEnd:r.positions.AuthorityEnd]
	return strings.TrimPrefix(authorityComponent, "//"), true
}

// Path returns the path component of the IRI. A path is always present,
// though it may be an empty string.
func (r *Ref) Path() string {
	return r.iri[r.positions.AuthorityEnd:r.positions.PathEnd]
}

// Query returns the query component of the IRI (the part after "?", without the "?")
// and a boolean indicating whether it was present.
func (r *Ref) Query() (string, bool) {
	if r.positions.PathEnd >= r.positions.QueryEnd {
		return "", false
	}
	// The query starts one character after the '?'.
	return r.iri[r.positions.PathEnd+1 : r.positions.QueryEnd], true
}

// Fragment returns the fragment component of the IRI (the part after "#", without the "#")
// and a boolean indicating whether it was present.
func (r *Ref) Fragment() (string, bool) {
	if r.positions.QueryEnd >= len(r.iri) {
		return "", false
	}
	// The fragment starts one character after the '#'.
	return r.iri[r.positions.QueryEnd+1:], true
}

// MarshalJSON implements the json.Marshaler interface, encoding the Ref as a JSON string.
func (r *Ref) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.iri)
}

// UnmarshalJSON implements the json.Unmarshaler interface. It decodes a JSON string
// into a Ref, performing validation in the process.
func (r *Ref) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	newRef, err := ParseRef(s)
	if err != nil {
		return err
	}
	*r = *newRef
	return nil
}

// Iri represents a guaranteed absolute IRI. It embeds a Ref and provides convenience
// methods for working with IRIs that must be absolute.
type Iri struct {
	Ref
}

// ParseIri parses and validates a string, ensuring it is an absolute IRI.
// If the string is a relative reference, it returns an error.
func ParseIri(s string) (*Iri, error) {
	ref, err := ParseRef(s)
	if err != nil {
		return nil, err
	}
	return NewIriFromRef(ref)
}

// ParseIriUnchecked parses a string as an absolute IRI without strict validation.
// It is faster than ParseIri but will panic if the input is not a valid,
// absolute IRI. It should only be used with known-good input.
func ParseIriUnchecked(s string) *Iri {
	ref := ParseRefUnchecked(s)
	if !ref.IsAbsolute() {
		panic("ParseIriUnchecked called with a relative IRI")
	}
	return &Iri{Ref: *ref}
}

// NewIriFromRef attempts to create an absolute Iri from an existing Ref.
// It returns an error if the provided Ref is not absolute.
func NewIriFromRef(ref *Ref) (*Iri, error) {
	if !ref.IsAbsolute() {
		return nil, newParseError(parser.ErrNoScheme)
	}
	return &Iri{Ref: *ref}, nil
}

// Scheme returns the scheme component of the IRI. It is guaranteed to be present.
func (i *Iri) Scheme() string {
	s, _ := i.Ref.Scheme()
	return s
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

// ResolveUnchecked resolves a relative IRI reference without validation.
// It is faster but may panic on invalid input.
func (i *Iri) ResolveUnchecked(relativeIRI string) *Iri {
	ref := i.Ref.ResolveUnchecked(relativeIRI)
	return &Iri{Ref: *ref}
}

// ResolveTo resolves a relative IRI and writes the resulting absolute IRI
// to the provided strings.Builder, avoiding allocations.
func (i *Iri) ResolveTo(relativeIRI string, target *strings.Builder) error {
	_, err := i.Ref.ResolveTo(relativeIRI, target)
	return err
}

// ResolveUncheckedTo resolves a relative IRI without validation and writes the
// result to the provided strings.Builder. It may panic on invalid input.
func (i *Iri) ResolveUncheckedTo(relativeIRI string, target *strings.Builder) {
	i.Ref.ResolveUncheckedTo(relativeIRI, target)
}

// MarshalJSON implements the json.Marshaler interface.
func (i *Iri) MarshalJSON() ([]byte, error) {
	return i.Ref.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface, ensuring the
// decoded IRI is absolute.
func (i *Iri) UnmarshalJSON(data []byte) error {
	var ref Ref
	if err := ref.UnmarshalJSON(data); err != nil {
		return err
	}
	newIri, err := NewIriFromRef(&ref)
	if err != nil {
		return err
	}
	*i = *newIri
	return nil
}

// Relativize computes a relative IRI reference that, when resolved against the
// base IRI `i`, will result in the target IRI `abs`. This is the inverse of the
// Resolve operation.
//
// The method will return the full target IRI or a scheme-relative IRI if the
// schemes or authorities differ. It returns `ErrIriRelativize` if the target
// IRI's path contains dot-segments ("." or "..").
func (i *Iri) Relativize(abs *Iri) (*Ref, error) {
	base := i
	absPath := abs.Path()

	// Pre-condition: The target path cannot contain dot segments for relativization.
	for _, segment := range strings.Split(absPath, "/") {
		if segment == "." || segment == ".." {
			return nil, ErrIriRelativize
		}
	}

	// If schemes are different, the result is just the absolute target IRI.
	if base.Scheme() != abs.Scheme() {
		return ParseRef(abs.String())
	}

	baseAuthority, hasBaseAuthority := base.Authority()
	absAuthority, hasAbsAuthority := abs.Authority()

	// If authorities differ, return a scheme-relative reference (e.g., "//example.org/path").
	if hasBaseAuthority != hasAbsAuthority || (hasBaseAuthority && baseAuthority != absAuthority) {
		if !hasAbsAuthority {
			// This is a rare case, like relativizing http://a against http:b
			return ParseRef(abs.String())
		}
		return ParseRef(abs.String()[abs.positions.SchemeEnd:])
	}

	basePath := base.Path()

	if absPath == "" && basePath != "" {
		if !hasAbsAuthority {
			return ParseRef(abs.String())
		}
		return ParseRef(abs.String()[abs.positions.SchemeEnd:])
	}

	// If paths are identical, the difference is only in the query and/or fragment.
	if basePath == absPath {
		return i.relativizeForSamePath(abs)
	}

	if !hasBaseAuthority {
		return i.relativizeForNoAuthority(abs)
	}

	return i.relativizeWithAuthority(abs)
}

// relativizeForSamePath handles relativization when base and target paths are identical.
func (i *Iri) relativizeForSamePath(abs *Iri) (*Ref, error) {
	base := i
	baseQuery, hasBaseQuery := base.Query()
	absQuery, hasAbsQuery := abs.Query()
	absFragment, hasAbsFragment := abs.Fragment()

	// If queries also match, the result is just the fragment or an empty string.
	if hasBaseQuery == hasAbsQuery && baseQuery == absQuery {
		if hasAbsFragment {
			return ParseRef("#" + absFragment)
		}
		return ParseRef("") // The IRIs are identical.
	}

	if !hasAbsQuery && hasBaseQuery {
		return i.relativizeForSamePathWithEmptyTargetQuery(abs)
	}

	// Otherwise, the result starts from the query part.
	return ParseRef(abs.String()[abs.positions.PathEnd:])
}

// relativizeForSamePathWithEmptyTargetQuery handles a specific edge case where
// paths match, but the target has no query while the base does.
func (i *Iri) relativizeForSamePathWithEmptyTargetQuery(abs *Iri) (*Ref, error) {
	absPath := abs.Path()
	if absPath != "" {
		// The result should be the filename part of the path.
		lastSlash := strings.LastIndex(absPath, "/")
		relPath := absPath[lastSlash+1:]
		if relPath == "" {
			relPath = "." // e.g. http://a/b/ resolves to "."
		}
		return buildRelativeRef(relPath, abs)
	}

	// If path is empty, we must include the authority.
	_, hasAbsAuthority := abs.Authority()
	if !hasAbsAuthority {
		return ParseRef(abs.String())
	}
	return ParseRef(abs.String()[abs.positions.SchemeEnd:])
}

// relativizeForNoAuthority handles relativization when both IRIs lack an authority part.
func (i *Iri) relativizeForNoAuthority(abs *Iri) (*Ref, error) {
	base := i
	basePath := base.Path()
	absPath := abs.Path()

	// A simple prefix cut might work if it doesn't create ambiguity.
	if relPath, ok := strings.CutPrefix(abs.String(), base.String()); ok {
		firstColon := strings.Index(relPath, ":")
		firstSlash := strings.Index(relPath, "/")
		isSchemeAmbiguous := firstColon != -1 && (firstSlash == -1 || firstColon < firstSlash)
		isPathAbsoluteAmbiguous := strings.HasPrefix(relPath, "/")

		if !isSchemeAmbiguous && !isPathAbsoluteAmbiguous {
			return ParseRef(relPath)
		}
	}

	// Find common parent directory.
	lastSlash := strings.LastIndex(basePath, "/")
	if lastSlash == -1 {
		return ParseRef(abs.String())
	}

	baseDir := basePath[:lastSlash+1]
	relPath, ok := strings.CutPrefix(absPath, baseDir)
	if !ok || strings.HasPrefix(relPath, "/") {
		return ParseRef(abs.String())
	}

	firstColon := strings.Index(relPath, ":")
	firstSlash := strings.Index(relPath, "/")
	isSchemeAmbiguous := firstColon != -1 && (firstSlash == -1 || firstColon < firstSlash)
	if isSchemeAmbiguous {
		return ParseRef(abs.String())
	}

	return buildRelativeRef(relPath, abs)
}

// relativizeWithAuthority handles the most complex case where both IRIs have
// an authority, and paths need to be compared.
func (i *Iri) relativizeWithAuthority(abs *Iri) (*Ref, error) {
	base := i
	basePath := base.Path()
	absPath := abs.Path()

	if strings.HasPrefix(absPath, "//") && !strings.HasPrefix(basePath, "//") {
		return ParseRef(abs.String()[abs.positions.SchemeEnd:])
	}

	lastSlash := strings.LastIndex(basePath, "/")
	if lastSlash == -1 {
		return ParseRef(abs.String()[abs.positions.AuthorityEnd:])
	}
	baseDir := basePath[:lastSlash+1]

	// If target is in the same directory, the result is simple.
	if relPath, ok := strings.CutPrefix(absPath, baseDir); ok && (len(relPath) == 0 || relPath[0] != ':') {
		if relPath == "" {
			relPath = "."
		}
		return buildRelativeRef(relPath, abs)
	}

	// Find the common path prefix by splitting into segments.
	baseSegs := strings.Split(baseDir, "/")
	absSegs := strings.Split(absPath, "/")

	commonSegs := 0
	for commonSegs < len(baseSegs) && commonSegs < len(absSegs) && baseSegs[commonSegs] == absSegs[commonSegs] {
		commonSegs++
	}

	// If no common path, return path from root.
	if commonSegs <= 1 {
		return ParseRef(abs.String()[abs.positions.AuthorityEnd:])
	}

	var b strings.Builder
	// Add "../" for each directory to go up from base to the common ancestor.
	for i := commonSegs; i < len(baseSegs); i++ {
		if baseSegs[i] != "" {
			b.WriteString("../")
		}
	}
	// Add the rest of the target path from the common ancestor.
	b.WriteString(strings.Join(absSegs[commonSegs:], "/"))

	return buildRelativeRef(b.String(), abs)
}

// buildRelativeRef constructs the final relative reference string from a relative path
// and the query/fragment parts of the absolute target IRI.
func buildRelativeRef(relPath string, abs *Iri) (*Ref, error) {
	absQuery, hasAbsQuery := abs.Query()
	absFragment, hasAbsFragment := abs.Fragment()

	var b strings.Builder
	b.WriteString(relPath)
	if hasAbsQuery {
		b.WriteRune('?')
		b.WriteString(absQuery)
	}
	if hasAbsFragment {
		b.WriteRune('#')
		b.WriteString(absFragment)
	}
	return ParseRef(b.String())
}
