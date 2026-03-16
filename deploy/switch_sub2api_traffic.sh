#!/usr/bin/env bash
set -euo pipefail

HOSTNAME_TARGET="${HOSTNAME_TARGET:-sub2api.airflow.eu.org}"
STAGING_HOSTNAME_TARGET="${STAGING_HOSTNAME_TARGET:-sub2api-staging.airflow.eu.org}"
PROD_PORT="${PROD_PORT:-65530}"
STAGING_PORT="${STAGING_PORT:-65531}"
PROD_HEALTH_HINT="${PROD_HEALTH_HINT:-}"
STAGING_HEALTH_HINT="${STAGING_HEALTH_HINT:-}"
CONFIG_FILE="${CLOUDFLARED_CONFIG:-$HOME/.cloudflared/config.yml}"
CLOUDFLARED_BIN="${CLOUDFLARED_BIN:-cloudflared}"
LOCK_DIR="${LOCK_DIR:-/tmp/sub2api-traffic-switch.lock.d}"

usage() {
  cat <<USAGE
Usage:
  $0 --status
  $0 --dry-run-report
  $0 --to prod
  $0 --to staging

Environment:
  HOSTNAME_TARGET      primary hostname to switch (default: sub2api.airflow.eu.org)
  STAGING_HOSTNAME_TARGET staging hostname for report checks (default: sub2api-staging.airflow.eu.org)
  PROD_PORT            prod local port (default: 65530)
  STAGING_PORT         staging local port (default: 65531)
  PROD_HEALTH_HINT     optional expected prod /health marker (e.g. env=prod)
  STAGING_HEALTH_HINT  optional expected staging /health marker (e.g. env=staging)
  CLOUDFLARED_CONFIG   cloudflared config path (default: ~/.cloudflared/config.yml)
  CLOUDFLARED_BIN      cloudflared binary (default: cloudflared)
  LOCK_DIR             lock directory path (default: /tmp/sub2api-traffic-switch.lock.d)
USAGE
}

log() { printf '[switch] %s\n' "$*"; }
err() { printf '[switch][error] %s\n' "$*" >&2; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { err "missing command: $1"; exit 1; }
}

acquire_lock() {
  if ! mkdir "$LOCK_DIR" 2>/dev/null; then
    err "another switch task is running (lock: $LOCK_DIR)"
    exit 1
  fi
  trap 'rmdir "$LOCK_DIR" >/dev/null 2>&1 || true' EXIT
}

current_port() {
  python3 - "$CONFIG_FILE" "$HOSTNAME_TARGET" <<'PY'
from pathlib import Path
import re, sys
cfg = Path(sys.argv[1])
host = sys.argv[2]
lines = cfg.read_text().splitlines()
for i, line in enumerate(lines):
    if f"hostname: {host}" in line:
        for j in range(i+1, min(i+10, len(lines))):
            s = lines[j].strip()
            if s.startswith("service:"):
                m = re.search(r"localhost:(\d+)", s)
                if m:
                    print(m.group(1))
                    raise SystemExit(0)
                raise SystemExit(2)
        break
raise SystemExit(1)
PY
}

set_port_with_backup() {
  local to_port="$1"
  local backup="$CONFIG_FILE.bak.$(date +%Y%m%d-%H%M%S)"

  cp "$CONFIG_FILE" "$backup"
  log "backup saved: $backup"

  python3 - "$CONFIG_FILE" "$HOSTNAME_TARGET" "$to_port" <<'PY'
from pathlib import Path
import sys
cfg = Path(sys.argv[1])
host = sys.argv[2]
port = sys.argv[3]
lines = cfg.read_text().splitlines()
found = False
for i, line in enumerate(lines):
    if f"hostname: {host}" in line:
        found = True
        for j in range(i+1, min(i+10, len(lines))):
            if lines[j].strip().startswith("service:"):
                indent = lines[j][:len(lines[j]) - len(lines[j].lstrip())]
                lines[j] = f"{indent}service: http://localhost:{port}"
                cfg.write_text("\n".join(lines) + "\n")
                raise SystemExit(0)
        raise SystemExit("service line not found under hostname block")
if not found:
    raise SystemExit("hostname block not found")
PY

  if ! "$CLOUDFLARED_BIN" tunnel --config "$CONFIG_FILE" ingress validate >/dev/null; then
    err "ingress validate failed after updating config; restoring backup"
    cp "$backup" "$CONFIG_FILE"
    if ! "$CLOUDFLARED_BIN" tunnel --config "$CONFIG_FILE" ingress validate >/dev/null; then
      err "failed to restore a valid config from backup: $backup"
    fi
    return 1
  fi

  echo "$backup"
}

