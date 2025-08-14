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

// registryParser holds the state for parsing a registry file.
type registryParser struct {
	registry      *Registry
	currentFields map[string][]string
	lastFieldName string
}

// processLine handles a single line from the registry file.
func (p *registryParser) processLine(line string) error {
	if line == "%%" {
		if err := addRecordFromFields(p.registry, p.currentFields); err != nil {
			return err
		}
		p.currentFields = make(map[string][]string)
		p.lastFieldName = ""
		return nil
	}

	if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
		if p.lastFieldName != "" && len(p.currentFields[p.lastFieldName]) > 0 {
			lastIdx := len(p.currentFields[p.lastFieldName]) - 1
			p.currentFields[p.lastFieldName][lastIdx] += " " + strings.TrimSpace(line)
		}
		return nil
	}

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

// ParseRegistry reads an IANA Language Subtag Registry file from the given
// reader and returns a populated Registry object. It correctly handles
// range notation (e.g., "qaa..qtz").
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

// addRecordFromFields builds a record from the collected fields and adds it
// to the registry, handling ranges.
func addRecordFromFields(registry *Registry, fields map[string][]string) error {
	if len(fields) == 0 {
		return nil
	}
	record := buildRecord(fields)
	return processAndAddRecord(registry, record)
}

// processAndAddRecord handles a parsed record, expanding ranges if necessary,
// and adds the resulting record(s) to the registry.
func processAndAddRecord(registry *Registry, record Record) error {
	switch {
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
func expandRange(rangeStr string) ([]string, error) {
	parts := strings.Split(rangeStr, "..")
	if len(parts) != rangeParts {
		return nil, fmt.Errorf("invalid range format: %s", rangeStr)
	}
	start, end := parts[0], parts[1]

	if len(start) != len(end) || len(start) == 0 {
		return nil, fmt.Errorf("range start/end must have same, non-zero length: %s", rangeStr)
	}

	if isNumeric(start) && isNumeric(end) {
		return expandNumericRange(start, end)
	}
	if isAlphabetic(start) && isAlphabetic(end) {
		return expandAlphabeticRange(start, end)
	}

	return nil, fmt.Errorf("range must be purely alphabetic or purely numeric: %s", rangeStr)
}

// expandNumericRange expands a numeric range (e.g., "001..003").
func expandNumericRange(start, end string) ([]string, error) {
	startNum, err1 := strconv.Atoi(start)
	endNum, err2 := strconv.Atoi(end)
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("invalid numeric range: %s..%s", start, end)
	}
	if startNum > endNum {
		return nil, fmt.Errorf("start of range cannot be greater than end: %s..%s", start, end)
	}
	if endNum-startNum > maxNumericExpansion {
		return nil, fmt.Errorf("numeric range is too large to expand: %s..%s", start, end)
	}

	var result []string
	format := fmt.Sprintf("%%0%dd", len(start))
	for i := startNum; i <= endNum; i++ {
		result = append(result, fmt.Sprintf(format, i))
	}
	return result, nil
}

// expandAlphabeticRange expands an alphabetic range (e.g., "qaa..qtz").
func expandAlphabeticRange(start, end string) ([]string, error) {
	current := []byte(strings.ToLower(start))
	endBytes := []byte(strings.ToLower(end))

	if bytes.Compare(current, endBytes) > 0 {
		return nil, fmt.Errorf("start of alphabetic range cannot be greater than end: %s..%s", start, end)
	}

	var result []string
	for {
		result = append(result, string(current))
		if bytes.Equal(current, endBytes) {
			break
		}
		if len(result) > maxAlphaExpansion {
			return nil, fmt.Errorf("alphabetic range is too large to expand: %s..%s", start, end)
		}

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
func buildRecord(fields map[string][]string) Record {
	getString := func(key string) string {
		if v, ok := fields[key]; ok && len(v) > 0 {
			return v[0]
		}
		return ""
	}
	rec := Record{
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
	return rec
}
