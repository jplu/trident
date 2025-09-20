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
	"net"
	"strings"

	"golang.org/x/net/idna"
)

const (
	// ipvFutureParts is the number of parts expected in an IPvFuture literal
	// (e.g., "v1.abc"), separated by a dot.
	ipvFutureParts = 2
)

// parseUserinfo handles the userinfo part of the authority.
func (p *iriParser) parseUserinfo(userinfo string) error {
	if userinfo == "" {
		return nil
	}
	if !p.unchecked {
		if err := validateBidiComponent(userinfo); err != nil {
			return err
		}
	}

	// Use a temporary buffer to ensure parsing is transactional.
	var tempBuffer strings.Builder
	tempParser := &iriParser{
		input:     newParserInput(userinfo),
		output:    &stringOutputBuffer{builder: &tempBuffer},
		unchecked: p.unchecked,
	}

	for {
		r, ok := tempParser.input.next()
		if !ok {
			break
		}
		if err := tempParser.readURLCodepointOrEchar(r, func(c rune) bool {
			return isIUnreservedOrSubDelims(c) || c == ':'
		}); err != nil {
			return err
		}
	}

	p.output.writeString(tempBuffer.String())
	p.output.writeRune('@')
	return nil
}

// validateHost checks the host component for structural validity (IP literal format, Bidi rules).
func (p *iriParser) validateHost(host string) error {
	if strings.HasPrefix(host, "[") {
		if !strings.HasSuffix(host, "]") {
			return &kindError{message: "Invalid host IP: unterminated IP literal", details: host}
		}
		ipLiteral := host[1 : len(host)-1]
		if err := p.validateIPLiteral(ipLiteral); err != nil {
			return err
		}
	} else if err := validateBidiHost(host); err != nil {
		return err
	}
	return nil
}

// parseHost handles the host part of the authority.
func (p *iriParser) parseHost(host string) error {
	if host == "" {
		return nil
	}
	if !p.unchecked {
		if err := p.validateHost(host); err != nil {
			return err
		}
	}

	var tempBuffer strings.Builder
	tempParser := &iriParser{
		input:     newParserInput(host),
		output:    &stringOutputBuffer{builder: &tempBuffer},
		unchecked: p.unchecked,
	}

	// This is the correct "consume-then-process" loop.
	for {
		r, ok := tempParser.input.next()
		if !ok {
			break
		}

		if r == '%' {
			// The '%' is now consumed. readEchar can correctly read the next two digits.
			if err := tempParser.readEchar(); err != nil {
				return err
			}
		} else {
			// Check against the allowed character set for a host.
			// The host component allows different characters depending on whether it's an
			// IP literal or a registered name. We must check for all valid possibilities.
			isIPLiteralChar := r == '[' || r == ']' || r == ':'
			if !p.unchecked && !isIUnreservedOrSubDelims(r) && !isIPLiteralChar {
				return &kindError{message: "Invalid character in host", char: r}
			}
			tempParser.output.writeRune(r)
		}
	}

	p.output.writeString(tempBuffer.String())
	return nil
}

// parsePort handles the port part of the authority.
func (p *iriParser) parsePort(port string) error {
	if port == "" {
		return nil
	}
	if !p.unchecked {
		for _, r := range port {
			if !isASCIIDigit(r) {
				return &kindError{message: "Invalid port character", char: r}
			}
		}
	}
	p.output.writeRune(':')
	p.output.writeString(port)
	return nil
}

// parseAuthority is a method on the iriParser that consumes and validates
// the authority component from the input stream.
func (p *iriParser) parseAuthority() error {
	authorityStr := p.input.asStr()
	end := len(authorityStr)
	for i, r := range authorityStr {
		if r == '/' || r == '?' || r == '#' {
			end = i
			break
		}
	}
	authorityPart := authorityStr[:end]

	userinfo, host, port := splitAuthority(authorityPart)

	if err := p.parseUserinfo(userinfo); err != nil {
		return err
	}
	if err := p.parseHost(host); err != nil {
		return err
	}
	if err := p.parsePort(port); err != nil {
		return err
	}

	p.input.reset(authorityStr[end:])
	p.outputPositions.AuthorityEnd = p.output.len()

	return nil
}

// validateIPLiteral checks if a string inside brackets is a valid IPv6 or IPvFuture address.
func (p *iriParser) validateIPLiteral(ipLiteral string) error {
	if strings.HasPrefix(ipLiteral, "v") || strings.HasPrefix(ipLiteral, "V") {
		return p.validateIPVFuture(ipLiteral)
	}
	if net.ParseIP(ipLiteral) == nil {
		return &kindError{message: "Invalid host IP", details: ipLiteral}
	}
	return nil
}

// validateIPVFuture validates an IPvFuture literal (e.g., "v1.something").
func (p *iriParser) validateIPVFuture(ip string) error {
	parts := strings.SplitN(ip[1:], ".", ipvFutureParts)
	if len(parts) != ipvFutureParts {
		return &kindError{message: "Invalid IPvFuture format: no dot separator", details: ip}
	}
	version, address := parts[0], parts[1]
	if version == "" {
		return &kindError{message: "Invalid IPvFuture: missing version", details: ip}
	}
	for _, r := range version {
		if !isASCIIHexDigit(r) {
			return &kindError{message: "Invalid IPvFuture version char", char: r}
		}
	}
	if address == "" {
		return &kindError{message: "Invalid IPvFuture: empty address part", details: ip}
	}
	for _, r := range address {
		if !isUnreservedOrSubDelims(r) && r != ':' {
			return &kindError{message: "Invalid IPvFuture address char", char: r}
		}
	}
	return nil
}

// splitAuthority is the single, stateless utility function that parses an authority
// string into its userinfo, host, and port components.
func splitAuthority(authority string) (string, string, string) {
	var userinfo, host, port string

	endUserinfo := strings.LastIndex(authority, "@")
	hostport := authority
	if endUserinfo != -1 {
		userinfo = authority[:endUserinfo]
		hostport = authority[endUserinfo+1:]
	}

	if strings.HasPrefix(hostport, "[") {
		endBracket := strings.LastIndex(hostport, "]")
		if endBracket == -1 {
			host = hostport
			return userinfo, host, port
		}
		host = hostport[:endBracket+1]
		if len(hostport) > endBracket+1 && hostport[endBracket+1] == ':' {
			port = hostport[endBracket+2:]
		}
		return userinfo, host, port
	}

	endHost := strings.LastIndex(hostport, ":")
	if endHost != -1 {
		host = hostport[:endHost]
		port = hostport[endHost+1:]
	} else {
		host = hostport
	}
	return userinfo, host, port
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
