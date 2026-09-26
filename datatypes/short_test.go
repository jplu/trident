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
	"errors"
	"testing"
)

// TestParseShort verifies the lexical parsing of valid and invalid short values.
func TestParseShort(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Short
		wantErr   error
		expectErr bool
	}{
		{
			name:      "valid zero",
			input:     "0",
			want:      0,
			expectErr: false,
		},
		{
			name:      "valid positive",
			input:     "32767",
			want:      32767,
			expectErr: false,
		},
		{
			name:      "valid positive with plus sign",
			input:     "+1234",
			want:      1234,
			expectErr: false,
		},
		{
			name:      "valid negative",
			input:     "-32768",
			want:      -32768,
			expectErr: false,
		},
		{
			name:      "overflow positive",
			input:     "32768",
			wantErr:   ErrOutOfBoundsShort,
			expectErr: true,
		},
		{
			name:      "overflow negative",
			input:     "-32769",
			wantErr:   ErrOutOfBoundsShort,
			expectErr: true,
		},
		{
			name:      "invalid format non-numeric",
			input:     "abc",
			wantErr:   ErrParseInteger,
			expectErr: true,
		},
		{
			name:      "invalid format empty",
			input:     "",
			wantErr:   ErrParseInteger,
			expectErr: true,
		},
		{
			name:      "invalid format decimal",
			input:     "12.34",
			wantErr:   ErrParseInteger,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseShort(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("ParseShort(%q) expected error, got nil", tt.input)
				}
				if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParseShort(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseShort(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseShort(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestShortString tests the canonical string serialization of Short values.
func TestShortString(t *testing.T) {
	tests := []struct {
		name  string
		value Short
		want  string
	}{
		{
			name:  "zero",
			value: Short(0),
			want:  "0",
		},
		{
			name:  "positive maximum",
			value: Short(32767),
			want:  "32767",
		},
		{
			name:  "negative minimum",
			value: Short(-32768),
			want:  "-32768",
		},
		{
			name:  "general positive",
			value: Short(42),
			want:  "42",
		},
		{
			name:  "general negative",
			value: Short(-42),
			want:  "-42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.String(); got != tt.want {
				t.Fatalf("Short(%d).String() = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestShortIsIdenticalWith tests the identity comparison method for Short values.
func TestShortIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name  string
		value Short
		other XSDValue
		want  bool
	}{
		{
			name:  "identical same value",
			value: Short(100),
			other: Short(100),
			want:  true,
		},
		{
			name:  "different value same type",
			value: Short(100),
			other: Short(101),
			want:  false,
		},
		{
			name:  "different type same numeric value",
			value: Short(100),
			other: Int(100),
			want:  false,
		},
		{
			name:  "nil comparison",
			value: Short(100),
			other: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsIdenticalWith(tt.other); got != tt.want {
				t.Fatalf("Short(%d).IsIdenticalWith(%v) = %v, want %v", tt.value, tt.other, got, tt.want)
			}
		})
	}
}

// TestShortCompare tests the comparison relation of Short values against other XSDValues.
func TestShortCompare(t *testing.T) {
	tests := []struct {
		name      string
		value     Short
		other     XSDValue
		want      int
		expectErr bool
	}{
		{
			name:      "equal to Short",
			value:     Short(10),
			other:     Short(10),
			want:      0,
			expectErr: false,
		},
		{
			name:      "less than Short",
			value:     Short(10),
			other:     Short(20),
			want:      -1,
			expectErr: false,
		},
		{
			name:      "greater than Short",
			value:     Short(20),
			other:     Short(10),
			want:      1,
			expectErr: false,
		},
		{
			name:      "equal to DecimalProvider",
			value:     Short(15),
			other:     NewDecimalFromInt64(15),
			want:      0,
			expectErr: false,
		},
		{
			name:      "less than DecimalProvider",
			value:     Short(10),
			other:     NewDecimalFromInt64(15),
			want:      -1,
			expectErr: false,
		},
		{
			name:      "greater than DecimalProvider",
			value:     Short(20),
			other:     NewDecimalFromInt64(15),
			want:      1,
			expectErr: false,
		},
		{
			name:      "incomparable type",
			value:     Short(10),
			other:     String("10"),
			want:      0,
			expectErr: true,
		},
		{
			name:      "nil incomparable",
			value:     Short(10),
			other:     nil,
			want:      0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.Compare(tt.other)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("Short(%d).Compare(%v) expected error, got nil", tt.value, tt.other)
				}
				return
			}
			if err != nil {
				t.Fatalf("Short(%d).Compare(%v) unexpected error: %v", tt.value, tt.other, err)
			}
			if got != tt.want {
				t.Fatalf("Short(%d).Compare(%v) = %d, want %d", tt.value, tt.other, got, tt.want)
			}
		})
	}
}

// TestShortToDecimal tests the conversion of a Short value to an arbitrary-precision Decimal.
func TestShortToDecimal(t *testing.T) {
	tests := []struct {
		name  string
		value Short
		want  string
	}{
		{
			name:  "zero",
			value: Short(0),
			want:  "0",
		},
		{
			name:  "positive short",
			value: Short(12345),
			want:  "12345",
		},
		{
			name:  "negative short",
			value: Short(-12345),
			want:  "-12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := tt.value.ToDecimal()
			if got := dec.String(); got != tt.want {
				t.Fatalf("Short(%d).ToDecimal().String() = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
