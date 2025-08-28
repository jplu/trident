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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package iri

import (
	"testing"
)

// TestNewParserInput verifies that the parserInput is initialized correctly.
// The test uses a string from RFC 3986, Section 1.1.2, which provides examples of various URI schemes.
func TestNewParserInput(t *testing.T) {
	t.Parallel()
	// RFC 3986, Section 1.1.2, provides a list of example URIs.
	// We use one of these to ensure the constructor handles valid, standard input.
	rfc3986Example := "mailto:John.Doe@example.com"

	tests := []struct {
		name         string
		input        string
		expectedStr  string
		expectedPos  int
		expectReader bool
	}{
		{
			name:         "Standard RFC3986 IRI",
			input:        rfc3986Example,
			expectedStr:  rfc3986Example,
			expectedPos:  0,
			expectReader: true,
		},
		{
			name:         "Empty String",
			input:        "",
			expectedStr:  "",
			expectedPos:  0,
			expectReader: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := newParserInput(tt.input)

			if p == nil {
				t.Fatalf("newParserInput() returned nil")
			}
			if (p.reader == nil) != !tt.expectReader {
				t.Errorf("newParserInput().reader is nil = %v, want %v", p.reader == nil, tt.expectReader)
			}
			if p.originalString != tt.expectedStr {
				t.Errorf("newParserInput().originalString = %q, want %q", p.originalString, tt.expectedStr)
			}
			if pos := p.position(); pos != tt.expectedPos {
				t.Errorf("newParserInput().position() = %d, want %d", pos, tt.expectedPos)
			}
		})
	}
}

// TestParserInput_Next tests the rune-by-rune consumption of the input string.
// RFC 3987 defines IRIs as a sequence of characters from the Universal Character Set.
// This test uses an example containing multi-byte UTF-8 characters to ensure `next()`
// correctly processes runes, not bytes, in compliance with the character-based definition.
func TestParserInput_Next(t *testing.T) {
	t.Parallel()
	// An example from RFC 3987, Section 3.1, using the character 'é' (U+00E9), which is
	// represented by two bytes in UTF-8 (0xC3 0xA9). This is ideal for testing rune-based processing.
	rfc3987Example := "résumé"
	expectedRunes := []rune(rfc3987Example)

	tests := []struct {
		name          string
		input         string
		expectedRunes []rune
	}{
		{
			name:          "RFC3987 Multi-byte string",
			input:         rfc3987Example,
			expectedRunes: expectedRunes,
		},
		{
			name:          "Simple ASCII string",
			input:         "abc",
			expectedRunes: []rune{'a', 'b', 'c'},
		},
		{
			name:          "Empty string",
			input:         "",
			expectedRunes: []rune{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := newParserInput(tt.input)

			for i, expectedRune := range tt.expectedRunes {
				r, ok := p.next()
				if !ok {
					t.Fatalf("next() returned ok=false unexpectedly at index %d", i)
				}
				if r != expectedRune {
					t.Errorf("next() at index %d returned rune %q, want %q", i, r, expectedRune)
				}
			}

			// After consuming all runes, next() should signal end of input.
			r, ok := p.next()
			if ok {
				t.Errorf("next() returned ok=true at end of input, with rune %q", r)
			}
			if r != 0 {
				t.Errorf("next() returned non-zero rune %q at end of input", r)
			}
		})
	}
}

