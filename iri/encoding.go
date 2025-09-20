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
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// percentEncode is a helper that percent-encodes non-ASCII characters in a string.
// It is used by Ref.ToURI() to convert an IRI to a URI.
func percentEncode(s string, b *strings.Builder) {
	for _, ru := range s {
		if ru <= unicode.MaxASCII {
			b.WriteRune(ru)
		} else {
			var buf [utf8.MaxRune]byte
			n := utf8.EncodeRune(buf[:], ru)
			for i := range n {
				fmt.Fprintf(b, "%%%02X", buf[i])
			}
		}
	}
}

// percentEncodeRune percent-encodes a single rune to the output buffer if it is not an
// unreserved character.
func percentEncodeRune(ru rune, output outputBuffer) {
	if isUnreserved(ru) {
		output.writeRune(ru)
		return
	}
	var buf [utf8.MaxRune]byte
	n := utf8.EncodeRune(buf[:], ru)
	for i := range n {
		output.writeString(fmt.Sprintf("%%%02X", buf[i]))
	}
}

// readURLCodepointOrEchar processes a single rune. If it's a '%' it handles
// percent-encoding. Otherwise, it validates the rune against the provided
// function and writes it to the output. It implements lenient parsing for
// certain disallowed ASCII characters by percent-encoding them.
func (p *iriParser) readURLCodepointOrEchar(r rune, valid func(rune) bool) error {
	if r == '%' {
		return p.readEchar()
	}

	if p.unchecked {
		p.output.writeRune(r)
		return nil
	}

	if valid(r) {
		p.output.writeRune(r)
		return nil
	}

	// Leniently parse certain disallowed ASCII characters by percent-encoding them.
	// This is an optional ("MAY") behavior from RFC 3987, Section 3.1.
	if isLaxASCII(r) {
		percentEncodeRune(r, p.output)
		return nil
	}

	return &kindError{message: "Invalid IRI character", char: r}
}

// readEchar handles a percent-encoded character (e.g., "%20").
func (p *iriParser) readEchar() error {
	c1, ok1 := p.input.next()
	c2, ok2 := p.input.next()
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
	p.output.writeRune('%')
	p.output.writeRune(c1)
	p.output.writeRune(c2)
	return nil
}

// normalizePercentEncoding decodes any percent-encoded octet that corresponds to an
// unreserved character, as per RFC 3986 Section 6.2.2.2.
func normalizePercentEncoding(s string) string {
	var b bytes.Buffer
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] == '%' && i+2 < len(s) && isASCIIHexDigit(rune(s[i+1])) && isASCIIHexDigit(rune(s[i+2])) {
			decoded, err := hex.DecodeString(s[i+1 : i+3])
			if err == nil {
				// Check if the decoded character is unreserved.
				c := rune(decoded[0])
				if isUnreserved(c) {
					b.WriteRune(c)
					i += 3
					continue
				}
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// validateDecodedBytes checks if a byte slice is valid UTF-8 and contains only allowed characters.
// Per RFC 3987, Section 4.1, bidi formatting characters are forbidden.
func validateDecodedBytes(decodedBytes []byte) bool {
	if !utf8.Valid(decodedBytes) {
		return false
	}
	decodedStr := string(decodedBytes)
	for _, r := range decodedStr {
		if isForbiddenBidiFormatting(r) {
			return false
		}
	}
	return true
}
