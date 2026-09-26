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

// unionMockXSDValue represents a test double implementing XSDValue for union testing.
type unionMockXSDValue struct {
	val string
}

// String returns the canonical string representation of the mock value.
func (m unionMockXSDValue) String() string {
	return m.val
}

// IsIdenticalWith reports whether two mock values are identical.
func (m unionMockXSDValue) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(unionMockXSDValue); ok {
		return m.val == o.val
	}
	return false
}

// unionMockSimpleType represents a test double implementing SimpleTypeDefinition for union testing.
type unionMockSimpleType struct {
	name            string
	targetNamespace string
	baseType        SimpleTypeDefinition
	variety         Variety
	primitiveType   SimpleTypeDefinition
	itemType        SimpleTypeDefinition
	memberTypes     []SimpleTypeDefinition
	facets          []Facet
	fundamental     FundamentalFacets
	validateFn      func(string) (XSDValue, error)
}

// Name returns the type name.
func (m *unionMockSimpleType) Name() string {
	return m.name
}

// TargetNamespace returns the target namespace URI.
func (m *unionMockSimpleType) TargetNamespace() string {
	return m.targetNamespace
}

// BaseType returns the base type definition.
func (m *unionMockSimpleType) BaseType() SimpleTypeDefinition {
	return m.baseType
}

// Variety returns the variety of the simple type.
func (m *unionMockSimpleType) Variety() Variety {
	return m.variety
}

// PrimitiveType returns the ancestor primitive type definition.
func (m *unionMockSimpleType) PrimitiveType() SimpleTypeDefinition {
	return m.primitiveType
}

// ItemType returns the item type definition.
func (m *unionMockSimpleType) ItemType() SimpleTypeDefinition {
	return m.itemType
}

// MemberTypes returns the union member types.
func (m *unionMockSimpleType) MemberTypes() []SimpleTypeDefinition {
	return m.memberTypes
}

// Facets returns the constraining facets.
func (m *unionMockSimpleType) Facets() []Facet {
	return m.facets
}

// FundamentalFacets returns the fundamental facets.
func (m *unionMockSimpleType) FundamentalFacets() FundamentalFacets {
	return m.fundamental
}

// ValidateLexical validates and parses a lexical string.
func (m *unionMockSimpleType) ValidateLexical(literal string) (XSDValue, error) {
	if m.validateFn != nil {
		return m.validateFn(literal)
	}
	return unionMockXSDValue{val: literal}, nil
}

// unionMockValueFacet represents a test double implementing a value-space Facet for union testing.
type unionMockValueFacet struct {
	checkFn func(XSDValue) error
}

// Check evaluates the value constraint.
func (m *unionMockValueFacet) Check(val XSDValue) error {
	if m.checkFn != nil {
		return m.checkFn(val)
	}
	return nil
}

// unionMockLexicalFacet represents a test double implementing both Facet and LexicalFacet for union testing.
type unionMockLexicalFacet struct {
	checkLexicalFn func(string) error
	checkFn        func(XSDValue) error
}

// Check evaluates the value-space constraint.
func (m *unionMockLexicalFacet) Check(val XSDValue) error {
	if m.checkFn != nil {
		return m.checkFn(val)
	}
	return nil
}

// CheckLexical evaluates the lexical-space constraint.
func (m *unionMockLexicalFacet) CheckLexical(literal string) error {
	if m.checkLexicalFn != nil {
		return m.checkLexicalFn(literal)
	}
	return nil
}

// TestUnionTypeAccessors tests all basic getter methods of UnionType.
func TestUnionTypeAccessors(t *testing.T) {
	base := &unionMockSimpleType{name: "baseType"}
	members := []SimpleTypeDefinition{&unionMockSimpleType{name: "member1"}}
	facets := []Facet{&unionMockValueFacet{}}
	fundamental := FundamentalFacets{
		Ordered:     OrderedTotal,
		Bounded:     true,
		Cardinality: Finite,
		Numeric:     true,
	}

	u := &UnionType{
		name:            "testUnion",
		targetNamespace: "http://example.com/ns",
		baseType:        base,
		memberTypes:     members,
		facets:          facets,
		fundamental:     fundamental,
	}

	if u.Name() != "testUnion" {
		t.Fatalf("expected testUnion, got %s", u.Name())
	}
	if u.TargetNamespace() != "http://example.com/ns" {
		t.Fatalf("expected http://example.com/ns, got %s", u.TargetNamespace())
	}
	if u.BaseType() != base {
		t.Fatalf("expected baseType, got %v", u.BaseType())
	}
	if u.Variety() != VarietyUnion {
		t.Fatalf("expected VarietyUnion, got %v", u.Variety())
	}
	if u.PrimitiveType() != nil {
		t.Fatalf("expected nil PrimitiveType, got %v", u.PrimitiveType())
	}
	if u.ItemType() != nil {
		t.Fatalf("expected nil ItemType, got %v", u.ItemType())
	}
	if len(u.MemberTypes()) != 1 || u.MemberTypes()[0] != members[0] {
		t.Fatalf("unexpected MemberTypes: %v", u.MemberTypes())
	}
	if len(u.Facets()) != 1 || u.Facets()[0] != facets[0] {
		t.Fatalf("unexpected Facets: %v", u.Facets())
	}
	if u.FundamentalFacets() != fundamental {
		t.Fatalf("unexpected FundamentalFacets: %v", u.FundamentalFacets())
	}
}

