#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

COMPOSE_FILES="-f deploy/docker-compose.yml -f deploy/docker-compose.business-worker.yml -f deploy/docker-compose.content-worker.yml -f deploy/docker-compose.images.yml"
PROFILES="--profile business-worker --profile content-worker"
ENV_FILE="${ENV_FILE:-deploy/.env}"

if [ -n "${GHCR_USERNAME:-}" ] && [ -n "${GHCR_TOKEN:-}" ]; then
  printf '%s' "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USERNAME" --password-stdin
fi

docker compose --env-file "$ENV_FILE" $COMPOSE_FILES $PROFILES pull cloudbase business-worker content-worker
docker compose --env-file "$ENV_FILE" $COMPOSE_FILES $PROFILES up -d --remove-orphans cloudbase business-worker content-worker
docker compose --env-file "$ENV_FILE" $COMPOSE_FILES $PROFILES ps

