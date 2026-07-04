// Package ai 实现 AI 代码生成功能。
// 通过调用本地 Ollama 的 /api/chat 端点，携带 write_file 工具定义，
// 模型以 Tool Calling 方式返回结构化的文件写入指令，
// 脚手架解析后执行写入并通过 go/ast 校验语法。
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/example/go-scaffold/internal/log"
)

// OllamaClient 封装与本地 Ollama 服务的 HTTP 通信。
// Host 为 Ollama 服务地址，如 http://localhost:11434。
// HTTPClient 为底层 HTTP 客户端，可配置超时。
type OllamaClient struct {
	Host       string       // Ollama 服务地址
	HTTPClient *http.Client // HTTP 客户端
}

// ChatMessage 定义 /api/chat 的消息结构。
type ChatMessage struct {
	Role      string     `json:"role"`                 // 角色：system、user、assistant
	Content   string     `json:"content,omitempty"`    // 消息内容
	ToolCalls []ToolCall `json:"tool_calls,omitempty"` // 工具调用列表
}

// ToolCall 定义模型返回的工具调用。
type ToolCall struct {
	Function FunctionCall `json:"function"` // 函数调用信息
}

// FunctionCall 定义工具调用的函数名与参数。
// Arguments 可以是 JSON 对象，也可以是转义后的 JSON 字符串。
type FunctionCall struct {
	Name      string          `json:"name"`      // 函数名，如 write_file
	Arguments json.RawMessage `json:"arguments"` // 参数 JSON（RawMessage 兼容对象与字符串）
}

// Tool 定义传给 Ollama 的工具描述。
type Tool struct {
	Type     string       `json:"type"`     // 工具类型，固定为 function
	Function ToolFunction `json:"function"` // 函数定义
}

// ToolFunction 定义工具函数的名称、描述与参数 schema。
type ToolFunction struct {
	Name        string          `json:"name"`        // 函数名
	Description string          `json:"description"` // 函数描述
	Parameters  json.RawMessage `json:"parameters"`  // 参数 JSON Schema
}

// ChatRequest 定义 /api/chat 请求体。
type ChatRequest struct {
	Model    string        `json:"model"`             // 模型名称
	Messages []ChatMessage `json:"messages"`           // 消息数组
	Tools    []Tool        `json:"tools,omitempty"`    // 工具列表
	Stream   bool          `json:"stream"`             // 是否流式返回
	Options  map[string]any `json:"options,omitempty"` // 模型推理参数
}

// ChatResponse 定义 /api/chat 响应体。
type ChatResponse struct {
	Model           string      `json:"model"`              // 使用的模型名称
	CreatedAt       string      `json:"created_at"`         // 响应创建时间
	Message         ChatMessage `json:"message"`            // 包含 role、content、tool_calls
	Done            bool        `json:"done"`               // 是否生成完成
	DoneReason      string      `json:"done_reason"`        // 停止原因
	TotalDuration   int64       `json:"total_duration"`     // 总耗时（纳秒）
	LoadDuration    int64       `json:"load_duration"`      // 模型加载耗时
	PromptEvalCount int         `json:"prompt_eval_count"`  // Prompt token 数
	EvalCount       int         `json:"eval_count"`         // 生成 token 数
}

// NewOllamaClient 创建 Ollama 客户端实例。
// host 为 Ollama 服务地址，如 http://localhost:11434。
func NewOllamaClient(host string) *OllamaClient {
	return &OllamaClient{
		Host: host,
		HTTPClient: &http.Client{
			// 设置较长的超时，本地模型推理可能耗时较长
			Timeout: 120 * time.Second,
		},
	}
}

// Chat 调用 Ollama /api/chat 接口进行对话。
// model 为模型名称，messages 为消息数组，tools 为工具列表，options 为推理参数。
// 返回 ChatResponse 或错误。
func (c *OllamaClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// 构建请求体
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	// 构建请求 URL
	url := c.Host + "/api/chat"
	log.Debug("调用 Ollama /api/chat", "url", url, "model", req.Model)

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("连接 Ollama 服务失败 (%s): %w", c.Host, err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama 返回错误状态码 %d: %s", resp.StatusCode, string(respBody))
	}
	// 反序列化响应
	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	log.Debug("Ollama 响应",
		"done", chatResp.Done,
		"tool_calls", len(chatResp.Message.ToolCalls),
		"eval_count", chatResp.EvalCount,
		"duration_ms", chatResp.TotalDuration/1e6,
	)
	return &chatResp, nil
}

// Ping 检查 Ollama 服务是否可用。
// 通过 GET /api/tags 接口验证服务连通性。
func (c *OllamaClient) Ping(ctx context.Context) error {
	url := c.Host + "/api/tags"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("无法连接到 Ollama 服务 (%s): %w", c.Host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama 服务返回状态码 %d", resp.StatusCode)
	}
	return nil
}
