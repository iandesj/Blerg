Part of #1

## What to build

Establish the Blerg CLI application skeleton: pick an implementation language for the greenfield repo, wire a command entrypoint, load a sources manifest, and persist sources and posts in embedded SQLite. Provide injectable HTTP for tests so integration tests can drive the same surface a user will invoke later, using canned feed responses and a temporary database without live network calls.

## Acceptance criteria

- [ ] Runnable CLI binary or documented entry command exists with a top-level help subcommand list.
- [ ] Sources manifest format is defined and loads at least one source record (name, feed URL, enabled flag).
- [ ] SQLite schema supports sources and posts with unique post URL constraint as described in the parent spec.
- [ ] Integration-test harness can run the CLI (or in-process command layer) with fake HTTP and isolated database.
- [ ] Build, test, and lint commands are documented in agent instructions for the chosen stack.

## Blocked by

None (can start immediately).
