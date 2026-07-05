// Package main 是生成项目的后端服务入口。
// 它负责加载配置、初始化数据库连接、注册路由和启动 HTTP 服务。
// 使用 Echo 作为 Web 框架，Ent 作为 ORM，PostgreSQL 作为数据库。
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"{{.ModulePath}}/internal/config"
	"{{.ModulePath}}/internal/handler"
	"{{.ModulePath}}/internal/middleware"
	"{{.ModulePath}}/internal/repository"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func main() {
	// 解析子命令（支持 migrate 子命令执行数据库迁移）
	migrateFlag := flag.Bool("migrate", false, "执行数据库迁移后退出")
	flag.Parse()

	// 初始化结构化日志
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// 加载应用配置（从环境变量）
	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", slog.Any("err", err))
		os.Exit(1)
	}
	slog.Info("配置加载完成", slog.String("db", cfg.DBName), slog.String("port", cfg.Port))

	// 初始化数据库连接（Ent 客户端）
	repo, err := repository.New(cfg)
	if err != nil {
		slog.Error("初始化数据库失败", slog.Any("err", err))
		os.Exit(1)
	}
	defer repo.Close()

	// 若指定 migrate 则执行迁移后退出
	if *migrateFlag {
		slog.Info("执行数据库迁移")
		if err := repo.AutoMigrate(context.Background()); err != nil {
			slog.Error("数据库迁移失败", slog.Any("err", err))
			os.Exit(1)
		}
		slog.Info("数据库迁移完成")
		return
	}

	// 创建 Echo 实例
	e := echo.New()
	e.HideBanner = true

	// 注册全局中间件
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// 注册错误处理中间件
	e.HTTPErrorHandler = middleware.ErrorHandler

	// 健康检查路由
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// OpenAPI 3.1 规格声明路由
	e.GET("/openapi.json", func(c echo.Context) error {
		openapiSpec := map[string]interface{}{
			"openapi": "3.1.0",
			"info": map[string]string{
				"title":       "{{.ProjectName}} API",
				"version":     "1.0.0",
				"description": "Golang AI 原生生成的 RESTful API 文档",
			},
			"paths": map[string]interface{}{
				"/api/v1/ping": map[string]interface{}{
					"get": map[string]interface{}{
						"summary":     "Ping 健康探活接口",
						"responses": map[string]interface{}{
							"200": map[string]string{"description": "成功返回 pong"},
						},
					},
				},
			},
		}
		return c.JSON(http.StatusOK, openapiSpec)
	})

	// Swagger UI 交互式文档页面
	e.GET("/docs", func(c echo.Context) error {
		htmlContent := `<!DOCTYPE html>
<html lang="zh">
<head>
  <meta charset="UTF-8">
  <title>OpenAPI 3.1 文档 - {{.ProjectName}}</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: '/openapi.json',
      dom_id: '#swagger-ui',
    });
  </script>
</body>
</html>`
		return c.HTML(http.StatusOK, htmlContent)
	})

	// 注册业务路由（API v1）
	handler.RegisterRoutes(e, repo)

	// 启动 HTTP 服务（支持优雅关闭）
	go func() {
		addr := ":" + cfg.Port
		slog.Info("HTTP 服务启动", slog.String("addr", addr))
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP 服务异常", slog.Any("err", err))
		}
	}()

	// 等待中断信号实现优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("正在关闭服务...")

	// 设置关闭超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		slog.Error("服务关闭异常", slog.Any("err", err))
	}
	slog.Info("服务已停止")
}

// init 用于在包初始化时打印启动信息（保留以兼容 Go 的 init 机制）
func init() {
	fmt.Println("{{.ProjectName}} 服务启动中...")
}
