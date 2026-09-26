# datatypes

A robust, high-performance implementation of the W3C XML Schema Definition (XSD) 1.1 Part 2 datatypes.

`datatypes` provides a comprehensive type system for parsing, validating, comparing, and canonicalizing XML Schema values. It supports all Simple Type Definition varieties—**atomic**, **list**, and **union**—alongside their constraining facets, exact value space identity semantics, and canonical lexical mappings. This implementation is limited to the datatypes supported by [RDF 1.2](https://www.w3.org/TR/rdf12-concepts/#xsd-datatypes).

## Features

- **Full XSD 1.1 Datatype Coverage**: Complete implementation of primitive and derived datatypes, including strings, numerics, temporal types, booleans, and binary formats.
- **Type Varieties**: Built-in support for atomic simple types (`AtomicType`), list simple types (`ListType`, `ListValue`), and union simple types (`UnionType`).
- **Constraining Facets**: Comprehensive suite of facets conforming to XSD 1.1:
  - *Length restrictions*: `length`, `minLength`, `maxLength`
  - *Regular expressions*: `pattern` (with character class subtraction and XML character classes)
  - *Enumerations*: `enumeration`
  - *Whitespace control*: `whiteSpace` (`preserve`, `replace`, `collapse`)
  - *Boundaries*: `minInclusive`, `minExclusive`, `maxInclusive`, `maxExclusive`
  - *Precision limits*: `totalDigits`, `fractionDigits`
  - *Timezone rules*: `explicitTimezone` (`required`, `prohibited`, `optional`)
- **Arbitrary-Precision Numerics**:
  - `xsd:decimal`: Arbitrary-precision exact arithmetic and canonical termination analysis powered by `math/big.Rat`.
  - `xsd:integer`: Infinite-precision integers via `math/big.Int`, along with all 12 integer restrictions (`long`, `int`, `short`, `byte`, `nonNegativeInteger`, `positiveInteger`, etc.) with strict bounds checking.
  - Cross-type numeric comparisons across the shared decimal value space.
- **IEEE 754 Compliant Floats**: `xsd:float` and `xsd:double` supporting special states (`INF`, `-INF`, `NaN`), signed zeros (`-0.0` vs. `+0.0`), and canonical scientific formatting.
- **Temporal Modeling & Timeline Engine**:
  - High-precision timeline coordinate calculation supporting Gregorian leap years, 400-year cycles, and timezone offsets from `-14:00` to `+14:00`.
  - Temporal types: `dateTime`, `dateTimeStamp`, `date`, `time`, `gYear`, `gYearMonth`, `gMonth`, `gMonthDay`, and `gDay`.
  - Full duration arithmetic (`duration`, `yearMonthDuration`, `dayTimeDuration`) with 4-anchor indeterminate partial-order comparison.
- **XSD Regex Engine Translation**: Translates XSD-specific regular expressions—including character class subtractions (`-[...]`), XML escapes (`\i`, `\c`, `\w`), and Unicode block categories (`\p{Is...}`)—into Go RE2-compatible expressions with implicit anchoring.
- **Standards Compliant Identity**: Implements `IsIdenticalWith` adhering to the strict identity semantics defined by W3C XSD 1.1 Part 2.

## Installation

```sh
go get github.com/jplu/trident/datatypes
```

## Quick Start

### 1. Parsing and Canonicalization

Parse lexical strings into their respective datatype structures and output their canonical representations according to XSD 1.1 rules.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/datatypes"
)

