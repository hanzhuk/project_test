// Package handler 实现 HTTP 请求处理函数。
// 它接收 HTTP 请求，调用 service 层处理业务逻辑，并返回统一格式的 JSON 响应。
package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"{{.ModulePath}}/ent"
	"{{.ModulePath}}/internal/response"

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

// RegisterRoutes 注册所有业务路由到 Echo 实例。
// 路由统一以 /api/v1 为前缀，遵循 RESTful 规范。
// repo 为数据访问层，需提供 Client() 方法返回 Ent 客户端。
func RegisterRoutes(e *echo.Echo, repo Repo) {
	h := New(repo.Client())

	// API v1 路由组
	v1 := e.Group("/api/v1")

	// 健康检查
	v1.GET("/ping", h.Ping)

	// TODO: 在此处注册业务路由，例如：
	// v1.POST("/users", h.CreateUser)
	// v1.GET("/users", h.ListUsers)
	// v1.GET("/users/:id", h.GetUser)
	// v1.PUT("/users/:id", h.UpdateUser)
	// v1.DELETE("/users/:id", h.DeleteUser)
}

// Repo 定义数据访问层接口，handler 通过它获取数据库客户端。
type Repo interface {
	Client() *ent.Client
}

// Ping 是健康检查接口，返回服务状态。
func (h *Handler) Ping(c echo.Context) error {
	slog.InfoContext(c.Request().Context(), "收到 ping 请求")
	return c.JSON(http.StatusOK, response.Success("pong"))
}

// parseID 从路径参数解析整数 ID。
// 若解析失败返回 0 和错误。
func parseID(c echo.Context) (int, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// parsePagination 解析分页参数。
// 默认 page=1, pageSize=10。
func parsePagination(c echo.Context) (int, int) {
	page := 1
	pageSize := 10
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
func offset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// contains 检查字符串是否包含子串（不区分大小写）。
func contains(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
