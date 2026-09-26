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

import (
	"errors"
	"strings"
)

// ListValue represents the list value variety as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.4.1.2)
//
// Representation:
// A slice of XSDValue items representing a finite-length sequence of atomic values.
type ListValue []XSDValue

// String returns the canonical lexical representation of the ListValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.4.1.2)
//
// Returns:
//   - string: The canonical space-separated string representation of the list.
func (l ListValue) String() string {
	strs := make([]string, len(l))
	for i, val := range l {
		strs[i] = val.String()
	}
	return strings.Join(strs, " ")
}

// IsIdenticalWith checks if this ListValue is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 2.2.1)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - bool: True if the two list values have the same length and their corresponding items are pairwise identical.
func (l ListValue) IsIdenticalWith(other XSDValue) bool {
	o, ok := other.(ListValue)
	if !ok || len(l) != len(o) {
		return false
	}
	for i := range l {
		if !l[i].IsIdenticalWith(o[i]) {
			return false
		}
	}
	return true
}

// Length returns the length of the ListValue value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1.1)
//
// Returns:
//   - int: The number of items in the list.
func (l ListValue) Length() int {
	return len(l)
}

// ListType represents the Simple Type Definition where the variety property is 'list' as defined in the governing
// specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Representation:
// A struct capturing properties of the List Simple Type Definition component, including its name, target namespace,
// base type, item type, facets, and fundamental facets.
type ListType struct {
	// name is the local name of the list type definition.
	name string
	// targetNamespace is the target namespace of the list type definition.
	targetNamespace string
	// baseType represents the simple type definition from which this list type is derived.
	baseType SimpleTypeDefinition
	// itemType represents the simple type definition of the items in the list.
	itemType SimpleTypeDefinition
	// facets is the set of constraining facets applied directly to this list type.
	facets []Facet
	// fundamental represents the fundamental facets of this list type definition.
	fundamental FundamentalFacets
}

// Name returns the local name of the ListType definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Returns:
//   - string: The local name of the list type definition, or empty if anonymous.
func (l *ListType) Name() string {
	return l.name
}

// TargetNamespace returns the target namespace of the ListType definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Returns:
//   - string: The target namespace URI of the list type definition, or empty.
func (l *ListType) TargetNamespace() string {
	return l.targetNamespace
}

// BaseType returns the simple type definition from which this ListType is derived.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Returns:
//   - SimpleTypeDefinition: The immediate base type definition of this list type.
func (l *ListType) BaseType() SimpleTypeDefinition {
	return l.baseType
}

// Variety returns the variety of the type definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Returns:
//   - Variety: The variety of the simple type, which is always VarietyList.
func (l *ListType) Variety() Variety {
	return VarietyList
}

// PrimitiveType returns the primitive type ancestor.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Returns:
//   - SimpleTypeDefinition: Always returns nil for list types since list types have no primitive type ancestor.
func (l *ListType) PrimitiveType() SimpleTypeDefinition {
	return nil
}

// ItemType returns the simple type definition of the items allowed within the list.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Returns:
//   - SimpleTypeDefinition: The item type definition of the list.
func (l *ListType) ItemType() SimpleTypeDefinition {
	return l.itemType
}

// MemberTypes returns the member types.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Returns:
//   - []SimpleTypeDefinition: Always returns nil for list types.
func (l *ListType) MemberTypes() []SimpleTypeDefinition {
	return nil
}

// Facets returns the slice of constraining facets applied to this ListType.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Returns:
//   - []Facet: The list of constraining facets applied to this type.
func (l *ListType) Facets() []Facet {
	return l.facets
}

// FundamentalFacets returns the fundamental facets associated with this ListType.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.1)
//
// Returns:
//   - FundamentalFacets: The set of fundamental facets.
func (l *ListType) FundamentalFacets() FundamentalFacets {
	return l.fundamental
}

// ValidateLexical validates a raw string literal against this list type definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.4)
//
// Parameters:
//   - literal: The raw lexical string to validate.
//
// Returns:
//   - XSDValue: The successfully parsed and validated ListValue.
//   - error: An error if validation against any of the facets or item type constraints fails.
func (l *ListType) ValidateLexical(literal string) (XSDValue, error) {
	if l.itemType == nil {
		return nil, errors.New("list type has no item type defined")
	}

	// Step 1: Normalize whitespace
	// Apply the 'collapse' whitespace normalization behavior prior to tokenization.
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.6)
	// Lists inherently apply the 'collapse' whiteSpace normalization before
	// tokenization, regardless of any whiteSpace facet specified on the item type.
	normalizedLiteral := collapseWhitespace(literal)

	// Step 2: Validate pattern and lexical facets
	// Check lexical-space constraining facets (like pattern) against the normalized string representation.
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.3.4)
	// The pattern facet of a list type is matched against the space-separated
	// literal representation of the entire list as a whole, rather than
	// against individual items.
	for _, f := range l.facets {
		if lf, ok := f.(LexicalFacet); ok {
			if err := lf.CheckLexical(normalizedLiteral); err != nil {
				return nil, err
			}
		}
	}

	// Step 3: Parse and validate list items
	// Split the normalized literal by single space characters and validate each token against the item type.
	var listVal ListValue
	if normalizedLiteral != "" {
		parts := strings.Split(normalizedLiteral, " ")
		listVal = make(ListValue, len(parts))
		for i, part := range parts {
			// Spec Rule: W3C XSD 1.1 Part 2 (Section 2.4.1.2)
			// Each space-separated item in the list must satisfy the lexical and value
			// constraints of the {item type}.
			item, err := l.itemType.ValidateLexical(part)
			if err != nil {
				return nil, err
			}
			listVal[i] = item
		}
	} else {
		listVal = make(ListValue, 0)
	}

	// Step 4: Validate value-space facets
	// Apply value-based constraining facets (length, minLength, maxLength, enumeration) to the list value.
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.1.4)
	// Value-space constraints on a list type (such as length, minLength,
	// maxLength, and enumeration) are evaluated against the parsed list value itself.
	for _, f := range l.facets {
		if _, isLexical := f.(LexicalFacet); !isLexical {
			if err := f.Check(listVal); err != nil {
				return nil, err
			}
		}
	}

	return listVal, nil
}
