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

// makeTestModel constructs a DateTimeSevenPropertyModel pointer for test setups.
func makeTestModel(yr int64, mo, da, hr, mi uint8, sec int64, tz *TimezoneOffset) *DateTimeSevenPropertyModel {
	dec := NewDecimalFromInt64(sec)
	return &DateTimeSevenPropertyModel{
		Year:           &yr,
		Month:          &mo,
		Day:            &da,
		Hour:           &hr,
		Minute:         &mi,
		Second:         &dec,
		TimezoneOffset: tz,
	}
}

// TestNewTimestamp tests the successful construction of Timestamp instances from valid models,
// as well as error handling when given an invalid model.
func TestNewTimestamp(t *testing.T) {
	tz, err := NewTimezoneOffset(120)
	if err != nil {
		t.Fatalf("unexpected timezone offset creation error: %v", err)
	}

	modelWithTZ := makeTestModel(2024, 2, 29, 12, 30, 45, &tz)
	tsWithTZ, err := newTimestamp(modelWithTZ)
	if err != nil {
		t.Fatalf("unexpected error creating timestamp with timezone: %v", err)
	}
	if tsWithTZ.timezoneOffset == nil || tsWithTZ.timezoneOffset.offset != 120 {
		t.Errorf("expected timezone offset 120, got %v", tsWithTZ.timezoneOffset)
	}

	modelUnzoned := makeTestModel(2023, 1, 15, 8, 10, 5, nil)
	tsUnzoned, err := newTimestamp(modelUnzoned)
	if err != nil {
		t.Fatalf("unexpected error creating unzoned timestamp: %v", err)
	}
	if tsUnzoned.timezoneOffset != nil {
		t.Errorf("expected nil timezone offset, got %v", tsUnzoned.timezoneOffset)
	}

	modelInvalid := makeTestModel(2024, 13, 1, 0, 0, 0, nil)
	if _, err = newTimestamp(modelInvalid); err == nil {
		t.Errorf("expected error for invalid month 13, got nil")
	}
}

// TestMustNewTimestamp tests the panic-free execution of MustNewTimestamp on valid models,
// and verifies that it panics on invalid models.
func TestMustNewTimestamp(t *testing.T) {
	model := makeTestModel(2021, 6, 1, 0, 0, 0, GetUTC())
	ts := MustNewTimestamp(model)
	if ts.year() != 2021 || ts.month() != 6 || ts.day() != 1 {
		t.Errorf("unexpected date components: %d-%d-%d", ts.year(), ts.month(), ts.day())
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected MustNewTimestamp to panic on invalid model, but it did not")
		}
	}()

	invalidModel := makeTestModel(2021, 13, 1, 0, 0, 0, GetUTC())
	MustNewTimestamp(invalidModel)
}

// TestTimestampToProperties tests converting a Timestamp back into a seven-property model.
func TestTimestampToProperties(t *testing.T) {
	tz, err := NewTimezoneOffset(-300)
	if err != nil {
		t.Fatalf("unexpected timezone creation error: %v", err)
	}

	model := makeTestModel(2020, 12, 31, 23, 59, 58, &tz)
	ts, err := newTimestamp(model)
	if err != nil {
		t.Fatalf("unexpected timestamp error: %v", err)
	}

	props := ts.ToProperties()
	if props == nil {
		t.Fatal("expected non-nil properties")
	}
	if props.Year == nil || *props.Year != 2020 {
		t.Errorf("expected year 2020, got %v", props.Year)
	}
	if props.Month == nil || *props.Month != 12 {
		t.Errorf("expected month 12, got %v", props.Month)
	}
	if props.Day == nil || *props.Day != 31 {
		t.Errorf("expected day 31, got %v", props.Day)
	}
	if props.Hour == nil || *props.Hour != 23 {
		t.Errorf("expected hour 23, got %v", props.Hour)
	}
	if props.Minute == nil || *props.Minute != 59 {
		t.Errorf("expected minute 59, got %v", props.Minute)
	}
	if props.Second == nil || props.Second.getValue().Cmp(NewDecimalFromInt64(58).getValue()) != 0 {
		t.Errorf("expected second 58, got %v", props.Second)
	}
	if props.TimezoneOffset == nil || props.TimezoneOffset.offset != -300 {
		t.Errorf("expected timezone offset -300, got %v", props.TimezoneOffset)
	}
}

