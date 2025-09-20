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

// applyDotSegmentRules handles rules 2A-2D of RFC 3986, Section 5.2.4.
// It modifies the input path `in` and output buffer `output` if a rule is matched.
// It returns the modified path, the modified output buffer, and a boolean
// indicating if a rule was successfully applied.
func applyDotSegmentRules(in string, output []string) (string, []string, bool) {
	// Rule 2A: "../" or "./"
	if strings.HasPrefix(in, "../") {
		return in[3:], output, true
	}
	if strings.HasPrefix(in, "./") {
		return in[2:], output, true
	}
	// Rule 2B: "/./" or "/."
	if strings.HasPrefix(in, "/./") {
		return "/" + in[3:], output, true
	}
	if in == "/." {
		return "/", output, true
	}
	// Rule 2C: "/../" or "/.."
	if strings.HasPrefix(in, "/../") || in == "/.." {
		newIn := "/"
		if len(in) > len("/..") { // Distinguishes "/../" from "/.."
			newIn += in[4:]
		}
		if len(output) > 0 {
			lastSegment := output[len(output)-1]
			output = output[:len(output)-1]

			if len(output) == 0 && !strings.HasPrefix(lastSegment, "/") {
				newIn = strings.TrimPrefix(newIn, "/")
			}
		}
		return newIn, output, true
	}
	// Rule 2D: "." or ".."
	if in == "." || in == ".." {
		return "", output, true
	}
	// No rule applied
	return in, output, false
}

// extractFirstSegment handles rule 2E of RFC 3986, Section 5.2.4.
// It extracts the first path segment from the input buffer `in` and returns
// that segment along with the remainder of the input buffer.
func extractFirstSegment(in string) (string, string) {
	slashIndex := strings.Index(in, "/")
	if slashIndex == 0 { // Path starts with a slash, e.g., "/a/b"
		nextSlash := strings.Index(in[1:], "/")
		if nextSlash == -1 {
			return in, ""
		}
		// The segment includes the slash
		return in[:nextSlash+1], in[nextSlash+1:]
	}

	// Path does not start with a slash, e.g., "a/b"
	if slashIndex == -1 {
		return in, ""
	}
	// The segment is up to the slash
	return in[:slashIndex], in[slashIndex:]
}

// removeDotSegments implements the "Remove Dot Segments" algorithm from
// RFC 3986, Section 5.2.4. It normalizes a path by resolving "." and ".." segments.
func removeDotSegments(input string) string {
	var output []string
	in := input

	for len(in) > 0 {
		var ruleApplied bool
		in, output, ruleApplied = applyDotSegmentRules(in, output)
		if ruleApplied {
			continue
		}

		// Rule 2E: No special rule applied, so move the first path segment
		// from the input buffer to the end of the output buffer.
		var segment, remainder string
		segment, remainder = extractFirstSegment(in)
		in = remainder
		output = append(output, segment)
	}

	return strings.Join(output, "")
}

// resolvePath resolves a relative path against a base path according to
// RFC 3986, Section 5.2.2. It merges the base path with the relative
// reference path.
func resolvePath(basePath, relPath string) string {
	lastSlash := strings.LastIndex(basePath, "/")
	if lastSlash == -1 {
		return removeDotSegments(relPath)
	}
	return removeDotSegments(basePath[:lastSlash+1] + relPath)
}
