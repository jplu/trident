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
	"math"
	"testing"
)

// TestNewDateTime tests creating a new DateTime instance and validating error handling.
func TestNewDateTime(t *testing.T) {
	utc := GetUTC()
	sec := NewDecimalFromInt64(30)

	dt, err := NewDateTime(2023, 10, 25, 14, 30, sec, utc)
	if err != nil {
		t.Fatalf("unexpected error creating DateTime: %v", err)
	}
	if dt.Year() != 2023 || dt.Month() != 10 || dt.Hour() != 14 || dt.Minute() != 30 {
		t.Errorf("unexpected DateTime components: %s", dt.String())
	}

	errTests := []struct {
		name   string
		year   int64
		month  uint8
		day    uint8
		hour   uint8
		minute uint8
		second Decimal
		tz     *TimezoneOffset
	}{
		{
			name:   "invalid month 0",
			year:   2023,
			month:  0,
			day:    1,
			hour:   0,
			minute: 0,
			second: DefaultDecimal(),
		},
		{
			name:   "invalid month 13",
			year:   2023,
			month:  13,
			day:    1,
			hour:   0,
			minute: 0,
			second: DefaultDecimal(),
		},
		{
			name:   "invalid day 0",
			year:   2023,
			month:  1,
			day:    0,
			hour:   0,
			minute: 0,
			second: DefaultDecimal(),
		},
		{
			name:   "invalid day for month (Feb 30)",
			year:   2023,
			month:  2,
			day:    30,
			hour:   0,
			minute: 0,
			second: DefaultDecimal(),
		},
		{
			name:   "invalid hour > 24",
			year:   2023,
			month:  1,
			day:    1,
			hour:   25,
			minute: 0,
			second: DefaultDecimal(),
		},
		{
			name:   "invalid hour 24 with non-zero minute",
			year:   2023,
			month:  1,
			day:    1,
			hour:   24,
			minute: 1,
			second: DefaultDecimal(),
		},
		{
			name:   "invalid hour 24 with non-zero second",
			year:   2023,
			month:  1,
			day:    1,
			hour:   24,
			minute: 0,
			second: NewDecimalFromInt64(1),
		},
		{
			name:   "invalid minute > 59",
			year:   2023,
			month:  1,
			day:    1,
			hour:   12,
			minute: 60,
			second: DefaultDecimal(),
		},
		{
			name:   "invalid second >= 60",
			year:   2023,
			month:  1,
			day:    1,
			hour:   12,
			minute: 0,
			second: NewDecimalFromInt64(60),
		},
	}

	for _, tt := range errTests {
		t.Run(tt.name, func(t *testing.T) {
			_, dtErr := NewDateTime(tt.year, tt.month, tt.day, tt.hour, tt.minute, tt.second, tt.tz)
			if dtErr == nil {
				t.Errorf("NewDateTime(%s) expected error, got nil", tt.name)
			}
		})
	}
}

// TestParseDateTimeType tests parsing various DateTime string formats.
func TestParseDateTimeType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid UTC dateTime",
			input:   "2023-10-25T14:30:00Z",
			wantErr: false,
		},
		{
			name:    "valid dateTime with positive offset",
			input:   "2023-10-25T14:30:00+02:00",
			wantErr: false,
		},
		{
			name:    "valid dateTime at hour 24:00:00",
			input:   "2023-10-25T24:00:00Z",
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid-datetime",
			wantErr: true,
		},
		{
			name:    "unrecognized value suffix",
			input:   "2023-10-25T14:30:00ZEXTRA",
			wantErr: true,
		},
		{
			name:    "invalid calendar date (Feb 30)",
			input:   "2023-02-30T14:30:00Z",
			wantErr: true,
		},
		{
			name:    "invalid hour 24 with non-zero minute",
			input:   "2023-10-25T24:01:00Z",
			wantErr: true,
		},
		{
			name:    "invalid hour 24 with non-zero second",
			input:   "2023-10-25T24:00:01Z",
			wantErr: true,
		},
		{
			name:    "invalid second >= 60",
			input:   "2023-10-25T14:30:60Z",
			wantErr: true,
		},
		{
			name:    "invalid second ending with dot",
			input:   "2023-10-25T14:30:00.Z",
			wantErr: true,
		},
		{
			name:    "invalid year with leading zero exceeding 4 digits",
			input:   "02023-10-25T14:30:00Z",
			wantErr: true,
		},
		{
			name:    "invalid timezone exceeding 14:00",
			input:   "2023-10-25T14:30:00+14:01",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateTime(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got.String() == "" {
				t.Errorf("ParseDateTime(%q) returned empty string representation", tt.input)
			}
		})
	}
}

