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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package iri

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestRef_MarshalJSON tests the JSON marshaling of a Ref.
func TestRef_MarshalJSON(t *testing.T) {
	ref := mustParseRef(t, "http://example.com/a?b#c")
	jsonData, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	expected := `"http://example.com/a?b#c"`
	if string(jsonData) != expected {
		t.Errorf("Expected JSON string '%s', got '%s'", expected, string(jsonData))
	}
}

// TestRef_UnmarshalJSON tests the JSON unmarshaling of a Ref.
func TestRef_UnmarshalJSON(t *testing.T) {
	t.Run("Valid IRI", func(t *testing.T) {
		var ref Ref
		jsonData := []byte(`"http://example.com/a?b#c"`)
		err := json.Unmarshal(jsonData, &ref)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
		expected := "http://example.com/a?b#c"
		if ref.String() != expected {
			t.Errorf("Expected unmarshaled string '%s', got '%s'", expected, ref.String())
		}
	})

	t.Run("Invalid IRI", func(t *testing.T) {
		var ref Ref
		jsonData := []byte(`"http://example.com/["`)
		err := json.Unmarshal(jsonData, &ref)
		if err == nil {
			t.Fatal("Expected an error for invalid IRI, but got none")
		}
		if !strings.Contains(err.Error(), "Invalid IRI character") {
			t.Errorf("Expected error message to contain 'Invalid IRI character', got '%s'", err.Error())
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		var ref Ref
		jsonData := []byte(`not-a-string`)
		err := json.Unmarshal(jsonData, &ref)
		if err == nil {
			t.Fatal("Expected an error for invalid JSON, but got none")
		}
	})
}

// TestIri_MarshalJSON tests the JSON marshaling of an Iri.
func TestIri_MarshalJSON(t *testing.T) {
	iri := mustParseIri(t, "http://example.com/a")
	jsonData, err := json.Marshal(iri)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	expected := `"http://example.com/a"`
	if string(jsonData) != expected {
		t.Errorf("Expected JSON string '%s', got '%s'", expected, string(jsonData))
	}
}

// TestIri_UnmarshalJSON tests the JSON unmarshaling of an Iri, including validation.
func TestIri_UnmarshalJSON(t *testing.T) {
	t.Run("Valid Absolute IRI", func(t *testing.T) {
		var iri Iri
		jsonData := []byte(`"http://example.com"`)
		err := json.Unmarshal(jsonData, &iri)
		if err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
		if iri.String() != "http://example.com" {
			t.Errorf("Expected unmarshaled string 'http://example.com', got '%s'", iri.String())
		}
	})

	t.Run("Relative IRI", func(t *testing.T) {
		var iri Iri
		jsonData := []byte(`"/relative/path"`)
		err := json.Unmarshal(jsonData, &iri)
		if err == nil {
			t.Fatal("Expected an error for relative IRI, but got none")
		}
		if !strings.Contains(err.Error(), "No scheme found") {
			t.Errorf("Expected error message to contain 'No scheme found', got '%s'", err.Error())
		}
	})

	t.Run("Invalid IRI", func(t *testing.T) {
		var iri Iri
		jsonData := []byte(`"http://["`)
		err := json.Unmarshal(jsonData, &iri)
		if err == nil {
			t.Fatal("Expected an error for invalid IRI, but got none")
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		var iri Iri
		err := iri.UnmarshalJSON([]byte("not-json"))
		if err == nil {
			t.Fatal("Expected an error for invalid JSON, but got none")
		}
	})
}
