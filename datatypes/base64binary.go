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

package datatypes

import (
	"bytes"
	"encoding/base64"
	"strings"
)

// Base64Binary represents the base64Binary primitive datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.16)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A finite-length sequence of zero or more binary octets, mapped from a lexical space
// of base64-encoded strings with 'collapse' whitespace normalization.
type Base64Binary []byte

// ParseBase64Binary parses a string literal matching the lexical representation of xsd:base64Binary.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.16.1)
//
// Parameters:
//   - s: A string representing the base64-encoded binary data.
//
// Returns:
//   - Base64Binary: The decoded byte slice representing the binary value.
//   - error: An error if the lexical representation contains invalid characters or has incorrect formatting.
func ParseBase64Binary(s string) (Base64Binary, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.16.1)
	// Whitespace characters (space #x20, carriage return #xD, line feed #xA, and tab #x9) are permitted within the
	// lexical representation of base64Binary and must be ignored during validation and decoding.

	// Step 1: Clean whitespace
	// Filter out all permitted whitespace characters prior to base64 decoding.
	clean := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, s)

	// Implementation Note: Standard base64 decoding
	// The Go standard library base64.StdEncoding decoder is utilized to map the cleaned lexical representation to the
	// value space.
	b, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		return nil, err
	}
	return Base64Binary(b), nil
}

// String returns the canonical lexical representation of the Base64Binary value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.16.2)
//
// Returns:
//   - string: The canonical base64-encoded string representing the binary sequence.
func (b Base64Binary) String() string {
	return base64.StdEncoding.EncodeToString(b)
}

// IsIdenticalWith checks if this Base64Binary value is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.16.4)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - bool: True if both values are identical sequences of binary octets, false otherwise.
func (b Base64Binary) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(Base64Binary); ok {
		return bytes.Equal(b, o)
	}
	return false
}

// Length returns the length of the Base64Binary value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.16.3)
//
// Returns:
//   - int: The number of binary octets contained within the sequence.
func (b Base64Binary) Length() int {
	return len(b)
}
