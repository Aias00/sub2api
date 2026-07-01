#!/usr/bin/env bash
set -euo pipefail

HOT_WORKER_STATUS_PATH="${HOT_WORKER_STATUS_PATH:-/tmp/sub2api-hot-worker-status.json}"
RUN_ONCE_CHECK="${RUN_HOT_COLLECTOR_APPLY_CHECK:-0}"
RUN_SCHEDULE_CHECK="${RUN_HOT_COLLECTOR_SCHEDULE_CHECK:-0}"
REQUIRE_READY="${REQUIRE_HOT_COLLECTOR_PRODUCTION_READY:-0}"
DATABASE_DSN="${HOT_DATABASE_URL:-${DATABASE_URL:-}}"
PGDOCKER_CONTAINER="${PGDOCKER_CONTAINER:-sub2api-postgres}"

failures=0

section() {
  echo
  echo "== $* =="
}

warn_or_fail() {
  local message="$1"
  failures=$((failures + 1))
  if [[ "$REQUIRE_READY" == "1" ]]; then
    echo "ERROR: ${message}" >&2
  else
    echo "WARN: ${message}" >&2
  fi
}

configured() {
  [[ -n "${1:-}" ]] && echo yes || echo no
}

derive_docker_database_url() {
  if [[ -n "$DATABASE_DSN" ]]; then
    return
  fi
  if ! command -v docker >/dev/null 2>&1; then
    return
  fi
  if ! docker inspect "$PGDOCKER_CONTAINER" >/dev/null 2>&1; then
    return
  fi

  local host user password db
  host="$(docker inspect "$PGDOCKER_CONTAINER" --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>/dev/null || true)"
  user="$(docker exec "$PGDOCKER_CONTAINER" printenv POSTGRES_USER 2>/dev/null || true)"
  password="$(docker exec "$PGDOCKER_CONTAINER" printenv POSTGRES_PASSWORD 2>/dev/null || true)"
  db="$(docker exec "$PGDOCKER_CONTAINER" printenv POSTGRES_DB 2>/dev/null || true)"
  if [[ -n "$host" && -n "$user" && -n "$password" && -n "$db" ]]; then
    DATABASE_DSN="postgres://${user}:${password}@${host}:5432/${db}?sslmode=disable"
  fi
}

derive_docker_database_url

section "Hot RSS collector production env"
echo "HOT_WORKER_STATUS_PATH=${HOT_WORKER_STATUS_PATH}"
echo "HOT_RSS_COLLECT_INTERVAL_MS=${HOT_RSS_COLLECT_INTERVAL_MS:-1800000}"
echo "HOT_RSS_COLLECT_MAX_BACKOFF_MS=${HOT_RSS_COLLECT_MAX_BACKOFF_MS:-600000}"
echo "HOT_RSS_COLLECT_ON_START=${HOT_RSS_COLLECT_ON_START:-true}"
echo "RUN_HOT_COLLECTOR_APPLY_CHECK=${RUN_ONCE_CHECK}"
echo "RUN_HOT_COLLECTOR_SCHEDULE_CHECK=${RUN_SCHEDULE_CHECK}"
echo "database_url_configured=$(configured "$DATABASE_DSN")"
echo "database_url_source=$([[ -n "${HOT_DATABASE_URL:-${DATABASE_URL:-}}" ]] && echo env || { [[ -n "$DATABASE_DSN" ]] && echo docker-container || echo missing; })"

if [[ ! -f tools/x-atuo/src/x_atuo/hot_rss_worker.py ]]; then
  echo "ERROR: tools/x-atuo/src/x_atuo/hot_rss_worker.py is missing" >&2
  exit 2
fi
if [[ ! -f tools/hot-collector-status-metrics.mjs ]]; then
  echo "ERROR: tools/hot-collector-status-metrics.mjs is missing" >&2
  exit 2
fi
if [[ ! -f tools/content-worker.Dockerfile ]]; then
  echo "ERROR: tools/content-worker.Dockerfile is missing" >&2
  exit 2
fi
if [[ ! -f deploy/docker-compose.content-worker.yml ]]; then
  echo "ERROR: deploy/docker-compose.content-worker.yml is missing" >&2
  exit 2
