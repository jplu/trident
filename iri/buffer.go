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

package iri

import "strings"

// outputBuffer represents the builder interface for constructing parsed IRI/URI representations.
//
// Representation:
// An abstraction over physical buffers that allows the parser to operate either in generating mode (allocating strings)
// or in validation-only mode (tracking bytes without allocation).
type outputBuffer interface {
	// writeRune appends a single Unicode code point to the buffer.
	//
	// Parameters:
	//   - r: The Unicode rune to write.
	writeRune(r rune)

	// writeString appends a string to the buffer.
	//
	// Parameters:
	//   - s: The string of characters to write.
	writeString(s string)

	// string returns the accumulated content of the buffer.
	//
	// Returns:
	//   - string: The fully built string.
	string() string

	// len returns the current length in bytes.
	//
	// Returns:
	//   - int: The total byte length.
	len() int

	// truncate reduces the buffer length to a target size.
	//
	// Parameters:
	//   - n: The target byte length.
	truncate(n int)

	// reset clears the state of the buffer.
	reset()
}

// voidOutputBuffer represents the zero-allocation output buffer.
//
// Representation:
// A structure that discards all written characters and only updates an internal counter to track the logical byte
// length of the would-be output.
type voidOutputBuffer struct {
	length int
}

// writeRune tracks the byte length of the rune that would have been written.
//
// Parameters:
//   - r: The rune to evaluate for byte length.
func (b *voidOutputBuffer) writeRune(r rune) { b.length += len(string(r)) }

// writeString tracks the byte length of the string that would have been written.
//
// Parameters:
//   - s: The string to evaluate for byte length.
func (b *voidOutputBuffer) writeString(s string) { b.length += len(s) }

// string returns an empty string because no characters are retained.
//
// Returns:
//   - string: An empty string.
func (b *voidOutputBuffer) string() string { return "" }

// len returns the total byte length that would have been written.
//
// Returns:
//   - int: The accumulated byte count.
func (b *voidOutputBuffer) len() int { return b.length }

// truncate updates the logical accumulated byte length.
//
// Parameters:
//   - n: The target byte length. If n is out of valid bounds, the operation has no effect.
func (b *voidOutputBuffer) truncate(n int) {
	// Implementation Note: Bounds checking for logical truncation
	// Slicing or resizing operations require validation to ensure the target offset remains within valid ranges.
	if n < 0 || n > b.length {
		return
	}
	b.length = n
}

// reset sets the tracked byte length back to zero.
func (b *voidOutputBuffer) reset() { b.length = 0 }

// stringOutputBuffer represents the memory-backed output buffer.
//
// Representation:
// A wrapper around strings.Builder that accumulates parsed runes and strings to construct the final output IRI/URI
// reference.
type stringOutputBuffer struct {
	builder *strings.Builder
}

// writeRune writes a single rune to the underlying strings.Builder.
//
// Parameters:
//   - r: The rune to append to the builder.
func (b *stringOutputBuffer) writeRune(r rune) { b.builder.WriteRune(r) }

// writeString writes a string to the underlying strings.Builder.
//
// Parameters:
//   - s: The string to append to the builder.
func (b *stringOutputBuffer) writeString(s string) { b.builder.WriteString(s) }

// string retrieves the accumulated content as a single string.
//
// Returns:
//   - string: The fully constructed string.
func (b *stringOutputBuffer) string() string { return b.builder.String() }

// len retrieves the current byte length of the accumulated buffer.
//
// Returns:
//   - int: The number of accumulated bytes.
func (b *stringOutputBuffer) len() int { return b.builder.Len() }

// truncate reslices the accumulated content to the target byte length.
//
// Parameters:
//   - n: The target byte length. If n is out of valid bounds, the operation has no effect.
func (b *stringOutputBuffer) truncate(n int) {
	// Implementation Note: Slicing restriction on builders
	// Reslicing a strings.Builder requires fetching the accumulated string, slicing it, and re-initializing the builder
	// to avoid unallocated writes.
	if n < 0 || n > b.builder.Len() {
		return
	}
	s := b.builder.String()[:n]
	b.builder.Reset()
	b.builder.WriteString(s)
}

// reset clears all data from the underlying builder.
func (b *stringOutputBuffer) reset() { b.builder.Reset() }
