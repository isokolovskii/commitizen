// Package adapters provides the built-in adapter registry used to resolve a
// configured adapter name into a wizard implementation.
package adapters

import (
	"errors"
	"fmt"
	"slices"
)

type (
	// Adapter is the contract every commit wizard adapter fulfills.
	Adapter interface {
		// Name returns the registry key of the adapter.
		Name() string
		// Description returns a one line explanation shown in listings.
		Description() string
		// Example returns a sample commit message produced from default answers.
		Example() (string, error)
	}

	// Registry resolves adapter names to instances.
	Registry struct {
		// ByName indexes adapters by their registry key.
		byName map[string]Adapter
	}
)

var (
	// ErrUnknownAdapter is returned when configuration names an adapter that
	// is not built in.
	ErrUnknownAdapter = errors.New("unknown adapter")
	// ErrEmptyName is returned when an adapter registration has no name.
	ErrEmptyName = errors.New("adapter name must not be empty")
	// ErrDuplicateAdapter is returned when two adapters share a name.
	ErrDuplicateAdapter = errors.New("duplicate adapter")
)

// NewRegistry indexes the given adapters by name.
func NewRegistry(all ...Adapter) (*Registry, error) {
	byName := make(map[string]Adapter, len(all))
	for _, adapter := range all {
		if adapter.Name() == "" {
			return nil, ErrEmptyName
		}

		if _, ok := byName[adapter.Name()]; ok {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateAdapter, adapter.Name())
		}

		byName[adapter.Name()] = adapter
	}

	return &Registry{byName: byName}, nil
}

// Get returns the adapter registered under name.
func (registry *Registry) Get(name string) (Adapter, error) {
	adapter, ok := registry.byName[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q, available: %v", ErrUnknownAdapter, name, registry.Names())
	}

	return adapter, nil
}

// Names returns all registered adapter names in alphabetical order.
func (registry *Registry) Names() []string {
	names := make([]string, 0, len(registry.byName))
	for name := range registry.byName {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}
