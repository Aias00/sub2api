#!/usr/bin/env bash
set -euo pipefail

RUN_REAL_PROVIDER_CHECK="${RUN_IMAGE_WORKSPACE_REAL_PROVIDER_CHECK:-0}"
RUN_REAL_E2E="${RUN_IMAGE_WORKSPACE_REAL_E2E:-0}"
REQUIRE_READY="${REQUIRE_IMAGE_WORKSPACE_PRODUCTION_READY:-0}"
TMP_RESULT=""

section() {
  echo
  echo "== $* =="
}

bool_env() {
  case "${1:-}" in
    1|true|TRUE|yes|YES|on|ON) return 0 ;;
    *) return 1 ;;
  esac
}

value_of_any() {
  local key
  for key in "$@"; do
    if [[ -n "${!key:-}" ]]; then
      printf '%s' "${!key}"
      return 0
    fi
  done
  return 1
}

configured_any() {
  value_of_any "$@" >/dev/null && echo yes || echo no
}

failures=0
require_any() {
  local label="$1"
  shift
  local configured
  configured="$(configured_any "$@")"
  echo "${label}_configured=${configured}"
  if [[ "$configured" != "yes" ]]; then
    failures=$((failures + 1))
  fi
}

warn_or_fail() {
  local message="$1"
  if [[ "$REQUIRE_READY" == "1" ]]; then
    echo "ERROR: ${message}" >&2
    failures=$((failures + 1))
  else
    echo "WARN: ${message}" >&2
  fi
}

cleanup() {
  if [[ -n "$TMP_RESULT" && -f "$TMP_RESULT" ]]; then
    rm -f "$TMP_RESULT"
  fi
}
trap cleanup EXIT

section "Image Workspace production env"
echo "IMAGE_WORKSPACE_API_BASE_URL=${IMAGE_WORKSPACE_API_BASE_URL:-${BASE_URL:-http://127.0.0.1:8080}}"
echo "IMAGE_WORKSPACE_UPSTREAM_URL=${IMAGE_WORKSPACE_UPSTREAM_URL:-https://api.openai.com/v1/images/generations}"
require_any "upstream_api_key" IMAGE_WORKSPACE_UPSTREAM_API_KEY
require_any "worker_token" IMAGE_WORKSPACE_WORKER_TOKEN
echo "private_worker_without_token=${IMAGE_WORKSPACE_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN:-false}"

if [[ -z "${IMAGE_WORKSPACE_WORKER_TOKEN:-}" ]] && bool_env "${IMAGE_WORKSPACE_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN:-false}"; then
  warn_or_fail "private worker token bypass is enabled; production should use IMAGE_WORKSPACE_WORKER_TOKEN"
fi

section "Image Workspace object-storage env"
object_storage_enabled="${IMAGE_WORKSPACE_OBJECT_STORAGE_ENABLED:-false}"
echo "IMAGE_WORKSPACE_OBJECT_STORAGE_ENABLED=${object_storage_enabled}"
echo "IMAGE_WORKSPACE_OBJECT_STORAGE_PROVIDER=${IMAGE_WORKSPACE_OBJECT_STORAGE_PROVIDER:-r2}"
if bool_env "$object_storage_enabled"; then
  require_any "object_storage_endpoint_or_account" IMAGE_WORKSPACE_OBJECT_STORAGE_ENDPOINT IMAGE_WORKSPACE_R2_ENDPOINT IMAGE_WORKSPACE_R2_ACCOUNT_ID
  require_any "object_storage_access_key" IMAGE_WORKSPACE_OBJECT_STORAGE_ACCESS_KEY_ID IMAGE_WORKSPACE_R2_ACCESS_KEY_ID IMAGE_WORKSPACE_R2_ACCESS_KEY
  require_any "object_storage_secret_key" IMAGE_WORKSPACE_OBJECT_STORAGE_SECRET_ACCESS_KEY IMAGE_WORKSPACE_R2_SECRET_ACCESS_KEY IMAGE_WORKSPACE_R2_SECRET_KEY
  require_any "object_storage_bucket" IMAGE_WORKSPACE_OBJECT_STORAGE_BUCKET IMAGE_WORKSPACE_R2_BUCKET IMAGE_WORKSPACE_R2_BUCKET_NAME
  require_any "object_storage_public_base_url" IMAGE_WORKSPACE_OBJECT_STORAGE_PUBLIC_BASE_URL IMAGE_WORKSPACE_R2_PUBLIC_BASE_URL IMAGE_WORKSPACE_R2_DOMAIN
  echo "object_storage_prefix=${IMAGE_WORKSPACE_OBJECT_STORAGE_PREFIX:-${IMAGE_WORKSPACE_R2_UPLOAD_PATH:-image-workspace}}"
