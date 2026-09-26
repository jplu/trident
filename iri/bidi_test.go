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
package iri

import (
	"errors"
	"strings"
	"testing"
)

// TestValidateBidiComponent tests the validation of individual IRI components against the bidirectional character
// rules.
func TestValidateBidiComponent(t *testing.T) {
	tests := []struct {
		name      string
		component string
		wantErr   bool
		errText   string
	}{
		{
			name:      "empty component",
			component: "",
			wantErr:   false,
		},
		{
			name:      "component with only LTR characters",
			component: "example-component",
			wantErr:   false,
		},
		{
			name:      "component with only RTL characters (Hebrew)",
			component: "\u05d0\u05d1\u05d2",
			wantErr:   false,
		},
		{
			name:      "component with only RTL characters (Arabic)",
			component: "\u0633\u0644\u0627\u0645",
			wantErr:   false,
		},
		{
			name:      "component with RTL and neutral characters (numbers)",
			component: "\u05d0123\u05d1",
			wantErr:   false,
		},
		{
			name:      "component with RTL and neutral characters (punctuation)",
			component: "\u05d0-\u05d1",
			wantErr:   false,
		},
		{
			name:      "component with only neutral characters",
			component: "123-.,_456",
			wantErr:   false,
		},
		{
			name:      "mixed LTR and RTL characters",
			component: "a\u05d0b",
			wantErr:   true,
			errText:   "bidirectional presentation violation (Rule 1) in component 'a\u05d0b': mixed left-to-right and right-to-left characters",
		},
		{
			name:      "mixed RTL and LTR characters",
			component: "\u05d0ab",
			wantErr:   true,
			errText:   "bidirectional presentation violation (Rule 1) in component '\u05d0ab': mixed left-to-right and right-to-left characters",
		},
		{
			name:      "RTL component starts with a number",
			component: "1\u05d0\u05d1",
			wantErr:   true,
			errText:   "bidirectional presentation violation (Rule 2) in component '1\u05d0\u05d1': right-to-left parts must start with right-to-left characters",
		},
		{
			name:      "RTL component ends with a number",
			component: "\u05d0\u05d11",
			wantErr:   true,
			errText:   "bidirectional presentation violation (Rule 2) in component '\u05d0\u05d11': right-to-left parts must end with right-to-left characters",
		},
		{
			name:      "RTL component starts with a neutral character",
			component: "-\u05d0\u05d1",
			wantErr:   true,
			errText:   "bidirectional presentation violation (Rule 2) in component '-\u05d0\u05d1': right-to-left parts must start with right-to-left characters",
		},
		{
			name:      "RTL component ends with a neutral character",
			component: "\u05d0\u05d1-",
			wantErr:   true,
			errText:   "bidirectional presentation violation (Rule 2) in component '\u05d0\u05d1-': right-to-left parts must end with right-to-left characters",
		},
		{
			name:      "component with RTL and skipped formatting characters (LRE)",
			component: "\u05d0\u202A\u05d1",
			wantErr:   false,
		},
		{
			name:      "component with RTL and skipped formatting characters (PDF)",
			component: "\u05d0\u202C\u05d1",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral EN (European Number)",
			component: "abc1",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral ES (European Number Separator)",
			component: "abc-",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral ET (European Number Terminator)",
			component: "abc$",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral AN (Arabic Number)",
			component: "abc\u0661",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral CS (Common Number Separator)",
			component: "abc.",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral B (Paragraph Separator)",
			component: "abc\n",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral S (Segment Separator)",
			component: "abc\t",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral WS (Whitespace)",
			component: "abc ",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral ON (Other Neutrals)",
			component: "abc@",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral BN (Boundary Neutral)",
			component: "abc\u00ad",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral NSM (Non-Spacing Mark)",
			component: "abc\u0300",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral Control",
			component: "abc\u0000",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral LRO",
			component: "abc\u202D",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral RLO",
			component: "abc\u202E",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral LRE",
			component: "abc\u202A",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral RLE",
			component: "abc\u202B",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral PDF",
			component: "abc\u202C",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral LRI",
			component: "abc\u2066",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral RLI",
			component: "abc\u2067",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral FSI",
			component: "abc\u2068",
			wantErr:   false,
		},
		{
			name:      "component with LTR and neutral PDI",
			component: "abc\u2069",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBidiComponent(tt.component)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBidiComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errText {
				t.Errorf("validateBidiComponent() error = %q, want %q", err.Error(), tt.errText)
			}
		})
	}
}