// TestUnionTypeFindActiveMember tests the resolution of active member types.
func TestUnionTypeFindActiveMember(t *testing.T) {
	emptyUnion := &UnionType{}
	_, _, errEmpty := emptyUnion.findActiveMember("test")
	if errEmpty == nil || errEmpty.Error() != "value does not match any union member type" {
		t.Fatalf("expected no member types error, got %v", errEmpty)
	}

	m1 := &unionMockSimpleType{
		name: "m1",
		validateFn: func(_ string) (XSDValue, error) {
			return nil, errors.New("m1 failed")
		},
	}
	m2 := &unionMockSimpleType{
		name: "m2",
		validateFn: func(s string) (XSDValue, error) {
			if s == "valid2" {
				return unionMockXSDValue{val: "m2Val"}, nil
			}
			return nil, errors.New("m2 failed")
		},
	}

	u := &UnionType{
		memberTypes: []SimpleTypeDefinition{m1, m2},
	}

	val, activeMember, err := u.findActiveMember("valid2")
	if err != nil {
		t.Fatalf("unexpected error finding active member: %v", err)
	}
	if val.String() != "m2Val" {
		t.Fatalf("expected m2Val, got %s", val.String())
	}
	if activeMember != m2 {
		t.Fatalf("expected m2 as active member, got %v", activeMember)
	}

	_, _, errNotFound := u.findActiveMember("invalid")
	if errNotFound == nil ||
		errNotFound.Error() != "value does not match any union member type. Last error: m2 failed" {
		t.Fatalf("unexpected error message on failure: %v", errNotFound)
	}
}

// TestUnionTypeValidateLexicalEmptyMembers verifies validation failure when no member types exist.
func TestUnionTypeValidateLexicalEmptyMembers(t *testing.T) {
	u := &UnionType{}
	_, err := u.ValidateLexical("test")
	if err == nil || err.Error() != "union type has no member types defined" {
		t.Fatalf("expected empty member types error, got %v", err)
	}
}

