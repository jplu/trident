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

//nolint:testpackage // White-box test file for an internal package to test unexported functions.
package iri

import (
	"strings"
	"testing"
)

// newPartParser creates a minimal parser instance for testing component parsing functions.
func newPartParser(unchecked bool) *iriParser {
	return &iriParser{
		output:    &stringOutputBuffer{builder: &strings.Builder{}},
		unchecked: unchecked,
		base:      &iriParserBase{},
	}
}

// TestSplitAuthority tests the stateless utility for deconstructing an authority string.
func TestSplitAuthority(t *testing.T) {
	tests := []struct {
		name                 string
		authority            string
		wantUserinfo         string
		wantHost             string
		wantPort             string
		wantHasPortDelimiter bool
	}{
		{
			name:                 "host only",
			authority:            "example.com",
			wantHost:             "example.com",
			wantHasPortDelimiter: false,
		},
		{
			name:                 "host and port",
			authority:            "example.com:8080",
			wantHost:             "example.com",
			wantPort:             "8080",
			wantHasPortDelimiter: true,
		},
		{
			name:                 "userinfo and host",
			authority:            "user@example.com",
			wantUserinfo:         "user",
			wantHost:             "example.com",
			wantHasPortDelimiter: false,
		},
		{
			name:                 "full authority",
			authority:            "user:pass@example.com:8080",
			wantUserinfo:         "user:pass",
			wantHost:             "example.com",
			wantPort:             "8080",
			wantHasPortDelimiter: true,
		},
		{
			name:                 "IPv6 literal host",
			authority:            "[::1]",
			wantHost:             "[::1]",
			wantHasPortDelimiter: false,
		},
		{
			name:                 "IPv6 literal with port",
			authority:            "[::1]:80",
			wantHost:             "[::1]",
			wantPort:             "80",
			wantHasPortDelimiter: true,
		},
		{
			name:                 "full authority with IPv6",
			authority:            "user@[::1]:80",
			wantUserinfo:         "user",
			wantHost:             "[::1]",
			wantPort:             "80",
			wantHasPortDelimiter: true,
		},
		{
			name:                 "empty authority",
			authority:            "",
			wantHasPortDelimiter: false,
		},
		{
			name:                 "multiple @ signs",
			authority:            "user@info@host",
			wantUserinfo:         "user@info",
			wantHost:             "host",
			wantHasPortDelimiter: false,
		},
		{
			name:                 "host with multiple colons (not IPv6)",
			authority:            "host:part:80",
			wantHost:             "host:part",
			wantPort:             "80",
			wantHasPortDelimiter: true,
		},
		{
			name:                 "empty port",
			authority:            "host:",
			wantHost:             "host",
			wantPort:             "",
			wantHasPortDelimiter: true,
		},
		{
			name:                 "empty userinfo",
			authority:            "@host",
			wantUserinfo:         "",
			wantHost:             "host",
			wantHasPortDelimiter: false,
		},
		{
			name:                 "malformed IPv6 literal without closing bracket",
			authority:            "[::1",
			wantHost:             "[::1",
			wantHasPortDelimiter: false,
		},
		{
			name:                 "bracketed host with trailing non-colon characters",
			authority:            "[::1]abc",
			wantUserinfo:         "",
			wantHost:             "[::1]",
			wantPort:             "",
			wantHasPortDelimiter: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserinfo, gotHost, gotPort, gotHasPortDelimiter := splitAuthority(tt.authority)
			if gotUserinfo != tt.wantUserinfo {
				t.Errorf("splitAuthority() gotUserinfo = %v, want %v",
					gotUserinfo, tt.wantUserinfo)
			}
			if gotHost != tt.wantHost {
				t.Errorf("splitAuthority() gotHost = %v, want %v",
					gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("splitAuthority() gotPort = %v, want %v",
					gotPort, tt.wantPort)
			}
			if gotHasPortDelimiter != tt.wantHasPortDelimiter {
				t.Errorf("splitAuthority() gotHasPortDelimiter = %v, want %v",
					gotHasPortDelimiter, tt.wantHasPortDelimiter)
			}
		})
	}
}

