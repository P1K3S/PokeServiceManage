FROM node:18-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.21-alpine AS backend
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -o server .

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata openssh-client sshpass
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
WORKDIR /app
COPY --from=backend /app/server/server .
COPY --from=backend /app/server/config.yaml .
COPY --from=frontend /app/web/dist ./dist
RUN mkdir -p logs
EXPOSE 8080
CMD ["./server"]