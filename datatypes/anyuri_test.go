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

//nolint:testpackage // White-box test in the same package to access unexported functions.
package datatypes

import (
	"testing"
)

// TestParseAnyURI tests the parsing of various URI strings into AnyURI values.
func TestParseAnyURI(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    AnyURI
		wantErr bool
	}{
		{
			name:    "valid absolute HTTP URI",
			input:   "http://example.com/path",
			want:    AnyURI("http://example.com/path"),
			wantErr: false,
		},
		{
			name:    "valid URI with leading and trailing whitespace",
			input:   "   http://example.com/path \n\t ",
			want:    AnyURI("http://example.com/path"),
			wantErr: false,
		},
		{
			name:    "valid URN",
			input:   "urn:isbn:0451450523",
			want:    AnyURI("urn:isbn:0451450523"),
			wantErr: false,
		},
		{
			name:    "valid relative reference",
			input:   "relative/path#fragment",
			want:    AnyURI("relative/path#fragment"),
			wantErr: false,
		},
		{
			name:    "empty URI reference",
			input:   "",
			want:    AnyURI(""),
			wantErr: false,
		},
		{
			name:    "invalid percent encoding",
			input:   "http://example.com/%ZZ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid characters in URI",
			input:   "http://[invalid-ipv6",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAnyURI(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseAnyURI(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseAnyURI(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestAnyURIString tests the String method of AnyURI.
func TestAnyURIString(t *testing.T) {
	uri := AnyURI("http://example.org/resource")
	want := "http://example.org/resource"
	if got := uri.String(); got != want {
		t.Errorf("AnyURI.String() = %q, want %q", got, want)
	}
}

// TestAnyURIIsIdenticalWith tests the IsIdenticalWith method for comparing AnyURI values.
func TestAnyURIIsIdenticalWith(t *testing.T) {
	uri1 := AnyURI("http://example.org/a")
	uri2 := AnyURI("http://example.org/a")
	uri3 := AnyURI("http://example.org/b")
	differentType := String("http://example.org/a")

	tests := []struct {
		name  string
		a     AnyURI
		other XSDValue
		want  bool
	}{
		{
			name:  "identical AnyURI values",
			a:     uri1,
			other: uri2,
			want:  true,
		},
		{
			name:  "different AnyURI values",
			a:     uri1,
			other: uri3,
			want:  false,
		},
		{
			name:  "different XSDValue type",
			a:     uri1,
			other: differentType,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsIdenticalWith(tt.other); got != tt.want {
				t.Errorf("AnyURI.IsIdenticalWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAnyURILength tests the Length method for counting runes in AnyURI values.
func TestAnyURILength(t *testing.T) {
	tests := []struct {
		name string
		uri  AnyURI
		want int
	}{
		{
			name: "ASCII characters",
			uri:  AnyURI("http://example.org"),
			want: 18,
		},
		{
			name: "Unicode characters (18 runes, 19 bytes)",
			uri:  AnyURI("http://exämpel.org"),
			want: 18,
		},
		{
			name: "empty URI",
			uri:  AnyURI(""),
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.uri.Length(); got != tt.want {
				t.Errorf("AnyURI.Length() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestAnyURIToRef tests conversion of AnyURI values to reference objects.
func TestAnyURIToRef(t *testing.T) {
	t.Run("valid AnyURI to Ref", func(t *testing.T) {
		uri := AnyURI("http://example.org/test#fragment")
		ref := uri.ToRef()
		if ref == nil {
			t.Fatalf("AnyURI.ToRef() returned nil for valid URI")
		}
		if got := ref.String(); got != string(uri) {
			t.Errorf("AnyURI.ToRef().String() = %q, want %q", got, uri)
		}
	})

	t.Run("invalid AnyURI to Ref returns nil", func(t *testing.T) {
		uri := AnyURI("http://example.org/%ZZ")
		ref := uri.ToRef()
		if ref != nil {
			t.Errorf("AnyURI.ToRef() expected nil for invalid URI, got %v", ref)
		}
	})
}
