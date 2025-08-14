# LangTag

`langtag` is a Go library for parsing, validating, and manipulating IETF BCP 47 language tags, as specified in [RFC 5646](https://tools.ietf.org/html/rfc5646).

This package provides a robust and efficient solution for applications requiring strict conformance to language tag standards. It includes the full IANA Language Subtag Registry, embedded at compile time, ensuring that the module works out of the box with no additional setup.

## Features

-   **Fully RFC 5646 Compliant**: Strictly validates tags against both "well-formed" syntax rules and IANA semantic rules.
-   **Comprehensive Canonicalization**: The `ParseAndNormalize` method automatically canonicalizes tags to their most optimal form, including:
    -   **Case Normalization**: Corrects the case of all subtags (e.g., `en-us` becomes `en-US`).
    -   **Deprecated Subtag Replacement**: Replaces deprecated and grandfathered tags with their modern, preferred equivalents from the IANA registry (e.g., `no-nyn` becomes `nn`; `i-klingon` becomes `tlh`).
    -   **Extlang Conversion**: Supports converting canonical tags to their "extlang form" (e.g., `hak-CN` to `zh-hak-CN`) via the `ToExtlangForm` method.
-   **High Performance & Flexible Parsing**:
    -   The `Parser` is reusable and thread-safe. Create it once and reuse it across your application for maximum performance.
    -   `ParseAndNormalize`: The recommended function for most use cases, providing a fully valid and canonical tag.
    -   `Parse`: A lighter-weight parser for checking "well-formedness" syntactically, ideal for performance-critical paths where full validation and canonicalization are not required.
-   **Rich Component Access**: Provides simple methods to access all parts of a tag:
    -   Primary Language (`PrimaryLanguage()`)
    -   Extended Language (`ExtendedLanguageSubtags()`)
    -   Script (`Script()`)
    -   Region (`Region()`)
    -   Variants (`VariantSubtags()`)
    -   Extensions (`ExtensionSubtags()`)
    -   Private-Use subtags (`PrivateUseSubtags()`)
-   **Built-in JSON Support**: Implements `json.Marshaler` and `json.Unmarshaler` for seamless integration. Unmarshaling automatically performs full validation and canonicalization.
-   **Self-Contained**: Has no runtime file dependencies, as the IANA registry is embedded in the library.
-   **Zero External Dependencies**: Relies only on the Go standard library.

## Installation

To add `langtag` to your project, use `go get`:

```sh
go get github.com/jplu/trident/langtag
```

## Quick Start

For efficiency, create a single `Parser` instance and reuse it. `NewParser()` is an expensive call that should ideally be made only once during your application's initialization.

```go
package main

import (
	"fmt"
	"log"

	"github.com/jplu/trident/langtag"
)

func main() {
	// Create a reusable parser. This is the recommended practice.
	p, err := langtag.NewParser()
	if err != nil {
		log.Fatalf("Failed to initialize parser: %v", err)
	}

	// --- Example 1: Parsing and normalizing a complex tag ---
	// ParseAndNormalize is the recommended method for most use cases.
	tag1, err := p.ParseAndNormalize("zh-cmn-hans-cn")
	if err != nil {
		log.Fatalf("Failed to parse language tag: %v", err)
	}

	// The String() method returns the normalized tag.
	fmt.Printf("Normalized Tag: %s\n", tag1) // Output: Normalized Tag: cmn-Hans-CN

	// Access individual components.
	if script, ok := tag1.Script(); ok {
		fmt.Printf("Script: %s\n", script) // Output: Script: Hans
	}
	if region, ok := tag1.Region(); ok {
		fmt.Printf("Region: %s\n", region) // Output: Region: CN
	}
	fmt.Println("------")

	// --- Example 2: Canonicalizing a deprecated grandfathered tag ---
	tag2, err := p.ParseAndNormalize("i-klingon")
	if err != nil {
		log.Fatalf("Failed to parse language tag: %v", err)
	}

	// ParseAndNormalize canonicalizes the tag, replacing the deprecated "i-klingon".
	fmt.Printf("Canonicalized Tag: %s\n", tag2) // Output: Canonicalized Tag: tlh
	fmt.Printf("Primary Language: %s\n", tag2.PrimaryLanguage()) // Output: Primary Language: tlh
	fmt.Println("------")

	// --- Example 3: Using Parse() for well-formedness checks ---
	// Use Parse() when you only need to check syntax and normalize case,
	// without performing IANA validation or replacing deprecated forms.
	rawTag, err := p.Parse("no-nyn")
	if err != nil {
		log.Fatalf("Failed to parse raw tag: %v", err)
	}
	// Note that the deprecated form is preserved, but case is normalized.
	fmt.Printf("Raw parsed tag (deprecated form preserved): %s\n", rawTag) // Output: Raw parsed tag (deprecated form preserved): no-nyn
}
```

## JSON Marshaling and Unmarshaling

The `LanguageTag` type natively supports JSON serialization and deserialization. Unmarshaling automatically performs full **validation and canonicalization**.

### Marshaling to JSON

A `LanguageTag` struct marshals to a simple JSON string.

```go
p, _ := langtag.NewParser()
tag, _ := p.ParseAndNormalize("en-GB")
data, err := json.Marshal(struct {
    UserLocale langtag.LanguageTag `json:"locale"`
}{
    UserLocale: tag,
})

// data will be: []byte(`{"locale":"en-GB"}`)
```

### Unmarshaling from JSON

When unmarshaling, the library automatically **parses, validates, and canonicalizes** the string value. This includes case normalization and replacement of deprecated tags. If the tag is invalid, an error is returned.

```go
var result struct {
    UserLocale langtag.LanguageTag `json:"locale"`
}

// Example 1: Input tag "de-at" is not in canonical case.
jsonData1 := []byte(`{"locale":"de-at"}`)
err := json.Unmarshal(jsonData1, &result)

// Unmarshaling validates and normalizes the tag's case.
// result.UserLocale.String() will be "de-AT".
fmt.Println(result.UserLocale) // Output: de-AT


// Example 2: Input tag "zh-min-nan" is a deprecated form.
jsonData2 := []byte(`{"locale":"zh-min-nan"}`)
err = json.Unmarshal(jsonData2, &result)

// Unmarshaling also canonicalizes it to its preferred modern equivalent.
// result.UserLocale.String() will be "nan".
fmt.Println(result.UserLocale) // Output: nan
```

> **Performance Warning**: The `LanguageTag.UnmarshalJSON` method creates a new parser on every call, which is an expensive operation. For performance-critical applications that unmarshal many tags, it is significantly more efficient to unmarshal the tag into a `string` variable first, and then use a single, pre-initialized `Parser` instance to parse it with `ParseAndNormalize`.

## Acknowledgements

This package is inspired by the excellent Rust [**oxilangtag**](https://github.com/oxigraph/oxilangtag) library from the Oxigraph project.

## License

This project is licensed under the APACHE-2.0 License.
