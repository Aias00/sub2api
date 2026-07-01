# Production Server Runbook

This legacy runbook documents the older non-Docker `cloudbase.eu.org` VM flow.
For the current local project runtime, use the Docker compose files under
`deploy/` plus `deploy/.env`.

## Host Identity

- GCE project: `project-f473f4a5-f3a5-4f31-a57`
- Instance: `instance-20260508-104214`
- Zone: `us-central1-a`
- Primary access command:

```bash
gcloud compute ssh --zone "us-central1-a" \
  "instance-20260508-104214" \
  --project "project-f473f4a5-f3a5-4f31-a57"
```

## Runtime Topology

- Public site: `https://cloudbase.eu.org`
- Local service health endpoint: `http://127.0.0.1:8081/health`
- External health endpoint: `https://cloudbase.eu.org/health`
- Server-side repo path: `/home/aias94coffee/sub2api-work/sub2api-main`
- Restart helper path: `/home/aias94coffee/sub2api-work/runtime/restart-sub2api.sh`
- Service owner for repo/build/runtime operations: `aias94coffee`
- Typical login user through `gcloud compute ssh`: `aias`
- Production operations should therefore run through:

```bash
sudo -n -u aias94coffee bash -lc '<commands>'
```

## Important Paths

- Repo root binary used by restart script:
  - `/home/aias94coffee/sub2api-work/sub2api-main/sub2api`
- Backend source tree:
  - `/home/aias94coffee/sub2api-work/sub2api-main/backend`
- Frontend source tree:
  - `/home/aias94coffee/sub2api-work/sub2api-main/frontend`
- Embedded frontend output:
  - `/home/aias94coffee/sub2api-work/sub2api-main/backend/internal/web/dist`
- Runtime logs:
  - `/home/aias94coffee/sub2api-work/runtime/sub2api-8081.log`

## Current Disk Layout

Snapshot as of `2026-06-03`:

- Boot disk: `instance-20260508-104214`
- Size: `50G`
- Root filesystem: `/dev/sda1`
- Filesystem type: `ext4`

Verified state after expansion:

```text
/dev/sda1 ext4 49G used 19G avail 29G use 39%
```

## Production Deployment Flow

For code changes on `main`, use this sequence:

```bash
gcloud compute ssh --zone "us-central1-a" \
  "instance-20260508-104214" \
  --project "project-f473f4a5-f3a5-4f31-a57" \
  --command "sudo -n -u aias94coffee bash -lc '
set -euo pipefail
cd /home/aias94coffee/sub2api-work/sub2api-main
git pull origin main
pnpm install --frozen-lockfile
pnpm run frontend:build
(cd backend && /usr/local/go/bin/go build -tags embed -o ../sub2api ./cmd/server)
/home/aias94coffee/sub2api-work/runtime/restart-sub2api.sh
curl --max-time 15 -fsS http://127.0.0.1:8081/health
'"
```

Then confirm external health:

```bash
curl --max-time 15 -fsS https://cloudbase.eu.org/health
```

## Critical Deployment Pitfall

Do **not** rely on either of these commands:

```bash
cd backend && /usr/local/go/bin/go build -tags embed ./cmd/server
/usr/local/go/bin/go build -tags embed -o sub2api ./backend/cmd/server
```

The first writes a binary like `backend/server`, but the restart script launches
the repo-root binary `./sub2api`. The second runs from the repo root, but this
repository's Go module is under `backend/`, so root builds fail with
`go: cannot find main module`.

If you build only inside `backend/` without `-o ../sub2api`, the running service
may continue using the old root binary even though the compile step succeeded.

Always build from the backend module with an explicit repo-root output:

```bash
cd backend && /usr/local/go/bin/go build -tags embed -o ../sub2api ./cmd/server
```

## Restart Behavior

The tracked helper already handles:

- `DATA_DIR` forwarding
- killing the old process on `SERVER_PORT=8081`
- starting the repo-root `./sub2api`
- local health verification

Default helper assumptions:

- `SERVER_PORT=8081`
- `DATA_DIR=/home/aias94coffee/sub2api-work/sub2api-main`
- runtime logs under `/home/aias94coffee/sub2api-work/runtime/`

## Source Build vs Admin Update API

The admin system update API is **not** the preferred path for this server.

Why:

- This production node runs in source-build mode.
- `/api/v1/admin/system/check-updates` tracks release versions, not arbitrary
  fork commits pushed to `main`.
- A fresh commit on `aias00/main` can therefore be deployed to GitHub while the
  admin update API still reports `has_update: false`.

For this server, deploy by `git pull + rebuild + restart`, not by relying on
the admin update endpoint.

## Safe Cleanup When Disk Gets Tight

If root disk usage climbs again, these are the first safe cleanup targets:

```bash
rm -rf /home/aias94coffee/.cache/go-build/*
rm -rf /home/aias94coffee/go/pkg/mod/cache/*
find /home/aias94coffee/sub2api-work/runtime -type f -name 'sub2api-*.log' -exec truncate -s 0 {} \;
```

Check space before and after:

```bash
df -h /
```

## Verification Checklist

After every production rollout:

1. `git rev-parse --short HEAD` matches the intended commit.
2. `curl -fsS http://127.0.0.1:8081/health` returns `{"env":"production","status":"ok"}`.
3. `curl -fsS https://cloudbase.eu.org/health` returns `{"env":"production","status":"ok"}`.
4. If the change is UI-facing, verify the exact production page in a browser.

## Notes

- Keep credentials, tokens, and one-off browser session artifacts out of this
  file.
- If the server topology changes, update this runbook first, then use the new
  steps for future operations.
