# Dockerfile - demo-api 多阶段构建
# 阶段一：编译
FROM golang:1.23-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum，利用缓存加速依赖下载
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译二进制文件，使用 -ldflags "-s -w" 减小体积
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o server ./cmd/server

# 阶段二：运行
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/server .

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./server"]
