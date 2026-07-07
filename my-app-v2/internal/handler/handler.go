// Package handler 实现 HTTP 请求处理函数。
// 使用 Huma 框架自动生成 OpenAPI 3.1 规范，所有 handler 方法通过 huma.Register 注册。
package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/my-app-v2/ent"
	"github.com/example/my-app-v2/internal/response"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
)

// Handler 是 HTTP 请求处理器，持有数据访问层引用。
type Handler struct {
	Client *ent.Client // Ent 数据库客户端
}

// New 创建 Handler 实例。
func New(client *ent.Client) *Handler {
	return &Handler{Client: client}
}

// RegisterRoutes 注册所有业务路由到 Huma API 实例。
// 路由统一以 /api/v1 为前缀，路由信息自动生成 OpenAPI 3.1 规范。
func RegisterRoutes(api huma.API, repo Repo) {
	h := New(repo.GetClient())

	// 健康检查接口
	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/api/v1/ping",
		Summary:     "Ping 健康探活",
		Description: "检查 API 服务是否正常运行",
		Tags:        []string{"健康检查"},
	}, h.Ping)

	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/api/v1/health",
		Summary:     "Health 健康检查",
		Description: "返回服务健康状态",
		Tags:        []string{"健康检查"},
	}, h.Ping)


	// Book 路由
	huma.Register(api, huma.Operation{Method: http.MethodPost,   Path: "/api/v1/books",     Summary: "创建Book",   Tags: []string{"Book"}}, h.CreateBook)
	huma.Register(api, huma.Operation{Method: http.MethodGet,    Path: "/api/v1/books",     Summary: "查询Book列表", Tags: []string{"Book"}}, h.ListBooks)
	huma.Register(api, huma.Operation{Method: http.MethodGet,    Path: "/api/v1/books/{id}", Summary: "查询单个Book", Tags: []string{"Book"}}, h.GetBook)
	huma.Register(api, huma.Operation{Method: http.MethodPut,    Path: "/api/v1/books/{id}", Summary: "更新Book",   Tags: []string{"Book"}}, h.UpdateBook)
	huma.Register(api, huma.Operation{Method: http.MethodDelete, Path: "/api/v1/books/{id}", Summary: "删除Book",   Tags: []string{"Book"}}, h.DeleteBook)
}

// Repo 定义数据访问层接口，handler 通过它获取数据库客户端。
type Repo interface {
	GetClient() *ent.Client
}

// PingInput Ping 接口输入（无参数）。
type PingInput struct{}

// PingOutput Ping 接口输出。
type PingOutput struct {
	Body response.Response
}

// Ping 是健康检查接口，返回服务状态。
func (h *Handler) Ping(ctx context.Context, input *PingInput) (*PingOutput, error) {
	slog.InfoContext(ctx, "收到 ping 请求")
	return &PingOutput{Body: response.Success("pong")}, nil
}

// parseID 从 Echo 上下文解析整数 ID（供混用 Echo 路由时使用）。
func parseID(c echo.Context) (int, error) {
	return strconv.Atoi(c.Param("id"))
}

// parsePagination 解析分页参数，默认 page=1, pageSize=10。
func parsePagination(c echo.Context) (int, int) {
	page, pageSize := 1, 10
	if p := c.QueryParam("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	if ps := c.QueryParam("page_size"); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil && n > 0 && n <= 100 {
			pageSize = n
		}
	}
	return page, pageSize
}

// offset 计算分页偏移量。
func offset(page, pageSize int) int { return (page - 1) * pageSize }

// contains 检查字符串是否包含子串（不区分大小写）。
func contains(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
