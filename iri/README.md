# IRI

A robust, standards-compliant Go package for working with Internationalized Resource Identifiers (IRIs).

This library provides types and functions for parsing, validating, resolving, and relativizing IRIs based on [RFC 3987](https://www.ietf.org/rfc/rfc3987.html) (for syntax) and [RFC 3986](https://www.ietf.org/rfc/rfc3986.html) (for resolution logic).

A key feature of this package is its type-safe distinction between an absolute `Iri` and a potentially relative `IriRef`, inspired by the design of its Rust counterpart.

## Features

- **RFC 3987 Compliant Parsing**: Strict validation of IRI syntax for schemes, authorities, paths, queries, and fragments.
- **RFC 3986 Relative IRI Resolution**: Correctly resolve relative references against a base IRI.
- **Type-Safe Distinction**: Use `Iri` for guaranteed absolute IRIs and `IriRef` for relative-or-absolute references. This prevents common errors at compile time.
- **IRI Relativization**: The inverse of resolution. Create a relative IRI from two absolute IRIs.
- **Efficient Component Access**: Get IRI components (`scheme`, `authority`, `path`, etc.) without extra allocations.
- **Checked and Unchecked Variants**: Use checked functions (`Parse`, `Resolve`) for safety with unknown input, or their `_Unchecked` counterparts for maximum performance when the input is known to be valid.
- **Built-in JSON Support**: `Iri` and `IriRef` types implement `json.Marshaler` and `json.Unmarshaler` for easy integration with web APIs.

## Installation

```sh
go get github.com/jplu/trident/iri
```

## Quick Start

### 1. Parsing and Validation

You can parse a string into a validated `Iri` (which must be absolute) or an `IriRef`.

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

	// IriRef can handle relative references
	relativeRef, err := iri.ParseIriRef("../other?key=val")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Parsed relative reference: %s\n", relativeRef)
}
```

### 2. Resolving a Relative IRI

The core use case is resolving a relative reference against a base IRI.

```go
package main

import (
	"fmt"
	"github.com/jplu/trident/iri"
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

## The `Iri` vs. `IriRef` Distinction

This library's design provides compile-time safety by separating two concepts:

-   **`IriRef`**: An "IRI Reference". This can be a full, absolute IRI or a relative reference (like `/path/only`, `../`, or `#fragment`). This type should be used when you are handling input that might not be a fully-qualified IRI.

-   **`Iri`**: A guaranteed absolute IRI. It always has a scheme. You cannot create an `Iri` from a relative reference. This is useful for types that require a non-ambiguous, absolute identifier, such as a base IRI for resolution.

You can convert an `IriRef` to an `Iri` using `NewIriFromRef()`, which will return an error if the `IriRef` is not absolute.

## JSON Marshaling and Unmarshaling

Both `Iri` and `IriRef` implement the `json.Marshaler` and `json.Unmarshaler` interfaces. The unmarshaler automatically validates the IRI string.

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/jplu/trident/iri"
)

type Config struct {
	BaseURL    *iri.Iri `json:"baseUrl"`
	Stylesheet *iri.IriRef `json:"stylesheet"`
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

This pacakge is inspired by the excellent Rust [**oxiri**](https://github.com/oxigraph/oxiri) library from the Oxigraph project.

## License

This project is licensed under the APACHE License - see the `LICENSE` file for details.
