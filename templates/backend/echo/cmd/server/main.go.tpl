// Package main 是生成项目的后端服务入口。
// 它负责加载配置、初始化数据库连接、注册路由和启动 HTTP 服务。
// 使用 Echo 作为 Web 框架，Ent 作为 ORM，PostgreSQL 作为数据库。
// OpenAPI 3.1 规范由 Huma 框架自动生成，访问 /docs 查看交互文档。
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

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func main() {
	// 自动加载 .env 文件（文件不存在时忽略，不影响正常启动）
	_ = godotenv.Load()

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

	// 初始化数据库连接
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

	// 健康检查路由（独立于 Huma，始终可用）
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// 初始化 Huma API（自动注册 /openapi.json 与 /docs 交互式在线文档）
	humaConfig := huma.DefaultConfig("{{.ProjectName}} API", "1.0.0")
	humaConfig.Info.Description = "基于 Go + Echo + Ent + PostgreSQL 的全栈 API，由 Go Scaffold 生成"
	humaConfig.OpenAPIPath = "/openapi.json"
	humaConfig.DocsPath = "/docs"
	api := humaecho.New(e, humaConfig)

	// 注册业务路由（API v1）
	handler.RegisterRoutes(api, repo)

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		slog.Error("服务关闭异常", slog.Any("err", err))
	}
	slog.Info("服务已停止")
}

// init 用于在包初始化时打印启动信息。
func init() {
	fmt.Println("{{.ProjectName}} 服务启动中...")
}
