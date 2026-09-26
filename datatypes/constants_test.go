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
	"math/big"
	"testing"
)

// TestConstantsValues verifies that all temporal, mathematical, cycle,
// and string constants defined in constants.go match their required values.
func TestConstantsValues(t *testing.T) {
	tests := []struct {
		name     string
		expected interface{}
		actual   interface{}
	}{
		{"secondsPerMinute", 60, secondsPerMinute},
		{"minutesPerHour", 60, minutesPerHour},
		{"secondsPerHour", 3600, secondsPerHour},
		{"hoursPerDay", 24, hoursPerDay},
		{"secondsPerDay", 86400, secondsPerDay},
		{"monthsPerYear", 12, monthsPerYear},
		{"secondsPerNormalYear", 31536000, secondsPerNormalYear},
		{"daysInNormalYear", 365, daysInNormalYear},
		{"daysInLeapYear", 366, daysInLeapYear},
		{"daysIn4Years", 1461, daysIn4Years},
		{"daysIn100Years", 36524, daysIn100Years},
		{"daysIn400Years", 146097, daysIn400Years},
		{"maxTimezoneHours", 14, maxTimezoneHours},
		{"maxTimezoneOffsetMinutes", 840, maxTimezoneOffsetMinutes},
		{"minTimezoneOffsetMinutes", -840, minTimezoneOffsetMinutes},
		{"maxMinutesPerHour", 59, maxMinutesPerHour},
		{"xsdDefaultEpochYear", 1971, xsdDefaultEpochYear},
		{"defaultTimeYear", 1972, defaultTimeYear},
		{"defaultTimeMonth", 12, defaultTimeMonth},
		{"defaultTimeDay", 31, defaultTimeDay},
		{"anchorYear1696", 1696, anchorYear1696},
		{"anchorYear1697", 1697, anchorYear1697},
		{"anchorYear1903", 1903, anchorYear1903},
		{"primitiveString", "string", primitiveString},
		{"strINF", "INF", strINF},
		{"strPosINF", "+INF", strPosINF},
		{"strNegINF", "-INF", strNegINF},
		{"strNaN", "NaN", strNaN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != tt.expected {
				t.Errorf("expected %s=%v, got %v", tt.name, tt.expected, tt.actual)
			}
		})
	}

	t.Run("LeapYearFactors", func(t *testing.T) {
		if leapYearFactor4 != 4 || leapYearFactor100 != 100 || leapYearFactor400 != 400 {
			t.Errorf("unexpected leap year factors: 4=%d, 100=%d, 400=%d",
				leapYearFactor4, leapYearFactor100, leapYearFactor400)
		}
	})

	t.Run("DaysInMonthConstants", func(t *testing.T) {
		if daysInFebruaryNormal != 28 || daysInFebruaryLeap != 29 ||
			daysInShortMonth != 30 || daysInLongMonth != 31 || maxDaysInMonth != 31 {
			t.Errorf("unexpected days-in-month constants")
		}
	})

	t.Run("MonthConstants", func(t *testing.T) {
		if monthFebruary != 2 || monthMarch != 3 || monthApril != 4 ||
			monthJune != 6 || monthJuly != 7 || monthSeptember != 9 || monthNovember != 11 {
			t.Errorf("unexpected month constants")
		}
	})

	t.Run("BitSizeConstants", func(t *testing.T) {
		if bitSize8 != 8 || bitSize16 != 16 || bitSize32 != 32 || bitSize64 != 64 {
			t.Errorf("unexpected integer bit size constants: 8=%d, 16=%d, 32=%d, 64=%d",
				bitSize8, bitSize16, bitSize32, bitSize64)
		}
	})

	t.Run("ParserScalingConstants", func(t *testing.T) {
		if nanosecondsScale != 9 || base5 != 5 || base10 != 10 ||
			builderGrowthFactor != 2 || maxFloatPrecision != 340 ||
			minYearDigits != 4 || twoDigits != 2 || maxRune16 != 0xFFFF {
			t.Errorf("unexpected numerical or parser scaling constants")
		}
	})
}

// TestGetAnchorDates asserts that getAnchorDates returns the four comparison anchor
// date timestamps.
func TestGetAnchorDates(t *testing.T) {
	anchors := getAnchorDates()
	if len(anchors) != 4 {
		t.Fatalf("expected 4 anchor dates, got %d", len(anchors))
	}

	for i, ts := range anchors {
		if !ts.value.IsPositive() {
			t.Errorf("anchor[%d] expected positive timeline value, got %s", i, ts.value.String())
		}
		if ts.timezoneOffset != nil {
			t.Errorf("anchor[%d] expected nil timezone offset, got %v", i, ts.timezoneOffset)
		}
	}

	for i := range len(anchors) - 1 {
		if anchors[i].value.getValue().Cmp(anchors[i+1].value.getValue()) >= 0 {
			t.Errorf("expected anchor[%d] < anchor[%d], got cmp >= 0", i, i+1)
		}
	}
}

// TestGetTimezoneMaxDuration asserts that getTimezoneMaxDuration returns a DayTimeDuration
// corresponding to the maximum positive 14-hour timezone offset limit.
func TestGetTimezoneMaxDuration(t *testing.T) {
	dur := getTimezoneMaxDuration()

	expectedSeconds := big.NewRat(int64(maxTimezoneHours*secondsPerHour), 1)
	if dur.AsSeconds().getValue().Cmp(expectedSeconds) != 0 {
		t.Errorf("expected max timezone duration to be %s seconds, got %s",
			expectedSeconds.String(), dur.AsSeconds().String())
	}

	if dur.Hours() != int64(maxTimezoneHours) {
		t.Errorf("expected %d hours, got %d", maxTimezoneHours, dur.Hours())
	}
	if dur.Days() != 0 {
		t.Errorf("expected 0 days, got %d", dur.Days())
	}
	if dur.Minutes() != 0 {
		t.Errorf("expected 0 minutes, got %d", dur.Minutes())
	}
}