// TestValidateIPVFuture tests validation of IPvFuture literals.
func TestValidateIPVFuture(t *testing.T) {
	p := &iriParser{}
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{name: "valid simple", ip: "v1.future-address"},
		{name: "valid hex version", ip: "vF9.more.stuff"},
		{name: "invalid - no v prefix", ip: "1.future-address", wantErr: true},
		{name: "invalid - non-hex version", ip: "vg.future-address", wantErr: true},
		{name: "invalid - no dot separator", ip: "v1future-address", wantErr: true},
		{name: "invalid - missing version", ip: "v.future-address", wantErr: true},
		{name: "invalid - empty address", ip: "v1.", wantErr: true},
		{name: "invalid - bad address character", ip: "v1.bad/char", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.validateIPVFuture(tt.ip); (err != nil) != tt.wantErr {
				t.Errorf("validateIPVFuture() error = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}
}

// TestValidateIPLiteral tests validation of IP literals (IPv6 and IPvFuture).
func TestValidateIPLiteral(t *testing.T) {
	p := &iriParser{}
	tests := []struct {
		name      string
		ipLiteral string
		wantErr   bool
	}{
		{name: "valid IPv6", ipLiteral: "2001:db8::1"},
		{name: "valid IPv6 mapped IPv4", ipLiteral: "::ffff:192.0.2.128", wantErr: true},
		{name: "valid IPvFuture", ipLiteral: "v1.example"},
		{name: "invalid IPv6", ipLiteral: "not-an-ip", wantErr: true},
		{name: "invalid IPv6 double colon", ipLiteral: "2001::db8::1", wantErr: true},
		{name: "invalid IPvFuture", ipLiteral: "v1.bad_char{", wantErr: true},
		{name: "invalid - starts with V (uppercase)", ipLiteral: "V1.is-valid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.validateIPLiteral(tt.ipLiteral); (err != nil) != tt.wantErr {
				t.Errorf("validateIPLiteral() error = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}
}

// TestValidateHost tests validation of the host component.
func TestValidateHost(t *testing.T) {
	p := &iriParser{}
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{name: "valid registered name", host: "example.com"},
		{name: "valid IP literal", host: "[::1]"},
		{name: "invalid IP literal", host: "[not-an-ip]", wantErr: true},
		{name: "valid Bidi host (Hebrew)", host: "xn--5db0a.xn--4dbrk0ce"},
		{
			name:    "invalid Bidi host (mixed script in label)",
			host:    "abc\u05d0\u05d1\u05d2def.com",
			wantErr: false,
		},
		{name: "unterminated IP literal", host: "[::1", wantErr: true},
		{name: "empty IP literal", host: "[]", wantErr: true},
		{name: "IP-literal IPv4", host: "[192.168.0.1]", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := p.validateHost(tt.host); (err != nil) != tt.wantErr {
				t.Errorf("validateHost() error = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}
}

// TestParsePort tests parsing of the port subcomponent.
func TestParsePort(t *testing.T) {
	tests := []struct {
		name             string
		port             string
		hasPortDelimiter bool
		unchecked        bool
		wantErr          bool
		wantOut          string
	}{
		{"valid", "8080", true, false, false, ":8080"},
		{"empty without delimiter", "", false, false, false, ""},
		{"empty with delimiter", "", true, false, false, ":"},
		{"invalid char", "80a80", true, false, true, ""},
		{"valid unchecked", "8080", true, true, false, ":8080"},
		{"invalid unchecked", "80a80", true, true, false, ":80a80"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPartParser(tt.unchecked)
			err := p.parsePort(tt.port, tt.hasPortDelimiter)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parsePort() error = %v, wantErr %v",
					err, tt.wantErr)
			}
			if gotOut := p.output.string(); gotOut != tt.wantOut {
				t.Errorf("parsePort() gotOut = %v, want %v",
					gotOut, tt.wantOut)
			}
		})
	}
}

// TestParseUserinfo tests parsing of the userinfo subcomponent.
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
		{
			"valid unreserved and sub-delims",
			"a-._~:!$&'()*+,;=",
			false,
			false,
			"a-._~:!$&'()*+,;=@",
		},
		{"valid percent-encoded", "a%20b", false, false, "a%20b@"},
		{"empty", "", false, false, ""},
		{"invalid char", "user/", false, true, ""},
		{"invalid percent-encoding parsed as literal", "a%2xb", false, true, ""},
		{"invalid bidi", "a\u05d0b", false, false, "a\u05d0b@"},
		{"unchecked with invalid char", "user/", true, false, "user/@"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPartParser(tt.unchecked)
			err := p.parseUserinfo(tt.userinfo)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseUserinfo() error = %v, wantErr %v",
					err, tt.wantErr)
			}
			if gotOut := p.output.string(); gotOut != tt.wantOut {
				t.Errorf("parseUserinfo() gotOut = %q, want %q",
					gotOut, tt.wantOut)
			}
		})
	}
}

// TestParseHost tests parsing of the host subcomponent.
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
		{"invalid percent-encoding parsed as literal", "a%2xb.com", false, true, ""},
		{"invalid ip literal", "[not-an-ip]", false, true, ""},
		{"unchecked with invalid char", "bad/host", true, false, "bad/host"},
		{"unchecked with invalid percent", "bad%2ghost", true, true, ""},
		{"unchecked with valid percent", "a%20b.com", true, false, "a%20b.com"},
		{"valid IP future", "[v1.abc]", false, false, "[v1.abc]"},
		{"valid IP future uppercase V", "[V1.abc]", false, false, "[V1.abc]"},
		{
			name:      "unterminated ip literal",
			host:      "[::1",
			unchecked: false,
			wantErr:   true,
			wantOut:   "",
		},
		{
			name:      "unterminated ip literal unchecked",
			host:      "[::1",
			unchecked: true,
			wantErr:   false,
			wantOut:   "[::1",
		},
		{
			name:      "invalid character in bracketed host",
			host:      "[::1/]",
			unchecked: false,
			wantErr:   true,
			wantOut:   "",
		},
		{
			name:      "invalid character in bracketed host unchecked",
			host:      "[::1/]",
			unchecked: true,
			wantErr:   false,
			wantOut:   "[::1/]",
		},
		{
			name:      "invalid percent-encoded character",
			host:      "bad%2ghost",
			unchecked: false,
			wantErr:   true,
			wantOut:   "",
		},
		{
			name:      "too short percent-encoding sequence",
			host:      "a%",
			unchecked: false,
			wantErr:   true,
			wantOut:   "",
		},
		{
			name:      "too short percent-encoding sequence unchecked",
			host:      "a%",
			unchecked: true,
			wantErr:   true,
			wantOut:   "",
		},
		{
			name:      "valid ip literal unchecked",
			host:      "[::1]",
			unchecked: true,
			wantErr:   false,
			wantOut:   "[::1]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newPartParser(tt.unchecked)
			err := p.parseHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseHost() error = %v, wantErr %v",
					err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotOut := p.output.string(); gotOut != tt.wantOut {
				t.Errorf("parseHost() gotOut = %q, want %q",
					gotOut, tt.wantOut)
			}
		})
	}
}

