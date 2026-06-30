#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
API_BASE="${API_BASE:-${BASE_URL%/}/api/v1}"
PGDOCKER_CONTAINER="${PGDOCKER_CONTAINER:-sub2api-postgres}"
REQUIRE_PARITY="${REQUIRE_PROMPT_CATALOG_PARITY:-0}"

failures=0

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

warn_or_fail() {
  local message="$1"
  failures=$((failures + 1))
  if [[ "$REQUIRE_PARITY" == "1" ]]; then
    echo "ERROR: ${message}" >&2
  else
    echo "WARN: ${message}" >&2
  fi
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
  echo "psql is required, or set PGDOCKER_CONTAINER to a running PostgreSQL container." >&2
  exit 2
}

count_query() {
  psql_query -Atc "$1" | tr -d '[:space:]'
}

section "Prompt Catalog compatibility scope"
cat <<'NOTE'
Checks encoded from the current Prompt Catalog compatibility contract:
- default case ordering is imported_at DESC NULLS LAST, then title/id;
- visible case/template facets keep source/category coverage and display labels;
- displayed primary images are owned static URLs or local paths, not raw remote URLs;
- X/Twitter imported rows keep imported_at timestamps for stable newest-first ordering.
NOTE

section "Default ordering"
order_violations="$(count_query "
WITH ordered AS (
  SELECT
    id,
    title,
    imported_at,
    LAG(imported_at) OVER (ORDER BY imported_at DESC NULLS LAST, title ASC, id ASC) AS previous_imported_at
  FROM prompt_catalog_items
  WHERE status = 'published'
    AND deleted_at IS NULL
    AND source_type = 'case'
)
SELECT COUNT(*)
FROM ordered
WHERE previous_imported_at IS NOT NULL
  AND imported_at IS NOT NULL
  AND previous_imported_at < imported_at;
")"
echo "imported_at_order_violations=${order_violations}"
if [[ "$order_violations" != "0" ]]; then
  warn_or_fail "published cases are not stable under imported_at DESC ordering"
fi

twitter_missing_imported_at="$(count_query "
SELECT COUNT(*)
FROM prompt_catalog_items
WHERE status = 'published'
  AND deleted_at IS NULL
  AND source_type = 'case'
  AND (
    source_url ILIKE '%x.com/%/status/%'
    OR source_url ILIKE '%twitter.com/%/status/%'
    OR id LIKE 'tw-%'
  )
  AND imported_at IS NULL;
")"
echo "twitter_imports_missing_imported_at=${twitter_missing_imported_at}"
if [[ "$twitter_missing_imported_at" != "0" ]]; then
  warn_or_fail "X/Twitter prompt rows are missing imported_at, which weakens newest-first ordering"
fi

section "Facet coverage"
source_project_count="$(count_query "
SELECT COUNT(DISTINCT NULLIF(source_project, ''))
FROM prompt_catalog_items
WHERE status = 'published' AND deleted_at IS NULL;
")"
category_count="$(count_query "
SELECT COUNT(DISTINCT NULLIF(category, ''))
FROM prompt_catalog_items
WHERE status = 'published' AND deleted_at IS NULL;
")"
missing_source_project="$(count_query "
SELECT COUNT(*)
FROM prompt_catalog_items
WHERE status = 'published'
  AND deleted_at IS NULL
  AND NULLIF(TRIM(COALESCE(source_project, '')), '') IS NULL;
")"
missing_category="$(count_query "
SELECT COUNT(*)
FROM prompt_catalog_items
WHERE status = 'published'
  AND deleted_at IS NULL
  AND NULLIF(TRIM(COALESCE(category, '')), '') IS NULL;
")"
echo "source_project_count=${source_project_count}"
echo "category_count=${category_count}"
echo "missing_source_project=${missing_source_project}"
echo "missing_category=${missing_category}"
if [[ "$source_project_count" -lt 3 ]]; then
  warn_or_fail "Prompt Catalog source facets look incomplete for the multi-source gallery"
fi
if [[ "$category_count" -lt 10 ]]; then
  warn_or_fail "Prompt Catalog category facets look incomplete for the gallery"
fi
if [[ "$missing_category" != "0" ]]; then
  warn_or_fail "published prompt rows are missing category values"
fi

