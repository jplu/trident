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

// TestNewFacetTotalDigits verifies that NewFacetTotalDigits correctly initializes
// the totalDigits facet with the provided value and fixed flag.
func TestNewFacetTotalDigits(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		fixed    bool
		wantName FacetName
	}{
		{
			name:     "fixed true",
			val:      5,
			fixed:    true,
			wantName: NameTotalDigits,
		},
		{
			name:     "fixed false",
			val:      3,
			fixed:    false,
			wantName: NameTotalDigits,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facet := NewFacetTotalDigits(tt.val, tt.fixed)
			if facet.Name != tt.wantName {
				t.Errorf("NewFacetTotalDigits().Name = %v, want %v", facet.Name, tt.wantName)
			}
			if facet.Fixed != tt.fixed {
				t.Errorf("NewFacetTotalDigits().Fixed = %v, want %v", facet.Fixed, tt.fixed)
			}
			if facet.Value != tt.val {
				t.Errorf("NewFacetTotalDigits().Value = %v, want %v", facet.Value, tt.val)
			}
		})
	}
}

// TestFacetTotalDigitsCheck tests the validation logic of the totalDigits facet
// against various valid and invalid XSD values.
func TestFacetTotalDigitsCheck(t *testing.T) {
	decZero := NewDecimalFromInt64(0)
	decValid, err := ParseDecimal("123.45")
	if err != nil {
		t.Fatalf("unexpected error parsing decimal: %v", err)
	}
	decNegative, err := ParseDecimal("-12.34")
	if err != nil {
		t.Fatalf("unexpected error parsing decimal: %v", err)
	}

	tests := []struct {
		name      string
		facetVal  int
		input     XSDValue
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "unsupported datatype",
			facetVal:  5,
			input:     String("not-a-decimal"),
			wantErr:   true,
			errSubstr: "does not support totalDigits facet",
		},
		{
			name:      "zero value with facet limit 0 (violation)",
			facetVal:  0,
			input:     decZero,
			wantErr:   true,
			errSubstr: "totalDigits violation: value 0 exceeds totalDigits limit 0",
		},
		{
			name:     "zero value with facet limit >= 1 (valid)",
			facetVal: 1,
			input:    decZero,
			wantErr:  false,
		},
		{
			name:      "digit count exceeds limit (violation)",
			facetVal:  4,
			input:     decValid,
			wantErr:   true,
			errSubstr: "totalDigits violation: value 123.45 has 5 digits, limit is 4",
		},
		{
			name:     "digit count equals limit (valid)",
			facetVal: 5,
			input:    decValid,
			wantErr:  false,
		},
		{
			name:     "digit count strictly less than limit (valid)",
			facetVal: 6,
			input:    decValid,
			wantErr:  false,
		},
		{
			name:     "negative value with sign stripped (valid)",
			facetVal: 4,
			input:    decNegative,
			wantErr:  false,
		},
		{
			name:      "negative value exceeding limit (violation)",
			facetVal:  3,
			input:     decNegative,
			wantErr:   true,
			errSubstr: "totalDigits violation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facet := NewFacetTotalDigits(tt.facetVal, false)
			checkErr := facet.Check(tt.input)
			if (checkErr != nil) != tt.wantErr {
				t.Fatalf("FacetTotalDigits.Check() error = %v, wantErr %v", checkErr, tt.wantErr)
			}
			if tt.wantErr && tt.errSubstr != "" && !strings.Contains(checkErr.Error(), tt.errSubstr) {
				t.Errorf(
					"FacetTotalDigits.Check() error message = %q, want substring %q",
					checkErr.Error(),
					tt.errSubstr,
				)
			}
		})
	}
}

// TestNewFacetFractionDigits verifies that NewFacetFractionDigits correctly initializes
// the fractionDigits facet with the provided value and fixed flag.
func TestNewFacetFractionDigits(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		fixed    bool
		wantName FacetName
	}{
		{
			name:     "fixed true",
			val:      2,
			fixed:    true,
			wantName: NameFractionDigits,
		},
		{
			name:     "fixed false",
			val:      0,
			fixed:    false,
			wantName: NameFractionDigits,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facet := NewFacetFractionDigits(tt.val, tt.fixed)
			if facet.Name != tt.wantName {
				t.Errorf("NewFacetFractionDigits().Name = %v, want %v", facet.Name, tt.wantName)
			}
			if facet.Fixed != tt.fixed {
				t.Errorf("NewFacetFractionDigits().Fixed = %v, want %v", facet.Fixed, tt.fixed)
			}
			if facet.Value != tt.val {
				t.Errorf("NewFacetFractionDigits().Value = %v, want %v", facet.Value, tt.val)
			}
		})
	}
}

// TestFacetFractionDigitsCheck tests the validation logic of the fractionDigits facet
// against various valid and invalid XSD values.
func TestFacetFractionDigitsCheck(t *testing.T) {
	decInt := NewDecimalFromInt64(42)
	decWithFrac, err := ParseDecimal("12.345")
	if err != nil {
		t.Fatalf("unexpected error parsing decimal: %v", err)
	}

	tests := []struct {
		name      string
		facetVal  int
		input     XSDValue
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "unsupported datatype",
			facetVal:  2,
			input:     String("not-a-decimal"),
			wantErr:   true,
			errSubstr: "does not support fractionDigits facet",
		},
		{
			name:     "integer decimal without fractional part",
			facetVal: 0,
			input:    decInt,
			wantErr:  false,
		},
		{
			name:      "fraction digits exceed limit (violation)",
			facetVal:  2,
			input:     decWithFrac, // 3 fractional digits
			wantErr:   true,
			errSubstr: "fractionDigits violation: value 12.345 has 3 fraction digits, limit is 2",
		},
		{
			name:     "fraction digits equal limit (valid)",
			facetVal: 3,
			input:    decWithFrac, // 3 fractional digits
			wantErr:  false,
		},
		{
			name:     "fraction digits strictly less than limit (valid)",
			facetVal: 4,
			input:    decWithFrac, // 3 fractional digits
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facet := NewFacetFractionDigits(tt.facetVal, false)
			checkErr := facet.Check(tt.input)
			if (checkErr != nil) != tt.wantErr {
				t.Fatalf("FacetFractionDigits.Check() error = %v, wantErr %v", checkErr, tt.wantErr)
			}
			if tt.wantErr && tt.errSubstr != "" && !strings.Contains(checkErr.Error(), tt.errSubstr) {
				t.Errorf(
					"FacetFractionDigits.Check() error message = %q, want substring %q",
					checkErr.Error(),
					tt.errSubstr,
				)
			}
		})
	}
}
