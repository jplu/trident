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

package langtag

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	keyValParts         = 2
	rangeParts          = 2
	maxNumericExpansion = 20000
	maxAlphaExpansion   = 40000
)

// Registry represents the parsed data from the IANA Language Subtag Registry file.
//
// Specification Reference:
// RFC 5646 (Section 3.1)
//
// Representation:
// A mapping database containing the registration records of various subtag categories and grandfathered tags indexed by
// their category and unique identifier, alongside the registry file's release date.
type Registry struct {
	Records  map[string]Record
	FileDate string
}

// Record represents a single entry in the IANA Language Subtag Registry.
//
// Specification Reference:
// RFC 5646 (Section 3.1.2)
//
// Representation:
// A structured element aggregating registration attributes such as type, subtag or tag name, preferred-value mappings,
// prefix directives, and deprecation markers.
type Record struct {
	Type           string   `json:"type"`
	Subtag         string   `json:"subtag,omitempty"`
	Tag            string   `json:"tag,omitempty"`
	Description    []string `json:"description"`
	Added          string   `json:"added"`
	Deprecated     string   `json:"deprecated,omitempty"`
	PreferredValue string   `json:"preferredValue,omitempty"`
	Prefix         []string `json:"prefix,omitempty"`
	SuppressScript string   `json:"suppressScript,omitempty"`
	Macrolanguage  string   `json:"macrolanguage,omitempty"`
	Scope          string   `json:"scope,omitempty"`
	Comments       []string `json:"comments,omitempty"`
}

// IsGrandfathered checks whether the record's type indicates it is grandfathered or redundant.
//
// Specification Reference:
// RFC 5646 (Section 2.2.8)
//
// Parameters:
//
// Returns:
//   - bool: true if the record matches grandfathered or redundant categories, false otherwise.
func (r *Record) IsGrandfathered() bool {
	return r.Type == "grandfathered" || r.Type == "redundant"
}

// ParseRegistry parses an IANA Language Subtag Registry file from the given reader and returns a populated Registry
// object.
//
// Specification Reference:
// RFC 5646 (Section 3.1.1)
//
// Parameters:
//   - r: io.Reader representing the raw stream of the registry document.
//
// Returns:
//   - *Registry: a pointer to the populated subtag registry database under successful execution.
//   - error: an error describing formatting, tokenizing, or sequence generation failures during parse.
func ParseRegistry(r io.Reader) (*Registry, error) {
	scanner := bufio.NewScanner(r)
	p := &registryParser{
		registry: &Registry{
			Records: make(map[string]Record),
		},
		currentFields: make(map[string][]string),
	}

	for scanner.Scan() {
		if err := p.processLine(scanner.Text()); err != nil {
			return nil, err
		}
	}

	if err := addRecordFromFields(p.registry, p.currentFields); err != nil {
		return nil, err
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return p.registry, nil
}

// registryParser represents the state machine context for parsing an IANA subtag registry data stream.
//
// Specification Reference:
// RFC 5646 (Section 3.1.1)
//
// Representation:
// A stateful parser unit buffering partially parsed field-value sets and holding the last observed key to support
// line-folding reconstruction.
type registryParser struct {
	registry      *Registry
	currentFields map[string][]string
	lastFieldName string
}

// processLine processes a single line from the registry file.
//
// Specification Reference:
// RFC 5646 (Section 3.1.1)
//
// Parameters:
//   - line: string containing the current unparsed text line.
//
// Returns:
//   - error: any semantic parsing error encountered when handling record transitions.
func (p *registryParser) processLine(line string) error {
	// Spec Rule: RFC 5646 (Section 3.1.1)
	// Records are separated by lines containing only the sequence "%%" (U+0025 U+0025).
	if line == "%%" {
		if err := addRecordFromFields(p.registry, p.currentFields); err != nil {
			return err
		}
		p.currentFields = make(map[string][]string)
		p.lastFieldName = ""
		return nil
	}

	// Spec Rule: RFC 5646 (Section 3.1.1)
	// A field-body can be split into a multiple-line representation; this is called "folding".
	if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
		if p.lastFieldName != "" && len(p.currentFields[p.lastFieldName]) > 0 {
			lastIdx := len(p.currentFields[p.lastFieldName]) - 1
			p.currentFields[p.lastFieldName][lastIdx] += " " + strings.TrimSpace(line)
		}
		return nil
	}

	// Spec Rule: RFC 5646 (Section 3.1.1)
	// Each field contains a "field-name" and a "field-body", separated by a COLON character (U+003A).
	parts := strings.SplitN(line, ":", keyValParts)
	if len(parts) != keyValParts {
		return nil
	}

	fieldName, fieldBody := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	if strings.EqualFold(fieldName, "File-Date") && len(p.registry.Records) == 0 {
		p.registry.FileDate = fieldBody
		return nil
	}

	fieldNameLower := strings.ToLower(fieldName)
	p.currentFields[fieldNameLower] = append(p.currentFields[fieldNameLower], fieldBody)
	p.lastFieldName = fieldNameLower
	return nil
}

