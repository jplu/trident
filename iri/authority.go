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
	"net"
	"strings"
)

const (
	// ipvFutureParts defines the number of parts expected in an IPvFuture literal (e.g., "v1.abc"), separated by a dot.
	ipvFutureParts = 2
)

// parseUserinfo processes and validates the userinfo subcomponent of the authority component.
//
// Specification Reference:
// RFC 3987 (Section 2.2), RFC 3986 (Section 3.2.1)
//
// Parameters:
//   - userinfo: The raw userinfo string to parse and validate.
//
// Returns:
//   - error: An error if validation or parsing fails, nil otherwise.
func (p *iriParser) parseUserinfo(userinfo string) error {
	if userinfo == "" {
		return nil
	}

	// Implementation Note: Transactional state buffering.
	// Uses a temporary buffer and secondary parser instance to ensure parsing of the userinfo string is transactional
	// and rollback-capable upon failure.
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

// validateHost checks the structural validity of the host component against IP literal rules.
//
// Specification Reference:
// RFC 3986 (Section 3.2.2)
//
// Parameters:
//   - host: The host string to be structurally checked.
//
// Returns:
//   - error: An error if the host does not conform to the required syntax, nil otherwise.
func (p *iriParser) validateHost(host string) error {
	if strings.HasPrefix(host, "[") {
		if !strings.HasSuffix(host, "]") {
			return &kindError{message: "Invalid host IP: unterminated IP literal", details: host}
		}
		ipLiteral := host[1 : len(host)-1]
		if err := p.validateIPLiteral(ipLiteral); err != nil {
			return err
		}
	}
	return nil
}

// validateHostChar checks if a single character is allowed in a host component.
//
// Specification Reference:
// RFC 3987 (Section 2.2), RFC 3986 (Section 3.2.2)
//
// Parameters:
//   - r: The rune to be validated.
//   - isIPLiteral: A boolean indicating whether the host is an IP literal.
//
// Returns:
//   - error: An error if the character is not allowed, nil otherwise.
func (p *iriParser) validateHostChar(r rune, isIPLiteral bool) error {
	if p.unchecked {
		return nil
	}
	if isIPLiteral {
		isIPLiteralChar := r == '[' || r == ']' || r == ':'
		if !isIUnreservedOrSubDelims(r) && !isIPLiteralChar {
			return &kindError{message: "Invalid character in host", char: r}
		}
	} else if !isIUnreservedOrSubDelims(r) {
		return &kindError{message: "Invalid character in host", char: r}
	}
	return nil
}

// parseHost processes, validates, and writes the host component to the internal parser output buffer.
//
// Specification Reference:
// RFC 3987 (Section 2.2), RFC 3986 (Section 3.2.2)
//
// Parameters:
//   - host: The host string to parse.
//
// Returns:
//   - error: An error if verification fails or an invalid character is encountered, nil otherwise.
func (p *iriParser) parseHost(host string) error {
	if host == "" {
		return nil
	}
	if !p.unchecked {
		if err := p.validateHost(host); err != nil {
			return err
		}
	}

	// Spec Rule: RFC 3986 (Section 3.2.2)
	// Bracketed IP literals must be identified prior to the character loop to restrict colon and bracket characters
	// strictly to IP formats.
	isIPLiteral := strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]")

	var tempBuffer strings.Builder
	tempParser := &iriParser{
		input:     newParserInput(host),
		output:    &stringOutputBuffer{builder: &tempBuffer},
		unchecked: p.unchecked,
	}

	for {
		r, ok := tempParser.input.next()
		if !ok {
			break
		}

		if r == '%' {
			if !tempParser.input.hasHexDigits() {
				return &kindError{message: "Invalid percent-encoding sequence", char: '%'}
			}
			// Implementation Note: Percent-encoding sequence validation.
			// Since hasHexDigits() returned true, readEchar() is guaranteed to succeed.
			_ = tempParser.readEchar()
		} else {
			// Spec Rule: RFC 3987 (Section 2.2)
			// Validates character legality against allowable iunreserved, sub-delims, or IP literal characters.
			if err := tempParser.validateHostChar(r, isIPLiteral); err != nil {
				return err
			}
			tempParser.output.writeRune(r)
		}
	}

	p.output.writeString(tempBuffer.String())
	return nil
}

