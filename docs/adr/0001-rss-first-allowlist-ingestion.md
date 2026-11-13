# ADR 0001: RSS-first ingestion with a curated allowlist

- **Status:** Accepted
- **Date:** 2026-03-14

## Context

Blerg returns **one** engineering blog post for a keyword query. The index must stay trustworthy (engineering-focused sources), cheap to refresh, and polite to publishers. Full-site HTML scraping and open-web search are out of scope for v1.

## Decision

1. **Ingestion:** RSS and Atom feeds only. No blog home page or listing-page HTML scrape in v1.
2. **Source scope:** A checked-in YAML **allowlist** (`config/sources.yaml`) with display name, feed URL, and `enabled` flag.
3. **Broken feeds:** Disable in the manifest (with a short comment noting the validation failure) instead of removing the row or failing the whole product.
4. **Post identity:** Canonical URL is the unique key; duplicates from multiple feeds collapse to one row.

## Consequences

- Adding a blog is editing the manifest and running `blerg refresh`, not shipping code.
- Sources without a stable public feed stay out of scope until a feed URL exists.
- Ranking uses feed metadata (title, summary, categories) until a later phase may fetch article HTML.

## Alternatives considered

- **HTML listing scrape:** Higher breakage and robots/paywall risk; deferred.
- **Search API (Google, Algolia, etc.):** Keys, cost, and non-engineering noise; deferred.
- **Dynamic discovery:** Out of v1; allowlist keeps scope explicit.
