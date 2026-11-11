# Documentation

Place long-form documentation here. The root `README.md` should stay short and link into this tree.

## Suggested layout

| Document | Contents |
|----------|----------|
| [architecture.md](architecture.md) | Components, boundaries, data flow |
| [agents/](agents/) | Issue tracker + domain-doc rules for Matt Pocock engineering skills |
| [adr/](adr/) | Architecture decision records |
| `runbooks/` | Operational procedures (deploy, rollback, incident) |

## Writing guidelines

- Use present tense and concrete names (services, env vars, queues).
- When a doc goes stale, fix it in the same change that alters behavior.
- Prefer diagrams in `architecture.md` over scattering context across issues.

## Placeholders

Until you add an application stack, treat these files as stubs and replace them as the project grows.
