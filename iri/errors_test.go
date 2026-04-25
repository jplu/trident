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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package iri

import (
	"errors"
	"fmt"
	"testing"
)

func TestKindError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *kindError
		expected string
	}{
		{
			name:     "Message Only",
			err:      &kindError{message: "base message"},
			expected: "base message",
		},
		{
			name:     "Message with Character",
			err:      &kindError{message: "invalid character", char: '<'},
			expected: "invalid character '<'",
		},
		{
			name:     "Message with Details",
			err:      &kindError{message: "invalid sequence", details: "%2G"},
			expected: "invalid sequence '%2G'",
		},
		{
			name:     "Character takes precedence over Details",
			err:      &kindError{message: "invalid character with details", char: '>', details: "some detail"},
			expected: "invalid character with details '>'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("kindError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewParseError(t *testing.T) {
	t.Run("Nil Error", func(t *testing.T) {
		if err := newParseError(nil); err != nil {
			t.Errorf("newParseError(nil) should return nil, but got %v", err)
		}
	})
	t.Run("Simple Error", func(t *testing.T) {
		originalErr := errors.New("a simple error")
		parseErr := newParseError(originalErr)
		if parseErr == nil {
			t.Fatal("newParseError should not return nil for a non-nil error")
			return
		}
		if parseErr.Message != originalErr.Error() {
			t.Errorf("ParseError.Message = %q, want %q", parseErr.Message, originalErr.Error())
		}
		if parseErr.Err != nil {
			t.Errorf("ParseError.Err should be nil for a simple error, but got %v", parseErr.Err)
		}
	})
	t.Run("Wrapped Error", func(t *testing.T) {
		innerErr := errors.New("inner cause")
		outerErr := fmt.Errorf("outer context: %w", innerErr)
		parseErr := newParseError(outerErr)
		if parseErr == nil {
			t.Fatal("newParseError should not return nil for a non-nil error")
			return
		}
		if parseErr.Message != outerErr.Error() {
			t.Errorf("ParseError.Message = %q, want %q", parseErr.Message, outerErr.Error())
		}
		if !errors.Is(parseErr.Err, innerErr) {
			t.Errorf("ParseError.Err should be the unwrapped error, but got %v", parseErr.Err)
		}
	})
}

func TestGlobalErrors(t *testing.T) {
	t.Run("errNoScheme", func(t *testing.T) {
		expected := "No scheme found in an absolute IRI"
		if got := errNoScheme.Error(); got != expected {
			t.Errorf("errNoScheme.Error() = %q, want %q", got, expected)
		}
	})
	t.Run("errPathStartingWithSlashes", func(t *testing.T) {
		expected := "An IRI path is not allowed to start with // if there is no authority"
		if got := errPathStartingWithSlashes.Error(); got != expected {
			t.Errorf("errPathStartingWithSlashes.Error() = %q, want %q", got, expected)
		}
	})
}
