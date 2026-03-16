#!/usr/bin/env bash
set -euo pipefail

PROD_PORT="${PROD_PORT:-65530}"
STAGING_PORT="${STAGING_PORT:-65531}"
PROD_PUBLIC_HOST="${PROD_PUBLIC_HOST:-sub2api.airflow.eu.org}"
STAGING_PUBLIC_HOST="${STAGING_PUBLIC_HOST:-sub2api-staging.airflow.eu.org}"
CLOUDFLARED_CONFIG="${CLOUDFLARED_CONFIG:-$HOME/.cloudflared/config.yml}"

PROD_DEPLOY_DIR="${PROD_DEPLOY_DIR:-/Volumes/data/sub2api/deploy}"
STAGING_DEPLOY_DIR="${STAGING_DEPLOY_DIR:-/Volumes/data/sub2api-staging/deploy}"
PROD_CONTAINERS="${PROD_CONTAINERS:-sub2api,sub2api-postgres,sub2api-redis}"
STAGING_CONTAINERS="${STAGING_CONTAINERS:-sub2api-staging,sub2api-staging-postgres,sub2api-staging-redis}"

PASS_COUNT=0
FAIL_COUNT=0

log() { printf '[optimize] %s\n' "$*"; }
pass() { PASS_COUNT=$((PASS_COUNT+1)); printf '[optimize][PASS] %s\n' "$*"; }
fail() { FAIL_COUNT=$((FAIL_COUNT+1)); printf '[optimize][FAIL] %s\n' "$*"; }

require_cmd() {
  local cmd="$1"
  if command -v "$cmd" >/dev/null 2>&1; then
    pass "command exists: $cmd"
  else
    fail "missing command: $cmd"
  fi
}

check_path() {
  local path="$1"
  if [[ -e "$path" ]]; then
    pass "path exists: $path"
  else
    fail "path missing: $path"
  fi
}

check_http_ok() {
  local name="$1"
  local url="$2"
  local out=""
  out="$(curl -fsS --max-time 10 "$url" || true)"
  if [[ "$out" == *"ok"* ]]; then
    pass "$name -> $url => $out"
  else
    fail "$name -> $url => ${out:-<no response>}"
  fi
}

load_containers() {
  local csv="$1"
  local raw_list=""
  raw_list="$(printf '%s' "$csv" | tr ',' '\n')"

  while IFS= read -r item; do
    item="$(printf '%s' "$item" | xargs)"
    [[ -n "$item" ]] && printf '%s\n' "$item"
  done <<< "$raw_list"
}

check_container_running() {
  local name="$1"
  local state
  state="$(docker inspect -f '{{.State.Status}}' "$name" 2>/dev/null || true)"
  if [[ "$state" == "running" ]]; then
    pass "container running: $name"
  else
    fail "container not running: $name (state=${state:-unknown})"
  fi
}

check_container_healthy_if_defined() {
  local name="$1"
  local health
  health="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$name" 2>/dev/null || true)"
  case "$health" in
    healthy|none)
      pass "container health: $name => $health"
      ;;
    *)
      fail "container health bad: $name => ${health:-unknown}"
      ;;
  esac
}

check_cloudflared_mapping() {
  local host="$1"
  local expect_port="$2"
  local actual
  actual="$(python3 - "$CLOUDFLARED_CONFIG" "$host" <<'PY'
from pathlib import Path
import re, sys
cfg = Path(sys.argv[1])
host = sys.argv[2]
if not cfg.exists():
    print('missing')
    raise SystemExit(0)
lines = cfg.read_text().splitlines()
for i, line in enumerate(lines):
    if f"hostname: {host}" in line:
        for j in range(i+1, min(i+12, len(lines))):
            s = lines[j].strip()
            if s.startswith('service:'):
                m = re.search(r'localhost:(\d+)', s)
                if m:
                    print(m.group(1))
                else:
                    print('unknown')
                raise SystemExit(0)
        print('missing-service')
        raise SystemExit(0)
print('missing-host')
PY
)"

  if [[ "$actual" == "$expect_port" ]]; then
    pass "cloudflared mapping: $host -> localhost:$actual"
  else
    fail "cloudflared mapping mismatch: $host -> ${actual:-unknown}, expect $expect_port"
  fi
}

main() {
  log "start new-version optimization precheck"

  require_cmd docker
  require_cmd curl
  require_cmd cloudflared
  require_cmd python3

  check_path "$PROD_DEPLOY_DIR"
  check_path "$STAGING_DEPLOY_DIR"
  check_path "$CLOUDFLARED_CONFIG"

  local prod_count=0
  local staging_count=0
  while IFS= read -r container; do
    prod_count=$((prod_count+1))
    check_container_running "$container"
    check_container_healthy_if_defined "$container"
  done < <(load_containers "$PROD_CONTAINERS")

  while IFS= read -r container; do
    staging_count=$((staging_count+1))
    check_container_running "$container"
    check_container_healthy_if_defined "$container"
  done < <(load_containers "$STAGING_CONTAINERS")

  if [[ "$prod_count" -eq 0 ]]; then
    fail "PROD_CONTAINERS is empty after parsing"
  else
    pass "prod container list parsed: $prod_count items"
  fi

  if [[ "$staging_count" -eq 0 ]]; then
    fail "STAGING_CONTAINERS is empty after parsing"
  else
    pass "staging container list parsed: $staging_count items"
  fi

  check_http_ok "prod local" "http://127.0.0.1:${PROD_PORT}/health"
  check_http_ok "staging local" "http://127.0.0.1:${STAGING_PORT}/health"
  check_http_ok "prod public" "https://${PROD_PUBLIC_HOST}/health"
  check_http_ok "staging public" "https://${STAGING_PUBLIC_HOST}/health"

  check_cloudflared_mapping "$PROD_PUBLIC_HOST" "$PROD_PORT"
  check_cloudflared_mapping "$STAGING_PUBLIC_HOST" "$STAGING_PORT"

  log "precheck finished: pass=$PASS_COUNT fail=$FAIL_COUNT"
  if [[ "$FAIL_COUNT" -gt 0 ]]; then
    exit 1
  fi
}

main "$@"
