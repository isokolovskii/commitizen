// Package conventional implements the built-in conventional commits adapter,
// the default wizard of this CLI.
package conventional

import (
	"fmt"

	"github.com/isokolovskii/commitizen/internal/adapters"
	"github.com/isokolovskii/commitizen/internal/config"
	"github.com/isokolovskii/commitizen/internal/conventional"
)

type (
	// Adapter answers registry queries and produces example messages for
	// the conventional commits style.
	Adapter struct {
		// Cfg is the resolved Commitizen configuration.
		cfg *config.Config
	}

	// BuiltinType describes one default commit type of the wizard.
	builtinType struct {
		// Value is the commit type keyword.
		value string
		// Description explains when to use the type.
		description string
	}
)

const (
	// Name is the registry and configuration key of this adapter.
	Name = "conventional"

	featTypeValue = "feat"
)

var (
	// BuiltinTypes mirrors the cz-conventional-changelog type list.
	builtinTypes = []builtinType{
		{value: featTypeValue, description: "A new feature"},
		{value: "fix", description: "A bug fix"},
		{value: "docs", description: "Documentation only changes"},
		{value: "style", description: "Changes that do not affect the meaning of the code"},
		{value: "refactor", description: "A code change that neither fixes a bug nor adds a feature"},
		{value: "perf", description: "A code change that improves performance"},
		{value: "test", description: "Adding missing tests or correcting existing tests"},
		{value: "build", description: "Changes that affect the build system or external dependencies"},
		{value: "ci", description: "Changes to our CI configuration files and scripts"},
		{value: "chore", description: "Other changes that do not modify src or test files"},
		{value: "revert", description: "Reverts a previous commit"},
	}

	// Compile-time assertion that Adapter fulfills the registry contract.
	_ adapters.Adapter = (*Adapter)(nil)
)

// New creates the conventional adapter for the given configuration.
func New(cfg *config.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

// Name implements adapters.Adapter.
func (*Adapter) Name() string {
	return Name
}

// Description implements adapters.Adapter.
func (*Adapter) Description() string {
	return "Conventional Commits wizard (feat, fix, docs, and friends)"
}

// Example implements adapters.Adapter.
func (a *Adapter) Example() (string, error) {
	subject := "add hello world handler"
	if a.cfg != nil && a.cfg.DefaultSubject != "" {
		subject = a.cfg.DefaultSubject
	}

	message, err := conventional.BuildCommitMessage(&conventional.Commit{
		Type:  featTypeValue,
		Scope: "core",
		Title: subject,
		Body:  "Provide a working baseline for the generated example message.",
	})
	if err != nil {
		return "", fmt.Errorf("building example message: %w", err)
	}

	return message, nil
}

// Types resolves the effective type table: configuration overrides win, then
// the built-in defaults.
func Types(cfg *config.Config) []config.Type {
	if cfg != nil && len(cfg.Types) > 0 {
		return cfg.Types
	}

	types := make([]config.Type, 0, len(builtinTypes))
	for _, builtin := range builtinTypes {
		types = append(types, config.Type{
			Value:       builtin.value,
			Description: builtin.description,
		})
	}

	return types
}
