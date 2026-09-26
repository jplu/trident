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

// Relativize computes a relative IRI reference from a base IRI and a target absolute IRI.
//
// Specification Reference:
// RFC 3987 (Section 6.5)
//
// Parameters:
//   - abs: The target absolute IRI to relativize against the base.
//
// Returns:
//   - *Ref: The generated relative IRI reference.
//   - error: An error if relativization is not possible, such as when the target contains dot-segments.
func (i *Iri) Relativize(abs *Iri) (*Ref, error) {
	base := i
	absPath := abs.Path()

	// Spec Rule: RFC 3986 (Section 5.2.4)
	// Relative reference resolution relies on removing dot-segments. Target paths containing
	// dot-segments must be rejected or normalized prior to computing relative references
	// to prevent resolution path traversal vulnerabilities.
	for _, segment := range strings.Split(absPath, "/") {
		if segment == "." || segment == ".." {
			return nil, ErrIriRelativize
		}
	}

	// Spec Rule: RFC 3986 (Section 5.2.2)
	// If the schemes differ, the target cannot be made relative to the base IRI,
	// so the absolute target IRI must be returned as is.
	if base.Scheme() != abs.Scheme() {
		return ParseRef(abs.String())
	}

	baseAuthority, hasBaseAuthority := base.Authority()
	absAuthority, hasAbsAuthority := abs.Authority()

	// Spec Rule: RFC 3986 (Section 5.2.2)
	// If the authorities differ, we cannot construct a relative-path reference. Instead,
	// we return a scheme-relative (network-path) reference if the target has an authority,
	// or the full target IRI if the target lacks one.
	if hasBaseAuthority != hasAbsAuthority || (hasBaseAuthority && baseAuthority != absAuthority) {
		if !hasAbsAuthority {
			return ParseRef(abs.String())
		}
		return ParseRef(abs.String()[abs.positions.SchemeEnd:])
	}

	basePath := base.Path()

	// Spec Rule: RFC 3986 (Section 5.2.2)
	// An empty path with a defined authority requires a scheme-relative reference
	// if the base has a path, to avoid inheriting the base path components.
	if absPath == "" && basePath != "" {
		if !hasAbsAuthority {
			return ParseRef(abs.String())
		}
		return ParseRef(abs.String()[abs.positions.SchemeEnd:])
	}

	if basePath == absPath {
		return i.relativizeForSamePath(abs)
	}

	if !hasBaseAuthority {
		return i.relativizeForNoAuthority(abs)
	}

	return i.relativizeWithAuthority(abs)
}

// relativizeWithAuthority computes a relative IRI reference when both base and target IRIs have an authority component.
//
// Specification Reference:
// RFC 3986 (Section 5.2.2)
//
// Parameters:
//   - abs: The target absolute IRI containing the authority and path.
//
// Returns:
//   - *Ref: The generated relative IRI reference.
//   - error: An error if parsing or reference construction fails.
func (i *Iri) relativizeWithAuthority(abs *Iri) (*Ref, error) {
	basePath := i.Path()
	targetPath := abs.Path()

	// Step 1: Normalize Empty Paths
	// Treat empty paths as the root directory "/" for consistent traversal logic.
	if basePath == "" {
		basePath = "/"
	}
	if targetPath == "" {
		targetPath = "/"
	}

	// Step 2: Segment Paths
	// Split base and target paths into directory segments and trailing file components to preserve structural
	// boundaries.
	baseDirSegs, _ := splitPathSegments(basePath)
	targetDirSegs, targetFile := splitPathSegments(targetPath)

	// Step 3: Compute Common Prefix Length
	// Identify how many segments of the path hierarchy are shared between the base and target.
	commonLen := 0
	for commonLen < len(baseDirSegs) && commonLen < len(targetDirSegs) && baseDirSegs[commonLen] == targetDirSegs[commonLen] {
		commonLen++
	}

	var b strings.Builder
	// Step 4: Append Upward Traversals
	// For each directory segment in the base path that is not shared with the target, append a relative parent
	// directory segment ("../").
	for i := commonLen; i < len(baseDirSegs); i++ {
		b.WriteString("../")
	}

	// Step 5: Append Remaining Target Segments
	// Append the remaining directory and file segments of the target path starting from the first non-common segment.
	if commonLen < len(targetDirSegs) {
		b.WriteString(strings.Join(targetDirSegs[commonLen:], "/"))
		b.WriteString("/")
	}
	b.WriteString(targetFile)

	relPath := b.String()

	// Implementation Note: Directory Edge Cases
	// If the relative path is empty, check if the target is a directory itself.
	// If the target is a directory, the correct relative reference is the current directory "." to avoid resolution
	// confusion.
	if relPath == "" {
		lastTargetSlash := strings.LastIndex(targetPath, "/")
		if lastTargetSlash > -1 && targetPath[lastTargetSlash+1:] == "" {
			return buildRelativeRef(".", abs)
		}
	}

	return buildRelativeRef(relPath, abs)
}