// TestParserInput_Peek verifies reading the next rune without advancing the cursor.
// This is fundamental for a parser that needs to look ahead to apply grammar rules,
// such as those in RFC 3986 Appendix A, without consuming the input.
func TestParserInput_Peek(t *testing.T) {
	t.Parallel()
	// A URI scheme, as defined in RFC 3986, Section 3.1, consists of a letter
	// followed by letters, digits, "+", "-", or ".". We use "http:" to test peeking.
	inputStr := "http:"

	p := newParserInput(inputStr)

	// Peek at the first character
	r, ok := p.peek()
	if !ok || r != 'h' {
		t.Errorf("peek() at start returned %q, %v, want 'h', true", r, ok)
	}
	if pos := p.position(); pos != 0 {
		t.Errorf("position() after peek() is %d, want 0", pos)
	}

	// Consume characters, peeking before each consumption
	for i, expectedRune := range []rune(inputStr) {
		r, ok = p.peek()
		if !ok || r != expectedRune {
			t.Errorf("peek() at index %d returned %q, %v, want %q, true", i, r, ok, expectedRune)
		}

		r, ok = p.next()
		if !ok || r != expectedRune {
			t.Errorf("next() at index %d returned %q, %v, want %q, true", i, r, ok, expectedRune)
		}
	}

	// Peek at the end of the string
	r, ok = p.peek()
	if ok || r != 0 {
		t.Errorf("peek() at end of string returned %q, %v, want 0, false", r, ok)
	}

	// Test on empty string
	pEmpty := newParserInput("")
	r, ok = pEmpty.peek()
	if ok || r != 0 {
		t.Errorf("peek() on empty string returned %q, %v, want 0, false", r, ok)
	}
}

// TestParserInput_StartsWith checks if the remaining input begins with a specific rune.
// RFC 3986 defines component delimiters like ':', '/', '?', '#'. This test
// verifies the utility function used to identify those delimiters.
func TestParserInput_StartsWith(t *testing.T) {
	t.Parallel()
	// From RFC 3986, Section 3, an example URI with all components.
	inputStr := "//example.com/path?query#fragment"
	p := newParserInput(inputStr)

	if !p.startsWith('/') {
		t.Error("startsWith('/') at start should be true")
	}
	if p.startsWith('e') {
		t.Error("startsWith('e') at start should be false")
	}

	// Advance past "//"
	p.next()
	p.next()

	if !p.startsWith('e') {
		t.Error("startsWith('e') after advancing should be true")
	}

	// Test at the end of input
	p.reset("")
	if p.startsWith('a') {
		t.Error("startsWith('a') on empty string should be false")
	}
}

// TestParserInput_Position validates the byte position tracking within the input.
// This is critical for correctly parsing multi-byte characters as defined in RFC 3987.
// The position must advance by the number of bytes in a rune, not by 1 for each character.
func TestParserInput_Position(t *testing.T) {
	t.Parallel()
	// String with ASCII and multi-byte characters to test byte-wise position.
	// f (1 byte), ó (2 bytes: 0xc3 0xb3), o (1 byte)
	inputStr := "fóo"
	p := newParserInput(inputStr)

	// Check initial position
	if pos := p.position(); pos != 0 {
		t.Errorf("Initial position = %d, want 0", pos)
	}

	// After reading 'f' (1 byte)
	p.next()
	if pos := p.position(); pos != 1 {
		t.Errorf("Position after 'f' = %d, want 1", pos)
	}

	// After reading 'ó' (2 bytes)
	p.next()
	if pos := p.position(); pos != 3 {
		t.Errorf("Position after 'ó' = %d, want 3", pos)
	}

	// After reading 'o' (1 byte)
	p.next()
	if pos := p.position(); pos != 4 {
		t.Errorf("Position after 'o' = %d, want 4", pos)
	}

	// At end of input
	p.next() // Read past the end
	if pos := p.position(); pos != 4 {
		t.Errorf("Position at EOF = %d, want 4", pos)
	}
}

