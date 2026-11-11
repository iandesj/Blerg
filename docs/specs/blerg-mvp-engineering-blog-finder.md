# Blerg: single engineering blog post by topic (MVP)

> Canonical tracker copy: [iandesj/Blerg#1](https://github.com/iandesj/Blerg/issues/1) (`ready-for-agent`).

## Problem Statement

When exploring an engineering topic, finding **one high-quality, relevant blog post** from reputable engineering sources takes too long: scattered sites, search noise, and marketing content dilute results. The user wants to type a topic or keywords and get **exactly one** curated match from **engineering-focused** public blogs—not a generic web search or an unbounded list of links.

## Solution

A **local CLI application** (working name: Blerg) that maintains an index of posts from a **curated allowlist** of popular engineering blogs, ingested primarily via **RSS/Atom feeds**. The user runs a **find** command with keywords; the app returns **one** best-matching post (title, URL, source, publication date, and a short match explanation). A **refresh** command updates the index from feeds. Scope is intentionally narrow: engineering content only, one result, no accounts.

## User Stories

1. As an engineer researching a topic, I want to enter keywords and receive a single recommended blog post, so that I can start reading without wading through search results.
2. As an engineer, I want the result to come from known engineering blogs, so that I trust the source quality.
3. As an engineer, I want to see the post title, URL, source name, and publish date, so that I can judge freshness and relevance quickly.
4. As an engineer, I want a brief explanation of why this post matched my query, so that I can decide whether to open it.
5. As an engineer, I want the app to prefer posts whose titles and summaries match my keywords, so that matches feel on-topic.
6. As an engineer, I want newer relevant posts to rank above older ones when scores are similar, so that guidance stays current.
7. As an engineer, I want a clear message when no post matches strongly enough, so that I know to broaden or change my keywords.
8. As an engineer, I want to refresh the local index from all configured sources, so that recently published posts appear in search.
9. As an engineer, I want feed updates to reuse HTTP caching where possible, so that refresh is polite and fast.
10. As an engineer, I want to list configured sources and see when each was last fetched successfully, so that I can tell if the index is stale.
11. As an engineer, I want the application to identify itself with a proper User-Agent, so that feed operators can contact the project if needed.
12. As an engineer, I want rate limiting and backoff on failed fetches, so that the app does not hammer source servers.
13. As an engineer, I want duplicate posts (same URL from re-ingestion) to appear once in the index, so that search results stay stable.
14. As an engineer, I want to run find without manually refreshing when the index is fresh enough, so that day-to-day use is frictionless.
15. As an engineer, I want find to trigger a refresh when the index exceeds a configurable staleness threshold, so that I still get recent posts without a separate habit.
16. As a maintainer, I want sources defined in a human-editable curated list, so that adding or disabling a blog does not require code changes.
17. As a maintainer, I want each source to have a display name and feed URL, so that configuration is auditable.
18. As a maintainer, I want optional per-source metadata (category, weight), so that ranking can prefer certain sources later without breaking v1.
19. As a developer, I want a persisted local index, so that find works offline after refresh.
20. As a developer, I want ingestion and search logic separated from CLI parsing, so that behavior is testable without subprocesses.
21. As a developer, I want errors from a single broken feed to not prevent other feeds from ingesting, so that partial failure is acceptable.
22. As a developer, I want invalid feed XML for one source logged and skipped, so that one bad source does not crash the whole refresh.
23. As an engineer, I want JSON output mode for scripting, so that I can pipe results into other tools.
24. As an engineer, I want human-readable terminal output by default, so that casual use needs no flags.
25. As an engineer, I want exit codes that distinguish success, no match, and operational failure, so that scripts can branch correctly.
26. As a maintainer, I want the initial allowlist to target company engineering blogs and similar public eng writing, so that scope stays "engineering blogs."
27. As an engineer, I want not to authenticate or manage API keys for the MVP, so that setup is minimal.
28. As a maintainer, I want user-visible behavior documented in the changelog when shipped, so that adopters know what changed.
29. As a developer, I want tests to follow red-green-refactor for new behavior, so that regressions are caught early.
30. As an engineer, I want paywalled or login-gated posts excluded from v1 behavior, so that find always yields a freely openable URL when possible.

## Implementation Decisions

### Product and boundaries

- **MVP interface:** CLI only (no web UI in this spec).
- **Result cardinality:** Exactly **one** post per find invocation (not a ranked list UI in v1).
- **Content scope:** Engineering/technical blogs via **allowlist**; marketing-heavy posts are out of scope for filtering beyond choosing reputable eng feeds.
- **Ingestion strategy:** **RSS/Atom first**; do not implement HTML scraping of listing pages in v1.
- **Ranking strategy (v1):** Keyword-based scoring on title, categories/tags when present, and feed summary/description; **recency boost** with a half-life on the order of months.
- **Match threshold:** If the top score is below a defined threshold, return **no strong match** (not a weak random post).

### Architecture

- **Layers:** (1) command entrypoint, (2) orchestration for refresh and find, (3) feed ingestion and parsing, (4) persistence, (5) ranking/search.
- **Persistence:** Embedded relational store (SQLite) holding sources, posts, and ingestion metadata (last fetch time, HTTP cache validators where available).
- **Post record (conceptual):** stable id, source id, title, canonical URL, published timestamp, summary text, optional raw categories; unique constraint on URL.
- **Source record (conceptual):** name, feed URL, enabled flag, optional weight/category, last successful ingest timestamp, last error message.

### Commands (behavioral contract)

- **refresh:** Fetch all enabled sources, upsert posts, update per-source status; may be invoked explicitly or implicitly when find detects stale index.
- **find \<query\>:** Tokenize query, score indexed posts, apply recency boost, return top-one or no-match; support human and JSON output.
- **sources list:** Show configured sources and last successful refresh (and optionally last error).

### Configuration

- Curated **sources manifest** (YAML or equivalent) checked into the repo; disabled sources skipped.
- **Staleness:** Configurable max age for index; find may trigger refresh when exceeded.
- **HTTP client:** Identifying User-Agent string including project name; sequential or small bounded concurrency; honor 429/503 with backoff.
- **Caching:** Store ETag / Last-Modified per feed when present to reduce bandwidth.

### Initial source set

- Approximately **15–25** feeds from well-known engineering blogs (company eng blogs, infra/platform writing); exact URLs validated at implementation time. Optional **dev.to tag feeds** as supplemental sources clearly labeled in output.
- Feeds that do not resolve or return invalid XML are disabled or marked failed in source status until fixed.

### Stack

- Greenfield repo: choose **one** implementation language during build (Python or Go) consistent with future Devbox tooling; single deployable CLI binary or equivalent entry script documented in agent instructions after choice.

### Documentation

- Update architecture overview and changelog for user-visible CLI behavior when implemented.
- ADR recommended at implementation time: **RSS-first ingestion and allowlist scope** (no listing-page scrape v1).

### Deferred within codebase (not v1)

- Full-article HTML fetch and body-text reranking.
- Semantic embeddings or LLM reranking.
- Hacker News Algolia or other external discovery APIs.
- Web UI, user accounts, saved searches.

## Testing Decisions

### Proposed test seam (single primary seam)

**Primary seam:** the **application command boundary**—the same surface the user invokes (`refresh`, `find`, `sources list`)—run in **integration tests** with:

- Replaceable **HTTP client** that serves canned RSS/Atom fixtures (valid, malformed, empty, duplicate entries).
- **Temporary on-disk or in-memory database** for persistence.

Rationale: greenfield with no existing seams; highest seam that still proves end-to-end behavior without real network. Inner units (parser, scorer) get unit tests only where integration tests would be opaque; keep cross-cutting count low.

**Good tests:** assert **observable outcomes**—stdout/stderr shape, JSON fields, exit codes, DB row counts after refresh, find returns expected URL for a fixture corpus—not private methods or parser internals.

**Modules exercised through the seam:** refresh + find flows together prove ingestion, deduplication, scoring, threshold, and staleness trigger.

**Seam check:** If you expected a library-only seam without CLI, or contract tests against a published HTTP API, comment on the issue before implementation.

## Out of Scope

- Scraping blog homepages or sitemaps without RSS when a feed exists elsewhere (v1).
- Returning multiple ranked results as the default UX (v1).
- Non-engineering news, press releases, or general tech journalism unless explicitly added to allowlist with different product scope.
- User authentication, multi-user state, cloud sync.
- Hosted SaaS API, browser extension, mobile app.
- Full-text extraction from article HTML and body-based reranking.
- Vector search, LLM-based reranking, query expansion.
- Google Custom Search or paid search APIs.
- Bypassing paywalls or automating login.
- Real-time push notifications for new posts.
- Moderation workflows for user-submitted sources (manual manifest edits only in v1).
- Internationalization of CLI messages (English only for v1).
- Compliance certification; basic robots.txt respect when/if article HTML fetch is added later.

## Further Notes

- Conversation planned **phased delivery:** v1 delivers allowlist + SQLite + keyword rank + refresh/find/sources; later phases may add optional article body fetch, HN tie-break, embeddings.
- Open product choices left for implementation defaults unless issue comments override: **CLI-only** confirmed in planning; **refresh on find when stale** preferred over refresh-only; **Go vs Python** left to implementer with Devbox alignment.
- Repository is **greenfield** (boilerplate layout, stub architecture doc, no glossary yet). Product name **Blerg** aligns with repo name.
