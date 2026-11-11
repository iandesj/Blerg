Part of #1

## What to build

Improve find so similarly scored posts prefer newer publication dates, weak matches produce a clear no strong match response instead of a poor guess, and the user sees a short rationale for why the chosen post matched (for example title versus summary match).

## Acceptance criteria

- [ ] Recency boost breaks ties or near-ties toward newer posts (half-life on the order of months per parent spec).
- [ ] When the best score is below the configured threshold, find reports no strong match and does not return a low-confidence post.
- [ ] Successful find includes a brief match explanation in human-readable output.
- [ ] Distinct exit code for no-match versus successful match.
- [ ] Tests cover threshold behavior and recency ordering with controlled fixture dates.

## Blocked by

- #4 (Find returns exactly one post for a keyword query)
