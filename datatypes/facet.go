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

// FacetName represents the constraining facet name as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.2)
//
// Representation:
// An XML non-colonized name (NCName) identifying a specific type of constraining facet.
type FacetName string

const (
	// NameLength represents the 'length' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.1).
	NameLength FacetName = "length"

	// NameMinLength represents the 'minLength' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.2).
	NameMinLength FacetName = "minLength"

	// NameMaxLength represents the 'maxLength' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.3).
	NameMaxLength FacetName = "maxLength"

	// NamePattern represents the 'pattern' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.4).
	NamePattern FacetName = "pattern"

	// NameEnumeration represents the 'enumeration' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.5).
	NameEnumeration FacetName = "enumeration"

	// NameWhiteSpace represents the 'whiteSpace' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.6).
	NameWhiteSpace FacetName = "whiteSpace"

	// NameMaxInclusive represents the 'maxInclusive' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.7).
	NameMaxInclusive FacetName = "maxInclusive"

	// NameMaxExclusive represents the 'maxExclusive' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.8).
	NameMaxExclusive FacetName = "maxExclusive"

	// NameMinExclusive represents the 'minExclusive' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.9).
	NameMinExclusive FacetName = "minExclusive"

	// NameMinInclusive represents the 'minInclusive' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.10).
	NameMinInclusive FacetName = "minInclusive"

	// NameTotalDigits represents the 'totalDigits' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.11).
	NameTotalDigits FacetName = "totalDigits"

	// NameFractionDigits represents the 'fractionDigits' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.12).
	NameFractionDigits FacetName = "fractionDigits"

	// NameExplicitTimezone represents the 'explicitTimezone' constraining facet.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.16).
	NameExplicitTimezone FacetName = "explicitTimezone"
)

// BaseFacet represents the baseline properties of a constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.2)
//
// Representation:
// A structured metadata wrapper tracking the name and the fixed override status of a facet.
type BaseFacet struct {
	// Name specifies the type of the constraining facet.
	Name FacetName

	// Fixed determines whether the facet's value cannot be overridden by further derived type definitions.
	//
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 2.4.1.2)
	// If the fixed property is true, the value of the facet cannot be overridden in derived simple types.
	Fixed bool
}

// LexicalFacet represents the lexical-space constraining facet as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.4)
//
// Representation:
// An interface for facets that evaluate and restrict the lexical representation prior to value mapping.
type LexicalFacet interface {
	// CheckLexical checks the pre-lexical normalized string against the facet constraint.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.1.4)
	//
	// Parameters:
	//   - literal: The pre-lexical normalized string literal to validate.
	//
	// Returns:
	//   - error: An error if validation fails or the literal does not conform to the pattern.
	CheckLexical(literal string) error
}

// LengthProvider represents the length-constrainable datatype provider as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.1)
//
// Representation:
// An interface for types whose value spaces support character, octet, or list item length measurements.
type LengthProvider interface {
	// Length measures the length of the conforming value.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.1.1)
	//
	// Parameters:
	//
	// Returns:
	//   - int: The exact length measured in characters, binary octets, or list items.
	Length() int
}

// DecimalProvider represents the decimal-projection provider as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.3.11)
//
// Representation:
// An interface for types that can map their value spaces into arbitrary-precision decimal representations.
type DecimalProvider interface {
	// ToDecimal converts the underlying value into an arbitrary-precision Decimal representation.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.3.11)
	//
	// Parameters:
	//
	// Returns:
	//   - Decimal: The projected arbitrary-precision decimal value.
	ToDecimal() Decimal
}

// TimezoneProvider represents the timezone offset provider as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Representation:
// An interface for temporal datatypes that possess an optional timezone offset component.
type TimezoneProvider interface {
	// TimezoneOffset retrieves the optional timezone offset of the temporal value.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 3.2.7)
	//
	// Parameters:
	//
	// Returns:
	//   - *TimezoneOffset: The pointer to the TimezoneOffset struct, or nil if unzoned.
	TimezoneOffset() *TimezoneOffset
}

// ComparableProvider represents the order-comparable datatype provider as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.2.1)
//
// Representation:
// An interface for datatypes possessing a partial or total order to validate boundary constraining facets.
type ComparableProvider interface {
	// Compare evaluates the ordering relationship of the value against another value.
	//
	// Specification Reference:
	// W3C XSD 1.1 Part 2 (Section 4.2.1)
	//
	// Parameters:
	//   - other: The other XSDValue to compare against.
	//
	// Returns:
	//   - int: An integer representing the comparison result (-1 if less, 0 if equal, 1 if greater).
	//   - error: An error if the types are incomparable in their value space.
	Compare(other XSDValue) (int, error)
}
