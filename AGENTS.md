# Agent instructions

This file orients automated coding assistants working in this repository.

## Project

- **Purpose:** Blerg is a CLI that ingests engineering-blog RSS/Atom feeds and returns a single best-matching post for a keyword query.
- **Stage:** MVP in progress (Go CLI, SQLite index).

## Conventions

- Prefer small, focused changes; match existing style once code lands.
- Update `CHANGELOG.md` under `[Unreleased]` for user-visible behavior changes.
- Put durable documentation in `docs/`; keep `README.md` as the front door.
- Do not commit secrets, credentials, or local-only config.

## Development rules

- Make small, concise, logical commits — one coherent step per commit, not drive-by mixes.
- Use TDD for new features, functionality, and behavioral changes: red → green → refactor (failing test, minimal pass, then improve design without changing behavior).
- Write code that is easy to test, read, and change (clear boundaries, narrow units, explicit dependencies).
- Do not add comments that justify why code exists; remove or refactor instead — comments only for non-obvious constraints or external contracts.

## Commands

Run inside `devbox shell` when Devbox is configured:

```bash
pre-commit run --all-files   # gitleaks, formatting, and sanity checks
```

Requires Go 1.25+ (see `go` line in `go.mod`; use `devbox shell` for the pinned toolchain).

```bash
go build -o bin/blerg ./cmd/blerg
go test ./...
go vet ./...
golangci-lint run ./...
```

## Boundaries

- Ask before large refactors, dependency upgrades, or license changes.
- Do not disable pre-commit hooks or skip CI without explicit approval.

## Agent skills

### Issue tracker

Work is tracked in GitHub Issues on `iandesj/Blerg` via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical roles (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`) match GitHub label names. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout: root `GLOSSARY.md` (when present) and `docs/adr/`. See `docs/agents/domain.md`.

## References

- `docs/README.md` — documentation index
- `docs/architecture.md` — system overview (stub until filled in)
- `docs/agents/` — issue tracker and domain-doc rules for engineering skills
- `.agents/skills/boilerplate/SKILL.md` — portable Agent Skills workflow ([agentskills.io](https://agentskills.io/)); auto-discovered in many harnesses via `.agents/skills/`