fi
if [[ ! -f deploy/hot-collector-alerts.example.yml ]]; then
  warn_or_fail "deploy/hot-collector-alerts.example.yml is missing"
fi
if [[ -z "$DATABASE_DSN" ]]; then
  warn_or_fail "HOT_DATABASE_URL or DATABASE_URL is required for the RSS collector"
fi
if ! command -v psql >/dev/null 2>&1; then
  warn_or_fail "psql is required for local RSS collector verification"
fi

section "Hot worker static checks"
python3 -m compileall -q tools/x-atuo/src/x_atuo/hot_rss_worker.py tools/x-atuo/src/x_atuo/content_worker.py
node --check tools/hot-collector-status-metrics.mjs
bash -n tools/hot-content-integrity.sh

section "Hot monitoring baseline"
if [[ -f deploy/hot-collector-alerts.example.yml ]]; then
  if grep -q "HotCollectorWorkerUnhealthy" deploy/hot-collector-alerts.example.yml \
    && grep -q "HotCollectorStatusStale" deploy/hot-collector-alerts.example.yml \
    && grep -q "sub2api_hot_collector_status_age_seconds" deploy/hot-collector-alerts.example.yml; then
    echo "hot_collector_alert_rules=ok"
  else
    warn_or_fail "deploy/hot-collector-alerts.example.yml is missing required Hot collector alerts"
  fi
fi

if [[ -n "$DATABASE_DSN" && "$(command -v psql || true)" != "" ]]; then
  section "PostgreSQL source baseline"
  psql "$DATABASE_DSN" -v ON_ERROR_STOP=1 -P null='NULL' -c "
SELECT
  (SELECT COUNT(*) FROM hot_sources WHERE enabled = TRUE AND adapter_kind = 'rss-generic') AS enabled_rss_sources,
  (SELECT COUNT(*) FROM hot_items) AS hot_items,
  (SELECT MAX(updated_at) FROM hot_items) AS hot_items_max_updated_at;
"
fi

section "Hot RSS worker once check"
if [[ "$RUN_ONCE_CHECK" == "1" ]]; then
  if [[ -z "$DATABASE_DSN" ]]; then
    echo "ERROR: RUN_HOT_COLLECTOR_APPLY_CHECK=1 requires HOT_DATABASE_URL or DATABASE_URL" >&2
    exit 2
  fi
  DATABASE_URL="$DATABASE_DSN" \
  HOT_WORKER_STATUS_PATH="$HOT_WORKER_STATUS_PATH" \
  HOT_RSS_WORKER_STATUS_PATH="$HOT_WORKER_STATUS_PATH" \
  PYTHONPATH=tools/x-atuo/src python3 -m x_atuo.hot_rss_worker --once
  DATABASE_URL="$DATABASE_DSN" \
  HOT_WORKER_STATUS_PATH="$HOT_WORKER_STATUS_PATH" \
  HOT_RSS_WORKER_STATUS_PATH="$HOT_WORKER_STATUS_PATH" \
  PYTHONPATH=tools/x-atuo/src python3 -m x_atuo.hot_rss_worker --healthcheck
  node - "$HOT_WORKER_STATUS_PATH" <<'NODE'
const { readFileSync } = await import('node:fs')
const status = JSON.parse(readFileSync(process.argv[2], 'utf8'))
function assert(condition, message) {
  if (!condition) throw new Error(message)
}
assert(status.apply === true, 'RSS collector status should record apply=true')
assert(String(status.storage || '') === 'postgres', 'RSS collector status should record storage=postgres')
assert(Number(status.run_count || 0) >= 1, 'RSS collector status is missing run_count')
assert(Number(status.success_count || 0) >= 1, 'RSS collector status is missing success_count')
console.log(`hot_rss_worker_status_apply=${status.apply}`)
console.log(`hot_rss_worker_status_storage=${status.storage}`)
console.log(`hot_rss_worker_status_mode=${status.mode}`)
NODE
  echo "hot_rss_worker_once_check=ok"
