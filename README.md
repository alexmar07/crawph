# Crawph

A concurrent web crawler written in Go that builds a **directed graph** of website link structures. Each crawled page becomes a vertex, and every hyperlink becomes an edge, producing a full map of how pages connect to each other.

## How it works

Crawph models a website as a directed graph:

- **Vertices** represent discovered URLs, indexed by full URL and by domain
- **Edges** represent hyperlinks between pages
- The graph is **thread-safe** — concurrent workers add vertices and edges without races
- URL normalization ensures each page appears as a single vertex, avoiding duplicates

The output graph can be serialized to **JSON** (for inspection and tooling) or **binary** (gob, for compact storage and fast reload).

## Features

- Directed graph construction with concurrent-safe vertex/edge insertion
- Dual indexing (full URL + domain) for fast lookups
- Pipeline architecture (fetch → extract → store)
- robots.txt compliance with Crawl-delay support
- Per-domain rate limiting
- URL normalization and deduplication
- Configurable crawl depth and worker pool
- JSON and binary graph serialization
- YAML configuration file support

## Installation

```bash
go install github.com/alexmar07/crawler-go/cmd@latest
```

Or build from source:

```bash
task
# Binary is at bin/crawph
```

## Usage

### Quick start

```bash
crawph -urls https://example.com
```

### With configuration file

```bash
crawph -config crawph.yml
```

### CLI flags

| Flag | Description | Default |
|------|-------------|---------|
| `-urls` | Comma-separated seed URLs | |
| `-config` | Path to YAML config file | |
| `-workers` | Number of concurrent workers | 5 |
| `-depth` | Maximum crawl depth | 10 |
| `-output` | Output file path | data/result |
| `-format` | Output format (json\|binary) | json |
| `-timeout` | HTTP request timeout | 30s |

CLI flags override config file values.

### Configuration file

```yaml
seeds:
  - https://example.com

crawl:
  max_depth: 10
  max_workers: 5
  timeout: 30s
  user_agent: "Crawph/1.0"

rate_limit:
  default_rps: 1.0
  respect_crawl_delay: true

robots:
  enabled: true

storage:
  format: json
  output: data/result
```

## Development

```bash
# Run tests
task test

# Build
task

# Clean
task clean
```

## License

MIT
