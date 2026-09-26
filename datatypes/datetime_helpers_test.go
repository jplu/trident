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
	"math"
	"testing"
)

// TestTimeOnTimeline tests the timeOnTimeline function with various property models.
func TestTimeOnTimeline(t *testing.T) {
	yr := int64(2024)
	mo := uint8(2)
	da := uint8(29)
	hr := uint8(14)
	mi := uint8(30)
	se := NewDecimalFromInt64(45)
	tz, _ := NewTimezoneOffset(120)

	propsFull := &DateTimeSevenPropertyModel{
		Year:           &yr,
		Month:          &mo,
		Day:            &da,
		Hour:           &hr,
		Minute:         &mi,
		Second:         &se,
		TimezoneOffset: &tz,
	}

	res := timeOnTimeline(propsFull)
	if res.getValue().Sign() == 0 {
		t.Errorf("expected non-zero timeline coordinate")
	}

	propsEmpty := &DateTimeSevenPropertyModel{}
	resEmpty := timeOnTimeline(propsEmpty)
	if resEmpty.getValue() == nil {
		t.Errorf("expected valid timeline decimal for empty props")
	}
}

// TestNormalizeMonth tests the normalizeMonth function across various edge cases.
func TestNormalizeMonth(t *testing.T) {
	tests := []struct {
		name    string
		yr      int64
		mo      int64
		wantYr  int64
		wantMo  uint8
		wantErr bool
	}{
		{
			name:   "standard month within range",
			yr:     2024,
			mo:     5,
			wantYr: 2024,
			wantMo: 5,
		},
		{
			name:   "month 12 boundary",
			yr:     2024,
			mo:     12,
			wantYr: 2024,
			wantMo: 12,
		},
		{
			name:   "month overflow into next year",
			yr:     2024,
			mo:     13,
			wantYr: 2025,
			wantMo: 1,
		},
		{
			name:   "month overflow by multiple years",
			yr:     2024,
			mo:     25,
			wantYr: 2026,
			wantMo: 1,
		},
		{
			name:   "month 0 roll back to previous December",
			yr:     2024,
			mo:     0,
			wantYr: 2023,
			wantMo: 12,
		},
		{
			name:   "negative month rollback",
			yr:     2024,
			mo:     -1,
			wantYr: 2023,
			wantMo: 11,
		},
		{
			name:   "negative month rollback multiple years",
			yr:     2024,
			mo:     -13,
			wantYr: 2022,
			wantMo: 11,
		},
		{
			name:    "positive year overflow",
			yr:      math.MaxInt64/2 + 10,
			mo:      1,
			wantErr: true,
		},
		{
			name:    "negative year underflow",
			yr:      math.MinInt64/2 - 10,
			mo:      1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotYr, gotMo, err := normalizeMonth(tt.yr, tt.mo)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotYr != tt.wantYr || gotMo != tt.wantMo {
				t.Errorf("normalizeMonth(%d, %d) = (%d, %d), want (%d, %d)",
					tt.yr, tt.mo, gotYr, gotMo, tt.wantYr, tt.wantMo)
			}
		})
	}
}

