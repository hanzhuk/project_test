package ai

import (
	"encoding/json"
	"fmt"
	"time"

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
	return fmt.Sprintf(`你是一名资深的 Go 后端开发专家。请根据用户需求生成高质量的 Go 代码。

项目技术栈：
- Web 框架：%s
- ORM：%s
- 数据库：%s
- 模块路径：%s

生成要求：
1. 使用 write_file 工具为每个新文件调用一次，一次生成所有相关文件。
2. path 为相对于项目根目录的路径，如 ent/schema/user.go、internal/handler/user_handler.go。
3. content 为完整的 Go 源代码，包含 package 声明和所有必要的 import。
4. 不要使用 Markdown 代码围栏或额外解释文字，所有内容通过 write_file 工具输出。
5. 所有函数必须处理错误并返回 error，禁止忽略错误。
6. 使用 slog.ErrorContext / slog.InfoContext 记录结构化日志，禁止使用 fmt.Println 或 log.Printf。
7. 使用 %s 框架的上下文处理方式注册路由。
8. 使用 %s 进行数据库操作，禁止手写 SQL 字符串拼接。
9. 字段类型使用 Go 标准类型，时间字段统一使用 time.Time。
10. 生成的代码必须可直接编译，import 路径使用项目模块路径 %s。`,
		backendName(meta.Backend),
		ormName(meta.ORM),
		databaseName(meta.Database),
		meta.ModulePath,
		backendName(meta.Backend),
		ormName(meta.ORM),
		meta.ModulePath,
	)
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
		// 反序列化参数 JSON 字符串
		var args WriteFileArgs
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return nil, fmt.Errorf("第 %d 个 tool_call 参数解析失败: %w", i+1, err)
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
