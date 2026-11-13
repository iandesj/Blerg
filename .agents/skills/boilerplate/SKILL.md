---
name: boilerplate
description: >-
  Scaffold and maintain a language-agnostic git repository template with docs,
  Devbox, pre-commit (Gitleaks, Lizard, sanity hooks), AGENTS.md, and
  Keep a Changelog. Use when starting a greenfield repo, copying this template,
  onboarding agents to repo conventions, or extending hooks and agent
  instructions without picking a language stack yet.
license: UNLICENSED
compatibility: Requires git. Optional Devbox and pre-commit for the documented dev workflow.
metadata:
  author: boilerplate-template
  version: "0.2.0"
  spec: https://agentskills.io/specification
---

# Boilerplate repository template

Portable workflow for a **language-agnostic** project skeleton. This skill describes the layout at the repository root (parent of `.agents/skills/boilerplate/`). Harness-specific files (for example `AGENTS.md`) are intentionally minimal so Cursor, Claude Code, Copilot, Amp, or custom runners can all follow the same rules.

**Discovery:** Project skills live under [`.agents/skills/`](https://agentskills.io/client-implementation/adding-skills-support) — the cross-client convention Cursor, Copilot, Codex, Gemini CLI, and others scan. The [Agent Skills spec](https://agentskills.io/specification) defines skill *content*, not location; clients also scan vendor paths (for example `.cursor/skills/`, `.claude/skills/`, `.github/skills/`) for compatibility.

## When to use

- User wants a new repo before choosing a language or framework.
- User asks to copy, fork, or “cookie cut” a standard project layout.
- User wants secret scanning, complexity limits, or git hooks on day zero.
- User wants agent instructions (TDD, commits, comments) in a harness-neutral way.
- User is **in** this template repo and needs to customize placeholders or add stack tooling.

Do **not** use for application logic, deployment runbooks, or language-specific frameworks until a stack is chosen.

## Quick start (new project from this template)

1. Copy or generate a repo containing the template files (see [repo layout](references/repo-layout.md)).
2. Replace placeholders: `README.md` title, `CHANGELOG.md` links, license section, `docs/architecture.md`, `AGENTS.md` purpose.
3. Initialize git, enter Devbox if used: `devbox shell`, then `pre-commit install`.
4. Run `pre-commit run --all-files` once to warm hook environments.
5. After choosing a stack, extend `devbox.json`, `.gitignore`, `.pre-commit-config.yaml`, and document build/test commands in `AGENTS.md` and `README.md`.

## Repository layout (summary)

| Path | Role |
|------|------|
| `README.md` | Human-facing entry; keep short, link to `docs/` |
| `AGENTS.md` | Agent/harness instructions (conventions + dev rules) |
| `CHANGELOG.md` | [Keep a Changelog](https://keepachangelog.com/) |
| `docs/` | Long-form design, runbooks, ADRs |
| `devbox.json` | Pinned dev tools (git, pre-commit, gitleaks) |
| `.pre-commit-config.yaml` | Git hooks: Gitleaks, Lizard, generic checks |
| `.gitignore` | Cross-language ignores; trim or extend per stack |

Full file-by-file notes: [references/repo-layout.md](references/repo-layout.md).

## Agent conventions (mirror `AGENTS.md`)

When editing code in a repo spawned from this template:

- **Commits:** Small, one logical step per commit.
- **TDD:** Red → green → refactor for new behavior.
- **Design:** Easy to test, read, change; narrow units; explicit dependencies.
- **Comments:** Only non-obvious constraints or external contracts—not justification comments.
- **Changelog:** User-visible changes under `[Unreleased]` in `CHANGELOG.md`.
- **Secrets:** Never commit credentials; rely on Gitleaks + `.gitignore` patterns.

## Pre-commit policy

Default hooks (see [pre-commit reference](references/pre-commit-hooks.md)):

- **Gitleaks** — secret scan on commit (staged content).
- **Lizard** — warnings-only complexity. This repo's limits are in `AGENTS.md`: production CCN ≤ 15, length ≤ 60, params ≤ 5; tests (`*_test.go`) CCN ≤ 20, length ≤ 120, params ≤ 5.
- **pre-commit-hooks** — whitespace, EOF, YAML/JSON, merge conflicts, large files, private keys, LF endings.
- **actionlint** — GitHub Actions workflows; remove that repo block if unused.

Commands:

```bash
pre-commit install
pre-commit run --all-files
devbox run hooks   # when Devbox is configured
```

Tune Lizard via `args` and `files` in `.pre-commit-config.yaml`. Pin hook versions on intentional upgrades, not drive-by bumps.

## Customization checklist

Use [references/customize-new-project.md](references/customize-new-project.md) when turning the template into a named product repo.

Typical order:

1. Identity (name, description, license).
2. Docs (`docs/architecture.md`, index in `docs/README.md`).
3. Agent commands section once build/test exist.
4. Stack-specific `devbox.json` packages and pre-commit linters.
5. CI workflow (then keep `actionlint`; otherwise delete it).

## Harness integration (agnostic)

This skill does not depend on a single IDE:

| Harness | Typical wiring |
|---------|----------------|
| **Cursor / generic agents** | Read repo-root `AGENTS.md` |
| **Claude Code** | Symlink or copy `AGENTS.md` → `CLAUDE.md`, or use `.claude/skills/` (compat) / `.agents/skills/boilerplate` |
| **GitHub Copilot / VS Code** | `.agents/skills/` or `.github/skills/` |
| **Skills CLI** | Install skill from this repo path or published package (see repo `README.md`) |
| **Custom runners** | Load `SKILL.md` body + `references/` on activation |

Keep one source of truth: **`AGENTS.md` for in-repo rules**, **`SKILL.md` for cross-repo workflow** when the skill is installed globally.

## Boundaries

- Ask before large refactors, dependency upgrades, or license changes.
- Do not disable pre-commit or secret scanning without explicit approval.
- Do not add language frameworks until the user picks a stack (avoid fake placeholders).

## References bundled with this skill

- [references/repo-layout.md](references/repo-layout.md) — artifacts and extension points
- [references/customize-new-project.md](references/customize-new-project.md) — placeholder replacement workflow
- [references/pre-commit-hooks.md](references/pre-commit-hooks.md) — hooks, versions, tuning