// TestParserInput_AsStr ensures the unread portion of the string is returned correctly.
// A parser often needs the remaining substring to pass to sub-parsers, for example,
// when parsing the authority part of a URI as defined in RFC 3986, Section 3.2.
func TestParserInput_AsStr(t *testing.T) {
	t.Parallel()
	// A simple hier-part from RFC 3986, Section 3.
	inputStr := "//a/b/c"
	p := newParserInput(inputStr)

	// Initial state
	if s := p.asStr(); s != "//a/b/c" {
		t.Errorf("Initial asStr() = %q, want %q", s, "//a/b/c")
	}

	// After reading "//"
	p.next()
	p.next()
	if s := p.asStr(); s != "a/b/c" {
		t.Errorf("asStr() after consuming '//' = %q, want %q", s, "a/b/c")
	}

	// After reading "a/b/"
	p.next()
	p.next()
	p.next()
	p.next()
	if s := p.asStr(); s != "c" {
		t.Errorf("asStr() after consuming 'a/b/' = %q, want %q", s, "c")
	}

	// After reading the final 'c'
	p.next()
	if s := p.asStr(); s != "" {
		t.Errorf("asStr() at end of input = %q, want %q", s, "")
	}
}

// TestParserInput_Reset verifies that the input can be re-initialized with a new string.
// This is a utility feature for reusing a parser object, reducing allocations.
func TestParserInput_Reset(t *testing.T) {
	t.Parallel()
	// Use examples from RFC 3986, Section 5.4.1 to simulate parsing different references.
	initialRef := "g:h"
	newRef := "./g"

	p := newParserInput(initialRef)

	// Consume part of the initial string
	p.next()
	if pos := p.position(); pos != 1 {
		t.Fatalf("Position before reset was %d, expected 1", pos)
	}
	if s := p.asStr(); s != ":h" {
		t.Fatalf("asStr() before reset was %q, expected %q", s, ":h")
	}

	// Reset the parser
	p.reset(newRef)

	// Verify the state is reset
	if p.originalString != newRef {
		t.Errorf("originalString after reset = %q, want %q", p.originalString, newRef)
	}
	if pos := p.position(); pos != 0 {
		t.Errorf("position() after reset = %d, want 0", pos)
	}
	if s := p.asStr(); s != newRef {
		t.Errorf("asStr() after reset = %q, want %q", s, newRef)
	}
	r, ok := p.peek()
	if !ok || r != '.' {
		t.Errorf("peek() after reset returned %q, %v, want '.', true", r, ok)
	}
	r, ok = p.next()
	if !ok || r != '.' {
		t.Errorf("next() after reset returned %q, %v, want '.', true", r, ok)
	}
	if pos := p.position(); pos != 1 {
		t.Errorf("position() after next() after reset = %d, want 1", pos)
	}
}

// TestParserInput_CombinedOperations performs a sequence of operations to ensure they
// interact correctly, simulating a real parsing scenario on a complex URI.
func TestParserInput_CombinedOperations(t *testing.T) {
	t.Parallel()
	// A complex URI from RFC 3986, Section 3, covering many syntax elements.
	inputStr := "foo://example.com:8042/over/there?name=ferret#nose"
	p := newParserInput(inputStr)

	// Consume "foo:"
	for _, expected := range "foo:" {
		r, ok := p.next()
		if !ok || r != expected {
			t.Fatalf("Failed to consume 'foo:'. Got %q, %v, want %q, true", r, ok, expected)
		}
	}

	// We should now be at the authority part.
	if !p.startsWith('/') {
		t.Fatalf("Expected to start with '/' after scheme, but got %q", p.asStr())
	}
	if pos := p.position(); pos != 4 {
		t.Errorf("Position after 'foo:' = %d, want 4", pos)
	}
	if s := p.asStr(); s != "//example.com:8042/over/there?name=ferret#nose" {
		t.Errorf("asStr() is incorrect. Got %q", s)
	}

	// Peek at the start of the authority.
	r, ok := p.peek()
	if !ok || r != '/' {
		t.Errorf("peek() should return '/', true but got %q, %v", r, ok)
	}

	// Reset and try another string.
	// RFC 3986, Section 5.4.2, abnormal example.
	p.reset("../../../g")
	if s := p.asStr(); s != "../../../g" {
		t.Errorf("asStr() after reset is incorrect. Got %q", s)
	}
	if pos := p.position(); pos != 0 {
		t.Errorf("Position after reset = %d, want 0", pos)
	}
}
