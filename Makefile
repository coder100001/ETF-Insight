# ETF-Insight Makefile
# 统一开发、测试、构建、部署流程

.PHONY: help dev build test clean docker-run docker-build lint backend-frontend restart

# 默认目标
help:
	@echo "ETF-Insight 开发命令："
	@echo ""
	@echo "  make dev          - 启动开发环境（前后端）"
	@echo "  make backend      - 仅启动后端服务"
	@echo "  make frontend     - 仅启动前端开发服务器"
	@echo "  make build        - 构建生产版本"
	@echo "  make test         - 运行所有测试"
	@echo "  make lint         - 运行代码检查"
	@echo "  make clean        - 清理构建产物"
	@echo "  make docker-build - 构建 Docker 镜像"
	@echo "  make docker-run   - 运行 Docker 容器"
	@echo "  make restart      - 重启服务"
	@echo ""

# 开发环境
dev:
	@echo "启动开发环境..."
	@# 启动后端
	@cd backend && go run main.go &
	@# 启动前端
	@cd frontend && npm run dev
	@echo "开发环境已启动！"
	@echo "  - 前端：http://localhost:5173"
	@echo "  - 后端：http://localhost:8080"

# 仅启动后端
backend:
	@echo "启动后端服务..."
	@cd backend && go run main.go

# 仅启动前端
frontend:
	@echo "启动前端开发服务器..."
	@cd frontend && npm run dev

# 构建生产版本
build:
	@echo "构建生产版本..."
	@# 构建前端
	@cd frontend && npm run build
	@# 构建后端
	@cd backend && CGO_ENABLED=1 go build -ldflags="-s -w" -o ../dist/main .
	@# 复制前端构建产物
	@cp -r frontend/dist ../dist/frontend/
	@echo "构建完成！输出目录：./dist/"

# 运行测试
test:
	@echo "运行后端测试..."
	@cd backend && go test -v ./...
	@echo "运行前端测试..."
	@cd frontend && npm run test

# 运行代码检查
lint:
	@echo "运行后端代码检查..."
	@cd backend && gofmt -l . && go vet ./...
	@echo "运行前端代码检查..."
	@cd frontend && npm run lint

# 清理构建产物
clean:
	@echo "清理构建产物..."
	@rm -rf dist/
	@rm -rf frontend/dist/
	@rm -rf frontend/node_modules/
	@cd backend && go clean -cache
	@echo "清理完成！"

# 构建 Docker 镜像
docker-build:
	@echo "构建 Docker 镜像..."
	@docker build -t etf-insight:latest .
	@echo "Docker 镜像构建完成！"
	@docker images etf-insight

# 运行 Docker 容器
docker-run:
	@echo "运行 Docker 容器..."
	@docker run -d -p 8080:8080 --name etf-insight etf-insight:latest
	@echo "容器已启动！访问 http://localhost:8080"

# 停止 Docker 容器
docker-stop:
	@docker stop etf-insight || true
	@docker rm etf-insight || true

# 重启服务
restart:
	@echo "重启服务..."
	@pkill -f "go run main.go" || true
	@pkill -f "npm run dev" || true
	@sleep 2
	@make dev

# 数据库迁移（如果有）
db-migrate:
	@echo "运行数据库迁移..."
	@# 添加数据库迁移命令

# 查看日志
logs:
	@docker logs -f etf-insight

# 进入容器
shell:
	@docker exec -it etf-insight /bin/sh
