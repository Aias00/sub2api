#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
API_BASE="${API_BASE:-${BASE_URL%/}/api/v1}"
REQUIRE_DB="${REQUIRE_HOT_DB:-0}"
RUN_API_SMOKE="${RUN_HOT_API_SMOKE:-1}"
PGDOCKER_CONTAINER="${PGDOCKER_CONTAINER:-sub2api-postgres}"

PSQL_ARGS=()
if [[ -n "${DATABASE_URL:-}" ]]; then
  PSQL_ARGS+=("${DATABASE_URL}")
else
  PSQL_ARGS+=(
    "host=${PGHOST:-127.0.0.1}"
    "port=${PGPORT:-5432}"
    "user=${PGUSER:-sub2api}"
    "dbname=${PGDATABASE:-sub2api}"
    "sslmode=${PGSSLMODE:-disable}"
  )
fi

section() {
  echo
  echo "== $* =="
}

psql_query() {
  if command -v psql >/dev/null 2>&1; then
    psql "${PSQL_ARGS[@]}" -v ON_ERROR_STOP=1 "$@"
    return
  fi
  if command -v docker >/dev/null 2>&1 && docker inspect "$PGDOCKER_CONTAINER" >/dev/null 2>&1; then
    docker exec -i "$PGDOCKER_CONTAINER" psql \
      -U "${PGUSER:-sub2api}" \
      -d "${PGDATABASE:-sub2api}" \
      -v ON_ERROR_STOP=1 \
      "$@"
    return
  fi
  return 127
}

go_hot_check() {
  (cd backend && go run ./cmd/hotimport -check-only)
}

section "Target PostgreSQL hot tables"
if ! command -v psql >/dev/null 2>&1 && ! { command -v docker >/dev/null 2>&1 && docker inspect "$PGDOCKER_CONTAINER" >/dev/null 2>&1; }; then
  echo "psql not found and PostgreSQL container fallback unavailable; trying Go PostgreSQL checker." >&2
  if go_hot_check; then
    echo "Go PostgreSQL checker ... ok"
  elif [[ "$REQUIRE_DB" == "1" ]]; then
    exit 2
  else
    echo "Go PostgreSQL checker failed; PostgreSQL checks skipped." >&2
  fi
else
  psql_query -Atc "
SELECT
  to_regclass('public.hot_sources') IS NOT NULL AS has_hot_sources,
  to_regclass('public.hot_runs') IS NOT NULL AS has_hot_runs,
  to_regclass('public.hot_checkpoints') IS NOT NULL AS has_hot_checkpoints,
  to_regclass('public.hot_items') IS NOT NULL AS has_hot_items,
  to_regclass('public.hot_item_media') IS NOT NULL AS has_hot_item_media,
  to_regclass('public.hot_media_assets') IS NOT NULL AS has_hot_media_assets,
  to_regclass('public.hot_run_events') IS NOT NULL AS has_hot_run_events,
  to_regclass('public.hot_feed_items') IS NOT NULL AS has_hot_feed_items,
  to_regclass('public.hot_daily_issues') IS NOT NULL AS has_hot_daily_issues,
  to_regclass('public.hot_daily_sections') IS NOT NULL AS has_hot_daily_sections,
  to_regclass('public.hot_daily_stories') IS NOT NULL AS has_hot_daily_stories,
  to_regclass('public.hot_mp_entries') IS NOT NULL AS has_hot_mp_entries;
"

  echo
  echo "== Target row counts =="
  psql_query -P null='NULL' -c "
SELECT
  (SELECT COUNT(*) FROM hot_sources) AS hot_sources,
  (SELECT COUNT(*) FROM hot_runs) AS hot_runs,
  (SELECT COUNT(*) FROM hot_checkpoints) AS hot_checkpoints,
  (SELECT COUNT(*) FROM hot_items) AS hot_items,
  (SELECT COUNT(*) FROM hot_item_media) AS hot_item_media,
  (SELECT COUNT(*) FROM hot_media_assets) AS hot_media_assets,
  (SELECT COUNT(*) FROM hot_run_events) AS hot_run_events,
  (SELECT COUNT(*) FROM hot_feed_items) AS hot_feed_items,
  (SELECT COUNT(*) FROM hot_daily_issues) AS hot_daily_issues,
  (SELECT COUNT(*) FROM hot_daily_sections) AS hot_daily_sections,
  (SELECT COUNT(*) FROM hot_daily_stories) AS hot_daily_stories,
  (SELECT COUNT(*) FROM hot_mp_entries) AS hot_mp_entries;
