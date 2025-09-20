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

package langtag

import (
	"strings"
	"unicode"
)

// isAlpha checks if a byte is an ASCII letter.
func isAlpha(b byte) bool { return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') }

// isDigit checks if a byte is an ASCII digit.
func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// isAlphanum checks if a byte is an ASCII letter or digit.
func isAlphanum(b byte) bool { return isAlpha(b) || isDigit(b) }

// isLangtagChar checks if a rune is a valid BCP 47 character (alphanumeric or hyphen).
func isLangtagChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-'
}

// isAlphabetic checks if a string contains only ASCII letters.
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
func writeTitleCase(b *strings.Builder, s string) {
	if len(s) == 0 {
		return
	}
	runes := []rune(s)
	b.WriteRune(unicode.ToTitle(runes[0]))
	if len(runes) > 1 {
		b.WriteString(strings.ToLower(string(runes[1:])))
	}
}
