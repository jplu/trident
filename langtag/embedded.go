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

package langtag

// Implementation Note: Blank import for go:embed
// The "embed" package must be imported, even if blank, to enable compiler support for the //go:embed directive.
import _ "embed"

// embeddedRegistryData represents the embedded raw IANA Language Subtag Registry data.
//
// Specification Reference:
// RFC 5646 (Section 3.1)
//
// Representation:
// A UTF-8 encoded byte slice of the record-jar formatted IANA Language Subtag Registry file.
//
//go:embed language-subtag-registry
var embeddedRegistryData []byte
