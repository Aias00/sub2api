#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
API_BASE="${API_BASE:-${BASE_URL%/}/api/v1}"
PGDOCKER_CONTAINER="${PGDOCKER_CONTAINER:-sub2api-postgres}"
STRICT="${REQUIRE_PROMPT_CATALOG_PRODUCTION_READY:-0}"
RUN_URL_SAMPLE="${RUN_PROMPT_CATALOG_URL_SAMPLE:-0}"
URL_SAMPLE_LIMIT="${PROMPT_CATALOG_URL_SAMPLE_LIMIT:-10}"
URL_CHECK_CONCURRENCY="${PROMPT_CATALOG_URL_CHECK_CONCURRENCY:-8}"
URL_CHECK_TIMEOUT="${PROMPT_CATALOG_URL_CHECK_TIMEOUT_SECONDS:-10}"
URL_CHECK_RETRIES="${PROMPT_CATALOG_URL_CHECK_RETRIES:-2}"
URL_CHECK_VERBOSE="${PROMPT_CATALOG_URL_CHECK_VERBOSE:-0}"
URL_REPORT_PATH="${PROMPT_CATALOG_URL_REPORT_PATH:-}"
ACCEPTANCE_REPORT_PATH="${PROMPT_CATALOG_ACCEPTANCE_REPORT_PATH:-}"
ALLOWED_HOSTS_CSV="${PROMPT_CATALOG_ALLOWED_IMAGE_HOSTS:-static.cloudbase.eu.org}"

missing_or_unsafe=0
visible_rows=0
visible_cases=0
cases_with_image=0
required_gaps=0
disallowed_hosts=0
total_public_urls=0
checked_urls=0
sample_failures=0
prompt_cache_files=""
prompt_cache_retired_files=""
prompt_cache_unresolved_files=""

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

is_truthy() {
  case "${1:-}" in
    1|true|TRUE|yes|YES|on|ON) return 0 ;;
    *) return 1 ;;
  esac
}

warn_or_fail() {
  echo "WARN: $*" >&2
  missing_or_unsafe=$((missing_or_unsafe + 1))
  return 0
}

check_urls() {
  local input_path="$1"
  local output_path="$2"
  PROMPT_CATALOG_URL_CHECK_TIMEOUT_FOR_CHILD="$URL_CHECK_TIMEOUT" \
    xargs -n 1 -P "$URL_CHECK_CONCURRENCY" bash -c '
    url="$1"
    timeout="${PROMPT_CATALOG_URL_CHECK_TIMEOUT_FOR_CHILD:-10}"
    if curl -fsSI --max-time "$timeout" "$url" >/dev/null 2>&1 \
      || curl -fsS --max-time "$timeout" -H "range: bytes=0-0" -o /dev/null "$url" >/dev/null 2>&1; then
      printf "ok\t%s\n" "$url"
    else
      printf "failed\t%s\n" "$url"
    fi
  ' _ < "$input_path" > "$output_path"
}

