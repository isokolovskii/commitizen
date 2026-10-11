# PRD — Commitizen for Go

Version: 1.0 (2026-10-11)
Status: Approved for implementation
Reference product: [commitizen/cz-cli](https://github.com/cz-cli) + [cz-conventional-changelog](https://github.com/conventional-changelog/cz-conventional-changelog)

## 1. Problem

Teams using [Conventional Commits](https://www.conventionalcommits.org) reject badly formatted commits at hook time, after the author has already composed the message. cz-cli solves this for Node.js projects by prompting the contributor through the required fields **at commit time**. Go projects (and polyglot contributors who dislike Node toolchains) have no equally polished, single-binary, dependency-free equivalent.

## 2. Vision & goals

A single static Go binary, `commitizen` (alias `git-cz`), that:

1. Works exactly like `git cz` — interactive wizard at commit time.
2. Is configurable per repo via a `.czrc` file, with the same mental model as cz-cli.
3. Installs with `go install` or one downloaded binary; runs everywhere goreleaser ships (linux/darwin/windows/freebsd/openbsd × amd64/arm64).
4. Is boring to operate: no runtime dependencies beyond the `git` executable.

Non-goals (v1): changelog generation (recommend git-cliff), commit linting (recommend commitlint equivalents), dynamic loading of Node.js adapters, `git commit` flag parity (`--amend`, `--no-verify` pass-through arrives in v1.x).

## 3. Users & personas

- **Contributor** — wants to commit correctly without memorizing the convention; interacts only with the wizard.
- **Maintainer** — wants to enforce a convention repo-wide; edits `.czrc`, runs `commitizen init`, wires optional hooks.
- **Release engineer** — tags releases via `make version`; consumes goreleaser artifacts; expects no release-pipeline changes in v1 beyond the `git-cz` alias binary.

## 4. Functional requirements

Priorities: **MUST** = v1 GA gate; **SHOULD** = v1 if cheap, else v1.x; **WON'T** = out of scope.

### FR-1 Interactive commit wizard (MUST — the core)

- `commitizen` with no subcommand, `commitizen commit`, or `git-cz` starts the adapter wizard.
- Question set (conventional adapter, mirrors cz-conventional-changelog):
  1. **Type** — selectable list with human descriptions (feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert).
  2. **Scope** — optional text; lowercased by default (disable-able later); may be skipped per config.
  3. **Subject** — required, imperative-mood hint shown, hard cap at `maxSubjectWidth` (default 100 chars); empty input re-prompts.
  4. **Body** — optional multi-line text.
  5. **Breaking change?** — confirm; if yes, ask for description; renders `!` in header + `BREAKING CHANGE:` footer.
  6. **Issues affected?** — confirm; if yes, ask for references; renders footer line.
  7. **Confirm** — render the full final message and ask to proceed / edit-free abort (configurable, default on).
- On confirm: run `git commit -m <message>` in the repo; on success print the message; wizard errors must abort **without committing**.
- Ctrl+C / "abort" exits non-zero and changes nothing.

### FR-2 Non-interactive mode (MUST — already exists, keep)

- `commitizen commit --type … --title … [--scope --body --breaking --issue]` builds and prints/commits the message without prompts. Used by scripts, bots, and tests.
- New: `--dry-run` prints the message and exits without running `git commit` (default behavior today becomes explicit).
- Validation errors (missing type/title) exit non-zero with a clear message.

### FR-3 Adapter selection & configuration (MUST)

- Config resolution order (first hit wins), matching cz-cli's model adapted to Go:
  1. `<git-root>/.czrc`
  2. `config.commitizen` key inside `<git-root>/package.json` (polyglot repos)
  3. `~/.czrc` (global default)
- `.czrc` shape (JSON):
  ```json
  {
    "path": "conventional",
    "types": [{ "value": "feat", "name": "feat", "description": "A new feature" }],
    "maxSubjectWidth": 100,
    "defaultType": "", "defaultScope": "", "defaultSubject": "", "defaultBody": "", "defaultIssues": "",
    "skipQuestions": [],
    "confirm": true
  }
  ```
- `path` accepts built-in adapter names (`conventional` — default). Unknown values fail with a list of available adapters (future: external adapters, ADR-0003).
- `commitizen ls` lists available adapters and, for the conventional adapter, the effective type table.

### FR-4 `commitizen init` (MUST)

- Writes a `.czrc` into the current git repo root pre-selecting an adapter: `commitizen init conventional [--force]`.
- Refuses to overwrite an existing `.czrc` unless `--force`; refuses outside a git repo; prints what it did and a Commitizen-friendly badge snippet.

### FR-5 Hook mode (MUST)

- `commitizen commit --hook [FILE]` (used from `prepare-commit-msg`): runs the wizard non-interactively-safe, then writes the final message into `FILE` (defaults to `.git/COMMIT_EDITMSG`) instead of committing, letting `git commit` proceed.
- Documented recipe for lefthook and raw git hooks in the README (with the `exec < /dev/tty` trick).

### FR-6 Retry (SHOULD)

- `commitizen commit --retry` re-runs the last wizard answers for this repo (keyed by absolute repo path), rebuilding the message and committing — for the "pre-commit hook failed, fix, re-run" loop.
- `--retry-without-editor` behaves identically minus any editor step (kept for cz-cli flag parity).
- Cache lives at `$XDG_CACHE_HOME|~/Library/Caches|…/commitizen/retry.json` (via `os.UserCacheDir`), never inside the repo; stale/absent cache → friendly error, not a crash.

### FR-7 Repository awareness (MUST)

- Every commit-path command: detect repo (`git rev-parse --show-toplevel`), fail cleanly outside one.
- Before committing: verify staged changes (`git diff --cached --name-only`); empty stage → error suggesting `git add` (or `--allow-empty` pass-through later; `git cz -a`-style staging is v1.x).
- Respect `GIT_DIR`/worktrees implicitly by shelling out to git.

### FR-8 Example & version (MUST)

- `commitizen example` prints a fully filled sample message from the active adapter.
- `commitizen version` prints `name version (commit)` to stdout, no log noise.

### FR-9 Distribution (MUST — mostly exists)

- goreleaser v2 pipeline unchanged in mechanics; **add** a `git-cz` binary build so `git cz` works after installing the release archive (ADR-0007).
- Release remains: `make version` from `main` → tag → `release.yml`.

## 5. Non-functional requirements

- **NFR-1 Zero deps for users**: static binary; only external process dependency is `git`.
- **NFR-2 Testability**: prompt layer behind an interface; golangci strict ruleset stays green; `t.Parallel` everywhere per repo linters.
- **NFR-3 Coverage**: ≥ 80% line coverage on `internal/…`; unit + integration (real git in temp dirs) + e2e (built binary, non-interactive paths) suites (ADR-0006).
- **NFR-4 Portability**: no TTY requirement for non-interactive flows; wizard requires a TTY and errors gracefully under pipes (`--dry-run`/flags still work).
- **NFR-5 Speed**: cold start < 100 ms excluding wizard; no network calls ever.
- **NFR-6 Errors**: user-facing errors are actionable sentences, not stack traces; `log` package dropped from the CLI surface.

## 6. Acceptance criteria (v1 GA)

1. In a clean clone of any repo, `commitizen` starts the wizard; completing it produces a `git log` entry matching Conventional Commits and the `.czrc` in force.
2. All FR-MUST items demo-able via documented commands; FR-SHOULD (retry) works per spec.
3. `make test` green on linux/darwin/windows; e2e suite green in CI; coverage gate satisfied.
4. `goreleaser release --snapshot --clean` succeeds locally; CI snapshot build unchanged-green.
5. README documents install, per-repo setup (`init`, `.czrc`), hook recipes (lefthook + raw), retry, non-interactive usage — the current TODO sections are gone.
6. Every merged PR carries conventional-commit titles and Co-authored-by trailers (Pi harness + GLM model).

## 7. Risks

| Risk | Mitigation |
| --- | --- |
| TUI library can't be driven in tests | Prompter interface + fake (ADR-0002/0006); wizard logic tested without TTY |
| Config creep (copying cz-conventional-changelog's 30 options) | Ship the documented v1 subset; unknown keys rejected with hint (fail-fast, no silent ignore) |
| Hook mode hangs without TTY in CI | Document `exec < /dev/tty`; detect non-TTY and exit 0 with unchanged COMMIT_EDITMSG (git proceeds) |
| Windows console quirks in prompts | huh/bubbletea has first-class Windows support; e2e suite runs on CI windows runner |

## 8. Milestones

| Milestone | Content | PR |
| --- | --- | --- |
| M0 Specs | Assessment, PRD, ADRs, plan | `docs/specs` |
| M1 Config | `.czrc`/package.json/home resolution + adapter registry + tests | `feat/config` |
| M2 Interactive commit | git layer, prompter, wizard, hook/retry/dry-run, init/ls/example, README usage | `feat/interactive-commit` |
| M3 Test hardening | coverage gaps, module + e2e suites polish | `test/coverage` |
| M4 Release readiness | `git-cz` build, goreleaser/CI cleanup, docs final pass | `chore/release-readiness` |
