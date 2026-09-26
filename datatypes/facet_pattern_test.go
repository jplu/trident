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
	"strings"
	"testing"
)

// TestNewFacetPattern verifies that NewFacetPattern correctly initializes
// a facet with the expected name, fixed status, and matchers slice.
func TestNewFacetPattern(t *testing.T) {
	matcher, err := CompileRegex("[a-z]+")
	if err != nil {
		t.Fatalf("unexpected error compiling regex: %v", err)
	}

	matchers := []*PatternMatcher{matcher}
	facet := NewFacetPattern(matchers)

	if facet.Name != NamePattern {
		t.Errorf("expected facet name %s, got %s", NamePattern, facet.Name)
	}
	if facet.Fixed {
		t.Errorf("expected fixed to be false, got true")
	}
	if len(facet.Matchers) != 1 || facet.Matchers[0] != matcher {
		t.Errorf("expected matchers to match input slice")
	}
}

// TestFacetPatternCheck verifies that Check returns nil for any XSDValue,
// conforming to the rule that pattern facet constraints apply to the lexical
// space rather than the value space.
func TestFacetPatternCheck(t *testing.T) {
	facet := NewFacetPattern(nil)
	if err := facet.Check(String("test")); err != nil {
		t.Errorf("expected Check to return nil, got %v", err)
	}
}

// TestFacetPatternCheckLexical verifies that CheckLexical correctly validates
// literal string values against one or more regular expression pattern matchers.
func TestFacetPatternCheckLexical(t *testing.T) {
	mDigits, err := CompileRegex(`\d+`)
	if err != nil {
		t.Fatalf("unexpected error compiling digits regex: %v", err)
	}

	mLetters, err := CompileRegex(`[a-zA-Z]+`)
	if err != nil {
		t.Fatalf("unexpected error compiling letters regex: %v", err)
	}

	tests := []struct {
		name      string
		matchers  []*PatternMatcher
		literal   string
		expectErr bool
	}{
		{
			name:      "empty matchers slice returns nil",
			matchers:  nil,
			literal:   "anything",
			expectErr: false,
		},
		{
			name:      "empty matchers slice (len 0) returns nil",
			matchers:  []*PatternMatcher{},
			literal:   "anything",
			expectErr: false,
		},
		{
			name:      "single matcher matches",
			matchers:  []*PatternMatcher{mDigits},
			literal:   "12345",
			expectErr: false,
		},
		{
			name:      "single matcher does not match",
			matchers:  []*PatternMatcher{mDigits},
			literal:   "abc",
			expectErr: true,
		},
		{
			name:      "multiple matchers, first matches (exercises break)",
			matchers:  []*PatternMatcher{mDigits, mLetters},
			literal:   "12345",
			expectErr: false,
		},
		{
			name:      "multiple matchers, second matches (first fails then second succeeds)",
			matchers:  []*PatternMatcher{mDigits, mLetters},
			literal:   "abcXYZ",
			expectErr: false,
		},
		{
			name:      "multiple matchers, none match",
			matchers:  []*PatternMatcher{mDigits, mLetters},
			literal:   "123abc!",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			facet := NewFacetPattern(tc.matchers)
			err = facet.CheckLexical(tc.literal)
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error for literal %q, got nil", tc.literal)
				} else if !strings.Contains(err.Error(), "pattern violation: lexical value '"+tc.literal+"'") {
					t.Errorf("unexpected error message: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for literal %q: %v", tc.literal, err)
				}
			}
		})
	}
}
