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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	// TODO: At some point implement my own IDNA2003 module (RFC 3490).
	"golang.org/x/net/idna"
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
// For comparison purposes where canonical equivalence is desired, use `ParseNormalizedRef`
// or the `Normalize()` method.
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
// use `ParseNormalizedRef` instead.
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

// ParseURIToRef converts a URI string into an IRI reference by decoding
// percent-encoded octets that form valid UTF-8 sequences. This is the
// reverse of the ToURI method and follows RFC 3987, Section 3.2.
//
// It cautiously decodes only valid sequences and re-validates the final
// string to ensure it forms a syntactically correct IRI reference. Any
// percent-encoded octets that do not form a valid UTF-8 sequence or that
// represent characters not permitted in IRIs (such as bidi control characters)
// are left in their percent-encoded form.
func ParseURIToRef(s string) (*Ref, error) {
	var builder strings.Builder
	builder.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] != '%' {
			builder.WriteByte(s[i])
			i++
			continue
		}

		start := i
		var decodedBytes []byte
		// Find a contiguous block of percent-encoded octets.
		for i < len(s) && s[i] == '%' {
			if i+2 >= len(s) || !isASCIIHexDigit(rune(s[i+1])) || !isASCIIHexDigit(rune(s[i+2])) {
				// Incomplete or invalid encoding, stop processing this block.
				break
			}
			b, _ := hex.DecodeString(s[i+1 : i+3])
			decodedBytes = append(decodedBytes, b[0])
			i += 3
		}

		// If the inner loop didn't advance, we found an invalid/incomplete sequence.
		if i == start {
			// Write the original '%' and advance past it to prevent an infinite loop.
			builder.WriteByte(s[start])
			i++
			continue
		}

		if validateDecodedBytes(decodedBytes) {
			builder.Write(decodedBytes)
		} else {
			// Not valid UTF-8 or contains forbidden characters, so keep original encoding.
			builder.WriteString(s[start:i])
		}
	}

	// The decoded string must be re-parsed to ensure it is a valid IRI.
	// ParseNormalizedRef is used here because URI-to-IRI conversion
	// implies a canonical representation is desired.
	return ParseNormalizedRef(builder.String())
}

