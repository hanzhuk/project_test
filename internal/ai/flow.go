// Package ai 提供基于 Ollama 本地模型的代码生成与 AI 流程控制。
// 本文件实现 Genkit Go / CloudWeGo Eino 风格的 AI Flow (DefineFlow) 管道抽象。
package ai

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/go-scaffold/internal/metadata"
)

// CodeGenFlow 代表一个 Genkit / Eino 风格的代码生成管道。
type CodeGenFlow struct {
	Name       string
	Client     *OllamaClient
	Metadata   *metadata.ProjectMetadata
	RetryPolicy *RetryPolicy
}

// NewCodeGenFlow 创建一个代码生成 Flow 管道。
func NewCodeGenFlow(client *OllamaClient, meta *metadata.ProjectMetadata) *CodeGenFlow {
	return &CodeGenFlow{
		Name:        "codeGeneratorFlow",
		Client:      client,
		Metadata:    meta,
		RetryPolicy: DefaultRetryPolicy(),
	}
}

// Run 执行 AI 代码生成流程 (类似于 Genkit DefineFlow 的 handler 逻辑)。
func (f *CodeGenFlow) Run(ctx context.Context, input *GenerateInput) (*GeneratedCode, error) {
	slog.InfoContext(ctx, "执行 Genkit/Eino AI 代码生成 Flow",
		slog.String("flow", f.Name),
		slog.String("entity", input.Entity),
		slog.String("code_type", input.CodeType),
	)

	// 1. 构建 Prompt
	messages := buildChatMessages(input, f.Metadata)

	// 2. 准备工具列表（包含 write_file）
	tools := []Tool{WriteFileToolDefinition}

	// 3. 带指数退避的重试循环
	var lastErr error
	backoff := f.RetryPolicy.Backoff

	for attempt := 0; attempt <= f.RetryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			slog.InfoContext(ctx, "Flow 重试 AI 代码生成", slog.Int("attempt", attempt), slog.Duration("wait", backoff))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff = time.Duration(float64(backoff) * f.RetryPolicy.BackoffFactor)
		}

		// 调用 Ollama /api/chat
		req := &ChatRequest{
			Model:    "qwen2.5-coder:7b",
			Messages: messages,
			Tools:    tools,
		}
		resp, err := f.Client.Chat(ctx, req)
		if err != nil {
			lastErr = fmt.Errorf("调用 Ollama 失败: %w", err)
			slog.WarnContext(ctx, "AI 交互异常，准备重试", slog.Int("attempt", attempt), slog.Any("err", err))
			continue
		}

		// 解析返回的 tool_calls
		result, err := parseToolCalls(resp.Message.ToolCalls)
		if err != nil {
			lastErr = fmt.Errorf("解析模型代码失败: %w", err)
			slog.WarnContext(ctx, "模型输出未匹配工具格式，准备重试", slog.Int("attempt", attempt), slog.Any("err", err))
			continue
		}

		slog.InfoContext(ctx, "Genkit/Eino AI Flow 执行成功", slog.Int("files", len(result.Files)))
		return result, nil
	}

	return nil, fmt.Errorf("AI 代码生成 Flow 失败 (已重试 %d 次): %w", f.RetryPolicy.MaxRetries, lastErr)
}
