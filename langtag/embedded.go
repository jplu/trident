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
	"bytes"
	_ "embed" // Note the blank import for go:embed
	"errors"
)

//go:embed language-subtag-registry
var embeddedRegistryData []byte

// NewParser creates a new parser instance from the embedded IANA registry.
//
// IMPORTANT: This function parses the entire IANA registry on every call and is
// therefore an expensive operation. For performance, it is strongly recommended
// to call this function only once at application startup and reuse the returned
// parser instance throughout your application.
func NewParser() (*Parser, error) {
	if len(embeddedRegistryData) == 0 {
		return nil, errors.New("embedded language-subtag-registry file is empty or not found")
	}

	reader := bytes.NewReader(embeddedRegistryData)
	registry, err := ParseRegistry(reader)
	if err != nil {
		return nil, err
	}

	return &Parser{
		registry: registry,
	}, nil
}
