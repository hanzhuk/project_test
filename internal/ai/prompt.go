package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/example/go-scaffold/internal/ast"
	"github.com/example/go-scaffold/internal/metadata"
)

// GenerateInput 定义 AI 代码生成输入。
// 技术栈（Backend/ORM/Database）不应由用户每次生成时重复指定，
// 而应从当前项目的 .go-scaffold.json 元数据中读取并作为上下文注入 Prompt。
type GenerateInput struct {
	Description string  // 用户自然语言描述
	CodeType    string  // 生成代码类型：auto/handler/model/service/route/test
	Entity      string  // 实体名称，如 User
	Fields      []Field // 实体字段列表
}

// Field 定义实体的单个字段。
type Field struct {
	Name string // 字段名称
	Type string // 字段类型简写：int/string/bool/time/float
	Tag  string // 可选的 struct tag
}

// WriteFileArgs 是 write_file 工具的参数结构。
// 对应 Ollama tools 定义中 write_file function 的 parameters schema。
type WriteFileArgs struct {
	Path    string `json:"path"`    // 文件相对路径，如 internal/handler/user_handler.go
	Content string `json:"content"` // 完整的 Go 源代码文件内容
}

// CodeFile 定义单个生成文件。
type CodeFile struct {
	Path    string // 文件相对路径
	Content string // 完整的文件内容
}

// GeneratedCode 定义 AI 代码生成输出。
type GeneratedCode struct {
	Files []CodeFile // 生成的代码文件列表
}

// RetryPolicy 定义 AI 生成失败时的重试策略。
type RetryPolicy struct {
	MaxRetries    int           // 最大重试次数，默认 3
	Backoff       time.Duration // 初始退避间隔
	BackoffFactor float64       // 退避倍数，默认 2.0
}

// DefaultRetryPolicy 返回默认重试策略：3 次、初始 2 秒、倍数 2.0。
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		Backoff:       2 * time.Second,
		BackoffFactor: 2.0,
	}
}

// WriteFileToolDefinition 是 write_file 工具的 JSON Schema 定义。
// 作为 Ollama tools 参数传入，约束模型输出结构化文件写入指令。
var WriteFileToolDefinition = Tool{
	Type: "function",
	Function: ToolFunction{
		Name:        "write_file",
		Description: "创建一个新的 Go 源代码文件。path 为相对于项目根目录的路径，content 为完整可编译的源代码，须包含 package 声明和所有 import。",
		Parameters: json.RawMessage(`{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "文件相对路径，如 ent/schema/user.go"
    },
    "content": {
      "type": "string",
      "description": "完整的 Go 源代码，包含 package 声明、import、错误处理和 slog 日志"
    }
  },
  "required": ["path", "content"]
}`),
	},
}

