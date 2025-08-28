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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package iri

import (
	"strings"
	"testing"
)

// newPartParser creates a minimal parser instance for testing component parsing functions
// that take a string as input (e.g., parseUserinfo, parseHost, parsePort).
func newPartParser(unchecked bool) *iriParser {
	return &iriParser{
		// No input reader needed as these functions don't consume from it.
		output:    &stringOutputBuffer{builder: &strings.Builder{}},
		unchecked: unchecked,
		base:      &iriParserBase{},
	}
}

// TestSplitAuthority tests the stateless utility for deconstructing an authority string.
// This is based on the ABNF from RFC 3986, Section 3.2.
func TestSplitAuthority(t *testing.T) {
	tests := []struct {
		name         string
		authority    string
		wantUserinfo string
		wantHost     string
		wantPort     string
	}{
		{
			name:      "host only",
			authority: "example.com",
			wantHost:  "example.com",
		},
		{
			name:      "host and port",
			authority: "example.com:8080",
			wantHost:  "example.com",
			wantPort:  "8080",
		},
		{
			name:         "userinfo and host",
			authority:    "user@example.com",
			wantUserinfo: "user",
			wantHost:     "example.com",
		},
		{
			name:         "full authority",
			authority:    "user:pass@example.com:8080",
			wantUserinfo: "user:pass",
			wantHost:     "example.com",
			wantPort:     "8080",
		},
		{
			name:      "IPv6 literal host",
			authority: "[::1]",
			wantHost:  "[::1]",
		},
		{
			name:      "IPv6 literal with port",
			authority: "[::1]:80",
			wantHost:  "[::1]",
			wantPort:  "80",
		},
		{
			name:         "full authority with IPv6",
			authority:    "user@[::1]:80",
			wantUserinfo: "user",
			wantHost:     "[::1]",
			wantPort:     "80",
		},
		{
			name:      "empty authority",
			authority: "",
		},
		{
			name:         "multiple @ signs",
			authority:    "user@info@host",
			wantUserinfo: "user@info",
			wantHost:     "host",
		},
		{
			name:      "host with multiple colons (not IPv6)",
			authority: "host:part:80",
			wantHost:  "host:part",
			wantPort:  "80",
		},
		{
			name:      "empty port",
			authority: "host:",
			wantHost:  "host",
			wantPort:  "",
		},
		{
			name:         "empty userinfo",
			authority:    "@host",
			wantUserinfo: "",
			wantHost:     "host",
		},
		{
			name:      "malformed IPv6 literal without closing bracket",
			authority: "[::1",
			wantHost:  "[::1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserinfo, gotHost, gotPort := splitAuthority(tt.authority)
			if gotUserinfo != tt.wantUserinfo {
				t.Errorf("splitAuthority() gotUserinfo = %v, want %v", gotUserinfo, tt.wantUserinfo)
			}
			if gotHost != tt.wantHost {
				t.Errorf("splitAuthority() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("splitAuthority() gotPort = %v, want %v", gotPort, tt.wantPort)
			}
		})
	}
}

// TestNormalizeHostAndPort tests the syntax-based and scheme-based normalization of host and port.
// This is based on RFC 3986, Sections 6.2.2.1 and 6.2.3.
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
			host:     "faß.de", // German Eszett, per RFC 3491 Nameprep must be mapped to 'ss'
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
			host:     "xn--fa-hia.de", // Punycode for "faß.de"
			port:     "",
			scheme:   "http",
			wantHost: "fass.de", // Should decode to Unicode, then apply Nameprep mapping 'ß' -> 'ss'
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

