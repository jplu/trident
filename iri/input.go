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

import "strings"

// parserInput provides a reader-like interface over the input string,
// allowing for peeking, advancing, and position tracking.
type parserInput struct {
	originalString string
	reader         *strings.Reader
}

// newParserInput creates a new parserInput wrapping the given string.
func newParserInput(s string) *parserInput {
	return &parserInput{
		originalString: s,
		reader:         strings.NewReader(s),
	}
}

// next reads and returns the next rune from the input, advancing the position.
func (p *parserInput) next() (rune, bool) {
	r, _, err := p.reader.ReadRune()
	return r, err == nil
}

// peek returns the next rune from the input without advancing the position.
func (p *parserInput) peek() (rune, bool) {
	r, _, err := p.reader.ReadRune()
	if err != nil {
		return 0, false
	}
	_ = p.reader.UnreadRune()
	return r, true
}

// startsWith checks if the remaining input starts with the given rune.
func (p *parserInput) startsWith(r rune) bool {
	pr, ok := p.peek()
	return ok && pr == r
}

// position returns the current read position in bytes from the start of the original string.
func (p *parserInput) position() int {
	return len(p.originalString) - p.reader.Len()
}

// asStr returns the unread portion of the input string.
func (p *parserInput) asStr() string {
	return p.originalString[p.position():]
}

// reset re-initializes the input with a new string.
func (p *parserInput) reset(s string) {
	p.originalString = s
	p.reader = strings.NewReader(s)
}