retry_failed_urls() {
  local results_path="$1"
  local retry_count=0
  if ! [[ "$URL_CHECK_RETRIES" =~ ^[0-9]+$ ]]; then
    warn_or_fail "PROMPT_CATALOG_URL_CHECK_RETRIES must be a non-negative integer"
    URL_CHECK_RETRIES=0
  fi
  while [[ "$retry_count" -lt "$URL_CHECK_RETRIES" ]]; do
    if ! grep -q '^failed	' "$results_path"; then
      break
    fi
    retry_count=$((retry_count + 1))
    local retry_urls retry_results merged_results
    retry_urls="$(mktemp)"
    retry_results="$(mktemp)"
    merged_results="$(mktemp)"
    awk -F '\t' '$1 == "failed" { print $2 }' "$results_path" > "$retry_urls"
    echo "prompt_image_url_retry_${retry_count}_count=$(wc -l < "$retry_urls" | tr -d '[:space:]')"
    check_urls "$retry_urls" "$retry_results"
    awk -F '\t' '
      FNR == NR {
        retry[$2] = $1
        next
      }
      $1 == "failed" && retry[$2] == "ok" {
        print "ok\t" $2
        next
      }
      print
    ' "$retry_results" "$results_path" > "$merged_results"
    mv "$merged_results" "$results_path"
    rm -f "$retry_urls" "$retry_results"
  done
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

csv_to_sql_array() {
  local csv="$1"
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

json_escape() {
  node -e 'process.stdout.write(JSON.stringify(process.argv[1] || ""))' "$1"
}

write_acceptance_report() {
  local status="$1"
  [[ -z "$ACCEPTANCE_REPORT_PATH" ]] && return 0
  mkdir -p "$(dirname "$ACCEPTANCE_REPORT_PATH")"
  cat > "$ACCEPTANCE_REPORT_PATH" <<JSON
{
  "schema": "sub2api-prompt-catalog-acceptance/v1",
  "status": $(json_escape "$status"),
  "generated_at": $(json_escape "$(date -u '+%Y-%m-%dT%H:%M:%SZ')"),
  "base_url": $(json_escape "$BASE_URL"),
  "api_base": $(json_escape "$API_BASE"),
  "strict": $(if is_truthy "$STRICT"; then printf 'true'; else printf 'false'; fi),
  "allowed_image_hosts": $(json_escape "$ALLOWED_HOSTS_CSV"),
  "database": {
    "visible_rows": ${visible_rows:-0},
    "visible_cases": ${visible_cases:-0},
    "cases_with_image": ${cases_with_image:-0},
    "required_field_gaps": ${required_gaps:-0}
  },
  "image_host_policy": {
    "disallowed_hosts": ${disallowed_hosts:-0}
  },
  "url_reachability": {
    "ran": $(if is_truthy "$RUN_URL_SAMPLE"; then printf 'true'; else printf 'false'; fi),
    "public_url_count": ${total_public_urls:-0},
    "checked_url_count": ${checked_urls:-0},
    "failure_count": ${sample_failures:-0},
    "report_path": $(json_escape "$URL_REPORT_PATH")
  },
  "prompt_cache_mapping": {
    "files": $(json_escape "$prompt_cache_files"),
    "retired_files": $(json_escape "$prompt_cache_retired_files"),
    "unresolved_files": $(json_escape "$prompt_cache_unresolved_files")
  },
  "missing_or_unsafe_items": ${missing_or_unsafe:-0}
}
JSON
  echo "prompt_catalog_acceptance_report_path=${ACCEPTANCE_REPORT_PATH}"
}

url_limit_clause() {
  case "$(printf '%s' "$URL_SAMPLE_LIMIT" | tr '[:upper:]' '[:lower:]')" in
    0|all|full)
      printf ''
      ;;
    *)
      printf 'LIMIT %s' "$URL_SAMPLE_LIMIT"
      ;;
  esac
}

section "Prompt Catalog production scope"
if is_truthy "$STRICT" && [[ -z "$URL_REPORT_PATH" ]]; then
  URL_REPORT_PATH="/tmp/prompt-catalog-url-reachability.tsv"
fi
if is_truthy "$STRICT" && [[ -z "$ACCEPTANCE_REPORT_PATH" ]]; then
  ACCEPTANCE_REPORT_PATH="/tmp/prompt-catalog-acceptance.json"
fi
echo "BASE_URL=${BASE_URL}"
echo "API_BASE=${API_BASE}"
echo "allowed_image_hosts=${ALLOWED_HOSTS_CSV}"
echo "run_url_sample=${RUN_URL_SAMPLE}"
echo "url_sample_limit=${URL_SAMPLE_LIMIT}"
echo "url_check_concurrency=${URL_CHECK_CONCURRENCY}"
echo "url_check_timeout_seconds=${URL_CHECK_TIMEOUT}"
echo "url_check_retries=${URL_CHECK_RETRIES}"
echo "url_check_verbose=${URL_CHECK_VERBOSE}"
echo "url_report_path=${URL_REPORT_PATH:-disabled}"
echo "acceptance_report_path=${ACCEPTANCE_REPORT_PATH:-disabled}"

section "Database coverage"
if [[ "$(count_query "SELECT CASE WHEN to_regclass('public.prompt_catalog_items') IS NULL THEN 0 ELSE 1 END;")" != "1" ]]; then
  warn_or_fail "prompt_catalog_items table is missing"
else
  psql_query -P null='NULL' -c "
