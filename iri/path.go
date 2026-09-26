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

// applyDotSegmentRules evaluates the prefix of the input path and applies rules 2A-2D of the remove dot segments
// algorithm.
//
// Specification Reference:
// RFC 3986 (Section 5.2.4)
//
// Parameters:
//   - in: The current remaining input path string to be processed.
//   - output: The slice of accumulated path segments.
//
// Returns:
//   - string: The remaining, unconsumed portion of the input path.
//   - []string: The modified slice of accumulated path segments.
//   - bool: True if a rule was successfully applied; otherwise, false.
func applyDotSegmentRules(in string, output []string) (string, []string, bool) {
	// Spec Rule: RFC 3986 (Section 5.2.4)
	// Rule 2A: Handle "../" or "./" prefixes by removing them from the input buffer.
	if strings.HasPrefix(in, "../") {
		return in[3:], output, true
	}
	if strings.HasPrefix(in, "./") {
		return in[2:], output, true
	}

	// Spec Rule: RFC 3986 (Section 5.2.4)
	// Rule 2B: Handle "/./" or "/." prefixes by replacing them with a single slash.
	if strings.HasPrefix(in, "/./") {
		return "/" + in[3:], output, true
	}
	if in == "/." {
		return "/", output, true
	}

	// Spec Rule: RFC 3986 (Section 5.2.4)
	// Rule 2C: Handle "/../" or "/.." prefixes by moving one segment up in the output buffer.
	if strings.HasPrefix(in, "/../") || in == "/.." {
		newIn := "/"
		if len(in) > len("/..") {
			// Append the remainder after the "/.." prefix if additional segments exist.
			newIn += in[4:]
		}
		if len(output) > 0 {
			output = output[:len(output)-1]
		}
		return newIn, output, true
	}

	// Spec Rule: RFC 3986 (Section 5.2.4)
	// Rule 2D: Handle standalone "." or ".." segments by discarding them.
	if in == "." || in == ".." {
		return "", output, true
	}

	// Implementation Note: Rule fallback
	// Return false to indicate that no dot-segment matching rule was applicable.
	return in, output, false
}

// extractFirstSegment parses the first segment from the input path, honoring leading slash boundaries.
//
// Specification Reference:
// RFC 3986 (Section 5.2.4)
//
// Parameters:
//   - in: The input path string from which the segment is extracted.
//
// Returns:
//   - string: The extracted first segment of the path, including the leading slash if present.
//   - string: The remaining unparsed portion of the input path.
func extractFirstSegment(in string) (string, string) {
	slashIndex := strings.Index(in, "/")
	if slashIndex == 0 {
		// Implementation Note: Absolute path segment extraction
		// Locate the next slash when the path starts with a slash (e.g., "/a/b").
		nextSlash := strings.Index(in[1:], "/")
		if nextSlash == -1 {
			return in, ""
		}
		// Implementation Note: Sliced segment bounds
		// The extracted segment includes the leading slash.
		return in[:nextSlash+1], in[nextSlash+1:]
	}

	// Implementation Note: Relative path segment extraction
	// Handle cases where the path does not start with a slash (e.g., "a/b").
	if slashIndex == -1 {
		return in, ""
	}
	// Implementation Note: Delimiter slice boundary
	// Extract everything up to but excluding the next slash.
	return in[:slashIndex], in[slashIndex:]
}

// removeDotSegments resolves and removes relative dot segments ("." and "..") from a path.
//
// Specification Reference:
// RFC 3986 (Section 5.2.4)
//
// Parameters:
//   - input: The raw path string potentially containing relative segments.
//
// Returns:
//   - string: The normalized path string with dot segments resolved.
func removeDotSegments(input string) string {
	// Step 1: Initialize buffers
	// Prepare the tracking output slice and load the original path input.
	var output []string
	in := input

	// Step 2: Loop while input buffer is not empty
	// Sequentially reduce potential dot segments.
	for len(in) > 0 {
		var ruleApplied bool
		in, output, ruleApplied = applyDotSegmentRules(in, output)
		if ruleApplied {
			continue
		}

		// Spec Rule: RFC 3986 (Section 5.2.4)
		// Rule 2E: Move the first path segment from the input buffer to the end of the output buffer.
		var segment, remainder string
		segment, remainder = extractFirstSegment(in)
		in = remainder
		output = append(output, segment)
	}

	// Step 3: Recompose path segments
	// Join parsed segments together to produce the final normalized path string.
	return strings.Join(output, "")
}

// resolvePath merges a base path with a relative path and resolves dot segments.
//
// Specification Reference:
// RFC 3986 (Section 5.2.3)
//
// Parameters:
//   - basePath: The path component of the base URI.
//   - relPath: The relative path reference being resolved.
//
// Returns:
//   - string: The merged and normalized absolute-equivalent path.
func resolvePath(basePath, relPath string) string {
	// Step 1: Find merging boundary
	// Locate the last slash to isolate the base directory prefix.
	lastSlash := strings.LastIndex(basePath, "/")
	if lastSlash == -1 {
		// Spec Rule: RFC 3986 (Section 5.2.3)
		// Resolve path relative to empty base path.
		return removeDotSegments(relPath)
	}
	// Spec Rule: RFC 3986 (Section 5.2.3)
	// Concatenate the isolated base prefix with the relative path and resolve.
	return removeDotSegments(basePath[:lastSlash+1] + relPath)
}
