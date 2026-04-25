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

import (
	"testing"

	"golang.org/x/text/unicode/norm"
)

// TestRef_Normalize tests the syntax-based and scheme-based normalization of a Ref.
// Based on RFC 3986, Section 6.2.2 and 6.2.3: Syntax-Based and Scheme-Based Normalization.
func TestRef_Normalize(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"Case normalization (scheme, host)",
			"HTTP://User@Example.COM/Path",
			"http://User@example.com/Path",
		},
		{
			"Percent-encoding normalization (decode unreserved)",
			"http://example.com/%7Euser",
			"http://example.com/~user",
		},
		{
			"Path segment normalization (remove dot segments)",
			"http://example.com/a/b/../c/./d",
			"http://example.com/a/c/d",
		},
		{
			"Scheme-based: add / for empty path with authority",
			"http://example.com",
			"http://example.com/",
		},
		{
			"Scheme-based: remove default port",
			"http://example.com:80/path",
			"http://example.com/path",
		},
		{
			"Scheme-based: keep non-default port",
			"http://example.com:8080/path",
			"http://example.com:8080/path",
		},
		{
			"NFC normalization",
			"http://example.com/re\u0301sume\u0301.html",
			"http://example.com/résumé.html",
		},
		{
			"Combination of normalizations",
			"HTTP://EXAMPLE.COM:80/a/../b/%7E",
			"http://example.com/b/~",
		},
		{"Empty IRI", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ref := mustParseRef(t, tc.input)
			normalizedRef := ref.Normalize()
			if normalizedRef.String() != tc.expected {
				t.Errorf("Expected normalized IRI '%s', got '%s'", tc.expected, normalizedRef.String())
			}
		})
	}

	t.Run("No-op returns same instance", func(t *testing.T) {
		iriStr := "http://example.com/already/normalized"
		ref := mustParseRef(t, iriStr)
		normalizedRef := ref.Normalize()
		if ref != normalizedRef {
			t.Error("Should return same instance if already normalized")
		}
	})
}

