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

package datatypes

// Variety represents the {variety} property of a Simple Type Definition as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Representation:
// An integer enumeration representing the allowed varieties: absent, atomic, list, or union.
type Variety int

const (
	// VarietyAbsent indicates that the simple type definition has an absent variety.
	//
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.1.1)
	// If the simple type definition is 'anySimpleType', the {variety} property is absent.
	// For all other simple types, it must be 'atomic', 'list', or 'union'.
	VarietyAbsent Variety = iota

	// VarietyAtomic indicates that the simple type definition is atomic.
	// Atomic types represent values that are indivisible in the value space.
	VarietyAtomic

	// VarietyList indicates that the simple type definition is a list.
	// List types represent finite-length sequences of atomic values.
	VarietyList

	// VarietyUnion indicates that the simple type definition is a union.
	// Union types represent values that are valid against one of their member types.
	VarietyUnion
)

// Facet represents the constraining facet component of a simple type definition as defined in the governing
// specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.2)
//
// Representation:
// An interface implemented by any component restricting the value space or lexical space of a datatype.
type Facet interface {
	// Check applies the facet's constraint directly to the given XSDValue in the value space.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.2)
	//
	// Parameters:
	//   - val: The XSD value to be validated against the facet's constraint.
	//
	// Returns:
	//   - error: An error if the value does not satisfy the constraint defined by the facet, or nil if valid.
	Check(val XSDValue) error
}

// SimpleTypeDefinition represents the Simple Type Definition schema component as defined in the governing
// specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Representation:
// An interface capturing the properties and behaviors mandated for atomic, list, and union varieties.
type SimpleTypeDefinition interface {
	// Name returns the name of the simple type definition as an XML NCName.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.1)
	//
	// Returns:
	//   - string: The local name of the simple type definition, or an empty string if anonymous.
	Name() string

	// TargetNamespace returns the target namespace of the simple type definition.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.1)
	//
	// Returns:
	//   - string: The target namespace of the simple type definition, or an empty string if none.
	TargetNamespace() string

	// BaseType returns the simple type definition from which this type is derived.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.1)
	//
	// Returns:
	//   - SimpleTypeDefinition: The immediate base type definition of this simple type.
	BaseType() SimpleTypeDefinition

	// Variety returns the variety of the simple type definition.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.1)
	//
	// Returns:
	//   - Variety: The variety of the simple type definition.
	Variety() Variety

	// PrimitiveType returns the ancestor built-in primitive type definition of the simple type definition.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.1)
	//
	// Returns:
	//   - SimpleTypeDefinition: The primitive type definition, or nil if the variety is not atomic.
	PrimitiveType() SimpleTypeDefinition

	// ItemType returns the simple type definition of the items in the list.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.1)
	//
	// Returns:
	// - SimpleTypeDefinition: The simple type definition of the items allowed in the list, or nil if the variety is not
	// list.
	ItemType() SimpleTypeDefinition

	// MemberTypes returns the ordered sequence of member types for a union type.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.1)
	//
	// Returns:
	//   - []SimpleTypeDefinition: The ordered slice of member types, or nil if the variety is not union.
	MemberTypes() []SimpleTypeDefinition

	// Facets returns the set of constraining facets applied to this simple type.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.1)
	//
	// Returns:
	//   - []Facet: The slice of constraining facets applied to this type.
	Facets() []Facet

	// FundamentalFacets returns the fundamental characteristics of this simple type.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.2)
	//
	// Returns:
	//   - FundamentalFacets: The fundamental facets of this simple type.
	FundamentalFacets() FundamentalFacets

	// ValidateLexical validates a raw string literal against this simple type definition.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.4)
	//
	// Parameters:
	//   - literal: The raw, unmodified string literal to be validated and mapped.
	//
	// Returns:
	//   - XSDValue: The resulting parsed value if validation succeeds.
	//   - error: An error indicating validation failure if constraints are violated.
	ValidateLexical(literal string) (XSDValue, error)
}
