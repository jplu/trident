# langtag

A robust and high-performance implementation for parsing, validating, and canonicalizing IETF BCP 47 language tags.

`langtag` provides a comprehensive toolset for working with language tags as specified in RFC 5646. It includes the full IANA Language Subtag Registry embedded at compile time, allowing for strict semantic validation and canonicalization without external dependencies or runtime file loading.

## Features

- **Strict BCP 47 Validation**: Performs full syntactic and semantic validation against the IANA registry, including checks for subtag length and character sets.
- **Full Canonicalization**: Normalizes tags by replacing deprecated subtags with preferred values, removing redundant scripts, and sorting extensions alphabetically.
- **Embedded IANA Registry**: The complete subtag registry is baked into the binary, ensuring consistency and zero-config deployment.
- **Extlang Support**: Easily convert canonical tags to their "extlang form" (e.g., converting `hak-CN` to `zh-hak-CN`) for legacy compatibility.
- **Component Access**: Granular methods to retrieve primary language, extended languages, scripts, regions, variants, extensions, and private-use subtags.
- **High Performance**: Features a thread-safe `Parser` designed for reuse, minimizing the overhead of registry lookups during high-volume processing.
- **Grandfathered Tag Handling**: Correctly identifies and handles non-compositional grandfathered tags like `i-klingon` or `art-lojban`.
- **JSON Support**: Implements `json.Marshaler` and `json.Unmarshaler` for seamless integration with web services and configuration files.

## Installation

```sh
go get github.com/jplu/trident/langtag
```

## Quick Start

### 1. Parsing and Component Access

The parser allows you to decompose a well-formed tag into its constituent parts for inspection.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/langtag"
)

func main() {
	// Initialize the parser once and reuse it
	p, err := langtag.NewParser()
	if err != nil {
		panic(err)
	}

	// Parse a tag for well-formedness
	lt, err := p.Parse("zh-Hans-CN-variant1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Full Tag: %s\n", lt.AsStr())

	// Access individual components
	fmt.Printf("Language: %s\n", lt.PrimaryLanguage())
	script, _ := lt.Script()
	fmt.Printf("Script:   %s\n", script)
	region, _ := lt.Region()
	fmt.Printf("Region:   %s\n", region)
}
```

### 2. Validation and Canonicalization

`ParseAndNormalize` goes beyond syntax, checking if subtags actually exist in the IANA registry and applying RFC-mandated transformations.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/langtag"
)

func main() {
	p, _ := langtag.NewParser()

	// "iw" is deprecated; the canonical form is "he" (Hebrew)
	// Extensions like -u- and -t- are sorted alphabetically
	input := "iw-u-co-phonebk-a-value"
	
	canonical, err := p.ParseAndNormalize(input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Input:     %s\n", input)
	fmt.Printf("Canonical: %s\n", canonical)
	// Output: he-a-value-u-co-phonebk
}
```

### 3. Converting to Extlang Form

RFC 5646 allows for two represents of certain languages. You can convert a canonical tag to its extended language format.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/langtag"
)

func main() {
	p, _ := langtag.NewParser()

	// Hakka Chinese canonical form is "hak"
	lt, _ := p.ParseAndNormalize("hak-CN")

	// Convert to extlang form (zh-hak-CN)
	extlangTag, err := p.ToExtlangForm(lt)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Canonical: %s\n", lt)
	fmt.Printf("Extlang:   %s\n", extlangTag)
	// Output: zh-hak-CN
}
```

### 4. Working with Extensions and Private Use

The package handles complex extensions and private-use sequences (`-x-`).

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/langtag"
)

func main() {
	p, _ := langtag.NewParser()
	input := "en-US-u-cu-usd-x-custom-data"

	lt, _ := p.Parse(input)

	// Access extensions
	for _, ext := range lt.ExtensionSubtags() {
		fmt.Printf("Extension: %c -> %s\n", ext.Singleton, ext.Value)
	}

	// Access private use data
	priv, _ := lt.PrivateUse()
	fmt.Printf("Private Use: %s\n", priv)
	// Output: custom-data
}
```

## The Well-formed vs. Valid Distinction

This package distinguishes between two levels of compliance to optimize for different use cases.

-   **Well-formed (`Parse`)**: Checks if the tag follows the ABNF syntax (e.g., subtags are 1-8 chars, correct order). It is fast and does not require extensive registry lookups for every subtag. Use this when you only need to ensure the tag won't break your data structures.

-   **Valid (`ParseAndNormalize`)**: Checks if every subtag (language, script, region, etc.) is actually defined in the IANA Language Subtag Registry. It also performs canonicalization, such as replacing deprecated codes. Use this when you need strict standards compliance and normalized tags for comparison.

## JSON Integration

`LanguageTag` implements the `json.Unmarshaler` interface, allowing you to use it directly in your configuration or API structs.

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/jplu/trident/langtag"
)

type UserProfile struct {
	ID       int                 `json:"id"`
	Language langtag.LanguageTag `json:"lang"`
}

func main() {
	jsonData := []byte(`{"id": 101, "lang": "en-us"}`)

	var profile UserProfile
	// Note: UnmarshalJSON is expensive as it initializes a parser internally.
	// For production performance, unmarshal to string and use a global Parser.
	err := json.Unmarshal(jsonData, &profile)
	if err != nil {
		panic(err)
	}

	fmt.Printf("User %d prefers language: %s (Region: %s)\n", 
		profile.ID, profile.Language.AsStr(), profile.Language.PrimaryLanguage())
}
```

## Acknowledgements

This package is inspired by the excellent Rust [**oxilangtag**](https://github.com/oxigraph/oxilangtag) library from the Oxigraph project.

## License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](../LICENSE) file for details.