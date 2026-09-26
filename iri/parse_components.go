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

import "strings"

// errPathStartingWithSlashes represents the error returned when an IRI path starts with double slashes but lacks an
// authority component.
//
// Specification Reference:
// RFC 3986 (Section 3.3) and RFC 3987 (Section 2.2)
//
// Representation:
// This is a kindError value returned to prevent parsing ambiguity where a path component is mistaken for a network-path
// reference.
var errPathStartingWithSlashes = &kindError{
	message: "An IRI path is not allowed to start with // if there is no authority",
}

// parsePathOrAuthority parses the component following the initial scheme delimiter slash to distinguish between an
// authority and an absolute path.
//
// Specification Reference:
// RFC 3986 (Section 3.2) and RFC 3987 (Section 2.2)
//
// Returns:
//   - error: An error if authority or path parsing fails.
func (p *iriParser) parsePathOrAuthority() error {
	if p.input.startsWith('/') {
		// Spec Rule: RFC 3986 (Section 3) and RFC 3987 (Section 2.2)
		// A double slash indicates that an authority component follows.
		p.input.next()
		p.output.writeRune('/')
		if err := p.parseAuthority(); err != nil {
			return err
		}
		r, ok := p.input.peek()
		return p.parsePathStart(r, ok)
	}
	// Spec Rule: RFC 3986 (Section 3) and RFC 3987 (Section 2.2)
	// A single slash after the scheme indicates the start of an absolute path without authority.
	p.outputPositions.AuthorityEnd = p.outputPositions.SchemeEnd
	return p.parsePath()
}

// parsePathStart dispatches the parser to the appropriate component based on the first character of the path or
// trailing delimiters.
//
// Specification Reference:
// RFC 3986 (Section 3.3) and RFC 3987 (Section 2.2)
//
// Parameters:
//   - r: The first rune of the path or trailing component.
//   - ok: A boolean indicating if a rune was successfully read from the input stream.
//
// Returns:
//   - error: An error if parsing the path, query, or fragment fails.
func (p *iriParser) parsePathStart(r rune, ok bool) error {
	if !ok {
		// Spec Rule: RFC 3986 (Section 3.3) and RFC 3987 (Section 2.2)
		// An empty path is valid after an authority component when the input terminates.
		p.outputPositions.PathEnd = p.output.len()
		p.outputPositions.QueryEnd = p.output.len()
		return nil
	}

	// Step 1: Component Dispatch
	// Dispatch the parser to query, fragment, or path depending on the leading character.
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
		// Spec Rule: RFC 3987 (Section 2.2)
		// The first character of an authority-based absolute path must belong to the allowed set of path characters.
		if err := p.readURLCodepointOrEchar(r, func(c rune) bool {
			return isIUnreservedOrSubDelims(c) || c == ':' || c == '@'
		}); err != nil {
			return err
		}
		return p.parsePath()
	}
}

