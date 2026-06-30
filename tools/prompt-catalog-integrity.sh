#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
API_BASE="${API_BASE:-${BASE_URL%/}/api/v1}"
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
  echo "psql is required for database integrity checks, or set PGDOCKER_CONTAINER to a running PostgreSQL container." >&2
  exit 2
}

echo "== Prompt Catalog table =="
psql_query -Atc "SELECT to_regclass('public.prompt_catalog_items');"

echo
echo "== Row counts =="
psql_query -P null='NULL' -c "
SELECT
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE status = 'published' AND deleted_at IS NULL) AS published_visible,
  COUNT(*) FILTER (WHERE status = 'draft' AND deleted_at IS NULL) AS draft_visible,
  COUNT(*) FILTER (WHERE deleted_at IS NOT NULL) AS deleted,
  COUNT(*) FILTER (WHERE status = 'published' AND deleted_at IS NULL AND source_type = 'case') AS published_cases,
  COUNT(*) FILTER (WHERE status = 'published' AND deleted_at IS NULL AND source_type = 'template') AS published_templates
FROM prompt_catalog_items;
"

echo
echo "== Required-field gaps =="
psql_query -P null='NULL' -c "
SELECT
  COUNT(*) FILTER (WHERE NULLIF(TRIM(id), '') IS NULL) AS missing_id,
  COUNT(*) FILTER (WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(title), '') IS NULL) AS published_missing_title,
  COUNT(*) FILTER (WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(prompt), '') IS NULL) AS published_missing_prompt,
  COUNT(*) FILTER (WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(source_type), '') IS NULL) AS published_missing_source_type,
  COUNT(*) FILTER (WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(category), '') IS NULL) AS published_missing_category
FROM prompt_catalog_items;
"

echo
echo "== Image coverage =="
psql_query -P null='NULL' -c "
WITH visible AS (
  SELECT *
  FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL
), image_flags AS (
  SELECT
    id,
    NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_preview_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_thumb_url, '')), '') IS NOT NULL
      OR jsonb_array_length(COALESCE(image_urls, '[]'::jsonb)) > 0 AS has_image
  FROM visible
)
SELECT
  COUNT(*) AS visible_total,
  COUNT(*) FILTER (WHERE has_image) AS with_image,
  COUNT(*) FILTER (WHERE NOT has_image) AS without_image
FROM image_flags;
"

echo
echo "== Top sources =="
psql_query -P null='NULL' -c "
SELECT
  COALESCE(NULLIF(source_project, ''), 'manual') AS source_project,
  COALESCE(NULLIF(source_label, ''), '-') AS source_label,
  source_type,
  COUNT(*) AS count
FROM prompt_catalog_items
WHERE status = 'published' AND deleted_at IS NULL
GROUP BY source_project, source_label, source_type
ORDER BY count DESC, source_project ASC
LIMIT 20;
"

echo
echo "== Top categories =="
psql_query -P null='NULL' -c "
SELECT COALESCE(NULLIF(category, ''), 'general') AS category, COUNT(*) AS count
FROM prompt_catalog_items
WHERE status = 'published' AND deleted_at IS NULL
GROUP BY category
ORDER BY count DESC, category ASC
LIMIT 20;
"

echo
echo "== Potential duplicates =="
psql_query -P null='NULL' -c "
SELECT source_url, COUNT(*) AS count
FROM prompt_catalog_items
WHERE deleted_at IS NULL
  AND NULLIF(TRIM(COALESCE(source_url, '')), '') IS NOT NULL
GROUP BY source_url
HAVING COUNT(*) > 1
ORDER BY count DESC, source_url ASC
LIMIT 20;
"

echo
echo "== Broken sample candidates =="
psql_query -P null='NULL' -c "
SELECT id, title, source_project, source_type, category
FROM prompt_catalog_items
WHERE status = 'published'
  AND deleted_at IS NULL
  AND (
    NULLIF(TRIM(title), '') IS NULL
    OR NULLIF(TRIM(prompt), '') IS NULL
    OR NULLIF(TRIM(source_type), '') IS NULL
  )
ORDER BY updated_at DESC NULLS LAST
LIMIT 20;
"

echo
echo "== Public API smoke =="
if command -v curl >/dev/null 2>&1; then
  curl -fsS "${API_BASE%/}/prompts/cases?page=1&page_size=1&source_type=case&has_image=true" >/tmp/prompt-catalog-api-smoke.json \
    && echo "GET ${API_BASE%/}/prompts/cases ... ok" \
    || echo "GET ${API_BASE%/}/prompts/cases ... skipped/failed; service may not be running" >&2
else
  echo "curl not found; API smoke skipped"
fi

echo
echo "Prompt Catalog integrity check complete."
