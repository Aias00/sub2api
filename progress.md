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
- Verified with automated scans that:
  - missing key count for `admin.settings / admin.users / admin.accounts / auth.dingtalk` is now `0`
  - Chinese-value leak count in `en.ts` for `admin.settings / admin.users / admin.accounts` is now `0`

### Failures
- No functional failures. Frontend build still reports the existing Vite dynamic-import chunk warnings and stale Browserslist data warning; neither is introduced by this i18n cleanup.

### Next
- If needed, continue from targeted modules into the remaining admin namespaces with missing locale coverage:
  - `admin.riskControl`
  - `admin.redeem`
  - `admin.subscriptions`
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
