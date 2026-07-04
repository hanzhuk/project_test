# Makefile - {{.ProjectName}} 项目构建脚本

# Go 相关变量
GO := go
GOFMT := gofmt
BINARY := server
OUTPUT_DIR := bin
MAIN_PKG := ./cmd/server

# 构建参数
LDFLAGS := -ldflags "-s -w"

# 默认目标
.PHONY: all
all: build

# 运行开发服务器
.PHONY: run
run:
	$(GO) run $(MAIN_PKG)

# 编译二进制
.PHONY: build
build:
	$(GO) build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY) $(MAIN_PKG)

# 运行单元测试
.PHONY: test
test:
	$(GO) test ./... -v

# 执行数据库迁移
.PHONY: migrate
migrate:
	$(GO) run $(MAIN_PKG) --migrate

# 生成 Ent 代码
.PHONY: ent-gen
ent-gen:
	$(GO) run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema

# 构建 Docker 镜像
.PHONY: docker-build
docker-build:
	docker build -t {{.ProjectName}}:latest .

# 启动 Docker Compose 服务
.PHONY: docker-up
docker-up:
	docker-compose up -d

# 停止 Docker Compose 服务
.PHONY: docker-down
docker-down:
	docker-compose down

# 代码格式化
.PHONY: fmt
fmt:
	$(GOFMT) -w .

# 代码检查
.PHONY: lint
lint:
	$(GO) vet ./...

# 清理构建产物
.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)

# 生成 OpenAPI 文档
.PHONY: openapi
openapi:
	@echo "请使用 Huma 或手动更新 api/openapi.yaml"
