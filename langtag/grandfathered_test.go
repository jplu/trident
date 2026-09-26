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
package langtag

import "testing"

// TestIsGrandfatheredTag tests the validation of grandfathered language tags.
func TestIsGrandfatheredTag(t *testing.T) {
	grandfatheredTags := []string{
		"en-gb-oed",
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
		"zh-xiang",
	}

	t.Run("registered grandfathered tags return true", func(t *testing.T) {
		for _, tag := range grandfatheredTags {
			t.Run(tag, func(t *testing.T) {
				if got := isGrandfatheredTag(tag); !got {
					t.Errorf("isGrandfatheredTag(%q) = false, expected true", tag)
				}
			})
		}
	})

	t.Run("non-grandfathered tags return false", func(t *testing.T) {
		tests := []string{
			"en",
			"en-us",
			"fr-fr",
			"de-de",
			"ja-jp",
			"zh-hans",
			"zh-hant-hk",
			"sr-cyrl",
			"sr-latn",

			"",
			" ",
			"-",

			"en-gb",
			"i",
			"sgn",
			"sgn-be",
			"sgn-ch",
			"art",
			"cel",
			"no",
			"zh",

			"en-gb-oed-variant",
			"i-ami-test",
			"i-default-x-extra",
			"art-lojban-extra",
			"zh-min-nan-tw",

			"oed",
			"default",
			"klingon",
			"lojban",
			"gaulish",
			"min-nan",

			"EN-GB-OED",
			"I-Ami",
			"I-DEFAULT",
			"SGN-BE-FR",
			"Art-Lojban",
			"Zh-Min-Nan",

			"foo-bar",
			"invalid-tag",
			"x-private",
		}

		for _, tag := range tests {
			t.Run(tag, func(t *testing.T) {
				if got := isGrandfatheredTag(tag); got {
					t.Errorf("isGrandfatheredTag(%q) = true, expected false", tag)
				}
			})
		}
	})
}
