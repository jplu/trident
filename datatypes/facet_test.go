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

// mockLexicalFacet implements LexicalFacet for testing interface compliance.
type mockLexicalFacet struct {
	err error
}

// CheckLexical returns the mock error for lexical validation.
func (m mockLexicalFacet) CheckLexical(_ string) error {
	return m.err
}

// mockLengthProvider implements LengthProvider for testing interface compliance.
type mockLengthProvider struct {
	length int
}

// Length returns the mock length value.
func (m mockLengthProvider) Length() int {
	return m.length
}

// mockFacetDecimalProvider implements DecimalProvider for testing interface compliance.
type mockFacetDecimalProvider struct {
	dec Decimal
}

// ToDecimal returns the mock decimal value.
func (m mockFacetDecimalProvider) ToDecimal() Decimal {
	return m.dec
}

// mockTimezoneProvider implements TimezoneProvider for testing interface compliance.
type mockTimezoneProvider struct {
	tz *TimezoneOffset
}

// TimezoneOffset returns the mock timezone offset.
func (m mockTimezoneProvider) TimezoneOffset() *TimezoneOffset {
	return m.tz
}

// mockComparableProvider implements ComparableProvider for testing interface compliance.
type mockComparableProvider struct {
	result int
	err    error
}

// Compare returns the mock comparison result and error.
func (m mockComparableProvider) Compare(_ XSDValue) (int, error) {
	return m.result, m.err
}

// TestFacetNameConstants verifies that all facet name constants match their expected string values.
func TestFacetNameConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      FacetName
		expected string
	}{
		{"NameLength", NameLength, "length"},
		{"NameMinLength", NameMinLength, "minLength"},
		{"NameMaxLength", NameMaxLength, "maxLength"},
		{"NamePattern", NamePattern, "pattern"},
		{"NameEnumeration", NameEnumeration, "enumeration"},
		{"NameWhiteSpace", NameWhiteSpace, "whiteSpace"},
		{"NameMaxInclusive", NameMaxInclusive, "maxInclusive"},
		{"NameMaxExclusive", NameMaxExclusive, "maxExclusive"},
		{"NameMinExclusive", NameMinExclusive, "minExclusive"},
		{"NameMinInclusive", NameMinInclusive, "minInclusive"},
		{"NameTotalDigits", NameTotalDigits, "totalDigits"},
		{"NameFractionDigits", NameFractionDigits, "fractionDigits"},
		{"NameExplicitTimezone", NameExplicitTimezone, "explicitTimezone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.expected {
				t.Errorf("expected %s to have value %q, got %q", tt.name, tt.expected, string(tt.got))
			}
		})
	}
}

// TestBaseFacet validates the proper initialization and field values of BaseFacet.
func TestBaseFacet(t *testing.T) {
	tests := []struct {
		name     string
		facet    BaseFacet
		expName  FacetName
		expFixed bool
	}{
		{
			name: "Length facet not fixed",
			facet: BaseFacet{
				Name:  NameLength,
				Fixed: false,
			},
			expName:  NameLength,
			expFixed: false,
		},
		{
			name: "WhiteSpace facet fixed",
			facet: BaseFacet{
				Name:  NameWhiteSpace,
				Fixed: true,
			},
			expName:  NameWhiteSpace,
			expFixed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.facet.Name != tt.expName {
				t.Errorf("expected BaseFacet.Name to be %q, got %q", tt.expName, tt.facet.Name)
			}
			if tt.facet.Fixed != tt.expFixed {
				t.Errorf("expected BaseFacet.Fixed to be %v, got %v", tt.expFixed, tt.facet.Fixed)
			}
		})
	}
}

// TestLexicalFacetInterface verifies lexical checking behavior and error handling for LexicalFacet.
func TestLexicalFacetInterface(t *testing.T) {
	var _ LexicalFacet = mockLexicalFacet{}

	expectedErr := errors.New("lexical violation")
	facetWithErr := mockLexicalFacet{err: expectedErr}
	if err := facetWithErr.CheckLexical("sample"); !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	facetSuccess := mockLexicalFacet{err: nil}
	if err := facetSuccess.CheckLexical("sample"); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// TestLengthProviderInterface checks that LengthProvider correctly returns expected length values.
func TestLengthProviderInterface(t *testing.T) {
	var _ LengthProvider = mockLengthProvider{}

	tests := []struct {
		name      string
		provider  LengthProvider
		expLength int
	}{
		{
			name:      "zero length",
			provider:  mockLengthProvider{length: 0},
			expLength: 0,
		},
		{
			name:      "positive length",
			provider:  mockLengthProvider{length: 42},
			expLength: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.provider.Length(); got != tt.expLength {
				t.Errorf("expected Length() = %d, got %d", tt.expLength, got)
			}
		})
	}
}

// TestFacetDecimalProviderInterface validates that DecimalProvider correctly returns decimal representations.
func TestFacetDecimalProviderInterface(t *testing.T) {
	var _ = mockFacetDecimalProvider{}

	expectedDec := NewDecimalFromInt64(100)
	provider := mockFacetDecimalProvider{dec: expectedDec}

	got := provider.ToDecimal()
	if !got.IsIdenticalWith(expectedDec) {
		t.Errorf("expected ToDecimal() = %s, got %s", expectedDec.String(), got.String())
	}
}

// TestTimezoneProviderInterface verifies TimezoneProvider returns the expected timezone offset or nil.
func TestTimezoneProviderInterface(t *testing.T) {
	var _ TimezoneProvider = mockTimezoneProvider{}

	tz := GetUTC()
	zonedProvider := mockTimezoneProvider{tz: tz}
	if got := zonedProvider.TimezoneOffset(); got == nil || *got != *tz {
		t.Errorf("expected TimezoneOffset() = %v, got %v", tz, got)
	}

	unzonedProvider := mockTimezoneProvider{tz: nil}
	if got := unzonedProvider.TimezoneOffset(); got != nil {
		t.Errorf("expected nil TimezoneOffset(), got %v", got)
	}
}

// TestComparableProviderInterface tests comparison outputs and error handling for ComparableProvider.
func TestComparableProviderInterface(t *testing.T) {
	var _ ComparableProvider = mockComparableProvider{}

	customErr := errors.New("incomparable types")
	tests := []struct {
		name      string
		provider  ComparableProvider
		other     XSDValue
		expResult int
		expErr    error
	}{
		{
			name:      "equal comparison",
			provider:  mockComparableProvider{result: 0, err: nil},
			other:     String("test"),
			expResult: 0,
			expErr:    nil,
		},
		{
			name:      "less than comparison",
			provider:  mockComparableProvider{result: -1, err: nil},
			other:     String("test"),
			expResult: -1,
			expErr:    nil,
		},
		{
			name:      "greater than comparison",
			provider:  mockComparableProvider{result: 1, err: nil},
			other:     String("test"),
			expResult: 1,
			expErr:    nil,
		},
		{
			name:      "error during comparison",
			provider:  mockComparableProvider{result: 0, err: customErr},
			other:     String("test"),
			expResult: 0,
			expErr:    customErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.provider.Compare(tt.other)
			if !errors.Is(err, tt.expErr) {
				t.Errorf("expected error %v, got %v", tt.expErr, err)
			}
			if res != tt.expResult {
				t.Errorf("expected comparison result %d, got %d", tt.expResult, res)
			}
		})
	}
}
