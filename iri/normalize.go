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
	"bytes"
	"encoding/hex"
	"strings"

	// TODO: At some point implement my own IDNA2003 module (RFC 3490).
	"golang.org/x/net/idna"
	// TODO: At some point implement my own NFC module.
	"golang.org/x/text/unicode/norm"
)

// Normalize applies syntax-based normalization and Unicode Form C normalization to the IRI reference.
//
// Specification Reference:
// RFC 3987 (Section 5.3.2) and RFC 3986 (Section 6.2.2)
//
// Returns:
//   - *Ref: The normalized IRI reference pointer, or the original Ref pointer if no modifications were needed.
func (r *Ref) Normalize() *Ref {
	if r.iri == "" {
		return &Ref{}
	}

	scheme, hasScheme := r.Scheme()
	authority, hasAuthority := r.Authority()
	path := r.Path()
	query, hasQuery := r.Query()
	fragment, hasFragment := r.Fragment()

	// Step 1: Case Normalization
	// Convert the scheme component to lowercase if present and delegate host normalization.
	if hasScheme {
		scheme = strings.ToLower(scheme)
	}
	var userinfo, host, port string
	if hasAuthority {
		userinfo, host, port, _ = splitAuthority(authority)
		host, port = normalizeHostAndPort(host, port, scheme)
	}

	// Step 2: Percent-Encoding Normalization
	// Decode any percent-encoded octet sequences that represent unreserved characters.
	userinfo = normalizePercentEncoding(userinfo)
	host = normalizePercentEncoding(host)
	path = normalizePercentEncoding(path)
	query = normalizePercentEncoding(query)
	fragment = normalizePercentEncoding(fragment)

	// Step 3: Path Segment Normalization
	// Remove dot-segments from the path component.
	path = removeDotSegments(path)

	// Step 4: Scheme-Based Path Normalization
	// Normalize empty paths to a single slash when an authority component is present.
	if hasAuthority && path == "" {
		path = "/"
	}

	// Step 5: Recompose and Re-parse Components
	// Recompose the normalized components into an IRI string and apply Unicode Normalization Form C (NFC).
	recomposedStr := recomposeNormalizedIRI(
		scheme, hasScheme,
		userinfo, host, port, hasAuthority,
		path,
		query, hasQuery,
		fragment, hasFragment,
	)

	normalizedStr := norm.NFC.String(recomposedStr)

	if normalizedStr == r.iri {
		return r
	}
	// Implementation Note: Safe Re-parsing of Normalized String
	// An error is not expected here as we are building from valid components. We use ParseRef since normalizedStr is
	// guaranteed to be in NFC.
	newRef, _ := ParseRef(normalizedStr)
	return newRef
}

// recomposeNormalizedIRI recomposes a normalized IRI string from its individual components.
//
// Specification Reference:
// RFC 3987 (Section 3.1) and RFC 3986 (Section 5.3)
//
// Parameters:
//   - scheme: The scheme component string.
//   - hasScheme: A boolean indicating whether the scheme is present.
//   - userinfo: The userinfo subcomponent string.
//   - host: The host subcomponent string.
//   - port: The port subcomponent string.
//   - hasAuthority: A boolean indicating whether the authority component is present.
//   - path: The path component string.
//   - query: The query component string.
//   - hasQuery: A boolean indicating whether the query component is present.
//   - fragment: The fragment component string.
//   - hasFragment: A boolean indicating whether the fragment component is present.
//
// Returns:
//   - string: The recomposed IRI string.
func recomposeNormalizedIRI(
	scheme string, hasScheme bool,
	userinfo, host, port string, hasAuthority bool,
	path string,
	query string, hasQuery bool,
	fragment string, hasFragment bool,
) string {
	var b strings.Builder
	if hasScheme {
		b.WriteString(scheme)
		b.WriteRune(':')
	}
	if hasAuthority {
		b.WriteString("//")
		if userinfo != "" {
			b.WriteString(userinfo)
			b.WriteRune('@')
		}
		b.WriteString(host)
		if port != "" {
			b.WriteRune(':')
			b.WriteString(port)
		}
	}
	b.WriteString(path)
	if hasQuery {
		b.WriteRune('?')
		b.WriteString(query)
	}
	if hasFragment {
		b.WriteRune('#')
		b.WriteString(fragment)
	}
	return b.String()
}

