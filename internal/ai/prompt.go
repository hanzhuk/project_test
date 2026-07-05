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
			buildFrontendStep(entity, entityLower, entityPlural, meta),
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
	system := fmt.Sprintf(`你是 Go 后端专家。只需调用一次 write_file 工具，生成文件 %s。
规则：
- 只能 import "entgo.io/ent"、"entgo.io/ent/schema/field"，需要时可加 "time"。
- 只使用 Ent 官方 Field 方法：NotEmpty, MinLen, MaxLen, Positive, Min, Max, Default, Optional。
- 生成后立即停止，不要生成其他文件。`, targetPath)

	user := fmt.Sprintf(`生成 Ent Schema 文件：%s
实体名：%s
字段：
%s
package 为 schema，struct 嵌入 ent.Schema，实现 Fields() 和 Edges() 方法。`, targetPath, entity, fieldDesc)

	return GenerationStep{Name: "schema", TargetPath: targetPath, SystemPrompt: system, UserPrompt: user}
}

func buildHandlerStep(entity, entityLower, entityPlural, fieldDesc string, meta *metadata.ProjectMetadata) GenerationStep {
	targetPath := fmt.Sprintf("internal/handler/%s_handler.go", entityLower)
	system := fmt.Sprintf(`你是 Go 后端专家。只需调用一次 write_file 工具，生成文件 %s。
规则：
- 使用 Echo 框架，handler 函数签名为 func (h *Handler) XxxYyy(c echo.Context) error。
- 使用 Ent client（类型为 *ent.Client）操作数据库，禁止手写 SQL。
- 所有错误必须返回，使用 slog.ErrorContext 记录日志。
- import 路径使用模块路径 %s。
- 生成后立即停止，不要生成其他文件。`, targetPath, meta.ModulePath)

	user := fmt.Sprintf(`生成 Echo HTTP Handler 文件：%s
实体名：%s，路由前缀：/api/v1/%s
字段：
%s
需包含以下 5 个方法：
- Create%s：POST /api/v1/%s，从 JSON body 读取字段，调用 ent.Client 插入数据库。
- List%ss：GET /api/v1/%s，查询全部记录并返回列表。
- Get%s：GET /api/v1/%s/:id，按 ID 查询单条记录。
- Update%s：PUT /api/v1/%s/:id，更新字段。
- Delete%s：DELETE /api/v1/%s/:id，删除记录。
Handler struct 持有 *ent.Client，构造函数为 NewHandler(client *ent.Client) *Handler。
使用 response.Success(data) 和 response.Fail(code, msg) 返回统一格式。`,
		targetPath, entity, entityPlural,
		fieldDesc,
		entity, entityPlural,
		entity, entityPlural,
		entity, entityPlural,
		entity, entityPlural,
		entity, entityPlural,
	)

	return GenerationStep{Name: "handler", TargetPath: targetPath, SystemPrompt: system, UserPrompt: user}
}

func buildServiceStep(entity, entityLower, fieldDesc string, meta *metadata.ProjectMetadata) GenerationStep {
	targetPath := fmt.Sprintf("internal/service/%s_service.go", entityLower)
	system := fmt.Sprintf(`你是 Go 后端专家。只需调用一次 write_file 工具，生成文件 %s。
规则：
- 使用 Ent client（*ent.Client）操作数据库。
- 所有方法接收 context.Context 作为第一个参数。
- import 路径使用模块路径 %s。
- 生成后立即停止，不要生成其他文件。`, targetPath, meta.ModulePath)

	user := fmt.Sprintf(`生成业务逻辑层文件：%s
实体名：%s
字段：
%s
包含 Create、List、GetByID、Update、Delete 五个方法，每个方法接收 ctx context.Context 和必要参数，返回结果与 error。
Service struct 持有 *ent.Client，构造函数为 New%sService(client *ent.Client) *%sService。`,
		targetPath, entity, fieldDesc, entity, entity)

	return GenerationStep{Name: "service", TargetPath: targetPath, SystemPrompt: system, UserPrompt: user}
}

func buildFrontendStep(entity, entityLower, entityPlural string, meta *metadata.ProjectMetadata) GenerationStep {
	targetPath := fmt.Sprintf("web/src/components/%sManager.tsx", entity)
	system := fmt.Sprintf(`你是 React TypeScript 前端专家。只需调用一次 write_file 工具，生成文件 %s。
规则：
- 使用 React 19 + TypeScript + Tailwind CSS。
- 使用 fetch 调用后端接口，不引入额外依赖库。
- 生成后立即停止，不要生成其他文件。`, targetPath)

	user := fmt.Sprintf(`生成 React CRUD 组件文件：%s
实体名：%s，后端接口路径：/api/v1/%s
组件功能：
- 页面顶部显示数据列表（table 展示所有记录）。
- 列表上方有"新增"按钮，点击弹出表单。
- 每行有"编辑"和"删除"按钮。
- 表单支持新增和编辑，提交后刷新列表。
- 使用 fetch 调用 GET /api/v1/%s 获取列表，POST/PUT/DELETE 进行增删改。
export default %sManager;`, targetPath, entity, entityPlural, entityPlural, entity)

	return GenerationStep{Name: "frontend", TargetPath: targetPath, SystemPrompt: system, UserPrompt: user}
}

func buildAppTsxStep(entity string, meta *metadata.ProjectMetadata) GenerationStep {
	targetPath := "web/src/App.tsx"
	componentName := fmt.Sprintf("%sManager", entity)
	system := fmt.Sprintf(`你是 React TypeScript 前端专家。只需调用一次 write_file 工具，生成文件 %s。
规则：
- 替换原有 App.tsx 全部内容。
- 生成后立即停止，不要生成其他文件。`, targetPath)

	user := fmt.Sprintf(`更新 web/src/App.tsx，导入并渲染 %s 组件作为主页面。
内容：
import %s from './components/%s'
export default function App() { return <div className="p-4"><%%s /></div> }
（将 %%s 替换为 %s 组件标签）`, componentName, componentName, componentName, componentName)

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
