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
	"strings"
	"testing"
)

type parseYearCase struct {
	name       string
	input      string
	wantYear   int64
	wantRem    string
	wantErrMsg string
}

// checkParseYear is a helper function to test parseYear cases.
func checkParseYear(t *testing.T, tt parseYearCase) {
	t.Helper()
	gotYear, gotRem, err := parseYear(tt.input)
	if tt.wantErrMsg != "" {
		if err == nil {
			t.Fatalf("parseYear(%q) expected error containing %q, got nil", tt.input, tt.wantErrMsg)
		}
		if !strings.Contains(err.Error(), tt.wantErrMsg) {
			t.Errorf("parseYear(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.wantErrMsg)
		}
		var parseErr ParseDateTimeError
		if errors.As(err, &parseErr) && parseErr.Error() == "" {
			t.Errorf("ParseDateTimeError.Error() returned empty string")
		}
		return
	}
	if err != nil {
		t.Fatalf("parseYear(%q) unexpected error: %v", tt.input, err)
	}
	if gotYear != tt.wantYear {
		t.Errorf("parseYear(%q) year = %d, want %d", tt.input, gotYear, tt.wantYear)
	}
	if gotRem != tt.wantRem {
		t.Errorf("parseYear(%q) rem = %q, want %q", tt.input, gotRem, tt.wantRem)
	}
}

// TestParseYear tests parsing of years.
func TestParseYear(t *testing.T) {
	tests := []parseYearCase{
		{
			name:     "valid 4-digit positive year",
			input:    "2023-01-01",
			wantYear: 2023,
			wantRem:  "-01-01",
		},
		{
			name:     "valid 4-digit negative year",
			input:    "-0044-03-15",
			wantYear: -44,
			wantRem:  "-03-15",
		},
		{
			name:     "valid 5-digit year without leading zero",
			input:    "12023T00:00:00",
			wantYear: 12023,
			wantRem:  "T00:00:00",
		},
		{
			name:       "invalid year with less than 4 digits",
			input:      "123-01-01",
			wantErrMsg: "year should be encoded on at least 4 digits",
		},
		{
			name:       "invalid year with negative sign and less than 4 digits",
			input:      "-123-01-01",
			wantErrMsg: "year should be encoded on at least 4 digits",
		},
		{
			name:       "invalid year > 4 digits starting with 0",
			input:      "02023-01-01",
			wantErrMsg: "years must not start with 0 if more than 4 digits",
		},
		{
			name:       "invalid year integer overflow",
			input:      "9999999999999999999999999999-01-01",
			wantErrMsg: "invalid year",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkParseYear(t, tt)
		})
	}
}

type parseTwoDigitCase struct {
	name       string
	input      string
	minVal     int
	maxVal     int
	fieldName  string
	wantVal    uint8
	wantRem    string
	wantErrMsg string
}

// checkParseTwoDigit is a helper function to test parseTwoDigit cases.
func checkParseTwoDigit(t *testing.T, tt parseTwoDigitCase) {
	t.Helper()
	gotVal, gotRem, err := parseTwoDigit(tt.input, tt.minVal, tt.maxVal, tt.fieldName)
	if tt.wantErrMsg != "" {
		if err == nil {
			t.Fatalf("parseTwoDigit(%q) expected error containing %q, got nil", tt.input, tt.wantErrMsg)
		}
		if !strings.Contains(err.Error(), tt.wantErrMsg) {
			t.Fatalf(
				"parseTwoDigit(%q) error = %q, want error containing %q",
				tt.input,
				err.Error(),
				tt.wantErrMsg,
			)
		}
		return
	}
	if err != nil {
		t.Fatalf("parseTwoDigit(%q) unexpected error: %v", tt.input, err)
	}
	if gotVal != tt.wantVal {
		t.Errorf("parseTwoDigit(%q) val = %d, want %d", tt.input, gotVal, tt.wantVal)
	}
	if gotRem != tt.wantRem {
		t.Errorf("parseTwoDigit(%q) rem = %q, want %q", tt.input, gotRem, tt.wantRem)
	}
}

// TestParseTwoDigit tests parsing of two-digit fields.
func TestParseTwoDigit(t *testing.T) {
	tests := []parseTwoDigitCase{
		{
			name:      "valid two digit",
			input:     "05-01",
			minVal:    1,
			maxVal:    12,
			fieldName: "month",
			wantVal:   5,
			wantRem:   "-01",
		},
		{
			name:       "invalid single digit",
			input:      "5-01",
			minVal:     1,
			maxVal:     12,
			fieldName:  "month",
			wantErrMsg: "month must be encoded with two digits",
		},
		{
			name:       "invalid three digits",
			input:      "005-01",
			minVal:     1,
			maxVal:     12,
			fieldName:  "month",
			wantErrMsg: "month must be encoded with two digits",
		},
		{
			name:       "below minVal",
			input:      "00-01",
			minVal:     1,
			maxVal:     12,
			fieldName:  "month",
			wantErrMsg: "month must be between 01 and 12",
		},
		{
			name:       "above maxVal",
			input:      "13-01",
			minVal:     1,
			maxVal:     12,
			fieldName:  "month",
			wantErrMsg: "month must be between 01 and 12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkParseTwoDigit(t, tt)
		})
	}
}

