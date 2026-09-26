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
	"testing"
)

// TestAtomicTypeAccessors tests all basic property accessors of AtomicType.
func TestAtomicTypeAccessors(t *testing.T) {
	base := &AtomicType{name: "baseType"}
	prim := &AtomicType{name: "primitiveType"}
	facets := []Facet{NewFacetWhiteSpace(WhiteSpaceCollapse, false)}
	funds := FundamentalFacets{Ordered: OrderedTotal, Bounded: true, Cardinality: Finite, Numeric: true}

	atomic := &AtomicType{
		name:            "customInt",
		targetNamespace: "http://example.org/ns",
		baseType:        base,
		primitiveType:   prim,
		facets:          facets,
		fundamental:     funds,
	}

	if atomic.Name() != "customInt" {
		t.Fatalf("expected customInt, got %s", atomic.Name())
	}
	if atomic.TargetNamespace() != "http://example.org/ns" {
		t.Fatalf("expected http://example.org/ns, got %s", atomic.TargetNamespace())
	}
	if atomic.BaseType() != base {
		t.Fatalf("unexpected base type")
	}
	if atomic.Variety() != VarietyAtomic {
		t.Fatalf("expected VarietyAtomic, got %v", atomic.Variety())
	}
	if atomic.PrimitiveType() != prim {
		t.Fatalf("unexpected primitive type")
	}
	if atomic.ItemType() != nil {
		t.Fatalf("expected ItemType to be nil")
	}
	if atomic.MemberTypes() != nil {
		t.Fatalf("expected MemberTypes to be nil")
	}
	if len(atomic.Facets()) != 1 {
		t.Fatalf("expected 1 facet, got %d", len(atomic.Facets()))
	}
	if atomic.FundamentalFacets() != funds {
		t.Fatalf("unexpected fundamental facets")
	}
}

// TestAtomicTypeValidateLexicalWhitespaceModes verifies whitespace handling logic.
func TestAtomicTypeValidateLexicalWhitespaceModes(t *testing.T) {
	stringPrim := &AtomicType{
		name: primitiveString,
		lexicalMapper: func(literal string) (XSDValue, error) {
			return String(literal), nil
		},
	}

	derivedFromString := &AtomicType{
		name:          "customString",
		primitiveType: stringPrim,
	}

	val, err := derivedFromString.ValidateLexical("  leading and trailing  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.String() != "  leading and trailing  " {
		t.Fatalf("expected whitespace preserved, got %q", val.String())
	}

	builtInString := &AtomicType{
		name: primitiveString,
		lexicalMapper: func(literal string) (XSDValue, error) {
			return String(literal), nil
		},
	}

	val, err = builtInString.ValidateLexical("  preserve me  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.String() != "  preserve me  " {
		t.Fatalf("expected whitespace preserved, got %q", val.String())
	}

	nonStringPrim := &AtomicType{
		name: "decimal",
		lexicalMapper: func(literal string) (XSDValue, error) {
			return String(literal), nil
		},
	}

	derivedNonString := &AtomicType{
		name:          "customDecimal",
		primitiveType: nonStringPrim,
	}

	val, err = derivedNonString.ValidateLexical("  collapsed   value  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.String() != "collapsed value" {
		t.Fatalf("expected whitespace collapsed, got %q", val.String())
	}

	explicitWS := &AtomicType{
		name:          "customWithReplace",
		primitiveType: stringPrim,
		facets: []Facet{
			NewFacetWhiteSpace(WhiteSpaceReplace, false),
		},
	}

	val, err = explicitWS.ValidateLexical("\tvalue\nwith\rtabs\t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.String() != " value with tabs " {
		t.Fatalf("expected whitespace replaced, got %q", val.String())
	}
}

