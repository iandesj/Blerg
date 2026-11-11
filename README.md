# Project Name

Short description of what this project does and who it is for.

## Getting started

1. Install [Devbox](https://www.jetify.com/devbox) (optional but recommended for a reproducible shell).
2. Run `devbox shell` to enter the development environment.
3. Install pre-commit hooks: `pre-commit install`
4. Follow language-specific setup in [docs/](docs/) once you add application code.

## Repository layout

| Path | Purpose |
|------|---------|
| `docs/` | Design notes, runbooks, and guides |
| `AGENTS.md` | Instructions for AI coding assistants |
| `CHANGELOG.md` | Notable changes by version |
| `devbox.json` | Pin dev tools via Devbox |
| `.pre-commit-config.yaml` | Local git hooks (Gitleaks, Lizard, sanity checks) |

## Development

- Document architecture and workflows under `docs/`.
- Record user-visible changes in `CHANGELOG.md` (see [Keep a Changelog](https://keepachangelog.com/)).
- Run `pre-commit run --all-files` before pushing when hooks are not installed.

## License

Specify your license here (e.g. MIT, Apache-2.0, proprietary).
