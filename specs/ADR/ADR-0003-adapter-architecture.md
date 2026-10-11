# ADR-0003 — Adapter architecture: built-in registry, config-driven customization

Status: Accepted
Date: 2026-10-11
Deciders: @isokolovskii, Pi (GLM)

## Context

cz-cli's power is its adapter ecosystem: any npm package that calls Inquirer can be loaded at runtime via `config.commitizen.path` (resolved with `require.resolve`). Go binaries are static; `plugin` packages are Linux/macOS-only, cgo-bound, and impractical for distribution. We must decide how "adapters" exist in the Go port.

Options considered:

1. **Built-in registry + config-driven customization (chosen)** — adapters are compiled in; `.czrc` customizes the conventional adapter (type table, defaults, lengths, skip list).
2. External-process adapters (`path: "/cmd"` protocol over stdin/stdout) — powerful, but a protocol design + security surface (arbitrary command execution from a repo-owned config) out of proportion for v1.
3. Embed a JS runtime (goja) to run real cz adapters — heavy, incompatible subset risk, big binary; rejected.
4. Go plugins (`plugin` package) — not portable to Windows/plan9, no goreleaser story; rejected.

## Decision

v1 ships a **built-in adapter registry**:

```go
// internal/adapters
type Adapter interface {
    Name() string
    Description() string
    Example() (string, error)          // commitizen example
    Prompt(p prompts.Prompter, cfg config.Adapter) (conventional.Commit, error)
}
```

- `conventional` (default): mirrors cz-conventional-changelog's question flow; its type table, defaults, `maxSubjectWidth`, `skipQuestions`, and `confirm` are overridable from `.czrc` (ADR-0004).
- Unknown `path` values error with the registry list attached.
- v1.x may add an `external` adapter with a line-oriented JSON protocol; the `Adapter` interface and config schema are designed so it can slot in without breaking `.czrc`.

## Rationale

- Covers the 90% case (conventional-changelog conventions) with zero distribution complexity.
- "Custom convention" projects are still served: overriding the type table and questions covers cz-customizable's common usage without code.
- Keeps the security story trivial: repo config can never execute code.

## Consequences

- Not drop-in compatible with the npm adapter ecosystem; documented in README ("adapters are built-in, request yours via PR").
- Registry lookup is pure Go map — trivially testable; `commitizen ls` lists adapters (PRD FR-3).
