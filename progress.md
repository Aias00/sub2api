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
