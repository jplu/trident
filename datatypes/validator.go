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

// Validate validates a raw XML string literal against an instantiated simple type definition.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 4.1.4)
//
// Parameters:
//   - literal: The raw, unmodified XML string literal to be validated.
//   - typeDef: The SimpleTypeDefinition schema component to validate against.
//
// Returns:
//   - XSDValue: The parsed and validated value conforming to the schema component.
//   - error: An error describing the validation or lexical parsing failure.
func Validate(literal string, typeDef SimpleTypeDefinition) (XSDValue, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 4.1.4)
	// Whitespace normalization must be context-dependent. Specifically, atomic types normalize based on their explicit
	// or inherited whiteSpace facet, list types inherently collapse whitespace, and union types normalize the raw
	// literal per member type during evaluation.
	//
	// Implementation Note: Preservation of the raw XML literal
	// The raw, unmodified literal is passed down directly to the SimpleTypeDefinition. This prevents Union types from
	// irreversibly altering whitespace before evaluating string-based member types that require raw formatting.
	return typeDef.ValidateLexical(literal)
}
