# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- GitHub Actions CI (test, vet, lint, build) and a tag-triggered release that publishes `blerg` archives. `scripts/install.sh` installs the latest release. `blerg --version` prints the release version.
- MIT license (`LICENSE`).
- `blerg find --interactive` (`-i`) browses strong matches with `j`/`k` and opens the selected post in the browser on Enter. The browser is colored unless `NO_COLOR` is set. `--random` with `--interactive` shuffles that list. Enter records `opened_at` on that post; refresh does not clear it.
- `blerg find --random` returns a random indexed post that meets the match threshold.
- Local fixture feeds: non-HTTP `feed_url` values in the manifest (paths relative to the manifest file) for offline `refresh`.
- Curated `config/sources.yaml` allowlist (~28 enabled engineering feeds; Uber and LinkedIn disabled with documented reason).
- The Morning Paper (`https://blog.acolyer.org/feed/`) on the allowlist. New posts stopped in February 2021; the feed still returns that archive window.
- `config/sources.smoke.yaml` for manual live-network smoke (three feeds).
- ADR 0001 (RSS-first ingestion, allowlist scope) and updated [architecture](docs/architecture.md).
- Conditional feed fetches (`ETag` / `Last-Modified`), bounded concurrent refresh, 429/503 backoff, `--stale-after` auto-refresh on find, and `find --json` output.
- `blerg sources list` shows per-source last successful ingest time and last error (when present).
- `blerg find` ranks indexed posts by keyword overlap with recency tie-break, match threshold, and exit code `2` for no strong match.
- `blerg refresh` ingests enabled RSS/Atom feeds into SQLite with partial-failure semantics.

### Changed

- `--sources` defaults to `sources.yaml` in the OS user config directory (`~/.config/Blerg` on Linux). The first `refresh`, `find`, or `sources` command installs the shipped `config/sources.yaml` there when the file is missing. An existing file is not overwritten.
- Lizard pre-commit limits are complexity ≤15 and length ≤60 for production code, and complexity ≤20 and length ≤120 for tests. Parameter limit stays at 5. See `AGENTS.md`.
- `blerg refresh` exits 0 when at least one enabled source succeeds; non-zero only when every enabled source fails (partial failures on stderr).
- Devbox `go@1.25` aligned with `go.mod` (required by `gofeed` v1.5).

### Added

- `blerg find` ranks with recency tie-break (6-month half-life), enforces a minimum match score, prints a short match explanation, and exits `2` on no strong match (`1` on operational errors).
- `blerg find` ranks indexed posts by keyword overlap (title, summary, categories) and prints one human-readable result.
- Go `blerg` CLI skeleton (`refresh`, `find`, `sources`) with SQLite schema for sources and posts.
- `blerg refresh` downloads enabled feeds, parses RSS/Atom, and upserts posts by canonical URL.
- `blerg sources list` lists manifest-backed sources.
- YAML sources manifest loader (`config/sources.example.yaml`, `testdata/sources.yaml`).
- Integration tests at the CLI seam (Cobra + subprocess, injectable HTTP, canned RSS).
- `golangci-lint` configuration and documented lint command.

- Matt Pocock engineering skills bundle under `.agents/skills/` (`to-spec`, `to-tickets`, `grill-with-docs`, `implement`, `code-review`, `retro`, `wayfinder`, `ask-matt`) with `skills-lock.json`.
- `triage` skill (mattpocock/skills) and `docs/agents/triage-labels.md` with GitHub triage labels.
- `docs/agents/` (GitHub issue tracker, single-context domain rules) and `## Agent skills` in `AGENTS.md` for engineering skills setup.
- `.agents/skills/boilerplate` Agent Skill (agentskills.io) under cross-client `.agents/skills/` layout.
- Lizard cyclomatic-complexity checks in pre-commit (CCN ≤10, length ≤80, args ≤5).
- Initial repository template (docs, Devbox, pre-commit, agent instructions).
- Gitleaks secret scanning via pre-commit.
- Development rules in `AGENTS.md` (commits, TDD, code quality, comments).

## [0.2.0] - 2026-11-13

### Added

- `blerg update` replaces the running binary with the latest GitHub release.

## [0.1.0] - YYYY-MM-DD

### Added

- First tagged release after you add application code.

[Unreleased]: https://github.com/iandesj/Blerg/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/iandesj/Blerg/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/iandesj/Blerg/releases/tag/v0.1.0