// TestDateTimeGetters tests retrieving individual fields of a DateTime.
func TestDateTimeGetters(t *testing.T) {
	tz, _ := NewTimezoneOffset(120) // +02:00
	sec, _ := ParseDecimal("30.500")

	dt, err := NewDateTime(2025, 5, 12, 10, 20, sec, &tz)
	if err != nil {
		t.Fatalf("failed to construct DateTime: %v", err)
	}

	if dt.Year() != 2025 {
		t.Errorf("Year() = %d, want 2025", dt.Year())
	}
	if dt.Month() != 5 {
		t.Errorf("Month() = %d, want 5", dt.Month())
	}
	if dt.Day() != 12 {
		t.Errorf("Day() = %d, want 12", dt.Day())
	}
	if dt.Hour() != 10 {
		t.Errorf("Hour() = %d, want 10", dt.Hour())
	}
	if dt.Minute() != 20 {
		t.Errorf("Minute() = %d, want 20", dt.Minute())
	}
	if dt.Second().String() != "30.5" {
		t.Errorf("Second() = %s, want 30.5", dt.Second().String())
	}
	if dt.TimezoneOffset() == nil || dt.TimezoneOffset().String() != "+02:00" {
		t.Errorf("TimezoneOffset() = %v, want +02:00", dt.TimezoneOffset())
	}

	unzoned, err := ParseTime("14:25:36")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if unzoned.TimezoneOffset() != nil {
		t.Errorf("expected nil timezone offset, got %v", unzoned.TimezoneOffset())
	}
}

// TestDateTimeCheckedAddDuration tests adding a duration to a DateTime.
func TestDateTimeCheckedAddDuration(t *testing.T) {
	dt, _ := ParseDateTime("2023-01-01T00:00:00Z")
	dur, _ := ParseDuration("P1Y1M1DT1H1M1S")

	res := dt.AddDuration(dur)
	if res.String() != "2024-02-02T01:01:01Z" {
		t.Errorf("CheckedAddDuration = %s, want 2024-02-02T01:01:01Z", res.String())
	}

	durNeg, _ := ParseDuration("-P1Y1M1DT1H1M1S")
	resNeg := res.AddDuration(durNeg)
	if resNeg.String() != "2023-01-01T00:00:00Z" {
		t.Errorf("CheckedAddDuration with negative duration = %s, want 2023-01-01T00:00:00Z", resNeg.String())
	}
}

// TestDateTimeCheckedSubDuration tests subtracting a duration from a DateTime.
func TestDateTimeCheckedSubDuration(t *testing.T) {
	dt, _ := ParseDateTime("2024-02-02T01:01:01Z")
	dur, _ := ParseDuration("P1Y1M1DT1H1M1S")

	res, err := dt.CheckedSubDuration(dur)
	if err != nil {
		t.Fatalf("unexpected error subtracting duration: %v", err)
	}
	if res.String() != "2023-01-01T00:00:00Z" {
		t.Errorf("CheckedSubDuration = %s, want 2023-01-01T00:00:00Z", res.String())
	}

	invalidNegDuration := Duration{yearMonth: NewYearMonthDuration(math.MinInt64)}
	_, err = dt.CheckedSubDuration(invalidNegDuration)
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected ErrDateTimeOverflow, got %v", err)
	}
}

// TestDateTimeIsIdenticalWith tests the identity comparison method.
func TestDateTimeIsIdenticalWith(t *testing.T) {
	dt1, _ := ParseDateTime("2023-10-25T14:30:00Z")
	dt2, _ := ParseDateTime("2023-10-25T14:30:00Z")
	dtDiffInstant, _ := ParseDateTime("2023-10-25T15:30:00Z")
	dtUnzoned1, _ := ParseDateTime("2023-10-25T14:30:00")
	dtUnzoned2, _ := ParseDateTime("2023-10-25T14:30:00")
	dtDiffTz, _ := ParseDateTime("2023-10-25T16:30:00+02:00")

	if !dt1.IsIdenticalWith(dt2) {
		t.Errorf("expected dt1 and dt2 to be identical")
	}
	if dt1.IsIdenticalWith(dtDiffInstant) {
		t.Errorf("expected dt1 and dtDiffInstant to not be identical")
	}
	if !dtUnzoned1.IsIdenticalWith(dtUnzoned2) {
		t.Errorf("expected dtUnzoned1 and dtUnzoned2 to be identical")
	}
	if dt1.IsIdenticalWith(dtUnzoned1) {
		t.Errorf("expected zoned and unzoned dateTime to not be identical")
	}
	if dt1.IsIdenticalWith(dtDiffTz) {
		t.Errorf("expected different timezone offsets to not be identical")
	}
	if dt1.IsIdenticalWith(String("2023-10-25T14:30:00Z")) {
		t.Errorf("expected non-DateTime value to return false for IsIdenticalWith")
	}
}

