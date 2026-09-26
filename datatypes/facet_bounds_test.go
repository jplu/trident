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
	"strings"
	"testing"
)

// mockNonComparable implements XSDValue without implementing ComparableProvider or DecimalProvider.
type mockNonComparable struct {
	val string
}

// String returns the string representation of mockNonComparable.
func (m mockNonComparable) String() string {
	return m.val
}

// IsIdenticalWith checks if mockNonComparable is identical to another XSDValue.
func (m mockNonComparable) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(mockNonComparable); ok {
		return m.val == o.val
	}
	return false
}

// mockDecimalOnly implements DecimalProvider and XSDValue, but intentionally omits ComparableProvider.
type mockDecimalOnly struct {
	dec Decimal
}

// String returns the string representation of mockDecimalOnly.
func (m mockDecimalOnly) String() string {
	return m.dec.String()
}

// IsIdenticalWith checks if mockDecimalOnly is identical to another XSDValue.
func (m mockDecimalOnly) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(mockDecimalOnly); ok {
		return m.dec.IsIdenticalWith(o.dec)
	}
	return false
}

// ToDecimal returns the underlying Decimal value.
func (m mockDecimalOnly) ToDecimal() Decimal {
	return m.dec
}

// mockFailingComparable implements ComparableProvider and simulates a comparison failure.
type mockFailingComparable struct {
	val string
}

// String returns the string representation of mockFailingComparable.
func (m mockFailingComparable) String() string {
	return m.val
}

// IsIdenticalWith checks if mockFailingComparable is identical to another XSDValue.
func (m mockFailingComparable) IsIdenticalWith(_ XSDValue) bool {
	return false
}

// Compare simulates a comparison error.
func (m mockFailingComparable) Compare(_ XSDValue) (int, error) {
	return 0, errors.New("simulated comparison error")
}