// Resolve resolves a relative IRI reference against the current Ref (which acts as the base IRI).
// It returns a new, absolute Ref. This operation is equivalent to resolving a hyperlink.
func (r *Ref) Resolve(relativeIRI string) (*Ref, error) {
	builder := &strings.Builder{}
	builder.Grow(len(r.iri) + len(relativeIRI)) // Pre-allocate for efficiency.
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

// String returns the underlying string representation of the IRI reference.
// The returned string is not guaranteed to be in any specific Unicode normalization form
// unless the Ref was created with `ParseNormalizedRef` or processed by `Normalize()`.
func (r *Ref) String() string {
	return r.iri
}

// ToURI converts the IRI reference to a URI reference string, strictly following
// RFC 3987, Section 3.1. It normalizes all components to NFC, percent-encodes
// any non-ASCII characters using their UTF-8 representation, and applies IDNA
// (ToASCII) to the host component to ensure the resulting URI is resolvable in DNS.
func (r *Ref) ToURI() string {
	var builder strings.Builder
	builder.Grow(len(r.iri))

	scheme, hasScheme := r.Scheme()
	authority, hasAuthority := r.Authority()
	path := r.Path()
	query, hasQuery := r.Query()
	fragment, hasFragment := r.Fragment()

	if hasScheme {
		builder.WriteString(scheme)
		builder.WriteRune(':')
	}

	if hasAuthority {
		builder.WriteString("//")
		userinfo, host, port := splitAuthority(authority)

		// Per RFC 3987, Section 3.1, Step 1, components must be in NFC
		// before percent-encoding.
		normalizedUserinfo := norm.NFC.String(userinfo)
		percentEncode(normalizedUserinfo, &builder)
		if userinfo != "" {
			builder.WriteRune('@')
		}

		// Normalize host to NFC before applying IDNA.
		normalizedHost := norm.NFC.String(host)

		// Apply IDNA ToASCII to the host for DNS resolvability.
		asciiHost, err := idna.ToASCII(normalizedHost)
		if err == nil {
			builder.WriteString(asciiHost)
		}

		if port != "" {
			builder.WriteRune(':')
			builder.WriteString(port)
		}
	}

	// Normalize path, query, and fragment to NFC before percent-encoding.
	percentEncode(norm.NFC.String(path), &builder)
	if hasQuery {
		builder.WriteRune('?')
		percentEncode(norm.NFC.String(query), &builder)
	}
	if hasFragment {
		builder.WriteRune('#')
		percentEncode(norm.NFC.String(fragment), &builder)
	}

	return builder.String()
}

// Normalize applies syntax-based normalization to the IRI reference according
// to RFC 3986, Section 6.2.2. This includes case-normalization of the scheme
// and host, percent-encoding normalization, and path-segment normalization.
// It also ensures the resulting IRI is in Unicode Normalization Form C (NFC).
// It returns a new, normalized Ref.
func (r *Ref) Normalize() *Ref {
	if r.iri == "" {
		return &Ref{}
	}

	scheme, hasScheme := r.Scheme()
	authority, hasAuthority := r.Authority()
	path := r.Path()
	query, hasQuery := r.Query()
	fragment, hasFragment := r.Fragment()

	// 1. Case Normalization
	if hasScheme {
		scheme = strings.ToLower(scheme)
	}
	var userinfo, host, port string
	if hasAuthority {
		userinfo, host, port = splitAuthority(authority)
		host, port = normalizeHostAndPort(host, port, scheme)
	}

	// 2. Percent-Encoding Normalization
	userinfo = normalizePercentEncoding(userinfo)
	host = normalizePercentEncoding(host)
	path = normalizePercentEncoding(path)
	query = normalizePercentEncoding(query)
	fragment = normalizePercentEncoding(fragment)

	// 3. Path Segment Normalization
	path = removeDotSegments(path)

	// 4. Scheme-based normalization for path
	if hasAuthority && path == "" {
		path = "/"
	}

	// Recompose and re-parse
	recomposedStr := recomposeNormalizedIRI(
		scheme, hasScheme,
		userinfo, host, port, hasAuthority,
		path,
		query, hasQuery,
		fragment, hasFragment,
	)

	normalizedStr := norm.NFC.String(recomposedStr)

	if normalizedStr == r.iri {
		return r
	}
	// An error is not expected here as we are building from valid components.
	// We use the compliant ParseRef because normalizedStr is now guaranteed to be NFC.
	newRef, _ := ParseRef(normalizedStr)
	return newRef
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
// into a Ref, performing validation in the process. It does not perform NFC normalization.
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
// If the string is a relative reference, it returns an error. The string is not
// NFC normalized; for that, use `ParseNormalizedIri`.
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

// ResolveTo resolves a relative IRI and writes the resulting absolute IRI
// to the provided strings.Builder, avoiding allocations.
func (i *Iri) ResolveTo(relativeIRI string, target *strings.Builder) error {
	_, err := i.Ref.ResolveTo(relativeIRI, target)
	return err
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

	for _, segment := range strings.Split(absPath, "/") {
		if segment == "." || segment == ".." {
			return nil, ErrIriRelativize
		}
	}

	if base.Scheme() != abs.Scheme() {
		return ParseRef(abs.String())
	}

	baseAuthority, hasBaseAuthority := base.Authority()
	absAuthority, hasAbsAuthority := abs.Authority()

	if hasBaseAuthority != hasAbsAuthority || (hasBaseAuthority && baseAuthority != absAuthority) {
		if !hasAbsAuthority {
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

	if basePath == absPath {
		return i.relativizeForSamePath(abs)
	}

	if !hasBaseAuthority {
		return i.relativizeForNoAuthority(abs)
	}

	return i.relativizeWithAuthority(abs)
}
