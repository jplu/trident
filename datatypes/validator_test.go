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

// mockSimpleTypeForValidator provides a mock implementation of SimpleTypeDefinition for testing Validate.
type mockSimpleTypeForValidator struct {
	validateLexicalFunc func(literal string) (XSDValue, error)
}

// Name returns an empty string for the mock type definition.
func (m mockSimpleTypeForValidator) Name() string {
	return ""
}

// TargetNamespace returns an empty string for the mock type definition.
func (m mockSimpleTypeForValidator) TargetNamespace() string {
	return ""
}

// BaseType returns nil for the mock type definition.
func (m mockSimpleTypeForValidator) BaseType() SimpleTypeDefinition {
	return nil
}

// Variety returns VarietyAbsent for the mock type definition.
func (m mockSimpleTypeForValidator) Variety() Variety {
	return VarietyAbsent
}

// PrimitiveType returns nil for the mock type definition.
func (m mockSimpleTypeForValidator) PrimitiveType() SimpleTypeDefinition {
	return nil
}

// ItemType returns nil for the mock type definition.
func (m mockSimpleTypeForValidator) ItemType() SimpleTypeDefinition {
	return nil
}

// MemberTypes returns nil for the mock type definition.
func (m mockSimpleTypeForValidator) MemberTypes() []SimpleTypeDefinition {
	return nil
}

// Facets returns nil for the mock type definition.
func (m mockSimpleTypeForValidator) Facets() []Facet {
	return nil
}

// FundamentalFacets returns empty FundamentalFacets for the mock type definition.
func (m mockSimpleTypeForValidator) FundamentalFacets() FundamentalFacets {
	return FundamentalFacets{}
}

// ValidateLexical delegates lexical validation to the configured mock function.
func (m mockSimpleTypeForValidator) ValidateLexical(literal string) (XSDValue, error) {
	return m.validateLexicalFunc(literal)
}

// mockValueForValidator implements XSDValue for validator testing.
type mockValueForValidator struct {
	val string
}

// String returns the underlying string representation.
func (m mockValueForValidator) String() string {
	return m.val
}

// IsIdenticalWith checks identity with another XSDValue.
func (m mockValueForValidator) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(mockValueForValidator); ok {
		return m.val == o.val
	}
	return false
}

// TestValidateSuccess verifies that Validate successfully returns the parsed XSDValue on valid input.
func TestValidateSuccess(t *testing.T) {
	expected := mockValueForValidator{val: "valid"}
	mockDef := mockSimpleTypeForValidator{
		validateLexicalFunc: func(literal string) (XSDValue, error) {
			if literal == "valid" {
				return expected, nil
			}
			return nil, errors.New("unexpected input")
		},
	}

	val, err := Validate("valid", mockDef)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if val != expected {
		t.Fatalf("expected %v, got %v", expected, val)
	}
}

// TestValidateError verifies that Validate propagates the error returned by the type definition.
func TestValidateError(t *testing.T) {
	expectedErr := errors.New("validation failure")
	mockDef := mockSimpleTypeForValidator{
		validateLexicalFunc: func(_ string) (XSDValue, error) {
			return nil, expectedErr
		},
	}

	val, err := Validate("invalid", mockDef)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got: %v", expectedErr, err)
	}
	if val != nil {
		t.Fatalf("expected nil value, got: %v", val)
	}
}
