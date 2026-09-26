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
package iri

import (
	"errors"
	"testing"
)

// TestRefIsAbsolute tests the IsAbsolute method of Ref.
func TestRefIsAbsolute(t *testing.T) {
	tests := []struct {
		name string
		iri  string
		want bool
	}{
		{
			name: "Absolute HTTP URI",
			iri:  "http://example.com",
			want: true,
		},
		{
			name: "Absolute HTTPS URI with path",
			iri:  "https://example.com/foo/bar",
			want: true,
		},
		{
			name: "Absolute URN",
			iri:  "urn:isbn:0451450523",
			want: true,
		},
		{
			name: "Absolute custom scheme",
			iri:  "custom-scheme://test",
			want: true,
		},
		{
			name: "Relative path starting with slash",
			iri:  "/path/to/resource",
			want: false,
		},
		{
			name: "Relative path without slash",
			iri:  "path/to/resource",
			want: false,
		},
		{
			name: "Network-path reference",
			iri:  "//example.com/foo",
			want: false,
		},
		{
			name: "Query-only reference",
			iri:  "?query=1",
			want: false,
		},
		{
			name: "Fragment-only reference",
			iri:  "#fragment",
			want: false,
		},
		{
			name: "Empty reference",
			iri:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref *Ref
			if tt.iri == "" {
				ref = &Ref{}
			} else {
				var err error
				ref, err = ParseRef(tt.iri)
				if err != nil {
					t.Fatalf("ParseRef() unexpected error: %v", err)
				}
			}

			if got := ref.IsAbsolute(); got != tt.want {
				t.Errorf("Ref.IsAbsolute() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRefScheme tests the Scheme method of Ref.
func TestRefScheme(t *testing.T) {
	tests := []struct {
		name       string
		iri        string
		wantScheme string
		wantOk     bool
	}{
		{
			name:       "HTTP scheme",
			iri:        "http://example.com",
			wantScheme: "http",
			wantOk:     true,
		},
		{
			name:       "HTTPS scheme",
			iri:        "https://example.com/path",
			wantScheme: "https",
			wantOk:     true,
		},
		{
			name:       "URN scheme",
			iri:        "urn:isbn:0451450523",
			wantScheme: "urn",
			wantOk:     true,
		},
		{
			name:       "Complex valid scheme with +, -, and .",
			iri:        "my-scheme.v1+test:foo",
			wantScheme: "my-scheme.v1+test",
			wantOk:     true,
		},
		{
			name:       "Relative path",
			iri:        "/path/to/resource",
			wantScheme: "",
			wantOk:     false,
		},
		{
			name:       "Network-path reference",
			iri:        "//example.com/foo",
			wantScheme: "",
			wantOk:     false,
		},
		{
			name:       "Empty reference",
			iri:        "",
			wantScheme: "",
			wantOk:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref *Ref
			if tt.iri == "" {
				ref = &Ref{}
			} else {
				var err error
				ref, err = ParseRef(tt.iri)
				if err != nil {
					t.Fatalf("ParseRef() unexpected error: %v", err)
				}
			}

			scheme, ok := ref.Scheme()
			if ok != tt.wantOk {
				t.Errorf("Ref.Scheme() ok = %v, wantOk %v", ok, tt.wantOk)
			}
			if scheme != tt.wantScheme {
				t.Errorf("Ref.Scheme() scheme = %q, wantScheme %q", scheme, tt.wantScheme)
			}
		})
	}
}

// TestIriScheme tests the Scheme method of Iri.
func TestIriScheme(t *testing.T) {
	tests := []struct {
		name       string
		iri        string
		wantScheme string
	}{
		{
			name:       "HTTP absolute IRI",
			iri:        "http://example.com",
			wantScheme: "http",
		},
		{
			name:       "HTTPS absolute IRI with query and fragment",
			iri:        "https://example.com/path?q=1#f",
			wantScheme: "https",
		},
		{
			name:       "FTP absolute IRI",
			iri:        "ftp://ftp.example.com",
			wantScheme: "ftp",
		},
		{
			name:       "URN absolute IRI",
			iri:        "urn:example:animal",
			wantScheme: "urn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absIri, err := ParseIri(tt.iri)
			if err != nil {
				t.Fatalf("ParseIri() unexpected error: %v", err)
			}

			if got := absIri.Scheme(); got != tt.wantScheme {
				t.Errorf("Iri.Scheme() = %q, want %q", got, tt.wantScheme)
			}
		})
	}
}

// TestRefString tests the String method of Ref.
func TestRefString(t *testing.T) {
	tests := []struct {
		name string
		iri  string
		want string
	}{
		{
			name: "Full absolute IRI",
			iri:  "http://user:pass@example.com:8080/path?query#frag",
			want: "http://user:pass@example.com:8080/path?query#frag",
		},
		{
			name: "Relative path reference",
			iri:  "/path/to/resource",
			want: "/path/to/resource",
		},
		{
			name: "Empty reference",
			iri:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref *Ref
			if tt.iri == "" {
				ref = &Ref{}
			} else {
				var err error
				ref, err = ParseRef(tt.iri)
				if err != nil {
					t.Fatalf("ParseRef() unexpected error: %v", err)
				}
			}

			if got := ref.String(); got != tt.want {
				t.Errorf("Ref.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRefAuthority tests the Authority method of Ref.
func TestRefAuthority(t *testing.T) {
	tests := []struct {
		name          string
		iri           string
		wantAuthority string
		wantOk        bool
	}{
		{
			name:          "Authority with userinfo, host, and port",
			iri:           "http://user:pass@example.com:8080/path",
			wantAuthority: "user:pass@example.com:8080",
			wantOk:        true,
		},
		{
			name:          "Authority with host only",
			iri:           "http://example.com/path",
			wantAuthority: "example.com",
			wantOk:        true,
		},
		{
			name:          "Network-path relative reference",
			iri:           "//example.com:9000/path",
			wantAuthority: "example.com:9000",
			wantOk:        true,
		},
		{
			name:          "No authority in URN",
			iri:           "urn:isbn:0451450523",
			wantAuthority: "",
			wantOk:        false,
		},
		{
			name:          "No authority in relative path",
			iri:           "/path/only",
			wantAuthority: "",
			wantOk:        false,
		},
		{
			name:          "Empty reference",
			iri:           "",
			wantAuthority: "",
			wantOk:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref *Ref
			if tt.iri == "" {
				ref = &Ref{}
			} else {
				var err error
				ref, err = ParseRef(tt.iri)
				if err != nil {
					t.Fatalf("ParseRef() unexpected error: %v", err)
				}
			}

			auth, ok := ref.Authority()
			if ok != tt.wantOk {
				t.Errorf("Ref.Authority() ok = %v, wantOk %v", ok, tt.wantOk)
			}
			if auth != tt.wantAuthority {
				t.Errorf("Ref.Authority() authority = %q, want %q", auth, tt.wantAuthority)
			}
		})
	}
}

// TestRefPath tests the Path method of Ref.
func TestRefPath(t *testing.T) {
	tests := []struct {
		name     string
		iri      string
		wantPath string
	}{
		{
			name:     "Absolute URI with multi-segment path",
			iri:      "http://example.com/path/to/resource",
			wantPath: "/path/to/resource",
		},
		{
			name:     "Absolute URI with empty path",
			iri:      "http://example.com",
			wantPath: "",
		},
		{
			name:     "Absolute URI with root slash path",
			iri:      "http://example.com/",
			wantPath: "/",
		},
		{
			name:     "Relative path with query and fragment",
			iri:      "/path/to/res?q=1#frag",
			wantPath: "/path/to/res",
		},
		{
			name:     "Relative path without leading slash",
			iri:      "relative/path",
			wantPath: "relative/path",
		},
		{
			name:     "Empty reference",
			iri:      "",
			wantPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref *Ref
			if tt.iri == "" {
				ref = &Ref{}
			} else {
				var err error
				ref, err = ParseRef(tt.iri)
				if err != nil {
					t.Fatalf("ParseRef() unexpected error: %v", err)
				}
			}

			if got := ref.Path(); got != tt.wantPath {
				t.Errorf("Ref.Path() = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

// TestRefQuery tests the Query method of Ref.
func TestRefQuery(t *testing.T) {
	tests := []struct {
		name      string
		iri       string
		wantQuery string
		wantOk    bool
	}{
		{
			name:      "With query parameters",
			iri:       "http://example.com/path?foo=bar&baz=qux",
			wantQuery: "foo=bar&baz=qux",
			wantOk:    true,
		},
		{
			name:      "With empty query string",
			iri:       "http://example.com/path?",
			wantQuery: "",
			wantOk:    true,
		},
		{
			name:      "Without query",
			iri:       "http://example.com/path",
			wantQuery: "",
			wantOk:    false,
		},
		{
			name:      "With query and fragment",
			iri:       "http://example.com/path?q=1#section",
			wantQuery: "q=1",
			wantOk:    true,
		},
		{
			name:      "Empty reference",
			iri:       "",
			wantQuery: "",
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref *Ref
			if tt.iri == "" {
				ref = &Ref{}
			} else {
				var err error
				ref, err = ParseRef(tt.iri)
				if err != nil {
					t.Fatalf("ParseRef() unexpected error: %v", err)
				}
			}

			query, ok := ref.Query()
			if ok != tt.wantOk {
				t.Errorf("Ref.Query() ok = %v, wantOk %v", ok, tt.wantOk)
			}
			if query != tt.wantQuery {
				t.Errorf("Ref.Query() query = %q, wantQuery %q", query, tt.wantQuery)
			}
		})
	}
}

// TestRefFragment tests the Fragment method of Ref.
func TestRefFragment(t *testing.T) {
	tests := []struct {
		name         string
		iri          string
		wantFragment string
		wantOk       bool
	}{
		{
			name:         "With fragment",
			iri:          "http://example.com/path#section1",
			wantFragment: "section1",
			wantOk:       true,
		},
		{
			name:         "With empty fragment",
			iri:          "http://example.com/path#",
			wantFragment: "",
			wantOk:       true,
		},
		{
			name:         "Without fragment",
			iri:          "http://example.com/path",
			wantFragment: "",
			wantOk:       false,
		},
		{
			name:         "With query and fragment",
			iri:          "http://example.com/path?query=val#section2",
			wantFragment: "section2",
			wantOk:       true,
		},
		{
			name:         "Empty reference",
			iri:          "",
			wantFragment: "",
			wantOk:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ref *Ref
			if tt.iri == "" {
				ref = &Ref{}
			} else {
				var err error
				ref, err = ParseRef(tt.iri)
				if err != nil {
					t.Fatalf("ParseRef() unexpected error: %v", err)
				}
			}

			fragment, ok := ref.Fragment()
			if ok != tt.wantOk {
				t.Errorf("Ref.Fragment() ok = %v, wantOk %v", ok, tt.wantOk)
			}
			if fragment != tt.wantFragment {
				t.Errorf("Ref.Fragment() fragment = %q, wantFragment %q", fragment, tt.wantFragment)
			}
		})
	}
}

// TestRefValidateBidiPresentation tests the ValidateBidiPresentation method of Ref.
func TestRefValidateBidiPresentation(t *testing.T) {
	tests := []struct {
		name    string
		iri     string
		wantErr bool
	}{
		{
			name:    "Empty IRI",
			iri:     "",
			wantErr: false,
		},
		{
			name:    "Normal LTR IRI",
			iri:     "http://user:pass@example.com/path/to/res?query=1#frag",
			wantErr: false,
		},
		{
			name:    "No authority, LTR",
			iri:     "mailto:test@example.com",
			wantErr: false,
		},
		{
			name:    "Valid RTL Hebrew",
			iri:     "http://\u05D0\u05D1@\u05D2\u05D3.\u05D4\u05D5/\u05D6\u05D7/\u05D8\u05D9?\u05DA\u05DB#\u05DC\u05DD",
			wantErr: false,
		},
		{
			name:    "Invalid Userinfo Rule 1 (Mixed)",
			iri:     "http://a\u05D0@example.com/path",
			wantErr: true,
		},
		{
			name:    "Invalid Userinfo Rule 2 (Ends with non-RTL)",
			iri:     "http://\u05D01@example.com/path",
			wantErr: true,
		},
		{
			name:    "Invalid Host Rule 1 (Mixed)",
			iri:     "http://example.com\u05D0/path",
			wantErr: true,
		},
		{
			name:    "Invalid Host Rule 2 (Starts with non-RTL)",
			iri:     "http://1\u05D0.com/path",
			wantErr: true,
		},
		{
			name:    "IP literal host bypassed",
			iri:     "http://[2001:db8::1]/path",
			wantErr: false,
		},
		{
			name:    "Invalid Path Segment Rule 1 (Mixed)",
			iri:     "http://example.com/path/a\u05D0",
			wantErr: true,
		},
		{
			name:    "Invalid Path Segment Rule 2 (Ends with non-RTL)",
			iri:     "http://example.com/path/\u05D01",
			wantErr: true,
		},
		{
			name:    "Empty Path bypasses path check",
			iri:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "Invalid Query Rule 1 (Mixed)",
			iri:     "http://example.com/path?a\u05D0",
			wantErr: true,
		},
		{
			name:    "Invalid Query Rule 2 (Starts with non-RTL)",
			iri:     "http://example.com/path?1\u05D0",
			wantErr: true,
		},
		{
			name:    "Invalid Fragment Rule 1 (Mixed)",
			iri:     "http://example.com/path?q#a\u05D0",
			wantErr: true,
		},
		{
			name:    "Invalid Fragment Rule 2 (Ends with non-RTL)",
			iri:     "http://example.com/path?q#\u05D01",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runBidiPresentation(t, tt.iri, tt.wantErr)
		})
	}
}

// runBidiPresentation is a helper function to test bidi presentation validation.
func runBidiPresentation(t *testing.T, iri string, wantErr bool) {
	t.Helper()

	var ref *Ref
	if iri == "" {
		ref = &Ref{}
	} else {
		var err error
		ref, err = ParseRef(iri)
		if err != nil {
			t.Fatalf("ParseRef() unexpected error: %v", err)
		}
	}

	err := ref.ValidateBidiPresentation()
	if (err != nil) != wantErr {
		t.Errorf("ValidateBidiPresentation() error = %v, wantErr %v", err, wantErr)
	}

	if wantErr && err != nil {
		var bidiErr *BidiGuidelineError
		if !errors.As(err, &bidiErr) {
			t.Errorf("Expected BidiGuidelineError, got %T: %v", err, err)
		} else if bidiErr.Error() == "" {
			t.Error("BidiGuidelineError.Error() returned empty string")
		}
	}
}

// TestRefValidateBidiAuthoritySpecifics tests specific cases for validateBidiAuthority.
func TestRefValidateBidiAuthoritySpecifics(t *testing.T) {
	ref := &Ref{}
	err := ref.validateBidiAuthority("")
	if err != nil {
		t.Errorf("validateBidiAuthority(\"\") expected nil, got %v", err)
	}

	refUserOnly, err := ParseRef("http://user@")
	if err != nil {
		t.Fatalf("ParseRef failed: %v", err)
	}
	auth, ok := refUserOnly.Authority()
	if !ok {
		t.Fatal("Expected authority to be present")
	}
	err = refUserOnly.validateBidiAuthority(auth)
	if err != nil {
		t.Errorf("validateBidiAuthority failed on user-only: %v", err)
	}

	refMixed, err := ParseRef("http://user@a\u05D0")
	if err != nil {
		t.Fatalf("ParseRef failed: %v", err)
	}
	authMixed, _ := refMixed.Authority()
	err = refMixed.validateBidiAuthority(authMixed)
	if err == nil {
		t.Error("Expected error for mixed host under userinfo, got nil")
	}
}

// TestValidateBidiPathSpecifics tests specific cases for validateBidiPath.
func TestValidateBidiPathSpecifics(t *testing.T) {
	err := validateBidiPath("")
	if err != nil {
		t.Errorf("validateBidiPath(\"\") expected nil, got %v", err)
	}

	err = validateBidiPath("a/b/c")
	if err != nil {
		t.Errorf("validateBidiPath(\"a/b/c\") expected nil, got %v", err)
	}

	err = validateBidiPath("a/a\u05D0")
	if err == nil {
		t.Error("validateBidiPath(\"a/a\u05D0\") expected error, got nil")
	}
}
