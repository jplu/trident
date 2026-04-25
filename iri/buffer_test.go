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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package iri

import (
	"strings"
	"testing"
)

func TestVoidOutputBuffer_WriteOperations(t *testing.T) {
	b := &voidOutputBuffer{}
	if b.len() != 0 {
		t.Errorf("Initial length should be 0, got %d", b.len())
	}
	if b.string() != "" {
		t.Errorf("string() should always return empty string, got '%s'", b.string())
	}
	b.writeRune('a')
	if b.len() != 1 {
		t.Errorf("Length after writing 'a' should be 1, got %d", b.len())
	}
	b.writeRune('\u00e9')
	expectedLen := 1 + len(string('\u00e9'))
	if b.len() != expectedLen {
		t.Errorf("Length after writing 'é' should be %d, got %d", expectedLen, b.len())
	}
	b.writeString("http:")
	expectedLen += len("http:")
	if b.len() != expectedLen {
		t.Errorf("Length after writing 'http:' should be %d, got %d", expectedLen, b.len())
	}
	if b.string() != "" {
		t.Errorf("string() should still be empty after writes, got '%s'", b.string())
	}
}

func TestVoidOutputBuffer_Truncate(t *testing.T) {
	b := &voidOutputBuffer{}
	b.writeString("http://example.com/path")
	b.truncate(15)
	if b.len() != 15 {
		t.Errorf("Length after truncating to 15 should be 15, got %d", b.len())
	}
	b.truncate(0)
	if b.len() != 0 {
		t.Errorf("Length after truncating to 0 should be 0, got %d", b.len())
	}
	b.writeString("test")
	b.truncate(-1)
	if b.len() != 4 {
		t.Errorf("Truncating to negative value should be a no-op, length is %d", b.len())
	}
	b.truncate(5)
	if b.len() != 4 {
		t.Errorf("Truncating to a larger value should be a no-op, length is %d", b.len())
	}
}

func TestVoidOutputBuffer_Reset(t *testing.T) {
	b := &voidOutputBuffer{}
	b.writeString("some-initial-data")
	if b.len() == 0 {
		t.Fatal("Buffer length is 0 before reset, test setup is wrong.")
	}
	b.reset()
	if b.len() != 0 {
		t.Errorf("Length after reset should be 0, got %d", b.len())
	}
	if b.string() != "" {
		t.Errorf("string() after reset should be empty, got '%s'", b.string())
	}
}

func TestStringOutputBuffer_WriteOperations(t *testing.T) {
	b := &stringOutputBuffer{builder: &strings.Builder{}}
	if b.len() != 0 {
		t.Errorf("Initial length should be 0, got %d", b.len())
	}
	if b.string() != "" {
		t.Errorf("Initial string should be empty, got '%s'", b.string())
	}
	b.writeRune('a')
	if b.len() != 1 {
		t.Errorf("Length after writing 'a' should be 1, got %d", b.len())
	}
	if b.string() != "a" {
		t.Errorf("String after writing 'a' should be 'a', got '%s'", b.string())
	}
	b.writeRune('\u00e9')
	expectedStr := "a\u00e9"
	expectedLen := len(expectedStr)
	if b.len() != expectedLen {
		t.Errorf("Length after writing 'é' should be %d, got %d", expectedLen, b.len())
	}
	if b.string() != expectedStr {
		t.Errorf("String after writing 'é' should be '%s', got '%s'", expectedStr, b.string())
	}
	b.writeString("http:")
	expectedStr += "http:"
	expectedLen += len("http:")
	if b.len() != expectedLen {
		t.Errorf("Length after writing 'http:' should be %d, got %d", expectedLen, b.len())
	}
	if b.string() != expectedStr {
		t.Errorf("String after writing 'http:' should be '%s', got '%s'", expectedStr, b.string())
	}
}

func TestStringOutputBuffer_Truncate(t *testing.T) {
	b := &stringOutputBuffer{builder: &strings.Builder{}}
	b.writeString("scheme://user@host:123/path?query#fragment")
	b.truncate(14)
	if b.len() != 14 {
		t.Errorf("Length after truncating to 14 should be 14, got %d", b.len())
	}
	if b.string() != "scheme://user@" {
		t.Errorf("String after truncating should be 'scheme://user@', got '%s'", b.string())
	}
	b.truncate(0)
	if b.len() != 0 {
		t.Errorf("Length after truncating to 0 should be 0, got %d", b.len())
	}
	if b.string() != "" {
		t.Errorf("String after truncating to 0 should be empty, got '%s'", b.string())
	}
	b.writeString("test")
	b.truncate(-1)
	if b.len() != 4 || b.string() != "test" {
		t.Errorf("Truncating to negative value should be a no-op, got len %d, str '%s'", b.len(), b.string())
	}
	b.truncate(5)
	if b.len() != 4 || b.string() != "test" {
		t.Errorf("Truncating to a larger value should be a no-op, got len %d, str '%s'", b.len(), b.string())
	}
	b.truncate(4)
	if b.len() != 4 || b.string() != "test" {
		t.Errorf("Truncating to the same length should be a no-op, got len %d, str '%s'", b.len(), b.string())
	}
}

func TestStringOutputBuffer_Reset(t *testing.T) {
	b := &stringOutputBuffer{builder: &strings.Builder{}}
	b.writeString("some-initial-data")
	if b.len() == 0 {
		t.Fatal("Buffer length is 0 before reset, test setup is wrong.")
	}
	b.reset()
	if b.len() != 0 {
		t.Errorf("Length after reset should be 0, got %d", b.len())
	}
	if b.string() != "" {
		t.Errorf("String after reset should be empty, got '%s'", b.string())
	}
}
