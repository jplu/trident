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

// nolint:testpackage // This is a white-box test file for an internal package. It needs to be in the same package to
// test unexported functions.
package datatypes

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

// JSONFacet represents the structure of each compiled facet inside tests.json.
type JSONFacet struct {
	Value interface{} `json:"value"` // Can be string or []string (pattern/enum)
	Fixed bool        `json:"fixed"`
}

// JSONType maps nested item types or member types for lists and unions.
type JSONType struct {
	Variety     string               `json:"variety"`
	BaseType    string               `json:"baseType,omitempty"`
	ItemType    *JSONType            `json:"itemType,omitempty"`
	MemberTypes []JSONType           `json:"memberTypes,omitempty"`
	Facets      map[string]JSONFacet `json:"facets"`
}

// JSONTest represents a single test definition structure.
type JSONTest struct {
	ID          string               `json:"id"`
	Variety     string               `json:"variety"`
	BaseType    string               `json:"baseType,omitempty"`
	ItemType    *JSONType            `json:"itemType,omitempty"`
	MemberTypes []JSONType           `json:"memberTypes,omitempty"`
	Facets      map[string]JSONFacet `json:"facets"`
	Cases       []struct {
		Literal  string `json:"literal"`
		Expected string `json:"expected"`
	} `json:"cases"`
}

type typeBuilder struct {
	primitives map[string]SimpleTypeDefinition
}

func newTypeBuilder() *typeBuilder {
	b := &typeBuilder{
		primitives: make(map[string]SimpleTypeDefinition),
	}
	b.initPrimitives()
	return b
}

func (b *typeBuilder) reg(name string, baseName string, primName string, ws *WhiteSpaceValue, mapper LexicalMapper) {
	at := &AtomicType{
		name:          name,
		lexicalMapper: mapper,
	}
	if baseName != "" {
		at.baseType = b.primitives[baseName]
	}
	if primName != "" {
		at.primitiveType = b.primitives[primName]
	} else {
		at.primitiveType = at
	}

	if ws != nil {
		at.facets = append(at.facets, NewFacetWhiteSpace(*ws, true))
	}

	b.primitives[name] = at
}

