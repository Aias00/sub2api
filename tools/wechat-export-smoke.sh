#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
API_BASE="${API_BASE:-${BASE_URL%/}/api/v1}"
AUTH_HEADER="${AUTH_HEADER:-}"
WORKER_TOKEN="${WECHAT_EXPORT_WORKER_TOKEN:-}"
RUN_WORKER_ONCE="${RUN_WORKER_ONCE:-0}"
RUN_QR_SIDE_EFFECTS="${RUN_QR_SIDE_EFFECTS:-0}"
RUN_COMPLETED_TASK_ACCEPTANCE="${RUN_WECHAT_EXPORT_COMPLETED_TASK_ACCEPTANCE:-0}"
ACCEPTANCE_TASK_ID="${WECHAT_EXPORT_ACCEPTANCE_TASK_ID:-}"
ACCEPTANCE_FORMATS="${WECHAT_EXPORT_ACCEPTANCE_FORMATS:-html,markdown,json}"

curl_json() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  shift 3 || true

  local args=(-fsS -X "$method" -H "content-type: application/json")
  if [[ -n "$AUTH_HEADER" ]]; then
    args+=(-H "$AUTH_HEADER")
  fi
  if [[ -n "$body" ]]; then
    args+=(-d "$body")
  fi
  curl "${args[@]}" "$url"
}

curl_worker() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  local args=(-fsS -X "$method" -H "content-type: application/json")
  if [[ -n "$WORKER_TOKEN" ]]; then
    args+=(-H "x-wechat-worker-token: ${WORKER_TOKEN}")
  fi
  if [[ -n "$body" ]]; then
    args+=(-d "$body")
  fi
  curl "${args[@]}" "$url"
}

require_curl() {
  if ! command -v curl >/dev/null 2>&1; then
    echo "curl is required for WeChat export smoke checks." >&2
    exit 2
  fi
}

section() {
  echo
  echo "== $* =="
}

require_curl

section "Browser routes"
curl -fsSI "${BASE_URL%/}/wechat" >/dev/null
echo "GET ${BASE_URL%/}/wechat ... ok"
curl -fsSI "${BASE_URL%/}/wechat-export" >/dev/null
echo "GET ${BASE_URL%/}/wechat-export alias ... ok"

section "Service health"
curl -fsS "${BASE_URL%/}/health" >/dev/null
echo "GET ${BASE_URL%/}/health ... ok"

section "Worker endpoint health"
if curl_worker GET "${API_BASE%/}/wechat/worker/health" "" >/tmp/wechat-worker-health.json; then
  echo "GET ${API_BASE%/}/wechat/worker/health ... ok"
else
  echo "GET ${API_BASE%/}/wechat/worker/health ... failed"
  echo "If running locally without WECHAT_EXPORT_WORKER_TOKEN, set WECHAT_EXPORT_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN=true on the backend and call from a private/loopback address." >&2
fi

section "Authenticated user API"
if [[ -z "$AUTH_HEADER" ]]; then
  echo "AUTH_HEADER not set; skipping user-scoped API checks."
  echo "Example: AUTH_HEADER='Authorization: Bearer <token>' BASE_URL=${BASE_URL} tools/wechat-export-smoke.sh"
else
  curl_json GET "${API_BASE%/}/wechat/session" "" >/tmp/wechat-session.json
  echo "GET /wechat/session ... ok"

  curl_json GET "${API_BASE%/}/wechat/worker/status" "" >/tmp/wechat-worker-status.json
  echo "GET /wechat/worker/status ... ok"

  curl_json GET "${API_BASE%/}/wechat/articles?page=1&page_size=5" "" >/tmp/wechat-articles.json
  echo "GET /wechat/articles ... ok"

  curl_json GET "${API_BASE%/}/wechat/tasks?page=1&page_size=5" "" >/tmp/wechat-tasks.json
  echo "GET /wechat/tasks ... ok"

  if [[ "$RUN_QR_SIDE_EFFECTS" == "1" ]]; then
    curl_json POST "${API_BASE%/}/wechat/session/qrcode" "{}" >/tmp/wechat-qrcode.json
    echo "POST /wechat/session/qrcode ... ok"
    echo "Scan the QR code from /tmp/wechat-qrcode.json, then keep polling until session.status is ready."
  else
    echo "QR creation skipped. Set RUN_QR_SIDE_EFFECTS=1 to create a real WeChat login session."
  fi
fi

section "Node worker"
if [[ "$RUN_WORKER_ONCE" == "1" ]]; then
  SUB2API_BASE_URL="${API_BASE%/}" npm --prefix tools/wechat-worker run worker -- --once
else
  echo "Worker --once skipped. Set RUN_WORKER_ONCE=1 to verify task claiming with the local Node worker."
fi

section "Completed task acceptance"
if [[ "$RUN_COMPLETED_TASK_ACCEPTANCE" == "1" ]]; then
  if [[ -z "$AUTH_HEADER" ]]; then
    echo "AUTH_HEADER is required when RUN_WECHAT_EXPORT_COMPLETED_TASK_ACCEPTANCE=1" >&2
    exit 2
  fi
  AUTH_HEADER="$AUTH_HEADER" \
    API_BASE="$API_BASE" \
    WECHAT_EXPORT_ACCEPTANCE_TASK_ID="$ACCEPTANCE_TASK_ID" \
    WECHAT_EXPORT_ACCEPTANCE_FORMATS="$ACCEPTANCE_FORMATS" \
    node tools/wechat-export-acceptance.mjs
else
  echo "Completed task acceptance skipped. Set RUN_WECHAT_EXPORT_COMPLETED_TASK_ACCEPTANCE=1 after a real export task completes."
fi

section "Manual real-chain acceptance"
cat <<'CHECKLIST'
Required manual checks for a full real WeChat export run:
1. Open /wechat and authenticate in Sub2API.
2. Create a WeChat QR session and scan it until status becomes ready.
3. Search a real official account and bind it, or import at least one mp.weixin.qq.com article link.
4. Sync account articles or select imported links.
5. Create an export task with html, markdown, json, and optionally engagement data.
6. Start the worker and verify the task moves queued -> running -> completed/completed_with_errors.
7. Expand the task timeline and confirm persisted log events are visible.
8. Load artifacts and download every generated file.
9. Download the task ZIP package and confirm it contains local worker artifacts.
10. Inspect HTML/Markdown/JSON for title, prompt/article body, images/media blocks, source URL, and engagementFetchStatus.
11. If engagement data is unavailable, confirm the manifest explains whether session, appmsg_token, or WeChat response caused the degradation.
12. Run automated completed-task acceptance:
    AUTH_HEADER='Authorization: Bearer <token>' RUN_WECHAT_EXPORT_COMPLETED_TASK_ACCEPTANCE=1 WECHAT_EXPORT_ACCEPTANCE_TASK_ID=<task_id> BASE_URL=<url> tools/wechat-export-smoke.sh
CHECKLIST

echo
echo "WeChat export smoke check complete."
