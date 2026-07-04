// Package command 实现脚手架的命令行命令注册与执行逻辑。
// 包含 root（根命令）、init（项目初始化）、generate（AI 代码生成）、build（项目构建）四个命令。
package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/example/go-scaffold/internal/config"
	"github.com/example/go-scaffold/internal/log"
)

// 全局配置实例，由 root 命令加载，供子命令使用。
var globalCfg *config.Config

// NewRootCmd 创建并返回脚手架的根命令。
// 根命令负责加载全局配置，并注册 init、generate、build 三个子命令。
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "go-scaffold",
		Short: "Golang AI 原生全栈应用快速开发脚手架",
		Long: `Go Scaffold 是一个 AI 原生的 Go 全栈应用快速开发脚手架。

通过命令行工具实现项目初始化、模板渲染、AI 代码生成、依赖管理和部署配置生成，
帮助开发者在几分钟内获得一个可运行的全栈项目骨架。

核心命令：
  init      初始化新项目，生成交互式配置和完整项目结构
  generate  基于自然语言描述，调用本地 Ollama 模型生成业务代码
  build     构建生成的项目，编译二进制或构建 Docker 镜像

示例：
  go-scaffold init my-api
  go-scaffold generate "user CRUD"
  go-scaffold build --docker`,
		// 根命令执行前的初始化：加载全局配置
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 加载全局配置（help 和 completion 命令跳过）
			if cmd.Name() == "help" || cmd.Name() == "completion" {
				return nil
			}
			cfg, err := config.LoadConfig()
			if err != nil {
				// 配置加载失败时使用默认配置并警告
				log.Warn("加载配置失败，使用默认配置", "err", err)
				cfg = config.DefaultConfig()
			}
			globalCfg = cfg
			// 根据配置重新初始化日志级别
			log.Init(cfg.LogLevel)
			return nil
		},
		SilenceUsage: true,
	}

	// 注册全局 flag
	rootCmd.PersistentFlags().String("module-prefix", "", "Go 模块路径前缀，默认 github.com/example")

	// 注册子命令
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newGenerateCmd())
	rootCmd.AddCommand(newBuildCmd())

	return rootCmd
}

// resolveModulePath 解析最终的模块路径。
// 优先使用 flag 值，其次使用配置文件的值，最后使用默认值。
func resolveModulePath(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if globalCfg != nil && globalCfg.ModulePrefix != "" {
		return globalCfg.ModulePrefix
	}
	return "github.com/example"
}

// resolveOllamaModel 解析最终使用的 Ollama 模型。
// 优先使用 flag 值，其次使用配置文件的值，最后使用默认值。
func resolveOllamaModel(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if globalCfg != nil && globalCfg.OllamaModel != "" {
		return globalCfg.OllamaModel
	}
	return "ornith:9b"
}

// resolveOllamaHost 解析最终的 Ollama 服务地址。
func resolveOllamaHost(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if globalCfg != nil && globalCfg.OllamaHost != "" {
		return globalCfg.OllamaHost
	}
	return "http://localhost:11434"
}

// exitWithError 以友好的方式输出错误并退出。
func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "错误: %v\n", err)
	os.Exit(1)
}
