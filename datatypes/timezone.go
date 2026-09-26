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
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

// TimezoneOffset represents the timezone offset component as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.3)
//
// Representation:
// The timezoneOffset is represented as a signed integer offset in minutes, bounded to the range [-840, 840] inclusive,
// representing the temporal difference from Coordinated Universal Time (UTC).
type TimezoneOffset struct {
	// offset holds the timezone offset in minutes.
	offset int16
}

// GetUTC returns a pointer to a TimezoneOffset representing Coordinated Universal Time (UTC).
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.3)
//
// Parameters:
//
// Returns:
//   - *TimezoneOffset: A pointer to a TimezoneOffset with an offset value of 0 minutes.
func GetUTC() *TimezoneOffset {
	return &TimezoneOffset{offset: 0}
}

// NewTimezoneOffset creates a new TimezoneOffset from an offset duration expressed in minutes.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.3)
//
// Parameters:
//   - offsetInMinutes: The timezone offset duration in minutes.
//
// Returns:
//   - TimezoneOffset: The initialized TimezoneOffset struct.
//   - error: An InvalidTimezoneError if the offset duration is outside the range [-840, 840] inclusive.
func NewTimezoneOffset(offsetInMinutes int16) (TimezoneOffset, error) {
	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.7.3)
	// Timezone offsets must be bounded between -14:00 (-840 minutes) and +14:00 (+840 minutes) inclusive.
	if offsetInMinutes < minTimezoneOffsetMinutes || offsetInMinutes > maxTimezoneOffsetMinutes {
		return TimezoneOffset{}, InvalidTimezoneError{OffsetInMinutes: int64(offsetInMinutes)}
	}
	return TimezoneOffset{offset: offsetInMinutes}, nil
}

// NewTimezoneOffsetFromDayTimeDuration converts a DayTimeDuration into a TimezoneOffset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.3)
//
// Parameters:
//   - d: The source DayTimeDuration representing the timezone offset duration.
//
// Returns:
//   - TimezoneOffset: The mapped timezone offset in minutes.
//
// - error: An InvalidTimezoneError if the duration contains non-zero seconds or exceeds the maximum timezone offset
// boundaries.
func NewTimezoneOffsetFromDayTimeDuration(d DayTimeDuration) (TimezoneOffset, error) {
	totalSeconds := d.AsSeconds().asI128()

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.2.7.3)
	// A timezone offset must represent an integral number of minutes. Any non-zero second or fractional second
	// component is invalid.
	if new(big.Int).Rem(totalSeconds, big.NewInt(secondsPerMinute)).Int64() != 0 {
		return TimezoneOffset{}, InvalidTimezoneError{}
	}
	minutes := new(big.Int).Quo(totalSeconds, big.NewInt(secondsPerMinute))
	if !minutes.IsInt64() {
		return TimezoneOffset{}, InvalidTimezoneError{}
	}
	m := minutes.Int64()
	if m < math.MinInt16 || m > math.MaxInt16 {
		return TimezoneOffset{}, InvalidTimezoneError{OffsetInMinutes: m}
	}

	return NewTimezoneOffset(int16(m))
}

// ToDayTimeDuration converts the TimezoneOffset into its equivalent DayTimeDuration representation.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.2.7.3)
//
// Parameters:
//
// Returns:
//   - DayTimeDuration: The timezone offset represented as a duration in seconds with decimal precision.
func (t TimezoneOffset) ToDayTimeDuration() DayTimeDuration {
	return NewDayTimeDuration(NewDecimalFromInt64(int64(t.offset) * secondsPerMinute))
}

// String returns the canonical lexical representation of the TimezoneOffset.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Appendix E.2)
//
// Parameters:
//
// Returns:
//   - string: The canonical representation of the timezone offset formatted either as 'Z' or '±hh:mm'.
func (t TimezoneOffset) String() string {
	// Spec Rule: W3C XSD 1.1 Part 2 (Appendix E.2)
	// The canonical representation of a zero timezone offset is "Z". Non-zero offsets must be formatted as "+hh:mm" or
	// "-hh:mm".
	if t.offset == 0 {
		return "Z"
	}
	sign := "+"
	offset := t.offset
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	hours := offset / minutesPerHour
	minutes := offset % minutesPerHour
	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}

// ToBEBytes serializes the internal timezone offset to a 2-byte big-endian byte array.
//
// Parameters:
//
// Returns:
//   - [2]byte: A 2-byte array containing the big-endian representation of the offset in minutes.
func (t TimezoneOffset) ToBEBytes() [2]byte {
	// Implementation Note: Bounds-checked two's complement conversion
	// Convert signed int16 offset into unsigned uint16 without triggering integer overflow warnings.
	var u uint16
	if t.offset >= 0 {
		u = uint16(t.offset)
	} else {
		u = ^uint16(-t.offset) + 1
	}
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], u)
	return b
}

// TimezoneOffsetFromBEBytes deserializes a 2-byte big-endian byte array back into a TimezoneOffset.
//
// Parameters:
//   - b: A 2-byte array containing the big-endian representation of the offset in minutes.
//
// Returns:
//   - TimezoneOffset: The deserialized TimezoneOffset value.
func TimezoneOffsetFromBEBytes(b [2]byte) TimezoneOffset {
	// Implementation Note: Bounds-checked two's complement conversion
	// Convert big-endian uint16 back into signed int16 without triggering integer overflow warnings.
	u := binary.BigEndian.Uint16(b[:])
	var offset int16
	if u <= math.MaxInt16 {
		offset = int16(u)
	} else {
		pos := ^u + 1
		if pos <= math.MaxInt16 {
			offset = -int16(pos)
		}
	}
	return TimezoneOffset{offset: offset}
}