// TestUnionTypeValidateLexicalMemberFailure verifies validation failure when no member validates the literal.
func TestUnionTypeValidateLexicalMemberFailure(t *testing.T) {
	m := &unionMockSimpleType{
		validateFn: func(string) (XSDValue, error) {
			return nil, errors.New("parse error")
		},
	}
	u := &UnionType{
		memberTypes: []SimpleTypeDefinition{m},
	}

	_, err := u.ValidateLexical("literal")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestUnionTypeValidateLexicalWhitespaceHandling tests whitespace normalization resolution based on active members.
func TestUnionTypeValidateLexicalWhitespaceHandling(t *testing.T) {
	stringPrim := &unionMockSimpleType{name: primitiveString}
	mAtomicString := &unionMockSimpleType{
		variety:       VarietyAtomic,
		primitiveType: stringPrim,
		validateFn: func(s string) (XSDValue, error) {
			return unionMockXSDValue{val: s}, nil
		},
	}

	var observedLexical string
	lexFacet := &unionMockLexicalFacet{
		checkLexicalFn: func(s string) error {
			observedLexical = s
			return nil
		},
	}

	uString := &UnionType{
		memberTypes: []SimpleTypeDefinition{mAtomicString},
		facets:      []Facet{lexFacet},
	}

	_, err := uString.ValidateLexical("  hello   world  ")
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if observedLexical != "  hello   world  " {
		t.Fatalf("expected preserved whitespace, got %q", observedLexical)
	}

	mAtomicWithWhitespaceFacet := &unionMockSimpleType{
		variety:       VarietyAtomic,
		primitiveType: stringPrim,
		facets: []Facet{
			&unionMockValueFacet{},
			NewFacetWhiteSpace(WhiteSpaceReplace, false),
		},
		validateFn: func(s string) (XSDValue, error) {
			return unionMockXSDValue{val: s}, nil
		},
	}

	uReplace := &UnionType{
		memberTypes: []SimpleTypeDefinition{mAtomicWithWhitespaceFacet},
		facets:      []Facet{lexFacet},
	}

	_, err = uReplace.ValidateLexical("hello\tworld")
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if observedLexical != "hello world" {
		t.Fatalf("expected replaced tab whitespace, got %q", observedLexical)
	}

	intPrim := &unionMockSimpleType{name: "integer"}
	mAtomicNonString := &unionMockSimpleType{
		variety:       VarietyAtomic,
		primitiveType: intPrim,
		validateFn: func(s string) (XSDValue, error) {
			return unionMockXSDValue{val: s}, nil
		},
	}

	uCollapseInt := &UnionType{
		memberTypes: []SimpleTypeDefinition{mAtomicNonString},
		facets:      []Facet{lexFacet},
	}

	_, err = uCollapseInt.ValidateLexical("  hello   world  ")
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if observedLexical != "hello world" {
		t.Fatalf("expected collapsed whitespace, got %q", observedLexical)
	}

	mAtomicNilPrim := &unionMockSimpleType{
		variety:       VarietyAtomic,
		primitiveType: nil,
		validateFn: func(s string) (XSDValue, error) {
			return unionMockXSDValue{val: s}, nil
		},
	}

	uCollapseNilPrim := &UnionType{
		memberTypes: []SimpleTypeDefinition{mAtomicNilPrim},
		facets:      []Facet{lexFacet},
	}

	_, err = uCollapseNilPrim.ValidateLexical("  hello   world  ")
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if observedLexical != "hello world" {
		t.Fatalf("expected collapsed whitespace, got %q", observedLexical)
	}

	mList := &unionMockSimpleType{
		variety: VarietyList,
		validateFn: func(s string) (XSDValue, error) {
			return unionMockXSDValue{val: s}, nil
		},
	}

	uList := &UnionType{
		memberTypes: []SimpleTypeDefinition{mList},
		facets:      []Facet{lexFacet},
	}

	_, err = uList.ValidateLexical("  hello   world  ")
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if observedLexical != "hello world" {
		t.Fatalf("expected collapsed whitespace for list member, got %q", observedLexical)
	}
}

// TestUnionTypeValidateLexicalFacetsEvaluation verifies evaluation of both lexical and value facets.
func TestUnionTypeValidateLexicalFacetsEvaluation(t *testing.T) {
	m := &unionMockSimpleType{
		variety: VarietyList,
		validateFn: func(s string) (XSDValue, error) {
			return unionMockXSDValue{val: s}, nil
		},
	}

	failingLexFacet := &unionMockLexicalFacet{
		checkLexicalFn: func(string) error {
			return errors.New("lexical facet violation")
		},
	}

	uFailingLex := &UnionType{
		memberTypes: []SimpleTypeDefinition{m},
		facets:      []Facet{failingLexFacet},
	}

	_, err := uFailingLex.ValidateLexical("test")
	if err == nil || err.Error() != "lexical facet violation" {
		t.Fatalf("expected lexical facet violation, got %v", err)
	}

	passingLexFacet := &unionMockLexicalFacet{
		checkLexicalFn: func(string) error {
			return nil
		},
	}
	failingValueFacet := &unionMockValueFacet{
		checkFn: func(XSDValue) error {
			return errors.New("value facet violation")
		},
	}

	uFailingValue := &UnionType{
		memberTypes: []SimpleTypeDefinition{m},
		facets:      []Facet{passingLexFacet, failingValueFacet},
	}

	_, err = uFailingValue.ValidateLexical("test")
	if err == nil || err.Error() != "value facet violation" {
		t.Fatalf("expected value facet violation, got %v", err)
	}

	passingValueFacet := &unionMockValueFacet{
		checkFn: func(XSDValue) error {
			return nil
		},
	}

	uSuccess := &UnionType{
		memberTypes: []SimpleTypeDefinition{m},
		facets:      []Facet{passingLexFacet, passingValueFacet},
	}

	val, err := uSuccess.ValidateLexical("hello")
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if val.String() != "hello" {
		t.Fatalf("expected hello, got %s", val.String())
	}
}
