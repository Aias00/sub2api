#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
API_BASE="${API_BASE:-${BASE_URL%/}/api/v1}"
AUTH_HEADER="${AUTH_HEADER:-}"
WORKER_TOKEN="${IMAGE_WORKSPACE_WORKER_TOKEN:-}"
RUN_TASK_SIDE_EFFECTS="${RUN_IMAGE_TASK_SIDE_EFFECTS:-0}"
RUN_WORKER_ONCE="${RUN_IMAGE_WORKER_ONCE:-0}"
RUN_STORAGE_CHECK="${RUN_IMAGE_STORAGE_CHECK:-1}"

curl_json() {
  local method="$1"
  local url="$2"
  local body="${3:-}"

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
    args+=(-H "x-image-workspace-worker-token: ${WORKER_TOKEN}")
  fi
  if [[ -n "$body" ]]; then
    args+=(-d "$body")
  fi
  curl "${args[@]}" "$url"
}

require_curl() {
  if ! command -v curl >/dev/null 2>&1; then
    echo "curl is required for Image Workspace smoke checks." >&2
    exit 2
  fi
}

section() {
  echo
  echo "== $* =="
}

require_curl

section "Browser route"
curl -fsSI "${BASE_URL%/}/image-generator" >/dev/null
echo "GET ${BASE_URL%/}/image-generator ... ok"

section "Service health"
curl -fsS "${BASE_URL%/}/health" >/dev/null
echo "GET ${BASE_URL%/}/health ... ok"

section "Worker endpoint health"
if curl_worker GET "${API_BASE%/}/image-workspace/worker/health" "" >/tmp/image-workspace-worker-health.json; then
  echo "GET /image-workspace/worker/health ... ok"
else
  echo "GET /image-workspace/worker/health ... failed" >&2
  echo "If running locally without IMAGE_WORKSPACE_WORKER_TOKEN, set IMAGE_WORKSPACE_ALLOW_PRIVATE_WORKER_WITHOUT_TOKEN=true on the backend and call from a private/loopback address." >&2
fi
if curl_worker GET "${API_BASE%/}/image-workspace/worker/status" "" >/tmp/image-workspace-worker-status.json; then
  echo "GET /image-workspace/worker/status ... ok"
else
  echo "GET /image-workspace/worker/status ... failed" >&2
  echo "Worker status uses the same worker token/private-network protection as worker health." >&2
fi

section "Authenticated user API"
if [[ -z "$AUTH_HEADER" ]]; then
  echo "AUTH_HEADER not set; skipping user-scoped API checks."
  echo "Example: AUTH_HEADER='Authorization: Bearer <token>' BASE_URL=${BASE_URL} tools/image-workspace-smoke.sh"
else
  curl_json GET "${API_BASE%/}/image-workspace/models" "" >/tmp/image-workspace-models.json
  echo "GET /image-workspace/models ... ok"

  curl_json GET "${API_BASE%/}/image-workspace/tasks?page=1&page_size=5" "" >/tmp/image-workspace-tasks.json
  echo "GET /image-workspace/tasks ... ok"

  curl_json GET "${API_BASE%/}/image-workspace/templates" "" >/tmp/image-workspace-templates.json
  echo "GET /image-workspace/templates ... ok"

  if [[ "$RUN_TASK_SIDE_EFFECTS" == "1" ]]; then
    curl_json POST "${API_BASE%/}/image-workspace/tasks" '{
      "prompt": "Smoke test image workspace task. Minimal abstract geometric icon.",
      "provider": "openai",
      "model": "gpt-image-2",
      "size": "1024x1024",
      "quality": "standard",
      "batch_size": 1
    }' >/tmp/image-workspace-created-task.json
    echo "POST /image-workspace/tasks ... ok"
    echo "Created a queued task; run worker with RUN_IMAGE_WORKER_ONCE=1 after upstream credentials are configured."
  else
    echo "Task creation skipped. Set RUN_IMAGE_TASK_SIDE_EFFECTS=1 to create a real image generation task."
  fi
fi

section "Node worker"
if [[ "$RUN_STORAGE_CHECK" == "1" ]]; then
  npm --prefix tools/image-workspace-worker run storage-check
else
  echo "Storage check skipped. Set RUN_IMAGE_STORAGE_CHECK=1 to validate worker storage config."
fi

if [[ "$RUN_WORKER_ONCE" == "1" ]]; then
  IMAGE_WORKSPACE_API_BASE_URL="${BASE_URL%/}" npm --prefix tools/image-workspace-worker run once
else
  echo "Worker --once skipped. Set RUN_IMAGE_WORKER_ONCE=1 to verify task claiming and execution with the local Node worker."
fi

section "Manual real-chain acceptance"
cat <<'CHECKLIST'
Required checks for a full real Image Workspace run:
1. Configure IMAGE_WORKSPACE_UPSTREAM_API_KEY and the target OpenAI-compatible image endpoint.
2. Configure IMAGE_WORKSPACE_WORKER_TOKEN, or private local worker fallback only for local development.
3. Configure local artifact storage, or enable object storage in Admin -> Runtime Settings
   with endpoint/credentials supplied through IMAGE_WORKSPACE_OBJECT_STORAGE_* / R2 env.
4. Authenticate in Cloudbase and open /image-generator.
5. Create a generation task and verify balance pre-authorization succeeds.
6. Start the image worker and verify the task moves queued -> running -> succeeded.
7. Confirm artifacts appear in the task detail with preview/download URLs.
8. If object storage is enabled, verify the artifact URL resolves from the public R2/static domain.
9. Confirm final cost, balance settlement, and image workspace usage audit rows are recorded.
10. Confirm failed upstream tasks refund/settle correctly and expose a clear error message.
CHECKLIST

echo
echo "Image Workspace smoke check complete."
