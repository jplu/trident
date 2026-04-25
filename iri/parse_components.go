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

// errPathStartingWithSlashes is returned when an IRI has a path that
// starts with "//" but does not have an authority component. This is
// disallowed by RFC 3987 to avoid ambiguity with network-path references.
// For example, "scheme:////path" is valid, but "scheme:/path" where the path
// starts with `//` is not.
var errPathStartingWithSlashes = &kindError{
	message: "An IRI path is not allowed to start with // if there is no authority",
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
