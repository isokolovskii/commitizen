# ADR-0004 — Configuration: `.czrc` JSON, package.json compat, home fallback

Status: Accepted
Date: 2026-10-11
Deciders: @isokolovskii, Pi (GLM)

## Context

cz-cli resolves adapter config from (in order): repo-local `package.json` `config.commitizen`, repo-local `.czrc`, then global `~/.czrc`. The Go port needs a resolution story that feels native to cz users, works in polyglot repos, and stays strict about unknown keys.

## Decision

Resolution order — **first file that exists wins, whole file, no merging**:

1. `<git-root>/.czrc` (JSON)
2. `<git-root>/package.json` → `config.commitizen` key (JSON, same schema)
3. `~/.czrc` (global default for all repos)

Schema (v1 subset, deliberately small — PRD FR-3):

```json
{
  "path": "conventional",
  "types": [
    { "value": "feat", "name": "feat", "description": "A new feature" }
  ],
  "maxSubjectWidth": 100,
  "defaultType": "", "defaultScope": "", "defaultSubject": "",
  "defaultBody": "", "defaultIssues": "",
  "skipQuestions": ["scope", "body", "breaking", "issues", "confirm"],
  "confirm": true
}
```

Rules:

- Unknown keys are a **hard error** listing the offending key and the allowed set (fail-fast beats silently ignored config; cz's 30-option surface is not v1).
- `types` replaces the built-in table wholesale when present; each entry requires `value` + `description`, unique `value`s.
- `maxSubjectWidth` default 100 (conventional-changelog ecosystem default), must be ≥ 1.
- Defaults: `confirm: true`, everything else empty/false; `skipQuestions` accepts only the five known question names.
- Parsing: `encoding/json` with `DisallowUnknownFields`; error messages rewritten to be human-actionable.

## Rationale

- `.czrc` + `~/.czrc` + `config.commitizen` are exactly the muscle memory cz-cli users already have; support costs three file reads.
- First-wins (no merge) matches cz-cli and keeps "which value won?" trivially explainable.
- Strictness (unknown-key errors) is safe at v1 scale and prevents the silent-drift bugs a lax reader would cause.

## Consequences

- `internal/config` is a pure, fully unit-testable package (fixtures in testdata).
- Adding schema keys later is backward compatible; removing/renaming requires an ADR.
- TOML/YAML `.czrc` variants are explicitly out of scope (cz-cli's format is JSON; polyglot-tempting but not worth the dependency).
