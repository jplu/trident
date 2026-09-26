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

// listTestSimpleType is a test double implementing SimpleTypeDefinition for list testing.
type listTestSimpleType struct {
	name                string
	targetNamespace     string
	baseType            SimpleTypeDefinition
	itemType            SimpleTypeDefinition
	memberTypes         []SimpleTypeDefinition
	facets              []Facet
	fundamental         FundamentalFacets
	variety             Variety
	primitiveType       SimpleTypeDefinition
	validateLexicalFunc func(literal string) (XSDValue, error)
}

// Name returns the mock simple type name.
func (m *listTestSimpleType) Name() string {
	return m.name
}

// TargetNamespace returns the mock target namespace.
func (m *listTestSimpleType) TargetNamespace() string {
	return m.targetNamespace
}

// BaseType returns the mock base type.
func (m *listTestSimpleType) BaseType() SimpleTypeDefinition {
	return m.baseType
}

// Variety returns the mock variety.
func (m *listTestSimpleType) Variety() Variety {
	return m.variety
}

// PrimitiveType returns the mock primitive type ancestor.
func (m *listTestSimpleType) PrimitiveType() SimpleTypeDefinition {
	return m.primitiveType
}

// ItemType returns the mock item type.
func (m *listTestSimpleType) ItemType() SimpleTypeDefinition {
	return m.itemType
}

// MemberTypes returns the mock member types.
func (m *listTestSimpleType) MemberTypes() []SimpleTypeDefinition {
	return m.memberTypes
}

// Facets returns the mock facets slice.
func (m *listTestSimpleType) Facets() []Facet {
	return m.facets
}

// FundamentalFacets returns the mock fundamental facets.
func (m *listTestSimpleType) FundamentalFacets() FundamentalFacets {
	return m.fundamental
}

// ValidateLexical executes the configured mock validation function or defaults to String mapping.
func (m *listTestSimpleType) ValidateLexical(literal string) (XSDValue, error) {
	if m.validateLexicalFunc != nil {
		return m.validateLexicalFunc(literal)
	}
	return String(literal), nil
}

// listTestLexicalFacet is a test double implementing both Facet and LexicalFacet.
type listTestLexicalFacet struct {
	checkFunc        func(val XSDValue) error
	checkLexicalFunc func(literal string) error
}

// Check delegates to the configured check function.
func (m *listTestLexicalFacet) Check(val XSDValue) error {
	if m.checkFunc != nil {
		return m.checkFunc(val)
	}
	return nil
}

// CheckLexical delegates to the configured lexical check function.
func (m *listTestLexicalFacet) CheckLexical(literal string) error {
	if m.checkLexicalFunc != nil {
		return m.checkLexicalFunc(literal)
	}
	return nil
}

// listTestValueFacet is a test double implementing only the Facet interface.
type listTestValueFacet struct {
	checkFunc func(val XSDValue) error
}

// Check delegates to the configured value check function.
func (m *listTestValueFacet) Check(val XSDValue) error {
	if m.checkFunc != nil {
		return m.checkFunc(val)
	}
	return nil
}

// TestListValueString tests serialization of empty and populated ListValue instances.
func TestListValueString(t *testing.T) {
	var empty ListValue
	if got := empty.String(); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}

	populated := ListValue{String("alpha"), String("beta"), String("gamma")}
	if got := populated.String(); got != "alpha beta gamma" {
		t.Fatalf("expected 'alpha beta gamma', got %q", got)
	}
}

// TestListValueIsIdenticalWith tests identity comparison between ListValue and other values.
func TestListValueIsIdenticalWith(t *testing.T) {
	list1 := ListValue{String("a"), String("b")}
	list2 := ListValue{String("a"), String("b")}
	listDiffLen := ListValue{String("a")}
	listDiffVal := ListValue{String("a"), String("c")}

	if list1.IsIdenticalWith(String("a b")) {
		t.Fatal("expected IsIdenticalWith to return false for non-ListValue")
	}

	if list1.IsIdenticalWith(listDiffLen) {
		t.Fatal("expected IsIdenticalWith to return false for differing length")
	}

	if list1.IsIdenticalWith(listDiffVal) {
		t.Fatal("expected IsIdenticalWith to return false for differing items")
	}

	if !list1.IsIdenticalWith(list2) {
		t.Fatal("expected IsIdenticalWith to return true for matching lists")
	}
}

// TestListValueLength tests measuring the length of ListValue.
func TestListValueLength(t *testing.T) {
	var empty ListValue
	if empty.Length() != 0 {
		t.Fatalf("expected length 0, got %d", empty.Length())
	}

	populated := ListValue{String("a"), String("b")}
	if populated.Length() != 2 {
		t.Fatalf("expected length 2, got %d", populated.Length())
	}
}

