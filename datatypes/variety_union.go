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

import "errors"

// UnionType represents the Simple Type Definition with a variety of union as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Representation:
// A union simple type whose value space, lexical space, and lexical mappings are the union of the value spaces, lexical
// spaces, and lexical mappings of its member types.
type UnionType struct {
	// name is the local name of the union type definition.
	name string
	// targetNamespace is the target namespace of the union type definition.
	targetNamespace string
	// baseType is the simple type definition from which this union type is derived.
	baseType SimpleTypeDefinition
	// memberTypes is the ordered list of member type definitions.
	memberTypes []SimpleTypeDefinition
	// facets is the set of constraining facets specified on this union type.
	facets []Facet
	// fundamental contains the fundamental facets of this union type.
	fundamental FundamentalFacets
}

// Name returns the local name of the UnionType definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - string: The local name of the union type definition.
func (u *UnionType) Name() string {
	return u.name
}

// TargetNamespace returns the target namespace of the UnionType definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - string: The target namespace URI of the union type definition.
func (u *UnionType) TargetNamespace() string {
	return u.targetNamespace
}

// BaseType returns the simple type definition from which this UnionType is derived.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - SimpleTypeDefinition: The parent simple type definition component.
func (u *UnionType) BaseType() SimpleTypeDefinition {
	return u.baseType
}

// Variety returns the variety of the simple type definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - Variety: The variety of the simple type, which is always VarietyUnion.
func (u *UnionType) Variety() Variety {
	return VarietyUnion
}

// PrimitiveType returns nil as union types do not have a direct primitive type definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - SimpleTypeDefinition: Always nil for union type definitions.
func (u *UnionType) PrimitiveType() SimpleTypeDefinition {
	return nil
}

// ItemType returns nil as union types do not have an item type definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - SimpleTypeDefinition: Always nil for union type definitions.
func (u *UnionType) ItemType() SimpleTypeDefinition {
	return nil
}

// MemberTypes returns the ordered sequence of member types for this UnionType.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - []SimpleTypeDefinition: The ordered list of simple type definitions representing the members of the union.
func (u *UnionType) MemberTypes() []SimpleTypeDefinition {
	return u.memberTypes
}

// Facets returns the set of constraining facets applied directly to this UnionType.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - []Facet: The slice of constraining facets applied to the union type definition.
func (u *UnionType) Facets() []Facet {
	return u.facets
}

// FundamentalFacets returns the fundamental characteristics associated with this UnionType.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.2)
//
// Parameters:
//
// Returns:
//   - FundamentalFacets: The fundamental facets derived for this union type definition.
func (u *UnionType) FundamentalFacets() FundamentalFacets {
	return u.fundamental
}

// ValidateLexical validates a raw XML string literal against the union type definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.4.1.3)
//
// Parameters:
//   - literal: The raw XML string literal to validate and parse.
//
// Returns:
//   - XSDValue: The successfully parsed value from the first matching member type.
//   - error: An error if the literal does not conform to any member type or if a union-level facet check fails.
func (u *UnionType) ValidateLexical(literal string) (XSDValue, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.1.1)
	// The union variety requires the {member type definitions} property to contain at least one simple type definition.
	if len(u.memberTypes) == 0 {
		return nil, errors.New("union type has no member types defined")
	}

	// Step 1: Resolve Active Member
	// Find the first member type that successfully validates and parses the raw input literal.
	activeValue, activeMember, err := u.findActiveMember(literal)
	if err != nil {
		return nil, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 2.4.1.3)
	// The whitespace normalization behavior of a union type is determined dynamically by its active member type. If the
	// active member type is atomic, its specific whiteSpace facet determines the normalization; otherwise, the default
	// 'collapse' behavior is assumed.
	wsVal := WhiteSpaceCollapse
	if activeMember.Variety() == VarietyAtomic {
		if activeMember.PrimitiveType() != nil && activeMember.PrimitiveType().Name() == primitiveString {
			wsVal = WhiteSpacePreserve
		}
		for _, f := range activeMember.Facets() {
			if wsf, ok := f.(FacetWhiteSpace); ok {
				wsVal = wsf.Value
				break
			}
		}
	}
	normalizedLiteral := NewFacetWhiteSpace(wsVal, false).Normalize(literal)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.1.4)
	// Constraining facets defined directly on the union type are evaluated against the active value or the normalized
	// literal corresponding to the active member's whitespace rules.
	for _, f := range u.facets {
		if lf, ok := f.(LexicalFacet); ok {
			if cErr := lf.CheckLexical(normalizedLiteral); cErr != nil {
				return nil, cErr
			}
		} else {
			if cErr := f.Check(activeValue); cErr != nil {
				return nil, cErr
			}
		}
	}

	return activeValue, nil
}

// findActiveMember iterates through the ordered member types to find the first type that successfully validates and
// parses the input literal.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.4.1.3)
//
// Parameters:
//   - literal: The raw lexical string literal to check against union member types.
//
// Returns:
//   - XSDValue: The parsed XSD value corresponding to the active member type.
//   - SimpleTypeDefinition: The first matching member simple type definition that accepted the literal.
//   - error: An error if no member type successfully validates the literal.
func (u *UnionType) findActiveMember(literal string) (XSDValue, SimpleTypeDefinition, error) {
	var lastErr error
	// Step 2: Sequentially Evaluate Members
	// Iterate through the member types in their specified order of definition until a match is found.
	for _, member := range u.memberTypes {
		val, err := member.ValidateLexical(literal)
		if err == nil {
			return val, member, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, nil, errors.New("value does not match any union member type. Last error: " + lastErr.Error())
	}
	return nil, nil, errors.New("value does not match any union member type")
}
