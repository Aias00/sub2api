#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
API_BASE="${API_BASE:-${BASE_URL%/}/api/v1}"
RUN_URL_SAMPLE="${RUN_OBJECT_STORAGE_URL_SAMPLE:-0}"
REQUIRE_URL_SAMPLE="${REQUIRE_OBJECT_STORAGE_URL_SAMPLE:-0}"
REQUIRE_DB="${REQUIRE_OBJECT_STORAGE_DB:-0}"
REQUIRE_NO_MOCK_URLS="${REQUIRE_OBJECT_STORAGE_NO_MOCK_URLS:-${REQUIRE_URL_SAMPLE}}"
REQUIRE_ALLOWED_PROMPT_HOSTS="${REQUIRE_OBJECT_STORAGE_ALLOWED_PROMPT_HOSTS:-${REQUIRE_URL_SAMPLE}}"
ALLOWED_PROMPT_IMAGE_HOSTS="${PROMPT_CATALOG_ALLOWED_IMAGE_HOSTS:-static.cloudbase.eu.org}"
URL_SAMPLE_LIMIT="${OBJECT_STORAGE_URL_SAMPLE_LIMIT:-10}"
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

go_object_storage_check() {
  local args=()
  if [[ "$RUN_URL_SAMPLE" == "1" ]]; then
    args+=("-sample" "-limit" "$URL_SAMPLE_LIMIT")
  fi
  if [[ ${#args[@]} -gt 0 ]]; then
    (cd backend && go run ./cmd/objectstoragecheck "${args[@]}")
  else
    (cd backend && go run ./cmd/objectstoragecheck)
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
  return 127
}

bool_env() {
  case "${1:-}" in
    1|true|TRUE|yes|YES|on|ON) return 0 ;;
    *) return 1 ;;
  esac
}

configured() {
  [[ -n "${!1:-}" ]] && echo "yes" || echo "no"
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

section "Image Workspace object-storage env"
echo "IMAGE_WORKSPACE_OBJECT_STORAGE_ENABLED=${IMAGE_WORKSPACE_OBJECT_STORAGE_ENABLED:-false}"
echo "IMAGE_WORKSPACE_OBJECT_STORAGE_PROVIDER=${IMAGE_WORKSPACE_OBJECT_STORAGE_PROVIDER:-r2}"
echo "endpoint_configured=$(configured IMAGE_WORKSPACE_OBJECT_STORAGE_ENDPOINT)"
echo "r2_account_id_configured=$(configured IMAGE_WORKSPACE_R2_ACCOUNT_ID)"
echo "access_key_configured=$([[ -n "${IMAGE_WORKSPACE_OBJECT_STORAGE_ACCESS_KEY_ID:-${IMAGE_WORKSPACE_R2_ACCESS_KEY_ID:-${IMAGE_WORKSPACE_R2_ACCESS_KEY:-}}}" ]] && echo yes || echo no)"
echo "secret_key_configured=$([[ -n "${IMAGE_WORKSPACE_OBJECT_STORAGE_SECRET_ACCESS_KEY:-${IMAGE_WORKSPACE_R2_SECRET_ACCESS_KEY:-${IMAGE_WORKSPACE_R2_SECRET_KEY:-}}}" ]] && echo yes || echo no)"
echo "bucket_configured=$([[ -n "${IMAGE_WORKSPACE_OBJECT_STORAGE_BUCKET:-${IMAGE_WORKSPACE_R2_BUCKET:-${IMAGE_WORKSPACE_R2_BUCKET_NAME:-}}}" ]] && echo yes || echo no)"
echo "public_base_url_configured=$([[ -n "${IMAGE_WORKSPACE_OBJECT_STORAGE_PUBLIC_BASE_URL:-${IMAGE_WORKSPACE_R2_PUBLIC_BASE_URL:-${IMAGE_WORKSPACE_R2_DOMAIN:-}}}" ]] && echo yes || echo no)"

section "Image Workspace worker storage check"
if [[ -f tools/image-workspace-worker/package.json ]]; then
  npm --prefix tools/image-workspace-worker run storage-check
else
  echo "tools/image-workspace-worker is missing; storage check skipped." >&2
fi

section "Target database object URL inventory"
if ! command -v psql >/dev/null 2>&1 && ! { command -v docker >/dev/null 2>&1 && docker inspect "$PGDOCKER_CONTAINER" >/dev/null 2>&1; }; then
  echo "psql not found and PostgreSQL container fallback unavailable; trying Go object-storage checker." >&2
  if go_object_storage_check; then
    echo "Go object-storage checker ... ok"
  elif [[ "$REQUIRE_DB" == "1" ]]; then
    exit 2
  else
    echo "Go object-storage checker failed; target database object URL inventory skipped." >&2
  fi
else
  psql_query -P null='NULL' -c "
SELECT
  to_regclass('public.prompt_catalog_items') IS NOT NULL AS has_prompt_catalog_items,
  to_regclass('public.image_workspace_artifacts') IS NOT NULL AS has_image_workspace_artifacts,
  to_regclass('public.hot_item_media') IS NOT NULL AS has_hot_item_media;
"
  psql_query -P null='NULL' -c "
WITH prompt_visible AS (
  SELECT *
  FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL
), prompt_flags AS (
  SELECT
    id,
    NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_preview_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_thumb_url, '')), '') IS NOT NULL
      OR jsonb_array_length(COALESCE(image_urls, '[]'::jsonb)) > 0 AS has_image
  FROM prompt_visible
)
SELECT
  (SELECT COUNT(*) FROM prompt_visible) AS prompt_visible,
  (SELECT COUNT(*) FROM prompt_flags WHERE has_image) AS prompt_with_image,
  (SELECT COUNT(*) FROM image_workspace_artifacts WHERE NULLIF(TRIM(image_url), '') IS NOT NULL) AS image_artifact_urls,
  (SELECT COUNT(*) FROM image_workspace_artifacts WHERE NULLIF(TRIM(storage_key), '') IS NOT NULL) AS image_artifact_storage_keys,
  (SELECT COUNT(*) FROM hot_item_media) AS hot_item_media_rows;
