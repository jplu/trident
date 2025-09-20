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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package iri

import (
	"errors"
	"testing"
)

// TestValidateBidiComponent tests the validation of individual IRI components
// against the bidirectional character rules outlined in RFC 3987, Section 4.2.
// Rule 1: A component should not mix LTR and RTL characters.
// Rule 2: A component with RTL characters should begin and end with an RTL character.
func TestValidateBidiComponent(t *testing.T) {
	tests := []struct {
		name      string
		component string
		wantErr   bool
		errText   string
	}{
		// --- Valid Cases (RFC 3987, Section 4.2) ---
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
			component: "\u05d0\u05d1\u05d2", // "אבג"
			wantErr:   false,
		},
		{
			name:      "component with only RTL characters (Arabic)",
			component: "\u0633\u0644\u0627\u0645", // "سلام"
			wantErr:   false,
		},
		{
			name:      "component with RTL and neutral characters (numbers)",
			component: "\u05d0123\u05d1", // "א123ב"
			wantErr:   false,
		},
		{
			name:      "component with RTL and neutral characters (punctuation)",
			component: "\u05d0-\u05d1", // "א-ב"
			wantErr:   false,
		},
		{
			name:      "component with only neutral characters",
			component: "123-.,_456",
			wantErr:   false,
		},
		// --- Invalid Cases: Rule 1 Violation (Mixed LTR and RTL) ---
		{
			name:      "mixed LTR and RTL characters",
			component: "a\u05d0b", // "aאb"
			wantErr:   true,
			errText:   "Invalid IRI component: mixed left-to-right and right-to-left characters 'a\u05d0b'",
		},
		{
			name:      "mixed RTL and LTR characters",
			component: "\u05d0ab", // "אab"
			wantErr:   true,
			errText:   "Invalid IRI component: mixed left-to-right and right-to-left characters '\u05d0ab'",
		},
		// --- Invalid Cases: Rule 2 Violation (RTL component must start/end with RTL) ---
		{
			name:      "RTL component starts with a number",
			component: "1\u05d0\u05d1", // "1אב"
			wantErr:   true,
			errText:   "Invalid IRI component: right-to-left parts must start and end with right-to-left characters '1\u05d0\u05d1'",
		},
		{
			name:      "RTL component ends with a number",
			component: "\u05d0\u05d11", // "אב1"
			wantErr:   true,
			errText:   "Invalid IRI component: right-to-left parts must start and end with right-to-left characters '\u05d0\u05d11'",
		},
		{
			name:      "RTL component starts with a neutral character",
			component: "-\u05d0\u05d1", // "-אב"
			wantErr:   true,
			errText:   "Invalid IRI component: right-to-left parts must start and end with right-to-left characters '-\u05d0\u05d1'",
		},
		{
			name:      "RTL component ends with a neutral character",
			component: "\u05d0\u05d1-", // "אב-"
			wantErr:   true,
			errText:   "Invalid IRI component: right-to-left parts must start and end with right-to-left characters '\u05d0\u05d1-'",
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

// TestValidateBidiHost tests the validation of an IRI host component against
// Bidi rules. Per RFC 3987, Section 4.2, each dot-separated label of a
// hostname is treated as an individual component for Bidi validation.
// IP literals are exempt from these checks.
func TestValidateBidiHost(t *testing.T) {
	// Reusable valid/invalid labels for testing
	validRtlLabel := "\u05d0\u05d1\u05d2"   // "אבג"
	invalidRtlLabelMixed := "a\u05d0b"      // "aאb"
	invalidRtlLabelStart := "1\u05d0\u05d1" // "1אב"
	invalidRtlLabelEnd := "\u05d0\u05d1-"   // "אב-"

	tests := []struct {
		name    string
		host    string
		wantErr bool
		errText string
	}{
		// --- Valid Cases (RFC 3987, Section 4.2) ---
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
			host:    validRtlLabel + "." + "\u0633\u0644\u0627\u0645", // "אבג.سلام"
			wantErr: false,
		},
		{
			name:    "hostname with mixed valid LTR and RTL labels",
			host:    "www." + validRtlLabel + ".com",
			wantErr: false,
		},
		// --- Valid Cases: IP Literals (exempt from Bidi checks) ---
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
		// --- Invalid Cases (RFC 3987, Section 4.2) ---
		{
			name:    "host with a mixed LTR/RTL label",
			host:    "www." + invalidRtlLabelMixed + ".com",
			wantErr: true,
			errText: "Invalid IRI host label '" + invalidRtlLabelMixed + " in host 'www." + invalidRtlLabelMixed + ".com''",
		},
		{
			name:    "host with an RTL label starting with a number",
			host:    "www." + invalidRtlLabelStart + ".com",
			wantErr: true,
			errText: "Invalid IRI host label '" + invalidRtlLabelStart + " in host 'www." + invalidRtlLabelStart + ".com''",
		},
		{
			name:    "host with an RTL label ending with a neutral char",
			host:    "www." + invalidRtlLabelEnd + ".com",
			wantErr: true,
			errText: "Invalid IRI host label '" + invalidRtlLabelEnd + " in host 'www." + invalidRtlLabelEnd + ".com''",
		},
		{
			name:    "first label is invalid",
			host:    invalidRtlLabelMixed + ".example.com",
			wantErr: true,
			errText: "Invalid IRI host label '" + invalidRtlLabelMixed + " in host '" + invalidRtlLabelMixed + ".example.com''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBidiHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBidiHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				// The function is expected to wrap the component error with host context.
				var e *kindError
				if !errors.As(err, &e) {
					t.Fatalf("validateBidiHost() returned wrong error type: got %T, want *kindError", err)
				}
				if e.Error() != tt.errText {
					t.Errorf("validateBidiHost() error = %q, want %q", e.Error(), tt.errText)
				}
			}
		})
	}
}
