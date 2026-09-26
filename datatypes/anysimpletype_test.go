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

import (
	"testing"
)

// TestAnySimpleTypeString tests the string representation of AnySimpleType.
func TestAnySimpleTypeString(t *testing.T) {
	tests := []struct {
		name     string
		input    AnySimpleType
		expected string
	}{
		{
			name:     "non-empty value",
			input:    AnySimpleType{Value: "hello world"},
			expected: "hello world",
		},
		{
			name:     "empty value",
			input:    AnySimpleType{Value: ""},
			expected: "",
		},
		{
			name:     "special characters",
			input:    AnySimpleType{Value: "<xml>&amp;</xml>"},
			expected: "<xml>&amp;</xml>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.String(); got != tt.expected {
				t.Errorf("AnySimpleType.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestAnySimpleTypeIsIdenticalWith tests the equality check for AnySimpleType.
func TestAnySimpleTypeIsIdenticalWith(t *testing.T) {
	tests := []struct {
		name     string
		a        AnySimpleType
		other    XSDValue
		expected bool
	}{
		{
			name:     "identical values",
			a:        AnySimpleType{Value: "test"},
			other:    AnySimpleType{Value: "test"},
			expected: true,
		},
		{
			name:     "both empty values",
			a:        AnySimpleType{Value: ""},
			other:    AnySimpleType{Value: ""},
			expected: true,
		},
		{
			name:     "different values",
			a:        AnySimpleType{Value: "foo"},
			other:    AnySimpleType{Value: "bar"},
			expected: false,
		},
		{
			name:     "different XSDValue type",
			a:        AnySimpleType{Value: "test"},
			other:    String("test"),
			expected: false,
		},
		{
			name:     "nil value",
			a:        AnySimpleType{Value: "test"},
			other:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsIdenticalWith(tt.other); got != tt.expected {
				t.Errorf("AnySimpleType.IsIdenticalWith() = %v, want %v", got, tt.expected)
			}
		})
	}
}
