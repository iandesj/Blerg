# Repository layout

Paths are relative to the template repository root (not inside `.agents/skills/boilerplate/`).

## Root files

### `README.md`

Front door for humans. Should include: one-line purpose, setup steps (Devbox optional, pre-commit install), layout table, link to `docs/`, license pointer. Avoid duplicating architecture detail here.

### `AGENTS.md`

In-repo instructions for automated assistants. Sections: project purpose, conventions, **development rules** (commits, TDD red-green-refactor, testability, comment policy), commands, boundaries, references to `docs/`. Update the commands section when build/test/lint exist.

Harnesses that only read `CLAUDE.md` or similar should reuse this content—do not fork divergent rules.

### `CHANGELOG.md`

[Keep a Changelog](https://keepachangelog.com/) format with `[Unreleased]` and version sections. Record template or product changes that affect users or contributors.

### `devbox.json`

Pins reproducible CLI tools:

- `git`, `pre-commit`, `gitleaks` (versions as pinned in template)
- `shell.init_hook` reminds developers to run `pre-commit install`
- `shell.scripts.hooks` → `pre-commit run --all-files`

Add stack packages (Node, Go, Rust, etc.) when the stack is chosen. Commit `devbox.lock` when present.

### `.pre-commit-config.yaml`

Ordered concerns:

1. Secrets (`gitleaks`)
2. Universal file hygiene (`pre-commit-hooks`)
3. CI lint (`actionlint`) — optional
4. Complexity (`local` + `lizard`)

### `.gitignore`

Language-agnostic baseline: OS cruft, editors, `.env`, common dependency and build dirs. Remove or narrow entries that conflict with the chosen stack (for example whether `Cargo.lock` is committed).

## `docs/`

### `docs/README.md`

Index: suggested `architecture.md`, `runbooks/`, `adr/`. Writing guidelines.

### `docs/architecture.md`

Stub until the system exists. Context, diagram placeholder, component table, cross-cutting concerns.

## `.agents/skills/boilerplate/`

Agent Skills package (this skill). Safe to copy to `~/.agents/skills/boilerplate/` or install via Skills CLI. Discovery: clients that scan `.agents/skills/` (and often `.cursor/skills/`, `.claude/skills/`, `.github/skills/` for compatibility).

## Extension points (after stack selection)

| Concern | Where to extend |
|---------|-----------------|
| Runtime / SDK | `devbox.json` `packages` |
| Dependency dirs | `.gitignore` |
| Formatters / linters | `.pre-commit-config.yaml` new repos or local hooks |
| CI | `.github/workflows/` (keeps `actionlint` relevant) |
| Agent build/test | `AGENTS.md` Commands + `README.md` |
| API / design docs | `docs/` |
