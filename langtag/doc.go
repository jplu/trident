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

// Package langtag provides a comprehensive implementation for parsing, validating,
// and manipulating IETF BCP 47 language tags, as specified in RFC 5646.
//
// The package offers a robust and efficient solution for applications requiring
// strict conformance to language tag standards. It includes the full IANA
// Language Subtag Registry, embedded at compile time, ensuring that the module
// works out of the box with no additional setup.
//
// # Key Features
//
//   - Strict Validation: Performs full syntactic and semantic validation against
//     the IANA registry, including checks for deprecated or invalid subtags.
//   - Canonicalization: Normalizes language tags to their canonical form as
//     per RFC 5646, simplifying comparisons and storage. It also supports
//     converting canonical tags to the alternative "extlang form" for
//     compatibility purposes.
//   - High Performance: The primary entry point, NewParser(), returns a reusable,
//     thread-safe parser instance that is initialized only once.
//   - Full Component Access: Provides methods to easily access all parts of a
//     tag, including language, script, region, variants, extensions, and
//     private-use subtags.
//   - Self-Contained: The required IANA registry data is embedded directly into
//     the library, so it has no external file dependencies at runtime.
package langtag
