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
	"errors"
	"regexp"
)

// languageRegex is the compiled regular expression used to validate the lexical
// representation of the xsd:language datatype.
var languageRegex = regexp.MustCompile(`^[a-zA-Z]{1,8}(-[a-zA-Z0-9]{1,8})*$`)

// Language represents the language datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.3)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A string representing natural language identifiers, derived from the token type
// by restriction and matching the regular expression `[a-zA-Z]{1,8}(-[a-zA-Z0-9]{1,8})*`.
type Language Token

// ParseLanguage parses a string literal matching the lexical representation of xsd:language.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.3)
//
// Parameters:
//   - s: The raw string literal to be parsed and validated.
//
// Returns:
//   - Language: The parsed Language value on success.
//   - error: An error if the input fails validation.
func ParseLanguage(s string) (Language, error) {
	tok := ParseToken(s)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.3)
	// The lexical representation of language is restricted to strings matching
	// the pattern `^[a-zA-Z]{1,8}(-[a-zA-Z0-9]{1,8})*$`. Extra structural or validator
	// constraints on subtags from BCP 47 must not be enforced at this level.
	if !languageRegex.MatchString(string(tok)) {
		return "", errors.New("value does not match the xsd:language pattern")
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.3)
	// Distinct case forms (such as 'MN' and 'mn') correspond to distinct values and
	// have distinct canonical forms. Returning the exact normalized token preserves
	// this case-sensitive identity.
	return Language(tok), nil
}

// String returns the canonical lexical representation of the Language value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.3.2)
//
// Parameters:
//
// Returns:
//   - string: The identity mapping of the language value as its canonical form.
func (l Language) String() string {
	return string(l)
}

// IsIdenticalWith checks if this Language value is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.3)
//
// Parameters:
//   - other: The other XSDValue to compare with the current Language value.
//
// Returns:
//   - bool: True if both values consist of the exact same sequence of characters.
func (l Language) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Language); ok {
		return l == o
	}
	return false
}

// Length returns the length of the Language value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1)
//
// Parameters:
//
// Returns:
//   - int: The number of XML characters (Unicode code points) contained in the value.
func (l Language) Length() int {
	return Token(l).Length()
}
