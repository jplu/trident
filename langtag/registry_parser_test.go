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

//nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to test unexported functions.
package langtag

import (
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

// errorReader is a helper type that implements io.Reader and always returns an error.
type errorReader struct{}

func (r errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("mock reader error")
}

// Test_buildRecord tests the buildRecord function, which is the lowest-level
// function for converting a map of fields into a Record struct.
// This is based on RFC 5646 Section 3.1.2, which defines the fields within a record.
func Test_buildRecord(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string][]string
		want   Record
	}{
		{
			name:   "empty fields map should produce an empty record",
			fields: map[string][]string{},
			want:   Record{},
		},
		{
			name: "minimal record with required fields",
			fields: map[string][]string{
				"type":        {"language"},
				"subtag":      {"en"},
				"description": {"English"},
				"added":       {"2005-10-16"},
			},
			want: Record{
				Type:        "language",
				Subtag:      "en",
				Description: []string{"English"},
				Added:       "2005-10-16",
			},
		},
		{
			name: "full record with all possible fields and multiple values",
			fields: map[string][]string{
				"type":            {"variant"},
				"subtag":          {"1996"},
				"description":     {"German orthography reform of 1996", "Second description"},
				"prefix":          {"de", "sl"},
				"added":           {"2005-10-16"},
				"deprecated":      {"2020-01-01"},
				"preferred-value": {"new-val"},
				"suppress-script": {"Latn"},
				"macrolanguage":   {"zh"},
				"scope":           {"special"},
				"comments":        {"A comment", "Another comment"},
			},
			want: Record{
				Type:           "variant",
				Subtag:         "1996",
				Description:    []string{"German orthography reform of 1996", "Second description"},
				Prefix:         []string{"de", "sl"},
				Added:          "2005-10-16",
				Deprecated:     "2020-01-01",
				PreferredValue: "new-val",
				SuppressScript: "Latn",
				Macrolanguage:  "zh",
				Scope:          "special",
				Comments:       []string{"A comment", "Another comment"},
			},
		},
		{
			name: "record for a full tag (grandfathered)",
			fields: map[string][]string{
				"type":        {"grandfathered"},
				"tag":         {"i-klingon"},
				"description": {"Klingon"},
				"added":       {"1996-09-17"},
			},
			want: Record{
				Type:        "grandfathered",
				Tag:         "i-klingon",
				Description: []string{"Klingon"},
				Added:       "1996-09-17",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildRecord(tt.fields); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildRecord() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Test_expandNumericRange tests the expansion of numeric ranges.
// RFC 5646 Section 3.1.1 states: "'11..13' denotes the values '11', '12', and '13'".
func Test_expandNumericRange(t *testing.T) {
	tests := []struct {
		name    string
		start   string
		end     string
		want    []string
		wantErr bool
	}{
		{
			name:  "rfc example 11..13",
			start: "11",
			end:   "13",
			want:  []string{"11", "12", "13"},
		},
		{
			name:  "padded range 001..003",
			start: "001",
			end:   "003",
			want:  []string{"001", "002", "003"},
		},
		{
			name:  "single element range",
			start: "42",
			end:   "42",
			want:  []string{"42"},
		},
		{
			name:    "invalid start > end",
			start:   "13",
			end:     "11",
			wantErr: true,
		},
		{
			name:    "invalid start is not numeric",
			start:   "a1",
			end:     "13",
			wantErr: true,
		},
		{
			name:    "invalid end is not numeric",
			start:   "11",
			end:     "b3",
			wantErr: true,
		},
		{
			name:    "range too large",
			start:   "0",
			end:     "20001", // maxNumericExpansion is 20000
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandNumericRange(tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandNumericRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandNumericRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_expandAlphabeticRange tests the expansion of alphabetic ranges.
// RFC 5646 Section 3.1.1 states: "'a..c' denotes the values 'a', 'b', and 'c'".
// RFC 5646 Section 2.2.1 provides another example: "'qaa' through 'qtz'".
func Test_expandAlphabeticRange(t *testing.T) {
	tests := []struct {
		name    string
		start   string
		end     string
		want    []string
		wantErr bool
	}{
		{
			name:  "rfc example a..c",
			start: "a",
			end:   "c",
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "rfc example qaa..qac (partial)",
			start: "qaa",
			end:   "qac",
			want:  []string{"qaa", "qab", "qac"},
		},
		{
			name:  "rollover behavior",
			start: "az",
			end:   "bc",
			want:  []string{"az", "ba", "bb", "bc"},
		},
		{
			name:  "single element range",
			start: "b",
			end:   "b",
			want:  []string{"b"},
		},
		{
			name:  "case is normalized to lower",
			start: "A",
			end:   "C",
			want:  []string{"a", "b", "c"},
		},
		{
			name:    "invalid start > end",
			start:   "c",
			end:     "a",
			wantErr: true,
		},
		{
			name:    "range too large",
			start:   "aaaa",
			end:     "zzzz", // Exceeds maxAlphaExpansion
			wantErr: true,
		},
		{
			name:  "complex rollover to test inner loop break",
			start: "ayz",
			end:   "aza",
			want:  []string{"ayz", "aza"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandAlphabeticRange(tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandAlphabeticRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandAlphabeticRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_expandRange tests the main range dispatcher.
// RFC 5646 Section 3.1.1 defines range notation and this function validates the format.
func Test_expandRange(t *testing.T) {
	tests := []struct {
		name     string
		rangeStr string
		want     []string
		wantErr  bool
	}{
		{
			name:     "valid numeric range",
			rangeStr: "005..007",
			want:     []string{"005", "006", "007"},
		},
		{
			name:     "valid alphabetic range",
			rangeStr: "ca..ce",
			want:     []string{"ca", "cb", "cc", "cd", "ce"},
		},
		{
			name:     "invalid format - too many parts",
			rangeStr: "a..b..c",
			wantErr:  true,
		},
		{
			name:     "invalid format - missing end",
			rangeStr: "a..",
			wantErr:  true,
		},
		{
			name:     "invalid format - missing start",
			rangeStr: "..b",
			wantErr:  true,
		},
		{
			name:     "invalid format - empty string",
			rangeStr: "",
			wantErr:  true,
		},
		{
			name:     "invalid format - unequal length",
			rangeStr: "a..bb",
			wantErr:  true,
		},
		{
			name:     "invalid format - zero length parts",
			rangeStr: "..",
			wantErr:  true,
		},
		{
			name:     "mixed type - alpha and numeric",
			rangeStr: "a..1",
			wantErr:  true,
		},
		{
			name:     "error from numeric expander",
			rangeStr: "3..1",
			wantErr:  true,
		},
		{
			name:     "error from alphabetic expander",
			rangeStr: "c..a",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandRange(tt.rangeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expandRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_processAndAddRecord tests the processing of a single parsed Record.
// It verifies that records are added correctly, ranges are expanded, and errors are handled.
// It also verifies the use of type-prefixed keys to avoid subtag collisions, e.g.,
// a language 'foo' and a variant 'foo' would occupy different keys ('language:foo', 'variant:foo').
func Test_processAndAddRecord(t *testing.T) {
	newTestRegistry := func() *Registry {
		return &Registry{Records: make(map[string]Record)}
	}

	tests := []struct {
		name         string
		record       Record
		wantRegistry *Registry
		wantErr      bool
	}{
		{
			name: "add single subtag record with type-prefixed key",
			record: Record{
				Type:   "language",
				Subtag: "en",
			},
			wantRegistry: &Registry{Records: map[string]Record{
				"language:en": {Type: "language", Subtag: "en"},
			}},
		},
		{
			name: "add single tag record (grandfathered)",
			record: Record{
				Type: "grandfathered",
				Tag:  "i-klingon",
			},
			wantRegistry: &Registry{Records: map[string]Record{
				"i-klingon": {Type: "grandfathered", Tag: "i-klingon"},
			}},
		},
		{
			name: "add expanded subtag range with type-prefixed keys",
			record: Record{
				Type:   "region",
				Subtag: "001..003",
			},
			wantRegistry: &Registry{Records: map[string]Record{
				"region:001": {Type: "region", Subtag: "001"},
				"region:002": {Type: "region", Subtag: "002"},
				"region:003": {Type: "region", Subtag: "003"},
			}},
		},
		{
			name: "add expanded tag range",
			record: Record{
				Type: "private-use",
				Tag:  "qaa..qac",
			},
			wantRegistry: &Registry{Records: map[string]Record{
				"qaa": {Type: "private-use", Tag: "qaa"},
				"qab": {Type: "private-use", Tag: "qab"},
				"qac": {Type: "private-use", Tag: "qac"},
			}},
		},
		{
			name: "key is lowercased and type-prefixed",
			record: Record{
				Type:   "language",
				Subtag: "EN",
			},
			wantRegistry: &Registry{Records: map[string]Record{
				"language:en": {Type: "language", Subtag: "EN"},
			}},
		},
		{
			name: "invalid subtag range",
			record: Record{
				Type:   "region",
				Subtag: "3..1",
			},
			wantRegistry: newTestRegistry(),
			wantErr:      true,
		},
		{
			name: "invalid tag range",
			record: Record{
				Type: "grandfathered",
				Tag:  "z..x",
			},
			wantRegistry: newTestRegistry(),
			wantErr:      true,
		},
		{
			name:         "record with no key (no subtag or tag)",
			record:       Record{Type: "language"},
			wantRegistry: newTestRegistry(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := newTestRegistry()
			err := processAndAddRecord(registry, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("processAndAddRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(registry, tt.wantRegistry) {
				t.Errorf("processAndAddRecord() registry = %+v, want %+v", registry, tt.wantRegistry)
			}
		})
	}
}

// Test_addRecordFromFields tests the wrapper that combines buildRecord and processAndAddRecord.
func Test_addRecordFromFields(t *testing.T) {
	newTestRegistry := func() *Registry {
		return &Registry{Records: make(map[string]Record)}
	}

	tests := []struct {
		name         string
		fields       map[string][]string
		wantRegistry *Registry
		wantErr      bool
	}{
		{
			name:         "empty fields map does nothing",
			fields:       map[string][]string{},
			wantRegistry: newTestRegistry(),
		},
		{
			name: "valid fields are built and added with type-prefixed key",
			fields: map[string][]string{
				"type":   {"language"},
				"subtag": {"de"},
			},
			wantRegistry: &Registry{Records: map[string]Record{
				"language:de": {Type: "language", Subtag: "de"},
			}},
		},
		{
			name: "error from processing is propagated",
			fields: map[string][]string{
				"type":   {"language"},
				"subtag": {"z..a"}, // invalid range
			},
			wantRegistry: newTestRegistry(),
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := newTestRegistry()
			err := addRecordFromFields(registry, tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("addRecordFromFields() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(registry, tt.wantRegistry) {
				t.Errorf("addRecordFromFields() registry = %+v, want %+v", registry, tt.wantRegistry)
			}
		})
	}
}

type processLineTestCase struct {
	name             string
	lines            []string
	wantErr          bool
	wantFileDate     string
	wantRecordsCount int
	wantFinalFields  map[string][]string
	wantLastField    string
}

func assertParserState(t *testing.T, p *registryParser, tt processLineTestCase) {
	t.Helper()
	if tt.wantFileDate != "" && p.registry.FileDate != tt.wantFileDate {
		t.Errorf("parser.registry.FileDate = %q, want %q", p.registry.FileDate, tt.wantFileDate)
	}
	if tt.wantRecordsCount != len(p.registry.Records) {
		t.Errorf("len(parser.registry.Records) = %d, want %d", len(p.registry.Records), tt.wantRecordsCount)
	}
	if !reflect.DeepEqual(p.currentFields, tt.wantFinalFields) {
		t.Errorf("parser.currentFields = %v, want %v", p.currentFields, tt.wantFinalFields)
	}
	if p.lastFieldName != tt.wantLastField {
		t.Errorf("parser.lastFieldName = %q, want %q", p.lastFieldName, tt.wantLastField)
	}
}

// Test_registryParser_processLine tests the line-by-line state machine of the parser.
// This is based on RFC 5646 Section 3.1.1 which describes the record-jar format,
// including field folding and record separators.
func Test_registryParser_processLine(t *testing.T) {
	newTestParser := func() *registryParser {
		return &registryParser{
			registry:      &Registry{Records: make(map[string]Record)},
			currentFields: make(map[string][]string),
		}
	}

	tests := []processLineTestCase{
		{
			name:            "file date sets registry field",
			lines:           []string{"File-Date: 2024-07-29"},
			wantFileDate:    "2024-07-29",
			wantFinalFields: map[string][]string{},
		},
		{
			name:             "file date ignored after first record",
			lines:            []string{"Type: language", "Subtag: en", "%%", "File-Date: 2024-07-29"},
			wantRecordsCount: 1,
			wantFileDate:     "",
			wantFinalFields:  map[string][]string{"file-date": {"2024-07-29"}},
			wantLastField:    "file-date",
		},
		{
			name:            "simple field adds to currentFields",
			lines:           []string{"Type: language"},
			wantFinalFields: map[string][]string{"type": {"language"}},
			wantLastField:   "type",
		},
		{
			name:            "folded line appends to last field",
			lines:           []string{"Description: A long description", "  that continues."},
			wantFinalFields: map[string][]string{"description": {"A long description that continues."}},
			wantLastField:   "description",
		},
		{
			name:            "folded line with no last field is ignored",
			lines:           []string{"  this should be ignored"},
			wantFinalFields: map[string][]string{},
		},
		{
			name:            "malformed line with no colon is ignored",
			lines:           []string{"this is malformed"},
			wantFinalFields: map[string][]string{},
		},
		{
			name:             "record separator %% adds record and clears state",
			lines:            []string{"Type: language", "Subtag: fr", "%%"},
			wantRecordsCount: 1,
			wantFinalFields:  map[string][]string{},
			wantLastField:    "",
		},
		{
			name:    "error on %% is propagated",
			lines:   []string{"Subtag: a..z..b", "%%"}, // bad range
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestParser()
			var err error
			for _, line := range tt.lines {
				err = p.processLine(line)
				if err != nil {
					break
				}
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("processLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			assertParserState(t, p, tt)
		})
	}
}

// parseRegistryTestCase holds the data for a single Test_ParseRegistry case.
type parseRegistryTestCase struct {
	name              string
	reader            io.Reader
	wantRecordCount   int
	wantFileDate      string
	wantSpecificCheck map[string]Record
	wantErr           bool
}

// checkParseRegistryResult contains the assertion logic for Test_ParseRegistry.
func checkParseRegistryResult(t *testing.T, got *Registry, want parseRegistryTestCase) {
	t.Helper()
	if got.FileDate != want.wantFileDate {
		t.Errorf("ParseRegistry() FileDate = %q, want %q", got.FileDate, want.wantFileDate)
	}
	if len(got.Records) != want.wantRecordCount {
		t.Errorf("ParseRegistry() record count = %d, want %d", len(got.Records), want.wantRecordCount)
	}

	for key, wantRecord := range want.wantSpecificCheck {
		gotRecord, ok := got.Records[key]
		if !ok {
			t.Errorf("ParseRegistry() did not find record with key %q", key)
			continue
		}
		if !reflect.DeepEqual(gotRecord, wantRecord) {
			t.Errorf("ParseRegistry() record %q = %+v, want %+v", key, gotRecord, wantRecord)
		}
	}
}

// Test_ParseRegistry tests the top-level registry parsing function.
// It verifies parsing of a complete registry file as described in RFC 5646 Section 3.1.
func Test_ParseRegistry(t *testing.T) {
	validRegistryContent := `File-Date: 2004-06-28
%%
Type: language
Subtag: de
Description: German
Added: 2005-10-16
Suppress-Script: Latn
%%
Type: script
Subtag: Latn
Description: Latin
Added: 2005-10-16
%%
Type: region
Subtag: qm..qz
Description: Private use
Added: 2005-10-16
%%
Type: grandfathered
Tag: i-klingon
Description: Klingon
  (folded description)
Added: 1996-09-17
Preferred-Value: tlh
` // No final %% to test handling of the last record in a file.

	tests := []parseRegistryTestCase{
		{
			name:   "valid registry parsing with ranges and folding",
			reader: strings.NewReader(validRegistryContent),
			// qm..qz expands to 14 records + de + Latn + i-klingon = 17 records
			// Keys will be: 'language:de', 'script:latn', 'region:qm'...'region:qz', 'i-klingon'
			wantRecordCount: 17,
			wantFileDate:    "2004-06-28",
			wantSpecificCheck: map[string]Record{
				"i-klingon": {
					Type:           "grandfathered",
					Tag:            "i-klingon",
					Description:    []string{"Klingon (folded description)"},
					Added:          "1996-09-17",
					PreferredValue: "tlh",
				},
			},
		},
		{
			name:            "empty registry",
			reader:          strings.NewReader(""),
			wantRecordCount: 0,
		},
		{
			name:         "registry with only file date",
			reader:       strings.NewReader("File-Date: 2024-01-01"),
			wantFileDate: "2024-01-01",
		},
		{
			name:    "reader error",
			reader:  errorReader{},
			wantErr: true,
		},
		{
			name:    "malformed content (bad range)",
			reader:  strings.NewReader("Type: region\nSubtag: 3..1\n%%"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRegistry(tt.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			checkParseRegistryResult(t, got, tt)
		})
	}
}
