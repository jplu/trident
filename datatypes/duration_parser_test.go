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
	"math"
	"strings"
	"testing"
)

// assertDurationPartsResult is a helper to verify the results of parsed duration parts.
func assertDurationPartsResult(
	t *testing.T,
	parts durationPartsInternal,
	wantYM *int64,
	wantDTHasVal bool,
	wantDTSecSign int,
) {
	t.Helper()
	switch {
	case wantYM == nil && parts.yearMonth != nil:
		t.Errorf("expected yearMonth to be nil, got %d", *parts.yearMonth)
	case wantYM != nil && parts.yearMonth == nil:
		t.Fatalf("expected yearMonth %d, got nil", *wantYM)
	case wantYM != nil && *parts.yearMonth != *wantYM:
		t.Errorf("yearMonth = %d, want %d", *parts.yearMonth, *wantYM)
	}

	switch {
	case wantDTHasVal && parts.dayTime == nil:
		t.Fatalf("expected dayTime to be non-nil")
	case wantDTHasVal:
		if sign := parts.dayTime.getValue().Sign(); sign != wantDTSecSign {
			t.Errorf("dayTime sign = %d, want %d", sign, wantDTSecSign)
		}
	case parts.dayTime != nil:
		t.Errorf("expected dayTime to be nil, got %v", parts.dayTime)
	}
}

// TestDurationPartsValid tests parsing valid duration strings into their components.
func TestDurationPartsValid(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantYM        *int64
		wantDTHasVal  bool
		wantDTSecSign int // -1, 0, or 1
	}{
		{
			name:          "full positive duration",
			input:         "P1Y2M3DT4H5M6S",
			wantYM:        int64Ptr(14),
			wantDTHasVal:  true,
			wantDTSecSign: 1,
		},
		{
			name:          "full negative duration",
			input:         "-P1Y2M3DT4H5M6S",
			wantYM:        int64Ptr(-14),
			wantDTHasVal:  true,
			wantDTSecSign: -1,
		},
		{
			name:         "years only",
			input:        "P5Y",
			wantYM:       int64Ptr(60),
			wantDTHasVal: false,
		},
		{
			name:         "months only",
			input:        "P7M",
			wantYM:       int64Ptr(7),
			wantDTHasVal: false,
		},
		{
			name:          "days only",
			input:         "P10D",
			wantYM:        nil,
			wantDTHasVal:  true,
			wantDTSecSign: 1,
		},
		{
			name:          "hours only",
			input:         "PT2H",
			wantYM:        nil,
			wantDTHasVal:  true,
			wantDTSecSign: 1,
		},
		{
			name:          "time minutes only",
			input:         "PT15M",
			wantYM:        nil,
			wantDTHasVal:  true,
			wantDTSecSign: 1,
		},
		{
			name:          "seconds only",
			input:         "PT30S",
			wantYM:        nil,
			wantDTHasVal:  true,
			wantDTSecSign: 1,
		},
		{
			name:          "fractional seconds",
			input:         "PT0.5S",
			wantYM:        nil,
			wantDTHasVal:  true,
			wantDTSecSign: 1,
		},
		{
			name:          "negative fractional seconds",
			input:         "-PT0.5S",
			wantYM:        nil,
			wantDTHasVal:  true,
			wantDTSecSign: -1,
		},
		{
			name:          "zero duration",
			input:         "PT0S",
			wantYM:        nil,
			wantDTHasVal:  true,
			wantDTSecSign: 0,
		},
		{
			name:         "empty duration parts P",
			input:        "P",
			wantYM:       nil,
			wantDTHasVal: false,
		},
		{
			name:         "empty duration parts PT",
			input:        "PT",
			wantYM:       nil,
			wantDTHasVal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, rem, err := durationParts(tt.input)
			if err != nil {
				t.Fatalf("unexpected error parsing %q: %v", tt.input, err)
			}
			if rem != "" {
				t.Errorf("expected empty remaining string, got %q", rem)
			}
			assertDurationPartsResult(t, parts, tt.wantYM, tt.wantDTHasVal, tt.wantDTSecSign)
		})
	}
}

