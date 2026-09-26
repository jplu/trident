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
	"testing"
)

// TestNewFacetWhiteSpace tests the creation and initialization of white-space facets.
func TestNewFacetWhiteSpace(t *testing.T) {
	tests := []struct {
		name     string
		val      WhiteSpaceValue
		fixed    bool
		wantName FacetName
	}{
		{
			name:     "preserve not fixed",
			val:      WhiteSpacePreserve,
			fixed:    false,
			wantName: NameWhiteSpace,
		},
		{
			name:     "replace fixed",
			val:      WhiteSpaceReplace,
			fixed:    true,
			wantName: NameWhiteSpace,
		},
		{
			name:     "collapse fixed",
			val:      WhiteSpaceCollapse,
			fixed:    true,
			wantName: NameWhiteSpace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFacetWhiteSpace(tt.val, tt.fixed)
			if f.Name != tt.wantName {
				t.Errorf("NewFacetWhiteSpace().Name = %v, want %v", f.Name, tt.wantName)
			}
			if f.Fixed != tt.fixed {
				t.Errorf("NewFacetWhiteSpace().Fixed = %v, want %v", f.Fixed, tt.fixed)
			}
			if f.Value != tt.val {
				t.Errorf("NewFacetWhiteSpace().Value = %v, want %v", f.Value, tt.val)
			}
		})
	}
}

// TestFacetWhiteSpaceCheck tests validation of XSD values against white-space facet rules.
func TestFacetWhiteSpaceCheck(t *testing.T) {
	tests := []struct {
		name      string
		facetVal  WhiteSpaceValue
		inputVal  XSDValue
		expectErr bool
	}{
		{
			name:      "String with preserve",
			facetVal:  WhiteSpacePreserve,
			inputVal:  String("test string"),
			expectErr: false,
		},
		{
			name:      "NormalizedString with replace",
			facetVal:  WhiteSpaceReplace,
			inputVal:  NormalizedString("test normalized"),
			expectErr: false,
		},
		{
			name:      "Token with collapse",
			facetVal:  WhiteSpaceCollapse,
			inputVal:  Token("test token"),
			expectErr: false,
		},
		{
			name:      "Name with preserve",
			facetVal:  WhiteSpacePreserve,
			inputVal:  Name("testName"),
			expectErr: false,
		},
		{
			name:      "NCName with collapse",
			facetVal:  WhiteSpaceCollapse,
			inputVal:  NCName("testNCName"),
			expectErr: false,
		},
		{
			name:      "AnyURI with collapse",
			facetVal:  WhiteSpaceCollapse,
			inputVal:  AnyURI("http://example.org"),
			expectErr: false,
		},
		{
			name:      "Language with preserve",
			facetVal:  WhiteSpacePreserve,
			inputVal:  Language("en-US"),
			expectErr: false,
		},
		{
			name:      "Non-string type (Boolean) with collapse valid",
			facetVal:  WhiteSpaceCollapse,
			inputVal:  Boolean(true),
			expectErr: false,
		},
		{
			name:      "Non-string type (Boolean) with preserve error",
			facetVal:  WhiteSpacePreserve,
			inputVal:  Boolean(true),
			expectErr: true,
		},
		{
			name:      "Non-string type (Byte) with replace error",
			facetVal:  WhiteSpaceReplace,
			inputVal:  Byte(42),
			expectErr: true,
		},
		{
			name:      "Non-string type (Int) with unknown value error",
			facetVal:  WhiteSpaceValue("custom"),
			inputVal:  Int(100),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFacetWhiteSpace(tt.facetVal, false)
			err := f.Check(tt.inputVal)
			if (err != nil) != tt.expectErr {
				t.Fatalf("FacetWhiteSpace.Check() error = %v, expectErr %v", err, tt.expectErr)
			}
			if tt.expectErr && err.Error() != "non-string types require whiteSpace=collapse" {
				t.Errorf("unexpected error message: %v", err.Error())
			}
		})
	}
}

// TestFacetWhiteSpaceNormalize tests the normalization of strings based on white-space facet modes.
func TestFacetWhiteSpaceNormalize(t *testing.T) {
	tests := []struct {
		name     string
		facetVal WhiteSpaceValue
		input    string
		want     string
	}{
		{
			name:     "replace facet mode",
			facetVal: WhiteSpaceReplace,
			input:    "hello\tworld\nfrom\rgo",
			want:     "hello world from go",
		},
		{
			name:     "collapse facet mode",
			facetVal: WhiteSpaceCollapse,
			input:    " \t hello \n  world \r from \t go  ",
			want:     "hello world from go",
		},
		{
			name:     "preserve facet mode",
			facetVal: WhiteSpacePreserve,
			input:    " \t hello \n  world \r from \t go  ",
			want:     " \t hello \n  world \r from \t go  ",
		},
		{
			name:     "default/unrecognized facet mode fallback",
			facetVal: WhiteSpaceValue("unknown"),
			input:    " \t hello \n world ",
			want:     " \t hello \n world ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFacetWhiteSpace(tt.facetVal, false)
			got := f.Normalize(tt.input)
			if got != tt.want {
				t.Errorf("FacetWhiteSpace.Normalize() = %q, want %q", got, tt.want)
			}
		})
	}
}
