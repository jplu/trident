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

import "strings"

// resolvedIRI represents the components of an IRI after reference resolution as defined in the governing specification.
//
// Specification Reference:
// RFC 3986 (Section 5.2.2) and RFC 3987 (Section 6.5)
//
// Representation:
// An internal parsed data model representing the deconstructed components of an Internationalized Resource Identifier
// (IRI) or Uniform Resource Identifier (URI) reference.
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
//
// Specification Reference:
// RFC 3986 (Section 3.1) and RFC 3987 (Section 2.2)
//
// Parameters:
//   - schemePart: The scheme substring to validate.
//
// Returns:
//   - bool: True if the string conforms to the allowed character set of a scheme, false otherwise.
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
//
// Specification Reference:
// RFC 3986 (Section 3.1) and RFC 3987 (Section 2.2)
//
// Parameters:
//   - ref: The reference string potentially containing a scheme prefix.
//
// Returns:
//   - string: The extracted scheme component (empty if not found).
//   - string: The remaining reference string after the scheme colon.
//   - bool: True if a valid scheme component was extracted, false otherwise.
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
//
// Specification Reference:
// RFC 3986 (Section 5.2.2) and RFC 3987 (Section 6.5)
//
// Parameters:
//   - ref: The relative reference string to parse.
//
// Returns:
//   - string: The extracted scheme component.
//   - string: The extracted authority component.
//   - string: The extracted path component.
//   - string: The extracted query component.
//   - string: The extracted fragment component.
//   - bool: Flag indicating if the reference explicitly contains an authority component.
//   - bool: Flag indicating if the reference explicitly contains a query component.
//   - bool: Flag indicating if the reference explicitly contains a fragment component.
func deconstructRef(ref string) (
	string, string, string, string, string,
	bool, bool, bool,
) {
	var scheme, authority, path, query, fragment string
	var hasAuthority, hasQuery, hasFragment bool

	// Step 1: Extract fragment component
	// Isolate and record the optional fragment identifier from the reference string.
	if i := strings.Index(ref, "#"); i != -1 {
		hasFragment = true
		fragment = ref[i+1:]
		ref = ref[:i]
	}

	// Step 2: Extract query component
	// Isolate and record the optional query string from the reference string.
	if i := strings.Index(ref, "?"); i != -1 {
		hasQuery = true
		query = ref[i+1:]
		ref = ref[:i]
	}

	// Step 3: Extract scheme component
	// Parse the optional scheme component from the reference string.
	scheme, ref, _ = extractRefScheme(ref)

	// Step 4: Extract authority and path components
	// Determine if an authority component is present and extract it along with the path.
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
//
// Specification Reference:
// RFC 3986 (Section 5.2.2) and RFC 3987 (Section 6.5)
//
// Parameters:
//   - t: The target resolvedIRI structure to populate with the resolved path and query.
//   - rPath: The relative path component.
//   - rQuery: The relative query component.
//   - rHasQuery: Flag indicating if the relative reference explicitly contains a query component.
//   - basePath: The base path component.
//   - baseQuery: The base query component.
//   - hasBaseQuery: Flag indicating if the base IRI contains a query component.
//   - hasBaseAuthority: Flag indicating if the base IRI contains an authority component.
func (p *iriParser) resolvePathAndQuery(
	t *resolvedIRI,
	rPath, rQuery string,
	rHasQuery bool,
	basePath, baseQuery string,
	hasBaseQuery, hasBaseAuthority bool,
) {
	if rPath != "" {
		// Spec Rule: RFC 3986 (Section 5.2.2)
		// If the relative path is not empty, check if it starts with a slash (absolute path) or requires merging.
		if strings.HasPrefix(rPath, "/") {
			t.Path = removeDotSegments(rPath)
		} else {
			// Spec Rule: RFC 3986 (Section 5.2.3)
			// Merge the relative path with the base path component.
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

	// Spec Rule: RFC 3986 (Section 5.2.2)
	// If the relative path is empty, inherit the base path component and determine query precedence.
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
//
// Specification Reference:
// RFC 3986 (Section 5.2) and RFC 3987 (Section 6.5)
//
// Parameters:
//   - relativeRef: The relative reference string to resolve against the base IRI.
//
// Returns:
//   - *resolvedIRI: A pointer to the populated resolvedIRI struct representing the resolved components.
func (p *iriParser) resolveComponents(relativeRef string) *resolvedIRI {
	rScheme, rAuthority, rPath, rQuery, rFragment, rHasAuthority, rHasQuery, rHasFragment := deconstructRef(
		relativeRef,
	)

	// Spec Rule: RFC 3986 (Section 5.2.2)
	// If the reference contains a scheme component, it is treated as absolute and base components are ignored.
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

	// Step 1: Extract base components
	// Isolate individual components from the parsed base IRI structure.
	baseScheme, baseAuthority, basePath, hasBaseAuthority, baseQuery, hasBaseQuery := p.getBaseComponents()

	t := &resolvedIRI{
		Fragment:    rFragment,
		HasFragment: rHasFragment,
		Scheme:      baseScheme,
	}

	// Step 2: Apply resolution precedence hierarchy
	// Check for the presence of authority or delegate path resolution when authority is absent.
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
//
// Specification Reference:
// RFC 3986 (Section 5.2.1) and RFC 3987 (Section 6.5)
//
// Returns:
//   - string: The base scheme component.
//   - string: The base authority component.
//   - string: The base path component.
//   - bool: Flag indicating if the base IRI contains an authority component.
//   - string: The base query component.
//   - bool: Flag indicating if the base IRI contains a query component.
func (p *iriParser) getBaseComponents() (string, string, string, bool, string, bool) {
	var scheme, authority, path, query string
	var hasAuthority, hasQuery bool
	b := p.base

	// Step 1: Isolate base scheme
	// Identify the index bound and extract the scheme if present.
	if b.schemeEnd > 0 {
		scheme = b.iri[:b.schemeEnd-1]
	}

	// Step 2: Isolate base authority
	// Check authority bounds, strip the double-slash prefix, and isolate the authority.
	if b.authorityEnd > b.schemeEnd {
		hasAuthority = true
		start := b.schemeEnd
		if strings.HasPrefix(b.iri[start:], "//") {
			start += 2
		}
		if b.authorityEnd > start {
			authority = b.iri[start:b.authorityEnd]
		}
	}

	// Step 3: Isolate base path and query
	// Extract the remaining segments and record query fields.
	path = b.iri[b.authorityEnd:b.pathEnd]
	if b.queryEnd > b.pathEnd {
		query = b.iri[b.pathEnd+1 : b.queryEnd]
		hasQuery = true
	}
	return scheme, authority, path, hasAuthority, query, hasQuery
}

// recomposeIRI assembles the final IRI from its resolved components into the output buffer.
//
// Specification Reference:
// RFC 3986 (Section 5.3) and RFC 3987 (Section 3.1)
//
// Parameters:
//   - t: The populated resolvedIRI structure containing parsed and resolved components.
func (p *iriParser) recomposeIRI(t *resolvedIRI) {
	// Step 1: Recompose scheme
	// Write scheme component and update scheme bounds.
	if t.Scheme != "" {
		p.output.writeString(t.Scheme)
		p.output.writeRune(':')
	}
	p.outputPositions.SchemeEnd = p.output.len()

	// Step 2: Recompose authority
	// Write authority component if present and update authority bounds.
	if t.HasAuthority {
		p.output.writeString("//")
		p.output.writeString(t.Authority)
	}
	p.outputPositions.AuthorityEnd = p.output.len()

	// Step 3: Recompose path
	// Write path component and update path bounds.
	p.output.writeString(t.Path)
	p.outputPositions.PathEnd = p.output.len()

	// Step 4: Recompose query
	// Write query component if present and update query bounds.
	if t.HasQuery {
		p.output.writeRune('?')
		p.output.writeString(t.Query)
	}
	p.outputPositions.QueryEnd = p.output.len()

	// Step 5: Recompose fragment
	// Write optional fragment component if present.
	if t.HasFragment {
		p.output.writeRune('#')
		p.output.writeString(t.Fragment)
	}
}
