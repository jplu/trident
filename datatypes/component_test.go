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

// TestVarietyConstants tests the variety constant values.
func TestVarietyConstants(t *testing.T) {
	tests := []struct {
		name     string
		variety  Variety
		expected int
	}{
		{"VarietyAbsent", VarietyAbsent, 0},
		{"VarietyAtomic", VarietyAtomic, 1},
		{"VarietyList", VarietyList, 2},
		{"VarietyUnion", VarietyUnion, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.variety) != tt.expected {
				t.Errorf("expected Variety %s to have integer value %d, got %d", tt.name, tt.expected, int(tt.variety))
			}
		})
	}
}

// mockFacet is a helper type implementing Facet.
type mockFacet struct{}

// Check is a helper method implementing Facet.Check.
func (m mockFacet) Check(_ XSDValue) error {
	return nil
}

// mockSimpleType is a helper type implementing SimpleTypeDefinition.
type mockSimpleType struct{}

// Name is a helper method returning the mock name.
func (m mockSimpleType) Name() string { return "mock" }

// TargetNamespace is a helper method returning the mock target namespace.
func (m mockSimpleType) TargetNamespace() string { return "http://example.org" }

// BaseType is a helper method returning the mock base type.
func (m mockSimpleType) BaseType() SimpleTypeDefinition { return nil }

// Variety is a helper method returning the mock variety.
func (m mockSimpleType) Variety() Variety { return VarietyAtomic }

// PrimitiveType is a helper method returning the mock primitive type.
func (m mockSimpleType) PrimitiveType() SimpleTypeDefinition { return nil }

// ItemType is a helper method returning the mock item type.
func (m mockSimpleType) ItemType() SimpleTypeDefinition { return nil }

// MemberTypes is a helper method returning mock member types.
func (m mockSimpleType) MemberTypes() []SimpleTypeDefinition { return nil }

// Facets is a helper method returning mock facets.
func (m mockSimpleType) Facets() []Facet { return nil }

// FundamentalFacets is a helper method returning fundamental facets.
func (m mockSimpleType) FundamentalFacets() FundamentalFacets {
	return DefaultFundamentalFacets()
}

// ValidateLexical is a helper method validating lexical representation.
//
//nolint:nilnil // Mock implementation explicitly returns nil value and nil error to test interface behavior.
func (m mockSimpleType) ValidateLexical(_ string) (XSDValue, error) {
	return nil, nil
}

// TestComponentInterfaces tests component interface implementations.
func TestComponentInterfaces(t *testing.T) {
	var f Facet = mockFacet{}
	if err := f.Check(nil); err != nil {
		t.Errorf("unexpected error from mockFacet.Check: %v", err)
	}

	var st SimpleTypeDefinition = mockSimpleType{}
	if st.Name() != "mock" {
		t.Errorf("expected Name 'mock', got %s", st.Name())
	}
	if st.TargetNamespace() != "http://example.org" {
		t.Errorf("expected TargetNamespace 'http://example.org', got %s", st.TargetNamespace())
	}
	if st.BaseType() != nil {
		t.Error("expected nil BaseType")
	}
	if st.Variety() != VarietyAtomic {
		t.Errorf("expected VarietyAtomic, got %v", st.Variety())
	}
	if st.PrimitiveType() != nil {
		t.Error("expected nil PrimitiveType")
	}
	if st.ItemType() != nil {
		t.Error("expected nil ItemType")
	}
	if st.MemberTypes() != nil {
		t.Error("expected nil MemberTypes")
	}
	if st.Facets() != nil {
		t.Error("expected nil Facets")
	}
	if st.FundamentalFacets() != DefaultFundamentalFacets() {
		t.Error("expected DefaultFundamentalFacets")
	}
	val, err := st.ValidateLexical("test")
	if val != nil || err != nil {
		t.Errorf("expected nil value and nil error from ValidateLexical, got val=%v, err=%v", val, err)
	}
}
