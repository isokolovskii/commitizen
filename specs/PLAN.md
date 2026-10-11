# Delivery Plan — Commitizen for Go

Source docs: [ASSESSMENT.md](ASSESSMENT.md) · [PRD.md](PRD.md) · [ADR/](ADR/)
Workflow: one branch/PR per milestone, conventional commits, Co-authored-by trailers (Pi harness + GLM model), CodeRabbit review before merge.

## M0 — Specs (branch `docs/specs`) ✅ this PR

- [x] Assessment of current state vs cz-cli
- [x] PRD with MUST/SHOULD scope + acceptance criteria
- [x] ADR-0001 … ADR-0007 (framework, prompts, adapters, config, git, testing, release)
- [x] This plan

## M1 — Config layer (branch `feat/config`)

Depends on: ADR-0003, ADR-0004.

- [ ] `internal/config`: `Config` struct (schema per ADR-0004), `Load(root string)` with first-wins resolution (`.czrc` → `package.json:config.commitizen` → `~/.czrc`), `DisallowUnknownFields`, human-readable errors
- [ ] `internal/config`: defaults (conventional type table, `maxSubjectWidth: 100`, `confirm: true`)
- [ ] `internal/adapters`: `Adapter` interface + registry (map), `conventional` adapter skeleton, `Names()` for `ls`/errors
- [ ] Unit tests: valid/invalid fixtures, resolution order, unknown key/type errors, types override validation
- [ ] Wire nothing into `cmd/` yet (framework-free packages land first)

Acceptance: `go test ./internal/...` green with new config+adapters suites; linter clean.

## M2 — Interactive commit flow (branch `feat/interactive-commit`)

Depends on: M1, ADR-0001/2/5.

- [ ] `internal/git`: `Runner` interface (`exec.CommandContext`), `Root()`, `StagedFiles()`, `Commit(msg)`, gittest helper (temp repo, pinned env)
- [ ] `internal/prompts`: `Prompter` interface + `huh` implementation + scripted fake for tests
- [ ] `internal/adapters/conventional`: full question flow (type → scope → subject → body → breaking → issues → confirm), `Example()`, skipQuestions/maxSubjectWidth/defaults from config
- [ ] `cmd/commit`: interactive default (`DefaultCommand`), flags mode retained + `--dry-run`, `--hook [FILE]`, `--retry`, `--retry-without-editor`; staged-files gate; clean stderr errors via `cli.Exit`
- [ ] `cmd/init` (`--force`), `cmd/ls`, `cmd/example`; `version` prints to stdout without `log`
- [ ] Retry cache at `os.UserCacheDir()/commitizen/retry.json`, keyed by repo path
- [ ] Unit tests (fake prompter, all decision branches) + integration tests (real git) + e2e tests (built binary: version/ls/example/init/hook/dry-run/retry/errors) per ADR-0006
- [ ] README: Usage TODO sections replaced (install, wizard, `.czrc`, init, hooks incl. lefthook recipe, retry, non-interactive)

Acceptance: PRD FR-1…FR-8 demonstrable; `make test` green 3 OS; coverage ≥ 80% `internal/…`.

## M3 — Test hardening (branch `test/coverage`)

Depends on: M2.

- [ ] Coverage gap sweep (`go tool cover -func`): remaining branches in builder/config/git/prompts
- [ ] Message-builder edge cases: max-width subjects, breaking-only footer, issue+breaking ordering, CRLF handling
- [ ] `--hook` non-TTY behavior test; retry-cache corruption test
- [ ] CI: add coverage diff comment gate (Codecov already wired; verify patch coverage)

Acceptance: coverage report shows every exported function exercised or explicitly `//nolint`-documented.

## M4 — Release readiness (branch `chore/release-readiness`)

Depends on: M2 (docs finalized there).

- [ ] ADR-0007: `git-cz` build in `.goreleaser.yaml`; drop Snapcraft from `release.yml`; unify goreleaser version pins; drop `go generate` hook
- [ ] Local `goreleaser release --snapshot --clean` verification; CI snapshot green
- [ ] README final pass (badges, git-cz alias instructions, commitizen-friendly badge snippet)
- [ ] CHANGELOG regenerated at next `make version` (documented, not run on feature branches)

Acceptance: snapshot release artifacts contain `commitizen` + `git-cz` for the full OS/arch matrix; release workflow contains no dead steps.

## Sequencing & review

1. PRs merge **fast-forward only via squash-free merge commits**? — No: plain merge to preserve granular conventional commits (changelog depends on them).
2. Each PR: push branch → open PR (compare URL provided to owner since `gh` is unavailable) → CodeRabbit review → address comments → merge.
3. Conventional commit types per repo changelog groups: `feat`, `fix`, `docs`, `refactor`, `test`, `build`, `chore`, `ci`.
4. Version bump + tag happen once, after M4, by the owner via `make version` from `main`.
