.PHONY: build build-core build-backend build-frontend build-all build-datamanagementd test test-core test-backend test-frontend test-frontend-critical test-datamanagementd validate-prompt-catalog validate-prompt-catalog-parity validate-prompt-catalog-external-images validate-prompt-catalog-production-preflight validate-prompt-catalog-production-full-urls validate-wechat-export validate-wechat-export-acceptance validate-wechat-export-fidelity validate-wechat-export-production-preflight validate-image-workspace validate-image-workspace-e2e validate-image-workspace-e2e-object-storage validate-image-workspace-real-e2e validate-image-workspace-acceptance validate-image-workspace-backend validate-image-workspace-upstream validate-image-workspace-worker-api validate-image-workspace-object-storage validate-image-workspace-production-preflight validate-business-worker validate-home-business-capabilities validate-hot-content validate-hot-collector-preflight validate-object-storage validate-image-workspace-clean-mock secret-scan

FRONTEND_CRITICAL_VITEST := \
	src/views/auth/__tests__/LinuxDoCallbackView.spec.ts \
	src/views/auth/__tests__/WechatCallbackView.spec.ts \
	src/views/user/__tests__/PaymentView.spec.ts \
	src/views/user/__tests__/PaymentResultView.spec.ts \
	src/components/user/profile/__tests__/ProfileInfoCard.spec.ts \
	src/views/admin/__tests__/SettingsView.spec.ts

# 一键编译完整平台（Go backend + Vue frontend/admin）
build: build-backend build-frontend

# 兼容旧核心平台目标
build-core: build-backend build-frontend

# 兼容旧完整平台目标
build-all: build

# 编译后端（复用 backend/Makefile）
build-backend:
	@$(MAKE) -C backend build

# 编译前端（需要已安装依赖）
build-frontend:
	@pnpm run frontend:build

# 编译 datamanagementd（宿主机数据管理进程）
build-datamanagementd:
	@cd datamanagement && go build -o datamanagementd ./cmd/datamanagementd

# 运行完整平台测试（后端 + Vue frontend/admin）
test: test-backend test-frontend

# 兼容旧核心平台测试目标
test-core: test-backend test-frontend

test-backend:
	@$(MAKE) -C backend test

test-frontend:
	@pnpm run frontend:test

test-frontend-critical:
	@pnpm run frontend:test:critical

test-datamanagementd:
	@cd datamanagement && go test ./...

validate-prompt-catalog:
	@tools/prompt-catalog-integrity.sh

validate-prompt-catalog-parity:
	@tools/prompt-catalog-parity-audit.sh

validate-prompt-catalog-external-images:
	@tools/prompt-catalog-retire-external-images.sh

validate-prompt-catalog-production-preflight:
	@tools/prompt-catalog-production-preflight.sh

validate-prompt-catalog-production-full-urls:
	@RUN_PROMPT_CATALOG_URL_SAMPLE=1 PROMPT_CATALOG_URL_SAMPLE_LIMIT=all REQUIRE_PROMPT_CATALOG_PRODUCTION_READY=1 tools/prompt-catalog-production-preflight.sh

validate-wechat-export:
	@tools/wechat-export-smoke.sh

validate-wechat-export-acceptance:
	@node tools/wechat-export-acceptance.mjs

validate-wechat-export-fidelity:
	@npm --prefix tools/wechat-worker run fidelity-check

validate-wechat-export-production-preflight:
	@tools/wechat-export-production-preflight.sh

validate-image-workspace:
	@tools/image-workspace-smoke.sh

validate-image-workspace-e2e:
	@node tools/image-workspace-e2e.mjs

validate-image-workspace-e2e-object-storage:
	@IMAGE_WORKSPACE_E2E_OBJECT_STORAGE=1 node tools/image-workspace-e2e.mjs

validate-image-workspace-real-e2e:
	@IMAGE_WORKSPACE_E2E_REAL_PROVIDER=1 node tools/image-workspace-e2e.mjs

validate-image-workspace-acceptance:
	@node tools/image-workspace-acceptance.mjs

validate-image-workspace-backend:
	@cd backend && go test -tags unit ./internal/service ./internal/repository -run 'TestImageWorkspace' -count=1

validate-image-workspace-upstream:
	@node tools/image-workspace-upstream-mock-check.mjs

validate-image-workspace-worker-api:
	@node tools/image-workspace-worker-api-mock-check.mjs

validate-image-workspace-object-storage:
	@node tools/image-workspace-object-storage-mock-check.mjs

validate-image-workspace-production-preflight:
	@tools/image-workspace-production-preflight.sh

validate-business-worker:
	@npm --prefix tools/wechat-worker run typecheck
	@node --check tools/image-workspace-worker/src/worker.mjs
	@POSTGRES_PASSWORD="$${POSTGRES_PASSWORD:-dummy}" docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.business-worker.yml --profile business-worker config >/tmp/cloudbase-business-workers-compose.yml

validate-home-business-capabilities:
	@node tools/home-business-capabilities-smoke.mjs

validate-hot-content:
	@tools/hot-content-integrity.sh

validate-hot-collector-preflight:
	@tools/hot-collector-production-preflight.sh

validate-object-storage:
	@tools/object-storage-integrity.sh

validate-image-workspace-clean-mock:
	@tools/image-workspace-clean-mock-artifacts.sh

secret-scan:
	@python3 tools/secret_scan.py
