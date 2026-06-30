#!/usr/bin/env bash
set -euo pipefail

APPLY="${IMAGE_WORKSPACE_CLEAN_MOCK_APPLY:-0}"
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

echo "== Image Workspace mock/test artifacts =="
psql_query -P null='NULL' -c "
SELECT
  a.id AS artifact_id,
  a.task_id,
  a.user_id,
  u.email,
  a.storage_provider,
  a.storage_key,
  a.image_url,
  a.created_at
FROM image_workspace_artifacts a
LEFT JOIN users u ON u.id = a.user_id
WHERE a.image_url LIKE '%assets.example.test%'
   OR a.storage_key LIKE 'image-workspace-e2e/%'
ORDER BY a.id;
"

echo
echo "== Candidate cleanup summary =="
psql_query -P null='NULL' -c "
WITH mock_artifacts AS (
  SELECT DISTINCT user_id
  FROM image_workspace_artifacts
  WHERE image_url LIKE '%assets.example.test%'
     OR storage_key LIKE 'image-workspace-e2e/%'
), smoke_users AS (
  SELECT u.id
  FROM users u
  JOIN mock_artifacts m ON m.user_id = u.id
  WHERE u.email LIKE 'iw-e2e-%@example.test'
)
SELECT
  (SELECT COUNT(*) FROM mock_artifacts) AS users_with_mock_artifacts,
  (SELECT COUNT(*) FROM smoke_users) AS deletable_smoke_users,
  (SELECT COUNT(*) FROM image_workspace_artifacts WHERE image_url LIKE '%assets.example.test%' OR storage_key LIKE 'image-workspace-e2e/%') AS mock_artifacts,
  (SELECT COUNT(*) FROM image_workspace_tasks WHERE user_id IN (SELECT id FROM smoke_users)) AS smoke_tasks,
  (SELECT COUNT(*) FROM image_workspace_templates WHERE user_id IN (SELECT id FROM smoke_users)) AS smoke_templates,
  (SELECT COUNT(*) FROM image_workspace_usage_records WHERE user_id IN (SELECT id FROM smoke_users)) AS smoke_usage_records;
"

if [[ "$APPLY" != "1" ]]; then
  echo
  echo "Dry run only. Set IMAGE_WORKSPACE_CLEAN_MOCK_APPLY=1 to delete matching iw-e2e smoke users and cascading task/artifact rows."
  exit 0
fi

echo
echo "== Applying cleanup =="
psql_query -P null='NULL' -c "
WITH mock_artifacts AS (
  SELECT DISTINCT user_id
  FROM image_workspace_artifacts
  WHERE image_url LIKE '%assets.example.test%'
     OR storage_key LIKE 'image-workspace-e2e/%'
), deleted_users AS (
  DELETE FROM users u
  USING mock_artifacts m
  WHERE u.id = m.user_id
    AND u.email LIKE 'iw-e2e-%@example.test'
  RETURNING u.id, u.email
)
SELECT COUNT(*) AS deleted_smoke_users FROM deleted_users;
"

echo
echo "Image Workspace mock artifact cleanup complete."
