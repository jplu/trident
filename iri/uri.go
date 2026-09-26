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
	"unicode/utf8"

	// TODO: At some point implement my own IDNA2003 module (RFC 3490).
	"golang.org/x/net/idna"
)

// ToURI converts the IRI reference to a URI reference string, strictly following RFC 3987 mapping rules.
//
// Specification Reference:
// RFC 3987 (Section 3.1)
//
// Returns:
//   - string: The corresponding RFC 3986-compliant URI reference string.
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
		userinfo, host, port, _ := splitAuthority(authority)

		// Step 1: Process and Encode Userinfo Component
		// Write the un-normalized userinfo component to the output buffer after percent-encoding.
		percentEncode(userinfo, &builder)
		if userinfo != "" {
			builder.WriteRune('@')
		}

		// Step 2: Translate Host via IDNA ToASCII
		// Convert the host component using ToASCII. Internal normalization is handled directly by the IDNA library.
		asciiHost, err := idna.ToASCII(host)
		if err == nil {
			builder.WriteString(asciiHost)
		} else {
			builder.WriteString(host)
		}

		if port != "" {
			builder.WriteRune(':')
			builder.WriteString(port)
		}
	}

	// Step 3: Process Remaining Components
	// Write path, query, and fragment components directly to the output buffer without applying NFC normalization.
	percentEncode(path, &builder)
	if hasQuery {
		builder.WriteRune('?')
		percentEncode(query, &builder)
	}
	if hasFragment {
		builder.WriteRune('#')
		percentEncode(fragment, &builder)
	}

	return builder.String()
}

// decodeNonASCIIBlock processes a block of contiguous non-ASCII percent-encoded octets.
//
// Specification Reference:
// RFC 3987 (Section 3.2)
//
// Parameters:
//   - s: The component string containing potentially percent-encoded characters.
//   - start: The starting index of the percent-encoded block.
//   - isValidRune: The validator function checking if a decoded rune is allowed in the component's context.
//
// Returns:
//   - string: The decoded representation or the original percent-encoded block.
//   - int: The updated index within the string after consuming the block.
func decodeNonASCIIBlock(s string, start int, isValidRune func(rune) bool) (string, int) {
	const maxASCII = 127
	var decodedBytes []byte
	i := start

	// Step 1: Scan Contiguous Non-ASCII Percent-Encoded Blocks
	// Aggregate consecutive non-ASCII percent-encoded sequences (bytes >= 128)
	// to evaluate them together as potential multi-byte UTF-8 characters.
	for i < len(s) && s[i] == '%' {
		if i+2 >= len(s) || !isASCIIHexDigit(rune(s[i+1])) || !isASCIIHexDigit(rune(s[i+2])) {
			break
		}
		octet, _ := hex.DecodeString(s[i+1 : i+3])
		if octet[0] <= maxASCII {
			break
		}
		decodedBytes = append(decodedBytes, octet[0])
		i += 3
	}

	// Step 2: Decode and Validate Non-ASCII Octets
	// Evaluate the decoded bytes to decide whether they can be written as raw IRI characters
	// or must remain percent-encoded under strict syntax/Bidi constraints.
	if utf8.Valid(decodedBytes) {
		decodedStr := string(decodedBytes)
		allValid := true
		for _, r := range decodedStr {
			if isForbiddenBidiFormatting(r) || !isValidRune(r) {
				allValid = false
				break
			}
		}
		if allValid {
			return decodedStr, i
		}
	}
	return s[start:i], i
}

