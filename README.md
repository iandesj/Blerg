# Blerg

[![CI](https://github.com/iandesj/Blerg/actions/workflows/ci.yml/badge.svg)](https://github.com/iandesj/Blerg/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/iandesj/Blerg)](https://github.com/iandesj/Blerg/releases)
[![Go](https://img.shields.io/github/go-mod/go-version/iandesj/Blerg)](go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/iandesj/blerg)](https://goreportcard.com/report/github.com/iandesj/blerg)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

CLI that indexes engineering-blog RSS/Atom feeds and finds **one** relevant post for your keywords.

## Install

Install the latest published release. This downloads the built binary. It does not clone the repo or compile it.

Linux and macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/iandesj/Blerg/main/scripts/install.sh | sh
```

The script picks `blerg_<os>_<arch>.tar.gz` from [the latest GitHub release](https://github.com/iandesj/Blerg/releases/latest) and copies `blerg` to `/usr/local/bin`, or to `~/.local/bin` if that directory is not writable. `blerg --version` prints the release tag without the leading `v`.

Windows: download `blerg_windows_amd64.zip` from that same release page.

A `vX.Y.Z` tag is what publishes a release. Until one exists, the install command has nothing to download.

## Quick start

**Offline (repo fixtures, no network):**

```bash
go build -o bin/blerg ./cmd/blerg
rm -f /tmp/blerg-test.db
./bin/blerg --db /tmp/blerg-test.db --sources testdata/sources.yaml refresh
./bin/blerg --db /tmp/blerg-test.db --sources testdata/sources.yaml find "pooling"
```

**With live feeds:**

```bash
blerg sources list                    # configured feeds + ingest status
blerg refresh                         # fetch all enabled feeds into ~/.config/Blerg/blerg.db
blerg find "connection pooling"       # one human-readable match
blerg find -i "connection pooling"    # j/k to browse, enter opens the browser
blerg find --json "raft consensus"    # scripting output
```

Default feed allowlist: `config/sources.yaml` (~28 engineering blogs). Override with `--sources` or copy `config/sources.example.yaml` for a minimal custom manifest.

## Commands

### `blerg refresh`

Fetches every **enabled** source in the manifest, parses RSS/Atom, and upserts posts by canonical URL. Uses stored `ETag` / `Last-Modified` when available. Exits **0** if at least one enabled source succeeds; **1** if every enabled source fails (partial failures are reported on stderr).

### `blerg find [query]`

Ranks indexed posts (title, summary, categories) with a recency boost and prints the highest-scoring result, with a short `Match:` explanation. `--random` picks uniformly among posts that clear the same score threshold. `--interactive` (`-i`) lists those posts best-first: `j`/`k` move, Enter opens the current URL in the browser and records when it was opened, `q` quits. `--random` with `--interactive` shuffles that list. The browser uses color unless `NO_COLOR` is set. If the index is older than `--stale-after` (default **24h**), refreshes first.

| Flag | Purpose |
|------|---------|
| `--json` | Emit `title`, `url`, `source`, `date`, `match_explanation` as JSON |
| `--random` | Pick one strong match at random, or shuffle the `--interactive` list |
| `--interactive`, `-i` | Browse strong matches and open one in the browser |

| Exit code | Meaning |
|-----------|---------|
| `0` | Match found |
| `1` | Operational error (empty index, refresh failure, etc.) |
| `2` | No strong match for query |

### `blerg sources list`

Tab-separated table: name, enabled, last successful refresh, last error, feed URL.

## Global flags

| Flag | Default | Purpose |
|------|---------|---------|
| `--db` | OS user config dir | SQLite index path |
| `--sources` | `config/sources.yaml` | Feed manifest |
| `--stale-after` | `24h` | Auto-refresh threshold for `find` |
| `--max-concurrent` | `4` | Parallel feed fetches during `refresh` |

## Manual smoke (live network)

Smoke against **three** small public feeds without pulling the full allowlist:

```bash
export BLERG_DB=/tmp/blerg-smoke.db
rm -f "$BLERG_DB"
blerg --db "$BLERG_DB" --sources config/sources.smoke.yaml refresh
blerg --db "$BLERG_DB" --sources config/sources.smoke.yaml find "latency"
blerg --db "$BLERG_DB" --sources config/sources.smoke.yaml sources list
```

Automated tests use fixture feeds under `testdata/` and do not require network access (`go test ./...`).

## Repository layout

| Path | Purpose |
|------|---------|
| `config/sources.yaml` | Shipped engineering-blog allowlist |
| `config/sources.smoke.yaml` | Small live-feed subset for manual smoke |
| `docs/` | Architecture, ADRs, specs |
| `AGENTS.md` | Instructions for AI coding assistants |
| `CHANGELOG.md` | Notable user-visible changes |

## Development

Build from a checkout when you are changing Blerg. Requires Go 1.26+ (`go.mod`). Optional: [Devbox](https://www.jetify.com/devbox) (`devbox shell`).

```bash
git clone https://github.com/iandesj/Blerg.git && cd Blerg
go build -o bin/blerg ./cmd/blerg
go test ./...
go vet ./...
golangci-lint run ./...
pre-commit run --all-files   # when hooks installed
```

See [architecture](docs/architecture.md) and [ADR 0001](docs/adr/0001-rss-first-allowlist-ingestion.md).

## License

[MIT](LICENSE)