type parseSecondCase struct {
	name       string
	input      string
	wantSec    string
	wantRem    string
	wantErrMsg string
}

// checkParseSecond is a helper function to test parseSecond cases.
func checkParseSecond(t *testing.T, tt parseSecondCase) {
	t.Helper()
	gotSec, gotRem, err := parseSecond(tt.input)
	if tt.wantErrMsg != "" {
		if err == nil {
			t.Fatalf("parseSecond(%q) expected error containing %q, got nil", tt.input, tt.wantErrMsg)
		}
		if !strings.Contains(err.Error(), tt.wantErrMsg) {
			t.Fatalf(
				"parseSecond(%q) error = %q, want error containing %q",
				tt.input,
				err.Error(),
				tt.wantErrMsg,
			)
		}
		return
	}
	if err != nil {
		t.Fatalf("parseSecond(%q) unexpected error: %v", tt.input, err)
	}
	if gotSec.String() != tt.wantSec {
		t.Errorf("parseSecond(%q) sec = %s, want %s", tt.input, gotSec.String(), tt.wantSec)
	}
	if gotRem != tt.wantRem {
		t.Errorf("parseSecond(%q) rem = %q, want %q", tt.input, gotRem, tt.wantRem)
	}
}

// TestParseSecond tests parsing of seconds.
func TestParseSecond(t *testing.T) {
	tests := []parseSecondCase{
		{
			name:    "integer seconds",
			input:   "45Z",
			wantSec: "45",
			wantRem: "Z",
		},
		{
			name:    "fractional seconds",
			input:   "05.1234+02:00",
			wantSec: "5.1234",
			wantRem: "+02:00",
		},
		{
			name:    "stops parsing prefix at non-decimal character",
			input:   "05.1a",
			wantSec: "5.1",
			wantRem: "a",
		},
		{
			name:       "less than 2 digits integer part",
			input:      "5.123",
			wantErrMsg: "seconds integer part must be two digits",
		},
		{
			name:       "more than 2 digits integer part with dot",
			input:      "105.123",
			wantErrMsg: "seconds integer part must be two digits",
		},
		{
			name:       "ends with a dot",
			input:      "05.",
			wantErrMsg: "seconds cannot end with a dot",
		},
		{
			name:       "seconds 60 or above",
			input:      "60.0",
			wantErrMsg: "seconds must be between 00.0 and 60.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkParseSecond(t, tt)
		})
	}
}

type parseTimezoneCase struct {
	name       string
	input      string
	wantTz     string
	wantRem    string
	isUnzoned  bool
	wantErrMsg string
}

// checkParseTimezone is a helper function to test parseTimezone cases.
func checkParseTimezone(t *testing.T, tt parseTimezoneCase) {
	t.Helper()
	gotTz, gotRem, err := parseTimezone(tt.input)
	if tt.wantErrMsg != "" {
		if err == nil {
			t.Fatalf("parseTimezone(%q) expected error containing %q, got nil", tt.input, tt.wantErrMsg)
		}
		if !strings.Contains(err.Error(), tt.wantErrMsg) {
			t.Errorf("parseTimezone(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.wantErrMsg)
		}
		return
	}
	if err != nil {
		t.Fatalf("parseTimezone(%q) unexpected error: %v", tt.input, err)
	}
	if tt.isUnzoned {
		if gotTz != nil {
			t.Errorf("parseTimezone(%q) expected nil timezone, got %v", tt.input, gotTz)
		}
	} else if gotTz == nil || gotTz.String() != tt.wantTz {
		t.Errorf("parseTimezone(%q) tz = %v, want %s", tt.input, gotTz, tt.wantTz)
	}
	if gotRem != tt.wantRem {
		t.Errorf("parseTimezone(%q) rem = %q, want %q", tt.input, gotRem, tt.wantRem)
	}
}

// TestParseTimezone tests parsing of timezones.
func TestParseTimezone(t *testing.T) {
	tests := []parseTimezoneCase{
		{
			name:      "empty string",
			input:     "",
			isUnzoned: true,
			wantRem:   "",
		},
		{
			name:    "UTC Z",
			input:   "Zextra",
			wantTz:  "Z",
			wantRem: "extra",
		},
		{
			name:    "positive timezone offset",
			input:   "+05:30",
			wantTz:  "+05:30",
			wantRem: "",
		},
		{
			name:    "negative timezone offset",
			input:   "-08:00",
			wantTz:  "-08:00",
			wantRem: "",
		},
		{
			name:      "no timezone prefix",
			input:     "ABC",
			isUnzoned: true,
			wantRem:   "ABC",
		},
		{
			name:       "invalid hour in timezone",
			input:      "+15:00",
			wantErrMsg: "timezone hour must be between 00 and 14",
		},
		{
			name:       "missing colon separator in timezone",
			input:      "+05X00",
			wantErrMsg: "timezone hours and minutes must be separated by ':'",
		},
		{
			name:       "invalid minute in timezone",
			input:      "+05:60",
			wantErrMsg: "timezone minute must be between 00 and 59",
		},
		{
			name:       "timezone exceeding 14:00 limit",
			input:      "+14:01",
			wantErrMsg: "timezone offset cannot exceed 14:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkParseTimezone(t, tt)
		})
	}
}

