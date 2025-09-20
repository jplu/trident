<p align="center">
  <img alt="trident logo" src="assets/logo.png" height="200" />
  <h3 align="center">trident</h3>
  <p align="center">Fast RDF and SPARQL graphs for Go</p>
</p>

# Trident 🔱

> Wield the power of linked data in Go.

[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)
[![Github CI](https://github.com/jplu/trident/actions/workflows/ci.yaml/badge.svg)](https://github.com/jplu/trident/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/jplu/trident/graph/badge.svg?token=UNSE94Z9TL)](https://codecov.io/gh/jplu/trident)
[![Go Report Card](https://goreportcard.com/badge/github.com/jplu/trident)](https://goreportcard.com/report/github.com/jplu/trident)
[![GoDoc](https://godocs.io/github.com/jplu/trident?status.svg)](https://godocs.io/github.com/jplu/trident)
[![License: Apache 2.0](https://img.shields.io/github/license/jplu/trident.svg)](https://github.com/jplu/trident/blob/main/LICENSE)
[![stability-unstable](https://img.shields.io/badge/stability-unstable-yellow.svg)](https://github.com/emersion/stability-badges#unstable)

**Trident** is a modern, expressive Go package for working with RDF graphs and executing SPARQL queries. It aims to provide a simple and idiomatic Go API for creating, manipulating, and querying linked data.

The name "Trident" is inspired by the fundamental structure of RDF: the **triple** (subject, predicate, object), where each tine of the trident represents a component of the triple.

## Why Trident?

*   **Go Idiomatic:** Designed from the ground up to feel natural for Go developers, with a clean and minimal API.
*   **Performance:** Aims for efficient in-memory graph processing and query execution.
*   **Type-Safe:** Leverages Go's type system to provide a safer, more predictable way to build RDF terms and queries.
*   **Extensible:** Planned support for custom storage backends and additional RDF serialization formats.

## Features (Planned and not Exhaustive)

- [ ] In-memory RDF graph store.
- [ ] Parsers for popular RDF formats.
- [ ] Serializers for popular RDF formats.
- [ ] SPARQL engine.
- [ ] Support for RDF and SPARQL 1.2 standards.
- [ ] A fluent and intuitive API for graph manipulation.

## Installation

```bash
go get github.com/jplu/trident
```
*(Note: This is the intended installation path.)*

## Usage

*(Note: The following API is a proposal and subject to change as development progresses.)*

### Creating a Graph and Adding Triples

```go
package main

import (
	"fmt"
	"github.com/your-username/trident"
	"github.com/your-username/trident/rdf"
)

func main() {
	// Create a new in-memory graph
	g := trident.NewGraph()

	// Define some terms
	picasso := rdf.NewIRI("http://example.org/artists#picasso")
	name := rdf.NewIRI("http://xmlns.com/foaf/0.1/name")
	guernica := rdf.NewIRI("http://example.org/artworks#guernica")
	created := rdf.NewIRI("http://purl.org/dc/terms/created")
	painter := rdf.NewIRI("http://example.org/vocabulary#painter")

	// Add triples to the graph
	g.Add(rdf.NewTriple(picasso, name, rdf.NewLiteral("Pablo Picasso")))
	g.Add(rdf.NewTriple(guernica, created, rdf.NewLiteralWithDatatype("1937", rdf.XSDYear)))
	g.Add(rdf.NewTriple(guernica, painter, picasso))

	// Serialize the graph to Turtle format
	turtle, err := g.Serialize(trident.Turtle)
	if err != nil {
		panic(err)
	}

	fmt.Println(turtle)
}
```

### Querying with SPARQL

```go
package main

// ... (assuming graph `g` is populated from the previous example)

func queryGraph(g *trident.Graph) {
	query := `
		PREFIX foaf: <http://xmlns.com/foaf/0.1/>
		SELECT ?name
		WHERE {
			?artist foaf:name ?name .
		}
	`

	results, err := g.Query(query)
	if err != nil {
		panic(err)
	}
	defer results.Close()

	// The results object allows for easy iteration
	for results.Next() {
		solution := results.Solution()
		if name, ok := solution["name"]; ok {
			fmt.Printf("Found artist: %s\n", name.Value())
		}
	}

	if err := results.Err(); err != nil {
		panic(err)
	}
}
```

## Project Status

**ALPHA:** This project is in the very early stages of planning and development. The API is not stable and is likely to change significantly. It is not yet ready for production use.

## License

This project is licensed under the APACHE License - see the [LICENSE](./LICENSE) file for details.
