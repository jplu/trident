# IRI

A robust, standards-compliant Go package for working with Internationalized Resource Identifiers (IRIs).

This library provides types and functions for parsing, validating, resolving, and relativizing IRIs based on [RFC 3987](https://www.ietf.org/rfc/rfc3987.html) (for syntax) and [RFC 3986](https://www.ietf.org/rfc/rfc3986.html) (for resolution logic).

A key feature of this package is its type-safe distinction between an absolute `Iri` and a potentially relative `Ref` (formerly `IriRef`).

## Features

- **RFC 3987 Compliant Parsing**: Strict validation of IRI syntax for schemes, authorities, paths, queries, and fragments.
- **RFC 3986 Relative IRI Resolution**: Correctly resolve relative references against a base IRI.
- **Type-Safe Distinction**: Use `Iri` for guaranteed absolute IRIs and `Ref` for relative-or-absolute references. This prevents common errors at compile time.
- **IRI Relativization**: The inverse of resolution. Create a relative IRI from two absolute IRIs.
- **Efficient Component Access**: Get IRI components (`scheme`, `authority`, `path`, etc.) without extra allocations.
- **High-Performance Resolution**: Zero-allocation resolution variants (`ResolveTo`) for performance-critical applications.
- **Unicode Normalization**: Support for Unicode Normalization Form C (NFC) for canonical representation of IRIs.
- **URI-to-IRI Conversion**: Convert URI strings to IRI references, handling percent-encoded UTF-8.
- **IRI-to-URI Conversion**: Convert IRI references to URI strings, applying IDNA (ToASCII) to the host and percent-encoding non-ASCII characters.
- **Syntax-Based Normalization**: Apply normalization rules (case, percent-encoding, path segment) from RFC 3986.
- **Built-in JSON Support**: `Iri` and `Ref` types implement `json.Marshaler` and `json.Unmarshaler` for easy integration with web APIs.

## Installation

```sh
go get github.com/jplu/trident/iri
```

## Quick Start

### 1. Parsing and Validation

You can parse a string into a validated `Iri` (which must be absolute) or a `Ref`. The library offers `ParseRef` for direct parsing and `ParseNormalizedRef` for inputs that require Unicode NFC normalization first.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/iri"
)

func main() {
	// Parse a known absolute IRI. This will fail if the IRI is relative.
	myIri, err := iri.ParseIri("http://user:pass@example.com:8080/path/to/resource?q=1#fragment")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully parsed IRI: %s\n\n", myIri)

	// Access its components
	fmt.Printf("Scheme:    %s\n", myIri.Scheme())
	authority, _ := myIri.Authority()
	fmt.Printf("Authority: %s\n", authority)
	fmt.Printf("Path:      %s\n", myIri.Path())
	query, _ := myIri.Query()
	fmt.Printf("Query:     %s\n", query)
	fragment, _ := myIri.Fragment()
	fmt.Printf("Fragment:  %s\n\n", fragment)

	// Ref can handle relative references
	relativeRef, err := iri.ParseRef("../other?key=val")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Parsed relative reference: %s\n", relativeRef)

	// Convert a URI to an IRI
	uriToConvert := "http://example.com/search?q=%D0%B1%D0%B5%D0%B3" // Russian "бег"
	iriFromUri, err := iri.ParseURIToRef(uriToConvert)
	if err != nil {
		panic(err)
	}
	fmt.Printf("IRI from URI: %s\n", iriFromUri)
}
```

### 2. Resolving a Relative IRI

The core use case is resolving a relative reference against a base IRI.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/iri"
	"strings"
)

func main() {
	// 1. Establish a base IRI.
	baseIri, err := iri.ParseIri("http://a/b/c/d;p?q")
	if err != nil {
		panic(err)
	}

	// 2. Define a relative reference.
	relativeRef := "../g"

	// 3. Resolve it.
	resolvedIri, err := baseIri.Resolve(relativeRef)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Base:     %s\n", baseIri)
	fmt.Printf("Relative: %s\n", relativeRef)
	fmt.Printf("Resolved: %s\n", resolvedIri)
	// Output: Resolved: http://a/b/g

	// Zero-allocation resolution with ResolveTo
	var builder strings.Builder
	err = baseIri.ResolveTo("../g", &builder)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Resolved (to builder): %s\n", builder.String())
}
```

