Commitizen
----------

[![codecov](https://codecov.io/github/isokolovskii/commitizen/graph/badge.svg?token=2ZDYRhKDly)](https://codecov.io/github/isokolovskii/commitizen) [![Lint](https://github.com/isokolovskii/commitizen/actions/workflows/lint.yml/badge.svg)](https://github.com/isokolovskii/commitizen/actions/workflows/lint.yml) [![Test](https://github.com/isokolovskii/commitizen/actions/workflows/test.yml/badge.svg)](https://github.com/isokolovskii/commitizen/actions/workflows/test.yml) [![CodeQL](https://github.com/isokolovskii/commitizen/actions/workflows/codeql.yml/badge.svg)](https://github.com/isokolovskii/commitizen/actions/workflows/codeql.yml)

CLI prompt tool for creation of [Conventional Commits](https://www.conventionalcommits.org). No more pain with commits rejected by validation tools or digging through [CONTRIBUTING.md](CONTRIBUTING.md) to find preferred commit format.

Based on wonderful work of [Commitizen](https://github.com/commitizen) with their [cz-cli](https://github.com/commitizen/cz-cli) which works like a charm in Node.js based projects.

This project replicated **cz-cli** way of work for Go projects

Installation
------------

Download a prebuilt binary from the [releases page](https://github.com/isokolovskii/commitizen/releases) — every release ships `commitizen` and `git-cz` for Linux, macOS, Windows, FreeBSD and OpenBSD on amd64 and arm64, for example:

```text
commitizen_v1.0.0_MacOS_x86_64
git-cz_v1.0.0_Linux_aarch64.gz
```

With **Go** (>= 1.25):

```bash
go install github.com/isokolovskii/commitizen@latest
```

-	or as a go tool

	```bash
	go get -tool github.com/isokolovskii/commitizen
	```

Usage
-----

Start the interactive wizard from the root of a git repository with staged changes:

```bash
commitizen
```

The wizard asks for a commit type, optional scope, subject, optional body, breaking changes, closing issues and a final confirmation, then runs `git commit` with the resulting message. `commitizen` is also the default command, so plain `commitizen` behaves the same way.

The releases also ship a `git-cz` binary: drop it on your `PATH` and the wizard is available as `git cz`, just like the original Node.js tool.

Using the command line tool
---------------------------

Skip the wizard and build a message from flags:

```bash
commitizen commit --type fix --scope core --title "Patch tricky bug" --body "Explain what went wrong" --issue JIRA-1
commitizen commit --type fix --title "Patch tricky bug" --dry-run   # print the message without committing
```

Run `commitizen commit --help` for the full flag list. `commitizen ls` lists available adapters, `commitizen example` prints an example message and `commitizen version` prints the current version.

### Configure the wizard

Create a `.czrc` file (or add a `config.commitizen` section to `package.json`) to pick an adapter, define custom commit types, skip questions and more:

```json
{
  "path": "conventional",
  "types": [
    { "value": "sparkle", "description": "Add glitter to the codebase" }
  ],
  "skipQuestions": ["body", "issues"],
  "maxSubjectWidth": 72
}
```

`commitizen init conventional` writes this file for you; add `--force` to overwrite an existing one.

### Retry a failed commit

If the wizard finished but `git commit` failed (for example a failing pre-commit hook), fix the problem and replay the cached message without answering the questions again:

```bash
commitizen commit --retry
commitizen commit --retry-without-editor   # skip the editor as well
```

As git hook with Lefthook
-------------------------

`--hook` integrates with `prepare-commit-msg`. With [Lefthook](https://lefthook.dev), add this to `lefthook.yml`:

```yaml
prepare-commit-msg:
  commands:
    commitizen:
      run: exec < /dev/tty && commitizen --hook "$1" || true
```

The `exec < /dev/tty` trick hands the terminal back to the wizard, because git hooks run without a TTY by default.

Development
-----------

The Makefile drives the whole toolchain through `go tool`, so no global installs are needed:

```bash
make build   # compile the commitizen binary
make test    # run the suite with the race detector
make lint    # golangci-lint fmt + run
```

Release artifacts are built with [goreleaser](https://goreleaser.com); a dry run against your current tree works without a tag:

```bash
go tool goreleaser release --snapshot --skip=publish --skip=validate --clean
```

Authors and Contributors
------------------------

@isokolovskii (Ivan Sokolovskii, author)

Special thanks to @JimTheDev, whose [cz-cli](https://github.com/commitizen/cz-cli) project makes conventional commits easy to use in Node.js projects and was inspiration and background for my work on this Go adaptation of his work.
