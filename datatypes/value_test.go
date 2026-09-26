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

// mockXSDValue represents a test implementation of the XSDValue interface.
type mockXSDValue struct {
	val string
}

// String returns the string representation for mockXSDValue.
func (m mockXSDValue) String() string {
	return m.val
}

// IsIdenticalWith checks value space identity for mockXSDValue.
func (m mockXSDValue) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(mockXSDValue); ok {
		return m.val == o.val
	}
	return false
}

// TestXSDValueInterface validates that the XSDValue interface contract is satisfied and functions as expected.
func TestXSDValueInterface(t *testing.T) {
	var v XSDValue = mockXSDValue{val: "sample"}

	if got := v.String(); got != "sample" {
		t.Fatalf("expected 'sample', got '%s'", got)
	}

	identicalVal := mockXSDValue{val: "sample"}
	if !v.IsIdenticalWith(identicalVal) {
		t.Fatal("expected values to be identical")
	}

	differentVal := mockXSDValue{val: "other"}
	if v.IsIdenticalWith(differentVal) {
		t.Fatal("expected values not to be identical")
	}

	var nilVal XSDValue
	if v.IsIdenticalWith(nilVal) {
		t.Fatal("expected non-nil value not to be identical with nil")
	}
}
