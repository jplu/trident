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

package iri

import "encoding/json"

// MarshalJSON marshals the Ref into its JSON representation.
//
// Specification Reference:
// RFC 3987 (Section 6.3)
//
// Returns:
//   - []byte: The JSON-encoded string representation of the IRI reference.
//   - error: Any error encountered during JSON serialization.
func (r *Ref) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.iri)
}

// UnmarshalJSON unmarshals a JSON string into a Ref.
//
// Specification Reference:
// RFC 3987 (Section 3.1)
//
// Parameters:
//   - data: The JSON-encoded data representing the IRI reference.
//
// Returns:
//   - error: An error if the JSON is invalid, or if the decoded string is not a valid IRI reference.
func (r *Ref) UnmarshalJSON(data []byte) error {
	// Step 1: JSON Decode
	// Decode the incoming JSON byte array into a Go string.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// Spec Rule: RFC 3987 (Section 3.1)
	// Validate and parse the decoded string as-is without applying NFC normalization to preserve the exact sequence.
	newRef, err := ParseRef(s)
	if err != nil {
		return err
	}
	*r = *newRef
	return nil
}

// MarshalJSON marshals the absolute Iri into its JSON representation.
//
// Specification Reference:
// RFC 3987 (Section 6.3)
//
// Returns:
//   - []byte: The JSON-encoded string representation of the absolute IRI.
//   - error: Any error encountered during JSON serialization.
func (i *Iri) MarshalJSON() ([]byte, error) {
	return i.Ref.MarshalJSON()
}

// UnmarshalJSON unmarshals a JSON string into an absolute Iri.
//
// Specification Reference:
// RFC 3987 (Section 2.2)
//
// Parameters:
//   - data: The JSON-encoded data representing the absolute IRI.
//
// Returns:
// - error: An error if the JSON is invalid, the decoded string is not a valid IRI reference, or the resulting IRI is
// not absolute.
func (i *Iri) UnmarshalJSON(data []byte) error {
	// Step 1: Base Reference Unmarshal
	// Unmarshal the JSON payload into a temporary generic Ref instance.
	var ref Ref
	if err := ref.UnmarshalJSON(data); err != nil {
		return err
	}

	// Spec Rule: RFC 3987 (Section 2.2)
	// Ensure the parsed IRI is absolute (i.e., has a valid scheme) to conform to the structural rules of an IRI.
	newIri, err := NewIriFromRef(&ref)
	if err != nil {
		return err
	}
	*i = *newIri
	return nil
}
