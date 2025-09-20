/*
Copyright 2025 Trident Authors

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

// Registry holds the parsed data from the IANA Language Subtag Registry file.
// It serves as the database for validating and canonicalizing language tags.
type Registry struct {
	Records  map[string]Record
	FileDate string
}

// Record represents a single entry in the IANA Language Subtag Registry.
// The fields correspond to the fields defined in RFC 5646, Section 3.1.
type Record struct {
	Type           string   `json:"type"`
	Subtag         string   `json:"subtag,omitempty"`
	Tag            string   `json:"tag,omitempty"`
	Description    []string `json:"description"`
	Added          string   `json:"added"`
	Deprecated     string   `json:"deprecated,omitempty"`
	PreferredValue string   `json:"preferredValue,omitempty"`
	Prefix         []string `json:"prefix,omitempty"`
	SuppressScript string   `json:"suppressScript,omitempty"`
	Macrolanguage  string   `json:"macrolanguage,omitempty"`
	Scope          string   `json:"scope,omitempty"`
	Comments       []string `json:"comments,omitempty"`
}

// IsGrandfathered returns true if the record type is 'grandfathered' or 'redundant'.
func (r *Record) IsGrandfathered() bool {
	return r.Type == "grandfathered" || r.Type == "redundant"
}