// TestExpectChar tests expected character matching.
func TestExpectChar(t *testing.T) {
	t.Run("matching character", func(t *testing.T) {
		rem, err := expectChar("-2023", '-', "error message")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rem != "2023" {
			t.Errorf("expectChar rem = %q, want %q", rem, "2023")
		}
	})

	t.Run("non-matching character", func(t *testing.T) {
		_, err := expectChar("2023", '-', "expected hyphen")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

// TestValidateDayOfMonth tests day of month validation.
func TestValidateDayOfMonth(t *testing.T) {
	y2023 := int64(2023)
	y2024 := int64(2024)

	t.Run("valid day in non-leap year", func(t *testing.T) {
		if err := validateDayOfMonth(&y2023, 2, 28); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid 29th day in non-leap year Feb", func(t *testing.T) {
		if err := validateDayOfMonth(&y2023, 2, 29); err == nil {
			t.Errorf("expected error for Feb 29 in 2023, got nil")
		}
	})

	t.Run("valid 29th day in leap year Feb", func(t *testing.T) {
		if err := validateDayOfMonth(&y2024, 2, 29); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nil year allows leap day 29 in Feb", func(t *testing.T) {
		if err := validateDayOfMonth(nil, 2, 29); err != nil {
			t.Errorf("unexpected error for nil year Feb 29: %v", err)
		}
	})
}

type parseDateTimeCase struct {
	name       string
	input      string
	wantYear   int64
	wantMonth  uint8
	wantDay    uint8
	wantHour   uint8
	wantMin    uint8
	wantSec    string
	wantRem    string
	wantErrMsg string
}

// checkInt64Prop is a helper function to check int64 date properties.
func checkInt64Prop(t *testing.T, input, field string, got *int64, want int64) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("parseDateTime(%q) %s = %v, want %d", input, field, got, want)
	}
}

// checkUint8Prop is a helper function to check uint8 date properties.
func checkUint8Prop(t *testing.T, input, field string, got *uint8, want uint8) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("parseDateTime(%q) %s = %v, want %d", input, field, got, want)
	}
}