WITH visible AS (
  SELECT *
  FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL
), flags AS (
  SELECT
    id,
    source_type,
    NULLIF(TRIM(COALESCE(title, '')), '') IS NULL AS missing_title,
    NULLIF(TRIM(COALESCE(prompt, '')), '') IS NULL AS missing_prompt,
    NULLIF(TRIM(COALESCE(category, '')), '') IS NULL AS missing_category,
    NULLIF(TRIM(COALESCE(source_project, '')), '') IS NULL AS missing_source_project,
    NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_preview_url, '')), '') IS NOT NULL
      OR NULLIF(TRIM(COALESCE(image_thumb_url, '')), '') IS NOT NULL
      OR jsonb_array_length(COALESCE(image_urls, '[]'::jsonb)) > 0 AS has_image
  FROM visible
)
SELECT
  COUNT(*) AS visible_rows,
  COUNT(*) FILTER (WHERE source_type = 'case') AS visible_cases,
  COUNT(*) FILTER (WHERE source_type = 'template') AS visible_templates,
  COUNT(*) FILTER (WHERE source_type = 'case' AND has_image) AS cases_with_image,
  COUNT(*) FILTER (WHERE source_type = 'case' AND NOT has_image) AS cases_without_image,
  COUNT(*) FILTER (WHERE missing_title OR missing_prompt OR missing_category OR missing_source_project) AS required_field_gaps
FROM flags;
"

  visible_rows="$(count_query "SELECT COUNT(*) FROM prompt_catalog_items WHERE status = 'published' AND deleted_at IS NULL;")"
  visible_cases="$(count_query "SELECT COUNT(*) FROM prompt_catalog_items WHERE status = 'published' AND deleted_at IS NULL AND source_type = 'case';")"
  cases_with_image="$(count_query "
WITH visible AS (
  SELECT *
  FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND source_type = 'case'
)
SELECT COUNT(*)
FROM visible
WHERE NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
   OR NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
   OR NULLIF(TRIM(COALESCE(image_preview_url, '')), '') IS NOT NULL
   OR NULLIF(TRIM(COALESCE(image_thumb_url, '')), '') IS NOT NULL
   OR jsonb_array_length(COALESCE(image_urls, '[]'::jsonb)) > 0;
")"
  required_gaps="$(count_query "
SELECT COUNT(*)
FROM prompt_catalog_items
WHERE status = 'published'
  AND deleted_at IS NULL
  AND (
    NULLIF(TRIM(COALESCE(title, '')), '') IS NULL
    OR NULLIF(TRIM(COALESCE(prompt, '')), '') IS NULL
    OR NULLIF(TRIM(COALESCE(category, '')), '') IS NULL
    OR NULLIF(TRIM(COALESCE(source_project, '')), '') IS NULL
  );
")"

  if [[ "$visible_rows" -lt 1 ]]; then
    warn_or_fail "Prompt Catalog has no visible rows"
  fi
  if [[ "$visible_cases" -lt 1 ]]; then
    warn_or_fail "Prompt Catalog has no visible cases"
  fi
  if [[ "$cases_with_image" -lt 1 ]]; then
    warn_or_fail "Prompt Catalog has no visible cases with image coverage"
  fi
  if [[ "$required_gaps" != "0" ]]; then
    warn_or_fail "Prompt Catalog visible rows have required field gaps"
  fi
fi

section "Prompt catalog integrity checks"
tools/prompt-catalog-integrity.sh
REQUIRE_PROMPT_CATALOG_PARITY="$STRICT" tools/prompt-catalog-parity-audit.sh

section "Image host policy"
allowed_hosts_sql="$(csv_to_sql_array "$ALLOWED_HOSTS_CSV")"
disallowed_hosts="$(count_query "
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
  UNION ALL
  SELECT jsonb_array_elements_text(COALESCE(image_urls, '[]'::jsonb)) AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL
), hosts AS (
  SELECT regexp_replace(url, '^https?://([^/?#]+).*$', '\\1') AS host
  FROM urls
  WHERE url LIKE 'http://%' OR url LIKE 'https://%'
)
SELECT COUNT(*)
FROM hosts
WHERE NOT (host = ANY(${allowed_hosts_sql}));
")"
echo "prompt_disallowed_image_hosts=${disallowed_hosts}"
if [[ "$disallowed_hosts" != "0" ]]; then
  warn_or_fail "Prompt Catalog image URLs include hosts outside PROMPT_CATALOG_ALLOWED_IMAGE_HOSTS"
fi

section "Prompt image URL reachability"
if is_truthy "$RUN_URL_SAMPLE"; then
  tmp_urls="$(mktemp)"
  total_public_urls="$(count_query "
WITH urls AS (
  SELECT image_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
  UNION
  SELECT image_original_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
  UNION
  SELECT image_preview_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_preview_url, '')), '') IS NOT NULL
  UNION
  SELECT image_thumb_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_thumb_url, '')), '') IS NOT NULL
  UNION
  SELECT jsonb_array_elements_text(COALESCE(image_urls, '[]'::jsonb)) AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL
)
SELECT COUNT(*)
FROM urls
WHERE url LIKE 'http://%' OR url LIKE 'https://%';
")"
  limit_clause="$(url_limit_clause)"
  psql_query -Atc "
