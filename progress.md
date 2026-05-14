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
