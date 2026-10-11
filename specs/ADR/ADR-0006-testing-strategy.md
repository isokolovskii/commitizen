# ADR-0006 — Testing strategy: ports, real git, layered suites

Status: Accepted
Date: 2026-10-11
Deciders: @isokolovskii, Pi (GLM)

## Context

Coverage is low (2 unit files). The repo lints with `paralleltest`, `funlen`, table-driven expectations, and CI runs `go test -race` on 3 OSes plus Codecov. The wizard is interactive — TUIs are classically hard to test — and git behavior is environment-sensitive.

## Decision

Four layers, all running in `make test` / CI:

1. **Unit** (pure logic, no I/O): message builder (exists, extend), config parsing/resolution (fixture files), adapter question-flow with a scripted **fake `Prompter`** (ADR-0002 port), retry cache store, message formatting edge cases (long subjects, breaking-only, footer ordering).
2. **Module/integration** (`internal/git`, cmd plumbing against real git): temp repo fixtures via `t.TempDir()` + isolated env (`GIT_CONFIG_GLOBAL` pointed at a fixture, `GIT_AUTHOR_*`/`GIT_COMMITTER_*` pinned) — repo detection, staged check (empty/non-empty), real `git commit` through the wrapper asserting `git log --format=%B`.
3. **E2E** (`e2e/` package, `TestMain` builds the binary once into `t.TempDir()`): built binary driven via `exec.Command` with piped stdin in **non-interactive paths**: `version`, `ls`, `example`, `init` (+`--force` refusal), flags-mode `commit --dry-run` (golden message), `--hook` writing `COMMIT_EDITMSG`, `--retry` replay, error exits (outside repo, no stage, bad `.czrc`). Interactive-TTY runs are out of scope (flaky across CI OSes); the wizard's decision logic is covered at layer 1 by the fake prompter.
4. **Golden files** for message rendering and `example`/`ls` output (`-update` flag via `goldie`-style helper or hand-rolled `testdata/` compare) to lock formats the changelog ecosystem depends on.

Conventions: `t.Parallel()` everywhere the linter demands; table-driven cases; `testify/assert` (already a dependency); no sleeps; no network; coverage gate ≥ 80% on `internal/...` (NFR-3) tracked by Codecov.

## Rationale

- Ports at the two boundaries (prompts, git) buy full decision-logic coverage without TTY/tty emulation, keeping CI fast and OS-portable.
- Real-git integration tests encode actual user contracts (hooks config, message normalization) that mocks would fake.
- E2E on the built binary catches ldflags/version wiring and flag-surface regressions unit tests can't see.

## Consequences

- `internal/git` must inject env cleanly per test (helper in `internal/git/gittest`).
- CI matrix (ubuntu/macos/windows) exercises path handling (`filepath`, `\n` normalization) for free.
- A small `e2e/testdata/` tree of repos-in-files or setup funcs; setup funcs chosen to stay Windows-safe.
