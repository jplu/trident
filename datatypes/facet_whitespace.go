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

import "errors"

// WhiteSpaceValue represents the whiteSpace normalization modes as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.6)
//
// Representation:
// A string type representing the whitespace normalization modes "preserve", "replace", and "collapse".
type WhiteSpaceValue string

const (
	// WhiteSpacePreserve instructs the processor to preserve all whitespace characters
	// exactly as they appear in the source XML literal.
	WhiteSpacePreserve WhiteSpaceValue = "preserve"

	// WhiteSpaceReplace instructs the processor to replace all occurrences of
	// tab (#x9), line feed (#xA), and carriage return (#xD) with a single space (#x20).
	WhiteSpaceReplace WhiteSpaceValue = "replace"

	// WhiteSpaceCollapse instructs the processor to perform the replace action,
	// then trim leading and trailing spaces, and collapse contiguous sequences
	// of spaces into a single space character.
	WhiteSpaceCollapse WhiteSpaceValue = "collapse"
)

// FacetWhiteSpace represents the 'whiteSpace' constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.6)
//
// Representation:
// A structure wrapping the BaseFacet properties alongside the specific WhiteSpaceValue configuration.
type FacetWhiteSpace struct {
	// BaseFacet provides the shared properties (name, fixed) for all XSD facets.
	BaseFacet

	// Value represents the specific whitespace normalization behavior to be applied.
	Value WhiteSpaceValue
}

// NewFacetWhiteSpace instantiates a new whiteSpace constraining facet with the specified value and fixed status.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.6)
//
// Parameters:
//   - val: The whitespace normalization mode to be applied.
//   - fixed: A boolean indicating whether the facet's value cannot be overridden by further derived type definitions.
//
// Returns:
//   - FacetWhiteSpace: An initialized whiteSpace constraining facet component.
func NewFacetWhiteSpace(val WhiteSpaceValue, fixed bool) FacetWhiteSpace {
	return FacetWhiteSpace{
		BaseFacet: BaseFacet{Name: NameWhiteSpace, Fixed: fixed},
		Value:     val,
	}
}

// Check validates whether the given value satisfies the restrictions of the whiteSpace facet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.6.4)
//
// Parameters:
//   - val: The XSD value to be validated against the whitespace facet restrictions.
//
// Returns:
//   - error: An error if a non-string-derived type specifies a normalization other than 'collapse'.
func (f FacetWhiteSpace) Check(val XSDValue) error {
	switch val.(type) {
	case String, NormalizedString, Token, Name, NCName, AnyURI, Language:
		return nil
	default:
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.6.4)
		// All datatypes other than string and those derived from string must
		// have a whiteSpace facet value of 'collapse'.
		if f.Value != WhiteSpaceCollapse {
			return errors.New("non-string types require whiteSpace=collapse")
		}
		return nil
	}
}

// Normalize transforms the raw input string based on the configured whitespace rule.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.6)
//
// Parameters:
//   - literal: The raw lexical string to be normalized.
//
// Returns:
//   - string: The normalized lexical string.
func (f FacetWhiteSpace) Normalize(literal string) string {
	switch f.Value {
	case WhiteSpaceReplace:
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.6.1)
		// Under the 'replace' facet, all occurrences of #x9 (tab), #xA (line feed),
		// and #xD (carriage return) are replaced with #x20 (space).
		return replaceWhitespace(literal)
	case WhiteSpaceCollapse:
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.6.1)
		// Under the 'collapse' facet, contiguous sequences of spaces are collapsed
		// to a single space, and leading/trailing spaces are removed.
		return collapseWhitespace(literal)
	case WhiteSpacePreserve:
		// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.6.1)
		// Under the 'preserve' facet, no whitespace normalization is performed,
		// and the raw input string is kept exactly as is.
		return literal
	default:
		return literal
	}
}
