FROM golang:1.21-alpine3.18 AS backend
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -o server .

FROM debian:bookworm-slim
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources 2>/dev/null; \
    sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list 2>/dev/null; \
    apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata openssh-client sshpass && \
    rm -rf /var/lib/apt/lists/*
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
WORKDIR /app
COPY --from=backend /app/server/server .
COPY --from=backend /app/server/config.yaml .
COPY web/dist ./dist
RUN mkdir -p logs
EXPOSE 8080
CMD ["./server"]