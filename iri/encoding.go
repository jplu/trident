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
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// percentEncode percent-encodes non-ASCII characters and disallowed lax ASCII characters in a string.
//
// Specification Reference:
// RFC 3987 (Section 3.1)
//
// Parameters:
//   - s: The input string containing characters to be encoded.
//   - b: The strings.Builder where the resulting encoded or unencoded characters are written.
func percentEncode(s string, b *strings.Builder) {
	for i := 0; i < len(s); {
		ru, size := utf8.DecodeRuneInString(s[i:])
		if ru <= unicode.MaxASCII {
			if isLaxASCII(ru) {
				// Spec Rule: RFC 3987 (Section 3.1)
				// Systems accepting IRIs may convert the printable characters in US-ASCII that are not allowed in URIs
				// by percent-encoding them.
				var buf [utf8.MaxRune]byte
				n := utf8.EncodeRune(buf[:], ru)
				for _, bb := range buf[:n] {
					fmt.Fprintf(b, "%%%02X", bb)
				}
			} else {
				b.WriteRune(ru)
			}
		} else {
			// Step 1: UTF-8 encoding
			// Convert the non-ASCII character to a sequence of one or more octets using UTF-8.
			var buf [utf8.MaxRune]byte
			n := utf8.EncodeRune(buf[:], ru)

			// Step 2: Hex conversion
			// Convert each octet to %HH, where HH is the uppercase hexadecimal notation.
			for _, bb := range buf[:n] {
				fmt.Fprintf(b, "%%%02X", bb)
			}
		}
		i += size
	}
}

// percentEncodeRune percent-encodes a single rune to the output buffer.
//
// Specification Reference:
// RFC 3986 (Section 2.1)
//
// Parameters:
//   - ru: The rune to be percent-encoded.
//   - output: The target outputBuffer where the percent-encoded string is written.
func percentEncodeRune(ru rune, output outputBuffer) {
	// Spec Rule: RFC 3986 (Section 2.3)
	// Unreserved characters should not be percent-encoded and are written directly.
	if isUnreserved(ru) {
		output.writeRune(ru)
		return
	}

	// Step 1: Convert rune to UTF-8 bytes and format as %HH percent-encoded values.
	var buf [utf8.MaxRune]byte
	n := utf8.EncodeRune(buf[:], ru)
	for _, bb := range buf[:n] {
		output.writeString(fmt.Sprintf("%%%02X", bb))
	}
}

// readURLCodepointOrEchar processes a single rune by validating or percent-encoding it.
//
// Specification Reference:
// RFC 3987 (Section 3.1)
//
// Parameters:
//   - r: The rune to be processed.
//   - valid: The predicate function to validate whether the rune is allowed in the current context.
//
// Returns:
//   - error: An error if the character is syntactically invalid.
func (p *iriParser) readURLCodepointOrEchar(r rune, valid func(rune) bool) error {
	// Step 1: Delegate percent-encoded sequence handling if '%' is encountered.
	if r == '%' {
		// Spec Rule: RFC 3986 (Section 2.1)
		// A percent-encoded octet must consist of the percent character '%' followed by two hexadecimal digits.
		if p.input.hasHexDigits() {
			return p.readEchar()
		}

		return &kindError{message: "Invalid percent-encoding sequence", char: '%'}
	}

	// Implementation Note: Unchecked Parsing
	// When unchecked parsing is enabled, verification is bypassed to optimize performance.
	if p.unchecked {
		p.output.writeRune(r)
		return nil
	}

	// Step 2: Validate the rune using the provided context-specific predicate.
	if valid(r) {
		p.output.writeRune(r)
		return nil
	}

	// Spec Rule: RFC 3987 (Section 3.1)
	// Systems accepting IRIs may leniently parse certain disallowed ASCII characters by percent-encoding them.
	if isLaxASCII(r) {
		percentEncodeRune(r, p.output)
		return nil
	}

	return &kindError{message: "Invalid IRI character", char: r}
}

// readEchar consumes and validates a percent-encoded character sequence from the input.
//
// Specification Reference:
// RFC 3986 (Section 2.1)
//
// Returns:
//   - error: An error if the percent-encoding format is incomplete or invalid.
func (p *iriParser) readEchar() error {
	// Step 1: Read the next two characters representing the hexadecimal digits.
	c1, ok1 := p.input.next()
	c2, ok2 := p.input.next()

	// Spec Rule: RFC 3986 (Section 2.1)
	// A percent-encoded octet must consist of the percent character '%' followed by two hexadecimal digits.
	if !ok1 || !ok2 || !isASCIIHexDigit(c1) || !isASCIIHexDigit(c2) {
		details := "%"
		if ok1 {
			details += string(c1)
		}
		if ok2 {
			details += string(c2)
		}
		return &kindError{message: "Invalid IRI percent encoding", details: details}
	}

	// Step 2: Write the fully validated percent-encoded sequence to the output.
	p.output.writeRune('%')
	p.output.writeRune(c1)
	p.output.writeRune(c2)
	return nil
}