// buildSystemPrompt 构建 System Prompt。
// 根据 ProjectMetadata 中的技术栈注入框架使用约束、工具使用指令和输出规范。
func buildSystemPrompt(meta *metadata.ProjectMetadata) string {
	prompt := fmt.Sprintf(`你是一名资深的 Go 后端开发专家。请根据用户需求生成高质量的 Go 代码。

项目技术栈：
- Web 框架：%s
- ORM：%s
- 数据库：%s
- 模块路径：%s

你必须为每个实体生成以下全部文件，缺少任何一个均视为不完整，必须全部调用 write_file：
【必须生成的文件清单】
  文件1: ent/schema/{实体小写}.go         —— Ent Schema 定义
  文件2: internal/handler/{实体小写}_handler.go —— Echo HTTP Handler，包含 Create/List/Get/Update/Delete 五个方法
  文件3: internal/service/{实体小写}_service.go  —— 业务逻辑层，调用 Ent Client 操作数据库
  文件4: internal/repository/{实体小写}_repo.go  —— 数据访问层接口与实现

生成规则：
1. 必须按顺序依次调用 write_file，将上述4个文件全部输出后才能停止。
2. path 为相对于项目根目录的路径。
3. content 为完整的 Go 源代码，包含 package 声明和所有必要的 import。
4. 不要使用 Markdown 代码围栏或额外解释文字，所有内容通过 write_file 工具输出。
5. 所有函数必须处理错误并返回 error，禁止忽略错误。
6. 使用 slog.ErrorContext / slog.InfoContext 记录结构化日志，禁止使用 fmt.Println 或 log.Printf。
7. 使用 %s 框架的上下文处理方式注册路由。
8. 使用 %s 进行数据库操作，禁止手写 SQL 字符串拼接。
9. 字段类型使用 Go 标准类型，时间字段统一使用 time.Time。
10. 生成的代码必须可直接编译，import 路径使用项目模块路径 %s。
11. 若生成 ent/schema/*.go，只能 import "entgo.io/ent"、"entgo.io/ent/schema/field" 和 (必要时) "time"，严禁引用 entgql、entql、entsql 或不存在的包。
12. 生成 ent/schema/*.go 时必须严格遵守 Ent 官方标准 Field 方法（如 NotEmpty(), MinLen(), MaxLen(), Positive(), Min(), Max(), Default(), Optional()），严禁虚构方法（如 Ceil, Negative, Round, Table 等）。
范例格式：
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Book struct {
	ent.Schema
}

func (Book) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty(),
		field.String("author").NotEmpty(),
		field.Float("price").Min(0).Optional(),
	}
}

func (Book) Edges() []ent.Edge {
	return nil
}
`,
		backendName(meta.Backend),
		ormName(meta.ORM),
		databaseName(meta.Database),
		meta.ModulePath,
		backendName(meta.Backend),
		ormName(meta.ORM),
		meta.ModulePath,
	)

	// 如果项目启用了前端，追加前端文件到必须生成清单
	if meta.Frontend != "" && meta.Frontend != "none" {
		prompt += fmt.Sprintf(`
【前端文件（同样必须生成，不可省略）】
  文件5: web/src/components/{实体}Manager.tsx —— React CRUD 组件，包含数据表格与新增/编辑表单
  文件6: web/src/App.tsx                      —— 必须更新，导入并渲染上面的组件

前端规则：
13. 项目启用了 %s 前端，必须额外调用 write_file 生成文件5和文件6，否则任务不完整。
14. {实体}Manager.tsx 使用 fetch 调用后端 /api/v1/{实体复数小写} 接口，展示列表并支持增删改。
15. App.tsx 必须替换原有内容，直接渲染该 CRUD 组件作为主页面。
`, meta.Frontend)
	}

	return prompt
}

// buildUserPrompt 构建 User Prompt。
// 包含用户输入的自然语言描述、实体名称、字段列表和代码类型。
func buildUserPrompt(input *GenerateInput) string {
	prompt := fmt.Sprintf("需求描述：%s\n", input.Description)
	// 追加实体信息
	if input.Entity != "" {
		prompt += fmt.Sprintf("实体名称：%s\n", input.Entity)
	}
	// 追加字段列表
	if len(input.Fields) > 0 {
		prompt += "字段列表：\n"
		for _, f := range input.Fields {
			prompt += fmt.Sprintf("- %s: %s\n", f.Name, f.Type)
		}
	}
	// 追加代码类型
	if input.CodeType != "" && input.CodeType != "auto" {
		prompt += fmt.Sprintf("生成代码类型：%s\n", input.CodeType)
	}
	prompt += "\n请使用 write_file 工具生成所有相关文件（包括 schema、handler、service、route 注册等）。"
	return prompt
}

// buildChatMessages 构建 /api/chat 的 messages 数组。
// 包含一条 system 消息（角色设定+技术栈约束）和一条 user 消息（需求描述+实体字段）。
func buildChatMessages(input *GenerateInput, meta *metadata.ProjectMetadata) []ChatMessage {
	return []ChatMessage{
		{
			Role:    "system",
			Content: buildSystemPrompt(meta),
		},
		{
			Role:    "user",
			Content: buildUserPrompt(input),
		},
	}
}

