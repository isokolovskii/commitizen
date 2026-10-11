# ADR-0001 — Keep `urfave/cli/v3` as the CLI framework

Status: Accepted
Date: 2026-10-11
Deciders: @isokolovskii, Pi (GLM)

## Context

The skeleton already uses `urfave/cli/v3` (root command + `commit`/`version` subcommands, shell completion). The interactive wizard adds requirements: a default subcommand (bare `commitizen` must commit), flag aliases, per-command flags, and clean error surfacing. Candidate alternatives: spf13/cobra, alecthomas/kong, hand-rolled flag parsing.

## Decision

Keep `urfave/cli/v3`.

## Rationale

- Already wired and working; switching costs nothing to the product but churns review.
- v3 supports `DefaultCommand` (bare invocation → wizard), `Destination` flag binding already used by `cmd/commit.go`, and shell completion — all required features.
- Cobra's strength (huge ecosystem, generated docs) is not needed for a 6-command CLI; kong's struct-tags style conflicts with the existing flag-Destination pattern.
- `cli.Exit(err, 1)` gives us `main()`-free error plumbing: actions return errors, the framework prints them to stderr and sets the exit code — removes today's `log.Fatalf` noise (see PRD NFR-6).

## Consequences

- `cmd/commitizen.go` gains `DefaultCommand: "commit"`; root `Version` handling stays.
- `main.go` simplifies to `cmd.Lefthook().RunAndExit()`-style flow; no `log` package in the CLI path.
- If a future need outgrows urfave/cli, the migration point is isolated in `cmd/` (all other packages are framework-free).