WITH urls AS (
  SELECT image_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_url, '')), '') IS NOT NULL
  UNION
  SELECT image_original_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_original_url, '')), '') IS NOT NULL
  UNION
  SELECT image_preview_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_preview_url, '')), '') IS NOT NULL
  UNION
  SELECT image_thumb_url AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL AND NULLIF(TRIM(COALESCE(image_thumb_url, '')), '') IS NOT NULL
  UNION
  SELECT jsonb_array_elements_text(COALESCE(image_urls, '[]'::jsonb)) AS url FROM prompt_catalog_items
  WHERE status = 'published' AND deleted_at IS NULL
)
SELECT url
FROM urls
WHERE url LIKE 'http://%' OR url LIKE 'https://%'
ORDER BY url
${limit_clause};
" > "$tmp_urls"
  checked_urls="$(wc -l < "$tmp_urls" | tr -d '[:space:]')"
  echo "prompt_public_image_urls=${total_public_urls}"
  echo "prompt_image_urls_checked=${checked_urls}"
  if [[ ! -s "$tmp_urls" ]]; then
    warn_or_fail "Prompt Catalog has no public image URLs to sample"
  else
    if ! [[ "$URL_CHECK_CONCURRENCY" =~ ^[0-9]+$ ]] || [[ "$URL_CHECK_CONCURRENCY" -lt 1 ]]; then
      warn_or_fail "PROMPT_CATALOG_URL_CHECK_CONCURRENCY must be a positive integer"
      URL_CHECK_CONCURRENCY=1
    fi
    if ! [[ "$URL_CHECK_TIMEOUT" =~ ^[0-9]+$ ]] || [[ "$URL_CHECK_TIMEOUT" -lt 1 ]]; then
      warn_or_fail "PROMPT_CATALOG_URL_CHECK_TIMEOUT_SECONDS must be a positive integer"
      URL_CHECK_TIMEOUT=10
    fi

    tmp_results="$(mktemp)"
    check_urls "$tmp_urls" "$tmp_results"
    retry_failed_urls "$tmp_results"

    sample_failures="$(grep -c '^failed	' "$tmp_results" || true)"
    if [[ -n "$URL_REPORT_PATH" ]]; then
      mkdir -p "$(dirname "$URL_REPORT_PATH")"
      {
        printf "status\turl\n"
        cat "$tmp_results"
      } > "$URL_REPORT_PATH"
      echo "prompt_image_url_report_path=${URL_REPORT_PATH}"
      echo "prompt_image_url_report_rows=$((checked_urls + 1))"
    fi
    if is_truthy "$URL_CHECK_VERBOSE"; then
      awk -F '\t' '
        $1 == "ok" { printf "HEAD %s ... ok\n", $2; next }
        $1 == "failed" { printf "HEAD %s ... failed\n", $2 > "/dev/stderr"; next }
      ' "$tmp_results"
    else
      echo "prompt_image_urls_ok=$(grep -c '^ok	' "$tmp_results" || true)"
      echo "prompt_image_urls_failed=${sample_failures}"
      if [[ "$sample_failures" -gt 0 ]]; then
        awk -F '\t' '
          $1 == "failed" && shown < 20 {
            printf "HEAD %s ... failed\n", $2 > "/dev/stderr"
            shown += 1
          }
        ' "$tmp_results"
        if [[ "$sample_failures" -gt 20 ]]; then
          echo "WARN: showing first 20 failed Prompt Catalog image URLs; see ${URL_REPORT_PATH:-$tmp_results} for the full report" >&2
        fi
      fi
    fi
    rm -f "$tmp_results"
    if [[ "$sample_failures" -gt 0 ]]; then
      warn_or_fail "Prompt Catalog image URL sample failures=${sample_failures}"
    fi
  fi
  rm -f "$tmp_urls"
else
  echo "URL sample skipped. Set RUN_PROMPT_CATALOG_URL_SAMPLE=1 to sample Prompt Catalog image URLs."
  warn_or_fail "Prompt Catalog image URL reachability sample has not run"
fi

section "Production readiness result"
if [[ "$missing_or_unsafe" -gt 0 ]]; then
  echo "prompt_catalog_production_ready=false"
  echo "missing_or_unsafe_items=${missing_or_unsafe}"
  write_acceptance_report "failed"
  if is_truthy "$STRICT"; then
    exit 2
  fi
else
  echo "prompt_catalog_production_ready=true"
  write_acceptance_report "passed"
fi

echo
echo "Prompt Catalog production preflight complete."
