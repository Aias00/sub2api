#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
API_BASE="${API_BASE:-${BASE_URL%/}/api/v1}"
STRICT="${REQUIRE_WECHAT_EXPORT_PRODUCTION_READY:-0}"
RUN_LIVE_WORKER_CHECK="${RUN_WECHAT_EXPORT_LIVE_WORKER_CHECK:-0}"

missing_or_unsafe=0

section() {
  echo
  echo "== $* =="
}

is_truthy() {
  case "${1:-}" in
    1|true|TRUE|yes|YES|on|ON) return 0 ;;
    *) return 1 ;;
  esac
}

warn_or_fail() {
  echo "WARN: $*" >&2
  missing_or_unsafe=$((missing_or_unsafe + 1))
  if is_truthy "$STRICT"; then
    return 2
  fi
  return 0
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    warn_or_fail "$1 is required for WeChat export production preflight"
    return 0
  fi
}

configured() {
  [[ -n "${!1:-}" ]] && echo yes || echo no
}

section "WeChat export production env"
echo "BASE_URL=${BASE_URL}"
echo "API_BASE=${API_BASE}"
echo "worker_token_configured=$(configured WECHAT_EXPORT_WORKER_TOKEN)"
echo "private_worker_without_token=${WECHAT_EXPORT_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN:-false}"
echo "output_dir=${WECHAT_EXPORT_OUTPUT_DIR:-/app/data/wechat-export}"
echo "storage_root=${WECHAT_EXPORT_STORAGE_ROOT:-/app/data/wechat-export}"
echo "storage_key_root=${WECHAT_EXPORT_STORAGE_KEY_ROOT:-/app/data/wechat-export}"
echo "zip_remote_host_allowlist_configured=$(configured WECHAT_EXPORT_ZIP_REMOTE_HOST_ALLOWLIST)"

if [[ -z "${WECHAT_EXPORT_WORKER_TOKEN:-}" ]]; then
  warn_or_fail "WECHAT_EXPORT_WORKER_TOKEN is empty; production worker APIs should require a shared token"
fi

if is_truthy "${WECHAT_EXPORT_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN:-false}"; then
  warn_or_fail "WECHAT_EXPORT_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN is enabled; production should use explicit worker tokens"
fi

if [[ -n "${WECHAT_EXPORT_ARTIFACT_PUBLIC_BASE_URL:-${WECHAT_EXPORT_PUBLIC_ARTIFACT_BASE_URL:-}}" && -z "${WECHAT_EXPORT_ZIP_REMOTE_HOST_ALLOWLIST:-}" ]]; then
  warn_or_fail "remote artifact public base URL is configured but WECHAT_EXPORT_ZIP_REMOTE_HOST_ALLOWLIST is empty"
fi

section "Worker static checks"
node --check tools/wechat-export-acceptance.mjs
if command -v npm >/dev/null 2>&1; then
  npm --prefix tools/wechat-worker run typecheck
  npm --prefix tools/wechat-worker run fidelity-check
else
  warn_or_fail "npm is not available; cannot run WeChat worker typecheck/fidelity checks"
fi

if [[ ! -f tools/business-worker.Dockerfile ]]; then
  warn_or_fail "tools/business-worker.Dockerfile is missing"
else
  if grep -q "HEALTHCHECK" tools/business-worker.Dockerfile \
    && grep -q "business-worker.mjs --healthcheck" tools/business-worker.Dockerfile; then
    echo "business_worker_docker_healthcheck=ok"
  else
    warn_or_fail "tools/business-worker.Dockerfile must include a worker healthcheck"
  fi
fi

section "Monitoring baseline"
if [[ -f deploy/wechat-export-alerts.example.yml ]]; then
  if grep -q "WeChatExportStaleRunningTask" deploy/wechat-export-alerts.example.yml \
    && grep -q "WeChatExportQueueBacklog" deploy/wechat-export-alerts.example.yml \
    && grep -q "WeChatExportFailureBacklog" deploy/wechat-export-alerts.example.yml \
    && grep -q "WeChatExportWorkerAttention" deploy/wechat-export-alerts.example.yml \
    && grep -q "cloudbase_wechat_export_worker_health" deploy/wechat-export-alerts.example.yml; then
    echo "wechat_export_alert_rules=ok"
  else
    warn_or_fail "deploy/wechat-export-alerts.example.yml is missing required WeChat export alert rules"
  fi
else
  warn_or_fail "deploy/wechat-export-alerts.example.yml is missing"
fi

if [[ -f tools/worker-status-metrics.mjs ]]; then
  tmp_status="$(mktemp)"
  tmp_metrics="$(mktemp)"
  cat >"$tmp_status" <<'JSON'
{"health":"attention","total_count":7,"queued_count":2,"running_count":1,"stale_running_count":1,"failed_count":3,"completed_count":1,"cancelled_count":0,"oldest_queued_seconds":901,"last_task_age_seconds":120,"attention_reasons":["stale_running_tasks"]}
JSON
  WORKER_STATUS_SERVICE=wechat \
    WORKER_STATUS_INPUT_PATH="$tmp_status" \
    WORKER_STATUS_METRICS_PATH="$tmp_metrics" \
    node tools/worker-status-metrics.mjs >/tmp/wechat-export-worker-metrics.out
  if grep -q "cloudbase_wechat_export_worker_health{health=\"attention\"} 1" "$tmp_metrics" \
    && grep -q "cloudbase_wechat_export_stale_running_count 1" "$tmp_metrics" \
    && grep -q "cloudbase_wechat_export_oldest_queued_seconds 901" "$tmp_metrics"; then
    echo "wechat_export_metrics_export=ok"
  else
    warn_or_fail "tools/worker-status-metrics.mjs did not export required WeChat export metrics"
  fi
  rm -f "$tmp_status" "$tmp_metrics" /tmp/wechat-export-worker-metrics.out
else
  warn_or_fail "tools/worker-status-metrics.mjs is missing"
fi

section "Docker compose overlay"
if command -v docker >/dev/null 2>&1; then
  if POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-dummy}" docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.business-worker.yml --profile business-worker config >/tmp/cloudbase-business-worker-compose.yml; then
    echo "docker_compose_business_worker_config=ok"
  else
    warn_or_fail "docker compose business worker overlay does not render"
  fi
else
  warn_or_fail "docker is not available; cannot validate business worker compose overlay"
fi

section "Live worker health"
if is_truthy "$RUN_LIVE_WORKER_CHECK"; then
  if command -v curl >/dev/null 2>&1; then
    args=(-fsS -H "accept: application/json")
    if [[ -n "${WECHAT_EXPORT_WORKER_TOKEN:-}" ]]; then
      args+=(-H "x-wechat-worker-token: ${WECHAT_EXPORT_WORKER_TOKEN}")
    fi
    curl "${args[@]}" "${API_BASE%/}/wechat/worker/health" >/tmp/wechat-export-worker-health.json
    echo "wechat_worker_live_health=ok"
  else
    warn_or_fail "curl is required for live worker health check"
  fi
else
  echo "Live worker check skipped. Set RUN_WECHAT_EXPORT_LIVE_WORKER_CHECK=1 to call /wechat/worker/health."
fi

section "Production readiness result"
if [[ "$missing_or_unsafe" -gt 0 ]]; then
  echo "wechat_export_production_ready=false"
  echo "missing_or_unsafe_items=${missing_or_unsafe}"
else
  echo "wechat_export_production_ready=true"
fi

echo
echo "WeChat export production preflight complete."