// TestListTypeMetadataMethods tests accessors and metadata methods of ListType.
func TestListTypeMetadataMethods(t *testing.T) {
	base := &listTestSimpleType{name: "baseType"}
	item := &listTestSimpleType{name: "itemType"}
	facets := []Facet{&listTestValueFacet{}}
	fund := FundamentalFacets{Ordered: OrderedTotal, Bounded: true}

	listType := &ListType{
		name:            "customList",
		targetNamespace: "http://example.com/ns",
		baseType:        base,
		itemType:        item,
		facets:          facets,
		fundamental:     fund,
	}

	if listType.Name() != "customList" {
		t.Fatalf("expected 'customList', got %q", listType.Name())
	}
	if listType.TargetNamespace() != "http://example.com/ns" {
		t.Fatalf("expected 'http://example.com/ns', got %q", listType.TargetNamespace())
	}
	if listType.BaseType() != base {
		t.Fatalf("expected base type pointer %v, got %v", base, listType.BaseType())
	}
	if listType.Variety() != VarietyList {
		t.Fatalf("expected VarietyList, got %v", listType.Variety())
	}
	if listType.PrimitiveType() != nil {
		t.Fatalf("expected nil primitive type, got %v", listType.PrimitiveType())
	}
	if listType.ItemType() != item {
		t.Fatalf("expected item type %v, got %v", item, listType.ItemType())
	}
	if listType.MemberTypes() != nil {
		t.Fatalf("expected nil member types, got %v", listType.MemberTypes())
	}
	if len(listType.Facets()) != 1 {
		t.Fatalf("expected 1 facet, got %d", len(listType.Facets()))
	}
	if listType.FundamentalFacets() != fund {
		t.Fatalf("expected fundamental facets %+v, got %+v", fund, listType.FundamentalFacets())
	}
}

// TestListTypeValidateLexicalNoItemType tests ValidateLexical failure when no item type is defined.
func TestListTypeValidateLexicalNoItemType(t *testing.T) {
	listType := &ListType{itemType: nil}
	_, err := listType.ValidateLexical("item1 item2")
	if err == nil || err.Error() != "list type has no item type defined" {
		t.Fatalf("expected 'list type has no item type defined', got %v", err)
	}
}

// TestListTypeValidateLexicalLexicalFacetFailure tests lexical facet check failure.
func TestListTypeValidateLexicalLexicalFacetFailure(t *testing.T) {
	expectedErr := errors.New("pattern error")
	lexFacet := &listTestLexicalFacet{
		checkLexicalFunc: func(_ string) error {
			return expectedErr
		},
	}

	listType := &ListType{
		itemType: &listTestSimpleType{},
		facets:   []Facet{lexFacet},
	}

	_, err := listType.ValidateLexical("test literal")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

// TestListTypeValidateLexicalEmptyLiteral tests ValidateLexical on empty or whitespace-only literal.
func TestListTypeValidateLexicalEmptyLiteral(t *testing.T) {
	var checkedValue XSDValue
	valFacet := &listTestValueFacet{
		checkFunc: func(val XSDValue) error {
			checkedValue = val
			return nil
		},
	}

	listType := &ListType{
		itemType: &listTestSimpleType{},
		facets:   []Facet{valFacet},
	}

	val, err := listType.ValidateLexical("   \t\n  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	listVal, ok := val.(ListValue)
	if !ok {
		t.Fatalf("expected ListValue, got %T", val)
	}
	if len(listVal) != 0 {
		t.Fatalf("expected 0 items, got %d", len(listVal))
	}
	if checkedValue == nil {
		t.Fatal("expected value facet check to be called")
	}
}

// TestListTypeValidateLexicalItemValidationFailure tests item parsing failure during list tokenization.
func TestListTypeValidateLexicalItemValidationFailure(t *testing.T) {
	expectedErr := errors.New("invalid item")
	mockItem := &listTestSimpleType{
		validateLexicalFunc: func(literal string) (XSDValue, error) {
			if literal == "bad" {
				return nil, expectedErr
			}
			return String(literal), nil
		},
	}

	listType := &ListType{
		itemType: mockItem,
	}

	_, err := listType.ValidateLexical("good bad")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

// TestListTypeValidateLexicalValueFacetFailure tests value facet failure after item tokenization.
func TestListTypeValidateLexicalValueFacetFailure(t *testing.T) {
	expectedErr := errors.New("length facet error")
	valFacet := &listTestValueFacet{
		checkFunc: func(_ XSDValue) error {
			return expectedErr
		},
	}

	listType := &ListType{
		itemType: &listTestSimpleType{},
		facets:   []Facet{valFacet},
	}

	_, err := listType.ValidateLexical("one two")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

// TestListTypeValidateLexicalSuccess tests successful validation with multiple items and facets.
func TestListTypeValidateLexicalSuccess(t *testing.T) {
	lexCheckCalled := false
	valCheckCalled := false

	lexFacet := &listTestLexicalFacet{
		checkLexicalFunc: func(literal string) error {
			lexCheckCalled = true
			if literal != "alpha beta" {
				t.Fatalf("expected collapsed string 'alpha beta', got %q", literal)
			}
			return nil
		},
	}

	valFacet := &listTestValueFacet{
		checkFunc: func(val XSDValue) error {
			valCheckCalled = true
			lv, ok := val.(ListValue)
			if !ok || len(lv) != 2 {
				t.Fatalf("expected ListValue with length 2, got %v", val)
			}
			return nil
		},
	}

	listType := &ListType{
		itemType: &listTestSimpleType{},
		facets:   []Facet{lexFacet, valFacet},
	}

	val, err := listType.ValidateLexical("  alpha   \n\t beta  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !lexCheckCalled {
		t.Fatal("expected lexical facet check to be called")
	}
	if !valCheckCalled {
		t.Fatal("expected value facet check to be called")
	}

	listVal, ok := val.(ListValue)
	if !ok {
		t.Fatalf("expected ListValue, got %T", val)
	}
	if len(listVal) != 2 {
		t.Fatalf("expected 2 items, got %d", len(listVal))
	}
	if listVal[0].String() != "alpha" || listVal[1].String() != "beta" {
		t.Fatalf("unexpected items: %v", listVal)
	}
}
