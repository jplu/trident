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

// LexicalMapper represents the lexical-to-value mapping function as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.3)
//
// Representation:
// Translates a normalized string literal into its corresponding value in the value space.
type LexicalMapper func(literal string) (XSDValue, error)

// AtomicType represents the atomic simple type definition component as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.2)
//
// Representation:
// A simple type definition with a variety of atomic, containing a value space
// of indivisible and distinct values, and a lexical space of string representations.
type AtomicType struct {
	// name represents the local name of the simple type.
	name string
	// targetNamespace represents the target namespace of the simple type.
	targetNamespace string
	// baseType represents the immediate base type definition of this type.
	baseType SimpleTypeDefinition
	// primitiveType represents the primitive type definition from which this type descends.
	primitiveType SimpleTypeDefinition
	// facets represents the collection of constraining facets applied to this type.
	facets []Facet
	// fundamental represents the fundamental facets of this simple type.
	fundamental FundamentalFacets
	// lexicalMapper represents the parser function for built-in primitive types.
	lexicalMapper LexicalMapper
}

// Name returns the local name of the atomic simple type definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - string: The local name of the simple type definition, or empty if anonymous.
func (a *AtomicType) Name() string {
	return a.name
}

// TargetNamespace returns the target namespace of the atomic simple type definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - string: The target namespace URI of the simple type definition, or empty if absent.
func (a *AtomicType) TargetNamespace() string {
	return a.targetNamespace
}

// BaseType returns the immediate base simple type definition of this atomic type.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - SimpleTypeDefinition: The base type definition from which this type is derived.
func (a *AtomicType) BaseType() SimpleTypeDefinition {
	return a.baseType
}

// Variety returns the variety of this simple type definition, which is always VarietyAtomic.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - Variety: The variety enum value representing atomic.
func (a *AtomicType) Variety() Variety {
	return VarietyAtomic
}

// PrimitiveType returns the ancestor built-in primitive type definition of this atomic type.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - SimpleTypeDefinition: The primitive type definition from which this type descends.
func (a *AtomicType) PrimitiveType() SimpleTypeDefinition {
	return a.primitiveType
}

// ItemType returns nil since atomic simple types do not have an item type.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - SimpleTypeDefinition: Always returns nil for atomic types.
func (a *AtomicType) ItemType() SimpleTypeDefinition {
	return nil
}

// MemberTypes returns nil since atomic simple types do not have union member types.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - []SimpleTypeDefinition: Always returns nil for atomic types.
func (a *AtomicType) MemberTypes() []SimpleTypeDefinition {
	return nil
}

// Facets returns the set of constraining facets applied directly to this atomic type.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Parameters:
//
// Returns:
//   - []Facet: The list of constraining facets specified on this type.
func (a *AtomicType) Facets() []Facet {
	return a.facets
}

// FundamentalFacets returns the derived fundamental facets of this atomic simple type.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.2)
//
// Parameters:
//
// Returns:
//   - FundamentalFacets: The fundamental characteristics associated with this type.
func (a *AtomicType) FundamentalFacets() FundamentalFacets {
	return a.fundamental
}

// ValidateLexical validates a raw string literal against this atomic type definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.4)
//
// Parameters:
//   - literal: The raw lexical string representation to validate.
//
// Returns:
//   - XSDValue: The successfully parsed value mapping on success.
//   - error: An error if validation or parsing constraints are violated.
func (a *AtomicType) ValidateLexical(literal string) (XSDValue, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.1.2.3)
	// Whitespace normalization must occur prior to validation and mapping. The
	// default behavior is 'collapse', except for the built-in primitive 'string'
	// and its derivatives, which default to 'preserve'.
	wsVal := WhiteSpaceCollapse
	if (a.primitiveType != nil && a.primitiveType.Name() == primitiveString) || a.name == primitiveString {
		wsVal = WhiteSpacePreserve
	}

	for _, f := range a.facets {
		if wsf, ok := f.(FacetWhiteSpace); ok {
			wsVal = wsf.Value
			break
		}
	}
	normalizedLiteral := NewFacetWhiteSpace(wsVal, false).Normalize(literal)

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.4)
	// Lexical-space facets (such as pattern) must be evaluated directly against
	// the normalized lexical representation prior to any value-space mapping.
	for _, f := range a.facets {
		if lf, ok := f.(LexicalFacet); ok {
			if err := lf.CheckLexical(normalizedLiteral); err != nil {
				return nil, err
			}
		}
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 2.2)
	// Translate the normalized lexical representation into its respective value space.
	var val XSDValue
	var err error

	switch {
	case a.lexicalMapper != nil:
		val, err = a.lexicalMapper(normalizedLiteral)
	case a.primitiveType != nil:
		val, err = a.primitiveType.ValidateLexical(normalizedLiteral)
	default:
		return nil, errors.New("atomic type has no primitive type mapping")
	}

	if err != nil {
		return nil, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.1.2.3)
	// Value-space constraining facets (such as bounds, enumeration, or digits)
	// are checked against the mapped value.
	for _, f := range a.facets {
		if _, isLexical := f.(LexicalFacet); !isLexical {
			if cErr := f.Check(val); cErr != nil {
				return nil, cErr
			}
		}
	}

	return val, nil
}
