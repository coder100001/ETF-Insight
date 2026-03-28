# 多阶段构建 Dockerfile

# 阶段 1: 构建 Go 后端
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app/backend

# 安装依赖
RUN apk add --no-cache git

# 复制依赖文件
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制源代码
COPY backend/ .

# 构建
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/main .

# 阶段 2: 构建前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 复制 package.json
COPY frontend/package*.json ./
RUN npm ci

# 复制源代码并构建
COPY frontend/ .
RUN npm run build

# 阶段 3: 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /root/

# 从后端构建阶段复制二进制文件
COPY --from=backend-builder /app/main .

# 从前端构建阶段复制构建产物
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./main"]