restore_backup() {
  local backup="$1"
  [[ -f "$backup" ]] || { err "backup not found: $backup"; return 1; }
  cp "$backup" "$CONFIG_FILE"
  "$CLOUDFLARED_BIN" tunnel --config "$CONFIG_FILE" ingress validate >/dev/null
  log "restored config from backup: $backup"
}

reload_cloudflared() {
  local cloudflared_name=""
  cloudflared_name="$(basename "$CLOUDFLARED_BIN")"
  local pattern="$cloudflared_name tunnel --config $CONFIG_FILE run"
  local old_pid=""
  old_pid="$(pgrep -f "$pattern" | head -n1 || true)"

  nohup "$CLOUDFLARED_BIN" tunnel --config "$CONFIG_FILE" run >/tmp/cloudflared-sub2api-switch.log 2>&1 &
  sleep 2
  local newest_pid=""
  newest_pid="$(pgrep -f "$pattern" | tail -n1 || true)"

  if [[ -n "$old_pid" && "$old_pid" != "$newest_pid" ]]; then
    kill "$old_pid" || true
  fi

  sleep 1
  local active_pid=""
  active_pid="$(pgrep -f "$pattern" | tail -n1 || true)"
  if [[ -z "$active_pid" ]]; then
    err "cloudflared not running after reload"
    return 1
  fi
  log "cloudflared active pid: $active_pid"
}

probe_once() {
  local url="$1"
  curl -fsS --max-time 10 "$url" || true
}

check_current_mapping() {
  local expect_port="$1"
  local current=""
  current="$(current_port || true)"
  if [[ "$current" == "$expect_port" ]]; then
    log "mapping OK: ${HOSTNAME_TARGET} -> localhost:${current}"
    return 0
  fi
  err "mapping mismatch: ${HOSTNAME_TARGET} -> localhost:${current:-unknown}, expect localhost:${expect_port}"
  return 1
}

wait_health_ok() {
  local url="$1"
  local label="$2"
  local tries="${3:-8}"
  local sleep_sec="${4:-2}"
  local expected_hint="${5:-}"
  local out=""

  for ((i=1; i<=tries; i++)); do
    out="$(probe_once "$url")"
    if [[ "$out" == *"ok"* ]]; then
      if [[ -z "$expected_hint" || "$out" == *"$expected_hint"* ]]; then
        log "$label OK: $url => $out"
        return 0
      fi
      log "$label hint-miss($i/$tries): expect hint ${expected_hint}, got => ${out:-<no response>}"
    else
      log "$label wait($i/$tries): $url => ${out:-<no response>}"
    fi
    sleep "$sleep_sec"
  done

  if [[ -n "$expected_hint" ]]; then
    err "$label failed: $url => ${out:-<no response>} (expected hint: $expected_hint)"
  else
    err "$label failed: $url => ${out:-<no response>}"
  fi
  return 1
}

health_hint_for_port() {
  local port="$1"
  if [[ "$port" == "$PROD_PORT" ]]; then
    printf %s "$PROD_HEALTH_HINT"
    return
  fi
  if [[ "$port" == "$STAGING_PORT" ]]; then
    printf %s "$STAGING_HEALTH_HINT"
    return
  fi
  printf ''
}

status_check() {
  local current=""
  current="$(current_port || true)"
  log "current route: ${HOSTNAME_TARGET} -> localhost:${current:-unknown}"

  local local_url="http://127.0.0.1:${PROD_PORT}/health"
  if [[ "$current" == "$STAGING_PORT" ]]; then
    local_url="http://127.0.0.1:${STAGING_PORT}/health"
  fi

  log "local  $local_url => $(probe_once "$local_url")"
  log "public https://${HOSTNAME_TARGET}/health => $(probe_once "https://${HOSTNAME_TARGET}/health")"
}

health_state() {
  local url="$1"
  local out=""
  out="$(probe_once "$url")"
  if [[ "$out" == *"ok"* ]]; then
    printf 'OK | %s | %s\n' "$url" "$out"
  else
    printf 'FAIL | %s | %s\n' "$url" "${out:-<no response>}"
  fi
}

