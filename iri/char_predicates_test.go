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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package iri

import "testing"

// TestIsASCIILetter tests the isASCIILetter function for compliance with RFC 3986, Appendix A (ALPHA).
func TestIsASCIILetter(t *testing.T) {
	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{"lowercase 'a'", 'a', true},
		{"lowercase 'z'", 'z', true},
		{"lowercase middle", 'm', true},
		{"uppercase 'A'", 'A', true},
		{"uppercase 'Z'", 'Z', true},
		{"uppercase middle", 'M', true},
		{"digit '0'", '0', false},
		{"digit '9'", '9', false},
		{"symbol '-'", '-', false},
		{"space", ' ', false},
		{"non-ascii", 'é', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isASCIILetter(tt.input); got != tt.want {
				t.Errorf("isASCIILetter('%c') = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestIsASCIIDigit tests the isASCIIDigit function for compliance with RFC 3986, Appendix A (DIGIT).
func TestIsASCIIDigit(t *testing.T) {
	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{"digit '0'", '0', true},
		{"digit '9'", '9', true},
		{"digit '5'", '5', true},
		{"lowercase 'a'", 'a', false},
		{"uppercase 'Z'", 'Z', false},
		{"symbol '-'", '-', false},
		{"space", ' ', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isASCIIDigit(tt.input); got != tt.want {
				t.Errorf("isASCIIDigit('%c') = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestIsASCIIHexDigit tests the isASCIIHexDigit function for compliance with RFC 3986, Appendix A (HEXDIG).
func TestIsASCIIHexDigit(t *testing.T) {
	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{"digit '0'", '0', true},
		{"digit '9'", '9', true},
		{"lowercase 'a'", 'a', true},
		{"lowercase 'f'", 'f', true},
		{"uppercase 'A'", 'A', true},
		{"uppercase 'F'", 'F', true},
		{"lowercase 'g'", 'g', false},
		{"uppercase 'G'", 'G', false},
		{"symbol '-'", '-', false},
		{"space", ' ', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isASCIIHexDigit(tt.input); got != tt.want {
				t.Errorf("isASCIIHexDigit('%c') = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestIsLaxASCII tests the isLaxASCII function based on RFC 3987, Section 3.1.
func TestIsLaxASCII(t *testing.T) {
	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{"less than", '<', true},
		{"greater than", '>', true},
		{"double quote", '"', true},
		{"space", ' ', true},
		{"open brace", '{', true},
		{"close brace", '}', true},
		{"pipe", '|', true},
		{"backslash", '\\', true},
		{"caret", '^', true},
		{"backtick", '`', true},
		{"hash (gen-delim)", '#', false},
		{"percent", '%', false},
		{"open bracket (gen-delim)", '[', false},
		{"close bracket (gen-delim)", ']', false},
		{"letter 'a'", 'a', false},
		{"digit '1'", '1', false},
		{"slash (gen-delim)", '/', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLaxASCII(tt.input); got != tt.want {
				t.Errorf("isLaxASCII('%c') = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestIsForbiddenBidiFormatting tests isForbiddenBidiFormatting against RFC 3987, Section 4.1.
func TestIsForbiddenBidiFormatting(t *testing.T) {
	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{"LRM U+200E", '\u200E', true},
		{"RLM U+200F", '\u200F', true},
		{"LRE U+202A", '\u202A', true},
		{"RLE U+202B", '\u202B', true},
		{"PDF U+202C", '\u202C', true},
		{"LRO U+202D", '\u202D', true},
		{"RLO U+202E", '\u202E', true},
		{"before LRM", '\u200D', false},
		{"before LRE", '\u2029', false},
		{"after RLO", '\u202F', false},
		{"letter 'a'", 'a', false},
		{"digit '1'", '1', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isForbiddenBidiFormatting(tt.input); got != tt.want {
				t.Errorf("isForbiddenBidiFormatting('%c') = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestIsUnreservedOrSubDelims tests isUnreservedOrSubDelims against RFC 3986, Appendix A.
func TestIsUnreservedOrSubDelims(t *testing.T) {
	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{"letter 'a'", 'a', true},
		{"letter 'Z'", 'Z', true},
		{"digit '0'", '0', true},
		{"digit '9'", '9', true},
		{"unreserved '-'", '-', true},
		{"unreserved '.'", '.', true},
		{"unreserved '_'", '_', true},
		{"unreserved '~'", '~', true},
		{"sub-delim '!'", '!', true},
		{"sub-delim '$'", '$', true},
		{"sub-delim '&'", '&', true},
		{"sub-delim '''", '\'', true},
		{"sub-delim '('", '(', true},
		{"sub-delim ')'", ')', true},
		{"sub-delim '*'", '*', true},
		{"sub-delim '+'", '+', true},
		{"sub-delim ','", ',', true},
		{"sub-delim ';'", ';', true},
		{"sub-delim '='", '=', true},
		{"gen-delim ':'", ':', false},
		{"gen-delim '/'", '/', false},
		{"gen-delim '?'", '?', false},
		{"gen-delim '#'", '#', false},
		{"gen-delim '['", '[', false},
		{"gen-delim ']'", ']', false},
		{"gen-delim '@'", '@', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUnreservedOrSubDelims(tt.input); got != tt.want {
				t.Errorf("isUnreservedOrSubDelims('%c') = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestIsIUnreservedOrSubDelims tests isIUnreservedOrSubDelims against RFC 3987, Section 2.2.
func TestIsIUnreservedOrSubDelims(t *testing.T) {
	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{"forbidden bidi LRE U+202A", '\u202A', false},
		{"unreserved letter 'a'", 'a', true},
		{"sub-delim '!'", '!', true},
		{"ucschar start U+00A0", '\u00A0', true},
		{"ucschar middle U+0ABC", '\u0ABC', true},
		{"ucschar end U+D7FF", '\uD7FF', true},
		{"ucschar start U+F900", '\uF900', true},
		{"ucschar middle U+FAAB", '\uFAAB', true},
		{"ucschar end U+FDCF", '\uFDCF', true},
		{"ucschar start U+FDF0", '\uFDF0', true},
		{"ucschar middle U+FEAB", '\uFEAB', true},
		{"ucschar end U+FFEF", '\uFFEF', true},
		{"ucschar start U+10000", '\U00010000', true},
		{"ucschar middle U+1ABCD", '\U0001ABCD', true},
		{"ucschar end U+1FFFD", '\U0001FFFD', true},
		{"ucschar start U+20000", '\U00020000', true},
		{"ucschar end U+2FFFD", '\U0002FFFD', true},
		{"ucschar start U+30000", '\U00030000', true},
		{"ucschar end U+3FFFD", '\U0003FFFD', true},
		{"ucschar start U+40000", '\U00040000', true},
		{"ucschar end U+4FFFD", '\U0004FFFD', true},
		{"ucschar start U+50000", '\U00050000', true},
		{"ucschar end U+5FFFD", '\U0005FFFD', true},
		{"ucschar start U+60000", '\U00060000', true},
		{"ucschar end U+6FFFD", '\U0006FFFD', true},
		{"ucschar start U+70000", '\U00070000', true},
		{"ucschar end U+7FFFD", '\U0007FFFD', true},
		{"ucschar start U+80000", '\U00080000', true},
		{"ucschar end U+8FFFD", '\U0008FFFD', true},
		{"ucschar start U+90000", '\U00090000', true},
		{"ucschar end U+9FFFD", '\U0009FFFD', true},
		{"ucschar start U+A0000", '\U000A0000', true},
		{"ucschar end U+AFFFD", '\U000AFFFD', true},
		{"ucschar start U+B0000", '\U000B0000', true},
		{"ucschar end U+BFFFD", '\U000BFFFD', true},
		{"ucschar start U+C0000", '\U000C0000', true},
		{"ucschar end U+CFFFD", '\U000CFFFD', true},
		{"ucschar start U+D0000", '\U000D0000', true},
		{"ucschar end U+DFFFD", '\U000DFFFD', true},
		{"ucschar start U+E1000", '\U000E1000', true},
		{"ucschar end U+EFFFD", '\U000EFFFD', true},
		{"boundary before ucschar U+009F", '\u009F', false},
		{"surrogate pair range U+D800", rune(0xD800), false},
		{"between ranges U+F8FF", '\uF8FF', false},
		{"between ranges U+FDDF", '\uFDDF', false},
		{"unassigned U+FFFE", '\uFFFE', false},
		{"plane boundary U+1FFFE", '\U0001FFFE', false},
		{"gen-delim '/'", '/', false},
		{"gen-delim '?'", '?', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIUnreservedOrSubDelims(tt.input); got != tt.want {
				t.Errorf("isIUnreservedOrSubDelims('%c') = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestIsUnreserved tests the isUnreserved function against RFC 3986, Appendix A.
func TestIsUnreserved(t *testing.T) {
	tests := []struct {
		name  string
		input rune
		want  bool
	}{
		{"lowercase letter 'a'", 'a', true},
		{"uppercase letter 'Z'", 'Z', true},
		{"digit '0'", '0', true},
		{"digit '9'", '9', true},
		{"hyphen", '-', true},
		{"period", '.', true},
		{"underscore", '_', true},
		{"tilde", '~', true},
		{"sub-delim '!'", '!', false},
		{"sub-delim '='", '=', false},
		{"gen-delim ':'", ':', false},
		{"gen-delim '/'", '/', false},
		{"space", ' ', false},
		{"non-ascii 'é'", 'é', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUnreserved(tt.input); got != tt.want {
				t.Errorf("isUnreserved('%c') = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