// addRecordFromFields builds a record from the collected fields and adds it to the registry.
//
// Specification Reference:
// RFC 5646 (Section 3.1.2)
//
// Parameters:
//   - registry: *Registry database receiving the constructed entry.
//   - fields: map[string][]string containing field-to-raw-value assignments.
//
// Returns:
//   - error: any generation or indexing validation failure.
func addRecordFromFields(registry *Registry, fields map[string][]string) error {
	if len(fields) == 0 {
		return nil
	}
	record := buildRecord(fields)
	return processAndAddRecord(registry, record)
}

// processAndAddRecord handles a parsed record, expanding ranges if necessary, and adds the resulting record(s) to the
// registry.
//
// Specification Reference:
// RFC 5646 (Section 3.1.1)
//
// Parameters:
//   - registry: *Registry to update with parsed database values.
//   - record: Record describing the currently parsed registry elements.
//
// Returns:
//   - error: any error generated by range-expansion math or collision rules.
func processAndAddRecord(registry *Registry, record Record) error {
	switch {
	// Spec Rule: RFC 5646 (Section 3.1.1)
	// The sequence '..' (U+002E U+002E) in a field-body denotes a range of values.
	case strings.Contains(record.Subtag, ".."):
		subtags, err := expandRange(record.Subtag)
		if err != nil {
			return fmt.Errorf("failed to expand subtag range '%s': %w", record.Subtag, err)
		}
		for _, sub := range subtags {
			newRec := record
			newRec.Subtag = sub
			key := newRec.Type + ":" + strings.ToLower(newRec.Subtag)
			registry.Records[key] = newRec
		}
	// Spec Rule: RFC 5646 (Section 3.1.1)
	// Ranges can also occur within grandfathered tag declarations to denote series.
	case strings.Contains(record.Tag, ".."):
		tags, err := expandRange(record.Tag)
		if err != nil {
			return fmt.Errorf("failed to expand tag range '%s': %w", record.Tag, err)
		}
		for _, t := range tags {
			newRec := record
			newRec.Tag = t
			registry.Records[strings.ToLower(newRec.Tag)] = newRec
		}
	default:
		// Implementation Note: Registry database index formatting
		// The index key maps the record's type prefix and the lowercase subtag value
		// to guarantee unique and fast lookup across record namespaces.
		var key string
		if record.Subtag != "" {
			key = record.Type + ":" + strings.ToLower(record.Subtag)
		} else if record.Tag != "" {
			key = strings.ToLower(record.Tag)
		}

		if key != "" {
			registry.Records[key] = record
		}
	}
	return nil
}

// expandRange expands a subtag range into a slice of individual subtags.
//
// Specification Reference:
// RFC 5646 (Section 3.1.1)
//
// Parameters:
//   - rangeStr: string containing the starting and ending parameters joined by '..'.
//
// Returns:
//   - []string: a generated sequence of individual subtag values representing the complete range.
//   - error: any error representing non-matching boundary structures or values.
func expandRange(rangeStr string) ([]string, error) {
	// Step 1: Range Delimiter Split
	// Verify and break the range string into start and end bounds using the ".." delimiter.
	parts := strings.Split(rangeStr, "..")
	if len(parts) != rangeParts {
		return nil, fmt.Errorf("invalid range format: %s", rangeStr)
	}
	start, end := parts[0], parts[1]

	// Step 2: Boundary Length Match
	// Enforce identical, non-zero length requirements for range boundaries as dictated by the specification.
	if len(start) != len(end) || len(start) == 0 {
		return nil, fmt.Errorf("range start/end must have same, non-zero length: %s", rangeStr)
	}

	// Step 3: Range Classification
	// Determine the subtag type (numeric vs alphabetic) and route to specialized range calculators.
	if isNumeric(start) && isNumeric(end) {
		return expandNumericRange(start, end)
	}
	if isAlphabetic(start) && isAlphabetic(end) {
		return expandAlphabeticRange(start, end)
	}

	return nil, fmt.Errorf("range must be purely alphabetic or purely numeric: %s", rangeStr)
}

