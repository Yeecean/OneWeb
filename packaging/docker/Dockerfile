# OneWeb 控制面多阶段构建
FROM golang:1.22-alpine AS backend
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/oneweb ./cmd/oneweb

# 运行时镜像（前端已 embed 进二进制）
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /out/oneweb /usr/local/bin/oneweb
EXPOSE 8080
ENTRYPOINT ["oneweb", "serve"]
