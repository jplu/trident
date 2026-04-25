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

// IsAbsolute returns true if the IRI reference is absolute (i.e., it has a scheme).
func (r *Ref) IsAbsolute() bool {
	return r.positions.SchemeEnd != 0
}

// Scheme returns the scheme component of the IRI (e.g., "http") and a boolean
// indicating whether it was present.
func (r *Ref) Scheme() (string, bool) {
	if !r.IsAbsolute() {
		return "", false
	}
	// The scheme ends one character before the colon.
	return r.iri[:r.positions.SchemeEnd-1], true
}

// Authority returns the authority component of the IRI (e.g., "example.com:80")
// and a boolean indicating whether it was present. The leading "//" is not included.
func (r *Ref) Authority() (string, bool) {
	if r.positions.AuthorityEnd <= r.positions.SchemeEnd {
		return "", false
	}
	authorityComponent := r.iri[r.positions.SchemeEnd:r.positions.AuthorityEnd]
	return strings.TrimPrefix(authorityComponent, "//"), true
}

// Path returns the path component of the IRI. A path is always present,
// though it may be an empty string.
func (r *Ref) Path() string {
	return r.iri[r.positions.AuthorityEnd:r.positions.PathEnd]
}

// Query returns the query component of the IRI (the part after "?", without the "?")
// and a boolean indicating whether it was present.
func (r *Ref) Query() (string, bool) {
	if r.positions.PathEnd >= r.positions.QueryEnd {
		return "", false
	}
	// The query starts one character after the '?'.
	return r.iri[r.positions.PathEnd+1 : r.positions.QueryEnd], true
}

// Fragment returns the fragment component of the IRI (the part after "#", without the "#")
// and a boolean indicating whether it was present.
func (r *Ref) Fragment() (string, bool) {
	if r.positions.QueryEnd >= len(r.iri) {
		return "", false
	}
	// The fragment starts one character after the '#'.
	return r.iri[r.positions.QueryEnd+1:], true
}

// String returns the underlying string representation of the IRI reference.
// The returned string is not guaranteed to be in any specific Unicode normalization form
// unless the Ref was created with ParseNormalizedRef or processed by Normalize().
func (r *Ref) String() string {
	return r.iri
}

// Scheme returns the scheme component of the IRI. It is guaranteed to be present.
func (i *Iri) Scheme() string {
	s, _ := i.Ref.Scheme()
	return s
}