// TestCompareValuesComparableProvider validates compareValues with ComparableProvider.
func TestCompareValuesComparableProvider(t *testing.T) {
	val := Int(10)
	bound := Int(20)

	res, err := compareValues(val, bound)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != -1 {
		t.Errorf("expected -1, got %d", res)
	}

	res, err = compareValues(bound, val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != 1 {
		t.Errorf("expected 1, got %d", res)
	}

	res, err = compareValues(val, val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != 0 {
		t.Errorf("expected 0, got %d", res)
	}
}

// TestCompareValuesComparableError validates compareValues when ComparableProvider returns an error.
func TestCompareValuesComparableError(t *testing.T) {
	val := mockFailingComparable{val: "fail"}
	bound := Int(10)

	_, err := compareValues(val, bound)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "simulated comparison error") {
		t.Errorf("expected 'simulated comparison error', got %v", err)
	}
}

// TestCompareValuesDecimalProviderFallback validates compareValues fallback to DecimalProvider.
func TestCompareValuesDecimalProviderFallback(t *testing.T) {
	d1 := NewDecimalFromInt64(15)
	d2 := NewDecimalFromInt64(30)
	val := mockDecimalOnly{dec: d1}
	bound := mockDecimalOnly{dec: d2}

	res, err := compareValues(val, bound)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != -1 {
		t.Errorf("expected -1, got %d", res)
	}

	res, err = compareValues(bound, val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != 1 {
		t.Errorf("expected 1, got %d", res)
	}

	res, err = compareValues(val, val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != 0 {
		t.Errorf("expected 0, got %d", res)
	}
}

// TestCompareValuesNonComparable validates compareValues when val is neither ComparableProvider nor DecimalProvider.
func TestCompareValuesNonComparable(t *testing.T) {
	val := mockNonComparable{val: "raw"}
	bound := Int(10)

	_, err := compareValues(val, bound)
	if err == nil {
		t.Fatal("expected error for non-comparable datatype, got nil")
	}
	if !strings.Contains(err.Error(), "does not support bound comparisons") {
		t.Errorf("expected unsupported bound comparisons error, got %v", err)
	}
}

// TestCompareValuesDecimalMismatch validates compareValues when val is DecimalProvider but bound is not.
func TestCompareValuesDecimalMismatch(t *testing.T) {
	val := mockDecimalOnly{dec: NewDecimalFromInt64(10)}
	bound := mockNonComparable{val: "non-decimal"}

	_, err := compareValues(val, bound)
	if err == nil {
		t.Fatal("expected error when bound is not DecimalProvider, got nil")
	}
	if !strings.Contains(err.Error(), "does not support bound comparisons") {
		t.Errorf("expected unsupported bound comparisons error, got %v", err)
	}
}

// TestFacetMinInclusive verifies creation, properties, and validation behavior of the minInclusive facet.
func TestFacetMinInclusive(t *testing.T) {
	bound := Int(10)
	facet := NewFacetMinInclusive(bound, true)

	if facet.Name != NameMinInclusive {
		t.Errorf("expected facet name %s, got %s", NameMinInclusive, facet.Name)
	}
	if !facet.Fixed {
		t.Errorf("expected Fixed to be true, got false")
	}
	if facet.Value != bound {
		t.Errorf("expected facet value %v, got %v", bound, facet.Value)
	}

	if err := facet.Check(Int(15)); err != nil {
		t.Errorf("unexpected error for value > bound: %v", err)
	}

	if err := facet.Check(Int(10)); err != nil {
		t.Errorf("unexpected error for value == bound: %v", err)
	}

	if err := facet.Check(Int(5)); err == nil {
		t.Error("expected minInclusive violation error for value < bound, got nil")
	} else if !strings.Contains(err.Error(), "minInclusive violation") {
		t.Errorf("expected minInclusive violation message, got: %v", err)
	}

	failingFacet := NewFacetMinInclusive(mockFailingComparable{val: "bound"}, false)
	if err := failingFacet.Check(mockFailingComparable{val: "val"}); err == nil {
		t.Error("expected error from failing comparison, got nil")
	}
}

// TestFacetMinExclusive verifies creation, properties, and validation behavior of the minExclusive facet.
func TestFacetMinExclusive(t *testing.T) {
	bound := Int(10)
	facet := NewFacetMinExclusive(bound, false)

	if facet.Name != NameMinExclusive {
		t.Errorf("expected facet name %s, got %s", NameMinExclusive, facet.Name)
	}
	if facet.Fixed {
		t.Errorf("expected Fixed to be false, got true")
	}
	if facet.Value != bound {
		t.Errorf("expected facet value %v, got %v", bound, facet.Value)
	}

	if err := facet.Check(Int(11)); err != nil {
		t.Errorf("unexpected error for value > bound: %v", err)
	}

	if err := facet.Check(Int(10)); err == nil {
		t.Error("expected minExclusive violation error for value == bound, got nil")
	} else if !strings.Contains(err.Error(), "minExclusive violation") {
		t.Errorf("expected minExclusive violation message, got: %v", err)
	}

	if err := facet.Check(Int(9)); err == nil {
		t.Error("expected minExclusive violation error for value < bound, got nil")
	} else if !strings.Contains(err.Error(), "minExclusive violation") {
		t.Errorf("expected minExclusive violation message, got: %v", err)
	}

	failingFacet := NewFacetMinExclusive(mockFailingComparable{val: "bound"}, false)
	if err := failingFacet.Check(mockFailingComparable{val: "val"}); err == nil {
		t.Error("expected error from failing comparison, got nil")
	}
}

// TestFacetMaxInclusive verifies creation, properties, and validation behavior of the maxInclusive facet.
func TestFacetMaxInclusive(t *testing.T) {
	bound := Int(20)
	facet := NewFacetMaxInclusive(bound, true)

	if facet.Name != NameMaxInclusive {
		t.Errorf("expected facet name %s, got %s", NameMaxInclusive, facet.Name)
	}
	if !facet.Fixed {
		t.Errorf("expected Fixed to be true, got false")
	}
	if facet.Value != bound {
		t.Errorf("expected facet value %v, got %v", bound, facet.Value)
	}

	if err := facet.Check(Int(15)); err != nil {
		t.Errorf("unexpected error for value < bound: %v", err)
	}

	if err := facet.Check(Int(20)); err != nil {
		t.Errorf("unexpected error for value == bound: %v", err)
	}

	if err := facet.Check(Int(25)); err == nil {
		t.Error("expected maxInclusive violation error for value > bound, got nil")
	} else if !strings.Contains(err.Error(), "maxInclusive violation") {
		t.Errorf("expected maxInclusive violation message, got: %v", err)
	}

	failingFacet := NewFacetMaxInclusive(mockFailingComparable{val: "bound"}, false)
	if err := failingFacet.Check(mockFailingComparable{val: "val"}); err == nil {
		t.Error("expected error from failing comparison, got nil")
	}
}

// TestFacetMaxExclusive verifies creation, properties, and validation behavior of the maxExclusive facet.
func TestFacetMaxExclusive(t *testing.T) {
	bound := Int(20)
	facet := NewFacetMaxExclusive(bound, false)

	if facet.Name != NameMaxExclusive {
		t.Errorf("expected facet name %s, got %s", NameMaxExclusive, facet.Name)
	}
	if facet.Fixed {
		t.Errorf("expected Fixed to be false, got true")
	}
	if facet.Value != bound {
		t.Errorf("expected facet value %v, got %v", bound, facet.Value)
	}

	if err := facet.Check(Int(19)); err != nil {
		t.Errorf("unexpected error for value < bound: %v", err)
	}

	if err := facet.Check(Int(20)); err == nil {
		t.Error("expected maxExclusive violation error for value == bound, got nil")
	} else if !strings.Contains(err.Error(), "maxExclusive violation") {
		t.Errorf("expected maxExclusive violation message, got: %v", err)
	}

	if err := facet.Check(Int(21)); err == nil {
		t.Error("expected maxExclusive violation error for value > bound, got nil")
	} else if !strings.Contains(err.Error(), "maxExclusive violation") {
		t.Errorf("expected maxExclusive violation message, got: %v", err)
	}

	failingFacet := NewFacetMaxExclusive(mockFailingComparable{val: "bound"}, false)
	if err := failingFacet.Check(mockFailingComparable{val: "val"}); err == nil {
		t.Error("expected error from failing comparison, got nil")
	}
}
