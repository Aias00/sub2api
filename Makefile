.PHONY: build build-core build-backend build-frontend build-all build-datamanagementd test test-core test-backend test-frontend test-frontend-critical test-datamanagementd secret-scan

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

secret-scan:
	@python3 tools/secret_scan.py
