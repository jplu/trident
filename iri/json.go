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

// MarshalJSON implements the json.Marshaler interface, encoding the Ref as a JSON string.
func (r *Ref) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.iri)
}

// UnmarshalJSON implements the json.Unmarshaler interface. It decodes a JSON string
// into a Ref, performing validation in the process. It does not perform NFC normalization.
func (r *Ref) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	newRef, err := ParseRef(s)
	if err != nil {
		return err
	}
	*r = *newRef
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (i *Iri) MarshalJSON() ([]byte, error) {
	return i.Ref.MarshalJSON()
}

// UnmarshalJSON implements the json.Unmarshaler interface, ensuring the
// decoded IRI is absolute.
func (i *Iri) UnmarshalJSON(data []byte) error {
	var ref Ref
	if err := ref.UnmarshalJSON(data); err != nil {
		return err
	}
	newIri, err := NewIriFromRef(&ref)
	if err != nil {
		return err
	}
	*i = *newIri
	return nil
}
