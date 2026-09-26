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

// TestCompileRegexSuccess validates successful compilation of an XSD regular expression and verifies MatchString and
// String behavior.
func TestCompileRegexSuccess(t *testing.T) {
	pattern := "[a-z]+"
	pm, err := CompileRegex(pattern)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if pm == nil {
		t.Fatal("expected non-nil PatternMatcher")
	}
	if pm.String() != pattern {
		t.Fatalf("expected original pattern %q, got %q", pattern, pm.String())
	}
	if !pm.MatchString("abc") {
		t.Fatalf("expected %q to match %q", "abc", pattern)
	}
	if pm.MatchString("123") {
		t.Fatalf("expected %q not to match %q", "123", pattern)
	}
	if pm.MatchString("abc123") {
		t.Fatalf("expected implicit anchoring to reject %q for %q", "abc123", pattern)
	}
}

// TestCompileRegexTranslateError validates that CompileRegex returns an error when translation fails.
func TestCompileRegexTranslateError(t *testing.T) {
	invalidXSD := "[a-z"
	pm, err := CompileRegex(invalidXSD)
	if err == nil {
		t.Fatalf("expected translation error for %q, got nil", invalidXSD)
	}
	if pm != nil {
		t.Fatalf("expected nil PatternMatcher, got: %v", pm)
	}
}

// TestCompileRegexCompileError validates that CompileRegex returns an error when RE2 compilation fails despite
// translation succeeding.
func TestCompileRegexCompileError(t *testing.T) {
	invalidRE2 := "("
	pm, err := CompileRegex(invalidRE2)
	if err == nil {
		t.Fatalf("expected compile error for %q, got nil", invalidRE2)
	}
	if pm != nil {
		t.Fatalf("expected nil PatternMatcher, got: %v", pm)
	}
}
