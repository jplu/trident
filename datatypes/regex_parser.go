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
	"strings"
)

// escapeLen represents the character length of a standard backslash escape sequence (e.g., `\c` or `\i`).
const escapeLen = 2

// translateRegex parses an XML Schema regular expression string and converts it into a string compliant with Go's RE2
// engine.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G)
//
// Parameters:
//   - xsd: The raw XML Schema regular expression string to be translated.
//
// Returns:
//   - string: The translated Go RE2-compliant regular expression string.
//   - error: An error if the translation fails due to malformed syntax.
func translateRegex(xsd string) (string, error) {
	var out strings.Builder
	out.Grow(len(xsd) * builderGrowthFactor)

	i := 0
	for i < len(xsd) {
		r := xsd[i]

		switch {
		case r == '\\' && i+1 < len(xsd):
			// Step 1: Handle Escape Sequences
			// Parse standard backslash escapes and map them to Go RE2-compliant representations.
			i += handleEscape(&out, xsd, i)
		case r == '[':
			// Step 2: Handle Brackets
			// Parse character classes and resolve nested set subtractions.
			adv, err := handleBracket(&out, xsd, i)
			if err != nil {
				return "", err
			}
			i += adv
		default:
			// Step 3: Copy Literal Characters
			// Write unescaped normal characters directly to the output buffer.
			out.WriteByte(r)
			i++
		}
	}

	return out.String(), nil
}

// handleEscape processes an escape sequence starting with a backslash.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2)
//
// Parameters:
//   - out: The string builder accumulator where the translated output is written.
//   - xsd: The complete regular expression string being parsed.
//   - i: The current reading index in the xsd string.
//
// Returns:
//   - int: The number of bytes consumed from the input string.
func handleEscape(out *strings.Builder, xsd string, i int) int {
	next := xsd[i+1]
	switch next {
	case 'i':
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
		// The '\i' escape matches any character that can start an XML Name, matching NameStartChar as defined in XML
		// 1.0.
		out.WriteString(xmlNameStartCharClass)
		return escapeLen
	case 'I':
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
		// The '\I' escape matches any character that cannot start an XML Name.
		out.WriteString("[^" + xmlNameStartCharClass[1:])
		return escapeLen
	case 'c':
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
		// The '\c' escape matches any character that can occur in an XML Name, matching NameChar as defined in XML 1.0.
		out.WriteString(xmlNameCharClass)
		return escapeLen
	case 'C':
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
		// The '\C' escape matches any character that cannot occur in an XML Name.
		out.WriteString("[^" + xmlNameCharClass[1:])
		return escapeLen
	case 'w':
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
		// The '\w' escape matches any character except those in the Unicode categories for Punctuation (P), Separator
		// (Z), and Other (C).
		out.WriteString(`[^\p{P}\p{Z}\p{C}]`)
		return escapeLen
	case 'W':
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
		// The '\W' escape matches any character belonging to the Unicode categories for Punctuation (P), Separator (Z),
		// and Other (C).
		out.WriteString(`[\p{P}\p{Z}\p{C}]`)
		return escapeLen
	case 'p', 'P':
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.4.2.2)
		// Supports Unicode character property and block escapes of the form \p{Is...} and \P{Is...}.
		//
		// Implementation Note: Translation of Unicode property/block syntax for Go compatibility.
		// These escapes are translated to align with Go's native Unicode properties because Go's RE2 engine does not
		// recognize the 'Is' prefix.
		end := strings.Index(xsd[i:], "}")
		if end != -1 {
			escape := xsd[i : i+end+1]
			out.WriteString(translateUnicodeEscape(escape))
			return end + 1
		}
		out.WriteByte('\\')
		return 1
	default:
		out.WriteByte('\\')
		out.WriteByte(next)
		return escapeLen
	}
}

// handleBracket processes character classes enclosed in brackets.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.1)
//
// Parameters:
//   - out: The string builder accumulator where the translated output is written.
//   - xsd: The complete regular expression string being parsed.
//   - i: The current reading index in the xsd string.
//
// Returns:
//   - int: The number of bytes consumed from the input string.
//   - error: An error if character class evaluation or nested regex translation fails.
func handleBracket(out *strings.Builder, xsd string, i int) (int, error) {
	classStr, length, err := extractBracket(xsd[i:])
	if err != nil {
		return 0, err
	}

	if strings.Contains(classStr, "-[") {
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.4.1)
		// Character class subtraction is permitted in the form [class1-[class2]].
		//
		// Implementation Note: Processing set subtraction in Go's RE2 engine.
		// Since Go's RE2 parser does not natively support set subtraction, the classes must be parsed and
		// mathematically evaluated into a single, flattened RuneSet.
		rs, e := evaluateCharacterClass(classStr)
		if e != nil {
			return 0, e
		}
		out.WriteString(rs.String())
	} else {
		inside, _ := translateRegex(classStr[1 : len(classStr)-1])
		out.WriteByte('[')
		out.WriteString(inside)
		out.WriteByte(']')
	}
	return length, nil
}

// extractBracket scans a bracketed character class, properly handling nested brackets and skipping escaped bracket
// characters.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.1)
//
// Parameters:
//   - s: The raw regular expression string slice starting with an opening bracket.
//
// Returns:
//   - string: The extracted bracketed character class substring.
//   - int: The total number of bytes processed.
//   - error: An error if the closing bracket is not found.
func extractBracket(s string) (string, int, error) {
	if len(s) == 0 || s[0] != '[' {
		return "", 0, errors.New("expected '['")
	}

	i := 1
	depth := 1
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			i += escapeLen
			continue
		}
		switch s[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return s[:i+1], i + 1, nil
			}
		}
		i++
	}
	return "", 0, errors.New("unclosed character class bracket '['")
}
