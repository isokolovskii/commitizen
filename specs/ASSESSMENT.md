# Project Assessment — Go Commitizen vs cz-cli

Date: 2026-10-11
Scope: snapshot of `main` @ `6b290dd`, compared against [commitizen/cz-cli](https://github.com/commitizen/cz-cli) (v4) and its reference adapter [cz-conventional-changelog](https://github.com/conventional-changelog/cz-conventional-changelog).

## 1. Current state of the Go port

The repository is a **working skeleton**, not yet usable as a commit wizard.

### What exists

| Area | State |
| --- | --- |
| CLI framework | `urfave/cli/v3`; root command `commitizen`, subcommands `commit`, `version`; shell completion enabled (`cmd/commitizen.go`) |
| Commit builder | `internal/conventional` — `Commit` struct + `BuildCommitMessage` producing `type(scope)!: title\n\nbody\n\nBREAKING CHANGE: ...\nIssue: ...`; table-driven tests (`message_test.go`) |
| Non-interactive commit | `cmd/commit.go` — flags-only mode: `--type` (required), `--scope`, `--title` (required), `--body`, `--breaking`, `--issue`; prints the message to stdout (does **not** run `git commit`) |
| Version | `internal/version` — hardcoded `1.0.0` + commit SHA injected via `-ldflags -X`; `Version()`; unused `Check()` semver helper (only referenced by its own test) |
| Release | `.goreleaser.yaml` v2 — 5 OSes / amd64+arm64 (386 excluded on non-Windows), binary + gz archives, sha256 checksums, grouped changelog, GitHub release; `release.yml` runs on tags with build-provenance attestations |
| Changelog | `cliff.toml` (git-cliff) + `make version` target: bumps `internal/version`, README install ref, CHANGELOG; commits + tags + pushes from `main` |
| CI | `test.yml` (3 OS × gotestsum + Codecov coverage/JUnit + goreleaser snapshot build artifact), `lint.yml` (golangci-lint v2), `codeql.yml`, `spelling.yml` (typos) |
| Repo hygiene | `.golangci.yml` (strict: paralleltest, funlen, err113, nlreturn, godot, mnd, cyclop, …), lefthook pre-commit (fmt+lint+tidy), `.coderabbit.yaml` (chill profile, PR reviews), dependabot, issue/PR templates, Codecov config |
| Docs | README with badges, install via `go install` / `go get -tool`; **Usage sections are TODO** |

### Tests today

- `internal/conventional/message_test.go` — 8 cases for the message builder.
- `internal/version/version_test.go` — `Check()` and `Version()`.
- Nothing covers the CLI layer, git integration, config, or end-to-end behavior. Coverage is low.

## 2. Feature gap vs cz-cli

cz-cli's value: at commit time, an adapter asks the contributor a fixed set of questions, renders a conventional message, and runs `git commit`. The Go port currently covers only "render a message from flags".

| # | cz-cli capability | Go port | Gap |
| --- | --- | --- | --- |
| 1 | Interactive commit wizard (`git cz` / `cz`) | — | **Missing entirely** — the core feature |
| 2 | Adapter system (`config.commitizen.path` / `.czrc` / `~/.czrc`) | — | Missing (Node module loading impossible in Go → needs Go-native design, see ADR-0003) |
| 3 | `commitizen init <adapter>` repo bootstrap | — | Missing |
| 4 | `commitizen version` | partial | Exists (should print cleanly to stdout) |
| 5 | `--hook` for `prepare-commit-msg` (write message into `COMMIT_EDITMSG`) | — | Missing |
| 6 | `--retry` / `--retry-without-editor` (reuse last answers, keyed by repo path, cache in home/temp) | — | Missing |
| 7 | Commit execution + pass-through of `git commit` flags (`git cz -a`) | prints to stdout only | Missing |
| 8 | Repo detection & sane errors outside a git repo / without staged changes | — | Missing |
| 9 | cz-conventional-changelog question set (type → scope → subject → body → breaking → issues) with descriptions and length guards | static flags | Missing prompts; builder supports the format already |
| 10 | Local/`package.json` + `.czrc` config resolution | — | Missing (Go equivalent defined in ADR-0004) |
| 11 | `git-cz` alias binary | — | Missing (goreleaser second build, ADR-0007) |

### Engineering gaps

- **Coverage**: no tests for `cmd/*`; no integration tests (real git repo); no e2e tests (built binary).
- **UX polish**: `main.go` logs errors via `log.Fatalf` (timestamps, stderr mixing); `version` uses `log.Println`.
- **Release leftovers**: `release.yml` installs Snapcraft and passes `SNAPCRAFT_STORE_CREDENTIALS`, but `.goreleaser.yaml` has no `snaps` section; goreleaser version pinned `v2.10.2` in release vs `~> v2` in test job.
- **Dead code**: `internal/version.Check` unused.
- **README**: Usage sections are TODO.

## 3. Constraints & inputs

- Go >= 1.25 (`.go-version` 1.25.4), CGO disabled in release builds.
- Strict golangci-lint v2 ruleset (enforces `t.Parallel`, table-driven style, error wrapping with `%w`, no magic numbers, line length, doc comments ending with periods…).
- Conventional Commits enforced on the repo itself (git-cliff + goreleaser changelog grouping + CodeRabbit title check).
- Releases: tag-driven via `make version` from `main`; artifacts via goreleaser GitHub release.

## 4. Decision

Proceed with the delivery plan in [PLAN.md](PLAN.md): specs + ADRs first (this PR), then config, interactive commit flow, tests, and release readiness — one feature branch/PR each, conventional commits, Co-authored-by trailers for the Pi harness and the GLM model.