// decodeComponent decodes valid percent-encoded octets in a component string according to context validation.
//
// Specification Reference:
// RFC 3987 (Section 3.2)
//
// Parameters:
//   - s: The component string containing potentially percent-encoded characters.
//   - isValidRune: The validator function checking if a decoded rune is allowed in the component's context.
//
// Returns:
//   - string: The decoded component string.
func decodeComponent(s string, isValidRune func(rune) bool) string {
	const maxASCII = 127
	var builder strings.Builder
	builder.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] != '%' {
			builder.WriteByte(s[i])
			i++
			continue
		}

		if i+2 >= len(s) || !isASCIIHexDigit(rune(s[i+1])) || !isASCIIHexDigit(rune(s[i+2])) {
			builder.WriteByte(s[i])
			i++
			continue
		}

		b, _ := hex.DecodeString(s[i+1 : i+3])
		if b[0] <= maxASCII {
			// Spec Rule: RFC 3987 (Section 3.2)
			// Percent-encoded octets corresponding to "%", characters in "reserved",
			// and characters in US-ASCII not allowed in URIs must remain percent-encoded.
			// Thus, only unreserved US-ASCII characters may be decoded.
			if isUnreserved(rune(b[0])) {
				builder.WriteByte(b[0])
			} else {
				builder.WriteString(s[i : i+3])
			}
			i += 3
		} else {
			var decodedStr string
			decodedStr, i = decodeNonASCIIBlock(s, i, isValidRune)
			builder.WriteString(decodedStr)
		}
	}
	return builder.String()
}

// ParseURIToRef converts a URI string into an IRI reference by decoding percent-encoded octets.
//
// Specification Reference:
// RFC 3987 (Section 3.2)
//
// Parameters:
//   - s: The URI reference string to decode and parse.
//
// Returns:
//   - *Ref: A pointer to the converted Ref representing the decoded IRI.
//   - error: An error condition under parsing or syntax validation failure.
func ParseURIToRef(s string) (*Ref, error) {
	// Step 1: Sub-parse the Input URI
	// Since any valid URI is syntactically a valid IRI, parse it first to determine component positions safely.
	uriRef, err := ParseRef(s)
	if err != nil {
		return nil, err
	}

	scheme, hasScheme := uriRef.Scheme()
	authority, hasAuthority := uriRef.Authority()
	path := uriRef.Path()
	query, hasQuery := uriRef.Query()
	fragment, hasFragment := uriRef.Fragment()

	var decodedUserinfo, decodedHost, port string
	if hasAuthority {
		var userinfo, host string
		userinfo, host, port, _ = splitAuthority(authority)

		// Step 2: Decode Userinfo Subcomponent
		// Allowed characters are iunreserved, sub-delims, and colon.
		decodedUserinfo = decodeComponent(userinfo, func(r rune) bool {
			return isIUnreservedOrSubDelims(r) || r == ':'
		})

		// Step 3: Decode Host Subcomponent
		// Allowed characters for host names are limited to iunreserved and sub-delims.
		decodedHost = decodeComponent(host, isIUnreservedOrSubDelims)
	}

	// Step 4: Decode Path Component
	// Allowed path characters as defined by syntax rules.
	decodedPath := decodeComponent(path, isPathChar)

	// Step 5: Decode Query Component
	// Query allows unreserved, sub-delims, slash, question mark, and private-use characters.
	var decodedQuery string
	if hasQuery {
		decodedQuery = decodeComponent(query, isQueryChar)
	}

	// Step 6: Decode Fragment Component
	// Fragment allows iunreserved, sub-delims, colon, at-sign, slash, and question mark.
	var decodedFragment string
	if hasFragment {
		decodedFragment = decodeComponent(fragment, func(r rune) bool {
			return isIUnreservedOrSubDelims(r) || r == ':' || r == '@' || r == '/' || r == '?'
		})
	}

	// Step 7: Recompose and Normalize the Decoded IRI Reference
	// Construct the final normalized string using the component-aware safe-decoded segments.
	var b strings.Builder
	b.Grow(len(s))
	if hasScheme {
		b.WriteString(scheme)
		b.WriteRune(':')
	}
	if hasAuthority {
		b.WriteString("//")
		if decodedUserinfo != "" {
			b.WriteString(decodedUserinfo)
			b.WriteRune('@')
		}
		b.WriteString(decodedHost)
		if port != "" {
			b.WriteRune(':')
			b.WriteString(port)
		}
	}
	b.WriteString(decodedPath)
	if hasQuery {
		b.WriteRune('?')
		b.WriteString(decodedQuery)
	}
	if hasFragment {
		b.WriteRune('#')
		b.WriteString(decodedFragment)
	}

	return ParseNormalizedRef(b.String())
}