// TestDurationPartsErrors tests error handling when parsing invalid duration strings.
func TestDurationPartsErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "missing P prefix",
			input:   "1Y2M",
			wantErr: "durations must start with 'P'",
		},
		{
			name:    "missing P prefix with minus",
			input:   "-1Y2M",
			wantErr: "durations must start with 'P'",
		},
		{
			name:    "duplicate T separator",
			input:   "PT1HT2M",
			wantErr: "duplicated time separator 'T'",
		},
		{
			name:    "duplicate T separator after year",
			input:   "P1YT1HT2M",
			wantErr: "duplicated time separator 'T'",
		},
		{
			name:    "component without number",
			input:   "PY",
			wantErr: "invalid duration component",
		},
		{
			name:    "number without unit",
			input:   "P1",
			wantErr: "invalid duration component",
		},
		{
			name:    "unexpected unit character",
			input:   "P1X",
			wantErr: "unexpected type character",
		},
		{
			name:    "unexpected unit character after T",
			input:   "PT1Z",
			wantErr: "unexpected type character",
		},
		{
			name:    "year out of order (after month)",
			input:   "P1M1Y",
			wantErr: "year out of order",
		},
		{
			name:    "year out of order (after day)",
			input:   "P1D1Y",
			wantErr: "year out of order",
		},
		{
			name:    "year out of order (after T)",
			input:   "PT1H1Y",
			wantErr: "year out of order",
		},
		{
			name:    "month out of order (after month)",
			input:   "P1M1M",
			wantErr: "month out of order",
		},
		{
			name:    "month out of order (after day)",
			input:   "P1D1M",
			wantErr: "month out of order",
		},
		{
			name:    "day out of order (after day)",
			input:   "P1D1D",
			wantErr: "day out of order",
		},
		{
			name:    "day out of order (after T)",
			input:   "PT1H1D",
			wantErr: "day out of order",
		},
		{
			name:    "hour before T",
			input:   "P1H",
			wantErr: "hour must follow 'T'",
		},
		{
			name:    "hour out of order (after hour)",
			input:   "PT1H1H",
			wantErr: "hour out of order",
		},
		{
			name:    "hour out of order (after minute)",
			input:   "PT1M1H",
			wantErr: "hour out of order",
		},
		{
			name:    "minute out of order (after minute)",
			input:   "PT1M1M",
			wantErr: "minute out of order",
		},
		{
			name:    "minute out of order (after second)",
			input:   "PT1S1M",
			wantErr: "minute out of order",
		},
		{
			name:    "second before T",
			input:   "P1S",
			wantErr: "second must follow 'T'",
		},
		{
			name:    "second out of order",
			input:   "PT1S1S",
			wantErr: "second out of order",
		},
		{
			name:    "fractional days",
			input:   "P1.5D",
			wantErr: "overflow error",
		},
		{
			name:    "fractional hours",
			input:   "PT1.5H",
			wantErr: "overflow error",
		},
		{
			name:    "fractional time minutes",
			input:   "PT1.5M",
			wantErr: "overflow error",
		},
		{
			name:    "invalid integer for year",
			input:   "P99999999999999999999Y",
			wantErr: "overflow error",
		},
		{
			name:    "invalid integer for month",
			input:   "P99999999999999999999M",
			wantErr: "overflow error",
		},
		{
			name:    "invalid decimal for seconds",
			input:   "PT.S",
			wantErr: "overflow error",
		},
		{
			name:    "invalid decimal for days",
			input:   "P.D",
			wantErr: "overflow error",
		},
		{
			name:    "multiple dots in component",
			input:   "PT1.2.3S",
			wantErr: "unexpected type character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := durationParts(tt.input)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// TestAddDurationYM tests adding year-month values to a duration with overflow checks.
func TestAddDurationYM(t *testing.T) {
	tests := []struct {
		name        string
		curr        *int64
		val         int64
		isNegative  bool
		wantVal     int64
		wantErrOver bool
	}{
		{
			name:       "nil current positive",
			curr:       nil,
			val:        12,
			isNegative: false,
			wantVal:    12,
		},
		{
			name:       "nil current negative",
			curr:       nil,
			val:        12,
			isNegative: true,
			wantVal:    -12,
		},
		{
			name:       "non-nil current positive add",
			curr:       int64Ptr(10),
			val:        5,
			isNegative: false,
			wantVal:    15,
		},
		{
			name:        "overflow positive addition",
			curr:        int64Ptr(math.MaxInt64),
			val:         1,
			isNegative:  false,
			wantErrOver: true,
		},
		{
			name:        "overflow positive addition via negative duration",
			curr:        int64Ptr(math.MaxInt64),
			val:         -1,
			isNegative:  true,
			wantErrOver: true,
		},
		{
			name:        "overflow negative addition",
			curr:        int64Ptr(math.MinInt64),
			val:         -1,
			isNegative:  false,
			wantErrOver: true,
		},
		{
			name:        "overflow negative addition via negative duration",
			curr:        int64Ptr(math.MinInt64),
			val:         1,
			isNegative:  true,
			wantErrOver: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := addDurationYM(tt.curr, tt.val, tt.isNegative)
			if tt.wantErrOver {
				if !errors.Is(err, ErrDurationOverflow) && !errors.Is(err, errParseOverflow) {
					t.Errorf("expected overflow error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res == nil || *res != tt.wantVal {
				t.Errorf("got %v, want %d", res, tt.wantVal)
			}
		})
	}
}

// TestAddDurationDT tests adding day-time decimal values to a duration.
func TestAddDurationDT(t *testing.T) {
	tests := []struct {
		name       string
		curr       *Decimal
		val        Decimal
		isNegative bool
		wantVal    Decimal
	}{
		{
			name:       "nil current positive",
			curr:       nil,
			val:        NewDecimalFromInt64(100),
			isNegative: false,
			wantVal:    NewDecimalFromInt64(100),
		},
		{
			name:       "nil current negative",
			curr:       nil,
			val:        NewDecimalFromInt64(100),
			isNegative: true,
			wantVal:    NewDecimalFromInt64(-100),
		},
		{
			name:       "non-nil current positive add",
			curr:       decimalPtr(NewDecimalFromInt64(50)),
			val:        NewDecimalFromInt64(30),
			isNegative: false,
			wantVal:    NewDecimalFromInt64(80),
		},
		{
			name:       "non-nil current negative add",
			curr:       decimalPtr(NewDecimalFromInt64(50)),
			val:        NewDecimalFromInt64(30),
			isNegative: true,
			wantVal:    NewDecimalFromInt64(20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := addDurationDT(tt.curr, tt.val, tt.isNegative)
			if res.getValue().Cmp(tt.wantVal.getValue()) != 0 {
				t.Errorf("got %v, want %v", res, tt.wantVal)
			}
		})
	}
}
