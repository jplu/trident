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

import "strings"

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

// isValidRefScheme checks if a given string is a valid scheme component.
func isValidRefScheme(schemePart string) bool {
	if len(schemePart) == 0 || !isASCIILetter(rune(schemePart[0])) {
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
	if i < 0 {
		return "", ref, false
	}

	schemePart := ref[:i]
	if !isValidRefScheme(schemePart) {
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

	// RFC 3986, Section 5.2.2: If the reference has a scheme, it is treated as absolute.
	if rScheme != "" {
		return &resolvedIRI{
			Scheme:       rScheme,
			Authority:    rAuthority,
			Path:         removeDotSegments(rPath),
			Query:        rQuery,
			Fragment:     rFragment,
			HasAuthority: rHasAuthority,
			HasQuery:     rHasQuery,
			HasFragment:  rHasFragment,
		}
	}

	baseScheme, baseAuthority, basePath, hasBaseAuthority, baseQuery, hasBaseQuery := p.getBaseComponents()

	t := &resolvedIRI{
		Fragment:    rFragment,
		HasFragment: rHasFragment,
		Scheme:      baseScheme,
	}

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
		p.output.writeString(t.Scheme)
		p.output.writeRune(':')
	}
	p.outputPositions.SchemeEnd = p.output.len()

	if t.HasAuthority {
		p.output.writeString("//")
		p.output.writeString(t.Authority)
	}
	p.outputPositions.AuthorityEnd = p.output.len()

	p.output.writeString(t.Path)
	p.outputPositions.PathEnd = p.output.len()

	if t.HasQuery {
		p.output.writeRune('?')
		p.output.writeString(t.Query)
	}
	p.outputPositions.QueryEnd = p.output.len()

	if t.HasFragment {
		p.output.writeRune('#')
		p.output.writeString(t.Fragment)
	}
}

// parseRelativeNoBase handles parsing a relative reference when no base IRI is provided.
// In this case, it's parsed as a relative-path reference.
func (p *iriParser) parseRelativeNoBase() error {
	p.outputPositions.SchemeEnd = 0
	p.inputSchemeEnd = 0
	if p.input.startsWith('/') {
		p.input.next()
		p.output.writeRune('/')
		return p.parsePath()
	}

	return p.parsePathNoScheme()
}

// validateRelativeRef runs a sub-parse on the relative reference string to ensure it's well-formed.
func (p *iriParser) validateRelativeRef(relativeRef string) error {
	validationParser := &iriParser{
		iri:       relativeRef,
		base:      &iriParserBase{hasBase: false},
		input:     newParserInput(relativeRef),
		output:    &voidOutputBuffer{},
		unchecked: false,
	}
	if err := validationParser.parseSchemeStart(); err != nil {
		return err
	}

	// According to RFC 3986 Section 4.2, a relative-path reference cannot
	// contain a colon in its first segment, as it would be mistaken for a scheme.
	// The generic parser will correctly parse such a string (e.g., "a:b") as an
	// absolute URI with scheme "a".
	// Since this validation function is specifically for references to be resolved
	// against a base, we must reject this ambiguous form.
	if validationParser.outputPositions.SchemeEnd > 0 {
		// It was parsed as an absolute URI. Check if it's the ambiguous form.
		// The ambiguous form is `scheme:path-rootless`.
		// It's unambiguous if it has an authority (`scheme://...`) or absolute path (`scheme:/...`).
		uriAfterScheme := relativeRef[validationParser.inputSchemeEnd:]
		if !strings.HasPrefix(uriAfterScheme, "/") {
			// This is the ambiguous case (e.g., "a:b"). Per RFC 3986, this form
			// is invalid as a relative-path reference.
			return &kindError{message: "Invalid IRI character in first path segment", char: ':'}
		}
	}

	return nil
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

	p.recomposeIRI(t)
	return nil
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
