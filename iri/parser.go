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
	"io"
	"strings"
)

const (
	// authorityPrefixLength is the length of the string "//".
	authorityPrefixLength = 2
)

// errNoScheme is returned when an absolute IRI is expected but no scheme is found.
// This typically occurs when the IRI string starts with a colon, which is invalid.
var errNoScheme = &kindError{message: "No scheme found in an absolute IRI"}

// Positions represents the parsed component offsets as defined in the governing specification.
//
// Specification Reference:
// RFC 3986 (Section 3)
//
// Representation:
// A struct holding the exclusive byte boundary markers for scheme, authority, path, and query components within the
// generated output buffer.
type Positions struct {
	// SchemeEnd is the byte offset just past the trailing colon of the scheme
	// (e.g., 5 for "http:"). Zero means no scheme is present.
	SchemeEnd int
	// AuthorityEnd is the byte offset just past the last byte of the authority
	// component (the "//" prefix and "host:port" are both included).
	// Equal to SchemeEnd when no authority is present.
	AuthorityEnd int
	// PathEnd is the byte offset just past the last byte of the path component.
	// Equal to AuthorityEnd when the path is empty.
	PathEnd int
	// QueryEnd is the byte offset just past the last byte of the query value
	// (the "?" delimiter is not included in the value but is counted in the offset).
	// Equal to PathEnd when no query is present.
	QueryEnd int
}

// base represents the parsed base IRI as defined in the governing specification.
//
// Specification Reference:
// RFC 3986 (Section 5.1) and RFC 3987 (Section 6.5)
//
// Representation:
// A data record associating a reference string with its parsed component positions to facilitate standard resolution.
type base struct {
	IRI string
	Pos Positions
}

// iriParserBase represents the base components of an IRI used for reference resolution as defined in the governing
// specification.
//
// Specification Reference:
// RFC 3986 (Section 5.2.1)
//
// Representation:
// A context structure holding cached index boundaries of the base IRI scheme, authority, path, and query components.
type iriParserBase struct {
	iri          string
	schemeEnd    int
	authorityEnd int
	pathEnd      int
	queryEnd     int
	hasBase      bool
}

// iriParser represents the parsing state and buffer configuration of the parser as defined in the governing
// specification.
//
// Specification Reference:
// RFC 3987 (Section 2)
//
// Representation:
// A stateful scanner wrapping input source, optional resolution context, output destination, and parsed position
// metadata.
type iriParser struct {
	iri             string
	base            *iriParserBase
	input           *parserInput
	output          outputBuffer
	outputPositions Positions
	inputSchemeEnd  int
	unchecked       bool
}

// run parses, validates, and resolves an input IRI reference against an optional base IRI.
//
// Specification Reference:
// RFC 3987 (Section 2.2) and RFC 3986 (Section 5.2.2).
//
// Parameters:
//   - iri: The raw IRI string to parse.
//   - baseIRI: An optional pointer to the pre-parsed base IRI component.
//   - unchecked: A boolean flag indicating whether syntax and bidirectional validation should be skipped.
//   - output: The output buffer implementation where the resolved string is written.
//
// Returns:
//   - Positions: The byte-offset positions of each component in the parsed output.
//   - error: An error under parsing or validation failures.
func run(iri string, baseIRI *base, unchecked bool, output outputBuffer) (Positions, error) {
	// Step 1: Initialize Base IRI Components
	// Set up components from the base IRI if provided, which is necessary for relative resolution.
	var b *iriParserBase
	if baseIRI != nil {
		b = &iriParserBase{
			iri:          baseIRI.IRI,
			schemeEnd:    baseIRI.Pos.SchemeEnd,
			authorityEnd: baseIRI.Pos.AuthorityEnd,
			pathEnd:      baseIRI.Pos.PathEnd,
			queryEnd:     baseIRI.Pos.QueryEnd,
			hasBase:      true,
		}
	} else {
		b = &iriParserBase{hasBase: false}
	}

	// Step 2: Initialize Parser State
	// Construct the iriParser instance with the input reader, output buffer, and configuration options.
	p := &iriParser{
		iri:       iri,
		base:      b,
		input:     newParserInput(iri),
		output:    output,
		unchecked: unchecked,
	}

	// Step 3: Begin Parsing Scheme or Relative Parts
	// Delegate execution to the scheme detection entry point.
	err := p.parseSchemeStart()
	return p.outputPositions, err
}

// parseSchemeStart determines the entry state of the parser by checking for a scheme or relative prefix.
//
// Specification Reference:
// RFC 3986 (Section 3.1) and RFC 3986 (Section 4.2).
//
// Returns:
//   - error: An error if the parsing or subsequent processing fails.
func (p *iriParser) parseSchemeStart() error {
	if !p.base.hasBase && strings.HasPrefix(p.iri, "//") {
		// Spec Rule: RFC 3986 (Section 4.2)
		// A relative reference that begins with two slash characters is a network-path reference.
		_, _ = p.input.reader.Seek(authorityPrefixLength, io.SeekStart)
		p.output.writeString("//")
		p.outputPositions.SchemeEnd = 0
		if err := p.parseAuthority(); err != nil {
			return err
		}
		r, ok := p.input.peek()
		return p.parsePathStart(r, ok)
	}

	r, ok := p.input.peek()
	if !ok {
		// Spec Rule: RFC 3986 (Section 4.2)
		// An empty relative reference is valid and resolves to the base path.
		return p.parseRelative()
	}
	if r == ':' {
		// Spec Rule: RFC 3986 (Section 3.1)
		// A scheme must begin with a letter. A colon at the start is invalid.
		return errNoScheme
	}
	if isASCIILetter(r) {
		// Spec Rule: RFC 3986 (Section 3.1)
		// A scheme name begins with an ASCII letter. Even if a reference starts
		// with a scheme, if we are resolving against a base, it must be resolved
		// via resolveComponents to properly normalize dot segments.
		if p.base.hasBase {
			return p.parseRelative()
		}
		return p.parseScheme()
	}
	// Spec Rule: RFC 3986 (Section 4.2)
	// If the string does not start with a scheme-valid prefix, it is parsed as relative.
	return p.parseRelative()
}

