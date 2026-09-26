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
package datatypes

import (
	"errors"
	"testing"
)

// parseTestDateTimeStamp parses a dateTimeStamp string for testing or fails the test.
func parseTestDateTimeStamp(t *testing.T, input string) DateTimeStamp {
	t.Helper()
	dts, err := ParseDateTimeStamp(input)
	if err != nil {
		t.Fatalf("failed to parse valid dateTimeStamp %q: %v", input, err)
	}
	return dts
}

// TestParseDateTimeStamp tests parsing various valid and invalid dateTimeStamp inputs.
func TestParseDateTimeStamp(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "valid UTC timezone Z",
			input:   "2025-01-15T12:00:00Z",
			wantErr: nil,
		},
		{
			name:    "valid positive timezone offset",
			input:   "2025-01-15T12:00:00+02:00",
			wantErr: nil,
		},
		{
			name:    "valid negative timezone offset",
			input:   "2025-01-15T12:00:00-05:00",
			wantErr: nil,
		},
		{
			name:    "invalid dateTime syntax",
			input:   "not-a-date-time",
			wantErr: newParseDateTimeError("year should be encoded on at least 4 digits"),
		},
		{
			name:    "missing timezone offset",
			input:   "2025-01-15T12:00:00",
			wantErr: ErrMissingTimezone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateTimeStamp(tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("ParseDateTimeStamp(%q) expected error %v, got nil", tt.input, tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) && err.Error() != tt.wantErr.Error() {
					t.Errorf("ParseDateTimeStamp(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseDateTimeStamp(%q) unexpected error: %v", tt.input, err)
			}
			if got.TimezoneOffset() == nil {
				t.Errorf("ParseDateTimeStamp(%q) returned DateTimeStamp with nil TimezoneOffset", tt.input)
			}
		})
	}
}

// TestDateTimeStampString tests the string conversion method of DateTimeStamp.
func TestDateTimeStampString(t *testing.T) {
	input := "2025-05-20T15:30:45.5Z"
	dts := parseTestDateTimeStamp(t, input)

	got := dts.String()
	want := dts.DateTime.String()
	if got != want {
		t.Errorf("DateTimeStamp.String() = %q, want %q", got, want)
	}
}

// TestDateTimeStampIsIdenticalWith tests the identity comparison method of DateTimeStamp.
func TestDateTimeStampIsIdenticalWith(t *testing.T) {
	dts1 := parseTestDateTimeStamp(t, "2025-01-15T12:00:00Z")
	dts2 := parseTestDateTimeStamp(t, "2025-01-15T12:00:00Z")
	dts3 := parseTestDateTimeStamp(t, "2025-01-15T14:00:00+02:00")
	dts4 := parseTestDateTimeStamp(t, "2025-01-16T12:00:00Z")

	tests := []struct {
		name  string
		dts   DateTimeStamp
		other XSDValue
		want  bool
	}{
		{
			name:  "identical dateTimeStamp values",
			dts:   dts1,
			other: dts2,
			want:  true,
		},
		{
			name:  "different timezone offset components",
			dts:   dts1,
			other: dts3,
			want:  false,
		},
		{
			name:  "different datetime values",
			dts:   dts1,
			other: dts4,
			want:  false,
		},
		{
			name:  "incompatible type (String)",
			dts:   dts1,
			other: String("2025-01-15T12:00:00Z"),
			want:  false,
		},
		{
			name:  "incompatible type (DateTime)",
			dts:   dts1,
			other: dts1.DateTime,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dts.IsIdenticalWith(tt.other)
			if got != tt.want {
				t.Errorf("DateTimeStamp.IsIdenticalWith() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDateTimeStampCompare tests the chronological comparison method of DateTimeStamp.
func TestDateTimeStampCompare(t *testing.T) {
	dts1 := parseTestDateTimeStamp(t, "2025-01-15T12:00:00Z")
	dts2 := parseTestDateTimeStamp(t, "2025-01-15T13:00:00Z")
	dtsEqual := parseTestDateTimeStamp(t, "2025-01-15T14:00:00+02:00")

	tests := []struct {
		name    string
		dts     DateTimeStamp
		other   XSDValue
		want    int
		wantErr error
	}{
		{
			name:    "less than",
			dts:     dts1,
			other:   dts2,
			want:    -1,
			wantErr: nil,
		},
		{
			name:    "greater than",
			dts:     dts2,
			other:   dts1,
			want:    1,
			wantErr: nil,
		},
		{
			name:    "equal timeline points with different timezones",
			dts:     dts1,
			other:   dtsEqual,
			want:    0,
			wantErr: nil,
		},
		{
			name:    "incomparable type",
			dts:     dts1,
			other:   String("2025-01-15T12:00:00Z"),
			want:    0,
			wantErr: ErrDateTimeOverflow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.dts.Compare(tt.other)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("DateTimeStamp.Compare() error = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("DateTimeStamp.Compare() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Errorf("DateTimeStamp.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}
