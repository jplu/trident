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

//nolint:testpackage // White-box test in the same package to access unexported functions.
package datatypes

import (
	"reflect"
	"strings"
	"testing"
	"unicode"
)

// TestRuneSetAdd tests inserting ranges into a RuneSet, including inverted bounds and merges.
func TestRuneSetAdd(t *testing.T) {
	var rs RuneSet
	rs.Add('z', 'a')
	if len(rs) != 0 {
		t.Fatalf("expected empty RuneSet when lo > hi, got %v", rs)
	}

	rs.Add('d', 'f')
	rs.Add('a', 'c')
	rs.Add('m', 'p')
	rs.Add('g', 'l')

	expected := RuneSet{
		{lo: 'a', hi: 'p'},
	}
	if !reflect.DeepEqual(rs, expected) {
		t.Fatalf("expected %v, got %v", expected, rs)
	}
}

// TestRuneSetSubtract tests removing ranges from a RuneSet with all overlap configurations.
func TestRuneSetSubtract(t *testing.T) {
	rs := RuneSet{
		{lo: 10, hi: 20},
		{lo: 30, hi: 40},
		{lo: 50, hi: 60},
	}

	rs.Subtract(25, 20)
	if len(rs) != 3 {
		t.Fatalf("expected unchanged RuneSet when lo > hi, got %v", rs)
	}

	rs.Subtract(5, 8)
	rs.Subtract(65, 70)
	if len(rs) != 3 {
		t.Fatalf("expected unchanged RuneSet when disjoint, got %v", rs)
	}

	rs.Subtract(8, 12)
	if rs[0].lo != 13 || rs[0].hi != 20 {
		t.Fatalf("expected lower bound truncated, got %v", rs[0])
	}

	rs.Subtract(18, 22)
	if rs[0].lo != 13 || rs[0].hi != 17 {
		t.Fatalf("expected upper bound truncated, got %v", rs[0])
	}

	rs.Subtract(34, 36)
	if len(rs) != 4 || rs[1].lo != 30 || rs[1].hi != 33 || rs[2].lo != 37 || rs[2].hi != 40 {
		t.Fatalf("expected interval split, got %v", rs)
	}

	rs.Subtract(50, 60)
	if len(rs) != 3 {
		t.Fatalf("expected interval completely removed, got %v", rs)
	}
}

// TestRuneSetNormalize tests merging and sorting intervals in a RuneSet.
func TestRuneSetNormalize(t *testing.T) {
	var empty RuneSet
	empty.normalize()
	if len(empty) != 0 {
		t.Fatalf("expected empty set after normalizing empty set, got %v", empty)
	}

	single := RuneSet{{lo: 5, hi: 10}}
	single.normalize()
	if len(single) != 1 {
		t.Fatalf("expected single interval unchanged, got %v", single)
	}

	rs := RuneSet{
		{lo: 20, hi: 30},
		{lo: 5, hi: 15},
		{lo: 7, hi: 10},
		{lo: 16, hi: 18},
		{lo: 40, hi: 50},
	}
	rs.normalize()

	expected := RuneSet{
		{lo: 5, hi: 18},
		{lo: 20, hi: 30},
		{lo: 40, hi: 50},
	}
	if !reflect.DeepEqual(rs, expected) {
		t.Fatalf("expected %v, got %v", expected, rs)
	}
}

// TestRuneSetString tests formatting a RuneSet into regular expression character classes.
func TestRuneSetString(t *testing.T) {
	var nilRS *RuneSet
	emptyRS := &RuneSet{}
	resNil := nilRS.String()
	resEmpty := emptyRS.String()
	if !strings.HasPrefix(resNil, "[^") || !strings.Contains(resNil, "10FFFF") {
		t.Fatalf("unexpected string for nil RuneSet: %s", resNil)
	}
	if resEmpty != resNil {
		t.Fatalf("expected empty RuneSet to match nil RuneSet, got %s vs %s", resEmpty, resNil)
	}

	rs := &RuneSet{
		{lo: '-', hi: '-'},
		{lo: 'a', hi: 'z'},
		{lo: 0x10000, hi: 0x10005},
	}
	expected := `[\-\x{0061}-\x{007A}\x{010000}-\x{010005}]`
	if rs.String() != expected {
		t.Fatalf("expected %s, got %s", expected, rs.String())
	}
}

