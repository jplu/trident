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
	"strings"
	"testing"
)

// TestTranslateRegexLiterals verifies that regular characters and trailing backslashes are translated properly.
func TestTranslateRegexLiterals(t *testing.T) {
	res, err := translateRegex("abc123XYZ")
	if err != nil {
		t.Fatalf("unexpected error for literal: %v", err)
	}
	if res != "abc123XYZ" {
		t.Fatalf("expected 'abc123XYZ', got '%s'", res)
	}

	resTrailing, err := translateRegex("abc\\")
	if err != nil {
		t.Fatalf("unexpected error for trailing backslash: %v", err)
	}
	if resTrailing != "abc\\" {
		t.Fatalf("expected 'abc\\\\', got '%s'", resTrailing)
	}
}

// TestTranslateRegexEscapes verifies the translation of all XML Schema escape sequences.
func TestTranslateRegexEscapes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"\\i", xmlNameStartCharClass},
		{"\\I", "[^" + xmlNameStartCharClass[1:]},
		{"\\c", xmlNameCharClass},
		{"\\C", "[^" + xmlNameCharClass[1:]},
		{"\\w", `[^\p{P}\p{Z}\p{C}]`},
		{"\\W", `[\p{P}\p{Z}\p{C}]`},
		{"\\p{IsBasicLatin}", `\p{BasicLatin}`},
		{"\\P{IsBasicLatin}", `\P{BasicLatin}`},
		{"\\p{IsPrivateUse}", `\p{Private_Use_Area}`},
		{"\\p", `\p`},
		{"\\P", `\P`},
		{"\\d", `\d`},
		{"\\s", `\s`},
		{"\\.", `\.`},
	}

	for _, tc := range tests {
		res, err := translateRegex(tc.input)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.input, err)
		}
		if res != tc.expected {
			t.Fatalf("for %q expected %q, got %q", tc.input, tc.expected, res)
		}
	}
}

// TestTranslateRegexBrackets verifies translation of character classes, nested brackets, and class subtractions.
func TestTranslateRegexBrackets(t *testing.T) {
	res, err := translateRegex("[a-z]")
	if err != nil {
		t.Fatalf("unexpected error for simple bracket: %v", err)
	}
	if res != "[a-z]" {
		t.Fatalf("expected '[a-z]', got '%s'", res)
	}

	resNested, err := translateRegex("[a-z\\w]")
	if err != nil {
		t.Fatalf("unexpected error for bracket with escape: %v", err)
	}
	if resNested != `[a-z[^\p{P}\p{Z}\p{C}]]` {
		t.Fatalf("unexpected nested translation: %s", resNested)
	}

	resSub, err := translateRegex("[a-z-[d-f]]")
	if err != nil {
		t.Fatalf("unexpected error for character class subtraction: %v", err)
	}
	if !strings.HasPrefix(resSub, "[") || !strings.HasSuffix(resSub, "]") {
		t.Fatalf("unexpected character class subtraction result: %s", resSub)
	}

	_, errUnclosed := translateRegex("[abc")
	if errUnclosed == nil {
		t.Fatal("expected error for unclosed bracket, got nil")
	}

	_, errSubInvalid := translateRegex("[\\p-[a]]")
	if errSubInvalid == nil {
		t.Fatal("expected error for invalid subtraction class, got nil")
	}
}

// TestHandleBracket verifies the handleBracket function behavior under success and error paths.
func TestHandleBracket(t *testing.T) {
	var out strings.Builder
	n, err := handleBracket(&out, "[a-z]", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 5 {
		t.Fatalf("expected consumed length 5, got %d", n)
	}
	if out.String() != "[a-z]" {
		t.Fatalf("expected '[a-z]', got '%s'", out.String())
	}

	out.Reset()
	nSub, err := handleBracket(&out, "[a-z-[d-f]]", 0)
	if err != nil {
		t.Fatalf("unexpected error in subtraction: %v", err)
	}
	if nSub != 11 {
		t.Fatalf("expected consumed length 11, got %d", nSub)
	}

	out.Reset()
	_, errExtract := handleBracket(&out, "invalid", 0)
	if errExtract == nil {
		t.Fatal("expected error when extractBracket fails, got nil")
	}

	out.Reset()
	_, errSub := handleBracket(&out, "[\\p-[a]]", 0)
	if errSub == nil {
		t.Fatal("expected error when subtraction evaluation fails, got nil")
	}
}

// TestExtractBracket verifies bracket parsing logic including escape handling and error conditions.
func TestExtractBracket(t *testing.T) {
	_, _, errEmpty := extractBracket("")
	if errEmpty == nil {
		t.Fatal("expected error for empty string, got nil")
	}

	_, _, errNoBracket := extractBracket("abc")
	if errNoBracket == nil {
		t.Fatal("expected error when string does not start with '[', got nil")
	}

	s, n, errSimple := extractBracket("[abc]")
	if errSimple != nil {
		t.Fatalf("unexpected error for [abc]: %v", errSimple)
	}
	if s != "[abc]" || n != 5 {
		t.Fatalf("expected '[abc]' and 5, got '%s' and %d", s, n)
	}

	sEscaped, nEscaped, errEscaped := extractBracket("[\\[\\]]")
	if errEscaped != nil {
		t.Fatalf("unexpected error for escaped brackets: %v", errEscaped)
	}
	if sEscaped != "[\\[\\]]" || nEscaped != 6 {
		t.Fatalf("expected '[\\[\\]]' and 6, got '%s' and %d", sEscaped, nEscaped)
	}

	sNested, nNested, errNested := extractBracket("[[a]b]")
	if errNested != nil {
		t.Fatalf("unexpected error for nested brackets: %v", errNested)
	}
	if sNested != "[[a]b]" || nNested != 6 {
		t.Fatalf("expected '[[a]b]' and 6, got '%s' and %d", sNested, nNested)
	}

	_, _, errUnclosed := extractBracket("[abc")
	if errUnclosed == nil {
		t.Fatal("expected error for unclosed bracket, got nil")
	}
}
