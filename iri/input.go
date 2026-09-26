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

const (
	// hexDigitsLength represents the number of hexadecimal characters required in a percent-encoded octet.
	hexDigitsLength = 2
)

// parserInput represents the input stream reader as defined in the parser implementation.
//
// Representation:
// A stateful cursor wrapping a strings.Reader that provides sequential character (rune) access, peeking, and position
// tracking.
type parserInput struct {
	originalString string
	reader         *strings.Reader
}

// newParserInput instantiates a new parserInput instance.
//
// Parameters:
//   - s: The raw input string to be processed.
//
// Returns:
//   - *parserInput: A pointer to the newly allocated parserInput.
func newParserInput(s string) *parserInput {
	return &parserInput{
		originalString: s,
		reader:         strings.NewReader(s),
	}
}

// next reads and returns the next rune from the input, advancing the position.
//
// Returns:
//   - rune: The Unicode code point read from the input stream.
//   - bool: True if a character was successfully read, false otherwise.
func (p *parserInput) next() (rune, bool) {
	r, _, err := p.reader.ReadRune()
	return r, err == nil
}

// peek returns the next rune from the input without advancing the position.
//
// Returns:
//   - rune: The Unicode code point peeked from the input stream.
//   - bool: True if a character was successfully peeked, false otherwise.
func (p *parserInput) peek() (rune, bool) {
	r, _, err := p.reader.ReadRune()
	if err != nil {
		return 0, false
	}
	_ = p.reader.UnreadRune()
	return r, true
}

// startsWith checks if the remaining input starts with the given rune.
//
// Parameters:
//   - r: The target rune to check at the current position.
//
// Returns:
//   - bool: True if the next unconsumed character is equal to r, false otherwise.
func (p *parserInput) startsWith(r rune) bool {
	pr, ok := p.peek()
	return ok && pr == r
}

// position returns the current read position in bytes from the start of the original string.
//
// Returns:
//   - int: The absolute byte offset of the unread input within the original string.
func (p *parserInput) position() int {
	return len(p.originalString) - p.reader.Len()
}

// asStr returns the unread portion of the input string.
//
// Returns:
//   - string: The remaining unconsumed slice of the original string.
func (p *parserInput) asStr() string {
	return p.originalString[p.position():]
}

// reset re-initializes the input reader with a new target string.
//
// Parameters:
//   - s: The new raw input string to parse.
func (p *parserInput) reset(s string) {
	p.originalString = s
	p.reader = strings.NewReader(s)
}

// hasHexDigits checks if the next two characters in the input stream represent valid hexadecimal digits.
//
// Specification Reference:
// RFC 3986 (Section 2.1)
//
// Returns:
//   - bool: True if the next two bytes conform to the ASCII hexadecimal range, false otherwise.
func (p *parserInput) hasHexDigits() bool {
	pos := p.position()
	// Spec Rule: RFC 3986 (Section 2.1)
	// A percent-encoded octet must consist of the percent character '%' followed by two hexadecimal digits.
	if len(p.originalString)-pos < hexDigitsLength {
		return false
	}
	return isASCIIHexDigit(rune(p.originalString[pos])) &&
		isASCIIHexDigit(rune(p.originalString[pos+1]))
}
