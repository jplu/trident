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

// Package datatypes implements the XML Schema Definition (XSD) 1.1 Part 2
// datatype library, supporting atomic, list, and union varieties alongside
// their constraining facets.
//
// The package provides Go representations for primitive and derived datatypes
// defined in the XSD specification, enabling parsing, validation, comparison,
// and canonical serialization of XML literals.
//
// Key features include:
//   - Type Varieties: Support for Atomic, List, and Union simple type definitions.
//   - Constraining Facets: Core implementation of facets including length, minLength,
//     maxLength, pattern, enumeration, whiteSpace, boundary limits (maxInclusive,
//     maxExclusive, minExclusive, minInclusive), digit counts (totalDigits, fractionDigits),
//     and explicitTimezone.
//   - Temporal representation: Modeling of temporal concepts, including dateTime,
//     time, date, gYear, gYearMonth, gMonth, gMonthDay, gDay, and durations (with support for
//     yearMonthDuration and dayTimeDuration) with timezone offset handling.
//   - Numeric modeling: Arbitrary-precision calculations for decimal and integer
//     types using big.Rat and big.Int, alongside fixed-precision mappings for IEEE 754
//     float/double types and standard Go integer types.
//   - Spec Alignment: Structured parsing behavior incorporating whitespace normalization
//     strategies ('preserve', 'replace', and 'collapse') and canonical lexical mappings
//     prescribed by the specification.
package datatypes
