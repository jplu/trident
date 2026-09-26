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

// TestParseNMTOKEN validates the ParseNMTOKEN parsing logic across valid and invalid inputs.
func TestParseNMTOKEN(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    NMTOKEN
		wantErr bool
	}{
		{
			name:    "valid simple token",
			input:   "validToken",
			want:    NMTOKEN("validToken"),
			wantErr: false,
		},
		{
			name:    "valid with colons, dots, hyphens, and digits",
			input:   "prefix:valid.name-123",
			want:    NMTOKEN("prefix:valid.name-123"),
			wantErr: false,
		},
		{
			name:    "valid with whitespace collapsed",
			input:   " \t\n\r token \t\n\r ",
			want:    NMTOKEN("token"),
			wantErr: false,
		},
		{
			name:    "valid starting with a digit",
			input:   "123token",
			want:    NMTOKEN("123token"),
			wantErr: false,
		},
		{
			name:    "valid with unicode character in name range",
			input:   "élement",
			want:    NMTOKEN("élement"),
			wantErr: false,
		},
		{
			name:    "invalid empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid whitespace only",
			input:   "   \t\n   ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid containing internal spaces",
			input:   "token with spaces",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid containing prohibited characters",
			input:   "token<name>",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNMTOKEN(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseNMTOKEN(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseNMTOKEN(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestNMTOKENString verifies the String method on NMTOKEN instances.
func TestNMTOKENString(t *testing.T) {
	val := NMTOKEN("example:token")
	expected := "example:token"
	if got := val.String(); got != expected {
		t.Errorf("NMTOKEN.String() = %q, want %q", got, expected)
	}
}

// TestNMTOKENIsIdenticalWith tests identity comparisons for NMTOKEN instances.
func TestNMTOKENIsIdenticalWith(t *testing.T) {
	tok := NMTOKEN("alpha")

	tests := []struct {
		name     string
		receiver NMTOKEN
		other    XSDValue
		want     bool
	}{
		{
			name:     "identical NMTOKEN values",
			receiver: tok,
			other:    NMTOKEN("alpha"),
			want:     true,
		},
		{
			name:     "different NMTOKEN values",
			receiver: tok,
			other:    NMTOKEN("beta"),
			want:     false,
		},
		{
			name:     "different type matching string",
			receiver: tok,
			other:    String("alpha"),
			want:     false,
		},
		{
			name:     "different type Token matching string",
			receiver: tok,
			other:    Token("alpha"),
			want:     false,
		},
		{
			name:     "nil other value",
			receiver: tok,
			other:    nil,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.receiver.IsIdenticalWith(tt.other); got != tt.want {
				t.Errorf("NMTOKEN.IsIdenticalWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNMTOKENLength tests the Length method calculation for NMTOKEN.
func TestNMTOKENLength(t *testing.T) {
	tests := []struct {
		name  string
		token NMTOKEN
		want  int
	}{
		{
			name:  "ascii characters",
			token: NMTOKEN("sample"),
			want:  6,
		},
		{
			name:  "unicode multi-byte characters",
			token: NMTOKEN("café"),
			want:  4,
		},
		{
			name:  "empty NMTOKEN",
			token: NMTOKEN(""),
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.Length(); got != tt.want {
				t.Errorf("NMTOKEN.Length() = %d, want %d", got, tt.want)
			}
		})
	}
}
