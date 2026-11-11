Part of #1

## What to build

Support multiple sources in the manifest with enabled and disabled entries. Refresh ingests all enabled feeds; a single bad feed must not abort the whole run. List sources shows each name, whether it is enabled, last successful ingest time, and last error when applicable. URLs deduplicate across sources in the index.

## Acceptance criteria

- [ ] Manifest supports multiple sources; disabled sources are skipped on refresh.
- [ ] One invalid or unreachable feed logs or records error and other sources still ingest.
- [ ] Per-source last success timestamp and last error message are persisted and visible.
- [ ] Sources list command displays status for each configured source.
- [ ] Same post URL from two feeds results in one indexed post.
- [ ] Integration test includes valid and invalid feed fixtures in one refresh.

## Blocked by

- #3 (Refresh ingests one feed into the local index)
