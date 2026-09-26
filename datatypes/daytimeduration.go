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

// DayTimeDuration represents the dayTimeDuration datatype as defined in the governing specification.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
// W3C RDF 1.2 (Section 5.1)
//
// Representation:
// A duration of time containing only day, hour, minute, and second components, mathematically modeled as a
// high-precision decimal number of total seconds.
type DayTimeDuration struct {
	// seconds represents the total duration value stored in seconds with decimal precision.
	seconds Decimal
}

// NewDayTimeDuration constructs a dayTimeDuration value from a Decimal value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//   - seconds: A Decimal value representing the total duration in seconds.
//
// Returns:
//   - DayTimeDuration: The constructed DayTimeDuration instance.
func NewDayTimeDuration(seconds Decimal) DayTimeDuration {
	return DayTimeDuration{seconds: seconds}
}

// DefaultDayTimeDuration returns a zero-valued dayTimeDuration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//
// Returns:
//   - DayTimeDuration: A DayTimeDuration instance representing zero seconds.
func DefaultDayTimeDuration() DayTimeDuration {
	return DayTimeDuration{seconds: DefaultDecimal()}
}

// NewDayTimeDurationFromStd converts a standard Go time.Duration into a DayTimeDuration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//   - stdDur: The standard library time.Duration to be converted.
//
// Returns:
//   - DayTimeDuration: The converted DayTimeDuration instance.
func NewDayTimeDurationFromStd(stdDur time.Duration) DayTimeDuration {
	nanos := big.NewInt(stdDur.Nanoseconds())
	dec := NewDecimal(nanos, nanosecondsScale)
	return DayTimeDuration{seconds: dec}
}

// Days extracts the day component of the duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//
// Returns:
//   - int64: The integer number of days contained in the duration.
func (d DayTimeDuration) Days() int64 {
	return new(big.Int).Quo(d.seconds.asI128(), big.NewInt(secondsPerDay)).Int64()
}

// Hours extracts the hour component of the duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//
// Returns:
//   - int64: The integer hour component, bounded from 0 to 23.
func (d DayTimeDuration) Hours() int64 {
	rem := new(big.Int).Rem(d.seconds.asI128(), big.NewInt(secondsPerDay))
	return new(big.Int).Quo(rem, big.NewInt(secondsPerHour)).Int64()
}

// Minutes extracts the minute component of the duration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//
// Returns:
//   - int64: The integer minute component, bounded from 0 to 59.
func (d DayTimeDuration) Minutes() int64 {
	rem := new(big.Int).Rem(d.seconds.asI128(), big.NewInt(secondsPerHour))
	return new(big.Int).Quo(rem, big.NewInt(secondsPerMinute)).Int64()
}

// Seconds extracts the second component of the duration, including fractional seconds.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//
// Returns:
//   - Decimal: The decimal second component of the duration.
func (d DayTimeDuration) Seconds() Decimal {
	res, _ := d.seconds.CheckedRemEuclid(NewDecimalFromInt64(secondsPerMinute))
	return res
}

// AsSeconds returns the total duration represented strictly as a Decimal of seconds.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//
// Returns:
//   - Decimal: The total duration in seconds.
func (d DayTimeDuration) AsSeconds() Decimal {
	return d.seconds
}

// Add performs addition of two DayTimeDuration values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//   - rhs: The DayTimeDuration to add to the receiver.
//
// Returns:
//   - DayTimeDuration: The resulting sum of the two durations.
func (d DayTimeDuration) Add(rhs DayTimeDuration) DayTimeDuration {
	return DayTimeDuration{seconds: d.seconds.Add(rhs.seconds)}
}

// Sub performs subtraction of two DayTimeDuration values.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//   - rhs: The DayTimeDuration to subtract from the receiver.
//
// Returns:
//   - DayTimeDuration: The resulting difference between the two durations.
func (d DayTimeDuration) Sub(rhs DayTimeDuration) DayTimeDuration {
	return DayTimeDuration{seconds: d.seconds.Sub(rhs.seconds)}
}

// Neg negates the DayTimeDuration value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//
// Returns:
//   - DayTimeDuration: The negated DayTimeDuration.
func (d DayTimeDuration) Neg() DayTimeDuration {
	return DayTimeDuration{seconds: d.seconds.Neg()}
}

// IsIdenticalWith checks if this DayTimeDuration is identical to another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27.1)
//
// Parameters:
//   - other: The other XSDValue to compare for identity.
//
// Returns:
//   - bool: True if both values are identical dayTimeDuration instances, false otherwise.
func (d DayTimeDuration) IsIdenticalWith(other XSDValue) bool {
	if o, ok := other.(DayTimeDuration); ok {
		return d.seconds.IsIdenticalWith(o.seconds)
	}
	return false
}

// Compare evaluates the order relation of this DayTimeDuration against another XSDValue.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//   - other: The other XSDValue to compare against.
//
// Returns:
//   - int: Negative if the receiver is less, positive if greater, zero if equal.
//   - error: An error if the types are incomparable or comparison overflows.
func (d DayTimeDuration) Compare(other XSDValue) (int, error) {
	if o, ok := other.(DayTimeDuration); ok {
		return d.seconds.Compare(o.seconds)
	}
	return 0, ErrDurationOverflow
}

// String returns the canonical lexical representation of the DayTimeDuration value.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27.1)
//
// Parameters:
//
// Returns:
//   - string: The canonical lexical representation matching dayTimeDurationLexicalRep.
func (d DayTimeDuration) String() string {
	if d.seconds.getValue().Sign() == 0 {
		return "PT0S"
	}
	dur, _ := NewDuration(0, d.seconds)
	return dur.String()
}

// ParseDayTimeDuration parses a string literal matching the lexical representation of xsd:dayTimeDuration.
//
// Specification Reference:
// W3C XSD 1.1 Part 2 (Section 3.4.27)
//
// Parameters:
//   - s: The raw string literal to be parsed.
//
// Returns:
//   - DayTimeDuration: The parsed DayTimeDuration instance.
//   - error: An error of type ParseDurationError if lexical parsing or validation fails.
func ParseDayTimeDuration(s string) (DayTimeDuration, error) {
	parts, _, err := durationParts(s)
	if err != nil {
		return DayTimeDuration{}, err
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.27.1)
	// No year or month components are permitted in the lexical representation of dayTimeDuration.
	if parts.yearMonth != nil {
		return DayTimeDuration{}, newParseDurationError(
			"there must not be any year or month component in a dayTimeDuration",
		)
	}

	// Spec Rule: W3C XSD 1.1 Part 2 (Section 3.4.27.1)
	// At least one of the day, hour, minute, or second components must be present.
	if parts.dayTime == nil {
		return DayTimeDuration{}, newParseDurationError("no day or time values found")
	}
	return NewDayTimeDuration(*parts.dayTime), nil
}
