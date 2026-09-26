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
	"strings"
	"testing"
)

// TestNewFacetEnumeration verifies that NewFacetEnumeration correctly initializes
// a facet with the expected name, fixed status, and provided values.
func TestNewFacetEnumeration(t *testing.T) {
	vals := []XSDValue{
		String("value1"),
		String("value2"),
	}

	facet := NewFacetEnumeration(vals)

	if facet.Name != NameEnumeration {
		t.Errorf("expected facet name %q, got %q", NameEnumeration, facet.Name)
	}

	if facet.Fixed {
		t.Errorf("expected facet Fixed to be false, got true")
	}

	if len(facet.Values) != len(vals) {
		t.Fatalf("expected %d values, got %d", len(vals), len(facet.Values))
	}

	for i, v := range facet.Values {
		if !v.IsIdenticalWith(vals[i]) {
			t.Errorf("expected value at index %d to be identical with %v, got %v", i, vals[i], v)
		}
	}
}

// TestFacetEnumerationCheckIdentityMatch tests that Check successfully validates
// a value using the fast-path identical match.
func TestFacetEnumerationCheckIdentityMatch(t *testing.T) {
	facet := NewFacetEnumeration([]XSDValue{
		String("alpha"),
		String("beta"),
		String("gamma"),
	})

	if err := facet.Check(String("beta")); err != nil {
		t.Fatalf("expected nil error on identical match, got: %v", err)
	}
}

// TestFacetEnumerationCheckValueComparisonMatch tests that Check successfully validates
// values of different underlying types that share an equivalent numeric/decimal value.
func TestFacetEnumerationCheckValueComparisonMatch(t *testing.T) {
	facet := NewFacetEnumeration([]XSDValue{
		Int(42),
	})

	if err := facet.Check(Short(42)); err != nil {
		t.Fatalf("expected nil error on value comparison match, got: %v", err)
	}
}

// TestFacetEnumerationCheckValueComparisonMismatch tests that Check returns an
// appropriate error when a comparable value is not present in the enumeration.
func TestFacetEnumerationCheckValueComparisonMismatch(t *testing.T) {
	facet := NewFacetEnumeration([]XSDValue{
		Short(10),
		Short(20),
	})

	err := facet.Check(Short(30))
	if err == nil {
		t.Fatal("expected error for value not present in enumeration, got nil")
	}

	expectedSub := "enumeration violation: value 30 is not in the set of allowed values"
	if !strings.Contains(err.Error(), expectedSub) {
		t.Errorf("expected error message to contain %q, got %q", expectedSub, err.Error())
	}
}

// TestFacetEnumerationCheckIncomparableType tests that Check returns an error
// when given a value whose type cannot be compared with the enumeration values.
func TestFacetEnumerationCheckIncomparableType(t *testing.T) {
	facet := NewFacetEnumeration([]XSDValue{
		Short(10),
	})

	err := facet.Check(String("10"))
	if err == nil {
		t.Fatal("expected error for incomparable type not in enumeration, got nil")
	}

	expectedSub := "enumeration violation: value 10 is not in the set of allowed values"
	if !strings.Contains(err.Error(), expectedSub) {
		t.Errorf("expected error message to contain %q, got %q", expectedSub, err.Error())
	}
}

// TestFacetEnumerationCheckEmptyValues tests that Check returns an error
// when validating against an enumeration that contains no allowed values.
func TestFacetEnumerationCheckEmptyValues(t *testing.T) {
	facet := NewFacetEnumeration(nil)

	err := facet.Check(String("test"))
	if err == nil {
		t.Fatal("expected error when facet has no values, got nil")
	}

	expectedSub := "enumeration violation: value test is not in the set of allowed values"
	if !strings.Contains(err.Error(), expectedSub) {
		t.Errorf("expected error message to contain %q, got %q", expectedSub, err.Error())
	}
}
