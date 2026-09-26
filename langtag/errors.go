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

import "errors"

// Errors that can occur during language tag parsing.
var (
	// ErrEmptyExtension represents the error when an extension singleton is not followed by any extension subtags.
	//
	// Specification Reference:
	// RFC 5646 (Section 2.2.6, Rule 6).
	ErrEmptyExtension = errors.New("if an extension subtag is present, it must not be empty")

	// ErrEmptyPrivateUse represents the error when a private-use singleton 'x' is not followed by any private-use
	// subtags.
	//
	// Specification Reference:
	// RFC 5646 (Section 2.2.7, Rule 3).
	ErrEmptyPrivateUse = errors.New("if the 'x' subtag is present, it must not be empty")

	// ErrForbiddenChar represents the error when a language tag contains characters outside the allowed US-ASCII
	// alphanumeric and hyphen set.
	//
	// Specification Reference:
	// RFC 5646 (Section 2.1).
	ErrForbiddenChar = errors.New("the langtag contains a char not allowed")

	// ErrInvalidSubtag represents the error when a subtag fails to parse or is not present in the IANA Language Subtag
	// Registry.
	//
	// Specification Reference:
	// RFC 5646 (Section 2.2.9).
	ErrInvalidSubtag = errors.New("a subtag fails to parse or is not a valid IANA subtag")

	// ErrInvalidLanguage represents the error when the primary language subtag is syntactically invalid or not found in
	// the registry.
	//
	// Specification Reference:
	// RFC 5646 (Section 2.2.1).
	ErrInvalidLanguage = errors.New("the given language subtag is invalid")

	// ErrSubtagTooLong represents the error when any individual subtag exceeds the maximum permitted length of eight
	// characters.
	//
	// Specification Reference:
	// RFC 5646 (Section 2.1).
	ErrSubtagTooLong = errors.New("a subtag may be eight characters in length at maximum")

	// ErrEmptySubtag represents the error when an empty subtag is encountered, typically caused by consecutive hyphens.
	//
	// Specification Reference:
	// RFC 5646 (Section 2.1).
	ErrEmptySubtag = errors.New("a subtag should not be empty")

	// ErrTooManyExtlangs represents the error when more than the allowed number of extended language subtags are
	// present.
	//
	// Specification Reference:
	// RFC 5646 (Section 2.2.2, Rule 4).
	ErrTooManyExtlangs = errors.New("at maximum one extlang is allowed")

	// ErrDuplicateVariant represents the error when the same variant subtag is repeated within a language tag.
	//
	// Specification Reference:
	// RFC 5646 (Section 2.2.5, Rule 5).
	ErrDuplicateVariant = errors.New("the same variant subtag appears more than once")

	// ErrDuplicateSingleton represents the error when the same extension singleton character is repeated within a
	// language tag.
	//
	// Specification Reference:
	// RFC 5646 (Section 2.2.6, Rule 3).
	ErrDuplicateSingleton = errors.New("the same extension singleton appears more than once")
)

// typeExtlang represents the extended language registry type identifier.
//
// Specification Reference:
// RFC 5646 (Section 3.1.3)
//
// Representation:
// The literal string "extlang" used to identify extended language records in the IANA registry.
const typeExtlang = "extlang"
