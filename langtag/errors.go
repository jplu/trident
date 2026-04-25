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
	ErrEmptyExtension     = errors.New("if an extension subtag is present, it must not be empty")
	ErrEmptyPrivateUse    = errors.New("if the 'x' subtag is present, it must not be empty")
	ErrForbiddenChar      = errors.New("the langtag contains a char not allowed")
	ErrInvalidSubtag      = errors.New("a subtag fails to parse or is not a valid IANA subtag")
	ErrInvalidLanguage    = errors.New("the given language subtag is invalid")
	ErrSubtagTooLong      = errors.New("a subtag may be eight characters in length at maximum")
	ErrEmptySubtag        = errors.New("a subtag should not be empty")
	ErrTooManyExtlangs    = errors.New("at maximum one extlang is allowed")
	ErrDuplicateVariant   = errors.New("the same variant subtag appears more than once")
	ErrDuplicateSingleton = errors.New("the same extension singleton appears more than once")
)

// typeExtlang is the IANA registry type string for extended language subtags.
const typeExtlang = "extlang"
