# ADR-0005 — Git integration via `os/exec` of the git CLI

Status: Accepted
Date: 2026-10-11
Deciders: @isokolovskii, Pi (GLM)

## Context

Commit-path features need: repo root detection, staged-files check, and commit execution (plus later pass-through flags). Candidates: shell out to the `git` binary (`os/exec`), or embed `github.com/go-git/go-git` (pure-Go implementation).

## Decision

Shell out to `git` through a narrow `internal/git` wrapper (`Runner` interface over `exec.CommandContext`), never linking go-git.

Operations v1: `rev-parse --show-toplevel`, `diff --cached --name-only` (staged check), `commit -m <msg>` (+ `--cleanup=strip`), `status --porcelain` (future niceties).

## Rationale

- Behavioral fidelity: the user's actual `git` applies their own config — hooks, `commit.gpgsign`, `core.hooksPath`, credential helpers, worktrees, sparse checkouts. A library re-implements these with subtle gaps (go-git notably ignores hooks and GPG signing — fatal for a commit tool).
- Zero-dependency distribution: binary size stays small; no library CVE surface.
- Testability: the `Runner` interface is faked in unit tests; integration tests assert against **real** git in temp repos (ADR-0006), which is the contract users care about.

## Consequences

- Requires `git` on PATH — acceptable (NFR-1 documents it); error message when missing is actionable.
- Output parsing limited to porcelain formats we control (`--porcelain`, `-z` where listing filenames to avoid quoting traps).
- Commit uses `-m` with the full message (trailing newline stripped; git normalizes), preserving cz-cli behavior.