dry_run_report() {
  local current=""
  current="$(current_port || true)"

  log "===== DRY RUN REPORT (NO TRAFFIC CHANGES) ====="
  log "generated at: $(date '+%Y-%m-%d %H:%M:%S %z')"
  log "config file: $CONFIG_FILE"
  log "current route: ${HOSTNAME_TARGET} -> localhost:${current:-unknown}"

  local current_target="unknown"
  if [[ "$current" == "$PROD_PORT" ]]; then
    current_target="prod"
  elif [[ "$current" == "$STAGING_PORT" ]]; then
    current_target="staging"
  fi
  log "active target: $current_target"

  log "----- Health snapshot -----"
  log "$(health_state "http://127.0.0.1:${PROD_PORT}/health")"
  log "$(health_state "http://127.0.0.1:${STAGING_PORT}/health")"
  log "$(health_state "https://${HOSTNAME_TARGET}/health")"
  log "$(health_state "https://${STAGING_HOSTNAME_TARGET}/health")"

  log "----- Suggested drill sequence -----"
  if [[ "$current_target" == "prod" ]]; then
    log "1) ./switch_sub2api_traffic.sh --to staging"
    log "2) verify business flow on ${HOSTNAME_TARGET}"
    log "3) ./switch_sub2api_traffic.sh --to prod"
  elif [[ "$current_target" == "staging" ]]; then
    log "1) ./switch_sub2api_traffic.sh --to prod"
    log "2) verify business flow on ${HOSTNAME_TARGET}"
    log "3) ./switch_sub2api_traffic.sh --to staging (optional rollback drill)"
  else
    log "1) ./switch_sub2api_traffic.sh --status"
    log "2) fix unknown route before any switch"
  fi

  log "----- Safety reminder -----"
  log "If prod/staging /health responses are indistinguishable (both only ok),"
  log "you should add an environment marker (e.g. env=prod/staging) to reduce false-positive switch verification risk."
  log "===== END REPORT ====="
}

main() {
  require_cmd python3
  require_cmd "$CLOUDFLARED_BIN"
  require_cmd curl

  [[ -f "$CONFIG_FILE" ]] || { err "config not found: $CONFIG_FILE"; exit 1; }

  local action=""
  local target=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --status) action="status"; shift ;;
      --dry-run-report) action="dryrun"; shift ;;
      --to) action="switch"; target="${2:-}"; shift 2 ;;
      -h|--help) usage; exit 0 ;;
      *) err "unknown arg: $1"; usage; exit 1 ;;
    esac
  done

  [[ -n "$action" ]] || { usage; exit 1; }

  acquire_lock

  if [[ "$action" == "status" ]]; then
    status_check
    exit 0
  fi

  if [[ "$action" == "dryrun" ]]; then
    dry_run_report
    exit 0
  fi

  local to_port=""
  case "$target" in
    prod) to_port="$PROD_PORT" ;;
    staging) to_port="$STAGING_PORT" ;;
    *) err "target must be prod|staging"; exit 1 ;;
  esac

  local from_port=""
  from_port="$(current_port || true)"
  if [[ "$from_port" == "$to_port" ]]; then
    log "already on target route: localhost:$to_port"
    status_check
    exit 0
  fi

  log "switching ${HOSTNAME_TARGET}: localhost:${from_port:-unknown} -> localhost:${to_port}"
  local backup=""
  backup="$(set_port_with_backup "$to_port" | tail -n1)"

  if ! reload_cloudflared; then
    err "reload failed; trying rollback"
    restore_backup "$backup" || true
    reload_cloudflared || true
    exit 1
  fi

  local local_url="http://127.0.0.1:${to_port}/health"
  local public_url="https://${HOSTNAME_TARGET}/health"
  local target_hint=""
  target_hint="$(health_hint_for_port "$to_port")"

  if ! check_current_mapping "$to_port" || ! wait_health_ok "$local_url" "local" 8 2 "$target_hint" || ! wait_health_ok "$public_url" "public" 8 2 "$target_hint"; then
    err "health check failed; rolling back to previous config"
    restore_backup "$backup"
    reload_cloudflared

    if [[ -n "$from_port" ]]; then
      local rollback_hint=""
      rollback_hint="$(health_hint_for_port "$from_port")"
      wait_health_ok "http://127.0.0.1:${from_port}/health" "rollback-local" 8 2 "$rollback_hint" || true
    fi
    if [[ -n "$from_port" ]]; then
      local rollback_public_hint=""
      rollback_public_hint="$(health_hint_for_port "$from_port")"
      wait_health_ok "https://${HOSTNAME_TARGET}/health" "rollback-public" 8 2 "$rollback_public_hint" || true
    else
      wait_health_ok "https://${HOSTNAME_TARGET}/health" "rollback-public" || true
    fi
    exit 1
  fi

  log "switch complete"
}

main "$@"
