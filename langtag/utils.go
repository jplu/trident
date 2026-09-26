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

package langtag

import (
	"strings"
	"unicode"
)

// isAlpha checks if a byte is an ASCII letter.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - b: The byte to evaluate.
//
// Returns:
//   - bool: True if the byte is in the ASCII range 'a'-'z' or 'A'-'Z', false otherwise.
func isAlpha(b byte) bool { return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') }

// isDigit checks if a byte is an ASCII digit.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - b: The byte to evaluate.
//
// Returns:
//   - bool: True if the byte is in the ASCII range '0'-'9', false otherwise.
func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// isAlphanum checks if a byte is an ASCII letter or digit.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - b: The byte to evaluate.
//
// Returns:
//   - bool: True if the byte is an ASCII letter or digit, false otherwise.
func isAlphanum(b byte) bool { return isAlpha(b) || isDigit(b) }

// isLangtagChar checks if a rune is a valid BCP 47 character (alphanumeric or hyphen).
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - r: The rune to evaluate.
//
// Returns:
//   - bool: True if the character is a US-ASCII alphanumeric character or a hyphen, false otherwise.
func isLangtagChar(r rune) bool {
	// Spec Rule: RFC 5646 (Section 2.1)
	// Language tags are sequences of US-ASCII alphanumeric characters and the hyphen character.
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-'
}

// isAlphabetic checks if a string contains only ASCII letters.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - s: The string to evaluate.
//
// Returns:
//   - bool: True if the string is non-empty and consists entirely of ASCII letters, false otherwise.
func isAlphabetic(s string) bool {
	if s == "" {
		return false
	}
	for i := range s {
		if !isAlpha(s[i]) {
			return false
		}
	}
	return true
}

// isNumeric checks if a string contains only ASCII digits.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - s: The string to evaluate.
//
// Returns:
//   - bool: True if the string is non-empty and consists entirely of ASCII digits, false otherwise.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for i := range s {
		if !isDigit(s[i]) {
			return false
		}
	}
	return true
}

// isAlphanumeric checks if a string contains only ASCII letters and digits.
//
// Specification Reference:
// RFC 5646 (Section 2.1)
//
// Parameters:
//   - s: The string to evaluate.
//
// Returns:
//   - bool: True if the string is non-empty and consists entirely of ASCII letters and digits, false otherwise.
func isAlphanumeric(s string) bool {
	if s == "" {
		return false
	}
	for i := range s {
		if !isAlphanum(s[i]) {
			return false
		}
	}
	return true
}

// writeTitleCase writes a string to a builder using title case (e.g., "Latn").
//
// Specification Reference:
// RFC 5646 (Section 2.1.1)
//
// Parameters:
//   - b: The strings.Builder pointer where the formatted characters are written.
//   - s: The string representing a subtag to format.
func writeTitleCase(b *strings.Builder, s string) {
	if len(s) == 0 {
		return
	}
	runes := []rune(s)
	// Step 1: Write titlecase first character
	// Convert the first character of the subtag to its titlecase equivalent and write it.
	b.WriteRune(unicode.ToTitle(runes[0]))
	if len(runes) > 1 {
		// Step 2: Write lowercase remaining characters
		// Convert all subsequent characters of the subtag to lowercase and write them.
		b.WriteString(strings.ToLower(string(runes[1:])))
	}
}
