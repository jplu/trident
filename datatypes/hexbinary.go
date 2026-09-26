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
	"encoding/hex"
	"errors"
	"strings"
)

// HexBinary represents the hexBinary datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.15)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// An arbitrary finite-length sequence of binary octets represented as a Go byte slice.
type HexBinary []byte

// ParseHexBinary parses a string literal matching the lexical representation of xsd:hexBinary.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.15.2)
//
// Parameters:
//   - s: A string literal containing pairs of hexadecimal digits representing binary octets.
//
// Returns:
//   - HexBinary: The parsed HexBinary value containing the decoded byte slice.
//   - error: An error if the input has an odd length or contains invalid hexadecimal characters.
func ParseHexBinary(s string) (HexBinary, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.15.2)
	// The lexical space of hexBinary consists of pairs of hexadecimal digits,
	// where each pair encodes a single octet. Consequently, the length of the
	// raw lexical string must be an even number of characters.
	if len(s)%2 != 0 {
		return nil, errors.New("hexBinary requires an even number of characters")
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return HexBinary(b), nil
}

// String returns the canonical lexical representation of the HexBinary value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section E.4.1)
//
// Parameters:
//
// Returns:
//   - string: The uppercase hexadecimal representation of the binary data.
func (h HexBinary) String() string {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section E.4.1)
	// The canonical mapping for hexBinary must use only the uppercase
	// alphanumeric forms of the characters A through F.
	return strings.ToUpper(hex.EncodeToString(h))
}

// IsIdenticalWith checks if this HexBinary is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.15.1)
//
// Parameters:
//   - other: The other XSDValue to compare with.
//
// Returns:
//   - bool: True if the values are identical, false otherwise.
func (h HexBinary) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(HexBinary); ok {
		return bytes.Equal(h, o)
	}
	return false
}

// Length returns the length of the HexBinary value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.3.15.1)
//
// Parameters:
//
// Returns:
//   - int: The number of binary octets.
func (h HexBinary) Length() int {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.3.15.1)
	// The length of a hexBinary value is mathematically defined as the number
	// of binary octets contained within the decoded sequence.
	return len(h)
}
