# ADR-0002 — Interactive prompts with `charmbracelet/huh`

Status: Accepted
Date: 2026-10-11
Deciders: @isokolovskii, Pi (GLM)

## Context

The wizard needs: single-select list (type), text input with validation and max length (scope, subject), multi-line text (body), confirm (breaking/issues/confirm), sensible defaults, and first-class Windows support. The repo's strict linter and testability demands mean prompts must be isolatable from business logic.

Candidates:

| Library | Status | Widgets | Testing story |
| --- | --- | --- | --- |
| **charmbracelet/huh** | active, high momentum | select/input/textarea/confirm/note, themes | forms driven via bubbletea program with injectable stdin/stdout |
| AlecAivazis/survey | archived (successor → huh) | rich | none maintained |
| juancano/promptui | stale since 2021 | select/input | none |
| Manifold Scholastic/tui? pterm | active | interactive text only | weaker select/confirm UX |

## Decision

Use `charmbracelet/huh` behind our own `Prompter` port interface; no prompt library types leak outside the prompts package.

```go
// internal/prompts
type Prompter interface {
    Select(message string, options []Option, def int) (int, error)
    Input(message string, def string, validate func(string) error) (string, error)
    MultiLine(message string, def string) (string, error)
    Confirm(message string, def bool) (bool, error)
}
```

## Rationale

- huh is the maintained successor of survey (which cz-cli's ecosystem used); select/input/textarea/confirm map 1:1 to our question set.
- Consistent TUI quality on macOS/Linux/Windows terminals (bubbletea runtime), including CJK-safe width handling for the subject length guard.
- Interface boundary (ports & adapters) keeps the wizard logic pure: unit tests use a scripted fake `Prompter`; the huh implementation gets a thin smoke test plus manual/e2e verification.

## Consequences

- New direct dependency `github.com/charmbracelet/huh` (pulls bubbletea/lipgloss family) — acceptable for a CLI; still a single static binary, CGO-free.
- Prompt styling centralized in one theme; not configurable in v1.
- If huh proves unsuitable mid-flight, only `internal/prompts/huh.go` is rewritten.
