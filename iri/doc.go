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

// Package iri provides types and functions for working with Internationalized
// Resource Identifiers (IRIs) and IRI references as defined by RFC 3987.
//
// The package offers two main types:
//   - Ref: Represents an IRI reference, which can be either absolute (e.g., "http://example.com/a")
//     or relative (e.g., "/a", "b", "#c").
//   - Iri: Represents a guaranteed absolute IRI, which always includes a scheme.
//
// Key features include:
//   - Strict parsing and validation against RFC 3987.
//   - High-performance "unchecked" parsing for known-valid inputs.
//   - Reference resolution (Resolve) to compute an absolute IRI from a base and a relative reference.
//   - Relativization (Relativize) to compute a relative reference between two absolute IRIs.
//   - Zero-allocation resolution variants (ResolveTo) for performance-critical applications.
//   - Support for JSON marshalling and unmarshalling.
package iri