else
  warn_or_fail "IMAGE_WORKSPACE_OBJECT_STORAGE_ENABLED is false; generated images will not be uploaded directly to R2/S3"
  require_any "local_output_dir" IMAGE_WORKSPACE_OUTPUT_DIR
  require_any "local_storage_key_root" IMAGE_WORKSPACE_STORAGE_KEY_ROOT
  public_base="$(value_of_any IMAGE_WORKSPACE_PUBLIC_ARTIFACT_BASE_URL || true)"
  if [[ -z "$public_base" ]]; then
    warn_or_fail "IMAGE_WORKSPACE_PUBLIC_ARTIFACT_BASE_URL is empty; local artifacts may not have public browser URLs"
  else
    echo "local_public_artifact_base_url_configured=yes"
  fi
fi

section "Worker static checks"
node --check tools/image-workspace-acceptance.mjs
npm --prefix tools/image-workspace-worker run check
npm --prefix tools/image-workspace-worker run storage-check
if [[ ! -f tools/image-workspace-worker/Dockerfile ]]; then
  warn_or_fail "tools/image-workspace-worker/Dockerfile is missing"
else
  if grep -q "HEALTHCHECK" tools/image-workspace-worker/Dockerfile \
    && grep -q "npm run healthcheck" tools/image-workspace-worker/Dockerfile; then
    echo "image_workspace_worker_docker_healthcheck=ok"
  else
    warn_or_fail "tools/image-workspace-worker/Dockerfile must include a worker healthcheck"
  fi
fi

section "Monitoring baseline"
if [[ -f deploy/image-workspace-alerts.example.yml ]]; then
  if grep -q "ImageWorkspaceStaleRunningTask" deploy/image-workspace-alerts.example.yml \
    && grep -q "ImageWorkspaceQueueBacklog" deploy/image-workspace-alerts.example.yml \
    && grep -q "ImageWorkspaceFailureBacklog" deploy/image-workspace-alerts.example.yml \
    && grep -q "ImageWorkspaceWorkerAttention" deploy/image-workspace-alerts.example.yml \
    && grep -q "sub2api_image_workspace_worker_health" deploy/image-workspace-alerts.example.yml; then
    echo "image_workspace_alert_rules=ok"
  else
    warn_or_fail "deploy/image-workspace-alerts.example.yml is missing required Image Workspace alert rules"
  fi
else
  warn_or_fail "deploy/image-workspace-alerts.example.yml is missing"
fi

if [[ -f tools/worker-status-metrics.mjs ]]; then
  tmp_status="$(mktemp)"
  tmp_metrics="$(mktemp)"
  cat >"$tmp_status" <<'JSON'
{"health":"attention","total_count":9,"queued_count":2,"running_count":1,"stale_running_count":1,"failed_count":3,"succeeded_count":2,"cancelled_count":1,"artifact_count":4,"oldest_queued_seconds":901,"last_task_age_seconds":120,"attention_reasons":["stale_running_tasks"]}
JSON
  WORKER_STATUS_SERVICE=image \
    WORKER_STATUS_INPUT_PATH="$tmp_status" \
    WORKER_STATUS_METRICS_PATH="$tmp_metrics" \
    node tools/worker-status-metrics.mjs >/tmp/image-workspace-worker-metrics.out
  if grep -q "sub2api_image_workspace_worker_health{health=\"attention\"} 1" "$tmp_metrics" \
    && grep -q "sub2api_image_workspace_stale_running_count 1" "$tmp_metrics" \
    && grep -q "sub2api_image_workspace_artifact_count 4" "$tmp_metrics"; then
    echo "image_workspace_metrics_export=ok"
  else
    warn_or_fail "tools/worker-status-metrics.mjs did not export required Image Workspace metrics"
  fi
  rm -f "$tmp_status" "$tmp_metrics" /tmp/image-workspace-worker-metrics.out
else
  warn_or_fail "tools/worker-status-metrics.mjs is missing"
fi

section "Docker compose overlay"
if command -v docker >/dev/null 2>&1; then
  if POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-dummy}" docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.image-worker.yml --profile image-worker config >/tmp/sub2api-image-worker-compose.yml; then
    echo "docker_compose_image_worker_config=ok"
  else
    warn_or_fail "docker compose Image Workspace worker overlay does not render"
  fi
else
  warn_or_fail "docker is not available; cannot validate Image Workspace worker compose overlay"
fi

