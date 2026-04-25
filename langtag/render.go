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

package langtag

import "strings"

// render reconstructs the language tag string from the parsed components.
func (cpr *canonicalParseRun) render(b *strings.Builder) {
	if cpr.language != "" {
		b.WriteString(strings.ToLower(cpr.language))
	} else if len(cpr.privateuse) > 0 {
		b.WriteByte('x')
		for _, subtag := range cpr.privateuse {
			b.WriteByte('-')
			b.WriteString(strings.ToLower(subtag))
		}
		return
	}

	for _, subtag := range cpr.extlangs {
		b.WriteByte('-')
		b.WriteString(strings.ToLower(subtag))
	}
	if cpr.script != "" {
		b.WriteByte('-')
		writeTitleCase(b, cpr.script)
	}
	if cpr.region != "" {
		b.WriteByte('-')
		b.WriteString(strings.ToUpper(cpr.region))
	}
	for _, subtag := range cpr.variants {
		b.WriteByte('-')
		b.WriteString(strings.ToLower(subtag))
	}
	for _, ext := range cpr.extensions {
		b.WriteByte('-')
		b.WriteRune(ext.Singleton)
		if ext.Value != "" {
			b.WriteByte('-')
			b.WriteString(strings.ToLower(ext.Value))
		}
	}
	if cpr.state == stateInPrivateUse && len(cpr.privateuse) > 0 {
		b.WriteByte('-')
		b.WriteByte('x')
		for _, subtag := range cpr.privateuse {
			b.WriteByte('-')
			b.WriteString(strings.ToLower(subtag))
		}
	}
}

// getPositions calculates the final end positions of each component in the
// rendered tag string.
func (cpr *canonicalParseRun) getPositions() tagElementsPositions {
	var pos tagElementsPositions
	cursor := 0
	if cpr.language != "" {
		cursor = len(cpr.language)
	}
	pos.languageEnd = cursor

	if len(cpr.extlangs) > 0 {
		for _, ext := range cpr.extlangs {
			cursor += 1 + len(ext)
		}
	}
	pos.extlangEnd = cursor

	if cpr.script != "" {
		cursor += 1 + len(cpr.script)
	}
	pos.scriptEnd = cursor

	if cpr.region != "" {
		cursor += 1 + len(cpr.region)
	}
	pos.regionEnd = cursor

	if len(cpr.variants) > 0 {
		for _, v := range cpr.variants {
			cursor += 1 + len(v)
		}
	}
	pos.variantEnd = cursor

	if len(cpr.extensions) > 0 {
		for _, ext := range cpr.extensions {
			cursor += 1 + 1 // -s
			if ext.Value != "" {
				cursor += 1 + len(ext.Value)
			}
		}
	}
	pos.extensionEnd = cursor

	return pos
}