// TestDateTimeCompare tests comparing two DateTime values.
func TestDateTimeCompare(t *testing.T) {
	dt1, _ := ParseDateTime("2023-10-25T14:30:00Z")
	dt2, _ := ParseDateTime("2023-10-25T15:30:00Z")
	dt3, _ := ParseDateTime("2023-10-25T14:30:00Z")
	dtUnzoned, _ := ParseDateTime("2023-10-25T14:30:00")

	_, err := dt1.Compare(String("2023-10-25T14:30:00Z"))
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected ErrDateTimeOverflow when comparing with non-DateTime, got %v", err)
	}

	cmp, err := dt1.Compare(dt2)
	if err != nil || cmp != -1 {
		t.Errorf("dt1.Compare(dt2) = (%d, %v), want (-1, nil)", cmp, err)
	}

	cmp, err = dt2.Compare(dt1)
	if err != nil || cmp != 1 {
		t.Errorf("dt2.Compare(dt1) = (%d, %v), want (1, nil)", cmp, err)
	}

	cmp, err = dt1.Compare(dt3)
	if err != nil || cmp != 0 {
		t.Errorf("dt1.Compare(dt3) = (%d, %v), want (0, nil)", cmp, err)
	}

	_, err = dt1.Compare(dtUnzoned)
	if !errors.Is(err, ErrDateTimeOverflow) {
		t.Errorf("expected ErrDateTimeOverflow for indeterminate comparison, got %v", err)
	}
}

// TestDateTimeString tests the string formatting of DateTime.
func TestDateTimeString(t *testing.T) {
	tests := []struct {
		name     string
		year     int64
		month    uint8
		day      uint8
		hour     uint8
		minute   uint8
		second   string
		timezone *TimezoneOffset
		expected string
	}{
		{
			name:     "positive year, single-digit integer second without fraction",
			year:     2023,
			month:    1,
			day:      1,
			hour:     0,
			minute:   0,
			second:   "5",
			timezone: GetUTC(),
			expected: "2023-01-01T00:00:05Z",
		},
		{
			name:     "positive year, single-digit integer second with fraction",
			year:     2023,
			month:    1,
			day:      1,
			hour:     0,
			minute:   0,
			second:   "5.25",
			timezone: GetUTC(),
			expected: "2023-01-01T00:00:05.25Z",
		},
		{
			name:     "positive year, double-digit integer second with fraction",
			year:     2023,
			month:    1,
			day:      1,
			hour:     0,
			minute:   0,
			second:   "15.25",
			timezone: GetUTC(),
			expected: "2023-01-01T00:00:15.25Z",
		},
		{
			name:     "negative year formatting",
			year:     -5,
			month:    3,
			day:      15,
			hour:     12,
			minute:   30,
			second:   "00",
			timezone: nil,
			expected: "-0005-03-15T12:30:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sec, err := ParseDecimal(tt.second)
			if err != nil {
				t.Fatalf("failed to parse second decimal: %v", err)
			}
			dt, err := NewDateTime(tt.year, tt.month, tt.day, tt.hour, tt.minute, sec, tt.timezone)
			if err != nil {
				t.Fatalf("failed to create DateTime: %v", err)
			}
			if dt.String() != tt.expected {
				t.Errorf("dt.String() = %q, want %q", dt.String(), tt.expected)
			}
		})
	}
}

// TestDateTimeAdjust tests adjusting the timezone offset of a DateTime.
func TestDateTimeAdjust(t *testing.T) {
	dt, _ := ParseDateTime("2023-10-25T14:30:00Z")
	tzPlus2, _ := NewTimezoneOffset(120)

	adjusted := dt.Adjust(&tzPlus2)
	if adjusted.String() != "2023-10-25T16:30:00+02:00" {
		t.Errorf("adjusted.String() = %q, want %q", adjusted.String(), "2023-10-25T16:30:00+02:00")
	}

	unzoned := dt.Adjust(nil)
	if unzoned.TimezoneOffset() != nil {
		t.Errorf("expected timezone to be nil, got %v", unzoned.TimezoneOffset())
	}

	dtUnzoned, _ := ParseDateTime("2023-10-25T14:30:00")
	zonedFromUnzoned := dtUnzoned.Adjust(GetUTC())
	if zonedFromUnzoned.TimezoneOffset() == nil {
		t.Errorf("expected timezone to not be nil")
	}
}