func main() {
	// Arbitrary-precision decimal: parses non-canonical literals and formats canonically
	dec, err := datatypes.ParseDecimal("+00123.4500")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Decimal Canonical: %s\n", dec.String()) // 123.45

	// 64-bit integer
	lng, err := datatypes.ParseLong("0000042")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Long: %s\n", lng.String()) // 42

	// Base64 binary: strips and collapses whitespace automatically
	b64, err := datatypes.ParseBase64Binary("  SGVsbG8g V29ybGQ= \n")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Base64 Decoded: %s\n", string(b64)) // Hello World
	fmt.Printf("Base64 Canonical: %s\n", b64.String()) // SGVsbG8gV29ybGQ=
}
```

### 2. Temporal Types and Timezones

Work with temporal datatypes, perform calendar arithmetic with durations, and adjust timezone offsets.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/datatypes"
)

func main() {
	// Parse an ISO/XSD dateTime
	dt, err := datatypes.ParseDateTime("2026-03-30T10:15:30.500+02:00")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Year: %d, Month: %d, Day: %d\n", dt.Year(), dt.Month(), dt.Day())
	fmt.Printf("Canonical DateTime: %s\n", dt.String())

	// Parse and add a duration
	dur, err := datatypes.ParseDuration("P1Y2M3DT4H")
	if err != nil {
		panic(err)
	}
	newDt := dt.AddDuration(dur)
	fmt.Printf("Added Duration:     %s\n", newDt.String())

	// Adjust timezone to UTC (+00:00 / Z)
	adjusted := newDt.Adjust(datatypes.GetUTC())
	fmt.Printf("Adjusted to UTC:    %s\n", adjusted.String())
}
```

### 3. Evaluating Constraining Facets

Enforce constraints on value spaces and lexical spaces using individual facets or compiled type definitions.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/datatypes"
)

func main() {
	// 1. Lexical constraint: Pattern facet with XSD regex syntax
	matcher, err := datatypes.CompileRegex(`[A-Z]{3}-\d{4}`)
	if err != nil {
		panic(err)
	}
	patternFacet := datatypes.NewFacetPattern([]*datatypes.PatternMatcher{matcher})

	if err := patternFacet.CheckLexical("XYZ-2026"); err != nil {
		fmt.Printf("Pattern check failed: %v\n", err)
	} else {
		fmt.Println("Pattern matched successfully!")
	}

	// 2. Value-space constraint: minInclusive and maxInclusive
	minBound := datatypes.NewFacetMinInclusive(datatypes.NewIntegerFromInt64(10), false)
	maxBound := datatypes.NewFacetMaxInclusive(datatypes.NewIntegerFromInt64(100), false)

	val, _ := datatypes.ParseInteger("42")
	if err := minBound.Check(val); err == nil && maxBound.Check(val) == nil {
		fmt.Printf("Value %s is within [10, 100]\n", val.String())
	}
}
```

### 4. Working with List and Union Varieties

Parse and validate sequences of atomic values or flexible union datatypes.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/datatypes"
)

func main() {
	// Construct and inspect a ListValue
	val1, _ := datatypes.ParseInteger("10")
	val2, _ := datatypes.ParseInteger("20")
	val3, _ := datatypes.ParseInteger("30")

	list := datatypes.ListValue{val1, val2, val3}
	fmt.Printf("List Value: %s\n", list.String()) // "10 20 30"
	fmt.Printf("List Length: %d items\n", list.Length())

	// Check facet constraints on lists
	lengthFacet := datatypes.NewFacetLength(3, false)
	if err := lengthFacet.Check(list); err == nil {
		fmt.Println("List satisfies length facet!")
	}
}
```

## Value Space, Lexical Space, and Identity

XSD 1.1 distinguishes between the **lexical space** (string representations) and the **value space** (abstract mathematical objects):

- **Whitespace Normalization**: Occurs before lexical validation and value mapping. The engine supports `preserve` (raw characters kept), `replace` (tabs, newlines, and carriage returns become spaces), and `collapse` (contiguous spaces merged, trimmed).
- **Identity (`IsIdenticalWith`) vs. Equality (`Compare`)**:
  - In IEEE 754 floating-point arithmetic, `+0.0 == -0.0` is true, but under XSD 1.1 Part 2 (Section 3.3.4.1), positive zero and negative zero are distinct elements in the value space: `posZero.IsIdenticalWith(negZero)` evaluates to `false`.
  - Two `dateTime` instances representing the exact same timeline instant in different timezones (e.g., `2026-01-01T12:00:00Z` and `2026-01-01T14:00:00+02:00`) are equal in comparison, but are **not** identical because their timezone components differ.
- **Cross-Datatype Numeric Comparison**: `xsd:decimal`, `xsd:integer`, and derived types share a unified numeric value space. You can compare an `Int` or `UnsignedLong` directly against an arbitrary-precision `Decimal`.

## Acknowledgements

This package is inspired by the excellent Rust [**oxsdatatypes**](https://github.com/oxigraph/oxigraph/tree/main/lib/oxsdatatypes) library from the Oxigraph project.

## License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](../LICENSE) file for details.