// buildRelativeRef constructs the final relative reference string from a relative path and the target query and
// fragment components.
//
// Specification Reference:
// RFC 3986 (Section 5.3)
//
// Parameters:
//   - relPath: The pre-computed relative path.
//   - abs: The absolute target IRI containing query and fragment components to preserve.
//
// Returns:
//   - *Ref: The constructed relative IRI reference.
//   - error: An error if the constructed reference fails to parse.
func buildRelativeRef(relPath string, abs *Iri) (*Ref, error) {
	absQuery, hasAbsQuery := abs.Query()
	absFragment, hasAbsFragment := abs.Fragment()

	var b strings.Builder
	// Step 1: Append Path
	// Write the resolved relative path to the output builder.
	b.WriteString(relPath)

	// Step 2: Append Query
	// If the target IRI contains a query component, append it prefixed by "?".
	if hasAbsQuery {
		b.WriteRune('?')
		b.WriteString(absQuery)
	}

	// Step 3: Append Fragment
	// If the target IRI contains a fragment component, append it prefixed by "#".
	if hasAbsFragment {
		b.WriteRune('#')
		b.WriteString(absFragment)
	}
	return ParseRef(b.String())
}

// relativizeForNoAuthority computes a relative reference when both IRIs lack an authority component.
//
// Specification Reference:
// RFC 3986 (Section 5.2.2)
//
// Parameters:
//   - abs: The target absolute IRI containing the unshaded relative path.
//
// Returns:
//   - *Ref: The generated relative IRI reference.
//   - error: An error if parsing or reference construction fails.
func (i *Iri) relativizeForNoAuthority(abs *Iri) (*Ref, error) {
	basePath := i.Path()
	absPath := abs.Path()

	// Step 1: Segment Paths
	// Split base and target paths into directory segments and trailing file components to preserve structural
	// boundaries.
	baseDirSegs, _ := splitPathSegments(basePath)
	absDirSegs, absFile := splitPathSegments(absPath)

	// Step 2: Count Shared Segments
	// Find the shared common prefix segments between the base directory and the target.
	commonSegs := 0
	for commonSegs < len(baseDirSegs) && commonSegs < len(absDirSegs) && baseDirSegs[commonSegs] == absDirSegs[commonSegs] {
		commonSegs++
	}

	var b strings.Builder
	// Step 3: Traverse Upward
	// Append parent directory segments ("../") for any unshared segments of the base path.
	for i := commonSegs; i < len(baseDirSegs); i++ {
		b.WriteString("../")
	}

	// Step 4: Append Remaining Target Segments
	// Append the remaining directory and file segments of the target path.
	if commonSegs < len(absDirSegs) {
		b.WriteString(strings.Join(absDirSegs[commonSegs:], "/"))
		b.WriteString("/")
	}
	b.WriteString(absFile)

	relPath := b.String()
	if relPath == "" && basePath != absPath {
		relPath = "."
	}

	// Step 5: Avoid Colon Ambiguity
	// Spec Rule: RFC 3986 (Section 4.2)
	// A relative-path reference cannot contain a colon in its first segment, as it would be
	// mistaken for a scheme component. If a colon appears in the first segment before any
	// slash, prepend "./" to disambiguate the relative path.
	if !strings.HasPrefix(relPath, ".") && !strings.HasPrefix(relPath, "/") {
		firstColon := strings.Index(relPath, ":")
		if firstColon != -1 {
			firstSlash := strings.Index(relPath, "/")
			if firstSlash == -1 || firstColon < firstSlash {
				relPath = "./" + relPath
			}
		}
	}

	return buildRelativeRef(relPath, abs)
}

