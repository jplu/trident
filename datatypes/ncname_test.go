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

// TestParseNCName tests the parsing and validation of valid and invalid NCName string literals.
func TestParseNCName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    NCName
		wantErr bool
	}{
		{
			name:    "valid simple ASCII",
			input:   "validName",
			want:    NCName("validName"),
			wantErr: false,
		},
		{
			name:    "valid with whitespace collapsing",
			input:   "  validName  ",
			want:    NCName("validName"),
			wantErr: false,
		},
		{
			name:    "valid with underscores hyphens and digits",
			input:   "_my-name.1",
			want:    NCName("_my-name.1"),
			wantErr: false,
		},
		{
			name:    "valid unicode start character",
			input:   "élement",
			want:    NCName("élement"),
			wantErr: false,
		},
		{
			name:    "invalid empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid spaces only",
			input:   "   ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid starting with digit",
			input:   "1element",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid starting with colon",
			input:   ":element",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid containing colon",
			input:   "prefix:element",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid special characters",
			input:   "invalid$name",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNCName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseNCName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseNCName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestNCNameString tests the String canonical representation method for NCName.
func TestNCNameString(t *testing.T) {
	nc := NCName("testName")
	if got := nc.String(); got != "testName" {
		t.Errorf("NCName.String() = %q, want %q", got, "testName")
	}
}

// TestNCNameIsIdenticalWith tests identity comparison for NCName against identical, non-identical, and mismatched
// types.
func TestNCNameIsIdenticalWith(t *testing.T) {
	base := NCName("element")
	identical := NCName("element")
	different := NCName("other")
	mismatchedType := String("element")

	if !base.IsIdenticalWith(identical) {
		t.Errorf("expected %v to be identical with %v", base, identical)
	}
	if base.IsIdenticalWith(different) {
		t.Errorf("expected %v to not be identical with %v", base, different)
	}
	if base.IsIdenticalWith(mismatchedType) {
		t.Errorf("expected %v to not be identical with non-NCName %T", base, mismatchedType)
	}
}

// TestNCNameLength tests the character length calculation for NCName including unicode characters.
func TestNCNameLength(t *testing.T) {
	tests := []struct {
		name     string
		val      NCName
		expected int
	}{
		{
			name:     "ascii string",
			val:      NCName("simple"),
			expected: 6,
		},
		{
			name:     "unicode string",
			val:      NCName("élém"),
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.val.Length(); got != tt.expected {
				t.Errorf("NCName.Length() = %d, want %d", got, tt.expected)
			}
		})
	}
}
