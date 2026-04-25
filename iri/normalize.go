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

// Normalize applies syntax-based normalization to the IRI reference according
// to RFC 3986, Section 6.2.2. This includes case-normalization of the scheme
// and host, percent-encoding normalization, and path-segment normalization.
// It also ensures the resulting IRI is in Unicode Normalization Form C (NFC).
// It returns a new, normalized Ref.
func (r *Ref) Normalize() *Ref {
	if r.iri == "" {
		return &Ref{}
	}

	scheme, hasScheme := r.Scheme()
	authority, hasAuthority := r.Authority()
	path := r.Path()
	query, hasQuery := r.Query()
	fragment, hasFragment := r.Fragment()

	// 1. Case Normalization
	if hasScheme {
		scheme = strings.ToLower(scheme)
	}
	var userinfo, host, port string
	if hasAuthority {
		userinfo, host, port = splitAuthority(authority)
		host, port = normalizeHostAndPort(host, port, scheme)
	}

	// 2. Percent-Encoding Normalization
	userinfo = normalizePercentEncoding(userinfo)
	host = normalizePercentEncoding(host)
	path = normalizePercentEncoding(path)
	query = normalizePercentEncoding(query)
	fragment = normalizePercentEncoding(fragment)

	// 3. Path Segment Normalization
	path = removeDotSegments(path)

	// 4. Scheme-based normalization for path
	if hasAuthority && path == "" {
		path = "/"
	}

	// Recompose and re-parse
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
	// An error is not expected here as we are building from valid components.
	// We use the compliant ParseRef because normalizedStr is now guaranteed to be NFC.
	newRef, _ := ParseRef(normalizedStr)
	return newRef
}

// recomposeNormalizedIRI builds an IRI string from its normalized components.
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

// normalizeHostAndPort applies case, IDNA, and scheme-based port normalization.
func normalizeHostAndPort(host, port, scheme string) (string, string) {
	// Case normalization for host.
	normalizedHost := strings.ToLower(host)

	// IDNA normalization.
	if !strings.HasPrefix(normalizedHost, "[") {
		unicodeHost := normalizedHost
		// First, get the canonical Unicode form using the library. This
		// handles both direct Unicode and Punycode input.
		if asciiHost, err := idna.ToASCII(normalizedHost); err == nil {
			if uh, errUnicode := idna.ToUnicode(asciiHost); errUnicode == nil {
				unicodeHost = uh
			}
		}

		// Apply specific mappings from Nameprep (RFC 3491, Table B.2)
		// that are part of IDNA2003 but not IDNA2008 (as implemented by x/net/idna).
		// The most prominent example is the mapping of German Eszett 'ß' to 'ss'
		// because 'ss' will always be translated to `ß` with `ToUnicode` even
		// if the `Transitional` option is set to `true`.
		normalizedHost = strings.ReplaceAll(unicodeHost, "ß", "ss")
	}

	// Scheme-based port normalization.
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