// TestValidateBidiHost tests the validation of an IRI host component against Bidi rules.
func TestValidateBidiHost(t *testing.T) {
	validRtlLabel := "\u05d0\u05d1\u05d2"
	invalidRtlLabelMixed := "a\u05d0b"
	invalidRtlLabelStart := "1\u05d0\u05d1"
	invalidRtlLabelEnd := "\u05d0\u05d1-"

	tests := []struct {
		name    string
		host    string
		wantErr bool
		errText string
	}{
		{
			name:    "empty host",
			host:    "",
			wantErr: false,
		},
		{
			name:    "standard LTR hostname",
			host:    "www.example.com",
			wantErr: false,
		},
		{
			name:    "hostname with all valid RTL labels",
			host:    validRtlLabel + "." + "\u0633\u0644\u0627\u0645",
			wantErr: false,
		},
		{
			name:    "hostname with mixed valid LTR and RTL labels",
			host:    "www." + validRtlLabel + ".com",
			wantErr: false,
		},
		{
			name:    "IPv6 literal should be ignored",
			host:    "[2001:db8::7]",
			wantErr: false,
		},
		{
			name:    "IPvFuture literal should be ignored",
			host:    "[v1.fe80::a+en0]",
			wantErr: false,
		},
		{
			name:    "host starting with [ but not ending with ]",
			host:    "[www.example.com",
			wantErr: false,
		},
		{
			name:    "host ending with ] but not starting with [",
			host:    "www.example.com]",
			wantErr: false,
		},
		{
			name:    "host with a mixed LTR/RTL label",
			host:    "www." + invalidRtlLabelMixed + ".com",
			wantErr: true,
			errText: "bidirectional presentation violation (Rule 1) in component 'invalidRtlLabelMixed in host 'www.invalidRtlLabelMixed.com'': Invalid IRI host label: mixed left-to-right and right-to-left characters",
		},
		{
			name:    "host with an RTL label starting with a number",
			host:    "www." + invalidRtlLabelStart + ".com",
			wantErr: true,
			errText: "bidirectional presentation violation (Rule 2) in component 'invalidRtlLabelStart in host 'www.invalidRtlLabelStart.com'': Invalid IRI host label: right-to-left parts must start with right-to-left characters",
		},
		{
			name:    "host with an RTL label ending with a neutral char",
			host:    "www." + invalidRtlLabelEnd + ".com",
			wantErr: true,
			errText: "bidirectional presentation violation (Rule 2) in component 'invalidRtlLabelEnd in host 'www.invalidRtlLabelEnd.com'': Invalid IRI host label: right-to-left parts must end with right-to-left characters",
		},
		{
			name:    "first label is invalid",
			host:    invalidRtlLabelMixed + ".example.com",
			wantErr: true,
			errText: "bidirectional presentation violation (Rule 1) in component 'invalidRtlLabelMixed in host 'invalidRtlLabelMixed.example.com'': Invalid IRI host label: mixed left-to-right and right-to-left characters",
		},
		{
			name:    "single invalid RTL label without dots",
			host:    invalidRtlLabelMixed,
			wantErr: true,
			errText: "bidirectional presentation violation (Rule 1) in component 'invalidRtlLabelMixed in host 'invalidRtlLabelMixed'': Invalid IRI host label: mixed left-to-right and right-to-left characters",
		},
		{
			name:    "host with multiple consecutive dots",
			host:    "www..com",
			wantErr: false,
		},
		{
			name:    "brackets with no content",
			host:    "[]",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedText := tt.errText
			expectedText = strings.ReplaceAll(
				expectedText,
				"invalidRtlLabelMixed",
				invalidRtlLabelMixed,
			)
			expectedText = strings.ReplaceAll(
				expectedText,
				"invalidRtlLabelStart",
				invalidRtlLabelStart,
			)
			expectedText = strings.ReplaceAll(
				expectedText,
				"invalidRtlLabelEnd",
				invalidRtlLabelEnd,
			)

			err := validateBidiHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBidiHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				var e *BidiGuidelineError
				if !errors.As(err, &e) {
					t.Fatalf(
						"validateBidiHost() returned wrong error type: got %T, want *BidiGuidelineError",
						err,
					)
				}
				if e.Error() != expectedText {
					t.Errorf("validateBidiHost() error = %q, want %q", e.Error(), expectedText)
				}
			}
		})
	}
}