func (b *typeBuilder) initPrimitives() {
	wsReplace := WhiteSpaceReplace
	wsCollapse := WhiteSpaceCollapse

	b.reg("string", "", "", nil, func(s string) (XSDValue, error) { return ParseString(s), nil })
	b.reg(
		"normalizedString",
		"string",
		"string",
		&wsReplace,
		func(s string) (XSDValue, error) { return ParseNormalizedString(s), nil },
	)
	b.reg(
		"token",
		"normalizedString",
		"string",
		&wsCollapse,
		func(s string) (XSDValue, error) { return ParseToken(s), nil },
	)
	b.reg("language", "token", "string", &wsCollapse, func(s string) (XSDValue, error) { return ParseLanguage(s) })
	b.reg("Name", "token", "string", &wsCollapse, func(s string) (XSDValue, error) { return ParseName(s) })
	b.reg("NCName", "Name", "string", &wsCollapse, func(s string) (XSDValue, error) { return ParseNCName(s) })
	b.reg("NMTOKEN", "token", "string", &wsCollapse, func(s string) (XSDValue, error) { return ParseNMTOKEN(s) })

	b.reg("decimal", "", "", nil, func(s string) (XSDValue, error) { return ParseDecimal(s) })
	b.reg("integer", "decimal", "decimal", nil, func(s string) (XSDValue, error) { return ParseInteger(s) })
	b.reg(
		"nonPositiveInteger",
		"integer",
		"decimal",
		nil,
		func(s string) (XSDValue, error) { return ParseNonPositiveInteger(s) },
	)
	b.reg(
		"negativeInteger",
		"nonPositiveInteger",
		"decimal",
		nil,
		func(s string) (XSDValue, error) { return ParseNegativeInteger(s) },
	)
	b.reg("long", "integer", "decimal", nil, func(s string) (XSDValue, error) { return ParseLong(s) })
	b.reg("int", "long", "decimal", nil, func(s string) (XSDValue, error) { return ParseInt(s) })
	b.reg("short", "int", "decimal", nil, func(s string) (XSDValue, error) { return ParseShort(s) })
	b.reg("byte", "short", "decimal", nil, func(s string) (XSDValue, error) { return ParseByte(s) })
	b.reg(
		"nonNegativeInteger",
		"integer",
		"decimal",
		nil,
		func(s string) (XSDValue, error) { return ParseNonNegativeInteger(s) },
	)
	b.reg(
		"unsignedLong",
		"nonNegativeInteger",
		"decimal",
		nil,
		func(s string) (XSDValue, error) { return ParseUnsignedLong(s) },
	)
	b.reg(
		"unsignedInt",
		"unsignedLong",
		"decimal",
		nil,
		func(s string) (XSDValue, error) { return ParseUnsignedInt(s) },
	)
	b.reg(
		"unsignedShort",
		"unsignedInt",
		"decimal",
		nil,
		func(s string) (XSDValue, error) { return ParseUnsignedShort(s) },
	)
	b.reg(
		"unsignedByte",
		"unsignedShort",
		"decimal",
		nil,
		func(s string) (XSDValue, error) { return ParseUnsignedByte(s) },
	)
	b.reg(
		"positiveInteger",
		"nonNegativeInteger",
		"decimal",
		nil,
		func(s string) (XSDValue, error) { return ParsePositiveInteger(s) },
	)

	b.reg("boolean", "", "", nil, func(s string) (XSDValue, error) { return ParseBoolean(s) })
	b.reg("float", "", "", nil, func(s string) (XSDValue, error) { return ParseFloat(s) })
	b.reg("double", "", "", nil, func(s string) (XSDValue, error) { return ParseDouble(s) })

	b.reg("dateTime", "", "", nil, func(s string) (XSDValue, error) { return ParseDateTime(s) })
	b.reg(
		"dateTimeStamp",
		"dateTime",
		"dateTime",
		nil,
		func(s string) (XSDValue, error) { return ParseDateTimeStamp(s) },
	)
	b.reg("date", "", "", nil, func(s string) (XSDValue, error) { return ParseDate(s) })
	b.reg("time", "", "", nil, func(s string) (XSDValue, error) { return ParseTime(s) })
	b.reg("gYear", "", "", nil, func(s string) (XSDValue, error) { return ParseGYear(s) })
	b.reg("gYearMonth", "", "", nil, func(s string) (XSDValue, error) { return ParseGYearMonth(s) })
	b.reg("gMonth", "", "", nil, func(s string) (XSDValue, error) { return ParseGMonth(s) })
	b.reg("gMonthDay", "", "", nil, func(s string) (XSDValue, error) { return ParseGMonthDay(s) })
	b.reg("gDay", "", "", nil, func(s string) (XSDValue, error) { return ParseGDay(s) })

	b.reg("duration", "", "", nil, func(s string) (XSDValue, error) { return ParseDuration(s) })
	b.reg(
		"yearMonthDuration",
		"duration",
		"duration",
		nil,
		func(s string) (XSDValue, error) { return ParseYearMonthDuration(s) },
	)
	b.reg(
		"dayTimeDuration",
		"duration",
		"duration",
		nil,
		func(s string) (XSDValue, error) { return ParseDayTimeDuration(s) },
	)

	b.reg("hexBinary", "", "", nil, func(s string) (XSDValue, error) { return ParseHexBinary(s) })
	b.reg("base64Binary", "", "", nil, func(s string) (XSDValue, error) { return ParseBase64Binary(s) })
	b.reg("anyURI", "", "", &wsCollapse, func(s string) (XSDValue, error) { return ParseAnyURI(s) })
}

func parseStringList(val interface{}) []string {
	switch v := val.(type) {
	case string:
		return []string{v}
	case []interface{}:
		var result []string
		for _, item := range v {
			if s, isStr := item.(string); isStr {
				result = append(result, s)
			}
		}
		return result
	default:
		return nil
	}
}

