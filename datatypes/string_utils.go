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

package datatypes

import "strings"

// replaceWhitespace normalizes whitespace characters in a string by replacing horizontal tabs, line feeds, and carriage
// returns with a single space rune.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.6)
//
// Parameters:
//   - s: The raw input string to be normalized.
//
// Returns:
//   - string: The normalized string with whitespace replaced by spaces.
func replaceWhitespace(s string) string {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.6.1)
	// Under the 'replace' facet, all occurrences of #x9 (tab), #xA (line feed), and #xD (carriage return) are replaced
	// with #x20 (space).
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\t' || r == '\n' || r == '\r' {
			b.WriteRune(' ')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// collapseWhitespace normalizes a string according to the collapse facet behavior by replacing all tabs, line feeds,
// and carriage returns with spaces, collapsing multiple consecutive spaces, and trimming leading and trailing spaces.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.6)
//
// Parameters:
//   - s: The raw input string to be collapsed.
//
// Returns:
//   - string: The collapsed and normalized string value.
func collapseWhitespace(s string) string {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.6.1)
	// Under the 'collapse' facet, contiguous sequences of spaces are collapsed to a single space, and leading/trailing
	// spaces are removed.

	// Step 1: Replace whitespace characters
	// Replace all tab, line feed, and carriage return characters with space characters.
	replaced := replaceWhitespace(s)

	// Step 2: Strip leading and trailing spaces
	// Remove any leading or trailing space characters from the replaced string.
	trimmed := strings.TrimSpace(replaced)
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(trimmed))
	inSpace := false

	// Step 3: Collapse internal spaces
	// Reduce contiguous sequences of internal space characters to a single space.
	for _, r := range trimmed {
		if r == ' ' {
			if !inSpace {
				b.WriteRune(' ')
				inSpace = true
			}
		} else {
			b.WriteRune(r)
			inSpace = false
		}
	}
	return b.String()
}
