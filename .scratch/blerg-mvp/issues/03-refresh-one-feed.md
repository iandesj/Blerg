Part of #1

## What to build

From a configured allowlist source, run refresh to download an RSS or Atom feed, parse entries, and upsert posts into the local index with title, canonical URL, published time, summary, and categories. Re-running refresh must not duplicate posts that share the same URL. The user sees whether refresh succeeded or failed, with a non-zero exit when refresh cannot complete successfully.

## Acceptance criteria

- [ ] Explicit refresh command fetches the feed URL from the manifest for enabled sources (tests use mocked HTTP with fixture XML).
- [ ] Parsed feed entries are stored as posts linked to their source.
- [ ] Upsert by URL keeps a single row per post across repeated refresh runs.
- [ ] CLI reports success or failure; total failure yields a non-zero exit code.
- [ ] Integration test proves refresh then inspectable index state without live network.

## Blocked by

- #2 (CLI scaffold, stack choice, and integration-test seam)
