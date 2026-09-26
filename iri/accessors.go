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

import "strings"

// IsAbsolute checks if the IRI reference is absolute (i.e., it has a scheme).
//
// Specification Reference:
// RFC 3987 (Section 2.2) and RFC 3986 (Section 4.3)
//
// Parameters:
//
// Returns:
//   - bool: True if the IRI reference contains a scheme component, and false otherwise.
func (r *Ref) IsAbsolute() bool {
	return r.positions.SchemeEnd != 0
}

// Scheme extracts the scheme component of the IRI and a boolean indicating whether it was present.
//
// Specification Reference:
// RFC 3987 (Section 2.2) and RFC 3986 (Section 3.1)
//
// Parameters:
//
// Returns:
//   - string: The extracted scheme component of the IRI reference (e.g., "http"), or an empty string if not present.
//   - bool: True if the scheme component is present in the IRI reference, and false otherwise.
func (r *Ref) Scheme() (string, bool) {
	if !r.IsAbsolute() {
		return "", false
	}
	// Implementation Note: Slice extraction of the scheme component
	// The scheme ends one character before the colon separator in the parsed IRI reference.
	return r.iri[:r.positions.SchemeEnd-1], true
}

// Authority extracts the authority component of the IRI and a boolean indicating whether it was present.
//
// Specification Reference:
// RFC 3987 (Section 2.2) and RFC 3986 (Section 3.2)
//
// Parameters:
//
// Returns:
// - string: The authority component of the IRI reference (e.g., "example.com:80") excluding the leading "//" prefix, or
// an empty string if not present.
//   - bool: True if the authority component is present in the IRI reference, and false otherwise.
func (r *Ref) Authority() (string, bool) {
	if r.positions.AuthorityEnd <= r.positions.SchemeEnd {
		return "", false
	}
	authorityComponent := r.iri[r.positions.SchemeEnd:r.positions.AuthorityEnd]
	return strings.TrimPrefix(authorityComponent, "//"), true
}

// Path extracts the path component of the IRI.
//
// Specification Reference:
// RFC 3987 (Section 2.2) and RFC 3986 (Section 3.3)
//
// Parameters:
//
// Returns:
//   - string: The path component of the IRI reference, which is always present but may be an empty string.
func (r *Ref) Path() string {
	return r.iri[r.positions.AuthorityEnd:r.positions.PathEnd]
}

// Query extracts the query component of the IRI (the part after "?", without the "?") and a boolean indicating whether
// it was present.
//
// Specification Reference:
// RFC 3987 (Section 2.2) and RFC 3986 (Section 3.4)
//
// Parameters:
//
// Returns:
// - string: The query component of the IRI reference (excluding the "?" delimiter), or an empty string if not present.
//   - bool: True if the query component is present, and false otherwise.
func (r *Ref) Query() (string, bool) {
	if r.positions.PathEnd >= r.positions.QueryEnd {
		return "", false
	}
	// Implementation Note: Slice extraction of the query component
	// The query starts one character after the "?" delimiter to exclude the delimiter itself from the returned value.
	return r.iri[r.positions.PathEnd+1 : r.positions.QueryEnd], true
}

// Fragment extracts the fragment component of the IRI (the part after "#", without the "#") and a boolean indicating
// whether it was present.
//
// Specification Reference:
// RFC 3987 (Section 2.2) and RFC 3986 (Section 3.5)
//
// Parameters:
//
// Returns:
// - string: The fragment component of the IRI reference (excluding the "#" delimiter), or an empty string if not
// present.
//   - bool: True if the fragment component is present, and false otherwise.
func (r *Ref) Fragment() (string, bool) {
	if r.positions.QueryEnd >= len(r.iri) {
		return "", false
	}
	// Implementation Note: Slice extraction of the fragment component
	// The fragment starts one character after the "#" delimiter to exclude the delimiter itself from the returned
	// value.
	return r.iri[r.positions.QueryEnd+1:], true
}

// String returns the underlying string representation of the IRI reference.
//
// Specification Reference:
// RFC 3987 (Section 3.1)
//
// Parameters:
//
// Returns:
// - string: The exact underlying string representation of the IRI reference, which is not guaranteed to be in any
// specific Unicode normalization form unless normalized during parsing or by calling Normalize().
func (r *Ref) String() string {
	return r.iri
}

// Scheme extracts the scheme component of the absolute IRI.
//
// Specification Reference:
// RFC 3987 (Section 2.2) and RFC 3986 (Section 3.1)
//
// Parameters:
//
// Returns:
//   - string: The guaranteed present scheme component of the absolute IRI.
func (i *Iri) Scheme() string {
	s, _ := i.Ref.Scheme()
	return s
}

// ValidateBidiPresentation validates the IRI's components against the bidirectional presentation guidelines.
//
// Specification Reference:
// RFC 3987 (Section 4.2)
//
// Parameters:
//
// Returns:
//   - error: An error if any component violates the bidirectional rendering guidelines.
func (r *Ref) ValidateBidiPresentation() error {
	if r.iri == "" {
		return nil
	}

	// Step 1: Validate Authority components
	// Verify that the userinfo and host components conform to bidirectional presentation guidelines.
	if authority, hasAuthority := r.Authority(); hasAuthority {
		if err := r.validateBidiAuthority(authority); err != nil {
			return err
		}
	}

	// Step 2: Validate Path segment components
	// Verify that each individual slash-separated path segment conforms to bidirectional presentation guidelines.
	if err := validateBidiPath(r.Path()); err != nil {
		return err
	}

	// Step 3: Validate Query component
	// Verify that the query string conforms to bidirectional presentation guidelines.
	if query, hasQuery := r.Query(); hasQuery {
		if err := validateBidiComponent(query); err != nil {
			return err
		}
	}

	// Step 4: Validate Fragment component
	// Verify that the fragment string conforms to bidirectional presentation guidelines.
	if fragment, hasFragment := r.Fragment(); hasFragment {
		if err := validateBidiComponent(fragment); err != nil {
			return err
		}
	}

	return nil
}

// validateBidiAuthority validates the userinfo and host parts of the authority against bidirectional rules.
//
// Specification Reference:
// RFC 3987 (Section 4.2)
//
// Parameters:
//   - authority: The raw authority string to validate.
//
// Returns:
//   - error: An error if either the userinfo or host component violates bidirectional presentation guidelines.
func (r *Ref) validateBidiAuthority(authority string) error {
	userinfo, host, _, _ := splitAuthority(authority)
	if userinfo != "" {
		if err := validateBidiComponent(userinfo); err != nil {
			return err
		}
	}
	if host != "" {
		if err := validateBidiHost(host); err != nil {
			return err
		}
	}
	return nil
}

// validateBidiPath validates the segments of the path against bidirectional rules.
//
// Specification Reference:
// RFC 3987 (Section 4.2)
//
// Parameters:
//   - path: The path component of the IRI.
//
// Returns:
//   - error: An error if any segment of the path violates bidirectional presentation guidelines.
func validateBidiPath(path string) error {
	if path == "" {
		return nil
	}
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if err := validateBidiComponent(segment); err != nil {
			return err
		}
	}
	return nil
}