func parseJSONInt(val interface{}) (int, bool) {
	switch v := val.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

func buildIntFacet(name string, jf JSONFacet) Facet {
	v, ok := parseJSONInt(jf.Value)
	if !ok {
		return nil
	}
	switch name {
	case "length":
		return NewFacetLength(v, jf.Fixed)
	case "minLength":
		return NewFacetMinLength(v, jf.Fixed)
	case "maxLength":
		return NewFacetMaxLength(v, jf.Fixed)
	case "totalDigits":
		return NewFacetTotalDigits(v, jf.Fixed)
	case "fractionDigits":
		return NewFacetFractionDigits(v, jf.Fixed)
	default:
		return nil
	}
}

func buildStringFacet(name string, jf JSONFacet) Facet {
	v, ok := jf.Value.(string)
	if !ok {
		return nil
	}
	switch name {
	case "whiteSpace":
		return NewFacetWhiteSpace(WhiteSpaceValue(v), jf.Fixed)
	case "explicitTimezone":
		return NewFacetExplicitTimezone(ExplicitTimezoneValue(v), jf.Fixed)
	default:
		return nil
	}
}

func buildPatternFacets(jf JSONFacet) []Facet {
	pats := parseStringList(jf.Value)
	var facets []Facet
	for _, p := range pats {
		if m, err := CompileRegex(p); err == nil {
			facets = append(facets, NewFacetPattern([]*PatternMatcher{m}))
		}
	}
	return facets
}

func buildEnumFacet(base SimpleTypeDefinition, jf JSONFacet) Facet {
	if base == nil {
		return nil
	}
	enums := parseStringList(jf.Value)
	var values []XSDValue
	for _, e := range enums {
		if val, err := base.ValidateLexical(e); err == nil {
			values = append(values, val)
		}
	}
	if len(values) == 0 {
		return nil
	}
	return NewFacetEnumeration(values)
}

func buildBoundaryFacet(base SimpleTypeDefinition, name string, jf JSONFacet) Facet {
	if base == nil {
		return nil
	}
	v, ok := jf.Value.(string)
	if !ok {
		return nil
	}
	val, err := base.ValidateLexical(v)
	if err != nil {
		return nil
	}
	switch name {
	case "minInclusive":
		return NewFacetMinInclusive(val, jf.Fixed)
	case "minExclusive":
		return NewFacetMinExclusive(val, jf.Fixed)
	case "maxInclusive":
		return NewFacetMaxInclusive(val, jf.Fixed)
	case "maxExclusive":
		return NewFacetMaxExclusive(val, jf.Fixed)
	default:
		return nil
	}
}

// buildFacets parses dynamically structured facets into strict constraining Facet implementations.
func buildFacets(base SimpleTypeDefinition, raw map[string]JSONFacet) []Facet {
	var facets []Facet
	for name, jf := range raw {
		if f := buildIntFacet(name, jf); f != nil {
			facets = append(facets, f)
			continue
		}
		if f := buildStringFacet(name, jf); f != nil {
			facets = append(facets, f)
			continue
		}
		if name == "pattern" {
			facets = append(facets, buildPatternFacets(jf)...)
			continue
		}
		if name == "enumeration" {
			if f := buildEnumFacet(base, jf); f != nil {
				facets = append(facets, f)
			}
			continue
		}
		if f := buildBoundaryFacet(base, name, jf); f != nil {
			facets = append(facets, f)
		}
	}
	return facets
}

func collectBaseFacets(base SimpleTypeDefinition) []Facet {
	var facets []Facet
	for curr := base; curr != nil; curr = curr.BaseType() {
		facets = append(facets, curr.Facets()...)
	}
	return facets
}

func (b *typeBuilder) createAtomicType(
	name string,
	baseTypeName string,
	rawFacets map[string]JSONFacet,
) SimpleTypeDefinition {
	base := b.primitives[baseTypeName]
	if base == nil {
		return nil
	}
	at := &AtomicType{
		name:          name,
		baseType:      base,
		primitiveType: base.PrimitiveType(),
	}
	at.facets = append(buildFacets(at, rawFacets), collectBaseFacets(base)...)
	return at
}

func (b *typeBuilder) buildListType(jt *JSONType) SimpleTypeDefinition {
	var item SimpleTypeDefinition
	if jt.ItemType != nil {
		item = b.buildType(jt.ItemType)
	} else if jt.BaseType != "" {
		item = b.primitives[jt.BaseType]
	}

	if item == nil {
		return nil
	}
	lt := &ListType{
		name:     "list_derived",
		itemType: item,
	}

	var baseFacets []Facet
	if jt.BaseType != "" && jt.ItemType == nil {
		baseFacets = collectBaseFacets(b.primitives[jt.BaseType])
	}
	lt.facets = append(buildFacets(lt, jt.Facets), baseFacets...)
	return lt
}

func (b *typeBuilder) buildMembers(memberTypes []JSONType) []SimpleTypeDefinition {
	var members []SimpleTypeDefinition
	for _, mt := range memberTypes {
		if m := b.buildType(&mt); m != nil {
			members = append(members, m)
		}
	}
	return members
}

func (b *typeBuilder) buildUnionType(jt *JSONType) SimpleTypeDefinition {
	members := b.buildMembers(jt.MemberTypes)
	ut := &UnionType{
		name:        "union_derived",
		memberTypes: members,
	}
	ut.facets = buildFacets(ut, jt.Facets)
	return ut
}

// buildType recursively converts structural JSON types to implementations of SimpleTypeDefinition.
func (b *typeBuilder) buildType(jt *JSONType) SimpleTypeDefinition {
	if jt == nil {
		return nil
	}

	switch jt.Variety {
	case "atomic":
		return b.createAtomicType(jt.BaseType+"_derived", jt.BaseType, jt.Facets)
	case "list":
		return b.buildListType(jt)
	case "union":
		return b.buildUnionType(jt)
	default:
		return nil
	}
}

func (b *typeBuilder) buildListTypeFromTest(jt *JSONTest) SimpleTypeDefinition {
	var item SimpleTypeDefinition
	if jt.ItemType != nil {
		item = b.buildType(jt.ItemType)
	} else if jt.BaseType != "" {
		item = b.createAtomicType(jt.ID+"_item", jt.BaseType, jt.Facets)
		jt.Facets = nil
	}

	if item == nil {
		return nil
	}
	lt := &ListType{
		name:     jt.ID,
		itemType: item,
	}
	if jt.Facets != nil {
		lt.facets = buildFacets(lt, jt.Facets)
	}
	return lt
}

func (b *typeBuilder) buildUnionTypeFromTest(jt *JSONTest) SimpleTypeDefinition {
	members := b.buildMembers(jt.MemberTypes)
	ut := &UnionType{
		name:        jt.ID,
		memberTypes: members,
	}
	if jt.Facets != nil {
		ut.facets = buildFacets(ut, jt.Facets)
	}
	return ut
}

// buildTypeFromTest creates the target type schema directly from the top-level test definition.
func (b *typeBuilder) buildTypeFromTest(jt *JSONTest) SimpleTypeDefinition {
	variety := jt.Variety
	if len(jt.MemberTypes) > 0 {
		variety = "union"
	} else if jt.ItemType != nil {
		variety = "list"
	}

	switch variety {
	case "atomic":
		return b.createAtomicType(jt.ID, jt.BaseType, jt.Facets)
	case "list":
		return b.buildListTypeFromTest(jt)
	case "union":
		return b.buildUnionTypeFromTest(jt)
	default:
		return nil
	}
}

// TestW3CSuite executes the dynamic schema conformance suite.
func TestW3CSuite(t *testing.T) {
	data, err := os.ReadFile("../tests/xsd_tests.json")
	if err != nil {
		t.Skip("tests.json file not found. Place it in datatypes/ directory to execute.")
	}

	var testDefs []JSONTest
	if unmarshalErr := json.Unmarshal(data, &testDefs); unmarshalErr != nil {
		t.Fatalf("Failed to parse tests.json: %v", unmarshalErr)
	}

	builder := newTypeBuilder()

	for _, tc := range testDefs {
		t.Run(tc.ID, func(t *testing.T) {
			typeDef := builder.buildTypeFromTest(&tc)
			if typeDef == nil {
				t.Skipf("Unresolvable base types or features on schema: %s", tc.ID)
			}

			for i, c := range tc.Cases {
				_, valErr := Validate(c.Literal, typeDef)
				hasError := valErr != nil

				if c.Expected == "valid" && hasError {
					t.Errorf(
						"Case %d: expected valid literal '%s', but validation failed with error: %v",
						i,
						c.Literal,
						valErr,
					)
				} else if c.Expected == "invalid" && !hasError {
					t.Errorf("Case %d: expected validation to fail for invalid literal '%s', but it passed.", i, c.Literal)
				}
			}
		})
	}
}
