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
	"strings"
	"testing"
)

// TestNewFacetLength tests the creation of length facets, ensuring proper initialization of Name, Fixed, and Value.
func TestNewFacetLength(t *testing.T) {
	tests := []struct {
		name      string
		val       int
		fixed     bool
		wantName  FacetName
		wantFixed bool
		wantVal   int
	}{
		{
			name:      "fixed length facet",
			val:       5,
			fixed:     true,
			wantName:  NameLength,
			wantFixed: true,
			wantVal:   5,
		},
		{
			name:      "unfixed length facet",
			val:       0,
			fixed:     false,
			wantName:  NameLength,
			wantFixed: false,
			wantVal:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFacetLength(tt.val, tt.fixed)
			if f.Name != tt.wantName {
				t.Errorf("NewFacetLength().Name = %v, want %v", f.Name, tt.wantName)
			}
			if f.Fixed != tt.wantFixed {
				t.Errorf("NewFacetLength().Fixed = %v, want %v", f.Fixed, tt.wantFixed)
			}
			if f.Value != tt.wantVal {
				t.Errorf("NewFacetLength().Value = %v, want %v", f.Value, tt.wantVal)
			}
		})
	}
}

// TestFacetLengthCheck tests the validation logic of length facets against various valid and invalid XSD values.
func TestFacetLengthCheck(t *testing.T) {
	facet := NewFacetLength(3, false)
	tests := []struct {
		name        string
		val         XSDValue
		wantErr     bool
		errContains string
	}{
		{
			name:        "non-LengthProvider value returns error",
			val:         Boolean(true),
			wantErr:     true,
			errContains: "does not support length facet",
		},
		{
			name:        "nil value returns error",
			val:         nil,
			wantErr:     true,
			errContains: "does not support length facet",
		},
		{
			name:        "length mismatch (shorter)",
			val:         String("ab"),
			wantErr:     true,
			errContains: "length violation: expected 3, got 2",
		},
		{
			name:        "length mismatch (longer)",
			val:         String("abcd"),
			wantErr:     true,
			errContains: "length violation: expected 3, got 4",
		},
		{
			name:    "exact length match with String",
			val:     String("abc"),
			wantErr: false,
		},
		{
			name:    "exact length match with HexBinary",
			val:     HexBinary([]byte{0x01, 0x02, 0x03}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := facet.Check(tt.val)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("FacetLength.Check() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("FacetLength.Check() error = %v, want substring %q", err, tt.errContains)
				}
			} else if err != nil {
				t.Fatalf("FacetLength.Check() unexpected error: %v", err)
			}
		})
	}
}

// TestNewFacetMinLength tests the creation of minimum length facets, ensuring proper initialization.
func TestNewFacetMinLength(t *testing.T) {
	tests := []struct {
		name      string
		val       int
		fixed     bool
		wantName  FacetName
		wantFixed bool
		wantVal   int
	}{
		{
			name:      "fixed minLength facet",
			val:       10,
			fixed:     true,
			wantName:  NameMinLength,
			wantFixed: true,
			wantVal:   10,
		},
		{
			name:      "unfixed minLength facet",
			val:       1,
			fixed:     false,
			wantName:  NameMinLength,
			wantFixed: false,
			wantVal:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFacetMinLength(tt.val, tt.fixed)
			if f.Name != tt.wantName {
				t.Errorf("NewFacetMinLength().Name = %v, want %v", f.Name, tt.wantName)
			}
			if f.Fixed != tt.wantFixed {
				t.Errorf("NewFacetMinLength().Fixed = %v, want %v", f.Fixed, tt.wantFixed)
			}
			if f.Value != tt.wantVal {
				t.Errorf("NewFacetMinLength().Value = %v, want %v", f.Value, tt.wantVal)
			}
		})
	}
}

// TestFacetMinLengthCheck tests the validation logic of minimum length facets against various values.
func TestFacetMinLengthCheck(t *testing.T) {
	facet := NewFacetMinLength(3, false)
	tests := []struct {
		name        string
		val         XSDValue
		wantErr     bool
		errContains string
	}{
		{
			name:        "non-LengthProvider value returns error",
			val:         Boolean(false),
			wantErr:     true,
			errContains: "does not support minLength facet",
		},
		{
			name:        "nil value returns error",
			val:         nil,
			wantErr:     true,
			errContains: "does not support minLength facet",
		},
		{
			name:        "length strictly less than minLength",
			val:         String("ab"),
			wantErr:     true,
			errContains: "minLength violation: expected at least 3, got 2",
		},
		{
			name:    "length equals minLength",
			val:     String("abc"),
			wantErr: false,
		},
		{
			name:    "length strictly greater than minLength",
			val:     String("abcd"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := facet.Check(tt.val)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("FacetMinLength.Check() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("FacetMinLength.Check() error = %v, want substring %q", err, tt.errContains)
				}
			} else if err != nil {
				t.Fatalf("FacetMinLength.Check() unexpected error: %v", err)
			}
		})
	}
}

// TestNewFacetMaxLength tests the creation of maximum length facets, ensuring proper initialization.
func TestNewFacetMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		val       int
		fixed     bool
		wantName  FacetName
		wantFixed bool
		wantVal   int
	}{
		{
			name:      "fixed maxLength facet",
			val:       20,
			fixed:     true,
			wantName:  NameMaxLength,
			wantFixed: true,
			wantVal:   20,
		},
		{
			name:      "unfixed maxLength facet",
			val:       5,
			fixed:     false,
			wantName:  NameMaxLength,
			wantFixed: false,
			wantVal:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFacetMaxLength(tt.val, tt.fixed)
			if f.Name != tt.wantName {
				t.Errorf("NewFacetMaxLength().Name = %v, want %v", f.Name, tt.wantName)
			}
			if f.Fixed != tt.wantFixed {
				t.Errorf("NewFacetMaxLength().Fixed = %v, want %v", f.Fixed, tt.wantFixed)
			}
			if f.Value != tt.wantVal {
				t.Errorf("NewFacetMaxLength().Value = %v, want %v", f.Value, tt.wantVal)
			}
		})
	}
}

// TestFacetMaxLengthCheck tests the validation logic of maximum length facets against various values.
func TestFacetMaxLengthCheck(t *testing.T) {
	facet := NewFacetMaxLength(3, false)

	tests := []struct {
		name        string
		val         XSDValue
		wantErr     bool
		errContains string
	}{
		{
			name:        "non-LengthProvider value returns error",
			val:         Boolean(true),
			wantErr:     true,
			errContains: "does not support maxLength facet",
		},
		{
			name:        "nil value returns error",
			val:         nil,
			wantErr:     true,
			errContains: "does not support maxLength facet",
		},
		{
			name:        "length strictly greater than maxLength",
			val:         String("abcd"),
			wantErr:     true,
			errContains: "maxLength violation: expected at most 3, got 4",
		},
		{
			name:    "length equals maxLength",
			val:     String("abc"),
			wantErr: false,
		},
		{
			name:    "length strictly less than maxLength",
			val:     String("ab"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := facet.Check(tt.val)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("FacetMaxLength.Check() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("FacetMaxLength.Check() error = %v, want substring %q", err, tt.errContains)
				}
			} else if err != nil {
				t.Fatalf("FacetMaxLength.Check() unexpected error: %v", err)
			}
		})
	}
}
