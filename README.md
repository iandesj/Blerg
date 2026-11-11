# Blerg

CLI that indexes engineering-blog feeds and finds **one** relevant post for your keywords.

## Getting started

1. Install [Devbox](https://www.jetify.com/devbox) (optional but recommended for a reproducible shell).
2. Run `devbox shell` to enter the development environment.
3. Install pre-commit hooks: `pre-commit install`
4. Build and run the CLI:

```bash
go build -o bin/blerg ./cmd/blerg
./bin/blerg --help
```

Copy `config/sources.example.yaml` when you need a manifest for future `refresh` / `find` work.

## Repository layout

| Path | Purpose |
|------|---------|
| `docs/` | Design notes, runbooks, and guides |
| `AGENTS.md` | Instructions for AI coding assistants |
| `CHANGELOG.md` | Notable changes by version |
| `devbox.json` | Pin dev tools via Devbox |
| `.pre-commit-config.yaml` | Local git hooks (Gitleaks, Lizard, sanity checks) |
| `.agents/skills/boilerplate/` | Portable [Agent Skill](https://agentskills.io/) (cross-client discovery) |

## Agent skill

The `boilerplate` skill ([`.agents/skills/boilerplate/SKILL.md`](.agents/skills/boilerplate/SKILL.md)) documents how to scaffold, customize, and extend this template. It lives under **`.agents/skills/`**, the interop path Cursor, Copilot, Codex, and [Skills CLI](https://skills.sh/) scan (see [`.agents/README.md`](.agents/README.md)):

```bash
# After publishing this repo to GitHub:
npx skills add <owner>/<repo>@boilerplate -g -y
```

See [`.agents/skills/README.md`](.agents/skills/README.md) for paths and global install.

## Development

- Document architecture and workflows under `docs/`.
- Record user-visible changes in `CHANGELOG.md` (see [Keep a Changelog](https://keepachangelog.com/)).
- Run `pre-commit run --all-files` before pushing when hooks are not installed.

## License

Specify your license here (e.g. MIT, Apache-2.0, proprietary).
