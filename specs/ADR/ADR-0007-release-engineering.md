# ADR-0007 — Release engineering: goreleaser v2 kept, `git-cz` alias build, snapcraft removal

Status: Accepted
Date: 2026-10-11
Deciders: @isokolovskii, Pi (GLM)

## Context

The pipeline already works: `make version` (git-cliff bump from `main`) → tag → `release.yml` (goreleaser v2 → GitHub release + build-provenance attestations). Observed issues:

- `release.yml` installs Snapcraft and passes `SNAPCRAFT_STORE_CREDENTIALS`, but `.goreleaser.yaml` defines no `snaps` — dead weight and a confusing secret dependency.
- goreleaser version is pinned `v2.10.2` in release but floating `~> v2` in the test job's snapshot build.
- cz-cli users expect `git cz` to work — git resolves `git <verb>` to a `git-<verb>` binary on PATH; we ship only `commitizen`.
- `before.hooks` runs `go generate ./...` though the repo has no directives (harmless, but noise).

## Decision

1. Keep goreleaser v2 + the tag-driven flow and attestation step unchanged in mechanics.
2. Add a second goreleaser **build id `git-cz`** producing a `git-cz` binary from the same `main.go` (release archives list both binaries; `go install` users get a README symlink one-liner). This is what makes the "use it like `git cz`" promise true (PRD FR-9).
3. Remove the Snapcraft install step and `SNAPCRAFT_STORE_CREDENTIALS` from `release.yml`; remove `--skip=snapcraft` from the test job's snapshot invocation (flag removed in goreleaser v2 — `--skip` no longer takes `snapcraft`, which currently only survives because goreleaser tolerates it).
4. Pin one goreleaser version for both workflows via the official action's version input (same `~> v2` range in both, avoiding drift).
5. Drop `go generate ./...` from `before.hooks` (no directives; re-add when first used).

## Rationale

- The alias binary is the cheapest possible parity win with cz-cli's headline UX (`git cz`).
- Snapcraft remnants actively mislead contributors (implying a snap deliverable that does not exist).
- Version drift between CI jobs can pass snapshot while failing release (or vice versa) — one pin removes the class of bug.

## Consequences

- Archive layout gains `git-cz` next to `commitizen` (checksums/attestations cover it automatically via checksum file).
- Windows `git-cz` binaries work for git-for-windows the same way (`git-cz.exe` on PATH).
- If a snap deliverable is ever wanted, it returns as its own ADR with registry credentials handled properly.
