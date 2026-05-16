# 第一阶段：构建前端 (Vue 3 + Vite)
FROM node:18-alpine AS frontend

WORKDIR /app/web

# 配置 npm 国内镜像源（可选，加速构建）
RUN npm config set registry https://registry.npmmirror.com

# 复制前端依赖文件
COPY web/package*.json ./
RUN npm ci --only=production || npm install

# 复制前端源码并构建
COPY web/ ./

# Vite 构建，确保输出到 dist 目录
RUN npm run build

# 第二阶段：构建后端 (Go 1.21+)
FROM golang:1.21-alpine3.18 AS backend

# 设置 Go 代理
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app/server

# 复制 Go 模块文件并下载依赖
COPY server/go.mod server/go.sum ./
RUN go mod download

# 复制后端源码并构建
COPY server/ ./

# 构建可执行文件
RUN go build -ldflags="-s -w" -o server .

# 第三阶段：最终运行镜像
FROM debian:bookworm-slim

# 配置 Debian 国内源并安装运行时依赖
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
    sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list 2>/dev/null; \
    apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    openssh-client \
    sshpass \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app

# 从后端构建阶段复制可执行文件
COPY --from=backend /app/server/server .

# 复制配置文件（如果存在）
COPY --from=backend /app/server/config.yaml ./config.yaml

# 从前端构建阶段复制静态文件
COPY --from=frontend /app/web/dist ./dist

# 创建必要目录
RUN mkdir -p logs uploads

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# 运行应用
CMD ["./server"]
