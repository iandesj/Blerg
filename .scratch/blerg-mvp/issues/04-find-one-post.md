Part of #1

## What to build

After the index is populated, run find with a keyword query and receive exactly one engineering blog post: title, URL, source name, and publish date. Ranking uses keyword overlap on title, summary, and categories. Default output is human-readable for terminal use.

## Acceptance criteria

- [ ] Find command accepts a free-text query string.
- [ ] Returns exactly one post when matches exist in the index.
- [ ] Output includes title, URL, source display name, and publication date.
- [ ] Integration test: refresh fixtures with known titles, then find returns the expected URL for a targeted query.
- [ ] Find works on an index populated only via refresh (no manual DB seeding in the happy path).

## Blocked by

- #3 (Refresh ingests one feed into the local index)