// TestTimestampComponentGetters tests year, month, day, hour, minute, and second getter methods.
func TestTimestampComponentGetters(t *testing.T) {
	tz, err := NewTimezoneOffset(60)
	if err != nil {
		t.Fatalf("unexpected timezone offset error: %v", err)
	}

	secDecimal, err := ParseDecimal("45.5")
	if err != nil {
		t.Fatalf("unexpected decimal parse error: %v", err)
	}

	yr := int64(2025)
	mo := uint8(7)
	da := uint8(14)
	hr := uint8(18)
	mi := uint8(42)

	model := &DateTimeSevenPropertyModel{
		Year:           &yr,
		Month:          &mo,
		Day:            &da,
		Hour:           &hr,
		Minute:         &mi,
		Second:         &secDecimal,
		TimezoneOffset: &tz,
	}

	ts, err := newTimestamp(model)
	if err != nil {
		t.Fatalf("unexpected newTimestamp error: %v", err)
	}

	if ts.year() != 2025 {
		t.Errorf("expected year 2025, got %d", ts.year())
	}
	if ts.month() != 7 {
		t.Errorf("expected month 7, got %d", ts.month())
	}
	if ts.day() != 14 {
		t.Errorf("expected day 14, got %d", ts.day())
	}
	if ts.hour() != 18 {
		t.Errorf("expected hour 18, got %d", ts.hour())
	}
	if ts.minute() != 42 {
		t.Errorf("expected minute 42, got %d", ts.minute())
	}
	if ts.second().String() != "45.5" {
		t.Errorf("expected second 45.5, got %s", ts.second().String())
	}
}

// TestTimestampNegativeTimelineAndRemNegative tests negative timeline coordinates to exercise negative remainder
// handling.
func TestTimestampNegativeTimelineAndRemNegative(t *testing.T) {
	tsNeg := Timestamp{
		value:          NewDecimalFromInt64(-100000),
		timezoneOffset: nil,
	}

	h := tsNeg.hour()
	if h > 23 {
		t.Errorf("hour out of range: %d", h)
	}

	m := tsNeg.minute()
	if m > 59 {
		t.Errorf("minute out of range: %d", m)
	}

	s := tsNeg.second()
	if s.IsNegative() {
		t.Errorf("second must be non-negative: %s", s)
	}

	yr, mo, da := tsNeg.yearMonthDay()
	if yr >= 1 {
		t.Errorf("expected BCE year, got %d", yr)
	}
	if mo < 1 || mo > 12 {
		t.Errorf("month out of bounds: %d", mo)
	}
	if da < 1 || da > 31 {
		t.Errorf("day out of bounds: %d", da)
	}

	tz, err := NewTimezoneOffset(-60)
	if err != nil {
		t.Fatalf("unexpected timezone offset error: %v", err)
	}
	tsNegZoned := Timestamp{
		value:          NewDecimalFromInt64(-100000),
		timezoneOffset: &tz,
	}
	if tsNegZoned.hour() > 23 {
		t.Errorf("hour out of range: %d", tsNegZoned.hour())
	}
	if tsNegZoned.minute() > 59 {
		t.Errorf("minute out of range: %d", tsNegZoned.minute())
	}
}