// TestValidateIPVFuture tests the validation of IPvFuture literals.
// This is based on the ABNF from RFC 3986, Section 3.2.2.
func TestValidateIPVFuture(t *testing.T) {
	p := &iriParser{}
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name: "valid simple",
			ip:   "v1.future-address",
		},
		{
			name: "valid hex version",
			ip:   "vF9.more.stuff",
		},
		{
			name:    "invalid - no v prefix",
			ip:      "1.future-address",
			wantErr: true,
		},
		{
			name:    "invalid - non-hex version",
			ip:      "vg.future-address",
			wantErr: true,
		},
		{
			name:    "invalid - no dot separator",
			ip:      "v1future-address",
			wantErr: true,
		},
		{
			name:    "invalid - missing version",
			ip:      "v.future-address",
			wantErr: true,
		},
		{
			name:    "invalid - empty address",
			ip:      "v1.",
			wantErr: true,
		},
		{
			name:    "invalid - bad address character",
			ip:      "v1.bad/char",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.validateIPVFuture(tt.ip); (err != nil) != tt.wantErr {
				t.Errorf("validateIPVFuture() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateIPLiteral tests the validation of IP literals (IPv6 and IPvFuture).
// This is based on RFC 3986, Section 3.2.2.
// It covers both valid IPv6 addresses and valid IPvFuture addresses.
func TestValidateIPLiteral(t *testing.T) {
	p := &iriParser{}
	tests := []struct {
		name      string
		ipLiteral string
		wantErr   bool
	}{
		{
			name:      "valid IPv6",
			ipLiteral: "2001:db8::1",
		},
		{
			name:      "valid IPv6 mapped IPv4",
			ipLiteral: "::ffff:192.0.2.128",
		},
		{
			name:      "valid IPvFuture",
			ipLiteral: "v1.example",
		},
		{
			name:      "invalid IPv6",
			ipLiteral: "not-an-ip",
			wantErr:   true,
		},
		{
			name:      "invalid IPv6 double colon",
			ipLiteral: "2001::db8::1",
			wantErr:   true,
		},
		{
			name:      "invalid IPvFuture",
			ipLiteral: "v1.bad_char{",
			wantErr:   true,
		},
		{
			name:      "invalid - starts with V (uppercase)",
			ipLiteral: "V1.is-valid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.validateIPLiteral(tt.ipLiteral); (err != nil) != tt.wantErr {
				t.Errorf("validateIPLiteral() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateHost tests the validation of the host component.
// This is based on RFC 3986, Section 3.2.2 and RFC 3987, Section 4.2.
func TestValidateHost(t *testing.T) {
	p := &iriParser{}
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{
			name: "valid registered name",
			host: "example.com",
		},
		{
			name: "valid IP literal",
			host: "[::1]",
		},
		{
			name:    "invalid IP literal",
			host:    "[not-an-ip]",
			wantErr: true,
		},
		{
			name: "valid Bidi host (Hebrew)",
			// Each label is a valid bidi component
			host: "xn--5db0a.xn--4dbrk0ce", // example.com in Hebrew
		},
		{
			name:    "invalid Bidi host (mixed script in label)",
			host:    "abc\u05d0\u05d1\u05d2def.com",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.validateHost(tt.host); (err != nil) != tt.wantErr {
				t.Errorf("validateHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestParsePort tests the parsing of the port subcomponent.
// This is based on RFC 3986, Section 3.2.3.
func TestParsePort(t *testing.T) {
	tests := []struct {
		name      string
		port      string
		unchecked bool
		wantErr   bool
		wantOut   string
	}{
		{"valid", "8080", false, false, ":8080"},
		{"empty", "", false, false, ""},
		{"invalid char", "80a80", false, true, ""},
		{"valid unchecked", "8080", true, false, ":8080"},
		{"invalid unchecked", "80a80", true, false, ":80a80"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPartParser(tt.unchecked)
			err := p.parsePort(tt.port)

			if (err != nil) != tt.wantErr {
				t.Fatalf("parsePort() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotOut := p.output.string(); gotOut != tt.wantOut {
				t.Errorf("parsePort() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

// TestParseUserinfo tests the parsing of the userinfo subcomponent.
// This is based on RFC 3986, Section 3.2.1 and RFC 3987 for Bidi rules.
func TestParseUserinfo(t *testing.T) {
	tests := []struct {
		name      string
		userinfo  string
		unchecked bool
		wantErr   bool
		wantOut   string
	}{
		{"valid simple", "user", false, false, "user@"},
		{"valid with password", "user:pass", false, false, "user:pass@"},
		{"valid unreserved and sub-delims", "a-._~:!$&'()*+,;=", false, false, "a-._~:!$&'()*+,;=@"},
		{"valid percent-encoded", "a%20b", false, false, "a%20b@"},
		{"empty", "", false, false, ""},
		{"invalid char", "user/", false, true, ""},
		{"invalid percent-encoding", "a%2xb", false, true, ""},
		{"invalid bidi", "a\u05d0b", false, true, ""}, // mixed ltr/rtl
		{"unchecked with invalid char", "user/", true, false, "user/@"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPartParser(tt.unchecked)
			err := p.parseUserinfo(tt.userinfo)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseUserinfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotOut := p.output.string(); gotOut != tt.wantOut {
				t.Errorf("parseUserinfo() gotOut = %q, want %q", gotOut, tt.wantOut)
			}
		})
	}
}

// TestParseHost tests the parsing of the host subcomponent.
// This is based on RFC 3986, Section 3.2.2.
func TestParseHost(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		unchecked bool
		wantErr   bool
		wantOut   string
	}{
		{"valid reg-name", "example.com", false, false, "example.com"},
		{"valid ip literal", "[::1]", false, false, "[::1]"},
		{"valid percent-encoded", "a%20b.com", false, false, "a%20b.com"},
		{"empty", "", false, false, ""},
		{"invalid char", "bad/host", false, true, ""},
		{"invalid percent-encoding", "a%2xb.com", false, true, ""},
		{"invalid ip literal", "[not-an-ip]", false, true, ""},
		{"unchecked with invalid char", "bad/host", true, false, "bad/host"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPartParser(tt.unchecked)
			err := p.parseHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseHost() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotOut := p.output.string(); gotOut != tt.wantOut {
				t.Errorf("parseHost() gotOut = %q, want %q", gotOut, tt.wantOut)
			}
		})
	}
}

// TestParseAuthority tests the main parser for the authority component.
// It orchestrates the parsing of userinfo, host, and port from the input stream.
// This is based on RFC 3986, Section 3.2.
func TestParseAuthority(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		unchecked     bool
		wantErr       bool
		wantOutput    string
		wantAuthority int
		wantRemainder string
	}{
		{
			name:          "full authority with path",
			input:         "user@example.com:8080/path",
			wantOutput:    "user@example.com:8080",
			wantAuthority: 21,
			wantRemainder: "/path",
		},
		{
			name:          "host only with query",
			input:         "example.com?query",
			wantOutput:    "example.com",
			wantAuthority: 11,
			wantRemainder: "?query",
		},
		{
			name:          "ip literal with fragment",
			input:         "[::1]#fragment",
			wantOutput:    "[::1]",
			wantAuthority: 5,
			wantRemainder: "#fragment",
		},
		{
			name:          "host only eof",
			input:         "example.com",
			wantOutput:    "example.com",
			wantAuthority: 11,
			wantRemainder: "",
		},
		{
			name:          "empty authority with path",
			input:         "/path",
			wantOutput:    "",
			wantAuthority: 0,
			wantRemainder: "/path",
		},
		{
			name:    "invalid userinfo (bad percent encoding)",
			input:   "user%@host.com/path",
			wantErr: true,
		},
		{
			name:    "invalid port",
			input:   "example.com:bad/path",
			wantErr: true,
		},
		{
			name:    "truly invalid host",
			input:   "bad{host}.com/path",
			wantErr: true,
		},
		{
			name:          "unchecked invalid port",
			input:         "example.com:bad/path",
			unchecked:     true,
			wantOutput:    "example.com:bad",
			wantAuthority: 15,
			wantRemainder: "/path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &iriParser{
				iri:       tt.input,
				input:     newParserInput(tt.input),
				output:    &stringOutputBuffer{builder: &strings.Builder{}},
				unchecked: tt.unchecked,
				base:      &iriParserBase{},
			}
			err := p.parseAuthority()

			if (err != nil) != tt.wantErr {
				t.Fatalf("parseAuthority() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotOutput := p.output.string(); gotOutput != tt.wantOutput {
				t.Errorf("parseAuthority() gotOutput = %q, want %q", gotOutput, tt.wantOutput)
			}
			if p.outputPositions.AuthorityEnd != tt.wantAuthority {
				t.Errorf(
					"parseAuthority() AuthorityEnd = %d, want %d",
					p.outputPositions.AuthorityEnd,
					tt.wantAuthority,
				)
			}
			if gotRemainder := p.input.asStr(); gotRemainder != tt.wantRemainder {
				t.Errorf("parseAuthority() gotRemainder = %q, want %q", gotRemainder, tt.wantRemainder)
			}
		})
	}
}
