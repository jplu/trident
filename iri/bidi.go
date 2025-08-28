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

package iri

import (
	"errors"
	"strings"

	// TODO: At some point implement my own Bidi module.
	"golang.org/x/text/unicode/bidi"
)

// validateBidiComponent checks a component string against the structural rules
// for bidirectional IRIs as defined in RFC 3987, Section 4.2.
//
// Rule 1: A component SHOULD NOT use both right-to-left and left-to-right characters.
// Rule 2: A component using right-to-left characters SHOULD start and end with
//
//	right-to-left characters.
//
// This function will return an error if a component violates these "SHOULD" rules,
// providing stricter validation.
func validateBidiComponent(component string) error {
	if component == "" {
		return nil
	}

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
		case bidi.EN, bidi.ES, bidi.ET, bidi.AN, bidi.CS, bidi.B, bidi.S, bidi.WS, bidi.ON, bidi.BN, bidi.NSM,
			bidi.Control, bidi.LRO, bidi.RLO, bidi.LRE, bidi.RLE, bidi.PDF, bidi.LRI, bidi.RLI, bidi.FSI, bidi.PDI:
			// These are neutral characters and do not affect LTR/RTL detection for the purpose of this validation.
		}
	}

	// Rule 1: Disallow mixing of LTR and RTL characters in the same component.
	if hasLTR && hasRTL {
		return &kindError{
			message: "Invalid IRI component: mixed left-to-right and right-to-left characters",
			details: component,
		}
	}

	// Rule 2 applies only if the component contains RTL characters.
	if hasRTL {
		// Check the first character of the component.
		propFirst, _ := bidi.LookupRune(runes[0])
		classFirst := propFirst.Class()
		isFirstRTL := classFirst == bidi.R || classFirst == bidi.AL
		if !isFirstRTL {
			return &kindError{
				message: "Invalid IRI component: right-to-left parts must start and end with right-to-left characters",
				details: component,
			}
		}

		// Check the last character of the component.
		propLast, _ := bidi.LookupRune(runes[len(runes)-1])
		classLast := propLast.Class()
		isLastRTL := classLast == bidi.R || classLast == bidi.AL
		if !isLastRTL {
			return &kindError{
				message: "Invalid IRI component: right-to-left parts must start and end with right-to-left characters",
				details: component,
			}
		}
	}

	return nil
}

// validateBidiHost checks a host string against the Bidi rules.
// RFC 3987, Section 4.2 requires that for hostnames, each dot-separated
// label be treated as an individual component for Bidi validation.
func validateBidiHost(host string) error {
	// For IP literals (e.g., [::1]), Bidi rules do not apply.
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return nil
	}
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if err := validateBidiComponent(label); err != nil {
			// Attach the full host context to the error for better diagnostics.
			var e *kindError
			if errors.As(err, &e) {
				e.message = "Invalid IRI host label"
				e.details = label + " in host '" + host + "'"
				return e
			}
		}
	}
	return nil
}
