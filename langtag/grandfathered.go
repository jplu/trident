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

// isGrandfatheredTag checks if a given lowercased language tag is a grandfathered tag.
//
// Specification Reference:
// RFC 5646 (Section 2.1 and Section 2.2.8)
//
// Parameters:
//   - tag: The lowercase language tag to evaluate.
//
// Returns:
//   - bool: True if the tag is registered as regular or irregular grandfathered, false otherwise.
func isGrandfatheredTag(tag string) bool {
	// Spec Rule: RFC 5646 (Section 2.1)
	// The regular and irregular grandfathered tags listed below match the explicit list
	// mandated by the specification to support parser well-formedness validation without
	// querying the external registry.
	switch tag {
	case "en-gb-oed",
		"i-ami",
		"i-bnn",
		"i-default",
		"i-enochian",
		"i-hak",
		"i-klingon",
		"i-lux",
		"i-mingo",
		"i-navajo",
		"i-pwn",
		"i-tao",
		"i-tay",
		"i-tsu",
		"sgn-be-fr",
		"sgn-be-nl",
		"sgn-ch-de",
		"art-lojban",
		"cel-gaulish",
		"no-bok",
		"no-nyn",
		"zh-guoyu",
		"zh-hakka",
		"zh-min",
		"zh-min-nan",
		"zh-xiang":
		return true
	default:
		return false
	}
}
