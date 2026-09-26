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

// Ordered represents the ordered fundamental facet of a datatype.
//
// Specification Reference:
// XSD 1.1 Part 2 (Section 4.2.1)
//
// Value Space:
// It describes the mathematical and logical ordering relationships
// among the values in the value space: false (unordered), partial, or total.
type Ordered int

const (
	// OrderedFalse indicates that no ordering relation is defined for the value space.
	OrderedFalse Ordered = iota

	// OrderedPartial indicates that a partial ordering relation is defined for the value space.
	OrderedPartial

	// OrderedTotal indicates that a total ordering relation is defined for the value space.
	OrderedTotal
)

// Cardinality represents the cardinality fundamental facet of a datatype.
//
// Specification Reference:
// XSD 1.1 Part 2 (Section 4.2.3)
//
// Value Space:
// It indicates whether the value space of the datatype is finite
// or countably infinite.
type Cardinality int

const (
	// Finite indicates that the value space of the datatype contains a finite number of values.
	Finite Cardinality = iota

	// CountablyInfinite indicates that the value space is countably infinite.
	CountablyInfinite
)

// FundamentalFacets contains the properties representing the fundamental
// characteristics of a datatype.
//
// Specification Reference:
// XSD 1.1 Part 2 (Section 4.2).
type FundamentalFacets struct {
	// Ordered specifies the ordering relationship of the datatype.
	// Reference: XSD 1.1 Part 2 (Section 4.2.1)
	Ordered Ordered

	// Bounded specifies whether the value space has both an upper and a lower bound.
	// Reference: XSD 1.1 Part 2 (Section 4.2.2)
	Bounded bool

	// Cardinality specifies the size of the value space.
	// Reference: XSD 1.1 Part 2 (Section 4.2.3)
	Cardinality Cardinality

	// Numeric specifies whether the values are mathematically numeric.
	// Reference: XSD 1.1 Part 2 (Section 4.2.4)
	Numeric bool
}

// DefaultFundamentalFacets returns a FundamentalFacets instance initialized with
// the default baseline characteristics: unordered, unbounded, countably infinite
// cardinality, and non-numeric.
func DefaultFundamentalFacets() FundamentalFacets {
	// Spec Rule: XSD 1.1 Part 2 (Section 4.2)
	// Types derived from anySimpleType default to unordered, unbounded,
	// and countably infinite unless constrained.
	return FundamentalFacets{
		Ordered:     OrderedFalse,
		Bounded:     false,
		Cardinality: CountablyInfinite,
		Numeric:     false,
	}
}
