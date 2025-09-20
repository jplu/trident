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

// relativizeWithAuthority handles the most complex case where both IRIs have
// an authority, and paths need to be compared.
func (i *Iri) relativizeWithAuthority(abs *Iri) (*Ref, error) {
	basePath := i.Path()
	targetPath := abs.Path()

	// Handle empty paths as root, per RFC 3986
	if basePath == "" {
		basePath = "/"
	}
	if targetPath == "" {
		targetPath = "/"
	}

	// Determine the "directory" of the base path.
	// If the base path ends with a '/', it's a directory.
	// Otherwise, it's a "file", and its directory is the path up to the last '/'.
	baseDir := basePath
	lastSlash := strings.LastIndex(baseDir, "/")
	if lastSlash > -1 {
		baseDir = baseDir[:lastSlash+1]
	}

	// Split the directories into segments for comparison.
	// We trim the slashes to get clean segment lists.
	baseSegs := strings.Split(strings.Trim(baseDir, "/"), "/")

	trimmedTargetPath := strings.TrimPrefix(targetPath, "/")
	targetSegs := strings.Split(trimmedTargetPath, "/")

	// An empty split result means it was the root directory.
	if baseDir == "/" {
		baseSegs = []string{}
	}
	if targetPath == "/" {
		targetSegs = []string{}
	}

	// Find the length of the common directory prefix.
	commonLen := 0
	for commonLen < len(baseSegs) && commonLen < len(targetSegs) && baseSegs[commonLen] == targetSegs[commonLen] {
		commonLen++
	}

	var b strings.Builder
	// For each directory in the base path that is not common, we need to go "up".
	for i := commonLen; i < len(baseSegs); i++ {
		b.WriteString("../")
	}

	// Now, append the remaining part of the target path.
	b.WriteString(strings.Join(targetSegs[commonLen:], "/"))
	relPath := b.String()

	// If we produce an empty string, it means the target is in the same directory
	// as the base "file". The correct representation for this is ".".
	if relPath == "" {
		// This handles the case where base is "a/b" and target is "a/c", producing "c".
		// But if base is "a/b" and target is "a/", we need "."
		lastTargetSlash := strings.LastIndex(targetPath, "/")
		if lastTargetSlash > -1 && targetPath[lastTargetSlash+1:] == "" { // target is a directory
			return buildRelativeRef(".", abs)
		}
	}

	return buildRelativeRef(relPath, abs)
}

// buildRelativeRef constructs the final relative reference string from a relative path
// and the query/fragment parts of the absolute target IRI.
func buildRelativeRef(relPath string, abs *Iri) (*Ref, error) {
	absQuery, hasAbsQuery := abs.Query()
	absFragment, hasAbsFragment := abs.Fragment()

	var b strings.Builder
	b.WriteString(relPath)
	if hasAbsQuery {
		b.WriteRune('?')
		b.WriteString(absQuery)
	}
	if hasAbsFragment {
		b.WriteRune('#')
		b.WriteString(absFragment)
	}
	return ParseRef(b.String())
}

// relativizeForNoAuthority handles relativization when both IRIs lack an authority part.
func (i *Iri) relativizeForNoAuthority(abs *Iri) (*Ref, error) {
	basePath := i.Path()
	absPath := abs.Path()

	baseSegs := strings.Split(basePath, "/")
	absSegs := strings.Split(absPath, "/")

	var baseDirSegs []string
	if !strings.HasSuffix(basePath, "/") {
		if len(baseSegs) > 0 {
			baseDirSegs = baseSegs[:len(baseSegs)-1]
		}
	} else {
		baseDirSegs = baseSegs[:len(baseSegs)-1]
	}

	commonSegs := 0
	for commonSegs < len(baseDirSegs) && commonSegs < len(absSegs) && baseDirSegs[commonSegs] == absSegs[commonSegs] {
		commonSegs++
	}

	var b strings.Builder
	for i := commonSegs; i < len(baseDirSegs); i++ {
		b.WriteString("../")
	}

	b.WriteString(strings.Join(absSegs[commonSegs:], "/"))

	relPath := b.String()
	if relPath == "" && basePath != absPath {
		relPath = "."
	}

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

// relativizeForSamePathWithEmptyTargetQuery handles a specific edge case where
// paths match, but the target has no query while the base does.
func (i *Iri) relativizeForSamePathWithEmptyTargetQuery(abs *Iri) (*Ref, error) {
	_, hasAbsAuthority := abs.Authority()

	// If the target has no authority, its structure is incompatible with a base
	// that has one. The only valid reference is the full absolute IRI.
	if !hasAbsAuthority {
		return ParseRef(abs.String())
	}

	absPath := abs.Path()
	if absPath != "" {
		lastSlash := strings.LastIndex(absPath, "/")
		relPath := absPath[lastSlash+1:]
		if relPath == "" {
			relPath = "."
		}
		return buildRelativeRef(relPath, abs)
	}

	// Path is empty and we know it has an authority, so create a scheme-relative ref.
	return ParseRef(abs.String()[abs.positions.SchemeEnd:])
}

// relativizeForSamePath handles relativization when base and target paths are identical.
func (i *Iri) relativizeForSamePath(abs *Iri) (*Ref, error) {
	base := i
	baseQuery, hasBaseQuery := base.Query()
	absQuery, hasAbsQuery := abs.Query()
	absFragment, hasAbsFragment := abs.Fragment()

	if hasBaseQuery == hasAbsQuery && baseQuery == absQuery {
		if hasAbsFragment {
			return ParseRef("#" + absFragment)
		}
		return ParseRef("")
	}

	if !hasAbsQuery && hasBaseQuery {
		return i.relativizeForSamePathWithEmptyTargetQuery(abs)
	}

	return ParseRef(abs.String()[abs.positions.PathEnd:])
}