// TestDateTimePlusDuration tests adding a duration to a date-time model.
func TestDateTimePlusDuration(t *testing.T) {
	yr := int64(2023)
	mo := uint8(1)
	da := uint8(31)
	hr := uint8(10)
	mi := uint8(30)
	se := NewDecimalFromInt64(0)
	tz, _ := NewTimezoneOffset(60)

	dt := &DateTimeSevenPropertyModel{
		Year:           &yr,
		Month:          &mo,
		Day:            &da,
		Hour:           &hr,
		Minute:         &mi,
		Second:         &se,
		TimezoneOffset: &tz,
	}

	dur, _ := NewDuration(1, DefaultDecimal())
	res, err := dateTimePlusDuration(dur, dt)
	if err != nil {
		t.Fatalf("unexpected error from dateTimePlusDuration: %v", err)
	}
	if res == nil || res.Year == nil || res.Month == nil || res.Day == nil {
		t.Fatalf("expected date fields to be populated")
	}
	if *res.Year != 2023 {
		t.Errorf("expected year 2023, got %d", *res.Year)
	}
	if res.TimezoneOffset == nil || *res.TimezoneOffset != tz {
		t.Errorf("expected timezone offset %v, got %v", tz, res.TimezoneOffset)
	}

	secVal := NewDecimalFromInt64(7200)
	durDT, _ := NewDuration(0, secVal)
	resDT, err := dateTimePlusDuration(durDT, dt)
	if err != nil {
		t.Fatalf("unexpected error from dateTimePlusDuration with day-time: %v", err)
	}
	if resDT == nil || resDT.Hour == nil {
		t.Fatalf("expected hour to be populated")
	}

	dtEmpty := &DateTimeSevenPropertyModel{}
	resEmpty, err := dateTimePlusDuration(dur, dtEmpty)
	if err != nil {
		t.Fatalf("unexpected error with empty fields: %v", err)
	}
	if resEmpty.Year != nil || resEmpty.Month != nil || resEmpty.Day != nil ||
		resEmpty.Hour != nil || resEmpty.Minute != nil || resEmpty.Second != nil ||
		resEmpty.TimezoneOffset != nil {
		t.Errorf("expected originally nil fields to remain nil in output")
	}

	overflowYr := int64(math.MaxInt64 / 2)
	dtOverflow := &DateTimeSevenPropertyModel{
		Year: &overflowYr,
	}
	overflowDur, _ := NewDuration(24, DefaultDecimal())
	_, err = dateTimePlusDuration(overflowDur, dtOverflow)
	if err == nil {
		t.Fatalf("expected overflow error, got nil")
	}
}

// TestDaysInMonth tests calculating the number of days in a month for various years.
func TestDaysInMonth(t *testing.T) {
	leap400 := int64(2000)
	leap4 := int64(2024)
	nonLeap100 := int64(1900)
	nonLeapNormal := int64(2023)

	tests := []struct {
		name     string
		y        *int64
		m        int64
		wantDays uint8
	}{
		{"Feb 2000 (divisible by 400)", &leap400, 2, 29},
		{"Feb 2024 (divisible by 4)", &leap4, 2, 29},
		{"Feb 1900 (divisible by 100, not 400)", &nonLeap100, 2, 28},
		{"Feb 2023 (non-leap year)", &nonLeapNormal, 2, 28},
		{"Feb absent year (leap year assumed)", nil, 2, 29},
		{"April (30 days)", &nonLeapNormal, 4, 30},
		{"June (30 days)", &nonLeapNormal, 6, 30},
		{"September (30 days)", &nonLeapNormal, 9, 30},
		{"November (30 days)", &nonLeapNormal, 11, 30},
		{"January (31 days)", &nonLeapNormal, 1, 31},
		{"July (31 days)", &nonLeapNormal, 7, 31},
		{"December (31 days)", &nonLeapNormal, 12, 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := daysInMonth(tt.y, tt.m)
			if got != tt.wantDays {
				t.Errorf("daysInMonth(%v, %d) = %d, want %d", tt.y, tt.m, got, tt.wantDays)
			}
		})
	}
}

// TestMinUint8 tests the minUint8 helper function.
func TestMinUint8(t *testing.T) {
	if minUint8(5, 10) != 5 {
		t.Errorf("minUint8(5, 10) != 5")
	}
	if minUint8(15, 10) != 10 {
		t.Errorf("minUint8(15, 10) != 10")
	}
	if minUint8(7, 7) != 7 {
		t.Errorf("minUint8(7, 7) != 7")
	}
}