// normalizeHostAndPort normalizes the host and port components by applying case, IDNA, and scheme-based port
// normalization.
//
// Specification Reference:
// RFC 3987 (Section 5.3.2.1), RFC 3490 (Section 4), and RFC 3986 (Section 6.2.3)
//
// Parameters:
//   - host: The host component string to be normalized.
//   - port: The port component string to be normalized.
//   - scheme: The scheme component string used to determine default ports.
//
// Returns:
//   - string: The normalized host string.
//   - string: The normalized port string, or empty if it matches the scheme's default port.
func normalizeHostAndPort(host, port, scheme string) (string, string) {
	// Step 1: Case Normalization
	// Convert the host component to lowercase since host names are case-insensitive.
	normalizedHost := strings.ToLower(host)

	// Step 2: IDNA Normalization
	// Apply IDNA ToASCII and ToUnicode mappings on the host labels.
	if !strings.HasPrefix(normalizedHost, "[") {
		unicodeHost := normalizedHost
		// Implementation Note: Canonical Unicode Form Conversion
		// Retrieve the canonical Unicode form by performing ToASCII followed by ToUnicode. This handles both native
		// Unicode and Punycode representations.
		if asciiHost, err := idna.ToASCII(normalizedHost); err == nil {
			if uh, errUnicode := idna.ToUnicode(asciiHost); errUnicode == nil {
				unicodeHost = uh
			}
		}

		// Spec Rule: RFC 3491 (Table B.2)
		// Map the German Eszett 'ß' to 'ss' as part of IDNA2003 Nameprep compatibility rules since x/net/idna
		// implements IDNA2008 by default.
		normalizedHost = strings.ReplaceAll(unicodeHost, "ß", "ss")
	}

	// Step 3: Scheme-Based Port Normalization
	// Elide the port delimiter and number if the port matches the default port defined by the scheme.
	normalizedPort := port
	if normalizedPort != "" {
		isDefaultPort := (scheme == "http" && normalizedPort == "80") ||
			(scheme == "https" && normalizedPort == "443") ||
			(scheme == "ftp" && normalizedPort == "21") ||
			(scheme == "ws" && normalizedPort == "80") ||
			(scheme == "wss" && normalizedPort == "443")
		if isDefaultPort {
			normalizedPort = ""
		}
	}

	return normalizedHost, normalizedPort
}

// normalizePercentEncoding decodes percent-encoded octets that correspond to unreserved characters and normalizes
// non-unreserved percent-encodings to uppercase.
//
// Specification Reference:
// RFC 3986 (Section 6.2.2.1), RFC 3986 (Section 6.2.2.2), and RFC 3987 (Section 5.3.2.3)
//
// Parameters:
//   - s: The component string containing potentially percent-encoded characters.
//
// Returns:
// - string: The normalized string where percent-encoded unreserved characters are decoded and reserved
// percent-encodings are uppercase.
func normalizePercentEncoding(s string) string {
	var b bytes.Buffer
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		// Spec Rule: RFC 3986 (Section 2.1)
		// A percent-encoded octet is denoted by a percent character "%" followed by two hexadecimal digits.
		if s[i] == '%' && i+2 < len(s) && isASCIIHexDigit(rune(s[i+1])) &&
			isASCIIHexDigit(rune(s[i+2])) {
			decoded, err := hex.DecodeString(s[i+1 : i+3])
			if err == nil {
				// Spec Rule: RFC 3986 (Section 2.3)
				// Decode any percent-encoded octet sequence that corresponds to an unreserved character, which includes
				// ALPHA, DIGIT, and "-", ".", "_", "~".
				c := rune(decoded[0])
				if isUnreserved(c) {
					b.WriteRune(c)
					i += 3
					continue
				}

				// Spec Rule: RFC 3986 (Section 6.2.2.1)
				// Normalize non-unreserved percent-encoded hex digits to uppercase.
				b.WriteByte('%')
				b.WriteByte(toUpperASCII(s[i+1]))
				b.WriteByte(toUpperASCII(s[i+2]))
				i += 3
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// toUpperASCII converts a lowercase ASCII byte to uppercase.
//
// Parameters:
//   - b: The ASCII byte to convert.
//
// Returns:
//   - byte: The uppercase ASCII byte, or the original byte if already uppercase or non-alphabetic.
func toUpperASCII(b byte) byte {
	if 'a' <= b && b <= 'z' {
		return b - 'a' + 'A'
	}
	return b
}