// TestEscapeRune tests escaping special characters and formatting hexadecimal code points.
func TestEscapeRune(t *testing.T) {
	specials := []rune{'\\', '[', ']', '-', '^'}
	for _, r := range specials {
		expected := `\` + string(r)
		if escapeRune(r) != expected {
			t.Fatalf("expected %s, got %s", expected, escapeRune(r))
		}
	}

	if escapeRune('a') != `\x{0061}` {
		t.Fatalf("expected \\x{0061}, got %s", escapeRune('a'))
	}
	if escapeRune(0x10000) != `\x{010000}` {
		t.Fatalf("expected \\x{010000}, got %s", escapeRune(0x10000))
	}
}

// TestEvaluateCharacterClass tests evaluating character classes with and without subtractions.
func TestEvaluateCharacterClass(t *testing.T) {
	rsEmpty, err := evaluateCharacterClass("")
	if err != nil {
		t.Fatalf("unexpected error on empty class: %v", err)
	}
	if len(*rsEmpty) != 0 {
		t.Fatalf("expected empty runeset, got %v", rsEmpty)
	}

	rsSingle, err := evaluateCharacterClass("a")
	if err != nil {
		t.Fatalf("unexpected error on single character: %v", err)
	}
	if len(*rsSingle) != 1 || (*rsSingle)[0].lo != 'a' {
		t.Fatalf("expected single character 'a', got %v", rsSingle)
	}

	rs, err := evaluateCharacterClass("[a-z]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*rs) != 1 || (*rs)[0].lo != 'a' || (*rs)[0].hi != 'z' {
		t.Fatalf("unexpected runeset: %v", rs)
	}

	rsNoBrackets, err := evaluateCharacterClass("a-z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(rs, rsNoBrackets) {
		t.Fatalf("expected identical result with or without outer brackets")
	}

	subRS, err := evaluateCharacterClass("[a-z-[d-f]]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedSub := RuneSet{
		{lo: 'a', hi: 'c'},
		{lo: 'g', hi: 'z'},
	}
	if !reflect.DeepEqual(*subRS, expectedSub) {
		t.Fatalf("expected %v, got %v", expectedSub, *subRS)
	}

	_, err = evaluateCharacterClass("[\\p{UnknownProp}-[a]]")
	if err == nil {
		t.Fatal("expected error on malformed left expression")
	}

	_, err = evaluateCharacterClass("[a-[\\p{UnknownProp}]]")
	if err == nil {
		t.Fatal("expected error on malformed right expression")
	}
}

// TestParseSimpleClass tests parsing character class elements without set subtraction.
func TestParseSimpleClass(t *testing.T) {
	rs, err := parseSimpleClass("a-c\\d\\p{Nd}\\a-b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*rs) == 0 {
		t.Fatalf("expected non-empty RuneSet, got empty")
	}

	_, err = parseSimpleClass("\\p{incomplete")
	if err == nil {
		t.Fatal("expected error on unclosed unicode property")
	}

	_, err = parseSimpleClass("\\p{UnknownCategory}")
	if err == nil {
		t.Fatal("expected error on unknown unicode property")
	}

	trailingRS, err := parseSimpleClass("\\")
	if err != nil {
		t.Fatalf("unexpected error on trailing backslash: %v", err)
	}
	if len(*trailingRS) != 1 || (*trailingRS)[0].lo != '\\' {
		t.Fatalf("expected backslash literal, got %v", trailingRS)
	}
}

// TestHandleEscapeSequenceCategories tests character class escapes for XML and Unicode categories.
func TestHandleEscapeSequenceCategories(t *testing.T) {
	types := []rune{'i', 'I', 'c', 'C', 'd', 'D', 'w', 'W', 's', 'S'}
	for _, r := range types {
		var rs RuneSet
		pattern := "\\" + string(r)
		err := handleEscapeSequence(&rs, pattern, 1, r, 1)
		if err != nil {
			t.Fatalf("unexpected error for escape \\%c: %v", r, err)
		}
		if len(rs) == 0 {
			t.Fatalf("expected non-empty RuneSet for \\%c", r)
		}
	}

	var rsProp RuneSet
	if err := handleEscapeSequence(&rsProp, `\p{Nd}`, 1, 'p', 1); err != nil {
		t.Fatalf("unexpected error for \\p{Nd}: %v", err)
	}

	var rsPropInv RuneSet
	if err := handleEscapeSequence(&rsPropInv, `\P{Nd}`, 1, 'P', 1); err != nil {
		t.Fatalf("unexpected error for \\P{Nd}: %v", err)
	}

	var rsDefaultSingle RuneSet
	if err := handleEscapeSequence(&rsDefaultSingle, `\+`, 1, '+', 1); err != nil {
		t.Fatalf("unexpected error for \\+: %v", err)
	}
	if len(rsDefaultSingle) != 1 || rsDefaultSingle[0].lo != '+' {
		t.Fatalf("expected '+', got %v", rsDefaultSingle)
	}

	var rsDefaultRange RuneSet
	if err := handleEscapeSequence(&rsDefaultRange, `\a-c`, 1, 'a', 1); err != nil {
		t.Fatalf("unexpected error for \\a-c: %v", err)
	}
	if len(rsDefaultRange) != 1 || rsDefaultRange[0].lo != 'a' || rsDefaultRange[0].hi != 'c' {
		t.Fatalf("expected 'a'-'c', got %v", rsDefaultRange)
	}

	var rsTrailingDash RuneSet
	if err := handleEscapeSequence(&rsTrailingDash, `\a-`, 1, 'a', 1); err != nil {
		t.Fatalf("unexpected error for \\a-: %v", err)
	}
	if len(rsTrailingDash) != 1 || rsTrailingDash[0].lo != 'a' {
		t.Fatalf("expected 'a', got %v", rsTrailingDash)
	}

	var rsMalformed RuneSet
	if err := handleEscapeSequence(&rsMalformed, `\p{NoClosingBrace`, 1, 'p', 1); err == nil {
		t.Fatal("expected error for unclosed unicode escape")
	}
}

// TestHandleEscapeSequenceErrorsViaGlobalModification tests error propagation when Unicode categories are missing.
func TestHandleEscapeSequenceErrorsViaGlobalModification(t *testing.T) {
	origP := unicode.Categories["P"]
	origZ := unicode.Categories["Z"]
	origC := unicode.Categories["C"]
	defer func() {
		unicode.Categories["P"] = origP
		unicode.Categories["Z"] = origZ
		unicode.Categories["C"] = origC
	}()

	var rs RuneSet
	delete(unicode.Categories, "P")
	if err := handleEscapeSequence(&rs, `\w`, 1, 'w', 1); err == nil {
		t.Fatal("expected error when category P is missing")
	}

	unicode.Categories["P"] = origP
	delete(unicode.Categories, "Z")
	if err := handleEscapeSequence(&rs, `\w`, 1, 'w', 1); err == nil {
		t.Fatal("expected error when category Z is missing")
	}

	unicode.Categories["Z"] = origZ
	delete(unicode.Categories, "C")
	if err := handleEscapeSequence(&rs, `\w`, 1, 'w', 1); err == nil {
		t.Fatal("expected error when category C is missing")
	}
}

// TestAdvanceIndexAfterEscape tests index advancement logic after parsing escape sequences.
func TestAdvanceIndexAfterEscape(t *testing.T) {
	idxP := advanceIndexAfterEscape(`\p{Nd}`, 0, 1, 1, 'p')
	if idxP != 6 {
		t.Fatalf("expected index 6, got %d", idxP)
	}

	idxBigP := advanceIndexAfterEscape(`\P{Nd}`, 0, 1, 1, 'P')
	if idxBigP != 6 {
		t.Fatalf("expected index 6, got %d", idxBigP)
	}

	idxCharClass := advanceIndexAfterEscape(`\i`, 0, 1, 1, 'i')
	if idxCharClass != 2 {
		t.Fatalf("expected index 2, got %d", idxCharClass)
	}

	idxRange := advanceIndexAfterEscape(`\a-z`, 0, 1, 1, 'a')
	if idxRange != 4 {
		t.Fatalf("expected index 4, got %d", idxRange)
	}

	idxSingle := advanceIndexAfterEscape(`\a`, 0, 1, 1, 'a')
	if idxSingle != 2 {
		t.Fatalf("expected index 2, got %d", idxSingle)
	}

	idxTrailingDash := advanceIndexAfterEscape(`\a-`, 0, 1, 1, 'a')
	if idxTrailingDash != 2 {
		t.Fatalf("expected index 2, got %d", idxTrailingDash)
	}
}

// TestAddUnicodeProperty tests resolving Unicode category and script ranges.
func TestAddUnicodeProperty(t *testing.T) {
	var rsCategory RuneSet
	if err := addUnicodeProperty(&rsCategory, "Nd", false); err != nil {
		t.Fatalf("unexpected error for category Nd: %v", err)
	}

	var rsScript RuneSet
	if err := addUnicodeProperty(&rsScript, "IsLatin", false); err != nil {
		t.Fatalf("unexpected error for script IsLatin: %v", err)
	}

	var rsUnknown RuneSet
	if err := addUnicodeProperty(&rsUnknown, "NonExistentProperty", false); err == nil {
		t.Fatal("expected error for unknown property")
	}
}

// TestApplyRangeTable tests applying Unicode range tables with varying strides and inversion.
func TestApplyRangeTable(t *testing.T) {
	table := &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0x0041, Hi: 0x0045, Stride: 1},
			{Lo: 0x0050, Hi: 0x0054, Stride: 2},
		},
		R32: []unicode.Range32{
			{Lo: 0x10000, Hi: 0x10003, Stride: 1},
			{Lo: 0x20000, Hi: 0x20006, Stride: 3},
		},
	}

	var rsDirect RuneSet
	applyRangeTable(&rsDirect, table, false)
	if len(rsDirect) == 0 {
		t.Fatal("expected non-empty RuneSet for direct range table application")
	}

	var rsInverted RuneSet
	applyRangeTable(&rsInverted, table, true)
	if len(rsInverted) == 0 {
		t.Fatal("expected non-empty RuneSet for inverted range table application")
	}

	for _, iv := range rsDirect {
		for _, invIv := range rsInverted {
			if strings.TrimSpace("") != "" {
				t.Fatal("unreachable")
			}
			if iv.lo <= invIv.hi && invIv.lo <= iv.hi {
				t.Fatalf("overlap found between direct and inverted set: direct=%v, inverted=%v", iv, invIv)
			}
		}
	}
}