"

  mock_artifact_count="$(psql_query -Atc "
SELECT COUNT(*)
FROM image_workspace_artifacts
WHERE image_url LIKE '%assets.example.test%'
   OR storage_key LIKE 'image-workspace-e2e/%';
" | tr -d '[:space:]')"
  echo "image_workspace_mock_artifacts=${mock_artifact_count}"
  if [[ "${mock_artifact_count:-0}" != "0" ]]; then
    echo "WARN: Image Workspace contains mock/test artifact URLs or storage keys." >&2
    if [[ "$REQUIRE_NO_MOCK_URLS" == "1" ]]; then
      echo "Mock artifact rows are not allowed when REQUIRE_OBJECT_STORAGE_NO_MOCK_URLS=1." >&2
      exit 2
    fi
  fi

  allowed_prompt_hosts_sql="$(csv_to_sql_array "$ALLOWED_PROMPT_IMAGE_HOSTS")"
  disallowed_prompt_hosts="$(psql_query -Atc "
WITH urls AS (
  SELECT image_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
  UNION ALL
  SELECT image_original_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
  UNION ALL
  SELECT image_preview_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_preview_url, '')), '') IS NOT NULL
  UNION ALL
  SELECT image_thumb_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_thumb_url, '')), '') IS NOT NULL
), hosts AS (
  SELECT regexp_replace(url, '^https?://([^/?#]+).*$', '\\1') AS host
  FROM urls
  WHERE url LIKE 'http://%' OR url LIKE 'https://%'
), disallowed AS (
  SELECT host, COUNT(*) AS count
  FROM hosts
  WHERE NOT (host = ANY(${allowed_prompt_hosts_sql}))
  GROUP BY host
)
SELECT COALESCE(SUM(count), 0)::text FROM disallowed;
" | tr -d '[:space:]')"
  echo "prompt_disallowed_image_hosts=${disallowed_prompt_hosts}"
  if [[ "${disallowed_prompt_hosts:-0}" != "0" ]]; then
    psql_query -P null='NULL' -c "
