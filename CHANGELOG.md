# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Shipped `config/sources.yaml` with validated engineering-blog feeds including AWS Architecture, Discord, and Pinterest (Uber and LinkedIn disabled after URL check).

### Changed

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

## [0.1.0] - YYYY-MM-DD

### Added

- First tagged release after you add application code.

[Unreleased]: https://github.com/ORG/REPO/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ORG/REPO/releases/tag/v0.1.0