// TestAtomicTypeValidateLexicalLexicalFacets verifies lexical facet validation.
func TestAtomicTypeValidateLexicalLexicalFacets(t *testing.T) {
	matcher, err := CompileRegex("[0-9]+")
	if err != nil {
		t.Fatalf("unexpected regex compilation error: %v", err)
	}
	patternFacet := NewFacetPattern([]*PatternMatcher{matcher})

	atomicPass := &AtomicType{
		name: "testType",
		facets: []Facet{
			patternFacet,
		},
		lexicalMapper: func(literal string) (XSDValue, error) {
			return String(literal), nil
		},
	}

	val, err := atomicPass.ValidateLexical("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.String() != "123" {
		t.Fatalf("expected '123', got %q", val.String())
	}

	atomicFail := &AtomicType{
		name: "testType",
		facets: []Facet{
			patternFacet,
		},
		lexicalMapper: func(literal string) (XSDValue, error) {
			return String(literal), nil
		},
	}

	_, err = atomicFail.ValidateLexical("invalid")
	if err == nil {
		t.Fatalf("expected pattern error, got nil")
	}
}

// TestAtomicTypeValidateLexicalMappingBranches verifies lexical mapping branches.
func TestAtomicTypeValidateLexicalMappingBranches(t *testing.T) {
	errMapping := errors.New("mapping error")

	withMapper := &AtomicType{
		name: "mappedType",
		lexicalMapper: func(literal string) (XSDValue, error) {
			return String("mapped:" + literal), nil
		},
	}
	val, err := withMapper.ValidateLexical("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.String() != "mapped:test" {
		t.Fatalf("unexpected value: %s", val.String())
	}

	withMapperErr := &AtomicType{
		name: "mappedTypeErr",
		lexicalMapper: func(_ string) (XSDValue, error) {
			return nil, errMapping
		},
	}
	_, err = withMapperErr.ValidateLexical("test")
	if !errors.Is(err, errMapping) {
		t.Fatalf("expected errMapping, got %v", err)
	}

	primSuccess := &AtomicType{
		name: "primSuccess",
		lexicalMapper: func(literal string) (XSDValue, error) {
			return String("prim:" + literal), nil
		},
	}
	withPrim := &AtomicType{
		name:          "derivedType",
		primitiveType: primSuccess,
	}
	val, err = withPrim.ValidateLexical("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.String() != "prim:test" {
		t.Fatalf("unexpected value: %s", val.String())
	}

	primFail := &AtomicType{
		name: "primFail",
		lexicalMapper: func(_ string) (XSDValue, error) {
			return nil, errMapping
		},
	}
	withPrimFail := &AtomicType{
		name:          "derivedTypeFail",
		primitiveType: primFail,
	}
	_, err = withPrimFail.ValidateLexical("test")
	if !errors.Is(err, errMapping) {
		t.Fatalf("expected errMapping, got %v", err)
	}

	noMapping := &AtomicType{
		name: "unmappedType",
	}
	_, err = noMapping.ValidateLexical("test")
	if err == nil || err.Error() != "atomic type has no primitive type mapping" {
		t.Fatalf("expected 'atomic type has no primitive type mapping' error, got %v", err)
	}
}

// TestAtomicTypeValidateLexicalValueFacets verifies value-space facet checking.
func TestAtomicTypeValidateLexicalValueFacets(t *testing.T) {
	matcher, err := CompileRegex("[a-z]+")
	if err != nil {
		t.Fatalf("unexpected regex compilation error: %v", err)
	}
	patternFacet := NewFacetPattern([]*PatternMatcher{matcher})
	lengthFacetFail := NewFacetLength(3, false)
	lengthFacetPass := NewFacetLength(5, false)

	atomicFail := &AtomicType{
		name: "valueFacetFail",
		facets: []Facet{
			patternFacet,
			lengthFacetFail,
		},
		lexicalMapper: func(literal string) (XSDValue, error) {
			return String(literal), nil
		},
	}

	_, err = atomicFail.ValidateLexical("hello")
	if err == nil {
		t.Fatalf("expected length facet violation error, got nil")
	}

	atomicPass := &AtomicType{
		name: "valueFacetPass",
		facets: []Facet{
			patternFacet,
			lengthFacetPass,
		},
		lexicalMapper: func(literal string) (XSDValue, error) {
			return String(literal), nil
		},
	}

	val, err := atomicPass.ValidateLexical("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val.String() != "hello" {
		t.Fatalf("expected 'hello', got %q", val.String())
	}
}
