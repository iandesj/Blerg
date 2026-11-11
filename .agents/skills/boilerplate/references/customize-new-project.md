# Customize a new project from the template

## 1. Identity

- [ ] `README.md`: project name, description, audience
- [ ] `CHANGELOG.md`: replace `ORG/REPO` link placeholders; set first release date when tagging
- [ ] License section in `README.md` and add `LICENSE` file if open source
- [ ] `AGENTS.md`: one-paragraph **Purpose** and realistic **Stage**

## 2. Documentation

- [ ] `docs/architecture.md`: replace stub with real context, diagram, components
- [ ] `docs/README.md`: link any new guides (runbooks, ADR index)

## 3. Tooling smoke test

- [ ] `git init` (if not already)
- [ ] `devbox shell` (optional)
- [ ] `pre-commit install`
- [ ] `pre-commit run --all-files` — fix any failures before first push

## 4. Stack selection (when ready)

- [ ] Add language SDK/runtime to `devbox.json`
- [ ] Trim `.gitignore` to match ecosystem conventions
- [ ] Add formatter/linter hooks to `.pre-commit-config.yaml`
- [ ] Document build, test, lint in `README.md` and `AGENTS.md`
- [ ] Remove `actionlint` repo block if not using GitHub Actions

## 5. Agent and skill wiring

- [ ] Ensure contributors’ harnesses read `AGENTS.md`
- [ ] Optionally install skill: `npx skills add <owner>/boilerplate@boilerplate -g -y` (once published); project copy is already at `.agents/skills/boilerplate/`

## 6. First release

- [ ] Move `[Unreleased]` entries to a versioned section in `CHANGELOG.md`
- [ ] Tag `v0.1.0` (or appropriate semver) when application code exists

## Anti-patterns

- Committing `.env` or secrets “temporarily”
- Duplicating agent rules in many markdown files with conflicting text
- Adding framework boilerplate before the stack decision is explicit
- Disabling Gitleaks or Lizard to unblock a commit without team approval
