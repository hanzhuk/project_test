# Go Scaffold

**Golang AI 原生全栈应用快速开发脚手架**

Go Scaffold 是一个面向 Go 生态的 AI 原生全栈应用快速开发脚手架。通过 CLI 工具实现项目初始化、模板渲染、AI 代码生成、依赖管理和部署配置生成，帮助开发者在几分钟内获得一个可运行的全栈项目骨架。

## 核心特性

- **项目初始化（`init`）**：通过交互式 TUI 或命令行参数收集配置，生成完整的项目目录结构
- **AI 代码生成（`generate`）**：基于自然语言描述，调用本地 Ollama 模型生成 CRUD 代码、handler、模型等
- **项目构建（`build`）**：编译生成的后端项目，打包 Docker 镜像或生成可执行文件
- **多技术栈支持**：Echo/Fiber/Gin + Ent/sqlc/GORM + PostgreSQL/MySQL/SQLite + React/Vue/Svelte
- **本地化优先**：所有 AI 推理通过本地 Ollama 完成，保护代码隐私

## 技术栈

| 层级 | 技术 | 版本 |
|:---|:---|---:|
| CLI 框架 | Cobra | v1.8+ |
| TUI 框架 | Bubble Tea | v1.1+ |
| 配置管理 | Viper | v1.19+ |
| 日志 | log/slog | 标准库 |
| 后端框架 | Echo | v4.12+ |
| ORM | Ent | v0.14+ |
| AI 框架 | Genkit Go | v0.9+ |
| 本地模型 | Ollama | v0.3+ |

## 安装

### 从源码构建

```bash
git clone <repository-url>
cd go-scaffold
go build -o go-scaffold ./cmd/go-scaffold
```

### 全局安装

```bash
go install github.com/example/go-scaffold/cmd/go-scaffold@latest
```

## 快速开始

### 1. 初始化新项目

```bash
# 交互式创建（推荐）
go-scaffold init my-api

# 指定技术栈创建
go-scaffold init my-api --backend echo --orm ent --db postgres --frontend react --jwt --docker --ci
```

### 2. 进入项目并运行

```bash
cd my-api
go mod tidy
docker-compose up -d db
make migrate
make run
```

### 3. AI 代码生成

```bash
# 确保已启动 Ollama 服务
ollama serve

# 在项目目录下执行 generate
go-scaffold generate "user CRUD，包含 JWT 鉴权"
go-scaffold generate "product CRUD，字段 name price category"
```

### 4. 构建项目

```bash
# 编译二进制
go-scaffold build

# 构建 Docker 镜像
go-scaffold build --docker
```

## 命令说明

### `init` - 初始化新项目

```bash
go-scaffold init <project-name> [flags]
```

| 参数 | 类型 | 默认值 | 说明 |
|:---|:---|:---|:---|
| `project-name` | string | - | 项目名称（必填） |
| `--backend` | string | echo | 后端框架：echo/fiber/gin |
| `--orm` | string | ent | ORM：ent/sqlc/gorm |
| `--db` | string | postgres | 数据库：postgres/mysql/sqlite |
| `--frontend` | string | react | 前端框架：react/vue/svelte |
| `--jwt` | bool | false | 是否启用 JWT 认证 |
| `--docker` | bool | false | 是否生成 Docker 配置 |
| `--ci` | bool | false | 是否生成 GitHub Actions 配置 |
| `--module-prefix` | string | github.com/example | Go 模块路径前缀 |
| `--interactive` | bool | true | 是否启用交互式 TUI |

### `generate` - AI 代码生成

```bash
go-scaffold generate "<description>" [flags]
```

| 参数 | 类型 | 默认值 | 说明 |
|:---|:---|:---|:---|
| `description` | string | - | 自然语言描述 |
| `--model` | string | ornith:9b | 指定 Ollama 模型 |
| `--dry-run` | bool | false | 仅预览不写入文件 |

### `build` - 构建项目

```bash
go-scaffold build [flags]
```

| 参数 | 类型 | 默认值 | 说明 |
|:---|:---|:---|:---|
| `--docker` | bool | false | 构建 Docker 镜像 |
| `--output` | string | ./bin | 二进制输出目录 |
| `--target` | string | server | 构建目标：server/web/all |

## 配置

脚手架自身配置文件位于 `~/.go-scaffold/config.yaml`：

```yaml
default_backend: echo
default_orm: ent
default_database: postgres
default_frontend: react
ollama_host: http://localhost:11434
ollama_model: ornith:9b
log_level: info
module_prefix: github.com/example
```

也支持环境变量（前缀 `GO_SCAFFOLD_`）覆盖配置。

## AI 代码生成原理

1. 从 `.go-scaffold.json` 读取项目技术栈上下文
2. 构建 System Prompt（角色 + 框架约束）和 User Prompt（需求描述）
3. 调用 Ollama `/api/chat`，携带 `write_file` 工具定义
4. 模型通过 Tool Calling 返回 `tool_calls`（`path` + `content`）
5. 解析并执行 `go/ast` 语法校验
6. 校验通过后写入项目文件

## 项目结构

```
go-scaffold/
├── cmd/
│   └── go-scaffold/
│       └── main.go              # CLI 入口
├── internal/
│   ├── ai/                      # AI 代码生成模块
│   │   ├── ollama_client.go     # Ollama HTTP 客户端
│   │   ├── prompt.go            # Prompt 构建与解析
│   │   └── generator.go         # 代码生成器（含重试）
│   ├── build/                   # 项目构建模块
│   ├── command/                 # 命令处理（root/init/generate/build）
│   ├── config/                  # 配置管理（Viper）
│   ├── dependency/              # 依赖管理
│   ├── errors/                  # 统一错误处理
│   ├── log/                     # 日志（slog）
│   ├── metadata/                # 项目元数据读写
│   ├── project/                 # 项目生成器
│   ├── template/                # 模板引擎
│   └── tui/                     # 交互式 TUI（Bubble Tea）
├── templates/                   # 项目模板
│   ├── backend/echo/            # Echo 后端模板
│   ├── database/postgres/       # PostgreSQL 模板
│   ├── deploy/
│   │   ├── docker/              # Docker 配置模板
│   │   └── github-actions/      # CI/CD 模板
│   ├── features/jwt/            # JWT 认证模板
│   └── frontend/react/          # React 前端模板
├── go.mod
└── README.md
```

## 许可证

MIT License