// TestHelperPointerFunctions tests various pointer and mapping helper utilities.
func TestHelperPointerFunctions(t *testing.T) {
	i := int64(100)
	iPtr := int64Ptr(i)
	if iPtr == nil || *iPtr != 100 {
		t.Errorf("int64Ptr failed")
	}

	u := uint8(20)
	uPtr := uint8Ptr(u)
	if uPtr == nil || *uPtr != 20 {
		t.Errorf("uint8Ptr failed")
	}

	d := NewDecimalFromInt64(5)
	dPtr := decimalPtr(d)
	if dPtr == nil || !dPtr.IsIdenticalWith(d) {
		t.Errorf("decimalPtr failed")
	}

	if mapIf[int64, int64](nil, 42) != nil {
		t.Errorf("mapIf with nil originalField should return nil")
	}
	orig := int64(1)
	mapped := mapIf(&orig, int64(42))
	if mapped == nil || *mapped != 42 {
		t.Errorf("mapIf with non-nil originalField should return pointer to value")
	}
}

// TestValidateDateTimeModel tests validateDateTimeModel covering all value-space validation branches.
func TestValidateDateTimeModel(t *testing.T) {
	if err := validateDateTimeModel(nil); err != nil {
		t.Errorf("validateDateTimeModel(nil) expected nil, got %v", err)
	}

	moZero := uint8(0)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{Month: &moZero}); err == nil {
		t.Errorf("expected error for month 0, got nil")
	}
	mo13 := uint8(13)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{Month: &mo13}); err == nil {
		t.Errorf("expected error for month 13, got nil")
	}

	daZero := uint8(0)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{Day: &daZero}); err == nil {
		t.Errorf("expected error for day 0, got nil")
	}
	yr2023 := int64(2023)
	mo2 := uint8(2)
	da29 := uint8(29)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{
		Year:  &yr2023,
		Month: &mo2,
		Day:   &da29,
	}); err == nil {
		t.Errorf("expected error for Feb 29 on non-leap year, got nil")
	}
	yr2024 := int64(2024)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{
		Year:  &yr2024,
		Month: &mo2,
		Day:   &da29,
	}); err != nil {
		t.Errorf("unexpected error for Feb 29 on leap year: %v", err)
	}
	da32 := uint8(32)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{Day: &da32}); err == nil {
		t.Errorf("expected error for day 32 without month, got nil")
	}
	da31 := uint8(31)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{Day: &da31}); err != nil {
		t.Errorf("unexpected error for day 31 without month: %v", err)
	}

	hr25 := uint8(25)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{Hour: &hr25}); err == nil {
		t.Errorf("expected error for hour 25, got nil")
	}
	hr24 := uint8(24)
	mi1 := uint8(1)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{
		Hour:   &hr24,
		Minute: &mi1,
	}); err == nil {
		t.Errorf("expected error for 24:01:00, got nil")
	}
	se1 := NewDecimalFromInt64(1)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{
		Hour:   &hr24,
		Second: &se1,
	}); err == nil {
		t.Errorf("expected error for 24:00:01, got nil")
	}
	mi0 := uint8(0)
	se0 := DefaultDecimal()
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{
		Hour:   &hr24,
		Minute: &mi0,
		Second: &se0,
	}); err != nil {
		t.Errorf("unexpected error for 24:00:00: %v", err)
	}
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{
		Hour: &hr24,
	}); err != nil {
		t.Errorf("unexpected error for hour 24 with nil min/sec: %v", err)
	}

	mi60 := uint8(60)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{Minute: &mi60}); err == nil {
		t.Errorf("expected error for minute 60, got nil")
	}

	seNeg := NewDecimalFromInt64(-1)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{Second: &seNeg}); err == nil {
		t.Errorf("expected error for negative seconds, got nil")
	}

	se60 := NewDecimalFromInt64(60)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{Second: &se60}); err == nil {
		t.Errorf("expected error for 60 seconds, got nil")
	}

	yr := int64(2024)
	mo := uint8(12)
	da := uint8(31)
	hr := uint8(23)
	mi := uint8(59)
	se := NewDecimalFromInt64(59)
	if err := validateDateTimeModel(&DateTimeSevenPropertyModel{
		Year:   &yr,
		Month:  &mo,
		Day:    &da,
		Hour:   &hr,
		Minute: &mi,
		Second: &se,
	}); err != nil {
		t.Errorf("unexpected error for fully valid model: %v", err)
	}
}
