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

package datatypes

import (
	"math/big"
	"time"
)

// uint8Lookup is a 256-byte constant mapping non-negative integer indices [0, 255] directly to uint8 (byte).
const uint8Lookup = "" +
	"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f" +
	"\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f" +
	"\x20\x21\x22\x23\x24\x25\x26\x27\x28\x29\x2a\x2b\x2c\x2d\x2e\x2f" +
	"\x30\x31\x32\x33\x34\x35\x36\x37\x38\x39\x3a\x3b\x3c\x3d\x3e\x3f" +
	"\x40\x41\x42\x43\x44\x45\x46\x47\x48\x49\x4a\x4b\x4c\x4d\x4e\x4f" +
	"\x50\x51\x52\x53\x54\x55\x56\x57\x58\x59\x5a\x5b\x5c\x5d\x5e\x5f" +
	"\x60\x61\x62\x63\x64\x65\x66\x67\x68\x69\x6a\x6b\x6c\x6d\x6e\x6f" +
	"\x70\x71\x72\x73\x74\x75\x76\x77\x78\x79\x7a\x7b\x7c\x7d\x7e\x7f" +
	"\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f" +
	"\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f" +
	"\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf" +
	"\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf" +
	"\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf" +
	"\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf" +
	"\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef" +
	"\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"

// Timestamp represents the timeline coordinate and optional timezone offset as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Representation:
// A high-precision linear timeline coordinate represented in seconds alongside an optional timezone offset, mapped
// internally without a direct user-facing lexical space.
type Timestamp struct {
	// value is the linear coordinate on the timeline in seconds.
	value Decimal

	// timezoneOffset represents the explicit timezone offset, or nil if the value is unzoned.
	timezoneOffset *TimezoneOffset
}

// newTimestamp constructs a Timestamp from the seven-property date/time model.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
// - props: A pointer to the seven-property date/time model containing year, month, day, hour, minute, second, and
// timezoneOffset.
//
// Returns:
//   - Timestamp: The resolved timeline coordinate.
//   - error: An error if calendar field boundaries are violated or if timeline coordinate resolution fails.
func newTimestamp(props *DateTimeSevenPropertyModel) (Timestamp, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2.1)
	// Validate the discrete components of the Seven-property Model against the value space constraints.
	if err := validateDateTimeModel(props); err != nil {
		return Timestamp{}, err
	}
	val := timeOnTimeline(props)
	return Timestamp{value: val, timezoneOffset: props.TimezoneOffset}, nil
}

// MustNewTimestamp constructs a Timestamp from the seven-property date/time model and panics on failure.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Parameters:
// - props: A pointer to the seven-property date/time model containing year, month, day, hour, minute, second, and
// timezoneOffset.
//
// Returns:
//   - Timestamp: The resolved timeline coordinate.
func MustNewTimestamp(props *DateTimeSevenPropertyModel) Timestamp {
	ts, err := newTimestamp(props)
	if err != nil {
		panic(err)
	}
	return ts
}

// ToProperties deconstructs the Timestamp back into the seven-property date/time model.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//
// Returns:
//   - *DateTimeSevenPropertyModel: A pointer to the reconstructed seven-property model.
func (t Timestamp) ToProperties() *DateTimeSevenPropertyModel {
	yr, mo, da := t.yearMonthDay()
	return &DateTimeSevenPropertyModel{
		Year:           &yr,
		Month:          &mo,
		Day:            &da,
		Hour:           uint8Ptr(t.hour()),
		Minute:         uint8Ptr(t.minute()),
		Second:         decimalPtr(t.second()),
		TimezoneOffset: t.timezoneOffset,
	}
}

