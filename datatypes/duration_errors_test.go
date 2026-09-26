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
	"errors"
	"testing"
)

// TestParseDurationErrorError tests the Error method of ParseDurationError.
func TestParseDurationErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      ParseDurationError
		expected string
	}{
		{
			name: "without wrapped error",
			err: ParseDurationError{
				msg: "invalid format",
			},
			expected: "parsing duration: invalid format",
		},
		{
			name: "with wrapped error",
			err: ParseDurationError{
				msg: "overflow",
				err: ErrDurationOverflow,
			},
			expected: "parsing duration: overflow: overflow during xsd:duration computation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ParseDurationError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestParseDurationErrorUnwrap tests the Unwrap method of ParseDurationError.
func TestParseDurationErrorUnwrap(t *testing.T) {
	t.Run("returns underlying wrapped error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		pErr := ParseDurationError{
			msg: "some error",
			err: underlying,
		}

		if got := pErr.Unwrap(); !errors.Is(got, underlying) {
			t.Errorf("ParseDurationError.Unwrap() = %v, want %v", got, underlying)
		}
	})

	t.Run("returns nil when no error is wrapped", func(t *testing.T) {
		pErr := ParseDurationError{
			msg: "some error",
		}

		if got := pErr.Unwrap(); got != nil {
			t.Errorf("ParseDurationError.Unwrap() = %v, want nil", got)
		}
	})

	t.Run("works with errors.Is", func(t *testing.T) {
		pErr := ParseDurationError{
			msg: "overflow occurred",
			err: ErrDurationOverflow,
		}

		if !errors.Is(pErr, ErrDurationOverflow) {
			t.Errorf("errors.Is(pErr, ErrDurationOverflow) = false, want true")
		}
	})
}

// TestNewParseDurationError tests the newParseDurationError constructor function.
func TestNewParseDurationError(t *testing.T) {
	msg := "test error message"
	pErr := newParseDurationError(msg)

	if pErr.msg != msg {
		t.Errorf("newParseDurationError().msg = %q, want %q", pErr.msg, msg)
	}

	if pErr.err != nil {
		t.Errorf("newParseDurationError().err = %v, want nil", pErr.err)
	}

	expectedStr := "parsing duration: test error message"
	if got := pErr.Error(); got != expectedStr {
		t.Errorf("newParseDurationError().Error() = %q, want %q", got, expectedStr)
	}
}

// TestErrParseOverflow tests the errParseOverflow error variable.
func TestErrParseOverflow(t *testing.T) {
	if errParseOverflow.msg != "overflow error" {
		t.Errorf("errParseOverflow.msg = %q, want %q", errParseOverflow.msg, "overflow error")
	}

	if !errors.Is(errParseOverflow, ErrDurationOverflow) {
		t.Errorf("errors.Is(errParseOverflow, ErrDurationOverflow) = false, want true")
	}

	expectedStr := "parsing duration: overflow error: overflow during xsd:duration computation"
	if got := errParseOverflow.Error(); got != expectedStr {
		t.Errorf("errParseOverflow.Error() = %q, want %q", got, expectedStr)
	}
}

// TestPredefinedErrors tests the error messages of predefined duration errors.
func TestPredefinedErrors(t *testing.T) {
	t.Run("ErrDurationOverflow message", func(t *testing.T) {
		expected := "overflow during xsd:duration computation"
		if got := ErrDurationOverflow.Error(); got != expected {
			t.Errorf("ErrDurationOverflow.Error() = %q, want %q", got, expected)
		}
	})

	t.Run("ErrOppositeSignInDurationComponents message", func(t *testing.T) {
		expected := "the xsd:yearMonthDuration and xsd:dayTimeDuration components of a xsd:duration can't have opposite sign"
		if got := ErrOppositeSignInDurationComponents.Error(); got != expected {
			t.Errorf("ErrOppositeSignInDurationComponents.Error() = %q, want %q", got, expected)
		}
	})
}
