# 邀请中心 / 邀请管理模块 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reorganize the existing invitation-code and affiliate-rebate capabilities into a dedicated user “邀请中心” and admin “邀请管理” module without changing rebate calculation rules.

**Architecture:** Reuse the current affiliate domain model and ledger logic, add missing user/admin aggregate/query APIs, then reorganize routing/navigation and page composition around those APIs. Keep old routes/data intact where possible and prefer additive interfaces over data-model changes.

**Tech Stack:** Go, Gin, Ent/PostgreSQL, Vue 3, Pinia, Vitest, existing affiliate/user/admin APIs

---

### Task 1: Add failing backend tests for new affiliate module APIs

**Files:**
- Create: `backend/internal/handler/admin/affiliate_handler_test.go`
- Modify: `backend/internal/handler/user_handler_test.go`
- Reference: `backend/internal/handler/admin/admin_basic_handlers_test.go`

- [ ] **Step 1: Write failing user-handler tests**
  - Cover `GET /api/v1/user/aff/rebates`
  - Cover `GET /api/v1/user/aff/transfers`
  - Assert auth subject scoping and response shape

- [ ] **Step 2: Write failing admin-handler tests**
  - Cover `GET /api/v1/admin/affiliates/overview`
  - Cover `GET /api/v1/admin/affiliates/rules`
  - Cover `PUT /api/v1/admin/affiliates/rules`

- [ ] **Step 3: Run focused handler tests and verify they fail**
  - Run: `cd backend && go test ./internal/handler ./internal/handler/admin -run 'Test(UserHandlerAffiliate|AffiliateHandler)' -count=1`

### Task 2: Implement backend affiliate module APIs

**Files:**
- Modify: `backend/internal/service/affiliate_service.go`
- Modify: `backend/internal/repository/affiliate_repo.go`
- Modify: `backend/internal/handler/user_handler.go`
- Modify: `backend/internal/handler/admin/affiliate_handler.go`
- Modify: `backend/internal/server/routes/user.go`
- Modify: `backend/internal/server/routes/admin.go`

- [ ] **Step 1: Add service/repository types for user rebates/transfers and admin overview/rules**
- [ ] **Step 2: Implement repository queries**
  - Current-user rebate records
  - Current-user transfer records
  - Admin overview aggregates

- [ ] **Step 3: Implement service methods**
  - User summary reuse
  - User rebates/transfers list
  - Admin overview read
  - Admin rules read/update

- [ ] **Step 4: Expose handler endpoints and routes**

- [ ] **Step 5: Run focused backend tests until green**
  - Run: `cd backend && go test ./internal/handler ./internal/handler/admin -run 'Test(UserHandlerAffiliate|AffiliateHandler)' -count=1`

### Task 3: Build admin invitation-management module UI

**Files:**
- Create: `frontend/src/views/admin/affiliates/AdminAffiliateOverviewView.vue`
- Create: `frontend/src/views/admin/affiliates/AdminAffiliateRulesView.vue`
- Create: `frontend/src/views/admin/affiliates/AdminAffiliateCodesView.vue`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/components/layout/AppSidebar.vue`
- Modify: `frontend/src/api/admin/affiliates.ts`
- Modify: `frontend/src/views/admin/SettingsView.vue`
- Modify: `frontend/src/i18n/locales/zh.ts`
- Modify: `frontend/src/i18n/locales/en.ts`

- [ ] **Step 1: Write or extend frontend tests for admin affiliate navigation/API usage**
- [ ] **Step 2: Add admin routes and sidebar children for overview/rules/codes**
- [ ] **Step 3: Implement overview/rules/codes views using existing affiliate APIs plus new rules/overview endpoints**
- [ ] **Step 4: Replace the old SettingsView affiliate section with a jump/entry into the new module**
- [ ] **Step 5: Run focused frontend tests and verify they pass**

### Task 4: Rework `/affiliate` into a full invite center

**Files:**
- Modify: `frontend/src/views/user/AffiliateView.vue`
- Modify: `frontend/src/api/user.ts`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/i18n/locales/zh.ts`
- Modify: `frontend/src/i18n/locales/en.ts`

- [ ] **Step 1: Write or extend frontend tests for invite-center sections**
- [ ] **Step 2: Add user API clients for rebates/transfers**
- [ ] **Step 3: Recompose `/affiliate` into overview + invitees + rebates + transfers**
- [ ] **Step 4: Keep existing transfer action and share actions intact**
- [ ] **Step 5: Run focused frontend tests and verify they pass**

### Task 5: Full verification and cleanup

**Files:**
- Modify: `progress.md`

- [ ] **Step 1: Run backend affiliate-related tests**
  - Run: `cd backend && go test ./internal/service ./internal/handler ./internal/handler/admin -count=1`

- [ ] **Step 2: Run full targeted frontend verification**
  - Run: `pnpm --dir frontend run typecheck`
  - Run: `pnpm --dir frontend exec eslint src/views/user/AffiliateView.vue src/views/admin/affiliates/*.vue src/components/layout/AppSidebar.vue src/router/index.ts src/api/admin/affiliates.ts src/api/user.ts src/i18n/locales/zh.ts src/i18n/locales/en.ts --ext .vue,.ts`

- [ ] **Step 3: Run any focused Vitest suites added in this work**

- [ ] **Step 4: Run `git diff --check`**

- [ ] **Step 5: Update `progress.md` with done/failures/next**