// yearMonthDay calculates the calendar year, month, and day represented by the Timestamp under Gregorian calendar
// rules.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.2)
//
// Parameters:
//
// Returns:
//   - int64: The calendar year, which may be negative or zero.
//   - uint8: The calendar month, ranging from 1 to 12.
//   - uint8: The calendar day of the month, ranging from 1 to 31.
func (t Timestamp) yearMonthDay() (int64, uint8, uint8) {
	// Step 1: Timezone Adjustment
	// Apply the timezone offset in minutes to align the local properties with the timeline coordinate.
	tzOffsetMinutes := int64(0)
	if t.timezoneOffset != nil {
		tzOffsetMinutes = int64(t.timezoneOffset.offset)
	}

	// Step 2: Conversion of Timeline Seconds to Days
	// Convert the cumulative timezone-adjusted seconds into integral calendar days.
	totalSeconds := new(big.Int).Add(t.value.asI128(), big.NewInt(tzOffsetMinutes*secondsPerMinute))
	days := new(big.Int).Div(totalSeconds, big.NewInt(secondsPerDay))

	// Step 3: Deconstruct days into Gregorian year, month, and day.
	// Timeline origin (days = 0) corresponds to 0001-01-01T00:00:00Z.
	d := days.Int64()
	tGo := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(d))

	return int64(tGo.Year()), uint8Lookup[tGo.Month()], uint8Lookup[tGo.Day()]
}

// year extracts the Gregorian calendar year from the Timestamp.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Parameters:
//
// Returns:
//   - int64: The calendar year.
func (t Timestamp) year() int64 { y, _, _ := t.yearMonthDay(); return y }

// month extracts the Gregorian calendar month from the Timestamp.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Parameters:
//
// Returns:
//   - uint8: The calendar month, bounded between 1 and 12.
func (t Timestamp) month() uint8 { _, m, _ := t.yearMonthDay(); return m }

// day extracts the Gregorian calendar day of the month from the Timestamp.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Parameters:
//
// Returns:
//   - uint8: The calendar day of the month, bounded between 1 and 31.
func (t Timestamp) day() uint8 { _, _, d := t.yearMonthDay(); return d }

// hour extracts the hour component from the Timestamp.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Parameters:
//
// Returns:
//   - uint8: The hour component of the timestamp, bounded between 0 and 23.
func (t Timestamp) hour() uint8 {
	tzOffsetMinutes := int64(0)
	if t.timezoneOffset != nil {
		tzOffsetMinutes = int64(t.timezoneOffset.offset)
	}
	val := new(big.Int).Add(t.value.asI128(), big.NewInt(tzOffsetMinutes*secondsPerMinute))
	remDay := new(big.Int).Rem(val, big.NewInt(secondsPerDay))
	if remDay.Sign() < 0 {
		remDay.Add(remDay, big.NewInt(secondsPerDay))
	}
	h := new(big.Int).Div(remDay, big.NewInt(secondsPerHour))

	return uint8Lookup[h.Int64()]
}

// minute extracts the minute component from the Timestamp.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Parameters:
//
// Returns:
//   - uint8: The minute component of the timestamp, bounded between 0 and 59.
func (t Timestamp) minute() uint8 {
	tzOffsetMinutes := int64(0)
	if t.timezoneOffset != nil {
		tzOffsetMinutes = int64(t.timezoneOffset.offset)
	}
	val := new(big.Int).Add(t.value.asI128(), big.NewInt(tzOffsetMinutes*secondsPerMinute))
	remHour := new(big.Int).Rem(val, big.NewInt(secondsPerHour))
	if remHour.Sign() < 0 {
		remHour.Add(remHour, big.NewInt(secondsPerHour))
	}
	m := new(big.Int).Div(remHour, big.NewInt(minutesPerHour))

	return uint8Lookup[m.Int64()]
}

// second extracts the second component as a high-precision Decimal from the Timestamp.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7)
//
// Parameters:
//
// Returns:
//   - Decimal: The high-precision decimal second component of the timestamp, strictly less than 60.
func (t Timestamp) second() Decimal {
	res, _ := t.value.CheckedRemEuclid(NewDecimalFromInt64(secondsPerMinute))
	return res.Abs()
}

