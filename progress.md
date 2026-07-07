## Worker split progress

Done:
- Added independent `hot-worker` and `x-auto-worker` Docker build targets.
- Split the content compose overlay into `hot-worker` and `x-auto-worker` services.
- Added CI image matrix entries for `hot-worker-sha-*` and `x-auto-worker-sha-*`.
- Updated Worker Management target mapping so Hot and X Auto deploy different images/services.
- Added PostgreSQL storage support for X Auto automation and author-alpha data.
- Added a one-shot SQLite to PostgreSQL migration CLI.
- Removed the legacy combined `content-worker` from regular image publishing,
  Worker Management targets, and example env configuration.
- Split WeChat export and Image Workspace default data mounts into
  `./data/wechat` and `./data/image-workspace`.

Simplifications:
- Hot RSS now has a container that only runs `python -m x_atuo.hot_rss_worker`.
- X Auto now has a container that only runs the API/scheduler surface.
- X Auto production compose now points storage at PostgreSQL instead of sqlite files.
- Production deploy scripts no longer restart the combined content worker by default.
- Business workers no longer share the top-level `./data` mount by default.

Remaining risks:
- SQLite fallback remains in code for rollback/local compatibility and can be deleted after production has run PostgreSQL storage stably.
- Rollback for the former combined content worker now relies on the production
  backup instead of regular compose/env entries.

Next:
- Sync the same env/compose cleanup to production and restart the affected
  workers.

Verification:
- Passed: compose render for API, WeChat, Image Workspace, Hot, and X Auto
  services.
- Passed: `go test ./internal/handler -run 'TestRuntimeWorker|TestHomeBusiness'`.
- Passed: `pnpm --dir frontend typecheck`.
- Passed: `PYTHONPATH=tools/x-atuo/src python3 -m unittest discover -s tools/x-atuo -p 'test*.py'`.
- Failed: `go test ./...` with the pre-existing local OAuth/SignupGrant changes
  using `NOW()` against SQLite; this is outside the worker cleanup diff and was
  not staged in this commit.

## User insight progress

Done:
- Added an admin aggregate user insights endpoint at `/api/v1/admin/users/profile-insights`.
- Added the `/admin/user-insights` frontend page and admin sidebar entry.
- Added registration request-context capture for IP, UA, language, device fingerprint, and a safe header snapshot.
- Added per-user profile summary entry from the user table menu.

Simplifications:
- Reused the existing admin service/handler route structure instead of adding a separate dashboard module.
- Kept raw DB/Redis/runtime access inside service methods; the frontend consumes a compact DTO.

Remaining risks:
- Aggregate labels are currently returned by the backend and are mostly Chinese; frontend i18n covers page chrome.
- Historical users before `user_registration_events` migration can only show partial registration context.

Next:
- Apply migration in production before relying on registration IP/UA aggregates for newly-created accounts.

Verification:
- Passed: `go test ./internal/service ./internal/handler ./internal/server/routes ./internal/repository`.
- Passed: `pnpm --dir frontend typecheck`.
- Passed: `pnpm --dir frontend test:run src/views/admin/__tests__/UsersView.spec.ts`.
- Passed: `git diff --check`.

## Image workspace upstream compatibility

Done:
- Switched production Image Workspace upstream to the 4Router image egress endpoint.
- Updated production model settings to expose only `gpt-image-2` with `auto` and `low` quality options.
- Added worker-side quality normalization for `img.4router.net` so legacy `standard`, `hd`, and `high` tasks are sent as `auto`.

Simplifications:
- Kept the compatibility rule scoped to `img.4router.net`; other OpenAI-compatible image providers keep their original quality behavior.

Remaining risks:
- 4Router image egress accepted `auto` and `low` during probing; other quality values are intentionally hidden until verified.

Next:
- Deploy a fresh image-workspace-worker image after CI publishes the new worker image.

Verification:
- Passed: `node --check tools/image-workspace-worker/src/worker.mjs`.
- Passed: production runtime config returns `https://img.4router.net/v1/images/generations`.
- Passed: production probe to `img.4router.net` with `gpt-image-2` and `auto` quality returned image data.

## User registration IP in admin list

Done:
- Added registration IP, registration UA, and registration accept-language fields to the admin user DTO.
- Hydrated admin user list rows from the latest `user_registration_events` record per user.
- Added a default-visible Registration IP column to the admin users table.

Simplifications:
- Kept registration context as a read-only list enrichment in admin service; repository list filtering remains unchanged.
- Displayed only IP in the table, with UA available in the cell title for quick inspection.

Remaining risks:
- Historical users without `user_registration_events` still show `-`, even if they were created before registration capture existed.

Verification:
- Passed: `go test -tags=unit ./internal/handler/dto -run TestUserFromServiceAdmin_MapsActivityTimestamps`.
- Passed: `go test -tags=unit ./internal/service -run 'TestAdminService_ListUsers|TestNonExistent'`.
- Passed: `pnpm --dir frontend test:run src/views/admin/__tests__/UsersView.spec.ts`.
- Passed: `pnpm --dir frontend typecheck`.
- Passed: `git diff --check`.

## User operation timeline

Done:
- Added a `timeline` section to the admin user profile summary response.
- Aggregated timestamped user activity from registration events, auth identities, API keys, gateway usage, payment orders, balance ledger, redeem codes, Image Workspace tasks, WeChat export tasks, and HTTP ops logs.
- Rendered the operation timeline in the user profile summary modal with source, title, status, amount, IP, and UA context.

Simplifications:
- Reused the existing profile summary endpoint instead of adding a separate timeline endpoint.
- Capped the combined timeline at 200 newest records to keep the profile modal responsive.

Remaining risks:
- Very old user actions that were never written to an event/log table cannot be reconstructed.
- High-volume API users see the newest records first; deeper pagination would need a dedicated timeline endpoint.

Verification:
- Passed: `go test -tags=unit ./internal/service -run 'TestSortUserProfileTimeline|TestAdminService_ListUsers'`.
- Passed: `go test -tags=unit ./internal/handler/dto -run TestUserFromServiceAdmin_MapsActivityTimestamps`.
- Passed: `pnpm --dir frontend test:run src/views/admin/__tests__/UsersView.spec.ts src/components/admin/user/__tests__/UserProfileSummaryModal.spec.ts`.
- Passed: `pnpm --dir frontend typecheck`.
- Passed: `go test -tags=unit ./internal/service ./internal/handler/dto`.
- Passed: `git diff --check`.