// TestValidateHostCharDirect targets missing branches of validateHostChar that
// are difficult to reach via high-level parser flows.
func TestValidateHostCharDirect(t *testing.T) {
	p := &iriParser{}

	pUnchecked := &iriParser{unchecked: true}
	if err := pUnchecked.validateHostChar('/', false); err != nil {
		t.Errorf("validateHostChar with unchecked=true should return nil, got %v", err)
	}

	for _, r := range []rune{'[', ']', ':', 'a', '1', '-', '.'} {
		if err := p.validateHostChar(r, true); err != nil {
			t.Errorf("validateHostChar with isIPLiteral=true, char=%q should return nil, got %v", r, err)
		}
	}
	if err := p.validateHostChar('/', true); err == nil {
		t.Errorf("validateHostChar with isIPLiteral=true, char='/' should return error, got nil")
	}

	for _, r := range []rune{'a', '1', '-', '.'} {
		if err := p.validateHostChar(r, false); err != nil {
			t.Errorf("validateHostChar with isIPLiteral=false, char=%q should return nil, got %v", r, err)
		}
	}
	for _, r := range []rune{'[', ']', ':', '/'} {
		if err := p.validateHostChar(r, false); err == nil {
			t.Errorf("validateHostChar with isIPLiteral=false, char=%q should return error, got nil", r)
		}
	}
}

// TestParseAuthority tests the main parser for the authority component.
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
			name:    "invalid userinfo (bad percent encoding) parsed as literal",
			input:   "user%@host.com/path",
			wantErr: true,
		},
		{name: "invalid port", input: "example.com:bad/path", wantErr: true},
		{name: "truly invalid host", input: "bad{host}.com/path", wantErr: true},
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
				t.Fatalf("parseAuthority() error = %v, wantErr %v",
					err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotOutput := p.output.string(); gotOutput != tt.wantOutput {
				t.Errorf("parseAuthority() gotOutput = %q, want %q",
					gotOutput, tt.wantOutput)
			}
			if p.outputPositions.AuthorityEnd != tt.wantAuthority {
				t.Errorf("parseAuthority() AuthorityEnd = %d, want %d",
					p.outputPositions.AuthorityEnd, tt.wantAuthority)
			}
			if gotRemainder := p.input.asStr(); gotRemainder != tt.wantRemainder {
				t.Errorf("parseAuthority() gotRemainder = %q, want %q",
					gotRemainder, tt.wantRemainder)
			}
		})
	}
}
