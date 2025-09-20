/*
Copyright 2025 Trident Authors

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

// outputBuffer is an interface for building the output string during parsing.
// This abstraction allows the parser to be used in different modes, such as
// full string generation (stringOutputBuffer) or validation-only without
// string allocation (voidOutputBuffer).
type outputBuffer interface {
	// writeRune appends a single rune to the buffer.
	writeRune(r rune)
	// writeString appends a string to the buffer.
	writeString(s string)
	// string returns the complete content of the buffer.
	string() string
	// len returns the number of bytes currently in the buffer.
	len() int
	// truncate reduces the buffer to n bytes.
	truncate(n int)
	// reset clears the buffer.
	reset()
}

// voidOutputBuffer is an implementation of outputBuffer that discards all
// writes and only tracks the length of the would-be output. It is useful for
// validation-only parsing where the final string is not needed, avoiding
// all allocations.
type voidOutputBuffer struct {
	length int
}

// writeRune tracks the length of the rune that would have been written.
func (b *voidOutputBuffer) writeRune(r rune) { b.length += len(string(r)) }

// writeString tracks the length of the string that would have been written.
func (b *voidOutputBuffer) writeString(s string) { b.length += len(s) }

// string returns an empty string, as no output is stored.
func (b *voidOutputBuffer) string() string { return "" }

// len returns the number of bytes that would have been written to the buffer.
func (b *voidOutputBuffer) len() int { return b.length }

// truncate sets the tracked length to n. If n is invalid (negative or
// greater than the current length), the operation is a no-op.
func (b *voidOutputBuffer) truncate(n int) {
	if n < 0 || n > b.length {
		return
	}
	b.length = n
}

// reset clears the tracked length by setting it to zero.
func (b *voidOutputBuffer) reset() { b.length = 0 }

// stringOutputBuffer is an implementation of outputBuffer that uses a
// strings.Builder to efficiently construct the output string.
type stringOutputBuffer struct {
	builder *strings.Builder
}

// writeRune appends a single rune to the underlying strings.Builder.
func (b *stringOutputBuffer) writeRune(r rune) { b.builder.WriteRune(r) }

// writeString appends a string to the underlying strings.Builder.
func (b *stringOutputBuffer) writeString(s string) { b.builder.WriteString(s) }

// string returns the complete content of the buffer as a string.
func (b *stringOutputBuffer) string() string { return b.builder.String() }

// len returns the number of bytes currently in the buffer.
func (b *stringOutputBuffer) len() int { return b.builder.Len() }

// truncate reduces the buffer to n bytes by reslicing the string content.
// If n is invalid, the buffer is not modified.
func (b *stringOutputBuffer) truncate(n int) {
	if n < 0 || n > b.builder.Len() {
		return
	}
	s := b.builder.String()[:n]
	b.builder.Reset()
	b.builder.WriteString(s)
}

// reset clears the underlying strings.Builder.
func (b *stringOutputBuffer) reset() { b.builder.Reset() }
