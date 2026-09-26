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

// TestParseNameValid verifies that valid XML Name literals are successfully parsed.
func TestParseNameValid(t *testing.T) {
	tests := []struct {
		input    string
		expected Name
	}{
		{input: "validName", expected: "validName"},
		{input: "_validName", expected: "_validName"},
		{input: ":validName", expected: ":validName"},
		{input: "valid:name.with-hyphen_123", expected: "valid:name.with-hyphen_123"},
		{input: "  collapsible:Name  ", expected: "collapsible:Name"},
		{input: "\tprefix:_name\n", expected: "prefix:_name"},
		{input: "élément", expected: "élément"},
		{input: "Ω_name", expected: "Ω_name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseName(tt.input)
			if err != nil {
				t.Fatalf("ParseName(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParseName(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestParseNameInvalid verifies that invalid XML Name literals return an error.
func TestParseNameInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty string", input: ""},
		{name: "whitespace only", input: "   \t\n  "},
		{name: "leading digit", input: "1name"},
		{name: "leading hyphen", input: "-name"},
		{name: "leading period", input: ".name"},
		{name: "embedded whitespace", input: "invalid name"},
		{name: "illegal symbol", input: "name@domain"},
		{name: "illegal punctuation", input: "name#tag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseName(tt.input)
			if err == nil {
				t.Fatalf("ParseName(%q) expected error, got %q", tt.input, got)
			}
			if got != "" {
				t.Errorf("ParseName(%q) returned non-empty value %q on error", tt.input, got)
			}
		})
	}
}

// TestNameString verifies the String method of Name.
func TestNameString(t *testing.T) {
	tests := []struct {
		name     Name
		expected string
	}{
		{name: Name("simpleName"), expected: "simpleName"},
		{name: Name("prefix:local"), expected: "prefix:local"},
		{name: Name("_"), expected: "_"},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			if got := tt.name.String(); got != tt.expected {
				t.Errorf("Name(%q).String() = %q, expected %q", tt.name, got, tt.expected)
			}
		})
	}
}

// TestNameIsIdenticalWith verifies the value space identity comparison for Name.
func TestNameIsIdenticalWith(t *testing.T) {
	n1 := Name("item")
	n2 := Name("item")
	n3 := Name("other")
	diffType := Token("item")

	if !n1.IsIdenticalWith(n2) {
		t.Errorf("Name(%q).IsIdenticalWith(%q) expected true, got false", n1, n2)
	}

	if n1.IsIdenticalWith(n3) {
		t.Errorf("Name(%q).IsIdenticalWith(%q) expected false, got true", n1, n3)
	}

	if n1.IsIdenticalWith(diffType) {
		t.Errorf("Name(%q).IsIdenticalWith(Token(%q)) expected false, got true", n1, diffType)
	}

	if n1.IsIdenticalWith(nil) {
		t.Errorf("Name(%q).IsIdenticalWith(nil) expected false, got true", n1)
	}
}

// TestNameLength verifies character-based length measurement for Name.
func TestNameLength(t *testing.T) {
	tests := []struct {
		name     Name
		expected int
	}{
		{name: Name(""), expected: 0},
		{name: Name("abc"), expected: 3},
		{name: Name("a:b"), expected: 3},
		{name: Name("élément"), expected: 7},
		{name: Name("日本語"), expected: 3},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			if got := tt.name.Length(); got != tt.expected {
				t.Errorf("Name(%q).Length() = %d, expected %d", tt.name, got, tt.expected)
			}
		})
	}
}
