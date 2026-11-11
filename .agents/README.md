# `.agents/`

Convention directory for agent-facing assets that should work across harnesses.

| Path | Purpose |
|------|---------|
| `skills/` | [Agent Skills](https://agentskills.io/) packages (`<name>/SKILL.md`) |

The specification defines what belongs inside each skill folder, not where skills must live on disk. `.agents/skills/` is the **interop default** documented by [agentskills.io client guidance](https://agentskills.io/client-implementation/adding-skills-support) and used by Cursor, Copilot, Codex, and the Skills CLI (`npx skills`).

In-repo behavioral rules for assistants remain in root [`AGENTS.md`](../AGENTS.md).
