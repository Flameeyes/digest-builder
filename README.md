<!--
SPDX-FileCopyrightText: 2026 Bob the Skull <bob.github@defp.uk>
SPDX-License-Identifier: 0BSD
-->

# digest-builder

Fetch articles from a [Wallabag](https://wallabag.org/) instance tagged
`digest-pending`, bundle them into a timestamped EPUB optimised for e-ink
readers, and mark them as consumed.

Designed to run as a batch job — on a schedule via cron, Docker, or a
TrueNAS custom app. The generated EPUBs can be picked up by a separate
sync tool (e.g. `rmapi` for reMarkable) for delivery to your reader.

## Quick Start

### From source

```bash
# Clone
git clone https://git.example.net/bob/digest-builder.git
cd digest-builder

# Build
go build -o digest-builder .

# Configure
cp config.example.yaml config.yaml
# Edit config.yaml with your Wallabag credentials

# Run
./digest-builder --config config.yaml

# Dry run (build EPUB but don't mark consumed)
./digest-builder --config config.yaml --dry-run
```

### With Docker

```bash
cp .env.example .env
# Edit .env with your Wallabag credentials

cp config.example.yaml config.yaml
# Edit config.yaml with your section preferences

docker compose up --build
```

### On TrueNAS SCALE

See [`truenas/README.md`](truenas/README.md) for detailed deployment
instructions.

## Configuration

digest-builder reads its configuration from a YAML file and/or environment
variables. Environment variables take precedence.

### Config file

See [`config.example.yaml`](config.example.yaml) for the full config with
comments.

Key sections:

| Section | Purpose |
|---------|---------|
| `wallabag` | API connection (URL, OAuth2 credentials) |
| `digest.sections` | Section ordering and tag-to-section mapping |
| `digest.pending_tag` | Wallabag tag that marks articles for inclusion (default: `digest-pending`) |
| `digest.exclude_patterns` | Title substrings to skip |
| `output` | EPUB output directory and filename pattern |
| `state` | OAuth2 token persistence path |

### Environment variables

| Variable | Overrides |
|----------|-----------|
| `WALLABAG_URL` | `wallabag.url` |
| `WALLABAG_CLIENT_ID` | `wallabag.client_id` |
| `WALLABAG_CLIENT_SECRET` | `wallabag.client_secret` |
| `WALLABAG_USERNAME` | `wallabag.username` |
| `WALLABAG_PASSWORD` | `wallabag.password` |
| `DIGEST_OUTPUT_DIR` | `output.dir` |
| `DIGEST_STATE_FILE` | `state.file` |

## How It Works

1. **Fetch**: Queries Wallabag for all entries tagged `digest-pending`
2. **Classify**: Groups articles into sections based on their Wallabag tags
   (AI & CASE Tools → `ai_case`, Financial Times → `ft_subscriber`, etc.)
3. **Build**: Generates an EPUB with a nested table of contents — one
   top-level entry per section, articles as sub-chapters, e-ink-optimised
   CSS
4. **Mark consumed**: Removes the `digest-pending` tag and adds a
   `digest-YYYY-MM-DD-HHMM` tag to each processed article

### Section ordering

Articles are classified by their first matching Wallabag tag. The section
order in the EPUB matches the order in `digest.sections` in the config.
Articles with no matching tag are placed in `general_tech`.

### Default sections

| Section Key | Label | Default Tag |
|-------------|-------|-------------|
| `ai_case` | AI & CASE Tools | AI & CASE Tools |
| `general_tech` | Hacker News & Tech | Hacker News & Tech |
| `technical_writing` | Technical Writing | Technical Writing |
| `photography` | Photography | Photography |
| `ft_subscriber` | Financial Times | Financial Times |
| `hbr_subscriber` | Harvard Business Review | Harvard Business Review |
| `ftav_further_reading` | FTAV Further Reading | FTAV Further Reading |

## CLI Flags

| Flag | Description |
|------|-------------|
| `--config <path>` | Path to config.yaml (optional if using env vars) |
| `--dry-run` | Build EPUB but skip marking articles consumed |
| `--version` | Print version and exit |

## Output

EPUBs are written to the configured output directory with timestamped filenames:

```
output/
  digest-2026-05-29-0500.epub
  digest-2026-05-30-0500.epub
```

## Building

```bash
# Standard build
go build -o digest-builder .

# With version tag
go build -ldflags="-s -w -X main.version=v1.0.0" -o digest-builder .

# Static binary for Docker
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o digest-builder .

# Cross-compile for ARM NAS
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o digest-builder .
```

## Docker image

```bash
docker build -t digest-builder:latest .
```

The multi-stage Dockerfile produces a ~25MB Alpine-based image.

## License

0BSD — see [`LICENSES/0BSD.txt`](LICENSES/0BSD.txt).
