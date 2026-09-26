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
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// runeInterval represents the inclusive lower and upper bounds of a Unicode code point range as defined in the
// governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.1)
//
// Representation:
// A struct containing the inclusive lower and upper rune limits for a character class range.
type runeInterval struct {
	// Lo is the inclusive lower bound of the rune interval.
	lo rune
	// Hi is the inclusive upper bound of the rune interval.
	hi rune
}

// xmlNameStartCharRanges returns the Unicode ranges for XML NameStartChar (\i) as defined in the governing
// specification.
//
// Specification Reference:
// XML 1.0 (Fifth Edition) (Section 2.3) / W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
//
// Returns:
//   - []runeInterval: The slice of runeInterval values defining NameStartChar ranges.
func xmlNameStartCharRanges() []runeInterval {
	return []runeInterval{
		{'A', 'Z'},
		{'a', 'z'},
		{'_', '_'},
		{':', ':'},
		{0xC0, 0xD6},
		{0xD8, 0xF6},
		{0xF8, 0x2FF},
		{0x370, 0x37D},
		{0x37F, 0x1FFF},
		{0x200C, 0x200D},
		{0x2070, 0x218F},
		{0x2C00, 0x2FEF},
		{0x3001, 0xD7FF},
		{0xF900, 0xFDCF},
		{0xFDF0, 0xFFFD},
		{0x10000, 0xEFFFF},
	}
}

// xmlNameCharExtraRanges returns the additional Unicode ranges for XML NameChar (\c) as defined in the governing
// specification.
//
// Specification Reference:
// XML 1.0 (Fifth Edition) (Section 2.3) / W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
//
// Returns:
//   - []runeInterval: The slice of runeInterval values defining additional NameChar ranges.
func xmlNameCharExtraRanges() []runeInterval {
	return []runeInterval{
		{'-', '-'},
		{'.', '.'},
		{'0', '9'},
		{0xB7, 0xB7},
		{0x0300, 0x036F},
		{0x203F, 0x2040},
	}
}

// RuneSet represents an ordered, mathematical set of Unicode code points as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.1)
//
// Representation:
// A slice of unique, non-overlapping, normalized Unicode rune intervals [lo, hi].
type RuneSet []runeInterval

// Add inserts a closed range of runes [lo, hi] into the RuneSet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.1)
//
// Parameters:
//   - lo: The inclusive lower bound of the rune range to add.
//   - hi: The inclusive upper bound of the rune range to add.
func (rs *RuneSet) Add(lo, hi rune) {
	if lo > hi {
		return
	}
	*rs = append(*rs, runeInterval{lo, hi})
	rs.normalize()
}

// Subtract removes a closed range of runes [lo, hi] from the RuneSet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.1)
//
// Parameters:
//   - lo: The inclusive lower bound of the rune range to subtract.
//   - hi: The inclusive upper bound of the rune range to subtract.
func (rs *RuneSet) Subtract(lo, hi rune) {
	if lo > hi {
		return
	}
	var next RuneSet
	for _, iv := range *rs {
		if hi < iv.lo || lo > iv.hi {
			next = append(next, iv)
		} else {
			if iv.lo < lo {
				next = append(next, runeInterval{iv.lo, lo - 1})
			}
			if iv.hi > hi {
				next = append(next, runeInterval{hi + 1, iv.hi})
			}
		}
	}
	*rs = next
	// Implementation Note: Normalization step omission.
	// Normalization is intentionally omitted here because splitting or removing sections from an already sorted,
	// non-overlapping set guarantees the resulting set maintains normalized properties.
}

// normalize sorts and collapses overlapping or contiguous intervals within the RuneSet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.1).
func (rs *RuneSet) normalize() {
	if len(*rs) <= 1 {
		return
	}
	set := *rs

	// Step 1: Sort intervals.
	// Sort intervals by lower bound in ascending order.
	sort.Slice(set, func(i, j int) bool {
		return set[i].lo < set[j].lo
	})

	// Step 2: Merge overlapping intervals.
	// Consolidate contiguous or overlapping interval boundaries.
	var merged RuneSet
	curr := set[0]
	for i := 1; i < len(set); i++ {
		if set[i].lo <= curr.hi+1 {
			if set[i].hi > curr.hi {
				curr.hi = set[i].hi
			}
		} else {
			merged = append(merged, curr)
			curr = set[i]
		}
	}
	merged = append(merged, curr)
	*rs = merged
}

