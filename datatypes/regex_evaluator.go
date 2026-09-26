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

import (
	"regexp"
)

// PatternMatcher represents the pattern constraining facet matcher as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.4)
//
// Representation:
// An immutable structure containing a compiled regular expression engine instance alongside its original raw expression
// literal.
type PatternMatcher struct {
	// Re holds the compiled Go standard library regular expression instance.
	re *regexp.Regexp
	// Original holds the raw XSD regular expression string as initially provided.
	original string
}

// CompileRegex compiles an XML Schema regular expression pattern into a matcher.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G)
//
// Parameters:
//   - xsdPattern: The raw regular expression pattern defined in the schema.
//
// Returns:
//   - *PatternMatcher: A pointer to the initialized pattern matcher.
//   - error: An error if the regular expression cannot be translated or compiled.
func CompileRegex(xsdPattern string) (*PatternMatcher, error) {
	// Step 1: Translate the regular expression dialect
	// Converts XSD-specific regex constructs into an RE2-compatible format.
	//
	// Implementation Note: Translation of regular expression dialect to RE2
	// W3C XSD 1.1 Part 2 (Appendix G) defines a specific regular expression language with features like character class
	// subtraction and special escapes (\i, \c, etc.) that are not natively supported by Go's RE2 engine.
	translated, err := translateRegex(xsdPattern)
	if err != nil {
		return nil, err
	}

	// Step 2: Apply implicit anchoring
	// Wraps the translated pattern to ensure it matches the entire input string from start to end.
	//
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.4)
	// The regular expression is implicitly anchored at both the beginning and the end. Unlike standard regex matchers
	// that search for substrings, the entire string must match. Wrapping the translated pattern in anchoring boundary
	// patterns enforces this behavior.
	anchored := "^(?:" + translated + ")$"

	// Step 3: Compile the Go regular expression
	// Compiles the anchored RE2 pattern using Go's regexp package.
	re, err := regexp.Compile(anchored)
	if err != nil {
		return nil, err
	}

	// Step 4: Construct the matcher
	// Instantiates and returns the PatternMatcher wrapper.
	return &PatternMatcher{
		re:       re,
		original: xsdPattern,
	}, nil
}

// MatchString tests if the provided string matches the compiled regular expression.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G)
//
// Parameters:
//   - s: The lexical representation string to evaluate.
//
// Returns:
//   - bool: True if the entire string matches the pattern, otherwise false.
func (p *PatternMatcher) MatchString(s string) bool {
	return p.re.MatchString(s)
}

// String returns the original, untranslated regular expression pattern.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.4)
//
// Returns:
//   - string: The raw, untranslated regular expression pattern as a string.
func (p *PatternMatcher) String() string {
	return p.original
}
