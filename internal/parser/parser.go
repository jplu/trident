/*
Copyright 2025 Trident Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUTHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package parser provides a fast, zero-allocation validating IRI (Internationalized
// Resource Identifier) parser compliant with RFC 3987. It is designed for
// high-performance scenarios where validation and resolution of IRIs are critical.
//
// The main entry point is the Run function, which processes an IRI string. The parser
// can operate in two modes: strict validation or a faster "unchecked" mode.
// It supports resolving relative IRI references against a base IRI, following the
// algorithm defined in RFC 3986.
//
// A key feature is the use of an OutputBuffer interface, which allows the parsed
// and normalized IRI to be written to different destinations. This enables a
// "validation-only" mode (using VoidOutputBuffer) that avoids string allocations
// entirely, making it suitable for simply checking IRI validity.
package parser

import (
	"fmt"
	"io"
	"net"
	"strings"
	"unicode"
)

const (
	// authorityPrefixLength is the length of the string "//".
	authorityPrefixLength = 2
	// ipvFutureParts is the number of parts expected in an IPvFuture literal
	// (e.g., "v1.abc"), separated by a dot.
	ipvFutureParts = 2
)

var (
	// ErrNoScheme is returned when an absolute IRI is expected but no scheme
	// (e.g., "http:") is found. This typically occurs when the IRI string
	// starts with a colon, which is invalid.
	ErrNoScheme = &kindError{message: "No scheme found in an absolute IRI"}
	// ErrPathStartingWithSlashes is returned when an IRI has a path that
	// starts with "//" but does not have an authority component. This is
	// disallowed by RFC 3987 to avoid ambiguity with network-path references.
	// For example, "scheme:////path" is valid, but "scheme:/path" where the path
	// starts with `//` is not.
	ErrPathStartingWithSlashes = &kindError{
		message: "An IRI path is not allowed to start with // if there is no authority",
	}
)

// kindError is a specialized error type used by the parser to provide
// detailed context about a parsing failure.
type kindError struct {
	message string
	char    rune
	details string
	err     error
}

// Error formats the error message with any available character, details, or
// wrapped error information.
func (e *kindError) Error() string {
	msg := e.message
	if e.char != 0 {
		msg = fmt.Sprintf("%s '%c'", msg, e.char)
	}
	if e.details != "" {
		msg = fmt.Sprintf("%s '%s'", msg, e.details)
	}
	if e.err != nil {
		msg = fmt.Sprintf("%s (%v)", msg, e.err)
	}
	return msg
}

// Unwrap provides compatibility with Go's standard errors package, allowing
// access to the underlying error if one exists.
func (e *kindError) Unwrap() error {
	return e.err
}

// Positions holds the end indices of the various components of a parsed IRI.
// Each index represents the offset in the *output* buffer where the component
// ends. The start of a component is the end of the previous one. An index is
// exclusive, i.e., it is the index of the first character *after* the component.
//
// For example, for the IRI "http://a/b?c#d":
//   - Scheme:    output[0:SchemeEnd] -> "http:"
//   - Authority: output[SchemeEnd:AuthorityEnd] -> "//a"
//   - Path:      output[AuthorityEnd:PathEnd] -> "/b"
//   - Query:     output[PathEnd:QueryEnd] -> "?c"
//   - Fragment:  output[QueryEnd:len(output)] -> "#d"
//
// If a component is absent, its end index will be the same as the previous
// component's end index. For example, in "http:path", AuthorityEnd will
// equal SchemeEnd.
type Positions struct {
	SchemeEnd    int
	AuthorityEnd int
	PathEnd      int
	QueryEnd     int
}

// Base represents a pre-parsed, absolute IRI that can be used as a base for
// resolving relative references. It contains the original IRI string and the
// positions of its components, which must be pre-calculated.
type Base struct {
	IRI string
	Pos Positions
}

// OutputBuffer is an interface for building the output string during parsing.
// This abstraction allows the parser to be used in different modes, such as
// full string generation (StringOutputBuffer) or validation-only without
// string allocation (VoidOutputBuffer).
type OutputBuffer interface {
	// WriteRune appends a single rune to the buffer.
	WriteRune(r rune)
	// WriteString appends a string to the buffer.
	WriteString(s string)
	// String returns the complete content of the buffer.
	String() string
	// Len returns the number of bytes currently in the buffer.
	Len() int
	// Truncate reduces the buffer to n bytes.
	Truncate(n int)
	// Reset clears the buffer.
	Reset()
}

// VoidOutputBuffer is an implementation of OutputBuffer that discards all
// writes and only tracks the length of the would-be output. It is useful for
// validation-only parsing where the final string is not needed, avoiding
// all allocations.
type VoidOutputBuffer struct {
	len int
}

func (b *VoidOutputBuffer) WriteRune(r rune)     { b.len += len(string(r)) }
func (b *VoidOutputBuffer) WriteString(s string) { b.len += len(s) }
func (b *VoidOutputBuffer) String() string       { return "" }
func (b *VoidOutputBuffer) Len() int             { return b.len }
func (b *VoidOutputBuffer) Truncate(n int)       { b.len = n }
func (b *VoidOutputBuffer) Reset()               { b.len = 0 }

// StringOutputBuffer is an implementation of OutputBuffer that uses a
// strings.Builder to efficiently construct the output string.
type StringOutputBuffer struct {
	Builder *strings.Builder
}

func (b *StringOutputBuffer) WriteRune(r rune)     { b.Builder.WriteRune(r) }
func (b *StringOutputBuffer) WriteString(s string) { b.Builder.WriteString(s) }
func (b *StringOutputBuffer) String() string       { return b.Builder.String() }
func (b *StringOutputBuffer) Len() int             { return b.Builder.Len() }
func (b *StringOutputBuffer) Truncate(n int) {
	if n < 0 || n > b.Builder.Len() {
		return
	}
	s := b.Builder.String()[:n]
	b.Builder.Reset()
	b.Builder.WriteString(s)
}
func (b *StringOutputBuffer) Reset() { b.Builder.Reset() }

// Run is the main entry point for the IRI parser. It parses, validates, and
// resolves an IRI reference against an optional base IRI. The result is written
// to the provided output buffer, and the component positions are returned.
//
// Parameters:
//   - iri: The IRI reference string to be parsed. This can be an absolute IRI
//     (e.g., "https://example.com/a") or a relative reference (e.g., "../b").
//   - base: An optional pre-parsed base IRI. If provided, the parser will
//     resolve the `iri` reference against it according to RFC 3986. If nil,
//     the `iri` is parsed as a standalone reference.
//   - unchecked: A boolean flag to control validation. If true, the parser
//     skips most character-level validation, significantly improving performance.
//     This is useful when the input is known to be well-formed. If false, the
//     parser performs strict validation against RFC 3987 rules.
//   - output: A buffer where the parsed (and potentially resolved and normalized)
//     IRI is written. Use a StringOutputBuffer to get the resulting string, or
//     a VoidOutputBuffer for validation-only runs without string allocation.
//
// Returns:
//   - Positions: The end indices of the parsed components in the `output` buffer.
//   - error: An error if parsing or validation fails. Nil on success.
func Run(iri string, base *Base, unchecked bool, output OutputBuffer) (Positions, error) {
	var b *iriParserBase
	if base != nil {
		b = &iriParserBase{
			iri:          base.IRI,
			schemeEnd:    base.Pos.SchemeEnd,
			authorityEnd: base.Pos.AuthorityEnd,
			pathEnd:      base.Pos.PathEnd,
			queryEnd:     base.Pos.QueryEnd,
			hasBase:      true,
		}
	} else {
		b = &iriParserBase{hasBase: false}
	}

	p := &iriParser{
		iri:       iri,
		base:      b,
		input:     newParserInput(iri),
		output:    output,
		unchecked: unchecked,
	}

	err := p.parseSchemeStart()
	return p.outputPositions, err
}

// iriParserBase holds the component data of a base IRI used for resolution.
type iriParserBase struct {
	iri          string
	schemeEnd    int
	authorityEnd int
	pathEnd      int
	queryEnd     int
	hasBase      bool
}

// iriParser holds the state for a single parsing operation.
type iriParser struct {
	iri             string         // The original input IRI string.
	base            *iriParserBase // The pre-parsed base IRI, if any.
	input           *parserInput   // The input reader for the IRI string.
	output          OutputBuffer   // The output buffer for the parsed result.
	outputPositions Positions      // The calculated positions of components in the output.
	inputSchemeEnd  int            // The position in the input string where the scheme ends.
	unchecked       bool           // Flag to disable strict validation.
}

// parseSchemeStart is the initial state of the parser. It determines whether
// the IRI is absolute, relative, or a network-path reference.
func (p *iriParser) parseSchemeStart() error {
	// A network-path reference (e.g., "//example.com/path") is a relative reference
	// that begins with a double slash. It's handled specially.
	if !p.base.hasBase && strings.HasPrefix(p.iri, "//") {
		if _, err := p.input.reader.Seek(authorityPrefixLength, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek past '//' prefix: %w", err)
		}
		p.output.WriteString("//")
		p.outputPositions.SchemeEnd = 0
		return p.parseAuthority()
	}

	r, ok := p.input.peek()
	if !ok {
		// Empty IRI is treated as a relative reference.
		return p.parseRelative()
	}
	if r == ':' {
		if p.unchecked {
			// In unchecked mode, we might try to parse it as a scheme,
			// but a valid scheme cannot start with ':'. Here we treat it as an
			// absolute IRI with an empty scheme part, which will later be handled.
			return p.parseScheme()
		}
		// An IRI cannot start with a colon.
		return ErrNoScheme
	}
	if isASCIILetter(r) {
		// If it starts with a letter, it must be a scheme.
		return p.parseScheme()
	}
	// Otherwise, it's a relative-path reference.
	return p.parseRelative()
}

// parseScheme consumes the scheme component (e.g., "http"). It reads until
// a colon is found. If other characters are encountered, it backtracks and
// treats the input as a relative path.
func (p *iriParser) parseScheme() error {
	initialInput := p.iri
	initialPos := p.input.position()
	for {
		r, ok := p.input.next()
		if !ok {
			// Reached end of string without finding ':', so it's a relative path.
			p.input.reset(initialInput[initialPos:])
			p.output.Reset()
			return p.parseRelative()
		}

		switch {
		case isASCIILetter(r) || isASCIIDigit(r) || r == '+' || r == '-' || r == '.':
			p.output.WriteRune(r)
		case r == ':':
			p.output.WriteRune(':')
			p.outputPositions.SchemeEnd = p.output.Len()
			p.inputSchemeEnd = p.input.position()
			// After the scheme, check for the slashes that might introduce an authority.
			if p.input.startsWith('/') {
				p.input.next()
				p.output.WriteRune('/')
				return p.parsePathOrAuthority()
			}
			// No slashes, so no authority component.
			p.outputPositions.AuthorityEnd = p.outputPositions.SchemeEnd
			return p.parsePath()
		default:
			// Invalid character for a scheme, so it must be a relative path.
			p.input.reset(initialInput[initialPos:])
			p.output.Reset()
			return p.parseRelative()
		}
	}
}

// parsePathOrAuthority handles the part of the IRI after "scheme:/". If another
// slash follows, it's an authority; otherwise, it's the start of the path.
func (p *iriParser) parsePathOrAuthority() error {
	if p.input.startsWith('/') {
		p.input.next()
		p.output.WriteRune('/')
		return p.parseAuthority()
	}
	// Only one slash, so no authority.
	p.outputPositions.AuthorityEnd = p.outputPositions.SchemeEnd
	return p.parsePath()
}

// isValidRefScheme checks if a given string is a valid scheme component.
func isValidRefScheme(schemePart string) bool {
	if len(schemePart) == 0 {
		return false
	}
	if !isASCIILetter(rune(schemePart[0])) {
		return false
	}
	for i := 1; i < len(schemePart); i++ {
		r := rune(schemePart[i])
		if !isASCIILetter(r) && !isASCIIDigit(r) && r != '+' && r != '-' && r != '.' {
			return false
		}
	}
	return true
}

// extractRefScheme attempts to extract a scheme from the beginning of a reference string.
func extractRefScheme(ref string) (string, string, bool) {
	i := strings.Index(ref, ":")
	if i <= 0 {
		return "", ref, false
	}

	schemePart := ref[:i]
	if !isValidRefScheme(schemePart) {
		return "", ref, false
	}

	// The colon must appear before any slash in the reference.
	firstSlash := strings.Index(ref, "/")
	if firstSlash != -1 && i > firstSlash {
		return "", ref, false
	}

	return schemePart, ref[i+1:], true
}

// deconstructRef breaks a relative reference string into its constituent parts.
// This is a pre-processing step for reference resolution.
func deconstructRef(ref string) (
	string, string, string, string, string,
	bool, bool, bool,
) {
	var scheme, authority, path, query, fragment string
	var hasAuthority, hasQuery, hasFragment bool

	if i := strings.Index(ref, "#"); i != -1 {
		hasFragment = true
		fragment = ref[i+1:]
		ref = ref[:i]
	}
	if i := strings.Index(ref, "?"); i != -1 {
		hasQuery = true
		query = ref[i+1:]
		ref = ref[:i]
	}

	scheme, ref, _ = extractRefScheme(ref)

	if strings.HasPrefix(ref, "//") {
		hasAuthority = true
		ref = ref[2:]
		endAuth := strings.Index(ref, "/")
		if endAuth == -1 {
			authority = ref
			path = ""
		} else {
			authority = ref[:endAuth]
			path = ref[endAuth:]
		}
	} else {
		path = ref
	}
	return scheme, authority, path, query, fragment, hasAuthority, hasQuery, hasFragment
}

// resolvedIRI holds the components of an IRI after reference resolution.
type resolvedIRI struct {
	Scheme       string
	Authority    string
	Path         string
	Query        string
	Fragment     string
	HasAuthority bool
	HasQuery     bool
	HasFragment  bool
}

// resolvePathAndQuery handles the path and query resolution logic from RFC 3986, Section 5.2.2.
func (p *iriParser) resolvePathAndQuery(
	t *resolvedIRI,
	rPath, rQuery string,
	rHasQuery bool,
	basePath, baseQuery string,
	hasBaseQuery, hasBaseAuthority bool,
) {
	if rPath != "" {
		if strings.HasPrefix(rPath, "/") {
			t.Path = removeDotSegments(rPath)
		} else {
			mergePath := basePath
			if mergePath == "" && hasBaseAuthority {
				mergePath = "/"
			}
			t.Path = resolvePath(mergePath, rPath)
		}
		t.Query = rQuery
		t.HasQuery = rHasQuery
		return
	}

	t.Path = basePath
	if rHasQuery {
		t.Query = rQuery
		t.HasQuery = true
	} else {
		t.Query = baseQuery
		t.HasQuery = hasBaseQuery
	}
}

// resolveComponents implements the reference resolution algorithm from RFC 3986, Section 5.2.
func (p *iriParser) resolveComponents(relativeRef string) *resolvedIRI {
	rScheme, rAuthority, rPath, rQuery, rFragment, rHasAuthority, rHasQuery, rHasFragment := deconstructRef(relativeRef)

	baseScheme, baseAuthority, basePath, hasBaseAuthority, baseQuery, hasBaseQuery := p.getBaseComponents()

	t := &resolvedIRI{Fragment: rFragment, HasFragment: rHasFragment}

	if rScheme != "" {
		t.Scheme = rScheme
		t.Authority = rAuthority
		t.HasAuthority = rHasAuthority
		t.Path = removeDotSegments(rPath)
		t.Query = rQuery
		t.HasQuery = rHasQuery
	} else {
		t.Scheme = baseScheme
		if rHasAuthority {
			t.Authority = rAuthority
			t.HasAuthority = true
			t.Path = removeDotSegments(rPath)
			t.Query = rQuery
			t.HasQuery = rHasQuery
		} else {
			p.resolvePathAndQuery(t, rPath, rQuery, rHasQuery, basePath, baseQuery, hasBaseQuery, hasBaseAuthority)
			t.Authority = baseAuthority
			t.HasAuthority = hasBaseAuthority
		}
	}
	return t
}

// getBaseComponents extracts the components from the base IRI for resolution.
func (p *iriParser) getBaseComponents() (string, string, string, bool, string, bool) {
	var scheme, authority, path, query string
	var hasAuthority, hasQuery bool
	base := p.base

	if base.schemeEnd > 0 {
		scheme = base.iri[:base.schemeEnd-1]
	}
	if base.authorityEnd > base.schemeEnd {
		hasAuthority = true
		start := base.schemeEnd
		if strings.HasPrefix(base.iri[start:], "//") {
			start += 2
		}
		if base.authorityEnd > start {
			authority = base.iri[start:base.authorityEnd]
		}
	}
	path = base.iri[base.authorityEnd:base.pathEnd]
	if base.queryEnd > base.pathEnd {
		query = base.iri[base.pathEnd+1 : base.queryEnd]
		hasQuery = true
	}
	return scheme, authority, path, hasAuthority, query, hasQuery
}

// recomposeIRI assembles the final IRI from its resolved components into the output buffer.
func (p *iriParser) recomposeIRI(t *resolvedIRI) {
	if t.Scheme != "" {
		p.output.WriteString(t.Scheme)
		p.output.WriteRune(':')
	}
	p.outputPositions.SchemeEnd = p.output.Len()

	if t.HasAuthority {
		p.output.WriteString("//")
		p.output.WriteString(t.Authority)
	}
	p.outputPositions.AuthorityEnd = p.output.Len()

	p.output.WriteString(t.Path)
	p.outputPositions.PathEnd = p.output.Len()

	if t.HasQuery {
		p.output.WriteRune('?')
		p.output.WriteString(t.Query)
	}
	p.outputPositions.QueryEnd = p.output.Len()

	if t.HasFragment {
		p.output.WriteRune('#')
		p.output.WriteString(t.Fragment)
	}
}

// parseRelativeNoBase handles parsing a relative reference when no base IRI is provided.
// In this case, it's parsed as a relative-path reference.
func (p *iriParser) parseRelativeNoBase() error {
	p.outputPositions.SchemeEnd = 0
	p.inputSchemeEnd = 0
	if p.input.startsWith('/') {
		p.input.next()
		p.output.WriteRune('/')
		// Path cannot start with "//" if there is no authority.
		if p.input.startsWith('/') {
			return ErrPathStartingWithSlashes
		}
		return p.parsePath()
	}

	return p.parsePathNoScheme()
}

// validateRelativeRef runs a sub-parse on the relative reference string to ensure it's well-formed.
func (p *iriParser) validateRelativeRef(relativeRef string) error {
	if p.unchecked {
		return nil
	}
	validationParser := &iriParser{
		iri:       relativeRef,
		base:      &iriParserBase{hasBase: false},
		input:     newParserInput(relativeRef),
		output:    &VoidOutputBuffer{},
		unchecked: false,
	}
	return validationParser.parseSchemeStart()
}

// parseRelative handles a relative IRI reference. If a base IRI is present,
// it resolves the reference against the base. Otherwise, it parses it as a
// relative-path reference.
func (p *iriParser) parseRelative() error {
	if !p.base.hasBase {
		return p.parseRelativeNoBase()
	}

	relativeRef := p.input.asStr()
	if err := p.validateRelativeRef(relativeRef); err != nil {
		return err
	}

	t := p.resolveComponents(relativeRef)

	if !t.HasAuthority && strings.HasPrefix(t.Path, "//") {
		return ErrPathStartingWithSlashes
	}

	p.recomposeIRI(t)
	return nil
}

// parseAuthority parses the authority component, which includes userinfo, host, and port.
func (p *iriParser) parseAuthority() error {
	loopInput := p.input.asStr()
	atIndex := strings.Index(loopInput, "@")
	slashIndex := strings.Index(loopInput, "/")
	// The '@' for userinfo must appear before the host ends.
	if atIndex != -1 && (slashIndex == -1 || atIndex < slashIndex) {
		userinfo := loopInput[:atIndex]
		for _, r := range userinfo {
			if err := p.readURLCodepointOrEchar(r, func(c rune) bool {
				return isIUnreservedOrSubDelims(c) || c == ':'
			}); err != nil {
				return err
			}
		}
		p.output.WriteRune('@')
		p.input.reset(loopInput[atIndex+1:])
	}
	return p.parseHost()
}

// parseHost dispatches to the correct host parser (IP literal or regular name).
func (p *iriParser) parseHost() error {
	if p.input.startsWith('[') {
		return p.parseHostIPV6()
	}
	return p.parseHostRegular()
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

// parseHostIPV6 parses an IP literal host (e.g., "[::1]" or "[v1.a]").
func (p *iriParser) parseHostIPV6() error {
	startPos := p.input.position()
	p.input.next()
	p.output.WriteRune('[')
	for {
		r, ok := p.input.next()
		if !ok {
			return &kindError{message: "Invalid host character: unterminated IPv6 literal"}
		}
		p.output.WriteRune(r)
		if r == ']' {
			ipLiteral := p.iri[startPos+1 : p.input.position()-1]
			if !p.unchecked {
				if err := p.validateIPLiteral(ipLiteral); err != nil {
					return err
				}
			}
			return p.parseHostPortEnd()
		}
	}
}

// parseHostRegular parses a registered name or IPv4 address as a host.
func (p *iriParser) parseHostRegular() error {
	for {
		r, ok := p.input.peek()
		if !ok {
			// End of string, authority ends here.
			p.outputPositions.AuthorityEnd = p.output.Len()
			return p.parsePathStart(0, false)
		}
		if r == ':' {
			p.input.next()
			p.output.WriteRune(':')
			return p.parsePort()
		}
		if r == '/' || r == '?' || r == '#' {
			// Start of path, query, or fragment means host is done.
			p.outputPositions.AuthorityEnd = p.output.Len()
			return p.parsePathStart(r, true)
		}
		p.input.next()

		if err := p.readURLCodepointOrEchar(r, func(c rune) bool {
			return isIUnreservedOrSubDelims(c) || c == '.'
		}); err != nil {
			return err
		}
	}
}

// parseHostPortEnd handles the characters that can legally follow an IP literal.
func (p *iriParser) parseHostPortEnd() error {
	r, ok := p.input.peek()
	if ok && r == ':' {
		p.input.next()
		p.output.WriteRune(':')
		return p.parsePort()
	}
	// After an IP literal, only ':', '/', '?', '#', or EOF are allowed.
	if ok && !p.unchecked && r != '/' && r != '?' && r != '#' {
		return &kindError{message: "Invalid character after IP literal", char: r}
	}
	p.outputPositions.AuthorityEnd = p.output.Len()
	return p.parsePathStart(r, ok)
}

// parsePort consumes the port number.
func (p *iriParser) parsePort() error {
	for {
		r, ok := p.input.peek()
		if !ok {
			// End of string, authority ends here.
			p.outputPositions.AuthorityEnd = p.output.Len()
			return p.parsePathStart(0, false)
		}

		if r == '/' || r == '?' || r == '#' {
			// Port is done, authority ends here.
			p.outputPositions.AuthorityEnd = p.output.Len()
			return p.parsePathStart(r, true)
		}

		p.input.next()
		if !p.unchecked && !isASCIIDigit(r) {
			return &kindError{message: "Invalid port character", char: r}
		}
		p.output.WriteRune(r)
	}
}

// parsePathStart begins parsing the path component, handling the first character
// which might also signal the start of a query or fragment.
func (p *iriParser) parsePathStart(r rune, ok bool) error {
	if !ok {
		// End of IRI.
		p.outputPositions.PathEnd = p.output.Len()
		p.outputPositions.QueryEnd = p.output.Len()
		return nil
	}
	p.input.next()
	switch r {
	case '?':
		p.outputPositions.PathEnd = p.output.Len()
		p.output.WriteRune('?')
		return p.parseQuery()
	case '#':
		p.outputPositions.PathEnd = p.output.Len()
		p.outputPositions.QueryEnd = p.output.Len()
		p.output.WriteRune('#')
		return p.parseFragment()
	case '/':
		p.output.WriteRune('/')
		return p.parsePath()
	default:
		// First character of a path segment.
		if err := p.readURLCodepointOrEchar(r, func(c rune) bool {
			return isIUnreservedOrSubDelims(c) || c == ':' || c == '@'
		}); err != nil {
			return err
		}
		return p.parsePath()
	}
}

// parsePathNoScheme parses a path that is not preceded by a scheme. The first
// segment of such a path cannot contain a colon.
func (p *iriParser) parsePathNoScheme() error {
	// First segment cannot contain a ':'
	for {
		c, ok := p.input.peek()
		if !ok || c == '/' || c == '?' || c == '#' {
			break
		}
		if c == ':' {
			// This would be ambiguous with a scheme.
			return &kindError{message: "Invalid IRI character in first path segment", char: c}
		}
		p.input.next()
		if err := p.readURLCodepointOrEchar(c, func(r rune) bool {
			return isIUnreservedOrSubDelims(r) || r == '@'
		}); err != nil {
			return err
		}
	}

	return p.parsePath()
}

// parsePath consumes the path component of the IRI.
func (p *iriParser) parsePath() error {
	for {
		c, ok := p.input.peek()
		if !ok {
			break // End of IRI.
		}
		if c == '?' {
			p.input.next()
			p.outputPositions.PathEnd = p.output.Len()
			p.output.WriteRune('?')
			return p.parseQuery()
		}
		if c == '#' {
			p.input.next()
			p.outputPositions.PathEnd = p.output.Len()
			p.outputPositions.QueryEnd = p.output.Len()
			p.output.WriteRune('#')
			return p.parseFragment()
		}
		p.input.next()
		if err := p.readURLCodepointOrEchar(c, func(c rune) bool {
			return isIUnreservedOrSubDelims(c) || c == ':' || c == '@' || c == '/'
		}); err != nil {
			return err
		}
	}
	p.outputPositions.PathEnd = p.output.Len()
	p.outputPositions.QueryEnd = p.output.Len()
	return nil
}

// parseQuery consumes the query component.
func (p *iriParser) parseQuery() error {
	for {
		r, ok := p.input.peek()
		if !ok {
			p.outputPositions.QueryEnd = p.output.Len()
			return nil
		}
		if r == '#' {
			p.input.next()
			p.outputPositions.QueryEnd = p.output.Len()
			p.output.WriteRune('#')
			return p.parseFragment()
		}
		p.input.next()
		err := p.readURLCodepointOrEchar(r, func(c rune) bool {
			return isIUnreservedOrSubDelims(c) || c == ':' || c == '@' || c == '/' || c == '?' ||
				// iprivate characters
				(c >= '\uE000' && c <= '\uF8FF') ||
				(c >= 0xF0000 && c <= 0xFFFFD) ||
				(c >= 0x100000 && c <= 0x10FFFD)
		})
		if err != nil {
			return err
		}
	}
}

// parseFragment consumes the fragment component.
func (p *iriParser) parseFragment() error {
	for {
		r, ok := p.input.next()
		if !ok {
			return nil
		}
		err := p.readURLCodepointOrEchar(r, func(c rune) bool {
			return isIUnreservedOrSubDelims(c) || c == ':' || c == '@' || c == '/' || c == '?'
		})
		if err != nil {
			return err
		}
	}
}

// readURLCodepointOrEchar processes a single rune. If it's a '%' it handles
// percent-encoding. Otherwise, it validates the rune against the provided
// function and writes it to the output.
func (p *iriParser) readURLCodepointOrEchar(r rune, valid func(rune) bool) error {
	if r == '%' {
		return p.readEchar()
	}
	if !p.unchecked && !valid(r) {
		return &kindError{message: "Invalid IRI character", char: r}
	}
	p.output.WriteRune(r)
	return nil
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
	p.output.WriteRune('%')
	p.output.WriteRune(c1)
	p.output.WriteRune(c2)
	return nil
}

// validateIPVFuture validates an IPvFuture literal (e.g., "v1.something").
func (p *iriParser) validateIPVFuture(ip string) error {
	if !strings.HasPrefix(ip, "v") && !strings.HasPrefix(ip, "V") {
		return &kindError{message: "Invalid IPvFuture format", details: ip}
	}
	parts := strings.SplitN(ip[1:], ".", ipvFutureParts)
	if len(parts) != ipvFutureParts {
		return &kindError{message: "Invalid IPvFuture format: no dot separator", details: ip}
	}
	version := parts[0]
	if version == "" {
		return &kindError{message: "Invalid IPvFuture: missing version", details: ip}
	}
	for _, r := range version {
		if !isASCIIHexDigit(r) {
			return &kindError{message: "Invalid IPvFuture version char", char: r}
		}
	}
	if parts[1] == "" {
		return &kindError{message: "Invalid IPvFuture: empty address part", details: ip}
	}
	for _, r := range parts[1] {
		if !isUnreservedOrSubDelims(r) && r != ':' {
			return &kindError{message: "Invalid IPvFuture address char", char: r}
		}
	}
	return nil
}

// extractAndAppendSegment is a helper for removeDotSegments. It extracts the
// next path segment from the input string and appends it to the output slice.
func extractAndAppendSegment(in string, output []string, isAbsolute bool) (string, []string) {
	startsWithSlash := strings.HasPrefix(in, "/")
	var i int
	if startsWithSlash {
		i = strings.Index(in[1:], "/")
		if i != -1 {
			i++
		}
	} else {
		i = strings.Index(in, "/")
	}

	var segment, nextIn string
	if i == -1 {
		segment = in
		nextIn = ""
	} else {
		segment = in[:i]
		nextIn = in[i:]
	}

	if len(output) == 0 && !isAbsolute && startsWithSlash {
		if segment != "/" {
			segment = segment[1:]
		}
	}

	if segment != "" {
		output = append(output, segment)
	}
	return nextIn, output
}

// processOneStepOfDotRemoval is a helper for removeDotSegments that processes
// one of the five cases for dot-segment removal from RFC 3986, Section 5.2.4.
func processOneStepOfDotRemoval(in string, output []string, isAbsolute bool) (string, []string) {
	if strings.HasPrefix(in, "../") { // Case B
		return in[3:], output
	}
	if strings.HasPrefix(in, "./") { // Case B
		return in[2:], output
	}
	if strings.HasPrefix(in, "/./") { // Case C
		return "/" + in[3:], output
	}
	if in == "/." { // Case C
		return "/", output
	}
	if strings.HasPrefix(in, "/../") || in == "/.." { // Case D
		if len(output) > 0 {
			output = output[:len(output)-1]
		}
		if in == "/.." {
			return "/", output
		}
		return "/" + in[4:], output
	}
	if in == "." || in == ".." { // Case E
		return "", output
	}
	// Case A: No dot segments found, extract next path segment.
	return extractAndAppendSegment(in, output, isAbsolute)
}

// removeDotSegments implements the "Remove Dot Segments" algorithm from
// RFC 3986, Section 5.2.4. It normalizes a path by resolving "." and ".."
// segments.
func removeDotSegments(input string) string {
	var output []string
	in := input
	isAbsolute := strings.HasPrefix(in, "/")

	for len(in) > 0 {
		in, output = processOneStepOfDotRemoval(in, output, isAbsolute)
	}

	result := strings.Join(output, "")
	if isAbsolute && result == "" {
		return "/"
	}
	return result
}

// resolvePath resolves a relative path against a base path according to
// RFC 3986, Section 5.2.2. It merges the base path with the relative
// reference path.
func resolvePath(basePath, relPath string) string {
	if basePath == "" {
		return removeDotSegments(relPath)
	}
	lastSlash := strings.LastIndex(basePath, "/")
	if lastSlash == -1 {
		// Base path has no slashes, e.g. "a"
		return removeDotSegments(relPath)
	}
	// Merge from the last slash.
	return removeDotSegments(basePath[:lastSlash+1] + relPath)
}

// parserInput provides a reader-like interface over the input string,
// allowing for peeking, advancing, and position tracking.
type parserInput struct {
	originalString string
	reader         *strings.Reader
}

func newParserInput(s string) *parserInput {
	return &parserInput{
		originalString: s,
		reader:         strings.NewReader(s),
	}
}
func (p *parserInput) next() (rune, bool) {
	r, _, err := p.reader.ReadRune()
	return r, err == nil
}
func (p *parserInput) peek() (rune, bool) {
	r, _, err := p.reader.ReadRune()
	if err != nil {
		return 0, false
	}
	_ = p.reader.UnreadRune()
	return r, true
}
func (p *parserInput) startsWith(r rune) bool {
	pr, ok := p.peek()
	return ok && pr == r
}
func (p *parserInput) position() int {
	return len(p.originalString) - p.reader.Len()
}
func (p *parserInput) asStr() string {
	return p.originalString[p.position():]
}
func (p *parserInput) reset(s string) {
	p.originalString = s
	p.reader = strings.NewReader(s)
}

// isASCIILetter checks if a rune is an ASCII letter.
func isASCIILetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

// isASCIIDigit checks if a rune is an ASCII digit.
func isASCIIDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

// isASCIIHexDigit checks if a rune is an ASCII hexadecimal digit.
func isASCIIHexDigit(r rune) bool {
	return isASCIIDigit(r) || ('a' <= unicode.ToLower(r) && unicode.ToLower(r) <= 'f')
}

// isIUnreservedOrSubDelims checks if a character is in the iunreserved or
// sub-delims sets as defined by RFC 3987. This includes unreserved US-ASCII
// characters and additional Unicode ranges.
func isIUnreservedOrSubDelims(c rune) bool {
	if isUnreservedOrSubDelims(c) {
		return true
	}

	// This is the "iunreserved" character set from RFC 3987.
	switch {
	case c >= '\u00A0' && c <= '\uD7FF',
		c >= '\uF900' && c <= '\uFDCF',
		c >= '\uFDF0' && c <= '\uFFEF',
		c >= 0x10000 && c <= 0x1FFFD,
		c >= 0x20000 && c <= 0x2FFFD,
		c >= 0x30000 && c <= 0x3FFFD,
		c >= 0x40000 && c <= 0x4FFFD,
		c >= 0x50000 && c <= 0x5FFFD,
		c >= 0x60000 && c <= 0x6FFFD,
		c >= 0x70000 && c <= 0x7FFFD,
		c >= 0x80000 && c <= 0x8FFFD,
		c >= 0x90000 && c <= 0x9FFFD,
		c >= 0xA0000 && c <= 0xAFFFD,
		c >= 0xB0000 && c <= 0xBFFFD,
		c >= 0xC0000 && c <= 0xCFFFD,
		c >= 0xD0000 && c <= 0xDFFFD,
		c >= 0xE1000 && c <= 0xEFFFD:
		return true
	}
	return false
}

// isUnreservedOrSubDelims checks if a character is in the unreserved or
// sub-delims sets as defined by RFC 3986 (US-ASCII only).
func isUnreservedOrSubDelims(c rune) bool {
	return isASCIILetter(c) || isASCIIDigit(c) || strings.ContainsRune("!$&'()*+,-.;=_~", c)
}
