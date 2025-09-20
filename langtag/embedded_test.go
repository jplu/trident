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
	"strings"
	"testing"
)

// TestNewParser_Success verifies that NewParser successfully creates a parser instance
// from the valid, embedded IANA registry.
// As per RFC 5646, Section 5.1, the IANA Language Subtag Registry is the
// source for valid subtags. NewParser is expected to successfully process
// this data when it is correctly formatted as described in Section 3.1.1.
func TestNewParser_Success(t *testing.T) {
	parser, err := NewParser()

	if err != nil {
		t.Fatalf("NewParser() returned an unexpected error with valid data: %v", err)
	}
	if parser.registry == nil {
		t.Fatal("parser.registry should not be nil after successful initialization")
	}
	if len(parser.registry.Records) == 0 {
		t.Fatal("parser.registry.Records should not be empty after successful initialization")
	}

	// RFC 5646, Section 2.2.1 specifies that two-character primary language
	// subtags are derived from ISO 639-1. 'en' is a fundamental example.
	expectedKey := "language:en"
	if _, ok := parser.registry.Records[expectedKey]; !ok {
		t.Errorf("registry missing fundamental record for subtag 'en' (expected key: %q)", expectedKey)
	}

	// RFC 5646, Section 3.1.2 requires a 'File-Date' record.
	// Its presence indicates the registry header was parsed correctly.
	if parser.registry.FileDate == "" {
		t.Error("registry missing expected 'File-Date'")
	}
}

// TestNewParser_EmptyRegistry ensures that NewParser returns an error when the
// embedded registry data is empty.
// While RFC 5646 does not explicitly define behavior for an empty registry file,
// a robust implementation should fail gracefully. This test verifies that
// NewParser returns an error, preventing panics or undefined behavior.
func TestNewParser_EmptyRegistry(t *testing.T) {
	originalData := embeddedRegistryData
	embeddedRegistryData = []byte{} // Simulate missing embedded file
	defer func() {
		embeddedRegistryData = originalData
	}()

	parser, err := NewParser()
	if err == nil {
		t.Fatal("NewParser() should have failed with empty data but did not")
	}
	if parser != nil {
		t.Fatalf("NewParser() should have returned a nil parser on failure, but got %v", parser)
	}

	expectedErr := "embedded language-subtag-registry file is empty or not found"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, but got: %v", expectedErr, err.Error())
	}
}

// TestNewParser_CorruptedRegistry verifies that NewParser fails when the embedded
// registry data is malformed.
// RFC 5646, Section 3.1.1, defines a strict "record-jar" format. This test
// provides data that violates this format to ensure NewParser propagates
// the parsing error. The specific violation tested is an invalid range
// that mixes numeric and alphabetic characters.
func TestNewParser_CorruptedRegistry(t *testing.T) {
	originalData := embeddedRegistryData
	corruptedData := []byte("File-Date: 2024-07-25\n%%\nType: region\nSubtag: 123..abc\nDescription: Corrupted")
	embeddedRegistryData = corruptedData
	defer func() {
		embeddedRegistryData = originalData // Restore for other tests
	}()

	parser, err := NewParser()
	if err == nil {
		t.Fatal("NewParser() should have failed with corrupted data but did not")
	}
	if parser != nil {
		t.Fatalf("NewParser() should have returned a nil parser on failure, but got %v", parser)
	}

	if err.Error() == "" {
		t.Error("expected a descriptive error message for corrupted data, but got an empty error string")
	}
}