else
  if [[ "$REQUIRE_READY" == "1" ]]; then
    echo "ERROR: RUN_HOT_COLLECTOR_APPLY_CHECK=1 is required for strict Hot collector production readiness" >&2
  else
    echo "WARN: RUN_HOT_COLLECTOR_APPLY_CHECK=1 was not set; target PostgreSQL write path is unproven" >&2
  fi
  failures=$((failures + 1))
fi

section "Hot worker metrics export"
if [[ -s "$HOT_WORKER_STATUS_PATH" ]]; then
  metrics_path="${HOT_WORKER_STATUS_PATH}.prom"
  HOT_WORKER_STATUS_PATH="$HOT_WORKER_STATUS_PATH" \
  HOT_WORKER_METRICS_PATH="$metrics_path" \
  node tools/hot-collector-status-metrics.mjs
  if grep -q "sub2api_hot_collector_status_age_seconds" "$metrics_path" \
    && grep -q "sub2api_hot_collector_last_success_age_seconds" "$metrics_path" \
    && grep -q "sub2api_hot_collector_run_count" "$metrics_path"; then
    echo "hot_worker_metrics_export=ok"
  else
    warn_or_fail "hot collector metrics export is missing required metrics"
  fi
else
  warn_or_fail "hot worker status file is missing; run once check or start the worker before exporting metrics"
fi

section "Hot worker schedule check"
if [[ "$RUN_SCHEDULE_CHECK" == "1" ]]; then
  if [[ -z "$DATABASE_DSN" ]]; then
    echo "ERROR: RUN_HOT_COLLECTOR_SCHEDULE_CHECK=1 requires HOT_DATABASE_URL or DATABASE_URL" >&2
    exit 2
  fi
  schedule_status_path="${HOT_WORKER_STATUS_PATH}.schedule.json"
  DATABASE_URL="$DATABASE_DSN" \
  HOT_WORKER_STATUS_PATH="$schedule_status_path" \
  HOT_RSS_WORKER_STATUS_PATH="$schedule_status_path" \
  HOT_RSS_COLLECT_INTERVAL_MS=1000 \
  HOT_RSS_COLLECT_MAX_BACKOFF_MS=1000 \
  HOT_RSS_COLLECT_ON_START=true \
  HOT_RSS_COLLECT_MAX_RUNS=2 \
  PYTHONPATH=tools/x-atuo/src python3 -m x_atuo.hot_rss_worker
  DATABASE_URL="$DATABASE_DSN" \
  HOT_WORKER_STATUS_PATH="$schedule_status_path" \
  HOT_RSS_WORKER_STATUS_PATH="$schedule_status_path" \
  HOT_RSS_COLLECT_INTERVAL_MS=1000 \
  HOT_RSS_COLLECT_MAX_BACKOFF_MS=1000 \
  PYTHONPATH=tools/x-atuo/src python3 -m x_atuo.hot_rss_worker --healthcheck
  echo "hot_rss_worker_schedule_check=ok"
else
  if [[ "$REQUIRE_READY" == "1" ]]; then
    echo "ERROR: RUN_HOT_COLLECTOR_SCHEDULE_CHECK=1 is required for strict Hot collector production readiness" >&2
    failures=$((failures + 1))
  else
    echo "Schedule check skipped. Set RUN_HOT_COLLECTOR_SCHEDULE_CHECK=1 to prove the worker loop executes more than one scheduled run."
  fi
fi

section "Docker compose overlay"
if command -v docker >/dev/null 2>&1; then
  if POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-dummy}" \
    docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.content-worker.yml --profile content-worker config >/tmp/sub2api-content-worker-compose.yml; then
    echo "docker_compose_content_worker_config=ok"
  else
    warn_or_fail "docker compose content-worker overlay did not render"
  fi
else
  warn_or_fail "docker command is unavailable; content-worker compose overlay was not rendered"
fi

section "Production readiness result"
if [[ "$failures" -gt 0 ]]; then
  echo "hot_collector_production_ready=false"
  echo "missing_or_unsafe_items=${failures}"
  if [[ "$REQUIRE_READY" == "1" ]]; then
    exit 2
  fi
else
  echo "hot_collector_production_ready=true"
fi

echo
echo "Hot RSS collector production preflight complete."