// parsePort processes, validates, and appends the port subcomponent to the internal parser output buffer.
//
// Specification Reference:
// RFC 3986 (Section 3.2.3)
//
// Parameters:
//   - port: The port string containing only numeric ASCII characters.
//   - hasPortDelimiter: A boolean indicating if the port delimiter (colon) was present.
//
// Returns:
//   - error: An error if any character is non-numeric, nil otherwise.
func (p *iriParser) parsePort(port string, hasPortDelimiter bool) error {
	if !hasPortDelimiter {
		return nil
	}
	if !p.unchecked {
		for _, r := range port {
			if !isASCIIDigit(r) {
				return &kindError{message: "Invalid port character", char: r}
			}
		}
	}
	// Spec Rule: RFC 3986 (Section 3.2.3)
	// Write the colon delimiter unconditionally to the output buffer to preserve empty ports during initial parsing.
	p.output.writeRune(':')
	p.output.writeString(port)
	return nil
}

// parseAuthority parses, validates, and splits the authority component into userinfo, host, and port.
//
// Specification Reference:
// RFC 3987 (Section 2.2), RFC 3986 (Section 3.2)
//
// Parameters:
//
// Returns:
//   - error: An error if any subcomponent validation fails, nil otherwise.
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

	userinfo, host, port, hasPortDelimiter := splitAuthority(authorityPart)

	if err := p.parseUserinfo(userinfo); err != nil {
		return err
	}
	if err := p.parseHost(host); err != nil {
		return err
	}
	if err := p.parsePort(port, hasPortDelimiter); err != nil {
		return err
	}

	p.input.reset(authorityStr[end:])
	p.outputPositions.AuthorityEnd = p.output.len()

	return nil
}

// validateIPLiteral checks if the enclosed IP literal string is a valid IPv6 or IPvFuture format.
//
// Specification Reference:
// RFC 3986 (Section 3.2.2)
//
// Parameters:
//   - ipLiteral: The string enclosed in square brackets.
//
// Returns:
// - error: An error if parsing fails, if the literal is a bracketed IPv4 address, or formatting is invalid, nil
// otherwise.
func (p *iriParser) validateIPLiteral(ipLiteral string) error {
	if strings.HasPrefix(ipLiteral, "v") || strings.HasPrefix(ipLiteral, "V") {
		return p.validateIPVFuture(ipLiteral)
	}

	ip := net.ParseIP(ipLiteral)
	// Spec Rule: RFC 3986 (Section 3.2.2)
	// IP-literal can only represent an IPv6 or IPvFuture address. Enclosing an IPv4 address
	// in square brackets (e.g., [192.168.0.1]) is structurally invalid.
	if ip == nil || ip.To4() != nil {
		return &kindError{message: "Invalid host IP", details: ipLiteral}
	}
	return nil
}

// validateIPVFuture validates the structural components of an IPvFuture literal identifier.
//
// Specification Reference:
// RFC 3986 (Section 3.2.2)
//
// Parameters:
//   - ip: The IPvFuture address string containing version and sub-delims characters.
//
// Returns:
//   - error: An error if format violations are encountered, nil otherwise.
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

// splitAuthority decomposes a raw authority string into its userinfo, host, and port components.
//
// Specification Reference:
// RFC 3986 (Section 3.2)
//
// Parameters:
//   - authority: The raw authority string to segment.
//
// Returns:
//   - string: The extracted userinfo subcomponent.
//   - string: The extracted host subcomponent.
//   - string: The extracted port subcomponent.
//   - bool: True if the authority contains a trailing port colon delimiter, false otherwise.
func splitAuthority(authority string) (string, string, string, bool) {
	var userinfo, host, port string
	var hasPortDelimiter bool

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
			return userinfo, host, port, false
		}
		host = hostport[:endBracket+1]
		if len(hostport) > endBracket+1 && hostport[endBracket+1] == ':' {
			port = hostport[endBracket+2:]
			hasPortDelimiter = true
		}
		return userinfo, host, port, hasPortDelimiter
	}

	endHost := strings.LastIndex(hostport, ":")
	if endHost != -1 {
		host = hostport[:endHost]
		port = hostport[endHost+1:]
		hasPortDelimiter = true
	} else {
		host = hostport
	}
	return userinfo, host, port, hasPortDelimiter
}