// parseScheme consumes the scheme component.
//
// Specification Reference:
// RFC 3986 (Section 3.1).
//
// Returns:
//   - error: An error if parsing or subsequent component validation fails.
func (p *iriParser) parseScheme() error {
	initialInput := p.iri
	initialPos := p.input.position()
	for {
		r, ok := p.input.next()
		if !ok {
			// Implementation Note: Backtracking logic.
			// Reached end of string without finding ':', so we treat the prefix as a relative path.
			p.input.reset(initialInput[initialPos:])
			p.output.reset()
			return p.parseRelative()
		}

		switch {
		case isASCIILetter(r) || isASCIIDigit(r) || r == '+' || r == '-' || r == '.':
			// Spec Rule: RFC 3986 (Section 3.1)
			// Scheme names consist of a sequence of characters beginning with a letter and
			// followed by letters, digits, plus, period, or hyphen.
			p.output.writeRune(r)
		case r == ':':
			// Spec Rule: RFC 3986 (Section 3.1)
			// The colon character marks the end of the scheme component.
			p.output.writeRune(':')
			p.outputPositions.SchemeEnd = p.output.len()
			p.inputSchemeEnd = p.input.position()
			if p.input.startsWith('/') {
				p.input.next()
				p.output.writeRune('/')
				return p.parsePathOrAuthority()
			}
			// Spec Rule: RFC 3986 (Section 3)
			// No authority exists, so path starts immediately.
			p.outputPositions.AuthorityEnd = p.outputPositions.SchemeEnd
			return p.parsePath()
		default:
			// Implementation Note: Backtracking logic.
			// Invalid character for a scheme, reset the parser to treat as a relative reference.
			p.input.reset(initialInput[initialPos:])
			p.output.reset()
			return p.parseRelative()
		}
	}
}

// parseRelativeNoBase handles parsing a relative reference when no base IRI is provided.
//
// Specification Reference:
// RFC 3986 (Section 4.2).
//
// Returns:
//   - error: An error if parsing the relative path fails.
func (p *iriParser) parseRelativeNoBase() error {
	p.outputPositions.SchemeEnd = 0
	p.inputSchemeEnd = 0
	if p.input.startsWith('/') {
		// Spec Rule: RFC 3986 (Section 4.2)
		// A relative reference that begins with a single slash is an absolute-path reference.
		p.input.next()
		p.output.writeRune('/')
		return p.parsePath()
	}
	// Spec Rule: RFC 3986 (Section 4.2)
	// A relative reference that does not begin with a slash is a relative-path reference.
	return p.parsePathNoScheme()
}

// validateRelativeRef runs a sub-parse on the relative reference string to ensure it's well-formed.
//
// Specification Reference:
// RFC 3986 (Section 4.2).
//
// Parameters:
//   - relativeRef: The relative reference string to validate.
//
// Returns:
//   - error: An error if the relative reference is invalid or contains ambiguous colons in the first segment.
func (p *iriParser) validateRelativeRef(relativeRef string) error {
	// Step 1: Sub-parse Initialization
	// Create a temporary parser with a void output buffer to validate the structure.
	validationParser := &iriParser{
		iri:       relativeRef,
		base:      &iriParserBase{hasBase: false},
		input:     newParserInput(relativeRef),
		output:    &voidOutputBuffer{},
		unchecked: false,
	}
	if err := validationParser.parseSchemeStart(); err != nil {
		return err
	}

	// Spec Rule: RFC 3986 (Section 4.2)
	// A path segment that contains a colon character cannot be used as the first segment
	// of a relative-path reference, as it would be mistaken for a scheme name.
	if validationParser.outputPositions.SchemeEnd > 0 {
		uriAfterScheme := relativeRef[validationParser.inputSchemeEnd:]
		if !strings.HasPrefix(uriAfterScheme, "/") {
			return &kindError{message: "Invalid IRI character in first path segment", char: ':'}
		}
	}

	return nil
}

// parseRelative handles a relative IRI reference.
//
// Specification Reference:
// RFC 3986 (Section 5.2) and RFC 3987 (Section 6.5).
//
// Returns:
//   - error: An error if validation or component resolution fails.
func (p *iriParser) parseRelative() error {
	if !p.base.hasBase {
		// Step 1: Parse without a base.
		// Fall back to simple relative-path reference parsing when no base is available.
		return p.parseRelativeNoBase()
	}

	// Step 2: Validate the relative reference.
	// Ensure that the reference is valid and does not violate first-segment ambiguity constraints.
	relativeRef := p.input.asStr()

	// Spec Rule: RFC 3986 (Section 4.2)
	// If the reference is absolute (i.e., contains a valid scheme), it is not a relative-path reference
	// and does not need first-segment colon validation.
	_, _, isAbsolute := extractRefScheme(relativeRef)
	if !isAbsolute {
		if err := p.validateRelativeRef(relativeRef); err != nil {
			return err
		}
	}

	// Step 3: Resolve components.
	// Apply the reference resolution algorithm against the established base IRI.
	t := p.resolveComponents(relativeRef)
	p.recomposeIRI(t)
	return nil
}
