# my-new-app

my-new-app 是一个基于 Go 语言的后端服务，使用以下技术栈：

- **Web 框架**: Echo
- **ORM**: Ent
- **数据库**: PostgreSQL
- **Go 版本**: 1.23+

## 项目结构

```
my-new-app/
├── cmd/
│   └── server/
│       └── main.go          # 服务入口
├── internal/
│   ├── config/              # 配置管理
│   ├── handler/             # HTTP 请求处理器
│   ├── middleware/          # 中间件
│   ├── repository/          # 数据访问层
│   └── response/            # 统一响应格式
├── ent/
│   └── schema/              # Ent 数据模型定义
├── migrations/              # 数据库迁移
├── api/
│   └── openapi.yaml         # OpenAPI 文档
├── tests/                   # 测试
├── Dockerfile               # Docker 构建文件
├── docker-compose.yml       # Docker Compose 编排
├── Makefile                 # 构建脚本
├── go.mod                   # Go 模块
├── go.sum                   # 依赖校验
├── .env.example             # 环境变量示例
└── .gitignore               # Git 忽略配置
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 修改数据库连接信息
```

### 3. 启动数据库

```bash
docker-compose up -d db
```

### 4. 执行数据库迁移

```bash
make migrate
```

### 5. 运行服务

```bash
make run
# 或直接运行
go run ./cmd/server
```

服务默认启动在 `http://localhost:8080`。

### 6. 健康检查

```bash
curl http://localhost:8080/health
# 返回: {"status":"ok"}
```

## 常用命令

| 命令 | 说明 |
|:---|:---|
| `make run` | 运行开发服务器 |
| `make build` | 编译二进制文件 |
| `make test` | 运行单元测试 |
| `make migrate` | 执行数据库迁移 |
| `make ent-gen` | 生成 Ent 代码 |
| `make docker-build` | 构建 Docker 镜像 |
| `make docker-up` | 启动 Docker Compose 服务 |

## API 文档

API 遵循 RESTful 规范，统一响应格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 业务错误码

| 错误码 | 含义 | HTTP 状态码 |
|:---:|:---|:---:|
| 0 | 成功 | 200/201/204 |
| 1001 | 参数校验失败 | 400 |
| 1002 | 资源不存在 | 404 |
| 1003 | 资源已存在 | 409 |
| 1004 | 认证失败 | 401 |
| 1005 | 权限不足 | 403 |
| 1006 | 服务器内部错误 | 500 |
| 1007 | 数据库操作失败 | 500 |

## Docker 部署

```bash
# 构建镜像
make docker-build

# 启动全部服务（数据库 + 后端）
make docker-up

# 查看日志
docker-compose logs -f server
```

## 开发说明

- 日志使用 Go 标准库 `log/slog` 结构化日志
- 数据库操作使用 Ent ORM，禁止手写 SQL 拼接
- 错误处理统一通过 `internal/middleware` 的 `ErrorHandler` 处理

## 许可证

MIT License