// TestTimestampCheckedAddSubSeconds tests adding and subtracting seconds with precision retention.
func TestTimestampCheckedAddSubSeconds(t *testing.T) {
	ts := Timestamp{
		value:          NewDecimalFromInt64(100),
		timezoneOffset: GetUTC(),
	}

	added := ts.addSeconds(NewDecimalFromInt64(50))
	if added.value.getValue().Cmp(NewDecimalFromInt64(150).getValue()) != 0 {
		t.Errorf("expected 150, got %s", added.value.String())
	}
	if added.timezoneOffset != ts.timezoneOffset {
		t.Errorf("expected matching timezone offset")
	}

	subbed := ts.subSeconds(NewDecimalFromInt64(40))
	if subbed.value.getValue().Cmp(NewDecimalFromInt64(60).getValue()) != 0 {
		t.Errorf("expected 60, got %s", subbed.value.String())
	}
	if subbed.timezoneOffset != ts.timezoneOffset {
		t.Errorf("expected matching timezone offset")
	}
}

// TestTimestampAdjust tests adjusting timezone configurations between zoned and unzoned forms.
func TestTimestampAdjust(t *testing.T) {
	tz1, err := NewTimezoneOffset(60)
	if err != nil {
		t.Fatalf("unexpected timezone creation error: %v", err)
	}
	tz2, err := NewTimezoneOffset(120)
	if err != nil {
		t.Fatalf("unexpected timezone creation error: %v", err)
	}

	tsZoned := Timestamp{
		value:          NewDecimalFromInt64(3600),
		timezoneOffset: &tz1,
	}

	adjZonedToZoned := tsZoned.adjust(&tz2)
	if adjZonedToZoned.timezoneOffset == nil || adjZonedToZoned.timezoneOffset.offset != 120 {
		t.Errorf("expected offset 120, got %v", adjZonedToZoned.timezoneOffset)
	}
	if adjZonedToZoned.value.getValue().Cmp(NewDecimalFromInt64(3600).getValue()) != 0 {
		t.Errorf("expected timeline value unchanged at 3600, got %s", adjZonedToZoned.value.String())
	}

	adjZonedToUnzoned := tsZoned.adjust(nil)
	if adjZonedToUnzoned.timezoneOffset != nil {
		t.Errorf("expected nil timezone offset, got %v", adjZonedToUnzoned.timezoneOffset)
	}
	if adjZonedToUnzoned.value.getValue().Cmp(NewDecimalFromInt64(7200).getValue()) != 0 {
		t.Errorf("expected timeline value shifted to 7200, got %s", adjZonedToUnzoned.value.String())
	}

	tsUnzoned := Timestamp{
		value:          NewDecimalFromInt64(7200),
		timezoneOffset: nil,
	}

	adjUnzonedToZoned := tsUnzoned.adjust(&tz1)
	if adjUnzonedToZoned.timezoneOffset == nil || adjUnzonedToZoned.timezoneOffset.offset != 60 {
		t.Errorf("expected offset 60, got %v", adjUnzonedToZoned.timezoneOffset)
	}
	if adjUnzonedToZoned.value.getValue().Cmp(NewDecimalFromInt64(3600).getValue()) != 0 {
		t.Errorf("expected timeline value shifted to 3600, got %s", adjUnzonedToZoned.value.String())
	}

	adjUnzonedToUnzoned := tsUnzoned.adjust(nil)
	if adjUnzonedToUnzoned.timezoneOffset != nil {
		t.Errorf("expected nil timezone offset, got %v", adjUnzonedToUnzoned.timezoneOffset)
	}
	if adjUnzonedToUnzoned.value.getValue().Cmp(NewDecimalFromInt64(7200).getValue()) != 0 {
		t.Errorf("expected timeline value 7200, got %s", adjUnzonedToUnzoned.value.String())
	}
}

