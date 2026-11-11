Part of #1

## What to build

Harden feed fetching and CLI ergonomics: conditional HTTP caching with stored validators, configurable index staleness so find can refresh automatically when data is old, JSON output for scripting, identifying User-Agent, bounded concurrency across feeds, backoff on rate-limit and server errors, and exit codes that distinguish success, no match, and operational failure.

## Acceptance criteria

- [ ] ETag and Last-Modified from feed responses are stored and honored on subsequent refresh (tests assert conditional request headers against fixtures).
- [ ] When index age exceeds configured staleness, find triggers refresh before scoring.
- [ ] Find supports JSON output mode with stable field names for title, URL, source, date, and match explanation.
- [ ] HTTP client sends a project-identifying User-Agent; concurrent fetches are bounded; 429 and 503 responses trigger backoff.
- [ ] Exit codes distinguish success, no match, and operational error as specified in the parent spec.
- [ ] Integration tests cover JSON output and at least one staleness or caching scenario.

## Blocked by

- #4 (Find returns exactly one post for a keyword query)
- #6 (Multi-source refresh, partial failure, and sources list)
