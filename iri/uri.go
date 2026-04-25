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
	"encoding/hex"
	"strings"

	// TODO: At some point implement my own IDNA2003 module (RFC 3490).
	"golang.org/x/net/idna"
	// TODO: At some point implement my own NFC module.
	"golang.org/x/text/unicode/norm"
)

// ToURI converts the IRI reference to a URI reference string, strictly following
// RFC 3987, Section 3.1. It normalizes all components to NFC, percent-encodes
// any non-ASCII characters using their UTF-8 representation, and applies IDNA
// (ToASCII) to the host component to ensure the resulting URI is resolvable in DNS.
func (r *Ref) ToURI() string {
	var builder strings.Builder
	builder.Grow(len(r.iri))

	scheme, hasScheme := r.Scheme()
	authority, hasAuthority := r.Authority()
	path := r.Path()
	query, hasQuery := r.Query()
	fragment, hasFragment := r.Fragment()

	if hasScheme {
		builder.WriteString(scheme)
		builder.WriteRune(':')
	}

	if hasAuthority {
		builder.WriteString("//")
		userinfo, host, port := splitAuthority(authority)

		// Per RFC 3987, Section 3.1, Step 1, components must be in NFC
		// before percent-encoding.
		normalizedUserinfo := norm.NFC.String(userinfo)
		percentEncode(normalizedUserinfo, &builder)
		if userinfo != "" {
			builder.WriteRune('@')
		}

		// Normalize host to NFC before applying IDNA.
		normalizedHost := norm.NFC.String(host)

		// Apply IDNA ToASCII to the host for DNS resolvability.
		asciiHost, err := idna.ToASCII(normalizedHost)
		if err == nil {
			builder.WriteString(asciiHost)
		}

		if port != "" {
			builder.WriteRune(':')
			builder.WriteString(port)
		}
	}

	// Normalize path, query, and fragment to NFC before percent-encoding.
	percentEncode(norm.NFC.String(path), &builder)
	if hasQuery {
		builder.WriteRune('?')
		percentEncode(norm.NFC.String(query), &builder)
	}
	if hasFragment {
		builder.WriteRune('#')
		percentEncode(norm.NFC.String(fragment), &builder)
	}

	return builder.String()
}

// ParseURIToRef converts a URI string into an IRI reference by decoding
// percent-encoded octets that form valid UTF-8 sequences. This is the
// reverse of the ToURI method and follows RFC 3987, Section 3.2.
//
// It cautiously decodes only valid sequences and re-validates the final
// string to ensure it forms a syntactically correct IRI reference. Any
// percent-encoded octets that do not form a valid UTF-8 sequence or that
// represent characters not permitted in IRIs (such as bidi control characters)
// are left in their percent-encoded form.
func ParseURIToRef(s string) (*Ref, error) {
	var builder strings.Builder
	builder.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] != '%' {
			builder.WriteByte(s[i])
			i++
			continue
		}

		start := i
		var decodedBytes []byte
		// Find a contiguous block of percent-encoded octets.
		for i < len(s) && s[i] == '%' {
			if i+2 >= len(s) || !isASCIIHexDigit(rune(s[i+1])) || !isASCIIHexDigit(rune(s[i+2])) {
				// Incomplete or invalid encoding, stop processing this block.
				break
			}
			b, _ := hex.DecodeString(s[i+1 : i+3])
			decodedBytes = append(decodedBytes, b[0])
			i += 3
		}

		// If the inner loop didn't advance, we found an invalid/incomplete sequence.
		if i == start {
			// Write the original '%' and advance past it to prevent an infinite loop.
			builder.WriteByte(s[start])
			i++
			continue
		}

		if validateDecodedBytes(decodedBytes) {
			builder.Write(decodedBytes)
		} else {
			// Not valid UTF-8 or contains forbidden characters, so keep original encoding.
			builder.WriteString(s[start:i])
		}
	}

	// The decoded string must be re-parsed to ensure it is a valid IRI.
	// ParseNormalizedRef is used here because URI-to-IRI conversion
	// implies a canonical representation is desired.
	return ParseNormalizedRef(builder.String())
}
