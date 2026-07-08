FROM golang:1.24-alpine AS builder

# 安装git
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
# RUN go mod download
RUN go mod tidy
# 复制全部源码
COPY . .

# 编译为静态二进制（禁用 CGO 以便在 alpine/scratch 中运行）,docker的架构一般是arm64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o myapp .

# ------------------- 运行阶段 -------------------
FROM alpine:latest

# 安装 ca-certificates（如果应用需要 HTTPS 请求，否则可省略）
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 从构建阶段拷贝二进制
COPY --from=builder /app/myapp .

# 拷贝配置文件和数据文件
RUN mkdir -p /app/configs
COPY configs/config.yaml configs/
COPY second-hand-mall.db .

# 暴露应用端口（根据你的实际端口修改，例如 8080）
EXPOSE 8080

# 启动应用
CMD ["./myapp", "web", "-config", "configs/config.yaml"]