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

import (
	"errors"
	"strings"

	// TODO: At some point implement my own Bidi module.
	"golang.org/x/text/unicode/bidi"
)

// validateBidiComponent checks a component string against the structural rules for bidirectional IRIs.
//
// Specification Reference:
// RFC 3987 (Section 4.2)
//
// Parameters:
//   - component: The component string to be checked.
//
// Returns:
//   - error: A BidiGuidelineError if the component violates the bidirectional constraint rules, or nil if valid.
func validateBidiComponent(component string) error {
	if component == "" {
		return nil
	}

	// Step 1: Scan for character directionality classes
	// We iterate over the runes to determine the presence of LTR and RTL characters.
	runes := []rune(component)
	var hasLTR, hasRTL bool

	for _, r := range runes {
		prop, _ := bidi.LookupRune(r)
		class := prop.Class()
		switch class {
		case bidi.R, bidi.AL:
			hasRTL = true
		case bidi.L:
			hasLTR = true
		case bidi.EN,
			bidi.ES,
			bidi.ET,
			bidi.AN,
			bidi.CS,
			bidi.B,
			bidi.S,
			bidi.WS,
			bidi.ON,
			bidi.BN,
			bidi.NSM,
			bidi.Control,
			bidi.LRO,
			bidi.RLO,
			bidi.LRE,
			bidi.RLE,
			bidi.PDF,
			bidi.LRI,
			bidi.RLI,
			bidi.FSI,
			bidi.PDI:
			// Implementation Note: Skip neutral and formatting characters
			// Neutral characters and bidirectional formatting characters are ignored during mixed-directionality
			// analysis.
		}
	}

	// Spec Rule: RFC 3987 (Section 4.2)
	// Rule 1: A component SHOULD NOT use both right-to-left and left-to-right characters.
	if hasLTR && hasRTL {
		return &BidiGuidelineError{
			Rule:      "Rule 1",
			Component: component,
			Message:   "mixed left-to-right and right-to-left characters",
		}
	}

	// Spec Rule: RFC 3987 (Section 4.2)
	// Rule 2: A component using right-to-left characters SHOULD start and end with right-to-left characters.
	if hasRTL {
		// Check the first character of the component.
		propFirst, _ := bidi.LookupRune(runes[0])
		classFirst := propFirst.Class()
		isFirstRTL := classFirst == bidi.R || classFirst == bidi.AL
		if !isFirstRTL {
			return &BidiGuidelineError{
				Rule:      "Rule 2",
				Component: component,
				Message:   "right-to-left parts must start with right-to-left characters",
			}
		}

		// Check the last character of the component.
		propLast, _ := bidi.LookupRune(runes[len(runes)-1])
		classLast := propLast.Class()
		isLastRTL := classLast == bidi.R || classLast == bidi.AL
		if !isLastRTL {
			return &BidiGuidelineError{
				Rule:      "Rule 2",
				Component: component,
				Message:   "right-to-left parts must end with right-to-left characters",
			}
		}
	}

	return nil
}

// validateBidiHost checks a host string against the bidirectional constraints.
//
// Specification Reference:
// RFC 3987 (Section 4.2)
//
// Parameters:
//   - host: The host string to be checked.
//
// Returns:
//   - error: A BidiGuidelineError if any label in the host violates the bidirectional rules, or nil if valid.
func validateBidiHost(host string) error {
	// Implementation Note: Exception for IP literal formatting
	// For IP literals enclosed in brackets, bidirectional constraints do not apply.
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return nil
	}

	// Step 1: Split the host into individual labels
	// Bidirectional rules mandate evaluating each dot-separated label independently.
	labels := strings.SplitSeq(host, ".")

	// Step 2: Validate each label against bidirectional constraints
	// Spec Rule: RFC 3987 (Section 4.2)
	// For hostnames, each dot-separated label is treated as an individual component for Bidi validation.
	for label := range labels {
		if err := validateBidiComponent(label); err != nil {
			// Implementation Note: Error context wrapping
			// Extract and enrich the internal error to provide detailed host diagnostics.
			bidiErr := func() *BidiGuidelineError {
				target := &BidiGuidelineError{}
				_ = errors.As(err, &target)
				return target
			}()
			bidiErr.Message = "Invalid IRI host label: " + bidiErr.Message
			bidiErr.Component = label + " in host '" + host + "'"
			return bidiErr
		}
	}
	return nil
}
