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
package datatypes

import (
	"errors"
	"testing"
)

// TestParseByte tests parsing strings into Byte values.
func TestParseByte(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Byte
		wantErr error
	}{
		{
			name:  "valid zero",
			input: "0",
			want:  Byte(0),
		},
		{
			name:  "valid positive max",
			input: "127",
			want:  Byte(127),
		},
		{
			name:  "valid negative min",
			input: "-128",
			want:  Byte(-128),
		},
		{
			name:  "valid with leading plus",
			input: "+42",
			want:  Byte(42),
		},
		{
			name:  "valid with leading zeros",
			input: "007",
			want:  Byte(7),
		},
		{
			name:    "out of bounds upper",
			input:   "128",
			wantErr: ErrOutOfBoundsByte,
		},
		{
			name:    "out of bounds lower",
			input:   "-129",
			wantErr: ErrOutOfBoundsByte,
		},
		{
			name:    "invalid format letters",
			input:   "abc",
			wantErr: ErrParseInteger,
		},
		{
			name:    "invalid format decimal",
			input:   "12.3",
			wantErr: ErrParseInteger,
		},
		{
			name:    "invalid format empty",
			input:   "",
			wantErr: ErrParseInteger,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseByte(tt.input)
			if tt.wantErr != nil {
				switch {
				case err == nil:
					t.Fatalf("ParseByte(%q) expected error %v, got nil", tt.input, tt.wantErr)
				case !errors.Is(err, tt.wantErr):
					t.Errorf("ParseByte(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseByte(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseByte(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestByteString tests converting a Byte to its string representation.
func TestByteString(t *testing.T) {
	tests := []struct {
		val  Byte
		want string
	}{
		{val: Byte(0), want: "0"},
		{val: Byte(127), want: "127"},
		{val: Byte(-128), want: "-128"},
		{val: Byte(-1), want: "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.val.String(); got != tt.want {
				t.Errorf("Byte(%d).String() = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

// TestByteIsIdenticalWith tests the identity comparison of a Byte with another XSDValue.
func TestByteIsIdenticalWith(t *testing.T) {
	b10 := Byte(10)
	b10Dup := Byte(10)
	b20 := Byte(20)
	otherType := Int(10)
	strType := String("10")

	tests := []struct {
		name  string
		b     Byte
		other XSDValue
		want  bool
	}{
		{
			name:  "identical same value and type",
			b:     b10,
			other: b10Dup,
			want:  true,
		},
		{
			name:  "different value same type",
			b:     b10,
			other: b20,
			want:  false,
		},
		{
			name:  "different type same numeric value",
			b:     b10,
			other: otherType,
			want:  false,
		},
		{
			name:  "completely different type",
			b:     b10,
			other: strType,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.IsIdenticalWith(tt.other); got != tt.want {
				t.Errorf("Byte(%d).IsIdenticalWith(%v) = %v, want %v", tt.b, tt.other, got, tt.want)
			}
		})
	}
}

// TestByteCompare tests the comparison of a Byte with other XSDValues.
func TestByteCompare(t *testing.T) {
	b10 := Byte(10)
	b20 := Byte(20)
	b5 := Byte(5)

	dec10 := NewDecimalFromInt64(10)
	dec5 := NewDecimalFromInt64(5)

	tests := []struct {
		name    string
		b       Byte
		other   XSDValue
		wantCmp int
		wantErr bool
	}{
		{
			name:    "compare Byte less",
			b:       b5,
			other:   b10,
			wantCmp: -1,
		},
		{
			name:    "compare Byte equal",
			b:       b10,
			other:   b10,
			wantCmp: 0,
		},
		{
			name:    "compare Byte greater",
			b:       b20,
			other:   b10,
			wantCmp: 1,
		},
		{
			name:    "compare DecimalProvider less",
			b:       b5,
			other:   dec10,
			wantCmp: -1,
		},
		{
			name:    "compare DecimalProvider equal",
			b:       b10,
			other:   dec10,
			wantCmp: 0,
		},
		{
			name:    "compare DecimalProvider greater",
			b:       b20,
			other:   dec5,
			wantCmp: 1,
		},
		{
			name:    "compare incomparable type",
			b:       b10,
			other:   String("10"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmp, err := tt.b.Compare(tt.other)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Byte(%d).Compare(%v) expected error, got nil", tt.b, tt.other)
				}
				return
			}

			if err != nil {
				t.Fatalf("Byte(%d).Compare(%v) unexpected error: %v", tt.b, tt.other, err)
			}
			if gotCmp != tt.wantCmp {
				t.Errorf("Byte(%d).Compare(%v) = %d, want %d", tt.b, tt.other, gotCmp, tt.wantCmp)
			}
		})
	}
}

// TestByteToDecimal tests converting a Byte to a Decimal value.
func TestByteToDecimal(t *testing.T) {
	b := Byte(-42)
	dec := b.ToDecimal()

	if dec.String() != "-42" {
		t.Errorf("Byte(-42).ToDecimal().String() = %q, want %q", dec.String(), "-42")
	}
}
