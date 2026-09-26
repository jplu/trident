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

import "fmt"

// FacetPattern represents the pattern constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.4)
//
// Representation:
// It constrains the lexical space of a datatype by restricting it to values that match one or more of the specified
// regular expressions.
type FacetPattern struct {
	// BaseFacet provides the common properties shared by all facets.
	BaseFacet
	// Matchers contains the compiled regular expression patterns to evaluate against.
	Matchers []*PatternMatcher
}

// NewFacetPattern instantiates a new FacetPattern constraining facet with the provided regular expression matchers.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.4)
//
// Parameters:
//   - matchers: The compiled regular expression matchers used to evaluate the lexical space.
//
// Returns:
//   - FacetPattern: A new instance of FacetPattern configured with the matchers and NamePattern.
func NewFacetPattern(matchers []*PatternMatcher) FacetPattern {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.4)
	// The 'fixed' property is not applicable to pattern facets.
	return FacetPattern{
		BaseFacet: BaseFacet{
			Name:  NamePattern,
			Fixed: false,
		},
		Matchers: matchers,
	}
}

// Check asserts that the given value satisfies the pattern constraint.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.4.1)
//
// Parameters:
//   - val: The XSDValue to check in the value space.
//
// Returns:
//   - error: An error if validation fails, though this is a no-op as pattern applies to lexical space.
func (f FacetPattern) Check(_ XSDValue) error {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.4.1)
	// The pattern facet constrains the lexical space, not the value space. Therefore, this value-space check is a
	// no-op, and actual evaluation is deferred to CheckLexical.
	return nil
}

// CheckLexical verifies that the normalized literal matches at least one of the configured patterns.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.4.4)
//
// Parameters:
//   - literal: The normalized lexical string to validate.
//
// Returns:
//   - error: An error if the lexical value does not match any required regular expression.
func (f FacetPattern) CheckLexical(literal string) error {
	if len(f.Matchers) == 0 {
		return nil
	}

	matched := false
	for _, matcher := range f.Matchers {
		if matcher.MatchString(literal) {
			matched = true
			break
		}
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.4.4)
	// If multiple patterns are defined on a single simple type, they are combined with an implicit logical OR. The
	// literal satisfies the constraint if it matches at least one pattern.
	if !matched {
		return fmt.Errorf("pattern violation: lexical value '%s' does not match required regular expressions", literal)
	}

	return nil
}