if [[ "$RUN_REAL_PROVIDER_CHECK" == "1" ]]; then
  section "Real provider/object-storage smoke"
  if [[ -z "${IMAGE_WORKSPACE_UPSTREAM_API_KEY:-}" ]]; then
    echo "IMAGE_WORKSPACE_UPSTREAM_API_KEY is required for RUN_IMAGE_WORKSPACE_REAL_PROVIDER_CHECK=1" >&2
    exit 2
  fi
  TMP_RESULT="$(mktemp)"
  npm --prefix tools/image-workspace-worker run upstream-check >"$TMP_RESULT"
  node - "$TMP_RESULT" <<'NODE'
const { readFileSync } = await import('node:fs')
const path = process.argv[2]
const text = readFileSync(path, 'utf8')
const start = text.indexOf('{')
if (start < 0) throw new Error('worker upstream-check did not print JSON')
const result = JSON.parse(text.slice(start))
if (result.ok !== true) throw new Error('worker upstream-check did not report ok')
if (!Array.isArray(result.artifacts) || result.artifacts.length < 1) {
  throw new Error('worker upstream-check returned no artifacts')
}
const artifact = result.artifacts[0]
console.log(`real_provider_artifact_count=${result.artifact_count}`)
console.log(`real_provider_storage_provider=${artifact.storage_provider || ''}`)
console.log(`real_provider_storage_key_configured=${artifact.storage_key ? 'yes' : 'no'}`)
console.log(`real_provider_image_url_configured=${artifact.image_url ? 'yes' : 'no'}`)
console.log(`real_provider_file_size=${artifact.file_size || 0}`)
if (artifact.image_url && /^https?:\/\//.test(artifact.image_url)) {
  console.log(`real_provider_image_url=${artifact.image_url}`)
}
NODE
  public_url="$(node - "$TMP_RESULT" <<'NODE'
const { readFileSync } = await import('node:fs')
const text = readFileSync(process.argv[2], 'utf8')
const start = text.indexOf('{')
const result = JSON.parse(text.slice(start))
const url = result.artifacts?.find((item) => /^https?:\/\//.test(item.image_url || ''))?.image_url || ''
process.stdout.write(url)
NODE
)"
  if [[ -n "$public_url" ]]; then
    if curl -fsSI --max-time 15 "$public_url" >/dev/null \
      || curl -fsS --max-time 15 -H "range: bytes=0-0" -o /dev/null "$public_url"; then
      echo "public real_provider_image_url ... ok"
    else
      warn_or_fail "real provider public artifact URL is not reachable"
    fi
  else
    warn_or_fail "real provider check did not return a public HTTP(S) artifact URL"
  fi
else
  section "Real provider/object-storage smoke skipped"
  if [[ "$REQUIRE_READY" == "1" ]]; then
    echo "ERROR: RUN_IMAGE_WORKSPACE_REAL_PROVIDER_CHECK=1 is required for strict production readiness" >&2
    failures=$((failures + 1))
  fi
  cat <<'NOTE'
Set RUN_IMAGE_WORKSPACE_REAL_PROVIDER_CHECK=1 to perform a real upstream image generation and artifact storage check.
This may consume provider quota and write an object to the configured bucket. The check intentionally does not print secrets.
Set REQUIRE_IMAGE_WORKSPACE_PRODUCTION_READY=1 to fail on missing production-required configuration.
NOTE
fi

if [[ "$RUN_REAL_E2E" == "1" ]]; then
  section "Real provider full backend E2E"
  IMAGE_WORKSPACE_E2E_REAL_PROVIDER=1 IMAGE_WORKSPACE_E2E_CLEANUP="${IMAGE_WORKSPACE_E2E_CLEANUP:-0}" node tools/image-workspace-e2e.mjs
else
  section "Real provider full backend E2E skipped"
  if [[ "$REQUIRE_READY" == "1" ]]; then
    echo "ERROR: RUN_IMAGE_WORKSPACE_REAL_E2E=1 is required for strict production readiness" >&2
    failures=$((failures + 1))
  fi
  cat <<'NOTE'
Set RUN_IMAGE_WORKSPACE_REAL_E2E=1 to create a real Image Workspace task through the Go API,
run the Node worker against the configured upstream provider and object storage, verify the
artifact URL, and check task history, balance deduction, usage records, and artifacts.
NOTE
fi

section "Production readiness result"
if [[ "$failures" -gt 0 ]]; then
  echo "production_ready=false"
  echo "missing_or_unsafe_items=${failures}"
  if [[ "$REQUIRE_READY" == "1" ]]; then
    exit 2
  fi
else
  echo "production_ready=true"
fi

echo
echo "Image Workspace production preflight complete."
