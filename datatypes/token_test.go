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

// nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to
// test unexported functions.
package datatypes

import (
	"testing"
)

// TestParseToken tests the ParseToken function with various string inputs.
func TestParseToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Token
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple word",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "tabs and newlines collapse to single spaces",
			input:    "hello\t\n\rworld",
			expected: "hello world",
		},
		{
			name:     "consecutive spaces collapse",
			input:    "hello     world",
			expected: "hello world",
		},
		{
			name:     "only whitespace",
			input:    " \t \n \r ",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseToken(tc.input)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

// TestTokenString tests the String method of Token.
func TestTokenString(t *testing.T) {
	tok := Token("sample token")
	if got := tok.String(); got != "sample token" {
		t.Errorf("expected %q, got %q", "sample token", got)
	}
}

// TestTokenIsIdenticalWith tests the IsIdenticalWith method of Token across identical, non-identical, and non-token
// values.
func TestTokenIsIdenticalWith(t *testing.T) {
	tok := Token("tokenA")

	if !tok.IsIdenticalWith(Token("tokenA")) {
		t.Errorf("expected tok to be identical with Token(\"tokenA\")")
	}

	if tok.IsIdenticalWith(Token("tokenB")) {
		t.Errorf("expected tok not to be identical with Token(\"tokenB\")")
	}

	if tok.IsIdenticalWith(String("tokenA")) {
		t.Errorf("expected tok not to be identical with String value")
	}

	if tok.IsIdenticalWith(nil) {
		t.Errorf("expected tok not to be identical with nil")
	}
}

// TestTokenLength tests the Length method of Token with ASCII, empty, and multi-byte Unicode characters.
func TestTokenLength(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected int
	}{
		{
			name:     "empty token",
			token:    Token(""),
			expected: 0,
		},
		{
			name:     "ascii string",
			token:    Token("hello"),
			expected: 5,
		},
		{
			name:     "multibyte unicode characters",
			token:    Token("世界"),
			expected: 2,
		},
		{
			name:     "mixed unicode and ascii",
			token:    Token("Go语言"),
			expected: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.token.Length(); got != tc.expected {
				t.Errorf("expected length %d, got %d", tc.expected, got)
			}
		})
	}
}