// parseToolCalls 将模型返回的 tool_calls 解析为 GeneratedCode。
// 校验每个 write_file 调用的 path 和 content 均为非空字符串，且 path 不包含 ".."。
func parseToolCalls(toolCalls []ToolCall) (*GeneratedCode, error) {
	if len(toolCalls) == 0 {
		return nil, fmt.Errorf("模型未返回任何 tool_calls，未能生成代码文件")
	}
	files := make([]CodeFile, 0, len(toolCalls))
	for i, tc := range toolCalls {
		// 仅处理 write_file 工具调用
		if tc.Function.Name != "write_file" {
			continue
		}
		// 反序列化参数 JSON（兼容 JSON 对象与转义字符串）
		var args WriteFileArgs
		raw := tc.Function.Arguments
		if err := json.Unmarshal(raw, &args); err != nil {
			// 尝试二次反序列化（防止模型返回被双重转义的字符串）
			var strArgs string
			if errStr := json.Unmarshal(raw, &strArgs); errStr == nil {
				if errInner := json.Unmarshal([]byte(strArgs), &args); errInner != nil {
					return nil, fmt.Errorf("第 %d 个 tool_call 参数解析失败: %w", i+1, errInner)
				}
			} else {
				return nil, fmt.Errorf("第 %d 个 tool_call 参数解析失败: %w", i+1, err)
			}
		}
		// 校验 path 非空
		if args.Path == "" {
			return nil, fmt.Errorf("第 %d 个 tool_call 的 path 为空", i+1)
		}
		// 校验 path 不包含父目录引用，防止路径穿越
		if containsParentDir(args.Path) {
			return nil, fmt.Errorf("第 %d 个 tool_call 的 path 包含非法路径 '..': %s", i+1, args.Path)
		}
		// 校验 content 非空
		if args.Content == "" {
			return nil, fmt.Errorf("第 %d 个 tool_call 的 content 为空 (path: %s)", i+1, args.Path)
		}
		// 如果生成的是 .go 源文件，执行 Go AST 语法自动校验
		if strings.HasSuffix(args.Path, ".go") {
			if err := ast.ValidateGoSource([]byte(args.Content)); err != nil {
				return nil, fmt.Errorf("第 %d 个生成文件 (%s) AST 语法校验失败: %w", i+1, args.Path, err)
			}
		}
		files = append(files, CodeFile{
			Path:    args.Path,
			Content: args.Content,
		})
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("模型返回的 tool_calls 中没有有效的 write_file 调用")
	}
	return &GeneratedCode{Files: files}, nil
}

// containsParentDir 检查路径是否包含父目录引用 ".."。
func containsParentDir(path string) bool {
	return containsStr(path, "..")
}

// containsStr 检查字符串是否包含子串。
func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// backendName 返回后端框架的友好显示名称。
func backendName(backend string) string {
	switch backend {
	case "echo":
		return "Echo"
	case "gin":
		return "Gin"
	case "fiber":
		return "Fiber"
	default:
		return backend
	}
}

// ormName 返回 ORM 的友好显示名称。
func ormName(orm string) string {
	switch orm {
	case "ent":
		return "Ent"
	case "gorm":
		return "GORM"
	case "sqlc":
		return "sqlc"
	default:
		return orm
	}
}

// GenerationStep 定义单个文件的生成步骤。
type GenerationStep struct {
	Name         string // 步骤名称，用于日志
	TargetPath   string // 期望生成的文件路径
	SystemPrompt string // 专注于单文件的 system prompt
	UserPrompt   string // 专注于单文件的 user prompt
}

// BuildSteps 根据输入和项目元数据构建分步生成计划。
// 每步只生成一个文件，适合 Tool Calling 能力较弱的小模型。
func BuildSteps(input *GenerateInput, meta *metadata.ProjectMetadata) []GenerationStep {
	entity := input.Entity
	if entity == "" {
		entity = "Entity"
	}
	entityLower := strings.ToLower(entity)
	entityPlural := entityLower + "s"

	fieldDesc := buildFieldDesc(input.Fields)

	steps := []GenerationStep{
		buildSchemaStep(entity, entityLower, fieldDesc, meta),
		buildHandlerStep(entity, entityLower, entityPlural, fieldDesc, meta),
		buildServiceStep(entity, entityLower, fieldDesc, meta),
	}

	if meta.Frontend != "" && meta.Frontend != "none" {
		steps = append(steps,
			buildFrontendStep(entity, entityLower, entityPlural, input.Fields, meta),
			buildAppTsxStep(entity, meta),
		)
	}

	return steps
}

func buildFieldDesc(fields []Field) string {
	if len(fields) == 0 {
		return "（字段由模型根据实体语义自行决定）"
	}
	var sb strings.Builder
	for _, f := range fields {
		sb.WriteString(fmt.Sprintf("  - %s: %s\n", f.Name, f.Type))
	}
	return sb.String()
}