// expandNumericRange expands a numeric range (e.g., "001..003").
//
// Specification Reference:
// RFC 5646 (Section 3.1.1)
//
// Parameters:
//   - start: string representing the numerical base boundary.
//   - end: string representing the numerical ceiling boundary.
//
// Returns:
//   - []string: an array of numeric strings, padded with leading zeroes matching the boundary length.
//   - error: any conversion or excessive buffer expansion size failure.
func expandNumericRange(start, end string) ([]string, error) {
	// Step 1: Boundary Numerical Parsing
	// Convert the string-encoded upper and lower bounds into machine integers.
	startNum, err1 := strconv.Atoi(start)
	endNum, err2 := strconv.Atoi(end)
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("invalid numeric range: %s..%s", start, end)
	}

	// Step 2: Span Constraint Validation
	// Ensure the parsed range flows positively and does not exceed memory expansion safety thresholds.
	if startNum > endNum {
		return nil, fmt.Errorf("start of range cannot be greater than end: %s..%s", start, end)
	}
	// Implementation Note: Expansion buffer overflow protection
	// Restricting numeric range sizes to maxNumericExpansion avoids excessive heap allocations
	// from malicious or erroneous range notation.
	if endNum-startNum > maxNumericExpansion {
		return nil, fmt.Errorf("numeric range is too large to expand: %s..%s", start, end)
	}

	// Step 3: Sequence Formatting
	// Walk the integer sequence, re-formatting values with zero-padding to preserve initial string width.
	var result []string
	format := fmt.Sprintf("%%0%dd", len(start))
	for i := startNum; i <= endNum; i++ {
		result = append(result, fmt.Sprintf(format, i))
	}
	return result, nil
}

// expandAlphabeticRange expands an alphabetic range (e.g., "qaa..qtz").
//
// Specification Reference:
// RFC 5646 (Section 3.1.1)
//
// Parameters:
//   - start: string representing the lowercase letter-bound prefix.
//   - end: string representing the lowercase letter-bound suffix.
//
// Returns:
//   - []string: a sequence of lowercase alphanumeric codes traversing standard alphabetic orders.
//   - error: any lexicographical inconsistency or buffer safety limits.
func expandAlphabeticRange(start, end string) ([]string, error) {
	current := []byte(strings.ToLower(start))
	endBytes := []byte(strings.ToLower(end))

	// Step 1: Boundary Comparison
	// Ensure starting byte sequence lexicographically precedes or matches the termination bound.
	if bytes.Compare(current, endBytes) > 0 {
		return nil, fmt.Errorf(
			"start of alphabetic range cannot be greater than end: %s..%s",
			start,
			end,
		)
	}

	var result []string
	for {
		result = append(result, string(current))
		if bytes.Equal(current, endBytes) {
			break
		}
		// Step 2: Allocation Guard
		// Prevent heap exhaustion by bounding the generated character sequence length.
		if len(result) > maxAlphaExpansion {
			return nil, fmt.Errorf("alphabetic range is too large to expand: %s..%s", start, end)
		}

		// Step 3: Lexicographical Increment Logic
		// Progress sequentially from the rightmost byte of the sequence, handling 'z' carry-over by carrying back.
		i := len(current) - 1
		for {
			current[i]++
			if current[i] <= 'z' {
				break
			}
			current[i] = 'a'
			i--
		}
	}
	return result, nil
}

// buildRecord converts a map of raw field strings into a Record struct.
//
// Specification Reference:
// RFC 5646 (Section 3.1.2)
//
// Parameters:
//   - fields: map[string][]string containing field identifiers mapped to multi-line records.
//
// Returns:
//   - Record: the populated structure with mapped domain attributes.
func buildRecord(fields map[string][]string) Record {
	getString := func(key string) string {
		if v, ok := fields[key]; ok && len(v) > 0 {
			return v[0]
		}
		return ""
	}
	return Record{
		Description:    fields["description"],
		Prefix:         fields["prefix"],
		Comments:       fields["comments"],
		Type:           getString("type"),
		Subtag:         getString("subtag"),
		Tag:            getString("tag"),
		Added:          getString("added"),
		Deprecated:     getString("deprecated"),
		PreferredValue: getString("preferred-value"),
		SuppressScript: getString("suppress-script"),
		Macrolanguage:  getString("macrolanguage"),
		Scope:          getString("scope"),
	}
}
