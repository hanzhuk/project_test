package command

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/example/go-scaffold/internal/dependency"
	"github.com/example/go-scaffold/internal/log"
	"github.com/example/go-scaffold/internal/project"
	"github.com/example/go-scaffold/internal/template"
	"github.com/example/go-scaffold/internal/tui"
)

// newInitCmd 创建 init 命令。
// init 命令用于初始化新项目，生成交互式配置和完整项目结构。
func newInitCmd() *cobra.Command {
	var (
		backend     string
		orm         string
		database    string
		frontend    string
		jwt         bool
		docker      bool
		ci          bool
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "init <project-name> [flags]",
		Short: "初始化新项目",
		Long: `初始化一个新的全栈项目。

根据指定的技术栈（后端框架、ORM、数据库、前端框架）和功能开关，
从内置模板渲染生成完整的项目目录结构、配置文件和基础代码。

若未通过 flag 指定技术栈，将启动交互式 TUI 引导选择。

示例：
  go-scaffold init my-api
  go-scaffold init my-api --backend echo --orm ent --db postgres --frontend react --jwt
  go-scaffold init my-api --docker --ci --module-prefix github.com/myorg`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			modulePrefix := resolveModulePath(modulePrefixFlag(cmd))

			// 构建初始选项
			opts := project.ProjectOptions{
				Name:         projectName,
				ModulePath:   fmt.Sprintf("%s/%s", modulePrefix, projectName),
				Backend:      backend,
				ORM:          orm,
				Database:     database,
				Frontend:     frontend,
				EnableJWT:    jwt,
				EnableDocker: docker,
				EnableCI:     ci,
			}

			// 若启用交互式且未指定关键配置，则启动 TUI
			needTUI := interactive && (opts.Backend == "" || opts.ORM == "" ||
				opts.Database == "" || opts.Frontend == "")
			if needTUI {
				log.Info("启动交互式 TUI 配置")
				// 应用配置默认值作为 TUI 初始选项
				if globalCfg != nil {
					if opts.Backend == "" {
						opts.Backend = globalCfg.DefaultBackend
						// 置空让 TUI 处理（若想跳过已填则保留）
					}
				}
				// 重新置空以触发 TUI 步骤（除非用户显式指定）
				tuiOpts := tui.TUIOptions{
					Initial: project.ProjectOptions{
						Name:         opts.Name,
						ModulePath:   opts.ModulePath,
						Backend:      backend,
						ORM:          orm,
						Database:     database,
						Frontend:     frontend,
						EnableJWT:    jwt,
						EnableDocker: docker,
						EnableCI:     ci,
					},
				}
				t := tui.NewTUI(tuiOpts)
				result, err := t.Run()
				if err != nil {
					return fmt.Errorf("TUI 配置失败: %w", err)
				}
				// 合并 TUI 结果，保留命令行已有的名称和模块路径
				result.Name = opts.Name
				result.ModulePath = opts.ModulePath
				opts = result
			}

			// 应用默认值（TUI 跳过后或非交互模式）
			if opts.Backend == "" {
				opts.Backend = "echo"
			}
			if opts.ORM == "" {
				opts.ORM = "ent"
			}
			if opts.Database == "" {
				opts.Database = "postgres"
			}
			if opts.Frontend == "" {
				opts.Frontend = "react"
			}

			// 定位模板目录（相对于可执行文件或工作目录的 templates/）
			tplDir, err := findTemplatesDir()
			if err != nil {
				return fmt.Errorf("定位模板目录失败: %w", err)
			}
			engine, err := template.NewTemplateEngine(tplDir)
			if err != nil {
				return fmt.Errorf("初始化模板引擎失败: %w", err)
			}

			// 创建依赖管理器（工作目录稍后由 Generate 设置）
			depMgr := dependency.NewDependencyManager("")

			// 创建项目生成器并执行
			generator := project.NewProjectGenerator(engine, depMgr)
			if err := generator.Generate(opts); err != nil {
				return err
			}

			// 输出成功信息
			printInitSuccess(opts)
			return nil
		},
	}

	// 注册命令 flag
	cmd.Flags().StringVar(&backend, "backend", "", "后端框架，可选值：echo、fiber、gin")
	cmd.Flags().StringVar(&orm, "orm", "", "ORM 框架，可选值：ent、sqlc、gorm")
	cmd.Flags().StringVar(&database, "db", "", "数据库，可选值：postgres、mysql、sqlite")
	cmd.Flags().StringVar(&database, "database", "", "数据库，可选值：postgres、mysql、sqlite")
	cmd.Flags().StringVar(&frontend, "frontend", "", "前端框架，可选值：react、vue、svelte")
	cmd.Flags().BoolVar(&jwt, "jwt", false, "是否启用 JWT 认证")
	cmd.Flags().BoolVar(&docker, "docker", false, "是否生成 Docker 配置")
	cmd.Flags().BoolVar(&ci, "ci", false, "是否生成 GitHub Actions 配置")
	cmd.Flags().BoolVar(&interactive, "interactive", true, "是否启用交互式 TUI")

	return cmd
}