// parsePathNoScheme parses a path component that is not preceded by a scheme.
//
// Specification Reference:
// RFC 3986 (Section 4.2) and RFC 3987 (Section 2.2)
//
// Returns:
// - error: An error if an invalid character such as a colon is encountered in the first path segment, or if path
// parsing fails.
func (p *iriParser) parsePathNoScheme() error {
	for {
		c, ok := p.input.peek()
		if !ok || c == '/' || c == '?' || c == '#' {
			break
		}
		if c == ':' {
			// Spec Rule: RFC 3986 (Section 4.2) and RFC 3987 (Section 2.2)
			// A path segment containing a colon cannot be used as the first segment of a relative-path reference.
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

// handlePathTerminator checks for and processes path termination characters including query and fragment delimiters.
//
// Specification Reference:
// RFC 3986 (Section 3.3) and RFC 3987 (Section 2.2)
//
// Parameters:
//   - c: The candidate delimiter rune being checked.
//
// Returns:
//   - bool: A boolean indicating if a terminator was found and handled.
//   - error: An error if parsing of the subsequent components fails.
func (p *iriParser) handlePathTerminator(c rune) (bool, error) {
	if c != '?' && c != '#' {
		return false, nil
	}

	p.input.next()
	p.outputPositions.PathEnd = p.output.len()

	if c == '?' {
		p.output.writeRune('?')
		return true, p.parseQuery()
	}

	// Spec Rule: RFC 3986 (Section 3.5) and RFC 3987 (Section 2.2)
	// Treat the '#' character as the start of the fragment component when query is absent.
	p.outputPositions.QueryEnd = p.output.len()
	p.output.writeRune('#')
	return true, p.parseFragment()
}

// isPathChar determines whether a rune is a permitted character within a path component.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Parameters:
//   - c: The candidate rune to check.
//
// Returns:
//   - bool: True if the character is allowed in a path component, false otherwise.
func isPathChar(c rune) bool {
	return isIUnreservedOrSubDelims(c) || c == ':' || c == '@' || c == '/'
}

// parsePath parses and consumes the path component from the input stream.
//
// Specification Reference:
// RFC 3986 (Section 3.3) and RFC 3987 (Section 2.2)
//
// Returns:
//   - error: An error if an illegal path structure is detected, or parsing fails.
func (p *iriParser) parsePath() error {
	hasAuthority := p.outputPositions.AuthorityEnd > p.outputPositions.SchemeEnd

	// Spec Rule: RFC 3986 (Section 3.3) and RFC 3987 (Section 2.2)
	// If an IRI does not contain an authority component, then the path cannot begin with two slash characters ("//").
	if !hasAuthority && strings.HasPrefix(p.iri[p.inputSchemeEnd:], "//") {
		return errPathStartingWithSlashes
	}

	// Step 1: Segment Parsing Loop
	// Iterate through characters, identifying path segments and checking for terminators.
	for {
		c, ok := p.input.peek()
		if !ok {
			break
		}

		isTerminator, err := p.handlePathTerminator(c)
		if isTerminator {
			return err
		}

		p.input.next()
		err = p.readURLCodepointOrEchar(c, isPathChar)
		if err != nil {
			return err
		}
	}

	p.outputPositions.PathEnd = p.output.len()
	p.outputPositions.QueryEnd = p.output.len()
	return nil
}

// isQueryChar determines whether a rune is a permitted character within a query component.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Parameters:
//   - c: The candidate rune to check.
//
// Returns:
//   - bool: True if the character is allowed in a query component, false otherwise.
func isQueryChar(c rune) bool {
	return isIUnreservedOrSubDelims(c) || c == ':' || c == '@' || c == '/' || c == '?' ||
		// Spec Rule: RFC 3987 (Section 2.2)
		// Query components are permitted to contain private-use Unicode characters.
		(c >= '\uE000' && c <= '\uF8FF') || // iprivate
		(c >= 0xF0000 && c <= 0xFFFFD) ||
		(c >= 0x100000 && c <= 0x10FFFD)
}

// handleQueryEnd finalizes query component parsing and optionally transitions to the fragment component parser.
//
// Specification Reference:
// RFC 3986 (Section 3.4) and RFC 3987 (Section 2.2)
//
// Parameters:
//   - isFragment: A boolean indicating whether a fragment component follows the query.
//
// Returns:
//   - error: An error if fragment parsing fails.
func (p *iriParser) handleQueryEnd(isFragment bool) error {
	p.outputPositions.QueryEnd = p.output.len()
	if !isFragment {
		return nil
	}
	p.input.next()
	p.output.writeRune('#')
	return p.parseFragment()
}

// parseQuery parses and consumes the query component from the input stream.
//
// Specification Reference:
// RFC 3986 (Section 3.4) and RFC 3987 (Section 2.2)
//
// Returns:
//   - error: An error if an invalid character is encountered.
func (p *iriParser) parseQuery() error {
	// Step 1: Parse Query Characters
	// Read and validate characters until EOF or fragment delimiter is reached.
	for {
		r, ok := p.input.peek()
		if !ok {
			return p.handleQueryEnd(false)
		}
		if r == '#' {
			return p.handleQueryEnd(true)
		}
		p.input.next()
		if err := p.readURLCodepointOrEchar(r, isQueryChar); err != nil {
			return err
		}
	}
}

// parseFragment parses and consumes the fragment component from the input stream.
//
// Specification Reference:
// RFC 3986 (Section 3.5) and RFC 3987 (Section 2.2)
//
// Returns:
//   - error: An error if an invalid character is encountered.
func (p *iriParser) parseFragment() error {
	// Step 1: Parse Fragment Characters
	// Consume and validate characters until the end of the input stream is reached.
	for {
		r, ok := p.input.next()
		if !ok {
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