// checkParseDateTime is a helper function to test parseDateTime cases.
func checkParseDateTime(t *testing.T, tt parseDateTimeCase) {
	t.Helper()
	props, rem, err := parseDateTime(tt.input)
	if tt.wantErrMsg != "" {
		if err == nil {
			t.Fatalf("parseDateTime(%q) expected error containing %q, got nil", tt.input, tt.wantErrMsg)
		}
		if !strings.Contains(err.Error(), tt.wantErrMsg) {
			t.Errorf("parseDateTime(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.wantErrMsg)
		}
		return
	}
	if err != nil {
		t.Fatalf("parseDateTime(%q) unexpected error: %v", tt.input, err)
	}
	checkInt64Prop(t, tt.input, "year", props.Year, tt.wantYear)
	checkUint8Prop(t, tt.input, "month", props.Month, tt.wantMonth)
	checkUint8Prop(t, tt.input, "day", props.Day, tt.wantDay)
	checkUint8Prop(t, tt.input, "hour", props.Hour, tt.wantHour)
	checkUint8Prop(t, tt.input, "minute", props.Minute, tt.wantMin)
	if props.Second == nil || props.Second.String() != tt.wantSec {
		t.Errorf("parseDateTime(%q) second = %v, want %s", tt.input, props.Second, tt.wantSec)
	}
	if rem != tt.wantRem {
		t.Errorf("parseDateTime(%q) rem = %q, want %q", tt.input, rem, tt.wantRem)
	}
}

// TestParseDateTime tests parsing of full date-time strings.
func TestParseDateTime(t *testing.T) {
	tests := []parseDateTimeCase{
		{
			name:      "valid full dateTime with timezone",
			input:     "2023-10-25T14:30:45.500+02:00REST",
			wantYear:  2023,
			wantMonth: 10,
			wantDay:   25,
			wantHour:  14,
			wantMin:   30,
			wantSec:   "45.5",
			wantRem:   "REST",
		},
		{
			name:      "valid midnight 24:00:00",
			input:     "2023-10-25T24:00:00Z",
			wantYear:  2023,
			wantMonth: 10,
			wantDay:   25,
			wantHour:  24,
			wantMin:   0,
			wantSec:   "0",
			wantRem:   "",
		},
		{
			name:       "invalid year format",
			input:      "23-10-25T14:30:45",
			wantErrMsg: "year should be encoded on at least 4 digits",
		},
		{
			name:       "missing year-month hyphen separator",
			input:      "2023X10-25T14:30:45",
			wantErrMsg: "year and month must be separated by '-'",
		},
		{
			name:       "invalid month value",
			input:      "2023-13-25T14:30:45",
			wantErrMsg: "month must be between 01 and 12",
		},
		{
			name:       "missing month-day hyphen separator",
			input:      "2023-10X25T14:30:45",
			wantErrMsg: "month and day must be separated by '-'",
		},
		{
			name:       "invalid day value",
			input:      "2023-10-32T14:30:45",
			wantErrMsg: "day must be between 01 and 31",
		},
		{
			name:       "missing T separator",
			input:      "2023-10-25 14:30:45",
			wantErrMsg: "date and time must be separated by 'T'",
		},
		{
			name:       "invalid hour value",
			input:      "2023-10-25T25:30:45",
			wantErrMsg: "hour must be between 00 and 24",
		},
		{
			name:       "missing hour-minute colon separator",
			input:      "2023-10-25T14X30:45",
			wantErrMsg: "hours and minutes must be separated by ':'",
		},
		{
			name:       "invalid minute value",
			input:      "2023-10-25T14:60:45",
			wantErrMsg: "minute must be between 00 and 59",
		},
		{
			name:       "missing minute-second colon separator",
			input:      "2023-10-25T14:30X45",
			wantErrMsg: "minutes and seconds must be separated by ':'",
		},
		{
			name:       "invalid second value",
			input:      "2023-10-25T14:30:60",
			wantErrMsg: "seconds must be between 00.0 and 60.0",
		},
		{
			name:       "24:00:01 time exceeding midnight 24:00:00",
			input:      "2023-10-25T24:00:01Z",
			wantErrMsg: "times are not allowed to be after 24:00:00",
		},
		{
			name:       "24:01:00 time exceeding midnight 24:00:00",
			input:      "2023-10-25T24:01:00Z",
			wantErrMsg: "times are not allowed to be after 24:00:00",
		},
		{
			name:       "invalid day for Feb non-leap year",
			input:      "2023-02-29T12:00:00",
			wantErrMsg: "29 is not a valid day of month 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkParseDateTime(t, tt)
		})
	}
}

// TestDecimalPrefix tests extraction of decimal prefixes.
func TestDecimalPrefix(t *testing.T) {
	tests := []struct {
		input      string
		wantPrefix string
		wantRem    string
	}{
		{"12.34abc", "12.34", "abc"},
		{"12.34.56", "12.34", ".56"},
		{"abc", "", "abc"},
		{"123", "123", ""},
	}

	for _, tt := range tests {
		gotPrefix, gotRem := decimalPrefix(tt.input)
		if gotPrefix != tt.wantPrefix || gotRem != tt.wantRem {
			t.Errorf(
				"decimalPrefix(%q) = (%q, %q), want (%q, %q)",
				tt.input,
				gotPrefix,
				gotRem,
				tt.wantPrefix,
				tt.wantRem,
			)
		}
	}
}

// TestIntegerPrefix tests extraction of integer prefixes.
func TestIntegerPrefix(t *testing.T) {
	tests := []struct {
		input      string
		wantPrefix string
		wantRem    string
	}{
		{"123abc", "123", "abc"},
		{"abc", "", "abc"},
		{"123", "123", ""},
	}

	for _, tt := range tests {
		gotPrefix, gotRem := integerPrefix(tt.input)
		if gotPrefix != tt.wantPrefix || gotRem != tt.wantRem {
			t.Errorf(
				"integerPrefix(%q) = (%q, %q), want (%q, %q)",
				tt.input,
				gotPrefix,
				gotRem,
				tt.wantPrefix,
				tt.wantRem,
			)
		}
	}
}
