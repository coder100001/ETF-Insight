# ETF-Insight Makefile

.PHONY: help build run dev init-db clean test docker-build docker-run

# 默认目标
help:
	@echo "ETF-Insight 项目管理"
	@echo ""
	@echo "可用命令:"
	@echo "  make build       - 构建 Go 后端"
	@echo "  make run         - 运行 Go 后端"
	@echo "  make dev         - 开发模式运行（自动重载）"
	@echo "  make init-db     - 初始化数据库"
	@echo "  make clean       - 清理构建文件"
	@echo "  make test        - 运行测试"
	@echo "  make docker-build - 构建 Docker 镜像"
	@echo "  make docker-run  - 运行 Docker 容器"
	@echo "  make frontend    - 构建前端"
	@echo "  make frontend-dev - 开发模式运行前端"

# 构建 Go 后端
build:
	cd backend && go build -o ../bin/etf-insight .

# 运行 Go 后端
run: build
	./bin/etf-insight

# 开发模式（需要安装 air: go install github.com/cosmtrek/air@latest）
dev:
	cd backend && air

# 初始化数据库
init-db:
	cd backend && go run main.go -init-db

# 清理构建文件
clean:
	rm -rf bin/
	rm -f backend/etf_insight.db

# 运行测试
test:
	cd backend && go test -v ./...

# 构建 Docker 镜像
docker-build:
	docker build -t etf-insight:latest .

# 运行 Docker 容器
docker-run:
	docker run -p 8080:8080 -v $(PWD)/data:/data etf-insight:latest

# 构建前端
frontend:
	cd frontend && npm install && npm run build

# 开发模式运行前端
frontend-dev:
	cd frontend && npm run dev

# 安装依赖
deps:
	cd backend && go mod tidy
	cd frontend && npm install

# 格式化代码
fmt:
	cd backend && gofmt -w .
	cd frontend && npx prettier --write "src/**/*.{ts,tsx}"

# 代码检查
lint:
	cd backend && golangci-lint run
	cd frontend && npm run lint