// TestTimestampCompare tests pairwise comparison across all combinations of zoned and unzoned timestamps.
func TestTimestampCompare(t *testing.T) {
	tsZoned1 := Timestamp{value: NewDecimalFromInt64(100), timezoneOffset: GetUTC()}
	tsZoned2 := Timestamp{value: NewDecimalFromInt64(200), timezoneOffset: GetUTC()}
	tsZonedEqual := Timestamp{value: NewDecimalFromInt64(100), timezoneOffset: GetUTC()}

	res, det := tsZoned1.compare(tsZoned2)
	if !det || res != -1 {
		t.Errorf("expected (-1, true), got (%d, %v)", res, det)
	}
	res, det = tsZoned2.compare(tsZoned1)
	if !det || res != 1 {
		t.Errorf("expected (1, true), got (%d, %v)", res, det)
	}
	res, det = tsZoned1.compare(tsZonedEqual)
	if !det || res != 0 {
		t.Errorf("expected (0, true), got (%d, %v)", res, det)
	}

	tsLocal1 := Timestamp{value: NewDecimalFromInt64(100), timezoneOffset: nil}
	tsLocal2 := Timestamp{value: NewDecimalFromInt64(200), timezoneOffset: nil}
	tsLocalEqual := Timestamp{value: NewDecimalFromInt64(100), timezoneOffset: nil}

	res, det = tsLocal1.compare(tsLocal2)
	if !det || res != -1 {
		t.Errorf("expected (-1, true), got (%d, %v)", res, det)
	}
	res, det = tsLocal2.compare(tsLocal1)
	if !det || res != 1 {
		t.Errorf("expected (1, true), got (%d, %v)", res, det)
	}
	res, det = tsLocal1.compare(tsLocalEqual)
	if !det || res != 0 {
		t.Errorf("expected (0, true), got (%d, %v)", res, det)
	}

	tsFarZoned := Timestamp{value: NewDecimalFromInt64(1_000_000_000), timezoneOffset: GetUTC()}
	tsLocalLow := Timestamp{value: NewDecimalFromInt64(0), timezoneOffset: nil}

	res, det = tsFarZoned.compare(tsLocalLow)
	if !det || res != -1 {
		t.Errorf("expected (-1, true), got (%d, %v)", res, det)
	}

	res, det = tsLocalLow.compare(tsFarZoned)
	if !det || res != 1 {
		t.Errorf("expected (1, true), got (%d, %v)", res, det)
	}

	tsZonedCenter := Timestamp{value: NewDecimalFromInt64(0), timezoneOffset: GetUTC()}
	tsLocalCenter := Timestamp{value: NewDecimalFromInt64(0), timezoneOffset: nil}

	res, det = tsZonedCenter.compare(tsLocalCenter)
	if det || res != 0 {
		t.Errorf("expected (0, false), got (%d, %v)", res, det)
	}
	res, det = tsLocalCenter.compare(tsZonedCenter)
	if det || res != 0 {
		t.Errorf("expected (0, false), got (%d, %v)", res, det)
	}
}

// TestTimestampYearMonthDayCycles tests calendar conversions covering 400-year, 100-year, 4-year, and normal year
// cycles.
func TestTimestampYearMonthDayCycles(t *testing.T) {
	testDates := []struct {
		yr int64
		mo uint8
		da uint8
	}{
		{2000, 2, 29},
		{1900, 2, 28},
		{2004, 2, 29},
		{2001, 1, 1},
		{2001, 12, 31},
		{1600, 1, 1},
		{2400, 12, 31},
		{1, 1, 1},
	}

	for _, td := range testDates {
		model := makeTestModel(td.yr, td.mo, td.da, 0, 0, 0, nil)
		ts, err := newTimestamp(model)
		if err != nil {
			t.Fatalf("unexpected newTimestamp error for %d-%d-%d: %v", td.yr, td.mo, td.da, err)
		}

		gotYr, gotMo, gotDa := ts.yearMonthDay()
		if gotYr != td.yr || gotMo != td.mo || gotDa != td.da {
			t.Errorf("expected %d-%d-%d, got %d-%d-%d", td.yr, td.mo, td.da, gotYr, gotMo, gotDa)
		}
	}
}
