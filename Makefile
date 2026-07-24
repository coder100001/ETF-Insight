.PHONY: build test lint check dev clean install-hooks

# 开发服务器
build:
	@echo "构建前端..."
	@cd frontend && npm run build
	@echo "构建后端..."
	@cd backend && go build -o ../bin/server ./cmd/server

dev:
	@echo "启动开发服务器..."
	@docker-compose up -d
	test:
	@./scripts/run-check.sh --stage=test

lint:
	@./scripts/run-check.sh --stage=format,static

check:
	@./scripts/run-check.sh --stage=all

# 安装 pre-commit hooks
install-hooks:
	@pre-commit install
	@pre-commit autoupdate

# 清理
clean:
	@rm -rf frontend/dist
	@rm -rf bin/
	@cd backend && go clean

# 后端快捷命令
backend-test:
	@cd backend && go test -v ./...

backend-vet:
	@cd backend && go vet ./...

backend-build:
	@cd backend && go build ./...

# 前端快捷命令
frontend-test:
	@cd frontend && npm run test:run

frontend-build:
	@cd frontend && npm run build

frontend-lint:
	@cd frontend && npm run lint

# 文档一致性检查（完整模式）
# docs-check: 已移除（工具已清理）
