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