func buildSchemaStep(entity, entityLower, fieldDesc string, meta *metadata.ProjectMetadata) GenerationStep {
	targetPath := fmt.Sprintf("ent/schema/%s.go", entityLower)
	system := fmt.Sprintf(`你是 Go 后端专家。只需调用一次 write_file 工具，生成文件 %s，生成后立即停止。

【Ent 字段类型映射，严格遵守，禁止使用其他名称】
  string  → field.String("name")
  int     → field.Int("name")
  float   → field.Float("name")   ← 注意：是 Float 不是 Float64
  bool    → field.Bool("name")
  time    → field.Time("name")

【禁止使用】field.Float64、field.Int64、field.Integer 等不存在的方法。
【允许的修饰方法】NotEmpty(), MinLen(n), MaxLen(n), Positive(), Min(n), Max(n), Default(v), Optional()
【只允许的 import】"entgo.io/ent"、"entgo.io/ent/schema/field"，需要时加 "time"`, targetPath)

	user := fmt.Sprintf(`生成文件 %s
实体名：%s
字段：
%s
示例格式（严格照此结构）：
package schema
import ("entgo.io/ent"; "entgo.io/ent/schema/field")
type %s struct { ent.Schema }
func (%s) Fields() []ent.Field { return []ent.Field{ field.String("title").NotEmpty() } }
func (%s) Edges() []ent.Edge { return nil }`, targetPath, entity, fieldDesc, entity, entity, entity)

	return GenerationStep{Name: "schema", TargetPath: targetPath, SystemPrompt: system, UserPrompt: user}
}

func buildHandlerStep(entity, entityLower, entityPlural, fieldDesc string, meta *metadata.ProjectMetadata) GenerationStep {
	targetPath := fmt.Sprintf("internal/handler/%s_handler.go", entityLower)
	system := fmt.Sprintf(`你是 Go 后端专家。只需调用一次 write_file 工具，生成文件 %s，生成后立即停止。

【绝对禁止——违反则代码无法编译】
❌ 禁止使用 ent.%sCreateInput、ent.%sUpdateInput、ent.%sMutation——这些类型在 Ent 中不存在
❌ 禁止忽略错误：_, _ = xxx 或 _, err = xxx 后不检查 err
❌ 禁止只返回 id 和 title，必须直接返回 ent 查询结果对象

【必须这样写 request struct（不能用 ent 的类型）】
type %sRequest struct {
  Title  string  `+"`"+`json:"title"`+"`"+`
  Author string  `+"`"+`json:"author"`+"`"+`
  Price  float64 `+"`"+`json:"price"`+"`"+`
}

【必须这样调用 Ent（严格照此写法，不能自创方法名）】
  创建：h.Client.%s.Create().SetTitle(req.Title).SetAuthor(req.Author).SetPrice(req.Price).Save(ctx)
  列表：h.Client.%s.Query().All(ctx)
  按ID：h.Client.%s.Get(ctx, id)      // id 是 int，用 strconv.Atoi 解析
  更新：h.Client.%s.UpdateOneID(id).SetTitle(req.Title).SetPrice(req.Price).Save(ctx)
  删除：h.Client.%s.DeleteOneID(id).Exec(ctx)

【响应】
  成功：return c.JSON(http.StatusOK, response.Success(result))   // result 是 ent 返回的对象
  错误：return c.JSON(http.StatusBadRequest, response.Error(400, "msg"))

【import】必须包含："%s/ent", "%s/internal/response", "github.com/labstack/echo/v4", "log/slog", "net/http", "strconv"`,
		targetPath,
		entity, entity, entity,
		entity,
		entity, entity, entity, entity, entity,
		meta.ModulePath, meta.ModulePath)

	user := fmt.Sprintf(`生成文件 %s
实体名：%s，路由前缀 /api/v1/%s
字段列表（根据字段生成 %sRequest struct 和对应的 SetXxx 调用）：
%s
5个方法：Create%s、List%ss、Get%s、Update%s、Delete%s
所有方法必须挂载在已有的 *Handler 上，Handler 已持有 h.Client *ent.Client。`,
		targetPath, entity, entityPlural, entity, fieldDesc,
		entity, entity, entity, entity, entity)

	return GenerationStep{Name: "handler", TargetPath: targetPath, SystemPrompt: system, UserPrompt: user}
}