// TestRecomposeNormalizedIRI tests the recomposition of an IRI from normalized components.
// RFC Reference: RFC 3986, Section 5.3.
func TestRecomposeNormalizedIRI(t *testing.T) {
	tests := []struct {
		name      string
		scheme    string
		hasScheme bool
		userinfo  string
		host      string
		port      string
		hasAuth   bool
		path      string
		query     string
		hasQuery  bool
		fragment  string
		hasFrag   bool
		want      string
	}{
		{"Full", "http", true, "user", "host", "80", true, "/p", "q", true, "f", true, "http://user@host:80/p?q#f"},
		{"No user/port", "http", true, "", "host", "", true, "/p", "q", true, "f", true, "http://host/p?q#f"},
		{"Scheme relative", "", false, "", "host", "", true, "/p", "", false, "", false, "//host/p"},
		{"No authority", "urn", true, "", "", "", false, "a:b", "", false, "", false, "urn:a:b"},
		{"Path only", "", false, "", "", "", false, "/p", "", false, "", false, "/p"},
		{"Empty query", "http", true, "", "h", "", true, "/p", "", true, "", false, "http://h/p?"},
		{"Empty fragment", "http", true, "", "h", "", true, "/p", "", false, "", true, "http://h/p#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := recomposeNormalizedIRI(
				tt.scheme, tt.hasScheme,
				tt.userinfo, tt.host, tt.port, tt.hasAuth,
				tt.path,
				tt.query, tt.hasQuery,
				tt.fragment, tt.hasFrag,
			)
			if got != tt.want {
				t.Errorf("recomposeNormalizedIRI() got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNormalizeHostAndPort tests case, IDNA, and scheme-based port normalization.
// Based on RFC 3986, Sections 6.2.2.1 and 6.2.3.
func TestNormalizeHostAndPort(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		scheme   string
		wantHost string
		wantPort string
	}{
		{
			name:     "host case normalization",
			host:     "EXAMPLE.COM",
			port:     "8080",
			scheme:   "http",
			wantHost: "example.com",
			wantPort: "8080",
		},
		{
			name:     "http default port removal",
			host:     "example.com",
			port:     "80",
			scheme:   "http",
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "https default port removal",
			host:     "example.com",
			port:     "443",
			scheme:   "https",
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "ftp default port removal",
			host:     "example.com",
			port:     "21",
			scheme:   "ftp",
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "ws default port removal",
			host:     "example.com",
			port:     "80",
			scheme:   "ws",
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "wss default port removal",
			host:     "example.com",
			port:     "443",
			scheme:   "wss",
			wantHost: "example.com",
			wantPort: "",
		},
		{
			name:     "non-default port preserved",
			host:     "example.com",
			port:     "8080",
			scheme:   "http",
			wantHost: "example.com",
			wantPort: "8080",
		},
		{
			name:     "port preserved for unknown scheme",
			host:     "example.com",
			port:     "80",
			scheme:   "gopher",
			wantHost: "example.com",
			wantPort: "80",
		},
		{
			name:     "IPv6 literal case normalization",
			host:     "[2001:DB8::7]",
			port:     "",
			scheme:   "http",
			wantHost: "[2001:db8::7]",
			wantPort: "",
		},
		{
			name:     "IDNA normalization",
			host:     "faß.de",
			port:     "",
			scheme:   "http",
			wantHost: "fass.de",
			wantPort: "",
		},
		{
			name:     "empty host and port",
			host:     "",
			port:     "",
			scheme:   "http",
			wantHost: "",
			wantPort: "",
		},
		{
			name:     "IP literal with default port",
			host:     "[::1]",
			port:     "80",
			scheme:   "http",
			wantHost: "[::1]",
			wantPort: "",
		},
		{
			name:     "normalize punycode",
			host:     "xn--fa-hia.de",
			port:     "",
			scheme:   "http",
			wantHost: "fass.de",
			wantPort: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort := normalizeHostAndPort(tt.host, tt.port, tt.scheme)
			if gotHost != tt.wantHost {
				t.Errorf("normalizeHostAndPort() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("normalizeHostAndPort() gotPort = %v, want %v", gotPort, tt.wantPort)
			}
		})
	}
}

// TestNormalizePercentEncoding tests the normalization of percent-encoded octets.
// RFC Reference: RFC 3986, Section 6.2.2.2.
func TestNormalizePercentEncoding(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Decode unreserved ALPHA",
			input:    "example.com/%41%42%43",
			expected: "example.com/ABC",
		},
		{
			name:     "Decode unreserved DIGIT",
			input:    "example.com/%31%32%33",
			expected: "example.com/123",
		},
		{
			name:     "Decode unreserved special chars (-._~)",
			input:    "%2D%2E%5F%7E",
			expected: "-._~",
		},
		{
			name:     "Do not decode reserved gen-delims",
			input:    "a%2Fb%3Ac",
			expected: "a%2Fb%3Ac",
		},
		{
			name:     "Do not decode reserved sub-delims",
			input:    "a%24b%26c",
			expected: "a%24b%26c",
		},
		{
			name:     "Do not decode non-ASCII UTF-8 sequence",
			input:    "r%C3%A9sum%C3%A9",
			expected: "r%C3%A9sum%C3%A9",
		},
		{
			name:     "Mixed reserved and unreserved",
			input:    "a%2Fb%2E%31",
			expected: "a%2Fb.1",
		},
		{
			name:     "Lowercase hex digits are preserved",
			input:    "a%2fb%2e%31",
			expected: "a%2fb.1",
		},
		{
			name:     "Invalid encoding - short",
			input:    "a%2",
			expected: "a%2",
		},
		{
			name:     "Invalid encoding - non-hex",
			input:    "a%2G",
			expected: "a%2G",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "No percent encoding",
			input:    "abc-123",
			expected: "abc-123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizePercentEncoding(tc.input)
			if result != tc.expected {
				t.Errorf("normalizePercentEncoding(%q) = %q; want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestParseNormalizedRef_NFC ensures ParseNormalizedRef applies NFC normalization.
// RFC Reference: RFC 3987, Sections 3.1 and 5.3.2.2.
func TestParseNormalizedRef_NFC(t *testing.T) {
	decomposed := "e\u0301"
	composed := "\u00e9"

	if !norm.NFC.IsNormalString(composed) {
		t.Fatalf("Test setup error: composed string '%s' is not in NFC", composed)
	}
	if norm.NFC.IsNormalString(decomposed) {
		t.Fatalf("Test setup error: decomposed string '%s' is in NFC", decomposed)
	}

	iriStr := "http://example.com/" + decomposed
	ref, err := ParseNormalizedRef(iriStr)
	if err != nil {
		t.Fatalf("ParseNormalizedRef failed: %v", err)
	}

	expectedStr := "http://example.com/" + composed
	if ref.String() != expectedStr {
		t.Errorf("Expected IRI string to be normalized to NFC '%s', got '%s'", expectedStr, ref.String())
	}
}
