# Agent skills (`.agents/skills/`)

Project-level skills for the [Agent Skills](https://agentskills.io/) open format. This path is the **cross-client convention**: Cursor, GitHub Copilot, OpenAI Codex, Gemini CLI, Amp, and others discover `SKILL.md` files here without vendor-specific folders.

| Scope | Typical path |
|-------|----------------|
| Project | `.agents/skills/<skill-name>/SKILL.md` |
| User | `~/.agents/skills/<skill-name>/SKILL.md` |

Vendor aliases (optional): `.cursor/skills/`, `.claude/skills/`, `.github/skills/` — often scanned for compatibility; prefer `.agents/skills/` for new skills.

## Available skills

| Skill | Path | Use |
|-------|------|-----|
| `boilerplate` | [boilerplate/SKILL.md](boilerplate/SKILL.md) | Scaffold and maintain the language-agnostic repo template |

## Installation

**In this repo:** open the project — compliant agents auto-discover `.agents/skills/boilerplate/SKILL.md`.

**Global install:** copy the skill folder to `~/.agents/skills/boilerplate/` or use the Skills CLI (below).

**Skills CLI (when published to GitHub):**

```bash
npx skills add <owner>/<repo>@boilerplate -g -y
```

Replace `<owner>/<repo>` with the published template repository coordinates.

## Authoring

- Skill folder name must match frontmatter `name` (lowercase, hyphens only).
- Put long detail in `.agents/skills/<name>/references/`, not only in `SKILL.md` (keep under ~500 lines).
- In-repo agent rules stay in root `AGENTS.md`; cross-repo workflow stays in the skill.
