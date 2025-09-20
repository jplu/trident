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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package langtag

import (
	"testing"
)

// TestRecord_IsGrandfathered tests the IsGrandfathered method of the Record struct.
//
// RFC 5646 Section 2.2.8 defines grandfathered and redundant registrations.
// It states: "A redundant tag is a grandfathered registration..." and
// "The remainder of the previously registered tags are 'grandfathered'".
// The implementation checks if the record's type is either 'grandfathered'
// or 'redundant'. This test validates that behavior.
func TestRecord_IsGrandfathered(t *testing.T) {
	testCases := []struct {
		name     string
		record   Record
		expected bool
	}{
		{
			name:     "Type is grandfathered",
			record:   Record{Type: "grandfathered"},
			expected: true,
		},
		{
			name:     "Type is redundant",
			record:   Record{Type: "redundant"},
			expected: true,
		},
		{
			name:     "Type is language",
			record:   Record{Type: "language"},
			expected: false,
		},
		{
			name:     "Type is extlang",
			record:   Record{Type: "extlang"},
			expected: false,
		},
		{
			name:     "Type is script",
			record:   Record{Type: "script"},
			expected: false,
		},
		{
			name:     "Type is region",
			record:   Record{Type: "region"},
			expected: false,
		},
		{
			name:     "Type is variant",
			record:   Record{Type: "variant"},
			expected: false,
		},
		{
			name:     "Type is empty",
			record:   Record{Type: ""},
			expected: false,
		},
		{
			name:     "Type is an arbitrary string",
			record:   Record{Type: "other"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.record.IsGrandfathered()
			if result != tc.expected {
				t.Errorf("IsGrandfathered() for type '%s' = %v; want %v", tc.record.Type, result, tc.expected)
			}
		})
	}
}
