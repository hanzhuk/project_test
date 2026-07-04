package ai

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"strings"
	"time"

	"github.com/example/go-scaffold/internal/errors"
	"github.com/example/go-scaffold/internal/log"
	"github.com/example/go-scaffold/internal/metadata"
)

// CodeGenerator 是 AI 代码生成器，负责调用 Ollama 生成代码并校验。
// client 为 Ollama 客户端，model 为使用的模型名称。
type CodeGenerator struct {
	client *OllamaClient // Ollama 客户端
	model  string        // 模型名称
}

// NewCodeGenerator 创建代码生成器实例。
// client 为 Ollama 客户端，model 为模型名称（如 ornith:9b）。
func NewCodeGenerator(client *OllamaClient, model string) *CodeGenerator {
	return &CodeGenerator{
		client: client,
		model:  model,
	}
}

// Generate 根据输入和项目元数据生成代码。
//
// 采用 Tool Calling 单次生成协议：
//  1. 从 meta 读取项目技术栈（Backend/ORM/Database/ModulePath）作为上下文；
//  2. 构建 System Prompt + User Prompt，通过 /api/chat 发送，携带 write_file 工具定义；
//  3. 模型返回 tool_calls，每个 tool_call 为一次 write_file 调用；
//  4. 将 tool_calls 参数反序列化为 WriteFileArgs，组装为 GeneratedCode；
//  5. 对每个 file.content 执行 go/parser 语法校验。
func (g *CodeGenerator) Generate(ctx context.Context, input *GenerateInput, meta *metadata.ProjectMetadata) (*GeneratedCode, error) {
	// 1. 构建 messages
	messages := buildChatMessages(input, meta)
	log.Info("开始 AI 代码生成",
		"model", g.model,
		"entity", input.Entity,
		"code_type", input.CodeType,
	)

	// 2. 构建 chat 请求，携带 write_file 工具定义
	req := &ChatRequest{
		Model:    g.model,
		Messages: messages,
		Tools:    []Tool{WriteFileToolDefinition},
		Stream:   false,
		Options: map[string]any{
			"temperature": 0.2,
			"top_p":       0.9,
			"top_k":       40,
			"num_predict": 4096,
		},
	}

	// 3. 调用 Ollama /api/chat
	resp, err := g.client.Chat(ctx, req)
	if err != nil {
		// Ollama 连接失败
		return nil, errors.NewScaffoldError(
			errors.CodeOllamaConnection,
			"调用 Ollama 服务失败，请确认 Ollama 已启动",
			err,
		)
	}

	// 4. 解析 tool_calls
	code, err := parseToolCalls(resp.Message.ToolCalls)
	if err != nil {
		// 模型输出格式错误
		return nil, errors.NewScaffoldError(
			errors.CodeInvalidOutput,
			"解析模型输出失败",
			err,
		)
	}

	// 5. 对每个文件执行 go/ast 语法校验
	for _, f := range code.Files {
		if err := g.ValidateSyntax(f.Content); err != nil {
			return nil, errors.NewScaffoldError(
				errors.CodeInvalidOutput,
				fmt.Sprintf("生成代码语法校验失败 [%s]", f.Path),
				err,
			).WithDetails("请检查模型输出或更换模型后重试")
		}
	}

	log.Info("AI 代码生成成功", "files", len(code.Files))
	return code, nil
}

// GenerateWithRetry 带重试的代码生成。
// 在 Ollama 连接失败、tool_calls 为空、参数解析失败或语法校验失败时重试。
// 超过最大重试次数后返回最终错误，并附带排查指引。
func (g *CodeGenerator) GenerateWithRetry(
	ctx context.Context,
	input *GenerateInput,
	meta *metadata.ProjectMetadata,
	policy *RetryPolicy,
) (*GeneratedCode, error) {
	// 使用默认重试策略
	if policy == nil {
		policy = DefaultRetryPolicy()
	}

	var lastErr error
	// 最多重试 MaxRetries+1 次（含首次）
	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		if attempt > 0 {
			// 计算退避等待时间
			wait := time.Duration(float64(policy.Backoff) * pow(policy.BackoffFactor, float64(attempt-1)))
			log.Info("重试 AI 代码生成", "attempt", attempt, "wait", wait)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}

		// 尝试生成
		code, err := g.Generate(ctx, input, meta)
		if err == nil {
			return code, nil
		}
		lastErr = err

		// 判断错误是否可重试
		if se, ok := errors.IsScaffoldError(err); ok {
			// Ollama 连接失败、模型输出格式错误可重试
			if se.Code == errors.CodeOllamaConnection || se.Code == errors.CodeInvalidOutput {
				log.Warn("AI 代码生成失败，准备重试", "attempt", attempt, "code", se.Code)
				continue
			}
		}
		// 其他错误不重试
		break
	}

	// 超过最大重试次数
	return nil, errors.NewScaffoldError(
		errors.CodeInvalidOutput,
		fmt.Sprintf("AI 代码生成失败，已重试 %d 次", policy.MaxRetries),
		lastErr,
	).WithDetails(
		"排查建议：",
		"1. 确认 Ollama 服务已启动：ollama serve",
		"2. 确认模型已拉取：ollama pull "+g.model,
		"3. 检查模型是否支持 Tool Calling",
		"4. 尝试更换模型或调整 Prompt",
	)
}

// ValidateSyntax 校验生成代码的语法正确性。
// 使用 go/parser 解析源代码，仅做语法层面校验，不进行类型检查。
// 非 .go 文件（如 .sql）跳过校验直接通过。
func (g *CodeGenerator) ValidateSyntax(code string) error {
	// 使用 go/parser 解析源代码
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "", code, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("Go 语法校验失败: %w", err)
	}
	return nil
}

// ValidateSyntaxForPath 根据文件路径决定是否校验语法。
// .go 文件执行 go/parser 校验，其他文件跳过。
func (g *CodeGenerator) ValidateSyntaxForPath(path, content string) error {
	// 仅对 .go 文件执行语法校验
	if strings.HasSuffix(path, ".go") {
		return g.ValidateSyntax(content)
	}
	return nil
}

// pow 计算 base^exp，用于退避时间计算。
// 仅支持非负整数指数。
func pow(base float64, exp float64) float64 {
	result := 1.0
	for i := 0.0; i < exp; i++ {
		result *= base
	}
	return result
}