// relativizeForSamePathWithEmptyTargetQuery computes a relative reference for the case where base and target paths
// match but the target has no query.
//
// Specification Reference:
// RFC 3986 (Section 5.2.2)
//
// Parameters:
//   - abs: The target absolute IRI.
//
// Returns:
//   - *Ref: The generated relative IRI reference.
//   - error: An error if reference parsing fails.
func (i *Iri) relativizeForSamePathWithEmptyTargetQuery(abs *Iri) (*Ref, error) {
	_, hasAbsAuthority := abs.Authority()

	// Step 1: Check Target Authority
	// If the target lacks an authority, it cannot be relativized against a base with
	// an authority, so return the full target IRI.
	if !hasAbsAuthority {
		return ParseRef(abs.String())
	}

	absPath := abs.Path()
	// Step 2: Compute Relative Path
	// If the target path is not empty, extract the last segment of the path.
	// If this segment is empty, fall back to the current directory indicator ".".
	if absPath != "" {
		lastSlash := strings.LastIndex(absPath, "/")
		relPath := absPath[lastSlash+1:]
		if relPath == "" {
			relPath = "."
		}
		return buildRelativeRef(relPath, abs)
	}

	// Step 3: Handle Scheme-Relative Reference
	// If the target path is empty and has an authority, return a scheme-relative reference.
	return ParseRef(abs.String()[abs.positions.SchemeEnd:])
}

// relativizeForSamePath computes a relative reference when the base and target IRI paths are identical.
//
// Specification Reference:
// RFC 3986 (Section 5.2.2)
//
// Parameters:
//   - abs: The target absolute IRI with the identical path.
//
// Returns:
//   - *Ref: The generated relative IRI reference.
//   - error: An error if reference construction or parsing fails.
func (i *Iri) relativizeForSamePath(abs *Iri) (*Ref, error) {
	base := i
	baseQuery, hasBaseQuery := base.Query()
	absQuery, hasAbsQuery := abs.Query()
	absFragment, hasAbsFragment := abs.Fragment()

	// Step 1: Compare Queries
	// If the base and target queries are identical, the relative reference only needs the
	// target fragment (or can be completely empty if there is no fragment).
	if hasBaseQuery == hasAbsQuery && baseQuery == absQuery {
		if hasAbsFragment {
			return ParseRef("#" + absFragment)
		}
		return ParseRef("")
	}

	// Step 2: Handle Empty Target Query Edge Case
	// If the target has no query but the base has one, delegate to the specialized handler.
	if !hasAbsQuery && hasBaseQuery {
		return i.relativizeForSamePathWithEmptyTargetQuery(abs)
	}

	// Step 3: Return Relative Reference after Path
	// If queries differ and the target has a query, slice the target string starting from the end
	// of the path to retain the query and fragment.
	return ParseRef(abs.String()[abs.positions.PathEnd:])
}

// splitPathSegments splits a path into directory segments and a final file segment.
//
// Representation:
// A slice of directory segment strings and a trailing file segment string.
func splitPathSegments(path string) ([]string, string) {
	if path == "" {
		return nil, ""
	}
	segs := strings.Split(path, "/")
	if strings.HasSuffix(path, "/") {
		return segs[:len(segs)-1], ""
	}
	return segs[:len(segs)-1], segs[len(segs)-1]
}
