Part of #1

## What to build

Ship the curated engineering-blog allowlist with validated public RSS or Atom URLs, document how to install and run Blerg, record the RSS-first allowlist decision in an ADR, and update architecture overview and changelog for user-visible CLI behavior.

## Acceptance criteria

- [ ] Sources manifest includes on the order of 15–25 engineering-focused feeds; broken feeds are disabled with a recorded reason rather than failing the product.
- [ ] README documents install, refresh, find, and sources list for end users.
- [ ] ADR captures RSS-first ingestion and allowlist scope (no listing-page scrape in v1).
- [ ] Architecture doc reflects real components; changelog notes user-visible CLI under Unreleased.
- [ ] Documented manual or automated smoke path to refresh against a small subset of live feeds (optional for CI).

## Blocked by

- #6 (Multi-source refresh, partial failure, and sources list)
- #7 (Operational polish: caching, staleness, JSON, polite HTTP)