// addSeconds returns a new Timestamp with the specified seconds added.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - seconds: The high-precision Decimal duration of seconds to add.
//
// Returns:
//   - Timestamp: The resulting Timestamp value.
func (t Timestamp) addSeconds(seconds Decimal) Timestamp {
	v := t.value.Add(seconds)
	return Timestamp{value: v, timezoneOffset: t.timezoneOffset}
}

// subSeconds returns a new Timestamp with the specified seconds subtracted.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - seconds: The high-precision Decimal duration of seconds to subtract.
//
// Returns:
//   - Timestamp: The resulting Timestamp value.
func (t Timestamp) subSeconds(seconds Decimal) Timestamp {
	v := t.value.Sub(seconds)
	return Timestamp{value: v, timezoneOffset: t.timezoneOffset}
}

// adjust maps the Timestamp to a different timezone offset, modifying the timeline value if moving between zoned and
// unzoned representations.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.3)
//
// Parameters:
//   - newTz: A pointer to the target timezone offset, or nil if the target is unzoned.
//
// Returns:
//   - Timestamp: The adjusted Timestamp value.
func (t Timestamp) adjust(newTz *TimezoneOffset) Timestamp {
	if t.timezoneOffset != nil {
		if newTz != nil {
			return Timestamp{value: t.value, timezoneOffset: newTz}
		}

		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix E.3)
		// When removing an explicit timezone from a zoned value to make it unzoned, the local timeline coordinate must
		// be shifted by the offset amount to preserve the original local time field values.
		offsetDur := t.timezoneOffset.ToDayTimeDuration()
		newVal := t.value.Add(offsetDur.AsSeconds())
		return Timestamp{value: newVal, timezoneOffset: nil}
	}
	if newTz != nil {
		// Spec Rule: W3C XSD 1.1 Part 2 (Appendix E.3)
		// When adding an explicit timezone to an unzoned value to make it zoned, the timeline coordinate is shifted in
		// the opposite direction of the offset to anchor the local time fields.
		offsetDur := newTz.ToDayTimeDuration()
		newVal := t.value.Sub(offsetDur.AsSeconds())
		return Timestamp{value: newVal, timezoneOffset: newTz}
	}
	return t
}

// compare evaluates the temporal order of two Timestamps.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix D.2)
//
// Parameters:
//   - other: The other Timestamp to compare against.
//
// Returns:
//   - int: Negative one (-1) if the receiver is chronologically before other, zero (0) if equal, or one (1) if after.
//   - bool: False if the comparison is indeterminate under partial ordering rules.
func (t Timestamp) compare(other Timestamp) (int, bool) {
	if (t.timezoneOffset != nil && other.timezoneOffset != nil) ||
		(t.timezoneOffset == nil && other.timezoneOffset == nil) {
		return t.value.getValue().Cmp(other.value.getValue()), true
	}

	var zoned, local Timestamp
	if t.timezoneOffset != nil {
		zoned, local = t, other
	} else {
		zoned, local = other, t
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix D.2)
	// Comparing zoned and unzoned times requires assuming a maximum timezone offset variation of +/- 14 hours.
	// The comparison is determinate if and only if the ordering remains identical when local is adjusted to both
	// the positive and negative extremes of the timezone range.
	tzRange := getTimezoneMaxDuration()
	localPlus := local.addSeconds(tzRange.seconds)
	localMinus := local.subSeconds(tzRange.seconds)

	cmpPlus := localPlus.value.getValue().Cmp(zoned.value.getValue())
	cmpMinus := localMinus.value.getValue().Cmp(zoned.value.getValue())

	if cmpPlus == cmpMinus {
		if t.timezoneOffset != nil {
			return cmpPlus, true
		}
		return -cmpPlus, true
	}
	return 0, false
}
