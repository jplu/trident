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

// TestParseBoolean tests the ParseBoolean function with various valid and invalid inputs.
func TestParseBoolean(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Boolean
		wantErr bool
	}{
		{
			name:    "valid true literal 'true'",
			input:   "true",
			want:    Boolean(true),
			wantErr: false,
		},
		{
			name:    "valid true literal '1'",
			input:   "1",
			want:    Boolean(true),
			wantErr: false,
		},
		{
			name:    "valid false literal 'false'",
			input:   "false",
			want:    Boolean(false),
			wantErr: false,
		},
		{
			name:    "valid false literal '0'",
			input:   "0",
			want:    Boolean(false),
			wantErr: false,
		},
		{
			name:    "invalid uppercase 'TRUE'",
			input:   "TRUE",
			want:    Boolean(false),
			wantErr: true,
		},
		{
			name:    "invalid uppercase 'FALSE'",
			input:   "FALSE",
			want:    Boolean(false),
			wantErr: true,
		},
		{
			name:    "invalid numeric '2'",
			input:   "2",
			want:    Boolean(false),
			wantErr: true,
		},
		{
			name:    "invalid empty string",
			input:   "",
			want:    Boolean(false),
			wantErr: true,
		},
		{
			name:    "invalid string with whitespace",
			input:   " true ",
			want:    Boolean(false),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBoolean(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBoolean(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseBoolean(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestBooleanString tests the String method of the Boolean type.
func TestBooleanString(t *testing.T) {
	tests := []struct {
		name string
		b    Boolean
		want string
	}{
		{
			name: "true boolean string",
			b:    Boolean(true),
			want: "true",
		},
		{
			name: "false boolean string",
			b:    Boolean(false),
			want: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.String(); got != tt.want {
				t.Errorf("Boolean.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBooleanIsIdenticalWith tests the IsIdenticalWith method for comparing Boolean values.
func TestBooleanIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name  string
		b     Boolean
		other XSDValue
		want  bool
	}{
		{
			name:  "identical true values",
			b:     Boolean(true),
			other: Boolean(true),
			want:  true,
		},
		{
			name:  "identical false values",
			b:     Boolean(false),
			other: Boolean(false),
			want:  true,
		},
		{
			name:  "non-identical true vs false",
			b:     Boolean(true),
			other: Boolean(false),
			want:  false,
		},
		{
			name:  "non-identical false vs true",
			b:     Boolean(false),
			other: Boolean(true),
			want:  false,
		},
		{
			name:  "different XSDValue type (String)",
			b:     Boolean(true),
			other: String("true"),
			want:  false,
		},
		{
			name:  "nil XSDValue",
			b:     Boolean(true),
			other: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.IsIdenticalWith(tt.other); got != tt.want {
				t.Errorf("Boolean.IsIdenticalWith() = %v, want %v", got, tt.want)
			}
		})
	}
}