section "Displayed image source policy"
remote_primary_images="$(count_query "
WITH visible AS (
  SELECT *
  FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL
), primary_images AS (
  SELECT
    id,
    COALESCE(
      NULLIF(image_url, ''),
      NULLIF(image_preview_url, ''),
      NULLIF(image_thumb_url, ''),
      NULLIF(image_original_url, ''),
      image_urls->>0
    ) AS primary_image_url
  FROM visible
)
SELECT COUNT(*)
FROM primary_images
WHERE NULLIF(TRIM(COALESCE(primary_image_url, '')), '') IS NOT NULL
  AND NOT (
    primary_image_url LIKE 'https://static.cloudbase.eu.org/%'
    OR primary_image_url LIKE '/%'
  );
")"
echo "remote_primary_images=${remote_primary_images}"
if [[ "$remote_primary_images" != "0" ]]; then
  warn_or_fail "displayed primary prompt images include raw remote URLs; expected local or owned static URLs"
fi

section "Duplicate source URL explanation"
duplicate_source_url_groups="$(count_query "
SELECT COUNT(*)
FROM (
  SELECT source_url
  FROM prompt_catalog_items
  WHERE deleted_at IS NULL
    AND NULLIF(TRIM(COALESCE(source_url, '')), '') IS NOT NULL
  GROUP BY source_url
  HAVING COUNT(*) > 1
) duplicates;
")"
duplicate_source_url_rows="$(count_query "
SELECT COALESCE(SUM(count), 0)
FROM (
  SELECT COUNT(*) AS count
  FROM prompt_catalog_items
  WHERE deleted_at IS NULL
    AND NULLIF(TRIM(COALESCE(source_url, '')), '') IS NOT NULL
  GROUP BY source_url
  HAVING COUNT(*) > 1
) duplicates;
")"
echo "duplicate_source_url_groups=${duplicate_source_url_groups}"
echo "duplicate_source_url_rows=${duplicate_source_url_rows}"
if [[ "$duplicate_source_url_groups" != "0" ]]; then
  echo "duplicate_source_url_note=duplicate source URLs are allowed only when they represent multi-image/thread source cases; review top groups in tools/prompt-catalog-integrity.sh output"
fi

section "Public API parity smoke"
if command -v curl >/dev/null 2>&1 && command -v node >/dev/null 2>&1; then
  tmp_json="$(mktemp -t prompt-catalog-parity.XXXXXX.json)"
  if curl -fsS "${API_BASE%/}/prompts/cases?page=1&page_size=12&source_type=case&has_image=true&sort_by=imported_at&sort_order=desc" >"$tmp_json"; then
    PROMPT_CATALOG_PARITY_JSON="$tmp_json" node <<'NODE'
const fs = require('node:fs')
const path = process.env.PROMPT_CATALOG_PARITY_JSON
const payload = JSON.parse(fs.readFileSync(path, 'utf8'))
const data = payload.data || payload
const items = Array.isArray(data.items) ? data.items : []
const sourceFacets = Array.isArray(data.summary?.sources) ? data.summary.sources : []
const categoryFacets = Array.isArray(data.summary?.categories) ? data.summary.categories : []
if (items.length === 0) {
  throw new Error('API returned no prompt cases')
}
if (sourceFacets.length === 0) {
  throw new Error('API summary returned no source facets')
}
if (categoryFacets.length === 0) {
  throw new Error('API summary returned no category facets')
}
for (const item of items) {
  const image = item.primary_image_url || item.image_url || item.image_preview_url || item.image_thumb_url || ''
  if (image && !(String(image).startsWith('https://static.cloudbase.eu.org/') || String(image).startsWith('/'))) {
    throw new Error(`API item ${item.id || ''} exposes non-owned primary image ${image}`)
  }
  if (!item.source_display_label) {
    throw new Error(`API item ${item.id || ''} is missing source_display_label`)
  }
}
for (const facet of [...sourceFacets, ...categoryFacets]) {
  if (facet?.value && !facet.display_label) {
    throw new Error(`facet ${facet.value} is missing display_label`)
  }
}
console.log(`api_items=${items.length}`)
console.log(`api_source_facets=${sourceFacets.length}`)
console.log(`api_category_facets=${categoryFacets.length}`)
NODE
    rm -f "$tmp_json"
    echo "api_parity_smoke=ok"
  else
    rm -f "$tmp_json"
    warn_or_fail "Prompt Catalog API smoke failed; service may not be running"
  fi
else
  warn_or_fail "curl and node are required for Prompt Catalog API parity smoke"
fi

section "Prompt Catalog parity result"
if [[ "$failures" -gt 0 ]]; then
  echo "prompt_catalog_parity=false"
  echo "missing_or_review_items=${failures}"
  if [[ "$REQUIRE_PARITY" == "1" ]]; then
    exit 2
  fi
else
  echo "prompt_catalog_parity=true"
fi

echo
echo "Prompt Catalog compatibility audit complete."
