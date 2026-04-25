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

// errNoScheme is returned when an absolute IRI is expected but no scheme
// (e.g., "http:") is found. This typically occurs when the IRI string
// starts with a colon, which is invalid.
var errNoScheme = &kindError{message: "No scheme found in an absolute IRI"}

// Positions holds the end byte-offsets of each component within a parsed IRI string.
// All values are exclusive upper bounds, measured in bytes from the start of the
// output string produced by the parser.
//
// The components are laid out as follows for "http://user@host:80/path?query#frag":
//
//	SchemeEnd    — byte index immediately after the scheme colon, e.g. 5 for "http:"
//	AuthorityEnd — byte index immediately after the authority, e.g. 21 for "//user@host:80"
//	PathEnd      — byte index immediately after the path, e.g. 26 for "/path"
//	QueryEnd     — byte index immediately after the query, e.g. 32 for "query"
//
// When a component is absent its end equals the previous component's end.
// For example, if there is no authority, AuthorityEnd == SchemeEnd.
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

// base represents a pre-parsed, absolute IRI that can be used as a base for
// resolving relative references.
type base struct {
	IRI string
	Pos Positions
}

// iriParserBase holds the component data of a base IRI used for resolution.
type iriParserBase struct {
	iri          string
	schemeEnd    int
	authorityEnd int
	pathEnd      int
	queryEnd     int
	hasBase      bool
}

// iriParser holds the state for a single parsing operation.
type iriParser struct {
	iri             string
	base            *iriParserBase
	input           *parserInput
	output          outputBuffer
	outputPositions Positions
	inputSchemeEnd  int
	unchecked       bool
}

// run is the main entry point for the IRI parser. It parses, validates, and
// resolves an IRI reference against an optional base IRI.
func run(iri string, baseIRI *base, unchecked bool, output outputBuffer) (Positions, error) {
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

	p := &iriParser{
		iri:       iri,
		base:      b,
		input:     newParserInput(iri),
		output:    output,
		unchecked: unchecked,
	}

	err := p.parseSchemeStart()
	return p.outputPositions, err
}

// parseSchemeStart is the initial state of the parser.
func (p *iriParser) parseSchemeStart() error {
	if !p.base.hasBase && strings.HasPrefix(p.iri, "//") {
		// This is a network-path reference like "//example.com/path"
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
		// Empty input, treat as relative reference.
		return p.parseRelative()
	}
	if r == ':' {
		return errNoScheme
	}
	if isASCIILetter(r) {
		return p.parseScheme()
	}
	// No scheme found, treat as a relative reference.
	return p.parseRelative()
}

// parseScheme consumes the scheme component.
func (p *iriParser) parseScheme() error {
	initialInput := p.iri
	initialPos := p.input.position()
	for {
		r, ok := p.input.next()
		if !ok {
			// Reached end of string without finding ':', so it's a relative path.
			p.input.reset(initialInput[initialPos:])
			p.output.reset()
			return p.parseRelative()
		}

		switch {
		case isASCIILetter(r) || isASCIIDigit(r) || r == '+' || r == '-' || r == '.':
			p.output.writeRune(r)
		case r == ':':
			p.output.writeRune(':')
			p.outputPositions.SchemeEnd = p.output.len()
			p.inputSchemeEnd = p.input.position()
			if p.input.startsWith('/') {
				p.input.next()
				p.output.writeRune('/')
				return p.parsePathOrAuthority()
			}
			// No authority, path starts immediately.
			p.outputPositions.AuthorityEnd = p.outputPositions.SchemeEnd
			return p.parsePath()
		default:
			// Invalid character for a scheme, so it must be a relative path.
			p.input.reset(initialInput[initialPos:])
			p.output.reset()
			return p.parseRelative()
		}
	}
}

// parseRelativeNoBase handles parsing a relative reference when no base IRI is provided.
// In this case, it's parsed as a relative-path reference.
func (p *iriParser) parseRelativeNoBase() error {
	p.outputPositions.SchemeEnd = 0
	p.inputSchemeEnd = 0
	if p.input.startsWith('/') {
		p.input.next()
		p.output.writeRune('/')
		return p.parsePath()
	}
	return p.parsePathNoScheme()
}

// validateRelativeRef runs a sub-parse on the relative reference string to ensure it's well-formed.
func (p *iriParser) validateRelativeRef(relativeRef string) error {
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

	// According to RFC 3986 Section 4.2, a relative-path reference cannot
	// contain a colon in its first segment, as it would be mistaken for a scheme.
	// The generic parser will correctly parse such a string (e.g., "a:b") as an
	// absolute URI with scheme "a".
	// Since this validation function is specifically for references to be resolved
	// against a base, we must reject this ambiguous form.
	if validationParser.outputPositions.SchemeEnd > 0 {
		// It was parsed as an absolute URI. Check if it's the ambiguous form.
		// The ambiguous form is `scheme:path-rootless`.
		// It's unambiguous if it has an authority (`scheme://...`) or absolute path (`scheme:/...`).
		uriAfterScheme := relativeRef[validationParser.inputSchemeEnd:]
		if !strings.HasPrefix(uriAfterScheme, "/") {
			// This is the ambiguous case (e.g., "a:b"). Per RFC 3986, this form
			// is invalid as a relative-path reference.
			return &kindError{message: "Invalid IRI character in first path segment", char: ':'}
		}
	}

	return nil
}

// parseRelative handles a relative IRI reference. If a base IRI is present,
// it resolves the reference against the base. Otherwise, it parses it as a
// relative-path reference.
func (p *iriParser) parseRelative() error {
	if !p.base.hasBase {
		return p.parseRelativeNoBase()
	}

	relativeRef := p.input.asStr()
	if err := p.validateRelativeRef(relativeRef); err != nil {
		return err
	}

	t := p.resolveComponents(relativeRef)
	p.recomposeIRI(t)
	return nil
}
