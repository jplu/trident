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

// TestParseNonNegativeInteger validates parsing behavior for valid and invalid inputs.
func TestParseNonNegativeInteger(t *testing.T) {
	tests := []struct {
		input       string
		wantErr     bool
		expectedErr error
		expectedVal string
	}{
		{
			input:       "0",
			wantErr:     false,
			expectedVal: "0",
		},
		{
			input:       "+0",
			wantErr:     false,
			expectedVal: "0",
		},
		{
			input:       "12345",
			wantErr:     false,
			expectedVal: "12345",
		},
		{
			input:       "+42",
			wantErr:     false,
			expectedVal: "42",
		},
		{
			input:       "-1",
			wantErr:     true,
			expectedErr: ErrNotNonNegativeInteger,
		},
		{
			input:       "-99999999999999999999",
			wantErr:     true,
			expectedErr: ErrNotNonNegativeInteger,
		},
		{
			input:       "invalid",
			wantErr:     true,
			expectedErr: ErrParseInteger,
		},
		{
			input:       "",
			wantErr:     true,
			expectedErr: ErrParseInteger,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseNonNegativeInteger(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tt.input)
				}
				if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.input, err)
			}
			if got.String() != tt.expectedVal {
				t.Fatalf("expected string %q, got %q", tt.expectedVal, got.String())
			}
		})
	}
}

// TestNonNegativeIntegerString validates the canonical string serialization of NonNegativeInteger.
func TestNonNegativeIntegerString(t *testing.T) {
	val, err := ParseNonNegativeInteger("42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := val.String(); got != "42" {
		t.Fatalf("expected %q, got %q", "42", got)
	}
}

// TestNonNegativeIntegerIsIdenticalWith validates the identity comparison logic for NonNegativeInteger.
func TestNonNegativeIntegerIsIdenticalWith(t *testing.T) {
	nni1, err := ParseNonNegativeInteger("100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nni2, err := ParseNonNegativeInteger("100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nni3, err := ParseNonNegativeInteger("200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !nni1.IsIdenticalWith(nni2) {
		t.Fatal("expected identical values to be recognized as identical")
	}

	if nni1.IsIdenticalWith(nni3) {
		t.Fatal("expected different values to not be identical")
	}

	if nni1.IsIdenticalWith(String("100")) {
		t.Fatal("expected different types to not be identical")
	}

	if nni1.IsIdenticalWith(nil) {
		t.Fatal("expected comparison with nil to return false")
	}
}