### 3. Relativizing an IRI

You can also perform the inverse of resolution: find the relative path between two absolute IRIs.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/iri"
)

func main() {
	base, _ := iri.ParseIri("http://example.com/foo/bar")
	target, _ := iri.ParseIri("http://example.com/foo/baz#section")

	relativeRef, err := base.Relativize(target)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Base:   %s\n", base)
	fmt.Printf("Target: %s\n", target)
	fmt.Printf("Relativized: %s\n\n", relativeRef)
	// Output: Relativized: baz#section

	// You can prove it works by resolving it back.
	resolved, _ := base.Resolve(relativeRef.String())
	fmt.Printf("Resolved back: %s\n", resolved)
	fmt.Printf("Matches target: %t\n", resolved.String() == target.String())
}
```

### 4. Normalization and URI Conversion

The `Ref` type provides methods for syntax-based normalization and converting to a URI.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/iri"
)

func main() {
	// An IRI with mixed case scheme, percent-encoded unreserved chars, and dot segments
	myRef, _ := iri.ParseRef("HTTP://example.com/a/./b/../c%7E?query%20param#fragment")

	fmt.Printf("Original Ref: %s\n", myRef)

	// Normalize applies case, percent-encoding, and path segment normalization,
	// and ensures NFC.
	normalizedRef := myRef.Normalize()
	fmt.Printf("Normalized Ref: %s\n", normalizedRef)
	// Output will be similar to: http://example.com/a/c~?query%20param#fragment

	// Convert to URI, which also involves IDNA for host and percent-encoding non-ASCII.
	// Let's use an IRI with a Unicode host and path.
	unicodeIri, _ := iri.ParseRef("http://bücher.example.com/Résumé.html")
	uri := unicodeIri.ToURI()
	fmt.Printf("Unicode IRI: %s\n", unicodeIri)
	fmt.Printf("Converted to URI: %s\n", uri)
	// Output: http://xn--bcher-kva.example.com/R%C3%A9sum%C3%A9.html
}
```

## The `Iri` vs. `Ref` Distinction

This library's design provides compile-time safety by separating two concepts:

-   **`Ref`**: An "IRI Reference". This can be a full, absolute IRI or a relative reference (like `/path/only`, `../`, or `#fragment`). This type should be used when you are handling input that might not be a fully-qualified IRI.

-   **`Iri`**: A guaranteed absolute IRI. It always has a scheme. You cannot create an `Iri` from a relative reference. This is useful for types that require a non-ambiguous, absolute identifier, such as a base IRI for resolution.

You can convert a `Ref` to an `Iri` using `NewIriFromRef()`, which will return an error if the `Ref` is not absolute.

## JSON Marshaling and Unmarshaling

Both `Iri` and `Ref` implement the `json.Marshaler` and `json.Unmarshaler` interfaces. The unmarshaler automatically validates the IRI string.

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/jplu/trident/iri"
)

type Config struct {
	BaseURL    *iri.Iri `json:"baseUrl"`
	Stylesheet *iri.Ref `json:"stylesheet"`
}

func main() {
	jsonData := []byte(`{
		"baseUrl": "https://api.example.com/v1/",
		"stylesheet": "styles/main.css"
	}`)

	var cfg Config
	err := json.Unmarshal(jsonData, &cfg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Base URL from JSON: %s\n", cfg.BaseURL)
	fmt.Printf("Stylesheet from JSON: %s\n", cfg.Stylesheet)

	// Now you can safely use the validated types
	resolvedStylesheet, _ := cfg.BaseURL.Resolve(cfg.Stylesheet.String())
	fmt.Printf("Full stylesheet path: %s\n", resolvedStylesheet)
}
```

## Acknowledgements

This package is inspired by the excellent Rust [**oxiri**](https://github.com/oxigraph/oxiri) library from the Oxigraph project.

## License

This project is licensed under the APACHE License - see the [LICENSE](../LICENSE) file for details.
