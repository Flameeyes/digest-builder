# TrueNAS SCALE Deployment Guide

Deploy digest-builder as a custom app on TrueNAS SCALE 24.10 (Electric Eel) or later.

## Prerequisites

- TrueNAS SCALE 24.10+ (Electric Eel / Dragonfish with Docker support)
- A running [Wallabag](https://wallabag.org/) instance accessible from TrueNAS
- A ZFS pool with space for datasets

## Quick Start

### 1. Create datasets

Create a parent dataset for the app and subdirectories for config, output, and state:

```bash
# Via TrueNAS UI: Datasets > Add Dataset
# Or via SSH:
mkdir -p /mnt/pool/apps/digest-builder/{config,output,state}
```

### 2. Add configuration file

Copy `config.example.yaml` from the repo root to your config dataset:

```bash
cp config.example.yaml /mnt/pool/apps/digest-builder/config/config.yaml
```

Edit it with your section ordering, article limits, and other preferences.
Wallabag credentials can go in the config file OR be set as environment
variables in the compose file (env vars take precedence).

### 3. Deploy via TrueNAS UI

1. Go to **Apps** → **Discover Apps**
2. Click the **⋮** (three-dot menu) next to "Custom App"
3. Select **Install via YAML**
4. Enter a name: `digest-builder`
5. Paste the contents of [`compose.yaml`](compose.yaml) into the **Custom Config** editor
6. **Edit the placeholder values**:
   - Replace all `«angle-quoted»` values with your real Wallabag credentials
   - Adjust the `/mnt/pool/apps/digest-builder/...` volume paths to match your dataset locations
7. Click **Save**

The app will pull the image, run the digest build, and exit.

### 3a. Alternative: Deploy via SSH

```bash
# Copy compose.yaml to your dataset
cp truenas/compose.yaml /mnt/pool/apps/digest-builder/compose.yaml

# Edit credentials and paths
nano /mnt/pool/apps/digest-builder/compose.yaml

# Run
cd /mnt/pool/apps/digest-builder
docker compose up
```

### 4. Schedule daily runs

digest-builder is a batch job — it runs once, builds the EPUB, and exits.
To run it automatically:

**Option A: TrueNAS Cron Job (recommended)**

1. Go to **System** → **Advanced** → **Cron Jobs** → **Add**
2. Command:
   ```
   docker compose -f /mnt/pool/apps/digest-builder/compose.yaml up
   ```
3. Schedule: `0 5 * * *` (daily at 05:00, adjust as needed)
4. User: `root`

**Option B: Self-contained scheduler sidecar**

Uncomment the `scheduler` service in `compose.yaml`. This runs a tiny
Alpine container that uses `crond` to `docker start` the main container
on schedule. Requires mounting the Docker socket.

## Building the image locally

If you prefer to build from source instead of pulling a pre-built image:

```bash
# Clone the repo to a dataset
cd /mnt/pool/apps/digest-builder
git clone https://git.example.net/bob/digest-builder.git src

# Build
cd src
docker build -t digest-builder:latest .
```

Then change the `image:` line in `compose.yaml` to `digest-builder:latest`.

## Volume layout

| Mount point | Purpose | Access |
|---|---|---|
| `/etc/digest-builder/config.yaml` | Application config | Read-only |
| `/output` | Generated EPUB files | Read-write |
| `/state` | OAuth2 token persistence | Read-write |

## Environment variables

All Wallabag credentials can be set via environment variables, which
override values from the config file:

| Variable | Description |
|---|---|
| `WALLABAG_URL` | Wallabag instance URL |
| `WALLABAG_CLIENT_ID` | OAuth2 client ID |
| `WALLABAG_CLIENT_SECRET` | OAuth2 client secret |
| `WALLABAG_USERNAME` | Wallabag username |
| `WALLABAG_PASSWORD` | Wallabag password |
| `TZ` | Timezone (default: `Europe/London`) |

## Output

EPUBs are written to the `/output` mount with timestamped filenames:

```
digest-2026-05-29-0500.epub
```

Point your e-reader sync tool (e.g. rmapi for reMarkable) at the output
dataset to pick up new digests automatically.

## Troubleshooting

**App shows "Stopped" immediately:**
This is normal — digest-builder is a batch job. Check the logs:
```bash
docker logs digest-builder
```

**Can't reach Wallabag:**
If Wallabag is on your local network, make sure the container can reach it.
TrueNAS custom apps use bridge networking by default. You may need to add
`network_mode: host` to the compose file, or ensure Wallabag is accessible
via the TrueNAS host IP.

**OAuth2 token errors:**
Delete the token state file and re-run:
```bash
rm /mnt/pool/apps/digest-builder/state/token.json
docker compose -f /mnt/pool/apps/digest-builder/compose.yaml up
```
