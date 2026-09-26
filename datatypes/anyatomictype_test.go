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

// nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to
// test unexported functions.
package datatypes

import "testing"

// TestAnyAtomicTypeString tests the String method of AnyAtomicType.
func TestAnyAtomicTypeString(t *testing.T) {
	tests := []struct {
		name     string
		instance AnyAtomicType
		expected string
	}{
		{
			name:     "empty value string",
			instance: AnyAtomicType{Value: ""},
			expected: "",
		},
		{
			name:     "non-empty value string",
			instance: AnyAtomicType{Value: "sampleText123"},
			expected: "sampleText123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.instance.String(); got != tt.expected {
				t.Errorf("AnyAtomicType.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestAnyAtomicTypeIsIdenticalWith tests the IsIdenticalWith method of AnyAtomicType.
func TestAnyAtomicTypeIsIdenticalWith(t *testing.T) {
	valA := AnyAtomicType{Value: "test"}
	valSame := AnyAtomicType{Value: "test"}
	valDiff := AnyAtomicType{Value: "other"}
	otherType := String("test")

	tests := []struct {
		name     string
		receiver AnyAtomicType
		other    XSDValue
		expected bool
	}{
		{
			name:     "identical AnyAtomicType values",
			receiver: valA,
			other:    valSame,
			expected: true,
		},
		{
			name:     "different AnyAtomicType values",
			receiver: valA,
			other:    valDiff,
			expected: false,
		},
		{
			name:     "different XSDValue type",
			receiver: valA,
			other:    otherType,
			expected: false,
		},
		{
			name:     "nil other XSDValue",
			receiver: valA,
			other:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.receiver.IsIdenticalWith(tt.other); got != tt.expected {
				t.Errorf("AnyAtomicType.IsIdenticalWith() = %v, want %v", got, tt.expected)
			}
		})
	}
}