func buildServiceStep(entity, entityLower, fieldDesc string, meta *metadata.ProjectMetadata) GenerationStep {
	targetPath := fmt.Sprintf("internal/service/%s_service.go", entityLower)
	system := fmt.Sprintf(`你是 Go 后端专家。只需调用一次 write_file 工具，生成文件 %s，生成后立即停止。

【绝对禁止】
❌ 禁止使用 book.Entity、ent.%sEntity——返回类型必须是 *ent.%s 或 []*ent.%s
❌ 禁止 import "%s/ent/%s"（除非用到排序常量，否则不要引入）
❌ 禁止忽略错误

【必须这样写】
- import 只需要："context" 和 "%s/ent"
- Service struct：type %sService struct { client *ent.Client }
- 返回类型：*ent.%s 或 []*ent.%s（不是 book.Entity）
- Ent API：s.client.%s.Create().SetTitle(t).Save(ctx)`,
		targetPath,
		entity, entity, entity,
		meta.ModulePath, entityLower,
		meta.ModulePath,
		entity, entity, entity, entity)

	user := fmt.Sprintf(`生成文件 %s
实体名：%s，字段：
%s
构造函数 New%sService(client *ent.Client) *%sService。
包含 Create、List、GetByID、Update、Delete 五个方法。`,
		targetPath, entity, fieldDesc, entity, entity)

	return GenerationStep{Name: "service", TargetPath: targetPath, SystemPrompt: system, UserPrompt: user}
}

func buildFrontendStep(entity, entityLower, entityPlural string, fields []Field, meta *metadata.ProjectMetadata) GenerationStep {
	targetPath := fmt.Sprintf("web/src/components/%sManager.tsx", entity)
	system := fmt.Sprintf(`你是 React TypeScript 前端专家。只需调用一次 write_file 工具，生成文件 %s，生成后立即停止。

【绝对禁止】
❌ 禁止写 import React from 'react' 或 React.FormEvent、React.ChangeEvent
❌ 禁止用动态 key 访问字段：item[header as keyof Item]
❌ 禁止直接 setItems(await resp.json())，后端响应是 {code, message, data} 包装格式
❌ 禁止在保存按钮上设置 disabled={editingId === -1}（保存按钮绝不能在新增时被禁用）

【第一行 import 必须完整列出所有用到的 hook】
import { useState, useEffect, FormEvent } from 'react'

【后端响应解包——必须这样写】
fetch('/api/v1/%s').then(r=>r.json()).then(json=>{ if(json.code===0) setItems(json.data) })

【新增/编辑 UI 模式——必须严格照此实现】
- 用 editingId: number|null 控制表单显示，null=隐藏，-1=新增，>0=编辑
- "新增"按钮：onClick={() => { setEditingId(-1); resetForm() }}
- 点击编辑时，必须将当前行的实体数据加载到表单状态中（不要调用 resetForm() 清空）
- 表单显示条件：{editingId !== null && <form>...</form>}
- 提交时判断：if(editingId === -1) 调用 POST，else 调用 PUT /{editingId}
- "取消"按钮：onClick={() => setEditingId(null)}
- 表单保存按钮：必须保持可用状态，不能禁用

【表格渲染——必须用明确字段名】
正确：<td>{item.title}</td><td>{item.author}</td>
错误：{headers.map(h => <td>{item[h]}</td>)}

【其他规则】使用 Tailwind CSS，使用 fetch 不用 axios。`,
		targetPath, entityPlural)

	fieldDesc := buildFieldDesc(fields)
	user := fmt.Sprintf(`生成文件 %s
实体名：%s，后端接口基路径：/api/v1/%s
实体包含以下字段，请在表格和表单中只使用这些字段（不要虚构其他字段）：
%s
功能：搜索过滤、新增、编辑、删除的完整 CRUD 界面。
export default function %sManager()`, targetPath, entity, entityPlural, fieldDesc, entity)

	return GenerationStep{Name: "frontend", TargetPath: targetPath, SystemPrompt: system, UserPrompt: user}
}

func buildAppTsxStep(entity string, meta *metadata.ProjectMetadata) GenerationStep {
	targetPath := "web/src/App.tsx"
	componentName := fmt.Sprintf("%sManager", entity)
	system := `你是 React TypeScript 前端专家。只需调用一次 write_file 工具，生成文件 web/src/App.tsx，生成后立即停止。
替换文件全部内容，只写导入和默认导出，不要添加其他逻辑。`

	user := fmt.Sprintf(`生成 web/src/App.tsx，内容如下（完整输出，不要省略）：
import %s from './components/%s'
export default function App() {
  return (
    <div>
      <%s />
    </div>
  )
}`, componentName, componentName, componentName)

	return GenerationStep{Name: "app-tsx", TargetPath: targetPath, SystemPrompt: system, UserPrompt: user}
}

// databaseName 返回数据库的友好显示名称。
func databaseName(db string) string {
	switch db {
	case "postgres":
		return "PostgreSQL"
	case "mysql":
		return "MySQL"
	case "sqlite":
		return "SQLite"
	default:
		return db
	}
}
