# Pre-commit hooks

Configuration file: repository root `.pre-commit-config.yaml`.

## Gitleaks

- **Repo:** `https://github.com/gitleaks/gitleaks`
- **Pinned rev:** `v8.30.1` (keep in sync with `devbox.json` `gitleaks` package when possible)
- **Behavior:** Scans staged changes for hardcoded secrets; non-zero exit on findings

Manual check (with Devbox): `gitleaks detect --source . --verbose`

## pre-commit-hooks

- **Rev:** `v5.0.0`
- **Hooks:** trailing whitespace, EOF fixer, merge conflict markers, YAML/JSON checks, 512 KiB large-file guard, private key detection, LF line endings

YAML uses `args: [--unsafe]` to allow custom tags if needed later.

## actionlint

- **Rev:** `v1.7.7`
- **When to keep:** `.github/workflows/*.yml` exist
- **When to remove:** No GitHub Actions — delete entire repo block to avoid noise

## Lizard (local hook)

- **Package:** `lizard==1.26.0` via `language: python` + `additional_dependencies`
- **Hook id:** `lizard-complexity`
- **Args:**
  - `--warnings_only`
  - `-C` `10` (cyclomatic complexity)
  - `--length` `80` (lines per function)
  - `-a` `5` (parameter count)
- **Files:** Extension regex for common languages (C/C++, C#, Go, Java, JS/TS, Kotlin, PHP, Python, Ruby, Rust, Scala, Swift, Vue)
- **Exclude:** `vendor/`, `node_modules/`, `.venv/`

Lizard exits non-zero when warnings exceed thresholds. Empty repos skip (no matching files).

### Tuning guidance

| Goal | Action |
|------|--------|
| Stricter complexity | Lower `-C` or `--length` |
| Legacy codebase | Raise thresholds temporarily; narrow `files` to new paths only |
| Single language | Replace extension regex with one pattern (for example `\.py$`) |
| Faster hooks | Exclude generated paths via `exclude` |

## Maintenance

- Run `pre-commit autoupdate` only when intentionally upgrading hook versions; re-run full hook suite after.
- Align `CHANGELOG.md` when hook behavior affects contributors.
- After adding CI, run hooks locally before push — `devbox run hooks` if scripted.
