# 使用官方 Go 1.24.4 镜像作为构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建 Go 应用，禁用 CGO，确保静态链接
RUN CGO_ENABLED=0 GOOS=linux go build -o ipquery .
# RUN CGO_ENABLED=0 GOOS=dawin GOARCH=arm64 go build -o ipquery .

# 使用轻量级 alpine 镜像作为运行阶段
FROM alpine:latest

# 安装必要的证书（如果应用需要 HTTPS 或其他 TLS 连接）
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /app/

# 从构建阶段复制可执行文件
COPY --from=builder /app/ipquery /app/cities.json /app/countries.json /app/regions.json ./

# 暴露应用端口（根据你的应用修改端口号）
EXPOSE 8080

# 运行应用
CMD ["./ipquery"]