// TestValidateBidiPath tests the validation of path segments against bidirectional rules.
func TestValidateBidiPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: false,
		},
		{
			name:    "valid LTR path",
			path:    "foo/bar/baz",
			wantErr: false,
		},
		{
			name:    "valid RTL path",
			path:    "\u05d0\u05d1/\u05d2\u05d3",
			wantErr: false,
		},
		{
			name:    "invalid RTL path segment",
			path:    "foo/a\u05d0b/bar",
			wantErr: true,
		},
		{
			name:    "path with empty segments (double slashes)",
			path:    "/foo//bar/",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBidiPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBidiPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateBidiPresentation tests the full Bidi validation process on parsed Refs.
func TestValidateBidiPresentation(t *testing.T) {
	tests := []struct {
		name    string
		iri     string
		wantErr bool
	}{
		{
			name:    "empty IRI",
			iri:     "",
			wantErr: false,
		},
		{
			name:    "valid LTR IRI with all components",
			iri:     "http://user:pass@example.com/foo/bar?q=1#frag",
			wantErr: false,
		},
		{
			name:    "valid mixed-direction IRI with consistent components",
			iri:     "http://example.com/\u05d0\u05d1\u05d2",
			wantErr: false,
		},
		{
			name:    "invalid bidi in userinfo",
			iri:     "http://a\u05d0b:pass@example.com/",
			wantErr: true,
		},
		{
			name:    "invalid bidi in host",
			iri:     "http://www.a\u05d0b.com/",
			wantErr: true,
		},
		{
			name:    "invalid bidi in path",
			iri:     "http://example.com/foo/a\u05d0b",
			wantErr: true,
		},
		{
			name:    "invalid bidi in query",
			iri:     "http://example.com/foo?a\u05d0b",
			wantErr: true,
		},
		{
			name:    "invalid bidi in fragment",
			iri:     "http://example.com/foo#a\u05d0b",
			wantErr: true,
		},
		{
			name:    "empty authority",
			iri:     "http://",
			wantErr: false,
		},
		{
			name:    "userinfo present but host empty",
			iri:     "http://userinfo@",
			wantErr: false,
		},
		{
			name:    "userinfo present with no bidi error and host has bidi error",
			iri:     "http://userinfo@www.a\u05d0b.com/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ParseRef(tt.iri)
			if err != nil {
				t.Fatalf("ParseRef() failed to parse test IRI: %v", err)
			}
			err = r.ValidateBidiPresentation()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBidiPresentation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBidiGuidelineError tests the Error formatting method of BidiGuidelineError.
func TestBidiGuidelineError(t *testing.T) {
	err := &BidiGuidelineError{
		Rule:      "Rule 1",
		Component: "testComponent",
		Message:   "testMessage",
	}
	expected := "bidirectional presentation violation (Rule 1) in component 'testComponent': testMessage"
	if err.Error() != expected {
		t.Errorf("BidiGuidelineError.Error() = %q, want %q", err.Error(), expected)
	}
}