// String returns the canonical lexical representation of the RuneSet as an RE2-compatible character class string of the
// form `[...]`.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.1)
//
// Returns:
//   - string: The RE2-compatible character class literal string representing the compiled RuneSet.
func (rs *RuneSet) String() string {
	if rs == nil || len(*rs) == 0 {
		return `[^\x00-\x{10FFFF}]`
	}

	var b strings.Builder
	b.WriteByte('[')
	for _, iv := range *rs {
		if iv.lo == iv.hi {
			b.WriteString(escapeRune(iv.lo))
		} else {
			b.WriteString(escapeRune(iv.lo))
			b.WriteByte('-')
			b.WriteString(escapeRune(iv.hi))
		}
	}
	b.WriteByte(']')
	return b.String()
}

// escapeRune handles the escaping of special regex characters within RE2 character classes.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.3)
//
// Parameters:
//   - r: The rune to be escaped.
//
// Returns:
//   - string: The escaped string representation or hexadecimal sequence.
func escapeRune(r rune) string {
	switch r {
	case '\\', '[', ']', '-', '^':
		return `\` + string(r)
	}
	if r <= maxRune16 {
		return fmt.Sprintf(`\x{%04X}`, r)
	}
	return fmt.Sprintf(`\x{%06X}`, r)
}

// evaluateCharacterClass parses and evaluates an XSD character class expression, processing nested subtractions
// mathematically.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.1)
//
// Parameters:
//   - s: The raw character class string expression.
//
// Returns:
//   - *RuneSet: A pointer to the evaluated and simplified set of Unicode code point intervals.
//   - error: An error if character class parsing or subtraction evaluation fails.
func evaluateCharacterClass(s string) (*RuneSet, error) {
	if len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']' {
		s = s[1 : len(s)-1]
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix G.1)
	// Character class subtraction uses the '-[' operator where all characters matching the right set are subtracted
	// from the left set.
	subIdx := strings.Index(s, "-[")
	if subIdx == -1 {
		return parseSimpleClass(s)
	}

	leftStr := s[:subIdx]
	rightStr := s[subIdx+1:]

	leftSet, err := parseSimpleClass(leftStr)
	if err != nil {
		return nil, err
	}

	rightSet, err := evaluateCharacterClass(rightStr)
	if err != nil {
		return nil, err
	}

	for _, iv := range *rightSet {
		leftSet.Subtract(iv.lo, iv.hi)
	}

	return leftSet, nil
}

// parseSimpleClass parses character class contents without nested subtraction.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.1)
//
// Parameters:
//   - s: The flat character class content string.
//
// Returns:
//   - *RuneSet: A pointer to the populated RuneSet containing the parsed ranges.
//   - error: An error if a malformed escape or range sequence is encountered.
func parseSimpleClass(s string) (*RuneSet, error) {
	var rs RuneSet
	i := 0

	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == '\\' && i+size < len(s) {
			next, nSize := utf8.DecodeRuneInString(s[i+size:])
			err := handleEscapeSequence(&rs, s, i+size, next, nSize)
			if err != nil {
				return nil, err
			}
			i = advanceIndexAfterEscape(s, i, size, nSize, next)
		} else {
			if i+size < len(s) && s[i+size] == '-' && i+size+1 < len(s) {
				endR, endSize := utf8.DecodeRuneInString(s[i+size+1:])
				rs.Add(r, endR)
				i += size + 1 + endSize
			} else {
				rs.Add(r, r)
				i += size
			}
		}
	}
	return &rs, nil
}

// handleEscapeSequence processes a character class escape sequence and appends the matched characters.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2)
//
// Parameters:
//   - rs: The destination RuneSet to populate.
//   - s: The source pattern string.
//   - offset: The index position of the escape sequence.
//   - next: The first character following the backslash.
//   - nSize: The UTF-8 byte size of the next character.
//
// Returns:
//   - error: An error if a malformed Unicode escape is parsed.
func handleEscapeSequence(rs *RuneSet, s string, offset int, next rune, nSize int) error {
	switch next {
	case 'p', 'P':
		return handleUnicodePropertyEscape(rs, s, offset, next, nSize)
	case 'i', 'I', 'c', 'C':
		handleNameCharEscape(rs, next)
	case 'd', 'D':
		return addUnicodeProperty(rs, "Nd", next == 'D')
	case 'w', 'W':
		return handleWordCharEscape(rs, next)
	case 's', 'S':
		handleSpaceCharEscape(rs, next)
	default:
		handleDefaultEscape(rs, s, offset, next, nSize)
	}
	return nil
}

// handleUnicodePropertyEscape handles \p{...} and \P{...} property escapes.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2.2)
//
// Parameters:
//   - rs: The destination RuneSet to populate.
//   - s: The source pattern string.
//   - offset: The index position of the escape sequence.
//   - next: The character following the backslash ('p' or 'P').
//   - nSize: The UTF-8 byte size of the next character.
//
// Returns:
//   - error: An error if the Unicode property escape is malformed or unknown.
func handleUnicodePropertyEscape(rs *RuneSet, s string, offset int, next rune, nSize int) error {
	end := strings.Index(s[offset:], "}")
	if end == -1 {
		return errors.New("malformed unicode escape in character class")
	}
	propName := s[offset+nSize+1 : offset+end]
	return addUnicodeProperty(rs, propName, next == 'P')
}

// handleNameCharEscape handles \i, \I, \c, and \C character class escapes.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2.5)
//
// Parameters:
//   - rs: The destination RuneSet to populate.
//   - next: The escape character ('i', 'I', 'c', or 'C').
func handleNameCharEscape(rs *RuneSet, next rune) {
	var temp RuneSet
	temp = append(temp, xmlNameStartCharRanges()...)
	if next == 'c' || next == 'C' {
		temp = append(temp, xmlNameCharExtraRanges()...)
	}
	if next == 'I' || next == 'C' {
		*rs = append(*rs, complementRuneSet(temp)...)
	} else {
		*rs = append(*rs, temp...)
	}
	rs.normalize()
}

// handleWordCharEscape handles \w and \W character class escapes.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2.4)
//
// Parameters:
//   - rs: The destination RuneSet to populate.
//   - next: The escape character ('w' or 'W').
//
// Returns:
//   - error: An error if adding Unicode properties fails.
func handleWordCharEscape(rs *RuneSet, next rune) error {
	var temp RuneSet
	for _, cat := range []string{"P", "Z", "C"} {
		if err := addUnicodeProperty(&temp, cat, false); err != nil {
			return err
		}
	}
	if next == 'w' {
		*rs = append(*rs, complementRuneSet(temp)...)
	} else {
		*rs = append(*rs, temp...)
	}
	rs.normalize()
	return nil
}

// handleSpaceCharEscape handles \s and \S character class escapes.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2.3)
//
// Parameters:
//   - rs: The destination RuneSet to populate.
//   - next: The escape character ('s' or 'S').
func handleSpaceCharEscape(rs *RuneSet, next rune) {
	var temp RuneSet
	temp.Add(' ', ' ')
	temp.Add('\t', '\t')
	temp.Add('\n', '\n')
	temp.Add('\r', '\r')
	if next == 'S' {
		*rs = append(*rs, complementRuneSet(temp)...)
	} else {
		*rs = append(*rs, temp...)
	}
	rs.normalize()
}

// handleDefaultEscape handles standard character escapes and literal range escapes.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2)
//
// Parameters:
//   - rs: The destination RuneSet to populate.
//   - s: The source pattern string.
//   - offset: The index position of the escape sequence.
//   - next: The character following the backslash.
//   - nSize: The UTF-8 byte size of the next character.
func handleDefaultEscape(rs *RuneSet, s string, offset int, next rune, nSize int) {
	if offset+nSize < len(s) && s[offset+nSize] == '-' && offset+nSize+1 < len(s) {
		endR, _ := utf8.DecodeRuneInString(s[offset+nSize+1:])
		rs.Add(next, endR)
	} else {
		rs.Add(next, next)
	}
}

// complementRuneSet computes the logical complement of a RuneSet across the full Unicode range [0, unicode.MaxRune].
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.1)
//
// Parameters:
//   - rs: The input RuneSet to complement.
//
// Returns:
//   - RuneSet: The logical complement of the input set across the valid Unicode range [0, unicode.MaxRune].
func complementRuneSet(rs RuneSet) RuneSet {
	inverted := RuneSet{{0, unicode.MaxRune}}
	for _, iv := range rs {
		inverted.Subtract(iv.lo, iv.hi)
	}
	return inverted
}

// advanceIndexAfterEscape calculates the index position in the source string immediately following a parsed escape
// sequence.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2)
//
// Parameters:
//   - s: The source pattern string.
//   - i: The current parser index at the backslash.
//   - size: The size of the backslash character.
//   - nSize: The size of the character following the backslash.
//   - next: The character following the backslash.
//
// Returns:
//   - int: The absolute index position following the parsed sequence.
func advanceIndexAfterEscape(s string, i, size, nSize int, next rune) int {
	switch next {
	case 'p', 'P':
		end := strings.Index(s[i+size:], "}")
		return i + size + end + 1
	case 'i', 'I', 'c', 'C', 'w', 'W', 'd', 'D', 's', 'S':
		return i + size + nSize
	default:
		if i+size+nSize < len(s) && s[i+size+nSize] == '-' && i+size+nSize+1 < len(s) {
			_, endSize := utf8.DecodeRuneInString(s[i+size+nSize+1:])
			return i + size + nSize + 1 + endSize
		}
		return i + size + nSize
	}
}

// addUnicodeProperty maps an XSD Unicode block or category name to standard Go unicode.RangeTables.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2.2)
//
// Parameters:
//   - rs: The target RuneSet to populate.
//   - name: The Unicode category or block name.
//   - invert: A boolean flag indicating whether the character set should be complemented.
//
// Returns:
//   - error: An error if the specified Unicode property or block name is unrecognized.
func addUnicodeProperty(rs *RuneSet, name string, invert bool) error {
	// Implementation Note: Go unicode table alignment.
	// W3C XSD 1.1 Part 2 patterns allow Unicode block names prefixed with 'Is' (e.g., \p{IsBasicLatin}). Go's unicode
	// package identifies categories and blocks directly without the 'Is' prefix, so it is stripped to align the lookup.
	name = strings.TrimPrefix(name, "Is")

	if table, ok := unicode.Categories[name]; ok {
		applyRangeTable(rs, table, invert)
		return nil
	}
	if table, ok := unicode.Scripts[name]; ok {
		applyRangeTable(rs, table, invert)
		return nil
	}

	return fmt.Errorf("unknown unicode property or block: %s", name)
}

// applyRangeTable extracts code point ranges from a Go unicode.RangeTable and adds them to the destination RuneSet.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix G.4.2.2)
//
// Parameters:
//   - rs: The target RuneSet to populate.
//   - table: A pointer to the standard library Unicode RangeTable.
//   - invert: A boolean flag indicating whether the RangeTable characters should be complemented.
func applyRangeTable(rs *RuneSet, table *unicode.RangeTable, invert bool) {
	var temp RuneSet

	// Implementation Note: Bulk range extraction optimization.
	// Contiguous block characters with a stride of 1 are appended directly as intervals to avoid iterating over single
	// code points.
	for _, r16 := range table.R16 {
		if r16.Stride == 1 {
			temp = append(temp, runeInterval{rune(r16.Lo), rune(r16.Hi)})
		} else {
			for c := uint32(r16.Lo); c <= uint32(r16.Hi); c += uint32(r16.Stride) {
				temp = append(temp, runeInterval{rune(c), rune(c)})
			}
		}
	}

	for _, r32 := range table.R32 {
		if r32.Stride == 1 {
			temp = append(temp, runeInterval{rune(r32.Lo), rune(r32.Hi)})
		} else {
			for c := r32.Lo; c <= r32.Hi; c += r32.Stride {
				temp = append(temp, runeInterval{rune(c), rune(c)})
			}
		}
	}

	// Implementation Note: Set normalization.
	// Normalize the bulk collected ranges once after extraction.
	temp.normalize()

	if invert {
		*rs = append(*rs, complementRuneSet(temp)...)
	} else {
		*rs = append(*rs, temp...)
	}

	rs.normalize()
}
