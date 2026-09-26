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

//nolint:testpackage // White-box test in the same package to access unexported functions.
package iri

import "testing"

// TestNewParserInput verifies that the parserInput is initialized correctly.
func TestNewParserInput(t *testing.T) {
	t.Parallel()
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
			if (p.reader == nil) != !tt.expectReader {
				t.Errorf(
					"newParserInput().reader is nil = %v, want %v",
					p.reader == nil,
					tt.expectReader,
				)
			}
			if p.originalString != tt.expectedStr {
				t.Errorf(
					"newParserInput().originalString = %q, want %q",
					p.originalString,
					tt.expectedStr,
				)
			}
			if pos := p.position(); pos != tt.expectedPos {
				t.Errorf("newParserInput().position() = %d, want %d", pos, tt.expectedPos)
			}
		})
	}
}

// TestParserInputNext tests the rune-by-rune consumption of the input string.
func TestParserInputNext(t *testing.T) {
	t.Parallel()
	rfc3987Example := "résumé"
	expectedRunes := []rune(rfc3987Example)

	tests := []struct {
		name          string
		input         string
		expectedRunes []rune
	}{
		{name: "RFC3987 Multi-byte string", input: rfc3987Example, expectedRunes: expectedRunes},
		{name: "Simple ASCII string", input: "abc", expectedRunes: []rune{'a', 'b', 'c'}},
		{name: "Empty string", input: "", expectedRunes: []rune{}},
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

// TestParserInputPeek verifies reading the next rune without advancing the cursor.
func TestParserInputPeek(t *testing.T) {
	t.Parallel()
	inputStr := "http:"
	p := newParserInput(inputStr)

	r, ok := p.peek()
	if !ok || r != 'h' {
		t.Errorf("peek() at start returned %q, %v, want 'h', true", r, ok)
	}
	if pos := p.position(); pos != 0 {
		t.Errorf("position() after peek() is %d, want 0", pos)
	}

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

	r, ok = p.peek()
	if ok || r != 0 {
		t.Errorf("peek() at end of string returned %q, %v, want 0, false", r, ok)
	}

	pEmpty := newParserInput("")
	r, ok = pEmpty.peek()
	if ok || r != 0 {
		t.Errorf("peek() on empty string returned %q, %v, want 0, false", r, ok)
	}
}

// TestParserInputStartsWith checks if the remaining input begins with a specific rune.
func TestParserInputStartsWith(t *testing.T) {
	t.Parallel()
	inputStr := "//example.com/path?query#fragment"
	p := newParserInput(inputStr)

	if !p.startsWith('/') {
		t.Error("startsWith('/') at start should be true")
	}
	if p.startsWith('e') {
		t.Error("startsWith('e') at start should be false")
	}

	p.next()
	p.next()

	if !p.startsWith('e') {
		t.Error("startsWith('e') after advancing should be true")
	}

	p.reset("")
	if p.startsWith('a') {
		t.Error("startsWith('a') on empty string should be false")
	}
}

// TestParserInputPosition validates the byte position tracking within the input.
func TestParserInputPosition(t *testing.T) {
	t.Parallel()
	inputStr := "fóo"
	p := newParserInput(inputStr)

	if pos := p.position(); pos != 0 {
		t.Errorf("Initial position = %d, want 0", pos)
	}
	p.next()
	if pos := p.position(); pos != 1 {
		t.Errorf("Position after 'f' = %d, want 1", pos)
	}
	p.next()
	if pos := p.position(); pos != 3 {
		t.Errorf("Position after 'ó' = %d, want 3", pos)
	}
	p.next()
	if pos := p.position(); pos != 4 {
		t.Errorf("Position after 'o' = %d, want 4", pos)
	}
	p.next()
	if pos := p.position(); pos != 4 {
		t.Errorf("Position at EOF = %d, want 4", pos)
	}
}

// TestParserInputAsStr ensures the unread portion of the string is returned correctly.
func TestParserInputAsStr(t *testing.T) {
	t.Parallel()
	inputStr := "//a/b/c"
	p := newParserInput(inputStr)

	if s := p.asStr(); s != "//a/b/c" {
		t.Errorf("Initial asStr() = %q, want %q", s, "//a/b/c")
	}
	p.next()
	p.next()
	if s := p.asStr(); s != "a/b/c" {
		t.Errorf("asStr() after consuming '//' = %q, want %q", s, "a/b/c")
	}
	p.next()
	p.next()
	p.next()
	p.next()
	if s := p.asStr(); s != "c" {
		t.Errorf("asStr() after consuming 'a/b/' = %q, want %q", s, "c")
	}
	p.next()
	if s := p.asStr(); s != "" {
		t.Errorf("asStr() at end of input = %q, want %q", s, "")
	}
}

// TestParserInputReset verifies that the input can be re-initialized with a new string.
func TestParserInputReset(t *testing.T) {
	t.Parallel()
	initialRef := "g:h"
	newRef := "./g"

	p := newParserInput(initialRef)
	p.next()
	if pos := p.position(); pos != 1 {
		t.Fatalf("Position before reset was %d, expected 1", pos)
	}
	if s := p.asStr(); s != ":h" {
		t.Fatalf("asStr() before reset was %q, expected %q", s, ":h")
	}

	p.reset(newRef)

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

// TestParserInputCombinedOperations performs a sequence of operations to ensure they interact correctly.
func TestParserInputCombinedOperations(t *testing.T) {
	t.Parallel()
	inputStr := "foo://example.com:8042/over/there?name=ferret#nose"
	p := newParserInput(inputStr)

	for _, expected := range "foo:" {
		r, ok := p.next()
		if !ok || r != expected {
			t.Fatalf("Failed to consume 'foo:'. Got %q, %v, want %q, true", r, ok, expected)
		}
	}

	if !p.startsWith('/') {
		t.Fatalf("Expected to start with '/' after scheme, but got %q", p.asStr())
	}
	if pos := p.position(); pos != 4 {
		t.Errorf("Position after 'foo:' = %d, want 4", pos)
	}
	if s := p.asStr(); s != "//example.com:8042/over/there?name=ferret#nose" {
		t.Errorf("asStr() is incorrect. Got %q", s)
	}

	r, ok := p.peek()
	if !ok || r != '/' {
		t.Errorf("peek() should return '/', true but got %q, %v", r, ok)
	}

	p.reset("../../../g")
	if s := p.asStr(); s != "../../../g" {
		t.Errorf("asStr() after reset is incorrect. Got %q", s)
	}
	if pos := p.position(); pos != 0 {
		t.Errorf("Position after reset = %d, want 0", pos)
	}
}

// TestParserInputHasHexDigits checks if the upcoming input represents a valid hexadecimal escape pair.
func TestParserInputHasHexDigits(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "Valid hex sequence", input: "%20", expected: true},
		{name: "Invalid hex sequence (second char non-hex)", input: "%2g", expected: false},
		{name: "Invalid hex sequence (first char non-hex)", input: "%g0", expected: false},
		{name: "Incomplete hex sequence (one char)", input: "%2", expected: false},
		{name: "Incomplete hex sequence (zero char)", input: "%", expected: false},
		{name: "No sequence", input: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var input string
			if tt.input != "" {
				input = tt.input
			} else {
				input = "foo"
			}
			p := newParserInput(input)
			if tt.input != "" {
				p.next()
			}
			if got := p.hasHexDigits(); got != tt.expected {
				t.Errorf("hasHexDigits() on %q = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
