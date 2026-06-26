## 2026-06-26 WeChat export closed-loop restoration
### Done
- Restored the first usable WeChat export loop on the integrated Sub2API surface.
- Added real PostgreSQL-backed session, article, task, and artifact repository/service behavior for the `wechat_*` tables.
- Replaced placeholder `/api/v1/wechat/*` handlers with authenticated article import, task creation/listing/status, artifact listing/download, and session create/poll/logout.
- Added local/private-network worker endpoints under `/api/v1/wechat/worker/*`; production can require `WECHAT_EXPORT_WORKER_TOKEN`.
- Reworked `/wechat-export` from a capability placeholder into a logged-in workspace for importing links, selecting `html / markdown / json`, creating tasks, refreshing status, and opening artifact downloads.
- Implemented the independent Node worker loop in `tools/wechat-worker`, reusing the Touch exporter core to fetch article HTML, generate artifacts, and complete/fail tasks via Go API.

### Validation
- `go test ./internal/handler ./internal/server/routes ./internal/service ./internal/repository -run 'TestWeChatExport|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/WeChatExportView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:build`
- Built backend with `-tags embed`, replaced the running `sub2api` container binary, and verified `http://127.0.0.1:8080/wechat-export` returns 200.
- Ran an end-to-end local task: imported article link `id=1`, created task `id=2`, ran Node worker once, task completed with `successful_article_count=1`, and produced 3 artifacts.

### Notes
- The QR session endpoint is persisted and pollable, but still uses an internal `wechat-export://login?token=...` token URL; a real WeChat QR login executor must still mark sessions `ready` with cookies.
- Synced account/article history endpoints remain structural; the first closed loop is direct-link import plus worker export.
- `tools/wechat-worker` installed its own npm dependency tree and lockfile; npm reported 2 moderate audit findings from the legacy dependency chain.

## 2026-06-25 WeChat export restoration foundation
### Done
- Added the first Go-side WeChat export domain migration with sessions, accounts, articles, export tasks, and artifacts tables.
- Added backend domain skeletons for WeChat export service and repository so the capability has explicit types and a future DI landing zone.
- Added a new `tools/wechat-worker/` Node worker workspace and copied the reusable Touch exporter/gateway/export-runner core needed for a `html / markdown / json` worker.
- Kept the worker entrypoint intentionally minimal for now; the next slice will connect it to PG task claiming and artifact upload instead of reviving old Touch-local business state.
- Registered the first authenticated `/api/v1/wechat/*` route skeletons and wired a dedicated WeChat export handler into the main Go service graph.

### Validation
- `go test ./internal/service ./internal/repository ./internal/server/routes -run '^$'`

### Notes
- This slice now exposes the route skeleton, but the handler methods still return placeholder responses until repository-backed session/account/article/task logic is filled in.

## 2026-06-25 Split /home and /sub homepage positioning
### Done
- Replaced the temporary `/sub` alias with a dedicated public `/sub` route that still reuses `HomeView`.
- Kept `/sub` on the current Sub2API platform-style homepage copy sourced from `home_shell_config`.
- Switched `/home` to a built-in business-capability homepage narrative focused on WeChat export, hot-topic tracking, prompt cases, and the image workspace.
- Avoided duplicating the homepage component by adding path-aware shell selection inside `HomeView`.
- Added focused tests proving `/sub` keeps runtime-configured copy while `/home` renders the new business-capability copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts src/utils/__tests__/homeShell.spec.ts src/router/__tests__/guards.spec.ts`

### Notes
- `/home` currently uses frontend-owned business copy defaults because the new business homepage shell is not yet exposed as a separate admin runtime setting.

## 2026-06-25 Add /sub alias for the homepage
### Done
- Added `/sub` as a public alias of the existing `/home` route.
- Kept `/sub` and `/home` on the same `HomeView` component so homepage content stays in sync.
- Added a router-guard regression test to keep `/sub` treated as a public route.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/router/__tests__/guards.spec.ts`

### Notes
- This is a route alias only; it does not create a second homepage implementation.

## 2026-06-18 Web runtime settings renamed away from Touch
### Done
- Confirmed the retired Next app is no longer present at `apps/touch`; the pnpm workspace builds only the Vue `frontend`.
- Confirmed runtime routes use generic `/api/v1/web/*`, `/api/v1/prompts/*`, and `/api/v1/admin/prompts/*`; legacy `/api/v1/touch/*` references are limited to negative tests and documentation guardrails.
- Added generic web/public runtime setting keys for app metadata, prompt gallery copy, pricing/credits shell copy, auth visibility, and public integration snippets.
- Switched settings update/default paths to write generic keys instead of new `touch_*` rows.
- Renamed internal service/admin runtime setting fields and variables from `Touch*`/`touch*` to `Web*`/`web*`.
- Changed legacy `touch_*` JSON tags on service-level public settings to `json:"-"`; public/admin DTOs continue to expose only generic runtime fields.
- Removed duplicate `Web*` compatibility fields from the HTML-injected `PublicSettingsInjectionPayload`; `window.__APP_CONFIG__` now carries only the generic public runtime keys while service/admin compatibility fields remain internal.
- Removed duplicate Web runtime setting keys from the `GetPublicSettings` fetch list.
- Renamed the public sitemap helper and comments from Touch-oriented wording to generic Vue/web wording.
- Fixed public settings legacy compatibility so empty/missing generic runtime keys no longer mask existing `touch_*` values; `GetPublicSettings` now reads the legacy keys it falls back to.
- Renamed non-legacy runtime setting test fixtures from Touch examples to Web examples, while keeping the dedicated legacy `touch_*` fallback test explicit.
- Tightened the runtime fallback rule so an explicitly stored generic key, even when empty, masks legacy `touch_*` data; only absent generic keys fall back to legacy rows.
- Renamed the Runtime Settings frontend test fixture from Touch examples to Web examples while preserving the `touch_*` exclusion assertion.
- Moved Prompt Gallery primary image selection into the Sub2API prompt catalog DTO via `primary_image_url`; the Vue gallery now consumes that field first and keeps old image fields only as compatibility fallback.
- Moved Prompt Gallery tag presentation into the Sub2API prompt catalog DTO via `all_tags` and `visible_tags`; the Vue gallery now consumes those fields first and only keeps local tag merging for older API compatibility.
- Moved Prompt Gallery source badge text and prompt character count into the Sub2API prompt catalog DTO via `source_display_label` and `prompt_char_count`; the Vue gallery consumes those fields first and keeps local fallbacks only for older API compatibility.
- Moved Prompt Gallery facet option text into the Sub2API prompt catalog summary via `display_label`; the Vue gallery now uses API-provided source/category option labels before old local label/count formatting fallback.
- Removed Prompt Gallery's old local display fallbacks for primary image selection, source label, character count, visible tags, full tags, and facet label formatting; those values are now treated as Sub2API DTO responsibilities.
- Moved the credits conversion ratio out of Touch/Vue hardcoding into the Sub2API `credits_per_balance` public/admin runtime setting; Credits view and Web session credits now use that setting, with backend default/validation at `10`.
- Kept legacy `touch_*` setting keys as read-only fallback compatibility for existing databases.
- Added `149_copy_touch_runtime_settings_to_web.sql` to copy existing non-empty `touch_*` runtime settings into generic keys without overwriting existing generic values.
- Added a regression test covering old `touch_*` fallback reads.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings|TestSettingService_UpdateSettings' -count=1`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings' -count=1`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings|TestSettingService_UpdateSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/admin -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler_UpdateSettings_AcceptsGenericRuntimeAliases|TestSettingHandler_GetSettings_InjectsAuthSourceDefaults|TestSettingHandler_GetRobotsTxt|TestSettingHandler_GetSitemapXML' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/admin -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler_GetAdsTxt|TestSettingHandler_GetFaviconICO|TestSettingHandler_GetRobotsTxt|TestSettingHandler_GetSitemapXML|TestSettingHandler_UpdateSettings_AcceptsGenericRuntimeAliases' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `go test ./internal/server/routes ./internal/handler -run 'TestPromptCasesPublicRouteFiltersAndSummary|TestPromptCasesPublicRoute|^$' -count=1`
- `go test ./internal/service ./internal/handler ./internal/server/routes ./cmd/server -run '^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts`
- `pnpm run frontend:typecheck`
- `go test ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler_GetFaviconICO|TestSettingHandler_GetRobotsTxt|TestSettingHandler_GetSitemapXML' -count=1`
- `go test ./internal/handler ./cmd/server -run '^$' -count=1`
- `go test ./internal/repository -run 'TestMigrationChecksumCompatibilityRules_CoverEditedUpgradeCompatibilityMigrations|TestIsMigrationChecksumCompatible_AdditionalCases' -count=1`
- `go test ./internal/service ./internal/handler ./internal/handler/admin ./internal/repository ./cmd/server -run '^$' -count=1`
- `git diff --check -- backend/internal/service/domain_constants.go backend/internal/service/settings_view.go backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go backend/internal/service/setting_service_update_test.go backend/internal/handler/setting_handler.go backend/internal/handler/setting_handler_public_test.go backend/internal/handler/admin/setting_handler.go backend/internal/handler/admin/setting_handler_auth_source_defaults_test.go backend/migrations/149_copy_touch_runtime_settings_to_web.sql progress.md`
- `rg -n "apps/touch|sub2api-touch|NEXT_PUBLIC_SUB2API_BASE_URL|SUB2API_BASE_URL|next dev|next build|next start|pnpm --filter.*touch" package.json pnpm-workspace.yaml Dockerfile deploy docs frontend backend -S`
- `rg -n "api/v1/touch|/touch/|touch/web|touch/prompts|Register.*Touch|TouchRoutes|touchRoutes" backend/internal backend/cmd frontend/src docs -S`

### Notes
- `golangci-lint` could not run in this environment because the installed binary was built with Go 1.25 while this project targets Go 1.26.3.
- Remaining `touch_*` runtime strings are intentionally limited to legacy compatibility constants/fallbacks, tests, and migration copy sources.

## 2026-06-18 Prompt catalog table renamed away from Touch
### Done
- Changed prompt catalog repository SQL to use `prompt_catalog_items` instead of `touch_prompt_items`.
- Updated the prompt catalog creation migration to create `prompt_catalog_items` and generic index names for fresh databases.
- Added `148_rename_touch_prompt_items.sql` to rename existing `touch_prompt_items` tables and recreate generic indexes on upgraded databases.
- Added a migration checksum compatibility rule for the old `145_touch_prompt_items.sql` contents so databases that already applied the old table-name migration can still advance to the rename migration.

### Validation
- `go test ./internal/repository -run 'TestMigrationChecksumCompatibilityRules_CoverEditedUpgradeCompatibilityMigrations|TestIsMigrationChecksumCompatible_AdditionalCases' -count=1`
- `go test ./internal/service -run 'TestPromptCatalog|TestTwitterImport' -count=1`
- `go test ./internal/repository ./cmd/server -run '^$' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/repository`
- `rg -n "touch_prompt_items|idx_touch_prompt_items" backend/internal backend/migrations -S`

### Notes
- Remaining `touch_prompt_items` strings are intentionally limited to the compatibility rename migration and its checksum rule.
- The original `145_touch_prompt_items.sql` filename remains for schema migration compatibility; the runtime table name is now generic.

## 2026-06-18 Legacy Touch browser redirects removed
### Done
- Removed `LegacyTouchRedirectMiddleware` from the main backend router.
- Deleted the legacy Touch browser redirect middleware and its tests.
- Updated web platform integration docs to state that `/touch/*` and `/admin/touch/*` browser aliases are retired rather than redirected.

### Validation
- `go test ./internal/server ./internal/web ./cmd/server -run '^$' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/server ./internal/web`
- `pnpm --filter sub2api-frontend exec vitest run src/router/__tests__/legacy-touch-alias-removal.spec.ts src/router/__tests__/runtime-settings-route.spec.ts src/router/__tests__/touch-oauth-compat.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "LegacyTouchRedirect|legacy_touch_redirect|/touch/|/admin/touch|Legacy Touch browser URLs still exist|handled by backend redirects" backend/internal frontend/src docs README* -S`

### Notes
- Old Touch browser paths now rely on the normal SPA/static fallback behavior instead of explicit backend redirects. Remaining `/api/v1/touch/*` strings are negative route tests only.

## 2026-06-18 Web auth bridge token parameter renamed to Sub2API
### Done
- Renamed the trusted web OAuth bridge query parameter from `touch_bridge_token` to `sub2api_web_bridge_token`.
- Kept `source=touch` identity isolation semantics unchanged, because Touch-origin web accounts still need a separate signup source.
- Added a negative test proving the old `touch_bridge_token` parameter name is no longer accepted.

### Validation
- `go test ./internal/handler -run 'TestEmailOAuthStartStoresWebSource|TestEmailOAuthStartRejectsWebSourceWithoutBridgeKey|TestAuthHandlerAllowsWebSourceOnlyWithBridgeKey|TestAuthHandlerAllowsWebOAuthBridgeToken|TestAuthHandlerRejectsLegacyTouchOAuthBridgeTokenName|TestAuthHandlerAllowsTrustedTouchOAuthSourceContext|TestAuthHandlerRejectsWebOAuthBridgeTokenWithoutProviderScope|TestEmailOAuthCallbackWebSourceSkipsPendingManualCompletion' -count=1`
- `go test ./internal/handler ./cmd/server -run 'TestAuthHandler|TestEmailOAuth|^$' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/handler`
- `rg -n "touch_bridge_token" backend/internal/handler frontend/src docs README* -S --glob '!**/*test.go'`

### Notes
- The old parameter name appears only in a negative regression test. The authenticated web surface should use `/api/v1/web/*` directly where possible.

## 2026-06-18 Web session stopped accepting legacy Touch cookies
### Done
- Changed Web session cookie reads to only accept the generic `sub2api_web_access_token` and `sub2api_web_refresh_token` names.
- Kept old `touch_sub2api_*` cookie names only as expired cleanup cookies during login/session refresh/logout.
- Updated cookie tests to assert old Touch cookie names are ignored as authentication input.

### Validation
- `go test ./internal/handler -run 'TestWebSessionCookiesUseGenericNamesAndClearLegacyNames|TestReadWebSessionCookieIgnoresLegacyName|TestClearWebSessionCookiesClearsGenericAndLegacyNames|TestWebCheckoutPaymentSourceDefaultsToGenericWebSource' -count=1`
- `go test ./internal/handler ./cmd/server -run 'TestWeb|^$' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/handler`
- `rg -n "readWebSessionCookie\\([^,]+,[^)]+," backend/internal/handler -S`

### Notes
- The old cookie constants remain only to expire stale browser cookies. They are no longer accepted for `/api/v1/web/*` session authentication.

## 2026-06-18 Settings API stopped exposing Touch runtime aliases
### Done
- Removed legacy `touch_*` runtime configuration fields from backend settings DTOs.
- Stopped accepting legacy `touch_*` runtime configuration aliases in the admin settings update request.
- Stopped returning legacy `touch_*` runtime configuration aliases from admin and public settings responses.
- Kept the existing internal `SettingKeyTouch*` storage fields active so this API cleanup does not require a data migration.

### Validation
- `go test ./internal/handler ./internal/handler/admin -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler_UpdateSettings_AcceptsGenericRuntimeAliases|TestSettingHandler_GetAdsTxt|TestSettingHandler_GetRobotsTxt|TestSettingHandler_GetSitemapXML' -count=1`
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings_ExposesGenericRuntimeSettingAliases|TestSettingHandler_GetAdsTxt|TestSettingHandler_GetFaviconICO|TestSettingHandler_GetRobotsTxt|TestSettingHandler_GetSitemapXML' -count=1`
- `go test ./internal/handler/dto -run '^$' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/handler ./internal/handler/admin`
- `pnpm run frontend:typecheck`
- `rg -n 'json:"touch_(app|prompt|workspace|pricing|credits|locale|email|google_auth|github_auth)' backend/internal/handler backend/internal/handler/dto -S`

### Notes
- Service-layer setting view structs and persisted setting keys still use `Touch*` names internally. That remaining cleanup should be paired with an explicit storage-key migration or compatibility read strategy.

## 2026-06-18 Runtime settings frontend stopped using Touch config fallbacks
### Done
- Removed Runtime Settings form fallback reads from legacy `touch_*` runtime configuration fields.
- Removed legacy Touch runtime configuration fields from frontend settings API and public settings types.
- Kept Runtime Settings save payloads generic-only (`app_name`, `pricing_shell_config`, `credits_shell_config`, auth visibility, integrations, etc.).

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/RuntimeSettingsView.spec.ts src/utils/__tests__/publicIntegrations.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `rg -n "touch_app_url|touch_app_name|touch_prompt_cases|touch_pricing|touch_credits|touch_locale_detect|touch_email_auth|touch_google_auth|touch_github_auth|legacyStringFieldFallbacks|legacyBooleanFieldFallbacks" frontend/src --glob '!**/__tests__/**' -S`

### Notes
- Backend storage constants and DTO internals still contain `Touch*` names for persisted legacy setting keys. The frontend no longer treats them as runtime inputs.

## 2026-06-18 Legacy Touch API discovery route removed
### Done
- Removed the final `/api/v1/touch/capabilities` route by deleting `RegisterLegacyTouchRoutes`.
- Stopped wiring legacy Touch API routes from the main server router.
- Updated route tests so generic `/api/v1/web/*`, `/api/v1/prompts/*`, and `/api/v1/admin/prompts/*` remain the primary API surfaces while old `/api/v1/touch/*` paths are asserted absent.

### Validation
- `go test ./internal/server/routes -run 'TestLegacyTouchAPIRoutesAreNotRegistered|TestWebRoutesExposeOnlyGenericPrimaryRoutes|TestPromptCatalogPublicAliasIsNotRegistered|TestPromptCatalogAdminAliasIsNotRegistered|TestPromptCasesPublicRoute|TestPromptAdminUpsertRouteRequiresAdminAuthAndWrites|TestPromptAdminImportTwitterRoute' -count=1`
- `go test ./internal/server ./cmd/server -run '^$' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/server ./internal/server/routes`
- `rg -n "RegisterLegacyTouchRoutes|/api/v1/touch/capabilities|touch/capabilities" backend frontend docs README* deploy -S`

### Notes
- Legacy browser URL redirects under `backend/internal/web` still remain for old public page URLs; this step removes the legacy Touch API surface only.
- Older `progress.md` entries intentionally keep historical references to `/api/v1/touch/capabilities`.

## 2026-06-18 Image generator draft stopped reading legacy Touch storage key
### Done
- Removed the `touch:image-generator:draft` sessionStorage fallback from the image generator draft utility.
- Kept the generic `sub2api:image-generator:draft` key as the only active draft storage key.
- Updated tests to assert old Touch draft storage is ignored instead of migrated.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/imageGeneratorDraft.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `rg -n "touch:image-generator:draft|LEGACY_TOUCH_IMAGE_GENERATOR_DRAFT_KEY|legacy Touch drafts|legacy draft" frontend/src -S`
- `git diff --check -- frontend/src/utils/imageGeneratorDraft.ts frontend/src/utils/__tests__/imageGeneratorDraft.spec.ts progress.md`

### Notes
- The old key may still exist in a user's browser sessionStorage, but it is no longer read by the Sub2API frontend.

## 2026-06-18 Legacy Touch integration settings removed from public/admin API surface
### Done
- Removed legacy `touch_*` integration fields from public settings DTOs and SSR public-settings injection payloads.
- Removed legacy `touch_*` integration aliases from the admin settings update request.
- Kept the generic integration fields (`google_analytics_id`, `crisp_enabled`, etc.) active; they still persist through the existing underlying setting keys until a separate storage-key migration is introduced.
- Removed frontend `PublicSettings` / admin settings API type fields for the legacy Touch integration keys.
- Updated Runtime Settings and public integration tests to assert old Touch integration fields are ignored or absent from responses.

### Validation
- `go test ./internal/handler ./internal/handler/admin -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler_GetAdsTxt|TestSettingHandler_UpdateSettings_AcceptsGenericRuntimeAliases|Test.*AuthSource' -count=1`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings|TestSettingService_GetPublicSettingsForInjection|TestPublicSettingsInjection|TestSettingService_UpdateSettings' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/handler ./internal/handler/admin ./internal/service`
- `pnpm run frontend:typecheck`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/publicIntegrations.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/utils/publicIntegrations.ts frontend/src/utils/__tests__/publicIntegrations.spec.ts frontend/src/views/admin/RuntimeSettingsView.vue frontend/src/views/admin/__tests__/RuntimeSettingsView.spec.ts frontend/src/types/index.ts frontend/src/api/admin/settings.ts backend/internal/handler/dto/settings.go backend/internal/service/setting_service.go backend/internal/handler/admin/setting_handler.go backend/internal/handler/admin/setting_handler_auth_source_defaults_test.go backend/internal/handler/setting_handler.go progress.md`

### Notes
- The storage constants are still named `SettingKeyTouch*` for these integrations. That is now an internal persistence detail; replacing the physical setting keys should be handled with an explicit data migration.

## 2026-06-18 Public integrations stopped reading Touch setting fallbacks
### Done
- Removed `touch_*` analytics/ad/chat/affiliate fallback reads from `frontend/src/utils/publicIntegrations.ts`.
- Updated Runtime Settings form loading so integration fields no longer backfill from legacy `touch_*` integration settings.
- Added tests that old Touch integration fields are ignored and do not inject public scripts or get saved back as generic integration values.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/publicIntegrations.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `rg -n "settings\\.touch_(google_analytics|clarity|plausible|openpanel|vercel_analytics|adsense|affonso|promotekit|crisp|tawk)|legacy.*touch_(google_analytics|clarity|plausible|openpanel|vercel_analytics|adsense|affonso|promotekit|crisp|tawk)" frontend/src -S`

### Notes
- Backend DTO/API type definitions still expose legacy `touch_*` integration fields. They are no longer active frontend runtime inputs and can be removed from the settings API in a follow-up cleanup.

## 2026-06-18 Legacy Touch admin bridge removed
### Done
- Removed the remaining `/api/v1/touch/admin/auth|users|payments|subscriptions/*` route registration.
- Kept `/api/v1/touch/capabilities` as a discovery-only legacy endpoint and marked all former Touch admin bridge groups as removed routes.
- Removed `LegacyTouchAuthHandler`, `LegacyTouchPaymentHandler`, and `LegacyTouchSubscriptionHandler` from runtime handler aggregation and Wire startup construction.
- Deleted the now-unreachable legacy Touch handler implementations.
- Removed orphan Touch sync/credit service helpers and the unused `touch_balance_events` migration that only supported the deleted bridge.
- Kept `signup_source = touch` identity separation logic intact because Touch-origin web/OAuth accounts still need source isolation.

### Validation
- `go test ./internal/server/routes ./internal/handler ./internal/repository ./cmd/server -run 'TestTouchCapabilitiesRoute|TestLegacyTouchAdminBridgeRoutesAreNotRegistered|TestPromptCatalogAdminAliasIsNotRegistered|TestUserRepo.*Email|^$' -count=1`
- `go test -tags unit ./internal/service -run 'TestUpdateBalance_Success|TestGetProfileIdentitySummaries|TestStartProfileIdentityBinding|TestCompleteProfileIdentityBinding' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/server/routes ./internal/handler ./internal/service ./internal/repository ./cmd/server`
- `rg -n "SyncTouchUser|DeductTouchUserBalance|GrantTouchUserBalance|TouchUserSyncInput|TouchBalance(Deduct|Grant)|GrantTouchBalanceOnce|GetTouchUserByTouchID|EnsureTouchAuthIdentity|touchCreditsPerBalance|touch_balance_events|LegacyTouch(Auth|Payment|Subscription)|NewLegacyTouch" backend/internal backend/cmd/server backend/migrations frontend/src -S`
- `git diff --check -- backend/internal/server/routes/touch.go backend/internal/server/routes/touch_test.go backend/internal/server/routes/prompt_catalog_test.go backend/internal/handler/handler.go backend/internal/handler/wire.go backend/internal/handler/web_handler.go backend/cmd/server/wire_gen.go backend/internal/service/user_service.go backend/internal/service/user_service_test.go backend/internal/repository/user_repo.go backend/migrations/146_touch_balance_events.sql progress.md`

### Notes
- `/api/v1/touch/capabilities` still intentionally contains Touch strings so older clients can discover replacement/removed route information.
- Older database instances that already applied `touch_balance_events` may keep that unused table until a separate cleanup migration is intentionally added.

## 2026-06-18 Legacy Touch admin prompt aliases removed
### Done
- Removed `/api/v1/touch/admin/prompts/*` registration from `RegisterLegacyTouchRoutes`.
- Kept `/api/v1/admin/prompts/*` as the only admin prompt write/import API surface.
- Added `removed_routes.admin_prompts = /api/v1/touch/admin/prompts/*` to `/api/v1/touch/capabilities`.
- Updated prompt route tests to assert the old Touch admin prompt alias is no longer registered while the generic admin route still writes.

### Validation
- `go test ./internal/server/routes -run 'TestTouchCapabilitiesRoute|TestPromptCatalogPublicAliasIsNotRegistered|TestPromptCatalogAdminAliasIsNotRegistered|TestPromptCasesPublicRoute|TestPromptAdminUpsertRouteRequiresAdminAuthAndWrites|TestPromptAdminImportTwitterRoute' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/server/routes`

### Notes
- The remaining `/api/v1/touch/admin/auth|users|payments|subscriptions/*` bridge routes still exist for legacy server-side compatibility clients.

## 2026-06-18 Legacy Touch web API aliases removed
### Done
- Removed `/api/v1/touch/web/*` registration from `RegisterLegacyTouchRoutes`.
- Kept the generic `/api/v1/web/*` auth/session/checkout/import routes as the single browser-safe Web API surface.
- Added `removed_routes.web = /api/v1/touch/web/*` to `/api/v1/touch/capabilities` so legacy clients can discover the replacement.
- Updated route and auth tests to use or assert the generic Web API path.

### Validation
- `go test ./internal/server/routes -run 'TestWebRoutesExposeOnlyGenericPrimaryRoutes|TestTouchCapabilitiesRoute|TestPromptCatalogPublicAliasIsNotRegistered|TestPromptCasesPublicRoute|TestPrompt.*Admin' -count=1`
- `go test ./internal/handler -run 'TestAuthHandlerAllowsTrustedTouchOAuthSourceContext|TestAuthHandlerAllowsTouchOAuthBridgeToken|TestAuthHandlerRejectsWebOAuthBridgeTokenWithoutProviderScope' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/server/routes`
- `git diff --check -- backend/internal/server/routes/touch.go backend/internal/server/routes/web_test.go backend/internal/server/routes/touch_test.go backend/internal/handler/auth_email_oauth_test.go progress.md`

### Notes
- Legacy `/api/v1/touch/admin/*` bridge routes still remain for server-side compatibility clients. This step removes only the browser-safe `/touch/web/*` API alias layer.

## 2026-06-18 Legacy Touch Vue aliases moved to web redirects
### Done
- Removed legacy `/touch/prompts`, `/touch/generator`, `/touch/pricing`, `/touch/credits`, and `/admin/touch/settings` aliases from the Vue Router.
- Added matching redirects to `LegacyTouchRedirectMiddleware` so old browser URLs still land on generic Sub2API frontend paths.
- Updated router tests to assert generic Vue routes are the only frontend route entries.
- Extended legacy redirect tests for the old Touch browser paths.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/router/__tests__/runtime-settings-route.spec.ts src/router/__tests__/legacy-touch-alias-removal.spec.ts src/router/__tests__/touch-oauth-compat.spec.ts`
- `go test ./internal/web -run TestLegacyTouchRedirectMiddleware -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/web`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/router/index.ts frontend/src/router/__tests__/runtime-settings-route.spec.ts frontend/src/router/__tests__/legacy-touch-alias-removal.spec.ts backend/internal/web/legacy_touch_redirect.go backend/internal/web/legacy_touch_redirect_test.go progress.md`

### Notes
- Backend `/api/v1/touch/*` compatibility routes still remain for API bridge clients. This step only removes legacy Touch browser route aliases from the Vue frontend.

## 2026-06-18 Home feature cards moved into home shell config
### Done
- Extended `home_shell_config` with `experienceCards` and `whyChooseCards` arrays.
- Updated the default Home page to merge configured cards by `key` over the built-in card structure, preserving icon/iconClass defaults when unset.
- Updated Runtime Settings placeholders in English and Chinese to document card override examples.
- Extended the Home page unit test to verify feature and why-choose card copy renders from public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/HomeView.vue frontend/src/views/__tests__/HomeView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts`

### Notes
- The Home page still keeps built-in card structure and visual classes as fallback. Runtime copy for the major Home shell and cards is now configurable from Sub2API public settings.

## 2026-06-18 Home page shell config moved into Sub2API public settings
### Done
- Added `home_shell_config` as a Sub2API setting key with admin update, public settings, SSR injection, and typed frontend/API exposure.
- Added a Runtime Settings editor field for Home page shell JSON.
- Updated the default Home page to read `home_shell_config.labels` for nav, hero, primary CTAs, section headings, footer, and model-family copy while keeping existing i18n text as fallback.
- Extended the Home page unit test to verify configured shell labels render from public settings.

### Validation
- `go test ./internal/handler ./internal/service -run 'Test.*PublicSettings|Test.*Settings|Test.*Schema|Test.*Runtime' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler`
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- backend/internal/service/domain_constants.go backend/internal/service/settings_view.go backend/internal/service/setting_service.go backend/internal/handler/dto/settings.go backend/internal/handler/setting_handler.go backend/internal/handler/admin/setting_handler.go frontend/src/types/index.ts frontend/src/api/admin/settings.ts frontend/src/views/admin/RuntimeSettingsView.vue frontend/src/views/HomeView.vue frontend/src/views/__tests__/HomeView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts progress.md`

### Notes
- Home feature cards and why-choose cards still use local i18n arrays. This step moves the larger shell/navigation/hero/footer/model-family copy into Sub2API settings.

## 2026-06-18 Pricing page shell labels expanded in public settings
### Done
- Expanded `pricing_shell_config.labels` to support the full Pricing page shell copy, including nav, hero, tabs, CTA, card labels, empty states, and validity labels.
- Kept `pricing_title` and `pricing_description` as higher-priority overrides for existing deployments.
- Updated Runtime Settings placeholders in English and Chinese to document the fuller `labels` shape.
- Extended the Pricing page unit test to verify nav, tab, CTA, catalog, and empty-state labels render from public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts`

### Notes
- Pricing catalog data and checkout already remain Sub2API-backed. This step reduces the remaining local Vue shell copy on the Pricing page.

## 2026-06-18 Docs page shell config moved into Sub2API public settings
### Done
- Added `docs_shell_config` as a Sub2API setting key with admin update, public settings, SSR injection, and typed frontend/API exposure.
- Added a Runtime Settings editor field for Docs page shell JSON.
- Updated `/docs` Vue page to read `docs_shell_config.labels` for the page title, login/dashboard buttons, and Docsify search placeholder/no-data copy while keeping existing i18n text as fallback.
- Extended the Docs page source-level regression test to verify the public-settings-driven shell path.

### Validation
- `go test ./internal/handler ./internal/service -run 'Test.*PublicSettings|Test.*Settings|Test.*Schema|Test.*Runtime' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- backend/internal/service/domain_constants.go backend/internal/service/settings_view.go backend/internal/service/setting_service.go backend/internal/handler/dto/settings.go backend/internal/handler/setting_handler.go backend/internal/handler/admin/setting_handler.go frontend/src/types/index.ts frontend/src/api/admin/settings.ts frontend/src/views/admin/RuntimeSettingsView.vue frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts progress.md`

### Notes
- Docs content still lives in the frontend docs bundle. This step moves the surrounding Docs page shell and Docsify search copy into Sub2API settings.

## 2026-06-18 Credits page shell labels moved further into public settings
### Done
- Extended `credits_shell_config` support on the Credits page with a `labels` object for page, balance, conversion, and action copy.
- Kept existing dedicated public settings (`credits_title`, `credits_description`, `credits_purchase_label`, `credits_balance_label`) as higher-priority overrides for current deployments.
- Updated Runtime Settings placeholders in English and Chinese to document the new `labels` shape.
- Added a Credits page unit test covering public-settings-driven labels and Sub2API balance conversion display.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/CreditsView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/CreditsView.vue frontend/src/views/user/__tests__/CreditsView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts`

### Notes
- Balance source and payment flow were already Sub2API-owned. This step reduces another hardcoded UI-copy cluster in the remaining frontend shell.

## 2026-06-18 Models Plaza shell config moved into Sub2API public settings
### Done
- Added `model_plaza_shell_config` as a Sub2API setting key with admin update, public settings, and typed DTO/API/frontend exposure.
- Added a Runtime Settings editor field for Models Plaza shell JSON.
- Updated `/models` Vue page to read `model_plaza_shell_config.labels` for hero, navigation, search, empty state, result, copy, group, and price labels while keeping existing i18n text as fallback.
- Extended the Models Plaza unit test to verify configured labels render from public settings.

### Validation
- `go test ./internal/handler ./internal/service -run 'Test.*Public|Test.*Settings|Test.*ModelPlaza' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- backend/internal/service/domain_constants.go backend/internal/service/settings_view.go backend/internal/service/setting_service.go backend/internal/handler/dto/settings.go backend/internal/handler/setting_handler.go backend/internal/handler/admin/setting_handler.go frontend/src/types/index.ts frontend/src/api/admin/settings.ts frontend/src/views/admin/RuntimeSettingsView.vue frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/__tests__/ModelsPlazaView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts progress.md`

### Notes
- Model card data was already Sub2API-owned via `model_plaza_items`; this step moves more of the page shell and operational copy into Sub2API public/admin settings.

## 2026-06-18 Prompt catalog detail copy moved further into public settings
### Done
- Extended `prompt_catalog_shell_config.labels` with a `charUnit` field for the prompt detail character-count unit.
- Updated Prompt Catalog detail rendering so the character-count unit comes from Sub2API public settings when configured.
- Updated Runtime Settings placeholders in English and Chinese to document the new `charUnit` shell label.
- Added a Prompt Catalog unit test covering the settings-driven detail metadata label.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts progress.md`

### Notes
- Prompt Catalog data, filtering, pagination, and import remain Sub2API-backed. This step only removes another hardcoded display-copy point from the Vue shell.

## 2026-06-18 Frontend bootstrap static fallbacks reduced
### Done
- Removed the hardcoded AdSense script from `frontend/index.html`; advertising script injection now stays on the public-settings-driven `PublicIntegrations` path.
- Changed the initial favicon in `frontend/index.html` from `/logo.png` to `/favicon.svg`.
- Replaced runtime logo display fallbacks from `/logo.png` to `/favicon.svg` across the public shell, auth layout, sidebar, pricing, prompt catalog, image generator, models plaza, legal, home, and key usage pages.

### Validation
- Runtime scan for hardcoded `pagead2.googlesyndication.com`, `ca-pub-8471022013462117`, and `/logo.png` fallback usage.
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/index.html frontend/src/views/HomeView.vue frontend/src/views/KeyUsageView.vue frontend/src/views/public/ImageGeneratorView.vue frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/PricingView.vue frontend/src/views/public/PromptCatalogView.vue frontend/src/components/layout/AppSidebar.vue frontend/src/components/layout/AuthLayout.vue progress.md`

### Notes
- `SUB2API_BASE_URL` / `VITE_API_BASE_URL`-style bootstrap discovery remains intentionally environment-based.
- Public settings still override all logo/site-name display paths when configured.

## 2026-06-18 Pricing page shell copy moved further into public settings
### Done
- Extended `pricing_shell_config.labels` with an `eyebrow` field for the Pricing page section eyebrow.
- Updated Pricing page rendering so the eyebrow comes from Sub2API public settings when configured, with the previous local text as fallback.
- Updated admin Runtime Settings placeholders in English and Chinese to show the new configurable `labels.eyebrow` field.
- Added a Pricing page unit test covering public-settings-driven shell labels.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts progress.md`

### Notes
- This does not change pricing data ownership; catalog data and checkout already remain Sub2API-backed. It only removes another fixed UI-copy point from the Vue shell.

## 2026-06-18 Web platform integration doc and legacy route capability cleanup
### Done
- Renamed the active integration document from `docs/TOUCH_PLATFORM_INTEGRATION.md` to `docs/WEB_PLATFORM_INTEGRATION.md`.
- Reworded the document around the Sub2API-owned web platform instead of a Touch-facing standalone product surface.
- Updated route comments so public prompt routes are described as shared Sub2API web frontend routes.
- Extended legacy `/api/v1/touch/capabilities` output with `replacement_routes` pointing callers to `/api/v1/web/*`, `/api/v1/prompts/*`, and `/api/v1/admin/prompts/*`.

### Validation
- `rg -n "TOUCH_PLATFORM_INTEGRATION|Touch Platform Integration|Touch-facing|Touch prompt|Touch pages|Touch-specific|RegisterPromptRoutes registers public prompt catalog routes shared by Touch" README.md README_CN.md README_JA.md docs backend/internal/server/routes -S`
- `go test ./internal/server/routes -run 'TestTouchCapabilitiesRoute|TestWebRoutesExposeGenericPrimaryAndLegacyTouchAlias' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/server/routes`
- `git diff --check -- backend/internal/server/routes/touch.go backend/internal/server/routes/touch_test.go backend/internal/server/routes/prompts.go docs/WEB_PLATFORM_INTEGRATION.md`

### Notes
- Legacy `/api/v1/touch/*` routes remain compatibility aliases only. New work should use the generic `/api/v1/web/*`, `/api/v1/prompts/*`, and `/api/v1/admin/prompts/*` surfaces.

## 2026-06-18 Web auth source internals renamed
### Done
- Renamed handler-level auth-source helpers from Touch-specific names to Web-source names: `webAuthSource`, `ensureWebAuthSourceAllowed`, `isWebAuthSource`, and `verifyWebOAuthBridgeToken`.
- Updated web/OAuth handlers to use the generic web-source helpers while preserving the existing persisted source value `touch` for account-source separation.
- Changed auth-source bridge errors from `TOUCH_SOURCE_*` to `WEB_SOURCE_*`.
- Renamed related auth tests to Web-source terminology.

### Validation
- `go test ./internal/handler -run 'TestEmailOAuth|TestAuthHandlerAllowsWebSource|TestAuthHandlerRejectsWebOAuthBridgeToken' -count=1`
- `go test ./internal/handler -run 'TestEmailOAuth|TestAuthHandlerAllowsWebSource|TestAuthHandlerRejectsWebOAuthBridgeToken|TestWebSessionCookie|TestReadWebSessionCookie|TestClearWebSessionCookies|TestWebCheckoutPaymentSource' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/handler ./internal/server/routes`
- internal Touch auth-source helper residue scan
- `git diff --check`

### Notes
- The actual stored signup/provider source value remains `touch` intentionally because Touch-origin accounts must stay separated from native Sub2API accounts with the same email.
- The external OAuth bridge query parameter `touch_bridge_token` remains for compatibility with existing bridge URLs.

## 2026-06-18 Web checkout default payment source made generic
### Done
- Changed the `/api/v1/web/payments/checkout` default payment source from `touch` to `sub2api_web`.
- Kept explicit frontend payment sources such as `hosted_redirect` and `wechat_in_app_resume` unchanged.
- Left the legacy `/api/v1/touch/admin/payments/checkout` bridge default as `touch` because it is a compatibility API.
- Added a handler test covering the generic web default and explicit source preservation.

### Validation
- `go test ./internal/handler -run 'TestWebSessionCookie|TestReadWebSessionCookie|TestClearWebSessionCookies|TestWebCheckoutPaymentSource' -count=1`
- `go test ./internal/server/routes -run 'TestWebRoutesExposeGenericPrimaryAndLegacyTouchAlias' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/handler ./internal/server/routes`
- frontend checkout payment source scan
- `git diff --check`

### Notes
- Existing frontend payment flow already sends generic/hosted payment source values and does not send `touch`.

## 2026-06-18 Web session cookies use generic Sub2API names
### Done
- Changed the primary browser web-session cookies from `touch_sub2api_access_token` / `touch_sub2api_refresh_token` to `sub2api_web_access_token` / `sub2api_web_refresh_token`.
- Kept read fallback for the old Touch cookie names so existing browser sessions can refresh or be logged out cleanly.
- Updated session set/clear paths to clear both generic and legacy cookie names, migrating active clients away from the old names.
- Added handler tests covering generic cookie writes, legacy fallback reads, and clearing both cookie name families.

### Validation
- `go test ./internal/handler -run 'TestWebSessionCookie|TestReadWebSessionCookie|TestClearWebSessionCookies' -count=1`
- `go test ./internal/server/routes -run 'TestWebRoutesExposeGenericPrimaryAndLegacyTouchAlias' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/handler ./internal/server/routes`
- cookie-name residue scan for primary vs legacy use
- `git diff --check`

### Notes
- Legacy cookie constants remain only for fallback reads and cleanup. The active `/api/v1/web/*` session writes generic Sub2API cookie names.

## 2026-06-18 OAuth popup compatibility component renamed
### Done
- Renamed the Vue OAuth popup compatibility page from `LegacyTouchAuthPopupView` to `AuthPopupView`.
- Changed the route name from `LegacyTouchAuthPopup` to `AuthPopup`.
- Kept `/auth-popup`, `/en/auth-popup`, and `/zh/auth-popup` compatibility URLs unchanged.
- Updated route and component tests to assert the generic route name while preserving legacy URL coverage.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/AuthPopupView.spec.ts src/router/__tests__/touch-oauth-compat.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- legacy popup component-name residue scan
- `git diff --check`

### Notes
- The route test filename still mentions touch compatibility because its purpose is to protect old URL aliases. The active component and route names are now generic.

## 2026-06-18 Runtime settings page copy made generic
### Done
- Removed the remaining visible Touch naming from the Runtime Settings page title in English and Chinese locales.
- Reworded the locale auto-detection hint so it refers to the public frontend instead of Touch.
- Confirmed the admin sidebar and primary route already use the generic Runtime Settings entry; `/admin/touch/settings` remains only as a legacy route alias.

### Validation
- Runtime Settings visible Touch copy scan across locale files and admin page
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check`

### Notes
- Legacy route aliases and compatibility component names still intentionally include Touch where they preserve old URLs or OAuth popup behavior.

## 2026-06-18 Admin runtime shell fields use generic settings fields
### Done
- Added generic admin settings aliases for core runtime shell fields: app metadata, theme/appearance/locale, auth visibility, prompt gallery text, workspace shell config, pricing shell text/config, and credits shell text.
- Updated admin settings PATCH handling so generic core runtime fields take precedence while legacy `touch_*` request fields remain accepted.
- Switched the Runtime Settings Vue form away from `form.touch_*` bindings for the core shell fields; save payload now uses generic fields.
- Kept frontend read fallback from legacy `touch_*` fields for older admin settings responses.
- Extended the admin handler runtime alias regression test to cover both core shell fields and integration fields.

### Validation
- `go test ./internal/handler/admin -run 'TestSettingHandler_UpdateSettings_AcceptsGenericRuntimeAliases' -count=1`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_ExposesTouchRuntimeSettings|TestSettingService_UpdateSettings_PersistsTouchRuntimeSettings' -count=1`
- `go test ./internal/handler ./internal/handler/dto ./internal/handler/admin -run 'TestSettingHandler_GetPublicSettings_ExposesGenericRuntimeSettingAliases|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|TestSettingHandler_UpdateSettings_AcceptsGenericRuntimeAliases' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler ./internal/handler/dto`
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check`

### Notes
- The persistent setting keys are still the existing `touch_*` keys. This step makes the active admin API/UI contract generic while preserving storage compatibility.

## 2026-06-18 Admin runtime integrations use generic settings fields
### Done
- Added generic admin settings aliases for analytics, ads, affiliate, and customer-support runtime integrations while keeping the existing `touch_*` response fields.
- Updated admin settings PATCH handling so generic fields take precedence and legacy `touch_*` fields remain accepted as compatibility fallback.
- Switched the Runtime Settings integrations sections to bind and submit generic fields: Google Analytics, Clarity, Plausible, OpenPanel, Vercel Analytics, AdSense, Affonso, PromoteKit, Crisp, and Tawk.
- Kept a frontend read fallback from legacy `touch_*` fields for older admin settings responses, but the save payload now uses the generic fields.
- Added an admin handler regression test covering generic integration aliases persisting into the existing runtime setting keys and returning both generic and legacy fields.

### Validation
- `go test ./internal/handler/admin -run 'TestSettingHandler_UpdateSettings_AcceptsGenericRuntimeAliases' -count=1`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_ExposesTouchRuntimeSettings|TestSettingService_UpdateSettings_PersistsTouchRuntimeSettings' -count=1`
- `go test ./internal/handler ./internal/handler/dto ./internal/handler/admin -run 'TestSettingHandler_GetPublicSettings_ExposesGenericRuntimeSettingAliases|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|TestSettingHandler_UpdateSettings_AcceptsGenericRuntimeAliases' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler ./internal/handler/dto`
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check`

### Notes
- Storage keys still use existing `touch_*` runtime settings for compatibility. The active admin API/UI path for integrations is now generic, matching the public settings path.

## 2026-06-18 Public integrations moved to generic settings aliases
### Done
- Added generic public settings aliases for analytics and marketing integrations: Google Analytics, Clarity, Plausible, OpenPanel, Vercel Analytics, AdSense, Affonso, PromoteKit, Crisp, and Tawk.
- Mapped those generic fields from the existing stored Touch runtime setting keys in Sub2API public settings and SSR public-settings injection.
- Updated `publicIntegrations` to prefer generic public settings fields and use `touch_*` fields only as legacy compatibility fallback.
- Updated frontend and backend tests to cover generic integration fields while preserving legacy fallback behavior.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_ExposesTouchRuntimeSettings' -count=1`
- `go test ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings_ExposesGenericRuntimeSettingAliases|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler ./internal/handler/dto`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/publicIntegrations.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check`

### Notes
- Admin/internal storage keys still use `touch_*`; this step moves the active public integration runtime to generic aliases without breaking old consumers.

## 2026-06-18 Public pages stopped reading Touch runtime aliases
### Done
- Removed active `cachedPublicSettings.touch_*` fallback reads from Prompt Catalog, Image Generator, Pricing, and Credits Vue pages.
- Public pages now consume the generic Sub2API public settings fields: `prompt_*`, `prompt_catalog_shell_config`, `workspace_shell_config`, `pricing_*`, and `credits_*`.
- Kept backend `touch_*` fields available for compatibility and admin persistence; active public Vue pages no longer depend on them.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- Public page scan for `cachedPublicSettings.touch_*`
- `git diff --check`

### Notes
- Admin Runtime Settings still writes existing `touch_*` storage keys while the backend mirrors them into generic public settings. A later cleanup can rename the admin/internal storage model once compatibility migration is planned.

## 2026-06-18 Image workspace draft key and shell copy made generic
### Done
- Added a shared image generator draft utility using `sub2api:image-generator:draft` as the primary sessionStorage key.
- Kept one-way legacy fallback for existing `touch:image-generator:draft` browser sessions and clear both keys on reset.
- Updated Prompt Catalog to save generator drafts through the shared utility instead of writing the old Touch key directly.
- Updated Image Generator to load/clear drafts through the shared utility.
- Extended `workspace_shell_config` examples and Image Generator parsing to cover catalog link, eyebrow, hero description, clear button, and back-to-catalog labels.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/imageGeneratorDraft.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- image-generator draft key residue scan
- `git diff --check`

### Notes
- A mistaken `pnpm --filter sub2api-frontend test -- imageGeneratorDraft.spec.ts` invocation entered Vitest watch mode and ran broad existing suites; it reported unrelated pre-existing failures in account usage, usage view, dashboard, and payment/catalog tests. The watch process was stopped and the targeted non-watch test passed.

## 2026-06-18 Prompt catalog shell copy moved to public settings
### Done
- Added `prompt_catalog_shell_config` as a generic Sub2API runtime setting exposed through admin settings, public settings, and SSR public-settings injection.
- Wired the Vue admin Runtime Settings page so admins can edit the prompt catalog shell JSON from Sub2API.
- Updated Prompt Catalog to merge locale-scoped `prompt_catalog_shell_config.labels` over its built-in fallback copy.
- Covered stats labels, filters, pagination, empty/error states, detail actions, copy/generate labels, and X import panel copy through the new labels surface.
- Kept catalog data, filtering, pagination, facets, and import behavior unchanged.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_ExposesTouchRuntimeSettings|TestSettingService_UpdateSettings_PersistsTouchRuntimeSettings' -count=1`
- `go test ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings_ExposesGenericRuntimeSettingAliases|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler ./internal/handler/dto`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check`

### Notes
- Prompt Catalog layout and UI state still live in Vue; this step moved configurable copy into Sub2API public settings.

## 2026-06-18 Pricing shell labels moved to public settings
### Done
- Extended `pricing_shell_config` parsing on the Vue pricing page to accept locale-scoped `labels`.
- Moved catalog status, product counts, CTA, error/empty states, recommended badge, balance/rate/quota, validity unit, and unlimited labels behind Sub2API public settings overrides.
- Kept built-in locale copy as the fallback when `pricing_shell_config` is empty or invalid.
- Updated admin runtime setting placeholders/hints to document the new `labels` surface.
- Removed the stale `[touch-pricing]` console tag from the public pricing catalog loader.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- Pricing local-copy residue scan for moved labels and old console tag
- `git diff --check`

### Notes
- Pricing layout and interaction remain in the Vue shell; this step only moves more page copy/control text to Sub2API public settings.

## 2026-06-18 Public Vue component names removed Touch prefix
### Done
- Renamed public `TouchImageGeneratorView.vue` to `ImageGeneratorView.vue` and updated router imports.
- Renamed public `TouchPricingView.vue` to `PricingView.vue` and updated router imports.
- Renamed user `TouchCreditsView.vue` to `CreditsView.vue` and updated the route name/import.
- Removed visible `Touch Pricing`, `Touch Prompt Workspace`, and `Touch Credits` copy from these migrated Vue pages.
- Renamed front-end credit conversion state from `touchCredits` / `TOUCH_CREDITS_PER_SUB2API_BALANCE` to generic `credits` / `CREDITS_PER_SUB2API_BALANCE`.

### Validation
- `pnpm run frontend:lint:check`
- `pnpm run frontend:typecheck`
- public Touch component/name residue scan

### Notes
- Remaining `TouchSettingsView` and `LegacyTouchAuthPopupView` names are still intentional compatibility/admin surfaces.

## 2026-06-18 Public Touch Vue routes made generic
### Done
- Changed the primary prompt catalog route from `/touch/prompts` to `/prompts`, keeping `/touch/prompts` as a Vue alias.
- Changed the primary image generator route from `/touch/generator` to `/image-generator`, keeping `/touch/generator` as a Vue alias.
- Changed the pricing route to use `/pricing` as the primary path and `/touch/pricing` as the alias.
- Updated old Touch URL redirects so `/prompts*` and `/ai-image-generator` land on generic Vue routes.
- Updated sitemap generation to advertise `/prompts` and `/image-generator` instead of `/touch/*` URLs.

### Validation
- `go test ./internal/web -run TestLegacyTouchRedirectMiddleware -count=1`
- `go test ./internal/handler -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/web ./internal/handler ./internal/server/routes`
- `pnpm run frontend:typecheck`
- active `/touch/prompts` and `/touch/generator` reference scan

### Notes
- `/touch/prompts` and `/touch/generator` remain as Vue aliases for compatibility.

## 2026-06-18 Prompt admin import moved off Touch route
### Done
- Added generic Sub2API admin prompt routes under `/api/v1/admin/prompts`.
- Moved the Vue prompt import client from `/api/v1/touch/admin/prompts/import-twitter` to `/api/v1/admin/prompts/import-twitter`.
- Added route coverage for generic admin prompt upsert at `/api/v1/admin/prompts/cases`.
- Kept the old `/api/v1/touch/admin/prompts/*` routes as compatibility-only while active Vue no longer depends on them.
- Updated Touch platform integration docs to mark Touch admin prompt routes as legacy compatibility.

### Validation
- `go test ./internal/server/routes -run 'TestPrompt|TestTouchPrompt' -count=1`
- `pnpm run frontend:typecheck`
- active route scan for `/touch/admin/prompts/import-twitter`

### Notes
- Historical `progress.md` entries still mention the old Touch prompt admin routes as prior migration steps.

## 2026-06-18 Touch Next subapp retired
### Done
- Removed `apps/touch` entirely after the remaining Next shell had been reduced to a placeholder build target.
- Removed Touch from the pnpm workspace and root package scripts; `build:web`/`test:web` now run the Sub2API Vue frontend only.
- Removed `build-touch` and `test-touch` from the Makefile; root `make build` and `make test` now cover Go backend plus Vue frontend/admin.
- Replaced the old Touch platform integration plan with the current Sub2API-owned state and verification entry points.
- Refreshed the root pnpm lockfile so Touch is no longer an importer.

### Validation
- `test ! -d apps/touch`
- active Touch Next build/runtime reference scan
- `pnpm install --lockfile-only`
- `pnpm run frontend:test`
- `pnpm run frontend:build`
- `go test ./internal/web ./internal/server -run 'Test|^$' -count=1`
- `make build`
- `PATH="/Users/aias/go/bin:$PATH" make test`
- `git diff --check`

### Notes
- Historical progress entries still mention `apps/touch` as prior work; those are archival, not active runtime/build ownership.
- Rebuilt `golangci-lint` v2.7.2 with local Go 1.26.3 under `/Users/aias/go/bin` so backend lint can run against the project's Go 1.26.3 target.

## 2026-06-18 Touch compatibility residue reduced
### Done
- Removed the unused Touch-side Sub2API web client helper after confirming no runtime imports.
- Removed leftover standalone deploy files: `apps/touch/.dockerignore`, `apps/touch/.vercelignore`, and `apps/touch/public/_headers`.
- Removed the duplicate Touch-local `apps/touch/tools/x-atuo` copy; Sub2API now owns the vendored runtime at repository-root `tools/x-atuo`.
- Updated ignore rules and Touch docs to keep x-atuo/deployment ownership outside the Touch compatibility shell.

### Validation
- stale active-reference scan
- `pnpm run touch:test`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch still needs the OAuth callback/popup compatibility runtime until those responsibilities move to Sub2API backend/Vue.

## 2026-06-18 Touch historical product docs removed
### Done
- Removed old `apps/touch/docs/superpowers` plans/specs for the pre-Sub2API Touch product, WeChat export, Hot integration, and template cleanup.
- Kept `apps/touch/docs/sub2api-integration.md` as the current Touch compatibility-shell integration document.

### Validation
- docs residue scan
- `pnpm run touch:test`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Old progress history still mentions the deleted docs as past work, but active Touch docs no longer preserve the obsolete product plans.

## 2026-06-18 Touch CSS theme minimized
### Done
- Replaced the old full product theme token set with the minimal colors/font tokens used by the remaining OAuth and not-found compatibility pages.
- Removed unused dark-mode, container, form, docs-table, scrollbar, and view-transition CSS from the Touch compatibility shell.

### Validation
- old-theme residue scan
- `pnpm run touch:test`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch still uses Tailwind/PostCSS because the remaining compatibility pages are styled with Tailwind utility classes.

## 2026-06-18 Legacy Touch URL redirects moved to Sub2API backend
### Done
- Added Sub2API web middleware for old Touch public URL redirects before Vue SPA fallback.
- Covered localized routes, safe auth callback forwarding, cache headers, and non-GET passthrough with backend tests.
- Removed the Touch Next proxy, legacy redirect resolver, and matching Touch test.
- Updated docs and Touch test script so route compatibility is owned by Sub2API backend, not Next.

### Validation
- `go test ./internal/web -run 'TestLegacyTouchRedirectMiddleware' -count=1`
- `go test ./internal/web ./internal/server -run 'Test|^$' -count=1`
- `pnpm run touch:test`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch still owns OAuth callback/popup handoff pages until those are replaced by Sub2API Vue/backend handling.

## 2026-06-18 Legacy Touch OAuth callback/popup moved to Sub2API Vue
### Done
- Added Vue routes for old `/auth-callback`, localized auth callback aliases, `/auth-popup`, and localized auth popup aliases.
- Added a Vue popup compatibility view that safely forwards `callbackUrl` into `/login?redirect=...`.
- Removed Touch OAuth pages, redirect components, public config/base-url helpers, and their node tests.
- Removed Touch runtime env docs and reduced `touch:test` to typecheck plus lint.

### Validation
- Pending.

### Notes
- Touch still exists as a minimal Next build shell until scripts/deployment no longer reference it.

## 2026-06-18 Touch formatter tooling simplified
### Done
- Removed Touch-local Prettier scripts and config files.
- Removed Touch-local Prettier dependencies, including Tailwind class sorting.
- Removed stale ESLint `.open-next` ignore entry.
- Refreshed the workspace lockfile.

### Validation
- `pnpm install --lockfile-only`
- stale Touch formatter/import-sort config scan
- `pnpm run touch:test`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch keeps lint/typecheck/test/build scripts; formatting can be handled by repository-level tooling or editor defaults.

## 2026-06-18 Touch Next config simplified
### Done
- Removed obsolete Touch bundle analyzer wrapper and `@next/bundle-analyzer` dependency.
- Removed standalone output, image optimizer settings, default page extension override, empty redirects hook, and unused Turbopack alias scaffold from `next.config.mjs`.
- Kept only the monorepo Turbopack root, OAuth popup COOP header, React compiler, and current Turbopack cache experiment.
- Refreshed the workspace lockfile.

### Validation
- `pnpm install --lockfile-only`
- stale analyzer/image/standalone Next config scan
- `pnpm run touch:test`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch remains a Next compatibility shell, but its Next config no longer carries settings for the removed full product app.

## 2026-06-18 Touch standalone docs/workflow cleanup
### Done
- Removed the obsolete Touch subdirectory GitHub Docker build workflow and VSCode i18n settings.
- Removed stale `preview-touch` / `deploy-touch` Make targets that pointed to deleted Cloudflare/OpenNext scripts.
- Trimmed Touch `.env.example` to the only remaining bootstrap values: `SUB2API_BASE_URL` and optional `NEXT_PUBLIC_SUB2API_BASE_URL`.
- Rewrote Touch README, CLAUDE.md, and Sub2API integration docs to describe the current compatibility-shell scope instead of the old standalone product app.

### Validation
- stale standalone deployment/docs scan
- `pnpm run touch:test`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- The remaining mention of Docker/Vercel/Cloudflare/OpenNext is an explicit note that those Touch standalone deployment entrypoints no longer exist.

## 2026-06-18 Touch standalone Docker/Vercel runtime removed
### Done
- Removed the Touch standalone Dockerfile and Vercel config.
- Removed the Touch `start` script plus obsolete standalone runner/env wrapper scripts.
- Removed the now-unused `dotenv-cli` dev dependency and refreshed the workspace lockfile.
- Removed stale standalone deploy ignores from Touch `.dockerignore`.

### Validation
- `pnpm install --lockfile-only`
- active standalone Docker/Vercel/start/dotenv scan
- `pnpm run touch:test`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Historical mentions in `apps/touch/progress.md` remain as archival records only.
- Touch still has `dev`, `build`, `lint`, and test scripts because it remains a Next compatibility subapp during the migration.

## 2026-06-17 Touch Cloudflare standalone deploy entry removed
### Done
- Removed root `touch:cf:*` scripts and Touch package `cf:*` scripts.
- Removed the Touch `wrangler` devDependency and refreshed the workspace lockfile.
- Deleted the obsolete Touch `wrangler.toml.example`.
- Removed stale wrangler/OpenNext ignore entries from Touch `.gitignore`, `.dockerignore`, and `.vercelignore`.

### Validation
- `pnpm install --lockfile-only`
- active Cloudflare/OpenNext entry scan across package scripts, ignore files, and lockfile
- `pnpm run touch:test`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Historical mentions in `apps/touch/progress.md` and old planning docs remain as archival records only; no active Touch Cloudflare deploy entry remains.

## 2026-06-17 Touch branding/runtime config removed from compatibility shell
### Done
- Made the Touch root layout static and removed runtime Sub2API branding/locale fetches from the layout.
- Simplified the Touch not-found page so it no longer fetches app name/logo or depends on Next image optimization.
- Reduced Touch `getPublicConfigs()` to only validate Sub2API public settings and expose `sub2api_public_base_url` for OAuth/legacy redirects.
- Removed Touch-side app URL/name/description/logo/favicon/default-locale env fallbacks from the compatibility shell.
- Deleted the obsolete public-site branding helper and test.
- Removed the now-unused Touch `logo.svg`; `apps/touch/public` now only keeps `_headers` and `favicon.svg`.
- Marked OAuth compatibility pages as dynamic so production builds do not prerender runtime-only Sub2API settings checks.

### Validation
- `pnpm run touch:test`
- `pnpm run touch:build`
- stale Touch branding/config helper scan
- `git diff --check`

### Notes
- Touch compatibility runtime config now only covers the Sub2API frontend base URL needed for redirect targets.
- Sub2API Vue/public settings own active branding, locale, SEO, and public page presentation.

## 2026-06-17 Unused Touch public assets removed
### Done
- Removed the old Touch public `imgs/` asset tree that belonged to deleted landing/showcase/payment page shells.
- Removed obsolete Touch `preview.png` and `preview.svg` assets after preview metadata moved out of the active shell.
- Removed the stale `/imgs/:path*` Next cache header and the matching `_headers` block.
- Kept only the remaining Touch shell static fallbacks: `_headers`, `logo.svg`, and `favicon.svg`.

### Validation
- `pnpm run touch:test`
- `pnpm run touch:build`
- stale Touch static asset reference scan
- `git diff --check`

### Notes
- The remaining `/images/` scan hits are Sub2API/OpenAI gateway and page-image routes, not Touch public assets.

## 2026-06-17 Touch favicon.ico moved to Sub2API
### Done
- Added Sub2API root `GET /favicon.ico`, backed by `touch_app_favicon` public settings with a Vue static `/favicon.svg` fallback.
- Added handler coverage for configured favicon redirects and fallback behavior.
- Updated embedded frontend bypass rules so `/favicon.ico` reaches backend routing instead of the SPA fallback.
- Added `frontend/public/favicon.svg` so the fallback favicon is shipped with the Sub2API Vue frontend.
- Removed the Touch Next `favicon.ico` redirect route.

### Validation
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_Get(AdsTxt|FaviconICO|RobotsTxt|SitemapXML|PublicSettings|SiteLogo)' -count=1`
- `go test ./internal/server ./internal/web -run 'Test|^$' -count=1`
- `pnpm run touch:test`
- `pnpm run frontend:build`
- `pnpm run touch:build`
- stale Touch root metadata/static route scan
- `git diff --check`

### Notes
- Touch build route output no longer includes `/favicon.ico`, `/robots.txt`, `/sitemap.xml`, or `/ads.txt`.
- The remaining Touch Next routes are compatibility root/catch-all, not-found, and OAuth redirect pages.

## 2026-06-17 Touch robots/sitemap moved to Sub2API
### Done
- Added Sub2API root `GET /robots.txt` and `GET /sitemap.xml`, backed by Touch public settings for base URL and default locale.
- Added handler coverage for configured Touch app URL, request-origin fallback, and localized sitemap output.
- Updated embedded frontend bypass rules so `/robots.txt` and `/sitemap.xml` reach backend routing instead of the SPA fallback.
- Removed the Touch Next `robots.ts` and `sitemap.ts` metadata routes.

### Validation
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_Get(AdsTxt|RobotsTxt|SitemapXML|PublicSettings|SiteLogo)' -count=1`
- `go test ./internal/server ./internal/web -run 'Test|^$' -count=1`
- `pnpm run touch:test`
- `pnpm run touch:build`
- stale Touch robots/sitemap route scan
- `git diff --check`

### Notes
- Touch build route output no longer includes `/robots.txt` or `/sitemap.xml`; both root SEO files are now Sub2API responsibilities.
- Touch still retains the favicon redirect and OAuth compatibility routes.

## 2026-06-17 Touch ads.txt moved to Sub2API
### Done
- Added Sub2API root `GET /ads.txt`, backed by `touch_adsense_code` from public settings.
- Added handler coverage for configured and unset ads.txt behavior.
- Updated embedded frontend bypass rules so `/ads.txt` reaches the backend route instead of the SPA fallback.
- Removed the Touch Next `/ads.txt` route.
- Removed `adsense_code` and old UI-only fields from Touch public config normalization: prompt gallery labels, auth visibility flags, locale detection, theme/appearance, and preview image.
- Deleted the now-unused Touch settings-name helper.

### Validation
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_Get(AdsTxt|PublicSettings|SiteLogo)' -count=1`
- `go test ./internal/server ./internal/web -run 'Test|^$' -count=1`
- `pnpm run touch:test`
- `pnpm run touch:build`
- stale Touch ads/config field scan
- `git diff --check`

### Notes
- Touch build route output no longer includes `/ads.txt`; the root ad verification file is now a Sub2API responsibility.
- Touch public config is now limited to fields consumed by the compatibility shell: Sub2API frontend base URL, app URL/name/description/logo/favicon, and locale.

## 2026-06-17 Touch compatibility shell tool residue removed
### Done
- Removed `nextjs-toploader` from the Touch root layout and package because the remaining compatibility shell no longer has navigable local UI.
- Removed the unused `tailwindcss-animate` plugin and dependency; the remaining OAuth spinner uses Tailwind's built-in `animate-spin`.
- Deleted stale Fumadocs `.source` generated files and the empty `.source` directory.
- Deleted obsolete `components.json` shadcn configuration after all shadcn/Radix UI wrappers were removed.
- Removed stale `.source` aliases/ignore rules from Touch `tsconfig`, eslint config, and `.gitignore`.

### Validation
- `pnpm run touch:typecheck`
- `pnpm run touch:test`
- `pnpm run touch:build`
- stale tool/dependency/source scan across Touch config, source, package, and lockfile
- `git diff --check`

### Notes
- Touch direct runtime dependencies are now down to `next`, `react`, and `react-dom`.
- Touch build/lint still emits the existing stale `baseline-browser-mapping` warning.

## 2026-06-17 Touch provider and UI shell removed
### Done
- Removed the remaining React provider shell from Touch root layout.
- OAuth compatibility pages now fetch Sub2API public settings on the server and pass them directly to tiny client redirect components.
- Simplified not-found to plain Next/Image/Link markup without shared Button/SmartIcon wrappers.
- Deleted obsolete AppContext, theme provider, toaster, button, smart-icon, class merge helper, unused hooks, unused browser/API utility libs, and stale block type declarations.
- Removed now-unused Touch direct dependencies: Radix slot, class-variance-authority, clsx, lucide-react, next-themes, sonner, and tailwind-merge.

### Validation
- `pnpm run touch:typecheck`
- `pnpm run touch:test`
- `pnpm run touch:build`
- stale provider/UI/hook/lib/dependency scan across Touch source, tests, and package
- `git diff --check`

### Notes
- Touch source is now limited to compatibility routes, OAuth redirect components, public settings helpers, legacy redirect helpers, and static metadata endpoints.
- Touch build/lint still emits the existing stale `baseline-browser-mapping` warning.

## 2026-06-17 Touch Next i18n shell removed
### Done
- Removed `next-intl` from the Touch compatibility shell.
- Replaced i18n middleware with a plain proxy pass-through after legacy route redirects.
- Added unlocalized `/auth-callback` and `/auth-popup` compatibility pages so OAuth redirects no longer depend on locale middleware rewrites.
- Extracted shared OAuth redirect client components for localized and unlocalized compatibility routes.
- Moved AppContext/theme/toaster providers to the root layout so all remaining compatibility routes receive Sub2API public settings.
- Reduced the locale layout to locale validation only.
- Deleted the old Touch i18n request/navigation config, translated metadata helper, and locale message JSON files.
- Removed the `next-intl` dependency from the Touch package and lockfile.

### Validation
- `pnpm run touch:typecheck`
- `pnpm run touch:test`
- `pnpm run touch:build`
- stale `next-intl`/i18n/SEO scan across Touch source, tests, package, config, and lockfile
- `git diff --check`

### Notes
- Touch Next is now a thinner runtime compatibility shell: OAuth hash-preserving redirects, metadata endpoints, and legacy proxy redirects remain.
- Root layout is intentionally dynamic because production public settings must come from Sub2API at runtime.
- Touch build/lint still emits the existing stale `baseline-browser-mapping` warning.

## 2026-06-17 Touch legacy routes centralized in proxy
### Done
- Added a pure legacy-route redirect resolver for the remaining Touch compatibility URLs.
- Moved legacy redirects for home, pricing, prompts, generator, settings, auth, docs, and legal pages into the Touch proxy before locale rendering.
- Removed the Touch-side settings cookie gate so settings navigation goes straight to Sub2API Vue, where auth/session ownership now lives.
- Deleted redundant old Next route page/layout files for landing, pricing, prompts, settings, docs, and auth pages.
- Added minimal root and localized catch-all compatibility pages so the slim Next app still satisfies App Router build requirements.
- Removed unused Touch common UI blocks, Radix UI wrappers, content-block mapping, and the now-unused Touch dependencies that only supported deleted page shells.
- Updated the root `touch:test` script to include the new legacy redirect tests.
- Fixed Touch Turbopack build under the pnpm monorepo by setting `turbopack.root` to the Sub2API workspace root, not `apps/touch`.

### Validation
- `pnpm run touch:typecheck`
- `pnpm run touch:test`
- `pnpm run touch:build`
- stale deleted Touch UI/dependency scan
- `git diff --check`

### Notes
- Touch Next now builds as a small compatibility/OAuth shell; active public pages route to Sub2API Vue.
- Touch build/lint still emits the existing stale `baseline-browser-mapping` warning.

## 2026-06-17 Touch runtime settings admin panel moved into Sub2API Vue
### Done
- Added a dedicated Sub2API Vue admin page at `/admin/touch/settings` for the retained Touch runtime settings.
- Wired the page into the admin router and sidebar as `Touch Settings` / `Touch 配置`.
- The new page loads from Sub2API admin settings, saves only Touch runtime fields as a partial settings payload, and refreshes public settings cache after save.
- Covered the new page with a focused unit test that verifies it does not submit unrelated global settings.
- Removed obsolete Touch locale copy for settings that were already deleted from runtime: prompt-gallery JSON, landing/docs/static JSON, old workspace page title/description, settings page labels, and auth-shell copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/TouchSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `pnpm run frontend:test:critical`
- `pnpm run frontend:build`
- stale Touch locale key scan for deleted settings
- `git diff --check`

### Notes
- Vue build still emits existing dynamic-import/chunk-size and stale Browserslist warnings.
- Existing `SettingsView` tests still emit the pre-existing `router-link` component warning in the test environment.

## 2026-06-17 Obsolete Touch page-shell settings pruned
### Done
- Removed dead Sub2API settings fields that only supported the old Touch Next page shell:
  - prompt-gallery label/page JSON
  - landing/docs/static page JSON
  - workspace/settings/auth title and copy fields
- Kept the active Touch runtime settings that are still consumed by the Vue frontend and compatibility shell:
  - app branding
  - theme/default locale
  - prompt section titles
  - pricing/credits shell fields
  - auth provider visibility
  - public integrations
- Restored `frontend/src/views/admin/SettingsView.vue` to the repository baseline after an overly broad mechanical cleanup damaged its template, then kept the type/API cleanup scoped to settings DTOs and public/admin config surfaces.

### Validation
- `go test ./internal/service ./internal/handler ./internal/handler/admin`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `pnpm run frontend:test:critical`
- `pnpm run frontend:build`
- `pnpm run touch:test`
- `pnpm run touch:build`
- obsolete-field scan for removed Touch settings in `frontend/src` and `backend/internal`
- `git diff --check`

### Notes
- Active Touch runtime settings are still present in Sub2API API/types.
- The admin UI is not yet a dedicated slim Touch config panel for only the retained Touch fields; that is the next sensible place to continue if admins need direct Vue-side editing for these settings.
- Vue build still emits existing dynamic-import/chunk-size and stale Browserslist warnings; Touch build/lint still emits the existing stale `baseline-browser-mapping` warning.

## 2026-06-17 Touch third-party integrations moved to Sub2API Vue
### Done
- Added Sub2API Vue public integration injection:
  - Google Analytics
  - Microsoft Clarity
  - Plausible
  - OpenPanel
  - Vercel Analytics script
  - Google Adsense
  - Affonso
  - PromoteKit
  - Crisp
  - Tawk
- Wired the injector into the Vue root app so these scripts are driven by Sub2API public settings instead of the Touch Next root layout.
- Removed Touch Next page-level third-party script injection from `apps/touch/src/app/layout.tsx`.
- Deleted Touch React/Next provider chains for ads, affiliate, analytics, customer-service, and UTM capture.
- Removed `@vercel/analytics` from the Touch package and updated the root lockfile.
- Reduced Touch public config normalization so script-only settings are no longer exposed through the Next shell. `adsense_code` remains for the Touch `ads.txt` compatibility endpoint.
- Updated the root `touch:test` script so it runs the currently retained Touch compatibility/runtime tests instead of deleted legacy prompt/pricing tests.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/publicIntegrations.spec.ts`
- `pnpm --filter touch exec node --import tsx --test tests/public-config.test.ts`
- `pnpm run frontend:typecheck`
- `pnpm run touch:typecheck`
- `pnpm run frontend:lint:check`
- `pnpm run touch:test`
- `pnpm run frontend:test:critical`
- `pnpm run frontend:build`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Vue build still emits existing dynamic-import/chunk-size and stale Browserslist warnings.
- Touch build/lint still emits the existing stale `baseline-browser-mapping` warning.
- Touch still keeps `ads.txt` as a Next compatibility route, so Adsense public config remains readable by the Touch shell for that file only.

## 2026-06-17 Touch locale and prompt catalog frontend logic removed
### Done
- Reduced Touch locale loading to the only active namespace, `common`.
- Deleted old locale JSON for redirected/removed Touch pages:
  - landing
  - settings sidebar/credits
  - AI image generator
  - pages index/pricing/prompts
  - prompts sidebar
- Removed Touch-side prompt catalog frontend logic that is no longer rendered after `/touch/prompts` moved to Sub2API Vue:
  - prompt catalog normalization/model helpers
  - prompt gallery labels/page config parsers
  - prompt gallery remote fetch/import client
  - prompt gallery local view-model/filter/stat helpers
  - Sub2API prompt catalog fetch helper used only by the deleted Touch-side UI/tests
- Removed corresponding Touch tests; active prompt catalog behavior is now covered by Sub2API backend/Vue code paths rather than the old Touch Next shell.
- Removed Touch public-config normalization for `touch_prompt_gallery_labels_config` and `touch_prompt_gallery_page_config`. Sub2API backend/Vue/admin settings still own those fields.

### Validation
- `rg -n "landing|settings/(sidebar|credits)|ai/image|pages/(index|pricing|prompts)|prompts/sidebar|useTranslations\\(['\\\"](landing|settings|ai|pages|prompts)|getTranslations\\(['\\\"](landing|settings|ai|pages|prompts)|metadataKey: ['\\\"](landing|settings|ai|pages|prompts)" apps/touch/src apps/touch/tests -S`
- `rg -n "prompt-gallery|prompt_gallery|prompt_catalog|sub2api-touch|PromptCase|normalizePromptCases|listPrompt|fetchPromptGallery|importPromptGallery|filterPromptCases|buildPrompt|PromptGalleryLabels|touch_prompt_gallery" apps/touch/src apps/touch/tests apps/touch/package.json -S`
- `pnpm --dir apps/touch exec node --import tsx --test $(find apps/touch/tests -maxdepth 1 -name '*.test.ts' -print | sed 's#^apps/touch/##' | sort)`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch tests are now reduced to the remaining compatibility/runtime shell helpers: auth redirect safety, public config, public site config, and Sub2API frontend URL resolution.
- `pnpm run touch:build` still emits the known `baseline-browser-mapping` stale-data warning.

## 2026-06-17 Touch legacy UI shell pruned after Vue handoff
### Done
- Removed inactive Touch React UI shells that were no longer reachable after the routes were redirected to Sub2API Vue:
  - generator workspace blocks
  - payment modal/provider blocks
  - console layout/sidebar helper
  - panel card wrapper
  - prompt gallery React component shell
  - SignUser/auth-shell blocks
- Simplified `AppContextProvider` to only provide runtime public configs; removed local user hydration, refresh, credits fetch, and sign-check state from the Touch Next shell.
- Removed unused Touch-side user/RBAC/auth client code that only supported the deleted local header/sign UI.
- Removed Touch-side normalization for `touch_workspace_*`, `touch_credits_*`, `touch_settings_*`, and `touch_auth_*` public settings. Sub2API backend/Vue still owns those settings for the unified frontend.
- Removed unused Touch UI wrappers and dependencies left behind by the deleted shells:
  - `accordion`, `avatar`, `dialog`, `drawer`, `label`, `navigation-menu`, `sheet`
  - `badge`, `card`, `input`, `textarea`, `highlighter`, `scroll-animation`
  - package dependencies for the removed Radix packages and `vaul`
- Removed the last Touch-side credit conversion helper/test; the active conversion now lives in the Sub2API Vue credits page.

### Validation
- `rg -n "@radix-ui/react-(accordion|avatar|dialog|label|navigation-menu)|vaul|shared/components/ui/(accordion|avatar|dialog|drawer|label|navigation-menu|sheet|badge|card|input|textarea|highlighter|scroll-animation)|shared/blocks/(generator|payment|panel|console|sign)|workspace-shell-config|shared/models/user|sub2api-auth-client|core/rbac|touch_workspace_|workspace_shell_config|touch_credits_|touch_settings_|touch_auth_|auth_shell_config|TOUCH_CREDITS_PER_SUB2API_BALANCE|sub2apiBalanceToTouchCredits" apps/touch/src apps/touch/tests apps/touch/package.json pnpm-lock.yaml -S`
- `pnpm --dir apps/touch exec node --import tsx --test $(find apps/touch/tests -maxdepth 1 -name '*.test.ts' -print | sed 's#^apps/touch/##' | sort)`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch still has compatibility Next routes and shared prompt catalog normalization services, but the removed UI shells no longer render inside Touch.
- `pnpm run touch:build` still emits the known `baseline-browser-mapping` stale-data warning.

## 2026-06-17 Touch landing and static pages redirected to Sub2API Vue
### Done
- Replaced the legacy Touch localized root landing page with a server-side redirect to the Sub2API Vue `/home` route.
- Replaced the Touch landing catch-all renderer with a small compatibility redirect map for old legal/static slugs:
  - `/privacy-policy` -> `/legal/privacy-policy`
  - `/terms-of-service` and `/terms` -> `/legal/terms`
- Simplified the landing segment layout to a passthrough layout because active pages under it now redirect or 404.
- Updated legacy Touch landing/footer legal links to `/legal/privacy-policy` and `/legal/terms`.
- Removed Touch static page rendering support: `content/pages`, `static-pages-config`, `static-page` model, static-page tests, and the remaining Fumadocs source file.
- Removed `touch_static_pages_config` from Touch public config normalization and tests.
- Removed the Fumadocs MDX Next wrapper, MDX page extensions, `source.config.ts`, Dockerfile copy reference, and package `postinstall` source generation hook from Touch.
- Removed obsolete Touch pricing content service/tests and the remaining Touch-side `pricing_title` / `pricing_description` / `pricing_shell_config` public-config normalization; Sub2API Vue still owns `touch_pricing_*` settings for the unified pricing page.
- Removed the last unused Touch Fumadocs helpers (`mdx-components`, docs TOC, docs CSS), Fumadocs package dependencies, the MDX type package, and the `mdxRs` experiment flag.
- Removed the stale nested `apps/touch/pnpm-lock.yaml` and changed the Touch Dockerfile to build from the repository root pnpm workspace instead of the old standalone Touch package context.
- Updated robots so the old `/privacy-policy` and `/terms-of-service` paths are no longer advertised as local blocked pages.

### Validation
- `rg -n "source.config|content/pages|static_pages_config|touch_static_pages_config|getStaticPageFromConfig|getLocalPage|pagesSource|core/docs/source|from '@/core/docs/source'|fumadocs-mdx/next|postinstall.*fumadocs|/terms-of-service|/privacy-policy" apps/touch/src apps/touch/tests apps/touch/content apps/touch/package.json apps/touch/Dockerfile apps/touch/next.config.mjs .github Makefile README.md docs deploy -S`
- `rg -n "fumadocs-(core|mdx|ui)|fumadocs|@types/mdx|mdxRs|source\\.config|content/pages|static_pages_config|touch_static_pages_config|getPricingSectionForLocale|getStaticPricingSection|pricing-content|core/docs/toc|config/style/docs.css|mdx-components" apps/touch/src apps/touch/tests apps/touch/package.json apps/touch/next.config.mjs pnpm-lock.yaml -S`
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-config.test.ts tests/public-site-config.test.ts`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch still contains compatibility routes in the Next route table for localized root and catch-all paths, but they no longer render the old dynamic/static page shells.
- `@types/mdx` still appears in the root lock through the Vue frontend dependency graph, not through Touch.
- `pnpm run touch:build` still emits the known `baseline-browser-mapping` stale-data warning.

## 2026-06-17 Touch docs route redirected to Sub2API Vue
### Done
- Replaced legacy Touch Fumadocs `/[locale]/docs/[[...slug]]` rendering with a server-side redirect to the Sub2API Vue `/docs` route, preserving nested docs slug paths.
- Simplified the Touch docs layout to a passthrough layout and removed the old Fumadocs layout config.
- Updated Touch sitemap generation so `/docs` is published as an unlocalized Vue-owned route and Touch no longer enumerates `docsSource` pages.
- Removed the unused Touch docs shell/pages config parsers and their focused tests.
- Removed `content/docs` MDX files and reduced `source.config.ts` / `core/docs/source.ts` to keep only `content/pages` for legacy static/legal pages.
- Removed now-unused `docs_shell_config` / `docs_pages_config` mapping from Touch public config normalization and tests.

### Validation
- `rg -n "docs_shell_config|docs_pages_config|touch_docs_shell_config|touch_docs_pages_config|docsSource|source\\.getPage|content/docs|auth-entry-visibility|sign-modal|social-providers" apps/touch/src apps/touch/tests apps/touch/content apps/touch/source.config.ts -S`
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-config.test.ts tests/public-site-config.test.ts`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch still has a compatibility `/[locale]/docs/[[...slug]]` route in the Next route table, but it no longer renders Fumadocs or local docs MDX.
- `content/pages` remains because legacy static/legal pages still use `pagesSource`.
- `pnpm run touch:build` still emits the known `baseline-browser-mapping` stale-data warning.

## 2026-06-17 Touch auth modal and OAuth bridge cleanup
### Done
- Removed the legacy Touch sign-in/sign-up modal components, embedded sign forms, social provider popup launcher, and the auth entry visibility helper after confirming no active route imports them.
- Updated the Touch header `SignUser` unauthenticated action to link to the Sub2API Vue `/login` entry instead of opening the local Touch sign modal.
- Updated the legacy pricing block fallback to redirect unauthenticated users to Sub2API Vue `/login` instead of opening the removed modal.
- Added a small client-safe `resolveSub2APIFrontendHref` helper for browser-side links to the Sub2API Vue frontend.
- Changed legacy Touch `/auth-popup` into a compatibility redirect to Sub2API Vue `/login`.
- Changed legacy Touch `/auth-callback` into a compatibility redirect to Sub2API Vue `/auth/callback`, preserving query and hash payloads for the Vue callback handler.
- Removed the AppContext `isShowSignModal` / `setIsShowSignModal` state after deleting its consumers.

### Validation
- `rg -n "isShowSignModal|setIsShowSignModal|sign-modal|sign-in-form|sign-up-form|social-providers|auth-entry-visibility|/api/v1/touch/web/auth/oauth/session|/api/v1/touch/web/auth/oauth/.*/start|auth-callback-success" apps/touch/src apps/touch/tests -S`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch still has compatibility `/[locale]/auth-popup` and `/[locale]/auth-callback` routes in the Next route table, but they no longer complete a Touch-specific OAuth session.
- `pnpm run touch:build` still emits the known `baseline-browser-mapping` stale-data warning.

## 2026-06-17 Touch auth pages redirect to Sub2API Vue
### Done
- Replaced legacy Touch Next `/sign-in` page rendering with a server-side redirect to Sub2API Vue `/login`.
- Replaced legacy Touch Next `/sign-up` page rendering with a server-side redirect to Sub2API Vue `/register`.
- Preserved safe internal callback forwarding by translating Touch `callbackUrl` into Sub2API Vue `redirect` query params.
- Removed route-level dependence on Touch local `getSignUser` session checks and SignIn/SignUp page components; auth state handling is left to the Sub2API Vue auth guard.

### Validation
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `rg -n "SignIn|SignUp|getSignUser|sub2api-auth-client|/api/v1/touch/web/auth/login|/api/v1/touch/web/auth/register" 'apps/touch/src/app/[locale]/(auth)' apps/touch/src/shared/blocks/sign apps/touch/src/shared/services -S`

### Notes
- Shared sign modal/components still exist for remaining legacy Touch UI surfaces, but `/[locale]/sign-in` and `/[locale]/sign-up` no longer render them.
- `pnpm run touch:build` still emits the known `baseline-browser-mapping` stale-data warning.

## 2026-06-17 Touch legacy links routed to Sub2API Vue
### Done
- Updated legacy Touch landing/static page copy, prompt sidebar copy, and docs index links so Prompt Gallery actions target Sub2API Vue `/touch/prompts` instead of localized Touch prompt routes.
- Updated docs shell nav to target `/touch/prompts`, `/touch/generator`, and `/pricing`, while keeping the docs home URL localized.
- Updated Touch sitemap generation to publish localized home entries plus unlocalized Vue-owned `/pricing`, `/touch/prompts`, and `/touch/generator` entries instead of advertising migrated localized pages.
- Cleaned remaining legacy component fallbacks: prompt card generator handoff now navigates to `/touch/generator`, and old pricing checkout return URL now uses `/settings/credits`.

### Validation
- `rg -n "(/touch/touch|/ai-image-generator|/zh/prompts|/en/prompts|/zh/pricing|/en/pricing|/zh/settings/credits|/en/settings/credits)" apps/touch/src/app apps/touch/src/config/locale/messages apps/touch/content apps/touch/src/shared apps/touch/src/themes -S`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Remaining `/api/v1/prompts/cases` strings are Sub2API API endpoints, and remaining localized prompt route files are compatibility redirects to Vue.
- `pnpm run touch:build` still emits the known `baseline-browser-mapping` stale-data warning.

## 2026-06-17 Touch Next prompt pricing generator redirects
### Done
- Replaced the legacy Touch Next `/pricing` page with a server-side redirect to the Sub2API Vue `/pricing` route.
- Replaced the legacy Touch Next `/ai-image-generator` page and layout with redirects to the Sub2API Vue `/touch/generator` route.
- Replaced the legacy Touch Next `/prompts`, `/prompts/cases`, and `/prompts/templates` pages plus prompts layout with redirects to the Sub2API Vue `/touch/prompts` route.
- These legacy localized routes remain reachable, but no longer render the React Prompt Gallery, generator workspace, pricing block, or console/sidebar shells.

### Validation
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `rg -n "PromptGallery|LazyImageGenerator|getPricingSectionForLocale|DynamicPage|ConsoleLayout|setRequestLocale|getTranslations\\(" 'apps/touch/src/app/[locale]/(landing)/prompts' 'apps/touch/src/app/[locale]/(landing)/(ai)/ai-image-generator' 'apps/touch/src/app/[locale]/(landing)/pricing/page.tsx' -S`
- `rg -n "resolveSub2APIFrontendUrl\\('/touch/prompts'|resolveSub2APIFrontendUrl\\('/touch/generator'|resolveSub2APIFrontendUrl\\('/pricing'" 'apps/touch/src/app/[locale]/(landing)' -S`

### Notes
- Prompt/gallery/generator/pricing helper modules still exist because tests and older non-route code reference them. The active legacy route surface now redirects to Vue instead of rendering those modules.

## 2026-06-17 Touch Next settings credits redirect
### Done
- Replaced the legacy Touch Next `settings` segment shell with a server-side redirect to the Sub2API Vue `/settings/credits` route.
- Replaced the legacy Touch Next `/settings` page and `/settings/credits` page with redirects instead of rendering the local credits UI.
- Added `resolveSub2APIFrontendUrl` so legacy Touch routes can target `sub2api_public_base_url` / `SUB2API_BASE_URL` and normalize paths.
- The redirect helper requires a Sub2API base URL in production and falls back to a relative path only outside production.
- This removes the duplicate Touch Next credits display while preserving legacy route reachability.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/sub2api-frontend-url.test.ts`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `rg -n "PanelCard|formatBalanceLabel|fetchUserCredits|settings_credits_label|ConsoleLayout title=\\{title\\}|resolveSub2APIFrontendUrl" 'apps/touch/src/app/[locale]/(landing)/settings' apps/touch/src/shared/services apps/touch/tests -S`

### Notes
- `/[locale]/settings` and `/[locale]/settings/credits` still exist in the legacy Touch Next route table, but they no longer render local settings/credits UI; they redirect to the Sub2API Vue implementation.

## 2026-06-17 Sub2API Vue Touch credits page
### Done
- Added a Sub2API Vue `/settings/credits` route for the Touch credits balance page and `/touch/credits` as a compatibility alias.
- The route requires Sub2API authentication but no longer requires payment to be enabled, because balance display should remain available even when checkout is disabled.
- The page uses the Sub2API auth store user balance as the only balance source and converts it with the existing Touch rule: `10 Touch credits = 1 Sub2API balance`.
- The page reads Sub2API public settings for `touch_credits_title`, `touch_credits_description`, `touch_credits_purchase_label`, and `touch_credits_balance_label`.
- The page routes balance actions to existing Sub2API `/purchase`, `/purchase?tab=recharge`, and `/orders` flows instead of Touch-local credit history or local credit mutation paths.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:build`
- `git diff --check`
- Browser smoke at `http://localhost:5174/settings/credits` while logged out redirected to `/login?redirect=/settings/credits`.
- Browser smoke with mocked Sub2API auth/public-settings responses rendered `12.5` Sub2API balance as `125` Touch credits and showed `/purchase`, `/purchase?tab=recharge`, and `/orders` links.

### Notes
- This migrates the Touch credits display shell. Detailed order/refund history remains the existing Sub2API `/orders` page.

## 2026-06-17 Sub2API Vue Touch pricing page
### Done
- Added a Sub2API Vue `/touch/pricing` public route for the Touch pricing shell.
- Added `/pricing` as an alias to the same Vue route so old Touch public pricing links can resolve inside the unified frontend.
- The page reads Sub2API public settings for `touch_pricing_title`, `touch_pricing_description`, and `touch_pricing_shell_config` button/group labels.
- The page reads `/api/v1/payment/public/catalog` through the existing `paymentAPI.getPublicCatalog()` client and renders recharge products plus subscription plans without any Touch-local pricing data source.
- Purchase CTAs route into the existing Sub2API `/purchase` payment flow: recharge uses `?tab=recharge`, subscription uses `?tab=subscription&group=<group_id>`.
- The route keeps pricing display inside the unified Vue frontend while avoiding duplicate checkout logic.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:build`
- `pnpm run frontend:typecheck` after adding the `/pricing` alias
- Browser smoke at `http://localhost:5174/touch/pricing`: route rendered and showed the explicit catalog-load failure state when the local Go backend was not running.
- Browser smoke with a mocked `/api/v1/payment/public/catalog` response: recharge card rendered `入门包`, subscription card rendered `GPT 开发包`, and purchase links resolved to `/purchase?tab=recharge` plus `/purchase?tab=subscription&group=11`.

### Notes
- This migrates the public pricing display shell, not the authenticated checkout page itself. Checkout remains owned by the existing Sub2API `/purchase` workflow.

## 2026-06-17 Sub2API Vue Touch generator workspace
### Done
- Added a Sub2API Vue `/touch/generator` route for the Touch image prompt workspace.
- Reused the existing `touch:image-generator:draft` sessionStorage contract so prompt catalog selections open inside the Vue frontend without relying on the legacy Next `/ai-image-generator` route.
- The Vue workspace reads Sub2API public `touch_workspace_shell_config` for title, prompt label, placeholder, draft notice, status, button copy, and length-warning text, with local safety copy only as fallback.
- The workspace supports imported draft display, prompt editing, copy prompt, clear draft, prompt length validation, and return to the Vue prompt catalog.
- Updated Vue `/touch/prompts` generator actions to navigate to `/touch/generator` instead of the legacy localized Touch route.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:build`
- `rg -n "TouchImageGenerator|/touch/generator|touch_workspace_shell_config|window\\.location\\.assign\\('/touch/generator'\\)|IMAGE_GENERATOR_DRAFT_KEY" frontend/src apps/touch/src -S`
- Browser smoke at `http://localhost:5174/touch/generator`: page rendered, imported-draft banner appeared after setting `touch:image-generator:draft`, and the textarea value was `test prompt from catalog`.
- `git diff --check`

### Notes
- This ports the existing prompt-prep/copy workspace behavior into Sub2API Vue. It does not add a new model-calling generation task flow.

## 2026-06-17 Sub2API Vue Prompt Catalog generator handoff
### Done
- Added a `Use in generator` / `去生图` action to Vue `/touch/prompts` cards and prompt detail dialogs.
- The action writes the same `touch:image-generator:draft` sessionStorage payload used by the legacy Touch Prompt Gallery.
- The handoff preserves prompt text, title, source prompt ID, and a Vue catalog source marker before navigating to the Vue `/touch/generator` route.
- This lets the Sub2API Vue catalog feed the Sub2API Vue generator workspace instead of the Touch Next shell.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:build`
- `rg -n "IMAGE_GENERATOR_DRAFT_KEY|openGenerator|sourcePromptId|ai-image-generator|generate:" frontend/src/views/public/PromptCatalogView.vue apps/touch/src/shared/blocks -S`
- `git diff --check`

### Notes
- This is a compatibility handoff, not a full generator migration. The next larger slice is to move the generator workspace itself into the Sub2API Vue frontend or expose a Sub2API-native generator route.

## 2026-06-17 Sub2API Vue Prompt Catalog admin import
### Done
- Added a typed Sub2API Vue prompt import client for `/api/v1/touch/admin/prompts/import-twitter`.
- Added an admin-only import panel to the Vue `/touch/prompts` catalog.
- The import panel currently supports X/Twitter post URLs, validates supported status URLs client-side, calls the Sub2API admin import API with the existing JWT admin auth path, and displays backend warnings.
- Successful imports clear the input, surface the imported case title, open the imported prompt in the detail dialog, and refresh the first server-backed catalog page.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:build`
- `git diff --check`

### Notes
- This moves the active X/Twitter prompt import UI into Sub2API Vue for Sub2API-native admins.
- The older Touch web-session import path still exists for the legacy Touch Next UI until that UI is fully retired.

## 2026-06-17 Sub2API Vue Prompt Catalog detail actions
### Done
- Added details and copy actions to the Sub2API Vue `/touch/prompts` catalog cards.
- Reused the existing Vue `BaseDialog` and `useClipboard` infrastructure instead of adding new UI/runtime dependencies.
- The details dialog now shows the primary prompt image, additional image thumbnails, full prompt text, source/category tags, all display/model/source tags, source link, and copy-prompt action.
- Card-level copy now uses the same Sub2API frontend toast/clipboard path as the existing model plaza copy action.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:build`
- `git diff --check`

### Notes
- This makes the Vue prompt catalog closer to the current Touch Prompt Gallery core read/copy workflow, but import/admin and image-generation actions still remain in the Touch Next UI.

## 2026-06-17 Touch Prompt Gallery Import State Helper
### Done
- Added `applyPromptGalleryImportResult` to `prompt-gallery-view-model.ts` to own the successful-import state transition.
- The helper now deduplicates the imported prompt, places it first, selects it, clears the source URL, carries import warnings, and formats the imported message.
- Updated `PromptGallery` to use the helper while preserving the previous functional `setCases` update so concurrent list changes are not overwritten by stale closures.
- Added focused coverage for import result deduplication, selected case, cleared input, warnings, and message output.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-view-model.test.ts tests/prompt-gallery-remote.test.ts tests/prompt-gallery-catalog.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Prompt Gallery has little business logic left in the component. Remaining work is mostly UI state/rendering or the larger move of Touch pages into the Sub2API Vue frontend.

## 2026-06-17 Touch Prompt Gallery Remote Service
### Done
- Added `prompt-gallery-remote.ts` to own Prompt Gallery browser-side Sub2API prompt list and X/Twitter import calls.
- Moved remote filter pagination, query parameter construction, Sub2API response normalization, prompt case normalization, summary normalization, import result normalization, warning filtering, and supported source URL validation out of the React component.
- Updated `PromptGallery` so it only decides when to fetch/import and how to apply the resulting state; endpoint paths and response shapes now live in the shared service.
- Added focused tests for supported import URLs, paginated remote filter requests, normalized summary/items, import request body, imported item normalization, and warning filtering.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-remote.test.ts tests/prompt-gallery-view-model.test.ts tests/prompt-gallery-catalog.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Prompt Gallery is now mainly UI state and rendering. Remaining migration choices are either extracting the final UI state machine pieces or starting the larger Touch pages into Sub2API Vue frontend consolidation.

## 2026-06-17 Touch Prompt Gallery Filtering View Model
### Done
- Expanded `prompt-gallery-view-model.ts` to own Prompt Gallery local filtering, imported-time fallback sorting, source value construction, summary/stat fallback calculation, category counts, and category ordering.
- Replaced the corresponding inline `PromptGallery` component calculations with shared helper calls, leaving the component closer to state orchestration and rendering.
- Added focused tests for backend-facet-first source values, server-backed stats, local stats fallback, local search/source/mode/image filtering, server-backed no-op filtering, category count/order fallback, and imported-time sorting behavior.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-view-model.test.ts tests/prompt-gallery-catalog.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Prompt Gallery is now mostly state orchestration, remote-fetch handling, import UI, and rendering. The next migration step should either extract the remaining remote-fetch/import interaction services or begin the larger Touch-into-Sub2API-Vue frontend consolidation.

## 2026-06-17 Touch Prompt Gallery View Model Service
### Done
- Extracted Prompt Gallery category/template grouping view-model logic from the React component into `prompt-gallery-view-model.ts`.
- Added shared helpers for backend-label-first category display, category/template facet label maps, and template group construction.
- The gallery component now delegates template group ordering/count/label handling to the shared service, making the remaining component code thinner and leaving a single replacement point for future Sub2API-returned view models.
- Added focused tests proving template grouping prefers backend facet order/label/count and falls back to local featured/count ordering when summary metadata is absent.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-view-model.test.ts tests/prompt-gallery-catalog.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Remaining Prompt Gallery work is mostly component interaction state and fallback rendering. The next larger migration step is to decide whether to start porting Touch pages into the Sub2API Vue frontend shell or keep shrinking Next UI modules first.

## 2026-06-17 Touch Prompt Template Group Metadata
### Done
- Changed Prompt Gallery template grouping to use structured group objects instead of raw `[category, items]` tuples.
- Template group ordering now prefers Sub2API `summary.template_groups` facet order whenever that metadata is available.
- Template group headings now use backend facet `label` when present, then fall back to configured Touch category labels.
- Template group count display now uses backend facet `count` when available instead of recomputing the heading count only from rendered items.
- Category badge/filter display now goes through a single backend-label-first helper, so future category/template facet labels from Sub2API will be consumed without more component rewrites.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-catalog.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Remaining Prompt Gallery local logic is mostly interaction state and fallback rendering. The larger remaining migration is still the Next subapp to Sub2API Vue frontend consolidation.

## 2026-06-17 Touch Prompt Source Facet Labels
### Done
- Extended Sub2API prompt source facets with an optional `label` field so prompt source display names can come from backend data instead of Touch-only `source_labels`.
- Updated the prompt repository summary query to aggregate non-empty `source_label` values alongside `source_project` counts.
- Threaded facet labels through `TouchPromptFacetCount`, prompt handler DTOs, and public prompt summary JSON.
- Updated Touch `PromptCatalogFacet` and Sub2API summary normalization to preserve facet labels without emitting `label: undefined` for unlabeled category/template facets.
- Changed Prompt Gallery source filter chips, source stat chips, case cards, and template cards to prefer backend `source_label` / summary facet labels, keeping Touch locale label maps only as fallback.
- Added route and Touch normalizer assertions for source facet labels.

### Validation
- `go test ./internal/server/routes -run 'TestPromptCasesPublicRouteFiltersAndSummary|TestPromptCasesPublicRoute' -count=1` passed from `backend/`.
- `go test ./internal/service ./internal/repository ./internal/handler ./internal/server/routes -count=1` passed from `backend/`.
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-catalog.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Continue moving Prompt Gallery category/template grouping presentation metadata into Sub2API response metadata, or start the larger Vue frontend consolidation plan once prompt view-model fallbacks are small enough.

## 2026-06-17 Touch Prompt Display Tags View Model
### Done
- Added backend-derived `DisplayTags` to `TouchPromptCase` so Sub2API owns the prompt detail tag view-model instead of Touch filtering raw tags in the browser.
- `deriveTouchPromptDisplayTags` now removes category duplicates, model tags, source project/label tags, author handles, `ui`, and source-style snake_case slugs while preserving useful style/scene tags.
- Exposed `display_tags` from the shared prompt case DTO used by public prompt list/detail and Touch import responses.
- Updated Touch prompt catalog types and Sub2API normalizers to consume `display_tags` / `displayTags`.
- Simplified the Prompt Gallery detail modal so it prefers backend `displayTags` and only falls back to raw tags/styles/scenes for older data.
- Added backend service/route coverage and Touch normalizer coverage for the new field.

### Validation
- `go test ./internal/service -run 'TestTouchPromptService(ListCasesDerivesDisplayTags|ListCasesInfersModelTags|ListCasesNormalizesInput)' -count=1` passed from `backend/`.
- `go test ./internal/server/routes -run 'TestPromptCasesPublicRouteFiltersAndSummary|TestPromptCasesPublicRoute' -count=1` passed from `backend/`.
- `go test ./internal/service ./internal/handler ./internal/server/routes -count=1` passed from `backend/`.
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-catalog.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Continue moving Prompt Gallery list/category/template grouping view-models into Sub2API response metadata, then reassess whether the remaining Touch shell work is best handled by Vue frontend consolidation.

## 2026-06-17 Touch Prompt Gallery Behavior Config
### Done
- Extended existing Sub2API-managed `touch_prompt_gallery_page_config` with a `behavior` section for Prompt Gallery interaction parameters.
- `behavior.initialVisibleCount`, `behavior.visibleIncrement`, and `behavior.remotePageSize` now control initial visible cards, load-more increments, and remote filter page size for cases/templates pages.
- Added bounded positive-integer handling in the Touch gallery so bad or excessive config values fall back or clamp safely.
- Updated Prompt Cases/Templates pages to pass parsed behavior config into the gallery component.
- Updated Sub2API admin settings zh/en placeholder and help text so operators can discover the behavior fields without adding another setting key.
- Added parser coverage for locale-scoped behavior overrides and invalid behavior fallback.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-page-config.test.ts tests/prompt-gallery-catalog.test.ts tests/public-config.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `pnpm run frontend:build` passed with existing Vite browserslist/dynamic-import/chunk-size warnings.
- `git diff --check` passed.

### Next
- Continue moving Prompt Gallery interaction/view-model decisions into Sub2API settings/API metadata, or start the larger Next-to-Sub2API-frontend consolidation.

## 2026-06-17 Touch Workspace Shell Settings
### Done
- Added `touch_workspace_shell_config` as a Sub2API public/admin JSON setting for Touch AI image workspace shell copy.
- Threaded the field through Sub2API setting constants, settings views, public settings payload, admin settings request/response DTOs, update persistence, defaults, Vue admin API types, and the Vue settings form.
- Added a Touch workspace shell parser with locale/default scoped JSON support and `{title}` template formatting for imported prompt drafts.
- Updated the Touch AI image generator to prefer `workspace_shell_config` for panel title, draft import copy, prompt label/placeholder/length warning, copy button, copy toast messages, workspace title/description, and status copy, while keeping local locale JSON as fallback.
- Added focused backend and Touch parser/public-config test coverage.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/workspace-shell-config.test.ts tests/public-config.test.ts` passed.
- `go test -tags unit ./internal/service -run 'TestSettingService_(GetPublicSettings_ExposesTouchRuntimeSettings|UpdateSettings_PersistsTouchRuntimeSettings)' -count=1 -v` passed from `backend/`.
- `go test ./internal/handler/dto -run TestPublicSettingsInjectionPayload_SchemaDoesNotDrift -count=1 -v` passed from `backend/`.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `pnpm run frontend:build` passed with existing Vite browserslist/dynamic-import/chunk-size warnings.
- `git diff --check` passed.

### Next
- Continue shrinking Touch-owned generator behavior and Prompt Gallery interaction state, or move toward the larger Next-to-Sub2API-frontend consolidation.

## 2026-06-17 Touch Public Branding Env Fallback Removal
### Done
- Removed local `envConfigs.app_name`, `envConfigs.app_logo`, and `envConfigs.app_preview_image` from the public branding helper fallback chain.
- Removed local brand/content/appearance fields from Touch `envConfigs`, so fallback public configs no longer expose `NEXT_PUBLIC_APP_NAME`, `NEXT_PUBLIC_APP_DESCRIPTION`, `NEXT_PUBLIC_APP_LOGO`, `NEXT_PUBLIC_APP_FAVICON`, `NEXT_PUBLIC_APP_PREVIEW_IMAGE`, or `NEXT_PUBLIC_APPEARANCE`.
- Touch public branding now prefers Sub2API public settings and otherwise uses code-level safe defaults (`touch`, `/logo.svg`, `/preview.png`), including error boundary, docs shell, auth layout, SEO preview image, and copyright consumers.
- Added a shared favicon helper so the root layout uses Sub2API `touch_app_favicon` when present and `/favicon.svg` as a code fallback when absent.
- Kept `app_url` bootstrap fallback unchanged because local URL fallback is still needed for development links, robots, sitemap, and auth redirects.
- Removed `NEXT_PUBLIC_APP_NAME`, `NEXT_PUBLIC_APP_DESCRIPTION`, and `NEXT_PUBLIC_APPEARANCE` from the Touch env example so brand/content/appearance configuration is documented as Sub2API-owned.
- Added a regression test proving local env branding values no longer affect public branding fallback behavior.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-site-config.test.ts tests/public-config.test.ts tests/auth-shell-config.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.
- `rg -n "NEXT_PUBLIC_APP_NAME|NEXT_PUBLIC_APP_DESCRIPTION|NEXT_PUBLIC_APP_LOGO|NEXT_PUBLIC_APP_FAVICON|NEXT_PUBLIC_APP_PREVIEW_IMAGE|NEXT_PUBLIC_APPEARANCE|envConfigs\\.(app_name|app_description|app_logo|app_favicon|app_preview_image|appearance)" apps/touch/src apps/touch/tests apps/touch/.env.example` now only finds test assertions proving those env values are ignored.

### Next
- Continue reducing Touch frontend-owned static shell, with Prompt Gallery interaction state and broad Next subapp ownership still remaining.

## 2026-06-17 Touch Auth User Menu Credits Label
### Done
- Extended `touch_auth_shell_config` parsing with a `creditsLabel` template for the signed-in user dropdown.
- Updated the Touch `SignUser` menu to render the credits item from runtime auth shell config with `{credits}` placeholder support, keeping local locale text only as fallback.
- Updated the Sub2API admin settings placeholder/hint for `touch_auth_shell_config` so this user-menu copy is visible to operators.
- Added focused auth-shell parser coverage for the new field.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/auth-shell-config.test.ts tests/public-config.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Continue shrinking Touch-owned UI shell text and interaction state, especially prompt-gallery controls and broader Next subapp component ownership.

## 2026-06-17 Touch Bootstrap Env Boundary Tightening
### Done
- Removed Touch local runtime env control for theme and locale detection from `envConfigs`; both now use safe code defaults unless Sub2API public settings provide `touch_theme` / `touch_locale_detect_enabled`.
- Removed `NEXT_PUBLIC_THEME` from the Touch env example and documented that theme and locale-detection behavior are no longer local `NEXT_PUBLIC_*` runtime flags.
- Added a public-config regression assertion that local `NEXT_PUBLIC_THEME` and `NEXT_PUBLIC_LOCALE_DETECT_ENABLED` values are ignored by fallback config loading.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-config.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Continue reducing Touch-owned frontend shell and static text, or start a separate plan for folding the Next subapp UI into the Sub2API Vue frontend.

## 2026-06-17 Touch Payment Modal Shell Settings
### Done
- Extended Touch pricing shell config parsing so `touch_pricing_shell_config` can also carry `payment` dialog copy and payment-provider display metadata.
- Added `PricingPaymentShell` / provider metadata types and passed the parsed shell through the pricing block into the payment modal/provider selector.
- Payment modal title, description, cancel label, no-methods message, provider labels/icons, and selection error copy can now be managed by Sub2API public settings while Touch keeps local locale/default provider metadata as fallback.
- Added pricing service coverage proving the runtime payment shell is parsed from Sub2API pricing shell JSON.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/pricing-plan-service.test.ts tests/public-config.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Continue slimming Touch UI shell: auth modal, settings/credits presentation, and prompt gallery interaction state still live in the Next subapp.

## 2026-06-17 Touch Credits Balance Label Setting
### Done
- Added `touch_credits_balance_label` as a Sub2API public/admin setting so the Touch credits page balance line is no longer fixed in Touch locale JSON.
- Threaded the field through Sub2API settings constants, service read/write models, public settings response, admin settings request/response DTOs, and the Vue admin settings form.
- Updated Touch public config mapping and `/settings/credits` rendering to prefer `credits_balance_label`, with `{balance}` placeholder support and local locale copy as fallback.
- Added backend public/update settings coverage and Touch public-config coverage for the new field.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_(GetPublicSettings_ExposesTouchRuntimeSettings|UpdateSettings_PersistsTouchRuntimeSettings)' -count=1 -v` passed from `backend/`.
- `go test ./internal/handler/dto -run TestPublicSettingsInjectionPayload_SchemaDoesNotDrift -count=1 -v` passed from `backend/`.
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-config.test.ts` passed.
- `pnpm run frontend:typecheck` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run frontend:build` passed with existing Vite browserslist/dynamic-import/chunk warnings.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Remaining Touch UI shell text exists in auth forms, sign user menu, and some prompt gallery interaction states; continue moving configurable text/view-model decisions into Sub2API public settings or API metadata.

## 2026-06-17 Touch Auth Form Shell Settings Expansion
### Done
- Extended `touch_auth_shell_config` parsing with auth form labels, placeholders, footer switch labels, required-field errors, failed-auth fallback errors, and OAuth popup-blocked copy.
- Updated Touch sign-in/sign-up pages and modal forms to prefer those runtime auth shell values with local locale JSON / safety strings as fallback.
- Updated social OAuth buttons to use the runtime popup-blocked message when browser popup creation fails.
- Added focused auth-shell parser coverage for the new fields.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/auth-shell-config.test.ts tests/public-config.test.ts` passed.
- `pnpm run touch:typecheck` passed.
- `pnpm run touch:build` passed with the existing `baseline-browser-mapping` stale-data warning.
- `git diff --check` passed.

### Next
- Remaining Touch UI shell text is mostly user menu labels, prompt-gallery interaction state text, and broader component ownership in the Next app.

## 2026-06-17 Touch Docs Pages Settings Migration
### Done
- Added `touch_docs_pages_config` as a Sub2API public/admin JSON setting for Touch docs page bodies.
- Exposed the setting through backend constants, settings service read/write models, public settings response, admin request/response DTOs, and the Vue admin settings form.
- Extracted a shared Touch structured-content block renderer and reused it for both static page overrides and docs page overrides.
- Added a Touch docs page parser that selects content by docs slug (`index` for `/docs`) and locale, rendering safe structured blocks (`paragraph`, `heading`, `list`, `quote`) with MDX fallback.
- Changed Touch `/docs/[[...slug]]` content and metadata generation to prefer `docs_pages_config` before rendering local MDX from `content/docs`.
- Added Touch tests for docs page config selection, invalid-config fallback behavior, and docs slug mapping.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_(GetPublicSettings_ExposesTouchRuntimeSettings|UpdateSettings_PersistsTouchRuntimeSettings)' -count=1 -v` passed from `backend/`.
- `go test ./internal/handler/dto -run TestPublicSettingsInjectionPayload_SchemaDoesNotDrift -count=1 -v` passed from `backend/`.
- `go test ./internal/service ./internal/handler ./internal/handler/dto -count=1` passed from `backend/`.
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-config.test.ts tests/docs-pages-config.test.tsx tests/static-pages-config.test.tsx tests/docs-shell-config.test.ts tests/landing-page-config.test.ts` passed.
- `pnpm --dir apps/touch exec tsc --noEmit --pretty false` passed.
- `pnpm --dir frontend exec vue-tsc --noEmit` passed.
- `pnpm --dir frontend build` passed with existing Vite dynamic/static import, caniuse-lite, and chunk-size warnings.
- `make build-touch` passed with existing `baseline-browser-mapping` stale-data warnings.

### Next
- Continue shrinking Touch UI ownership: prompt gallery UI composition, auth/payment modal shell, settings shell, and broader Next subapp structure remain in Touch even though data/config is now more Sub2API-managed.

## 2026-06-17 Touch Static Pages Settings Migration
### Done
- Added `touch_static_pages_config` as a Sub2API public/admin JSON setting for Touch static page content.
- Exposed the setting through backend constants, settings service read/write models, public settings response, admin request/response DTOs, and the Vue admin settings form.
- Added a Touch static-page runtime parser that selects content by page slug and locale, renders safe structured blocks (`paragraph`, `heading`, `list`, `quote`) into the existing `StaticPage.body`, and falls back to local MDX on empty/invalid config.
- Changed Touch `/(landing)/[...slug]` rendering and metadata generation to prefer `static_pages_config` before reading `content/pages/*.mdx`.
- Added Touch tests for locale-scoped static page config selection and invalid-config fallback behavior.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_(GetPublicSettings_ExposesTouchRuntimeSettings|UpdateSettings_PersistsTouchRuntimeSettings)' -count=1 -v` passed from `backend/`.
- `go test ./internal/handler/dto -run TestPublicSettingsInjectionPayload_SchemaDoesNotDrift -count=1 -v` passed from `backend/`.
- `go test ./internal/service ./internal/handler ./internal/handler/dto -count=1` passed from `backend/`.
- `pnpm --dir apps/touch exec node --import tsx --test tests/static-pages-config.test.tsx tests/public-config.test.ts` passed.
- `pnpm --dir apps/touch exec tsc --noEmit --pretty false` passed.
- `pnpm --dir frontend exec vue-tsc --noEmit && pnpm --dir frontend build` passed with existing Vite dynamic/static import, caniuse-lite, and chunk-size warnings.
- `make build-touch` passed with existing `baseline-browser-mapping` stale-data warnings.

### Notes
- `pnpm --dir frontend exec prettier ...` is not available because the Sub2API Vue frontend package does not install a `prettier` binary; formatting verification for that package is covered by typecheck/build.

### Next
- Continue reducing Touch-owned content: docs MDX body content still lives under `apps/touch/content/docs`, and broader UI composition remains in the Touch Next subapp.

## 2026-06-17 Touch Docs Shell Settings Migration
### Done
- Added `touch_docs_shell_config` as a Sub2API public/admin JSON setting for Touch docs shell text.
- Exposed the setting through backend settings constants, service read/write models, public settings response, admin request/response DTOs, and the Vue admin settings form.
- Changed Touch docs layout to read `docs_shell_config` from `/api/v1/settings/public` and use it for docs nav labels, nav subtitle, search copy, and locale switch options.
- Kept invalid or empty JSON as a no-op fallback to the existing Touch local docs shell copy.
- Added Touch parser tests for locale-scoped docs shell config and fallback behavior.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_(GetPublicSettings_ExposesTouchRuntimeSettings|UpdateSettings_PersistsTouchRuntimeSettings)' -count=1 -v` passed from `backend/`.
- `go test ./internal/handler/dto -run TestPublicSettingsInjectionPayload_SchemaDoesNotDrift -count=1 -v` passed from `backend/`.
- `go test ./internal/service ./internal/handler ./internal/handler/dto -count=1` passed from `backend/`.
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-config.test.ts tests/docs-shell-config.test.ts tests/landing-page-config.test.ts` passed.
- `pnpm --dir apps/touch exec tsc --noEmit --pretty false` passed.
- `pnpm --dir frontend exec vue-tsc --noEmit` passed.
- `pnpm --dir frontend build` passed with existing Vite dynamic/static import and chunk-size warnings.
- `make build-touch` passed with existing `baseline-browser-mapping` stale-data warnings.

### Next
- Continue shrinking Touch static content ownership: docs/static MDX bodies and landing/docs page content are still stored in Touch code/content files, even though more shell configuration now comes from Sub2API settings.

## 2026-06-16 Touch Public Runtime Settings Migration
### Done
- Shifted the repository run/build shape toward a unified platform entry:
  - root `make build` now builds backend, Vue frontend/admin, and Touch Next subapp
  - root `make test` now runs backend, Vue frontend/admin, and Touch tests
  - added `make build-core` / `make test-core` for intentionally limited Go+Vue checks
  - kept `make build-all` as a compatibility alias for the full platform build
  - updated Touch integration docs so `apps/touch` is described as a platform-managed subapp instead of an optional external checkout
- Reduced Touch-specific API surface for prompt reads:
  - added shared public `/api/v1/prompts/cases` and `/api/v1/prompts/cases/:id` routes
  - switched Touch prompt catalog fetches and gallery remote filtering from `/api/v1/touch/prompts/*` to `/api/v1/prompts/*`
  - kept old Touch prompt read routes as compatibility aliases while session/admin-only capabilities stay under `/api/v1/touch/*`
  - updated integration docs to distinguish shared prompt catalog APIs from Touch web-session routes
- Trimmed a leftover Touch pricing UI orchestration branch:
  - removed the unused `currentSubscription` prop path from the pricing block
  - removed the pricing page's empty `currentSubscription: undefined` injection
  - deleted the now-unused local `current_plan` pricing copy
- Removed the error-boundary branding dependency on local `envConfigs`; the boundary now accepts runtime configs and falls back only to literal safety defaults.
- Shifted Prompt Gallery category filter counts to Sub2API-first metadata:
  - category badges now use `summary.categories` facets returned by Sub2API when available
  - source filter badges now use `summary.sources` facets returned by Sub2API when available
  - local category counting remains only as a no-summary fallback
- Reduced Prompt Gallery hardcoded copy:
  - moved load-more and category expand/collapse labels into prompt locale JSON
  - made import-source and model-tag labels required locale-backed fields instead of `isZh` fallback branches
- Moved Prompt Gallery page shell copy toward Sub2API public settings:
  - added public/admin settings fields for cases/templates page titles and descriptions
  - exposed those fields through `/api/v1/settings/public` and the Sub2API admin settings UI
  - changed Touch cases/templates pages to prefer Sub2API-provided copy with locale JSON as fallback
- Moved Touch workspace page shell copy toward Sub2API public settings:
  - added public/admin settings fields for workspace page title and description
  - exposed those fields through `/api/v1/settings/public` and the Sub2API admin settings UI
  - changed Touch `/ai-image-generator` to prefer Sub2API-provided copy with locale JSON as fallback
- Moved Touch pricing page shell copy toward Sub2API public settings:
  - added public/admin settings fields for pricing page title and description
  - exposed those fields through `/api/v1/settings/public` and the Sub2API admin settings UI
  - changed Touch `/pricing` to prefer Sub2API-provided copy with locale/static shell as fallback
- Moved more Touch pricing shell configuration toward Sub2API public settings:
  - added `touch_pricing_shell_config` as a public/admin JSON setting for locale-scoped pricing groups and button templates
  - exposed the field through the Sub2API admin settings UI and `/api/v1/settings/public`
  - changed Touch pricing rendering to apply Sub2API-provided shell JSON with the existing locale shell as fallback
- Moved Touch credits page shell copy toward Sub2API public settings:
  - added public/admin settings fields for credits page title, description, and purchase button label
  - exposed those fields through `/api/v1/settings/public` and the Sub2API admin settings UI
  - changed Touch `/settings/credits` to prefer AppContext public config with locale JSON as fallback
- Moved Touch settings sidebar shell copy toward Sub2API public settings:
  - added public/admin settings fields for settings page title and credits navigation label
  - exposed those fields through `/api/v1/settings/public` and the Sub2API admin settings UI
  - changed Touch settings layout to prefer Sub2API-provided sidebar copy with locale JSON as fallback
- Moved Touch auth shell copy toward Sub2API public settings:
  - added public/admin settings fields for sign-in/sign-up titles and descriptions
  - exposed those fields through `/api/v1/settings/public` and the Sub2API admin settings UI
  - changed Touch sign-in/sign-up pages and sign modal to prefer Sub2API-provided copy with locale JSON as fallback
- Moved more Touch auth UI labels toward Sub2API public settings:
  - added `touch_auth_shell_config` as a public/admin JSON setting for locale-scoped sign-in/sign-up buttons, switch links, social provider labels, and sign-out copy
  - exposed the field through the Sub2API admin settings UI and `/api/v1/settings/public`
  - changed Touch auth components to apply Sub2API-provided auth shell JSON with local locale copy as fallback
- Moved Touch landing page shell configuration toward Sub2API public settings:
  - added `touch_landing_page_config` as a public/admin JSON setting for locale-scoped DynamicPage home configuration
  - exposed the field through the Sub2API admin settings UI and `/api/v1/settings/public`
  - changed Touch `/` rendering to prefer Sub2API-provided landing page JSON with local `pages/index.json` as fallback
- Extended Sub2API public settings with Touch runtime fields for branding, appearance, locale detection, auth-entry visibility, analytics, ads, affiliate, and customer-service integrations.
- Added Sub2API settings-table read/write support for the same `touch_*` fields through admin settings service/request/response models, so they are no longer only environment-backed.
- Changed Touch runtime config loading so `getEnvironmentConfigs()` and `getPublicConfigs()` prefer `/api/v1/settings/public`, map `touch_*` fields into the existing Touch config keys, and keep R2 credentials server-only.
- Updated Touch root layout, locale theme provider, and SEO metadata generation to use merged runtime configs instead of static `NEXT_PUBLIC_*` values where the setting can now come from Sub2API.
- Updated Touch docs/env examples to mark `NEXT_PUBLIC_APP_*` and auth visibility values as local fallbacks rather than production configuration sources.
- Added a Sub2API admin settings UI section for editing the new Touch runtime settings from the General tab, including branding, appearance, auth-entry visibility, analytics, affiliate, ads, and customer-service fields.
- Verified the Sub2API frontend with `pnpm --dir frontend typecheck`, `pnpm --dir frontend build`, and `git diff --check`.
- Moved more Touch public runtime consumers to Sub2API-first config:
  - `robots.txt` and `sitemap.xml` now read `touch_app_url` at runtime and are marked dynamic.
  - sign-in/sign-up/static-page canonical URLs now use merged runtime config.
  - docs layout branding, auth layout branding, 404 branding, copyright fallback, locale detection, and theme loading now prefer merged runtime config instead of static `NEXT_PUBLIC_*` values.
- Verified the Touch follow-up with `pnpm --dir apps/touch exec tsc --noEmit --pretty false`, `make test-touch`, `make build-touch`, and `git diff --check`.
- Removed the unused Touch-local storage layer now that imported image sync is owned by Sub2API:
  - deleted the local generic storage service and R2 provider implementation
  - removed R2 env names from Touch runtime setting discovery, `.env.example`, and README
  - removed the now-unused `aws4fetch` dependency from Touch
  - simplified public-config tests so Touch no longer carries R2 environment assertions
- Removed direct static `envConfigs` fallback use from the SEO helper; SEO metadata now uses merged runtime config with literal local defaults only as final safety fallback.
- Verified the storage cleanup with `pnpm --dir apps/touch exec tsc --noEmit --pretty false`, `make build-touch`, `make test-touch`, and `git diff --check`.
- Moved prompt model-tag derivation out of Touch frontend normalization and into Sub2API `TouchPromptService`:
  - Sub2API now infers missing `model_tags` during prompt list/detail/upsert normalization.
  - X/Twitter imports now rely on the same service normalization instead of a separate local rule set.
  - Touch frontend now only sanitizes and displays `modelTags` returned by Sub2API.
- Verified the prompt tag migration with backend Touch prompt/import service tests, Touch prompt catalog tests, `make test-touch`, `make build-touch`, and `git diff --check`.
- Moved prompt imported-time derivation out of Touch frontend normalization and into Sub2API `TouchPromptService`:
  - Sub2API now infers missing `imported_at` from X/Twitter snowflake ids in source URLs or `tw-*` case ids.
  - Existing stored `imported_at` values are preserved.
  - Touch frontend no longer exports or runs `inferPromptImportedAt`.
- Verified the prompt imported-time migration with backend Touch prompt/import service tests, focused Touch prompt catalog tests, Touch `tsc`, `make test-touch`, `make build-touch`, and stale frontend inference scans.
- Moved prompt gallery page source-type filtering into Sub2API list requests:
  - cases page now calls `getPromptCases({ sourceType: "case" })` instead of loading all prompt items and filtering out templates in Touch
  - templates page now calls `getPromptCases({ sourceType: "template" })` instead of loading all prompt items and filtering locally
  - the shared Touch prompt catalog loader now forwards source-type filters to Sub2API while preserving the existing all-cases default
- Verified the source-type filter migration with Touch `tsc`, focused prompt catalog tests, and backend Touch prompt route tests.
- Added Sub2API prompt catalog summary metadata to the Touch prompt list API:
  - `/api/v1/touch/prompts/cases` now returns `summary` with total/case/template/source/category counts plus source/category facets
  - Touch parses the summary through the Sub2API fetch layer and passes it to the gallery
  - the gallery uses Sub2API summary/source facets as the initial no-interaction baseline, while keeping local instant filtering for search/source/category changes
- Verified the prompt summary contract with backend Touch prompt service/route tests, Touch `tsc`, and focused prompt catalog tests.
- Extended the Sub2API prompt catalog filter contract for the remaining gallery controls:
  - list and summary queries now support `source_project`
  - list and summary queries now support `has_image=true|false`
  - search now includes title, prompt, category, source project, and source label
  - Touch fetch/model helpers forward source project, category, search, and image-presence filters
  - cases page now requests `has_image=true` from Sub2API instead of loading no-image cases and dropping them locally
- Verified the prompt filter contract with backend Touch prompt service/route tests, Touch `tsc`, and focused prompt catalog tests.
- Switched prompt gallery interactions toward Sub2API-filtered data:
  - search/source/category/mode changes now debounce and request filtered `/api/v1/touch/prompts/cases` pages directly from Sub2API
  - filtered requests use the browser-safe Sub2API public base URL and do not reintroduce a Touch API route
  - the gallery keeps local filtering as a fallback if the remote filter request fails
  - cases layout keeps `has_image=true` in filtered requests so no-image cases stay out of the case grid
- Verified the gallery remote-filtering client path with Touch `tsc` and focused ESLint on `prompt-gallery/gallery.tsx`.

### Failures
- Initial Touch public-config test expected the older narrow public whitelist; updated it to include the newly browser-safe public runtime IDs while still asserting R2 secrets are not exposed.

### Next
- Remaining direct `envConfigs` uses are now limited to Sub2API bootstrap base URL, default locale bootstrap, shared config assembly, and the client error-boundary fallback.
- Continue reducing Touch UI orchestration by moving remaining client-side category/source/stat recomputation into Sub2API response metadata where doing so still preserves instant interactions.

## 2026-06-14 Touch Web Session And BFF Reduction
### Done
- Added browser-facing Sub2API Touch web endpoints under `/api/v1/touch/web` for email login, registration, OAuth token session establishment, refresh, logout, current user, current credits, checkout, and admin-gated X/Twitter prompt import.
- Switched Touch client flows to call those Sub2API web endpoints directly with `credentials: include`:
  - auth modal/page login, registration, logout
  - app user hydration, refresh, and credits lookup
  - pricing checkout
  - prompt-gallery X/Twitter import
  - OAuth callback session handoff
- Removed the now-replaced Touch BFF routes for:
  - `/api/auth/sub2api/login`
  - `/api/auth/sub2api/logout`
  - `/api/auth/sub2api/oauth/session`
  - `/api/auth/sub2api/refresh`
  - `/api/auth/sub2api/register`
  - `/api/user/get-user-info`
  - `/api/user/get-user-credits`
  - `/api/payment/checkout`
  - `/api/admin/prompts/import-source`
- Reduced Touch's remaining Next API routes from 13 to 4:
  - `/api/auth/sub2api/oauth/start`
  - `/api/payment/callback`
  - `/api/prompts/image`
  - `/api/docs/search`
- Updated `apps/touch/docs/sub2api-integration.md` so the remaining-route list reflects the current 4-route state and documents the Sub2API web replacements.
- Kept the remaining 4 routes because they are still tied to OAuth popup start mechanics, legacy payment return redirect compatibility, local/owned static image proxying, and local Fumadocs search.
- Continued the bridge reduction to zero Touch Next.js API routes:
  - moved Touch OAuth start to Sub2API `/api/v1/touch/web/auth/oauth/:provider/start`
  - changed the Touch OAuth popup to navigate directly to Sub2API
  - removed the old Touch OAuth bridge route and orphaned admin-key bridge helper
  - removed the legacy payment callback route after checkout return URLs were confirmed to point directly at `/settings/credits`
  - removed the prompt image proxy route by serving local public images directly and keeping owned static URLs direct
  - disabled Fumadocs docs search and removed the local docs search route
  - confirmed `apps/touch/src/app/api` no longer exists
- Removed the remaining Touch-side admin-key bridge dependency:
  - server-rendered admin-import visibility now uses the current Sub2API session role instead of calling `/api/v1/touch/admin/auth/check`
  - deleted unused Touch helpers for admin-key X import, user sync, user credit lookup, and admin checkout bridge calls
  - removed `SUB2API_ADMIN_API_KEY` from Touch active env examples and docs
  - added backend coverage that `/api/v1/auth/me` continues returning `role`, which Touch uses for admin UI gating

### Failures
- None from code verification. One initial backend verification command used a path relative to the wrong working directory; the corrected command passed.

### Next
- Keep future Touch server capabilities in Sub2API web/admin routes rather than reintroducing Next.js API routes.
- If docs search is needed again, implement it in Sub2API or use an external/static search provider instead of a Touch-local API route.

## 2026-06-13 Touch Subapp Merge
### Done
- Added the current Touch Next.js app into the Sub2API repository under `apps/touch` as the first reversible monorepo step.
- Kept the original Touch checkout intact during migration; deployment can switch after verification without losing the standalone rollback path.
- Added Sub2API root Makefile targets:
  - `make build-touch`
  - `make test-touch`
  - `make build-all`
  - `make preview-touch`
  - `make deploy-touch`
- Left the historical `make build` target unchanged so existing backend + Vue build pipelines do not suddenly require Touch dependencies.
- Updated root `.gitignore` so `apps/touch` source, scripts, tests, and docs are trackable while local env files, runtime data, dependencies, and build outputs remain ignored.
- Documented the remaining Touch Next API routes in `apps/touch/docs/sub2api-integration.md`, including each route's current server-side security role and the condition required before it can be removed.

### Failures
- None. `make test-touch` passed. `make build-touch` passed with only the existing `baseline-browser-mapping` stale-data warning.

### Next
- Continue replacing BFF routes only after Sub2API exposes equivalent browser-safe web/session endpoints.

## 2026-06-03 Frontend i18n Coverage Cleanup
### Done
- Fixed high-priority public/auth i18n gaps and hardcoded runtime copy by wiring these pages back to shared locale keys:
  - `frontend/src/views/public/LegalDocumentView.vue`
  - `frontend/src/views/public/DocsView.vue`
  - `frontend/src/views/public/ModelsPlazaView.vue`
  - `frontend/src/views/auth/LoginView.vue`
  - `frontend/src/views/auth/RegisterView.vue`
- Added missing public/auth locale coverage in both `zh.ts` and `en.ts` for:
  - login-agreement warnings
  - DingTalk auth flow labels
  - legal document page copy
  - model plaza public-page copy
  - email suffix overflow wording
  - DingTalk provider labels in profile binding UI
- Filled the targeted missing i18n keys for:
  - `admin.settings`
  - `admin.users`
  - `admin.accounts`
- Added the previously missing locale trees for:
  - `admin.settings.apiKeyAcl`
  - `admin.settings.dingtalk`
  - `admin.settings.emailTemplates` (zh)
  - `admin.settings.subscriptionExpiryNotify` (en)
  - `admin.settings.platformQuota`
  - `admin.users.platformQuota`
  - multiple `admin.accounts.openai/*` and account scheduling labels
- Cleaned the remaining English-locale Chinese leftovers inside the same targeted admin modules.
- Added a focused locale regression test:
  - `frontend/src/i18n/__tests__/localeCoverage.spec.ts`
  - `frontend/src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- Verified with automated scans that:
  - missing key count for `admin.settings / admin.users / admin.accounts / auth.dingtalk` is now `0`
  - Chinese-value leak count in `en.ts` for `admin.settings / admin.users / admin.accounts` is now `0`
- Continued the admin sweep for:
  - `admin.riskControl`
  - `admin.redeem`
  - `admin.subscriptions`
- Added the missing locale keys for those three namespaces and cleaned the English-locale Chinese leftovers in `admin.redeem`.
- Verified with automated scans that:
  - missing key count for `admin.riskControl / admin.redeem / admin.subscriptions` is now `0`
  - Chinese-value leak count in `en.ts` for those three namespaces is now `0`
- Continued the remaining admin namespace sweep for:
  - `admin.ops`
  - `admin.groups`
  - `admin.channels`
  - `admin.backup`
  - `admin.channelMonitor`
  - `admin.proxies`
  - `admin.dashboard`
  - `admin.usage`
- Added the remaining missing keys for those namespaces and cleaned the final English-locale Chinese leftovers in:
  - `admin.groups`
  - `admin.dashboard`
  - `admin.proxies`
  - `admin.ops`
- Verified with automated scans that:
  - missing key count for all `admin.*` namespaces is now `0`
  - Chinese-value leak count in `en.ts` for all `admin.*` namespaces is now `0`
- Continued the global frontend locale sweep beyond `admin.*` and fixed the remaining referenced missing keys for:
  - `apiTest`
  - `dashboard`
  - `keyUsage`
  - `payment`
  - `usage`
  - `userSubscriptions`
  - shared `common.*` action/status copy
- Fixed two obvious runtime hardcoded strings by routing them through locale keys:
  - `frontend/src/components/account/UsageProgressBar.vue`
  - `frontend/src/views/admin/ops/components/OpsSystemLogTable.vue`
- Verified with automated scans that:
  - missing key count for the entire frontend locale tree is now `0`
  - Chinese-value leak count in `frontend/src/i18n/locales/en.ts` is now `0`

### Failures
- No functional failures. Frontend build still reports the existing Vite dynamic-import chunk warnings and stale Browserslist data warning; neither is introduced by this i18n cleanup.

### Next
- If needed, continue converting intentionally bilingual helper-based copy into centralized locale keys in shared/public/admin components such as:
  - `views/admin/SettingsView.vue` legacy `localText(...)` sections
  - `views/admin/orders/*` bilingual helper-based marketing/config copy
  - `utils/loginAgreementTemplates.ts` document titles/templates that are currently authored as fixed source content
- If the current scope is acceptable, commit and push this i18n cleanup as a standalone frontend/docs-quality change.

## 2026-05-19 Backend Login Agreement Enforcement
### Done
- Added persistent user-level login agreement acceptance tracking:
  - `login_agreement_accepted_revision`
  - `login_agreement_accepted_at`
- Added migration `096_add_user_login_agreement_acceptance.sql`.
- Regenerated Ent code and wired the new acceptance fields through:
  - service user model
  - repository create/update/entity mapping
- Added backend login-agreement enforcement helpers so the server now validates the current agreement revision instead of relying on frontend-only localStorage.
- Enforced the current agreement revision on:
  - password login
  - email registration
  - pending OAuth bind-login
  - pending OAuth create-account
  - pending OAuth exchange/token issuance
  - GitHub/Google complete-registration
  - LinuxDo / OIDC / WeChat complete-registration
  - direct GitHub/Google verified-email OAuth callback completion
- Kept the auth-page UX working by adding agreement payload propagation from browser state to:
  - login/register submit payloads
  - email verify registration completion
  - pending OAuth bind/create/exchange requests
  - legacy complete-registration requests
  - GitHub/Google OAuth start URL so direct callback completion can still prove the accepted revision
- Added/updated regression coverage for:
  - password login agreement enforcement
  - email registration agreement enforcement
  - pending OAuth token-exchange agreement enforcement
  - direct email OAuth callback agreement enforcement
  - existing auth callback/auth view payload tests

### Failures
- `pnpm exec prettier` is not available in this workspace, so formatting was handled with `gofmt` plus the existing frontend lint/build pipeline instead.

### Next
- Commit the agreement-enforcement changes with a Lore commit message.
- Push to `main`.
- Pull on the Google server, rebuild the embedded frontend/backend binary, restart `SERVER_PORT=8081`, and update `contact_info` to `cc@cloudbase.eu.org`.

### Follow-up
- Found that a partial `PUT /admin/settings` call could still flip `login_agreement_enabled` back to `false` because the request field was a non-pointer bool.
- Hardened the admin settings handler so omitted `login_agreement_enabled` keeps the previous value instead of silently defaulting to `false`.

## 2026-05-21 Privacy Policy Login Agreement Addendum
### Done
- Added a dedicated `privacy-policy` login-agreement document template with runtime placeholders for effective date, update date, and contact info.
- Extended the commercial legal template bundle from 4 documents to 5 documents by including:
  - 商业服务条款
  - 隐私条款
  - 使用政策
  - 支持的国家和地区
  - 服务特定条款
- Kept backward compatibility for legacy blank agreement bundles that still only contain the original 4 empty docs.
- Added a non-destructive admin action to append the privacy policy document into the current agreement list without overwriting existing configured documents.
- Switched the admin agreement document icon logic from index-based mapping to title-based mapping so adding a fifth document does not mislabel existing cards.
- Added regression coverage for:
  - 5-document template generation
  - privacy-policy rendering contact substitution
  - non-destructive privacy-policy append behavior
  - duplicate privacy-policy avoidance

### Failures
- None during local implementation. Existing Vite chunk-size warnings remain informational only.

### Next
- Commit and push the privacy-policy additions.
- Deploy the updated frontend/backend bundle to production.
- Use the production admin settings flow to append the privacy-policy document without overwriting the existing 4 document bodies.

## 2026-05-21 Remove Onboarding Popup And Entry
### Done
- Removed the onboarding tour auto-start from the shared app layout so the welcome walkthrough no longer opens automatically on dashboard entry.
- Removed the user-dropdown entry used to replay/restart the onboarding tour.
- Left the underlying onboarding store/composable code in place so unrelated guided-step hooks on pages do not break, but they are now inert because no top-level tour is initialized.

### Failures
- None. Verification only surfaced the existing Vite chunk-size warnings.

### Next
- Commit and push the onboarding removal.
- Deploy the rebuilt frontend/backend bundle to production.
- Verify that dashboard no longer shows the onboarding modal and the user dropdown no longer contains the onboarding entry.

## 2026-05-21 Account Test Copy Output
### Done
- Added regression coverage for the account-test modal copy button in both:
  - `frontend/src/components/admin/account/AccountTestModal.vue`
  - `frontend/src/components/account/AccountTestModal.vue`
- Verified the old behavior failed the new tests because the copy button concatenated the full terminal log instead of the actual model response body.
- Updated both modal implementations to maintain a dedicated response-content buffer during streamed test output.
- Changed the copy button to copy only the streamed response text while keeping the existing terminal log display unchanged.
- Verified:
  - `pnpm --dir frontend exec vitest run src/components/admin/account/__tests__/AccountTestModal.spec.ts src/components/account/__tests__/AccountTestModal.spec.ts`
  - `pnpm --dir frontend exec eslint src/components/admin/account/AccountTestModal.vue src/components/account/AccountTestModal.vue src/components/admin/account/__tests__/AccountTestModal.spec.ts src/components/account/__tests__/AccountTestModal.spec.ts --ext .vue,.ts`
  - `pnpm --dir frontend run typecheck`
  - `pnpm --dir frontend run build`

### Failures
- None. Existing Vite chunk-size warnings remain informational only.

### Next
- Commit and push the copy-behavior fix.
- Deploy the rebuilt frontend/backend bundle to production.
- Open the account-test modal and verify the copy button only copies the response body.

## 2026-05-21 Account Test Error Copy Regression
### Done
- Added regression coverage for failed account-test runs in both modal variants so the copy button must preserve the surfaced error response payload.
- Verified the current implementation failed those tests because it only copied streamed response text and returned an empty string when the test completed with an error.
- Updated both modal variants so copy behavior now:
  - copies streamed response text on success
  - copies `errorMessage` on failure
- Re-verified:
  - `pnpm --dir frontend exec vitest run src/components/admin/account/__tests__/AccountTestModal.spec.ts src/components/account/__tests__/AccountTestModal.spec.ts`
  - `pnpm --dir frontend exec eslint src/components/admin/account/AccountTestModal.vue src/components/account/AccountTestModal.vue src/components/admin/account/__tests__/AccountTestModal.spec.ts src/components/account/__tests__/AccountTestModal.spec.ts --ext .vue,.ts`
  - `pnpm --dir frontend run typecheck`
  - `pnpm --dir frontend run build`

### Failures
- None locally. Existing Vite chunk-size warnings remain informational only.

### Next
- Commit and push the error-copy regression fix.
- Deploy to production.
- Reproduce a failed account test and verify the copy button includes the error response details.

## 2026-05-21 Claude Code Native Account Test Mode
### Done
- Added a dedicated `claude_code` account-test mode so Claude-compatible accounts can be tested with a more native Claude Code request fingerprint instead of only the generic account probe.
- Backend changes:
  - introduced `AccountTestModeClaudeCode`
  - added a deterministic Claude Code-native payload builder with stable field order
  - added native request headers for this mode:
    - `Accept: application/json`
    - `X-Stainless-Helper-Method: stream`
    - `X-Client-Request-Id`
    - expanded Claude Code mimic beta header set
  - threaded the mode through Anthropic and Antigravity API-key Claude test paths
- Frontend changes:
  - added a request-mode selector for Claude-compatible account test modals
  - preserved the existing OpenAI default/compact test modes
  - sent `mode: "claude_code"` to the backend when the native mode is selected
- Added regression coverage for:
  - backend native-mode header/payload construction
  - admin account modal sending `claude_code`
  - user account modal sending `claude_code`
- Extended account test content editing so non-image probes now send the operator-provided prompt instead of hard-coded `"hi"` for:
  - Anthropic default probe
  - Anthropic Claude Code native probe
  - OpenAI default probe
  - OpenAI compact probe
  - Anthropic service-account probe
  - Bedrock text probe
- Updated both account test modals to:
  - show an editable request-content textarea for non-image tests
  - show the current request mode inside the result panel
  - show the effective prompt value inside the result panel summary row
  - render raw upstream request / response debug panels with redacted sensitive headers
- Backend account-test SSE now emits upstream debug blocks for supported test paths, including:
  - method + URL
  - redacted request / response headers
  - request body
  - response body or a streamed-body marker
- Verified:
  - `cd backend && go test ./... -count=1`
  - `pnpm --dir frontend exec vitest run src/components/admin/account/__tests__/AccountTestModal.spec.ts src/components/account/__tests__/AccountTestModal.spec.ts`
  - `pnpm --dir frontend exec eslint src/components/admin/account/AccountTestModal.vue src/components/account/AccountTestModal.vue src/components/admin/account/__tests__/AccountTestModal.spec.ts src/components/account/__tests__/AccountTestModal.spec.ts src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .vue,.ts`
  - `pnpm --dir frontend run typecheck`
  - `pnpm --dir frontend run build`
  - `git diff --check`

### Failures
- None locally. Existing Vite chunk-size warnings remain informational only.

### Next
- Completed:
  - committed and pushed the native test-mode feature
  - deployed the rebuilt frontend/backend bundle to production
  - verified on production that Claude-compatible account tests now expose:
    - `请求方式 -> 常规请求`
    - `请求方式 -> Claude Code 原生测试`
  - confirmed the live request body includes `mode: "claude_code"` when the native mode is selected

## 2026-05-13 Active Email Daily Rate Limit
### Done
- Added a shared daily quota for user-initiated email sends to reduce SMTP abuse risk.
- Limited active email triggers across:
  - registration verification codes
  - pending OAuth email verification codes
  - email identity binding codes
  - notification email binding codes
  - TOTP verification codes
  - password reset requests
- Scoped the quota by authenticated user ID when available, and by normalized email address for pre-login/pre-registration flows.
- Stored counters in Redis with hashed keys and a TTL that expires at the next local midnight.
- Kept system/admin-driven emails outside this limit:
  - welcome emails
  - admin registration push notifications
  - balance/quota/ops alerts
- Added regression coverage for registration verification code daily-limit behavior and updated affected test doubles.
- Verified:
  - `go test -tags=unit ./internal/... -count=1`
  - `make build`
  - `git diff --check`

### Failures
- None during implementation.

### Next
- Commit and push the backend-only limit.
- Deploy to the Google server and verify the live backend restarts on `SERVER_PORT=8081`.

## 2026-05-13 CI Follow-up After Email Rate Limit
### Done
- Investigated the GitHub Actions failure for commit `ceac08cd`.
- Fixed frontend lint by replacing the Docs build-time global with `import.meta.env.VITE_DOCS_CONTENT_VERSION`.
- Fixed backend lint by:
  - checking/explicitly ignoring email template `strings.Builder` writes
  - checking/explicitly ignoring registration webhook HMAC writes and response body close
  - removing obsolete no-logo email-template wrapper functions that were only referenced by unit tests
  - updating unit tests to call the current `WithLogo(..., "")` helpers directly
- Follow-up: fixed the same unchecked `strings.Builder` writes in the welcome email template after CI surfaced the remaining file.
- Verified:
  - `make test-frontend`
  - `go test -tags=unit ./internal/service -count=1`
  - `go test -tags=unit ./internal/service -run 'TestBuild.*Email|TestAuthService_SendVerifyCode_ActiveEmailDailyLimit' -count=1`
  - `make build`
  - `git diff --check`

### Failures
- Local `golangci-lint run --timeout=30m` could not execute because the installed binary was built with Go 1.25 while the project targets Go 1.26.3; CI uses the correct v2.9 binary.

### Next
- Commit and push the CI follow-up.
- Monitor the new CI run and deploy if it passes.

## 2026-05-13 Admin Stats and Registration Push
### Done
- Verified live admin dashboard/statistics APIs with the provided admin account:
  - stats, snapshot, trend, model stats, group stats, user ranking, usage list, and users list all returned HTTP 200 / code 0
  - trend/model/group/ranking data is empty because live usage totals are currently zero
- Added configurable new-user registration push notifications:
  - supports DingTalk and Feishu custom bot webhooks
  - supports optional bot signing secrets without returning secrets to the frontend
  - sends asynchronously after successful email registration or finalized OAuth email registration
  - notification failures are logged and do not fail registration
- Added admin settings API fields and admin settings UI controls for the registration push provider, webhook, and secret.
- Added regression tests for DingTalk payload/sign query, Feishu signed text payload, disabled notification behavior, and updated admin settings API contracts.
- Committed and pushed the implementation to `aias00/main`.
- Deployed commit `5d67cb37` to the Google server using the server-side build flow:
  - `pnpm --dir frontend install --frozen-lockfile`
  - `pnpm --dir frontend run build`
  - `go build -tags embed -o ../sub2api.new ./cmd/server`
- Restarted the live `./sub2api` process on `SERVER_PORT=8081`; the first manual restart used the default port `8080`, which caused a temporary Cloudflare 502 until the process was restarted with the correct port.
- Verified the live domain is healthy again and admin APIs return HTTP 200 / code 0 after deployment.
- Verified `/api/v1/admin/settings` now exposes:
  - `registration_notify_enabled`
  - `registration_notify_provider`
  - `registration_notify_webhook_url`
  - `registration_notify_secret_configured`

### Failures
- The first live API login attempt returned HTTP 403 without a browser user-agent; retrying with a browser-like user-agent succeeded.

## 2026-05-15 Claude Code Gateway Diagnosis
### Done
- Investigated production `/v1/messages` failures for Claude Code traffic.
- Confirmed the API key authenticates and the default group now selects upstream account `10`.
- Reproduced `/v1/messages` with Claude Code-style headers and `claude-opus-4-7`; the simple message request returned HTTP 200.
- Reproduced `/v1/messages/count_tokens`; the upstream relay returned HTTP 404 with `Invalid URL (POST /v1/messages/count_tokens)`.
- Updated the generic count_tokens forwarding branch to return `not_found_error` for unsupported count_tokens 404s, matching the API-key passthrough branch so Claude Code can fall back to local token estimation.

### Failures
- Production logs also show large streaming Claude Code requests can hit upstream HTTP 504 from `https://www.fkclaude.xyz`; this is separate from the count_tokens fallback issue.

### Next
- Run focused backend tests for count_tokens fallback.
- If tests pass, deploy the fix and retest production `/v1/messages/count_tokens`.
- Full unit tests initially failed because the admin settings API contract expected the old response shape; updated the contract snapshot with the new fields.
- Deployment health check initially targeted the wrong local port (`8081`) after a default-port restart; corrected by restarting with `SERVER_PORT=8081`.

### Next
- If the user provides a real DingTalk or Feishu bot webhook, configure it under admin settings and run a live registration smoke test to confirm delivery.

## 2026-05-13 Email Template Logo
### Done
- Updated the shared HTML email shell to render a real image logo when a logo URL is available.
- Email logo resolution now prefers configured `site_logo`, then falls back to the public site origin plus `/logo.png`.
- Wired the resolved logo into:
  - registration verification code emails
  - password reset emails
  - notification email verification emails
  - balance low and account quota alert emails
  - admin SMTP test emails
- Added regression coverage for image-logo rendering and default `/logo.png` resolution from `frontend_url`.
- Verified backend unit coverage with `go test -tags unit ./internal/...`.
- Committed and pushed the logo change as `3431e2f0`.
- Deployed the embedded frontend/backend build on the Google server and restarted the live `SERVER_PORT=8081` process.
- Verified `https://cloudbase.eu.org/dashboard` returns HTTP 200 and `https://cloudbase.eu.org/logo.png` serves the real PNG logo.
- Follow-up: added `GET /api/v1/settings/site-logo` so an uploaded base64 `site_logo` is exposed as a normal public image URL for email clients, instead of falling back to `/logo.png`.
- Deployed the uploaded-logo endpoint and verified `https://cloudbase.eu.org/api/v1/settings/site-logo` returns a 1024x1024 PNG with an ETag.

### Failures
- Initial test command used the backend directory with root-relative paths, so `gofmt` could not find files; reran from the repository root successfully.

### Next
- If needed, send a live SMTP test email from admin settings to inspect real mailbox rendering.

## 2026-05-13 Docs Content Rewrite
### Done
- Replaced the small built-in docs set with a DragonCode-inspired information architecture:
  - site introduction
  - environment preparation
  - quickstarts
  - advanced guides
- Rewrote the content for cloudbase instead of copying third-party wording verbatim.
- Updated all project URLs in the docs to `https://cloudbase.eu.org`.
- Added matching English docs so the existing docs i18n switch remains useful.
- Removed stale `getting-started` / `guides` docs that were no longer linked.
- Fixed Docsify sidebar loading so localized docs load `_sidebar.md` from the active `basePath`.
- Follow-up: replaced hardcoded `https://cloudbase.eu.org` docs links with same-origin app routes and added a regression test to prevent reintroducing the production domain in Markdown.

### Failures
- Local Vite preview reports `/setup/status` and `/api/v1/settings/public` proxy errors because no backend is attached to the preview server; docs content itself renders normally.

### Next
- Commit, push, build on the server, restart the service, and verify the live `/docs` page.

## 2026-05-13 Docs Mobile and i18n Follow-Up
### Done
- Added the shared language switcher to the public Docsify docs header.
- Routed Docsify content by active locale, keeping Chinese at `/docs-content/` and adding English docs under `/docs-content/en/`.
- Capped the mobile Docsify sidebar height and reduced its padding so the table of contents no longer consumes the whole first screen.
- Kept the mobile dashboard CTA compact by showing the shorter dashboard label on small screens.
- Deployed commit `ffaaf999` to the Google server with server-side frontend and backend builds.
- Verified live docs in Chinese and English on mobile, plus dashboard, purchase, and profile desktop pages for overflow, raw i18n keys, and console errors.

### Failures
- None; local verification passed.

### Next
- Continue lower-priority frontend style and interaction review after the current high-priority docs pass.

## 2026-05-10 Global Password Minimum Length
### Done
- Added a shared backend password policy setting `password_min_length` with a hard floor of `8` and a default of `8`.
- Exposed `password_min_length` through:
  - admin settings payloads
  - public settings API
  - SSR-injected public settings
  - frontend `PublicSettings` / admin settings typings
- Unified backend password validation for:
  - email registration
  - password reset
  - OAuth pending account creation
  - verified email OAuth completion
  - first-time email binding
  - profile password changes
  - admin-created / admin-updated user passwords
- Unified frontend password-length UX for:
  - register
  - reset password
  - Google/GitHub OAuth completion
  - pending OAuth account creation
  - profile password setup/change
  - email binding in profile
- Added the admin settings UI control for minimum password length and documented that it cannot be lower than `8`.
- Updated the stale auth docs/examples that still said `6` characters.

### Failures
- One service unit test still asserted the old `email_bound` semantics for compat-backfilled email identities; updated the assertion to match the current OAuth/passwordless semantics before rerunning the unit-tag test suite.
- Accidentally invoked the Corepack-managed `pnpm` shim once during verification; switched back to the repo's stable local/Homebrew toolchain path immediately.

### Next
- Finish the in-flight `go test -tags unit ...` rerun after the assertion fix.
- If green, run the final default Go test pass, then deploy the rebuilt frontend/backend to the live server.

## 2026-05-12 Login Turnstile Scope
### Done
- Traced the live login flow and confirmed the current behavior:
  - frontend rendered Turnstile on the password-login page whenever `turnstile_enabled=true`
  - the login button stayed disabled until a Turnstile token existed
  - backend `/api/v1/auth/login` also hard-required `turnstile_token`
- Confirmed the live site currently produces Turnstile/CSP/Trusted Types console errors on the login page, making the password-login chain too brittle.
- Narrowed Turnstile scope so it no longer blocks standard email/password login.
- Kept Turnstile on higher-friction flows that still benefit from it:
  - register
  - forgot password
  - verification-code sending
  - pending OAuth account creation code send
- Updated admin Turnstile copy to reflect the new scope instead of still claiming it protects login.
- Added a focused LoginView regression test to lock:
  - no Turnstile widget on password login even when public settings report it enabled
  - login submission no longer sends `turnstile_token`

### Failures
- None during the code change itself; the original issue was reproducible in live QA rather than in unit tests.

### Next
- Finish the in-flight frontend/backend builds.
- If green, deploy to the server and verify that plain password login works without Turnstile while register/forgot-password still require it.

## 2026-05-14 Admin Settings Recharge Product Blank Page
### Done
- Investigated the settings page blank screen after saving recharge products.
- Root cause: save response could contain `payment_recharge_products[].features` as `null`, while the settings template expects an array and calls `.join()`.
- Patched backend admin settings DTO output to serialize empty recharge product features as `[]`.
- Patched frontend save handling to normalize returned recharge products before assigning them back into the form.
- Added frontend/backend regression tests for null/empty feature lists.
- Verified:
  - `pnpm --dir frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
  - `pnpm --dir frontend exec vue-tsc --noEmit`
  - `pnpm --dir frontend exec eslint src/views/admin/SettingsView.vue src/views/admin/__tests__/SettingsView.spec.ts`
  - `go test ./internal/handler/admin -run TestDTORechargeProductsEncodesEmptyFeaturesAsArray -count=1`
  - `git diff --check`

### Failures
- An initial backend test command was run from the repository root, but this project keeps `go.mod` under `backend`; reran from `backend` successfully.
- The new backend test initially used an old module import path; corrected to `github.com/Wei-Shaw/sub2api/internal/service`.

### Next
- If this needs to go live immediately, commit/push and deploy with the server-side build flow.

## 2026-05-12 Turnstile CSP Nonce Propagation
### Done
- Narrowed the likely root cause of the remaining register/forgot-password Turnstile instability:
  - the page is served with a CSP nonce
  - the frontend `TurnstileWidget` was dynamically injecting `api.js` without reusing that nonce
- Added a small CSP nonce resolver utility on the frontend.
- Updated `TurnstileWidget` so the dynamically inserted Turnstile script copies the existing page nonce before appending.
- Added a focused frontend regression test for nonce discovery.
- Confirmed the server-side \"build on server\" path is close but not yet the default:
  - Node and pnpm are available on the server
  - the missing piece is `frontend/node_modules` / install step, not the toolchain binary itself

### Failures
- None yet for this nonce propagation patch; live re-validation is still pending.

### Next
- Finish the in-flight frontend test/typecheck/build runs.
- Deploy the nonce propagation patch and re-check the register / forgot-password Turnstile behavior.
- If that stabilizes, switch the deployment workflow to server-side frontend builds instead of uploading `dist`.

## 2026-05-12 Server-side Frontend Build + UI Anomalies
### Done
- Installed the frontend build toolchain on the server:
  - `nodejs`
  - `npm`
  - global `pnpm`
- Installed server-side frontend dependencies under `/home/aias94coffee/sub2api-work/sub2api-main/frontend/node_modules`.
- Verified the server can now run:
  - `pnpm install --frozen-lockfile`
  - `pnpm exec vite build`
  - backend `go build -tags embed`
- Narrowed admin onboarding auto-start so it only launches automatically from dashboard entry routes, not arbitrary admin pages like settings or plan management.
- Updated the recharge page CTA so it no longer shows a misleading `¥0.00` payment label before the user has selected an amount.

### Failures
- None for these two UI changes; both passed local targeted tests and typecheck.

### Next
- Deploy the onboarding + recharge CTA adjustments with the server-side build flow.
- Continue the remaining anomaly sweep after those are live.

## Done
- Confirmed GitHub Actions run 25547187360 failed in golangci-lint.
- Retrieved the failing annotations from the job page.
- Fixed the reported issues:
  - removed unused helper in google_validation_error.go
  - gofmt-aligned proxy_service.go var block
  - checked/explicitly ignored Close, ReadAll, Write, and WriteString return values at the reported sites
- Re-ran gofmt on all touched Go files.
- Re-ran backend go test ./... successfully.

## Failures
- GitHub CLI access to the Actions API timed out; used the web job page instead.
- Full golangci-lint re-run could not complete locally because the environment first hit a v1/v2 mismatch, then upstream module download timeouts via the proxy.
- Standalone errcheck re-run could not complete because proxy access to proxy.golang.org timed out.

## Next
- Commit the five-file CI fix.
- Push to aias00/main to trigger CI again.

## Next
- Monitor the follow-up GitHub Actions run triggered by commit fe3807a0.
- If it fails, inspect the failed job and apply the smallest verified fix.
- Fixed a remaining unchecked type assertion in Gemini Web session handler tests after reviewing CI annotations.
- Reproduced the CI test job locally with make test-unit and make test-integration.
- Restored proxy normalization and duplicate checks inside adminServiceImpl for CI unit tests.

## 2026-05-09 Local Run
### Done
- Confirmed this repository was not the source of the already-running `sub2api` instance on `127.0.0.1:8080`; that container is mounted from `/Users/aias/Work/github/2Api`.
- Built the current repository backend binary at `backend/bin/sub2api-local`.
- Started isolated dependency containers for this repository on host ports:
  - PostgreSQL: `127.0.0.1:15432` (`sub2api-codex-pg`)
  - Redis: `127.0.0.1:16379` (`sub2api-codex-redis`)
- Started the current repository server process on `http://127.0.0.1:18082`.
- Verified:
  - `GET /health` returns `{"env":"production","status":"ok"}`
  - `GET /` returns `HTTP/1.1 200 OK`
  - Browser snapshot shows the `Home - Sub2API` page loaded from `http://127.0.0.1:18082/home`
- Created admin account:
  - email: `admin@sub2api.local`
  - password: `admin123456`

### Failures
- Docker image rebuild from current source failed because external package registries timed out during `apk add` and `pnpm install`.
- Frontend rebuild via local `vue-tsc` failed because local `node_modules` is missing `@stripe/stripe-js` type resolution.

### Next
- If a fully fresh frontend build is required, restore/install the missing frontend package metadata and rerun the frontend build.
- If a containerized run from current source is required, retry the Docker build when external registry connectivity is stable.

## 2026-05-09 GitHub Menu Visibility
### Done
- Confirmed the source dropdown in [frontend/src/components/layout/AppHeader.vue](/Users/aias/Work/github/sub2api/frontend/src/components/layout/AppHeader.vue) already intended the GitHub item to be admin-only, but the embedded frontend dist served by the local app was stale and still rendered it unconditionally.
- Added a focused regression check in [frontend/src/components/layout/__tests__/AppHeader.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/components/layout/__tests__/AppHeader.spec.ts) to lock both:
  - source intent (`showGithubLink` admin-only)
  - embedded dist behavior (admin-only compiled branch before the GitHub link)
- Updated `AppHeader.vue` to use an explicit `showGithubLink` computed guard.
- Patched the currently embedded frontend asset `backend/internal/web/dist/assets/AppLayout.vue_vue_type_script_setup_true_lang-Cyx7fIER.js` to match source behavior, then rebuilt and restarted the local backend on `http://127.0.0.1:18082`.
- Verified with a real regular user account (`regular-hide-gh@example.com`) in the Codex app browser that the user dropdown now contains only:
  - `个人资料`
  - `API 密钥`
  - `退出登录`
  and no `GitHub` item.

### Failures
- A clean frontend rebuild is still blocked locally because `frontend/node_modules` is missing `@stripe/stripe-js`, so the served fix had to be synchronized by patching the embedded dist directly instead of rerunning `vite build`.

### Next
- Restore the missing Stripe frontend dependency and rerun a clean frontend build so the generated dist can be refreshed from source instead of patched in place.

## 2026-05-09 Regular User Sidebar Trim
### Done
- Removed `API 密钥` and `个人资料` from the **regular user** sidebar only; admins still keep their personal account section links.
- Kept both entries in the top-right user dropdown for regular users.
- Moved the regular-user onboarding anchor from the hidden sidebar key item to the dashboard quick action button:
  - added `data-tour="dashboard-create-key-shortcut"` to the dashboard “创建 API 密钥” shortcut
  - updated `getUserSteps` to target the new shortcut
- Added file-level regression coverage in [frontend/src/components/layout/__tests__/AppSidebar.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/components/layout/__tests__/AppSidebar.spec.ts).
- Synced the currently served embedded frontend dist and rebuilt/restarted the local backend on `http://127.0.0.1:18082`.
- Verified in the Codex app browser with the regular user account `regular-hide-gh@example.com`:
  - sidebar now shows `仪表盘 / 调用说明 / 调用测试 / 使用记录 / 我的订阅 / 兑换`
  - sidebar no longer shows `API 密钥` or `个人资料`
  - user dropdown still shows `个人资料` and `API 密钥`
  - dashboard still shows the `创建 API 密钥` quick action button

### Failures
- A clean frontend rebuild is still blocked locally by the missing `@stripe/stripe-js` dependency in `frontend/node_modules`, so the embedded dist had to be patched directly again.

### Next
- Restore the missing Stripe dependency and rerun a full frontend build so future runtime changes can come from generated assets instead of manual dist synchronization.

## 2026-05-09 Dropdown vs Sidebar Clarification
### Done
- Reverted the earlier regular-user sidebar trim after clarifying the intended behavior.
- Restored `API 密钥` and `个人资料` in the regular user sidebar.
- Restored the regular-user onboarding anchor back to the sidebar key item (`[data-tour="sidebar-my-keys"]`) and removed the temporary dashboard shortcut tour anchor.
- Hid the duplicate `个人资料` and `API 密钥` entries from the **regular user avatar dropdown** instead.
- Kept admin-only dropdown items gated in the header (`个人资料` / `API 密钥` / `GitHub` stay admin-only there).
- Verified in the Codex app browser for `regular-hide-gh@example.com`:
  - left sidebar shows both `API 密钥` and `个人资料`
  - avatar dropdown now only shows `退出登录`
  - `/health` still returns `{"env":"production","status":"ok"}`

### Failures
- Full frontend rebuild is still blocked locally by the missing `@stripe/stripe-js` dependency, so the served embedded dist was synchronized manually again before rebuilding the backend binary.

### Next
- Restore the missing frontend dependency and rerun a clean frontend build so future UI changes can come from generated assets instead of direct dist synchronization.

## 2026-05-09 ApiGuide Dark Contrast
### Done
- Added shared source-side surface tokens in [frontend/src/style.css](/Users/aias/Work/github/sub2api/frontend/src/style.css) for:
  - `page-hero`
  - `page-kicker`
  - `sticky-panel`
  - `surface-panel`
  - `surface-panel-strong`
  - `surface-panel-muted`
  - `metric-panel`
  - `metric-icon`
- Strengthened dark-mode contrast for those surfaces and improved `btn-secondary` dark styling.
- Increased page-header subtitle contrast in [frontend/src/components/layout/AppHeader.vue](/Users/aias/Work/github/sub2api/frontend/src/components/layout/AppHeader.vue).
- Added regression coverage in [frontend/src/views/user/__tests__/apiGuideDarkContrast.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/views/user/__tests__/apiGuideDarkContrast.spec.ts).
- Synced the currently served embedded frontend assets:
  - [backend/internal/web/dist/assets/vendor-misc-DB0Q8XAf.css](/Users/aias/Work/github/sub2api/backend/internal/web/dist/assets/vendor-misc-DB0Q8XAf.css)
  - [backend/internal/web/dist/assets/AppLayout.vue_vue_type_script_setup_true_lang-Cyx7fIER.js](/Users/aias/Work/github/sub2api/backend/internal/web/dist/assets/AppLayout.vue_vue_type_script_setup_true_lang-Cyx7fIER.js)
- Rebuilt and restarted the local backend on `http://127.0.0.1:18082`.
- Verified in the Codex app browser on `/gateway-guide` that dark-mode hero text, metric cards, side panels, and the secondary button are visibly higher contrast than before.

### Failures
- Full frontend rebuild remains blocked locally by the missing `@stripe/stripe-js` dependency, so the embedded dist required manual synchronization again.

### Next
- Restore the missing Stripe dependency and rerun a clean frontend build so these source-side surface tokens become part of the generated dist without manual patching.

## 2026-05-09 Clean Rebuild Path
### Done
- Diagnosed the clean rebuild blocker into two separate issues:
  - the default `pnpm` on this machine was a Corepack shim that tried to download pnpm from npm before doing anything
  - the existing frontend install had been incomplete, leaving `@stripe/stripe-js` missing from `node_modules`
- Bypassed the Corepack shim by using the real Homebrew pnpm binary (`/opt/homebrew/bin/pnpm`).
- Added repo-local pnpm build-script allowlisting in [frontend/.npmrc](/Users/aias/Work/github/sub2api/frontend/.npmrc):
  - `only-built-dependencies[]=esbuild`
  - `only-built-dependencies[]=vue-demi`
- Reinstalled frontend dependencies from a clean state:
  - moved away the old `frontend/node_modules`
  - ran `/opt/homebrew/bin/pnpm --dir frontend install --offline --frozen-lockfile`
  - confirmed `esbuild` and `vue-demi` postinstall scripts executed successfully
  - removed the temporary backup after success
- Verified the stable clean-install path now works without warnings:
  - `/opt/homebrew/bin/pnpm --dir frontend install --offline --frozen-lockfile`
- Verified full clean rebuild from source:
  - `./node_modules/.bin/vue-tsc -b && ./node_modules/.bin/vite build`
  - build completed successfully and regenerated `backend/internal/web/dist`
- Rebuilt the backend with embedded frontend and restarted the local app on `http://127.0.0.1:18082`
- Verified runtime health after the clean rebuild with `/health -> {"env":"production","status":"ok"}`

### Failures
- Direct installs through the Corepack-managed `pnpm` shim still fail in this network environment because it attempts to download pnpm itself from `registry.npmjs.org` before running.
- Online package fetches against the default npm registry still time out intermittently in this environment, so the reliable path here is:
  1. use the real pnpm binary
  2. warm the store once if needed
  3. use offline frozen-lockfile installs afterward

### Next
- If you want, I can add a tiny repo-local `make frontend-install` / `scripts/frontend-install.sh` helper so the clean install/build path is one command instead of relying on remembering the Homebrew pnpm binary.

## 2026-05-09 Dashboard Hero Heading Size
### Done
- Reduced the dashboard welcome heading in [frontend/src/views/user/DashboardView.vue](/Users/aias/Work/github/sub2api/frontend/src/views/user/DashboardView.vue) from `text-3xl md:text-4xl` to `text-2xl md:text-3xl`.
- Added a focused regression check in [frontend/src/views/user/__tests__/dashboardHeroTypography.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/views/user/__tests__/dashboardHeroTypography.spec.ts).
- Rebuilt the frontend from source, rebuilt the backend with embedded assets, and restarted the local app on `http://127.0.0.1:18082`.
- Verified in the Codex app browser on `/dashboard` that the “欢迎回来！这是您账户的概览。” hero title is visibly smaller than before.

### Failures
- None for this change. The clean rebuild path restored earlier was sufficient for a full frontend rebuild and embed cycle.

### Next
- If needed, tune the hero typography one more step by making it responsive-only (`text-xl md:text-3xl`) or tightening only the desktop breakpoint.

## 2026-05-09 Hide Unconfigured Profile Bindings
### Done
- Updated [frontend/src/components/user/profile/ProfileIdentityBindingsSection.vue](/Users/aias/Work/github/sub2api/frontend/src/components/user/profile/ProfileIdentityBindingsSection.vue) so third-party provider cards render only when at least one of these is true:
  - the provider is already bound
  - the provider can currently be bound
  - the provider can currently be unbound
- Kept the email binding card always visible.
- Added regression coverage in [frontend/src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts) for:
  - unconfigured + unbound providers are hidden
  - already-bound providers remain visible even if the provider is currently not configured
- Rebuilt frontend from source, rebuilt backend with embedded assets, and restarted the local app on `http://127.0.0.1:18082`.
- Verified in the Codex app browser on `/profile` that the `LinuxDo` / `OIDC` / `微信` cards no longer appear for the current user, while the `邮箱` card remains.

### Failures
- None for this change after the visibility rule was tightened from “enabled or bound” to “bound / canBind / canUnbind”.

### Next
- If you want, I can apply the same “hide when unavailable” rule to any other profile-side integration panels that currently show inert placeholders.

## 2026-05-09 Hide Unopened TOTP
### Done
- Confirmed the backend public settings payload already includes `totp_enabled`; the frontend had simply not wired it into the profile page.
- Added `totp_enabled` to [frontend/src/types/index.ts](/Users/aias/Work/github/sub2api/frontend/src/types/index.ts) and the app-store fallback object in [frontend/src/stores/app.ts](/Users/aias/Work/github/sub2api/frontend/src/stores/app.ts).
- Updated [frontend/src/views/user/ProfileView.vue](/Users/aias/Work/github/sub2api/frontend/src/views/user/ProfileView.vue) to load `settings.totp_enabled` and render `ProfileTotpCard` only when that flag is enabled.
- Added page-level regression coverage in [frontend/src/views/user/__tests__/ProfileView.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/views/user/__tests__/ProfileView.spec.ts) for:
  - hidden when `totp_enabled` is false
  - shown when `totp_enabled` is true
- Rebuilt frontend from source, rebuilt backend with embedded assets, and restarted the local app on `http://127.0.0.1:18082`.
- Verified in the Codex app browser on `/profile` that the entire `双因素认证 (2FA)` panel is gone when the feature is not opened.

### Failures
- None for this change.

### Next
- If you want the same treatment elsewhere, I can sweep the profile page for any other “功能未开放” placeholder blocks and convert them to hidden-by-default.

## 2026-05-09 Purchase Page Catalog
### Done
- Confirmed `/purchase` is the actual “充值/订阅” page and `/subscriptions` remains the “我的订阅” list page.
- Added backend support for configurable recharge products via payment config:
  - new payment-config setting key in [backend/internal/service/payment_config_service.go](/Users/aias/Work/github/sub2api/backend/internal/service/payment_config_service.go)
  - parsed + normalized into `PaymentConfig.RechargeProducts`
  - returned from `/api/v1/payment/checkout-info` in [backend/internal/handler/payment_handler.go](/Users/aias/Work/github/sub2api/backend/internal/handler/payment_handler.go)
- Added admin settings schema support for `payment_recharge_products`:
  - backend DTO/handler wiring
  - frontend admin settings types
  - settings form state + editor UI in [frontend/src/views/admin/SettingsView.vue](/Users/aias/Work/github/sub2api/frontend/src/views/admin/SettingsView.vue)
- Added frontend payment types support for `RechargeProduct` and checkout payload typing.
- Reworked [frontend/src/views/user/PaymentView.vue](/Users/aias/Work/github/sub2api/frontend/src/views/user/PaymentView.vue):
  - recharge tab now prefers configured product cards over the legacy quick-amount matrix
  - if no recharge products are configured, it still falls back to the old amount input
  - subscription tab continues using existing admin-configured plans
- Added [frontend/src/components/payment/RechargeProductCard.vue](/Users/aias/Work/github/sub2api/frontend/src/components/payment/RechargeProductCard.vue) for the card-style recharge catalog.
- Added/updated tests:
  - [backend/internal/service/payment_config_service_test.go](/Users/aias/Work/github/sub2api/backend/internal/service/payment_config_service_test.go)
  - [frontend/src/views/user/__tests__/PaymentView.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/views/user/__tests__/PaymentView.spec.ts)
- Seeded the local instance with:
  - payment enabled
  - one dummy EasyPay provider (for page visibility only)
  - six recharge products
  - one subscription plan
- Verified in the Codex app browser:
  - `/purchase` recharge tab shows one card per configured recharge product
  - `/purchase` subscription tab shows the admin-configured subscription plan card

### Failures
- No blocking implementation failures after the config chain was added.
- Vite still reports existing chunk-splitting warnings; they do not block build or runtime for this change.

### Next
- If needed, further align `/purchase` visual details with the reference by tightening tab styling, card spacing, and header copy once you review this first pass in-browser.

## 2026-05-10 TOTP Issuer Domain
### Done
- Traced the authenticator display label to [backend/internal/service/totp_service.go](/Users/aias/Work/github/sub2api/backend/internal/service/totp_service.go), where the TOTP issuer had been hard-coded as `Sub2API`.
- Changed TOTP setup so the issuer now prefers the configured `frontend_url` host, which matches the current server domain when that setting is present.
- Kept fallbacks in place:
  - `site_name` when `frontend_url` is empty or unparsable
  - `Sub2API` as the final default
- Added focused unit coverage in [backend/internal/service/totp_service_test.go](/Users/aias/Work/github/sub2api/backend/internal/service/totp_service_test.go) for:
  - frontend URL host priority
  - site name fallback
  - default fallback

### Failures
- None for this change.

### Next
- Push and deploy this issuer update to the server, then re-bind 2FA once so the authenticator entry is regenerated with the domain-based issuer.

## 2026-05-10 Dashboard Hero And User Dropdown Cleanup
### Done
- Removed the blank middle section and extra divider from the regular user avatar dropdown in [frontend/src/components/layout/AppHeader.vue](/Users/aias/Work/github/sub2api/frontend/src/components/layout/AppHeader.vue).
- Kept the richer separated dropdown layout for admin/support/onboarding cases by making the compact variant conditional instead of flattening every menu state.
- Removed the small `仪表盘` eyebrow label and the user email line from the dashboard hero in [frontend/src/views/user/DashboardView.vue](/Users/aias/Work/github/sub2api/frontend/src/views/user/DashboardView.vue), leaving just the site kicker and welcome headline.
- Added/updated focused regression checks:
  - [frontend/src/components/layout/__tests__/AppHeader.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/components/layout/__tests__/AppHeader.spec.ts)
  - [frontend/src/views/user/__tests__/dashboardHeroTypography.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/views/user/__tests__/dashboardHeroTypography.spec.ts)
- Rebuilt the frontend so embedded assets under `backend/internal/web/dist` match the source changes.

### Failures
- The existing AppHeader dist assertion was too brittle against current Vite minification output, so it was relaxed to a stable generated-asset check before proceeding.

### Next
- Push and deploy the updated frontend bundle to the server, then visually re-check `/dashboard` and the user dropdown on the live domain.

## 2026-05-10 CI Contract And Formatting Fixes
### Done
- Investigated GitHub Actions run `25622045409` and confirmed two real failures:
  - `golangci-lint` rejected [backend/internal/service/payment_config_service_test.go](/Users/aias/Work/github/sub2api/backend/internal/service/payment_config_service_test.go) because it was not `gofmt`-formatted.
  - unit tests failed in [backend/internal/server/api_contract_test.go](/Users/aias/Work/github/sub2api/backend/internal/server/api_contract_test.go) because the admin settings contract still expected the old payload shape without `payment_recharge_products`.
- Added `payment_recharge_products: []` to the affected admin settings contract expectations.
- Ran `gofmt` on the touched Go test files so lint and CI formatting checks align with the repo rules.
- Re-ran the CI-equivalent backend unit entrypoint with `make test-unit`, and it completed successfully locally.

### Failures
- Local `golangci-lint` binary is built with Go 1.25 while the repo targets Go 1.26.3, so I relied on the concrete CI failure log plus `gofmt` remediation instead of a local lint run.

### Next
- Push the fixes and watch the new GitHub Actions run to completion. If anything else fails, continue iterating on that run instead of guessing.

## 2026-05-10 Frontend Docs Framework
### Done
- Added a built-in frontend docs route at `/docs/:pathMatch(.*)*` in [frontend/src/router/index.ts](/Users/aias/Work/github/sub2api/frontend/src/router/index.ts).
- Introduced a lightweight docs framework using existing markdown tooling instead of a new dependency stack:
  - docs registry + link resolution in [frontend/src/utils/docs.ts](/Users/aias/Work/github/sub2api/frontend/src/utils/docs.ts)
  - reusable smart link wrapper in [frontend/src/components/common/DocsLink.vue](/Users/aias/Work/github/sub2api/frontend/src/components/common/DocsLink.vue)
  - docs shell/page in [frontend/src/views/public/DocsView.vue](/Users/aias/Work/github/sub2api/frontend/src/views/public/DocsView.vue)
- Seeded initial docs content in:
  - [frontend/src/docs/pages/getting-started/overview.md](/Users/aias/Work/github/sub2api/frontend/src/docs/pages/getting-started/overview.md)
  - [frontend/src/docs/pages/getting-started/quick-start.md](/Users/aias/Work/github/sub2api/frontend/src/docs/pages/getting-started/quick-start.md)
  - [frontend/src/docs/pages/guides/api-keys.md](/Users/aias/Work/github/sub2api/frontend/src/docs/pages/guides/api-keys.md)
  - [frontend/src/docs/pages/guides/gateway.md](/Users/aias/Work/github/sub2api/frontend/src/docs/pages/guides/gateway.md)
  - [frontend/src/docs/pages/guides/billing.md](/Users/aias/Work/github/sub2api/frontend/src/docs/pages/guides/billing.md)
  - [frontend/src/docs/pages/guides/security.md](/Users/aias/Work/github/sub2api/frontend/src/docs/pages/guides/security.md)
- Rewired existing docs entry points in the authenticated header, home page, and key-usage page so same-origin `doc_url` values like `/`, `/home`, or `/docs` now resolve to the internal docs route instead of bouncing back to home.
- Added focused regression coverage:
  - [frontend/src/utils/__tests__/docs.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/utils/__tests__/docs.spec.ts)
  - [frontend/src/router/__tests__/docsRoute.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/router/__tests__/docsRoute.spec.ts)
- Rebuilt the frontend bundle so `backend/internal/web/dist` includes the docs route and link changes.

### Failures
- No blocking implementation failures after switching to the built-in markdown-based docs approach.

### Next
- Push and deploy the docs route to the server, then verify that clicking the header docs button on the live domain lands on `/docs` instead of redirecting to `/home`.

## 2026-05-10 Docsify Migration
### Done
- Replaced the custom in-app markdown docs renderer with a Docsify-backed shell in [frontend/src/views/public/DocsView.vue](/Users/aias/Work/github/sub2api/frontend/src/views/public/DocsView.vue).
- Added explicit frontend docs framework dependencies in [frontend/package.json](/Users/aias/Work/github/sub2api/frontend/package.json) and updated [frontend/pnpm-lock.yaml](/Users/aias/Work/github/sub2api/frontend/pnpm-lock.yaml) so the runtime assets resolve locally at build time.
- Moved docs content from the previous source-only registry into Docsify static markdown files under:
  - [frontend/public/docs-content/README.md](/Users/aias/Work/github/sub2api/frontend/public/docs-content/README.md)
  - [frontend/public/docs-content/_sidebar.md](/Users/aias/Work/github/sub2api/frontend/public/docs-content/_sidebar.md)
  - [frontend/public/docs-content/getting-started/quick-start.md](/Users/aias/Work/github/sub2api/frontend/public/docs-content/getting-started/quick-start.md)
  - [frontend/public/docs-content/guides/api-keys.md](/Users/aias/Work/github/sub2api/frontend/public/docs-content/guides/api-keys.md)
  - [frontend/public/docs-content/guides/gateway.md](/Users/aias/Work/github/sub2api/frontend/public/docs-content/guides/gateway.md)
  - [frontend/public/docs-content/guides/billing.md](/Users/aias/Work/github/sub2api/frontend/public/docs-content/guides/billing.md)
  - [frontend/public/docs-content/guides/security.md](/Users/aias/Work/github/sub2api/frontend/public/docs-content/guides/security.md)
- Simplified [frontend/src/utils/docs.ts](/Users/aias/Work/github/sub2api/frontend/src/utils/docs.ts) to only keep docs-link normalization and Docsify hash normalization.
- Kept the existing `/docs` route and existing docs-entry buttons, but changed same-origin docs links to land on the Docsify route instead of external/home redirects.
- Added Docsify-focused regression checks:
  - [frontend/src/views/public/__tests__/DocsView.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/views/public/__tests__/DocsView.spec.ts)
  - [frontend/src/utils/__tests__/docs.spec.ts](/Users/aias/Work/github/sub2api/frontend/src/utils/__tests__/docs.spec.ts)

### Failures
- None after switching from the custom markdown registry to the Docsify shell.

### Next
- Push and deploy the Docsify version to the server, then verify that the live docs button lands on `/docs` and the page loads Docsify content correctly.

## 2026-05-12 Stripe Webhook Non-Payment Events
### Done
- Investigated Stripe retry email for `https://cloudbase.eu.org/api/v1/payment/webhook/stripe` and confirmed the provided failing payload is `capability.updated`, not a payment completion/failure event.
- Updated [backend/internal/payment/provider/stripe.go](/Users/aias/Work/github/sub2api/backend/internal/payment/provider/stripe.go) so unsupported non-payment Stripe events are acknowledged as ignored before order fulfillment logic runs.
- Kept `payment_intent.succeeded` and `payment_intent.payment_failed` on the strict signed webhook path so order-affecting events still require the configured Stripe webhook secret.
- Added focused regression coverage in [backend/internal/payment/provider/stripe_test.go](/Users/aias/Work/github/sub2api/backend/internal/payment/provider/stripe_test.go).

### Failures
- The server's bundled `psql` client cannot currently run because `libpq.so.5` is missing, so provider secret inspection through direct SQL was not completed.

### Next
- Deploy the Stripe event ACK fix to the server and verify the exact `capability.updated` payload now returns `200`; separately confirm the Stripe provider has the live endpoint signing secret configured for real payment events.

## 2026-05-13 Frontend Purchase Flow Follow-Up
### Done
- Ran a live UI sweep across `/dashboard`, `/purchase`, `/profile`, `/docs`, and auth pages on `https://cloudbase.eu.org`.
- Confirmed no current horizontal overflow, raw i18n key leak, or console error on the refreshed page sweep.
- Found that `/purchase` still falls back to the legacy quick amount/custom amount UI when no recharge products are configured, contradicting the product-card catalog direction.
- Updated the user purchase page so recharge now only renders configured product cards; an empty product catalog shows a proper empty state instead of legacy amount inputs.
- Updated the admin settings copy so an empty recharge product list describes the new empty-catalog behavior.
- Verified the change with the focused `PaymentView` test suite, `vue-tsc --noEmit`, a production Vite build, and a local Vite preview proxied to the live API.

### Failures
- The Codex in-app browser automation channel timed out twice, so the sweep used the available Playwright browser context instead.
- A temporary rendered component assertion was removed because `PaymentView`'s existing shallow `AppLayout` stubbing does not render default slots reliably; the existing source regression plus browser preview covered the behavior instead.

### Next
- Commit, push, deploy to the server, then re-check live `/purchase` for the empty catalog state.

## 2026-05-13 Docs Content Domain Portability
### Done
- Replaced production-domain links inside `frontend/public/docs-content/**/*.md` with same-origin routes such as `/home`, `/dashboard`, `/register`, and `/purchase`.
- Added `DocsContent` regression coverage so Markdown docs cannot reintroduce `cloudbase.eu.org`.
- Added a Docsify content cache-buster built at frontend build time, rewrote Docsify hash links with `_docs_v`, isolated the search index by version, and sent no-cache headers for Markdown requests.
- Corrected Docsify's rewriting of app-route links so `/home`, `/dashboard`, `/register`, and `/purchase` remain application routes instead of document hashes.
- Switched Docsify link post-processing to the stable `.docsify-shell` container because Docsify replaces the original Vue ref element during initialization.

### Failures
- First live verification showed the server content was updated but the browser still rendered stale Docsify Markdown from the old page cache.
- Second live verification showed Docsify converted same-origin app links into `#/...` document routes; corrected before final deployment.

### Next
- Commit, push, deploy the Docsify cache-busting fix, then verify the live docs page no longer displays hard-coded production-domain links.

## 2026-05-13 User Dashboard Hero Removal
### Done
- Removed the normal-user dashboard overview hero so the page starts directly with the main dashboard stats, charts, recent usage, and quick actions.
- Replaced the obsolete hero typography regression with a no-hero layout regression.
- Verified the focused dashboard regression and a production frontend build locally.

### Failures
- None.

### Next
- Commit, push, deploy to the server, then verify live `/dashboard` no longer shows the redundant overview hero for normal users.

## 2026-05-13 Email Template Refresh
### Done
- Added a shared backend email shell inspired by the Cursor one-time-code email: white bordered card, compact logo block, large title, primary code/action, divider, and muted safety copy.
- Migrated verification code, password reset, notification email verification, balance low, quota alert, ops alert, content moderation, and SMTP test email bodies to the shared shell.
- Added focused unit coverage for the shared template and updated existing notification email assertions.
- Verified template tests, service/handler tests, and backend server build locally.

### Failures
- `GOTOOLCHAIN=local go build` failed on this machine because local Go is 1.25.5 while the project requires Go 1.26.3; `go build` with automatic toolchain selection passed.

### Next
- Commit, push, deploy to the server, then send or inspect a live test email from admin settings if needed.

## 2026-05-13 Email Template Polish
### Done
- Tightened the shared email card spacing, typography, divider, and main value/code hierarchy to better match the Cursor reference.
- Replaced the SVG email logo with a pure HTML/CSS mark so real mail clients are less likely to strip it.
- Generated local browser previews for test and verification-code emails.

### Failures
- The first local preview was too narrow; adjusted the card width back up while keeping the vertical rhythm tighter.
- The initial code preview used overly loose monospace digits; switched to system digits with tighter tracking.

### Next
- Commit, push, deploy to the server, then use the admin SMTP test email if a real mailbox render check is needed.

## 2026-05-13 Welcome Email And Subscription Links
### Done
- Added user-level welcome email and marketing email unsubscribe timestamps.
- Added a welcome email template with uploaded site logo support, onboarding CTA, and Manage subscriptions / Unsubscribe footer links.
- Added signed public email preference endpoints for manage, unsubscribe, and resubscribe actions.
- Wired first-time email registration, verified Google/GitHub OAuth signup, pending OAuth finalization, and legacy OAuth signup paths to enqueue one welcome email.
- Generated ent code and verified focused service/handler tests plus backend server build.

### Failures
- `make generate` completed ent generation but failed at Wire because the local `go.sum` lacks `github.com/google/subcommands` for the Wire tool path; this feature did not require Wire regeneration.
- Initial handler test exposed a transaction lock in the welcome-email dedupe update; fixed by reusing the current ent transaction client.

### Next
- Commit, push, deploy to the server, then create a fresh test account to verify the real welcome email and preference links in production.

## 2026-05-13 Welcome Email Deployment
### Done
- Pushed the welcome email and subscription preference implementation to the `aias00/main` remote.
- Pulled commit `f66a58a9` on the Google server, built the backend binary on the server with Go 1.26.3, and replaced the live `sub2api` binary.
- Restarted the live process with the required `SERVER_PORT=8081` environment and verified `https://cloudbase.eu.org/api/v1/settings/public` returns 200.
- Verified the new public preference route `https://cloudbase.eu.org/api/v1/email-preferences/manage?token=bad` reaches the app and returns the expected invalid-token HTML response.

### Failures
- The first manual restart omitted `SERVER_PORT=8081`, so the app listened on the default `8080` and Cloudflare returned 502 until it was restarted on the correct port.
- Pushing to the `origin` remote failed with 403 because the current GitHub credentials cannot write to `Wei-Shaw/sub2api`; pushed to `aias00/main` instead.

### Next
- Create a fresh test account on production to verify a real welcome email is delivered and that Manage subscriptions / Unsubscribe links mutate the account preference as expected.

## 2026-05-14 Affiliate Invitation Gate
### Done
- Investigated invite-only registration, reusable affiliate invite codes, OAuth first-login, and rebate binding paths.
- Found that invitation-only mode accepted only one-time redeem invitation codes, so a valid friend invite link could not onboard a new user.
- Added a shared registration gate that still prefers one-time invitation codes, but accepts enabled affiliate invite codes as reusable onboarding credentials and preserves inviter binding for rebates.
- Updated email registration and Google verified-email OAuth auto-registration so a valid affiliate invite link can pass the gate without a separate invitation code.
- Added UI hints so users arriving through a friend invite link are not blocked by an empty invitation-code field.
- Verified focused backend service/handler tests, full backend unit tests, frontend typecheck, production frontend build, and affected OAuth frontend tests.

### Failures
- Accidentally overwrote `progress.md` with only the new entry during the first update; restored the tracked history and appended this entry instead.

### Next
- Commit, push, deploy to the server, then verify the live invitation-only affiliate signup path.

## 2026-05-14 Multi SMTP Channel Fallback
### Done
- Started implementation for multiple SMTP channels with per-channel daily limits.
- Chose settings-backed JSON configuration to avoid introducing a new database table while preserving existing primary SMTP compatibility.
- Added backend channel parsing, password preservation, primary-channel daily limit, and fallback-send selection.
- Added the admin settings form controls for primary daily limit and fallback SMTP channels.

### Failures
- None so far.

### Next
- Run backend unit tests, frontend typecheck/build, then fix any regressions before deployment.

## 2026-05-14 Docs Layout Spacing
### Done
- Investigated the large desktop gap between the Docsify sidebar and documentation content.
- Removed the redundant desktop `margin-left` on Docsify content after the shell had already placed sidebar/content with flex layout.
- Tightened the desktop sidebar width and changed the Markdown content column from centered to left-aligned inside the content area.
- Verified the production frontend build and local preview layout: sidebar width dropped from 272px to 248px, and Markdown content now starts at the content column instead of leaving an extra centered gap.
- Follow-up: made Docsify sidebar active links explicit with a blue pill, left accent, and dark-mode-safe text color so selected menu entries stay visible after clicking.
- Follow-up: fixed nested Docsify page navigation by aliasing every nested `_sidebar.md` request back to the root `_sidebar.md`, avoiding SPA fallback HTML replacing the sidebar menu.
- Follow-up: disabled Docsify relative path resolution so root sidebar links remain root-relative when switching between nested quickstart/advanced pages.
- Follow-up: disabled Docsify sidebar auto-heading injection and added path-based active-link syncing so clicking nested left-menu entries keeps the selected item visible in the stable root navigation.

### Failures
- Nested docs pages initially loaded the right Markdown body but fetched a missing nested sidebar file; the backend SPA fallback returned `index.html`, which corrupted the sidebar content.
- With only the sidebar alias applied, Docsify still resolved sidebar links relative to the current nested page and produced duplicated routes such as `quickstart/quickstart/gpt-image-2`.
- After the alias and path fix, Docsify still treated the root overview link as active on nested pages and inserted current-page headings under it; the fix now avoids Docsify's generated sidebar subsection for this layout and marks the exact sidebar route manually.

### Next
- Commit, push, deploy, and verify the nested docs sidebar fix on the live site.

## 2026-05-14 Aura Theme Refresh
### Done
- Started a global visual refresh using the provided Brevo/Aura references: warm off-white surfaces, lavender backgrounds, deep-purple primary actions, and black-forward typography.
- Changed the Tailwind primary palette from deep blue to an Aura purple scale, with a muted green accent for sparse brand highlights.
- Updated shared global UI tokens for buttons, cards, panels, sidebar active states, page hero surfaces, glass surfaces, and the app background mesh.
- Restyled the Docsify selected sidebar item into a rounded outlined lavender pill and aligned docs headings/search/borders with the new palette.

### Failures
- None so far.

### Next
- Run frontend validation, inspect the docs/dashboard visuals locally, then commit, push, deploy, and verify production.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed; only existing Vite chunk/import warnings were reported.
- Local browser preview checked `/docs#/cloudbase-guide` and `/login`; the docs active item now renders as a rounded lavender outline pill and the public auth page uses the updated purple/off-white theme.

## 2026-05-14 Fixed Light Theme
### Done
- Removed user-facing theme toggle controls from the main sidebar, home page, and key usage page.
- Changed app bootstrap to always remove the `dark` class and persist `theme=light`, so cached or OS dark-mode preferences no longer switch the UI theme.

### Failures
- None so far.

### Next
- Run frontend validation, inspect the sidebar/home controls locally, then deploy to production.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed; only existing Vite chunk/import warnings were reported.
- Local browser preview forced `localStorage.theme=dark` and reloaded `/home`; the root `dark` class was removed, `theme` was reset to `light`, and no sun/moon theme toggle button remained.

## 2026-05-14 Solid App Background
### Done
- Removed the global app background decoration layer that added lavender/green mesh gradients behind every authenticated page.
- Removed the unused Tailwind `mesh-gradient` token so the colored page backdrop is not reintroduced accidentally through that utility.

### Failures
- None so far.

### Next
- None.

### Validation
- `rg` confirmed `bg-mesh-gradient` / `mesh-gradient` are no longer referenced.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed; only existing Vite chunk/import warnings were reported.
- Deployed commit `ac7b0c32` to production; `https://cloudbase.eu.org/api/v1/settings/public` and `/dashboard` returned HTTP 200.
- Browser verification on `/dashboard?verify=ac7b0c32` found `meshClassCount=0`, solid shell background `rgb(247, 246, 241)`, and no console errors.

## 2026-05-14 Empty State Polish
### Done
- Started the first local-first optimization pass with a narrow scope: common empty states and the user channel status empty page.
- Updated the shared `EmptyState` component to support a framed `panel` presentation while preserving the default plain presentation for table slots.
- Fixed the legacy `message` prop so existing `<EmptyState :message="...">` usages display their message instead of falling back to the generic no-data title.
- Changed the channel status empty view to use the framed empty state, reducing the impression of an oversized blank page.

### Failures
- None so far.

### Next
- Continue with the next local-first optimization item.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed; only existing Vite chunk/import warnings were reported.
- Local browser verification on `http://127.0.0.1:4173/monitor?localVerify=empty-state-clean` used mocked API responses and found one framed empty state (`672x274`) with no console errors.
- Deployed commit `b1ef8bc4` to production; `https://cloudbase.eu.org/api/v1/settings/public` and `/monitor` returned HTTP 200.
- Browser verification on `https://cloudbase.eu.org/monitor?verify=b1ef8bc4` found one framed empty state (`672x274`) and no console errors.

## 2026-05-14 Docs Sidebar Polish
### Done
- Started the second local-first optimization pass for Docsify navigation.
- Removed Docsify sidebar list marker clutter by normalizing sidebar list spacing and list styles.
- Changed active sidebar links to a full-row pill style so the selected document item stays visible and easier to scan.
- Added sidebar-only scroll correction after active-link sync so clicking deeper documentation items keeps the selected entry inside the left navigation viewport.
- Corrected the active-link scroll target from the fixed sidebar shell to Docsify's actual scrollable `.sidebar-nav` container.

### Failures
- First production check on `a86c9ae2` showed the active pill style was applied, but the selected deep link was still below the visible sidebar because `.sidebar-nav`, not `.sidebar`, owns scrolling.

### Next
- Continue with the next local-first optimization item.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed; only existing Vite chunk/import warnings were reported.
- Local browser verification on `http://127.0.0.1:18084/docs?localVerify=docs-sidebar-clean#/advanced/vscode-gui` found marker style `none`, active link display `flex`, selected item visible inside the sidebar, and a 40px sidebar-to-content gap with no console errors.
- After the production scroll-target issue was fixed locally, `pnpm --dir frontend run typecheck` and `pnpm --dir frontend run build` passed again.
- Local browser verification on `http://127.0.0.1:18084/docs?localVerify=docs-scroll-target#/advanced/vscode-gui` found marker style `none`, active link display `flex`, selected item visible inside `.sidebar-nav`, `sidebarNavScrollTop=249.5`, and a 40px sidebar-to-content gap with no console errors.
- Deployed commit `15f8d89b` to production; `https://cloudbase.eu.org/api/v1/settings/public`, `/docs`, and `/monitor` returned HTTP 200.
- Browser verification on `https://cloudbase.eu.org/docs?verify=15f8d89b#/advanced/vscode-gui` found title `VS Code 图形化操作教程`, active link text `VS Code 图形化操作教程`, marker style `none`, active display `flex`, active item visible inside `.sidebar-nav`, `sidebarNavScrollTop=249.5`, a 40px sidebar-to-content gap, and no console errors.

## 2026-05-14 Docs Deep Link Preservation
### Done
- Started the next local-first docs interaction pass after production sidebar verification.
- Preserved existing Docsify hash deep links when `/docs` is opened directly, instead of always replacing the initial hash with the README route.
- Added a source-level regression check for preserving initial Docsify hash routes.

### Failures
- None so far.

### Next
- Continue with the next local-first optimization item.

### Validation
- `pnpm --dir frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts src/utils/__tests__/docs.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed; only existing Vite chunk/import warnings were reported.
- Local browser verification on `http://127.0.0.1:18084/docs?localVerify=docs-deep-link-final#/advanced/vscode-gui` found title `VS Code 图形化操作教程`, hash preserved as `#/advanced/vscode-gui?...`, active link text `VS Code 图形化操作教程`, selected item visible inside `.sidebar-nav`, `sidebarNavScrollTop=249`, and no console errors.
- Deployed commit `a1214648` to production; `https://cloudbase.eu.org/api/v1/settings/public` and `/docs` returned HTTP 200.
- `/monitor` returned HTTP 200 but the full body download exceeded the 20s curl limit, so page weight/transfer time remains a separate follow-up optimization candidate.
- Browser verification on `https://cloudbase.eu.org/docs?verify=a1214648#/advanced/vscode-gui` found title `VS Code 图形化操作教程`, hash preserved as `#/advanced/vscode-gui?...`, active link text `VS Code 图形化操作教程`, selected item visible inside `.sidebar-nav`, `sidebarNavScrollTop=249`, and no console errors.

## 2026-05-14 Sidebar Version Badge Removal
### Done
- Started a narrow production cleanup pass for the admin sidebar version badge.
- Removed the sidebar `VersionBadge` mount point so the visible version pill/dropdown no longer appears in the navigation header.
- Removed the sidebar import and `siteVersion` binding so mounting the sidebar no longer triggers the update-check component path.
- Updated the source-level sidebar regression test to lock the version badge out of `AppSidebar`.

### Failures
- The first server deploy command for the previous admin settings fix lost SSH during the frontend build before restart; the running service was not restarted from that interrupted command.

### Next
- None.

### Validation
- `pnpm --dir frontend exec vitest run src/components/layout/__tests__/AppSidebar.spec.ts` passed.
- `pnpm --dir frontend exec eslint src/components/layout/AppSidebar.vue src/components/layout/__tests__/AppSidebar.spec.ts` passed.
- `pnpm --dir frontend exec vue-tsc --noEmit` passed.
- `pnpm --dir frontend run build` passed; only existing Vite chunk/import warnings were reported.
- Deployed commit `deb60f2c` to production with server-side frontend and backend builds; local server health returned `{"env":"production","status":"ok"}`.
- `https://cloudbase.eu.org/health`, `https://cloudbase.eu.org/api/v1/settings/public`, and `/dashboard` returned HTTP 200.
- Browser verification on `https://cloudbase.eu.org/dashboard?verify=deb60f2c` found no `v0.1.x` text in the page/sidebar, no `/check-updates` network request, and no console warnings/errors.

## 2026-05-14 Channel Monitor Health Layer
### Done
- Started the New API-inspired channel health enhancement pass.
- Added monitor-level health state fields for automatic disable/recover tracking without turning off the scheduled probes.
- Added error classification to channel monitor histories for auth, rate limit, quota, server, network, timeout, challenge, slow response, empty response, invalid request, and unknown failures.
- Added health snapshot computation for recent success rate, average latency, consecutive failed/successful check runs, top error categories, and latest failure context.
- Added admin API exposure for health state and a dedicated `GET /api/v1/admin/channel-monitors/:id/health` endpoint.
- Wired auto-disable/auto-recover transitions into the existing Ops Alert event stream so monitor health incidents are visible in the admin alert center.
- Added focused unit coverage for error classification and health snapshot derivation.

### Failures
- Initially overwrote the existing `progress.md` while creating the task record; restored the original history and appended this section only.

### Next
- Optionally add a management UI panel for the health snapshot and route these Ops Alert events through email/Feishu/DingTalk delivery if product-level push notifications are required.

### Validation
- `go test -tags unit ./internal/service -run 'Test(ClassifyMonitorError|BuildChannelMonitorHealthSnapshot|RunCheckForModel)'` passed.
- `go test ./internal/service ./internal/handler/admin ./internal/server ./internal/repository` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed with existing Vite chunk warnings.
- `go test ./...` passed.

## 2026-05-15 Per-Channel SMTP Test Email
### Done
- Added per-fallback SMTP channel test email support in the admin settings UI.
- Extended the admin test-email API with optional `smtp_channel_id` so a selected saved channel can reuse its stored password when the password field is intentionally left blank.
- Allowed disabled fallback SMTP channels to be selected for admin test sends, so a channel can be verified before enabling production fallback use.
- Added frontend and backend regression coverage for testing a selected fallback SMTP channel.

### Failures
- An initial frontend typecheck caught the global send-test button passing the click event into the optional channel parameter; fixed by calling `sendTestEmail()` explicitly.
- An initial backend test command was run from the repository root, but the Go module lives in `backend/`; reran from the correct directory successfully.

### Next
- Deploy only when requested; current changes are local and verified.

### Validation
- `go test ./internal/handler/admin ./internal/service -run 'Test.*(Setting|Email|SMTP)'` passed from `backend/`.
- `go test ./...` passed from `backend/`.
- `pnpm --dir frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts` passed; existing `router-link` warnings remain in this test file.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed; only existing Vite chunk/import warnings were reported.
- `git diff --check` passed.

## 2026-05-17 Products And Plans Catalog
### Done
- Started a local-first admin commerce management pass.
- Reworked the existing admin subscription plan page into a unified 商品/套餐 management entry.
- Added a recharge product manager that edits `PAYMENT_RECHARGE_PRODUCTS` through the settings API, with add/delete/edit, feature list editing, recommended badge, preview, dirty-state tracking, and a single save action.
- Kept subscription plan CRUD on the existing `/admin/payment/plans` API path and preserved the existing group-bound plan validation flow.
- Renamed the admin sidebar route from 订阅套餐 / Plans to 商品/套餐 / Products & Plans.
- Added source-level regression coverage for the combined page, settings-backed recharge products, and navigation labels.

### Failures
- None.

### Next
- Continue with focused admin commerce UI polish after the live 商品/套餐 page is reviewed in an authenticated admin session.

### Validation
- `pnpm --dir frontend exec vitest run src/views/admin/__tests__/AdminPaymentCatalogView.spec.ts` passed, including a component-level smoke test for rendering the combined catalog shell and switching to the recharge product manager.
- `pnpm --dir frontend exec vitest run src/views/admin/__tests__/AdminPaymentCatalogView.spec.ts src/views/admin/__tests__/SettingsView.spec.ts` passed; existing `router-link` warnings remain in `SettingsView.spec.ts`.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/admin/orders/AdminPaymentPlansView.vue src/views/admin/orders/RechargeProductsManager.vue src/views/admin/__tests__/AdminPaymentCatalogView.spec.ts --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed; only existing Vite chunk/import warnings were reported.
- `go test ./internal/handler/admin ./internal/service -run 'Test.*(Setting|Email|SMTP)'` passed from `backend/`.
- `go test ./...` passed from `backend/`.
- Full local browser smoke test against a real admin session was not available because the backend/API target on `127.0.0.1:18082` refused the connection; a standalone Vite run on `18083` confirmed routing reaches the app and then correctly redirects without local auth/backend.
- Pushed implementation commit `4d6dced9` to `aias00/main`.
- Deployed commit `4d6dced9` to the Google server with the server-side build flow:
  - `pnpm --dir frontend install --frozen-lockfile`
  - `pnpm --dir frontend run build`
  - `go build -tags embed -o ../sub2api.new ./cmd/server`
- Restarted production on `SERVER_PORT=8081` with process `925212`; local server health returned `{"env":"production","status":"ok"}`.
- `https://cloudbase.eu.org/health`, `https://cloudbase.eu.org/api/v1/settings/public`, `/admin/orders/plans`, and `/login` returned HTTP 200 through Cloudflare.

## 2026-05-18 Commercial Legal Templates
### Done
- Added a reusable commercial login-agreement template generator for:
  - 商业服务条款
  - 使用政策
  - 支持的国家和地区
  - 服务特定条款
- Added a one-click “套用商业条款模板” action in admin settings so non-technical operators can refresh the legal document set without editing raw Markdown line-by-line.
- Reworked public legal document rendering so effective date / updated date / contact info / site URL are resolved at render time instead of being permanently baked into stored Markdown.
- Expanded `supported-regions` from a thin policy shell into a fuller operations-facing regional policy with:
  - general support regions by geography
  - exception handling
  - review requirements
  - dynamic restriction notes
- Normalized the current local legal-document records so the public `/legal/*` routes match the new template structure.

### Failures
- The first local pass still showed stale embedded frontend assets because the backend binary had not yet been rebuilt and restarted after the template changes.
- The first dynamic legal-content implementation used `replaceAll` and a missing public-settings field, which failed frontend typecheck; replaced with compatibility-safe string handling and legal-view local resolution.

### Next
- Commit, push, deploy the legal template refresh to the server, then verify production `/legal/terms` and `/legal/supported-regions`.

### Validation
- `pnpm --dir frontend exec vitest run src/utils/__tests__/loginAgreementTemplates.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- Local browser verification confirmed `/legal/terms` now shows one title, dynamic dates, and no hardcoded localhost URL in the正文地址说明。

## 2026-05-18 Production Settings Partial-Update Fix
### Done
- Investigated the production brand regression after enabling login agreement content and confirmed `site_name` / `site_logo` were overwritten by empty values.
- Located the root cause in `admin/settings` partial updates: several OEM branding fields were plain request fields in the backend handler, so omitted JSON fields were treated as zero values and persisted as resets.
- Changed the affected branding request fields to pointer semantics in the admin settings handler and now preserve previous values when the client omits them.
- Confirmed the production legal-document data and login-agreement toggle were updated successfully after the hotfix path was identified.

### Failures
- The first production legal-content update was applied through a partial admin settings API payload before the backend fix existed, which reset branding fields on the live site.

### Next
- Deploy the partial-update hotfix so future targeted settings writes do not clobber branding and other optional OEM values.

### Validation
- `go build ./cmd/server` passed from `backend/`.
- `go test ./internal/handler/admin -run 'Test.*Setting'` passed from `backend/`.

## 2026-05-18 Persistent Login Agreement Entry
### Done
- Updated the shared login agreement prompt so modal-mode legal entry remains visible below the login/register submit buttons even after the user has already accepted the current revision.
- Kept the existing gate behavior intact: users still cannot submit login/register while a required agreement revision is unaccepted.
- Added focused component coverage for:
  - accepted modal mode still showing a persistent “查看条款” entry
  - unaccepted modal mode still showing the stronger gate copy and agreement CTA

### Failures
- None.

### Next
- Commit, push, deploy, and verify production `/login` and `/register` show the persistent agreement entry under the primary action button.

### Validation
- `pnpm --dir frontend exec vitest run src/components/auth/__tests__/LoginAgreementPrompt.spec.ts src/utils/__tests__/loginAgreementTemplates.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed; only existing Vite chunk/import warnings were reported.

## 2026-05-21 Local Creem/Waffo Payment Integration
### Done
- Added `creem` and `waffo` provider support to the payment subsystem and factory wiring.
- Implemented local provider logic without adding new SDK dependencies:
  - `Creem` checkout creation / order query / webhook signature verification
  - `Waffo` signed order creation / query / webhook verification / signed webhook acknowledgement
- Extended payment provider request shape with:
  - `customer_email`
  - `customer_name`
  - provider-specific product id support for `Creem`
- Extended admin payment configuration to expose:
  - `creem` / `waffo` provider options
  - recharge product `creem_product_id`
  - subscription plan `creem_product_id`
  - Waffo callback URL fields in the provider config form
- Added schema + migration support for `subscription_plans.creem_product_id`.
- Wired user checkout payloads so recharge orders can send the selected local `product_id` to the backend.
- Added backend order-time resolution for Creem product ids:
  - subscription orders use the selected plan’s `creem_product_id`
  - balance recharge orders use the selected recharge product’s `creem_product_id`
  - invalid or missing Creem product bindings now fail before order creation.
- Extended frontend payment method normalization so `creem` and `waffo` appear as first-class hosted payment methods.
- Added regression coverage for:
  - Creem provider constructor/create/query/webhook behavior
  - Waffo provider constructor/create/query/webhook behavior
  - Creem product id resolution in order creation
  - frontend payment payload building for `creem`

### Failures
- I briefly drifted into an unrelated account-test debugging path and rolled that work back before continuing. No unrelated code was kept from that detour.
- `gofmt` was accidentally invoked against frontend files during verification; it failed fast and did not modify the Vue/TypeScript sources.

### Next
- Start a local app instance and manually exercise the admin payment configuration flow for:
  - adding a `creem` provider instance
  - adding a `waffo` provider instance
  - assigning `creem_product_id` values to recharge products / plans
  - confirming the user payment page renders the new methods end-to-end
- After local manual verification, decide whether to extend the integration further (for example EPay parity or richer provider-specific UX) before any production rollout.

### Validation
- `cd backend && go test ./internal/payment/provider ./internal/service -run 'Test(NewCreemRequiresKeys|CreemCreatePaymentAndQueryOrder|CreemVerifyNotification|NewWaffoValidatesKeys|WaffoCreateQueryAndWebhook|ResolveProviderProductID.*)' -count=1` passed.
- `cd backend && go test ./... -count=1` passed.
- `pnpm --dir frontend exec vitest run src/components/payment/__tests__/paymentFlow.spec.ts src/views/user/__tests__/PaymentView.spec.ts` passed.
- `pnpm --dir frontend exec eslint src/components/payment/paymentFlow.ts src/components/payment/providerConfig.ts src/components/payment/__tests__/paymentFlow.spec.ts src/views/user/PaymentView.vue src/views/admin/orders/RechargeProductsManager.vue src/views/admin/orders/PlanEditDialog.vue src/types/payment.ts --ext .ts,.vue` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed; existing Vite chunk-size warnings remain informational only.
- `git diff --check` passed.
- Manual local API verification passed:
  - started the workspace backend from source on `http://127.0.0.1:18082`
  - created a local `creem-local-test` provider instance through `POST /api/v1/admin/payment/providers`
  - created a local `waffo-local-test` provider instance through `POST /api/v1/admin/payment/providers`
  - updated local payment config so `checkout-info` exposed `["creem", "waffo"]`
  - ran a balance order against a local Creem mock backend and received `pay_url = https://checkout.creem.local/session/chk_local_123`

## 2026-05-22 Creem/Waffo Follow-up Fixes
### Done
- Bound Creem recharge orders to the server-side recharge catalog so:
  - the selected `product_id` determines the provider product id
  - the selected `product_id` also determines the local recharge amount used for order creation
  - local order amounts can no longer drift from the upstream Creem product being purchased
- Added dedicated `resolveCreemRechargeProduct` coverage for:
  - configured recharge product binding
  - non-Creem bypass behavior
- Fixed Waffo amount formatting so zero-decimal currencies (for example `JPY/KRW/VND/IDR`) are no longer forced into `%.2f`.
- Extended webhook order-id extraction to support:
  - Creem `object.request_id`
  - Waffo `result.merchantOrderId`
  This allows the existing pinned-order/provider-instance resolution path to work for multi-instance Creem/Waffo setups.
- Added service/handler regression coverage for:
  - Creem/Waffo webhook order-id extraction
  - Creem multi-instance webhook provider resolution with a pinned order
  - Waffo zero-decimal amount formatting
- Tightened frontend hosted-method availability so `Creem` is disabled unless:
  - the selected recharge product has `creem_product_id`
  - or the selected subscription plan has `creem_product_id`
- Updated recharge-page tests to lock the Creem selection-support logic in a pure helper.

### Failures
- I twice invoked `gofmt` against frontend files by mistake while batching verification commands; both invocations failed immediately and did not modify the Vue/TypeScript sources.

### Next
- Decide whether to add a local Waffo mock checkout/inquiry harness similar to the Creem mock before any production rollout.
- If rollout is approved later, commit the current payment integration as a single Lore commit and deploy only after one more browser-level user-flow check against a dev build that serves the frontend routes locally.

## 2026-05-27 Waffo Service-Level End-to-End Verification Follow-up
### Done
- Added a new service-level unit test scaffold at:
  - `backend/internal/service/payment_order_waffo_test.go`
- The new test exercises the real `PaymentService.CreateOrder` path for `waffo` with:
  - a temporary provider instance stored through `PaymentConfigService`
  - a local `httptest.Server` simulating Waffo `/order/create`
  - assertion that the created payment order returns the hosted `pay_url`
  - assertion that the upstream request body includes the generated merchant order id and customer email
- Re-ran the existing frontend payment regression coverage after the recent payment integration work:
  - `src/components/payment/__tests__/paymentFlow.spec.ts`
  - `src/views/user/__tests__/PaymentView.spec.ts`
- Re-ran the new Waffo service-level test with working proxy settings and fixed the test fixture:
  - created the backing user row in the Ent sqlite test database
  - aligned the mock-user repository with that real user ID
  - fixed the provider-instance creation call to reuse the existing `err` binding
- Verified the new service-level Waffo path now passes end-to-end.

### Failures
- The first two reruns were blocked by Go module proxy timeouts until proxy environment variables were provided.
- After network access was restored, the new Waffo service test exposed two real fixture issues:
  - duplicate `err :=` declaration in the test
  - missing persisted user row for the payment-order foreign key path
  Both issues were fixed in the test fixture.

### Next
- Decide whether to also add a manual local API verification path for Waffo comparable to the earlier Creem mock flow.
- If not, the current local verification bar is now:
  - provider-level Waffo tests
  - service-level Waffo order-creation path
  - existing frontend payment flow tests

### Validation
- `pnpm --dir frontend exec vitest run src/components/payment/__tests__/paymentFlow.spec.ts src/views/user/__tests__/PaymentView.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/components/payment/paymentFlow.ts src/components/payment/providerConfig.ts src/components/payment/__tests__/paymentFlow.spec.ts src/views/user/PaymentView.vue src/views/user/__tests__/PaymentView.spec.ts src/views/admin/orders/RechargeProductsManager.vue src/views/admin/orders/PlanEditDialog.vue src/types/payment.ts --ext .ts,.vue` passed.
- `cd backend && env https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890 all_proxy=socks5://127.0.0.1:7890 GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org go test -tags=unit ./internal/service -run 'TestCreateOrderWithWaffoProviderInstance' -count=1` passed.
- `cd backend && env https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890 all_proxy=socks5://127.0.0.1:7890 GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org go test -tags=unit ./internal/payment/provider -run 'TestWaffoCreateQueryAndWebhook|TestFormatWaffoAmountHonorsZeroDecimalCurrencies' -count=1` passed.
- `cd backend && env https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890 all_proxy=socks5://127.0.0.1:7890 GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org go test -tags=unit ./internal/service -run 'Test(CreateOrderWithWaffoProviderInstance|ResolveCreemRechargeProductUsesConfiguredCatalogValues|GetWebhookProvidersUsesPinnedCreemOrderProviderWhenMultipleInstancesExist)' -count=1` passed.

## 2026-05-22 Per-Account Login Agreement Consent Isolation
### Done
- Reworked login-agreement browser persistence from a single global revision record into a structured store with:
  - anonymous consent (used before identity is known)
  - per-email consent records
- Preserved backward compatibility with the old single-record storage shape so existing browsers do not break on upgrade.
- Updated login/register pages so agreement acceptance is now recalculated against the currently entered email address instead of reusing a browser-global acceptance flag.
- Added email watchers on both auth forms so switching from one account email to another in the same browser immediately reopens the agreement gate.
- Changed login/register submit payloads to send the agreement revision associated with the current email rather than a browser-global revision.
- Bound anonymous agreement acceptance to the authenticated user’s email on successful auth so OAuth and other email-less pre-auth flows do not leave behind a global acceptance token that leaks across accounts.
- Tightened rejection / server-enforced reaccept flows so both the anonymous record and the current email-specific record are cleared before reopening the agreement prompt.
- Added regression coverage for:
  - anonymous-vs-email acceptance isolation
  - binding anonymous acceptance to a concrete email
  - multiple email records coexisting independently
  - backward compatibility with legacy storage
  - LoginView re-showing the agreement prompt when the operator switches to a different email in the same browser

### Failures
- None so far during implementation.

### Next
- Re-run focused frontend tests plus typecheck/lint/build.
- After local verification, decide whether to apply the same UX tightening to any additional OAuth pre-auth entry pages beyond the shared login/register screens.

### Validation
- `pnpm --dir frontend exec vitest run src/utils/__tests__/loginAgreementConsent.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/utils/loginAgreementConsent.ts src/utils/__tests__/loginAgreementConsent.spec.ts src/views/auth/LoginView.vue src/views/auth/RegisterView.vue src/views/auth/__tests__/LoginView.turnstile.spec.ts src/stores/auth.ts --ext .ts,.vue` passed.
- `git diff --check` passed.

## 2026-05-25 Restore Login Turnstile
### Done
- Restored the Cloudflare Turnstile widget on the password login page.
- Reintroduced the login-page public settings wiring for:
  - `turnstile_enabled`
  - `turnstile_site_key`
- Re-enabled login-form validation so password login now requires a verified Turnstile token when the feature is enabled.
- Added the Turnstile token back into the login request payload.
- Restored backend `/api/v1/auth/login` Turnstile verification before credential authentication.
- Extended the auth test helper to allow injecting a Turnstile verifier.
- Added handler-level regression coverage that `/auth/login`:
  - rejects when Turnstile is enabled but no token is provided
  - succeeds and forwards the token when verification passes
- Updated the existing login-page frontend test to assert:
  - the Turnstile widget renders on login
  - the verified token is submitted to the login API

### Failures
- One LoginView test initially failed because the Turnstile stub did not emit a verify event; switched the stub to a clickable button emitter.
- One RegisterView compile issue surfaced from a stale `publicSettingsLoaded` identifier and was corrected earlier in the consent-isolation pass.

### Next
- If these validations stay green, deploy the Turnstile restoration to production and confirm the live `/login` page visibly renders the Cloudflare challenge again.

## 2026-05-25 Allow AdSense Bootstrap Script In CSP
### Done
- Added the Google AdSense bootstrap domain `https://pagead2.googlesyndication.com` to the default CSP `script-src`.
- Extended the CSP enhancement middleware so older configured policies are automatically upgraded to include the same AdSense domain.
- Added middleware regression coverage to ensure the AdSense domain is injected exactly once.

### Failures
- None during implementation.

### Next
- If backend validation stays green, deploy the CSP adjustment and confirm the browser console no longer reports the AdSense script being blocked by CSP.

## 2026-05-25 Allow AdSense Frame Domains In CSP
### Done
- Extended the default CSP `frame-src` to allow the Google ad iframe domains currently used after the AdSense bootstrap script loads:
  - `https://googleads.g.doubleclick.net`
  - `https://tpc.googlesyndication.com`
  - `https://ep2.adtrafficquality.google`
  - `https://www.google.com`
- Updated the CSP enhancement middleware so older configured policies are also upgraded with the same iframe allowlist.
- Added middleware regression coverage for the new AdSense frame/embed domains.

### Failures
- None during implementation.

### Next
- If backend validation stays green, redeploy and confirm the login-page browser console no longer reports AdSense iframe CSP violations.

### Validation
- `cd backend && go test ./internal/server/middleware -run TestEnhanceCSPPolicy -count=1` passed.
- `cd backend && go test ./... -count=1` passed.
- Deployed commit `8dea082a` to the production host and rebuilt the embedded frontend/backend bundle.
- The production runtime was still serving a stale custom `security.csp.policy` from `config.yaml`, so the live header continued to omit the new iframe domains even though the source and binary were updated.
- Patched the production `config.yaml` CSP policy to include:
  - `https://googleads.g.doubleclick.net`
  - `https://tpc.googlesyndication.com`
  - `https://ep2.adtrafficquality.google`
  - `https://www.google.com`
- Restarted the live service and verified both source-of-truth headers now include the full AdSense allowlist:
  - `http://127.0.0.1:8081/login`
  - `https://cloudbase.eu.org/login`
- Playwright verification against `https://cloudbase.eu.org/login?verify=adsense-csp-fix` no longer reports AdSense CSP violations. Remaining console noise comes from Cloudflare Turnstile, not AdSense.

## 2026-05-22 AdSense Verification Script
### Done
- Added the provided Google AdSense verification script to the shared SPA entry HTML head.
- This places the script into the generated app shell for all frontend routes that are served through `frontend/index.html`.

### Failures
- None during implementation.

### Next
- Rebuild the frontend bundle and verify the generated embedded `backend/internal/web/dist/index.html` contains the AdSense script before any production rollout.

## 2026-05-24 Homepage Supported-Model Card Trim
### Done
- Removed the public homepage “已支持的 AI 模型” cards for:
  - Gemini
  - Antigravity
- Kept the rest of the homepage provider strip unchanged so Claude / GPT / 更多 still render in the same layout.

### Failures
- None during implementation.

### Next
- Rebuild the frontend bundle and, if green, deploy this narrow homepage-only change to production without bundling unrelated payment work.

### Validation
- `pnpm --dir frontend run build` passed.
- Pushed deploy commit `9f251d5e` to `aias00/main`.
- Pulled commit `9f251d5e` on the production GCE host, rebuilt the embedded frontend/backend bundle, and replaced the live binary.
- Production health checks passed:
  - `http://127.0.0.1:8081/health`
  - `https://cloudbase.eu.org/health`

## 2026-05-25 Reduce Turnstile Console Noise
### Done
- Switched the shared Turnstile widget loader to Cloudflare's explicit render script path instead of the legacy global `onload=onTurnstileLoad` callback pattern.
- Replaced the global load callback with a module-scoped singleton script promise to avoid callback clobbering across SPA route transitions.
- Disabled Turnstile automatic retry and automatic refresh-on-expire/timeout so unsupported PAT or unattended challenge flows do not keep spamming the console in the background.
- Mapped Turnstile timeout events into the existing expire handling path so the login form still clears stale tokens without silent failures.

### Failures
- None during implementation.

### Next
- Deploy the frontend bundle update and verify the live login page only shows residual Turnstile-origin noise, not repeated avoidable retries caused by our component lifecycle.

### Validation
- `pnpm --dir frontend exec vitest run src/views/auth/__tests__/LoginView.turnstile.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/components/TurnstileWidget.vue src/views/auth/LoginView.vue --ext .vue` passed.
- `pnpm --dir frontend run build` passed.
- Deployed frontend commit `eec50ab4` and then follow-up commit `3d425f34` to production after narrowing the Turnstile callbacks to return truthy values.
- Production now serves the updated login bundle from commit `3d425f34` and remains healthy on:
  - `http://127.0.0.1:8081/health`
  - `https://cloudbase.eu.org/health`
- Browser verification against `https://cloudbase.eu.org/login?verify=turnstile-noise-fix-final-2` shows the avoidable app-side retry churn and extra callback logging have been reduced.
- Remaining console noise still originates inside Cloudflare's own Turnstile script under automated/PAT challenge flows (for example the upstream `postMessage`/PAT/`NaN` diagnostics), not from our app code.

## 2026-05-25 Clarify Dashboard Token Breakdown
### Done
- Updated the user dashboard token cards so the large aggregate total continues to include cache tokens, but the supporting copy now also shows:
  - cache write tokens
  - cache read tokens
- Kept the existing aggregate number and split the small print into two lines to avoid cramped wrapping.
- Added missing top-level dashboard i18n keys for the cache labels in both Chinese and English.

### Failures
- None during implementation.

### Next
- Deploy the dashboard copy change so the production token cards match the actual backend aggregation formula.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/components/user/dashboard/UserDashboardStats.vue src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .vue,.ts` passed.

## 2026-05-25 Remove User-Side Channel Navigation Entries
### Done
- Removed both ordinary-user sidebar entries related to channels:
  - available channels
  - channel status
- Kept the underlying routes intact so direct links and future re-exposure remain possible without backend changes.
- Left the admin-side channel management and monitor navigation unchanged.

### Failures
- None during implementation.

### Next
- Rebuild the frontend bundle and deploy this navigation-only simplification to production.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/components/layout/AppSidebar.vue --ext .vue` passed.

## 2026-05-26 Remove Duplicate Subscription Entry For Ordinary Users
### Done
- Kept the ordinary-user purchase entry (`充值/订阅`) in the sidebar.
- Removed the separate ordinary-user `我的订阅` sidebar entry so users no longer see two subscription-related first-level menu items.
- Left the `/subscriptions` route intact for direct links and internal jump actions.
- Left the admin-side personal navigation unchanged.

### Failures
- None during implementation.

### Next
- Rebuild the frontend bundle and deploy this user-navigation-only simplification to production.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/components/layout/AppSidebar.vue --ext .vue` passed.

### Follow-up
- The first pass only affected the pure ordinary-user sidebar.
- The admin "我的账户" section still reused the same self-navigation builder, so administrators checking their personal menu could still see both subscription-related entries.
- Narrowed the shared self-navigation builder itself so both ordinary users and the admin personal section now expose only the purchase entry while keeping `/subscriptions` reachable by direct link.

## 2026-05-26 Restore My Subscriptions Navigation Entry
### Done
- Restored the shared `我的订阅` sidebar entry in the self-navigation builder.
- This re-enables the menu item for:
  - ordinary users
  - the admin personal "我的账户" section
- Left the rest of the subscription/payment navigation unchanged.

### Failures
- None during implementation.

### Next
- Rebuild the frontend bundle and redeploy this navigation reversal to production.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/components/layout/AppSidebar.vue --ext .vue` passed.

## 2026-05-26 Defer Login Agreement Modal Until Submit
### Done
- Removed the login-page behavior that reopened the agreement modal immediately when the email field changed to another account.
- Kept the agreement entry point permanently visible on the login page.
- Changed the login flow so:
  - the user can switch accounts and fill credentials without interruption
  - the backend remains the source of truth for whether the current account must re-accept the latest agreement revision
  - a `LOGIN_AGREEMENT_REQUIRED` response now opens the agreement modal and records a pending retry
  - accepting the agreement automatically replays the login request
- Updated the non-accepted prompt copy so it no longer incorrectly claims that credential inputs are disabled before submit.
- Added regression coverage for:
  - switching emails no longer auto-opens the modal
  - agreement-required login errors open the modal and auto-retry after acceptance

### Failures
- The first pass left an unused `agreementGateActive` computed in `LoginView`; removed it after `vue-tsc` surfaced the dead state.

### Next
- Deploy the login-page agreement interaction change to production and verify the live flow against a login that requires agreement confirmation.

### Validation
- `pnpm --dir frontend exec vitest run src/views/auth/__tests__/LoginView.turnstile.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/auth/LoginView.vue src/components/auth/LoginAgreementPrompt.vue src/views/auth/__tests__/LoginView.turnstile.spec.ts --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.

### Follow-up
- Adjusted the login/register OAuth divider copy so the line below the email/password form no longer says “or continue with email”. It now reads as an alternative-method separator, which matches the actual layout.

## 2026-05-26 Treat Login Agreement Confirmation As Per-Attempt, Not Per-Account
### Done
- Replaced the old per-account/per-browser agreement persistence helper with a session-scoped “current auth attempt” helper.
- Login and register pages now clear any stale agreement attempt state on mount, so a fresh visit always starts unchecked.
- OAuth start buttons now forward the current in-memory agreement revision through the start URL instead of reading a long-lived remembered acceptance.
- LinuxDo / OIDC / WeChat OAuth start handlers now preserve the current agreement revision in short-lived OAuth cookies so direct callback logins keep the same “this attempt was confirmed” context.
- Backend agreement enforcement no longer treats `user.login_agreement_accepted_revision` as a reusable login bypass and no longer persists acceptance back to the user record during login/registration flows.
- Auth success now clears the current attempt agreement state instead of binding it to the logged-in account.

### Failures
- The first pass left an unused import in `EmailOAuthButtons.vue`; `vue-tsc` caught it and it was removed before final verification.

### Next
- Deploy the full per-attempt agreement confirmation behavior to production and validate at least one live login flow that hits `LOGIN_AGREEMENT_REQUIRED`.

### Validation
- `cd backend && go test ./internal/handler -run 'Test(LoginRequiresCurrentAgreementWhenEnabled|RegisterRequiresCurrentAgreementWhenEnabled|ExchangePendingOAuthCompletionRequiresCurrentAgreementForTokenIssue|EmailOAuthCallbackWithExistingUserRequiresAgreement)' -count=1` passed.
- `cd backend && go test ./internal/handler -count=1` passed.
- `pnpm --dir frontend exec vitest run src/utils/__tests__/loginAgreementConsent.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/auth/LoginView.vue src/views/auth/RegisterView.vue src/components/auth/LoginAgreementPrompt.vue src/components/auth/EmailOAuthButtons.vue src/components/auth/LinuxDoOAuthSection.vue src/components/auth/OidcOAuthSection.vue src/components/auth/WechatOAuthSection.vue src/utils/loginAgreementConsent.ts src/utils/__tests__/loginAgreementConsent.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.

## 2026-05-26 Require Turnstile Before Third-Party Login Starts
### Done
- Added a shared handler helper so OAuth start endpoints now verify `turnstile_token` before redirecting to upstream providers.
- Covered these start endpoints:
  - GitHub
  - Google
  - LinuxDo
  - OIDC
  - WeChat
- Updated all login/register-side OAuth buttons to forward the active Turnstile token in the start URL query.
- Updated auth-page action gating so third-party login buttons stay disabled until Turnstile completes, matching the password login requirement.

### Failures
- The first pass tried to add LinuxDo/OIDC/WeChat backend tests through helpers that do not expose Turnstile injection cleanly. Those tests were removed instead of forcing awkward fixture surgery.
- One Google OAuth start test initially referenced non-existent settings keys for authorize/token/userinfo URLs; those redundant overrides were removed.

### Next
- Deploy the OAuth-start Turnstile enforcement to production and verify one live third-party login button stays disabled until Turnstile succeeds.

### Validation
- `cd backend && go test ./internal/handler -run 'Test(GoogleOAuthStartRequiresTurnstileWhenEnabled|LoginRequiresCurrentAgreementWhenEnabled|RegisterRequiresCurrentAgreementWhenEnabled|ExchangePendingOAuthCompletionRequiresCurrentAgreementForTokenIssue|EmailOAuthCallbackWithExistingUserRequiresAgreement)' -count=1` passed.
- `pnpm --dir frontend exec vitest run src/components/auth/__tests__/EmailOAuthButtons.spec.ts src/utils/__tests__/loginAgreementConsent.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/auth/LoginView.vue src/views/auth/RegisterView.vue src/components/auth/EmailOAuthButtons.vue src/components/auth/LinuxDoOAuthSection.vue src/components/auth/OidcOAuthSection.vue src/components/auth/WechatOAuthSection.vue src/stores/auth.ts src/utils/loginAgreementConsent.ts src/utils/__tests__/loginAgreementConsent.spec.ts src/components/auth/__tests__/EmailOAuthButtons.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts --ext .vue,.ts` passed.

## 2026-05-26 Shorten Generated Redeem Codes To 8-Char Uppercase Alphanumerics
### Done
- Changed generated redeem codes from long hex strings to `8`-character uppercase alphanumeric codes.
- Added a shared uniqueness helper with bounded retries so generated codes are checked against the repository before being used.
- Updated both user-facing redeem generation and admin-side adjustment code generation to use the same short-code helper.
- Kept manual/fixed redeem codes untouched.
- Made redeem lookup backward-compatible:
  - first try the raw input
  - then try uppercase canonical form
  This preserves existing lowercase historical codes while allowing new uppercase codes to be entered in lowercase by users.
- Updated the user redeem hint copy so it no longer claims codes are case-sensitive.

### Failures
- None during implementation.

### Next
- Deploy the redeem-code format change to production and verify newly generated codes use the short format.

### Validation
- `cd backend && go test ./internal/service -run 'TestGenerateRedeemCodeFormat|TestRedeemServiceFindRedeemCodeByInputFallsBackToUppercase' -count=1` passed.
- `cd backend && go test ./... -count=1` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .ts` passed.

## 2026-05-26 Simplify Home Provider Section
### Done
- Removed the three pill-style home-page capability tags below the hero section.
- Removed the helper subtitle under “已支持的 AI 模型”.
- Kept the rest of the home-page structure, CTA, feature cards, and provider cards intact.

### Failures
- None during implementation.

### Next
- Deploy the home-page content trim to production and verify the public page no longer shows the removed capability tags or provider subtitle.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/HomeView.vue --ext .vue` passed.
- `pnpm --dir frontend run build` passed.

## 2026-05-26 Affiliate Center / Affiliate Management Design
### Done
- Audited the current affiliate/invitation implementation surface:
  - user page `/affiliate`
  - admin record pages under `/admin/affiliates/*`
  - invitation-code and affiliate rules embedded in `SettingsView`
  - backend `AffiliateService` and existing admin/user APIs
- Confirmed this round should not change rebate calculation rules; scope is module boundaries, entry points, and information architecture only.
- Completed the approved A2 design:
  - user-side “邀请中心”
  - admin-side “邀请管理” module
  - reuse existing invite binding / rebate accrual / freeze / transfer logic
  - add focused user/admin APIs where current surfaces are missing
- Wrote the design spec to:
  - `docs/superpowers/specs/2026-05-26-affiliate-center-design.md`

### Failures
- Did not run the brainstorming skill's requested subagent spec-review loop because this session's higher-priority tool policy only permits spawning child agents when the user explicitly requests delegation.

### Next
- Ask the user to review `docs/superpowers/specs/2026-05-26-affiliate-center-design.md`.
- After approval, convert the design into an implementation plan and then execute milestone-by-milestone.

## 2026-05-26 Affiliate Center / Affiliate Management Implementation
### Done
- Added backend affiliate module APIs:
  - user rebate records
  - user transfer records
  - admin affiliate overview
  - admin affiliate rules read/update
- Added backend handler coverage for:
  - `GET /api/v1/user/aff/rebates`
  - `GET /api/v1/user/aff/transfers`
  - `GET /api/v1/admin/affiliates/overview`
  - `GET /api/v1/admin/affiliates/rules`
  - `PUT /api/v1/admin/affiliates/rules`
- Extended the affiliate repository/service layer with:
  - current-user rebate/transfer list queries
  - admin overview aggregation query
  - dedicated affiliate rules read/update service methods
- Expanded admin affiliate routing/navigation into a standalone module:
  - `/admin/affiliates/overview`
  - `/admin/affiliates/rules`
  - `/admin/affiliates/codes`
  - existing invites/rebates/transfers pages kept intact under the same namespace
- Added new admin pages for:
  - module overview
  - rules configuration
  - invite-code management
- Reworked `/affiliate` into a fuller invite center:
  - retained overview/share/transfer actions
  - added rebate-records section
  - added transfer-records section
- Updated related API clients, shared frontend types, routing metadata, sidebar labels, and i18n copy.

### Failures
- The first focused handler run missed the `unit` build tag and reported “no tests to run”; reran with `-tags=unit`.
- Initial handler tests assumed paginated responses returned raw arrays instead of the project’s `{ data: { items, total, ... } }` envelope; fixed the test contracts.
- `AdminUpdateRules` tests initially panicked because the minimal settings stub used `SettingService.GetAllSettings()` without a config default; fixed by providing a minimal config object in the test.

### Next
- Optional follow-up: move the legacy affiliate settings/custom-user UI out of `SettingsView` and replace it with a module jump/notice once the new module is accepted as the primary entry.
- Optional follow-up: add focused Vitest coverage for the new admin affiliate pages and the expanded user invite-center sections.

### Validation
- `cd backend && go test -tags=unit ./internal/handler ./internal/handler/admin -run 'Test(UserHandlerAffiliate|AffiliateHandlerOverviewAndRulesEndpoints)' -count=1` passed.
- `cd backend && go test ./... -count=1` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/admin/affiliates/AdminAffiliateOverviewView.vue src/views/admin/affiliates/AdminAffiliateRulesView.vue src/views/admin/affiliates/AdminAffiliateCodesView.vue src/views/admin/affiliates/AdminAffiliateRecordsTable.vue src/views/user/AffiliateView.vue src/api/admin/affiliates.ts src/api/user.ts src/router/index.ts src/components/layout/AppSidebar.vue src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.

## 2026-05-27 SettingsView Affiliate Entry Shrink
### Done
- Replaced the old `SettingsView` invitation-code toggle area with a lightweight status summary plus a jump button into `/admin/affiliates/rules`.
- Replaced the old `SettingsView` affiliate card that previously contained:
  - rebate rule form fields
  - custom-user invite code table
  - add/edit/reset/batch dialogs
  with a transition-state summary card plus direct links into:
  - `/admin/affiliates/overview`
  - `/admin/affiliates/rules`
  - `/admin/affiliates/codes`
  - `/admin/affiliates/rebates`
- Removed the now-dead affiliate management state, modal, batch action, and confirm-dialog logic from `SettingsView`.
- Kept the underlying form fields in `SettingsView` state so saving unrelated settings continues to round-trip current affiliate values instead of forcing them to defaults.

### Failures
- Initial large `apply_patch` deletions were too broad and failed to match cleanly inside `SettingsView`; switched to smaller targeted patches.
- The first cleanup pass left an unused `watch` import, which `vue-tsc`/ESLint surfaced and was removed immediately.

### Next
- Optional follow-up: stop including affiliate fields in the generic settings save payload only after the backend `UpdateSettingsRequest` safely supports omitted `invitation_code_enabled` and related values without coercing them to false/defaults.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/admin/SettingsView.vue src/views/admin/affiliates/AdminAffiliateOverviewView.vue src/views/admin/affiliates/AdminAffiliateRulesView.vue src/views/admin/affiliates/AdminAffiliateCodesView.vue src/views/admin/affiliates/AdminAffiliateRecordsTable.vue src/views/user/AffiliateView.vue src/api/admin/affiliates.ts src/api/user.ts src/router/index.ts src/components/layout/AppSidebar.vue src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .vue,.ts` passed.

## 2026-05-27 DragonCode-Style Homepage Redesign
### Done
- Replaced the old split hero/feature-grid homepage structure with a new landing-page skeleton modeled on the pacing of `dragoncode.codes`:
  - minimal top navigation
  - centered hero
  - model matrix directly below the hero
  - pricing section directly below the model matrix
  - experience section
  - why-choose-us section
  - grouped footer
- Preserved `homeContent` override behavior exactly as before, including iframe mode and raw HTML mode.
- Added a new public payment catalog endpoint:
  - `GET /api/v1/payment/public/catalog`
  - exposes public recharge products and for-sale plans for homepage rendering without requiring login
- Added a frontend homepage catalog helper:
  - derives visible model families from current plan platforms / model scopes
  - groups pricing into recharge products and subscription plans
- Reworked `HomeView.vue` to consume the new public catalog endpoint and render model/pricing sections dynamically from backend/payment configuration instead of a separately maintained static marketing table.
- Added regression coverage for:
  - homepage catalog helper mapping
  - public payment catalog handler
  - `HomeView` rendering the new hero and dynamic pricing sections

### Failures
- The first typecheck pass failed because `Icon` does not support a `signal` name; changed the experience-card icon to an existing supported icon.
- Local browser verification initially still showed the old homepage because the running local `go run -tags embed ./cmd/server` process had started before the homepage rewrite. Restarting the local embedded server fixed the mismatch.

### Next
- Decide whether the homepage should keep showing empty-state model/pricing blocks when the public catalog is empty, or hide those sections entirely until products/plans are configured.
- Optional follow-up: further tune the visual fidelity (spacing, copy, decorative background treatment) after a side-by-side screenshot review against the reference site.

### Validation
- `cd backend && env https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890 all_proxy=socks5://127.0.0.1:7890 GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org go test -tags=unit ./internal/handler -run 'TestPaymentHandlerGetPublicCatalog' -count=1` passed.
- `pnpm --dir frontend exec vitest run src/views/home/__tests__/homeCatalog.spec.ts src/views/__tests__/HomeView.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/HomeView.vue src/views/home/homeCatalog.ts src/views/home/__tests__/homeCatalog.spec.ts src/views/__tests__/HomeView.spec.ts src/api/payment.ts src/types/payment.ts src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.
- Local visual check passed at `http://127.0.0.1:18082/` after restarting the embedded local server.

## 2026-05-27 DragonCode-Style Homepage Redesign — Visual Polish Pass 2
### Done
- Tightened the post-hero sections so the page reads more like a product landing page and less like an operational dashboard:
  - narrower top nav width
  - more whitespace in the hero
  - lighter model matrix / pricing section copy
  - removed pricing item counters
  - replaced the heavy dark "why choose us" block with a pale blue band and lighter cards
  - simplified the footer into a sparser brand + 3-column layout
- Kept the public catalog-driven model/pricing sections intact while making their empty states more marketing-friendly.
- Rebuilt the frontend bundle and restarted the local embedded backend so the rendered page reflects the latest `HomeView` implementation.

### Failures
- None in this pass.

### Next
- Decide whether the homepage should be committed as-is or receive a final content pass once public recharge products and plans exist in local/production data.
- Optional follow-up: if you want even closer similarity, the next pass should reduce copy density further instead of adding more sections.

### Validation
- `pnpm --dir frontend exec vitest run src/views/home/__tests__/homeCatalog.spec.ts src/views/__tests__/HomeView.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/HomeView.vue src/views/home/homeCatalog.ts src/views/home/__tests__/homeCatalog.spec.ts src/views/__tests__/HomeView.spec.ts src/api/payment.ts src/types/payment.ts src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.
- `git diff --check` passed.
- Local health check: `curl -sS http://127.0.0.1:18082/health` -> `{"env":"production","status":"ok"}`
- Local homepage rebuilt bundle confirmed in HTML:
  - `/assets/index-Dow1BNoA.js`
  - `/assets/index-ClB7GPr7.css`

## 2026-05-27 Homepage Public Catalog Demo Data
### Done
- Added a reproducible local seed file at [backend/dev/seed_homepage_public_catalog.sql](/Users/aias/Work/github/sub2api/backend/dev/seed_homepage_public_catalog.sql) to populate the homepage's public pricing/model catalog.
- Seeded local payment settings with:
  - `payment_enabled = true`
  - `BALANCE_RECHARGE_MULTIPLIER = 1.2`
  - three recharge products: `体验包` / `开发包` / `团队包`
- Seeded local group + plan data for public homepage rendering:
  - `Claude`
  - `GPT`
  - `Gemini`
  - four sellable plans across those families
- Verified that the local homepage now replaces the placeholder cards with real model names, real recharge products, and real subscription plans.

### Failures
- None in this pass.

### Next
- If you want the same non-placeholder effect on production, the next step is not more frontend work — it is populating the production payment catalog with real sellable products and plans.

### Validation
- `docker exec -i sub2api-local-postgres psql -U sub2api -d sub2api < backend/dev/seed_homepage_public_catalog.sql` applied successfully.
- `curl -sS http://127.0.0.1:18082/api/v1/payment/public/catalog | jq '.data.recharge_products, .data.plans | length'` returned `3` and `4`.
- `curl -sS http://127.0.0.1:18082/api/v1/payment/public/catalog | jq '.data.recharge_products[].name, .data.plans[].name'` returned the seeded product and plan names.
- Local visual verification passed at `http://127.0.0.1:18082/home` after refreshing the page against the seeded catalog.

## 2026-05-27 Homepage Content Reduction
### Done
- Hid Gemini-related content from the homepage display layer:
  - removed Gemini from the visible model matrix
  - filtered Gemini plans out of homepage-facing helpers
  - updated hero and model-matrix copy so it no longer promises Gemini on the homepage
- Removed the entire homepage pricing / plans section:
  - deleted the pricing anchor from homepage navigation
  - changed the secondary hero CTA from pricing to model browsing
  - removed pricing links from the footer
  - updated supporting copy so the remaining cards no longer claim public homepage pricing
- Made the homepage model matrix auto-optimize by card count:
  - one card -> single centered column
  - two cards -> centered two-column grid
  - three or more cards -> standard three-column layout
- Replaced model-name pills on the homepage with capability tags:
  - Claude -> `复杂推理 / 系统设计 / 代码审查`
  - GPT -> `代码生成 / 功能迭代 / Agent 调用`
  - the homepage no longer exposes concrete model identifiers like `Claude Opus 4.6` or `GPT-5.4`
- Removed the homepage model-matrix section entirely:
  - removed the section body
  - removed the "查看模型" secondary CTA
  - removed model-matrix navigation / footer links
  - stopped homepage-side public catalog loading
- Later restored the model-matrix section and the "查看模型" CTA:
  - model matrix is visible again
  - hero secondary CTA is visible again
  - model-matrix nav / footer links are restored
  - homepage-side public catalog loading is re-enabled
  - Gemini remains hidden from the homepage
  - pricing section remains removed from the homepage

### Failures
- After deleting the pricing section, `platformLabel` became unused and broke `vue-tsc`; removed the dead helper and reran verification.

### Next
- If you later want pricing back, restore it as a separate decision instead of letting homepage copy drift into talking about packages that are no longer shown.

### Validation
- `pnpm --dir frontend exec vitest run src/views/home/__tests__/homeCatalog.spec.ts src/views/__tests__/HomeView.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/HomeView.vue src/views/home/homeCatalog.ts src/views/home/__tests__/homeCatalog.spec.ts src/views/__tests__/HomeView.spec.ts src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.
- Local embedded server restarted successfully on `http://127.0.0.1:18082`.
- Visual verification passed at `http://127.0.0.1:18082/home`: pricing section absent, Gemini card absent, capability tags removed with the whole model-matrix section, and `查看模型` text absent.
- Final visual verification passed at `http://127.0.0.1:18082/home`: model matrix visible again, `查看模型` text restored, Gemini still absent, pricing section still absent.

## 2026-05-27 Model Plaza
### Done
- Added a backend-configurable model plaza data path using a new settings key:
  - `model_plaza_items`
- Extended public settings so the frontend can read configured model plaza cards without requiring auth.
- Extended admin settings save/load so model plaza items can be edited from the existing settings backend.
- Added a new public page:
  - `/models`
- Wired homepage discovery back in:
  - Hero secondary CTA now links to `/models`
  - `模型矩阵` navigation/footer entries point to `/models`
- Implemented a dark card-grid public plaza page inspired by the reference structure but driven by local config:
  - provider-aware card styling
  - capability tags
  - pricing text lines
  - copy model ID button
  - hidden-item filtering
- Seeded local demo data for the plaza with 4 Claude-family cards:
  - `claude-haiku-4-5-20251001`
  - `claude-opus-4-6`
  - `claude-opus-4-7`
  - `claude-sonnet-4-6`
- Promoted the admin-side model plaza editor from a subsection inside `通用设置` to its own dedicated `模型广场` tab in `SettingsView`.

### Failures
- `ModelsPlazaView` initially forgot to expose `t()` from `useI18n`, which broke the page test and typecheck; fixed by pulling `t` into the component setup.
- Adding `model_plaza_items` to `PublicSettings` surfaced a fallback-object type error in `appStore.fetchPublicSettings`; fixed by adding an empty array to the fallback shape.

### Next
- Decide whether the model plaza should stay under `SettingsView` or be moved into a dedicated admin page later if this list grows.
- Decide whether homepage should eventually drop the preview matrix and only keep `/models` as the canonical public catalog.

### Validation
- `cd backend && env https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890 all_proxy=socks5://127.0.0.1:7890 GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org go test -tags=unit ./internal/handler ./internal/handler/admin ./internal/service -run 'TestSettingHandler_GetPublicSettings_(ExposesModelPlazaItems|ExposesForceEmailOnThirdPartySignup|ExposesPasswordMinLength)|TestPaymentHandlerGetPublicCatalog' -count=1` passed.
- `pnpm --dir frontend exec vitest run src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/__tests__/HomeView.spec.ts src/views/home/__tests__/homeCatalog.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/public/ModelsPlazaView.vue src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/HomeView.vue src/views/__tests__/HomeView.spec.ts src/views/home/homeCatalog.ts src/views/home/__tests__/homeCatalog.spec.ts src/views/admin/SettingsView.vue src/i18n/locales/zh.ts src/i18n/locales/en.ts src/router/index.ts src/types/index.ts src/api/admin/settings.ts src/stores/app.ts --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.
- Local visual verification passed at:
  - `http://127.0.0.1:18082/models`
  - `http://127.0.0.1:18082/home`
- Admin visual verification passed at:
  - `http://127.0.0.1:18082/admin/settings`
  - dedicated `模型广场` tab is visible and switches to the model-plaza editor content

## 2026-05-27 模型广场左侧分组与搜索

### Done
- Added a left-side filter rail to `/models` with provider-based grouping:
  - `全部模型`
  - provider-derived groups such as `Claude` and `GPT`
- Added client-side search on the public model plaza page.
- Search now matches against:
  - title
  - provider
  - badge
  - description
  - pricing text
  - capability tags
  - model IDs
- Added a filtered empty state for "no matching cards" so search/group combinations do not fall back to the unconfigured-plaza state.
- Extended `ModelsPlazaView` tests to cover:
  - visible-card rendering
  - group filtering
  - search filtering
  - empty search result state

### Next
- Decide whether provider-based grouping is sufficient, or whether the admin editor should later expose an explicit custom group field for the plaza sidebar.
- If `/models` grows beyond a few dozen cards, consider adding sticky section anchors or secondary sort options.

## 2026-05-27 cloudbase logo and icon candidates

### Done
- Created three production-oriented `cloudbase` branding candidate pairs under:
  - `assets/branding/cloudbase/`
- Added these SVG deliverables:
  - `cloud-gate-logo.svg`
  - `cloud-gate-icon.svg`
  - `neural-portal-logo.svg`
  - `neural-portal-icon.svg`
  - `orbit-gateway-logo.svg`
  - `orbit-gateway-icon.svg`
- Added usage notes and a local preview page:
  - `assets/branding/cloudbase/README.md`
  - `assets/branding/cloudbase/preview.html`
- Kept the current live logo untouched; these are candidate assets, not a live brand swap.

### Validation
- `xmllint --noout assets/branding/cloudbase/*.svg` passed.
- `git diff --check` passed.
- Rendered the SVGs to PNG thumbnails with Quick Look to confirm they display correctly:
  - `/tmp/cloudbase-brand-renders/cloud-gate-logo.svg.png`
  - `/tmp/cloudbase-brand-renders/neural-portal-logo.svg.png`
  - `/tmp/cloudbase-brand-renders/orbit-gateway-logo.svg.png`

### Next
- Pick one direction as the canonical live logo.
- If one direction is selected, generate the final rollout set next:
  - favicon-sized PNGs
  - high-resolution app icon exports
  - replacement for `frontend/public/logo.png`

## 2026-06-02 用户可用分组入口

### Done
- Added a dedicated user-facing `可用分组 / Available Groups` page powered by the existing `/groups/available` API.
- Added a clear sidebar entry for regular users:
  - `/available-groups`
- Split the page into two user-facing sections:
  - public groups
  - exclusive / subscription groups
- Reused existing group rate data from `/groups/rates` so user-specific multipliers stay visible.
- Added search across group name, description, platform, and type.

### Validation
- `pnpm --dir frontend exec vitest run src/views/user/__tests__/AvailableGroupsView.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/views/user/AvailableGroupsView.vue src/views/user/__tests__/AvailableGroupsView.spec.ts src/components/layout/AppSidebar.vue src/router/index.ts src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.
- Local browser verification passed at `http://127.0.0.1:18082/available-groups`:
  - sidebar entry visible
  - page title renders
  - public group cards render from `/groups/available`

### Next
- Decide whether `可用分组` should eventually replace part of the current `可用渠道` mental model for regular users.
- If users need stronger purchase guidance later, connect each group card to the relevant subscription / purchase path.

## 2026-06-02 Merge upstream main into local main

### Done
- Fetched the latest `origin/main` from `Wei-Shaw/sub2api` and merged it into the local `main` branch.
- Used a local-favoring merge strategy for overlapping hunks, then repaired the merge fallout explicitly where upstream additions were required to keep the branch buildable.
- Reconciled payment/provider type declarations so local `creem` / `waffo` support and upstream `airwallex` support coexist.
- Restored missing frontend exports and typings introduced by the upstream merge:
  - account upstream model sync APIs
  - user platform quota API
  - airwallex-visible payment method typing
- Reconciled backend service/runtime merge gaps:
  - regenerated / repaired ent-generated files around `platform_quotas`
  - reconnected `UserPlatformQuotaUsageFlusher` into `wire_gen.go`
  - restored OpenAI gateway upstream error helpers
  - restored DingTalk settings write-through fields
  - removed stale ops retry handler endpoints that no longer matched current service APIs
- Confirmed the merged local branch is buildable again after the merge.

### Validation
- `pnpm --dir frontend install --frozen-lockfile` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend run build` passed.
- `cd backend && env https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890 all_proxy=socks5://127.0.0.1:7890 GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org go test ./cmd/server ./internal/payment/provider ./internal/repository ./internal/service ./internal/handler ./internal/handler/admin -count=1` passed.

### Next
- Decide whether to push the local upstream-merge result to `aias00/main`.
- Review the remaining untracked local-only assets (`assets/branding`, `.superpowers`, screenshots) separately from the code merge.

## 2026-06-02 Public docs refresh

### Done
- Restructured the docs landing page to feel more like a product guide instead of a loose article collection.
- Rewrote the Chinese and English docs homepages to better explain:
  - what cloudbase is
  - who it is for
  - the recommended onboarding path
  - the console concepts users should learn first
- Added a new `控制台能力 / Console` section in the docs sidebar.
- Added new docs pages for:
  - API keys
  - available groups
  - model plaza
  - Gemini CLI quickstart
- Kept the existing docs shell and docsify setup; only improved content architecture and coverage.

### Validation
- Verified all new markdown files exist in both zh/en trees.
- `git diff --check` passed.
- `pnpm --dir frontend run build` passed.
- Local docs preview verified after rebuilding embedded frontend assets:
  - `/docs`
  - `/docs#/console/available-groups`
- Confirmed the new sidebar sections and the new pages render in the local browser.

### Next
- Consider adding a dedicated `Gateway Guide` docs page if users still need a more explicit bridge from console-generated configuration to client setup.
- Consider polishing docs branding and shell styling later if you want the visual language to move closer to the reference site, not just the information architecture.

## 2026-06-03 Frontend i18n tree unification

### Done
- Eliminated the last frontend missing locale keys and Chinese leaks in `en.ts`.
- Migrated helper-based bilingual copy in:
  - `AdminPaymentPlansView.vue`
  - `RechargeProductsManager.vue`
  - `EmailTemplateEditor.vue`
- Migrated the `SettingsView.vue` login-agreement, payment recharge-product, and model-plaza sections from `localText(...)` helper copy into the locale tree.
- Replaced remaining runtime strings tied to those sections with locale-backed messages, including agreement validation errors and model-plaza defaults.

### Validation
- `pnpm --dir frontend exec vitest run src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts` passed.
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/components/auth/LoginAgreementPrompt.vue src/components/auth/WechatOAuthSection.vue src/views/admin/ops/components/OpsSystemLogTable.vue src/views/admin/SettingsView.vue src/views/admin/orders/AdminPaymentPlansView.vue src/views/admin/orders/RechargeProductsManager.vue src/views/admin/settings/EmailTemplateEditor.vue src/i18n/locales/zh.ts src/i18n/locales/en.ts src/components/account/UsageProgressBar.vue --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.

### Next
- If needed, continue beyond locale-tree completeness into deeper content governance:
  - authored legal template content in `frontend/src/utils/loginAgreementTemplates.ts`
  - hardcoded product labels in model-whitelist mapping helpers
  - non-user-facing comments and developer annotations

## 2026-06-03 Dashboard i18n hotfix

### Done
- Verified the production deployment reached commit `48204a53` but the dashboard still rendered raw `dashboard.platformBreakdown` and `dashboard.platformCount` keys.
- Added a narrow fallback in `UserDashboardStats.vue` so the platform-breakdown heading and platform-count summary still render localized copy even if those two runtime keys fail to resolve.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/components/user/dashboard/UserDashboardStats.vue --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.

### Next
- Redeploy this small dashboard patch to production.

## 2026-06-03 Production server runbook

### Done
- Added a dedicated production runbook at `deploy/PRODUCTION_SERVER_RUNBOOK.md`.
- Recorded the current GCE project / zone / instance, runtime paths, deploy flow, binary output rule, health checks, disk layout, and low-space cleanup procedure.
- Linked `deploy/README.md` to the production runbook so future server work defaults to the tracked repo document.

### Validation
- `git diff --check` passed.

### Next
- Use `deploy/PRODUCTION_SERVER_RUNBOOK.md` as the default reference for future production operations.

## 2026-06-03 Docs entry restoration

### Done
- Confirmed `/docs` still worked, but the public docs entry had been hidden by UI conditions rather than broken routing.
- Restored a stable docs entry in the logged-in header by removing the `docUrl`-required gate.
- Restored a stable docs entry on the public home page by:
  - keeping the header docs button visible even when `doc_url` is empty
  - adding an explicit `文档 / Docs` item back into the top nav
- Kept the existing fallback behavior so empty `doc_url` still resolves to the internal `/docs` route.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/components/layout/AppHeader.vue src/views/HomeView.vue --ext .vue,.ts` passed.
- `pnpm --dir frontend run build` passed.

### Next
- Deploy the docs-entry fix if the current local patch should go live immediately.

## 2026-06-04 Docs content review fixes

### Done
- Reviewed the current docs tree for navigation completeness, cross-page link correctness, and zh/en naming consistency.
- Fixed the broken relative links inside:
  - `quickstart/README.md`
  - `en/quickstart/README.md`
  - `console/api-keys.md`
  - `en/console/api-keys.md`
  - `console/models-plaza.md`
  - `en/console/models-plaza.md`
- Unified the English docs terminology around the console access page to `API Guide`.
- Expanded the zh/en docs homepages so their quickstart recommendation lists match the actual sidebar coverage.
- Kept the docs shell visual refresh in `DocsView.vue` as a presentation-only change; content fixes remain in markdown.

### Validation
- Relative-link scan confirms only the docsify directory-index entries in `_sidebar.md` still use directory targets such as `quickstart/`, which is expected.
- `pnpm --dir frontend run build` passed.
- `git diff --check` pending final pre-commit rerun.

### Next
- Commit the docs shell and docs-content fixes together, then deploy to production.

## 2026-06-04 Docsify nested-link hotfix

### Done
- Identified a second docs-specific issue after the first production rollout: nested docs pages were still relying on docsify-relative `./` and `../` links, which rendered as malformed hash targets in production.
- Replaced those links with explicit hash targets in:
  - `quickstart/README.md`
  - `en/quickstart/README.md`
  - `console/api-keys.md`
  - `en/console/api-keys.md`
  - `console/models-plaza.md`
  - `en/console/models-plaza.md`
  - `quickstart/gemini-cli.md`
  - `en/quickstart/gemini-cli.md`

### Validation
- Relative-link scan returned `[]`.
- `pnpm --dir frontend run build` passed.

### Next
- Commit and deploy the docsify nested-link hotfix so production matches the verified local state.

## 2026-06-04 common.login i18n key

Done:
- Added the missing `common.login` key in zh/en locale files for the login route title.

Validation:
- `pnpm --dir frontend exec eslint src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .ts` passed.
- A focused key-presence check confirmed `common.login` exists in both locale files.

Next:
- No follow-up needed for this key.

## 2026-06-04 Public UI audit polish

### Done
- Audited the production public pages across desktop and mobile:
  - `/home`
  - `/models`
  - `/docs#/quickstart/README`
  - `/login`
- Added defensive numeric formatting in the user dashboard stats component so incomplete platform usage data does not throw on `toFixed`.
- Reduced mobile header pressure on the public home page by hiding the secondary docs button below the `sm` breakpoint.
- Reworded the public home model-card empty state so it reads as a capability entry instead of an internal catalog-sync status.
- Tightened the models plaza mobile layout:
  - compacted the filter sidebar on narrow viewports
  - changed group filters into horizontal chips on mobile
  - improved long model-title wrapping
  - split model pricing into a clearer two-column grid on larger cards
- Restyled the auth shell and login button to better match the home page's white/slate/sky direction without changing the global admin button style.

### Validation
- `pnpm --dir frontend run typecheck` passed.
- `pnpm --dir frontend exec eslint src/components/user/dashboard/UserDashboardStats.vue src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts src/views/HomeView.vue src/views/public/ModelsPlazaView.vue src/components/layout/AuthLayout.vue src/views/auth/LoginView.vue src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .vue,.ts` passed.
- `pnpm --dir frontend exec vitest run src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/__tests__/HomeView.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts` passed.
- `pnpm --dir frontend run build` passed with existing Vite chunk warnings.
- `git diff --check` passed.
- Local preview is running at `http://127.0.0.1:18082/` with the dev proxy pointed at the production API.

### Next
- Review the local preview and decide whether to deploy this UI polish patch.

## 2026-06-09 Touch platform integration groundwork

### Done
- Added `docs/TOUCH_PLATFORM_INTEGRATION.md` to define the target Sub2API + Touch unified project shape:
  - Sub2API remains the platform source of truth.
  - Touch moves under `apps/touch` only after contracts and cutovers are proven.
  - Touch-derived backend domains stay isolated under `/api/v1/touch/*`.
  - Capability migration uses a fixed API/data/read/write/worker/rollback/deletion template.
- Added the first Touch route namespace mount:
  - `GET /api/v1/touch/capabilities`
  - Route implementation: `backend/internal/server/routes/touch.go`
  - Route registration: `backend/internal/server/router.go`
  - Regression test: `backend/internal/server/routes/touch_test.go`
- Updated `.gitignore` so the new Touch integration doc is tracked despite the existing `docs/*` ignore rule.

### Validation
- `go test ./internal/server/routes` passed from `backend/`.

### Next
- Implement the first real Sub2API Touch domain behind `/api/v1/touch/*`, starting with prompt/content read APIs or auth/session contract verification.

## 2026-06-09 Touch prompt case read API

### Done
- Added the first real Touch domain in Sub2API: prompt case catalog read APIs under `/api/v1/touch/prompts/*`.
- Added `touch_prompt_items` migration with fields covering the current Touch case catalog model:
  - prompt text and preview
  - category/tags/model tags/styles/scenes
  - source URL/project/type/label
  - single and multi-image URLs
  - import source/raw JSON/status/imported time
- Added repository/service/handler layers for:
  - `GET /api/v1/touch/prompts/cases`
  - `GET /api/v1/touch/prompts/cases/:id`
- Added service/repository upsert support for future import workers, including default normalization for category/source/status/arrays.
- Added an admin-protected HTTP write contract for future importers:
  - `POST /api/v1/touch/admin/prompts/cases`
  - registered only when `adminAuth` is explicitly provided
- Kept list ordering aligned with the current Touch behavior:
  - `imported_at DESC NULLS LAST`
  - `title ASC`
  - `id ASC` tie-breaker
- Updated Wire providers and regenerated `backend/cmd/server/wire_gen.go`.

### Validation
- `go test ./internal/service ./internal/server/routes ./internal/handler ./internal/repository` passed from `backend/`.
- `go generate ./cmd/server` completed and regenerated Wire output.
- `go test ./cmd/server ./internal/server ./internal/server/routes ./internal/service ./internal/handler ./internal/repository` passed from `backend/`.
- After adding upsert support, reran:
  - `go test ./internal/service ./internal/server/routes ./internal/handler ./internal/repository`
  - `go test ./cmd/server ./internal/server ./internal/server/routes ./internal/service ./internal/handler ./internal/repository`
- After adding the admin upsert route, reran the same package sets and both passed.

### Next
- Migrate Touch prompt import/write path into Sub2API so X imports write to `touch_prompt_items`.
- Add a one-time migration/import tool for existing Touch prompt rows once the database source mapping is finalized.

## 2026-06-09 Touch X import backend

### Done
- Added Sub2API-native X/Twitter prompt import service:
  - parses and normalizes `x.com` / `twitter.com` status URLs
  - reads X content from `X_AUTO_BASE_URL` / `X_ATUO_BASE_URL` when configured
  - falls back to the vendored local `tools/x-atuo` runtime when no HTTP x-auto endpoint is configured
  - falls back to Twitter oEmbed, FxTwitter, and page metadata
  - extracts prompt text, author label, model tags, and `pbs.twimg.com` media URLs
- Copied hot's `x_atuo` runtime into Sub2API under `tools/x-atuo/src/x_atuo` so production does not need to call another project checkout.
- Added `tools/x-atuo/pyproject.toml` with the minimal runtime dependencies for local x-auto execution.
- Added R2 image sync for imported prompt cases using existing AWS S3 SDK dependencies:
  - supports Touch env names: `R2_ACCESS_KEY`, `R2_SECRET_KEY`, `R2_BUCKET_NAME`, `R2_ENDPOINT`, `R2_DOMAIN`, `R2_ACCOUNT_ID`, `R2_UPLOAD_PATH`
  - also supports hot-style env names: `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`
  - replaces imported image URLs with public static URLs when upload succeeds
  - drops unsynced source images and returns warnings when R2 is unavailable, matching the current Touch import safety behavior
- Added admin import endpoint:
  - `POST /api/v1/touch/admin/prompts/import-twitter`
  - protected by Sub2API `adminAuth`
  - upserts into `touch_prompt_items` through `TouchPromptService`
- Added handler/service tests for X import and static image URL upsert behavior.

### Validation
- `go test ./internal/service -run 'TestTouch(TwitterImport|Prompt)' -count=1 -v` passed.
- `go generate ./cmd/server` completed and regenerated Wire output.
- `go test ./cmd/server ./internal/server ./internal/server/routes ./internal/service ./internal/handler ./internal/repository` passed from `backend/`.
- After vendoring local x-auto fallback, reran:
  - `go test ./internal/service -run 'TestTouch(TwitterImport|Prompt)' -count=1`
  - `go generate ./cmd/server`
  - `go test ./cmd/server ./internal/server ./internal/server/routes ./internal/service ./internal/handler ./internal/repository`

### Next
- Wire the Touch frontend “from link import” action to call Sub2API instead of the old Touch API route.
- Install/sync `tools/x-atuo` Python dependencies in production when relying on local x-auto fallback instead of `X_AUTO_BASE_URL`.

## 2026-06-09 Touch import cutover start

### Done
- Updated the separate Touch app's legacy import API routes to proxy X imports into Sub2API:
  - `SUB2API_BASE_URL`
  - `SUB2API_ADMIN_API_KEY`
- Touch calls Sub2API `POST /api/v1/touch/admin/prompts/import-twitter` with `x-api-key`.
- The proxy normalizes Sub2API snake_case prompt case fields back to Touch's camelCase shape so the current gallery UI can consume the response without UI changes.
- Legacy Touch local import fallback was removed; missing Sub2API config now fails fast.

### Validation
- In `/Users/aias/Work/github/touch`, targeted ESLint passed for the changed files.
- In `/Users/aias/Work/github/touch`, `pnpm exec tsc --noEmit --pretty false` passed.

### Next
- Deploy Touch with `SUB2API_BASE_URL` and `SUB2API_ADMIN_API_KEY` so production prompt reads/imports use Sub2API only.
- Continue moving Touch auth/billing calls to Sub2API.

## 2026-06-09 Touch admin authority check

### Done
- Added `POST /api/v1/touch/admin/auth/check`, protected by Sub2API `adminAuth`, for Touch backend-to-backend admin checks.
- Added `UserService.IsAdminByEmail`, mapping Touch's current logged-in email to the Sub2API user row and treating only active `admin` users as admins.
- Regenerated Wire output after adding `TouchAuthHandler`.

### Validation
- `go generate ./cmd/server` passed from `backend/`.
- `go test ./internal/server/routes ./internal/service ./internal/handler -run 'TestTouch|TestUserService' -count=1` passed.

### Next
- Keep replacing Touch local RBAC/admin page guards with Sub2API authority.
- Plan the next identity bridge step so Touch registration/login can be backed by Sub2API instead of Touch-local user tables.

## 2026-06-09 Touch user sync bridge

### Done
- Added `UserService.SyncTouchUser`, an idempotent Touch identity bridge:
  - existing Sub2API users are returned by normalized email
  - missing users are created as active regular users with `signup_source=touch`
  - `touch_user_id` is retained in notes for traceability
- Added `POST /api/v1/touch/admin/users/sync`, protected by Sub2API `adminAuth`.
- Added route coverage for admin check and user sync.

### Validation
- `go generate ./cmd/server` passed from `backend/`.
- `go test ./internal/server/routes ./internal/service ./internal/handler -run 'TestTouch|TestUserService|TestAuth' -count=1` passed.

### Next
- Move Touch credit/balance reads to Sub2API now that Touch users can be materialized in Sub2API.
- Replace Touch local signup grants once Sub2API balance/credit display is wired.

## 2026-06-09 Touch balance bridge

### Done
- Extended `POST /api/v1/touch/admin/users/sync` to return the Sub2API user's `balance`.
- Kept the existing sync endpoint as the single Touch identity/profile bridge instead of adding a separate balance-read endpoint.
- Added route coverage proving the balance field is returned through the Touch admin API.

### Validation
- `go test ./internal/server/routes ./internal/service ./internal/handler -run 'TestTouch|TestUserService|TestAuth' -count=1` passed from `backend/`.
- `git diff --check` passed.

### Next
- Add Touch-side paid-operation pre-deduct endpoints against Sub2API balance for image generation and WeChat export.
- Keep Touch local credit tables read-only/legacy until deduction paths have been migrated.

## 2026-06-10 Touch image-generation pre-deduct

### Done
- Added `POST /api/v1/touch/admin/users/deduct-balance`, protected by the existing Sub2API Touch admin API key middleware.
- Added `UserService.DeductTouchUserBalance`, converting Touch credits to Sub2API balance at `10 touch credits = 1 sub2api balance`.
- Added route and service tests for successful deduction and insufficient balance.

### Validation
- `go test ./internal/server/routes ./internal/service ./internal/handler -run 'TestTouch|TestUserService|TestAuth|TestDeductTouchUserBalance' -count=1` passed from `backend/`.
- `git diff --check` passed.

### Next
- Move WeChat export task creation to the same pre-deduct endpoint.
- Consider replacing the current read-then-update deduction with a repository-level conditional update if strict concurrent no-overdraft semantics are required.

## 2026-06-10 Touch WeChat-export pre-deduct

### Done
- Reused `POST /api/v1/touch/admin/users/deduct-balance` for Touch WeChat export task creation.
- No new Sub2API endpoint was needed; image generation and WeChat export now share the same Touch paid-operation deduction bridge.

### Validation
- `go test ./internal/server/routes ./internal/service ./internal/handler -run 'TestTouch|TestUserService|TestAuth|TestDeductTouchUserBalance' -count=1` passed from `backend/`.
- `git diff --check` passed.

### Next
- Add operation/idempotency metadata to Touch deduction requests if duplicate-submit protection becomes required.

## 2026-06-10 Touch payment balance grant bridge

### Done
- Added `POST /api/v1/touch/admin/users/grant-balance`, protected by the existing Touch admin API-key middleware.
- Added `UserService.GrantTouchUserBalance`, converting Touch payment credits to Sub2API balance at `10 touch credits = 1 sub2api balance`.
- The grant path syncs/materializes the Touch user by email before balance grant, so payment success can top up newly synced Touch users.
- Added route and service tests for the grant bridge.

### Validation
- `go test ./internal/server/routes ./internal/service ./internal/handler -run 'TestTouch|TestUserService|TestAuth|TestDeductTouchUserBalance|TestGrantTouchUserBalance' -count=1` passed from `backend/`.
- `git diff --check` passed.

### Next
- Add a durable idempotency key/ledger for Touch payment grant events before removing all Touch-local payment/order screens.

## 2026-06-10 Touch payment grant idempotency ledger

### Done
- Added migration `146_touch_balance_events.sql` with a unique `event_key` ledger for Touch balance events.
- Extended `POST /api/v1/touch/admin/users/grant-balance` to accept `event_key` and return `idempotent`.
- Implemented `GrantTouchBalanceOnce` in the real user repository: it inserts the Touch balance event and updates the user balance in one transaction; duplicate `event_key` requests return success without adding balance again.
- Updated `UserService.GrantTouchUserBalance` to use the persistent event ledger when `event_key` is present, while keeping the old non-event path for compatibility.

### Validation
- `go test ./internal/server/routes -run 'TestTouchGrantBalanceRoute|TestTouchDeductBalanceRoute|TestTouchAdmin'` passed from `backend/`.
- `go test ./internal/handler -run TestNonexistent` passed from `backend/`.
- `go test ./internal/service -run TestNonexistent` passed from `backend/`.
- `go test ./internal/repository -run TestNonexistent` passed from `backend/`.
- `git diff --check` passed.
- `go test -tags unit ./internal/service -run 'TestGrantTouchUserBalance|TestDeductTouchUserBalance'` is currently blocked by pre-existing unit-test compile issues in unrelated service tests (`valueOrEmpty` redeclaration and stale `newAuthService`/quota test helpers).

### Next
- Migrate checkout/provider/webhook/order administration into Sub2API so Touch no longer owns the payment domain itself.
- Add an admin repair/reporting surface for Touch payment grant events after the order domain moves.

## 2026-06-10 Touch checkout bridge into Sub2API orders

### Done
- Added `TouchPaymentHandler` with `POST /api/v1/touch/admin/payments/checkout`.
- The endpoint syncs/materializes the Touch user by email, then calls Sub2API `PaymentService.CreateOrder` as that Sub2API user.
- Added the handler to Touch routes, wire provider setup, and the generated server wiring.
- Added route coverage proving the checkout bridge syncs the Touch user and creates a Sub2API balance order with the expected `payment_source=touch`.

### Validation
- `go test ./internal/server/routes -run 'TestTouchPaymentCheckoutRoute|TestTouchGrantBalanceRoute|TestTouchDeductBalanceRoute|TestTouchAdmin'` passed from `backend/`.
- `go test ./internal/handler -run TestNonexistent` passed from `backend/`.
- `go test ./cmd/server -run TestNonexistent` passed from `backend/`.
- `git diff --check` passed.

### Next
- Move Touch pricing package reads to Sub2API payment catalog/config so checkout amount, recharge multiplier, and payment methods all come from one source.
- Expose or reuse Sub2API order list/detail APIs for Touch payment admin pages.

## 2026-06-10 Touch OAuth frontend redirect override

### Done
- Added a short-lived `email_oauth_frontend_callback` cookie to Google/GitHub OAuth start so Touch can pass an absolute `frontend_redirect_url` for the current OAuth attempt.
- The Google/GitHub callback now prefers that per-attempt callback override before falling back to configured Sub2API frontend redirect URLs.
- Added handler coverage for storing the Touch callback override.

### Validation
- `go test ./internal/handler -run 'Test(EmailOAuthStartStoresFrontendRedirectOverride|GoogleOAuth|GitHubOAuth|EmailOAuth)' -count=1` passed from `backend/`.
- `git diff --check` passed.

### Next
- Pending/manual OAuth completion still returns to the supplied frontend callback without tokens; Touch currently handles token-return completions and leaves manual completion UX to a later migration slice.

## 2026-06-10 Touch identity source separation

### Done
- Added `touch` as a first-class `users.signup_source` and `auth_identities.provider_type`.
- Changed Touch user sync/balance paths to identify users by `auth_identities(provider_type='touch', provider_key='touch', provider_subject=touch_user_id)` instead of email.
- Changed native email lookups to exclude `signup_source='touch'`, so Touch-marked users cannot log in through normal Sub2API email/password auth.
- Added migration `147_touch_identity_source_separation.sql`:
  - allows `touch` in source/provider check constraints
  - replaces the active email unique index with a non-Touch-only unique index
  - removes legacy email auth identities from Touch users
  - backfills Touch auth identities from historical `notes` values like `touch_user_id=...`
- Added service coverage proving a Touch user is created separately even when the same email already exists as a native Sub2API user.

### Validation
- `go test ./internal/service ./internal/repository ./internal/server/routes ./internal/handler ./ent/schema -run 'Test(SyncTouchUserCreatesTouchUserEvenWhenNativeEmailExists|DeductTouchUserBalance|GrantTouchUserBalance|Touch|AuthIdentityFoundationSchemas|EmailOAuthStartStoresFrontendRedirectOverride|GoogleOAuth|GitHubOAuth|EmailOAuth)' -count=1` passed from `backend/`.
- `git diff --check` passed.

### Next
- Touch email/password UI now creates local better-auth users and syncs them as `touch`; keep watching for any remaining Touch backend path that calls native `/api/v1/auth/register` or `/api/v1/auth/login` directly.

## 2026-06-10 Touch OAuth pending/manual bypass

### Done
- Added a short-lived `email_oauth_source=touch` cookie when Google/GitHub OAuth starts with `source=touch`.
- Touch-sourced OAuth callbacks now skip Sub2API pending/manual completion when a verified email first-login would otherwise require manual GitHub completion or invite-only completion.
- Added `AuthService.LoginOrRegisterVerifiedEmailOAuthBypassInvitation` so the Touch-specific bypass is explicit and does not affect normal Sub2API frontend OAuth.
- Added handler coverage proving Touch source is stored and an invite-only GitHub first login returns tokens without creating a pending session.

### Validation
- `go test ./internal/handler ./internal/service -run 'Test(EmailOAuthStartStoresFrontendRedirectOverride|EmailOAuthStartStoresTouchSource|EmailOAuthCallbackTouchSourceSkipsPendingManualCompletion|GoogleOAuth|GitHubOAuth|EmailOAuth|LoginOrRegisterVerifiedEmailOAuth)' -count=1` passed from `backend/`.
- `git diff --check` passed.

### Next
- Keep this bypass scoped to `source=touch`; do not apply it to Sub2API-native OAuth buttons unless product policy changes.

## 2026-06-10 Touch pricing catalog source of truth

### Done
- Extended the public payment catalog response with `enabled_payment_types`, `balance_disabled`, and `recharge_fee_rate` so Touch can render recharge packages without local payment config.
- Updated the Touch checkout bridge to resolve balance recharge `product_id` through Sub2API payment config and override the client-supplied amount with the configured recharge product amount.
- Added route coverage proving Touch checkout ignores a tampered client amount and rejects unknown recharge products.

### Validation
- `go test ./internal/server/routes ./internal/handler -run 'TestTouchPaymentCheckoutRoute|TestTouchCapabilitiesRoute' -count=1` passed from `backend/`.
- `git diff --check` passed.
- `go test -tags unit ./internal/handler -run TestPaymentHandlerGetPublicCatalog -count=1` is blocked by pre-existing unrelated unit-test compile errors in `user_affiliate_handler_test.go` and `user_handler_test.go` for stale `NewUserHandler` arguments.

### Next
- Finish moving Touch pricing/payment admin configuration to Sub2API; Touch admin pricing pages still use local tables.
- Restore subscription checkout by mapping Touch subscription products to Sub2API subscription plans.

## 2026-06-10 Touch subscription checkout bridge

### Done
- Extended the Touch payment checkout bridge request with `plan_id`.
- Touch subscription orders now pass through to `PaymentService.CreateOrder` with `order_type=subscription` and the selected Sub2API subscription plan ID.
- Added route coverage proving subscription checkout preserves `plan_id`, return URL, and subscription order type without applying recharge-product amount overrides.

### Validation
- `go test ./internal/server/routes ./internal/handler ./cmd/server -run 'TestTouchPaymentCheckoutRoute|TestTouchCapabilitiesRoute|TestNonexistent' -count=1` passed from `backend/`.
- `git diff --check` passed.

### Next
- Move Touch billing/subscription display pages to Sub2API user subscription/order APIs so successful subscription purchases are visible in Touch without local subscription rows.

## 2026-06-10 Touch subscription list bridge

### Done
- Added `TouchSubscriptionHandler` with `POST /api/v1/touch/admin/subscriptions/list`.
- The endpoint syncs/materializes the Touch user, then returns only that Sub2API user's subscriptions using the existing subscription service and public user subscription DTO.
- Registered the handler in Touch routes, wire provider setup, and generated server wiring.
- Added route coverage proving the list endpoint syncs the Touch identity and queries subscriptions for the synced Sub2API user ID.

### Validation
- `go test ./internal/server/routes ./internal/handler ./cmd/server -run 'TestTouchSubscriptionListRoute|TestTouchPaymentCheckoutRoute|TestTouchCapabilitiesRoute|TestNonexistent' -count=1` passed from `backend/`.
- `git diff --check` passed.

### Next
- Add a user-order bridge for `/settings/payments`, then retire Touch-local order history from user-facing pages.

## 2026-06-10 Touch payment order list bridge

### Done
- Added `POST /api/v1/touch/admin/payments/orders`.
- The endpoint syncs/materializes the Touch user, then returns only that synced Sub2API user's payment orders through `PaymentService.GetUserOrders`.
- Reuses the existing sanitized payment order DTO, so Touch can show order history without exposing provider internals or reading its local order table.
- Added route coverage proving the endpoint syncs the Touch identity, forwards pagination/filter params, and scopes the query to the synced Sub2API user ID.

### Validation
- `go test ./internal/server/routes ./internal/handler ./cmd/server -run 'TestTouchPaymentOrdersRoute|TestTouchSubscriptionListRoute|TestTouchPaymentCheckoutRoute|TestTouchCapabilitiesRoute|TestNonexistent' -count=1` passed from `backend/`.
- `git diff --check` passed.

### Next
- Move any remaining user-facing invoice/retrieve behavior or explicitly retire it if Sub2API will not expose invoice metadata to Touch.

## 2026-06-17 Touch Prompt Gallery labels settings migration

### Done
- Added `touch_prompt_gallery_labels_config` to Sub2API system/public settings so Prompt Gallery labels can be managed from Sub2API instead of only Touch locale JSON.
- Exposed the new JSON field in the Sub2API admin settings API and Vue admin settings page with zh/en help text.
- Updated Touch public config mapping and Prompt Cases/Templates pages to merge locale-scoped runtime label overrides with existing local fallback labels.
- Added a focused Touch helper for safe partial label merges; invalid JSON or non-string leaves fall back to the local Touch locale copy.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_(GetPublicSettings_ExposesTouchRuntimeSettings|UpdateSettings_PersistsTouchRuntimeSettings)' -count=1 -v`
- `go test ./internal/handler/dto -run TestPublicSettingsInjectionPayload_SchemaDoesNotDrift -count=1 -v`
- `go test ./internal/service ./internal/handler ./internal/handler/dto -count=1`
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-config.test.ts tests/prompt-gallery-labels-config.test.ts`
- `pnpm --dir apps/touch exec tsc --noEmit --pretty false`
- `pnpm --dir frontend exec vue-tsc --noEmit`
- `pnpm --dir frontend build`
- `make build-touch`
- `git diff --check`

### Notes
- Frontend build still emits existing browserslist/chunk warnings; Touch build still emits the existing `baseline-browser-mapping` stale-data warning.
- Next useful slice: move Prompt Gallery section/hero copy and remaining local fallback calculations into Sub2API public settings/API responses, or start the larger Vue-front-end consolidation.

## 2026-06-17 Touch Prompt Gallery page shell settings migration

### Done
- Added `touch_prompt_gallery_page_config` to Sub2API system/public settings so Prompt Cases/Templates page shell copy can be managed as locale-scoped JSON.
- Exposed the new field in public settings, admin settings read/write APIs, and the Sub2API Vue settings page with zh/en guidance.
- Updated Touch to map `touch_prompt_gallery_page_config` to `prompt_gallery_page_config`.
- Prompt Cases/Templates pages now merge runtime page config over existing title/description fallbacks; the templates page no longer hardcodes `template forge`.
- Added focused Touch tests for locale/default scoped page shell overrides and invalid JSON fallback.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_(GetPublicSettings_ExposesTouchRuntimeSettings|UpdateSettings_PersistsTouchRuntimeSettings)' -count=1 -v`
- `go test ./internal/handler/dto -run TestPublicSettingsInjectionPayload_SchemaDoesNotDrift -count=1 -v`
- `go test ./internal/service ./internal/handler ./internal/handler/dto -count=1`
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-config.test.ts tests/prompt-gallery-labels-config.test.ts tests/prompt-gallery-page-config.test.ts`
- `pnpm --dir apps/touch exec tsc --noEmit --pretty false`
- `pnpm --dir frontend exec vue-tsc --noEmit`
- `pnpm --dir frontend build`
- `make build-touch`
- `git diff --check`

### Notes
- Frontend build still emits existing browserslist/chunk warnings; Touch build still emits the existing `baseline-browser-mapping` stale-data warning.
- Remaining Prompt Gallery slimming is mostly interaction/view-model logic: local facet/stat fallback calculations, local sorting fallback, and Gallery component state orchestration.

## 2026-06-17 Prompt Catalog server-side sort contract

### Done
- Added explicit `sort_by` / `sort_order` handling to the public prompt catalog API, backed by a repository SQL whitelist.
- Default prompt catalog sorting is now `imported_at desc`, with supported sort fields `imported_at`, `title`, `created_at`, and `updated_at`.
- Touch catalog fetches now always send explicit server-side sort params.
- Prompt Gallery trusts Sub2API order for cases/templates pages when API summary data is present; local imported-date sorting is retained only as a fallback for non-Sub2API-backed/mixed cases.
- Added service, route, and Touch request tests for prompt catalog sorting.

### Validation
- `go test ./internal/service -run TestTouchPromptService -count=1 -v`
- `go test ./internal/server/routes -run 'Test.*Prompt.*' -count=1 -v`
- `go test ./internal/service ./internal/server/routes ./internal/handler -count=1`
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-catalog.test.ts`
- `pnpm --dir apps/touch exec tsc --noEmit --pretty false`
- `make build-touch`
- `git diff --check`

### Notes
- Touch still owns client-side search/filter state and fallback facet/stat/grouping calculations; this slice moves only the ordering contract to Sub2API.

## 2026-06-17 Prompt Gallery server-backed view-model trust

### Done
- Prompt Gallery now has a clear `serverBackedCatalogReady` boundary for cases/templates pages.
- When Sub2API-backed data and summary are ready, Touch no longer reapplies local search/source/category/image filtering or local imported-date sorting.
- Cases total stats now read from the Sub2API summary instead of `filtered.length` on server-backed views.
- Local filter/sort behavior remains only while a remote filter request is pending or for non-server-backed/mixed fallback views.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-catalog.test.ts`
- `pnpm --dir apps/touch exec tsc --noEmit --pretty false`
- `make build-touch`
- `git diff --check`

### Notes
- Facet lists and template grouping still have Touch-side fallback derivations; the next slice can move grouped template view-model or facet payloads further into Sub2API.

## 2026-06-17 Prompt Catalog template groups payload

### Done
- Added `template_groups` to the Sub2API prompt catalog summary payload.
- The prompt repository now computes template category group counts from the same filtered summary query used for facets and stats.
- Touch normalizes `template_groups` into `summary.templateGroups` and uses the server-provided group order for template pages when available.
- Local template grouping remains as a fallback for non-server-backed or partially loaded views.

### Validation
- `go test ./internal/service -run TestTouchPromptService -count=1 -v`
- `go test ./internal/server/routes -run 'Test.*Prompt.*' -count=1 -v`
- `go test ./internal/service ./internal/server/routes ./internal/handler -count=1`
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-catalog.test.ts`
- `pnpm --dir apps/touch exec tsc --noEmit --pretty false`
- `make build-touch`
- `git diff --check`

### Notes
- Touch build still emits the existing `baseline-browser-mapping` stale-data warning.
- Remaining local Prompt Gallery logic is category/source facet presentation, local fallback view-models, and component interaction state.

## 2026-06-17 Root pnpm workspace for Vue and Touch frontends

### Done
- Added a repository root `package.json`, `pnpm-workspace.yaml`, and root `pnpm-lock.yaml` covering `frontend` and `apps/touch`.
- Moved package override policy from child package manifests to the root workspace so pnpm applies it consistently.
- Updated root Makefile frontend/touch targets to call root workspace scripts instead of per-directory pnpm commands.
- Updated CI, release workflow, production runbook, README variants, and Touch integration docs to install/build through the root workspace entrypoint.
- Added root scripts for Vue build/dev/typecheck/test and Touch build/dev/typecheck/lint/test/deploy.

### Validation
- `pnpm install --frozen-lockfile --lockfile-only`
- `pnpm run touch:typecheck`
- `pnpm run frontend:typecheck`
- `make test-frontend-critical`
- `pnpm run frontend:build`
- `pnpm run touch:build`

### Notes
- Vue build still emits existing browserslist/chunk warnings.
- Touch build still emits the existing `baseline-browser-mapping` stale-data warning.
- This unifies install/build/test entrypoints but does not yet merge Touch pages into the Vue frontend runtime.

## 2026-06-17 Public prompt API surface cleanup

### Done
- Removed the duplicate public `/api/v1/touch/prompts/cases` and `/api/v1/touch/prompts/cases/:id` aliases from Touch routes.
- Kept the shared public prompt catalog under `/api/v1/prompts/cases` and `/api/v1/prompts/cases/:id`.
- Updated prompt route tests so filter/summary coverage targets the shared route and old Touch public alias returns 404.

### Validation
- `go test ./internal/server/routes -run 'Test.*Prompt.*' -count=1 -v`
- `go test ./internal/service ./internal/server/routes ./internal/handler -count=1`
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-catalog.test.ts`
- `rg -n "/api/v1/touch/prompts|touch/prompts/cases" backend apps/touch/src apps/touch/tests`

### Notes
- The remaining `/api/v1/touch/web/*` routes still carry Touch browser-session semantics and are intentionally not removed in this slice.

## 2026-06-17 Touch public site config fallback consolidation

### Done
- Added a shared Touch `public-site-config` helper for public app URL, name, logo, and preview image fallbacks.
- Updated public-facing not-found, robots, sitemap, auth layout, auth metadata, docs shell, static-page metadata, copyright, error boundary, and SEO helpers to use the shared runtime config helper.
- Public site chrome now prefers Sub2API public settings, then public Sub2API base URL where appropriate, then the remaining local env/default fallback.
- Removed scattered `touch`, `/logo.svg`, and `http://localhost:3000` fallback decisions from public UI modules; remaining occurrences are the centralized helper and bottom-level env default.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-site-config.test.ts tests/public-config.test.ts`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `rg -n "http://localhost:3000|app_name \\|\\| 'touch'|app_logo \\|\\| '/logo.svg'|app_preview_image \\|\\| '/preview.png'" apps/touch/src apps/touch/tests`
- `git diff --check`

### Notes
- Touch build still emits the existing `baseline-browser-mapping` stale-data warning.
- `defaultLocale` remains tied to Next/i18n static routing and still needs a larger routing-level design before it can be fully runtime-managed.

## 2026-06-17 Prompt Gallery display counts prefer Sub2API summary

### Done
- Added `getPromptGalleryDisplayCount` as the single helper for choosing server-backed summary counts versus local fallback counts.
- Prompt Gallery filter "All" badge now uses the Sub2API summary total when the catalog is server-backed.
- Prompt Gallery compact sidebar total row now uses `stats.total`, which is already sourced from Sub2API summary on server-backed cases/templates pages.
- Local counts remain as fallback for mixed/non-server-backed states.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/prompt-gallery-catalog.test.ts`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Touch still owns UI interaction state and pagination display, but total-count display is no longer tied to local current-array length on server-backed views.

## 2026-06-17 Touch public UI reads browser-safe public configs

### Done
- Public Touch pages and services now read `getPublicConfigs()` instead of `getEnvironmentConfigs()`.
- Covered root layout scripts, SEO metadata, ads.txt, docs shell/pages, landing/static pages, pricing, AI workspace, settings shell, auth metadata, prompt gallery pages, and public ads/analytics/affiliate/customer-service helpers.
- `getEnvironmentConfigs()` is now only used inside the config module as the fallback source for public config construction.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-config.test.ts tests/public-site-config.test.ts`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `rg -n "getEnvironmentConfigs\\(" apps/touch/src/app apps/touch/src/shared apps/touch/tests`
- `git diff --check`

### Notes
- This tightens the Touch UI shell around Sub2API public settings and reduces direct local env participation in public rendering.
- Touch build still emits the existing `baseline-browser-mapping` stale-data warning.

## 2026-06-17 Production public settings are Sub2API-required

### Done
- `getPublicConfigs()` and the config fallback path now throw in production if `SUB2API_BASE_URL` is missing.
- Production also fails if `/api/v1/settings/public` returns an error or cannot provide usable data, instead of silently falling back to Touch env values.
- Docs content pages are now dynamic SSR so production runtime reads Sub2API public settings while local/CI builds without Sub2API can still complete.
- Updated Touch README and `.env.example` to state that production runtime must configure `SUB2API_BASE_URL` and have public settings reachable.

### Validation
- `pnpm --dir apps/touch exec node --import tsx --test tests/public-config.test.ts tests/public-site-config.test.ts`
- `pnpm run touch:typecheck`
- `pnpm run touch:build`
- `git diff --check`

### Notes
- Local/test environments still retain env fallback for development convenience.
- Touch build still emits the existing `baseline-browser-mapping` stale-data warning.

## 2026-06-17 Sub2API Vue Prompt Catalog entry

### Done
- Added a Sub2API Vue API wrapper for the shared public prompt catalog endpoint under `/api/v1/prompts/cases`.
- Added a public Vue route at `/touch/prompts` as the first Sub2API frontend entry for the migrated Touch prompt catalog.
- The new Vue page supports server-backed search, source type, source project, category, image-only filtering, imported-time sorting, and pagination.
- The page renders prompt images from the URLs returned by Sub2API/R2 and no longer depends on Touch-local prompt catalog data or Touch BFF routes.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:build`
- `rg -n "PromptCatalogView|promptsAPI|TouchPromptCatalog|/touch/prompts|PromptCatalog" frontend/src`
- Browser smoke on `http://localhost:5174/touch/prompts` rendered the Vue route and visible error state when local backend `:8080` was not running.
- `git diff --check`

### Notes
- This is the first concrete page-level move into the Sub2API Vue frontend, not a full Touch UI parity migration.
- Touch Next still owns the richer Prompt Gallery UI, import panel, image-generation entry points, docs/landing shell, and related interaction polish.
- Local browser smoke had expected proxy 500s for `/setup/status`, `/api/v1/settings/public`, and `/api/v1/prompts/cases` because the backend service was not running.

## 2026-06-18 Public runtime settings expose generic aliases

### Done
- Added generic public settings aliases for workspace shell, pricing shell/text, and credits text:
  - `workspace_shell_config`
  - `pricing_title`
  - `pricing_description`
  - `pricing_shell_config`
  - `credits_title`
  - `credits_description`
  - `credits_purchase_label`
  - `credits_balance_label`
- Kept the existing `touch_*` public fields as compatibility output sourced from the same stored settings.
- Updated the Vue Image Generator, Pricing, and Credits pages to prefer the generic fields and fall back to the legacy `touch_*` fields.
- Added service and handler regression assertions so public settings expose both the generic aliases and legacy fields.

### Validation
- `go test -tags unit ./internal/service -run PublicSettings -count=1`
- `go test -tags unit ./internal/handler -run GenericRuntimeSettingAliases -count=1`
- `go test ./internal/handler -count=1`
- `go test ./internal/handler/dto -count=1`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler ./internal/handler/dto`
- `git diff --check`

### Notes
- The admin settings form and database keys still use the existing `touch_*` names for compatibility; this step only moves public runtime consumption to generic aliases.
- Remaining cleanup can continue with prompt gallery text/content aliases and then the broader admin/settings naming cleanup.

## 2026-06-18 Prompt catalog public text uses generic aliases

### Done
- Added generic public settings aliases for prompt catalog titles and descriptions:
  - `prompt_cases_title`
  - `prompt_cases_description`
  - `prompt_templates_title`
  - `prompt_templates_description`
- Kept the existing `touch_prompt_*` fields as compatibility output sourced from the same stored settings.
- Updated the Vue Prompt Catalog page to use generic prompt case/template text first and fall back to the legacy fields.
- Removed the hardcoded visible "Touch Prompt Catalog" eyebrow and Touch-specific default catalog description from the public page.
- Removed Touch-specific default wording from the public Image Generator workspace description and the Credits page console tag.
- Extended service and handler tests to assert prompt text aliases alongside legacy compatibility fields.

### Validation
- `go test -tags unit ./internal/service -run PublicSettings -count=1`
- `go test -tags unit ./internal/handler -run GenericRuntimeSettingAliases -count=1`
- `go test ./internal/handler/dto -count=1`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler ./internal/handler/dto`
- `git diff --check`
- `rg -n "Touch Prompt Catalog|Browse Touch prompt|Sub2API 中的 Touch|cachedPublicSettings\\?\\.touch_prompt_(cases|templates)" frontend/src/views/public/PromptCatalogView.vue frontend/src/types/index.ts -S`
- `rg -n "Touch Prompt Catalog|Browse Touch prompt|Sub2API 中的 Touch|original Touch prompt|Touch 原有|\\[touch-credits\\]" frontend/src/views/public frontend/src/views/user -S`

### Notes
- The remaining `touch_prompt_*` reads in `PromptCatalogView.vue` are intentional legacy fallback paths while existing stored settings still use old keys.
- Admin settings still writes the old `touch_*` keys; a later pass can add generic admin fields or migrate persisted keys.

## 2026-06-18 Runtime settings admin entry made generic

### Done
- Renamed the Vue admin settings component from `TouchSettingsView` to `RuntimeSettingsView`.
- Changed the primary admin route from `/admin/touch/settings` to `/admin/runtime-settings`, keeping `/admin/touch/settings` as a compatibility alias.
- Renamed the route from `AdminTouchSettings` to `AdminRuntimeSettings`.
- Updated the sidebar entry and navigation copy from `Touch Settings` / `Touch 配置` to `Runtime Settings` / `运行配置`.
- Moved admin i18n keys from `admin.settings.touch.*` to `admin.settings.runtime.*` and removed visible Touch-specific wording from this page.
- Kept the persisted payload fields as `touch_*` because those are still the current stored setting keys.
- Added router coverage proving `/admin/runtime-settings` is primary and `/admin/touch/settings` is only an alias.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/RuntimeSettingsView.spec.ts src/router/__tests__/runtime-settings-route.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `rg -n "TouchSettingsView|AdminTouchSettings|admin\\.settings\\.touch|nav\\.touchSettings|touchSettings|Touch Settings|Touch 配置|Touch Frontend|Touch 前端|Touch Site URL|Touch 站点地址|Touch local env|Touch 本地文案|Touch settings" frontend/src -S`
- `git diff --check`

### Notes
- The old `/admin/touch/settings` URL remains for compatibility only.
- The admin form still writes legacy `touch_*` setting keys; a later persistence migration can introduce generic stored keys once compatibility policy is decided.

## 2026-06-18 Legacy Touch route registration marked compatibility-only

### Done
- Renamed the backend route registration function from `RegisterTouchRoutes` to `RegisterLegacyTouchRoutes`.
- Updated the server router and route tests to call the legacy-specific function name.
- Updated the route comment to describe `/api/v1/touch/*` as compatibility URLs while active surfaces move to generic Sub2API routes.
- Added `mode: "legacy_compatibility"` to `/api/v1/touch/capabilities` so clients and tests can distinguish it from a primary API surface.

### Validation
- `go test ./internal/server/routes -run 'TestTouch|Test.*Prompt' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/server ./internal/server/routes`
- `rg -n "RegisterTouchRoutes|registers routes used by the migrated Touch frontend" backend/internal/server -S`

### Notes
- The `/api/v1/touch/*` URLs are still registered for compatibility; this step changes ownership semantics, not the external URL contract.

## 2026-06-18 Prompt catalog backend internals renamed

### Done
- Renamed the prompt catalog backend service, handler, repository, and tests from `TouchPrompt*` file/type naming to `PromptCatalog*`.
- Renamed the X/Twitter import service and handler internals from `TouchTwitterImport*` to `TwitterImport*`.
- Updated Wire/server route assembly to inject `PromptCatalog` and `TwitterImport` handlers.
- Changed prompt image sync and Twitter import user agents from Touch-specific names to Sub2API prompt catalog names.
- Removed duplicate import helper code introduced during the rename and reused the existing service-level `firstNonEmpty` helper.
- Kept the existing `touch_prompt_items` table, `touch_prompt_*` settings keys, and `/api/v1/touch/*` compatibility URLs unchanged.

### Validation
- `go test ./internal/service -run 'TestPromptCatalog|TestTwitterImport|TestValidate|TestPrompt|Test.*Image' -count=1`
- `go test ./internal/handler ./internal/repository ./internal/server/routes ./cmd/server -run 'Test|^$' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler ./internal/repository ./internal/server/routes ./cmd/server`
- `rg -n "TouchPrompt|touchPrompt|TOUCH_PROMPT|NewTouchPrompt|ErrTouchPrompt|TouchPromptCase|TouchPromptService|touch_prompt_handler|touch_prompt_repo|touch_prompt_test|sub2api-touch-prompt|TouchTwitter|touchTwitter|NewTouchTwitter|sub2api-touch-twitter" backend/internal backend/cmd -S`
- `rg -n "TouchHTTPClient|touchTweetPayload|normalizeTouch|cleanTouch|extractTouch|sanitizeTouch|isTouchStorage|clearTouch|buildTouch|detectTouch|loadTouch|newTouch|putTouch|touchFirst|touchValue|touchXAuto" backend/internal/service/twitter_import.go backend/internal/service/twitter_import_test.go backend/internal/handler/twitter_import_handler.go -S`
- `git diff --check`

### Notes
- The remaining `TouchPrompt*` matches are compatibility DTO/settings fields for existing `touch_prompt_*` persisted keys.
- The remaining `TouchWeb*` and `/api/v1/touch/*` surface is still a legacy web compatibility surface and needs a separate endpoint-contract migration before removal.

## 2026-06-18 Generic web API aliases added

### Done
- Added `RegisterWebRoutes` for browser-safe web session, checkout, and Twitter import endpoints under `/api/v1/web/*`.
- Reused the same web route registration for legacy `/api/v1/touch/web/*` so old clients remain compatible without a separate route table.
- Registered the generic web routes before legacy Touch compatibility routes in the main backend router.
- Added route-table coverage for generic `/api/v1/web/*` and legacy `/api/v1/touch/web/*` aliases.
- Updated `docs/TOUCH_PLATFORM_INTEGRATION.md` to make `/api/v1/web/*` the preferred web surface.

### Validation
- `go test ./internal/server/routes -run 'TestWebRoutes|TestTouch|TestPromptCatalog' -count=1`
- `go test ./internal/server -run 'Test|^$' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/server ./internal/server/routes`
- `rg -n "RegisterWebRoutes|RegisterLegacyTouchRoutes|/api/v1/web|/api/v1/touch/web" backend/internal/server backend/internal/server/routes docs -S`
- `git diff --check`

### Notes
- The handler type is still `TouchWebHandler` and still enforces the `touch` signup/source boundary; this step only changes the public API path ownership.
- Legacy `/api/v1/touch/web/*` remains registered for existing clients and can be removed only after all clients use `/api/v1/web/*`.

## 2026-06-18 Frontend API bootstrap env made optional

### Done
- Audited the Sub2API Vue frontend API bootstrap path and confirmed it already defaults to same-origin `/api/v1`.
- Changed `VITE_API_BASE_URL` typing from required to optional so the type contract matches the runtime fallback.
- Kept `VITE_API_BASE_URL` as an override for split frontend/backend deployments.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check`

### Notes
- The remaining `import.meta.env.BASE_URL` usage is Vite router base configuration, not a Touch/Sub2API API discovery setting.
- Runtime API discovery no longer needs a Touch-specific `SUB2API_BASE_URL` or `NEXT_PUBLIC_SUB2API_BASE_URL` equivalent in the unified Vue frontend.

## 2026-06-18 Web session handler renamed to platform-neutral surface

### Done
- Renamed the browser session handler file and type from `TouchWebHandler` / `NewTouchWebHandler` to `WebHandler` / `NewWebHandler`.
- Renamed the top-level handler aggregate field from `Handlers.TouchWeb` to `Handlers.Web`.
- Updated route registration, Wire provider assembly, generated server wiring, and route tests to use the generic web handler field.
- Replaced remaining Web handler error messages that referred to "Touch OAuth", "Touch session", or "Touch Twitter importer" with platform-neutral wording.
- Kept `touchAuthSource`, existing cookie names, and default `payment_source: touch` unchanged for identity separation and existing browser session/order compatibility.

### Validation
- `go test ./internal/handler ./internal/server/routes ./cmd/server -run 'Test|^$' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/handler ./internal/server ./internal/server/routes ./cmd/server`
- `rg -n "TouchWebHandler|NewTouchWebHandler|TouchWeb|touchWeb|touch_web_handler|Touch OAuth|Touch Twitter importer|Touch session" backend/internal/handler backend/internal/server backend/cmd -S`
- `git diff --check`

### Notes
- This is an internal ownership rename only. The legacy `/api/v1/touch/web/*` alias remains registered and backed by the same `WebHandler`.
- The `touch` signup source is still required by product rules: Touch-origin users remain separate from native Sub2API users even if the email is the same.

## 2026-06-18 Prompt catalog frontend trusts API summary

### Done
- Initialized Prompt Catalog summary with the API response shape instead of nullable local fallback state.
- Changed hero stats to render `summary.total`, `summary.source_count`, `summary.case_count`, and `summary.template_count` directly from Sub2API API data.
- Removed the local `sourceOptions.length` fallback count.
- Removed manual list prepending after Twitter import; the page now refetches the catalog and lets Sub2API define ordering, filtering, and pagination.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `rg -n "summary\\?\\.|sourceOptions\\.length|items\\.value = \\[data\\.item|PromptCatalogSummary \\| null" frontend/src/views/public/PromptCatalogView.vue -S`
- `git diff --check`

### Notes
- Prompt Catalog still owns visual layout, selection state, and navigation in Vue.
- The remaining display tag merging in `allTags()` is presentation-only; model/display/source tags already come from the Prompt API.

## 2026-06-18 Credits shell content moved to public settings

### Done
- Added a generic Sub2API setting key `credits_shell_config` for Credits page presentation copy.
- Exposed `credits_shell_config` through service public settings, injection payloads, handler DTOs, and admin settings responses/updates.
- Added Runtime Settings admin form support for editing `credits_shell_config`.
- Updated the Credits page to parse `credits_shell_config` by locale and use it for:
  - actions title
  - actions description
  - recharge/orders button labels
  - conversion text
- Kept existing `touch_credits_*` title/description/purchase/balance fields unchanged for compatibility.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_ExposesTouchRuntimeSettings|TestSettingService_UpdateSettings_PersistsTouchRuntimeSettings' -count=1`
- `go test ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings_ExposesGenericRuntimeSettingAliases|Test' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/service ./internal/handler ./internal/handler/dto`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `rg -n "credits_shell_config|creditsShellConfig|CreditsShellConfig|creditsShell" backend/internal frontend/src -S`
- `git diff --check`

### Notes
- This reduces local Credits page copy, but the page layout and balance presentation remain in Vue.
- The legacy `touch_credits_*` settings are still emitted and editable; full key migration can happen after existing stored settings are migrated or compatibility is no longer required.

## 2026-06-18 Legacy Touch admin bridge handlers marked explicitly

### Done
- Renamed legacy admin bridge handlers from primary-looking names to explicit compatibility names:
  - `TouchAuthHandler` -> `LegacyTouchAuthHandler`
  - `TouchPaymentHandler` -> `LegacyTouchPaymentHandler`
  - `TouchSubscriptionHandler` -> `LegacyTouchSubscriptionHandler`
- Renamed the backing files to `legacy_touch_*_handler.go`.
- Updated route registration, handler aggregation, Wire provider assembly, generated server wiring, and route tests to use the `LegacyTouch*` handler names.
- Added comments on each handler explaining that it serves legacy `/api/v1/touch/admin/*` clients and should not be used as the new platform API surface.
- Kept all legacy URL paths and JSON contracts unchanged, including `touch_user_id` and `touch_credits`.

### Validation
- `go test ./internal/handler ./internal/server/routes ./cmd/server -run 'TestLegacyTouch|TestTouch|TestWebRoutes|^$' -count=1`
- `PATH="/Users/aias/go/bin:$PATH" golangci-lint run ./internal/handler ./internal/server/routes ./cmd/server`
- `rg -n "\\bTouchAuthHandler\\b|\\bNewTouchAuthHandler\\b|\\bTouchPaymentHandler\\b|\\bNewTouchPaymentHandler\\b|\\bTouchSubscriptionHandler\\b|\\bNewTouchSubscriptionHandler\\b|LegacyLegacyTouch|touch_auth_handler|touch_payment_handler|touch_subscription_handler" backend/internal/handler backend/internal/server backend/cmd -S`
- `git diff --check`

### Notes
- The legacy bridge still exists because external Touch-era admin clients may still call `/api/v1/touch/admin/*`.
- Future platform work should avoid adding to `LegacyTouch*`; new functionality should use generic web/admin/payment/subscription APIs instead.

## 2026-06-18 Pricing currency display moved into runtime settings

### Done
- Added `pricing_currency_symbol` as a Sub2API public/admin runtime setting with default `¥`.
- Exposed the setting through public settings, HTML injection settings, and admin runtime settings.
- Added the admin runtime settings field so operators can change the public pricing card currency symbol without editing the frontend.
- Updated the public Pricing page to read the symbol from Sub2API public settings instead of hardcoding `¥` in `formatCurrency`.
- Updated payment amount displays in the user checkout flow, Stripe popup/inline components, QR success dialog, order table, and recharge product cards to use the existing payment currency formatter instead of hardcoded `¥`.
- Kept checkout/payment amount calculation unchanged; this only changes UI formatting.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings|TestSettingService_UpdateSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/admin -run 'TestSettingHandler_GetPublicSettings_ExposesGenericRuntimeSettingAliases|TestSettingHandler_UpdateSettings_AcceptsGenericRuntimeAliases' -count=1`
- `go test ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `go test ./internal/service ./internal/handler ./internal/handler/admin ./cmd/server -run '^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n '¥\\{\\{|createOrder.*¥|CREDITS_PER_SUB2API_BALANCE|balance \\* 10|credits \\* 10' frontend/src/views/user frontend/src/views/public frontend/src/components/payment backend/internal -S`
- `git diff --check`

### Notes
- Pricing/payment UI layout still lives in Vue. This removes the remaining user-facing hardcoded payment symbol from the public/user/payment component path.
- Other pricing presentation rules remain local, including plan platform label mapping and validity/quota formatting.

## 2026-06-18 Pricing plan platform label moved into catalog DTO

### Done
- Added `group_display_label` to Sub2API payment plan DTOs returned by:
  - `/api/v1/payment/plans`
  - `/api/v1/payment/checkout-info`
  - `/api/v1/payment/public/catalog`
- Centralized platform-code display mapping in the backend payment handler:
  - `anthropic` -> `Claude`
  - `openai` -> `OpenAI`
  - `gemini` -> `Gemini`
  - `antigravity` -> `Antigravity`
  - unknown platforms fall back to group name, then raw platform.
- Updated public Pricing page to render `plan.group_display_label` instead of doing local `platformLabel()` mapping.
- Updated user purchase subscription confirmation and subscription plan cards to prefer `group_display_label` from checkout-info before falling back to local platform labels.
- Removed the now-unused `PaymentConfigService.GetGroupPlatformMap` helper.

### Validation
- `go test -tags unit ./internal/handler -run 'TestPaymentHandlerGetPublicCatalog' -count=1`
- `go test ./internal/handler ./internal/server/routes ./cmd/server -run '^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/views/public/__tests__/PricingView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "function platformLabel|platformLabel\\(plan|group_display_label|GroupDisplayLabel|planGroupDisplayLabel" backend/internal/handler/payment_handler.go backend/internal/handler/payment_handler_public_catalog_test.go frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts frontend/src/types/payment.ts -S`

### Notes
- Public Pricing no longer owns plan platform label mapping.
- User purchase plan cards now consume the same backend-provided display label, while still using local platform helpers for colors and fallback labels.
- Other pricing presentation rules remain local, especially validity text and quota label formatting.

## 2026-06-18 Pricing plan quota label moved into catalog DTO

### Done
- Added `quota_label` to Sub2API payment plan DTOs returned by:
  - `/api/v1/payment/plans`
  - `/api/v1/payment/checkout-info`
  - `/api/v1/payment/public/catalog`
- Centralized pricing quota display calculation in the backend payment handler by taking the maximum configured daily/weekly/monthly USD limit and formatting it as a compact dollar label.
- Updated the public Pricing page to render `plan.quota_label` from Sub2API and fall back only to the localized unlimited label.
- Removed the local `quotaLabel(plan)` array/max calculation from the Pricing page.
- Added catalog and Pricing page coverage for API-provided quota labels.

### Validation
- `go test -tags unit ./internal/handler -run 'TestPaymentHandlerGetPublicCatalog' -count=1`
- `go test ./internal/handler ./internal/server/routes ./cmd/server -run '^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "quotaLabel\\(|quota_label|planQuotaLabel|GetGroupPlatformMap\\(" backend/internal/handler/payment_handler.go backend/internal/handler/payment_handler_public_catalog_test.go backend/internal/service/payment_config_plans.go frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts frontend/src/types/payment.ts -S`
- `git diff --check`

### Notes
- Validity text still lives in the Vue presentation layer because it is locale-sensitive.
- User/payment plan detail views still render individual daily/weekly/monthly limits locally; this step only removes the public Pricing card's aggregate quota label calculation.

## 2026-06-18 Image generator workspace default shell moved to Sub2API settings

### Done
- Added a Sub2API-owned default `workspace_shell_config` JSON payload with zh/en labels for the public Image Generator workspace.
- Changed public/admin settings reads so empty `workspace_shell_config` returns the backend default while still honoring explicit configured values and legacy `touch_workspace_shell_config` fallback when the generic key is absent.
- Updated default settings initialization to seed `workspace_shell_config` with the backend default JSON for new installations.
- Removed the large zh/en default copy block from `ImageGeneratorView.vue`; the page now merges parsed public settings with only a minimal frontend resilience fallback.
- Added a frontend regression test proving Image Generator renders workspace copy from public settings.
- Added a backend regression test proving the default workspace shell config is valid JSON and contains the expected zh/en labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings|TestSettingService_GetPublicSettings_FallsBackToLegacyTouchRuntimeKeys' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/ImageGeneratorView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/PricingView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "AI 生图工作区|提示词工作台|AI Image Workspace|Prompt Workspace|copyEmpty|copySuccess|copyPromptLabel|defaultWorkspaceShellConfig|workspaceShellConfigSetting" backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go frontend/src/views/public/ImageGeneratorView.vue frontend/src/views/public/__tests__/ImageGeneratorView.spec.ts -S`
- `git diff --check`

### Notes
- Image Generator visual layout and prompt draft interaction still live in Vue.
- The remaining frontend fallback is intentionally small and only protects the page if settings fail to load or the JSON is invalid.

## 2026-06-18 Prompt catalog default shell moved to Sub2API settings

### Done
- Added a Sub2API-owned default `prompt_catalog_shell_config` JSON payload with zh/en labels for the public Prompt Catalog.
- Added separate configured labels for authenticated and anonymous account actions so the frontend can still choose the right button text from backend-provided copy.
- Changed public/admin settings reads so empty `prompt_catalog_shell_config` returns the backend default while still honoring explicit configured values.
- Updated default settings initialization to seed `prompt_catalog_shell_config` with the backend default JSON for new installations.
- Removed the large zh/en default copy block from `PromptCatalogView.vue`; the page now merges parsed public settings with only a minimal frontend resilience fallback.
- Replaced the hardcoded Prompt Catalog eyebrow in the template with the shell-provided label.
- Extended Prompt Catalog frontend coverage to prove configured shell copy controls the account action and eyebrow.
- Added backend coverage proving the default prompt catalog shell config is valid JSON and contains expected zh/en labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig|TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts src/views/public/__tests__/PricingView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "提示词案例库|直接浏览 Sub2API 中的提示词案例|defaultPromptCatalogShellConfig|FALLBACK_PROMPT_CATALOG_COPY|accountActionAuthenticated|Configured eyebrow" backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts -S`
- `git diff --check`

### Notes
- Prompt Catalog visual layout, filter state, pagination, and admin import interaction still live in Vue.
- The remaining frontend fallback is intentionally small and only protects the page if settings fail to load or the JSON is invalid.

## 2026-06-18 Pricing default shell moved to Sub2API settings

### Done
- Added a Sub2API-owned default `pricing_shell_config` JSON payload with zh/en labels, button text, and tab group labels for the public Pricing page.
- Changed public/admin settings reads so empty `pricing_shell_config` returns the backend default while still honoring explicit configured values and legacy `touch_pricing_shell_config` fallback when the generic key is absent.
- Updated default settings initialization to seed `pricing_shell_config` with the backend default JSON for new installations.
- Removed the large zh/en default copy block from `PricingView.vue`; the page now merges parsed public settings with only a minimal frontend resilience fallback.
- Extended Pricing frontend coverage to prove configured shell groups and buy button text control the page.
- Added backend coverage proving the default pricing shell config is valid JSON and contains expected zh/en labels, button text, and group labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPricingShellConfig|TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig|TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "价格与套餐|浏览由 Sub2API 统一配置|defaultPricingShellConfig|FALLBACK_PRICING_COPY|Configured recharge group|DefaultsPricingShellConfig" backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts -S`
- `git diff --check`

### Notes
- Pricing visual layout, tab state, catalog loading, and checkout navigation still live in Vue.
- Validity formatting remains frontend-local because it is locale-sensitive and tied to displayed plan duration.
- The remaining frontend fallback is intentionally small and only protects the page if settings fail to load or the JSON is invalid.

## 2026-06-18 Credits default shell moved to Sub2API settings

### Done
- Added a Sub2API-owned default `credits_shell_config` JSON payload with zh/en labels, action block copy, button labels, and conversion text for the Credits page.
- Changed public/admin settings reads so empty `credits_shell_config` returns the backend default while still honoring explicit configured values.
- Updated default settings initialization to seed `credits_shell_config` with the backend default JSON for new installations.
- Removed the large zh/en default copy block from `CreditsView.vue`; the page now merges parsed public settings with only a minimal frontend resilience fallback.
- Extended Credits frontend coverage to prove configured conversion, action block, and button labels control the page.
- Added backend coverage proving the default credits shell config is valid JSON and contains expected zh/en labels, conversion text, action title, and buttons.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsCreditsShellConfig|TestSettingService_GetPublicSettings_DefaultsPricingShellConfig|TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig|TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/CreditsView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/CreditsView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "积分余额|积分已由 Sub2API|defaultCreditsShellConfig|FALLBACK_CREDITS_COPY|Configured conversion|DefaultsCreditsShellConfig" backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go frontend/src/views/user/CreditsView.vue frontend/src/views/user/__tests__/CreditsView.spec.ts -S`
- `git diff --check`

### Notes
- Credits visual layout, balance presentation, and refresh interaction still live in Vue.
- The balance-to-credit calculation still reads `credits_per_balance` from Sub2API public settings.
- The remaining frontend fallback is intentionally small and only protects the page if settings fail to load or the JSON is invalid.

## 2026-06-18 Home default shell labels moved to Sub2API settings

### Done
- Added a Sub2API-owned default `home_shell_config` JSON payload with zh/en labels for the Home page hero, navigation, model matrix, footer, and model family copy.
- Changed public/admin settings reads so empty `home_shell_config` returns the backend default while still honoring explicit configured values.
- Updated default settings initialization to seed `home_shell_config` with the backend default JSON for new installations.
- Removed the large i18n-backed Home labels default from `HomeView.vue`; the page now merges parsed public settings with only a minimal frontend resilience fallback.
- Added backend coverage proving the default home shell config is valid JSON and contains expected zh/en labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsHomeShellConfig|TestSettingService_GetPublicSettings_DefaultsCreditsShellConfig|TestSettingService_GetPublicSettings_DefaultsPricingShellConfig|TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig|TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Home visual layout, model matrix behavior, and card merging still live in Vue.
- Home experience/why card default titles and descriptions still use frontend i18n and should be the next shell-config slice.
- The remaining frontend label fallback is intentionally small and only protects the page if settings fail to load or the JSON is invalid.

## 2026-06-18 Home default cards moved to Sub2API settings

### Done
- Extended the Sub2API-owned default `home_shell_config` JSON payload with zh/en `experienceCards` and `whyChooseCards`.
- Kept the existing Home card override shape, so configured cards still merge by `key` and retain icon/iconClass control from settings.
- Removed Home card default titles/descriptions from frontend i18n assembly; `HomeView.vue` now keeps only minimal card fallback data for invalid/missing settings.
- Extended backend coverage to assert default Home shell card arrays and key titles are present.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsHomeShellConfig|TestSettingService_GetPublicSettings_DefaultsCreditsShellConfig|TestSettingService_GetPublicSettings_DefaultsPricingShellConfig|TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig|TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Home visual layout, model catalog loading, and card ordering/merge behavior still live in Vue.
- The remaining frontend Home fallback is intentionally small and only protects the page if settings fail to load or the JSON is invalid.

## 2026-06-18 Model Plaza default shell moved to Sub2API settings

### Done
- Added a Sub2API-owned default `model_plaza_shell_config` JSON payload with zh/en labels for the public Models Plaza page.
- Changed public/admin settings reads so empty `model_plaza_shell_config` returns the backend default while still honoring explicit configured values.
- Updated default settings initialization to seed `model_plaza_shell_config` with the backend default JSON for new installations.
- Removed the large i18n-backed Models Plaza default copy from `ModelsPlazaView.vue`; the page now merges parsed public settings with only a minimal frontend resilience fallback.
- Updated Models Plaza test data to explicitly configure the empty-filter label it asserts, matching the new Sub2API-owned shell behavior.
- Added backend coverage proving the default Models Plaza shell config is valid JSON and contains expected zh/en labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsModelPlazaShellConfig|TestSettingService_GetPublicSettings_DefaultsHomeShellConfig|TestSettingService_GetPublicSettings_DefaultsCreditsShellConfig|TestSettingService_GetPublicSettings_DefaultsPricingShellConfig|TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig|TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/ModelsPlazaView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/__tests__/HomeView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Models Plaza filtering, provider grouping, copy-to-clipboard, and item ordering still live in Vue.
- The remaining frontend Models Plaza fallback is intentionally small and only protects the page if settings fail to load or the JSON is invalid.
- Initial Models Plaza frontend test failed because the test asserted an old i18n fallback label while only partially configuring shell labels; fixed by explicitly putting `emptyFilteredTitle` in the mocked `model_plaza_shell_config`.

## 2026-06-18 Docs default shell moved to Sub2API settings

### Done
- Added a Sub2API-owned default `docs_shell_config` JSON payload with zh/en labels for the public Docs page and Docsify search UI.
- Changed public/admin settings reads so empty `docs_shell_config` returns the backend default while still honoring explicit configured values.
- Updated default settings initialization to seed `docs_shell_config` with the backend default JSON for new installations.
- Removed the i18n-backed Docs default copy from `DocsView.vue`; the page now merges parsed public settings with only a minimal frontend resilience fallback.
- Changed the compact dashboard button label to use the shell-provided `copy.dashboard`, keeping the Docs header fully controlled by shell config.
- Added backend coverage proving the default Docs shell config is valid JSON and contains expected zh/en labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsDocsShellConfig|TestSettingService_GetPublicSettings_DefaultsModelPlazaShellConfig|TestSettingService_GetPublicSettings_DefaultsHomeShellConfig|TestSettingService_GetPublicSettings_DefaultsCreditsShellConfig|TestSettingService_GetPublicSettings_DefaultsPricingShellConfig|TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig|TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/__tests__/HomeView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Docsify runtime loading, content path selection, hash rewriting, and route cleanup still live in Vue.
- The remaining frontend Docs fallback is intentionally small and only protects the page if settings fail to load or the JSON is invalid.

## 2026-06-18 Key Usage shell copy no longer reuses Home i18n

### Done
- Removed the remaining `home.*` runtime copy references from `KeyUsageView.vue`.
- Key Usage now reads the Docs link label from `docs_shell_config.labels.title` and the footer copyright suffix from `home_shell_config.labels.allRightsReserved`, with minimal zh/en fallbacks for invalid or missing settings.
- Cleaned stale Home/ModelsPlaza i18n mock labels from Key Usage and Models Plaza tests so tests reflect the settings-first shell behavior.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/KeyUsageView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/KeyUsageView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/KeyUsageView.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/__tests__/HomeView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Key Usage business copy remains under `keyUsage.*` i18n because that page has not yet been modeled as a Sub2API shell config.
- The remaining search hit for `docs.frameworkHint` is a negative assertion in `DocsView.spec.ts`, not runtime usage.

## 2026-06-18 Remaining Touch migration status after public shell pass

### Done
- Public shell defaults now live in Sub2API settings for Home, Models Plaza, Docs, Prompt Catalog, Image Generator workspace, Pricing, and Credits.
- Frontend public pages now keep only minimal resilience fallbacks for those shells, instead of carrying full zh/en default copy blocks.
- Key Usage no longer reuses Home i18n for its shared header/footer labels.

### Still Open
- Touch/current Vue frontend is still an independent frontend shell; it has not been merged into a single Sub2API frontend runtime beyond API/settings consolidation.
- Legal Document page shell still uses `legalDocument.*` i18n for chrome/error labels, while the document content itself already comes from Sub2API `login_agreement_documents`.
- Key Usage business UI still uses `keyUsage.*` i18n and has not been modeled as a Sub2API shell config.
- Auth/payment/user business pages still use normal Vue i18n and local page orchestration; only their data/payment/auth sources have been migrated.
- Runtime page layouts, filters, pagination, Docsify bootstrap, model matrix grouping, payment modal flow, and other UI interactions still live in frontend code.

### Suggested Next Slices
- Add a `legal_document_shell_config` setting only if legal page chrome must be admin-configurable rather than locale-driven.
- Add a `key_usage_shell_config` setting if the public Key Usage page needs to be managed from Sub2API settings.
- Larger remaining work is structural: collapsing this frontend shell further into the Sub2API application/runtime rather than just moving page copy into public settings.

## 2026-06-18 Legal Document shell moved to Sub2API settings

### Done
- Added a Sub2API-owned default `legal_document_shell_config` JSON payload with zh/en labels for login-agreement page chrome, loading failures, missing-document states, update date text, and empty content.
- Exposed `legal_document_shell_config` through public settings, admin settings DTOs, runtime settings admin form, and public settings injection.
- Changed `LegalDocumentView.vue` to merge public settings labels with a minimal frontend resilience fallback instead of using `legalDocument.*` / `auth.signIn` i18n at runtime.
- Added backend coverage for the default legal document shell config and frontend coverage proving shell labels render from public settings.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsLegalDocumentShellConfig|TestSettingService_GetPublicSettings_DefaultsDocsShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/LegalDocumentView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/RuntimeSettingsView.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/LegalDocumentView.spec.ts src/views/__tests__/KeyUsageView.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/__tests__/HomeView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Legal document content and status still come from Sub2API `login_agreement_documents`; this change only moves the surrounding page copy/config into public settings.
- The remaining frontend fallback is intentionally small and only protects the page if settings fail to load or the JSON is invalid.
- Key Usage, auth/payment/user pages, and structural frontend/runtime consolidation remain open migration areas.

## 2026-06-18 Key Usage shell moved to Sub2API settings

### Done
- Added a Sub2API-owned default `key_usage_shell_config` JSON payload with zh/en labels for the public API key usage page.
- Exposed `key_usage_shell_config` through public settings, admin settings DTOs, runtime settings admin form, and public settings injection.
- Changed `KeyUsageView.vue` to render page chrome, date range labels, result headings, table headers, quota labels, query messages, and reset/expiry copy from public settings labels.
- Kept Key Usage query behavior, chart/ring calculations, result shaping, and `/v1/usage` request flow in the frontend; this slice only moved display labels/config.
- Added backend coverage for the default Key Usage shell config and frontend coverage proving configured labels render from public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/KeyUsageView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsKeyUsageShellConfig|TestSettingService_GetPublicSettings_DefaultsLegalDocumentShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/KeyUsageView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/__tests__/HomeView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- `KeyUsageView.vue` no longer has runtime `keyUsage.*` i18n usage.
- The remaining frontend fallback is intentionally small and only protects the page if settings fail to load or the JSON is invalid.
- Key Usage UI interaction, chart animation, API request composition, result normalization, and table rendering still live in Vue.

## 2026-06-18 Auth entry shell moved to Sub2API settings

### Done
- Added a Sub2API-owned default `auth_shell_config` JSON payload with zh/en labels for the public login and registration entry pages.
- Exposed `auth_shell_config` through public settings, admin settings DTOs, runtime settings admin form, and public settings injection.
- Changed `LoginView.vue` to render the login page title, subtitle, form labels/placeholders, forgot-password link, submit button, OAuth divider, and register footer from public settings labels.
- Changed `RegisterView.vue` to render registration title/subtitle, disabled-registration state, form labels/placeholders, password hint, invitation/promo field labels, submit button, OAuth divider, and login footer from public settings labels.
- Kept validation errors, login agreement messages, 2FA, OAuth callbacks, password reset, invitation/promo validation logic, and auth request behavior unchanged.
- Added backend coverage for the default auth shell config and frontend coverage proving Login/Register labels render from public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAuthShellConfig|TestSettingService_GetPublicSettings_DefaultsKeyUsageShellConfig|TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm run frontend:typecheck`
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts src/views/__tests__/KeyUsageView.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts`
- `git diff --check`

### Notes
- Login/Register entry chrome is now settings-first, but auth validation/error copy, callback pages, forgot/reset password pages, TOTP modal text, OAuth provider components, and login agreement prompt still use frontend i18n.
- Auth API calls and session behavior remain unchanged in this slice.

## 2026-06-18 Payment page shell moved to Sub2API settings

### Done
- Added a Sub2API-owned default `payment_shell_config` JSON payload with zh/en labels for the logged-in payment page: recharge/subscription tabs, account summary, empty states, amount summary, submit/cancel buttons, plan quota labels, and active subscription badges.
- Exposed `payment_shell_config` through public settings, admin settings DTOs, runtime settings admin form, public settings injection, and settings update persistence.
- Changed `PaymentView.vue` to render payment page shell labels from `appStore.cachedPublicSettings.payment_shell_config` with locale fallback and `{value}` interpolation.
- Kept checkout catalog loading, payment method selection, order creation, WeChat/Stripe/Airwallex launch decisions, QR/status handling, recovery snapshots, and payment error translations unchanged.
- Updated backend and frontend tests to cover the new payment shell setting and settings-first payment labels.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/dto -run 'TestSettingHandler_GetPublicSettings|TestPublicSettingsInjectionPayload_SchemaDoesNotDrift|^$' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Payment page chrome is settings-first, but the payment flow and provider-specific components still live in Vue.
- Payment validation/error copy intentionally remains on the existing i18n/error-descriptor path to avoid changing checkout behavior.
- Subscription plan card internals and payment method selector labels are still component-local/i18n driven; this slice only moved the parent payment page shell.

## 2026-06-18 Payment child component labels moved under payment shell

### Done
- Extended Sub2API `payment_shell_config` defaults with labels for payment method selector, payment provider names, recharge product card badges/CTA/credited-balance line, subscription plan card quota/model labels, subscribe/renew buttons, and validity suffixes.
- Changed `PaymentView.vue` to pass settings-derived labels into `PaymentMethodSelector`, `RechargeProductCard`, and `SubscriptionPlanCard`.
- Added optional label props to the three payment child components while preserving their existing i18n fallback when used outside `PaymentView`.
- Updated tests to assert `PaymentView` wires settings labels into child components and `SubscriptionPlanCard` renders configured labels from props.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentView.spec.ts src/components/payment/__tests__/SubscriptionPlanCard.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Payment status/result/QR/Stripe/Airwallex views still use frontend i18n because those are provider-flow and result-state surfaces, not the purchase selection shell.
- Component fallback i18n remains intentionally available for admin payment components and standalone use outside the purchase page.

## 2026-06-18 Payment status panel labels moved under payment shell

### Done
- Extended Sub2API `payment_shell_config` defaults with labels for the in-purchase payment status panel: QR scan titles/hints, popup reopen copy, countdown labels, waiting/cancel actions, terminal success/cancelled/expired states, order summary labels, and confirm button.
- Added an optional `labels` prop to `PaymentStatusPanel` while preserving existing i18n fallback for standalone use.
- Changed `PaymentView.vue` to pass settings-derived `paymentStatusPanelLabels` into `PaymentStatusPanel`.
- Added tests proving `PaymentStatusPanel` renders configured labels and `PaymentView` wires the label object into the status panel.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/PaymentStatusPanel.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- `PaymentStatusPanel.vue` now uses configured labels for visible status-panel chrome; remaining `PaymentView` i18n hits in this area are cancellation toast and payment error fallback paths.
- Standalone result/return pages such as `PaymentResultView`, `StripePaymentView`, and `AirwallexPaymentView` still use frontend i18n and remain open for later slices.

## 2026-06-18 Payment result page labels moved under payment shell

### Done
- Extended Sub2API `payment_shell_config` defaults with labels for `PaymentResultView`: terminal success/processing/failed titles, processing hint, order summary fields, payment method names, status label, and back/view-orders buttons.
- Changed `PaymentResultView.vue` to read `appStore.cachedPublicSettings.payment_shell_config` and render result-page chrome from settings with locale fallback.
- Preserved all resume-token lookup, order polling, legacy out_trade_no verification, recovery snapshot cleanup, amount formatting, and unknown payment-method fallback behavior.
- Added a frontend test proving `PaymentResultView` renders configured labels from public payment shell config.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentResultView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- `PaymentResultView.vue` still falls back to `paymentMethodI18nKey` for unknown/unconfigured payment method names so new providers do not render blank labels.
- Stripe and Airwallex dedicated payment pages remain frontend-i18n driven and are still open migration slices.

## 2026-06-18 Stripe and Airwallex carrier labels moved under payment shell

### Done
- Extended Sub2API `payment_shell_config` defaults with labels for the dedicated Stripe and Airwallex payment carrier pages: load failures, missing payment parameters, Stripe submit button, Stripe success-processing hint, and return-to-recharge copy.
- Changed `StripePaymentView.vue` to read `appStore.cachedPublicSettings.payment_shell_config` and render visible carrier-page chrome from settings with locale fallback.
- Changed `AirwallexPaymentView.vue` to read the same payment shell config for load/missing-parameter/error action labels.
- Preserved Stripe Payment Element mounting, Alipay/WeChat confirmation, polling, Airwallex snapshot restore, redirect construction, and provider error handling.
- Added frontend tests proving configured Stripe/Airwallex carrier labels render from public settings.

### Validation
- `pnpm --filter sub2api-frontend exec eslint --fix src/views/user/StripePaymentView.vue src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/AirwallexPaymentView.vue src/views/user/__tests__/AirwallexPaymentView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Provider/runtime error details still surface as provider messages when Stripe or Airwallex returns a concrete error; this slice only moved stable UI chrome and fallback copy into Sub2API settings.
- `StripePopupView.vue` and `StripePaymentInline.vue` still have their own frontend-side payment text and remain open cleanup candidates if they are still active routes/components.

## 2026-06-18 Stripe popup and inline labels moved under payment shell

### Done
- Extended Sub2API `payment_shell_config` defaults with Stripe popup labels for close, redirecting, loading WeChat QR, and timeout states.
- Changed `StripePopupView.vue` to read `appStore.cachedPublicSettings.payment_shell_config` for order ID, close, success/error fallback, redirect/loading/timeout copy, and provider load/missing-parameter fallbacks.
- Added optional `labels` support to `StripePaymentInline.vue` so parent surfaces can pass payment shell labels while standalone use keeps frontend i18n fallback.
- Preserved popup postMessage handshakes, Stripe Alipay/WeChat confirmation, polling, inline Payment Element mounting, popup launch, cancel-order behavior, and provider error details.
- Added frontend tests covering configured labels for both the standalone popup route and the inline component.

### Validation
- `pnpm --filter sub2api-frontend exec eslint --fix src/components/payment/StripePaymentInline.vue src/components/payment/__tests__/StripePaymentInline.spec.ts src/views/user/StripePopupView.vue src/views/user/__tests__/StripePopupView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/StripePaymentInline.spec.ts src/views/user/__tests__/StripePopupView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- `StripePaymentInline.vue` does not appear to be referenced by the current main payment page, but it is now ready for settings-driven labels if reintroduced.
- `PaymentQRCodeView.vue`, `PaymentQRDialog.vue`, and `UserOrdersView.vue` still retain payment-related frontend i18n and are separate cleanup candidates.

## 2026-06-18 QR and user order labels moved under payment shell

### Done
- Extended Sub2API `payment_shell_config` defaults with QR/new-window, order-list, refund, status-filter, and table action labels used by legacy payment QR surfaces and the user orders page.
- Added optional `labels` support to `PaymentQRDialog.vue` and `OrderTable.vue`, preserving existing i18n fallback for admin/standalone callers.
- Changed `PaymentQRCodeView.vue` to read `appStore.cachedPublicSettings.payment_shell_config` for QR page titles, hints, countdown labels, open-window action, expired state, back-to-recharge, processing, and cancel-order copy.
- Changed `UserOrdersView.vue` to read the same payment shell config for status filters, refresh/back actions, cancel/refund buttons, cancel confirmation, refund form labels, and labels passed into `OrderTable`.
- Preserved QR rendering, logo overlay, polling, cancellation, refund eligibility, pagination, refund submission, and payment amount formatting behavior.
- Added frontend tests proving configured labels render for `PaymentQRDialog`, `PaymentQRCodeView`, and `UserOrdersView`.

### Validation
- `pnpm --filter sub2api-frontend exec eslint --fix src/components/payment/PaymentQRDialog.vue src/components/payment/OrderTable.vue src/views/user/PaymentQRCodeView.vue src/views/user/UserOrdersView.vue src/components/payment/__tests__/PaymentQRDialog.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/PaymentQRDialog.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts src/components/payment/__tests__/StripePaymentInline.spec.ts src/views/user/__tests__/StripePopupView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- `PaymentQRCodeView.vue`, `PaymentQRDialog.vue`, `UserOrdersView.vue`, and `OrderTable.vue` no longer directly call frontend i18n for stable payment/order UI copy; their fallback mappings remain for resilience and non-configured deployments.
- Admin payment order surfaces still use frontend i18n and remain separate from the Touch/user-facing shell migration.

## 2026-06-18 User subscription labels moved under payment shell

### Done
- Extended Sub2API `payment_shell_config` defaults with user subscription page labels: empty state, status badges, expiration copy, daily/weekly/monthly usage labels, unlimited state, reset/quota window text, today/tomorrow suffixes, renew action, and load-failure fallback.
- Changed `SubscriptionsView.vue` to read `appStore.cachedPublicSettings.payment_shell_config` for stable subscription page UI copy with locale fallback and `{value}` interpolation.
- Preserved subscription API loading, platform badge styling, renewal navigation, usage progress calculations, expiration color logic, and quota window duration calculations.
- Added frontend tests proving empty state and subscription cards render configured labels from public settings.

### Validation
- `pnpm --filter sub2api-frontend exec eslint --fix src/views/user/SubscriptionsView.vue src/views/user/__tests__/SubscriptionsView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/SubscriptionsView.spec.ts src/components/payment/__tests__/PaymentQRDialog.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- `SubscriptionsView.vue` no longer directly calls frontend i18n for stable subscription display copy; fallback mappings remain.
- User account/workbench pages such as `UsageView`, `KeysView`, `ProfileView`, `ApiGuideView`, and `ApiTestView` still retain local frontend i18n and are later shell/config migration candidates.

## 2026-06-19 User usage labels moved under public settings

### Done
- Added Sub2API `usage_shell_config` as a public/admin settings field with default zh/en JSON for user usage-history stat cards, filters, buttons, and table headers.
- Included `usage_shell_config` in public settings responses, SSR public settings injection, admin settings read/update, frontend public/admin types, and the runtime settings admin page.
- Changed `UsageView.vue` to render stable page chrome from `appStore.cachedPublicSettings.usage_shell_config` with locale-aware fallback to existing frontend i18n.
- Added backend and frontend tests proving the default config is valid JSON, SSR injection schema stays in sync, runtime settings can save the field, and `UsageView` renders configured labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsUsageShellConfig|TestSettingService_GetPublicSettings_DefaultsKeyUsageShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler' -count=1`
- `go test -tags unit ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/UsageView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- `UsageView.vue` still keeps fallback i18n mappings for non-configured deployments and for detailed tooltip/body copy outside the stable shell labels.
- This reduces Touch-side user usage page chrome, but the page interaction/state orchestration remains in the frontend.

## 2026-06-19 API guide labels moved under public settings

### Done
- Added Sub2API `api_guide_shell_config` as a public/admin settings field with default zh/en JSON for the API guide page hero, actions, key selector, key summary, auth hint, endpoint cards, and curl actions.
- Included `api_guide_shell_config` in public settings responses, SSR public settings injection, admin settings read/update, frontend public/admin types, and the runtime settings admin page.
- Changed `ApiGuideView.vue` to render stable page chrome from `appStore.cachedPublicSettings.api_guide_shell_config` with locale-aware fallback to existing frontend i18n.
- Added backend and frontend tests proving the default config is valid JSON, SSR injection schema stays in sync, runtime settings can save the field, and `ApiGuideView` renders configured labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAPIGuideShellConfig|TestSettingService_GetPublicSettings_DefaultsUsageShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler' -count=1`
- `go test -tags unit ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiGuideView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts src/views/user/__tests__/apiGuideDarkContrast.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- `ApiGuideView.vue` still keeps gateway protocol/platform/variant labels in shared frontend i18n because those are reused capability labels, not page-specific shell copy.
- API tester (`ApiTestView.vue`) still retains local frontend i18n and remains a later shell/config migration candidate.

## 2026-06-19 API test labels moved under public settings

### Done
- Added Sub2API `api_test_shell_config` as a public/admin settings field with default zh/en JSON for the API test page hero, controls, model selector, prompt editor, billing warning, request/response panels, copy actions, and usage-sync messages.
- Included `api_test_shell_config` in public settings responses, SSR public settings injection, admin settings read/update, frontend public/admin types, and the runtime settings admin page.
- Changed `ApiTestView.vue` to render stable page chrome and user-facing action/status copy from `appStore.cachedPublicSettings.api_test_shell_config` with locale-aware fallback to existing frontend i18n.
- Added backend and frontend tests proving the default config is valid JSON, SSR injection schema stays in sync, runtime settings can save the field, and `ApiTestView` renders configured labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAPITestShellConfig|TestSettingService_GetPublicSettings_DefaultsAPIGuideShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler' -count=1`
- `go test -tags unit ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiTestView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- `ApiTestView.vue` still keeps gateway protocol/platform/variant labels in shared frontend i18n because those are reused capability labels.
- The page remains a Vue frontend orchestration surface for actual request execution, model loading, and usage-record polling; this slice only moved stable shell/action/status copy to Sub2API settings.

## 2026-06-19 Available groups labels moved under public settings

### Done
- Added Sub2API `available_groups_shell_config` as a public/admin settings field with default zh/en JSON for the available groups page title, description, stat cards, search placeholder, empty states, public/member sections, badges, fields, and quota labels.
- Included `available_groups_shell_config` in public settings responses, SSR public settings injection, admin settings read/update, frontend public/admin types, and the runtime settings admin page.
- Changed `AvailableGroupsView.vue` to render stable page chrome from `appStore.cachedPublicSettings.available_groups_shell_config` with locale-aware fallback to existing frontend i18n.
- Added backend and frontend tests proving the default config is valid JSON, SSR injection schema stays in sync, runtime settings can save the field, and `AvailableGroupsView` renders configured labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAvailableGroupsShellConfig|TestSettingService_GetPublicSettings_DefaultsAPITestShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler' -count=1`
- `go test -tags unit ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AvailableGroupsView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- `AvailableGroupsView.vue` still keeps platform names in shared frontend i18n because those labels are reused across admin/user group surfaces.
- The page still owns frontend filtering and grouping display logic; this slice only moved stable shell and display labels to Sub2API settings.

## 2026-06-19 Available channels labels moved under public settings

### Done
- Added Sub2API `available_channels_shell_config` as a public/admin settings field with default zh/en JSON for the available channels page search, refresh title, empty state, table headers, and public/exclusive group labels.
- Included `available_channels_shell_config` in public settings responses, SSR public settings injection, admin settings read/update, frontend public/admin types, and the runtime settings admin page.
- Changed `AvailableChannelsView.vue` to render stable page/table labels from `appStore.cachedPublicSettings.available_channels_shell_config` with locale-aware fallback to existing frontend i18n.
- Added optional label props to `AvailableChannelsTable.vue` so page-level config can drive public/exclusive labels without changing its default i18n behavior.
- Added backend and frontend tests proving the default config is valid JSON, SSR injection schema stays in sync, runtime settings can save the field, and `AvailableChannelsView` passes configured labels into the table.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAvailableChannelsShellConfig|TestSettingService_GetPublicSettings_DefaultsAvailableGroupsShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler' -count=1`
- `go test -tags unit ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AvailableChannelsView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Pricing detail labels remain in shared frontend i18n because they are part of the reusable supported-model pricing component, not only this page shell.
- The page still owns frontend search/filtering over the already fetched channel payload; this slice only moved stable display labels to Sub2API settings.

## 2026-06-19 Channel status labels moved under public settings

### Done
- Added Sub2API `channel_status_shell_config` as a public/admin settings field with default zh/en JSON for channel status window tabs, overall status labels, refresh title, empty state, card metric labels, detail table headers, close button, and load-error copy.
- Included `channel_status_shell_config` in public settings responses, SSR public settings injection, admin settings read/update, frontend public/admin types, and the runtime settings admin page.
- Changed `ChannelStatusView.vue` to parse `appStore.cachedPublicSettings.channel_status_shell_config` once and pass locale-aware labels into the monitor hero, card grid, and detail dialog.
- Added optional label props to `MonitorHero.vue`, `MonitorCardGrid.vue`, `MonitorCard.vue`, and `MonitorDetailDialog.vue` while preserving their existing i18n fallback behavior.
- Added backend and frontend tests proving the default config is valid JSON, SSR injection schema stays in sync, runtime settings can save the field, and `ChannelStatusView` passes configured labels into all channel-status child surfaces.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsChannelStatusShellConfig|TestSettingService_GetPublicSettings_DefaultsAvailableChannelsShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler' -count=1`
- `go test -tags unit ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ChannelStatusView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- `monitorCommon` status, provider, relative-time, and timeline labels remain in shared frontend i18n because admin and user monitor views share those semantics.
- Channel status still owns frontend polling, auto-refresh state, and detail caching; this slice only moved stable page/display copy to Sub2API settings.

## 2026-06-19 Custom page labels moved under public settings

### Done
- Added Sub2API `custom_page_shell_config` as a public/admin settings field with default zh/en JSON for custom page missing states, invalid URL state, Markdown TOC labels, open-in-new-tab action, Markdown load failures, and code-copy button copy.
- Included `custom_page_shell_config` in public settings responses, SSR public settings injection, admin settings read/update, frontend public/admin types, and the runtime settings admin page.
- Changed `CustomPageView.vue` to read `appStore.cachedPublicSettings.custom_page_shell_config` and render missing/not-configured/open-action/TOC/copy/load-failure copy from Sub2API settings with locale-aware fallback.
- Escaped configured Markdown load-error copy before injecting it into `v-html` error output.
- Added backend and frontend tests proving the default config is valid JSON, SSR injection schema stays in sync, runtime settings can save the field, and `CustomPageView` renders configured missing/invalid/open labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsCustomPageShellConfig|TestSettingService_GetPublicSettings_DefaultsChannelStatusShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler' -count=1`
- `go test -tags unit ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/CustomPageView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Custom page body content, Markdown files, relative image resolution, iframe URL construction, and custom menu source data remain in their existing Sub2API page/custom-menu flows.
- The page still owns frontend Markdown rendering, TOC generation, and iframe embedding behavior; this slice only moved stable display copy to Sub2API settings.

## 2026-06-19 Profile overview labels moved under public settings

### Done
- Added Sub2API `profile_shell_config` as a public/admin settings field with default zh/en JSON for profile overview role labels, metrics, basic-profile section, linked-source section, support card title, provider labels, and profile-source hints.
- Included `profile_shell_config` in public settings responses, SSR public settings injection, admin settings read/update, frontend public/admin types, and the runtime settings admin page.
- Changed `ProfileView.vue` to parse `appStore.cachedPublicSettings.profile_shell_config`, use it for the support card title, and pass labels into `ProfileInfoCard`.
- Changed `ProfileInfoCard.vue` to render overview role/metric/basic/source/provider labels from configured labels with existing i18n fallback.
- Added backend and frontend tests proving the default config is valid JSON, SSR injection schema stays in sync, runtime settings can save the field, and `ProfileView` passes configured labels into the profile overview.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig|TestSettingService_GetPublicSettings_DefaultsCustomPageShellConfig|TestSettingService_GetPublicSettings_WebRuntimeSettings|TestSettingService_UpdateSettings_TrimsWebRuntimeSettings' -count=1`
- `go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings|TestSettingHandler' -count=1`
- `go test -tags unit ./internal/handler/dto -run 'TestPublicSettingsInjectionPayload_SchemaDoesNotDrift' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ProfileView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- Profile password, avatar upload, identity binding, balance-notify, and TOTP form/dialog copy still remain in their existing component i18n and should be migrated as separate, narrower slices.
- Profile account data, OAuth capability flags, and notification settings remain sourced from existing Sub2API user/public-settings APIs; this slice only moved stable overview display copy to Sub2API settings.

## 2026-06-19 Profile password labels moved under public settings

### Done
- Extended Sub2API `profile_shell_config` defaults with zh/en labels for the profile password form title, fields, hint, submit/loading states, validation errors, and success/failure messages.
- Changed `ProfileView.vue` to parse the password-form label keys from `appStore.cachedPublicSettings.profile_shell_config` and pass the shared profile shell labels into `ProfilePasswordForm`.
- Changed `ProfilePasswordForm.vue` to render configured labels with existing i18n fallback and interpolate `{count}` for configured password-policy messages.
- Added backend and frontend tests proving the default config exposes password labels, the profile page passes configured labels into the password form, and the form renders/uses configured validation copy.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfilePasswordForm.spec.ts src/views/user/__tests__/ProfileView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Profile avatar upload, identity binding, balance-notify, and TOTP form/dialog copy still remain in their existing component i18n and should be migrated as separate, narrower slices.
- Profile account data, OAuth capability flags, and notification settings remain sourced from existing Sub2API user/public-settings APIs; this slice only moved password-form display and feedback copy to Sub2API settings.

## 2026-06-19 Profile balance notification labels moved under public settings

### Done
- Extended Sub2API `profile_shell_config` defaults with zh/en labels for the profile balance notification card title, description, toggle, threshold controls, notification email controls, verification states, and toast feedback messages.
- Changed `ProfileView.vue` to parse the balance-notification label keys from `appStore.cachedPublicSettings.profile_shell_config` and pass the shared profile shell labels into `ProfileBalanceNotifyCard`.
- Changed `ProfileBalanceNotifyCard.vue` to render configured labels with existing i18n fallback while keeping the update-profile, email-code, verify, remove, and countdown logic unchanged.
- Added backend and frontend tests proving the default config exposes balance-notification labels, the profile page passes configured labels into the card, and the card renders/uses configured copy.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfileBalanceNotifyCard.spec.ts src/views/user/__tests__/ProfileView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Profile avatar upload, identity binding, and TOTP form/dialog copy still remain in their existing component i18n and should be migrated as separate, narrower slices.
- Profile account data, OAuth capability flags, and notification settings remain sourced from existing Sub2API user/public-settings APIs; this slice only moved balance-notification display and feedback copy to Sub2API settings.

## 2026-06-19 Profile avatar labels moved under public settings

### Done
- Extended Sub2API `profile_shell_config` defaults with zh/en labels for the profile avatar card title, description, upload hint/action, upload validation failures, compression failures, and save/delete feedback.
- Changed `ProfileView.vue` to parse avatar label keys from `appStore.cachedPublicSettings.profile_shell_config`.
- Changed `ProfileInfoCard.vue` to pass the shared profile shell labels into `ProfileAvatarCard`.
- Changed `ProfileAvatarCard.vue` to render configured labels with existing i18n fallback while keeping avatar compression, preview, save, and delete behavior unchanged.
- Added backend and frontend tests proving the default config exposes avatar labels, the profile info card passes configured labels into the avatar card, and the avatar card renders/uses configured copy.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfileAvatarCard.spec.ts src/components/user/profile/__tests__/ProfileInfoCard.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Profile identity binding and TOTP form/dialog copy still remain in their existing component i18n and should be migrated as separate, narrower slices.
- Profile account data, OAuth capability flags, and notification settings remain sourced from existing Sub2API user/public-settings APIs; this slice only moved avatar-card display and feedback copy to Sub2API settings.

## 2026-06-19 Profile TOTP card labels moved under public settings

### Done
- Extended Sub2API `profile_shell_config` defaults with zh/en labels for the profile TOTP card title, description, feature-disabled state, enabled state, enabled-at label, disabled state, and enable/disable actions.
- Changed `ProfileView.vue` to parse TOTP card label keys from `appStore.cachedPublicSettings.profile_shell_config` and pass the shared profile shell labels into `ProfileTotpCard`.
- Changed `ProfileTotpCard.vue` to render configured labels with existing i18n fallback while keeping status loading, setup modal opening, disable dialog opening, and refresh behavior unchanged.
- Added backend and frontend tests proving the default config exposes TOTP card labels, the profile page passes configured labels into the TOTP card, and the card renders configured labels in disabled/enabled/not-enabled states.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfileTotpCard.spec.ts src/views/user/__tests__/ProfileView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Profile identity binding plus the TOTP setup/disable dialogs still remain in their existing component i18n and should be migrated as separate, narrower slices.
- Profile account data, OAuth capability flags, TOTP status, setup, and disable operations remain sourced from existing Sub2API APIs; this slice only moved TOTP card shell/status copy to Sub2API settings.

## 2026-06-19 Profile identity binding labels moved under public settings

### Done
- Extended Sub2API `profile_shell_config` defaults with zh/en labels for the profile identity binding section title, description, binding statuses, bind/unbind actions, email binding form placeholders/actions, bound-count text, note text, and success feedback.
- Changed `ProfileView.vue` to parse identity-binding label keys from `appStore.cachedPublicSettings.profile_shell_config`.
- Changed `ProfileInfoCard.vue` to pass the shared profile shell labels into `ProfileIdentityBindingsSection`.
- Changed `ProfileIdentityBindingsSection.vue` to render configured labels with existing i18n fallback while preserving OAuth bind redirects, email bind/replace, unbind, validation, and cached WeChat capability behavior.
- Added backend and frontend tests proving the default config exposes identity-binding labels, the profile page and info card pass configured labels through, and the identity binding section renders/uses configured copy with interpolation.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts src/components/user/profile/__tests__/ProfileInfoCard.spec.ts src/views/user/__tests__/ProfileView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- TOTP setup/disable dialogs still remain in their existing component i18n and should be migrated as separate, narrower slices.
- Profile account data, OAuth capability flags, binding state, email-code, bind, and unbind operations remain sourced from existing Sub2API APIs; this slice only moved identity-binding display and success-feedback copy to Sub2API settings.

## 2026-06-19 Profile TOTP dialogs labels moved under public settings

### Done
- Extended Sub2API `profile_shell_config` defaults with zh/en labels for TOTP setup and disable dialogs, including setup step descriptions, verification labels/placeholders/actions, manual secret entry, verify action, disable warning/action, and success/failure feedback.
- Changed `ProfileView.vue` to parse TOTP setup/disable dialog label keys from `appStore.cachedPublicSettings.profile_shell_config`.
- Changed `ProfileTotpCard.vue` to pass shared profile shell labels into `TotpSetupModal` and `TotpDisableDialog`.
- Changed `TotpSetupModal.vue` and `TotpDisableDialog.vue` to render configured labels with existing i18n fallback while preserving verification-method loading, email-code cooldown, QR setup, enable/disable requests, toast behavior, and timer cleanup.
- Added backend and frontend tests proving the default config exposes TOTP dialog labels, the card passes labels into both dialogs, and both dialogs render/use configured copy while preserving their existing flows.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/totp-timer-cleanup.spec.ts src/components/user/profile/__tests__/ProfileTotpCard.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- The Profile page’s stable visible copy is now largely under `profile_shell_config`; remaining Profile-local i18n is mostly shared validation/common copy or reusable auth/login strings.
- Profile account data, OAuth capability flags, TOTP setup/disable status, and verification APIs remain sourced from existing Sub2API APIs; this slice only moved TOTP dialog display and feedback copy to Sub2API settings.

## 2026-06-19 Redeem page labels moved under public settings

### Done
- Added Sub2API `redeem_shell_config` as a public/admin runtime setting with zh/en defaults for the redeem page balance card, redeem form, success/failure messages, info card, history labels, toast copy, and history item titles.
- Exposed `redeem_shell_config` through public settings, HTML public-settings injection, admin settings read/write DTOs, frontend public-settings types, admin runtime settings form, and admin i18n labels.
- Changed `RedeemView.vue` to prefer labels from `appStore.cachedPublicSettings.redeem_shell_config` with existing i18n fallback and interpolation support.
- Added `RedeemView` coverage proving configured labels render and are used by the redeem success flow.
- Fixed two service test failures surfaced during wider verification:
  - OAuth email signup/finalize now snapshot default platform quotas like normal signup.
  - Payment provider config validation now includes Airwallex in valid provider keys, sensitive config fields, and pending-order protected config fields; Stripe currency is protected while pending orders exist.

### Validation
- `go test -tags unit ./internal/service -run 'Test(EmailOAuthAuto_SnapshotsPlatformQuotaDefaults|FinalizeOAuthEmailAccount_SnapshotsPlatformQuotaDefaults)' -count=1 -v`
- `go test -tags unit ./internal/service -run 'Test(ValidateProviderRequest|IsSensitiveProviderConfigField|UpdateProviderInstanceRejectsProtectedConfigChangesWhilePendingOrders|UpdateProviderInstanceClearsAirwallexAccountID|AdminService_UpdateUserBalance_InvalidatesAuthCache|SettingService_GetPublicSettings_DefaultsRedeemShellConfig)' -count=1`
- `go test -tags unit ./internal/service ./internal/handler -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/RedeemView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`

### Notes
- This reduces the remaining Touch/Sub2API frontend shell by moving the redeem page’s stable display and feedback copy into Sub2API runtime settings.
- Redeem API behavior, history data, auth user state, and subscription refresh logic remain unchanged.

## 2026-06-19 Affiliate page labels moved under public settings

### Done
- Added Sub2API `affiliate_shell_config` as a public/admin runtime setting with zh/en defaults for the invite center stats, invite code/link blocks, tips, rebate transfer card, invitee table, rebate table, transfer table, and feedback copy.
- Exposed `affiliate_shell_config` through public settings, HTML public-settings injection, admin settings read/write DTOs, frontend public-settings types, admin runtime settings form, and admin i18n labels.
- Changed `AffiliateView.vue` to prefer labels from `appStore.cachedPublicSettings.affiliate_shell_config` with existing i18n fallback and interpolation support.
- Added `AffiliateView` coverage proving configured labels render and are used by copy/transfer success flows.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAffiliateShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AffiliateView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `go test -tags unit ./internal/service ./internal/handler -count=1`
- `git diff --check`

### Notes
- This continues shrinking the remaining Touch/Sub2API frontend shell by moving another stable user-page copy surface into Sub2API runtime settings.
- Affiliate account data, rebate records, transfer records, clipboard behavior, and transfer API behavior remain unchanged.

## 2026-06-19 Usage page labels completed under public settings

### Done
- Extended Sub2API `usage_shell_config` defaults with zh/en labels for the remaining user usage page copy: empty state, token/cost tooltip headings and fields, image billing labels, request type badges, load failures, and CSV export feedback.
- Changed `UsageView.vue` so all user usage page `usage.*` copy now goes through `usageText()` backed by `appStore.cachedPublicSettings.usage_shell_config`, with existing i18n fallback.
- Kept shared `admin.usage.*` labels and utility-level image/billing formatter translations unchanged because they are shared cross-page/tooling copy rather than Usage page shell copy.
- Extended `UsageView` coverage to prove configured labels render in the page shell/empty state and that CSV export feedback uses configured public settings labels.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsUsageShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/UsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `go test -tags unit ./internal/service ./internal/handler -count=1`
- `git diff --check`

### Notes
- `UsageView.vue` no longer has direct `t('usage.*')` references.
- Usage data loading, stats queries, tooltip calculations, CSV row shape, API key filters, sorting, and pagination remain unchanged.

## 2026-06-19 Payment page residual labels moved under public settings

### Done
- Extended Sub2API `payment_shell_config` defaults with zh/en labels for PaymentView’s remaining local page feedback: amount validation, too-many-pending orders, cancel rate limiting, mobile QR fallback, and failed-payment fallback copy.
- Changed `PaymentView.vue` so the remaining `payment.*` page-shell feedback goes through `paymentText()` backed by `appStore.cachedPublicSettings.payment_shell_config`.
- Updated PaymentView tests to prove configured payment shell labels are used for JSAPI cancellation and mobile QR fallback feedback.
- Added a source-level guard in the PaymentView test to prevent reintroducing direct `t('payment.*')` calls in the page.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `go test -tags unit ./internal/service ./internal/handler -count=1`
- `git diff --check`

### Notes
- `PaymentView.vue` no longer has direct `t('payment.*')` references.
- Checkout data loading, order creation, WeChat JSAPI flow, mobile QR fallback behavior, subscription selection, recharge products, payment status state, and payment error extraction behavior remain unchanged.

## 2026-06-19 User order feedback labels moved under public settings

### Done
- Extended Sub2API `payment_shell_config` defaults with zh/en labels for user order action feedback: cancel success, refund request success, and generic order-operation fallback errors.
- Changed `UserOrdersView.vue` so order loading, cancel, and refund fallback messages go through `paymentText()` backed by `appStore.cachedPublicSettings.payment_shell_config`.
- Extended `UserOrdersView` coverage to prove configured labels render and configured cancel/refund success messages are used by the page actions.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/UserOrdersView.spec.ts`
- `pnpm run frontend:typecheck`
- `go test -tags unit ./internal/service ./internal/handler -count=1`
- `git diff --check`

### Notes
- `UserOrdersView.vue` no longer has direct `t('payment.*')` or `t('common.*')` page-feedback references.
- Order query, pagination, cancellation, refund eligibility, refund request payloads, and payment API calls remain unchanged.

## 2026-06-19 Payment flow error fallbacks moved under public settings

### Done
- Extended the remaining user-facing payment flow components to use the configured `payment_shell_config.errorFallback` label for cancel/init failure fallback messages instead of directly falling back to local `common.error`.
- Updated `PaymentQRDialog.vue`, `StripePaymentInline.vue`, `PaymentStatusPanel.vue`, `PaymentQRCodeView.vue`, and `PaymentView.vue` without changing payment polling, cancellation, redirect, QR rendering, or Stripe behavior.
- Added regression coverage for QR dialog, Stripe inline, payment status panel, and standalone QR page cancellation failures using configured public-settings fallback labels.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/PaymentQRDialog.spec.ts src/components/payment/__tests__/StripePaymentInline.spec.ts src/components/payment/__tests__/PaymentStatusPanel.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check`
- `rg -n "t\\('common\\.error'\\)" frontend/src/components/payment/PaymentQRDialog.vue frontend/src/components/payment/StripePaymentInline.vue frontend/src/components/payment/PaymentStatusPanel.vue frontend/src/views/user/PaymentQRCodeView.vue frontend/src/views/user/PaymentView.vue`

### Notes
- This continues shrinking the local Touch/Vue shell by moving another stable feedback-copy path under Sub2API runtime settings.
- Shared/admin payment provider components still use generic local admin/common labels and were intentionally left out of this user-facing payment shell slice.

## 2026-06-19 User dashboard shell moved under public settings

### Done
- Added `dashboard_shell_config` as a Sub2API runtime/public setting with zh/en defaults for signed-in dashboard stat cards, charts, recent usage, platform quota labels, and quick actions.
- Exposed the new setting through public settings, admin settings DTOs, admin update persistence, frontend public/admin types, and the Runtime Settings admin form.
- Changed `DashboardView.vue` to parse `appStore.cachedPublicSettings.dashboard_shell_config` once and pass labels to `UserDashboardStats`, `UserDashboardCharts`, `UserDashboardRecentUsage`, and `UserDashboardQuickActions`.
- Updated the Dashboard child components to render their stable page chrome from provided dashboard shell labels, with existing i18n fallbacks preserved.
- Added backend default coverage and extended the Dashboard stats component test to prove configured labels render.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsDashboardShellConfig' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts`
- `pnpm run frontend:typecheck`
- `go test -tags unit ./internal/service ./internal/handler -count=1`
- `git diff --check`

### Notes
- Dashboard data loading, chart API calls, recent usage queries, platform quota calculations, and quick-action navigation behavior remain unchanged.
- This moves another user-facing Touch/Vue shell surface into Sub2API runtime settings; larger structural work remains for the API Keys page and for fully collapsing the frontend runtime shape.

## 2026-06-19 Profile balance notification feedback completed under public settings

### Done
- Extended `profile_shell_config` defaults with the remaining balance-notification action/feedback labels: saving, save, cancel, add, saved, and generic error fallback.
- Changed `ProfileBalanceNotifyCard.vue` so its save/add/cancel buttons and save/error feedback use the provided `profile_shell_config` labels instead of direct local `common.*` calls.
- Extended `ProfileBalanceNotifyCard` coverage to prove configured action labels, success feedback, duplicate-email feedback, and generic error fallback are used.
- Extended the backend profile shell default test to assert the new balance-notification labels are present.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfileBalanceNotifyCard.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `go test -tags unit ./internal/service ./internal/handler -count=1`
- `git diff --check`

### Notes
- Balance notification API calls, email verification timers, threshold update behavior, and auth-store synchronization remain unchanged.
- Other profile subcomponents still have small local fallback labels and can be handled in later narrow slices.

## 2026-06-19 Profile avatar actions moved under public settings

### Done
- Extended `profile_shell_config` defaults with avatar action and generic error labels: save, delete, and operation failure.
- Changed `ProfileAvatarCard.vue` so save/delete buttons and avatar update error fallbacks use configured `profile_shell_config` labels instead of direct local `common.*` fallback text.
- Extended `ProfileAvatarCard` coverage to prove configured avatar shell labels, action labels, success feedback, and error fallback are used.
- Extended the backend profile shell default test to assert the new avatar action/error labels are present in zh/en defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfileAvatarCard.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/components/user/profile/ProfileAvatarCard.vue frontend/src/components/user/profile/__tests__/ProfileAvatarCard.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Avatar upload normalization, compression, profile update calls, delete behavior, and auth-store synchronization remain unchanged.
- Remaining profile shell work is now concentrated in TOTP setup/disable, identity binding validation feedback, edit form labels, and small status labels.

## 2026-06-19 TOTP disable dialog actions moved under public settings

### Done
- Extended `profile_shell_config` defaults with TOTP disable dialog action/status/error labels: sending, cancel, processing, and generic operation failure.
- Changed `TotpDisableDialog.vue` so current-password label, send-code pending text, cancel button, processing button text, and verification-method load error fallback use `profile_shell_config` labels instead of direct local `common.*` / `profile.*` fallbacks.
- Extended TOTP dialog coverage to prove configured disable labels render, the processing state uses configured copy while the disable request is pending, success feedback remains configured, and the generic load-error fallback uses configured copy.
- Extended the backend profile shell default test to assert the new TOTP disable labels are present in zh/en defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/totp-timer-cleanup.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- TOTP verification-method loading, send-code cooldown cleanup, disable request payloads, success event emission, and API behavior remain unchanged.
- TOTP setup modal still has similar residual local action/fallback labels and is a good next narrow slice.

## 2026-06-19 TOTP setup modal actions moved under public settings

### Done
- Extended `profile_shell_config` defaults with TOTP setup action/status/clipboard labels: next, back, loading, verifying, copied, and copy failed.
- Changed `TotpSetupModal.vue` so current-password label, send-code pending text, cancel/next/back/loading/verifying buttons, clipboard feedback, and verification-method load error fallback use configured `profile_shell_config` labels instead of direct local `common.*` / `profile.*` fallbacks.
- Extended TOTP modal coverage to prove configured setup labels render through identity verification and QR/code steps, clipboard success feedback uses configured copy, and the generic load-error fallback uses configured copy.
- Extended the backend profile shell default test to assert the new setup labels are present in zh/en defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/totp-timer-cleanup.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- TOTP verification-method loading, setup initiation, QR generation, clipboard write target, enable request payloads, cooldown cleanup, and success event emission remain unchanged.
- Profile shell residuals are now mostly in identity binding validation feedback, profile edit form labels, and small status labels.

## 2026-06-19 Profile identity binding feedback moved under public settings

### Done
- Extended `profile_shell_config` defaults with identity-binding loading, retry, email validation, password validation, and send-code failure labels.
- Changed `ProfileIdentityBindingsSection.vue` so email send/bind/unbind loading states, email validation errors, password validation errors, send-code fallback errors, bind fallback errors, and unbind fallback errors use configured `profile_shell_config` labels instead of direct local `common.*` / `auth.*` fallbacks.
- Extended identity binding component coverage to prove configured validation feedback, send-code fallback feedback, and third-party unbind retry fallback are used.
- Extended the backend profile shell default test to assert the new identity-binding feedback labels are present in zh/en defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- OAuth binding starts, WeChat capability resolution, email code sending, email binding payloads, unbind payloads, local user updates, and auth-store synchronization remain unchanged.
- Remaining profile shell residuals are now mainly `ProfileEditForm` and small status/provider fallback labels in profile info.

## 2026-06-19 Profile edit form moved under public settings

### Done
- Extended `profile_shell_config` defaults with profile edit form labels and feedback: title, username label, placeholder, updating state, submit action, username-required validation, success, and failure fallback.
- Changed `ProfileEditForm.vue` so its title, username label, placeholder, submit/loading text, validation feedback, success feedback, and update failure fallback use configured `profile_shell_config` labels instead of direct local `profile.*` calls.
- Passed profile shell labels from `ProfileInfoCard` into `ProfileEditForm`, and added the new label keys to `ProfileView` parsing so public settings can reach the embedded form.
- Added focused `ProfileEditForm` coverage and extended `ProfileInfoCard` coverage to prove configured labels are rendered and propagated.
- Extended the backend profile shell default test to assert the new profile edit labels are present in zh/en defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfileEditForm.spec.ts src/components/user/profile/__tests__/ProfileInfoCard.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- Username update payloads, auth-store synchronization, server error detail precedence, and embedded/non-embedded rendering behavior remain unchanged.
- Remaining profile shell residuals are now mainly small status/provider fallback labels in `ProfileInfoCard`.

## 2026-06-19 Profile overview status labels moved under public settings

### Done
- Extended `profile_shell_config` defaults with profile overview active/disabled status labels.
- Changed `ProfileInfoCard.vue` so the overview status badge uses configured `profile_shell_config` labels instead of direct local `common.active` / `common.disabled` calls.
- Added ProfileInfoCard coverage for configured active/disabled labels and configured GitHub/Google provider labels in profile source hints.
- Extended the backend profile shell default test to assert the new status labels and GitHub/Google provider defaults are present.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfileInfoCard.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- Profile source resolution, role badges, currency/date formatting, layout, and provider source hint interpolation remain unchanged.
- The remaining profile shell residue is down to provider fallback calls used only when configured provider labels are absent.

## 2026-06-19 API Keys page shell config added to public settings

### Done
- Added `api_keys_shell_config` as a Sub2API public/admin runtime setting, including service defaults, DTO mappings, admin update handling, and runtime settings form support.
- Added default zh/en API Keys page labels for the migrated slice: search, refresh, create key, table columns, filters, status labels, group labels, copy text, common save/delete/group/status feedback, and reset-soon text.
- Changed `KeysView.vue` so the API Keys list shell reads configured labels from `api_keys_shell_config` with existing i18n as fallback.
- Added backend coverage proving the public settings API returns valid default API Keys shell JSON.

### Validation
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_Defaults(APIKeys|KeyUsage|Profile)ShellConfig' -count=1`
- `go test -tags unit ./internal/handler/... ./internal/service -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- backend/internal/service/domain_constants.go backend/internal/service/settings_view.go backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go backend/internal/handler/dto/settings.go backend/internal/handler/setting_handler.go backend/internal/handler/admin/setting_handler.go frontend/src/views/user/KeysView.vue frontend/src/views/admin/RuntimeSettingsView.vue frontend/src/api/admin/settings.ts frontend/src/types/index.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts`

### Notes
- API Key CRUD payloads, pagination, sorting, filtering request parameters, usage stat loading, group assignment behavior, and server error precedence remain unchanged.
- `pnpm exec prettier --write ...` is unavailable in this workspace (`prettier` command not found); typecheck and diff whitespace validation passed instead.
- Remaining API Keys shell residue includes deeper create/edit modal labels, custom-key validation labels, quota/rate-limit reset feedback, CCS import labels, and empty-state copy.

## 2026-06-19 API Keys deep shell labels moved under public settings

### Done
- Extended `api_keys_shell_config` defaults with the remaining API Keys page shell copy: action buttons, empty state, create/edit modal labels, custom-key validation, IP restriction fields, quota/rate-limit controls, expiration fields, confirmation dialogs, reset success/failure feedback, CCS import dialog labels, and CCS protocol-handler fallback errors.
- Changed `KeysView.vue` so API Keys page copy now goes through `apiKeysText(...)` instead of direct `keys.*` / `common.*` i18n reads.
- Added `apiKeysStatusText(...)` for configured status badge copy and kept existing i18n keys only as fallback when a public setting omits a label.
- Extended the backend API Keys shell default test to assert deeper labels such as delete, reset rate limit, and CCS client selection are present.

### Validation
- `rg -n "t\\('keys\\.|t\\('common\\." frontend/src/views/user/KeysView.vue || true`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAPIKeysShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go frontend/src/views/user/KeysView.vue`

### Notes
- API Key CRUD payloads, delete/reset payloads, CCS deeplink construction, pagination, sorting, filtering, usage stat loading, and server-detail error precedence remain unchanged.
- `UseKeyModal` is still a separate component with its own UI copy and can be moved under a dedicated or shared shell config in a later slice.

## 2026-06-19 UseKeyModal copy moved under API Keys shell config

### Done
- Passed parsed `api_keys_shell_config` labels from `KeysView.vue` into `UseKeyModal`.
- Changed `UseKeyModal.vue` so modal title, no-group warning, copy button labels, close button, client tab labels, platform descriptions, platform notes, Gemini model comment, OpenAI config hint, and OpenCode hint read configured shell labels with existing i18n as fallback.
- Extended `api_keys_shell_config` defaults with UseKeyModal labels via `apiKeysShellConfigDefault()` so both public settings responses and seeded defaults include the modal copy.
- Added component coverage proving configured modal chrome/warning labels render.
- Extended the backend API Keys shell default test to assert UseKeyModal defaults are present.

### Validation
- `rg -n "t\\('keys\\.useKeyModal|t\\('common\\." frontend/src/components/keys/UseKeyModal.vue || true`
- `pnpm --filter sub2api-frontend exec vitest run src/components/keys/__tests__/UseKeyModal.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAPIKeysShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go frontend/src/components/keys/UseKeyModal.vue frontend/src/components/keys/__tests__/UseKeyModal.spec.ts frontend/src/views/user/KeysView.vue`

### Notes
- Generated terminal/config file contents, default tab selection, platform-specific branching, copy behavior, and OpenCode/Codex/Gemini config models remain unchanged.
- API Keys page and `UseKeyModal` no longer directly read `keys.*` / `common.*` UI copy for their visible shell; remaining work should move on to another frontend shell area instead of this page.

## 2026-06-19 Usage tooltip detail labels moved under public settings

### Done
- Extended `usage_shell_config` defaults with token/cost tooltip detail labels: input/output tokens, cache creation 5m/1h/aggregate tokens, cache read tokens, input/output cost, cache creation cost, and cache read cost.
- Changed `UsageView.vue` so the user usage token/cost tooltip labels use `usageText(...)` instead of direct `admin.usage.*` i18n reads.
- Added default-setting test assertions for the new usage tooltip labels.
- Extended `UsageView.spec.ts` to prove configured tooltip labels render in both token and cost tooltips.

### Validation
- `rg -n "t\\('admin\\.usage\\." frontend/src/views/user/UsageView.vue || true`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/UsageView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsUsageShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go frontend/src/views/user/UsageView.vue frontend/src/views/user/__tests__/UsageView.spec.ts`

### Notes
- Usage query parameters, export behavior, tooltip positioning, image billing labels, pricing calculations, request type labels, and billing mode labels remain unchanged.
- The only verification warning was Browserslist/caniuse-lite being six months old during Vitest; tests passed.

## 2026-06-19 Channel Status labels centralized under public settings

### Done
- Added `frontend/src/utils/channelStatusShell.ts` to resolve channel status page labels from `channel_status_shell_config`, including locale selection, nested label merging, and a final zh/en bootstrap fallback matching the Sub2API default payload.
- Changed `ChannelStatusView.vue` to pass one complete labels object into the monitor hero, card grid, cards, and detail dialog.
- Removed component-level `channelStatus.*` / `monitorCommon.*` i18n fallback reads from the channel status user shell; visible page copy now comes through the resolved shell labels.
- Kept the existing Sub2API backend default setting test as the source contract for `channel_status_shell_config`.

### Validation
- `rg -n "channelStatus\\.|monitorCommon\\.availabilityPrefix|monitorCommon\\.extraModelsCount|labels\\?\\." frontend/src/views/user/ChannelStatusView.vue frontend/src/components/user/monitor/MonitorHero.vue frontend/src/components/user/monitor/MonitorCardGrid.vue frontend/src/components/user/monitor/MonitorCard.vue frontend/src/components/user/MonitorDetailDialog.vue`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ChannelStatusView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsChannelStatusShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/channelStatusShell.ts frontend/src/views/user/ChannelStatusView.vue frontend/src/components/user/monitor/MonitorHero.vue frontend/src/components/user/monitor/MonitorCardGrid.vue frontend/src/components/user/monitor/MonitorCard.vue frontend/src/components/user/MonitorDetailDialog.vue`

### Notes
- Channel monitor API calls, auto-refresh behavior, detail cache behavior, status/provider formatting, latency/availability formatting, and card layout remain unchanged.
- `MonitorDetailDialog` still uses shared `common.loading` for the transient loading label; the channel-status-specific shell copy is now centralized.
- The only verification warning was Browserslist/caniuse-lite being six months old during Vitest; tests passed.

## 2026-06-19 Dashboard shell fallback moved out of component i18n

### Done
- Extended `dashboardShellLabels.ts` with complete zh/en default dashboard labels matching the Sub2API `dashboard_shell_config` default payload.
- Changed `parseDashboardShellLabels(...)` to return a complete label map by merging configured public settings over the locale default labels.
- Updated `UserDashboardStats`, `UserDashboardCharts`, `UserDashboardRecentUsage`, and `UserDashboardQuickActions` so visible dashboard copy reads from passed shell labels or static shell defaults instead of local `vue-i18n` fallback calls.
- Updated the dashboard layout guard test to assert `dashboardLabels` is passed into quick actions.

### Validation
- `rg -n "useI18n|\\bt\\(|\\bte\\(|dashboardShellFallbackKeys" frontend/src/components/user/dashboard/*.vue frontend/src/components/user/dashboard/dashboardShellLabels.ts frontend/src/views/user/DashboardView.vue`
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsDashboardShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/components/user/dashboard/dashboardShellLabels.ts frontend/src/components/user/dashboard/UserDashboardStats.vue frontend/src/components/user/dashboard/UserDashboardCharts.vue frontend/src/components/user/dashboard/UserDashboardRecentUsage.vue frontend/src/components/user/dashboard/UserDashboardQuickActions.vue frontend/src/views/user/__tests__/dashboardNoHero.spec.ts`

### Notes
- Dashboard API calls, chart loading, recent usage slicing, platform quota rendering, date range state, and quick-action routes remain unchanged.
- `DashboardView.vue` still uses `useI18n().locale` only to choose the locale branch from `dashboard_shell_config`.

## 2026-06-19 Available Channels shell labels completed

### Done
- Added `loadError` and `publicTooltip` to the Sub2API default `available_channels_shell_config` payload.
- Changed `AvailableChannelsView.vue` so `available_channels_shell_config` resolves into a complete locale label map with zh/en bootstrap defaults.
- Removed the view-level local i18n fallback map for available-channel labels; visible copy and load-error fallback now come from the resolved shell labels.
- Changed `AvailableChannelsTable.vue` so exclusive/public labels and tooltips are required props instead of component-local `availableChannels.*` i18n fallbacks.
- Extended the backend public settings test to prove `loadError` and `publicTooltip` are present in default zh/en settings.

### Validation
- `rg -n "availableChannels\\.|useI18n|\\bt\\(" frontend/src/views/user/AvailableChannelsView.vue frontend/src/components/channels/AvailableChannelsTable.vue`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AvailableChannelsView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAvailableChannelsShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go frontend/src/views/user/AvailableChannelsView.vue frontend/src/components/channels/AvailableChannelsTable.vue`

### Notes
- Available channel API calls, user group rate loading, search filtering behavior, pricing key prefix behavior, group/model rendering, and table layout remain unchanged.
- `AvailableChannelsView.vue` still uses `useI18n().locale` only to select the locale branch from public settings.

## 2026-06-19 Legal document shell fallback localized

### Done
- Changed `LegalDocumentView.vue` so `legal_document_shell_config` resolves to a complete zh/en copy map instead of merging configured labels over an English-only fallback.
- Added locale-specific default copy for login, agreement label, load failure, missing document, updated-at template, and empty-content states.
- Added coverage proving invalid/missing legal shell JSON falls back to Chinese labels when the runtime locale is `zh-CN`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/LegalDocumentView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsLegalDocumentShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/__tests__/LegalDocumentView.spec.ts`

### Notes
- Legal document loading, agreement document selection, markdown rendering, DOMPurify sanitization, URL sanitization, icon selection, and template rendering remain unchanged.
- `LegalDocumentView.vue` still uses `useI18n().locale` only to choose the locale branch from `legal_document_shell_config`.

## 2026-06-19 Models Plaza shell fallback localized

### Done
- Changed `ModelsPlazaView.vue` so `model_plaza_shell_config` resolves to a complete zh/en copy map instead of merging configured labels over an English-only fallback.
- Added locale-specific default copy for docs/login/dashboard links, hero copy, empty states, search/filter shell, model price labels, copy-model action, and group labels.
- Added coverage proving invalid model plaza shell JSON falls back to Chinese labels when the runtime locale is `zh-CN`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/ModelsPlazaView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsModelPlazaShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/__tests__/ModelsPlazaView.spec.ts`

### Notes
- Model plaza item filtering, visibility handling, provider grouping, sort order, copy-to-clipboard behavior, auth-state link routing, and public settings fetch behavior remain unchanged.
- `ModelsPlazaView.vue` still uses `useI18n().locale` only to choose the locale branch from `model_plaza_shell_config`.

## 2026-06-19 Docs shell fallback localized

### Done
- Changed `DocsView.vue` so `docs_shell_config` resolves to a complete zh/en copy map instead of merging configured labels over an English-only fallback.
- Added locale-specific default copy for docs title, dashboard/login actions, Docsify search placeholder, and no-result state.
- Added coverage proving the docs page keeps locale-specific bootstrap fallback copy and no longer contains the old `FALLBACK_DOCS_COPY` path.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsDocsShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- Docsify runtime loading, route/hash sync, cache-busting, sidebar link rewriting, resource cleanup, and locale-specific docs content paths remain unchanged.
- `DocsView.vue` still uses `useI18n().locale` only to choose the locale branch from `docs_shell_config` and docs content base path.

## 2026-06-19 Pricing shell fallback localized

### Done
- Changed `PricingView.vue` so `pricing_shell_config` resolves to a complete zh/en labels map instead of merging configured labels over an English-only fallback.
- Added locale-specific default copy aligned with the Sub2API backend default pricing shell config for tabs, CTAs, catalog stats, empty states, plan labels, quota/rate labels, and validity units.
- Added coverage proving invalid pricing shell JSON falls back to Chinese copy when the runtime locale is `zh`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPricingShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- Pricing catalog fetching, product/plan sorting, purchase links, currency formatting, group labels, button override behavior, and public settings fetch behavior remain unchanged.
- `PricingView.vue` still uses `getLocale()` only to choose the locale branch from `pricing_shell_config`.

## 2026-06-19 Workspace shell fallback localized

### Done
- Changed `ImageGeneratorView.vue` so `workspace_shell_config` resolves to a complete zh/en copy map instead of merging configured values over an English-only fallback.
- Added locale-specific default copy aligned with the Sub2API backend default workspace shell config for catalog navigation, hero copy, draft import notice, prompt controls, copy messages, and workspace status.
- Added coverage proving invalid workspace shell JSON falls back to Chinese copy when the runtime locale is `zh`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- Prompt draft loading, draft cleanup, prompt length guard, clipboard copy behavior, error toast behavior, public settings fetch behavior, and catalog navigation remain unchanged.
- `ImageGeneratorView.vue` still uses `getLocale()` only to choose the locale branch from `workspace_shell_config`.

## 2026-06-19 Credits shell fallback localized

### Done
- Changed `CreditsView.vue` so `credits_shell_config` resolves to a complete zh/en labels map instead of merging configured values over an English-only fallback.
- Moved the credits page eyebrow into `credits_shell_config.labels.eyebrow` and added it to the Sub2API backend default public settings payload.
- Added locale-specific default copy aligned with the Sub2API backend default credits shell config for balance labels, conversion copy, action block copy, and buttons.
- Added coverage proving invalid credits shell JSON falls back to Chinese copy when the runtime locale is `zh`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/CreditsView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsCreditsShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- Balance refresh, credits conversion math, public settings fetch behavior, purchase/order routes, and Sub2API balance display remain unchanged.
- `CreditsView.vue` still uses `getLocale()` only to choose the locale branch from `credits_shell_config`.

## 2026-06-19 Profile shell fetch result used directly

### Done
- Changed `ProfileView.vue` so the `profile_shell_config` returned by `fetchPublicSettings()` is cached locally and used immediately when `cachedPublicSettings` does not already contain profile shell labels.
- Added coverage proving the profile page and child card labels consume `profile_shell_config` from the Sub2API public settings fetch response even when the app store cache starts empty.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ProfileView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- User refresh, contact support display, OAuth capability flags, balance notification feature flags, TOTP visibility, and child component props remain unchanged.
- `ProfileView.vue` still has an i18n fallback for label keys if both cached settings and fetched Sub2API settings omit `profile_shell_config`; backend defaults already provide the broad profile shell config on the public settings path.

## 2026-06-19 Profile shell coverage guarded

### Done
- Verified the final backend default `profile_shell_config` produced by `profileShellConfigDefault()` covers all 124 `ProfileView` fallback label keys for both zh and en.
- Strengthened `TestSettingService_GetPublicSettings_DefaultsProfileShellConfig` so default profile shell config must include at least the full label-key count for zh/en.
- Added zh provider assertions alongside the existing en provider assertions so provider labels remain in the Sub2API public settings payload.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ProfileView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsProfileShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- No user-facing Profile behavior changed in this slice; this guards the Sub2API default profile shell coverage used by the previous fetch-result integration.

## 2026-06-19 Auth shell parser shared

### Done
- Added `frontend/src/utils/authShell.ts` as the single shared resolver for `auth_shell_config` labels, locale defaults, and template interpolation.
- Changed `LoginView.vue` and `RegisterView.vue` to use the shared auth shell resolver instead of maintaining duplicate fallback labels and duplicate JSON parsing logic.
- Added direct unit coverage for the shared resolver: zh fallback, configured overrides, invalid JSON fallback, and template parameter replacement.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAuthShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- Login/register public settings loading, OAuth visibility, Turnstile behavior, login agreement behavior, validation errors, promo/invitation logic, and submit flows remain unchanged.
- Auth pages still use Sub2API `auth_shell_config` from the public settings response; local auth label defaults now live in one utility instead of two page components.

## 2026-06-19 Payment shell parser shared for result and orders

### Done
- Added `frontend/src/utils/paymentShell.ts` as a shared parser for `payment_shell_config` locale selection and allowed-key filtering.
- Added `interpolatePaymentShellLabel()` for shared payment label template interpolation.
- Migrated `PaymentResultView.vue` from its local `parsePaymentResultLabels()` to the shared payment shell resolver.
- Migrated `UserOrdersView.vue` from its local `parseUserOrdersLabels()` to the shared payment shell resolver.
- Added direct unit coverage for payment shell locale resolution, unknown-key filtering, English fallback, invalid JSON fallback, and interpolation behavior.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- Payment result verification, recovery-token handling, order polling, order table rendering, cancellation, refund request flow, and payment method display behavior remain unchanged.
- Remaining local payment shell parsers are still present in the main purchase page, QR payment page, Stripe page, Airwallex page, and subscription-related pages.

## 2026-06-19 Payment shell parser shared for QR Stripe Airwallex

### Done
- Migrated `PaymentQRCodeView.vue`, `StripePaymentView.vue`, and `AirwallexPaymentView.vue` to the shared `resolvePaymentShellLabels()` parser for `payment_shell_config`.
- Removed the local payment-shell JSON parser copies from those payment-flow pages.
- Kept payment flow fallback behavior by mapping the same i18n keys through page-local allowed label keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentQRCodeView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- QR countdown/cancel/polling, Stripe amount display/confirmation, and Airwallex recovery snapshot/redirect behavior remain unchanged.
- The main purchase page still has a local `payment_shell_config` parser and interpolation helper; that is the next payment shell cleanup target.

## 2026-06-19 Payment shell parser shared for purchase popup subscriptions

### Done
- Migrated `PaymentView.vue`, `StripePopupView.vue`, and `SubscriptionsView.vue` to the shared `resolvePaymentShellLabels()` parser.
- Reused `interpolatePaymentShellLabel()` for purchase-page and subscription template labels such as recharge previews, group fallback, and remaining-days copy.
- Removed the remaining local page-level parser copies for `payment_shell_config` in `frontend/src/views/user`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/views/user/__tests__/StripePopupView.spec.ts src/views/user/__tests__/SubscriptionsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "payment_shell_config|parse[A-Za-z]*(Labels|Config)|resolvePaymentShellLabels|interpolatePayment" frontend/src/views/user frontend/src/components/payment frontend/src/utils/paymentShell.ts -g '*.vue' -g '*.ts'`

### Notes
- Purchase selection, WeChat JSAPI resume, popup payment initialization, subscription status/usage display, and existing i18n fallbacks remain unchanged.
- `payment_shell_config` is now parsed through the shared utility everywhere it is used by `frontend/src/views/user`; other non-payment shell domains still have their own local parsers.

## 2026-06-19 Prompt catalog shell parser extracted

### Done
- Added `frontend/src/utils/promptCatalogShell.ts` for `prompt_catalog_shell_config` locale resolution, allowed label filtering, and JSON error fallback.
- Moved Prompt Catalog copy types and parser logic out of `PromptCatalogView.vue`.
- Kept Prompt Catalog UI state, filtering, import form, and Sub2API data fetching in the page while removing page-local config parsing.
- Added direct unit coverage for the new parser.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/promptCatalogShell.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Prompt Gallery source/category/stat rendering still happens in the page from Sub2API API response fields.
- This reduces local shell parsing but does not yet move Prompt Gallery UI composition into the Sub2API Vue admin/public shell.

## 2026-06-19 Public localized shell parser shared

### Done
- Added `frontend/src/utils/localizedShell.ts` as a shared resolver for locale-scoped `{ labels: ... }` public settings payloads.
- Added unit coverage for locale branch selection, fallback merging, root-label handling, unknown-key filtering, and invalid JSON fallback.
- Migrated `DocsView.vue` from local `parseDocsShellConfig()` to `resolveLocalizedShellLabels()`.
- Migrated `LegalDocumentView.vue` from local `parseLegalDocumentShellConfig()` to `resolveLocalizedShellLabels()`.
- Migrated `ModelsPlazaView.vue` from local `parseModelsPlazaShellConfig()` to `resolveLocalizedShellLabels()`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/localizedShell.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/localizedShell.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/localizedShell.ts frontend/src/utils/__tests__/localizedShell.spec.ts frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/__tests__/LegalDocumentView.spec.ts frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/__tests__/ModelsPlazaView.spec.ts`

### Notes
- Docsify lifecycle/resource handling, legal document markdown rendering/sanitization, and Models Plaza filtering/copy behavior remain unchanged.
- Remaining local public parser targets include `HomeView.vue`, `PricingView.vue`, and `ImageGeneratorView.vue`; `PricingView.vue` and `ImageGeneratorView.vue` have richer non-label config shapes, so they need more targeted extraction than this generic labels-only resolver.

## 2026-06-19 Public shell parsers extracted

### Done
- Added `frontend/src/utils/pricingShell.ts` for `pricing_shell_config` labels, button title, and group tab overrides.
- Added `frontend/src/utils/imageWorkspaceShell.ts` for root-field `workspace_shell_config` parsing used by the image prompt workspace.
- Added `frontend/src/utils/homeShell.ts` for `home_shell_config` labels plus homepage experience/why-choose card overrides and merge behavior.
- Migrated `PricingView.vue`, `ImageGeneratorView.vue`, and `HomeView.vue` to consume the new utilities instead of page-local parsers.
- Removed the remaining local shell parser functions from public/Home views.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/pricingShell.spec.ts src/views/public/__tests__/PricingView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/pricingShell.spec.ts src/views/public/__tests__/PricingView.spec.ts src/utils/__tests__/imageWorkspaceShell.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/homeShell.spec.ts src/views/__tests__/HomeView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/pricingShell.spec.ts src/views/public/__tests__/PricingView.spec.ts src/utils/__tests__/imageWorkspaceShell.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts src/utils/__tests__/homeShell.spec.ts src/views/__tests__/HomeView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "function parse[A-Za-z]*(Shell|Labels|Config)|parse[A-Za-z]*(Shell|Labels|Config)\\(" frontend/src/views/public frontend/src/views/HomeView.vue -g '*.vue'`

### Notes
- Public/Home views no longer contain local shell parser functions.
- Homepage card override merge semantics intentionally remain unchanged: if an override key does not match the current default card, the original behavior falls back to the override at the same array index.
- Remaining parser work is now concentrated in user/account pages such as usage, credits, available groups/channels, redeem, API guide/test, custom page, affiliate, and profile.

## 2026-06-19 User shell parsers extracted

### Done
- Added `frontend/src/utils/shellLabelOverrides.ts` for standard locale-scoped `{ labels }` override parsing.
- Added `frontend/src/utils/creditsShell.ts` for `credits_shell_config` labels, action block, buttons, and conversion override parsing.
- Added `frontend/src/utils/profileShell.ts` for `profile_shell_config` labels plus nested provider labels.
- Added `frontend/src/utils/availableChannelsShell.ts` for `available_channels_shell_config` labels plus nested table column labels.
- Migrated `UsageView.vue`, `AvailableGroupsView.vue`, `RedeemView.vue`, `ApiGuideView.vue`, `ApiTestView.vue`, `CustomPageView.vue`, `AffiliateView.vue`, `CreditsView.vue`, `ProfileView.vue`, and `AvailableChannelsView.vue` away from page-local parser functions.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/shellLabelOverrides.spec.ts src/utils/__tests__/creditsShell.spec.ts src/views/user/__tests__/UsageView.spec.ts src/views/user/__tests__/AvailableGroupsView.spec.ts src/views/user/__tests__/RedeemView.spec.ts src/views/user/__tests__/ApiTestView.spec.ts src/views/user/__tests__/CustomPageView.spec.ts src/views/user/__tests__/AffiliateView.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/views/user/__tests__/CreditsView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/shellLabelOverrides.spec.ts src/utils/__tests__/creditsShell.spec.ts src/utils/__tests__/profileShell.spec.ts src/views/user/__tests__/UsageView.spec.ts src/views/user/__tests__/AvailableGroupsView.spec.ts src/views/user/__tests__/RedeemView.spec.ts src/views/user/__tests__/ApiTestView.spec.ts src/views/user/__tests__/CustomPageView.spec.ts src/views/user/__tests__/AffiliateView.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/user/__tests__/ProfileView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/shellLabelOverrides.spec.ts src/utils/__tests__/creditsShell.spec.ts src/utils/__tests__/profileShell.spec.ts src/utils/__tests__/availableChannelsShell.spec.ts src/views/user/__tests__/UsageView.spec.ts src/views/user/__tests__/AvailableGroupsView.spec.ts src/views/user/__tests__/RedeemView.spec.ts src/views/user/__tests__/ApiTestView.spec.ts src/views/user/__tests__/CustomPageView.spec.ts src/views/user/__tests__/AffiliateView.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/user/__tests__/ProfileView.spec.ts src/views/user/__tests__/AvailableChannelsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "function parse[A-Za-z]*(Shell|Labels|Config)|parse[A-Za-z]*(Shell|Labels|Config)\\(" frontend/src/views/user frontend/src/views/public frontend/src/views/HomeView.vue frontend/src/components frontend/src/utils -g '*.vue' -g '*.ts'`

### Notes
- User/public/Home views no longer contain page-local shell parser functions.
- Remaining parser functions are now inside shared utility files: `authShell.ts`, `channelStatusShell.ts`, and `dashboardShellLabels.ts`.
- This does not yet move all UI composition into Sub2API backend/admin settings; it removes repeated front-end parsing code and keeps public settings as the source for runtime shell copy.

## 2026-06-19 Dashboard/auth/channel shell parser cleanup

### Done
- Migrated dashboard runtime shell labels to `resolveLocalizedShellLabels()` through `resolveDashboardShellLabels()`.
- Added direct dashboard shell label coverage for configured overrides and invalid JSON fallback.
- Renamed the remaining internal auth and channel status shell readers away from `parse*` parser-style entry points while preserving behavior.
- Confirmed no `parse*Shell/Labels/Config` functions remain in the migrated public, user, component, and utility frontend scope.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/views/user/__tests__/ChannelStatusView.spec.ts src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "function parse[A-Za-z]*(Shell|Labels|Config)|parse[A-Za-z]*(Shell|Labels|Config)\\(" frontend/src/views/user frontend/src/views/public frontend/src/views/HomeView.vue frontend/src/components frontend/src/utils -g '*.vue' -g '*.ts'`

### Notes
- The frontend still owns dashboard/auth/channel status presentation and interaction state.
- This is a shell-parser consolidation step, not a full merge of Touch UI into the Sub2API Vue app shell.

## 2026-06-19 Frontend bootstrap settings tightened

### Done
- Exposed `default_locale` through Sub2API public settings and injected `window.__APP_CONFIG__`.
- Updated frontend i18n bootstrap priority to use: saved user locale, then public settings `default_locale`, then browser locale, then English fallback.
- Removed the docs page dependency on `VITE_DOCS_CONTENT_VERSION`; Docsify hash/search cache busting now uses public settings `version`.
- Added regression coverage for runtime default locale selection.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/i18n/__tests__/runtimeDefaultLocale.spec.ts src/views/public/__tests__/DocsView.spec.ts`
- `pnpm run frontend:typecheck`
- `go test ./internal/handler/dto ./internal/handler -run 'TestPublicSettingsInjectionPayload|TestSettingHandler_GetPublicSettings'` from `backend/`

### Notes
- `VITE_API_BASE_URL` and related OAuth/API bootstrap paths remain local runtime configuration because the frontend still needs an initial backend base URL before public settings can be fetched.
- This reduces local frontend bootstrap settings but does not merge the remaining independent Vue shell into a single backend-rendered/admin-managed UI.

## 2026-06-19 Auth shell footer copy moved to public settings

### Done
- Added `allRightsReserved` to the shared auth shell label defaults.
- Updated `AuthLayout.vue` to render the auth footer copy through `auth_shell_config` via `resolveAuthShellLabels()`.
- Removed the layout-level hardcoded English copyright phrase while preserving an English default in the auth shell fallback map.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/components/layout/__tests__/AuthLayout.spec.ts src/i18n/__tests__/runtimeDefaultLocale.spec.ts src/views/public/__tests__/DocsView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Auth pages still own the form flow and OAuth button orchestration in Vue.
- This only moves another static auth shell copy string into Sub2API-managed public settings.

## 2026-06-19 Key Usage shell config boundary tightened

### Done
- Stopped `KeyUsageView.vue` from reading `docs_shell_config` and `home_shell_config` for its own header/footer copy.
- Added `docs` and `allRightsReserved` labels to `key_usage_shell_config` fallback copy.
- Updated the Sub2API backend default `key_usage_shell_config` so fresh/public settings include those labels.
- Added regression coverage proving docs/footer labels come from `key_usage_shell_config`, not the docs/home shell configs.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/KeyUsageView.spec.ts src/utils/__tests__/authShell.spec.ts src/components/layout/__tests__/AuthLayout.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsKeyUsageShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- Key Usage remains a Vue-owned public page, but its public shell copy is now scoped to its own Sub2API runtime setting.

## 2026-06-19 Keys shell parser shared resolver

### Done
- Replaced `KeysView.vue` page-local shell label parsing with the shared `resolveShellLabelOverrides()` utility.
- Removed the Keys page-local `readLocalizedShellLabels` / `isRecord` parser helpers.
- Added a source regression test to keep Keys page shell parsing on the shared resolver.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/KeysView.shell.spec.ts src/utils/__tests__/shellLabelOverrides.spec.ts src/views/__tests__/KeyUsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "function readLocalizedShellLabels|function isRecord|resolveShellLabelOverrides" frontend/src/views/user/KeysView.vue`

### Notes
- Keys UI and API key CRUD orchestration still live in the Vue frontend.
- This is a shell parser consolidation step, not a full move of the Keys experience into backend/admin settings.

## 2026-06-19 Key Usage shell parser shared resolver

### Done
- Replaced `KeyUsageView.vue` page-local `readLocalizedShellLabels` / `isRecord` parsing with shared `resolveLocalizedShellLabels()`.
- Kept `key_usage_shell_config` as the only runtime shell copy source for Key Usage page labels.
- Added a source regression assertion so the page does not reintroduce local parser helpers.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/KeyUsageView.spec.ts src/utils/__tests__/localizedShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "function readLocalizedShellLabels|function isRecord|resolveLocalizedShellLabels" frontend/src/views/KeyUsageView.vue`
- `git diff --check -- frontend/src/views/KeyUsageView.vue frontend/src/views/__tests__/KeyUsageView.spec.ts progress.md`

### Notes
- Key Usage remains a Vue-owned public page.
- This removes another page-local shell parser and keeps shell parsing centralized in shared frontend utilities while Sub2API public settings remain the runtime source.

## 2026-06-19 Legacy Touch runtime setting fallback removed

### Done
- Removed runtime reads of legacy `touch_*` setting keys from `GetPublicSettings()` / settings view construction.
- Removed `SettingKeyTouch*` constants from service domain constants.
- Kept `backend/migrations/149_copy_touch_runtime_settings_to_web.sql` as the historical migration path from old Touch keys to current Web/Sub2API keys.
- Updated public settings tests to assert legacy `touch_*` keys no longer affect runtime public settings.

### Validation
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings' -count=1`
- `cd backend && go test -tags unit ./internal/handler -run 'TestSettingHandler_GetPublicSettings_(ExposesGenericRuntimeSettingAliases|ExposesForceEmailOnThirdPartySignup|ExposesPasswordMinLength)' -count=1`
- `cd backend && go test -tags unit ./internal/handler/admin -run 'Test.*Setting.*Touch|Test.*AuthSource' -count=1`
- `rg -n "SettingKeyTouch|webRuntimeSetting" backend/internal/service backend/internal/handler backend/internal/server -g '*.go'`

### Notes
- Legacy `touch_*` strings intentionally remain in migration SQL and tests that assert legacy keys are not exposed.
- Existing installations must have migration 149 applied before depending on old Touch runtime settings under the new Web/Sub2API keys.

## 2026-06-19 Docs shell fallback copy moved out of Vue view

### Done
- Removed locale-specific docs shell fallback copy from `DocsView.vue`.
- Kept only an empty structural fallback object in the view; real default docs labels now come from Sub2API public settings via backend `defaultDocsShellConfig`.
- Updated DocsView regression coverage to assert the Vue view no longer embeds Chinese/English docs default copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts src/utils/__tests__/localizedShell.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsDocsShellConfig' -count=1`
- `rg -n "DEFAULT_DOCS_COPY|title: '文档'|dashboard: '控制台'|searchPlaceholder: '搜索文档'|title: 'Docs'|searchPlaceholder: 'Search docs'|EMPTY_DOCS_COPY" frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Docs rendering is still owned by the Vue frontend and Docsify.
- This specifically removes duplicated default copy from the Vue view; `docs_shell_config` remains the runtime source.

## 2026-06-19 Model Plaza shell fallback copy moved out of Vue view

### Done
- Removed locale-specific Model Plaza fallback copy from `ModelsPlazaView.vue`.
- Kept only an empty structural fallback object in the view; real default Model Plaza labels now come from Sub2API public settings via backend `defaultModelPlazaShellConfig`.
- Updated Model Plaza tests so rendered copy comes from `model_plaza_shell_config`, and added a source regression assertion that the Vue view no longer embeds Chinese/English default Model Plaza copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/ModelsPlazaView.spec.ts src/utils/__tests__/localizedShell.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsModelPlazaShellConfig' -count=1`
- `rg -n "DEFAULT_MODELS_PLAZA_COPY|badge: '模型广场'|title: '公开模型目录'|searchPlaceholder: '搜索模型、能力或标签'|badge: 'Model Plaza'|title: 'Public Model Catalog'|searchPlaceholder: 'Search models, capabilities, or tags'|EMPTY_MODELS_PLAZA_COPY" frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/__tests__/ModelsPlazaView.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Model Plaza rendering and filtering remain Vue-owned.
- This removes duplicated default copy from the Vue view; `model_plaza_shell_config` remains the runtime source.

## 2026-06-19 Legal document shell fallback copy moved out of Vue view

### Done
- Removed locale-specific legal document fallback copy from `LegalDocumentView.vue`.
- Kept only an empty structural fallback object in the view; real default legal document labels now come from Sub2API public settings via backend `defaultLegalDocumentShellConfig`.
- Updated LegalDocument tests so rendered copy comes from `legal_document_shell_config`, and added a source regression assertion that the Vue view no longer embeds Chinese/English default legal document copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/LegalDocumentView.spec.ts src/utils/__tests__/localizedShell.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsLegalDocumentShellConfig' -count=1`
- `rg -n "DEFAULT_LEGAL_DOCUMENT_COPY|login: '登录'|agreementLabel: '登录条款'|missingTitle: '文档不存在'|login: 'Log in'|agreementLabel: 'Login agreement'|missingTitle: 'Document not found'|EMPTY_LEGAL_DOCUMENT_COPY" frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/__tests__/LegalDocumentView.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Legal document rendering remains Vue-owned.
- This removes duplicated default copy from the Vue view; `legal_document_shell_config` remains the runtime source.

## 2026-06-19 Pricing shell fallback copy moved out of Vue view

### Done
- Removed locale-specific pricing fallback copy from `PricingView.vue`.
- Kept only an empty structural fallback object in the view; real default pricing labels now come from Sub2API public settings via backend `defaultPricingShellConfig`.
- Updated Pricing tests so rendered copy comes from `pricing_shell_config`, and added a source regression assertion that the Vue view no longer embeds Chinese/English default pricing copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/utils/__tests__/pricingShell.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPricingShellConfig' -count=1`
- `rg -n "DEFAULT_PRICING_COPY|title: '价格与套餐'|catalogStatus: '目录状态'|prompts: '提示词案例'|rechargeCta: '购买充值包'|title: 'Pricing'|catalogStatus: 'Catalog status'|prompts: 'Prompt cases'|rechargeCta: 'Buy credits'|EMPTY_PRICING_COPY" frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Pricing rendering and checkout routing remain Vue-owned.
- This removes duplicated default copy from the Vue view; `pricing_shell_config` remains the runtime source.

## 2026-06-19 Credits shell fallback copy moved out of Vue view

### Done
- Removed locale-specific credits fallback copy from `CreditsView.vue`.
- Kept only an empty structural fallback object in the view; real default credits labels now come from Sub2API public settings via backend `defaultCreditsShellConfig`.
- Updated Credits tests so rendered copy comes from `credits_shell_config`, and added a source regression assertion that the Vue view no longer embeds Chinese/English default credits copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/CreditsView.spec.ts src/utils/__tests__/creditsShell.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsCreditsShellConfig' -count=1`
- `rg -n "DEFAULT_CREDITS_COPY|title: '积分余额'|purchase: '购买积分'|orders: '订单记录'|recharge: '去充值'|viewOrders: '查看订单'|title: 'Credit Balance'|purchase: 'Purchase credits'|orders: 'Orders'|recharge: 'Recharge'|viewOrders: 'View orders'|EMPTY_CREDITS_COPY" frontend/src/views/user/CreditsView.vue frontend/src/views/user/__tests__/CreditsView.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Credits page rendering remains Vue-owned.
- This removes duplicated default copy from the Vue view; `credits_shell_config` remains the runtime source.

## 2026-06-19 Image workspace shell fallback copy moved out of Vue view

### Done
- Removed locale-specific workspace fallback copy from `ImageGeneratorView.vue`.
- Kept only an empty structural fallback object in the view; real default workspace labels now come from Sub2API public settings via backend `defaultWorkspaceShellConfig`.
- Updated ImageGenerator tests so rendered copy comes from `workspace_shell_config`, and added a source regression assertion that the Vue view no longer embeds Chinese/English default workspace copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/ImageGeneratorView.spec.ts src/utils/__tests__/imageWorkspaceShell.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsWorkspaceShellConfig' -count=1`
- `rg -n "DEFAULT_WORKSPACE_SHELL|catalogLabel: '提示词案例'|eyebrow: '提示词工作台'|title: 'AI 生图工作区'|promptPlaceholder: '输入或从案例库导入提示词'|copyPromptLabel: '复制提示词'|backToCatalogLabel: '返回案例库'|catalogLabel: 'Prompt catalog'|eyebrow: 'Prompt Workspace'|title: 'AI Image Workspace'|promptPlaceholder: 'Enter a prompt or import one from the catalog'|copyPromptLabel: 'Copy prompt'|backToCatalogLabel: 'Back to catalog'|EMPTY_WORKSPACE_SHELL" frontend/src/views/public/ImageGeneratorView.vue frontend/src/views/public/__tests__/ImageGeneratorView.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Image workspace rendering and prompt copy interactions remain Vue-owned.
- This removes duplicated default copy from the Vue view; `workspace_shell_config` remains the runtime source.

## 2026-06-19 Auth shell defaults moved out of frontend parser

### Done
- Removed `DEFAULT_AUTH_SHELL_LABELS` and all embedded auth/login/register default copy from `authShell.ts`.
- Changed `resolveAuthShellLabels()` to return labels only from `auth_shell_config`; missing or invalid config now exposes missing keys instead of silently using frontend defaults.
- Added `allRightsReserved` to backend `defaultAuthShellConfig` so AuthLayout footer copy is supplied by Sub2API public settings.
- Updated auth shell tests and backend public settings tests to cover the new source of truth.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/components/layout/__tests__/AuthLayout.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAuthShellConfig' -count=1`
- `rg -n "DEFAULT_AUTH_SHELL_LABELS|welcomeBack: '欢迎回来'|welcomeBack: 'Welcome Back'|allRightsReserved: '保留所有权利。'|allRightsReserved: 'All rights reserved.'" frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts frontend/src/components/layout/AuthLayout.vue frontend/src/views/auth -g '*.ts' -g '*.vue'`

### Notes
- Login/Register/AuthLayout UI remains Vue-owned.
- This removes the last scanned frontend `DEFAULT_*SHELL/LABELS` default-copy table from the auth parser; backend `auth_shell_config` is now the runtime default source.

## 2026-06-19 Channel status shell defaults moved out of frontend parser

### Done
- Removed `DEFAULT_LABELS` and embedded channel status copy from `channelStatusShell.ts`.
- Kept only an empty structural fallback for nested labels; real channel status defaults now come from Sub2API public settings via backend `defaultChannelStatusShellConfig`.
- Added channel status shell utility coverage to verify configured labels are used and invalid config no longer silently falls back to frontend copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/channelStatusShell.spec.ts src/views/user/__tests__/ChannelStatusView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsChannelStatusShellConfig' -count=1`
- `rg -n "const DEFAULT_.*COPY|DEFAULT_.*SHELL|DEFAULT_.*LABELS|const DEFAULT_LABELS|falls back to .* shell copy" frontend/src/views frontend/src/components frontend/src/utils -g '*.vue' -g '*.ts'`

### Notes
- Channel status page rendering and monitor polling remain Vue-owned.
- This removes another frontend default-copy parser; `channel_status_shell_config` remains the runtime source.

## 2026-06-19 Prompt catalog fallback copy moved out of Vue view

### Done
- Removed `FALLBACK_PROMPT_CATALOG_COPY` and embedded English Prompt Catalog copy from `PromptCatalogView.vue`.
- Kept only an empty structural fallback object in the view; real default Prompt Catalog labels now come from Sub2API public settings via backend `defaultPromptCatalogShellConfig`.
- Updated Prompt Catalog tests so all rendered labels come from `prompt_catalog_shell_config`, and added a source regression assertion that the Vue view no longer embeds default catalog copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig' -count=1`
- `rg -n "FALLBACK_PROMPT_CATALOG_COPY|title: 'Prompt Catalog'|description: 'Browse prompt cases from the shared prompt API.'|searchPlaceholder: 'Search prompts'|importTitle: 'Import from link'|importPlaceholder: 'Paste an X/Twitter post URL'|loadError: 'Failed to load prompt cases'|EMPTY_PROMPT_CATALOG_COPY" frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Prompt Catalog rendering, filters, modal interaction, and import form remain Vue-owned.
- This removes duplicated default copy from the Vue view; `prompt_catalog_shell_config` remains the runtime source.

## 2026-06-19 Home shell fallback copy moved out of Vue view

### Done
- Removed `FALLBACK_HOME_COPY` and embedded Home shell labels from `HomeView.vue`.
- Replaced embedded Home experience/why-choose card titles and descriptions with empty structural card placeholders keyed for config merging.
- Kept visual/card merge structure in Vue, while real Home labels and card copy now come from Sub2API public settings via backend `defaultHomeShellConfig`.
- Updated Home tests to assert configured labels/cards render and the Vue view no longer embeds default Home copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts src/utils/__tests__/homeShell.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsHomeShellConfig' -count=1`
- `rg -n "FALLBACK_.*COPY|FALLBACK_.*CARDS|DEFAULT_.*COPY|DEFAULT_.*SHELL|DEFAULT_.*LABELS|const DEFAULT_LABELS|AI Coding Workspace|One key, unified access|Less setup overhead|Browse prompt cases from the shared prompt API" frontend/src/views frontend/src/components frontend/src/utils -g '*.vue' -g '*.ts'`

### Notes
- Home layout, model-family rendering, and section composition remain Vue-owned.
- This removes duplicated default copy from the Home view; `home_shell_config` remains the runtime source.

## 2026-06-19 Key Usage shell fallback copy moved out of Vue view

### Done
- Removed `FALLBACK_KEY_USAGE_LABELS` and embedded Key Usage English/Chinese copy from `KeyUsageView.vue`.
- Kept only a fixed allowed-key list plus empty structural labels for shell config merging.
- Updated Key Usage tests so rendered page labels are supplied by `key_usage_shell_config`, and added a source regression assertion that the view no longer embeds default Key Usage copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/KeyUsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsKeyUsageShellConfig' -count=1`

### Notes
- Key Usage querying, API-key validation flow, and usage rendering remain Vue-owned.
- This removes duplicated default copy from the Vue page; `key_usage_shell_config` remains the runtime source.

## 2026-06-19 Usage shell i18n fallback keys moved out of Vue view

### Done
- Removed `usageShellFallbackKeys` from `UsageView.vue`.
- Changed `usageText()` so Usage page shell labels come only from `usage_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated Usage tests to inject runtime shell labels through mocked public settings, including tooltip and export status labels.
- Added a source regression assertion that the view no longer embeds usage shell i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/UsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsUsageShellConfig' -count=1`

### Notes
- Usage query/export behavior, CSV header names, billing-mode labels, service-tier labels, and image-size source labels remain Vue/business utility concerns.
- This removes another public page shell fallback; `usage_shell_config` remains the runtime source.

## 2026-06-19 API Guide shell i18n fallback keys moved out of Vue view

### Done
- Removed `apiGuideFallbackKeys` from `ApiGuideView.vue`.
- Changed `apiGuideText()` so API Guide labels come only from `api_guide_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated API Guide tests to build labels from mocked public settings and added a source regression assertion against reintroducing i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiGuideView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAPIGuideShellConfig' -count=1`

### Notes
- Gateway variant labels/protocol labels still come from gateway utility/i18n because they are shared enum metadata, not API Guide shell copy.
- This removes another local shell fallback; `api_guide_shell_config` remains the runtime source.

## 2026-06-19 API Test shell i18n fallback keys moved out of Vue view

### Done
- Removed `apiTestFallbackKeys` from `ApiTestView.vue`.
- Changed `apiTestText()` so API Test labels come only from `api_test_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated API Test tests to build labels from mocked public settings and added a source regression assertion against reintroducing i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiTestView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAPITestShellConfig' -count=1`

### Notes
- Gateway variants, model option labels, default prompt content, and request construction remain shared gateway utility concerns.
- This removes another local shell fallback; `api_test_shell_config` remains the runtime source.

## 2026-06-19 Available Groups shell i18n fallback keys moved out of Vue view

### Done
- Removed `availableGroupsFallbackKeys` from `AvailableGroupsView.vue`.
- Changed `availableGroupsText()` so Available Groups labels come only from `available_groups_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated Available Groups tests to build labels from mocked public settings and added a source regression assertion against reintroducing i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AvailableGroupsView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAvailableGroupsShellConfig' -count=1`

### Notes
- Group data loading, filtering, grouping, badges, and quota formatting remain Vue-owned.
- This removes another local shell fallback; `available_groups_shell_config` remains the runtime source.

## 2026-06-19 Redeem shell i18n fallback keys moved out of Vue view

### Done
- Removed `redeemFallbackKeys` from `RedeemView.vue`.
- Changed `redeemText()` so Redeem page labels come only from `redeem_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated Redeem tests with a source regression assertion against reintroducing i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/RedeemView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsRedeemShellConfig' -count=1`

### Notes
- Redeem submission, history rendering, subscription refresh, and contact info loading remain Vue-owned.
- This removes another local shell fallback; `redeem_shell_config` remains the runtime source.

## 2026-06-19 Affiliate shell i18n fallback keys moved out of Vue view

### Done
- Removed `affiliateFallbackKeys` from `AffiliateView.vue`.
- Changed `affiliateText()` so Affiliate page labels come only from `affiliate_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated Affiliate tests with a source regression assertion against reintroducing i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AffiliateView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAffiliateShellConfig' -count=1`

### Notes
- Affiliate data loading, clipboard operations, transfer behavior, payment-type display, and pagination remain Vue-owned.
- This removes another local shell fallback; `affiliate_shell_config` remains the runtime source.

## 2026-06-19 Custom Page shell i18n fallback keys moved out of Vue view

### Done
- Removed `customPageFallbackKeys` from `CustomPageView.vue`.
- Changed `customPageText()` so custom page labels come only from `custom_page_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated Custom Page tests with a source regression assertion against reintroducing i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/CustomPageView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsCustomPageShellConfig' -count=1`

### Notes
- Custom menu item resolution, admin fallback visibility, markdown loading/rendering, iframe behavior, and code-copy wiring remain Vue-owned.
- This removes another local shell fallback; `custom_page_shell_config` remains the runtime source.

## 2026-06-19 API Keys shell i18n fallback keys moved out of Vue view

### Done
- Removed `API_KEYS_I18N_KEYS` from `KeysView.vue`.
- Changed `apiKeysText()` so API Keys page labels come only from `api_keys_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated Keys shell tests with a source regression assertion against reintroducing i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/KeysView.shell.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAPIKeysShellConfig' -count=1`

### Notes
- API Key CRUD, group switching, usage display, CC Switch import, modal flow, and status mapping remain Vue-owned.
- This removes another local shell fallback; `api_keys_shell_config` remains the runtime source.

## 2026-06-19 Payment result and popup shell i18n fallback keys moved out of Vue views

### Done
- Removed local payment-result, Stripe popup, and Airwallex carrier-page fallback label maps from the Vue views.
- Changed `PaymentResultView.vue`, `StripePopupView.vue`, and `AirwallexPaymentView.vue` so shell labels come from `payment_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated the focused payment view tests to inject mocked public settings labels and added source regression assertions against reintroducing the old fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/StripePopupView.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig' -count=1`

### Notes
- Payment result verification, resume-token recovery, polling, Stripe popup message handling, and Airwallex local snapshot recovery remain Vue-owned.
- This removes another set of local payment shell fallbacks; `payment_shell_config` remains the runtime source for these small payment pages.

## 2026-06-19 Payment QR and Stripe carrier shell i18n fallback keys moved out of Vue views

### Done
- Removed local payment QR and Stripe carrier-page fallback label maps from `PaymentQRCodeView.vue` and `StripePaymentView.vue`.
- Changed both pages so shell labels come from `payment_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Added source regression assertions to the focused tests to prevent reintroducing the old payment i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentQRCodeView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig' -count=1`

### Notes
- QR rendering, countdown/polling, cancellation, Stripe element mounting, Alipay/WeChat confirmation, and popup redirect behavior remain Vue-owned.
- This removes another local payment shell fallback set; `payment_shell_config` remains the runtime source for these payment carrier pages.

## 2026-06-19 Payment status panel i18n fallback keys moved out of Vue component

### Done
- Removed `panelFallbackKeys` from `PaymentStatusPanel.vue`.
- Changed `panelText()` so status-panel labels come from parent-provided `payment_shell_config` labels; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated `PaymentStatusPanel` tests and added a source regression assertion against reintroducing the old payment i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/PaymentStatusPanel.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Polling, pending-order verification, QR rendering, countdown, cancellation, popup reopening, terminal state emission, and payment amount formatting remain Vue-owned.
- The parent payment page still owns label resolution from `payment_shell_config`; this component no longer carries its own local payment shell fallback map.

## 2026-06-19 Payment QR dialog and Stripe inline i18n fallback keys moved out of Vue components

### Done
- Removed `paymentQRDialogFallbackKeys` from `PaymentQRDialog.vue`.
- Removed `stripeInlineFallbackKeys` from `StripePaymentInline.vue`.
- Changed both components so labels come from parent-provided `payment_shell_config` labels; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Added source regression assertions to both component tests.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/PaymentStatusPanel.spec.ts src/components/payment/__tests__/PaymentQRDialog.spec.ts src/components/payment/__tests__/StripePaymentInline.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- QR dialog countdown, polling, pending-order verification, popup reopen, cancellation, Stripe element mounting, inline Stripe confirmation, and redirect event behavior remain Vue-owned.
- The parent payment page still owns label resolution from `payment_shell_config`; these child components no longer carry local payment shell fallback maps.

## 2026-06-19 User orders shell i18n fallback keys moved out of Vue view

### Done
- Removed `userOrdersFallbackKeys` from `UserOrdersView.vue`.
- Changed `paymentText()` so the user orders page labels come only from `payment_shell_config`; missing labels now surface their key instead of falling back to local Vue/i18n mappings.
- Updated `UserOrdersView` tests with a source regression assertion against reintroducing the old orders/payment i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/UserOrdersView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Order fetching, status filtering, pagination, cancellation, refund eligibility, refund request flow, and API error localization remain Vue-owned.
- `payment_shell_config` remains the runtime source for user order page shell labels.

## 2026-06-19 Subscriptions shell i18n fallback keys moved out of Vue view

### Done
- Removed `subscriptionFallbackKeys` from `SubscriptionsView.vue`.
- Changed `paymentText()` so subscription page labels come only from `payment_shell_config`; missing labels now surface their key with interpolation instead of falling back to local Vue/i18n mappings.
- Updated `SubscriptionsView` tests with a source regression assertion against reintroducing the old subscription/payment i18n fallback keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/UserOrdersView.spec.ts src/views/user/__tests__/SubscriptionsView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Subscription fetching, status display, renewal navigation, quota window formatting, expiry formatting, platform styling, and API error display remain Vue-owned.
- `payment_shell_config` remains the runtime source for subscription page shell labels.

## 2026-06-19 Order table and payment method selector fallbacks moved out of Vue components

### Done
- Removed `orderTableFallbackKeys` from `OrderTable.vue`.
- Removed local payment i18n fallback calls from `PaymentMethodSelector.vue`.
- Extended `OrderTable` labels to include known payment method labels and wired `UserOrdersView` plus `AdminOrdersView` to pass labels explicitly.
- Added focused component tests and source regression assertions for `OrderTable` and `PaymentMethodSelector`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/OrderTable.spec.ts src/components/payment/__tests__/PaymentMethodSelector.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- `OrderTable` still uses the runtime locale for amount formatting only; table shell labels now come from the caller.
- `PaymentMethodSelector` now relies on caller-provided shell labels and falls back to stable label keys/method ids instead of local Vue/i18n mappings.

## 2026-06-19 Recharge product and subscription plan card fallbacks moved out of Vue components

### Done
- Removed local recharge product i18n fallback calls from `RechargeProductCard.vue`.
- Removed local subscription plan i18n fallback calls and `useI18n` from `SubscriptionPlanCard.vue`.
- Added focused component tests and source regression assertions for both card components.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/RechargeProductCard.spec.ts src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- `RechargeProductCard` still uses the runtime locale for amount formatting only; labels now come from the parent payment page.
- `SubscriptionPlanCard` now relies entirely on caller-provided shell labels and stable label keys for missing labels.

## 2026-06-19 Order status badge fallback removed and unused amount input deleted

### Done
- Removed local `payment.status.*` i18n fallback mappings from `OrderStatusBadge.vue`.
- Changed `OrderStatusBadge` to render caller-provided status labels, falling back to stable status codes when labels are absent.
- Wired status badge labels through `OrderTable`, `UserOrdersView`, `PaymentResultView`, and `AdminOrdersView`.
- Deleted unused `AmountInput.vue`; production code no longer imports or renders it.
- Added focused tests for `OrderStatusBadge` and expanded related table/result tests.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/OrderStatusBadge.spec.ts src/components/payment/__tests__/OrderTable.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Affiliate rebate rows currently fall back to stable status codes because they are driven by `affiliate_shell_config`, not `payment_shell_config`.
- `OrderStatusBadge` no longer owns any payment shell text; callers choose their own runtime label source.

## 2026-06-19 PaymentView shell i18n fallback map removed

### Done
- Removed `paymentShellFallbackKeys` from `PaymentView.vue`.
- Changed `paymentText()` so the main payment page labels come only from `payment_shell_config`; missing labels now surface their key with interpolation instead of falling back to local Vue/i18n mappings.
- Added source regression assertions to `PaymentView` tests against reintroducing the local fallback map or old payment i18n keys.
- Added default `payment_shell_config` status badge labels (`statusPending`, `statusPaid`, etc.) in the backend defaults and covered them in the public settings test.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentView.spec.ts src/utils/__tests__/paymentShell.spec.ts`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig' -count=1`
- `pnpm run frontend:typecheck`

### Notes
- Payment order creation, WeChat resume/fallback handling, Stripe/Airwallex routing, recovery snapshots, subscription selection, and API error-code localization remain Vue-owned.
- The public payment shell path no longer carries local shell text fallback maps; remaining `payment.errors` references are error-code namespace lookups, not page shell labels.

## 2026-06-19 Profile shell fallback maps removed from Vue profile components

### Done
- Removed local Profile shell fallback maps from `ProfileView.vue` and the Profile child components: info card, avatar card, edit form, password form, balance notification card, TOTP card/dialogs, and identity bindings section.
- Changed Profile shell text resolution so labels come from `profile_shell_config`; missing labels now surface stable label keys instead of falling back to local Vue/i18n mappings.
- Updated Profile component tests to reflect the Sub2API-config-owned label contract while keeping existing business-flow coverage.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ProfileView.spec.ts src/components/user/profile/__tests__/ProfileInfoCard.spec.ts src/components/user/profile/__tests__/ProfileAvatarCard.spec.ts src/components/user/profile/__tests__/ProfileEditForm.spec.ts src/components/user/profile/__tests__/ProfilePasswordForm.spec.ts src/components/user/profile/__tests__/ProfileBalanceNotifyCard.spec.ts src/components/user/profile/__tests__/ProfileTotpCard.spec.ts src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts src/components/user/profile/__tests__/totp-timer-cleanup.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "FallbackKeys|fallbackKeys|const fallbackKeys|profileFallbackKeys|avatarFallbackKeys|profileEditFallbackKeys|passwordFallbackKeys|balanceNotifyFallbackKeys|totpSetupFallbackKeys|totpDisableFallbackKeys|totpFallbackKeys|authBindingFallbackKeys|\\bt\\(|profile\\.authBindings\\.providers|profile\\.avatar|profile\\.totp|profile\\.balanceNotify|profile\\.changePassword|common\\.save|common\\.cancel|auth\\.emailRequired" frontend/src/views/user/ProfileView.vue frontend/src/components/user/profile -g '*.vue'`

### Notes
- Profile data loading, profile updates, avatar compression, balance notification mutations, TOTP setup/disable, and auth binding/unbinding remain Vue-owned interaction logic.
- `ProfileView.vue` still uses `useI18n().locale` only to choose the locale branch from `profile_shell_config`.

## 2026-06-19 Dashboard shell frontend default copy removed

### Done
- Removed `dashboardShellFallbackKeys` and the frontend zh/en `defaultDashboardShellLabels` copy table from `dashboardShellLabels.ts`.
- Changed Dashboard shell label resolution to parse only `dashboard_shell_config`; missing or invalid config now yields empty labels.
- Updated Dashboard stats, charts, recent usage, and quick actions components so missing labels surface stable label keys instead of local default copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "dashboardShellFallbackKeys|defaultDashboardShellLabels|dashboard\\.balance|dashboard\\.apiKeys|common\\.available|common\\.refresh|\\bt\\(" frontend/src/components/user/dashboard frontend/src/views/user/DashboardView.vue -g '*.vue' -g '*.ts'`

### Notes
- Dashboard data fetching, chart ranges, recent usage loading, quick-action routing, and platform quota calculations remain Vue-owned interaction logic.
- `DashboardView.vue` still uses `useI18n().locale` only to choose the locale branch from `dashboard_shell_config`.

## 2026-06-19 Shared localized shell fallback copy contract removed

### Done
- Changed `resolveLocalizedShellLabels()` so it no longer accepts or merges caller-provided fallback copy.
- Updated Docs, Model Plaza, Legal Document, and Key Usage shell label parsing to rely only on their Sub2API public settings shell config.
- Removed empty fallback-copy constants from those views; missing configured labels now resolve to empty strings at the helper layer and page helpers can still surface stable keys where they already do so.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/localizedShell.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts src/views/__tests__/KeyUsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "EMPTY_DOCS_COPY|EMPTY_MODELS_PLAZA_COPY|EMPTY_LEGAL_DOCUMENT_COPY|EMPTY_KEY_USAGE_LABELS|DEFAULT_DOCS_COPY|DEFAULT_MODELS_PLAZA_COPY|DEFAULT_LEGAL_DOCUMENT_COPY|FALLBACK_KEY_USAGE_LABELS|resolveLocalizedShellLabels\\([^\\n]*,[^\\n]*,[^\\n]*,[^\\n]*\\)" frontend/src/views/public frontend/src/views/KeyUsageView.vue frontend/src/utils/localizedShell.ts frontend/src/utils/__tests__/localizedShell.spec.ts -S`

### Notes
- Docsify routing, docs content loading, model catalog filtering, legal document rendering, and key-usage query behavior remain Vue-owned interaction logic.
- The backend public settings defaults remain the source of configured copy for these shell configs.

## 2026-06-19 Credits pricing prompt catalog and channel status empty-copy constants removed

### Done
- Changed `resolveCreditsShellConfig()` and `resolvePricingShellConfig()` so they no longer accept caller-provided fallback copy.
- Removed `EMPTY_CREDITS_COPY`, `EMPTY_PRICING_COPY`, and `EMPTY_PROMPT_CATALOG_COPY` from their Vue views.
- Exported `promptCatalogCopyKeys` from the prompt catalog shell parser so the view can construct a typed empty label map without duplicating the key list.
- Replaced the `EMPTY_LABELS` channel-status parser constant with a local structure factory, keeping configured labels as the only copy source.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/creditsShell.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/utils/__tests__/pricingShell.spec.ts src/views/public/__tests__/PricingView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/channelStatusShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "EMPTY_CREDITS_COPY|EMPTY_PRICING_COPY|EMPTY_PROMPT_CATALOG_COPY|EMPTY_LABELS|DEFAULT_CREDITS_COPY|DEFAULT_PRICING_COPY|FALLBACK_PROMPT_CATALOG_COPY|DEFAULT_LABELS|resolveCreditsShellConfig\\([^\\n]*,[^\\n]*,[^\\n]*\\)|resolvePricingShellConfig\\([^\\n]*,[^\\n]*,[^\\n]*\\)" frontend/src/views/user/CreditsView.vue frontend/src/views/public/PricingView.vue frontend/src/views/public/PromptCatalogView.vue frontend/src/utils frontend/src/views/user/__tests__/CreditsView.spec.ts frontend/src/views/public/__tests__/PricingView.spec.ts frontend/src/views/public/__tests__/PromptCatalogView.spec.ts -S`

### Notes
- Credits balance math, pricing catalog loading, prompt catalog filtering/importing/detail modal behavior, and channel status calculations remain Vue-owned interaction logic.

## 2026-06-19 Image workspace shell frontend fallback contract removed

### Done
- Changed `resolveWorkspaceShellConfig()` so it no longer accepts caller-provided fallback copy.
- Removed `EMPTY_WORKSPACE_SHELL` from `ImageGeneratorView.vue`; the page now reads workspace shell copy only from Sub2API public settings.
- Updated Image Generator tests to guard against reintroducing local workspace shell copy or the old three-argument resolver contract.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/imageWorkspaceShell.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Image draft loading, prompt length validation, clipboard actions, and catalog back navigation remain Vue-owned interaction logic.
- Missing or invalid `workspace_shell_config` now yields empty workspace labels rather than frontend default copy.

## 2026-06-19 Home and available-channels shell fallback copy removed

### Done
- Removed `EMPTY_HOME_COPY`, `EMPTY_HOME_EXPERIENCE_CARDS`, and `EMPTY_HOME_WHY_CHOOSE_CARDS` from `HomeView.vue`.
- Changed `resolveHomeShellConfig()` to return complete empty labels and card arrays internally; the homepage now consumes only `home_shell_config` from Sub2API public settings for copy and card content.
- Removed `defaultAvailableChannelsLabels` from `AvailableChannelsView.vue`.
- Changed `resolveAvailableChannelsShellLabels()` so available-channel page labels come only from `available_channels_shell_config`; missing or invalid config yields empty labels.
- Added source-level regression tests to prevent reintroducing local homepage/available-channel shell copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/homeShell.spec.ts src/views/__tests__/HomeView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/availableChannelsShell.spec.ts src/views/user/__tests__/AvailableChannelsView.spec.ts`
- `pnpm run frontend:typecheck`
- Shell fallback scan now only reports negative regression assertions plus `openaiWsMode` business fallback keys.

### Notes
- Homepage public catalog loading, model-family matrix building, CTA routing, and footer section composition remain Vue-owned presentation logic.
- Available-channel search/filtering, group-rate loading, and table composition remain Vue-owned interaction logic.

## 2026-06-19 Frontend bootstrap brand fallbacks reduced

### Done
- Removed repeated page-level `Sub2API` site-name fallbacks from public pages, auth layout, register/email verification, and legal document rendering.
- Changed the global app store so `siteName` starts empty and is populated by public settings instead of a frontend default brand.
- Changed document title resolution so missing site name no longer falls back to a local brand; route titles render alone until settings provide a site name.
- Cleared the static `index.html` title so the runtime title comes from settings/router instead of local HTML bootstrap copy.
- Removed the login-agreement template default site name and changed payment product preview to use the configured site name rather than a hard-coded product prefix.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts src/views/__tests__/HomeView.spec.ts src/views/__tests__/KeyUsageView.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/LegalDocumentView.spec.ts src/views/auth/__tests__/EmailVerifyView.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/router/__tests__/title.spec.ts src/stores/__tests__/app.spec.ts src/router/__tests__/legacy-touch-alias-removal.spec.ts src/router/__tests__/runtime-settings-route.spec.ts src/router/__tests__/wechat-route.spec.ts src/router/__tests__/touch-oauth-compat.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts src/utils/__tests__/loginAgreementTemplates.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "\\|\\| 'Sub2API'|\\|\\| \\\"Sub2API\\\"|ref<string>\\('Sub2API'\\)|site_name \\|\\| 'Sub2API'|siteName\\.value = settings\\.site_name \\|\\| 'Sub2API'|AI API Gateway|<title>Sub2API|return value \\|\\| \\\"Sub2API\\\"|payment_product_name_prefix \\|\\| \\\"Sub2API\\\"|placeholder=\\\"Sub2API\\\"" frontend/src frontend/index.html -S`

### Notes
- Admin setup/settings still contain Sub2API-specific defaults and explanatory copy because that is the Sub2API management product itself, not retired Touch public-shell bootstrap.
- Static docs content still contains cloudbase/Sub2API instructional references and should be migrated separately if documentation content becomes settings-managed.

## 2026-06-19 Docs content base path moved to public settings

### Done
- Added `docs_content_base_path` to Sub2API public/admin settings, update payloads, defaults, DTOs, and public settings injection.
- Exposed the docs content base path in admin runtime settings so deployments can point Docsify at configured content paths instead of relying only on bundled frontend paths.
- Changed `DocsView` to resolve the Docsify content base path from public settings with the bundled `/docs-content/` and `/docs-content/en/` paths kept only as compatibility fallback.
- Added focused frontend utility coverage and backend public-settings coverage for the new setting.
- Adjusted the admin settings API contract test so extensible settings responses validate the required core field subset instead of embedding every shell-config string snapshot.

### Validation
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_Defaults(DocsShellConfig|DocsContentBasePath)' -count=1`
- `cd backend && go test -tags unit ./internal/server -run 'TestAPIContracts' -count=1`
- `cd backend && go test -tags unit ./internal/handler ./internal/server -run 'Test.*(Settings|PublicSettings|Runtime|Contract)' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/docsContentBasePath.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check` on the docs-settings related files

### Notes
- Bundled static docs content still exists under `frontend/public/docs-content` as a fallback path.
- This moves the runtime path selection into Sub2API settings; it does not yet migrate docs content authoring/storage into a CMS or admin editor.

## 2026-06-19 Pricing view local shell fallbacks removed

### Done
- Removed the remaining `shellLabel(..., fallback)` and `shellGroupLabel(..., fallback)` contracts from `PricingView`.
- Pricing tab labels and page labels now come from `pricing_shell_config` labels/groups, with empty strings from the shell parser used when config is missing.
- Added source-level regression assertions so the Pricing view cannot reintroduce local label/group fallback parameters.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/pricingShell.spec.ts src/views/public/__tests__/PricingView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check` on the pricing view/test/progress files

### Notes
- Pricing catalog loading, sorting, price formatting, CTA routing, and subscription/recharge card rendering remain Vue-owned interaction and presentation logic.
- The currency symbol still has a local compatibility fallback for missing settings; that is runtime formatting rather than shell copy.

## 2026-06-19 Prompt import URL validation delegated to Sub2API

### Done
- Removed `PromptCatalogView`'s local X/Twitter URL shape validator from the import flow.
- Removed the unused `importInvalidUrl` shell label from the frontend prompt catalog shell contract.
- Removed `importInvalidUrl` from the backend default `prompt_catalog_shell_config`, so public settings no longer publish frontend-only import validation copy.
- Added source-level regression checks to keep the prompt catalog view from reintroducing local URL validation or `copy.value.importInvalidUrl`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/promptCatalogShell.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts`
- `cd backend && go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `cd backend && go test -tags unit ./internal/handler ./internal/server -run 'Test.*(PublicSettings|Settings|PromptCatalog)' -count=1`
- `git diff --check` on the prompt-catalog/settings/progress files

### Notes
- X/Twitter import URL validation is now owned by the Sub2API import endpoint.
- Prompt Catalog filtering, pagination, detail modal state, and import form state remain Vue-owned UI orchestration.

## 2026-06-19 Web session cookies use only generic Sub2API names

### Done
- Removed the `touch_sub2api_access_token` and `touch_sub2api_refresh_token` legacy cookie constants from `WebHandler`.
- Stopped emitting legacy cookie clear headers during web login/session refresh/logout.
- Kept a regression test that injects a legacy Touch cookie by literal name and verifies generic web-session reads ignore it.

### Validation
- `cd backend && go test -tags unit ./internal/handler -run 'TestWeb(SessionCookies|CheckoutPaymentSource|ReadWebSessionCookie|ClearWebSessionCookies)' -count=1`
- `cd backend && go test -tags unit ./internal/server/routes -run 'Test(WebRoutesExposeOnlyGenericPrimaryRoutes|LegacyTouchAPIRoutesAreNotRegistered|PromptCatalog.*Alias)' -count=1`
- `rg -n "legacyWebAccessTokenCookie|legacyWebRefreshTokenCookie|touch_sub2api" backend/internal/handler/web_handler.go backend/internal/handler/web_handler_cookie_test.go -S`

### Notes
- Production web-session code now only names `sub2api_web_access_token` and `sub2api_web_refresh_token`.
- The only remaining `touch_sub2api` occurrence in this area is a negative regression test fixture.

## 2026-06-19 Pricing and credits runtime defaults moved out of frontend

### Done
- Removed the Pricing page's local `¥` fallback for `pricing_currency_symbol`; missing public settings now render without a frontend currency default.
- Removed the Runtime Settings form's local `pricing_currency_symbol: '¥'` default.
- Removed the Credits page's local `10` fallback for `credits_per_balance`; missing public settings now produce `0` until Sub2API settings load.
- Removed the Runtime Settings form's local `credits_per_balance: '10'` default.
- Added source-level regression checks for both removed frontend defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "pricing_currency_symbol:\\s*'¥'|pricing_currency_symbol\\?\\.trim\\(\\) \\|\\| '¥'|credits_per_balance:\\s*'10'|parsed : 10|\\? parsed : 10" frontend/src/views/public/PricingView.vue frontend/src/views/user/CreditsView.vue frontend/src/views/admin/RuntimeSettingsView.vue frontend/src/views/public/__tests__/PricingView.spec.ts frontend/src/views/user/__tests__/CreditsView.spec.ts frontend/src/views/admin/__tests__/RuntimeSettingsView.spec.ts -S`

### Notes
- The authoritative defaults remain in Sub2API settings (`pricing_currency_symbol`, `credits_per_balance`) rather than Vue page bootstrap state.
- Backend web-session credits payload still uses the backend settings helper and its backend default path.

## 2026-06-19 Admin payment currency display follows Sub2API data/settings

### Done
- Added a shared frontend payment-currency formatter that renders fiat paid amounts from each order's `currency` code instead of hard-coded `¥`.
- Updated admin order detail, order table, orders view, and refund dialog to use order currency for paid/refund displays and plain balance amounts for balance credits.
- Updated recharge product management so the highest-amount stat and preview price use `pricing_currency_symbol` from Sub2API public settings rather than a local `¥`.
- Added regression coverage for payment currency formatting and source-level checks that recharge product management stays settings-backed.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/AdminPaymentCatalogView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/utils/__tests__/paymentCurrency.spec.ts src/components/payment/__tests__/OrderTable.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "order_type === 'balance' \\? '\\$' : '¥'|>¥\\{\\{|>¥</span>|>\\$\\{\\{|\\$\\{\\{ userBalance|¥\\{\\{ order|\\$\\{\\{ order|pricing_currency_symbol:\\s*'¥'|credits_per_balance:\\s*'10'|parsed\\s*:\\s*10" frontend/src/views/admin/orders frontend/src/components/admin/payment frontend/src/views/public frontend/src/views/user frontend/src/utils frontend/src/views/admin/__tests__ frontend/src/views/public/__tests__ frontend/src/views/user/__tests__ -S`

### Notes
- Payment amount currency now follows order/provider data where available.
- Runtime defaults still belong in Sub2API settings; the Vue shell now renders empty prefixes when settings are absent instead of inventing local currency defaults.

## 2026-06-19 User payment result/order amount display uses shared formatter

### Done
- Updated `PaymentResultView` so credited amounts use the shared order amount formatter rather than a local `$` prefix for balance orders.
- Updated `PaymentStatusPanel` so successful paid-order amount display uses the shared credited-amount formatter.
- Updated `UserOrdersView` refund dialog amount display to use the shared credited-amount formatter.
- Removed the payment-result fallback to local `payment.methods.*` i18n keys for unknown providers; known providers still use `payment_shell_config` labels and unknown providers render their normalized provider key directly.
- Added source-level regression assertions for the removed local `$` and provider-label fallback paths.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts src/components/payment/__tests__/PaymentStatusPanel.spec.ts src/utils/__tests__/paymentCurrency.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "\\$\\{\\{[^\\n]*(order|refund|amount|pay|balance|price)|>\\$\\{[^\\n]*(order|refund|amount|pay|balance|price)|'\\$' \\+|\\\"\\$\\\" \\+|>\\$</span>|\\$\\{\\{ (order|refundTarget|paidOrder)|\\$\\{\\{ .*\\.toFixed\\(2\\)" frontend/src/views/user frontend/src/components/payment frontend/src/components/admin/payment frontend/src/views/admin/orders -S`

### Notes
- This continues the same migration line as the admin payment currency cleanup: user-visible payment amounts no longer have local frontend currency assumptions.
- Subscription plan card pricing still has some local presentation formatting and should be reviewed separately if the goal is full payment UI thinness.

## 2026-06-19 Subscription plan payment amounts use checkout currency

### Done
- Updated `SubscriptionPlanCard` so plan prices and original prices render through the shared payment amount formatter instead of hard-coded `$` markup.
- Passed the selected checkout currency and locale from `PaymentView` into subscription plan cards, including the renewal modal.
- Updated the subscription confirmation panel so `daily_limit_usd`, `weekly_limit_usd`, and `monthly_limit_usd` render through USD formatting instead of template-level `$` prefixes.
- Added regression assertions that the subscription plan card and payment view do not reintroduce template-level dollar prefixes for plan prices or USD limits.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/utils/__tests__/paymentCurrency.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n ">\\$</span>|\\$\\{\\{ (plan|selectedPlan|refundTarget|paidOrder|order)\\.|'\\$' \\+|\\\"\\$\\\" \\+|\\$\\{\\{ .*_(usd|price|amount).*\\}\\}" frontend/src/views/user/PaymentView.vue frontend/src/components/payment/SubscriptionPlanCard.vue frontend/src/views/user/PaymentResultView.vue frontend/src/views/user/UserOrdersView.vue frontend/src/components/payment/PaymentStatusPanel.vue frontend/src/components/admin/payment frontend/src/views/admin/orders -S`

### Notes
- Plan payment price currency still derives from the currently selected checkout method because the plan DTO itself does not expose a currency field.
- USD quota fields remain semantically USD, but the symbol/formatting is now centralized through Intl formatting instead of hard-coded template text.

## 2026-06-19 Payment child components stop rendering local label keys

### Done
- Removed local visible key fallbacks from payment child components (`PaymentMethodSelector`, `RechargeProductCard`, `SubscriptionPlanCard`, `OrderTable`, `PaymentQRDialog`, `StripePaymentInline`, and `PaymentStatusPanel`).
- Payment child components now render labels only from the parent-provided payment shell labels, with empty text when Sub2API public settings do not provide a label.
- Updated `OrderTable` credited amount display to use the shared payment currency formatter instead of a local `$` balance prefix.
- Added regression assertions to prevent reintroducing local label-key fallbacks or local balance amount prefixes in the payment components.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/PaymentMethodSelector.spec.ts src/components/payment/__tests__/RechargeProductCard.spec.ts src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/components/payment/__tests__/OrderTable.spec.ts src/components/payment/__tests__/PaymentQRDialog.spec.ts src/components/payment/__tests__/StripePaymentInline.spec.ts src/components/payment/__tests__/PaymentStatusPanel.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "label \\|\\| 'paymentMethod'|feeLabel \\|\\| 'fee'|rechargeProductRecommended|rechargeProductCreditLine|rechargeProductCta|labels\\?\\.[a-zA-Z]+ \\|\\| '[A-Za-z][A-Za-z0-9]+'|return props\\.labels\\?\\.\\[key\\] \\|\\| key|'\\$' \\+ row\\.amount\\.toFixed\\(2\\)" frontend/src/components/payment frontend/src/views/user/PaymentView.vue -S`

### Notes
- `PaymentView` remains responsible for translating Sub2API `payment_shell_config` into child-component labels.
- This removes another layer of Touch-local payment UI copy, but the payment page orchestration itself still lives in the Vue frontend.

## 2026-06-19 Profile child components stop rendering local label keys

### Done
- Removed visible label-key fallbacks from profile child components: edit form, avatar card, balance notification card, TOTP card/dialogs, profile info card, and identity bindings.
- Removed local provider-name fallbacks from profile info and identity-binding surfaces; provider display names now come from `profile_shell_config.providers`.
- Updated profile tests so any UI text needed by an interaction is supplied through configured shell labels rather than relying on component-local key names.
- Added source-level regression assertions for the removed key/provider fallback patterns.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/profile/__tests__/ProfileInfoCard.spec.ts src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts src/components/user/profile/__tests__/ProfileEditForm.spec.ts src/components/user/profile/__tests__/ProfileAvatarCard.spec.ts src/components/user/profile/__tests__/ProfileBalanceNotifyCard.spec.ts src/components/user/profile/__tests__/ProfileTotpCard.spec.ts src/components/user/profile/__tests__/totp-timer-cleanup.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "return props\\.labels\\?\\.\\[key\\] \\|\\| key|configured \\|\\| key|return interpolateLabel\\(configured \\|\\| key|props\\.labels\\?\\.providers\\?\\.[a-z]+ \\|\\| '[^']+'|return provider|configured \\|\\| props\\.oidcProviderName" frontend/src/components/user/profile -S`

### Notes
- `ProfileView` still owns profile-page layout and passes parsed `profile_shell_config` labels into these child components.
- This continues reducing Touch/Vue-local profile copy; it does not move profile interaction state out of the Vue frontend.

## 2026-06-19 Dashboard child components stop rendering local label keys

### Done
- Removed visible dashboard label-key fallbacks from dashboard stats, charts, recent usage, and quick action components.
- Dashboard child components now display labels only from `dashboard_shell_config` values parsed by `DashboardView`.
- Added regression coverage to prevent reintroducing `props.labels?.[key] || key` fallback copy in dashboard children.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "return props\\.labels\\?\\.\\[key\\] \\|\\| key|props\\.labels\\?\\.\\[key\\] \\|\\| key|interpolateDashboardLabel\\(props\\.labels\\?\\.\\[key\\] \\|\\| key" frontend/src/components/user/dashboard -S`

### Notes
- `DashboardView` still owns dashboard data loading, layout, date range state, chart state, and route interactions.
- This is a shell-copy cleanup step; it does not move dashboard interaction orchestration out of the Vue frontend.

## 2026-06-19 API guide and usage views stop rendering local label keys

### Done
- Removed visible shell label-key fallbacks from `ApiGuideView` and `UsageView`.
- Both views now rely on Sub2API public settings shell config (`api_guide_shell_config`, `usage_shell_config`) for display labels, with empty text when settings are absent.
- Extended existing source-level tests to prevent reintroducing `apiGuideLabels.value[key] || key` and `usageShellLabels.value[key] || key`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiGuideView.spec.ts src/views/user/__tests__/UsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "apiGuideLabels\\.value\\[key\\] \\|\\| key|usageShellLabels\\.value\\[key\\] \\|\\| key" frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/UsageView.vue frontend/src/views/user/__tests__/ApiGuideView.spec.ts frontend/src/views/user/__tests__/UsageView.spec.ts -S`

### Notes
- These views still own API guide and usage interaction state locally; this step only removes another layer of frontend-local copy fallback.

## 2026-06-19 Profile view stops rendering local label keys

### Done
- Removed the top-level `ProfileView` shell label-key fallback.
- `ProfileView` now relies on `profile_shell_config` from public settings or the fetched settings response for visible profile-page shell labels.
- Added a source-level regression assertion so `profileShellLabels.value[key] || key` cannot be reintroduced.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ProfileView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "profileShellLabels\\.value\\[key\\] \\|\\| key" frontend/src/views/user/ProfileView.vue frontend/src/views/user/__tests__/ProfileView.spec.ts -S`

### Notes
- This pairs with the profile child-component cleanup; profile layout and interaction state still remain in the Vue frontend.

## 2026-06-19 Remaining shell label key fallbacks removed from frontend

### Done
- Removed remaining visible `label || key` / `text || key` / `value || key` shell fallbacks from payment carrier/result/order pages, user feature pages, API key usage, and the shared auth shell renderer.
- Covered these surfaces: Stripe popup, Airwallex payment, payment result, QR payment, Stripe payment, user orders, available groups, redeem, API test, affiliate, custom page, API keys, key usage, and auth shell labels.
- Extended source-level regression checks across the affected tests so those frontend-local key fallbacks are not reintroduced.
- Verified the production frontend scan for shell key fallback patterns now returns no matches.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/StripePopupView.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AvailableGroupsView.spec.ts src/views/user/__tests__/RedeemView.spec.ts src/views/user/__tests__/ApiTestView.spec.ts src/views/user/__tests__/AffiliateView.spec.ts src/views/user/__tests__/CustomPageView.spec.ts src/views/user/__tests__/KeysView.shell.spec.ts src/views/__tests__/KeyUsageView.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "return .*\\|\\| key|\\[key\\] \\|\\| key|let text = .*\\|\\| key|let value = .*\\|\\| key|labels\\[key\\] \\|\\| key" frontend/src/views frontend/src/components frontend/src/utils -S --glob '!**/__tests__/**' || true`

### Notes
- Frontend pages still own layout, state, and interaction orchestration; this step removes the remaining visible local shell-copy key fallback layer.
- Some non-copy fallbacks remain intentionally, such as route defaults, numeric placeholders, status CSS defaults, and product/domain defaults.

## 2026-06-19 OIDC and provider-name bootstrap defaults removed from frontend

### Done
- Removed frontend-local `OIDC` provider-name defaults from the OIDC login section, login/register pages, OIDC callback, profile page, profile auth-binding components, and app public-settings fallback state.
- Removed the local `sub2api` provider-name fallback from CCS import link generation in `KeysView`; it now uses the configured public `site_name` or an empty value.
- Added source-level regression coverage so these frontend bootstrap defaults are not reintroduced.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/auth/__tests__/OidcOAuthSection.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts src/views/user/__tests__/ProfileView.spec.ts src/views/user/__tests__/KeysView.shell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "oidc_oauth_provider_name: 'OIDC'|oidcOAuthProviderName = ref<string>\\('OIDC'\\)|oidcOAuthProviderName = ref\\('OIDC'\\)|providerName = ref\\('OIDC'\\)|oidcProviderName: 'OIDC'|return name \\|\\| 'OIDC'|providerName: 'OIDC'|\\|\\| 'sub2api'|\\|\\| \\\"sub2api\\\"" frontend/src -S --glob '!**/__tests__/**'`

### Notes
- Admin settings still keep OIDC wording where it is part of the Sub2API management configuration UI.
- This continues reducing frontend bootstrap assumptions; auth/profile/page interaction state still lives in the Vue frontend.

## 2026-06-19 Site logo fallback removed from public frontend shell

### Done
- Removed bundled `/favicon.svg` fallback rendering from public page headers, auth layout, sidebar, home footer, prompt/image/pricing/model/legal/key-usage surfaces.
- Logo images now render only when Sub2API public settings provide a `site_logo`; missing logo settings no longer cause the Vue frontend to invent a local static brand asset.
- Added source-level regression coverage for public/layout logo surfaces.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/siteLogoFallback.spec.ts src/views/__tests__/HomeView.spec.ts src/views/__tests__/KeyUsageView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/components/layout/__tests__/AuthLayout.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "siteLogo \\|\\| '/favicon\\.svg'|siteLogo \\|\\| \\\"/favicon\\.svg\\\"|:src=\\\"siteLogo \\|\\||favicon\\.svg" frontend/src -S --glob '!**/__tests__/**'`

### Notes
- `frontend/index.html` still references `/favicon.svg` as the browser favicon; this step removes only page-rendered logo fallback assets.
- Page layout and interaction state remain in Vue; this step removes another frontend-local branding fallback.

## 2026-06-19 Docs content path frontend fallback removed

### Done
- Removed the frontend-bundled `/docs-content/` and `/docs-content/en/` fallback paths from `resolveDocsContentBasePath`.
- Docs content base path now comes from Sub2API public settings (`docs_content_base_path`); missing or unsafe values resolve to an empty base path instead of a frontend-owned static content path.
- Updated Docs view and resolver tests to assert docs path defaults are delegated to public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/docsContentBasePath.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "const bundledDocsBasePaths|/docs-content/|/docs-content/en/|fallback = bundledDocsBasePaths|return fallback|normalizeDocsBasePath\\([^\\n]*fallback" frontend/src/utils/docsContentBasePath.ts frontend/src/views/public/DocsView.vue -S`

### Notes
- Sub2API backend settings still provide the default `docs_content_base_path`; this change removes the duplicate frontend fallback.
- Bundled static files can still exist under `frontend/public/docs-content`, but the frontend no longer selects them unless settings point there.

## 2026-06-19 Usage cost currency prefix moved to public settings

### Done
- Removed hard-coded `$` prefixes from the user Usage page's visible cost summary, table cost cell, cost tooltip, image unit price, image total, token unit prices, cache cost, original cost, and billed cost.
- Added `formatUsageCost` / `formatUsageTokenPricePerMillion` in `UsageView` so display prefixes come from Sub2API public settings `pricing_currency_symbol`; missing settings render plain numeric amounts.
- Updated UsageView tests to assert configured currency prefixes are used and the view source no longer hard-codes dollar-prefixed template output.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/UsageView.spec.ts src/utils/__tests__/usageServiceTier.spec.ts src/utils/__tests__/paymentCurrency.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n -F '${{' frontend/src/views/user/UsageView.vue || true`
- `rg -n -F '>${{' frontend/src/views/user/UsageView.vue || true`

### Notes
- CSV export continues to emit raw numeric costs, unchanged.
- API Keys quota/rate-limit displays still have local dollar prefixes and should be migrated separately.

## 2026-06-19 API Keys quota currency prefix moved to public settings

### Done
- Removed hard-coded `$` prefixes from the user API Keys page's usage stats, quota display, rate-limit display, edit-modal quota usage, and rate-limit usage rows.
- Added `formatApiKeyCost` and `apiKeyCurrencyPrefix` so API Keys amount displays use Sub2API public settings `pricing_currency_symbol`; missing settings render plain numeric amounts.
- Updated API Keys shell source tests to prevent reintroducing dollar-prefixed template output.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/KeysView.shell.spec.ts src/utils/__tests__/ccswitchImport.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n '\\$\\{\\{|>\\$|\\$\\{\\{|absolute left-3[^\\n]*>\\$|>\\$</span>' frontend/src/views/user/KeysView.vue -S`

### Notes
- This changes presentation only; quota/rate-limit numeric payloads sent to Sub2API are unchanged.

## 2026-06-19 User balance and dashboard currency prefix moved to public settings

### Done
- Added `formatPublicMoneyAmount` so frontend money-like balances and costs can render with Sub2API public settings `pricing_currency_symbol` instead of local `$` template prefixes.
- Removed hard-coded dollar-prefixed output from Credits, Redeem, Payment credited balance, Subscriptions quota usage, AppHeader mobile balance, user Dashboard stats/charts/recent usage, and platform usage cost cells.
- Dashboard children now receive the runtime currency prefix from `DashboardView`; platform cost cells read the same public setting directly for the admin user list reuse path.
- Added regression coverage for configured currency symbols and source-level checks that the touched runtime views do not reintroduce dollar-prefixed Vue template output.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentCurrency.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/user/__tests__/RedeemView.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/SubscriptionsView.spec.ts src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/paymentCurrency.ts frontend/src/utils/__tests__/paymentCurrency.spec.ts frontend/src/views/user/CreditsView.vue frontend/src/views/user/RedeemView.vue frontend/src/views/user/PaymentView.vue frontend/src/views/user/DashboardView.vue frontend/src/views/user/SubscriptionsView.vue frontend/src/components/layout/AppHeader.vue frontend/src/components/user/dashboard/UserDashboardStats.vue frontend/src/components/user/dashboard/UserDashboardRecentUsage.vue frontend/src/components/user/dashboard/UserDashboardCharts.vue frontend/src/components/user/PlatformUsageBreakdown.vue frontend/src/components/user/PlatformCostCell.vue frontend/src/views/user/__tests__/CreditsView.spec.ts frontend/src/views/user/__tests__/RedeemView.spec.ts frontend/src/views/user/__tests__/PaymentView.spec.ts frontend/src/views/user/__tests__/SubscriptionsView.spec.ts frontend/src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts frontend/src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts`
- `rg -n -F '${{' frontend/src/views/user frontend/src/components/layout frontend/src/components/user frontend/src/components/payment -S --glob '!**/__tests__/**'`

### Notes
- The remaining `${{ ... }}` hits are only in `frontend/src/components/layout/EXAMPLES.md`; no touched runtime user page or component keeps dollar-prefixed template output.
- This batch is display-only. Payment currencies, `_usd` backend field names, order payloads, quota values, and billing semantics are unchanged.

## 2026-06-19 OAuth API base URL bootstrap centralized in API client

### Done
- Added `resolveApiBaseUrl` / `buildApiUrl` to `frontend/src/api/client.ts`, with priority: explicit public settings, injected `window.__APP_CONFIG__.api_base_url`, `VITE_API_BASE_URL`, then `/api/v1`.
- Removed direct `import.meta.env.VITE_API_BASE_URL || '/api/v1'` usage from OAuth start components, OAuth callback forwarding, WeChat callback restart, and user account-binding URL construction.
- Profile identity binding now passes Sub2API public settings into the OAuth binding URL builder, so configured `api_base_url` can drive bind redirects without component-local env reads.
- Added regression coverage for runtime API base URL resolution, email OAuth start URLs, account binding URLs, and OAuth component source scans.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/api/__tests__/client.spec.ts src/api/__tests__/user.spec.ts src/components/auth/__tests__/EmailOAuthButtons.spec.ts src/components/auth/__tests__/OidcOAuthSection.spec.ts src/components/auth/__tests__/WechatOAuthSection.spec.ts src/views/auth/__tests__/OAuthCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/api/client.ts frontend/src/api/user.ts frontend/src/api/__tests__/client.spec.ts frontend/src/api/__tests__/user.spec.ts frontend/src/components/auth/EmailOAuthButtons.vue frontend/src/components/auth/OidcOAuthSection.vue frontend/src/components/auth/DingTalkOAuthSection.vue frontend/src/components/auth/LinuxDoOAuthSection.vue frontend/src/components/auth/WechatOAuthSection.vue frontend/src/components/auth/__tests__/EmailOAuthButtons.spec.ts frontend/src/components/auth/__tests__/OidcOAuthSection.spec.ts frontend/src/components/user/profile/ProfileIdentityBindingsSection.vue frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/__tests__/OAuthCallbackView.spec.ts frontend/src/views/auth/__tests__/WechatCallbackView.spec.ts`
- `rg -n "VITE_API_BASE_URL|import\\.meta\\.env.*API" frontend/src -S --glob '!**/__tests__/**'`

### Notes
- `frontend/src/api/client.ts` remains the single frontend bootstrap point that can read `VITE_API_BASE_URL`; this is still needed before public settings are available.
- `frontend/src/vite-env.d.ts` still declares the env variable type. OAuth components and callback views no longer read it directly.

## 2026-06-19 Payment carrier defaults centralized in payment utilities

### Done
- Added `normalizePaymentCountryCode` and a centralized `DEFAULT_PAYMENT_COUNTRY_CODE` beside the existing payment currency normalization utility.
- Removed page-local `CNY` fallback usage from `StripePopupView`; the view now passes missing route currency through the shared normalizer.
- Removed page-local `CNY/CN` fallback usage from `AirwallexPaymentView`; the view now normalizes restored snapshot currency and country code through shared payment utilities.
- Removed page-local `ref('CNY')` defaults from `PaymentResultView` and `StripePaymentView`; both pages now initialize currency through the shared payment normalizer.
- Added regression coverage so these carrier pages do not reintroduce scattered `CNY/CN` fallback expressions or page-local currency refs.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/currency.spec.ts src/views/user/__tests__/StripePopupView.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/components/payment/currency.ts frontend/src/components/payment/__tests__/currency.spec.ts frontend/src/views/user/StripePopupView.vue frontend/src/views/user/AirwallexPaymentView.vue frontend/src/views/user/PaymentResultView.vue frontend/src/views/user/StripePaymentView.vue frontend/src/views/user/__tests__/StripePopupView.spec.ts frontend/src/views/user/__tests__/AirwallexPaymentView.spec.ts frontend/src/views/user/__tests__/PaymentResultView.spec.ts frontend/src/views/user/__tests__/StripePaymentView.spec.ts progress.md`
- `rg -n "\\|\\| 'CNY'|\\?\\? 'CNY'|ref\\('CNY'\\)|currency = ref\\('CNY'\\)|snapshot\\.countryCode \\|\\| 'CN'|\\?\\? 'CN'" frontend/src/views/user frontend/src/components/payment -S --glob '!**/__tests__/**'`

### Notes
- Payment fallback semantics remain unchanged: missing or invalid payment currency still resolves to `CNY`, and missing or invalid Airwallex country code still resolves to `CN`.
- This removes another layer of page-local payment bootstrap config, but those defaults are still centralized frontend utilities rather than Sub2API public settings.

## 2026-06-19 Payment method defaults centralized in payment utilities

### Done
- Added `paymentMethod.ts` with centralized visible-method and Stripe-method defaults plus Stripe popup method colors.
- Removed page-local Stripe popup method color/default declarations from `StripePopupView`.
- Removed scattered `wxpay` default handling from WeChat resume parsing and the payment OAuth redirect builder by routing missing methods through `resolveVisiblePaymentMethod`.
- Added regression coverage for centralized payment method defaults and source-level guards preventing page-local method defaults from returning.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/paymentMethod.spec.ts src/views/user/__tests__/StripePopupView.spec.ts src/views/user/__tests__/paymentWechatResume.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/components/payment/paymentMethod.ts frontend/src/components/payment/__tests__/paymentMethod.spec.ts frontend/src/views/user/StripePopupView.vue frontend/src/views/user/paymentWechatResume.ts frontend/src/views/user/PaymentView.vue frontend/src/views/user/__tests__/StripePopupView.spec.ts frontend/src/views/user/__tests__/paymentWechatResume.spec.ts frontend/src/views/user/__tests__/PaymentView.spec.ts`
- `rg -n "METHOD_COLORS|DEFAULT_METHOD_COLOR|route\\.query\\.method \\|\\| 'alipay'|normalizeVisibleMethod\\([^\\n]*\\) \\|\\| 'wxpay'|\\|\\| 'wxpay'" frontend/src/views/user frontend/src/components/payment -S --glob '!**/__tests__/**'`

### Notes
- Payment method fallback semantics remain unchanged: missing visible payment method still resolves to `wxpay`, and missing Stripe popup method still resolves to `alipay`.
- The remaining method literals in payment flows are domain values used for provider branching, API payloads, icon matching, and tests rather than page-local bootstrap defaults.

## 2026-06-19 OAuth start redirect defaults centralized

### Done
- Added `authRedirect.ts` with centralized auth login redirect, auth bind redirect, route-query redirect resolution, and safe local-path validation.
- Updated email, OIDC, DingTalk, LinuxDo, and WeChat OAuth start components to use the shared redirect resolver instead of page-local `/dashboard` fallback expressions.
- Updated OAuth binding URL construction and the profile identity-binding entry point to use the shared `/profile` bind redirect resolver.
- Added regression coverage for safe redirect handling, bind URL defaults, and OAuth start component source scans.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authRedirect.spec.ts src/components/auth/__tests__/EmailOAuthButtons.spec.ts src/components/auth/__tests__/OidcOAuthSection.spec.ts src/components/auth/__tests__/WechatOAuthSection.spec.ts src/api/__tests__/user.spec.ts src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/authRedirect.ts frontend/src/utils/__tests__/authRedirect.spec.ts frontend/src/components/auth/EmailOAuthButtons.vue frontend/src/components/auth/OidcOAuthSection.vue frontend/src/components/auth/DingTalkOAuthSection.vue frontend/src/components/auth/LinuxDoOAuthSection.vue frontend/src/components/auth/WechatOAuthSection.vue frontend/src/components/auth/__tests__/OidcOAuthSection.spec.ts frontend/src/api/user.ts frontend/src/api/__tests__/user.spec.ts frontend/src/components/user/profile/ProfileIdentityBindingsSection.vue`
- `rg -n "\\(route\\.query\\.redirect as string\\) \\|\\| '/dashboard'|redirectTo = .*\\|\\| '/dashboard'|options\\.redirectTo\\?\\.trim\\(\\) \\|\\| '/profile'|route\\.fullPath \\|\\| '/profile'" frontend/src/components/auth frontend/src/api/user.ts frontend/src/components/user/profile -S --glob '!**/__tests__/**'`

### Notes
- Login redirect semantics remain unchanged for normal local paths: missing login redirect still resolves to `/dashboard`, and missing binding redirect still resolves to `/profile`.
- OAuth callback views still contain their own local `sanitizeRedirectPath` implementations and `/dashboard` defaults; those larger callback files can be migrated in a separate focused pass.

## 2026-06-19 OAuth callback redirect defaults centralized

### Done
- Reused `authRedirect.ts` across OAuth callback and auth completion flows.
- Removed duplicated `sanitizeRedirectPath` implementations from generic OAuth, OIDC, LinuxDo, DingTalk, WeChat, and DingTalk email-completion views.
- Replaced callback-local `/dashboard` and bind-local `/profile` defaults with `DEFAULT_AUTH_REDIRECT_PATH`, `DEFAULT_AUTH_BIND_REDIRECT_PATH`, and `sanitizeAuthRedirectPath`.
- Updated login, registration, and email-verification success redirects to use the same centralized redirect defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authRedirect.spec.ts src/views/auth/__tests__/OAuthCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts src/views/auth/__tests__/EmailVerifyView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/authRedirect.ts frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/DingTalkEmailCompletionView.vue frontend/src/views/auth/LoginView.vue frontend/src/views/auth/RegisterView.vue frontend/src/views/auth/EmailVerifyView.vue`
- `rg -n "function sanitizeRedirectPath|sanitizeRedirectPath|const redirectTo = ref\\('/dashboard'\\)|\\|\\| '/dashboard'|router\\.push\\('/dashboard'\\)|\\|\\| '/profile'|route\\.fullPath \\|\\| '/profile'" frontend/src/views/auth frontend/src/components/auth frontend/src/api/user.ts frontend/src/components/user/profile -S --glob '!**/__tests__/**' --glob '!**/*.md'`

### Notes
- Redirect behavior remains local-path-only; unsafe external or malformed redirects still fall back to `/dashboard`, and bind completion still falls back to `/profile`.
- Auth views now share one redirect contract; remaining `/dashboard` references in tests and auth documentation are expected examples/assertions.

## 2026-06-19 Models Plaza provider display defaults centralized

### Done
- Added `modelPlazaDisplay.ts` to centralize Models Plaza provider group keys, group ordering, avatar initials, and provider icon gradient classes.
- Updated `ModelsPlazaView` to consume the shared display helpers instead of keeping page-local `all`/`other`/`M` and provider-group mapping logic.
- Moved Models Plaza shell copy schema, allowed copy keys, provider group labels, and simple template formatting out of the view and into the shared Models Plaza display helper.
- Added regression coverage for the shared display helper behavior and source-level guards so the page does not reintroduce local provider display or copy-schema defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/modelPlazaDisplay.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/modelPlazaDisplay.ts frontend/src/utils/__tests__/modelPlazaDisplay.spec.ts frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/__tests__/ModelsPlazaView.spec.ts progress.md`
- `rg -n "activeGroup = ref\('all'\)|return normalized \|\| 'other'|\|\| 'M'|function providerGroupKey|function providerGroupRank|bg-\[linear-gradient\(135deg,#ef8e67" frontend/src/views/public/ModelsPlazaView.vue -S`
- `rg -n "type ModelsPlazaCopy|const modelsPlazaCopyKeys|function formatTemplate|MODEL_PLAZA_OTHER_GROUP_KEY|resolveLocalizedShellLabels|function providerGroupKey|function providerGroupRank|\|\| 'M'|return normalized \|\| 'other'" frontend/src/views/public/ModelsPlazaView.vue -S`

### Notes
- Display behavior remains unchanged: empty provider groups still fall under `other`, and an empty provider avatar initial still falls back to `M`.
- This reduces another Models Plaza page-local UI default and copy-schema block; the page still owns layout, search state, and visible-card filtering.

## 2026-06-19 Image workspace template formatting centralized

### Done
- Moved image workspace shell template interpolation from `ImageGeneratorView` into `imageWorkspaceShell.ts`.
- Updated `ImageGeneratorView` to call `formatWorkspaceShellTemplate` instead of keeping a page-local formatter.
- Added regression coverage for workspace shell template formatting and a view source guard preventing the local formatter from returning.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/imageWorkspaceShell.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/imageWorkspaceShell.ts frontend/src/utils/__tests__/imageWorkspaceShell.spec.ts frontend/src/views/public/ImageGeneratorView.vue frontend/src/views/public/__tests__/ImageGeneratorView.spec.ts progress.md`
- `rg -n "function formatTemplate|template\.replace\(/\\\{\(\\w\+\)\\\}/g|formatTemplate\(" frontend/src/views/public/ImageGeneratorView.vue -S`

### Notes
- Workspace shell content still comes from Sub2API public settings; this only removes another small piece of page-local presentation plumbing.
- `ImageGeneratorView` still owns the prompt draft state and copy interaction.

## 2026-06-19 Public document shell schemas centralized

### Done
- Added `legalDocumentShell.ts` to centralize Legal Document shell copy keys, public-settings copy resolution, and updated-date template formatting.
- Updated `LegalDocumentView` to consume the legal document shell helper instead of keeping page-local copy schema and formatter logic.
- Added `docsShell.ts` to centralize Docs shell copy keys and public-settings copy resolution.
- Updated `DocsView` to consume the Docs shell helper instead of keeping page-local copy schema and localized shell parser wiring.
- Added regression coverage and source guards for both public document pages.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/legalDocumentShell.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts src/utils/__tests__/docsShell.spec.ts src/views/public/__tests__/DocsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/legalDocumentShell.ts frontend/src/utils/__tests__/legalDocumentShell.spec.ts frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/__tests__/LegalDocumentView.spec.ts frontend/src/utils/docsShell.ts frontend/src/utils/__tests__/docsShell.spec.ts frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts progress.md`
- `rg -n "type LegalDocumentCopy|const legalDocumentCopyKeys|function formatTemplate|resolveLocalizedShellLabels\(" frontend/src/views/public/LegalDocumentView.vue -S`
- `rg -n "type DocsShellCopy|const docsShellCopyKeys|resolveLocalizedShellLabels\(" frontend/src/views/public/DocsView.vue -S`

### Notes
- Page behavior remains unchanged; legal/docs shell content still comes from Sub2API public settings.
- The residual matches in the scans are type imports from the new helpers, not page-local schema definitions.

## 2026-06-19 Profile auth binding shell schema centralized

### Done
- Moved profile auth binding label keys, provider label interpolation, legacy note compatibility mapping, and auth binding text interpolation into `profileShell.ts`.
- Updated `ProfileIdentityBindingsSection` to consume the shared profile shell helpers instead of maintaining its own shell schema and formatter.
- Added regression coverage for auth binding helper behavior and source guards preventing the component-local schema from returning.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/profileShell.spec.ts src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/legalDocumentShell.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts src/utils/__tests__/docsShell.spec.ts src/views/public/__tests__/DocsView.spec.ts src/utils/__tests__/profileShell.spec.ts src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/legalDocumentShell.ts frontend/src/utils/__tests__/legalDocumentShell.spec.ts frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/__tests__/LegalDocumentView.spec.ts frontend/src/utils/docsShell.ts frontend/src/utils/__tests__/docsShell.spec.ts frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts frontend/src/utils/profileShell.ts frontend/src/utils/__tests__/profileShell.spec.ts frontend/src/components/user/profile/ProfileIdentityBindingsSection.vue frontend/src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts progress.md`
- `rg -n "const authBindingLabelKeys|const legacyBindingNoteKeys|const noteKeyMap|function interpolateLabel|template\.replace\(/\\\{\(\\w\+\)\\\}/g" frontend/src/components/user/profile/ProfileIdentityBindingsSection.vue -S`

### Notes
- Profile auth binding labels still come from the profile shell config passed by the parent view; this change only removes local schema/formatter ownership from the component.
- Component business behavior for email binding, OAuth binding, and unbinding is unchanged.

## 2026-06-19 User shell label contracts centralized

### Done
- Moved Available Channels label keys and full-label resolver into `availableChannelsShell.ts`; `AvailableChannelsView` now passes only shell config and locale.
- Added `redeemShell.ts` to centralize Redeem label keys, shell label resolution, and placeholder rendering.
- Updated `RedeemView` to consume the Redeem shell helper instead of keeping page-local label keys and interpolation logic.
- Added `affiliateShell.ts` to centralize Affiliate label keys, shell label resolution, and placeholder rendering.
- Updated `AffiliateView` to consume the Affiliate shell helper instead of keeping page-local label keys and interpolation logic.
- Added focused helper tests and source guards for all three user pages.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/availableChannelsShell.spec.ts src/views/user/__tests__/AvailableChannelsView.spec.ts src/utils/__tests__/redeemShell.spec.ts src/views/user/__tests__/RedeemView.spec.ts src/utils/__tests__/affiliateShell.spec.ts src/views/user/__tests__/AffiliateView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/availableChannelsShell.ts frontend/src/utils/__tests__/availableChannelsShell.spec.ts frontend/src/views/user/AvailableChannelsView.vue frontend/src/views/user/__tests__/AvailableChannelsView.spec.ts frontend/src/utils/redeemShell.ts frontend/src/utils/__tests__/redeemShell.spec.ts frontend/src/views/user/RedeemView.vue frontend/src/views/user/__tests__/RedeemView.spec.ts frontend/src/utils/affiliateShell.ts frontend/src/utils/__tests__/affiliateShell.spec.ts frontend/src/views/user/AffiliateView.vue frontend/src/views/user/__tests__/AffiliateView.spec.ts progress.md`
- `rg -n "const availableChannelsLabelKeys|resolveAvailableChannelsShellLabels\([^\n]*,[^\n]*,[^\n]*\)|const redeemLabelKeys|resolveShellLabelOverrides\(|const affiliateLabelKeys|function affiliateText\([^\n]*\)[\s\S]*?Object\.entries|function redeemText\([^\n]*\)[\s\S]*?Object\.entries" frontend/src/views/user/AvailableChannelsView.vue frontend/src/views/user/RedeemView.vue frontend/src/views/user/AffiliateView.vue -S`

### Notes
- Runtime behavior remains unchanged; all three pages still read copy from Sub2API public shell settings.
- This removes another set of page-local shell contracts, but the pages still own their data loading and interaction state.

## 2026-06-19 More user shell label contracts centralized

### Done
- Added `availableGroupsShell.ts` to centralize Available Groups label keys, shell label resolution, and placeholder rendering.
- Updated `AvailableGroupsView` to consume the Available Groups shell helper instead of keeping page-local label keys/interpolation.
- Added `customPageShell.ts` to centralize Custom Page label keys and shell label resolution.
- Updated `CustomPageView` to consume the Custom Page shell helper instead of keeping page-local label keys.
- Added `userOrdersShell.ts` to centralize User Orders label keys and payment-shell resolution.
- Updated `UserOrdersView` to consume the User Orders shell helper instead of keeping page-local order label keys.
- Added focused helper tests and source guards for all three pages.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/availableGroupsShell.spec.ts src/views/user/__tests__/AvailableGroupsView.spec.ts src/utils/__tests__/customPageShell.spec.ts src/views/user/__tests__/CustomPageView.spec.ts src/utils/__tests__/userOrdersShell.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/availableGroupsShell.ts frontend/src/utils/__tests__/availableGroupsShell.spec.ts frontend/src/views/user/AvailableGroupsView.vue frontend/src/views/user/__tests__/AvailableGroupsView.spec.ts frontend/src/utils/customPageShell.ts frontend/src/utils/__tests__/customPageShell.spec.ts frontend/src/views/user/CustomPageView.vue frontend/src/views/user/__tests__/CustomPageView.spec.ts frontend/src/utils/userOrdersShell.ts frontend/src/utils/__tests__/userOrdersShell.spec.ts frontend/src/views/user/UserOrdersView.vue frontend/src/views/user/__tests__/UserOrdersView.spec.ts progress.md`
- `rg -n "const availableGroupsLabelKeys|resolveShellLabelOverrides\(|const customPageLabelKeys|const userOrdersLabelKeys|resolvePaymentShellLabels\(" frontend/src/views/user/AvailableGroupsView.vue frontend/src/views/user/CustomPageView.vue frontend/src/views/user/UserOrdersView.vue -S`

### Notes
- Runtime behavior remains unchanged; copy still comes from Sub2API public shell settings.
- User Orders continues to read `payment_shell_config`; the new helper only centralizes the order-page label contract.

## 2026-06-19 API and usage shell label contracts centralized

### Done
- Added `usageShell.ts` to centralize Usage page label keys and shell label resolution.
- Updated `UsageView` to consume the Usage shell helper instead of keeping page-local label keys.
- Added `apiGuideShell.ts` to centralize API Guide label keys and shell label resolution.
- Updated `ApiGuideView` to consume the API Guide shell helper instead of keeping page-local label keys.
- Added `apiTestShell.ts` to centralize API Test label keys, shell label resolution, and placeholder rendering.
- Updated `ApiTestView` to consume the API Test shell helper instead of keeping page-local label keys/interpolation.
- Added focused helper tests and source guards for all three pages.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/usageShell.spec.ts src/views/user/__tests__/UsageView.spec.ts src/utils/__tests__/apiGuideShell.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/utils/__tests__/apiTestShell.spec.ts src/views/user/__tests__/ApiTestView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/usageShell.ts frontend/src/utils/__tests__/usageShell.spec.ts frontend/src/views/user/UsageView.vue frontend/src/views/user/__tests__/UsageView.spec.ts frontend/src/utils/apiGuideShell.ts frontend/src/utils/__tests__/apiGuideShell.spec.ts frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/__tests__/ApiGuideView.spec.ts frontend/src/utils/apiTestShell.ts frontend/src/utils/__tests__/apiTestShell.spec.ts frontend/src/views/user/ApiTestView.vue frontend/src/views/user/__tests__/ApiTestView.spec.ts progress.md`
- `rg -n "const usageShellLabelKeys|const apiGuideLabelKeys|const apiTestLabelKeys|resolveShellLabelOverrides\(" frontend/src/views/user/UsageView.vue frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/ApiTestView.vue -S`

### Notes
- Runtime behavior remains unchanged; all three pages still read copy from Sub2API public shell settings.
- API Test still owns request execution and usage-record synchronization state; this only removes page-local shell label contracts.

## 2026-06-19 Payment component shell contracts centralized

### Done
- Moved Payment QR Dialog label keys and text rendering into `paymentShell.ts`.
- Moved Stripe inline payment label keys and text rendering into `paymentShell.ts`.
- Moved Order Table label keys and text rendering into `paymentShell.ts`.
- Updated the three payment components to consume shared payment shell contracts instead of keeping component-local key lists.
- Added focused helper coverage and source guards to prevent the component-local shell schemas from returning.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/components/payment/__tests__/PaymentQRDialog.spec.ts src/components/payment/__tests__/StripePaymentInline.spec.ts src/components/payment/__tests__/OrderTable.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/paymentShell.ts frontend/src/utils/__tests__/paymentShell.spec.ts frontend/src/components/payment/PaymentQRDialog.vue frontend/src/components/payment/StripePaymentInline.vue frontend/src/components/payment/OrderTable.vue frontend/src/components/payment/__tests__/PaymentQRDialog.spec.ts frontend/src/components/payment/__tests__/StripePaymentInline.spec.ts frontend/src/components/payment/__tests__/OrderTable.spec.ts`
- `rg -n "const paymentQRDialogLabelKeys|const stripeInlineLabelKeys|const orderTableLabelKeys|resolvePaymentShellLabels\\(" frontend/src/components/payment/PaymentQRDialog.vue frontend/src/components/payment/StripePaymentInline.vue frontend/src/components/payment/OrderTable.vue -S`

### Notes
- Runtime behavior is unchanged; parent views still provide payment shell labels from Sub2API public settings.
- This removes one more batch of Touch-style frontend shell contracts, but payment pages still own their flow state and gateway interactions.

## 2026-06-19 Payment page shell contracts centralized

### Done
- Moved Stripe Popup page label keys, shell resolution, and text rendering into `paymentShell.ts`.
- Moved Airwallex redirect page label keys, shell resolution, and text rendering into `paymentShell.ts`.
- Moved standalone Payment QR page label keys, shell resolution, and text rendering into `paymentShell.ts`.
- Updated all three pages to consume shared payment shell contracts while keeping their route, recovery, and gateway behavior local.
- Added focused helper coverage and source guards for the three pages.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/views/user/__tests__/StripePopupView.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/paymentShell.ts frontend/src/utils/__tests__/paymentShell.spec.ts frontend/src/views/user/StripePopupView.vue frontend/src/views/user/AirwallexPaymentView.vue frontend/src/views/user/PaymentQRCodeView.vue frontend/src/views/user/__tests__/StripePopupView.spec.ts frontend/src/views/user/__tests__/AirwallexPaymentView.spec.ts frontend/src/views/user/__tests__/PaymentQRCodeView.spec.ts progress.md`
- `rg -n "const stripePopupLabelKeys|const airwallexPaymentLabelKeys|const paymentQRLabelKeys|resolvePaymentShellLabels\\(" frontend/src/views/user/StripePopupView.vue frontend/src/views/user/AirwallexPaymentView.vue frontend/src/views/user/PaymentQRCodeView.vue -S`

### Notes
- Runtime behavior is unchanged; all three pages still read `payment_shell_config` from Sub2API public settings.
- These pages still own payment flow state, popup messaging, Airwallex recovery, and QR polling; only shell label contracts moved out.

## 2026-06-19 Payment result and subscription shell contracts centralized

### Done
- Moved Payment Result page label keys, shell resolution, and text rendering into `paymentShell.ts`.
- Moved Subscriptions page label keys, shell resolution, and placeholder rendering into `paymentShell.ts`.
- Moved Stripe carrier page label keys, shell resolution, and text rendering into `paymentShell.ts`.
- Updated the three pages to consume shared payment shell contracts while preserving payment status, subscription quota, and Stripe interaction logic.
- Added helper coverage and source guards for all three pages.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/SubscriptionsView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/paymentShell.ts frontend/src/utils/__tests__/paymentShell.spec.ts frontend/src/views/user/PaymentResultView.vue frontend/src/views/user/SubscriptionsView.vue frontend/src/views/user/StripePaymentView.vue frontend/src/views/user/__tests__/PaymentResultView.spec.ts frontend/src/views/user/__tests__/SubscriptionsView.spec.ts frontend/src/views/user/__tests__/StripePaymentView.spec.ts progress.md`
- `rg -n "const paymentResultLabelKeys|const subscriptionLabelKeys|const stripePaymentLabelKeys|resolvePaymentShellLabels\\(" frontend/src/views/user/PaymentResultView.vue frontend/src/views/user/SubscriptionsView.vue frontend/src/views/user/StripePaymentView.vue -S`

### Notes
- Runtime behavior is unchanged; all three pages still read `payment_shell_config` from Sub2API public settings.
- `PaymentView` remains the last large payment page with a local shell label key list and should be handled as a standalone cleanup batch.

## 2026-06-19 Payment view shell contract centralized

### Done
- Moved the main Payment page label keys, shell resolution, and placeholder rendering into `paymentShell.ts`.
- Updated `PaymentView` to consume `resolvePaymentViewLabels` and `renderPaymentViewText` instead of keeping its own large label key list.
- Added helper coverage for the Payment page shell contract and source guards to keep the schema out of `PaymentView`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/paymentShell.ts frontend/src/utils/__tests__/paymentShell.spec.ts frontend/src/views/user/PaymentView.vue frontend/src/views/user/__tests__/PaymentView.spec.ts progress.md`
- `rg -n "const paymentShellLabelKeys|resolvePaymentShellLabels\\(" frontend/src/views/user/PaymentView.vue frontend/src/components/payment frontend/src/views/user -S --glob '!**/__tests__/**'`
- `rg -n "const .*LabelKeys = \\[|resolvePaymentShellLabels\\(" frontend/src/views/user frontend/src/components/payment -S --glob '!**/__tests__/**'`

### Notes
- Runtime behavior is unchanged; `PaymentView` still reads `payment_shell_config` from Sub2API public settings and owns checkout/payment flow state.
- Payment shell contract ownership is now centralized in `paymentShell.ts`; remaining local frontend shell key lists are outside payment (`KeysView`, `ProfileView`).

## 2026-06-19 Profile view shell schema centralized

### Done
- Moved `ProfileView` label keys and provider keys into `profileShell.ts`.
- Updated `ProfileView` to import the shared profile shell schema and keep only runtime locale/config wiring locally.
- Added helper coverage for the exported profile/provider schema and source guards to prevent local schema ownership returning to `ProfileView`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/profileShell.spec.ts src/views/user/__tests__/ProfileView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/profileShell.ts frontend/src/utils/__tests__/profileShell.spec.ts frontend/src/views/user/ProfileView.vue frontend/src/views/user/__tests__/ProfileView.spec.ts progress.md`
- `rg -n "const profileLabelKeys|const profileProviderKeys|resolveProfileShellLabels\\([^\\n]*,[^\\n]*,[^\\n]*,[^\\n]*\\)" frontend/src/views/user/ProfileView.vue -S`
- `rg -n "const .*LabelKeys = \\[|resolveShellLabelOverrides\\(" frontend/src/views/user frontend/src/components/payment -S --glob '!**/__tests__/**'`

### Notes
- Runtime behavior is unchanged; `ProfileView` still reads `profile_shell_config` from Sub2API public settings/fetched public settings.
- `KeysView` is now the only remaining user/payment page with a local shell label key list and direct `resolveShellLabelOverrides` usage.

## 2026-06-19 API keys shell schema centralized

### Done
- Added `apiKeysShell.ts` to own the API Keys page label key schema, shell resolution, and placeholder rendering.
- Updated `KeysView` to consume `resolveApiKeysShellLabels` and `renderApiKeysShellText` instead of keeping a page-local schema and direct `resolveShellLabelOverrides` usage.
- Updated `KeysView.shell.spec.ts` to guard against local schema/parser ownership returning.
- Added focused helper tests for API Keys shell label filtering, interpolation, and schema ownership.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiKeysShell.spec.ts src/views/user/__tests__/KeysView.shell.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/apiKeysShell.ts frontend/src/utils/__tests__/apiKeysShell.spec.ts frontend/src/views/user/KeysView.vue frontend/src/views/user/__tests__/KeysView.shell.spec.ts progress.md`
- `rg -n "const .*LabelKeys = \\[|resolveShellLabelOverrides\\(|resolvePaymentShellLabels\\(" frontend/src/views/user frontend/src/components/payment -S --glob '!**/__tests__/**'`
- `rg -n "apiKeysShellLabelKeys|resolveShellLabelOverrides\\(" frontend/src/views/user/KeysView.vue frontend/src/views/user/__tests__/KeysView.shell.spec.ts frontend/src/utils/apiKeysShell.ts -S`

### Notes
- Runtime behavior is unchanged; `KeysView` still reads `api_keys_shell_config` from Sub2API public settings/cache.
- User/payment page shell label schemas are now centralized in shared utils rather than owned by the page components.

## 2026-06-19 Key usage shell schema centralized

### Done
- Added `keyUsageShell.ts` to own the public Key Usage page label key schema, shell resolution, and placeholder rendering.
- Updated `KeyUsageView` to consume `resolveKeyUsageShellLabels` and `renderKeyUsageShellText` instead of keeping a page-local label schema and direct localized resolver usage.
- Updated `KeyUsageView` source guards so the page cannot regain local shell parsing/schema ownership.
- Added focused helper tests for Key Usage shell label filtering, interpolation, and schema ownership.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/keyUsageShell.spec.ts src/views/__tests__/KeyUsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/keyUsageShell.ts frontend/src/utils/__tests__/keyUsageShell.spec.ts frontend/src/views/KeyUsageView.vue frontend/src/views/__tests__/KeyUsageView.spec.ts progress.md`
- `rg -n "const .*LabelKeys = \\[|labelKeys = \\[|resolveShellLabelOverrides\\(|resolvePaymentShellLabels\\(|resolveLocalizedShellLabels\\(" frontend/src/views frontend/src/components -S --glob '!**/__tests__/**'`
- `rg -n "keyUsageShellLabelKeys|resolveLocalizedShellLabels\\(" frontend/src/views/KeyUsageView.vue frontend/src/views/__tests__/KeyUsageView.spec.ts frontend/src/utils/keyUsageShell.ts -S`

### Notes
- Runtime behavior is unchanged; `KeyUsageView` still reads `key_usage_shell_config` from Sub2API public settings.
- Page/component-level shell label schemas are now removed from `frontend/src/views` and `frontend/src/components`; remaining schema lists are shared helper modules such as `dashboardShellLabels.ts`.
- Renamed the OAuth compatibility router test suite wording from Touch-specific to generic legacy OAuth wording; the compatibility route behavior is unchanged.

## 2026-06-19 Runtime locale helper centralized

### Done
- Added `runtimeLocale.ts` to own ref/string runtime locale normalization and `zh`/`en` runtime language detection.
- Replaced duplicated `currentRuntimeLocale()` implementations across user/payment pages with `resolveRuntimeLocale`.
- Replaced the remaining direct `locale.value.startsWith('zh')` style runtime language checks in Airwallex checkout and API Keys shell wiring.
- Added focused helper tests for raw strings, ref-like locale objects, and missing locale fallback behavior.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/runtimeLocale.spec.ts src/views/user/__tests__/UsageView.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/views/user/__tests__/ApiTestView.spec.ts src/views/user/__tests__/AvailableGroupsView.spec.ts src/views/user/__tests__/AvailableChannelsView.spec.ts src/views/user/__tests__/RedeemView.spec.ts src/views/user/__tests__/AffiliateView.spec.ts src/views/user/__tests__/CustomPageView.spec.ts src/views/user/__tests__/ChannelStatusView.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts src/views/user/__tests__/ProfileView.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/StripePopupView.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/views/user/__tests__/SubscriptionsView.spec.ts src/views/user/__tests__/KeysView.shell.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/runtimeLocale.ts frontend/src/utils/__tests__/runtimeLocale.spec.ts frontend/src/views/user/StripePopupView.vue frontend/src/views/user/AirwallexPaymentView.vue frontend/src/views/user/AvailableGroupsView.vue frontend/src/views/user/RedeemView.vue frontend/src/views/user/PaymentView.vue frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/ApiTestView.vue frontend/src/views/user/PaymentResultView.vue frontend/src/views/user/AvailableChannelsView.vue frontend/src/views/user/ChannelStatusView.vue frontend/src/views/user/StripePaymentView.vue frontend/src/views/user/UserOrdersView.vue frontend/src/views/user/AffiliateView.vue frontend/src/views/user/CustomPageView.vue frontend/src/views/user/PaymentQRCodeView.vue frontend/src/views/user/ProfileView.vue frontend/src/views/user/SubscriptionsView.vue frontend/src/views/user/UsageView.vue frontend/src/views/user/KeysView.vue`
- `rg -n "function currentRuntimeLocale\\(|currentRuntimeLocale\\(\\)|toLowerCase\\(\\)\\.startsWith\\('zh'\\)|locale\\.value\\.startsWith\\('zh'\\)" frontend/src/views/user frontend/src/components -S --glob '!**/__tests__/**'`

### Notes
- Runtime behavior is unchanged; pages still consume Sub2API public settings for their shell config.
- This removes another class of local bootstrap/config parsing from the Touch-derived frontend shell, but the pages themselves still own UI state and interaction orchestration.

## 2026-06-19 Public shell locale parsing centralized

### Done
- Reused the shared runtime locale helper across more public/user/auth shell config entry points.
- Updated Home, Docs, Legal Document, Prompt Catalog, Models Plaza, Image Generator, Key Usage, Dashboard, AuthLayout, Login, and Register shell config language selection to avoid page-local `locale.value` / `startsWith('zh')` parsing.
- Kept actual formatting paths such as date, money, embedded URL locale, and validation-message separators unchanged.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/runtimeLocale.spec.ts src/views/__tests__/HomeView.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts src/views/__tests__/KeyUsageView.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts src/components/layout/__tests__/AuthLayout.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/HomeView.vue frontend/src/views/public/DocsView.vue frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/ImageGeneratorView.vue frontend/src/views/KeyUsageView.vue frontend/src/views/user/DashboardView.vue frontend/src/components/layout/AuthLayout.vue frontend/src/views/auth/LoginView.vue frontend/src/views/auth/RegisterView.vue progress.md`
- `rg -n "resolve.*Shell.*\\([^\\n]*locale\\.value|locale\\.value\\.startsWith\\('zh'\\)|locale\\.value\\.startsWith\\(\\\"zh\\\"\\)" frontend/src/views frontend/src/components -S --glob '!**/__tests__/**'`

### Notes
- The residual search hits are limited to admin `SettingsView` local preview/copy language selection, not public/user shell runtime settings.
- This continues shrinking Touch-derived frontend bootstrap/config parsing; the Vue frontend still owns the page interaction state and UI composition.

## 2026-06-19 Direct locale branch checks centralized

### Done
- Replaced the remaining direct `getLocale() === 'zh'` public/user page checks in Pricing and Credits with `resolveRuntimeLanguage`.
- Replaced direct auth-page email whitelist separator language checks in Register and Email Verify with `resolveRuntimeLanguage`.
- Replaced the remaining admin Settings payment documentation link language checks with `resolveRuntimeLanguage`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/runtimeLocale.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/user/__tests__/CreditsView.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts src/views/auth/__tests__/EmailVerifyView.spec.ts src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/public/PricingView.vue frontend/src/views/user/CreditsView.vue frontend/src/views/auth/RegisterView.vue frontend/src/views/auth/EmailVerifyView.vue frontend/src/views/admin/SettingsView.vue progress.md`
- `rg -n "getLocale\\(\\) === 'zh'|getLocale\\(\\) == 'zh'|locale\\.value\\.startsWith\\(['\\\"]zh['\\\"]\\)|String\\(locale\\.value \\|\\| ''\\)\\.toLowerCase\\(\\)\\.startsWith\\('zh'\\)|resolve.*Shell.*\\([^\\n]*locale\\.value" frontend/src/views frontend/src/components frontend/src/utils -S --glob '!**/__tests__/**'`

### Notes
- SettingsView tests still emit existing `router-link` mount warnings, but all assertions pass.
- The direct language-branch cleanup keeps the previous output characters and documentation URLs unchanged; only the locale normalization path moved to the shared helper.

## 2026-06-19 Profile child label schema centralized

### Done
- Removed local `*LabelKey` union ownership from Profile child components and reused the shared `ProfileLabelKey` / `ProfileViewShellLabels` contracts from `profileShell.ts`.
- Updated ProfileInfo provider-label interpolation to reuse `resolveAuthBindingProviderLabel`.
- Updated ProfileInfo text interpolation to reuse `interpolateProfileShellLabel`.
- Added a source-level regression guard in `profileShell.spec.ts` so Profile child components do not reintroduce local label-key unions.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/profileShell.spec.ts src/components/user/profile/__tests__/ProfileAvatarCard.spec.ts src/components/user/profile/__tests__/ProfileBalanceNotifyCard.spec.ts src/components/user/profile/__tests__/ProfileEditForm.spec.ts src/components/user/profile/__tests__/ProfileInfoCard.spec.ts src/components/user/profile/__tests__/ProfilePasswordForm.spec.ts src/components/user/profile/__tests__/ProfileTotpCard.spec.ts src/components/user/profile/__tests__/ProfileIdentityBindingsSection.spec.ts src/components/user/profile/__tests__/totp-timer-cleanup.spec.ts src/views/user/__tests__/ProfileView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/__tests__/profileShell.spec.ts frontend/src/components/user/profile/ProfileBalanceNotifyCard.vue frontend/src/components/user/profile/ProfileTotpCard.vue frontend/src/components/user/profile/ProfileEditForm.vue frontend/src/components/user/profile/ProfilePasswordForm.vue frontend/src/components/user/profile/ProfileAvatarCard.vue frontend/src/components/user/profile/TotpSetupModal.vue frontend/src/components/user/profile/TotpDisableDialog.vue frontend/src/components/user/profile/ProfileInfoCard.vue progress.md`
- `rg -n "type \\w*LabelKey\\s*=|interface \\w*Labels" frontend/src/components/user/profile -S --glob '!**/__tests__/**'`

### Notes
- Runtime behavior is unchanged; ProfileView still reads `profile_shell_config` from Sub2API public settings and passes the resolved labels down.
- This removes another component-level UI schema ownership point from the Touch-derived frontend shell; Profile interaction state still remains in Vue components.

## 2026-06-19 Payment child label schema centralized

### Done
- Moved `PaymentStatusPanelLabels` and `SubscriptionPlanCardLabels` ownership into `paymentShell.ts`.
- Updated PaymentStatusPanel to use shared `PaymentStatusPanelLabelKey` / `PaymentStatusPanelLabels` and `renderPaymentStatusPanelText`.
- Updated SubscriptionPlanCard to consume shared `SubscriptionPlanCardLabels`.
- Added payment shell tests for the new shared component label contracts and source guards preventing local component label interfaces from returning.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/components/payment/__tests__/PaymentStatusPanel.spec.ts src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/paymentShell.ts frontend/src/utils/__tests__/paymentShell.spec.ts frontend/src/components/payment/PaymentStatusPanel.vue frontend/src/components/payment/SubscriptionPlanCard.vue progress.md`
- `rg -n "interface \\w*Labels|type \\w*LabelKey\\s*=" frontend/src/components/payment -S --glob '!**/__tests__/**'`

### Notes
- Runtime behavior is unchanged; PaymentView still resolves `payment_shell_config` from Sub2API public settings and passes labels down.
- Payment components no longer own local label schema definitions; their shell copy contracts now live in `paymentShell.ts`.

## 2026-06-19 Payment and profile child label prop types centralized

### Done
- Replaced remaining payment child component `Partial<Record<...LabelKey, string>>` label props with shared `PaymentQRDialogLabels`, `StripeInlineLabels`, and `OrderTableLabels` from `paymentShell.ts`.
- Added shared `ProfileLabels` in `profileShell.ts` and updated Profile child components to consume that type instead of repeating the label prop shape.
- Extended payment/profile source-level regression guards so child components do not reintroduce local label prop schema expressions.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/components/payment/__tests__/PaymentQRDialog.spec.ts src/components/payment/__tests__/StripePaymentInline.spec.ts src/components/payment/__tests__/OrderTable.spec.ts src/components/payment/__tests__/PaymentStatusPanel.spec.ts src/components/payment/__tests__/SubscriptionPlanCard.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/profileShell.spec.ts src/components/user/profile/__tests__/ProfileAvatarCard.spec.ts src/components/user/profile/__tests__/ProfileBalanceNotifyCard.spec.ts src/components/user/profile/__tests__/ProfileEditForm.spec.ts src/components/user/profile/__tests__/ProfilePasswordForm.spec.ts src/components/user/profile/__tests__/ProfileTotpCard.spec.ts src/components/user/profile/__tests__/totp-timer-cleanup.spec.ts src/views/user/__tests__/ProfileView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "Partial<Record<\\w*LabelKey,\\s*string>>|interface \\w*Labels|type \\w*LabelKey\\s*=" frontend/src/components/payment -S --glob '!**/__tests__/**'`
- `rg -n "Partial<Record<ProfileLabelKey,\\s*string>>|type \\w*LabelKey\\s*=" frontend/src/components/user/profile -S --glob '!**/__tests__/**'`

### Notes
- Runtime behavior is unchanged; parent views still provide payment/profile shell labels from Sub2API public settings.
- This removes another batch of Touch-derived component-level shell contracts, but the Vue frontend still owns payment/profile interaction state.

## 2026-06-19 API Keys modal shell copy delegated to public settings

### Done
- Added Use Key modal label keys to the shared `apiKeysShell.ts` schema so modal copy is part of `api_keys_shell_config`.
- Removed the UseKeyModal local i18n fallback map and `vue-i18n` dependency; modal text now renders only from the parent-provided API Keys shell labels.
- Tightened API Keys shell label types from broad `Record<string, string>` to `Partial<Record<ApiKeysShellLabelKey, string>>`.
- Updated KeysView to consume shared `ApiKeysShellLabels` / `ApiKeysShellLabelKey` types and kept unknown status text as raw status instead of sending arbitrary keys through the shell renderer.
- Added source-level regression guards so UseKeyModal cannot reintroduce local fallback copy and KeysView cannot revert to broad shell label typing.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiKeysShell.spec.ts src/components/keys/__tests__/UseKeyModal.spec.ts src/views/user/__tests__/KeysView.shell.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/apiKeysShell.ts frontend/src/utils/__tests__/apiKeysShell.spec.ts frontend/src/components/keys/UseKeyModal.vue frontend/src/components/keys/__tests__/UseKeyModal.spec.ts frontend/src/views/user/KeysView.vue frontend/src/views/user/__tests__/KeysView.shell.spec.ts`
- `rg -n "USE_KEY_MODAL_I18N_KEYS|shellLabels\\?: Record<string, string>|from 'vue-i18n'|\\bt\\(" frontend/src/components/keys/UseKeyModal.vue -S`
- `rg -n "computed<Record<string, string>>|const apiKeysText = \\(key: string" frontend/src/views/user/KeysView.vue -S`

### Notes
- Runtime behavior now depends on Sub2API default/public `api_keys_shell_config` for UseKeyModal visible copy; backend defaults already include the modal labels.
- This removes another Touch-derived frontend-local copy fallback, but API Keys page state and modal layout still remain in Vue.

## 2026-06-19 Key Usage runtime locale checks centralized

### Done
- Replaced the remaining direct `locale.value === 'zh'` checks in `KeyUsageView` with the shared runtime locale helpers.
- Added `resolveRuntimeLocale(locale)` usage for date formatting so locale-code resolution no longer reads the i18n ref directly.
- Extended the Key Usage source regression test to prevent direct `locale.value` language checks from returning.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/runtimeLocale.spec.ts src/views/__tests__/KeyUsageView.spec.ts src/utils/__tests__/keyUsageShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "locale\\.value === ['\\\"]zh['\\\"]|locale\\.value\\.startsWith\\(['\\\"]zh|String\\(locale\\.value" frontend/src/views/KeyUsageView.vue -S`

### Notes
- Runtime behavior is unchanged; Key Usage still reads visible shell labels from Sub2API `key_usage_shell_config`.
- This removes another frontend-local bootstrap/locale branch from the Touch-derived public page shell.

## 2026-06-20 API Keys endpoint popover shell copy delegated to public settings

### Done
- Added endpoint popover label keys to the shared `apiKeysShell.ts` schema so endpoint copy belongs to `api_keys_shell_config`.
- Removed the EndpointPopover direct `vue-i18n` dependency and `keys.endpoints.*` local copy reads.
- Passed `apiKeysShellLabels` from KeysView into EndpointPopover, reusing the same Sub2API public settings-backed label source as the rest of the API Keys page.
- Extended Sub2API backend default `api_keys_shell_config` with endpoint popover labels for zh/en.
- Added regression guards for the endpoint popover and backend default public settings payload.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiKeysShell.spec.ts src/components/keys/__tests__/EndpointPopover.spec.ts src/views/user/__tests__/KeysView.shell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAPIKeysShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/apiKeysShell.ts frontend/src/utils/__tests__/apiKeysShell.spec.ts frontend/src/components/keys/EndpointPopover.vue frontend/src/components/keys/__tests__/EndpointPopover.spec.ts frontend/src/views/user/KeysView.vue frontend/src/views/user/__tests__/KeysView.shell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`
- `rg -n "from 'vue-i18n'|keys\\.endpoints\\.|shellLabels\\?: Record<string, string>" frontend/src/components/keys/EndpointPopover.vue frontend/src/utils/apiKeysShell.ts frontend/src/views/user/KeysView.vue -S`

### Notes
- Runtime behavior is unchanged when defaults are used, but EndpointPopover visible copy is now configurable through Sub2API public/admin settings.
- API Keys page layout and interaction state still remain in Vue; this only removes another local copy ownership point from the Touch-derived frontend shell.

## 2026-06-20 Payment method labels typed through payment shell

### Done
- Added a shared `PaymentMethodLabelKey` / `PaymentMethodLabels` contract to `paymentShell.ts`.
- Updated PaymentMethodSelector to consume the shared label type instead of accepting arbitrary `Record<string, string>` labels.
- Normalized known payment method variants such as `wxpay_direct` to the shared `wxpay` label while preserving unknown method types as raw display text.
- Updated PaymentView so `paymentMethodLabels` is explicitly typed as `PaymentMethodLabels` and still derives labels from `payment_shell_config`.
- Added source-level regression guards to prevent the selector from returning to broad local label maps.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/components/payment/__tests__/PaymentMethodSelector.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/paymentShell.ts frontend/src/utils/__tests__/paymentShell.spec.ts frontend/src/components/payment/PaymentMethodSelector.vue frontend/src/components/payment/__tests__/PaymentMethodSelector.spec.ts frontend/src/views/user/PaymentView.vue progress.md`
- `rg -n "methodLabels\\?: Record<string, string>|computed<Record<string, string>>|paymentMethodLabels = computed\\(\\)|from 'vue-i18n'|t\\('payment\\." frontend/src/components/payment/PaymentMethodSelector.vue frontend/src/views/user/PaymentView.vue frontend/src/utils/paymentShell.ts -S`

### Notes
- Runtime labels still come from Sub2API `payment_shell_config`; this change tightens the frontend component contract and removes another arbitrary local label-map boundary.
- PaymentView still imports `vue-i18n` for broader payment-page locale and error translation behavior; that path was intentionally left unchanged.

## 2026-06-20 API Guide status label moved to shell config

### Done
- Added `status` to the shared API Guide shell label schema.
- Replaced the last local `t('common.status')` label in ApiGuideView with `apiGuideText('status')`.
- Extended the Sub2API default `api_guide_shell_config` with zh/en status labels.
- Added frontend and backend regression coverage so the status label remains controlled by public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiGuideShell.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAPIGuideShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/apiGuideShell.ts frontend/src/utils/__tests__/apiGuideShell.spec.ts frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/__tests__/ApiGuideView.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`
- `rg -n "t\\('common\\.status'\\)|apiGuide\\.status|status.*common" frontend/src/views/user/ApiGuideView.vue frontend/src/utils/apiGuideShell.ts frontend/src/views/user/__tests__/ApiGuideView.spec.ts -S`

### Notes
- Runtime behavior is unchanged with backend defaults, but the status field label is now configurable through Sub2API `api_guide_shell_config`.
- ApiGuideView still uses i18n for gateway protocol/platform labels; this step only removes the remaining generic local status label from the page shell.

## 2026-06-20 API Test loading labels moved to shell config

### Done
- Added `loading` and `noOptionsFound` to the shared API Test shell label schema.
- Replaced ApiTestView's model selector loading/empty text from `common.loading` / `common.noOptionsFound` with `apiTestText(...)`.
- Extended the Sub2API default `api_test_shell_config` with zh/en loading and empty-option labels.
- Added frontend and backend regression coverage so these labels remain controlled by public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiTestShell.spec.ts src/views/user/__tests__/ApiTestView.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAPITestShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/apiTestShell.ts frontend/src/utils/__tests__/apiTestShell.spec.ts frontend/src/views/user/ApiTestView.vue frontend/src/views/user/__tests__/ApiTestView.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`
- `rg -n "t\\('common\\.(loading|noOptionsFound)'\\)|apiTest\\.(loading|noOptionsFound)|loading.*common|noOptionsFound.*common" frontend/src/views/user/ApiTestView.vue frontend/src/utils/apiTestShell.ts frontend/src/views/user/__tests__/ApiTestView.spec.ts -S`

### Notes
- Runtime behavior is unchanged with defaults, but the model selector's loading/empty labels now come from Sub2API `api_test_shell_config`.
- ApiTestView still uses i18n for gateway platform/variant labels and unknown-error fallback; those remain separate follow-up candidates.

## 2026-06-20 API Test unknown-error fallback moved to shell config

### Done
- Added `unknownError` to the shared API Test shell label schema.
- Replaced ApiTestView's non-Error exception fallback from `common.unknownError` with `apiTestText('unknownError')`.
- Extended the Sub2API default `api_test_shell_config` with zh/en unknown-error labels.
- Added frontend and backend regression coverage so the fallback remains controlled by public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiTestShell.spec.ts src/views/user/__tests__/ApiTestView.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAPITestShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/apiTestShell.ts frontend/src/utils/__tests__/apiTestShell.spec.ts frontend/src/views/user/ApiTestView.vue frontend/src/views/user/__tests__/ApiTestView.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`
- `rg -n "t\\('common\\.(loading|noOptionsFound|unknownError)'\\)|apiTest\\.(loading|noOptionsFound|unknownError)|(loading|noOptionsFound|unknownError).*common" frontend/src/views/user/ApiTestView.vue frontend/src/utils/apiTestShell.ts frontend/src/views/user/__tests__/ApiTestView.spec.ts -S`

### Notes
- Normal Error instances still display their actual error message; this only moves the last fallback string into Sub2API `api_test_shell_config`.
- ApiTestView still uses i18n for gateway platform and variant labels, which are shared gateway taxonomy labels rather than single-page shell copy.

## 2026-06-20 Auth shell labels typed through shared schema

### Done
- Added an explicit `AuthShellLabelKey` / `AuthShellLabels` schema to the auth shell parser.
- Filtered `auth_shell_config` labels to known keys so arbitrary public-settings fields do not become frontend label surface.
- Updated AuthLayout, LoginView, and RegisterView to call `authText` with typed auth shell keys.
- Added regression guards to keep auth shell consumers from drifting back to broad `Record<string, string>` maps.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/components/layout/__tests__/AuthLayout.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "AuthShellLabels = Record<string, string>|ref<Record<string, string>>|authText\\(key: string|renderAuthShellText\\([^\\n]*,\\s*key:\\s*string" frontend/src/utils/authShell.ts frontend/src/components/layout/AuthLayout.vue frontend/src/views/auth/LoginView.vue frontend/src/views/auth/RegisterView.vue -S`

### Notes
- Runtime copy still comes from Sub2API `auth_shell_config`; this step tightens the frontend contract and removes another loose Touch-derived shell boundary.
- Form validation and agreement messages still use existing i18n/error helpers; this change only targets configurable auth shell labels.

## 2026-06-20 Register optional label moved to auth shell config

### Done
- Added `optional` to the shared auth shell label schema.
- Replaced RegisterView's promo-code optional label from local `common.optional` with `authText('optional')`.
- Extended Sub2API default `auth_shell_config` with zh/en optional labels.
- Added frontend and backend regression coverage so the optional label remains controlled by public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/components/layout/__tests__/AuthLayout.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts frontend/src/components/layout/AuthLayout.vue frontend/src/components/layout/__tests__/AuthLayout.spec.ts frontend/src/views/auth/LoginView.vue frontend/src/views/auth/__tests__/LoginView.turnstile.spec.ts frontend/src/views/auth/RegisterView.vue frontend/src/views/auth/__tests__/RegisterView.auth-shell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`
- `rg -n "AuthShellLabels = Record<string, string>|ref<Record<string, string>>|authText\\(key: string|renderAuthShellText\\([^\\n]*,\\s*key:\\s*string|t\\('common\\.optional'\\)" frontend/src/utils/authShell.ts frontend/src/components/layout/AuthLayout.vue frontend/src/views/auth/LoginView.vue frontend/src/views/auth/RegisterView.vue -S`

### Notes
- OAuth callback optional labels were handled in the next milestone; broader auth-flow copy still remains a separate follow-up candidate.

## 2026-06-20 OAuth callback status labels moved to auth shell config

### Done
- Replaced OAuthCallbackView's password optional labels from local `common.optional` with `authText('optional')`.
- Replaced OAuthCallbackView's submit processing label from local `common.processing` with `authText('processing')`.
- Reused the shared auth shell parser and `auth_shell_config` key schema instead of adding callback-specific copy.
- Loaded public auth shell labels asynchronously so OAuth callback token handling and provider redirects are not blocked by public settings fetch latency.
- Added OAuth callback regression coverage for configured optional/processing labels.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/OAuthCallbackView.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/__tests__/OAuthCallbackView.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`
- `rg -n "t\\('common\\.(optional|processing)'\\)|common\\.(optional|processing)" frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/RegisterView.vue -S`

### Notes
- The OAuth callback page still owns broader callback/debug UI and other auth-flow copy through i18n; this step only removes the remaining optional/processing common-label dependency from this page.

## 2026-06-20 Auth provider callback processing labels moved to auth shell config

### Done
- Added `useAuthShellText` so auth callback pages can share `auth_shell_config` label loading and rendering.
- Reused that composable in OAuthCallbackView instead of keeping a page-local auth shell loader.
- Replaced OIDC, LinuxDo, DingTalk, and WeChat callback submit processing labels from local `common.processing` with `authText('processing')`.
- Replaced PendingOAuthCreateAccountForm's submit processing label with `authText('processing')`.
- Added source-level regression coverage across callback views and the pending OAuth create-account form.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/CallbackAuthShell.spec.ts src/components/auth/__tests__/PendingOAuthCreateAccountForm.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts src/views/auth/__tests__/OAuthCallbackView.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "t\\('common\\.processing'\\)|common\\.processing" frontend/src/views/auth frontend/src/components/auth -S`

### Notes
- Runtime code is now clear of `common.processing` in auth views/components; remaining matches are regression assertions.
- Provider-specific callback titles, hints, and action labels still use existing i18n and remain follow-up candidates if the goal is to make all auth-flow copy admin-configurable.

## 2026-06-20 Pending OAuth create-account form copy moved to auth shell config

### Done
- Added pending OAuth create-account form labels to the shared auth shell schema: send-code, countdown, code-sent, and verification hint.
- Extended Sub2API default `auth_shell_config` with zh/en values for those labels.
- Replaced PendingOAuthCreateAccountForm placeholders, send-code labels, submit label, and switch-to-bind label with `authText(...)`.
- Updated frontend and backend regression tests so the form copy remains controlled by public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/auth/__tests__/PendingOAuthCreateAccountForm.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "t\\('auth\\.(emailPlaceholder|passwordPlaceholder|sendingCode|resendCountdown|sendCode|codeSentSuccess|verificationCodeHint|invitationCodePlaceholder|createAccount|alreadyHaveAccount)'\\)" frontend/src/components/auth/PendingOAuthCreateAccountForm.vue -S`

### Notes
- The component still uses i18n for validation/error fallback strings such as Turnstile and send-code failure messages; visible form shell copy now comes from Sub2API public settings.

## 2026-06-20 OAuth entry button copy moved to auth shell config

### Done
- Added `signInWithProvider` to the shared auth shell schema and Sub2API default `auth_shell_config`.
- Updated EmailOAuthButtons, OidcOAuthSection, LinuxDoOAuthSection, DingTalkOAuthSection, and WechatOAuthSection to render provider login buttons from `auth_shell_config`.
- Updated the OAuth divider in those components to use the existing `oauthAlternativeMethods` auth shell label.
- Passed Login/Register `authShellLabels` into all OAuth entry components instead of letting those components own public settings loading.
- Added regression coverage so OAuth entry components do not drift back to local auth i18n keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/auth/__tests__/EmailOAuthButtons.spec.ts src/components/auth/__tests__/WechatOAuthSection.spec.ts src/components/auth/__tests__/OidcOAuthSection.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "t\\('auth\\.(oauthOrContinue|emailOAuth\\.signIn|oidc\\.signIn|linuxdo\\.signIn|dingtalk\\.signIn)'\\)|auth\\.oauthOrContinue|auth\\.emailOAuth\\.signIn|auth\\.oidc\\.signIn|auth\\.linuxdo\\.signIn|auth\\.dingtalk\\.signIn" frontend/src/components/auth/EmailOAuthButtons.vue frontend/src/components/auth/OidcOAuthSection.vue frontend/src/components/auth/LinuxDoOAuthSection.vue frontend/src/components/auth/DingTalkOAuthSection.vue frontend/src/components/auth/WechatOAuthSection.vue -S`

### Notes
- WeChat unavailable-reason hints still use existing i18n because they are capability/error state copy rather than the shared OAuth entry shell.

## 2026-06-20 TOTP login modal copy moved to auth shell config

### Done
- Added TOTP login modal labels to the shared auth shell schema: title, hint, verifying, and cancel.
- Extended Sub2API default `auth_shell_config` with zh/en TOTP login modal labels.
- Updated TotpLoginModal to consume parent-provided auth shell labels instead of local `profile.totp.*` and `common.*` i18n keys.
- Passed LoginView `authShellLabels` into TotpLoginModal.
- Added regression coverage for configured TOTP login modal copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/auth/__tests__/TotpLoginModal.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "t\\('profile\\.totp\\.loginTitle'\\)|t\\('profile\\.totp\\.loginHint'\\)|t\\('common\\.(verifying|cancel)'\\)|useI18n" frontend/src/components/auth/TotpLoginModal.vue -S`

### Notes
- Login error handling and validation text still use existing i18n/error helpers; this step only moves the visible TOTP modal shell copy.

## 2026-06-20 Login agreement prompt copy moved to auth shell config

### Done
- Added login agreement prompt labels to the shared auth shell schema, including checkbox text, modal titles, descriptions, date template, and accept/reject actions.
- Extended Sub2API default `auth_shell_config` with zh/en login agreement prompt labels.
- Updated LoginAgreementPrompt to consume parent-provided auth shell labels instead of local auth i18n keys.
- Passed Login/Register `authShellLabels` into LoginAgreementPrompt.
- Added regression coverage for configured login agreement prompt copy and backend default public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/auth/__tests__/LoginAgreementPrompt.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/components/auth/LoginAgreementPrompt.vue frontend/src/components/auth/__tests__/LoginAgreementPrompt.spec.ts frontend/src/views/auth/LoginView.vue frontend/src/views/auth/RegisterView.vue frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`
- `rg -n "auth\\.loginAgreementPrompt|useI18n" frontend/src/components/auth/LoginAgreementPrompt.vue -S`

### Notes
- Agreement acceptance/rejection state and document routing remain unchanged; only prompt shell copy moved to Sub2API public settings.

## 2026-06-20 Auth popup hint moved to auth shell config

### Done
- Added `oauthCallbackHint` to the shared auth shell schema.
- Extended Sub2API default `auth_shell_config` with zh/en OAuth popup redirect hint copy.
- Updated AuthPopupView to render the hint from public auth shell settings via the shared composable.
- Updated AuthPopupView tests to mock public settings instead of local i18n.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/AuthPopupView.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "auth\\.oauth\\.callbackHint|useI18n|\\bt\\('" frontend/src/views/auth/AuthPopupView.vue -S`
- `git diff --check -- frontend/src/views/auth/AuthPopupView.vue frontend/src/views/auth/__tests__/AuthPopupView.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- Redirect sanitization and login handoff behavior remain unchanged; only the transient popup hint copy moved to Sub2API public settings.

## 2026-06-20 Forgot password page shell copy moved to auth shell config

### Done
- Added forgot-password page labels to the shared auth shell schema: title, hint, success state, reset action, footer, and back-to-login labels.
- Extended Sub2API default `auth_shell_config` with zh/en forgot-password labels.
- Updated ForgotPasswordView to parse auth shell labels from its existing public settings load instead of adding another request.
- Replaced the visible forgot-password page shell copy and success toast with auth shell labels.
- Added regression coverage for configured forgot-password page copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/ForgotPasswordView.auth-shell.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "auth\\.(forgotPasswordTitle|forgotPasswordHint|resetEmailSent|resetEmailSentHint|backToLogin|emailLabel|emailPlaceholder|sendingResetLink|sendResetLink|rememberedPassword|signIn)'" frontend/src/views/auth/ForgotPasswordView.vue -S`
- `git diff --check -- frontend/src/views/auth/ForgotPasswordView.vue frontend/src/views/auth/__tests__/ForgotPasswordView.auth-shell.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- Validation errors and failure fallback messages still use existing i18n/error helpers; this step only moves the page shell and success copy.

## 2026-06-20 Reset password page shell copy moved to auth shell config

### Done
- Added reset-password page labels to the shared auth shell schema: title, hint, invalid-link state, success state, password labels/placeholders, submit labels, and request-new-link action.
- Extended Sub2API default `auth_shell_config` with zh/en reset-password labels.
- Updated ResetPasswordView to load auth shell labels through the shared composable and render page shell copy from public settings.
- Replaced invalid-link and password-reset success toast copy with auth shell labels.
- Added regression coverage for configured reset-password page copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/ResetPasswordView.auth-shell.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "auth\\.(resetPasswordTitle|resetPasswordHint|invalidResetLink|invalidResetLinkHint|requestNewResetLink|passwordResetSuccess|passwordResetSuccessHint|emailLabel|newPassword|newPasswordPlaceholder|confirmPassword|confirmPasswordPlaceholder|resettingPassword|resetPassword|rememberedPassword|signIn)'" frontend/src/views/auth/ResetPasswordView.vue -S`
- `git diff --check -- frontend/src/views/auth/ResetPasswordView.vue frontend/src/views/auth/__tests__/ResetPasswordView.auth-shell.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- Password validation errors and reset failure fallback messages still use existing i18n/error helpers; this step only moves the page shell and success/invalid-link copy.

## 2026-06-20 Email verification page shell copy moved to auth shell config

### Done
- Added email verification page labels to the shared auth shell schema with an `emailVerify*` prefix to avoid reusing semantically different generic verification labels.
- Extended Sub2API default `auth_shell_config` with zh/en email verification labels.
- Updated EmailVerifyView to parse auth shell labels from its existing public settings response.
- Replaced page title, destination hint, expired-session copy, code label/hint, sent-state copy, submit/resend actions, and back-to-registration copy with auth shell labels.
- Added regression coverage for configured email verification page copy while keeping existing registration and pending OAuth flow tests intact.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/EmailVerifyView.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "auth\\.(verifyYourEmail|sendCodeDesc|sessionExpired|sessionExpiredDesc|verificationCode|verificationCodeHint|codeSentSuccess|verifying|verifyAndCreate|resendCountdown|sendingCode|clickToResend|resendCode|backToRegistration)'" frontend/src/views/auth/EmailVerifyView.vue -S`
- `git diff --check -- frontend/src/views/auth/EmailVerifyView.vue frontend/src/views/auth/__tests__/EmailVerifyView.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- Validation errors, email-suffix messages, and backend failure fallbacks still use existing i18n/error helpers; this step only moves the page shell and resend/status copy.

## 2026-06-20 OAuth callback page shell copy moved to auth shell config

### Done
- Added OAuth callback shell labels to the shared auth shell schema, including callback title/hint, invalid callback state, manual copy field labels, registration completion labels, and optional password hint.
- Extended Sub2API default `auth_shell_config` with zh/en OAuth callback labels.
- Updated OAuthCallbackView to render safe page shell copy from auth shell labels instead of local auth i18n keys.
- Kept error handling, login success messages, password validation errors, and copy button text on their existing helpers/i18n paths.
- Updated OAuthCallbackView regression tests to assert configured public settings copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/OAuthCallbackView.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "auth\\.(oauth\\.callbackTitle|oauth\\.callbackHint|oauth\\.invalidCallbackTitle|oauth\\.invalidCallbackHint|oauth\\.code|oauth\\.state|oauth\\.fullUrl|emailOAuth\\.passwordOptionalHint|oidc\\.completeRegistration|oidc\\.invitationRequired|emailLabel|passwordLabel|createPasswordPlaceholder|confirmPassword|confirmPasswordPlaceholder|invitationCodeLabel|invitationCodePlaceholder|affiliateInvitationDetected|backToLogin)'" frontend/src/views/auth/OAuthCallbackView.vue -S`
- `git diff --check -- frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/__tests__/OAuthCallbackView.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- Provider-specific callback pages (OIDC/LinuxDo/DingTalk/WeChat) still have their own local flow copy and remain follow-up candidates.

## 2026-06-20 OIDC callback shared OAuth flow copy moved to auth shell config

### Done
- Added shared OAuth flow labels to the auth shell schema: profile adoption, account action chooser, create-account hint, bind-login hint/actions, TOTP hint/action, and fallback account label.
- Extended Sub2API default `auth_shell_config` with zh/en shared OAuth flow labels.
- Updated OidcCallbackView to render those common flow labels from auth shell settings.
- Left provider-specific callback title/processing/invitation strings and error handling on existing i18n/error paths.
- Updated OIDC callback tests to assert configured public settings copy across adoption, chooser, bind-login, and TOTP states.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/OidcCallbackView.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "auth\\.oauthFlow\\.(profileDetailsTitle|profileDetailsDescription|useDisplayName|avatarAlt|useAvatar|reviewProfileBeforeContinue|chooseHowToContinue|suggestedEmail|chooseAccountActionHint|bindExistingAccount|createNewAccount|createAccountHint|bindLoginHint|logInAndBind|useDifferentEmail|totpHint|yourAccount|verifyAndContinue)'" frontend/src/views/auth/OidcCallbackView.vue -S`
- `git diff --check -- frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/__tests__/OidcCallbackView.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- LinuxDo, DingTalk, and WeChat callback views still need the same shared OAuth flow label migration; the shared schema/defaults are now available for reuse.

## 2026-06-20 LinuxDo and DingTalk shared OAuth flow copy moved to auth shell config

### Done
- Reused the shared `oauthFlow*` auth shell labels added for OIDC.
- Updated LinuxDoCallbackView and DingTalkCallbackView to render shared OAuth flow shell copy from auth shell settings.
- Migrated profile adoption, account chooser, create-account hint, bind-login hint/actions, TOTP hint/action, and fallback account label.
- Added source-level regression coverage for OIDC/LinuxDo/DingTalk shared OAuth flow labels.
- Updated LinuxDo callback tests to provide auth shell config and assert configured chooser/TOTP copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/CallbackAuthShell.spec.ts src/views/auth/__tests__/LinuxDoCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "auth\\.oauthFlow\\.(profileDetailsTitle|profileDetailsDescription|useDisplayName|avatarAlt|useAvatar|reviewProfileBeforeContinue|chooseHowToContinue|suggestedEmail|chooseAccountActionHint|bindExistingAccount|createNewAccount|createAccountHint|bindLoginHint|logInAndBind|useDifferentEmail|totpHint|yourAccount|verifyAndContinue)'" frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue -S`
- `git diff --check -- frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/__tests__/CallbackAuthShell.spec.ts frontend/src/views/auth/__tests__/LinuxDoCallbackView.spec.ts progress.md`

### Notes
- DingTalk currently has source-level and typecheck coverage, but no dedicated callback spec file. WeChat still has additional auth-flow branches and remains a follow-up migration target.

## 2026-06-20 WeChat callback OAuth flow copy moved to auth shell config

### Done
- Added WeChat/current-account binding labels to the shared auth shell schema.
- Extended Sub2API default `auth_shell_config` with zh/en labels for current-account binding and sign-in-then-bind guidance.
- Updated WechatCallbackView to render shared OAuth flow copy, current-account binding copy, placeholders, and common actions from auth shell settings.
- Kept WeChat capability/unavailable error messages on existing i18n/error paths.
- Expanded CallbackAuthShell source guards to include WeChat.
- Updated WeChat callback tests to provide auth shell config and assert configured adoption, chooser, existing-account, current-account, and TOTP copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/WechatCallbackView.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "auth\\.oauthFlow\\.(profileDetailsTitle|profileDetailsDescription|useDisplayName|avatarAlt|useAvatar|bindCurrentAccount|bindCurrentAccountDescription|bindCurrentAccountTitle|bindSignInToExistingAccount|signInThenBindDescription|reviewProfileBeforeContinue|chooseHowToContinue|chooseAccountActionHint|bindExistingAccount|createNewAccount|createAccountHint|logInAndBind|totpHint|yourAccount|verifyAndContinue)'|auth\\.(alreadyHaveAccount|emailPlaceholder|passwordPlaceholder|signIn|continue)'" frontend/src/views/auth/WechatCallbackView.vue -S`
- `git diff --check -- frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/__tests__/WechatCallbackView.spec.ts frontend/src/views/auth/__tests__/CallbackAuthShell.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- Provider callback common flow copy is now covered for OIDC, LinuxDo, DingTalk, and WeChat. Remaining provider-specific callback title/processing/invitation/error copy is still local i18n.

## 2026-06-20 Provider callback shell copy moved to auth shell config

### Done
- Added generic provider callback labels to the auth shell schema for callback title, processing hint, fallback hint, invitation-required copy, and registration submit states.
- Extended Sub2API default `auth_shell_config` with zh/en provider callback labels using `{providerName}`.
- Updated OIDC, LinuxDo, DingTalk, and WeChat callback views to render provider callback title/processing/invitation/register shell copy from public auth shell settings.
- Replaced local invitation-code placeholder and continue button copy in those callback shells with auth shell labels.
- Added source-level regression coverage to keep provider callback shell copy out of local Touch i18n.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/CallbackAuthShell.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "t\\('auth\\.(oidc|linuxdo|dingtalk)\\.(callbackTitle|callbackProcessing|callbackHint|invitationRequired|completing|completeRegistration)'\\)|auth\\.invitationCodePlaceholder|t\\('auth\\.continue'\\)" frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue -S`
- `git diff --check -- frontend/src/utils/authShell.ts frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/__tests__/CallbackAuthShell.spec.ts frontend/src/views/auth/__tests__/OidcCallbackView.spec.ts frontend/src/views/auth/__tests__/LinuxDoCallbackView.spec.ts frontend/src/views/auth/__tests__/WechatCallbackView.spec.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Provider callback error/failure messages still use existing i18n/error paths; this step only moved safe page shell copy and common action labels.

## 2026-06-20 DingTalk email completion shell copy moved to auth shell config

### Done
- Added `oauthFlowCreateAccountTitle` and `dingtalkProviderName` to the shared auth shell schema.
- Extended Sub2API default `auth_shell_config` with zh/en DingTalk email-completion page shell labels.
- Updated DingTalkEmailCompletionView to render the create-account title and description from auth shell settings.
- Kept login success, registration-disabled guidance, and failure fallbacks on their existing i18n/error paths.
- Added source-level regression coverage so the DingTalk email-completion shell does not return to local Touch auth i18n.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/CallbackAuthShell.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "t\\('auth\\.dingtalk\\.createAccountTitle'\\)|t\\('auth\\.oauthFlow\\.createAccountHint'\\)" frontend/src/views/auth/DingTalkEmailCompletionView.vue -S`
- `git diff --check -- frontend/src/views/auth/DingTalkEmailCompletionView.vue frontend/src/views/auth/__tests__/CallbackAuthShell.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- This is another UI-shell-only migration; DingTalk business and error messages remain on existing localized error paths.

## 2026-06-20 WeChat payment callback shell copy moved to payment shell config

### Done
- Added WeChat payment callback labels to the payment shell contract.
- Extended Sub2API default `payment_shell_config` with zh/en labels for callback title, processing state, back-to-payment action, and missing resume-token message.
- Updated WechatPaymentCallbackView to read visible callback labels from public payment shell settings instead of local auth i18n.
- Added regression coverage for configured callback labels and source guards against `auth.wechatPayment.*` usage.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/WechatPaymentCallbackView.spec.ts src/utils/__tests__/paymentShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsPaymentShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "auth\\.wechatPayment\\.(callbackTitle|callbackProcessing|backToPayment|callbackMissingResumeToken)|t\\('auth\\.wechatPayment" frontend/src/views/auth/WechatPaymentCallbackView.vue -S`
- `git diff --check -- frontend/src/views/auth/WechatPaymentCallbackView.vue frontend/src/views/auth/__tests__/WechatPaymentCallbackView.spec.ts frontend/src/utils/paymentShell.ts frontend/src/utils/__tests__/paymentShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- This keeps WeChat payment callback copy with payment runtime settings rather than auth shell settings.

## 2026-06-20 Provider callback bind-login placeholders moved to auth shell config

### Done
- Updated OIDC, LinuxDo, and DingTalk callback bind-login forms to render email/password placeholders from auth shell settings.
- Reused existing `emailPlaceholder` and `passwordPlaceholder` labels; no new settings schema was needed.
- Strengthened callback source guards to prevent provider callback views from using local auth placeholder i18n.
- Updated OIDC and LinuxDo callback test fixtures with configured bind-login placeholder labels.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/CallbackAuthShell.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n ":placeholder=\\"t\\('auth\\.(emailPlaceholder|passwordPlaceholder)'\\)\\"|t\\('auth\\.(emailPlaceholder|passwordPlaceholder)'\\)" frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue -S`
- `git diff --check -- frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/__tests__/CallbackAuthShell.spec.ts frontend/src/views/auth/__tests__/OidcCallbackView.spec.ts frontend/src/views/auth/__tests__/LinuxDoCallbackView.spec.ts progress.md`

### Notes
- This was a safe shell-only cleanup. Provider callback success/error messages remain on existing i18n/error paths.

## 2026-06-20 WeChat OAuth entry hints moved to auth shell config

### Done
- Added WeChat OAuth provider and availability hint labels to the auth shell schema.
- Extended Sub2API default `auth_shell_config` with zh/en WeChat provider name, system-browser-only, WeChat-browser-only, native-app-only, and not-configured labels.
- Updated WechatOAuthSection to render provider name and unavailable hints from auth shell labels instead of local auth i18n.
- Strengthened OAuth section source guards so WeChat OAuth entry copy does not return to `auth.oauthFlow.wechat*` local i18n.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/auth/__tests__/WechatOAuthSection.spec.ts src/components/auth/__tests__/OidcOAuthSection.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "auth\\.oauthFlow\\.wechat(SystemBrowserOnly|BrowserOnly|NativeAppOnly|NotConfigured)|t\\('auth\\.wechatProviderName'\\)|t\\('auth\\.oauthFlow\\.wechat" frontend/src/components/auth/WechatOAuthSection.vue -S`
- `git diff --check -- frontend/src/components/auth/WechatOAuthSection.vue frontend/src/components/auth/__tests__/WechatOAuthSection.spec.ts frontend/src/components/auth/__tests__/OidcOAuthSection.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- This only moves the login entry UI copy. WeChat callback availability/error handling still has separate runtime logic and remains a follow-up candidate.

## 2026-06-20 WeChat callback availability copy moved to auth shell config

### Done
- Added `wechatAvailabilityUnknown` to the auth shell schema and Sub2API zh/en defaults.
- Updated WechatCallbackView to render the WeChat provider name and availability/unavailable messages from auth shell settings.
- Reused the existing auth shell labels for system-browser-only, WeChat-browser-only, native-app-only, and not-configured callback states.
- Updated WeChat callback tests and source guards so provider text and availability labels do not return to local Touch auth i18n.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/WechatCallbackView.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts src/utils/__tests__/authShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAuthShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "t\\('auth\\.wechatProviderName'\\)|auth\\.oauthFlow\\.wechat(AvailabilityUnknown|SystemBrowserOnly|BrowserOnly|NativeAppOnly|NotConfigured)|t\\('auth\\.oauthFlow\\.wechat" frontend/src/views/auth/WechatCallbackView.vue -S`
- `git diff --check -- frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/__tests__/WechatCallbackView.spec.ts frontend/src/views/auth/__tests__/CallbackAuthShell.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- Default login failure and business errors remain on existing i18n/error paths.

## 2026-06-20 Prompt import provider label moved to prompt catalog shell config

### Done
- Added `importProviderX` to the prompt catalog shell schema.
- Extended Sub2API default `prompt_catalog_shell_config` with zh/en import provider labels.
- Updated PromptCatalogView to render the X/Twitter import provider option from public prompt catalog shell settings.
- Added view and shell tests so the import source label does not return to hardcoded Vue copy.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig -count=1` from `backend/`
- `pnpm run frontend:typecheck`
- `rg -n "<option value=\\"x\\">X / Twitter</option>|importProviderX" frontend/src/views/public/PromptCatalogView.vue frontend/src/utils/promptCatalogShell.ts frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/utils/__tests__/promptCatalogShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go -S`
- `git diff --check -- frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/utils/promptCatalogShell.ts frontend/src/utils/__tests__/promptCatalogShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- This keeps the current supported provider as X only, but moves the visible provider label into public settings so the frontend shell is thinner.

## 2026-06-20 Prompt catalog default filters moved to prompt catalog shell config

### Done
- Added `defaults` to `prompt_catalog_shell_config` for `sourceType`, `hasImage`, and `pageSize`.
- Extended Sub2API default prompt catalog public settings with zh/en default filter and pagination behavior.
- Updated PromptCatalogView to apply configured defaults after public settings load and before the first catalog request.
- Removed the fixed frontend `PAGE_SIZE = 24` constant and hardcoded initial `case` / images-only filters from the view.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig -count=1` from `backend/`
- `pnpm run frontend:typecheck`
- `rg -n "const PAGE_SIZE = 24|sourceType: 'case'|hasImage: true|<option value=\\"x\\">X / Twitter</option>|defaults" frontend/src/views/public/PromptCatalogView.vue frontend/src/utils/promptCatalogShell.ts frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/utils/__tests__/promptCatalogShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go -S`
- `git diff --check -- frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/utils/promptCatalogShell.ts frontend/src/utils/__tests__/promptCatalogShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- This keeps sorting in the frontend request for now; moving sort defaults into API/config remains a follow-up candidate.

## 2026-06-20 Prompt catalog sort defaults moved to prompt catalog shell config

### Done
- Added `sortBy` and `sortOrder` to `prompt_catalog_shell_config.defaults`.
- Extended Sub2API default prompt catalog public settings with `imported_at desc`, preserving the existing runtime behavior as configurable defaults.
- Updated PromptCatalogView to send configured sort params instead of hardcoding `imported_at` / `desc` in the request.
- Added resolver normalization for allowed sort fields and directions so invalid config falls back safely.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig -count=1` from `backend/`
- `pnpm run frontend:typecheck`
- `rg -n "sort_by: 'imported_at'|sort_order: 'desc'|sortBy|sortOrder" frontend/src/views/public/PromptCatalogView.vue frontend/src/utils/promptCatalogShell.ts frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/utils/__tests__/promptCatalogShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go -S`
- `git diff --check -- frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/utils/promptCatalogShell.ts frontend/src/utils/__tests__/promptCatalogShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- Prompt catalog list defaults are now all driven by Sub2API public settings: source type, image filter, page size, sort field, and sort order.

## 2026-06-20 Prompt catalog generator handoff moved to prompt catalog shell config

### Done
- Added `generatorPath` and `generatorDraftSource` to `prompt_catalog_shell_config.defaults`.
- Extended Sub2API default prompt catalog public settings with the current generator path and draft source marker.
- Updated PromptCatalogView to save generator drafts and navigate using configured defaults instead of hardcoded action values.
- Added safe internal-path normalization so invalid external generator paths fall back to `/image-generator`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig -count=1` from `backend/`
- `pnpm run frontend:typecheck`
- `rg -n "window\\.location\\.assign\\('/image-generator'\\)|source: 'sub2api-vue-prompt-catalog'|generatorPath|generatorDraftSource" frontend/src/views/public/PromptCatalogView.vue frontend/src/utils/promptCatalogShell.ts frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/utils/__tests__/promptCatalogShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go -S`
- `git diff --check -- frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/utils/promptCatalogShell.ts frontend/src/utils/__tests__/promptCatalogShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- Prompt catalog's list defaults and generator handoff behavior are now controlled from Sub2API public settings, with frontend fallbacks only for malformed or missing config.

## 2026-06-20 Prompt catalog case/template headings moved into prompt catalog shell config

### Done
- Added `caseTitle`, `caseDescription`, `templateTitle`, and `templateDescription` to prompt catalog shell labels.
- Extended Sub2API default `prompt_catalog_shell_config` with separate cases and templates page heading copy.
- Updated PromptCatalogView to prefer shell-configured case/template headings over standalone prompt title settings.
- Kept standalone `prompt_cases_*` and `prompt_templates_*` values as temporary fallback so existing runtime settings do not break immediately.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsPromptCatalogShellConfig -count=1` from `backend/`
- `pnpm run frontend:typecheck`
- `rg -n "caseTitle|caseDescription|templateTitle|templateDescription|prompt_cases_title|prompt_templates_title" frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/utils/promptCatalogShell.ts frontend/src/utils/__tests__/promptCatalogShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go -S`
- `git diff --check -- frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/utils/promptCatalogShell.ts frontend/src/utils/__tests__/promptCatalogShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- The admin/runtime settings form still exposes standalone prompt cases/templates fields. A follow-up can migrate that editor into the prompt catalog shell JSON and then remove the fallback path.

## 2026-06-20 Runtime settings prompt heading editor moved to prompt catalog shell config

### Done
- Removed standalone Prompt Catalog cases/templates title and description fields from RuntimeSettingsView.
- Runtime Settings now edits Prompt Catalog headings through `prompt_catalog_shell_config` only.
- Stopped submitting `prompt_cases_*` and `prompt_templates_*` fields from the runtime settings form.
- Kept backend DTO/settings compatibility for existing clients and persisted settings while the frontend management entry is consolidated.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "form\\.prompt_cases_|form\\.prompt_templates_|'prompt_cases_title'|'prompt_cases_description'|'prompt_templates_title'|'prompt_templates_description'|promptCatalogShellConfig" frontend/src/views/admin/RuntimeSettingsView.vue frontend/src/views/admin/__tests__/RuntimeSettingsView.spec.ts -S`
- `git diff --check -- frontend/src/views/admin/RuntimeSettingsView.vue frontend/src/views/admin/__tests__/RuntimeSettingsView.spec.ts progress.md`

### Notes
- Public API fields for standalone prompt headings still exist as compatibility surface. The user-facing PromptCatalogView now prefers shell labels and the admin UI no longer edits the standalone fields.

## 2026-06-20 Prompt catalog frontend stopped reading standalone heading settings

### Done
- Removed PromptCatalogView fallback reads for `prompt_cases_*` and `prompt_templates_*`.
- Removed standalone prompt heading fields from frontend public settings/admin settings TypeScript interfaces.
- Prompt Catalog page headings now come from `prompt_catalog_shell_config.labels` only in the Vue frontend.
- Kept backend public/admin DTO fields for compatibility with existing API clients and persisted settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "prompt_cases_title|prompt_cases_description|prompt_templates_title|prompt_templates_description" frontend/src -S`
- `git diff --check -- frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/views/admin/RuntimeSettingsView.vue frontend/src/views/admin/__tests__/RuntimeSettingsView.spec.ts frontend/src/utils/promptCatalogShell.ts frontend/src/api/admin/settings.ts frontend/src/types/index.ts progress.md`

### Notes
- Remaining references are backend compatibility and frontend regression assertions only.

## 2026-06-20 Pricing page title and description moved to pricing shell config

### Done
- Removed PricingView fallback reads for standalone `pricing_title` and `pricing_description`.
- Pricing page heading copy now comes from `pricing_shell_config.labels.title` and `.description`.
- Updated PricingView tests so configured title/description are provided by the shell config.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/utils/__tests__/pricingShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "pricing_title|pricing_description|copy\\.value\\.title|copy\\.value\\.description" frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts frontend/src/utils/pricingShell.ts frontend/src/utils/__tests__/pricingShell.spec.ts -S`
- `git diff --check -- frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts progress.md`

### Notes
- Standalone pricing title/description fields still exist in admin/runtime settings and backend compatibility. The public pricing page no longer reads them.

## 2026-06-20 Credits page copy moved to credits shell config

### Done
- Removed CreditsView fallback reads for standalone `credits_title`, `credits_description`, `credits_purchase_label`, and `credits_balance_label`.
- Credits page title, description, purchase label, and balance label now come from `credits_shell_config.labels`.
- Kept `credits_per_balance` and `pricing_currency_symbol` as runtime numeric/formatting settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/CreditsView.spec.ts src/utils/__tests__/creditsShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "credits_title|credits_description|credits_purchase_label|credits_balance_label|copy\\.value\\.(title|description|purchase|balanceLabel)|credits_per_balance|pricing_currency_symbol" frontend/src/views/user/CreditsView.vue frontend/src/views/user/__tests__/CreditsView.spec.ts frontend/src/utils/creditsShell.ts frontend/src/utils/__tests__/creditsShell.spec.ts -S`
- `git diff --check -- frontend/src/views/user/CreditsView.vue frontend/src/views/user/__tests__/CreditsView.spec.ts progress.md`

### Notes
- Standalone credits copy fields may still exist in backend/settings compatibility, but the Credits Vue page no longer reads them.

## 2026-06-20 Runtime settings pricing and credits copy moved to shell configs

### Done
- Removed standalone pricing copy inputs from Runtime Settings: `pricing_title` and `pricing_description`.
- Removed standalone credits copy inputs from Runtime Settings: `credits_title`, `credits_description`, `credits_purchase_label`, and `credits_balance_label`.
- Runtime Settings now edits pricing/credits page copy through `pricing_shell_config` and `credits_shell_config`.
- Removed the frontend admin settings/public type surface for those standalone copy fields.
- Removed unused Runtime Settings i18n labels for those standalone copy fields.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "pricing_title|pricing_description|credits_title|credits_description|credits_purchase_label|credits_balance_label|pricingTitle|pricingDescription|creditsTitle|creditsDescription|creditsPurchaseLabel|creditsBalanceLabel" frontend/src/views/admin frontend/src/api/admin/settings.ts frontend/src/types/index.ts frontend/src/i18n/locales -S`
- `git diff --check -- frontend/src/views/admin/RuntimeSettingsView.vue frontend/src/views/admin/__tests__/RuntimeSettingsView.spec.ts frontend/src/api/admin/settings.ts frontend/src/types/index.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts`

### Notes
- Backend DTO/service fields remain for compatibility with existing persisted settings and API clients.
- Remaining frontend matches from the scan are Runtime Settings regression assertions and unrelated homepage pricing copy keys.

## 2026-06-20 Removed stale homepage pricing locale copy

### Done
- Removed unused homepage pricing locale keys from Chinese and English bundles after HomeView stopped rendering homepage pricing content.
- Removed stale keys for homepage pricing headings, empty states, recharge/subscription headings, top-up labels, and plan validity labels.
- Added a HomeView regression guard so the removed homepage pricing locale keys do not reappear in the public shell layer.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts src/utils/__tests__/homeShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "pricingKicker|pricingTitle|pricingDescription|pricingUnavailable|pricingEmptyPlans|pricingEmptyRecharge|rechargeProductsTitle|subscriptionPlansTitle|topUpCreditLine|topUpPriceLabel|planValidityDays|planValidityMonths|planValidityYears" frontend/src -S`
- `git diff --check -- frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/views/__tests__/HomeView.spec.ts`

### Notes
- The public pricing page still uses `pricing_shell_config`; this cleanup only removes legacy homepage pricing copy from locale bundles.

## 2026-06-20 Removed legacy Home locale section from frontend bundles

### Done
- Removed the unused `home` locale section from Chinese and English frontend bundles.
- HomeView already resolves all visible homepage copy from `home_shell_config`, backed by Sub2API public settings/defaults.
- Added a HomeView regression guard so the legacy `home` locale block does not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts src/utils/__tests__/homeShell.spec.ts src/i18n/__tests__/runtimeDefaultLocale.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "home:\\s*\\{|home\\.([A-Za-z0-9_]+)|['\\\"]home['\\\"]|\\$t\\(['\\\"]home|t\\(['\\\"]home" frontend/src -S`
- `git diff --check -- frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/views/__tests__/HomeView.spec.ts`

### Notes
- Sub2API backend still owns the default homepage shell in `defaultHomeShellConfig`; frontend locale bundles no longer carry duplicate Home page copy.

## 2026-06-20 Removed legacy Docs locale section from frontend bundles

### Done
- Removed the unused `docs` locale section from Chinese and English frontend bundles.
- DocsView already reads visible header/search copy from `docs_shell_config`, backed by Sub2API public settings/defaults.
- Added a DocsView regression guard so the legacy `docs` locale block does not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts src/utils/__tests__/docsShell.spec.ts src/utils/__tests__/docsContentBasePath.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "^  docs: \\{|docs\\.([A-Za-z0-9_]+)|\\$t\\(['\\\"]docs|t\\(['\\\"]docs|frameworkHint|On this page|本页目录" frontend/src -S`
- `git diff --check -- frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/views/public/__tests__/DocsView.spec.ts`

### Notes
- Remaining `docs` scan matches are plain identifiers, admin nested i18n paths, and regression assertions, not the public Docs page locale section.

## 2026-06-20 Removed legacy Model Plaza locale section from frontend bundles

### Done
- Removed the unused `modelsPlaza` locale section from Chinese and English frontend bundles.
- ModelsPlazaView already reads visible public model catalog copy from `model_plaza_shell_config`, backed by Sub2API public settings/defaults.
- Updated locale coverage so Model Plaza copy is expected to live in shell config instead of frontend locale bundles.
- Added a ModelsPlazaView regression guard so the legacy `modelsPlaza` locale block does not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/ModelsPlazaView.spec.ts src/utils/__tests__/modelPlazaDisplay.spec.ts src/i18n/__tests__/localeCoverage.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "^  modelsPlaza: \\{|modelsPlaza\\.([A-Za-z0-9_]+)|\\$t\\(['\\\"]modelsPlaza|t\\(['\\\"]modelsPlaza" frontend/src -S`
- `git diff --check -- frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/views/public/__tests__/ModelsPlazaView.spec.ts frontend/src/i18n/__tests__/localeCoverage.spec.ts`

### Notes
- Runtime Settings still contains JSON examples for `model_plaza_shell_config`; those are the intended admin configuration path, not public page locale fallback copy.

## 2026-06-20 Removed legacy Available Groups locale section from frontend bundles

### Done
- Removed the unused `availableGroups` locale section from Chinese and English frontend bundles.
- AvailableGroupsView already reads all page copy from `available_groups_shell_config`, backed by Sub2API public settings/defaults.
- Removed `availableGroups.title` and `availableGroups.description` from the route meta so the header/document title falls back to the static route title.
- Added an AvailableGroupsView regression guard so the legacy locale block and route title keys do not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AvailableGroupsView.spec.ts src/utils/__tests__/availableGroupsShell.spec.ts src/router/__tests__/title.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "^  availableGroups: \\{|availableGroups\\.([A-Za-z0-9_]+)|\\$t\\(['\\\"]availableGroups|t\\(['\\\"]availableGroups|titleKey: 'availableGroups|descriptionKey: 'availableGroups" frontend/src -S`
- `git diff --check -- frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/views/user/__tests__/AvailableGroupsView.spec.ts frontend/src/router/index.ts`

### Notes
- Remaining scan matches are regression assertions and local variable names in admin components, not i18n locale dependencies.

## 2026-06-20 Removed legacy API Guide locale section from frontend bundles

### Done
- Removed the unused `apiGuide` locale section from Chinese and English frontend bundles.
- ApiGuideView already reads all page copy from `api_guide_shell_config`, backed by Sub2API public settings/defaults.
- Removed `apiGuide.title` and `apiGuide.description` from the route meta so the header/document title falls back to the static route title.
- Added an ApiGuideView regression guard so the legacy locale block and route title keys do not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiGuideView.spec.ts src/utils/__tests__/apiGuideShell.spec.ts src/router/__tests__/title.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "^  apiGuide: \\{|apiGuide\\.([A-Za-z0-9_]+)|\\$t\\(['\\\"]apiGuide|t\\(['\\\"]apiGuide|titleKey: 'apiGuide|descriptionKey: 'apiGuide" frontend/src -S`
- `git diff --check -- frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/views/user/__tests__/ApiGuideView.spec.ts frontend/src/router/index.ts`

### Notes
- Remaining scan matches are test-local mock labels and regression assertions, not runtime i18n locale dependencies.

## 2026-06-20 Removed legacy API Test locale section from frontend bundles

### Done
- Removed the unused `apiTest` locale section from Chinese and English frontend bundles.
- ApiTestView already reads all page copy from `api_test_shell_config`, backed by Sub2API public settings/defaults.
- Removed `apiTest.title` and `apiTest.description` from the `/gateway-test` route meta so document/title behavior no longer depends on frontend locale fallback copy.
- Moved the API Guide streaming badge from `apiTest.stream` to the API Guide shell label set, and added `stream` to Sub2API's default `api_guide_shell_config`.
- Added regression guards so the legacy `apiTest` locale block and route title keys do not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiTestView.spec.ts src/utils/__tests__/apiTestShell.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/utils/__tests__/apiGuideShell.spec.ts src/router/__tests__/title.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_DefaultsAPIGuideShellConfig|TestSettingService_GetPublicSettings_DefaultsAPITestShellConfig' -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "apiTest\\." frontend/src -S`
- `git diff --check -- frontend/src/router/index.ts frontend/src/utils/apiGuideShell.ts frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/__tests__/ApiGuideView.spec.ts frontend/src/utils/__tests__/apiGuideShell.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/views/user/__tests__/ApiTestView.spec.ts`

### Notes
- Remaining `apiTest.*` scan matches are test-local shell-config fixtures and regression assertions, not production runtime locale dependencies.

## 2026-06-20 Removed legacy Available Channels locale section from frontend bundles

### Done
- Removed the unused `availableChannels` locale section from Chinese and English frontend bundles.
- AvailableChannelsView already reads table/search/status copy from `available_channels_shell_config`, backed by Sub2API public settings/defaults.
- Moved the model pricing popover labels into `available_channels_shell_config.pricing` and pass them through the user table path, so the user Available Channels page no longer depends on `availableChannels.pricing`.
- Removed `availableChannels.title` and `availableChannels.description` from route meta so document/title behavior no longer depends on frontend locale fallback copy.
- Added regression guards so the legacy locale block, route title keys, and user-page pricing prefix do not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AvailableChannelsView.spec.ts src/utils/__tests__/availableChannelsShell.spec.ts src/router/__tests__/title.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAvailableChannelsShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "availableChannels\\." frontend/src/views/user/AvailableChannelsView.vue frontend/src/router/index.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts -S`
- `git diff --check -- frontend/src/utils/availableChannelsShell.ts frontend/src/components/channels/SupportedModelChip.vue frontend/src/components/channels/AvailableChannelsTable.vue frontend/src/views/user/AvailableChannelsView.vue frontend/src/router/index.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/views/user/__tests__/AvailableChannelsView.spec.ts frontend/src/utils/__tests__/availableChannelsShell.spec.ts`

### Notes
- `SupportedModelChip` still keeps `availableChannels.pricing` as a generic fallback prefix for older/default callers; the user Available Channels route no longer passes or relies on it.

## 2026-06-20 Removed legacy Channel Status locale section from frontend bundles

### Done
- Removed the unused `channelStatus` locale section from Chinese and English frontend bundles.
- ChannelStatusView already passes all visible monitor copy through `channel_status_shell_config`, backed by Sub2API public settings/defaults.
- Removed the `nav.channelStatus` route title key from `/channel-status` so the page route no longer needs local i18n to resolve its title.
- Added regression guards so the legacy locale block and route title key do not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ChannelStatusView.spec.ts src/utils/__tests__/channelStatusShell.spec.ts src/router/__tests__/title.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsChannelStatusShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "channelStatus\\." frontend/src/views/user/ChannelStatusView.vue frontend/src/router/index.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts -S`
- `git diff --check -- frontend/src/utils/availableChannelsShell.ts frontend/src/components/channels/SupportedModelChip.vue frontend/src/components/channels/AvailableChannelsTable.vue frontend/src/views/user/AvailableChannelsView.vue frontend/src/views/user/__tests__/AvailableChannelsView.spec.ts frontend/src/utils/__tests__/availableChannelsShell.spec.ts frontend/src/views/user/ChannelStatusView.vue frontend/src/views/user/__tests__/ChannelStatusView.spec.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/router/index.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go progress.md`

### Notes
- The navigation label `nav.channelStatus` remains in locale files for menu/navigation copy; the page-specific `channelStatus` block has been removed.

## 2026-06-20 Removed legacy Redeem locale section from frontend bundles

### Done
- Removed the legacy user-facing `redeem` locale section from Chinese and English frontend bundles.
- RedeemView already reads page copy from `redeem_shell_config`, backed by Sub2API public settings/defaults.
- Migrated the admin user balance-history modal off `t('redeem.*')` and onto the same `redeem_shell_config` history labels.
- Added `balanceAddedAffiliate` to the Redeem shell label contract and Sub2API default `redeem_shell_config`, because the admin balance-history modal needs the affiliate-transfer history label.
- Removed `redeem.title` and `redeem.description` from the `/redeem` route meta.
- Added regression guards so the user-facing `redeem` locale block, route title keys, and modal `t('redeem.*')` dependency do not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/RedeemView.spec.ts src/components/admin/user/__tests__/UserBalanceHistoryModal.spec.ts src/utils/__tests__/redeemShell.spec.ts src/router/__tests__/title.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsRedeemShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "redeem\\." frontend/src -S`
- `git diff --check -- frontend/src/components/admin/user/UserBalanceHistoryModal.vue frontend/src/components/admin/user/__tests__/UserBalanceHistoryModal.spec.ts frontend/src/views/user/RedeemView.vue frontend/src/views/user/__tests__/RedeemView.spec.ts frontend/src/utils/redeemShell.ts frontend/src/router/index.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Remaining `redeem.*` scan matches are `admin.redeem.*` management-console labels plus tests/assertions, not the removed user-facing Redeem locale namespace.

## 2026-06-20 Removed legacy Affiliate locale section from frontend bundles

### Done
- Removed the legacy user-facing `affiliate` locale section from Chinese and English frontend bundles.
- AffiliateView already reads all user invite/rebate/transfer copy from `affiliate_shell_config`, backed by Sub2API public settings/defaults.
- Removed `affiliate.title` and `affiliate.description` from the `/affiliate` route meta.
- Added regression guards so the user-facing `affiliate` locale block and route title keys do not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AffiliateView.spec.ts src/utils/__tests__/affiliateShell.spec.ts src/router/__tests__/title.spec.ts`
- `go test -tags unit ./internal/service -run TestSettingService_GetPublicSettings_DefaultsAffiliateShellConfig -count=1`
- `pnpm run frontend:typecheck`
- `rg -n "(^|[^A-Za-z])affiliate\\." frontend/src/views/user/AffiliateView.vue frontend/src/router/index.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts -S`
- `git diff --check -- frontend/src/router/index.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/views/user/AffiliateView.vue frontend/src/views/user/__tests__/AffiliateView.spec.ts frontend/src/utils/affiliateShell.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go`

### Notes
- Remaining affiliate i18n usage is under admin settings namespaces such as `admin.settings.features.affiliate.*`; the user-facing `affiliate` namespace has been removed.

## 2026-06-20 Removed local i18n title metadata from dashboard, keys, and usage routes

### Done
- Removed `dashboard.title` / `dashboard.welcomeMessage` from the `/dashboard` route meta.
- Removed `keys.title` / `keys.description` from the `/keys` route meta.
- Removed `usage.title` / `usage.description` from the `/usage` route meta.
- Added a router regression guard so these shell-backed user routes do not regain local i18n title/description metadata.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/router/__tests__/title.spec.ts src/views/user/__tests__/UsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "titleKey: '(dashboard|keys|usage)\\.|descriptionKey: '(dashboard|keys|usage)\\." frontend/src/router/index.ts`
- `git diff --check -- frontend/src/router/index.ts frontend/src/router/__tests__/title.spec.ts`

### Notes
- The `dashboard`, `keys`, and `usage` locale namespaces still exist because shared/admin components still consume portions of them. This step only removes user-route bootstrap dependence on those local namespaces.

## 2026-06-20 Removed legacy Dashboard locale section from frontend bundles

### Done
- Removed the user-facing `dashboard` locale section from Chinese and English frontend bundles.
- DashboardView and its stat/chart/quick-action/recent-usage child components already read labels from `dashboard_shell_config`, backed by Sub2API public settings/defaults.
- Added a dashboard shell regression guard so the removed local `dashboard` locale namespace does not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/components/user/dashboard/__tests__/UserDashboardStats.spec.ts src/router/__tests__/title.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "^  dashboard: \\{|dashboard\\.title|dashboard\\.welcomeMessage" frontend/src/i18n/locales frontend/src/router frontend/src/components/user/dashboard frontend/src/views/user/DashboardView.vue -S`

### Notes
- Remaining `dashboard.*` scan matches are `admin.dashboard.*`, router regression assertions, or test fixtures; they are not the removed user-facing Dashboard locale namespace.

## 2026-06-20 Removed legacy API Keys locale section from frontend bundles

### Done
- Removed the user-facing `keys` locale section from Chinese and English frontend bundles.
- Kept KeysView, EndpointPopover, and UseKeyModal delegated to `api_keys_shell_config` through `apiKeysShell`.
- Moved the last production `keys.*` consumers out of shared user copy: monitor key picker now uses `admin.channelMonitor.form.*`, and admin promo/redeem copy tooltips use `common.copy`.
- Added a KeysView shell regression guard so the removed local `keys` locale namespace does not return.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/KeysView.shell.spec.ts src/utils/__tests__/apiKeysShell.spec.ts src/router/__tests__/title.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "^  keys: \\{|t\\('keys\\.|\\$t\\('keys\\." frontend/src -S`

### Notes
- The API keys shell schema and defaults remain in Sub2API public settings via `api_keys_shell_config`; the frontend no longer needs a top-level local `keys` i18n namespace.

## 2026-06-20 Moved export progress copy out of the shared usage namespace

### Done
- Audited `usage.*` production references and confirmed the namespace is still shared by admin usage, account stats, charts, image billing helpers, and user usage shell flow.
- Moved the reusable `ExportProgressDialog` copy from `usage.*` to `common.exportProgress.*`.
- Removed the export-progress-only keys from the local `usage` locale section while keeping user Usage page export button copy delegated through `usage_shell_config`.
- Added a focused ExportProgressDialog regression test covering rendered copy, progress aria label, and cancel event.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/common/__tests__/ExportProgressDialog.spec.ts src/components/admin/usage/__tests__/UsageFilters.spec.ts src/views/user/__tests__/UsageView.spec.ts src/utils/__tests__/usageShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "usage\\.(exportingProgress|exportedCount|estimatedTime|cancelExport)|t\\('usage\\.exporting'\\)|\\$t\\('usage\\.exporting'\\)" frontend/src -S`

### Notes
- The remaining top-level `usage` locale section is still required by admin/shared usage components. It should be shrunk incrementally by moving admin-only chart/table/account-stat labels to `admin.usage` or narrower shared namespaces.

## 2026-06-20 Moved image usage copy out of usage shell

### Done
- Moved shared image billing/size labels from top-level `usage.*` and `usage_shell_config` into `common.imageUsage.*`.
- Updated the shared `imageUsage` formatter helper to read `common.imageUsage.*` labels.
- Updated both user UsageView and admin UsageTable image billing UI to read `common.imageUsage.*`.
- Removed image-specific keys from `usageShellLabelKeys`, so user Usage shell config no longer owns image billing metadata copy.
- Added formatter regression coverage for missing sizes, legacy sizes, size source labels, and stable size-breakdown ordering.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/imageUsage.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/UsageView.spec.ts src/utils/__tests__/usageShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "usage\\.image|image(Unit|Count|BillingSize|InputSize|OutputSize|SizeSource|SizeBreakdown|UnitPrice|TotalPrice):" frontend/src -S`

### Notes
- Remaining `usage` locale keys still cover non-image usage records, account billing, service tier, tokens, charts, and admin/shared usage table labels. Those can be split in later passes.

## 2026-06-20 Moved service tier copy out of usage shell

### Done
- Moved service tier display labels from top-level `usage.*` and `usage_shell_config` into `common.serviceTier.*`.
- Updated `getUsageServiceTierLabel` to resolve `common.serviceTier.priority/flex/standard`.
- Updated both user UsageView and admin UsageTable to display the service tier label from `common.serviceTier.label`.
- Removed `serviceTier` from `usageShellLabelKeys`, so user Usage shell config no longer owns the enum label.
- Updated locale and formatter tests to guard the new common namespace.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/usageServiceTier.spec.ts src/i18n/__tests__/usageServiceTierLocales.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/UsageView.spec.ts src/utils/__tests__/usageShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "usage\\.serviceTier|serviceTier(Priority|Flex|Standard)|'serviceTier'" frontend/src -S`

### Notes
- Remaining `usage` locale keys still include generic usage record labels, account billing labels, token/cost table labels, and chart labels.

## 2026-06-20 Moved usage metric tooltip copy out of usage shell

### Done
- Moved shared token/cost tooltip labels from top-level `usage.*`, `admin.usage.*`, and `usage_shell_config` into `common.usageMetrics.*`.
- Updated user UsageView token/cost tooltips to read `common.usageMetrics.*` through `usageMetricText`.
- Updated admin UsageTable token/cost tooltips to read `common.usageMetrics.*`.
- Removed token/cost/cache TTL tooltip labels from `usageShellLabelKeys`, so user Usage shell config no longer owns those internal metric labels.
- Kept table/export column labels in their existing domains (`usage_shell_config` for user table copy and `admin.usage` for admin table/export copy).

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/user/__tests__/UsageView.spec.ts src/utils/__tests__/usageShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "usage\\.(tokenDetails|costDetails|inputTokenPrice|outputTokenPrice|perMillionTokens|unitPrice|cacheTtlOverridden|inputTokens|outputTokens|cacheReadTokens)|usageText\\('(tokenDetails|inputTokens|outputTokens|cacheCreation|cacheReadTokens|costDetails|inputCost|outputCost|inputTokenPrice|outputTokenPrice|perMillionTokens|unitPrice|cacheCreationCost|cacheReadCost|cacheTtlOverridden)'\\)|admin\\.usage\\.(inputCost|outputCost|cacheCreationCost|cacheReadCost|inputTokens|outputTokens|cacheCreation.*Tokens|cacheReadTokens)" frontend/src/components/admin/usage frontend/src/views/user frontend/src/utils frontend/src/i18n -S`

### Notes
- Remaining `usage` locale keys are now mostly broader usage table/record/chart labels and account billing summary labels.

## 2026-06-20 Moved account billing copy out of usage namespace

### Done
- Moved shared account/user billed labels from top-level `usage.*` into `common.billingMetrics.*`.
- Updated account usage badges, account stats modals, admin usage stats/cards/table, and admin usage CSV headers to use `common.billingMetrics.*`.
- Removed `accountBilled`, `userBilled`, `accountMultiplier`, and `accountCost` from the local top-level `usage` locale section.
- Updated stale focused tests on the verification path so they reflect the current `getUsage(accountId, source)` signature and current no-implicit-refetch behavior.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/account/__tests__/UsageProgressBar.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/i18n/__tests__/localeCoverage.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "usage\\.(accountBilled|userBilled|accountMultiplier|accountCost)|accountBilled:|userBilled:|accountMultiplier:|accountCost:" frontend/src -S`

### Notes
- Remaining `usage` locale keys are now concentrated in generic usage tables, export actions, request type labels, endpoint/model chart labels, and user Usage shell labels.

## 2026-06-20 Moved usage routing copy out of usage namespace

### Done
- Moved upstream/inbound endpoint and model mapping labels from top-level `usage.*` into `common.usageRouting.*`.
- Updated endpoint/model distribution charts, admin UsageView export headers, admin usage table endpoint labels, and account stats modals to use `common.usageRouting.*`.
- Removed routing-only keys from the local top-level `usage` locale section: `endpointDistribution`, `inbound`, `inboundEndpoint`, `upstream`, `upstreamEndpoint`, `requestedModel`, `upstreamModel`, `mapping`, and `path`.
- Updated the ModelDistributionChart test fixture to include the current `account_cost` field required by the chart table.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/charts/__tests__/ModelDistributionChart.spec.ts src/views/admin/__tests__/UsageView.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/i18n/__tests__/localeCoverage.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "usage\\.(endpointDistribution|inboundEndpoint|upstreamEndpoint|requestedModel|upstreamModel|mapping|inbound|upstream|path)" frontend/src/components/account frontend/src/components/admin/account frontend/src/components/admin/usage frontend/src/components/charts frontend/src/views/admin frontend/src/i18n/locales -S`

### Notes
- Remaining `usage` locale keys now mostly cover generic usage table/filter/export labels and user Usage shell labels.

## 2026-06-20 Moved request type copy out of usage shell

### Done
- Moved request type labels (`ws`, `stream`, `sync`, `unknown`) from top-level `usage.*` into `common.requestType.*`.
- Updated admin UsageView, UsageTable, UsageFilters, and user UsageView request type rendering to use `common.requestType.*`.
- Removed request type labels from `usageShellLabelKeys`, so `usage_shell_config` no longer owns request type enum copy.
- Trimmed Sub2API's default `usage_shell_config` to remove request type labels and earlier migrated tooltip/image/service-tier fields.
- Simplified `usageShellConfigDefault()` after the old token/cost detail injection path became obsolete.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/admin/usage/__tests__/UsageFilters.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/admin/__tests__/UsageView.spec.ts src/views/user/__tests__/UsageView.spec.ts src/utils/__tests__/usageShell.spec.ts src/i18n/__tests__/localeCoverage.spec.ts`
- `pnpm run frontend:typecheck`
- `go test ./internal/service` from `backend/`
- `rg -n "usage\\.(ws|stream|sync|unknown)|usageText\\('(ws|stream|sync|unknown)'\\)" frontend/src/views/admin/UsageView.vue frontend/src/components/admin/usage frontend/src/views/user/UsageView.vue frontend/src/views/user/__tests__/UsageView.spec.ts frontend/src/utils/usageShell.ts frontend/src/i18n/locales backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go -S`

### Notes
- Remaining `usage` locale keys are now mainly generic table/export labels and user Usage shell labels such as `model`, `endpoint`, `tokens`, `cost`, `rate`, `original`, and `billed`.

## 2026-06-20 Moved admin usage copy out of top-level usage namespace

### Done
- Moved admin Usage page, filter, stats, table, and export labels from top-level `usage.*` into `admin.usage.*`.
- Moved the remaining endpoint chart table header to `common.usageRouting.endpoint`.
- Kept user Usage page labels on `usage_shell_config`, since those are public runtime settings owned by Sub2API.
- Verified there are no remaining direct `t('usage.*')` or `$t('usage.*')` calls in `frontend/src`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/admin/usage/__tests__/UsageFilters.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts src/views/admin/__tests__/UsageView.spec.ts src/components/charts/__tests__/ModelDistributionChart.spec.ts src/i18n/__tests__/localeCoverage.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "(t|\\$t)\\('usage\\." frontend/src -S`

### Notes
- The remaining local `usage` locale object is no longer directly referenced by Vue i18n calls. The live user Usage page shell still reads from Sub2API public `usage_shell_config`.

## 2026-06-20 Removed unused local usage locale fallback

### Done
- Deleted the unused top-level `usage` locale object from both Chinese and English locale files.
- Confirmed user Usage labels remain backed by Sub2API public `usage_shell_config`, not local Vue i18n fallback copy.
- Confirmed no direct `t('usage.*')` or `$t('usage.*')` calls remain in `frontend/src`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/i18n/__tests__/localeCoverage.spec.ts src/views/user/__tests__/UsageView.spec.ts src/views/admin/__tests__/UsageView.spec.ts src/components/admin/usage/__tests__/UsageFilters.spec.ts src/components/admin/usage/__tests__/UsageTable.spec.ts`
- `rg -n "^  usage: \\{" frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts -S`
- `rg -n "(t|\\$t)\\('usage\\." frontend/src -S`

### Notes
- User Usage tests still contain local mock keys named `usage.*`, but the production locale files and production Vue i18n calls no longer depend on them.

## 2026-06-20 Slimmed Prompt Catalog display derivation

### Done
- Removed PromptCatalogView wrapper functions for image URL, source label, visible tags, all tags, and prompt character count.
- Updated the Vue template to render Sub2API API fields directly: `primary_image_url`, `source_display_label`, `visible_tags`, `all_tags`, and `prompt_char_count`.
- Added source-level regression checks so these display derivations stay owned by the Sub2API prompt catalog API instead of returning to local Vue calculations.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts`
- `rg -n "cardImageUrl|sourceDisplayLabel|promptCharCount|visibleTags\\(|allTags\\(" frontend/src/views/public/PromptCatalogView.vue -S`
- `rg -n "primary_image_url|source_display_label|prompt_char_count|visible_tags|all_tags" frontend/src/views/public/PromptCatalogView.vue frontend/src/api/prompts.ts backend/internal/handler/prompt_catalog_handler.go -S`

### Notes
- This does not remove Prompt Catalog UI state itself; it narrows the Vue page to API-field rendering for display metadata that Sub2API already returns.

## 2026-06-20 Removed Prompt Catalog frontend runtime defaults

### Done
- Changed `resolvePromptCatalogShellConfig` to drop invalid or missing defaults instead of recreating Sub2API's prompt catalog business defaults in the frontend.
- Updated PromptCatalogView so page size, sorting, generator path, and draft source come from `prompt_catalog_shell_config.defaults`.
- Omitted list query defaults when Sub2API public settings do not provide them, allowing the prompt catalog API to apply its own defaults.
- Made the generator action a no-op when no configured generator path is present instead of falling back to a local route.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts`
- `rg -n "\\|\\| 24|\\|\\| 'imported_at'|\\|\\| 'desc'|\\|\\| '/image-generator'|\\|\\| 'sub2api-vue-prompt-catalog'|'/image-generator'|'sub2api-vue-prompt-catalog'|'imported_at'|'desc'" frontend/src/views/public/PromptCatalogView.vue frontend/src/utils/promptCatalogShell.ts -S`
- `rg -n "pageSize: 24|sortBy: 'imported_at'|sortOrder: 'desc'|generatorPath: '/image-generator'|generatorDraftSource: 'sub2api-vue-prompt-catalog'" frontend/src/utils/promptCatalogShell.ts frontend/src/views/public/PromptCatalogView.vue frontend/src/utils/__tests__/promptCatalogShell.spec.ts -S`

### Notes
- `promptCatalogShell.ts` still lists allowed enum values such as `imported_at` and `desc` for validation, but it no longer uses them as local runtime defaults.

## 2026-06-20 Removed Home shell frontend default icons

### Done
- Removed `defaultExperienceIcons` from the Home shell parser.
- Made experience card icons optional, so Sub2API `home_shell_config.experienceCards[].icon` is the only source for configured card icons.
- Updated HomeView to render the icon container only when an icon is provided by public settings.
- Added parser coverage for missing/invalid icon values to ensure the frontend no longer assigns default icon choices.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/homeShell.spec.ts src/views/__tests__/HomeView.spec.ts`
- `rg -n "defaultExperienceIcons|\\|\\| defaultExperienceIcons|icon: readExperienceIcon\\(card\\.icon\\) \\|\\||icon: 'server'|icon: 'key'|icon: 'sparkles'|icon: 'chart'" frontend/src/utils/homeShell.ts frontend/src/views/HomeView.vue frontend/src/utils/__tests__/homeShell.spec.ts frontend/src/views/__tests__/HomeView.spec.ts -S`
- `rg -n "v-if=\"feature\\.icon\"|Icon :name=\"feature\\.icon\"|icon\\?:" frontend/src/views/HomeView.vue frontend/src/utils/homeShell.ts -S`

### Notes
- The frontend still validates the allowed icon enum for rendering safety, but visual icon choices now come from Sub2API public settings.

## 2026-06-20 Removed Pricing validity duration fallback

### Done
- Removed the Pricing page's local `days || 1` fallback for monthly subscription validity display.
- Updated `formatValidity` to render duration only when Sub2API catalog data provides a positive `validity_days` and a supported `validity_unit`.
- Added a regression test that a month plan with missing/zero duration no longer displays a synthesized `/1...` duration.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/utils/__tests__/pricingShell.spec.ts`
- `rg -n "days \\|\\| 1|validity_days \\|\\||validity_unit \\|\\| 'day'|monthLabel\\.value\\}|dayLabel\\.value : daysLabel" frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts -S`
- `rg -n "formatValidity|validity_days|validity_unit" frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts -S`

### Notes
- Pricing labels and currency remain driven by Sub2API public settings. This slice only removes a local duration-defaulting behavior from the Vue shell.

## 2026-06-20 Removed PaymentView validity unit fallback

### Done
- Removed the user Payment page's local `validity_unit || 'day'` fallback.
- Updated subscription validity suffix rendering to display day counts only when Sub2API checkout data explicitly returns `validity_unit: "day"` with a positive `validity_days`.
- Kept month/year labels driven by the payment shell labels instead of local i18n fallback.
- Added a source-level regression check to prevent the local day-unit fallback from returning.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentView.spec.ts src/utils/__tests__/paymentShell.spec.ts`
- `rg -n "validity_unit \\|\\| 'day'|validity_unit \\|\\| \\\"day\\\"|validity_days \\|\\| 0|selectedPlan\\.value\\.validity_days\\}\\$\\{paymentText\\('days'\\)" frontend/src/views/user/PaymentView.vue frontend/src/views/user/__tests__/PaymentView.spec.ts -S`
- `rg -n "planValiditySuffix|validity_unit|validity_days" frontend/src/views/user/PaymentView.vue frontend/src/views/user/__tests__/PaymentView.spec.ts -S`

### Notes
- This keeps payment display behavior aligned with the checkout/catalog data returned by Sub2API instead of interpreting missing units in the Vue shell.

## 2026-06-20 Removed PaymentView platform label fallback

### Done
- Changed shared platform display helpers to accept missing platform data without synthesizing an `API` label.
- Updated PaymentView selected-plan and active-subscription badges to render platform labels only when Sub2API data provides a display label or platform value.
- Removed `|| ''` platform arguments from PaymentView badge/text/accent class calls so the frontend no longer normalizes missing backend data into local fallback labels.
- Added regression coverage for missing platform labels and neutral styling defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentView.spec.ts src/utils/__tests__/paymentShell.spec.ts src/utils/__tests__/platformColors.spec.ts`
- `rg -n "group_platform \\|\\| ''|group\\?\\.platform \\|\\| ''|platformLabel\\([^\\n]*\\|\\| ''|platformBadgeClass\\([^\\n]*\\|\\| ''|platformTextClass\\([^\\n]*\\|\\| ''|platformAccentBarClass\\([^\\n]*\\|\\| ''|platformBadgeLightClass\\([^\\n]*\\|\\| ''" frontend/src/views/user/PaymentView.vue -S`
- `rg -n "platformLabel\\(|platformBadgeClass\\(|platformTextClass\\(|platformAccentBarClass\\(|platformBadgeLightClass\\(" frontend/src/views/user/PaymentView.vue frontend/src/utils/platformColors.ts frontend/src/utils/__tests__/platformColors.spec.ts -S`

### Notes
- Payment UI still owns layout and interaction state, but platform naming is now treated as Sub2API/catalog data instead of a Vue-side default.

## 2026-06-20 Removed SubscriptionPlanCard display fallbacks

### Done
- Updated SubscriptionPlanCard so missing `group_platform` no longer renders a synthesized platform badge or `API` label.
- Removed the local `validity_unit || 'day'` fallback from the subscription card.
- Changed validity suffix rendering to show day duration only when Sub2API returns `validity_unit: "day"` and a positive `validity_days`.
- Added component regression tests for missing platform and missing validity-unit data.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/SubscriptionPlanCard.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/utils/__tests__/paymentShell.spec.ts src/utils/__tests__/platformColors.spec.ts`
- `rg -n "validity_unit \\|\\| 'day'|validity_unit \\|\\| \\\"day\\\"|group_platform \\|\\| ''|group_platform \\|\\| \\\"\\\"|platformLabel\\([^\\n]*\\|\\| ''|platformLabel\\([^\\n]*\\|\\| \\\"\\\"" frontend/src/components/payment/SubscriptionPlanCard.vue frontend/src/views/user/PaymentView.vue -S`
- `rg -n "validitySuffix|platformLabel\\(|group_platform|validity_unit" frontend/src/components/payment/SubscriptionPlanCard.vue frontend/src/components/payment/__tests__/SubscriptionPlanCard.spec.ts -S`

### Notes
- This keeps payment plan-card display aligned with Sub2API checkout/catalog fields instead of recreating missing defaults in the Vue component.

## 2026-06-20 Moved Affonso cookie duration fallback to Sub2API settings

### Done
- Removed the frontend `Affonso` public integration fallback that synthesized `data-cookie_duration="30"` when public settings omitted `affonso_cookie_duration`.
- Added a Sub2API service-level default for `web_affonso_cookie_duration`, so public settings remain the single source for the Affonso cookie duration.
- Updated public settings service and handler tests to assert configured Affonso cookie duration is exposed through the API.
- Added frontend regression coverage that public integration injection uses the provided setting and does not recreate the `30` fallback locally.

### Validation
- `go test -tags unit ./internal/service ./internal/handler -run 'TestSettingService_GetPublicSettings|TestSettingHandler_GetPublicSettings' -count=1`
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/publicIntegrations.spec.ts`
- `rg -n "data-cookie_duration.*\\|\\| '30'|affonso_cookie_duration\\).*\\|\\| '30'|affonso_cookie_duration.*30" frontend/src/utils/publicIntegrations.ts frontend/src/utils/__tests__/publicIntegrations.spec.ts backend/internal/service/setting_service.go backend/internal/service/setting_service_public_test.go backend/internal/handler/setting_handler_public_test.go -S`

### Notes
- Existing databases with an empty `web_affonso_cookie_duration` now receive the same default from Sub2API public settings instead of the Vue runtime.

## 2026-06-20 Removed CCS import platform synthesis from API Keys

### Done
- Removed API Keys page logic that converted missing key group platform data into `anthropic` before building CCS import links.
- Updated the CCS import helper so an omitted platform falls through to the Claude import behavior without pretending the backend returned Anthropic.
- Added regression tests for missing platform imports and source-level checks against the old `row.group?.platform || 'anthropic'` pattern.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/KeysView.shell.spec.ts src/utils/__tests__/ccswitchImport.spec.ts src/utils/__tests__/publicIntegrations.spec.ts`
- `rg -n "platform \\|\\| 'anthropic'|platform \\|\\| \\\"anthropic\\\"|row\\.group\\?\\.platform \\|\\| 'anthropic'|row\\.group\\?\\.platform \\|\\| \\\"anthropic\\\"|switch \\(platform \\|\\| 'anthropic'\\)" frontend/src/views/user/KeysView.vue frontend/src/utils/ccswitchImport.ts frontend/src/views/user/__tests__/KeysView.shell.spec.ts frontend/src/utils/__tests__/ccswitchImport.spec.ts -S`

### Notes
- The default CCS target remains Claude for unknown/missing platforms, but the Vue page no longer fabricates an Anthropic platform value.

## 2026-06-20 Removed login agreement dynamic defaults from frontend templates

### Done
- Removed frontend defaults for login agreement contact info, effective date, and site URL in `loginAgreementTemplates`.
- Kept placeholder rendering and legacy dynamic-line normalization, but missing values now render as empty instead of injecting local static copy.
- Added regression coverage so the frontend utility does not reintroduce the old fixed date, fallback contact sentence, or fallback current-domain sentence.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/loginAgreementTemplates.spec.ts`
- `rg -n "请通过站点设置中的客服联系方式与运营方联系|以您当前访问本服务时所使用的域名为准|return value \\|\\| \\\"2026-03-31\\\"|return value \\|\\| '2026-03-31'" frontend/src/utils/loginAgreementTemplates.ts frontend/src/utils/__tests__/loginAgreementTemplates.spec.ts -S`

### Notes
- The legal document template body is still a frontend utility; this step only removes dynamic value synthesis so settings-provided values remain authoritative.

## 2026-06-20 Removed CCS usage unit fallback from API Keys

### Done
- Removed the `USD` fallback from the CCS import usage script generated by the API Keys page.
- The generated script now returns `response.unit` or `response.quota.unit` as provided by the usage API, without synthesizing a frontend business unit.
- Added a source-level regression check to keep the fallback out of the generated script.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/KeysView.shell.spec.ts`
- `rg -n "response\\?\\.quota\\?\\.unit \\?\\? ['\\\"]USD['\\\"]|unit \\?\\? ['\\\"]USD['\\\"]" frontend/src/views/user/KeysView.vue frontend/src/views/user/__tests__/KeysView.shell.spec.ts -S`

### Notes
- This keeps CCS import metadata aligned with the Sub2API usage response instead of assuming all missing units are USD.

## 2026-06-20 Removed frontend payment currency and country defaults

### Done
- Removed the Vue payment helper defaults that synthesized missing currencies as `CNY` and missing country codes as `CN`.
- Updated payment amount formatting so a missing or invalid currency renders as a plain numeric amount instead of a fake fiat currency.
- Changed Stripe result/carrier pages to initialize payment currency state explicitly as empty until an order or recovery snapshot provides currency data.
- Added Airwallex regression coverage so missing snapshot currency/country data is passed through as empty values instead of frontend-generated defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/currency.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/StripePopupView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "DEFAULT_PAYMENT_CURRENCY|DEFAULT_PAYMENT_COUNTRY_CODE|currency = ref\\(normalizePaymentCurrency\\(\\)\\)|snapshot\\.currency \\|\\| ['\\\"]CNY['\\\"]|snapshot\\.countryCode \\|\\| ['\\\"]CN['\\\"]|route\\.query\\.currency \\|\\| ['\\\"]CNY['\\\"]" frontend/src -S`

### Notes
- Currency/country defaults now need to come from Sub2API order/payment configuration data. The frontend still formats explicit currencies, but it no longer masks missing payment configuration as CNY/CN.

## 2026-06-20 Removed frontend payment method defaults

### Done
- Removed the Vue helper defaults that synthesized missing visible payment methods as `wxpay` and missing Stripe popup methods as `alipay`.
- Changed payment method normalization to return an empty value when Sub2API/URL data does not provide a known method.
- Updated WeChat resume handling so missing `payment_type` no longer overwrites the checkout-selected method with an empty value, while still avoiding a hardcoded `wxpay` fallback.
- Changed Stripe popup initialization to surface a missing-params error when no explicit Stripe sub-method is provided instead of silently treating it as Alipay.
- Renamed the Stripe unknown-method color fallback to avoid implying a payment method default.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/paymentMethod.spec.ts src/views/user/__tests__/paymentWechatResume.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/views/user/__tests__/StripePopupView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "DEFAULT_VISIBLE_PAYMENT_METHOD|DEFAULT_STRIPE_PAYMENT_METHOD|centralizes the visible payment method default|Stripe popup method defaults|route\\.query\\.method \\|\\| ['\\\"]alipay['\\\"]|resolveVisiblePaymentMethod\\([^\\n]*\\) \\|\\||normalizeStripePaymentMethod\\([^\\n]*\\) \\|\\|" frontend/src -S`

### Notes
- Payment method defaults now come from Sub2API checkout configuration or explicit route/order data. The frontend keeps only normalization and neutral display coloring for unknown Stripe sub-methods.

## 2026-06-20 Removed public pricing catalog metadata fallbacks

### Done
- Updated the public Pricing page so missing recharge product names no longer render the product ID as a synthetic title.
- Updated subscription cards on the public Pricing page so missing source labels no longer fall back to `siteName`.
- Removed public Pricing page display fallbacks that synthesized missing plan rate as `x1` and missing quota labels as configured unlimited copy.
- Added regression coverage for sparse catalog items and source checks against the removed fallback patterns.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/utils/__tests__/pricingShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "product\\.name \\|\\| product\\.id|group_display_label \\|\\| plan\\.group_name \\|\\| plan\\.group_platform \\|\\| siteName|plan\\.rate_multiplier \\?\\? 1|plan\\.quota_label \\|\\| unlimitedLabel|const unlimitedLabel" frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts -S`

### Notes
- Public Pricing still owns layout and tab interaction, but catalog metadata now needs to be provided by Sub2API instead of being invented by the Vue page.

## 2026-06-20 Removed unknown payment method icon fallback

### Done
- Updated `PaymentMethodSelector` so unknown/custom payment methods no longer render the Alipay icon as a synthetic fallback.
- Unknown methods still display the explicit method type returned by Sub2API, but no provider icon is shown unless the method maps to a known icon.
- Added regression coverage for a custom provider method and a source check against the removed `METHOD_ICONS[type] || alipayIcon` pattern.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/PaymentMethodSelector.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "METHOD_ICONS\\[type\\] \\|\\| alipayIcon|return METHOD_ICONS\\[type\\] \\|\\| alipayIcon|unknown.*alipay|custom_provider" frontend/src/components/payment/PaymentMethodSelector.vue frontend/src/components/payment/__tests__/PaymentMethodSelector.spec.ts -S`

### Notes
- Payment method identity is now driven by the method type/config data instead of being visually coerced to Alipay when the frontend does not recognize it.

## 2026-06-20 Removed order-table synthetic user identity fallback

### Done
- Updated the payment order table user column so it only renders explicit `user_email` or `user_name` values returned by Sub2API.
- Removed the frontend fallback that displayed missing user identity as `#user_id`.
- Added regression coverage for missing user display data and explicit email/name display data.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/OrderTable.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Admin order rows with no user email/name now render no synthetic identity. If an ID should be shown, Sub2API should provide it as an explicit display field instead of relying on Vue-side formatting.

## 2026-06-20 Removed local payment expiry timeout fallbacks

### Done
- Removed the frontend `30 * 60` payment waiting timeout fallback from the QR route view, QR dialog, and embedded payment status panel.
- Missing or invalid `expires_at` now expires immediately instead of pretending Sub2API returned a 30-minute payment window.
- Prevented payment polling from starting when the waiting UI is already expired due to missing expiry data.
- Fixed `PaymentQRDialog` initialization so an initially open dialog runs the same setup path as a dialog opened after mount.
- Added regression coverage for all three payment waiting surfaces.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentQRCodeView.spec.ts src/components/payment/__tests__/PaymentQRDialog.spec.ts src/components/payment/__tests__/PaymentStatusPanel.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Payment expiry is now a Sub2API/order API responsibility for these active payment waiting surfaces. The frontend still counts down explicit expiry timestamps, but it no longer creates a hidden timeout policy.

## 2026-06-20 Removed admin login agreement date fallback

### Done
- Removed the frontend `2026-03-31` login-agreement updated-at default from `SettingsView`.
- Changed default login agreement document generation in the admin form to use an empty effective date unless Sub2API settings provide one.
- Stopped the commercial-template action from writing a hardcoded date into the form.
- Added a source-level regression check so the fixed date is not reintroduced in the admin settings frontend.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- The backend still owns the platform default date through settings service defaults. The frontend no longer carries a separate policy date.

## 2026-06-20 Removed admin payment product suffix fallback

### Done
- Removed the `CNY` placeholder/default from the admin payment product-name suffix field.
- Changed the payment product-name preview to join only explicit configured parts instead of appending a frontend currency suffix.
- Added a source-level regression check so the admin settings frontend does not reintroduce the `CNY` suffix fallback.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Payment product suffix/currency display now needs to come from Sub2API settings. The admin page previews explicit settings but no longer supplies a business currency on its own.

## 2026-06-20 Removed admin provider creation default

### Done
- Changed the payment provider dialog's initial provider key from `easypay` to an empty value.
- Changed the admin provider creation entry point to use the first enabled provider key when present, otherwise pass an empty key instead of falling back to `easypay`.
- Added regression checks for both the dialog and SettingsView entry point.
- Added the missing Airwallex webhook path so the existing Airwallex provider guidance can render its webhook hint and callback URL.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/PaymentProviderDialog.spec.ts src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- New provider creation now depends on configured/enabled provider options. The frontend no longer silently chooses EasyPay when no provider key is available.

## 2026-06-20 Removed admin payment strategy/rate-limit save fallbacks

### Done
- Removed the admin SettingsView load-time fallback that converted an empty `payment_load_balance_strategy` into `round-robin`.
- Changed payment cancel-rate save payloads so empty/zero max and window values remain `0` instead of being rewritten to `10` and `1` by the frontend.
- Added source-level and payload regression coverage for these payment settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Payment load-balance and cancel-rate defaults are already normalized by Sub2API backend config parsing. The admin frontend now sends the form state without adding its own policy defaults.

## 2026-06-20 Removed admin subscription validity-unit fallbacks

### Done
- Updated the admin subscription plan list so missing `validity_unit` no longer displays as `days`.
- Updated the subscription plan edit dialog so existing plans with missing `validity_unit` stay empty instead of being rewritten to `days`.
- Added source-level regression coverage for both the list and edit dialog.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/AdminPaymentCatalogView.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- New plan creation still offers the existing default `days` selection. This change only removes frontend synthesis for existing Sub2API plan data.

## 2026-06-20 Removed admin channel/group platform defaults

### Done
- Updated `ChannelsView` so missing channel pricing `platform` values are no longer treated as Anthropic in the frontend.
- Added a source-level regression test for the channel pricing platform filter.
- Updated `GroupsView` edit handling so missing `subscription_type` stays empty instead of being rewritten to `standard`.
- Converted an empty edited `subscription_type` to `undefined` in the update payload so backend `omitempty` semantics remain authoritative.
- Added source-level regression coverage for the group edit path.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/ChannelsView.source.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/groupsModelsListLayout.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Backend compatibility defaults remain in backend handlers/services where they already existed. The frontend no longer synthesizes these defaults while loading/editing existing data.

## 2026-06-20 Removed available-channel subscription-type fallback

### Done
- Changed available-channel group rendering to pass through Sub2API's `subscription_type` value instead of rewriting missing values to `standard`.
- Updated the available-channel API type so `subscription_type` can be absent rather than forced into a frontend string.
- Removed `standard` defaults from `GroupBadge` and `GroupOptionItem` so shared group UI components no longer synthesize business subscription types.
- Added source-level regression coverage to prevent reintroducing the available-channel `standard` fallback.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AvailableChannelsView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/api/channels.ts frontend/src/components/channels/AvailableChannelsTable.vue frontend/src/components/common/GroupBadge.vue frontend/src/components/common/GroupOptionItem.vue frontend/src/views/user/__tests__/AvailableChannelsView.spec.ts progress.md`

### Notes
- Existing callers that pass explicit `standard` or `subscription` still render the same way. Missing values now remain missing in the UI path, keeping subscription semantics owned by Sub2API data.

## 2026-06-20 Removed account-edit platform fallback

### Done
- Updated `EditAccountModal` model whitelist selectors to pass the account platform as returned by Sub2API instead of rewriting missing values to `anthropic`.
- Changed account preset mappings so missing account/platform data produces no presets instead of Anthropic presets.
- Added a regression test proving an account without `platform` no longer gives the model selector an Anthropic platform.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/components/account/EditAccountModal.vue frontend/src/components/account/__tests__/EditAccountModal.spec.ts progress.md`

### Notes
- Create-account defaults remain unchanged because that is an explicit creation flow. This change only removes frontend synthesis while editing existing account data.

## 2026-06-20 Stopped auto-persisting payment provider config defaults

### Done
- Removed the `PaymentProviderDialog` `applyDefaults` path that copied frontend `defaultValue` hints into provider config state.
- Provider reset, provider-key changes, and provider loading now keep missing provider config fields missing instead of filling values like Airwallex API base, country, or currency from the frontend.
- Kept placeholders/hints visible so admins can still see suggested values without the frontend silently persisting them.
- Added regression coverage proving provider config defaults are not applied into form values.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/PaymentProviderDialog.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/components/payment/PaymentProviderDialog.vue frontend/src/components/payment/__tests__/PaymentProviderDialog.spec.ts progress.md`

### Notes
- Required provider fields still validate before save. The change only removes automatic frontend persistence of suggested defaults; actual provider defaults/config should come from Sub2API settings or explicit admin input.

## 2026-06-20 Moved gateway guide model defaults into public settings overrides

### Done
- Added `gatewayVariants` parsing under `api_guide_shell_config` so Sub2API public settings can override gateway default models, model placeholders, and fallback model lists per variant.
- Updated API Guide to render configured gateway defaults instead of only the frontend hardcoded defaults.
- Updated API Test to use the same configured gateway defaults and fallback model lists when initializing variants and loading model options.
- Kept existing built-in gateway defaults as compatibility fallback when Sub2API settings do not provide overrides.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/gatewayDocs.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/views/user/__tests__/ApiTestView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/gatewayDocs.ts frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/ApiTestView.vue frontend/src/utils/__tests__/gatewayDocs.spec.ts frontend/src/views/user/__tests__/ApiGuideView.spec.ts frontend/src/views/user/__tests__/ApiTestView.spec.ts progress.md`

### Notes
- This reuses the existing `api_guide_shell_config` public setting rather than adding a new endpoint. Suggested config shape: `{ "zh": { "gatewayVariants": { "openaiChat": { "defaultModel": "...", "fallbackModels": ["..."] } } } }`.

## 2026-06-20 Removed payment provider config default hints

### Done
- Removed `ConfigFieldDef.defaultValue` from frontend payment provider config definitions.
- Removed static frontend provider config placeholders/default suggestions for Stripe, Airwallex, Creem, and Waffo.
- Kept required-field validation intact while leaving provider config values to Sub2API settings or explicit admin input.
- Ensured a new payment provider form no longer defaults locally to `easypay`.
- Added Airwallex to the payment webhook path map so its callback URL is generated from the same Sub2API webhook path table as other providers.
- Extended regression coverage to prevent reintroducing local payment provider defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/providerConfig.spec.ts src/components/payment/__tests__/PaymentProviderDialog.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "defaultValue|field\\.defaultValue|https://api\\.airwallex\\.com/api/v1|defaultValue: ['\"](CNY|USD|CN|false)" frontend/src/components/payment/providerConfig.ts frontend/src/components/payment/PaymentProviderDialog.vue -S`
- `git diff --check -- frontend/src/components/payment/providerConfig.ts frontend/src/components/payment/PaymentProviderDialog.vue frontend/src/components/payment/__tests__/providerConfig.spec.ts frontend/src/components/payment/__tests__/PaymentProviderDialog.spec.ts progress.md`

### Notes
- This continues moving payment provider behavior out of static Touch/Sub2API frontend hints. Provider defaults should be owned by Sub2API settings, backend provider behavior, or explicit admin configuration.

## 2026-06-20 Removed local public settings object fallback

### Done
- Removed the synthesized frontend `PublicSettings` fallback object from `useAppStore.fetchPublicSettings()`.
- Kept the real cache path unchanged: injected `window.__APP_CONFIG__`, cached Sub2API public settings, or a fresh Sub2API fetch.
- Added regression coverage so the frontend store does not reintroduce local defaults such as page-size options, monitor intervals, or site-name-derived settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/stores/__tests__/app.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/stores/app.ts frontend/src/stores/__tests__/app.spec.ts progress.md`

### Notes
- This trims the bootstrap configuration surface: missing public settings now remains missing instead of being expanded into a frontend-owned config object.

## 2026-06-20 Removed frontend API base env fallback

### Done
- Removed `import.meta.env.VITE_API_BASE_URL` from the runtime API base URL resolver.
- Removed `VITE_API_BASE_URL` from the frontend Vite env type surface.
- Kept the runtime resolution order as explicit public settings, injected `window.__APP_CONFIG__`, then same-origin `/api/v1`.
- Updated OAuth URL tests so default URLs no longer depend on a build-time API env value.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/api/__tests__/client.spec.ts src/api/__tests__/user.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "VITE_API_BASE_URL" frontend/src --glob '!**/__tests__/**' -S`
- `git diff --check -- frontend/src/api/client.ts frontend/src/api/__tests__/client.spec.ts frontend/src/api/__tests__/user.spec.ts frontend/src/vite-env.d.ts progress.md`

### Notes
- Standalone frontend development still uses the same-origin `/api/v1` path and Vite's dev proxy. Cross-origin API URLs should now be provided by Sub2API public settings/injected runtime config instead of a frontend build env.

## 2026-06-20 Stopped persisting table page size locally

### Done
- Removed `usePersistedPageSize` reads from the old `table-page-size` localStorage key.
- Changed `setPersistedPageSize` into a compatibility no-op so user-selected page size is no longer stored locally.
- Kept page-size resolution tied to Sub2API public settings via `table_default_page_size` and `table_page_size_options`.
- Removed the frontend's full built-in `[10,20,50,100]` selectable page-size fallback; when injected settings are missing, the frontend now uses only the configured/default page size as an emergency option.
- Removed the admin Settings form's local `[10,20,50,100]` initialization and load/save fallbacks; the form now stays empty until Sub2API settings provide page-size options.
- Updated `Pagination` docs to describe Sub2API public table settings instead of a local component default list.
- Added regression coverage that stale localStorage no longer overrides system table defaults and that page-size choices are not written back locally.
- Added SettingsView source-level regression coverage to prevent reintroducing frontend-owned table page-size defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/composables/__tests__/usePersistedPageSize.spec.ts src/utils/__tests__/tablePreferences.spec.ts`
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts src/composables/__tests__/usePersistedPageSize.spec.ts src/utils/__tests__/tablePreferences.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "table-page-size|table-page-size-source|localStorage\\.(getItem|setItem).*page-size|STORAGE_KEY" frontend/src/composables frontend/src/utils frontend/src/components/common frontend/src/views -S`
- `rg -n "DEFAULT_TABLE_PAGE_SIZE_OPTIONS|\\[10, 20, 50, 100\\]|return \\[\\.\\.\\.DEFAULT_TABLE_PAGE_SIZE_OPTIONS\\]" frontend/src/utils/tablePreferences.ts frontend/src/utils/__tests__/tablePreferences.spec.ts frontend/src/composables/usePersistedPageSize.ts -S`
- `rg -n "tablePageSizeOptionsInput = ref\\(\\"10, 20, 50, 100\\"\\)|table_page_size_options: \\[10, 20, 50, 100\\]|: \\[10, 20, 50, 100\\]|Available page size options \\(default: \\[10, 20, 50, 100\\]\\)" frontend/src/views/admin/SettingsView.vue frontend/src/components/common/README.md frontend/src/utils/tablePreferences.ts frontend/src/composables/usePersistedPageSize.ts -S`
- `git diff --check -- frontend/src/composables/usePersistedPageSize.ts frontend/src/composables/__tests__/usePersistedPageSize.spec.ts progress.md`

### Notes
- The frontend still has a single generic emergency default page size when injected public settings are missing. Normal selectable options should come from Sub2API public settings.

## 2026-06-20 Removed admin payment settings form defaults

### Done
- Removed frontend initialization defaults for payment limits and order policy values in `SettingsView`.
- Changed payment min/max/daily limits, max pending orders, order timeout, load-balance strategy, and cancel-rate limit fields to start from zero/empty form state until Sub2API settings are loaded.
- Added source-level regression coverage so SettingsView does not reintroduce frontend-owned defaults such as `10000`, `50000`, `30`, `round-robin`, or cancel-rate fallback values.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "payment_min_amount: 1,|payment_max_amount: 10000,|payment_daily_limit: 50000,|payment_max_pending_orders: 3,|payment_order_timeout_minutes: 30,|payment_load_balance_strategy: \\"round-robin\\"|payment_cancel_rate_limit_max: 10,|payment_cancel_rate_limit_window: 1,|payment_cancel_rate_limit_unit: \\"day\\"|payment_cancel_rate_limit_window_mode: \\"rolling\\"" frontend/src/views/admin/SettingsView.vue -S`
- `git diff --check -- frontend/src/views/admin/SettingsView.vue frontend/src/views/admin/__tests__/SettingsView.spec.ts progress.md`

### Notes
- Backend defaults and existing persisted settings are unchanged. This only removes admin frontend synthesis before Sub2API settings are loaded or when a field is absent.

## 2026-06-20 Removed admin site/user/recharge form defaults

### Done
- Removed frontend initialization defaults for `site_name`, `site_subtitle`, `default_concurrency`, `affiliate_rebate_rate`, and `payment_balance_recharge_multiplier` in `SettingsView`.
- Changed payment balance recharge multiplier save behavior so an empty value is no longer rewritten to `1` by the frontend.
- Added source-level regression coverage to keep brand, user, affiliate, and recharge multiplier defaults owned by Sub2API settings/backend behavior.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "site_name: \\"Sub2API\\"|site_subtitle: \\"Subscription to API Conversion Platform\\"|default_concurrency: 1,|affiliate_rebate_rate: 20,|payment_balance_recharge_multiplier: 1,|Number\\(form\\.payment_balance_recharge_multiplier\\) \\|\\| 1" frontend/src/views/admin/SettingsView.vue -S`
- `git diff --check -- frontend/src/views/admin/SettingsView.vue frontend/src/views/admin/__tests__/SettingsView.spec.ts progress.md`

### Notes
- Existing backend defaults and loaded persisted settings still apply. The admin Vue form now starts from empty/zero values until Sub2API settings are loaded.

## 2026-06-20 Removed admin AI fallback model defaults

### Done
- Removed frontend initialization defaults for `fallback_model_anthropic`, `fallback_model_openai`, `fallback_model_gemini`, and `fallback_model_antigravity` in `SettingsView`.
- Added source-level regression coverage so AI fallback model choices remain owned by Sub2API settings instead of frontend hardcoded model names.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "fallback_model_anthropic: \\"claude-3-5-sonnet-20241022\\"|fallback_model_openai: \\"gpt-4o\\"|fallback_model_gemini: \\"gemini-2.5-pro\\"|fallback_model_antigravity: \\"gemini-2.5-pro\\"" frontend/src/views/admin/SettingsView.vue -S`
- `git diff --check -- frontend/src/views/admin/SettingsView.vue frontend/src/views/admin/__tests__/SettingsView.spec.ts progress.md`

### Notes
- Backend model fallback defaults and persisted settings are unchanged. This only removes model choice synthesis in the admin frontend before settings are loaded.

## 2026-06-20 Removed admin connection/oauth/ops form defaults

### Done
- Removed frontend initialization defaults for SMTP port/TLS, WeChat frontend callback/mode/scopes, OIDC provider/scopes/token auth/signing algs/clock skew, GitHub/Google OAuth frontend callback URLs, identity patch, ops monitoring, fingerprint unification, subscription expiry notifications, and channel monitor settings in `SettingsView`.
- Changed the channel monitor save payload so an empty interval is no longer rewritten to `60` by the frontend.
- Added source-level regression coverage to keep connection, OAuth, notification, and ops defaults owned by Sub2API settings/backend behavior instead of the admin frontend.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "smtp_port: 587,|smtp_use_tls: true,|wechat_connect_mode: \\"open\\"|wechat_connect_scopes: \\"snsapi_login\\"|wechat_connect_frontend_redirect_url: \\"/auth/wechat/callback\\"|oidc_connect_provider_name: \\"OIDC\\"|oidc_connect_scopes: \\"openid email profile\\"|oidc_connect_frontend_redirect_url: \\"/auth/oidc/callback\\"|oidc_connect_token_auth_method: \\"client_secret_post\\"|oidc_connect_allowed_signing_algs: \\"RS256,ES256,PS256\\"|oidc_connect_clock_skew_seconds: 120,|github_oauth_frontend_redirect_url: \\"/auth/oauth/callback\\"|google_oauth_frontend_redirect_url: \\"/auth/oauth/callback\\"|enable_identity_patch: true,|ops_monitoring_enabled: true,|ops_realtime_monitoring_enabled: true,|ops_query_mode_default: \\"auto\\"|ops_metrics_interval_seconds: 60,|enable_fingerprint_unification: true,|subscription_expiry_notify_enabled: true,|channel_monitor_enabled: true,|channel_monitor_default_interval_seconds: 60,|Number\\(form\\.channel_monitor_default_interval_seconds\\) \\|\\| 60" frontend/src/views/admin/SettingsView.vue -S`
- `git diff --check -- frontend/src/views/admin/SettingsView.vue frontend/src/views/admin/__tests__/SettingsView.spec.ts progress.md`

### Notes
- Existing persisted settings still load and display normally. This only removes admin frontend synthesis before Sub2API settings are loaded or when fields are absent.

## 2026-06-20 Tightened Prompt Gallery API response ownership

### Done
- Removed the Prompt Gallery list response fallback that replaced missing `summary` with a frontend-owned empty summary.
- Removed the Prompt Gallery page-count fallback that rewrote missing `pages` to `1` before clamping.
- Added source-level regression coverage so prompt catalog summary/page metadata stays owned by the Sub2API prompt catalog API contract.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PromptCatalogView.spec.ts src/utils/__tests__/promptCatalogShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "data\\.summary \\|\\| emptySummary\\(\\)|data\\.pages \\|\\| 1" frontend/src/views/public/PromptCatalogView.vue -S`
- `git diff --check -- frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts progress.md`

### Notes
- The view still keeps an initial empty summary for pre-load rendering only. Loaded catalog facets, counts, image URLs, tags, and pagination metadata now rely on the API response.

## 2026-06-20 Removed Docsify Cloudbase search namespace default

### Done
- Replaced the hardcoded `cloudbase-docs-*` Docsify search namespace with a runtime namespace derived from public site settings (`site_name`), locale, and public settings version.
- Added normalization for namespace parts so search cache keys remain stable and storage-safe without embedding local product branding.
- Updated DocsView source tests to prevent reintroducing the Cloudbase namespace default.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts src/utils/__tests__/docsShell.spec.ts src/utils/__tests__/docsContentBasePath.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n 'cloudbase-docs|VITE_DOCS_CONTENT_VERSION' frontend/src/views/public/DocsView.vue -S`
- `git diff --check -- frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts progress.md`

### Notes
- This only changes Docsify search cache namespace ownership. Docs content path, Docsify runtime loading, and application route rewriting behavior are unchanged.

## 2026-06-20 Moved Docsify app-route allowlist into docs shell config

### Done
- Extended `docs_shell_config` parsing with `defaults.appRouteLinks`.
- Changed DocsView to derive application-route hash rewrite allowlist from public docs shell settings instead of a local `#/home`, `#/dashboard`, `#/register`, `#/purchase` constant.
- Added route-link normalization and filtering so only same-origin application paths are accepted.
- Kept `resolveDocsShellCopy` as a compatibility wrapper while DocsView now reads the richer `resolveDocsShellConfig` result.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts src/utils/__tests__/docsShell.spec.ts src/utils/__tests__/docsContentBasePath.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "const appRouteDocsLinks = new Set\\(\\['#/home'|#/purchase|#/register|cloudbase-docs|resolveDocsShellCopy\\(" frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts frontend/src/utils/docsShell.ts frontend/src/utils/__tests__/docsShell.spec.ts -S`
- `git diff --check -- frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts frontend/src/utils/docsShell.ts frontend/src/utils/__tests__/docsShell.spec.ts progress.md`

### Notes
- Existing docs labels still come from `docs_shell_config.labels`. Application route rewrites now require explicit `docs_shell_config.defaults.appRouteLinks`, keeping the Docs page from owning product navigation defaults.

## 2026-06-20 Moved Pricing route defaults into pricing shell config

### Done
- Extended `pricing_shell_config` parsing with `defaults.promptsPath` and `defaults.purchasePath`.
- Changed PricingView to render the prompt-catalog link and purchase CTAs only when those paths are provided by public pricing shell settings.
- Replaced hardcoded `/prompts` and `/purchase` route usage with configured, same-origin internal paths.
- Added route default filtering so external URLs, protocol-relative URLs, and malformed paths are ignored.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/PricingView.spec.ts src/utils/__tests__/pricingShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n 'to="/prompts"|path: .*/purchase|/purchase\\?tab=|/prompts' frontend/src/views/public/PricingView.vue frontend/src/utils/pricingShell.ts -S`
- `git diff --check -- frontend/src/utils/pricingShell.ts frontend/src/utils/__tests__/pricingShell.spec.ts frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts progress.md`

### Notes
- Pricing labels, groups, and button text were already config-driven. This change moves navigation defaults as well, so the page no longer owns Touch-era prompt/purchase route decisions.

## 2026-06-20 Moved Credits route defaults into credits shell config

### Done
- Extended `credits_shell_config` parsing with `defaults.purchasePath` and `defaults.ordersPath`.
- Changed CreditsView to render purchase/recharge/order links only when those paths are provided by public credits shell settings.
- Replaced hardcoded `/purchase`, `/purchase?tab=recharge`, and `/orders` usage with configured, same-origin internal paths.
- Added route default filtering so external URLs, protocol-relative URLs, and malformed paths are ignored.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/CreditsView.spec.ts src/utils/__tests__/creditsShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n 'to="/purchase"|to="/purchase\\?tab=recharge"|to="/orders"|/purchase\\?tab=recharge' frontend/src/views/user/CreditsView.vue frontend/src/utils/creditsShell.ts -S`
- `git diff --check -- frontend/src/utils/creditsShell.ts frontend/src/utils/__tests__/creditsShell.spec.ts frontend/src/views/user/CreditsView.vue frontend/src/views/user/__tests__/CreditsView.spec.ts progress.md`

### Notes
- Credits labels, action text, conversion text, and currency display were already config-driven. This change moves the remaining navigation defaults out of the frontend shell.

## 2026-06-20 Added auth shell redirect defaults for password auth

### Done
- Extended `auth_shell_config` parsing with `defaults.defaultRedirectPath` and `defaults.bindRedirectPath`.
- Changed LoginView password login and 2FA login success redirects to prefer `auth_shell_config.defaults.defaultRedirectPath`.
- Changed RegisterView successful registration redirect to prefer `auth_shell_config.defaults.defaultRedirectPath`.
- Added route default filtering so external URLs, protocol-relative URLs, and malformed paths are ignored.
- Kept the existing `/dashboard` compatibility fallback only when the public auth shell redirect default is absent.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "resolveRouteAuthRedirect\\(router\\.currentRoute\\.value\\.query\\.redirect\\)|router\\.push\\(DEFAULT_AUTH_REDIRECT_PATH\\)|resolveAuthShellLabels\\(settings\\.auth_shell_config" frontend/src/views/auth/LoginView.vue frontend/src/views/auth/RegisterView.vue -S`
- `git diff --check -- frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts frontend/src/views/auth/LoginView.vue frontend/src/views/auth/RegisterView.vue frontend/src/views/auth/__tests__/LoginView.turnstile.spec.ts frontend/src/views/auth/__tests__/RegisterView.auth-shell.spec.ts progress.md`

### Notes
- This is the first auth redirect migration slice. OAuth callback views and router guards still retain local `/dashboard`, `/profile`, `/login`, or admin-dashboard defaults and should be migrated separately.

## 2026-06-20 Added auth shell redirect defaults for email OAuth callback

### Done
- Extended `useAuthShellText` to expose full auth shell config/defaults while keeping `loadAuthShellLabels` compatibility.
- Changed `OAuthCallbackView` direct token callbacks and pending email OAuth completion flows to prefer `auth_shell_config.defaults.defaultRedirectPath`.
- Added coverage for direct token callbacks without redirect params and registration completion without backend redirect.
- Kept `/dashboard` compatibility fallback when public auth shell redirect default is absent.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/OAuthCallbackView.spec.ts src/utils/__tests__/authShell.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "loadAuthShellLabels\\(|completion\\.redirect \\|\\| DEFAULT_AUTH_REDIRECT_PATH|params\\.get\\('redirect'\\) \\|\\| DEFAULT_AUTH_REDIRECT_PATH|sanitizeAuthRedirectPath\\(redirect\\)" frontend/src/views/auth/OAuthCallbackView.vue frontend/src/composables/useAuthShellText.ts -S`
- `git diff --check -- frontend/src/composables/useAuthShellText.ts frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/__tests__/OAuthCallbackView.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts progress.md`

### Notes
- This migrates only GitHub/Google email OAuth callback. LinuxDo, OIDC, WeChat, DingTalk callback views and router guards still retain local `/dashboard`, `/profile`, `/login`, or admin-dashboard defaults.

## 2026-06-20 Added auth shell redirect defaults for LinuxDo callback

### Done
- Changed `LinuxDoCallbackView` to load full `auth_shell_config` before resolving callback redirects.
- Changed legacy token callbacks and pending-account OAuth flows to prefer `auth_shell_config.defaults.defaultRedirectPath` when no redirect is supplied.
- Changed bind completions to prefer `auth_shell_config.defaults.bindRedirectPath` when the backend does not return a bind redirect.
- Added regression coverage for legacy token callbacks without redirect params and updated bind completion expectations to use configured defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/utils/__tests__/authShell.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "loadAuthShellLabels\\(|DEFAULT_AUTH_BIND_REDIRECT_PATH\\)|DEFAULT_AUTH_REDIRECT_PATH\\)|DEFAULT_AUTH_REDIRECT_PATH|completion\\.redirect \\|\\| \\(route\\.query\\.redirect|params\\.get\\('redirect'\\) \\|\\| \\(route\\.query\\.redirect|sanitizeAuthRedirectPath\\(redirect \\|\\| redirectTo\\.value\\)" frontend/src/views/auth/LinuxDoCallbackView.vue -S`
- `git diff --check -- frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/__tests__/LinuxDoCallbackView.spec.ts frontend/src/composables/useAuthShellText.ts frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/__tests__/OAuthCallbackView.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts progress.md`

### Notes
- LinuxDo now shares the auth shell redirect-default path with password auth and email OAuth callback. OIDC, WeChat, DingTalk callback views and router guards still retain local redirect defaults.

## 2026-06-20 Added auth shell redirect defaults for OIDC callback

### Done
- Changed `OidcCallbackView` to load full `auth_shell_config` before resolving callback redirects.
- Changed legacy token callbacks and pending-account OAuth flows to prefer `auth_shell_config.defaults.defaultRedirectPath` when no redirect is supplied.
- Changed bind completions to prefer `auth_shell_config.defaults.bindRedirectPath` when the backend does not return a bind redirect.
- Added regression coverage for legacy token callbacks without redirect params and bind completions without backend redirect.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/OidcCallbackView.spec.ts src/utils/__tests__/authShell.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "loadAuthShellLabels\\(|DEFAULT_AUTH_BIND_REDIRECT_PATH\\)|DEFAULT_AUTH_REDIRECT_PATH\\)|DEFAULT_AUTH_REDIRECT_PATH|completion\\.redirect \\|\\| \\(route\\.query\\.redirect|params\\.get\\('redirect'\\) \\|\\| \\(route\\.query\\.redirect|sanitizeAuthRedirectPath\\(redirect \\|\\| redirectTo\\.value\\)" frontend/src/views/auth/OidcCallbackView.vue -S`
- `git diff --check -- frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/__tests__/OidcCallbackView.spec.ts progress.md`

### Notes
- OIDC now shares the auth shell redirect-default path with password auth, email OAuth callback, and LinuxDo. WeChat, DingTalk callback views and router guards still retain local redirect defaults.

## 2026-06-20 Added auth shell redirect defaults for WeChat callback

### Done
- Changed `WechatCallbackView` to load full `auth_shell_config` before resolving callback redirects.
- Changed legacy token callbacks, pending-account OAuth flows, and WeChat resume targets to prefer `auth_shell_config.defaults.defaultRedirectPath` when no redirect is supplied.
- Changed bind completions to prefer `auth_shell_config.defaults.bindRedirectPath` when the backend does not return a bind redirect.
- Added regression coverage for legacy token callbacks without redirect params and bind completions without backend redirect.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/WechatCallbackView.spec.ts src/utils/__tests__/authShell.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "loadAuthShellLabels\\(|DEFAULT_AUTH_BIND_REDIRECT_PATH\\)|DEFAULT_AUTH_REDIRECT_PATH\\)|DEFAULT_AUTH_REDIRECT_PATH|completion\\.redirect \\|\\| \\(route\\.query\\.redirect|params\\.get\\('redirect'\\) \\|\\| \\(route\\.query\\.redirect|sanitizeAuthRedirectPath\\(redirect \\|\\| redirectTo\\.value\\)" frontend/src/views/auth/WechatCallbackView.vue -S`
- `git diff --check -- frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/__tests__/WechatCallbackView.spec.ts progress.md`

### Notes
- WeChat now shares the auth shell redirect-default path with password auth, email OAuth callback, LinuxDo, and OIDC. DingTalk callback views and router guards still retain local redirect defaults.

## 2026-06-20 Added auth shell redirect defaults for DingTalk callback views

### Done
- Changed `DingTalkCallbackView` to load full `auth_shell_config` before resolving callback redirects.
- Changed legacy token callbacks, pending-account OAuth flows, and DingTalk email-completion hops to prefer `auth_shell_config.defaults.defaultRedirectPath` when no redirect is supplied.
- Changed bind completions to prefer `auth_shell_config.defaults.bindRedirectPath` when the backend does not return a bind redirect.
- Changed `DingTalkEmailCompletionView` successful account creation redirect fallback to use `auth_shell_config.defaults.defaultRedirectPath`.
- Added focused DingTalk callback regression coverage for legacy token callbacks without redirect params and bind completions without backend redirect.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/DingTalkCallbackView.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "loadAuthShellLabels\\(|DEFAULT_AUTH_BIND_REDIRECT_PATH\\)|DEFAULT_AUTH_REDIRECT_PATH\\)|DEFAULT_AUTH_REDIRECT_PATH|completion\\.redirect \\|\\| \\(route\\.query\\.redirect|params\\.get\\('redirect'\\) \\|\\| \\(route\\.query\\.redirect|sanitizeAuthRedirectPath\\(redirect \\|\\| redirectTo\\.value\\)" frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/DingTalkEmailCompletionView.vue -S`
- `git diff --check -- frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/DingTalkEmailCompletionView.vue frontend/src/views/auth/__tests__/DingTalkCallbackView.spec.ts progress.md`

### Notes
- Password auth plus GitHub/Google email OAuth, LinuxDo, OIDC, WeChat, and DingTalk callback redirects now share the Sub2API public `auth_shell_config.defaults` path. Router guards still retain local dashboard/login/admin-dashboard defaults.

## 2026-06-20 Moved router guard redirect defaults into auth shell config

### Done
- Extended `auth_shell_config.defaults` parsing with `loginPath`, `adminRedirectPath`, and `adminSettingsPath`.
- Added `resolveAuthRouteDefaults` and `resolveRoleHomeRedirect` so router guard fallback redirects share one public-settings-driven helper.
- Changed `router.beforeEach` guard redirects to use `auth_shell_config.defaults` for unauthenticated login redirects, authenticated user/admin home redirects, setup completion redirects, payment-disabled redirects, risk-control redirects, simple-mode redirects, and backend-mode blocking redirects.
- Kept static route definitions such as `/login`, `/dashboard`, and `/admin/dashboard` unchanged; only guard fallback decisions moved out of local constants.
- Added guard tests proving configured login/user/admin defaults override the frontend fallbacks and source checks preventing `next('/login')` / `next('/dashboard')` from returning to the real guard.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/router/__tests__/guards.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "next\\('/login'\\)|next\\('/dashboard'\\)|next\\('/admin/dashboard'\\)|authStore\\.isAdmin \\? '/admin/dashboard' : '/dashboard'|path: '/login'," frontend/src/router/index.ts -S`
- `git diff --check -- frontend/src/router/index.ts frontend/src/router/setupRedirect.ts frontend/src/router/__tests__/guards.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts progress.md`

### Notes
- Router guard runtime defaults now come from Sub2API public `auth_shell_config.defaults`. Remaining local route strings in `router/index.ts` are route definitions and aliases, not guard fallback configuration.

## 2026-06-20 Reused auth route defaults in public Vue shells

### Done
- Added `useAuthRouteDefaults` so public pages can reuse the same `auth_shell_config.defaults` parsing as the router guard.
- Changed Home, Docs, Models Plaza, and Prompt Catalog header/footer account links to use configured `loginPath`, `defaultRedirectPath`, and `adminRedirectPath` instead of local `/login`, `/dashboard`, and `/admin/dashboard` fallback decisions.
- Kept static navigation routes such as `/home`, `/models`, `/docs`, and legal links unchanged.
- Added source-level regression checks so these public pages keep using `useAuthRouteDefaults` and do not reintroduce local account-link fallback expressions.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/HomeView.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/router/__tests__/guards.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "isAuthenticated \\? dashboardPath : '/login'|to=\\\"/login\\\"|authStore\\.isAdmin \\? '/admin/dashboard' : '/dashboard'|isAdmin\\.value \\? '/admin/dashboard' : '/dashboard'" frontend/src/views/HomeView.vue frontend/src/views/public/DocsView.vue frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/PromptCatalogView.vue -S`
- `git diff --check -- frontend/src/composables/useAuthRouteDefaults.ts frontend/src/views/HomeView.vue frontend/src/views/public/DocsView.vue frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/PromptCatalogView.vue frontend/src/views/__tests__/HomeView.spec.ts frontend/src/views/public/__tests__/DocsView.spec.ts frontend/src/views/public/__tests__/ModelsPlazaView.spec.ts frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/router/index.ts frontend/src/router/setupRedirect.ts frontend/src/router/__tests__/guards.spec.ts frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts progress.md`

### Notes
- This removes account-link fallback decisions from the main public Vue shells. Layout navigation components still contain fixed route menu entries like sidebar dashboard/profile/admin routes, which are menu definitions rather than public account-link fallbacks.

## 2026-06-20 Reused auth route defaults in layout and fallback entry points

### Done
- Changed `AppHeader` logout redirect to use `auth_shell_config.defaults.loginPath` through `useAuthRouteDefaults`.
- Changed `LegalDocumentView` login CTA to resolve `loginPath` from the page's public settings payload instead of hardcoding `/login`.
- Changed `NotFoundView` dashboard action to resolve the user/admin destination through `useAuthRouteDefaults`.
- Added source-level regression tests for AppHeader, LegalDocumentView, and NotFoundView so these fallback links do not return to local `/login`, `/dashboard`, or `/admin/dashboard` constants.
- Updated stale AppHeader source assertions to match the current disabled GitHub/profile shortcut behavior.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/layout/__tests__/AppHeader.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts src/views/__tests__/NotFoundView.spec.ts src/router/__tests__/guards.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "router\\.push\\('/login'\\)|to=\\\"/login\\\"|to=\\\"/dashboard\\\"|('/admin/dashboard'|\\\"/admin/dashboard\\\")" frontend/src/components/layout/AppHeader.vue frontend/src/views/public/LegalDocumentView.vue frontend/src/views/NotFoundView.vue -S`
- `git diff --check -- frontend/src/components/layout/AppHeader.vue frontend/src/components/layout/__tests__/AppHeader.spec.ts frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/__tests__/LegalDocumentView.spec.ts frontend/src/views/NotFoundView.vue frontend/src/views/__tests__/NotFoundView.spec.ts frontend/src/composables/useAuthRouteDefaults.ts progress.md`

### Notes
- Remaining `/dashboard`, `/profile`, `/keys`, and `/admin/...` strings in `AppSidebar` and dashboard quick actions are concrete menu destinations, not auth/account fallback defaults. Moving those would require configurable navigation schemas rather than auth route defaults.

## 2026-06-20 Moved dashboard quick action destinations into dashboard shell config

### Done
- Extended `dashboard_shell_config` parsing with `defaults.quickActions.createApiKeyPath`, `usagePath`, and `redeemPath`.
- Changed `UserDashboardQuickActions` to navigate through configured quick action defaults instead of hardcoded `/keys`, `/usage`, and `/redeem` button handlers.
- Changed `DashboardView` to resolve the full dashboard shell config once and pass quick action defaults into the dashboard quick action component.
- Updated admin runtime-settings placeholder and hint copy so operators can configure quick action destinations through Sub2API settings.
- Added regression tests for configured quick action paths, unsafe path fallback, and source checks that prevent reintroducing hardcoded quick action button destinations.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/components/user/dashboard/dashboardShellLabels.ts frontend/src/components/user/dashboard/UserDashboardQuickActions.vue frontend/src/views/user/DashboardView.vue frontend/src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts frontend/src/views/user/__tests__/dashboardNoHero.spec.ts frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts`

### Notes
- This keeps the dashboard UI in the Vue frontend, but removes another Touch-era local navigation decision from the component layer. The sidebar remains a larger static menu-definition surface and should be migrated separately if a full navigation schema is introduced.

## 2026-06-20 Moved home shell navigation links into public settings

### Done
- Verified the unified runtime shape: `apps/touch` is absent and `apps/` is empty, so Touch is no longer a standalone Next subapp in this repository.
- Verified the Touch-specific API surface is removed from active route registration; `/api/v1/touch/*` appears only in negative route tests and historical progress records.
- Extended `home_shell_config` parsing with `defaults.links` for home anchor, models path, experience anchor, docs fallback path, terms path, and privacy path.
- Changed `HomeView` nav and footer link composition to consume `home_shell_config.defaults.links` instead of local `/models`, `/docs`, and `/legal/*` path literals.
- Updated runtime-settings placeholder and hint copy so operators can configure Home shell links through Sub2API public settings.
- Added regression coverage for configured Home links and unsafe-link fallback, plus HomeView source checks that keep those route decisions out of the Vue view layer.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/homeShell.spec.ts src/views/__tests__/HomeView.spec.ts src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts`
- `pnpm run frontend:typecheck`
- `go test ./internal/server/routes` from `backend/`
- `git diff --check -- frontend/src/utils/homeShell.ts frontend/src/utils/__tests__/homeShell.spec.ts frontend/src/views/HomeView.vue frontend/src/views/__tests__/HomeView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts frontend/src/components/user/dashboard/dashboardShellLabels.ts frontend/src/components/user/dashboard/UserDashboardQuickActions.vue frontend/src/views/user/DashboardView.vue frontend/src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts frontend/src/views/user/__tests__/dashboardNoHero.spec.ts progress.md`

### Notes
- Attempted `go test ./backend/internal/server/routes` from the repository root first; it failed because the root is not the Go module. The successful route verification was run from `backend/`.
- Remaining frontend shell ownership is now mostly larger UI composition and static route definitions, especially `AppSidebar` menu structure and public page layouts. Prompt Catalog already consumes Sub2API summary/facets/items and has no obvious local data-source fallback left in the active view.

## 2026-06-20 Reused dashboard shell usage path in recent usage card

### Done
- Changed `UserDashboardRecentUsage` to accept a configured `usagePath` prop instead of hardcoding `to="/usage"`.
- Changed `DashboardView` to pass `dashboard_shell_config.defaults.quickActions.usagePath` to both the quick action usage button and the recent-usage "view all" link.
- Added regression checks so the recent usage card keeps using the configured dashboard shell path and does not reintroduce a local `/usage` route literal.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/components/user/dashboard/UserDashboardRecentUsage.vue frontend/src/views/user/DashboardView.vue frontend/src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts frontend/src/views/user/__tests__/dashboardNoHero.spec.ts progress.md`

### Notes
- This keeps dashboard behavior consistent with the Sub2API-managed dashboard shell defaults. Static router definitions and `AppSidebar` menu paths remain separate route/menu schema work.

## 2026-06-20 Finished remaining HomeView link usage of home shell defaults

### Done
- Changed the Home hero secondary CTA to use `home_shell_config.defaults.links.modelsPath` instead of template-level `to="/models"`.
- Changed the Home footer legal anchors to use `home_shell_config.defaults.links.termsPath` and `privacyPath` instead of template-level `/legal/*` hrefs.
- Added source guards so HomeView does not reintroduce direct `/models` or `/legal/*` template literals while the parsed shell defaults exist.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/homeShell.spec.ts src/views/__tests__/HomeView.spec.ts src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "to=\\\"/models\\\"|href=\\\"/legal/terms\\\"|href=\\\"/legal/privacy-policy\\\"|to=\\\"/usage\\\"" frontend/src/views/HomeView.vue frontend/src/components/user/dashboard/UserDashboardRecentUsage.vue -S`
- `git diff --check -- frontend/src/views/HomeView.vue frontend/src/views/__tests__/HomeView.spec.ts frontend/src/components/user/dashboard/UserDashboardRecentUsage.vue frontend/src/views/user/DashboardView.vue frontend/src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts frontend/src/views/user/__tests__/dashboardNoHero.spec.ts progress.md`

### Notes
- Public pages still contain brand links back to `/home` as static route definitions. Those should be handled with a shared public-shell route schema rather than piecemeal auth/default-link cleanup.

## 2026-06-20 Moved public brand home links into auth route defaults

### Done
- Extended `auth_shell_config.defaults` with `homePath`, surfaced through `resolveAuthRouteDefaults` and `useAuthRouteDefaults`.
- Changed public/utility brand links in Docs, Models Plaza, Prompt Catalog, Legal Document, Pricing, Image Generator, and Key Usage views to use `authRouteDefaults.homePath` instead of static `to="/home"`.
- Updated auth shell runtime-settings placeholder and hint copy so admins can configure the shared public home route through Sub2API public settings.
- Added parser, router-default, and page source regression checks for `homePath`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/router/__tests__/guards.spec.ts src/views/public/__tests__/DocsView.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts src/views/public/__tests__/PricingView.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts src/views/__tests__/KeyUsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "to=\\\"/home\\\"|href=\\\"/home\\\"|router-link to=\\\"/home\\\"|RouterLink to=\\\"/home\\\"" frontend/src -S` now only matches negative source assertions in tests.
- `git diff --check -- frontend/src/utils/authShell.ts frontend/src/utils/__tests__/authShell.spec.ts frontend/src/router/setupRedirect.ts frontend/src/router/__tests__/guards.spec.ts frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/__tests__/ModelsPlazaView.spec.ts frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/PromptCatalogView.spec.ts frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/__tests__/LegalDocumentView.spec.ts frontend/src/views/public/PricingView.vue frontend/src/views/public/__tests__/PricingView.spec.ts frontend/src/views/public/ImageGeneratorView.vue frontend/src/views/public/__tests__/ImageGeneratorView.spec.ts frontend/src/views/KeyUsageView.vue frontend/src/views/__tests__/KeyUsageView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts progress.md`

### Notes
- This removes the remaining public-page brand-home route literal from active Vue views. Static route definitions in `router/index.ts`, docs, tests, and app-menu schemas remain separate concerns.

## 2026-06-20 Moved sidebar shell paths into auth route defaults

### Done
- Extended `auth_shell_config.defaults` with regular user navigation paths for keys, usage, available channels/groups, subscriptions, purchase, orders, redeem, affiliate, and profile.
- Extended `auth_shell_config.defaults` with `adminRuntimeSettingsPath`, alongside the existing admin dashboard/settings defaults.
- Surfaced those paths through `resolveAuthRouteDefaults` with existing route fallbacks, so Sub2API public settings can configure user navigation without breaking unloaded-settings behavior.
- Changed `AppSidebar` regular-user and admin personal-account menu construction to use the configured route defaults instead of local user route literals.
- Changed the primary admin dashboard/runtime-settings/settings sidebar entries to use the configured route defaults.
- Kept the onboarding API-key anchor tied to the configured API keys path, including simple-mode admin insertion and menu click handling.
- Updated runtime-settings placeholder and hint copy so admins can see the new configurable navigation defaults.
- Added parser, router-default, and sidebar source regression checks to prevent reintroducing hardcoded regular user menu paths and primary admin entry paths.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/router/__tests__/guards.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/authShell.ts frontend/src/router/setupRedirect.ts frontend/src/components/layout/AppSidebar.vue frontend/src/components/layout/__tests__/AppSidebar.spec.ts frontend/src/utils/__tests__/authShell.spec.ts frontend/src/router/__tests__/guards.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts`
- `rg -n "path: '/dashboard'|path: '/keys'|path: '/usage'|path: '/available-channels'|path: '/available-groups'|path: '/subscriptions'|path: '/purchase'|path: '/orders'|path: '/redeem'|path: '/affiliate'|path: '/profile'|path: '/admin/dashboard'|path: '/admin/runtime-settings'|path: '/admin/settings'|item.path === '/keys'" frontend/src/components/layout/AppSidebar.vue -S` returned no matches.

### Notes
- This moves the user-facing sidebar routes and the primary admin dashboard/settings entries into Sub2API public settings. The broader admin menu structure is still local Vue menu composition and should be handled as a separate, larger navigation schema if we want a fully settings-driven sidebar.

## 2026-06-20 Reused auth route defaults in header account shortcuts

### Done
- Changed the `AppHeader` account dropdown profile and API key shortcuts to use `authRouteDefaults.profilePath` and `authRouteDefaults.apiKeysPath`.
- Added regression checks preventing hidden dropdown shortcuts from falling back to static `to="/profile"` or `to="/keys"` literals if they are re-enabled later.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/layout/__tests__/AppHeader.spec.ts src/utils/__tests__/authShell.spec.ts src/router/__tests__/guards.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- The account shortcuts are currently disabled by `showDropdownAccountLinks`, but their route targets now follow the same Sub2API-managed defaults as the sidebar.

## 2026-06-20 Reused auth route defaults in payment result actions

### Done
- Changed `PaymentResultView` action buttons to navigate through `authRouteDefaults.purchasePath` and `authRouteDefaults.ordersPath`.
- Added source regression checks preventing the payment result page from reintroducing direct `router.push('/purchase')` or `router.push('/orders')` calls.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentResultView.spec.ts src/components/layout/__tests__/AppHeader.spec.ts src/utils/__tests__/authShell.spec.ts src/router/__tests__/guards.spec.ts`
- `pnpm run frontend:typecheck`

### Notes
- Other payment flow pages still contain direct recharge/order navigation and should be migrated to the same route defaults in follow-up slices.

## 2026-06-20 Reused auth route defaults in payment flow recharge links

### Done
- Changed `PaymentQRCodeView`, `StripePaymentView`, `AirwallexPaymentView`, and `UserOrdersView` recharge/back buttons to use `authRouteDefaults.purchasePath`.
- Changed QR payment cancellation success navigation to use the same configured purchase path.
- Added source regression checks across the affected payment views so direct `router.push('/purchase')` calls are not reintroduced.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentQRCodeView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "router\\.push\\('/purchase'|router\\.push\\('/orders'" frontend/src/views/user/PaymentQRCodeView.vue frontend/src/views/user/StripePaymentView.vue frontend/src/views/user/AirwallexPaymentView.vue frontend/src/views/user/UserOrdersView.vue frontend/src/views/user/PaymentResultView.vue -S` returned no matches.

### Notes
- The remaining direct payment-route literals are outside this slice, including callback normalization and subscription-specific query construction. Those should be migrated carefully because they preserve payment provider return compatibility and tab/group query state.

## 2026-06-20 Reused auth route defaults in subscription renewal and WeChat callback fallback

### Done
- Changed `SubscriptionsView` renewal navigation to use `authRouteDefaults.purchasePath` while preserving the existing `tab=subscription` and `group` query state.
- Changed `WechatPaymentCallbackView` empty/invalid redirect fallback and error recovery button to use `auth_shell_config.defaults.purchasePath`.
- Preserved explicit valid callback redirect paths as-is, and preserved legacy `/payment?...` compatibility by merging its query into the configured purchase fallback.
- Added tests for configured subscription renewal navigation, configured WeChat callback fallback, error recovery navigation, and source guards against reintroducing direct `/purchase` fallback returns.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/SubscriptionsView.spec.ts src/views/auth/__tests__/WechatPaymentCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "router\\.push\\('/purchase'|router\\.replace\\('/purchase'|path: '/purchase'|return '/purchase'|'/purchase\\?" frontend/src/views/user frontend/src/views/auth frontend/src/views/public frontend/src/components -S` now reports only tests and `PaymentView` self-route expectations.

### Notes
- `PaymentView` still naturally has `/purchase` in route-state tests because it is the purchase page itself. This is route identity, not an outbound navigation decision.

## 2026-06-20 Reused auth route defaults in API guide/test navigation

### Done
- Changed `ApiGuideView` manage-key CTA and empty-state action to use `authRouteDefaults.apiKeysPath`.
- Changed `ApiTestView` empty-state manage-key action to use `authRouteDefaults.apiKeysPath`.
- Changed `ApiTestView` usage inspection link to use `authRouteDefaults.usagePath` while preserving the selected key and date query.
- Added source regression checks preventing direct `to="/keys"`, `action-to="/keys"`, and `path: '/usage'` route literals in the active API guide/test views.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiGuideView.spec.ts src/views/user/__tests__/ApiTestView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "to=\\\"/keys\\\"|action-to=\\\"/keys\\\"|path: '/usage'|router\\.push\\('/usage'" frontend/src/views/user frontend/src/components -S` now reports only defaults, test guards, and docs examples.

### Notes
- `/gateway-test` remains a real route identity and has not been moved into auth route defaults. If we want that configurable too, it should be added deliberately as an API tools shell default rather than overloading auth defaults further.

## 2026-06-20 Moved image workspace and subscription mini shortcuts into public settings

### Done
- Added `resolveWorkspaceShellDefaults` for `workspace_shell_config.defaults.catalogPath`, with internal-path validation.
- Changed public `ImageGeneratorView` catalog/back links to use the configured workspace catalog path instead of static `to="/prompts"`.
- Updated runtime-settings placeholder and hint copy to document `defaults.catalogPath`.
- Changed `SubscriptionProgressMini` "view all" link to use `authRouteDefaults.subscriptionsPath`.
- Added source and parser regression tests for the workspace catalog path and subscription mini shortcut.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/imageWorkspaceShell.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts src/components/common/__tests__/SubscriptionProgressMini.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "to=\\\"/prompts\\\"|to=\\\"/subscriptions\\\"" frontend/src/views/public/ImageGeneratorView.vue frontend/src/components/common/SubscriptionProgressMini.vue frontend/src/views/public/__tests__/ImageGeneratorView.spec.ts frontend/src/components/common/__tests__/SubscriptionProgressMini.spec.ts -S` now reports only negative source assertions in tests.

### Notes
- `ImageGeneratorView` still owns the workspace UI composition locally; this change only moves the catalog route decision into Sub2API public settings.

## 2026-06-20 Moved auth page navigation defaults into public settings

### Done
- Extended `auth_shell_config.defaults` with `registerPath`, `forgotPasswordPath`, and `emailVerifyPath`.
- Changed Login, Register, Forgot Password, Reset Password, Email Verify, OAuth Callback, and Auth Popup views to use Sub2API public settings for auth-page navigation instead of hardcoded cross-page links.
- Updated runtime-settings placeholder and hint copy so admins can configure the new auth route defaults.
- Added parser, guard, and page regression tests covering configured auth route defaults and source guards against reintroducing local static auth navigation.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/router/__tests__/guards.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts src/views/auth/__tests__/ForgotPasswordView.auth-shell.spec.ts src/views/auth/__tests__/ResetPasswordView.auth-shell.spec.ts src/views/auth/__tests__/EmailVerifyView.spec.ts src/views/auth/__tests__/OAuthCallbackView.spec.ts src/views/auth/__tests__/AuthPopupView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/authShell.ts frontend/src/router/setupRedirect.ts frontend/src/views/auth/LoginView.vue frontend/src/views/auth/RegisterView.vue frontend/src/views/auth/ForgotPasswordView.vue frontend/src/views/auth/ResetPasswordView.vue frontend/src/views/auth/EmailVerifyView.vue frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/AuthPopupView.vue frontend/src/views/auth/__tests__/LoginView.turnstile.spec.ts frontend/src/views/auth/__tests__/RegisterView.auth-shell.spec.ts frontend/src/views/auth/__tests__/ForgotPasswordView.auth-shell.spec.ts frontend/src/views/auth/__tests__/ResetPasswordView.auth-shell.spec.ts frontend/src/views/auth/__tests__/EmailVerifyView.spec.ts frontend/src/views/auth/__tests__/OAuthCallbackView.spec.ts frontend/src/views/auth/__tests__/AuthPopupView.spec.ts frontend/src/utils/__tests__/authShell.spec.ts frontend/src/router/__tests__/guards.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts progress.md`
- `rg -n "to=\"/login\"|to=\"/forgot-password\"|to=\"/register\"|router\\.push\\('/register'|router\\.push\\('/email-verify'|router\\.replace\\('/login'|path: '/login'" frontend/src/views/auth/*.vue frontend/src/router -S` now reports only route registration and fallback defaults.

### Notes
- `/login`, `/register`, `/forgot-password`, and `/email-verify` remain real route identities. This slice removed outbound auth-page navigation decisions from the active auth views, not the route definitions themselves.

## 2026-06-20 Reused auth route defaults in Docsify chrome

### Done
- Changed the Docsify `nameLink` target from the local `/home` literal to `auth_shell_config.defaults.homePath` via `authRouteDefaults`.
- Updated Docs view source regression coverage so the internal Docsify site-name link stays aligned with the configured public home path.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/DocsView.spec.ts src/utils/__tests__/docsShell.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts progress.md`
- `rg -n "nameLink: '/home'|to=\"/home\"|'/home'" frontend/src/views/public/DocsView.vue frontend/src/views/public/__tests__/DocsView.spec.ts -S` now reports only negative source assertions in tests.

### Notes
- The `/docs` path itself remains a real Vue Router route identity. It was not moved into public settings because configuring it without registering a matching route would break direct Docs navigation.

## 2026-06-20 Centralized image workspace catalog fallback

### Done
- Moved the Image Generator catalog fallback path out of `ImageGeneratorView` and into `resolveWorkspaceShellDefaults`.
- Added `DEFAULT_WORKSPACE_CATALOG_PATH` so the shared workspace shell parser owns the safe `/prompts` fallback when `workspace_shell_config.defaults.catalogPath` is missing or unsafe.
- Tightened Image Generator source coverage so the page does not reintroduce a local `|| '/prompts'` route fallback.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/imageWorkspaceShell.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/imageWorkspaceShell.ts frontend/src/views/public/ImageGeneratorView.vue frontend/src/utils/__tests__/imageWorkspaceShell.spec.ts frontend/src/views/public/__tests__/ImageGeneratorView.spec.ts progress.md`
- `rg -n "\\|\\| '/prompts'|to=\"/prompts\"|catalogPath\\?:|resolveWorkspaceShellDefaults" frontend/src/utils/imageWorkspaceShell.ts frontend/src/views/public/ImageGeneratorView.vue frontend/src/utils/__tests__/imageWorkspaceShell.spec.ts frontend/src/views/public/__tests__/ImageGeneratorView.spec.ts -S` now reports only parser/test references and negative assertions.

### Notes
- This does not remove the Image Generator UI shell; it narrows route fallback ownership so page code consumes Sub2API runtime shell parsing rather than carrying its own local path default.

## 2026-06-20 Moved API guide/test cross-links into shell defaults

### Done
- Added `api_guide_shell_config.defaults.testPath` parsing with a safe `/gateway-test` fallback.
- Added `api_test_shell_config.defaults.guidePath` parsing with a safe `/gateway-guide` fallback.
- Changed `ApiGuideView` and `ApiTestView` cross-page buttons to use those Sub2API public settings defaults instead of local static route literals.
- Updated Runtime Settings placeholder and hint copy to document the new guide/test route defaults.
- Added parser and source regression tests to prevent reintroducing static API tool cross-links in the active views.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiGuideShell.spec.ts src/utils/__tests__/apiTestShell.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/views/user/__tests__/ApiTestView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/apiGuideShell.ts frontend/src/utils/apiTestShell.ts frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/ApiTestView.vue frontend/src/utils/__tests__/apiGuideShell.spec.ts frontend/src/utils/__tests__/apiTestShell.spec.ts frontend/src/views/user/__tests__/ApiGuideView.spec.ts frontend/src/views/user/__tests__/ApiTestView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts progress.md`
- `rg -n "to=\"/gateway-test\"|path: '/gateway-test'|to=\"/gateway-guide\"|DEFAULT_API_(GUIDE_TEST|TEST_GUIDE)_PATH|resolveAPI(Guide|Test)ShellDefaults" frontend/src/utils/apiGuideShell.ts frontend/src/utils/apiTestShell.ts frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/ApiTestView.vue frontend/src/utils/__tests__/apiGuideShell.spec.ts frontend/src/utils/__tests__/apiTestShell.spec.ts frontend/src/views/user/__tests__/ApiGuideView.spec.ts frontend/src/views/user/__tests__/ApiTestView.spec.ts -S`

### Notes
- `/gateway-test` and `/gateway-guide` remain real Vue Router route identities. This slice moves the API tool cross-link decisions out of the pages and into Sub2API-managed shell config defaults.

## 2026-06-20 Moved WeChat payment return fallback into auth route defaults

### Done
- Changed the WeChat payment OAuth authorize URL builder to use `auth_shell_config.defaults.purchasePath` through `useAuthRouteDefaults`.
- Preserved server-provided `redirect` parameters exactly as before; the configured purchase path is only used when the authorize URL omits a redirect.
- Added a regression test for configured purchase-path fallback and a source guard against reintroducing `|| '/purchase'` in `PaymentView`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/PaymentView.spec.ts src/router/__tests__/guards.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/user/PaymentView.vue frontend/src/views/user/__tests__/PaymentView.spec.ts progress.md`

### Notes
- `/purchase` remains the router fallback default in `FALLBACK_AUTH_ROUTE_DEFAULTS`. This slice removes the local payment-page fallback and keeps the route decision under Sub2API public auth shell settings.

## 2026-06-20 Reused auth route defaults for dashboard quick-action fallbacks

### Done
- Extended `resolveDashboardShellConfig` to accept quick-action fallback paths from the caller.
- Changed `DashboardView` so dashboard quick actions fall back to `auth_shell_config` route defaults for API keys, usage, and redeem pages before using the final built-in defaults.
- Kept `dashboard_shell_config.defaults.quickActions` as the highest-priority override.
- Added parser and source regression tests covering auth-route fallback handoff.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/components/user/dashboard/dashboardShellLabels.ts frontend/src/views/user/DashboardView.vue frontend/src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts progress.md`

### Notes
- The dashboard components still receive concrete routes through props; this change only moves fallback route ownership from the dashboard shell module toward shared Sub2API auth shell defaults.

## 2026-06-20 Moved public integration injection switch into Sub2API settings

### Done
- Added `web_public_integrations_enabled` as the Sub2API runtime setting behind public `public_integrations_enabled`.
- Exposed the setting through public settings, SSR injection payload, admin settings read/update DTOs, and the Runtime Settings admin page.
- Changed `PublicIntegrations` to use `cachedPublicSettings.public_integrations_enabled !== false` instead of frontend env flags.
- Kept the default enabled so existing production integration injection behavior remains unchanged unless an admin disables it.
- Added frontend component/source tests and backend public/update/schema coverage.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/common/__tests__/PublicIntegrations.spec.ts src/utils/__tests__/publicIntegrations.spec.ts src/views/admin/__tests__/RuntimeSettingsView.spec.ts`
- `go test -tags unit ./internal/service -run 'TestSettingService_GetPublicSettings_ExposesWebRuntimeSettings|TestSettingService_GetPublicSettings_IgnoresLegacyTouchRuntimeKeys|TestSettingService_GetPublicSettings_UsesOnlyGenericRuntimeKeys|TestSettingService_UpdateSettings_PersistsWebRuntimeSettings' -count=1` from `backend/`
- `go test -tags unit ./internal/handler/dto -run TestPublicSettingsInjectionPayload_SchemaDoesNotDrift -count=1` from `backend/`
- `go test -tags unit ./internal/handler/admin -run 'Test.*Runtime|Test.*Settings|Test.*AuthSource' -count=1` from `backend/`

### Notes
- This removes one remaining frontend bootstrap/env decision. The integration script IDs and individual provider settings still live in Sub2API public runtime settings.

## 2026-06-20 Removed ops websocket frontend env fallback

### Done
- Changed the admin ops QPS websocket URL builder to derive its default origin/path from Sub2API public `api_base_url` via `resolveApiBaseUrl`.
- Kept explicit `wsBaseUrl` overrides for tests and special deployments, but removed the `VITE_WS_BASE_URL` build-time fallback.
- Added regression coverage for same-origin defaults, injected remote `api_base_url`, explicit websocket host overrides, and a source guard against reintroducing `VITE_WS_BASE_URL`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/api/admin/__tests__/ops.spec.ts src/api/__tests__/client.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/api/admin/ops.ts frontend/src/api/admin/__tests__/ops.spec.ts progress.md`
- `rg -n "import\\.meta\\.env|VITE_|NEXT_PUBLIC_|SUB2API_BASE_URL" frontend/src -g '!**/__tests__/**' -S`

### Notes
- This removes another frontend runtime env decision. If a deployment needs a websocket origin that is different from the configured Sub2API API origin, it should pass an explicit `wsBaseUrl` or add a dedicated DB-backed runtime setting rather than using frontend build env.
- The remaining non-test env reads are Vite router base (`import.meta.env.BASE_URL`) and dev-only route-prefetch logging (`import.meta.env.DEV`), not Touch/Sub2API runtime bootstrap configuration.

## 2026-06-20 Moved payment result route into auth shell defaults

### Done
- Added `auth_shell_config.defaults.paymentResultPath` to the shared frontend route-default parser and fallback map.
- Changed payment completion redirects in `PaymentView`, `StripePaymentView`, `PaymentQRCodeView`, `StripePaymentInline`, and `StripePopupView` to use `authRouteDefaults.paymentResultPath`.
- Updated Runtime Settings auth shell JSON placeholders and hints to document `paymentResultPath`.
- Added source and parser regression checks so payment pages no longer own local `/payment/result` route literals.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/router/__tests__/guards.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/components/payment/__tests__/StripePaymentInline.spec.ts src/views/user/__tests__/StripePopupView.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/utils/authShell.ts frontend/src/router/setupRedirect.ts frontend/src/views/user/PaymentView.vue frontend/src/views/user/StripePaymentView.vue frontend/src/views/user/PaymentQRCodeView.vue frontend/src/components/payment/StripePaymentInline.vue frontend/src/views/user/StripePopupView.vue frontend/src/utils/__tests__/authShell.spec.ts frontend/src/router/__tests__/guards.spec.ts frontend/src/views/user/__tests__/PaymentView.spec.ts frontend/src/views/user/__tests__/StripePaymentView.spec.ts frontend/src/views/user/__tests__/PaymentQRCodeView.spec.ts frontend/src/components/payment/__tests__/StripePaymentInline.spec.ts frontend/src/views/user/__tests__/StripePopupView.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts progress.md`
- `rg -n "path: '/payment/result'|window\\.location\\.origin \\+ '/payment/result|return_url: .*payment/result|paymentResultPath" frontend/src/views/user frontend/src/views/auth frontend/src/components/payment frontend/src/utils/authShell.ts frontend/src/router/setupRedirect.ts -S`

### Notes
- `/payment/result` remains the built-in fallback route identity. Active payment UI now consumes the Sub2API-managed auth shell route default instead of owning page-local result paths.

## 2026-06-20 Moved DingTalk OAuth internal routes into auth shell defaults

### Done
- Added `auth_shell_config.defaults.dingtalkCallbackPath` and `dingtalkEmailCompletionPath` to the shared route-default parser and fallback map.
- Changed the DingTalk callback and email-completion views to route through configured auth shell defaults instead of page-owned internal path literals.
- Updated Runtime Settings auth shell JSON placeholders and hints to document the DingTalk route defaults.
- Added parser, router-default, and source regression checks so the DingTalk OAuth flow stays configurable from Sub2API public settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authShell.spec.ts src/router/__tests__/guards.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "'/auth/dingtalk/email-completion\\?|\\\"/auth/dingtalk/email-completion\\?|path: '/auth/dingtalk/callback'|path: \\\"/auth/dingtalk/callback\\\"|dingtalkCallbackPath|dingtalkEmailCompletionPath" frontend/src/views/auth frontend/src/utils/authShell.ts frontend/src/router/setupRedirect.ts frontend/src/views/auth/__tests__ frontend/src/utils/__tests__/authShell.spec.ts frontend/src/router/__tests__/guards.spec.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts -S`

### Notes
- `/auth/dingtalk/callback` and `/auth/dingtalk/email-completion` remain built-in fallback route identities. Active DingTalk auth UI now consumes the Sub2API-managed auth shell route defaults.

## 2026-06-20 Reused shared auth defaults for setup completion redirect

### Done
- Changed the post-install setup wizard redirect to use `FALLBACK_AUTH_ROUTE_DEFAULTS.loginPath` instead of a page-local `/login` literal.
- Added a source regression test that prevents reintroducing a hardcoded setup completion login URL.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/setup/__tests__/SetupWizardView.spec.ts src/router/__tests__/guards.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "window\\.location\\.href\\s*=\\s*['\\\"]/(login|register|dashboard|purchase)|FALLBACK_AUTH_ROUTE_DEFAULTS\\.loginPath" frontend/src/views/setup frontend/src/views/setup/__tests__ -S`

### Notes
- The setup wizard runs before normal public runtime settings are guaranteed to be available, so this slice reuses the shared fallback route map rather than adding a setup-time public-settings dependency.

## 2026-06-20 Reused auth route defaults for ops settings redirect

### Done
- Changed the admin ops dashboard disabled-state redirect to use `authRouteDefaults.adminSettingsPath` instead of a page-local `/admin/settings` literal.
- Added a source regression test that keeps the ops management entry wired to shared Sub2API route defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/ops/__tests__/OpsDashboard.spec.ts src/router/__tests__/guards.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "router\\.replace\\(['\\\"]/(admin/settings|admin/runtime-settings)|authRouteDefaults\\.value\\.adminSettingsPath|useAuthRouteDefaults" frontend/src/views/admin/ops frontend/src/views/admin/ops/__tests__ -S`

### Notes
- `adminSettingsPath` remains a built-in fallback route identity in the shared defaults map, while the active admin ops flow now follows the Sub2API-managed auth shell route defaults.

## 2026-06-20 Reused auth shell route defaults in the API client interceptor

### Done
- Added `resolveClientAuthRouteDefaults` so low-level API interception can read `auth_shell_config` from injected Sub2API public settings without depending on Vue stores.
- Changed token-expiry login redirects to use `auth_shell_config.defaults.loginPath` instead of client-local `/login` literals.
- Changed the ops-disabled interceptor redirect to use `auth_shell_config.defaults.adminSettingsPath` instead of a client-local `/admin/settings` literal.
- Added regression coverage for injected auth shell route defaults and source guards against reintroducing hardcoded browser redirects.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/api/__tests__/client.spec.ts src/router/__tests__/guards.spec.ts`
- `pnpm run frontend:typecheck`
- `rg -n "window\\.location\\.href\\s*=\\s*['\\\"]/(login|admin/settings)|resolveClientAuthRouteDefaults\\(\\)\\.(loginPath|adminSettingsPath)|VITE_API_BASE_URL" frontend/src/api/client.ts frontend/src/api/__tests__/client.spec.ts -S`

### Notes
- This removes another frontend bootstrap decision from the API layer. Router route identities still live in Vue, but browser redirects now follow Sub2API-managed public auth shell defaults.

## 2026-06-20 Fixed public integrations renderless component lint gate

### Done
- Replaced the empty `PublicIntegrations` template with a non-rendering guarded root element so the component remains side-effect-only while satisfying Vue template lint rules.
- Kept public integration injection controlled by Sub2API public `public_integrations_enabled` settings.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/common/__tests__/PublicIntegrations.spec.ts src/utils/__tests__/publicIntegrations.spec.ts`
- `pnpm run frontend:lint:check`
- `pnpm run frontend:typecheck`

### Notes
- This fixes the full frontend lint gate uncovered while continuing the frontend runtime-setting cleanup. It does not add visible DOM because the root node is guarded by `v-if="false"`.

## 2026-06-20 Reused auth route fallbacks in auth redirect utilities

### Done
- Changed `DEFAULT_AUTH_REDIRECT_PATH` and `DEFAULT_AUTH_BIND_REDIRECT_PATH` to reuse `FALLBACK_AUTH_ROUTE_DEFAULTS.userRedirectPath` and `profilePath`.
- Removed another duplicated frontend-owned `/dashboard` and `/profile` default from the auth redirect utility layer.
- Added source coverage so `authRedirect.ts` does not reintroduce local redirect path literals.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/authRedirect.spec.ts src/router/__tests__/guards.spec.ts src/views/auth/__tests__/OAuthCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts src/views/auth/__tests__/LoginView.turnstile.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `rg -n "DEFAULT_AUTH_REDIRECT_PATH\\s*=\\s*['\\\"]|DEFAULT_AUTH_BIND_REDIRECT_PATH\\s*=\\s*['\\\"]|FALLBACK_AUTH_ROUTE_DEFAULTS\\.(userRedirectPath|profilePath)" frontend/src/utils/authRedirect.ts frontend/src/utils/__tests__/authRedirect.spec.ts -S`

### Notes
- OAuth/login callbacks still resolve configured `auth_shell_config.defaults.defaultRedirectPath` and provider bind redirects at runtime. This slice only removes duplicated fallback route ownership from the shared sanitizer defaults.

## 2026-06-20 Reused auth route fallbacks in dashboard quick-action parser

### Done
- Changed `dashboardShellLabels.ts` built-in quick-action fallbacks to use `FALLBACK_AUTH_ROUTE_DEFAULTS.apiKeysPath`, `usagePath`, and `redeemPath`.
- Kept `dashboard_shell_config.defaults.quickActions` as the highest-priority runtime override and the caller-supplied auth defaults as the next fallback layer.
- Added source coverage so the dashboard shell parser does not reintroduce local `/keys`, `/usage`, or `/redeem` fallback literals.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts src/router/__tests__/guards.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `rg -n "createApiKeyPath: '/keys'|usagePath: '/usage'|redeemPath: '/redeem'|FALLBACK_AUTH_ROUTE_DEFAULTS\\.(apiKeysPath|usagePath|redeemPath)" frontend/src/components/user/dashboard/dashboardShellLabels.ts frontend/src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts -S`

### Notes
- The Dashboard page already passes configured auth route defaults into the dashboard shell parser. This slice removes the parser's final duplicated path ownership so fallback route identities stay centralized.

## 2026-06-20 Moved API test default prompt into public shell settings

### Done
- Extended `api_test_shell_config.defaults` with `defaultPrompt` so the API Test page can receive its initial prompt from Sub2API public settings.
- Changed `ApiTestView` to initialize and refresh the prompt from parsed shell defaults instead of importing the fixed gateway prompt directly.
- Kept `DEFAULT_GATEWAY_TEST_PROMPT` as the low-level request-body fallback for empty prompts and missing config.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover `defaults.defaultPrompt`.
- Added helper and page regression coverage for configured prompt defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiTestShell.spec.ts src/utils/__tests__/gatewayDocs.spec.ts src/views/user/__tests__/ApiTestView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- This removes another API Test page bootstrap value from the Touch/Sub2API frontend shell while preserving the existing request fallback behavior for blank user input.

## 2026-06-20 Moved prompt catalog X import automation mode into public shell settings

### Done
- Added `prompt_catalog_shell_config.defaults.importXAuto` to the prompt catalog shell parser.
- Changed `PromptCatalogView` to send `x_auto` from parsed shell defaults, with the existing behavior preserved by defaulting to `true`.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover prompt catalog query defaults and X automation mode.
- Added parser and page regression coverage so the Vue page no longer hardcodes `x_auto: true`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/promptCatalogShell.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- This keeps the current import path compatible while moving another prompt-gallery behavior switch under Sub2API-managed public settings.

## 2026-06-20 Moved API test max-token request setting into public shell settings

### Done
- Added `api_test_shell_config.defaults.maxTokens` to the API Test shell parser, defaulting to the existing 256-token behavior.
- Changed API Test request preview, curl generation, and actual send path to pass the configured max-token limit into gateway request body construction.
- Updated gateway request body generation to map the configured limit to protocol-specific fields: `max_tokens`, `max_output_tokens`, or Gemini `generationConfig.maxOutputTokens`.
- Kept API Guide examples compatible by only applying non-Anthropic max-token fields when an explicit option is provided.
- Updated Runtime Settings zh/en placeholders and hints so administrators can configure `defaults.maxTokens`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiTestShell.spec.ts src/utils/__tests__/gatewayDocs.spec.ts src/views/user/__tests__/ApiTestView.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- This moves another model-call behavior knob from the frontend implementation into Sub2API-managed public runtime settings while preserving existing defaults.

## 2026-06-20 Moved image workspace prompt length limit into public shell settings

### Done
- Added `workspace_shell_config.defaults.maxPromptLength` to the image workspace shell parser, defaulting to the existing 2000-character limit.
- Changed `ImageGeneratorView` to use the configured prompt length for the counter, too-long validation, and imported-draft truncation.
- Updated Runtime Settings zh/en placeholders and hints so administrators can configure `defaults.maxPromptLength`.
- Added parser and page regression coverage so the image workspace no longer owns a local `MAX_PROMPT_LENGTH` constant.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/imageWorkspaceShell.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- This keeps the current image workspace behavior by default while moving another frontend-shell constraint into Sub2API-managed public settings.

## 2026-06-20 Reused auth shell payment result path for checkout return URLs

### Done
- Added a sanitized `returnPath` option to `buildCreateOrderPayload`, preserving `/payment/result` as the fallback.
- Changed `PaymentView` checkout creation to pass `authRouteDefaults.value.paymentResultPath` into the order payload builder.
- Added regression coverage for configured payment return paths and unsafe path fallback.
- Added a page source guard so PaymentView keeps using the Sub2API-managed auth shell payment result path for checkout `return_url`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/paymentFlow.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- This removes another payment-flow path decision from the frontend checkout implementation. Admin provider callback display still has provider-specific static path constants, but the user checkout payload now follows public auth shell route defaults.

## 2026-06-20 Reused auth shell payment result path for Airwallex success URLs

### Done
- Changed `AirwallexPaymentView` to build Airwallex `successUrl` from `authRouteDefaults.value.paymentResultPath` instead of hardcoding `/payment/result`.
- Added regression coverage proving a Sub2API-managed auth shell payment result path is propagated into the Airwallex checkout success URL.
- Added a source guard so the Airwallex page no longer reintroduces `new URL('/payment/result', ...)`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/AirwallexPaymentView.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- This continues the payment bootstrap cleanup: the remaining `/payment/result` references are now mostly router fallback/default declarations, tests, and admin provider callback display constants rather than Airwallex runtime checkout behavior.

## 2026-06-20 Reused auth shell payment result path for provider callback return URLs

### Done
- Added `buildProviderCallbackPaths(paymentResultPath)` so payment provider callback return paths can be derived from Sub2API-managed auth shell route defaults.
- Changed `PaymentProviderDialog` to display, save, and extract provider `returnUrl` values with `authRouteDefaults.value.paymentResultPath` instead of the static callback-path map.
- Kept `PROVIDER_CALLBACK_PATHS` as the default compatibility map for existing imports.
- Added regression coverage for configured provider return paths and unsafe path fallback.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/payment/__tests__/providerConfig.spec.ts src/components/payment/__tests__/PaymentProviderDialog.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts src/utils/__tests__/authShell.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- `/payment/result` remains the sanitized fallback/default route, but provider admin return URL generation now follows the same Sub2API public route default as user checkout flows.

## 2026-06-20 Moved channel status refresh interval into public shell settings

### Done
- Extended `channel_status_shell_config` with `defaults.refreshIntervalSeconds`, preserving the existing 60-second fallback.
- Changed `ChannelStatusView` to use the configured refresh interval for the hero display, auto-refresh initialization, settings changes, and post-reload countdown reset.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover the channel status refresh interval default.
- Added parser and page regression coverage so the channel status page no longer imports the frontend `DEFAULT_INTERVAL_SECONDS` constant.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/channelStatusShell.spec.ts src/views/user/__tests__/ChannelStatusView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- The channel status page still owns UI state and monitor presentation, but its refresh timing default is now a Sub2API-managed public runtime setting rather than a Vue-local constant.

## 2026-06-20 Moved API guide curl defaults into public shell settings

### Done
- Extended `api_guide_shell_config.defaults` with `defaultPrompt` and `maxTokens`, preserving the existing gateway prompt and 256-token fallback.
- Changed `ApiGuideView` curl generation to use `apiGuideDefaults.value.defaultPrompt` and `apiGuideDefaults.value.maxTokens` instead of treating the `defaultPrompt` label as request behavior.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover the API guide curl defaults.
- Added helper and page regression coverage proving configured prompt/maxTokens flow into the generated curl body.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiGuideShell.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts src/utils/__tests__/gatewayDocs.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- API Guide still renders gateway variant cards in Vue, but another behavior-bearing default for examples is now owned by Sub2API public runtime settings.

## 2026-06-20 Moved dashboard range and recent-usage defaults into public shell settings

### Done
- Extended `dashboard_shell_config.defaults` with `dateRangeDays`, `defaultGranularity`, and `recentUsageLimit`, preserving the existing 7-day/day/5-row behavior.
- Changed `DashboardView` to initialize and synchronize its date window, chart granularity, and recent usage slice from parsed dashboard shell defaults.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover the dashboard behavior defaults.
- Added parser/source regression coverage so the dashboard no longer owns hardcoded `6 * 86400000`, `granularity = ref('day')`, or `res.items.slice(0, 5)` defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts src/views/user/__tests__/UsageView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- The dashboard UI composition still lives in Vue, but the primary dashboard behavior defaults now come from Sub2API-managed public runtime settings.

## 2026-06-20 Moved usage CSV export page size into public shell settings

### Done
- Extended `usage_shell_config.defaults` with `exportPageSize`, preserving the existing 100-row export batch fallback.
- Changed `UsageView` CSV export to use `usageShell.value.defaults.exportPageSize` instead of a local `const pageSize = 100`.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover the usage export batch-size default.
- Added parser and page regression coverage proving configured export page size is sent to the usage query API.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/usageShell.spec.ts src/views/user/__tests__/UsageView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- Usage history still owns table/export UI composition in Vue, but another behavior-bearing batch default now lives in Sub2API-managed public runtime settings.

## 2026-06-20 Moved usage date/key fetch defaults into public shell settings

### Done
- Extended `usage_shell_config.defaults` with `dateRangeDays` and `apiKeyPageSize`, preserving the existing 7-day date window and 100-key fetch fallback.
- Changed `UsageView` initial date range, reset date range, and API key filter fetch size to use parsed usage shell defaults instead of local constants.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover the usage date range and key-fetch defaults.
- Added parser and page regression coverage proving configured date/key defaults drive initial usage queries.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/usageShell.spec.ts src/views/user/__tests__/UsageView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- Usage history still owns table/filter UI composition in Vue, but its primary date/key bootstrap defaults now live in Sub2API-managed public runtime settings.

## 2026-06-20 Moved API guide key fetch page size into public shell settings

### Done
- Extended `api_guide_shell_config.defaults` with `apiKeyPageSize`, preserving the existing 100-key fetch fallback.
- Changed `ApiGuideView` API key selector loading to use `apiGuideDefaults.value.apiKeyPageSize` instead of `keysAPI.list(1, 100)`.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover the API Guide key-fetch default.
- Added helper and page regression coverage so API Guide no longer owns the key-fetch page size.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiGuideShell.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts src/utils/__tests__/gatewayDocs.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- API Guide still owns endpoint/card UI composition in Vue, but another API bootstrap default now lives in Sub2API-managed public runtime settings.

## 2026-06-20 Moved API test key and usage-sync page sizes into public shell settings

### Done
- Extended `api_test_shell_config.defaults` with `apiKeyPageSize` and `usageSyncPageSize`, preserving the existing 100-key fetch and 10-row usage-sync fallbacks.
- Changed `ApiTestView` API key selector loading and post-request usage record synchronization to use parsed API test shell defaults.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover the API Test key-fetch and usage-sync defaults.
- Added helper and page regression coverage so API Test no longer owns those request page sizes.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/apiTestShell.spec.ts src/views/user/__tests__/ApiTestView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts src/utils/__tests__/gatewayDocs.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- API Test still owns request execution UI and usage notice composition in Vue, but its key/usage bootstrap query sizes now live in Sub2API-managed public runtime settings.

## 2026-06-20 Moved payment result polling defaults into public shell settings

### Done
- Added `resolvePaymentResultDefaults` for `payment_shell_config.defaults.paymentResultRefreshIntervalMs` and `paymentResultMaxRefreshAttempts`, preserving the existing 2000ms interval and 15-attempt fallback.
- Changed `PaymentResultView` pending-order refresh scheduling to use parsed payment shell defaults instead of local constants.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover payment-result polling defaults.
- Added helper and page regression coverage for configured interval and max-attempt behavior.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- Payment result still owns resume-token/order-status UI composition in Vue, but its polling behavior defaults now live in Sub2API-managed public runtime settings.

## 2026-06-20 Moved key usage date defaults into public shell settings

### Done
- Added `resolveKeyUsageShellConfig` with `key_usage_shell_config.defaults.defaultDateRange` and `dailyUsageDays`, preserving the existing `today` query range and 30-day daily-detail fallback.
- Changed `KeyUsageView` to parse key usage shell settings once and initialize its date range and daily-detail request days from the parsed defaults.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover key usage date defaults.
- Added helper and page regression coverage for configured default range and daily-detail days.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/keyUsageShell.spec.ts src/views/__tests__/KeyUsageView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- Public API key usage still owns the query UI and visualization composition in Vue, but its initial date-query defaults now live in Sub2API-managed public runtime settings.

## 2026-06-20 Moved shared payment status polling interval into public shell settings

### Done
- Added `resolvePaymentStatusPollingDefaults` for `payment_shell_config.defaults.paymentStatusPollIntervalMs`, preserving the existing 3000ms payment status polling fallback.
- Changed `PaymentQRCodeView` to read the polling interval directly from public payment shell settings.
- Changed `PaymentStatusPanel` and `PaymentQRDialog` to accept `pollIntervalMs`, with `PaymentView` passing the Sub2API-managed setting into `PaymentStatusPanel`.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover the shared payment status polling default.
- Added helper, page, component, and source regression coverage for configured payment status polling intervals.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts src/components/payment/__tests__/PaymentQRDialog.spec.ts src/components/payment/__tests__/PaymentStatusPanel.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- Payment QR and inline payment status UI composition still lives in Vue, but the shared status polling interval now lives in Sub2API-managed public runtime settings.

## 2026-06-22 Moved Stripe runtime delays into public shell settings

### Done
- Added `resolveStripePaymentRuntimeDefaults` for `payment_shell_config.defaults.stripePollIntervalMs`, `stripeCloseDelayMs`, and `stripePopupInitTimeoutMs`, preserving the existing 3000ms poll, 2000ms close delay, and 15000ms popup init timeout fallbacks.
- Changed `StripePaymentView` to use parsed Stripe runtime defaults for WeChat Pay status polling and success close/redirect scheduling.
- Changed `StripePopupView` to use parsed Stripe runtime defaults for popup init timeout, WeChat Pay status polling, and success close scheduling.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover the Stripe runtime defaults.
- Added helper and page regression coverage for configured Stripe runtime timing.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts src/views/user/__tests__/StripePopupView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- Stripe carrier/popup UI composition still lives in Vue, but another set of payment timing defaults now lives in Sub2API-managed public runtime settings.

## 2026-06-23 Moved WeChat payment verify retry defaults into public shell settings

### Done
- Extended `payment_shell_config.defaults` parsing with `paymentVerifyRetryIntervalMs` and `paymentVerifyRetryMaxAttempts`, preserving the existing 15000ms interval and 6-attempt fallback.
- Changed `PaymentStatusPanel` and `PaymentQRDialog` to consume configurable verify retry defaults instead of owning local WeChat pending-order retry constants.
- Changed `PaymentView` to pass the parsed Sub2API public payment shell defaults into the inline payment status panel.
- Updated Runtime Settings zh/en placeholders and hints so administrators can discover the WeChat payment verify retry defaults.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/paymentShell.spec.ts src/components/payment/__tests__/PaymentQRDialog.spec.ts src/components/payment/__tests__/PaymentStatusPanel.spec.ts src/views/user/__tests__/PaymentView.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/utils/paymentShell.ts frontend/src/components/payment/PaymentQRDialog.vue frontend/src/components/payment/PaymentStatusPanel.vue frontend/src/views/user/PaymentView.vue frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts progress.md`

### Notes
- Per request, this slice did not add new test code. Payment waiting UI composition still lives in Vue, but another behavior-bearing payment recovery default now lives in Sub2API-managed public runtime settings.

## 2026-06-23 Moved admin sidebar route paths into auth shell runtime defaults

### Done
- Added `resolveAdminSidebarRouteDefaults` so the admin sidebar can read manager entry paths from localized `auth_shell_config.defaults` without changing the existing shared auth-shell parser contract.
- Changed `AppSidebar` to use Sub2API-configurable paths for admin Ops, users, groups, channel management, subscriptions, accounts, announcements, proxies, risk control, redeem, promo code, affiliate, order, and usage entries.
- Replaced the remaining active hardcoded admin tour-selector path checks in `AppSidebar` with the same configured admin sidebar paths.
- Updated Runtime Settings zh/en auth-shell placeholders and hints so operators can discover the admin sidebar path keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/layout/__tests__/AppSidebar.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/utils/adminSidebarShell.ts frontend/src/components/layout/AppSidebar.vue frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts progress.md`

### Notes
- This slice moves admin sidebar route decisions under Sub2API public settings, but the sidebar grouping, labels, ordering, and visibility logic still remain a Vue-owned navigation shell rather than a full backend-driven menu schema.

## 2026-06-23 Removed legacy OAuth callback success and invitation fragment compatibility

### Done
- Removed the retired legacy success-token fragment handling from `LinuxDoCallbackView`, `OidcCallbackView`, `WechatCallbackView`, and `DingTalkCallbackView`.
- Removed the retired legacy `pending_oauth_token` fragment invitation-registration branch from those callback views, leaving the active cookie-based pending-session flow as the only supported registration continuation path.
- Simplified DingTalk invitation completion to stop sending an optional legacy `pending_oauth_token` field from the frontend.
- Updated the auth callback view tests so they assert the current pending-session exchange flow instead of legacy fragment compatibility behavior.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/__tests__/LinuxDoCallbackView.spec.ts frontend/src/views/auth/__tests__/OidcCallbackView.spec.ts frontend/src/views/auth/__tests__/WechatCallbackView.spec.ts frontend/src/views/auth/__tests__/DingTalkCallbackView.spec.ts progress.md`

### Notes
- This slice intentionally drops retired fragment-based compatibility in favor of the current Sub2API-managed OAuth pending-session flow. Error fragments remain supported because the active backend still redirects provider failures that way.

## 2026-06-23 Moved admin sidebar grouping and ordering into auth shell schema

### Done
- Added `resolveAdminSidebarSections` so localized `auth_shell_config.defaults.adminSidebarSections` can define admin sidebar groups and item ordering without changing the shared auth-shell parser contract.
- Changed `AppSidebar` to render admin navigation from section data instead of one hardcoded flat admin item list, while keeping the current fallback structure when no schema is configured.
- Kept simple-mode admin behavior conservative with a compact built-in section, while allowing normal admin mode to reorder/group built-in entries through runtime settings.
- Updated Runtime Settings zh/en auth-shell placeholders and hints so operators can discover the `adminSidebarSections` schema and supported item keys.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/layout/__tests__/AppSidebar.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/utils/adminSidebarSchema.ts frontend/src/components/layout/AppSidebar.vue frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts progress.md`

### Notes
- This is a fast ROI schema step, not a full backend-driven menu system. Menu labels, icon selection, feature gating, and personal-account/admin split still remain in the Vue shell.

## 2026-06-23 Moved user and admin-personal sidebar grouping into auth shell schema

### Done
- Extended the sidebar schema reader with `defaults.userSidebarSections` and `defaults.adminPersonalSidebarSections` so the regular user sidebar and the admin “My Account” area can be reordered/grouped from runtime settings.
- Changed `AppSidebar` to render regular-user navigation and admin personal navigation from section data instead of one fixed flat list each, while keeping built-in fallback sections when no schema is configured.
- Preserved existing feature-flag filtering, simple-mode filtering, onboarding anchors, and custom user menu item injection while moving grouping/order ownership closer to Sub2API settings.
- Updated Runtime Settings zh/en auth-shell placeholders and hints so operators can discover the new user/admin-personal sidebar schema keys and supported item names.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/layout/__tests__/AppSidebar.spec.ts src/i18n/__tests__/localeCoverage.spec.ts src/i18n/__tests__/adminNamespaceLocaleAudit.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/utils/adminSidebarSchema.ts frontend/src/components/layout/AppSidebar.vue frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts progress.md`

### Notes
- This continues the quick-win navigation refactor. Sidebar item labels, icon mapping, feature gating policy, and route definitions still remain in the Vue layer rather than a complete backend-driven navigation model.

## 2026-06-23 Removed legacy blank login-agreement template compatibility

### Done
- Removed `isLegacyBlankLoginAgreementDocuments` and the old blank-document compatibility path from `loginAgreementTemplates.ts`.
- Changed admin `SettingsView` to treat login-agreement documents as current structured data only: empty backend payload now loads the commercial template bundle, while any non-empty payload is used directly without legacy blank-shape detection.
- Simplified “apply commercial template” confirmation logic so it only checks whether the current form already contains custom content.
- Removed the legacy blank-template unit coverage and kept the active commercial-template and privacy-merge behaviors covered.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/utils/__tests__/loginAgreementTemplates.spec.ts src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/utils/loginAgreementTemplates.ts frontend/src/views/admin/SettingsView.vue frontend/src/utils/__tests__/loginAgreementTemplates.spec.ts progress.md`

### Notes
- This slice assumes there is no historical blank-template business data to preserve, and intentionally keeps only the current structured agreement-document flow.

## 2026-06-23 Removed legacy WeChat single-app setting backfill in admin settings

### Done
- Removed the admin `SettingsView` load-time compatibility logic that copied `wechat_connect_app_id` / `wechat_connect_app_secret_configured` into the newer Open/MP/Mobile WeChat fields.
- Kept the current form behavior focused on the explicit `wechat_connect_open_*`, `wechat_connect_mp_*`, and `wechat_connect_mobile_*` settings returned by the backend.
- Updated the admin settings test fixture so WeChat Connect coverage uses the current MP-field payload instead of the retired single-app compatibility shape.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/admin/SettingsView.vue frontend/src/views/admin/__tests__/SettingsView.spec.ts progress.md`

### Notes
- This assumes there is no historical single-AppID WeChat settings data that still needs to be auto-upgraded in the frontend form layer.

## 2026-06-24 Removed legacy WeChat single-app save fallback in admin settings

### Done
- Removed the admin `SettingsView` save-path fallback that derived `wechat_connect_app_id` from the newer Open/MP/Mobile WeChat fields.
- Kept the old `wechat_connect_app_id` payload field only as an explicit form field value, so the frontend no longer silently backfills a retired single-AppID shape from current WeChat settings.
- Updated the WeChat Connect save assertion in `SettingsView.spec.ts` so it verifies the current MP-field-driven save behavior instead of the old implicit fallback.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/admin/SettingsView.vue frontend/src/views/admin/__tests__/SettingsView.spec.ts progress.md`

### Notes
- This assumes the active admin flow should no longer synthesize the retired single-AppID WeChat config shape during saves.

## 2026-06-25 Removed legacy WeChat single-app fields from frontend settings form state

### Done
- Removed `wechat_connect_app_id`, `wechat_connect_app_secret`, and `wechat_connect_app_secret_configured` from the active frontend `SettingsForm` state in `SettingsView.vue`.
- Removed the remaining frontend save-path emission of the retired single-AppID WeChat fields so the admin UI now operates only on the current Open/MP/Mobile field set.
- Updated the admin settings test fixture and WeChat save assertion so they only cover the current WeChat field contract used by the frontend.
- Kept backend compatibility untouched for now; this slice only stops the Vue admin form from carrying the retired single-app fields.

### Validation
- `pnpm run frontend:typecheck`
- `pnpm --filter sub2api-frontend exec vitest run src/views/admin/__tests__/SettingsView.spec.ts`
- `git diff --check -- frontend/src/views/admin/SettingsView.vue frontend/src/views/admin/__tests__/SettingsView.spec.ts progress.md`

### Notes
- This makes the frontend form layer cleanly current-state only, while backend DTO/service compatibility for older WeChat settings can be handled separately later if still needed.

## 2026-06-25 Stopped exposing legacy WeChat single-app fields in backend response DTOs

### Done
- Removed `wechat_connect_app_id` and `wechat_connect_app_secret_configured` from the backend response DTO surface used by settings/public-settings responses.
- Removed the corresponding admin/public response mapping assignments so the backend no longer emits the retired single-AppID WeChat fields to current clients.
- Updated the API contract fixture assertions to stop expecting those legacy response fields while preserving the current Open/MP/Mobile WeChat capability fields.

### Validation
- `(cd backend && go test ./internal/handler -run 'TestSettingHandler_GetPublicSettings_ExposesWeChatOAuthModeCapabilities' -count=1)`
- `(cd backend && go test ./internal/handler -run 'Test.*(LinuxDo|OIDC|WeChat|DingTalk|Pending).*' -count=1)`
- `git diff --check -- backend/internal/handler/dto/settings.go backend/internal/handler/admin/setting_handler.go backend/internal/server/api_contract_test.go progress.md`

### Notes
- Backend request/internal settings compatibility still remains for this WeChat single-app shape; this slice only removes it from the active response/API surface consumed by the unified frontend.

## 2026-06-25 Removed legacy WeChat single-app fields from admin settings request surface

### Done
- Removed `wechat_connect_app_id` / `wechat_connect_app_secret` from the active admin settings request contract in `backend/internal/handler/admin/setting_handler.go`.
- Updated WeChat settings normalization so the handler derives any internal legacy single-app value from the current Open/MP/Mobile fields instead of accepting the retired single-field request shape from clients.
- Kept backend internal compatibility by still populating service-layer legacy fields from the current three-field model, while stopping the admin API from treating the old shape as an external contract.
- Narrowed the “secret changed” detection for WeChat settings so it now follows the current Open/MP/Mobile secret fields.

### Validation
- `(cd backend && go test ./internal/handler/admin ./internal/handler -run 'Test.*Setting|Test.*WeChat.*|Test.*PublicSettings.*' -count=1)`
- `git diff --check -- backend/internal/handler/admin/setting_handler.go progress.md`

### Notes
- At this point the old WeChat single-app shape is no longer active in the frontend form, frontend types, backend response surface, or backend admin request contract. Remaining compatibility now lives only inside backend settings/service internals.

## 2026-06-25 Removed legacy single-app fallback from admin WeChat settings normalization

### Done
- Removed `previousSettings.WeChatConnectAppID` / `previousSettings.WeChatConnectAppSecret` as fallback sources during admin WeChat settings normalization in `setting_handler.go`.
- Kept the admin save flow anchored only to the current Open/MP/Mobile field set and their previous values, while still deriving internal legacy aggregate values from the current fields when needed downstream.
- Further isolated the retired single-app WeChat shape to backend settings/service internals instead of letting it influence active admin form update decisions.

### Validation
- `(cd backend && go test ./internal/handler/admin ./internal/handler -run 'Test.*Setting|Test.*WeChat.*|Test.*PublicSettings.*' -count=1)`
- `git diff --check -- backend/internal/handler/admin/setting_handler.go progress.md`

### Notes
- This assumes the active admin settings flow should not inherit stale legacy single-app values when the current three-mode WeChat fields are empty or partially updated.

## 2026-06-25 Cleared legacy WeChat single-app keys on active settings writes

### Done
- Changed the admin settings handler to stop deriving active service-layer values for `WeChatConnectAppID` / `WeChatConnectAppSecret` from the current Open/MP/Mobile fields during request assembly.
- Changed the settings service write path to explicitly write the legacy single-app WeChat secret key as well, allowing active saves to clear retired `wechat_connect_app_id` / `wechat_connect_app_secret` values instead of leaving stale compatibility data behind.
- This means current admin saves now keep the active three-field model authoritative while progressively scrubbing the retired single-app write path.

### Validation
- `(cd backend && go test ./internal/handler/admin ./internal/handler ./internal/service -run 'Test.*(Setting|WeChat|PublicSettings).*' -count=1)`
- `git diff --check -- backend/internal/handler/admin/setting_handler.go backend/internal/service/setting_service.go progress.md`

### Notes
- Backend internal read compatibility still remains, but new writes no longer reinforce the retired single-app WeChat settings shape and now actively clear it.

## 2026-06-25 Removed legacy single-app DB-key reads from active WeChat runtime tests and mode resolution path

### Done
- Removed `LegacyAppID` / `LegacyAppSecret` fallback usage from `WeChatConnectOAuthConfig` mode-resolution helpers so runtime mode selection now reads only the current Open/MP/Mobile fields.
- Updated `effectiveWeChatConnectOAuthConfig` to stop reading `SettingKeyWeChatConnectAppID` / `SettingKeyWeChatConnectAppSecret` as active database override sources.
- Migrated the remaining focused service/handler tests on this path to seed current Open/MP-specific WeChat keys instead of the retired single-app DB keys.

### Validation
- `(cd backend && go test ./internal/service ./internal/handler -run 'Test.*WeChat.*|Test.*PaymentOrder.*|Test.*Setting.*' -count=1)`
- `git diff --check -- backend/internal/service/settings_view.go backend/internal/service/setting_service.go backend/internal/service/setting_service_wechat_config_test.go backend/internal/service/setting_service_public_test.go backend/internal/service/payment_order_result_test.go backend/internal/handler/auth_wechat_oauth_test.go progress.md`

### Notes
- This keeps config/env compatibility outside the DB settings path for now, but the active settings-backed WeChat runtime path no longer depends on the retired single-app database keys.

## 2026-06-25 Removed legacy WeChat env/config fallback from active runtime config loading

### Done
- Removed `applyLegacyWeChatConnectEnvCompatibility` and the old `WECHAT_OAUTH_*` environment-variable fallback path from `backend/internal/config/config.go`.
- Removed the old `AppID` / `AppSecret` fan-out normalization from `normalizeWeChatConnectConfig`, so runtime config no longer derives the current Open/MP/Mobile fields from the retired single-app config shape.
- Updated the config test to assert that legacy `WECHAT_OAUTH_*` environment variables no longer activate WeChat Connect runtime settings.
- Kept the current Open/MP/Mobile config path and its related service/handler tests green.

### Validation
- `(cd backend && go test ./internal/config ./internal/service ./internal/handler -run 'Test.*WeChat.*|Test.*PaymentOrder.*|Test.*Setting.*' -count=1)`
- `git diff --check -- backend/internal/config/config.go backend/internal/config/config_test.go backend/internal/service/settings_view.go backend/internal/service/setting_service.go backend/internal/service/setting_service_wechat_config_test.go backend/internal/service/setting_service_public_test.go backend/internal/service/payment_order_result_test.go backend/internal/handler/auth_wechat_oauth_test.go progress.md`

### Notes
- With this slice, the active WeChat configuration path now consistently expects the current Open/MP/Mobile model across frontend, handler, settings DB, service runtime, and config/env loading. Remaining old single-app fields are now only deeper internal compatibility residue.

## 2026-06-25 Removed legacy single-app fields from WeChat runtime config model

### Done
- Removed `AppID` and `AppSecret` from `backend/internal/config.WeChatConnectConfig`, leaving only the current Open/MP/Mobile field model in the runtime config struct.
- Verified there were no remaining active code paths reading those two legacy config fields after the earlier runtime normalization cleanup.

### Validation
- `(cd backend && go test ./internal/config ./internal/service ./internal/handler -run 'Test.*WeChat.*|Test.*PaymentOrder.*|Test.*Setting.*' -count=1)`
- `git diff --check -- backend/internal/config/config.go progress.md`

### Notes
- This is a model cleanup step: active WeChat runtime config now has no legacy single-app fields at the struct level, matching the already-cleaned frontend, handler, settings, and runtime paths.

## 2026-06-25 Removed legacy WeChat provider-key and openid repair compatibility from active auth flow

### Done
- Removed `wechatOAuthLegacyProviderKey` and the legacy `providerKey=wechat` compatibility handling from the active WeChat OAuth flow.
- Removed `findWeChatUserByLegacyOpenID` and the old openid/unionid identity repair fallback from `auth_wechat_oauth.go`.
- Simplified pending OAuth identity/channel reconciliation in `auth_oauth_pending_flow.go` so WeChat identity resolution now uses the current provider key directly instead of searching compatible legacy key sets and openid repair candidates.
- Kept the active `wechat-main` identity/channel path intact and verified the focused WeChat/pending auth handler tests still pass.

### Validation
- `(cd backend && go test ./internal/handler -run 'Test.*WeChat.*|Test.*Pending.*' -count=1)`
- `git diff --check -- backend/internal/handler/auth_wechat_oauth.go backend/internal/handler/auth_oauth_pending_flow.go progress.md`

### Notes
- This assumes there is no historical WeChat auth data stored under the retired legacy provider key that still needs runtime repair or lookup in the active unified system.

## 2026-06-25 Removed legacy openid-based WeChat payment resume fallback from frontend flow

### Done
- Simplified `WechatPaymentCallbackView` so it now requires the opaque `wechat_resume_token` and no longer forwards legacy `openid/state/scope/payment_type/amount/order_type/plan_id` resume payloads back into the purchase route.
- Simplified `paymentWechatResume.ts` to parse only the opaque WeChat resume token path and return `null` when that token is absent.
- Removed the corresponding openid-based resume branch from `PaymentView` so the active in-app WeChat resume flow now uses only the opaque resume token path.
- Updated focused frontend tests to assert the new token-only behavior and the missing-token error path.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/WechatPaymentCallbackView.spec.ts src/views/user/__tests__/paymentWechatResume.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/WechatPaymentCallbackView.vue frontend/src/views/user/paymentWechatResume.ts frontend/src/views/user/PaymentView.vue frontend/src/views/auth/__tests__/WechatPaymentCallbackView.spec.ts frontend/src/views/user/__tests__/paymentWechatResume.spec.ts progress.md`

### Notes
- This assumes there is no active production flow still depending on the old openid-based WeChat payment callback resume parameters; the unified flow now treats the opaque resume token as the only supported return contract.

## 2026-06-25 Removed legacy encryption-key fallback from active WeChat payment resume verification

### Done
- Changed the active WeChat payment resume service factory so it now uses only `PAYMENT_RESUME_SIGNING_KEY` and no longer accepts the old encryption key as a verification fallback.
- Updated `auth_wechat_oauth.go` to construct the WeChat payment resume service from the explicit payment-resume signing key only.
- Removed the focused tests that asserted legacy encryption-key verification fallback and replaced the remaining payment-order expectation with the current explicit-signing-key-required behavior.

### Validation
- `(cd backend && go test ./internal/service ./internal/handler -run 'Test.*WeChat.*|Test.*PaymentResume.*|Test.*PaymentOrder.*' -count=1)`
- `git diff --check -- backend/internal/service/payment_service.go backend/internal/service/payment_resume_service_test.go backend/internal/service/payment_order_result_test.go backend/internal/handler/auth_wechat_oauth.go progress.md`

### Notes
- This assumes there are no active payment resume tokens in circulation that still rely on the old encryption-key verification fallback. The live WeChat payment resume path now requires the explicit payment resume signing key end to end.

## 2026-06-25 Removed legacy single-app fields from backend service active model

### Done
- Removed `WeChatConnectAppID`, `WeChatConnectAppSecret`, and `WeChatConnectAppSecretConfigured` from the active `service.SystemSettings` WeChat model in `settings_view.go`.
- Removed the remaining active `parseSettings` population of those fields and changed the settings write path to clear the retired single-app DB keys explicitly instead of reading values from the service model.
- Removed the corresponding struct-literal assignments in the admin settings handler, so active handler/service flow now uses only the current Open/MP/Mobile fields end to end.

### Validation
- `(cd backend && go test ./internal/service ./internal/handler/admin ./internal/handler -run 'Test.*WeChat.*|Test.*PaymentOrder.*|Test.*Setting.*' -count=1)`
- `git diff --check -- backend/internal/service/settings_view.go backend/internal/service/setting_service.go backend/internal/handler/admin/setting_handler.go progress.md`

### Notes
- This completes the removal of the old WeChat single-app shape from the active frontend, handler, response, request, runtime, and service-model paths. Any remaining references are deeper compatibility residue rather than part of the live unified flow.

## 2026-06-25 Removed legacy single-app WeChat DB keys from active settings fetch path

### Done
- Removed `SettingKeyWeChatConnectAppID` and `SettingKeyWeChatConnectAppSecret` from `GetWeChatConnectOAuthConfig()` fetch keys so the active settings-backed WeChat config loader no longer even requests the retired single-app DB settings.
- Verified the current WeChat Open/MP/Mobile runtime path still passes focused service/handler tests without those DB keys being loaded.

### Validation
- `(cd backend && go test ./internal/service ./internal/handler -run 'Test.*WeChat.*|Test.*PaymentOrder.*|Test.*Setting.*' -count=1)`
- `git diff --check -- backend/internal/service/setting_service.go progress.md`

### Notes
- At this point the old single-app WeChat DB keys are no longer part of the active frontend flow, active handler flow, active write path, active settings fetch path, or active runtime mode resolution path.

## 2026-06-25 Removed final active references to retired single-app WeChat setting keys

### Done
- Removed `SettingKeyWeChatConnectAppID` and `SettingKeyWeChatConnectAppSecret` from the active service constants and write path.
- Updated the remaining WeChat public-settings test fixture to use the current MP-specific keys instead of the retired single-app keys.
- Verified there are no remaining active backend code references to those two retired WeChat setting keys.

### Validation
- `(cd backend && go test ./internal/service ./internal/handler ./internal/config -run 'Test.*WeChat.*|Test.*PaymentOrder.*|Test.*Setting.*' -count=1)`
- `rg -n "SettingKeyWeChatConnectAppID|SettingKeyWeChatConnectAppSecret" backend -S`
- `git diff --check -- backend/internal/service/domain_constants.go backend/internal/service/setting_service.go backend/internal/handler/setting_handler_public_test.go progress.md`

### Notes
- This effectively closes the active-code cleanup for the old single-app WeChat settings shape. Any remaining references after this point would be outside the live backend/frontend flow and should be evaluated separately.

## 2026-06-25 Removed unused OIDC provider_fallback field from active flow

### Done
- Removed `provider_fallback` from the OIDC OAuth upstream claims payload because the active frontend flow did not consume it.
- Removed the corresponding unused field from the `PendingOidcCompletion` frontend type in `OidcCallbackView.vue`.
- Verified the OIDC callback flow still passes focused frontend and backend tests without this unused compatibility-style payload field.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/OidcCallbackView.spec.ts`
- `(cd backend && go test ./internal/handler -run 'Test.*OIDC.*' -count=1)`
- `pnpm run frontend:typecheck`
- `git diff --check -- backend/internal/handler/auth_oidc_oauth.go frontend/src/views/auth/OidcCallbackView.vue progress.md`

### Notes
- This is a narrow contract cleanup step: it removes a field that looked legacy/fallback-oriented but was not part of the active user-visible flow.

## 2026-06-25 Removed unused choice payload noise from active OIDC and LinuxDo auth flows

### Done
- Removed `choice_reason` and `existing_account_bindable` from the active OIDC/LinuxDo pending choice payloads because the current frontend flow does not consume them.
- Kept the actually used fields such as `step`, `redirect`, `email`, and `existing_account_email` intact, so the active bind/create-account routing behavior is unchanged.
- Updated focused backend handler tests to stop asserting those non-consumed payload fields and keep coverage on the fields that still drive runtime behavior.

### Validation
- `(cd backend && go test ./internal/handler -run 'Test.*(OIDC|LinuxDo).*' -count=1)`
- `git diff --check -- backend/internal/handler/auth_linuxdo_oauth.go backend/internal/handler/auth_oidc_oauth.go backend/internal/handler/auth_linuxdo_oauth_test.go backend/internal/handler/auth_oidc_oauth_test.go progress.md`

### Notes
- DingTalk still retains `existing_account_bindable` because its email-completion page currently uses that field to decide whether to route users back into the bind/choice flow.

## 2026-06-25 Removed unused choice payload noise from active Email OAuth flow

### Done
- Removed `choice_reason` and `create_account_allowed` from the Email OAuth pending completion payload because the current frontend does not consume them.
- Updated focused backend handler tests to stop asserting those fields while keeping the active registration-completion flow assertions on `step`, `error`, `invitation_required`, `email`, and `resolved_email`.

### Validation
- `(cd backend && go test ./internal/handler -run 'Test.*EmailOAuth.*' -count=1)`
- `git diff --check -- backend/internal/handler/auth_email_oauth.go backend/internal/handler/auth_email_oauth_test.go progress.md`

### Notes
- This is another small contract cleanup step that removes non-consumed payload noise without changing the active Email OAuth registration-completion behavior.

## 2026-06-25 Reused shared email-oauth provider parsing in OAuthCallbackView

### Done
- Added `emailOAuthFlow.ts` with shared `resolveEmailOAuthProvider()` parsing for the Email OAuth provider marker.
- Updated `OAuthCallbackView.vue` to use the shared helper for session-stored provider parsing and pending completion provider resolution instead of duplicating string normalization logic inline.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/OAuthCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/emailOAuthFlow.ts frontend/src/views/auth/OAuthCallbackView.vue progress.md`

### Notes
- This is a small shell-thinning step, but it starts isolating Email OAuth callback-specific logic into its own helper surface instead of keeping everything inline in the page component.

## 2026-06-25 Extracted shared Email OAuth pending-completion state derivation

### Done
- Extended `emailOAuthFlow.ts` with shared `EmailOAuthPendingCompletion` typing plus helpers for token-response detection and registration-state derivation.
- Updated `OAuthCallbackView.vue` to use the shared helpers for pending completion provider/redirect/registration state parsing instead of duplicating that interpretation inline.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/OAuthCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/emailOAuthFlow.ts frontend/src/views/auth/OAuthCallbackView.vue progress.md`

### Notes
- This is another page-thinning step: Email OAuth callback behavior stays unchanged while more of its pending completion interpretation moves out of the page component and into a shared helper surface.

## 2026-06-25 Removed unused choice_reason payload noise from active DingTalk auth flow

### Done
- Removed `choice_reason` from the active DingTalk pending choice payload because the current frontend flow does not consume it.
- Kept `existing_account_bindable` intact for DingTalk because the email-completion page still relies on that field to route users back into bind/choice handling.
- Verified focused DingTalk handler tests still pass with the slimmer payload.

### Validation
- `(cd backend && go test ./internal/handler -run 'Test.*DingTalk.*' -count=1)`
- `git diff --check -- backend/internal/handler/auth_dingtalk_oauth.go progress.md`

### Notes
- This keeps the minimal DingTalk payload still needed by the active email-completion flow while removing another non-consumed reason/debug field from the live contract.

## 2026-06-25 Removed never-populated suggested_email fields from active auth callback views

### Done
- Removed `suggested_email` from the active frontend callback completion types in `LinuxDoCallbackView.vue`, `OidcCallbackView.vue`, and `DingTalkCallbackView.vue`.
- Simplified `extractPendingAccountEmail()` in those views to stop checking a field that the active backend flow does not populate.
- Verified the focused auth callback view tests and frontend typecheck still pass without the dead field.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue progress.md`

### Notes
- This is another payload-noise cleanup step: it removes frontend-only type and fallback branches for a field that is not emitted by the active backend auth flows.

## 2026-06-25 Extracted shared pending-account action parsing from auth callback views

### Done
- Added `frontend/src/views/auth/pendingAccountFlow.ts` with shared helpers for pending-state normalization, pending-email extraction, and common pending-action resolution.
- Replaced duplicated pending-account parsing logic in `LinuxDoCallbackView.vue`, `OidcCallbackView.vue`, and `DingTalkCallbackView.vue` with the shared helper module.
- Kept behavior unchanged while reducing another cluster of repeated Touch-derived frontend orchestration code from the active auth callback surfaces.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/pendingAccountFlow.ts frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue progress.md`

### Notes
- This is a structural shell-thinning step rather than a contract change. WeChat remains slightly different for now because its callback flow still carries a special `choice` branch plus resume-email fallback behavior.

## 2026-06-25 Reused shared pending-account parsing in WeChat callback flow

### Done
- Updated `WechatCallbackView.vue` to reuse the shared `pendingAccountFlow.ts` helpers for pending email extraction and pending action resolution.
- Kept the WeChat-specific differences limited to its `choice` branch naming and `resolveResumeEmail()` fallback while removing another copy of the core parsing logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/WechatCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/pendingAccountFlow.ts frontend/src/views/auth/WechatCallbackView.vue progress.md`

### Notes
- With this slice, all four main auth callback views now share the same pending-account parsing foundation, reducing another chunk of Touch-derived repeated frontend orchestration.

## 2026-06-25 Extracted shared standard pending-account state application for auth callback views

### Done
- Extended `pendingAccountFlow.ts` with a shared `buildStandardPendingAccountState()` helper for the common choose/create/bind account-action state transitions.
- Replaced the repeated `applyPendingAccountAction()` state assignment blocks in `LinuxDoCallbackView.vue`, `OidcCallbackView.vue`, and `DingTalkCallbackView.vue` with the shared helper output.
- Kept WeChat separate for now because its callback flow still carries a different `choice` branch plus resume-email fallback behavior.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/pendingAccountFlow.ts frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue progress.md`

### Notes
- This is another shell-thinning step. The auth callback pages still own some page-local state, but a larger chunk of duplicated transition logic is now centralized.

## 2026-06-25 Reused shared standard state application in WeChat callback flow

### Done
- Extended `pendingAccountFlow.ts` with `buildStandardPendingAccountStateForChoiceMode()` to cover WeChat’s special `choice` branch while still sharing the common create/bind/default state wiring.
- Updated `WechatCallbackView.vue` to reuse the shared helper for standard pending-account state assignment, leaving only the `needsChooser` toggle as the WeChat-specific extra step.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/WechatCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/pendingAccountFlow.ts frontend/src/views/auth/WechatCallbackView.vue progress.md`

### Notes
- With this slice, all four callback pages now share not only pending-state parsing, but also almost all standard pending-account state assignment logic. Remaining divergence is limited to true page-specific behavior.

## 2026-06-25 Extracted shared standard 2FA challenge state for auth callback views

### Done
- Extended `pendingAccountFlow.ts` with `buildStandardTotpChallengeState()` to centralize the common 2FA challenge state transition used by callback views.
- Replaced the repeated `applyTotpChallenge()` state-assignment blocks in `LinuxDoCallbackView.vue`, `OidcCallbackView.vue`, and `DingTalkCallbackView.vue` with the shared helper output.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/pendingAccountFlow.ts frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue progress.md`

### Notes
- This is another shell-thinning step that leaves page-specific 2FA UI behavior intact while removing duplicated callback-state setup code.

## 2026-06-25 Extracted shared bind/create mode switch state for auth callback views

### Done
- Extended `pendingAccountFlow.ts` with shared helpers for the repeated `switchToBindLoginMode()` and `switchToCreateAccountMode()` state transitions.
- Replaced the duplicated bind/create switch state setup in `LinuxDoCallbackView.vue`, `OidcCallbackView.vue`, and `DingTalkCallbackView.vue` with the shared helper outputs.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/pendingAccountFlow.ts frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue progress.md`

### Notes
- This continues shrinking repeated callback orchestration while leaving each page’s remaining unique UI state and routing behavior intact.

## 2026-06-25 Extracted shared request error parsing for auth pages

### Done
- Added `frontend/src/views/auth/requestError.ts` with the shared request-error extraction logic previously duplicated across auth callback/completion pages.
- Replaced the duplicated `getRequestErrorMessage()` implementations in LinuxDo/OIDC/DingTalk/WeChat callback views and DingTalk email completion with the shared helper.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/requestError.ts frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/DingTalkEmailCompletionView.vue progress.md`

### Notes
- This is another small shell-thinning step that centralizes repeated frontend error-message parsing without changing any active auth behavior.

## 2026-06-25 Extracted shared create-account recovery error detection for auth callback views

### Done
- Added `isCreateAccountRecoveryError()` to `pendingAccountFlow.ts` and removed the duplicated local implementations from `LinuxDoCallbackView.vue`, `OidcCallbackView.vue`, and `DingTalkCallbackView.vue`.
- Kept the existing behavior intact while removing another repeated chunk of callback recovery-state parsing logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/pendingAccountFlow.ts frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue progress.md`

### Notes
- This continues the same auth-shell thinning pattern: shared parsing helpers are centralized while page-specific state and routing behavior remain local.

## 2026-06-25 Reused shared auth login-success finalization across auth pages

### Done
- Reused `finalizeAuthLoginSuccess()` across `OAuthCallbackView`, `DingTalkEmailCompletionView`, and the main LinuxDo/OIDC/DingTalk/WeChat callback success paths.
- Removed another repeated cluster of token persistence, auth-store update, affiliate cleanup, success toast, and redirect logic from the auth page layer.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts src/views/auth/__tests__/OAuthCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `git diff --check -- frontend/src/views/auth/finalizeAuthLogin.ts frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/DingTalkEmailCompletionView.vue frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue progress.md`

### Notes
- This is another shell-thinning step: active auth page behavior stays unchanged while another shared success-path helper replaces repeated per-page implementations.

## 2026-06-24 Unified frontend pending-auth token field to the current single key

### Done
- Simplified frontend pending-auth session state to use only `pending_auth_token` instead of keeping a `pending_auth_token` / `pending_oauth_token` dual-field concept in the Vue auth store and auth views.
- Updated LinuxDo, OIDC, WeChat, and DingTalk callback flows so persisted pending sessions now consistently store `token_field: "pending_auth_token"`.
- Simplified `EmailVerifyView` pending-session typing to the single current field while keeping backend request compatibility intact.
- Updated auth-related store/view tests to assert the unified frontend field contract.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/stores/__tests__/auth.spec.ts src/views/auth/__tests__/EmailVerifyView.spec.ts src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/stores/auth.ts frontend/src/views/auth/EmailVerifyView.vue frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/stores/__tests__/auth.spec.ts frontend/src/views/auth/__tests__/EmailVerifyView.spec.ts frontend/src/views/auth/__tests__/LinuxDoCallbackView.spec.ts frontend/src/views/auth/__tests__/OidcCallbackView.spec.ts frontend/src/views/auth/__tests__/WechatCallbackView.spec.ts progress.md`

### Notes
- The backend still accepts both request fields for now; this slice only removes the dual-field concept from the active frontend flow so future cleanup can tighten the handler contract safely.

## 2026-06-24 Removed unused pending_oauth_token request field from pending-auth verify handler

### Done
- Removed the unused `PendingOAuthToken` field from `sendPendingOAuthVerifyCodeRequest` in `auth_oauth_pending_flow.go`.
- Removed the corresponding frontend test expectation that still asserted a `pending_oauth_token: undefined` payload member during email verification send-code requests.
- Verified that the active pending-auth verify flow still uses only `pending_auth_token` and that no handler logic depended on the retired request field.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/EmailVerifyView.spec.ts`
- `(cd backend && go test ./internal/handler -run 'Test.*Pending.*OAuth|Test.*EmailVerify.*' -count=1)`
- `pnpm run frontend:typecheck`
- `git diff --check -- backend/internal/handler/auth_oauth_pending_flow.go frontend/src/views/auth/__tests__/EmailVerifyView.spec.ts progress.md`

### Notes
- This is a narrow contract cleanup step following the frontend-side pending token unification. The backend can now be tightened further from a simpler baseline.

## 2026-06-24 Removed legacy pending-session auto-repair while preserving current pending-session short-circuit

### Done
- Deleted `legacyCompleteRegistrationSessionStatus` and its old auto-repair/update path for historical pending OAuth sessions that lacked a `step`.
- Added a smaller `currentCompleteRegistrationSessionStatus` helper that only preserves the active behavior: if the current pending session already carries a `step`, the complete-registration handlers return the normalized `pending_session` payload immediately.
- Simplified LinuxDo, OIDC, WeChat, and DingTalk complete-registration handlers to stop mutating old session shapes while keeping the live step-driven pending flow intact.

### Validation
- `(cd backend && go test ./internal/handler -run 'Test.*(LinuxDo|OIDC|WeChat|DingTalk|Pending).*' -count=1)`
- `git diff --check -- backend/internal/handler/auth_oauth_pending_flow.go backend/internal/handler/auth_linuxdo_oauth.go backend/internal/handler/auth_oidc_oauth.go backend/internal/handler/auth_wechat_oauth.go backend/internal/handler/auth_dingtalk_oauth.go progress.md`

### Notes
- This assumes there are no historical pending OAuth sessions in production that still depend on backend-side step auto-repair.

## 2026-06-24 Removed token_field from frontend pending-auth session state

### Done
- Removed `token_field` from the frontend pending-auth session store shape and persistence format in `frontend/src/stores/auth.ts`.
- Simplified `EmailVerifyView` to always submit `pending_auth_token` directly instead of carrying a field-selector variable through session storage and runtime state.
- Updated LinuxDo, OIDC, WeChat, and DingTalk callback views so `setPendingAuthSession()` only stores the actual pending session data still used by the active flow.
- Updated auth store and auth view tests to assert the reduced pending-auth session shape without `token_field`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/stores/__tests__/auth.spec.ts src/views/auth/__tests__/EmailVerifyView.spec.ts src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/stores/auth.ts frontend/src/views/auth/EmailVerifyView.vue frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/stores/__tests__/auth.spec.ts frontend/src/views/auth/__tests__/EmailVerifyView.spec.ts frontend/src/views/auth/__tests__/LinuxDoCallbackView.spec.ts frontend/src/views/auth/__tests__/OidcCallbackView.spec.ts frontend/src/views/auth/__tests__/WechatCallbackView.spec.ts progress.md`

### Notes
- The backend still accepts a `pending_auth_token` request field as before; this slice removes only the redundant frontend field-selector abstraction from the live flow.

## 2026-06-25 Extracted Prompt Catalog runtime helpers

### Done
- Added `frontend/src/views/public/promptCatalogRuntime.ts` to hold Prompt Catalog page-title, description, paging, sorting, generator, import, date-format, facet-label, and import-success helper logic.
- Simplified `PromptCatalogView.vue` so it now consumes those shared helpers instead of keeping the runtime decision logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/public/__tests__/promptCatalogRuntime.spec.ts`.
- Updated the Prompt Catalog source-guard test so it asserts the new helper usage instead of the old inline implementation details.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/promptCatalogRuntime.spec.ts src/views/public/__tests__/PromptCatalogView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/public/promptCatalogRuntime.ts frontend/src/views/public/PromptCatalogView.vue frontend/src/views/public/__tests__/promptCatalogRuntime.spec.ts frontend/src/views/public/__tests__/PromptCatalogView.spec.ts progress.md`

### Notes
- This slice only thins the Vue page shell. Prompt Catalog layout, interaction state, and routing still remain in the frontend as planned.

## 2026-06-25 Auth shell route defaults moved deeper into shared composable

### Done
- Added `resolveAuthRouteDefaultsFromShellDefaults()` in `frontend/src/router/setupRedirect.ts` so route-default merging no longer requires reparsing raw `auth_shell_config` everywhere.
- Extended `useAuthShellText()` to expose `applyAuthShellConfig()`, `authRouteDefaults`, `defaultRedirectPath`, and `defaultBindRedirectPath`.
- Simplified `AuthPopupView.vue`, `ForgotPasswordView.vue`, and `ResetPasswordView.vue` so they now consume shared auth route defaults instead of each view maintaining its own `FALLBACK_AUTH_ROUTE_DEFAULTS` route wiring.
- Kept the existing page behavior unchanged while removing duplicate auth-shell parsing in the thin auth pages.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/AuthPopupView.spec.ts src/views/auth/__tests__/ForgotPasswordView.auth-shell.spec.ts src/views/auth/__tests__/ResetPasswordView.auth-shell.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- This slice only thins the frontend auth shell. The main login/OAuth callback flow still has additional local runtime logic to keep cleaning up.

## 2026-06-25 OAuth callback pages switched to shared auth route/default outputs

### Done
- Updated `OAuthCallbackView.vue` to use `authRouteDefaults.loginPath` and the shared `defaultRedirectPath` from `useAuthShellText()` instead of keeping its own fallback route wiring.
- Updated `DingTalkEmailCompletionView.vue` to use `authRouteDefaults.value.dingtalkCallbackPath` and shared redirect defaults from the auth-shell composable.
- Kept the callback/create-account flow behavior unchanged while removing another layer of per-page auth route fallback state.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/OAuthCallbackView.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/composables/useAuthShellText.ts frontend/src/router/setupRedirect.ts frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/DingTalkEmailCompletionView.vue frontend/src/views/auth/__tests__/OAuthCallbackView.spec.ts frontend/src/views/auth/__tests__/CallbackAuthShell.spec.ts progress.md`

### Notes
- Larger provider callback pages such as LinuxDo/OIDC/WeChat/DingTalk still keep more local runtime state and remain the next auth-shell cleanup candidates.

## 2026-06-25 LinuxDo and DingTalk callback pages dropped local redirect fallback refs

### Done
- Updated `LinuxDoCallbackView.vue` to consume shared `defaultRedirectPath` and `defaultBindRedirectPath` from `useAuthShellText()` instead of keeping page-local fallback refs.
- Updated `DingTalkCallbackView.vue` to consume shared `defaultRedirectPath`, `defaultBindRedirectPath`, and `authRouteDefaults.value.dingtalkEmailCompletionPath`.
- Kept both callback pages' invitation, bind-login, TOTP, and pending-session state machines unchanged while removing another layer of duplicated auth-shell fallback state.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LinuxDoCallbackView.spec.ts src/views/auth/__tests__/DingTalkCallbackView.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- OIDC and WeChat callback pages still follow the same older pattern and remain the next straightforward auth-shell cleanup targets.

## 2026-06-25 OIDC and WeChat callback pages dropped local redirect fallback refs

### Done
- Updated `OidcCallbackView.vue` to consume shared `defaultRedirectPath` and `defaultBindRedirectPath` from `useAuthShellText()` instead of page-local fallback refs.
- Updated `WechatCallbackView.vue` to consume the same shared redirect defaults, while leaving its WeChat-specific availability and account-binding flow intact.
- Kept both callback pages' invitation, chooser, bind-login, TOTP, and pending-session behavior unchanged while removing another duplicated auth-shell fallback layer.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/OidcCallbackView.spec.ts src/views/auth/__tests__/WechatCallbackView.spec.ts src/views/auth/__tests__/CallbackAuthShell.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- Callback pages across all current OAuth providers now share the same auth-shell redirect-default source instead of each page owning its own fallback refs.

## 2026-06-25 Login and Register pages switched to shared auth-shell config outputs

### Done
- Updated `LoginView.vue` to use `useAuthShellText()` for `authText`, `authShellLabels`, `authRouteDefaults`, `defaultRedirectPath`, and `applyAuthShellConfig()`.
- Updated `RegisterView.vue` to use the same shared auth-shell composable outputs instead of manually parsing `auth_shell_config` in-page.
- Extended `useAuthShellText()` to expose `authShellLabels`, which keeps child auth components on the same shared label source.
- Kept both pages' submit, validation, agreement, OAuth-entry, and turnstile logic unchanged while removing duplicate auth-shell parsing.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/LoginView.turnstile.spec.ts src/views/auth/__tests__/RegisterView.auth-shell.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/composables/useAuthShellText.ts frontend/src/router/setupRedirect.ts frontend/src/views/auth/LoginView.vue frontend/src/views/auth/RegisterView.vue frontend/src/views/auth/OidcCallbackView.vue frontend/src/views/auth/WechatCallbackView.vue frontend/src/views/auth/DingTalkCallbackView.vue frontend/src/views/auth/LinuxDoCallbackView.vue frontend/src/views/auth/OAuthCallbackView.vue frontend/src/views/auth/DingTalkEmailCompletionView.vue frontend/src/views/auth/__tests__/LoginView.turnstile.spec.ts frontend/src/views/auth/__tests__/RegisterView.auth-shell.spec.ts frontend/src/views/auth/__tests__/CallbackAuthShell.spec.ts frontend/src/views/auth/__tests__/OAuthCallbackView.spec.ts frontend/src/views/auth/__tests__/LinuxDoCallbackView.spec.ts frontend/src/views/auth/__tests__/DingTalkCallbackView.spec.ts progress.md`

### Notes
- Remaining auth-shell cleanup is now concentrated in smaller pages like `EmailVerifyView.vue`, plus any page that still fetches public settings only to derive auth-shell routes or labels.

## 2026-06-25 EmailVerify page switched to shared auth-shell config outputs

### Done
- Updated `EmailVerifyView.vue` to use `useAuthShellText()` for `authText`, `authRouteDefaults`, and `applyAuthShellConfig()` instead of manually parsing `auth_shell_config` in-page.
- Kept both normal email registration verification and pending OAuth completion verification flows unchanged while removing duplicate auth-shell parsing from the page.
- Updated the EmailVerify source-guard test to assert the new shared auth-shell entry points.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/EmailVerifyView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/auth/EmailVerifyView.vue frontend/src/views/auth/__tests__/EmailVerifyView.spec.ts progress.md`

### Notes
- The remaining auth-shell work is now mostly scattered route/default readers such as `WechatPaymentCallbackView.vue` and other smaller auth-related entry pages.

## 2026-06-25 WeChat payment callback page switched to shared auth route defaults

### Done
- Updated `WechatPaymentCallbackView.vue` to use `useAuthRouteDefaults()` instead of resolving auth route defaults directly inside the page.
- Kept the payment-shell label source unchanged while removing one more page-local auth route parsing path.
- Updated the source-guard test to assert the shared auth-route composable usage.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/auth/__tests__/WechatPaymentCallbackView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/auth/WechatPaymentCallbackView.vue frontend/src/views/auth/__tests__/WechatPaymentCallbackView.spec.ts progress.md`

### Notes
- Remaining frontend shell work is now less about auth fallback parsing and more about other small local shell/default readers outside the main auth flow.

## 2026-06-25 User subscriptions and API guide pages switched to shared auth route defaults

### Done
- Updated `SubscriptionsView.vue` to consume `useAuthRouteDefaults()` instead of resolving auth route defaults directly in-page.
- Updated `ApiGuideView.vue` to use the same shared auth-route composable for API key navigation targets.
- Kept payment-shell/API-guide-shell behavior unchanged while removing two more page-local auth route parsing paths.
- Updated source-guard tests in `SubscriptionsView.spec.ts` and `ApiGuideView.spec.ts` to assert the shared auth-route composable usage.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/SubscriptionsView.spec.ts src/views/user/__tests__/ApiGuideView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/SubscriptionsView.vue frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/__tests__/SubscriptionsView.spec.ts frontend/src/views/user/__tests__/ApiGuideView.spec.ts progress.md`

### Notes
- Remaining shell-thinning work is now increasingly in page-specific fallback/render helpers rather than direct auth route parsing.

## 2026-06-25 Credits page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/creditsRuntime.ts` to hold credits ratio parsing, conversion/balance label rendering, button/action fallback resolution, and purchase-route building.
- Simplified `CreditsView.vue` so it now consumes shared credits runtime helpers instead of keeping those rules inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/creditsRuntime.spec.ts`.
- Updated the Credits page source-guard test to assert helper usage instead of inline purchase-path and route builder logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/creditsRuntime.spec.ts src/views/user/__tests__/CreditsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/CreditsView.vue frontend/src/views/user/creditsRuntime.ts frontend/src/views/user/__tests__/CreditsView.spec.ts frontend/src/views/user/__tests__/creditsRuntime.spec.ts progress.md`

### Notes
- This slice thins one more frontend shell page without changing the underlying user balance/refresh flow.

## 2026-06-25 Legal document page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/public/legalDocumentRuntime.ts` to hold current-document resolution, document-icon selection, and sanitized HTML rendering for agreement documents.
- Simplified `LegalDocumentView.vue` so it now consumes those helpers instead of keeping the runtime document-selection and render logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/public/__tests__/legalDocumentRuntime.spec.ts`.
- Updated the Legal Document page source-guard test to assert helper usage instead of inline runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/legalDocumentRuntime.spec.ts src/views/public/__tests__/LegalDocumentView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/legalDocumentRuntime.ts frontend/src/views/public/__tests__/LegalDocumentView.spec.ts frontend/src/views/public/__tests__/legalDocumentRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing public-settings fetch path intact and only removes page-local runtime shaping logic from the view.

## 2026-06-25 Pricing page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/public/pricingRuntime.ts` to hold catalog sorting, shell label/group lookup, purchase-route building, currency/credits formatting, plan source label, and validity rendering logic.
- Simplified `PricingView.vue` so it now consumes those shared runtime helpers instead of keeping the pricing page rules inline in the view component.
- Added focused helper coverage in `frontend/src/views/public/__tests__/pricingRuntime.spec.ts`.
- Updated the Pricing page source-guard test to assert helper usage instead of inline purchase-path and runtime helper logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/pricingRuntime.spec.ts src/views/public/__tests__/PricingView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/public/PricingView.vue frontend/src/views/public/pricingRuntime.ts frontend/src/views/public/__tests__/PricingView.spec.ts frontend/src/views/public/__tests__/pricingRuntime.spec.ts progress.md`

### Notes
- This slice only thins the frontend pricing shell and does not change the catalog API fetch path or purchase flow behavior.

## 2026-06-25 Docs page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/public/docsRuntime.ts` to hold docs search namespace building, docs hash normalization, docs content version cache-busting, and initial deep-link preservation helpers.
- Simplified `DocsView.vue` so it now consumes those shared docs runtime helpers instead of keeping the hash/version/namespace rules inline in the page component.
- Added focused helper coverage in `frontend/src/views/public/__tests__/docsRuntime.spec.ts`.
- Updated the Docs page source-guard test to assert helper usage instead of inline runtime helper implementations.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/docsRuntime.spec.ts src/views/public/__tests__/DocsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/public/DocsView.vue frontend/src/views/public/docsRuntime.ts frontend/src/views/public/__tests__/DocsView.spec.ts frontend/src/views/public/__tests__/docsRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing Docsify runtime/bootstrap path intact and only removes page-local runtime shaping logic from the docs view.

## 2026-06-25 Models plaza page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/public/modelsPlazaRuntime.ts` to hold visible-item filtering, provider-group option building, active-group label resolution, and search filtering for the model plaza page.
- Simplified `ModelsPlazaView.vue` so it now consumes those runtime helpers instead of keeping the item/group/search shaping logic inline in the view component.
- Added focused helper coverage in `frontend/src/views/public/__tests__/modelsPlazaRuntime.spec.ts`.
- Updated the Models Plaza page source-guard test to assert helper usage instead of inline item/group/search runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/modelsPlazaRuntime.spec.ts src/views/public/__tests__/ModelsPlazaView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/public/ModelsPlazaView.vue frontend/src/views/public/modelsPlazaRuntime.ts frontend/src/views/public/__tests__/ModelsPlazaView.spec.ts frontend/src/views/public/__tests__/modelsPlazaRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing provider badge/icon/copy helpers in `modelPlazaDisplay.ts` and only removes page-local runtime shaping logic from the plaza view.

## 2026-06-25 Dashboard page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/dashboardRuntime.ts` to hold dashboard local-date formatting, default start/end date resolution, and recent-usage slicing logic.
- Simplified `DashboardView.vue` so it now consumes those shared helpers instead of keeping the date-range and recent-usage runtime rules inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/dashboardRuntime.spec.ts`.
- Updated the dashboard source-guard test to assert helper usage instead of inline date/recent-usage runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/dashboardRuntime.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/DashboardView.vue frontend/src/views/user/dashboardRuntime.ts frontend/src/views/user/__tests__/dashboardRuntime.spec.ts frontend/src/views/user/__tests__/dashboardNoHero.spec.ts progress.md`

### Notes
- This slice does not change the dashboard API request structure; it only removes page-local runtime shaping logic from the dashboard view.

## 2026-06-25 Usage page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/usageRuntime.ts` to hold usage local-date formatting, default date-range resolution, and table query param building.
- Simplified `UsageView.vue` so it now consumes those shared helpers instead of keeping the date-range and query-shaping rules inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/usageRuntime.spec.ts`.
- Updated the usage source-guard test to assert helper usage instead of inline date-range/query runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/usageRuntime.spec.ts src/views/user/__tests__/UsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/UsageView.vue frontend/src/views/user/usageRuntime.ts frontend/src/views/user/__tests__/UsageView.spec.ts frontend/src/views/user/__tests__/usageRuntime.spec.ts progress.md`

### Notes
- This slice leaves the existing tooltip, export, and usage metric rendering logic intact; it only removes page-local runtime shaping logic from the usage view.

## 2026-06-25 API test page switched to shared auth route defaults

### Done
- Updated `frontend/src/views/user/ApiTestView.vue` to consume `useAuthRouteDefaults()` instead of resolving auth route defaults directly in-page.
- Kept API test shell labels/defaults and gateway override logic unchanged while removing one more page-local auth route parsing path.
- Updated the API test source-guard test to assert the shared auth-route composable usage.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiTestView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/ApiTestView.vue frontend/src/views/user/__tests__/ApiTestView.spec.ts progress.md`

### Notes
- `LegalDocumentView.vue` is still the only page-level direct `resolveAuthRouteDefaults(...)` reader in the current frontend views because it uses a page-local public-settings fetch instead of `appStore.cachedPublicSettings`.

## 2026-06-25 Legal document page switched to app store public settings cache

### Done
- Updated `frontend/src/views/public/LegalDocumentView.vue` to read public settings from `appStore.cachedPublicSettings` instead of issuing its own page-local `getPublicSettings()` call.
- Reused `useAuthRouteDefaults()` for login/home navigation targets, which removes the last page-level direct `resolveAuthRouteDefaults(...)` usage from current frontend views.
- Kept the legal document runtime helpers and document rendering behavior unchanged while unifying the settings source with the rest of the frontend.
- Updated `LegalDocumentView.spec.ts` to mock the app store cache path and verify cold-cache fetch behavior through `appStore.fetchPublicSettings()`.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/LegalDocumentView.spec.ts src/views/setup/__tests__/SetupWizardView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/public/LegalDocumentView.vue frontend/src/views/public/__tests__/LegalDocumentView.spec.ts frontend/src/views/setup/SetupWizardView.vue frontend/src/views/setup/__tests__/SetupWizardView.spec.ts progress.md`

### Notes
- This removes the last frontend view-level direct auth-route resolver call while keeping the legal document page’s fetch timing unchanged.

## 2026-06-25 Setup wizard stopped referencing fallback auth route constants directly

### Done
- Updated `frontend/src/views/setup/SetupWizardView.vue` to derive the post-install login redirect path from `resolveAuthRouteDefaultsFromShellDefaults().loginPath` instead of importing `FALLBACK_AUTH_ROUTE_DEFAULTS` directly.
- Updated the setup wizard source-guard test to assert the centralized route-default resolver usage.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/LegalDocumentView.spec.ts src/views/setup/__tests__/SetupWizardView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`

### Notes
- The setup wizard still uses a pre-auth bootstrap redirect, but it now routes through the centralized auth-route default resolver.

## 2026-06-25 Dashboard shell config stopped importing auth fallback route constants

### Done
- Updated `frontend/src/components/user/dashboard/dashboardShellLabels.ts` so `resolveDashboardShellConfig()` now requires explicit quick-action fallback paths instead of importing `FALLBACK_AUTH_ROUTE_DEFAULTS` itself.
- Kept `DashboardView.vue` as the caller that supplies current auth-route defaults, which makes the dashboard shell config a pure shell parser instead of a mixed parser + auth fallback source.
- Updated dashboard shell tests to pass explicit fallback paths and assert the shared shell config no longer imports auth fallback constants directly.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts src/views/user/__tests__/dashboardNoHero.spec.ts src/views/user/__tests__/dashboardRuntime.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/components/user/dashboard/dashboardShellLabels.ts frontend/src/components/user/dashboard/__tests__/dashboardShellLabels.spec.ts frontend/src/views/user/DashboardView.vue frontend/src/views/user/dashboardRuntime.ts frontend/src/views/user/__tests__/dashboardNoHero.spec.ts frontend/src/views/user/__tests__/dashboardRuntime.spec.ts progress.md`

### Notes
- This removes one of the last shared-shell imports of `FALLBACK_AUTH_ROUTE_DEFAULTS` from the user dashboard stack.

## 2026-06-25 App sidebar section assembly extracted into shared runtime helpers

### Done
- Added `frontend/src/components/layout/sidebarRuntime.ts` to hold sidebar feature-flag filtering, visible item map building, and section assembly for shared/user/admin sidebar nav structures.
- Simplified `frontend/src/components/layout/AppSidebar.vue` so it now delegates sidebar section assembly to shared runtime helpers instead of keeping that logic inline in the layout component.
- Added focused helper coverage in `frontend/src/components/layout/__tests__/sidebarRuntime.spec.ts`.
- Updated the AppSidebar source-guard test to assert shared runtime helper usage.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/layout/__tests__/sidebarRuntime.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/components/layout/AppSidebar.vue frontend/src/components/layout/sidebarRuntime.ts frontend/src/components/layout/__tests__/AppSidebar.spec.ts frontend/src/components/layout/__tests__/sidebarRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing sidebar template, icon definitions, and onboarding/toggle interactions intact while removing the heavier shared nav assembly logic from the component body.

## 2026-06-25 Auth layout switched to shared auth shell composable

### Done
- Updated `frontend/src/components/layout/AuthLayout.vue` to consume `useAuthShellText()` instead of resolving auth shell labels directly in the layout.
- Kept the existing brand block and auth footer behavior unchanged while removing one more shared-layout-local auth shell parsing path.
- Updated `AuthLayout.spec.ts` to assert the shared auth shell composable usage.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/layout/__tests__/AuthLayout.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts src/components/layout/__tests__/sidebarRuntime.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/components/layout/AuthLayout.vue frontend/src/components/layout/__tests__/AuthLayout.spec.ts progress.md`

### Notes
- This removes the last obvious shared-layout-local auth shell label parsing path from the auth container.

## 2026-06-25 App header runtime helpers extracted out of the component

### Done
- Added `frontend/src/components/layout/appHeaderRuntime.ts` to hold shared header display-name, initials, compact-dropdown, and page-title shaping logic.
- Simplified `frontend/src/components/layout/AppHeader.vue` so it now consumes those runtime helpers instead of keeping the header runtime shaping logic inline in the component body.
- Added focused helper coverage in `frontend/src/components/layout/__tests__/appHeaderRuntime.spec.ts`.
- Updated the AppHeader source-guard test to assert helper usage instead of inline runtime shaping logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/components/layout/__tests__/appHeaderRuntime.spec.ts src/components/layout/__tests__/AppHeader.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/components/layout/AppHeader.vue frontend/src/components/layout/appHeaderRuntime.ts frontend/src/components/layout/__tests__/AppHeader.spec.ts frontend/src/components/layout/__tests__/appHeaderRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing docs shortcut, dropdown behavior, and logout flow intact while removing more shared-layout-local runtime shaping logic from the header component.

## 2026-06-25 Profile page removed local profile shell config fallback

### Done
- Updated `frontend/src/views/user/ProfileView.vue` so profile shell labels now read only from `appStore.cachedPublicSettings?.profile_shell_config` instead of keeping a local `fetchedProfileShellConfig` fallback.
- Kept the existing public-settings fetch for runtime flags such as contact info, notify settings, TOTP, and OAuth capability booleans, while aligning the shell config source with the rest of the frontend.
- Updated `ProfileView.spec.ts` to model the real cache-based shell config path and added a source-guard assertion that the local fetched shell-config fallback no longer exists.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ProfileView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/ProfileView.vue frontend/src/views/user/__tests__/ProfileView.spec.ts progress.md`

### Notes
- This leaves the profile page’s runtime flag fetch path intact while removing one more local shell-config fallback from the view.

## 2026-06-25 Image generator page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/public/imageGeneratorRuntime.ts` to hold prompt draft application and catalog-path normalization for the image generator page.
- Simplified `ImageGeneratorView.vue` so it now consumes those shared helpers instead of keeping that runtime shaping logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/public/__tests__/imageGeneratorRuntime.spec.ts`.
- Updated the Image Generator page source-guard test to assert helper usage instead of inline draft/path runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/public/__tests__/imageGeneratorRuntime.spec.ts src/views/public/__tests__/ImageGeneratorView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/public/ImageGeneratorView.vue frontend/src/views/public/imageGeneratorRuntime.ts frontend/src/views/public/__tests__/ImageGeneratorView.spec.ts frontend/src/views/public/__tests__/imageGeneratorRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing workspace shell config parsing intact and only removes page-local runtime shaping logic from the image generator view.

## 2026-06-25 Stripe popup page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/stripePopupRuntime.ts` to hold Stripe popup route-state normalization, display-amount formatting, and payment-result return URL construction.
- Simplified `StripePopupView.vue` so it now consumes those helpers instead of keeping the popup route parsing and return URL logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/stripePopupRuntime.spec.ts`.
- Updated the Stripe popup source-guard test to assert helper usage instead of inline popup route/runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/stripePopupRuntime.spec.ts src/views/user/__tests__/StripePopupView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/StripePopupView.vue frontend/src/views/user/stripePopupRuntime.ts frontend/src/views/user/__tests__/StripePopupView.spec.ts frontend/src/views/user/__tests__/stripePopupRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing Stripe popup polling and payment-method-specific behavior intact while removing page-local runtime shaping logic from the popup view.

## 2026-06-25 Payment result page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/paymentResultRuntime.ts` to hold payment base/fee amount calculation, result status normalization, route query parsing, resolved-order currency application, and recovery snapshot matching/cleanup helpers.
- Simplified `PaymentResultView.vue` so it now consumes those helpers instead of keeping the payment-result runtime shaping logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/paymentResultRuntime.spec.ts`.
- Updated the payment result source-guard test to assert helper usage instead of inline status/query/recovery runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/paymentResultRuntime.spec.ts src/views/user/__tests__/PaymentResultView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/PaymentResultView.vue frontend/src/views/user/paymentResultRuntime.ts frontend/src/views/user/__tests__/PaymentResultView.spec.ts frontend/src/views/user/__tests__/paymentResultRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing payment-result polling and order resolution flow intact while removing page-local runtime shaping logic from the result view.

## 2026-06-25 Airwallex payment page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/airwallexPaymentRuntime.ts` to hold Airwallex route query parsing, success URL construction, and payment recovery snapshot restoration.
- Simplified `AirwallexPaymentView.vue` so it now consumes those shared helpers instead of keeping the Airwallex route/snapshot runtime logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/airwallexPaymentRuntime.spec.ts`.
- Updated the Airwallex payment source-guard test to assert helper usage instead of inline route/snapshot runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/airwallexPaymentRuntime.spec.ts src/views/user/__tests__/AirwallexPaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/AirwallexPaymentView.vue frontend/src/views/user/airwallexPaymentRuntime.ts frontend/src/views/user/__tests__/AirwallexPaymentView.spec.ts frontend/src/views/user/__tests__/airwallexPaymentRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing Airwallex init/redirect behavior intact while removing page-local runtime shaping logic from the Airwallex payment view.

## 2026-06-25 Payment QR page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/paymentQrRuntime.ts` to hold QR page route-state parsing, countdown formatting, seconds-until-expiry calculation, and payment status classification helpers.
- Simplified `PaymentQRCodeView.vue` so it now consumes those shared helpers instead of keeping the QR-page route/countdown/status runtime logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/paymentQrRuntime.spec.ts`.
- Updated the QR page source-guard test to assert helper usage instead of inline route/countdown/runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/paymentQrRuntime.spec.ts src/views/user/__tests__/PaymentQRCodeView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/PaymentQRCodeView.vue frontend/src/views/user/paymentQrRuntime.ts frontend/src/views/user/__tests__/PaymentQRCodeView.spec.ts frontend/src/views/user/__tests__/paymentQrRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing QR rendering, payment polling, and cancel flow intact while removing page-local runtime shaping logic from the QR payment view.

## 2026-06-25 User orders page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/userOrdersRuntime.ts` to hold user-orders status filter construction, order-table label mapping, and refund eligibility checks.
- Simplified `UserOrdersView.vue` so it now consumes those helpers instead of keeping those runtime shaping rules inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/userOrdersRuntime.spec.ts`.
- Updated the user orders source-guard test to assert helper usage instead of inline orders-page runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/userOrdersRuntime.spec.ts src/views/user/__tests__/UserOrdersView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/UserOrdersView.vue frontend/src/views/user/userOrdersRuntime.ts frontend/src/views/user/__tests__/UserOrdersView.spec.ts frontend/src/views/user/__tests__/userOrdersRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing orders fetch, cancel, and refund request flows intact while removing page-local runtime shaping logic from the user orders view.

## 2026-06-25 Stripe payment page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/stripePaymentRuntime.ts` to hold Stripe payment route-state parsing, gateway amount formatting, restored currency resolution, and payment-result route/URL construction.
- Simplified `StripePaymentView.vue` so it now consumes those helpers instead of keeping the Stripe route/currency/result runtime logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/stripePaymentRuntime.spec.ts`.
- Updated the Stripe payment source-guard test to assert helper usage instead of inline route/result runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/stripePaymentRuntime.spec.ts src/views/user/__tests__/StripePaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/StripePaymentView.vue frontend/src/views/user/stripePaymentRuntime.ts frontend/src/views/user/__tests__/StripePaymentView.spec.ts frontend/src/views/user/__tests__/stripePaymentRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing Stripe method-specific confirmation flow and polling behavior intact while removing page-local runtime shaping logic from the Stripe payment view.

## 2026-06-25 Available groups page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/availableGroupsRuntime.ts` to hold available-groups search filtering, public/member grouping, subscription badge selection, and quota summary rendering.
- Simplified `AvailableGroupsView.vue` so it now consumes those shared helpers instead of keeping the available-groups runtime shaping logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/availableGroupsRuntime.spec.ts`.
- Updated the available-groups source-guard test to assert helper usage instead of inline grouping/search/quota runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/availableGroupsRuntime.spec.ts src/views/user/__tests__/AvailableGroupsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/AvailableGroupsView.vue frontend/src/views/user/availableGroupsRuntime.ts frontend/src/views/user/__tests__/AvailableGroupsView.spec.ts frontend/src/views/user/__tests__/availableGroupsRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing available-groups fetch flow intact while removing page-local runtime shaping logic from the available groups view.

## 2026-06-25 Available channels page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/availableChannelsRuntime.ts` to hold available-channels search filtering plus shared column/pricing label mapping.
- Simplified `AvailableChannelsView.vue` so it now consumes those helpers instead of keeping the available-channels runtime shaping logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/availableChannelsRuntime.spec.ts`.
- Updated the available-channels source-guard test to assert helper usage instead of inline search/label runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/availableChannelsRuntime.spec.ts src/views/user/__tests__/AvailableChannelsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/AvailableChannelsView.vue frontend/src/views/user/availableChannelsRuntime.ts frontend/src/views/user/__tests__/AvailableChannelsView.spec.ts frontend/src/views/user/__tests__/availableChannelsRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing available-channels fetch flow intact while removing page-local runtime shaping logic from the available channels view.

## 2026-06-25 Key usage page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/keyUsageRuntime.ts` to hold key-usage date param construction, status info shaping, and reset-time formatting helpers.
- Simplified `KeyUsageView.vue` so it now consumes those helpers instead of keeping that runtime logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/__tests__/keyUsageRuntime.spec.ts`.
- Updated the key usage source-guard test to assert helper usage instead of inline date/status runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/__tests__/keyUsageRuntime.spec.ts src/views/__tests__/KeyUsageView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/KeyUsageView.vue frontend/src/views/keyUsageRuntime.ts frontend/src/views/__tests__/KeyUsageView.spec.ts frontend/src/views/__tests__/keyUsageRuntime.spec.ts progress.md`

### Notes
- This slice leaves the existing ring rendering, tooltip details, and fetch flow intact while removing page-local runtime shaping logic from the key usage view.

## 2026-06-25 API guide page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/apiGuideRuntime.ts` to hold API guide key masking, key-option construction, and auth-header preview selection.
- Simplified `ApiGuideView.vue` so it now consumes those shared helpers instead of keeping the API guide runtime shaping logic inline in the page component.
- Updated the API guide source-guard test to assert helper usage instead of inline key/auth-header runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiGuideView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/apiGuideRuntime.ts frontend/src/views/user/__tests__/ApiGuideView.spec.ts progress.md`

### Notes
- This slice keeps the existing API guide shell config parsing and gateway variant resolution intact while removing a small chunk of page-local runtime shaping logic from the API guide view.

## 2026-06-25 Custom page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/customPageRuntime.ts` to hold custom page menu item resolution, markdown slug resolution, relative markdown asset checks, and markdown image URL construction.
- Simplified `CustomPageView.vue` so it now consumes those helpers instead of keeping that runtime shaping logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/customPageRuntime.spec.ts`.
- Updated the custom page source-guard test to assert helper usage instead of inline menu/markdown runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/customPageRuntime.spec.ts src/views/user/__tests__/CustomPageView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/CustomPageView.vue frontend/src/views/user/customPageRuntime.ts frontend/src/views/user/__tests__/CustomPageView.spec.ts frontend/src/views/user/__tests__/customPageRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing embedded/markdown fetch behavior intact while removing page-local runtime shaping logic from the custom page view.

## 2026-06-25 Redeem page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/redeemRuntime.ts` to hold redeem history type classification, history title/value formatting, and related display helpers.
- Simplified `RedeemView.vue` so it now consumes those shared helpers instead of keeping the redeem history runtime shaping logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/redeemRuntime.spec.ts`.
- Updated the redeem source-guard test to assert helper usage instead of inline redeem history runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/redeemRuntime.spec.ts src/views/user/__tests__/RedeemView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/RedeemView.vue frontend/src/views/user/redeemRuntime.ts frontend/src/views/user/__tests__/RedeemView.spec.ts frontend/src/views/user/__tests__/redeemRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing redeem submit/history fetch flow intact while removing page-local runtime shaping logic from the redeem view.

## 2026-06-25 API guide page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/apiGuideRuntime.ts` to hold API guide key masking, key option construction, and auth-header preview resolution.
- Simplified `ApiGuideView.vue` so it now consumes those shared helpers instead of keeping that runtime shaping logic inline in the page component.
- Updated the API guide source-guard test to assert helper usage instead of inline key/auth-header runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/ApiGuideView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/ApiGuideView.vue frontend/src/views/user/apiGuideRuntime.ts frontend/src/views/user/__tests__/ApiGuideView.spec.ts progress.md`

### Notes
- This slice keeps the existing API guide shell config parsing and gateway variant resolution intact while removing a small chunk of page-local runtime shaping logic from the API guide view.

## 2026-06-25 Payment page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/paymentViewRuntime.ts` to hold empty payment recovery state creation, payment-result query construction, and WeChat OAuth redirect URL assembly.
- Simplified `PaymentView.vue` so it now consumes those helpers instead of keeping those runtime shaping rules inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/paymentViewRuntime.spec.ts`.
- Updated the payment page source-guard test to assert helper usage instead of inline payment-result / WeChat redirect runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/paymentViewRuntime.spec.ts src/views/user/__tests__/PaymentView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/PaymentView.vue frontend/src/views/user/paymentViewRuntime.ts frontend/src/views/user/__tests__/PaymentView.spec.ts frontend/src/views/user/__tests__/paymentViewRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing checkout creation, WeChat JSAPI/H5 fallback, and payment-launch decision flow intact while removing another chunk of page-local runtime shaping logic from the main payment view.

## 2026-06-25 Channel status page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/channelStatusRuntime.ts` to hold overall status reduction, detail title resolution, detail-window loading rules, and immutable detail-cache updates.
- Simplified `ChannelStatusView.vue` so it now consumes those shared helpers instead of keeping the channel-status runtime shaping logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/channelStatusRuntime.spec.ts`.
- Updated the channel-status source-guard test to assert helper usage instead of inline status/detail runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/channelStatusRuntime.spec.ts src/views/user/__tests__/ChannelStatusView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/ChannelStatusView.vue frontend/src/views/user/channelStatusRuntime.ts frontend/src/views/user/__tests__/ChannelStatusView.spec.ts frontend/src/views/user/__tests__/channelStatusRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing channel monitor fetch and auto-refresh flow intact while removing page-local runtime shaping logic from the channel status view.

## 2026-06-25 Subscriptions page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/subscriptionsRuntime.ts` to hold subscription status text, progress styling, expiration formatting, and quota-window/reset-time rendering helpers.
- Simplified `SubscriptionsView.vue` so it now consumes those shared helpers instead of keeping the subscriptions runtime shaping logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/subscriptionsRuntime.spec.ts`.
- Updated the subscriptions source-guard test to assert helper usage instead of inline subscription runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/subscriptionsRuntime.spec.ts src/views/user/__tests__/SubscriptionsView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/SubscriptionsView.vue frontend/src/views/user/subscriptionsRuntime.ts frontend/src/views/user/__tests__/SubscriptionsView.spec.ts frontend/src/views/user/__tests__/subscriptionsRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing subscriptions fetch flow intact while removing page-local runtime shaping logic from the subscriptions view.

## 2026-06-25 Affiliate page runtime helpers extracted out of the view

### Done
- Added `frontend/src/views/user/affiliateRuntime.ts` to hold affiliate count/currency/payment-type formatting, order-status casting, and pagination change helpers.
- Simplified `AffiliateView.vue` so it now consumes those shared helpers instead of keeping the affiliate runtime shaping logic inline in the page component.
- Added focused helper coverage in `frontend/src/views/user/__tests__/affiliateRuntime.spec.ts`.
- Updated the affiliate source-guard test to assert helper usage instead of inline affiliate runtime logic.

### Validation
- `pnpm --filter sub2api-frontend exec vitest run src/views/user/__tests__/affiliateRuntime.spec.ts src/views/user/__tests__/AffiliateView.spec.ts`
- `pnpm run frontend:typecheck`
- `pnpm run frontend:lint:check`
- `git diff --check -- frontend/src/views/user/AffiliateView.vue frontend/src/views/user/affiliateRuntime.ts frontend/src/views/user/__tests__/AffiliateView.spec.ts frontend/src/views/user/__tests__/affiliateRuntime.spec.ts progress.md`

### Notes
- This slice keeps the existing affiliate detail/rebate/transfer fetch flow intact while removing page-local runtime shaping logic from the affiliate view.

## 2026-06-25 Focused inventory after shell/runtime thinning pass

### Current high-ROI remaining items
- Several user/public pages still keep page-specific runtime helpers inline, but the remaining ones are smaller than the pages already extracted.
- Shared layout hotspots are now much smaller; remaining work is more about scattered page-level runtime shaping than one dominant shared-shell bottleneck.
- Remaining higher-value cleanup is now mostly scattered page-level runtime/helper extraction and selective preload/fetch-timing cleanup, rather than any single obvious Touch-specific choke point.

### Lower-priority/expected leftovers
- `touch_*` strings remain in migrations, compatibility tests, and docs guardrails by design.
- `signup_source = touch` separation remains active in backend auth/service/repository code by design and is still required by the chosen identity model.
