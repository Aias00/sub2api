#!/usr/bin/env bash
set -euo pipefail

APPLY="${PROMPT_CATALOG_RETIRE_EXTERNAL_IMAGES_APPLY:-0}"
ALLOWED_HOSTS_CSV="${PROMPT_CATALOG_ALLOWED_IMAGE_HOSTS:-static.cloudbase.eu.org}"
TARGET_HOSTS_CSV="${PROMPT_CATALOG_RETIRE_IMAGE_HOSTS:-}"
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
  echo "psql is required, or set PGDOCKER_CONTAINER to a running PostgreSQL container." >&2
  exit 2
}

csv_to_sql_array() {
  local csv="$1"
  if [[ -z "$csv" ]]; then
    printf 'ARRAY[]::text[]'
    return
  fi
  local out=""
  local -a parts=()
  IFS=',' read -r -a parts <<< "$csv"
  for part in "${parts[@]}"; do
    part="$(printf '%s' "$part" | xargs)"
    [[ -z "$part" ]] && continue
    part="${part//\'/\'\'}"
    if [[ -n "$out" ]]; then
      out+=","
    fi
    out+="'$part'"
  done
  printf 'ARRAY[%s]::text[]' "$out"
}

allowed_hosts_sql="$(csv_to_sql_array "$ALLOWED_HOSTS_CSV")"
target_hosts_sql="$(csv_to_sql_array "$TARGET_HOSTS_CSV")"

candidate_where="
WITH image_rows AS (
  SELECT
    id,
    title,
    source_project,
    source_url,
    image_url,
    image_urls,
    regexp_replace(image_url, '^https?://([^/?#]+).*$', '\\1') AS image_host
  FROM prompt_catalog_items
  WHERE status = 'published'
    AND deleted_at IS NULL
    AND image_url LIKE 'http%'
), candidates AS (
  SELECT *
  FROM image_rows
  WHERE NOT (image_host = ANY(${allowed_hosts_sql}))
    AND (
      cardinality(${target_hosts_sql}) = 0
      OR image_host = ANY(${target_hosts_sql})
    )
)
"

echo "== Prompt Catalog external image candidates =="
psql_query -P null='NULL' -c "
${candidate_where}
SELECT id, title, source_project, image_host, source_url, image_url
FROM candidates
ORDER BY image_host, id;
"

echo
echo "== Candidate summary =="
psql_query -P null='NULL' -c "
${candidate_where}
SELECT image_host, COUNT(*) AS count
FROM candidates
GROUP BY image_host
ORDER BY count DESC, image_host;
"

if [[ "$APPLY" != "1" ]]; then
  echo
  echo "Dry run only. Set PROMPT_CATALOG_RETIRE_EXTERNAL_IMAGES_APPLY=1 to retire these external image URLs."
  exit 0
fi

echo
echo "== Applying retirement =="
psql_query -P null='NULL' -c "
${candidate_where},
updated AS (
  UPDATE prompt_catalog_items p
  SET
    image_url = NULL,
    image_original_url = NULL,
    image_preview_url = NULL,
    image_thumb_url = NULL,
    image_urls = '[]'::jsonb,
    raw_json = (
      (COALESCE(p.raw_json, '{}'::jsonb) - 'imageUrl' - 'imageUrls')
      || jsonb_build_object(
        'retiredImageUrls', jsonb_build_array(c.image_url),
        'imageRetiredAt', to_jsonb(now()),
        'imageRetirementReason', 'external image URL was not reachable during object-storage cutover validation',
        'imageRetiredHost', c.image_host
      )
    ),
    updated_at = now()
  FROM candidates c
  WHERE p.id = c.id
  RETURNING p.id
)
SELECT COUNT(*) AS retired_prompt_images FROM updated;
"

echo
echo "Prompt Catalog external image retirement complete."
