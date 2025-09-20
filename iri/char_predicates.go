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
	"strings"
	"unicode"
)

// isASCIILetter checks if a rune is an ASCII letter.
func isASCIILetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

// isASCIIDigit checks if a rune is an ASCII digit.
func isASCIIDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

// isASCIIHexDigit checks if a rune is an ASCII hexadecimal digit.
func isASCIIHexDigit(r rune) bool {
	return isASCIIDigit(r) || ('a' <= unicode.ToLower(r) && unicode.ToLower(r) <= 'f')
}

// isLaxASCII checks if a character is one of the US-ASCII characters
// that are not allowed in URIs but may be accepted and percent-encoded
// by a lenient IRI parser, as per RFC 3987, Section 3.1.
func isLaxASCII(c rune) bool {
	// Note: The set is "<", ">", '"', space, "{", "}", "|", "\", "^", and "`".
	// The characters "#", "%", "[", and "]" are explicitly excluded and must not be converted.
	return strings.ContainsRune("<>\" {}|\\^`", c)
}

// isForbiddenBidiFormatting checks for bidirectional formatting characters that are forbidden in IRIs.
func isForbiddenBidiFormatting(c rune) bool {
	// RFC 3987, Section 4.1: "IRIs MUST NOT contain bidirectional formatting characters"
	// These characters are LRM (U+200E), RLM (U+200F), and LRE, RLE, PDF, LRO, RLO (U+202A to U+202E).
	return (c >= '\u202A' && c <= '\u202E') || c == '\u200E' || c == '\u200F'
}

// isIUnreservedOrSubDelims checks if a character is in the iunreserved or
// sub-delims sets as defined by RFC 3987. This includes unreserved US-ASCII
// characters and additional Unicode ranges.
func isIUnreservedOrSubDelims(c rune) bool {
	if isForbiddenBidiFormatting(c) {
		return false
	}

	if isUnreservedOrSubDelims(c) {
		return true
	}

	switch {
	case c >= '\u00A0' && c <= '\uD7FF',
		c >= '\uF900' && c <= '\uFDCF',
		c >= '\uFDF0' && c <= '\uFFEF',
		c >= 0x10000 && c <= 0x1FFFD,
		c >= 0x20000 && c <= 0x2FFFD,
		c >= 0x30000 && c <= 0x3FFFD,
		c >= 0x40000 && c <= 0x4FFFD,
		c >= 0x50000 && c <= 0x5FFFD,
		c >= 0x60000 && c <= 0x6FFFD,
		c >= 0x70000 && c <= 0x7FFFD,
		c >= 0x80000 && c <= 0x8FFFD,
		c >= 0x90000 && c <= 0x9FFFD,
		c >= 0xA0000 && c <= 0xAFFFD,
		c >= 0xB0000 && c <= 0xBFFFD,
		c >= 0xC0000 && c <= 0xCFFFD,
		c >= 0xD0000 && c <= 0xDFFFD,
		c >= 0xE1000 && c <= 0xEFFFD:
		return true
	}
	return false
}

// isUnreservedOrSubDelims checks if a character is in the unreserved or
// sub-delims sets as defined by RFC 3986 (US-ASCII only).
func isUnreservedOrSubDelims(c rune) bool {
	return isASCIILetter(c) || isASCIIDigit(c) || strings.ContainsRune("!$&'()*+,-.;=_~", c)
}

// isUnreserved checks if a character is in the unreserved set as defined by RFC 3986.
func isUnreserved(c rune) bool {
	return isASCIILetter(c) || isASCIIDigit(c) || c == '-' || c == '.' || c == '_' || c == '~'
}
