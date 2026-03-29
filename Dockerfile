# ETF-Insight 多阶段构建 Dockerfile
# 优化目标：减小镜像体积、提升安全性、添加健康检查

# ============================================
# 阶段 1: 构建 Go 后端
# ============================================
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app/backend

# 安装必要的构建依赖
RUN apk add --no-cache git gcc musl-dev

# 复制依赖文件（利用 Docker 缓存层）
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制源代码
COPY backend/ .

# 构建优化的二进制文件（静态链接，去除调试信息）
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o /app/main .

# ============================================
# 阶段 2: 构建前端
# ============================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制 package 文件（利用 Docker 缓存层）
COPY frontend/package*.json ./

# 安装依赖（使用 npm ci 确保一致性）
RUN npm ci --omit=dev

# 复制源代码并构建
COPY frontend/ .
RUN npm run build

# ============================================
# 阶段 3: 最终运行镜像
# ============================================
FROM alpine:3.19

# 安装运行时依赖（sqlite 和 CA 证书）
RUN apk --no-cache add ca-certificates sqlite-libs tzdata

# 创建非 root 用户（安全最佳实践）
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -D appuser

WORKDIR /root/

# 从后端构建阶段复制二进制文件
COPY --from=backend-builder /app/main .

# 从前端构建阶段复制构建产物
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# 设置正确的权限
RUN chown -R appuser:appgroup /root/
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查（每 30 秒检查一次）
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# 设置环境变量
ENV GIN_MODE=release

# 运行
CMD ["./main"]
