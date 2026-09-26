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
	"math"
	"testing"
)

// TestParseInt verifies parsing valid and invalid lexical representations of xsd:int.
func TestParseInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Int
		wantErr error
	}{
		{
			name:    "valid zero",
			input:   "0",
			want:    Int(0),
			wantErr: nil,
		},
		{
			name:    "valid positive with plus sign",
			input:   "+42",
			want:    Int(42),
			wantErr: nil,
		},
		{
			name:    "valid negative",
			input:   "-100",
			want:    Int(-100),
			wantErr: nil,
		},
		{
			name:    "valid max inclusive",
			input:   "2147483647",
			want:    Int(math.MaxInt32),
			wantErr: nil,
		},
		{
			name:    "valid min inclusive",
			input:   "-2147483648",
			want:    Int(math.MinInt32),
			wantErr: nil,
		},
		{
			name:    "error overflow",
			input:   "2147483648",
			want:    Int(0),
			wantErr: ErrOutOfBoundsInt,
		},
		{
			name:    "error underflow",
			input:   "-2147483649",
			want:    Int(0),
			wantErr: ErrOutOfBoundsInt,
		},
		{
			name:    "error invalid format letters",
			input:   "abc",
			want:    Int(0),
			wantErr: ErrParseInteger,
		},
		{
			name:    "error invalid format empty",
			input:   "",
			want:    Int(0),
			wantErr: ErrParseInteger,
		},
		{
			name:    "error invalid format decimal point",
			input:   "12.34",
			want:    Int(0),
			wantErr: ErrParseInteger,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParseInt(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseInt(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseInt(%q) = %v, want = %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestIntString verifies the canonical string representation of Int values.
func TestIntString(t *testing.T) {
	tests := []struct {
		name  string
		value Int
		want  string
	}{
		{
			name:  "zero",
			value: Int(0),
			want:  "0",
		},
		{
			name:  "positive",
			value: Int(123456),
			want:  "123456",
		},
		{
			name:  "negative",
			value: Int(-98765),
			want:  "-98765",
		},
		{
			name:  "max value",
			value: Int(math.MaxInt32),
			want:  "2147483647",
		},
		{
			name:  "min value",
			value: Int(math.MinInt32),
			want:  "-2147483648",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.value.String()
			if got != tt.want {
				t.Errorf("Int(%d).String() = %q, want = %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestIntIsIdenticalWith validates identity checks between Int and other XSD values.
func TestIntIsIdenticalWith(t *testing.T) {
	val := Int(42)

	tests := []struct {
		name  string
		other XSDValue
		want  bool
	}{
		{
			name:  "identical value and type",
			other: Int(42),
			want:  true,
		},
		{
			name:  "different value same type",
			other: Int(43),
			want:  false,
		},
		{
			name:  "different type same numeric value",
			other: Short(42),
			want:  false,
		},
		{
			name:  "different type string",
			other: String("42"),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := val.IsIdenticalWith(tt.other)
			if got != tt.want {
				t.Errorf("Int(%d).IsIdenticalWith(%v) = %v, want = %v", val, tt.other, got, tt.want)
			}
		})
	}
}

// TestIntCompare tests comparing Int against same-type and cross-type values.
func TestIntCompare(t *testing.T) {
	val := Int(10)

	tests := []struct {
		name      string
		receiver  Int
		other     XSDValue
		wantRes   int
		expectErr bool
	}{
		{
			name:      "same type less",
			receiver:  val,
			other:     Int(20),
			wantRes:   -1,
			expectErr: false,
		},
		{
			name:      "same type equal",
			receiver:  val,
			other:     Int(10),
			wantRes:   0,
			expectErr: false,
		},
		{
			name:      "same type greater",
			receiver:  val,
			other:     Int(5),
			wantRes:   1,
			expectErr: false,
		},
		{
			name:      "decimal provider less",
			receiver:  val,
			other:     NewDecimalFromInt64(20),
			wantRes:   -1,
			expectErr: false,
		},
		{
			name:      "decimal provider equal",
			receiver:  val,
			other:     NewDecimalFromInt64(10),
			wantRes:   0,
			expectErr: false,
		},
		{
			name:      "decimal provider greater",
			receiver:  val,
			other:     NewDecimalFromInt64(5),
			wantRes:   1,
			expectErr: false,
		},
		{
			name:      "incomparable type",
			receiver:  val,
			other:     String("10"),
			wantRes:   0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.receiver.Compare(tt.other)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("Int(%d).Compare(%v) expected error, got nil", tt.receiver, tt.other)
				}
				return
			}
			if err != nil {
				t.Fatalf("Int(%d).Compare(%v) unexpected error: %v", tt.receiver, tt.other, err)
			}
			if res != tt.wantRes {
				t.Errorf("Int(%d).Compare(%v) = %d, want = %d", tt.receiver, tt.other, res, tt.wantRes)
			}
		})
	}
}

// TestIntToDecimal verifies conversion of Int to arbitrary-precision Decimal.
func TestIntToDecimal(t *testing.T) {
	val := Int(-12345)
	dec := val.ToDecimal()

	expected := NewDecimalFromInt64(-12345)
	if !dec.IsIdenticalWith(expected) {
		t.Errorf("Int(%d).ToDecimal() = %s, want = %s", val, dec.String(), expected.String())
	}
}