"

  echo
  echo "== Legacy table retention =="
  psql_query -P null='NULL' -c "
SELECT
  (SELECT COUNT(*) FROM hot_item_media) AS hot_item_media_rows,
  (SELECT COUNT(*) FROM hot_media_assets) AS hot_media_assets_rows,
  (SELECT COUNT(*) FROM hot_feed_items) AS hot_feed_items_rows,
  (SELECT COUNT(*) FROM hot_daily_issues) AS hot_daily_issues_rows,
  (SELECT COUNT(*) FROM hot_daily_sections) AS hot_daily_sections_rows,
  (SELECT COUNT(*) FROM hot_daily_stories) AS hot_daily_stories_rows,
  (SELECT COUNT(*) FROM hot_mp_entries) AS hot_mp_entries_rows;
"

  echo
  echo "== Target data quality =="
  psql_query -P null='NULL' -c "
SELECT
  COUNT(*) FILTER (WHERE NULLIF(TRIM(source_id), '') IS NULL) AS missing_source_id,
  COUNT(*) FILTER (WHERE NULLIF(TRIM(external_id), '') IS NULL) AS missing_external_id,
  COUNT(*) FILTER (WHERE NULLIF(TRIM(title), '') IS NULL) AS missing_title,
  COUNT(*) FILTER (WHERE status = 'published') AS published,
  COUNT(*) FILTER (WHERE status <> 'published') AS non_published,
  COUNT(DISTINCT source_id) AS source_count
FROM hot_items;
"

  echo
  echo "== Top target sources =="
  psql_query -P null='NULL' -c "
SELECT source_id, COUNT(*) AS count
FROM hot_items
GROUP BY source_id
ORDER BY count DESC, source_id ASC
LIMIT 20;
"
fi

section "Hot surface mapping"
cat <<'NOTE'
The live Sub2API Hot collector reads enabled RSS sources from hot_sources and
writes runs, checkpoints, run events, and hot_items directly into PostgreSQL.
Legacy presentation tables are preserved as empty table structures only:
hot_media_assets, hot_feed_items, hot_daily_issues, hot_daily_sections,
hot_daily_stories, and hot_mp_entries. They are not part of the live product
surface or startup readiness.
NOTE

section "Public API smoke"
if [[ "$RUN_API_SMOKE" != "1" ]]; then
  echo "API smoke skipped. Set RUN_HOT_API_SMOKE=1 to enable it."
elif ! command -v curl >/dev/null 2>&1; then
  echo "curl not found; API smoke skipped."
else
  curl -fsS "${API_BASE%/}/hot/sources" >/tmp/hot-sources-smoke.json \
    && echo "GET ${API_BASE%/}/hot/sources ... ok" \
    || echo "GET ${API_BASE%/}/hot/sources ... skipped/failed; service may not be running" >&2
  curl -fsS "${API_BASE%/}/hot/items?page=1&page_size=1" >/tmp/hot-items-smoke.json \
    && echo "GET ${API_BASE%/}/hot/items ... ok" \
    || echo "GET ${API_BASE%/}/hot/items ... skipped/failed; service may not be running" >&2
  run_id="$(psql_query -Atc "SELECT run_id FROM hot_run_events ORDER BY created_at ASC, id ASC LIMIT 1;" 2>/dev/null | tr -d '[:space:]' || true)"
  if [[ -n "$run_id" ]]; then
    curl -fsS "${API_BASE%/}/hot/run-events?run_id=${run_id}&page=1&page_size=1" >/tmp/hot-run-events-smoke.json \
      && echo "GET ${API_BASE%/}/hot/run-events ... ok" \
      || echo "GET ${API_BASE%/}/hot/run-events ... skipped/failed; service may not be running" >&2
  fi
fi

echo
echo "Hot Content integrity check complete."
