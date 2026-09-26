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

// TestDefaultFundamentalFacets verifies that DefaultFundamentalFacets returns the expected default values.
func TestDefaultFundamentalFacets(t *testing.T) {
	facets := DefaultFundamentalFacets()

	if facets.Ordered != OrderedFalse {
		t.Errorf("expected Ordered to be OrderedFalse (%d), got %d", OrderedFalse, facets.Ordered)
	}
	if facets.Bounded != false {
		t.Errorf("expected Bounded to be false, got %t", facets.Bounded)
	}
	if facets.Cardinality != CountablyInfinite {
		t.Errorf("expected Cardinality to be CountablyInfinite (%d), got %d", CountablyInfinite, facets.Cardinality)
	}
	if facets.Numeric != false {
		t.Errorf("expected Numeric to be false, got %t", facets.Numeric)
	}
}

// TestOrderedConstants verifies that the Ordered enumeration constants have the correct underlying integer values.
func TestOrderedConstants(t *testing.T) {
	tests := []struct {
		name     string
		ordered  Ordered
		expected int
	}{
		{name: "OrderedFalse", ordered: OrderedFalse, expected: 0},
		{name: "OrderedPartial", ordered: OrderedPartial, expected: 1},
		{name: "OrderedTotal", ordered: OrderedTotal, expected: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.ordered) != tt.expected {
				t.Errorf("expected %s value %d, got %d", tt.name, tt.expected, tt.ordered)
			}
		})
	}
}

// TestCardinalityConstants verifies that the Cardinality enumeration constants have the correct underlying integer
// values.
func TestCardinalityConstants(t *testing.T) {
	tests := []struct {
		name        string
		cardinality Cardinality
		expected    int
	}{
		{name: "Finite", cardinality: Finite, expected: 0},
		{name: "CountablyInfinite", cardinality: CountablyInfinite, expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.cardinality) != tt.expected {
				t.Errorf("expected %s value %d, got %d", tt.name, tt.expected, tt.cardinality)
			}
		})
	}
}

// TestFundamentalFacetsFields verifies that FundamentalFacets struct fields can be assigned and retrieved properly.
func TestFundamentalFacetsFields(t *testing.T) {
	custom := FundamentalFacets{
		Ordered:     OrderedTotal,
		Bounded:     true,
		Cardinality: Finite,
		Numeric:     true,
	}

	if custom.Ordered != OrderedTotal {
		t.Errorf("expected OrderedTotal, got %v", custom.Ordered)
	}
	if !custom.Bounded {
		t.Errorf("expected Bounded=true, got %v", custom.Bounded)
	}
	if custom.Cardinality != Finite {
		t.Errorf("expected Cardinality=Finite, got %v", custom.Cardinality)
	}
	if !custom.Numeric {
		t.Errorf("expected Numeric=true, got %v", custom.Numeric)
	}
}
