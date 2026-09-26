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
package iri

import (
	"errors"
	"strings"
	"testing"
)

// setupTestParser creates and configures a parser instance for testing.
// It uses a stringOutputBuffer to allow inspection of the parsed output.
func setupTestParser(input string, unchecked bool) *iriParser {
	return &iriParser{
		iri:       input,
		base:      &iriParserBase{hasBase: false},
		input:     newParserInput(input),
		output:    &stringOutputBuffer{builder: &strings.Builder{}},
		unchecked: unchecked,
	}
}

// assertError checks if the received error matches the expected error.
// It handles nil errors, sentinel errors, and custom *kindError types by message prefix.
func assertError(t *testing.T, got, want error) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Errorf("unexpected error: %v", got)
		}
		return
	}

	if got == nil {
		t.Errorf("expected error %v, but got nil", want)
		return
	}

	if errors.Is(want, errNoScheme) || errors.Is(want, errPathStartingWithSlashes) {
		if !errors.Is(got, want) {
			t.Errorf("got error %v, want sentinel error %v", got, want)
		}
		return
	}

	var wantKE *kindError
	if errors.As(want, &wantKE) {
		var gotKE *kindError
		if !errors.As(got, &gotKE) {
			t.Errorf("got error of type %T, want *kindError", got)
			return
		}
		if !strings.HasPrefix(gotKE.message, wantKE.message) {
			t.Errorf("got error message %q, want prefix %q", gotKE.message, wantKE.message)
		}
		return
	}

	if !errors.Is(got, want) {
		t.Errorf("got error %v, want %v", got, want)
	}
}

// mustParseRef parses a string as a Ref and fails the test if there is an error.
func mustParseRef(t *testing.T, s string) *Ref {
	t.Helper()
	r, err := ParseRef(s)
	if err != nil {
		t.Fatalf("mustParseRef failed for input '%s': %v", s, err)
	}
	return r
}

// mustParseIri parses a string as an Iri and fails the test if there is an error.
func mustParseIri(t *testing.T, s string) *Iri {
	t.Helper()
	i, err := ParseIri(s)
	if err != nil {
		t.Fatalf("mustParseIri failed for input '%s': %v", s, err)
	}
	return i
}

// mustParseAbsoluteIri parses a string into an Iri for tests, panicking if the string is invalid.
func mustParseAbsoluteIri(s string) *Iri {
	iri, err := ParseIri(s)
	if err != nil {
		panic("test setup failed: could not parse base IRI: " + s)
	}
	return iri
}