WITH urls AS (
  SELECT image_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
  UNION ALL
  SELECT image_original_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
  UNION ALL
  SELECT image_preview_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_preview_url, '')), '') IS NOT NULL
  UNION ALL
  SELECT image_thumb_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_thumb_url, '')), '') IS NOT NULL
), hosts AS (
  SELECT regexp_replace(url, '^https?://([^/?#]+).*$', '\\1') AS host
  FROM urls
  WHERE url LIKE 'http://%' OR url LIKE 'https://%'
)
SELECT host, COUNT(*) AS count
FROM hosts
WHERE NOT (host = ANY(${allowed_prompt_hosts_sql}))
GROUP BY host
ORDER BY count DESC, host;
"
    if [[ "$REQUIRE_ALLOWED_PROMPT_HOSTS" == "1" ]]; then
      echo "Disallowed Prompt Catalog image hosts are not allowed when REQUIRE_OBJECT_STORAGE_ALLOWED_PROMPT_HOSTS=1." >&2
      exit 2
    fi
  fi

  if [[ "$RUN_URL_SAMPLE" == "1" ]]; then
    section "Target URL reachability sample"
    tmp_urls="$(mktemp)"
    psql_query -Atc "
WITH urls AS (
  SELECT image_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
  UNION
  SELECT image_original_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
  UNION
  SELECT image_url AS url FROM image_workspace_artifacts
  WHERE NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
  UNION
  SELECT original_url AS url FROM hot_item_media
  WHERE NULLIF(TRIM(COALESCE(original_url, '')), '') IS NOT NULL
)
SELECT url FROM urls
WHERE url LIKE 'http://%' OR url LIKE 'https://%'
ORDER BY url
LIMIT ${URL_SAMPLE_LIMIT};
" > "$tmp_urls"
    if [[ -s "$tmp_urls" ]]; then
      sample_failures=0
      while IFS= read -r url; do
        curl -fsSI --max-time 10 "$url" >/dev/null \
          && echo "HEAD $url ... ok" \
          || {
            sample_failures=$((sample_failures + 1))
            echo "HEAD $url ... failed" >&2
          }
      done < "$tmp_urls"
      if [[ "$sample_failures" -gt 0 ]]; then
        echo "URL sample failures=${sample_failures}" >&2
        if [[ "$REQUIRE_URL_SAMPLE" == "1" ]]; then
          rm -f "$tmp_urls"
          exit 2
        fi
      fi
    else
      echo "No public HTTP(S) URLs found for sampling."
    fi
    rm -f "$tmp_urls"
  else
    echo "URL reachability sample skipped. Set RUN_OBJECT_STORAGE_URL_SAMPLE=1 to enable it."
  fi
fi

section "Manual production acceptance"
cat <<'CHECKLIST'
Required before claiming object-storage migration complete:
1. `make validate-image-workspace-object-storage` passes to prove worker PUT upload behavior against a local S3/R2-compatible mock.
2. Image Workspace worker has object storage enabled with endpoint/account, bucket, access key, secret, and public base URL.
3. `npm --prefix tools/image-workspace-worker run storage-check` passes with production-like R2 variables.
4. A real image generation task writes an artifact to object storage and the public URL loads.
5. Prompt Catalog image URLs sampled from PostgreSQL load from the configured public/static domain.
6. If Hot item media is enabled later, sampled hot_item_media URLs load from the configured public/static domain.
CHECKLIST

echo
echo "Object storage integrity check complete."