// modulePrefixFlag 从命令中读取 --module-prefix flag 的值。
func modulePrefixFlag(cmd *cobra.Command) string {
	val, err := cmd.Flags().GetString("module-prefix")
	if err != nil {
		return ""
	}
	return val
}

// findTemplatesDir 定位脚手架的模板目录。
// 依次尝试：工作目录 templates/、可执行文件同级 templates/。
func findTemplatesDir() (string, error) {
	candidates := []string{
		"templates",
		filepath.Join(".", "templates"),
	}
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		// 检查目录是否存在且包含 backend 子目录
		isDir, err := osStat(abs)
		if err == nil && isDir {
			return abs, nil
		}
	}
	return "", fmt.Errorf("未找到模板目录 templates/")
}

// printInitSuccess 输出项目创建成功的提示信息。
func printInitSuccess(opts project.ProjectOptions) {
	fmt.Printf("\n✓ 项目 %s 创建成功\n", opts.Name)
	fmt.Printf("  后端框架: %s\n", titleCaseStr(opts.Backend))
	fmt.Printf("  ORM: %s\n", titleCaseStr(opts.ORM))
	fmt.Printf("  数据库: %s\n", dbTitleStr(opts.Database))
	fmt.Printf("  前端框架: %s\n", frontendTitleStr(opts.Frontend))
	fmt.Printf("  模块路径: %s\n", opts.ModulePath)
	if opts.EnableJWT {
		fmt.Printf("  JWT 认证: 已启用\n")
	}
	if opts.EnableDocker {
		fmt.Printf("  Docker: 已生成\n")
	}
	if opts.EnableCI {
		fmt.Printf("  CI/CD: 已生成\n")
	}
	fmt.Printf("\n后续步骤:\n")
	fmt.Printf("  1. 进入项目目录:\n")
	fmt.Printf("     cd %s\n", opts.Name)
	fmt.Printf("  2. AI 自动生成业务代码 (可选):\n")
	fmt.Printf("     ..\\go-scaffold.exe generate \"user CRUD\" --no-tui\n")
	if opts.ORM == "ent" {
		fmt.Printf("  3. 编译 Ent Schema 客户端:\n")
		fmt.Printf("     go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema\n")
	}
	fmt.Printf("  4. 整理并下载包依赖:\n")
	fmt.Printf("     go mod tidy\n")
	if opts.EnableDocker {
		fmt.Printf("  5. 启动数据库容器:\n")
		fmt.Printf("     docker compose up -d db\n")
	}
	fmt.Printf("  6. 运行后端服务:\n")
	fmt.Printf("     go run ./cmd/server\n")
	if opts.Frontend != "" {
		fmt.Printf("  7. 启动前端界面 (另开终端窗口):\n")
		fmt.Printf("     cd web && npm install && npm run dev\n")
	}
}
