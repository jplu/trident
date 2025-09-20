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
	"io"
	"strings"
)

const (
	// authorityPrefixLength is the length of the string "//".
	authorityPrefixLength = 2
)

// Positions holds the end indices of the various components of a parsed IRI.
// See the full struct definition in iri_parser.go for detailed documentation.
type Positions struct {
	SchemeEnd    int
	AuthorityEnd int
	PathEnd      int
	QueryEnd     int
}

// base represents a pre-parsed, absolute IRI that can be used as a base for
// resolving relative references.
type base struct {
	IRI string
	Pos Positions
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

// parsePathOrAuthority handles the part of the IRI after "scheme:/".
func (p *iriParser) parsePathOrAuthority() error {
	if p.input.startsWith('/') {
		// This is an authority-based IRI like "scheme://host/path"
		p.input.next()
		p.output.writeRune('/')
		if err := p.parseAuthority(); err != nil {
			return err
		}
		r, ok := p.input.peek()
		return p.parsePathStart(r, ok)
	}
	// No second slash, so no authority. Path starts here.
	p.outputPositions.AuthorityEnd = p.outputPositions.SchemeEnd
	return p.parsePath()
}

// parsePathStart begins parsing the path component.
func (p *iriParser) parsePathStart(r rune, ok bool) error {
	if !ok {
		// End of input after authority (e.g., "http://host.com")
		p.outputPositions.PathEnd = p.output.len()
		p.outputPositions.QueryEnd = p.output.len()
		return nil
	}

	// The dispatcher logic determines what to parse next based on the first character.
	switch r {
	case '?':
		p.input.next() // Consume '?'
		p.outputPositions.PathEnd = p.output.len()
		p.output.writeRune('?')
		return p.parseQuery()
	case '#':
		p.input.next() // Consume '#'
		p.outputPositions.PathEnd = p.output.len()
		p.outputPositions.QueryEnd = p.output.len()
		p.output.writeRune('#')
		return p.parseFragment()
	case '/':
		p.input.next() // Consume '/'
		p.output.writeRune('/')
		return p.parsePath()
	default:
		p.input.next() // Consume the character
		// This is the first character of the first path segment.
		if err := p.readURLCodepointOrEchar(r, func(c rune) bool {
			return isIUnreservedOrSubDelims(c) || c == ':' || c == '@'
		}); err != nil {
			return err
		}
		return p.parsePath()
	}
}

// parsePathNoScheme parses a path that is not preceded by a scheme.
func (p *iriParser) parsePathNoScheme() error {
	for {
		c, ok := p.input.peek()
		if !ok || c == '/' || c == '?' || c == '#' {
			break
		}
		if c == ':' {
			// RFC 3986, Section 4.2: A path segment that contains a colon
			// cannot be used as the first segment of a relative-path reference.
			return &kindError{message: "Invalid IRI character in first path segment", char: c}
		}
		p.input.next()
		if err := p.readURLCodepointOrEchar(c, func(r rune) bool {
			return isIUnreservedOrSubDelims(r) || r == '@'
		}); err != nil {
			return err
		}
	}
	return p.parsePath()
}

// validateBidiPart checks the bidi validity of the current component part if validation is enabled.
func (p *iriParser) validateBidiPart(startIndex int) error {
	if p.unchecked {
		return nil
	}
	if _, ok := p.output.(*voidOutputBuffer); ok {
		return nil
	}
	part := p.output.string()[startIndex:]
	return validateBidiComponent(part)
}

// handlePathTerminator checks for and processes path terminators ('?' or '#').
// It returns true if a terminator was found and handled, along with any error.
func (p *iriParser) handlePathTerminator(c rune, segmentStartIndex int) (bool, error) {
	if c != '?' && c != '#' {
		return false, nil
	}

	if err := p.validateBidiPart(segmentStartIndex); err != nil {
		return true, err
	}

	p.input.next()
	p.outputPositions.PathEnd = p.output.len()

	if c == '?' {
		p.output.writeRune('?')
		return true, p.parseQuery()
	}

	// c must be '#'
	p.outputPositions.QueryEnd = p.output.len()
	p.output.writeRune('#')
	return true, p.parseFragment()
}

// isPathChar is a predicate for characters allowed in a path.
func isPathChar(c rune) bool {
	return isIUnreservedOrSubDelims(c) || c == ':' || c == '@' || c == '/'
}

// parsePath consumes the path component of the IRI.
func (p *iriParser) parsePath() error {
	hasAuthority := p.outputPositions.AuthorityEnd > p.outputPositions.SchemeEnd
	var prev rune
	segmentStartIndex := p.output.len()

	for {
		c, ok := p.input.peek()
		if !ok {
			break
		}

		isTerminator, err := p.handlePathTerminator(c, segmentStartIndex)
		if isTerminator {
			return err
		}

		// RFC 3986, Section 3.3: if a URI does not contain an authority component,
		// then the path cannot begin with two slash characters ("//").
		if !hasAuthority && c == '/' && prev == '/' {
			return errPathStartingWithSlashes
		}

		p.input.next()
		if c == '/' {
			if err = p.validateBidiPart(segmentStartIndex); err != nil {
				return err
			}
		}
		err = p.readURLCodepointOrEchar(c, isPathChar)
		if err != nil {
			return err
		}
		if c == '/' {
			segmentStartIndex = p.output.len()
		}
		prev = c
	}

	if err := p.validateBidiPart(segmentStartIndex); err != nil {
		return err
	}

	p.outputPositions.PathEnd = p.output.len()
	p.outputPositions.QueryEnd = p.output.len()
	return nil
}

// isQueryChar is a predicate for characters allowed in a query.
func isQueryChar(c rune) bool {
	return isIUnreservedOrSubDelims(c) || c == ':' || c == '@' || c == '/' || c == '?' ||
		(c >= '\uE000' && c <= '\uF8FF') || // iprivate
		(c >= 0xF0000 && c <= 0xFFFFD) ||
		(c >= 0x100000 && c <= 0x10FFFD)
}

// handleQueryEnd handles the end of a query, either by EOF or a '#' terminator.
func (p *iriParser) handleQueryEnd(isFragment bool, queryStart int) error {
	if err := p.validateBidiPart(queryStart); err != nil {
		return err
	}
	p.outputPositions.QueryEnd = p.output.len()
	if !isFragment {
		return nil
	}
	p.input.next()
	p.output.writeRune('#')
	return p.parseFragment()
}

// parseQuery consumes the query component.
func (p *iriParser) parseQuery() error {
	queryStart := p.output.len()
	for {
		r, ok := p.input.peek()
		if !ok {
			return p.handleQueryEnd(false, queryStart)
		}
		if r == '#' {
			return p.handleQueryEnd(true, queryStart)
		}
		p.input.next()
		if err := p.readURLCodepointOrEchar(r, isQueryChar); err != nil {
			return err
		}
	}
}

// parseFragment consumes the fragment component.
func (p *iriParser) parseFragment() error {
	fragmentStart := p.output.len()
	for {
		r, ok := p.input.next()
		if !ok {
			if !p.unchecked {
				if err := p.validateBidiPart(fragmentStart); err != nil {
					return err
				}
			}
			return nil
		}
		err := p.readURLCodepointOrEchar(r, func(c rune) bool {
			return isIUnreservedOrSubDelims(c) || c == ':' || c == '@' || c == '/' || c == '?'
		})
		if err != nil {
			return err
		}
	}
}
