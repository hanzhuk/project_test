// Package project 负责根据用户选择的技术栈和配置，调用模板引擎生成完整项目结构。
// 它组合模板渲染、依赖初始化和元数据写入，完成项目初始化的全部工作。
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/go-scaffold/internal/dependency"
	"github.com/example/go-scaffold/internal/errors"
	"github.com/example/go-scaffold/internal/log"
	"github.com/example/go-scaffold/internal/metadata"
	"github.com/example/go-scaffold/internal/template"
)

// ProjectOptions 定义项目初始化参数。
// 这些参数由命令行或 TUI 收集，传递给 ProjectGenerator.Generate。
type ProjectOptions struct {
	Name          string // 项目名称
	ModulePath    string // Go 模块路径
	Backend       string // 后端框架：echo/fiber/gin
	ORM           string // ORM：ent/sqlc/gorm
	Database      string // 数据库：postgres/mysql/sqlite
	Frontend      string // 前端框架：react/vue/svelte
	EnableJWT     bool   // 是否启用 JWT 认证
	EnableDocker  bool   // 是否生成 Docker 配置
	EnableCI      bool   // 是否生成 CI 配置
	OutputDir     string // 输出目录
}

// ProjectGenerator 是项目生成器，组合模板引擎和依赖管理器。
type ProjectGenerator struct {
	engine  *template.TemplateEngine    // 模板引擎
	depMgr  *dependency.DependencyManager // 依赖管理器
}

// NewProjectGenerator 创建项目生成器。
// engine 为模板引擎实例，depMgr 为依赖管理器实例。
func NewProjectGenerator(
	engine *template.TemplateEngine,
	depMgr *dependency.DependencyManager,
) *ProjectGenerator {
	return &ProjectGenerator{
		engine: engine,
		depMgr: depMgr,
	}
}

// Generate 根据选项生成完整项目，包含模板渲染、依赖初始化和元数据写入。
func (g *ProjectGenerator) Generate(opts ProjectOptions) error {
	// 1. 校验选项合法性
	if err := g.Validate(opts); err != nil {
		return err
	}

	// 确定输出目录
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = opts.Name
	}
	// 转为绝对路径
	absDir, err := filepath.Abs(outputDir)
	if err != nil {
		return errors.NewScaffoldError(errors.CodeInternal, "解析输出目录路径失败", err)
	}

	log.Info("开始生成项目", "name", opts.Name, "dir", absDir)

	// 2. 检查目录是否已存在
	if info, err := os.Stat(absDir); err == nil && info.IsDir() {
		// 目录非空时报错
		entries, _ := os.ReadDir(absDir)
		if len(entries) > 0 {
			return errors.NewScaffoldError(
				errors.CodeAlreadyExists,
				fmt.Sprintf("项目目录已存在且非空: %s", absDir),
				nil,
			)
		}
	}

	// 3. 创建项目目录
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return errors.NewScaffoldError(errors.CodeInternal, "创建项目目录失败", err)
	}

	// 4. 构建模板渲染数据
	data := g.buildTemplateData(opts)

	// 5. 渲染后端模板
	backendTplDir := filepath.Join("backend", opts.Backend)
	if err := g.engine.RenderDir(backendTplDir, absDir, data); err != nil {
		return errors.NewScaffoldError(
			errors.CodeInternal,
			fmt.Sprintf("渲染后端模板失败 (%s)", opts.Backend),
			err,
		)
	}

	// 6. 渲染数据库相关模板
	dbTplDir := filepath.Join("database", opts.Database)
	if err := g.engine.RenderDir(dbTplDir, absDir, data); err != nil {
		log.Warn("渲染数据库模板失败，跳过", "db", opts.Database, "err", err)
	}

	// 7. 渲染 JWT 模板（若启用）
	if opts.EnableJWT {
		jwtTplDir := filepath.Join("features", "jwt")
		if err := g.engine.RenderDir(jwtTplDir, absDir, data); err != nil {
			log.Warn("渲染 JWT 模板失败，跳过", "err", err)
		}
	}

	// 8. 渲染 Docker 配置（若启用）
	if opts.EnableDocker {
		dockerTplDir := filepath.Join("deploy", "docker")
		if err := g.engine.RenderDir(dockerTplDir, absDir, data); err != nil {
			log.Warn("渲染 Docker 模板失败，跳过", "err", err)
		}
	}

	// 9. 渲染 CI 配置（若启用）
	if opts.EnableCI {
		ciTplDir := filepath.Join("deploy", "github-actions")
		if err := g.engine.RenderDir(ciTplDir, absDir, data); err != nil {
			log.Warn("渲染 CI 模板失败，跳过", "err", err)
		}
	}

	// 10. 渲染前端模板
	if opts.Frontend != "" {
		frontendDir := filepath.Join(absDir, "web")
		if err := os.MkdirAll(frontendDir, 0o755); err == nil {
			frontendTplDir := filepath.Join("frontend", opts.Frontend)
			if err := g.engine.RenderDir(frontendTplDir, frontendDir, data); err != nil {
				log.Warn("渲染前端模板失败，跳过", "err", err)
			}
		}
	}

	// 11. 写入项目元数据 .go-scaffold.json
	meta := &metadata.ProjectMetadata{
		ProjectName:  opts.Name,
		Backend:      opts.Backend,
		ORM:          opts.ORM,
		Database:     opts.Database,
		Frontend:     opts.Frontend,
		ModulePath:   opts.ModulePath,
		EnableJWT:    opts.EnableJWT,
		EnableDocker: opts.EnableDocker,
		EnableCI:     opts.EnableCI,
		GeneratedAt:  time.Now(),
	}
	if err := metadata.SaveMetadata(absDir, meta); err != nil {
		log.Warn("写入项目元数据失败", "err", err)
	}

	// 12. 初始化 Go 模块并添加依赖
	if g.depMgr != nil {
		g.depMgr.WorkDir = absDir
		if err := g.depMgr.InitGoModule(opts.ModulePath); err != nil {
			log.Warn("初始化 Go 模块失败，请稍后手动执行 go mod init", "err", err)
		}
		if err := g.depMgr.AddGoDeps(opts.Backend, opts.ORM, opts.Database); err != nil {
			log.Warn("添加 Go 依赖失败，请稍后手动执行 go mod tidy", "err", err)
		}
	}

	log.Info("项目生成完成", "dir", absDir)
	return nil
}

// Validate 校验项目选项的合法性。
// 检查项目名称、后端框架、ORM、数据库、前端框架是否为支持的取值。
func (g *ProjectGenerator) Validate(opts ProjectOptions) error {
	var details []string
	// 校验项目名称
	if opts.Name == "" {
		details = append(details, "项目名称不能为空")
	} else if !isValidProjectName(opts.Name) {
		details = append(details, "项目名称只能包含字母、数字、下划线和连字符")
	}
	// 校验后端框架
	validBackends := map[string]bool{"echo": true, "fiber": true, "gin": true}
	if !validBackends[opts.Backend] {
		details = append(details, fmt.Sprintf("不支持的后端框架: %s（可选 echo/fiber/gin）", opts.Backend))
	}
	// 校验 ORM
	validORMs := map[string]bool{"ent": true, "sqlc": true, "gorm": true}
	if !validORMs[opts.ORM] {
		details = append(details, fmt.Sprintf("不支持的 ORM: %s（可选 ent/sqlc/gorm）", opts.ORM))
	}
	// 校验数据库
	validDBs := map[string]bool{"postgres": true, "mysql": true, "sqlite": true}
	if !validDBs[opts.Database] {
		details = append(details, fmt.Sprintf("不支持的数据库: %s（可选 postgres/mysql/sqlite）", opts.Database))
	}
	// 校验前端框架
	validFrontends := map[string]bool{"react": true, "vue": true, "svelte": true}
	if opts.Frontend != "" && !validFrontends[opts.Frontend] {
		details = append(details, fmt.Sprintf("不支持的前端框架: %s（可选 react/vue/svelte）", opts.Frontend))
	}
	// 校验模块路径
	if opts.ModulePath == "" {
		details = append(details, "模块路径不能为空")
	}
	if len(details) > 0 {
		return errors.NewScaffoldError(
			errors.CodeValidation,
			"项目选项校验失败",
			nil,
		).WithDetails(details...)
	}
	return nil
}

// buildTemplateData 构建注入模板的变量数据。
func (g *ProjectGenerator) buildTemplateData(opts ProjectOptions) map[string]any {
	return map[string]any{
		"ProjectName":   opts.Name,
		"ModulePath":    opts.ModulePath,
		"Backend":       opts.Backend,
		"ORM":           opts.ORM,
		"Database":      opts.Database,
		"Frontend":      opts.Frontend,
		"EnableJWT":     opts.EnableJWT,
		"EnableDocker":  opts.EnableDocker,
		"EnableCI":      opts.EnableCI,
		// 派生变量
		"BackendTitle":  titleCase(opts.Backend),
		"ORMTitle":      titleCase(opts.ORM),
		"DBTitle":       dbTitle(opts.Database),
		"FrontendTitle": frontendTitle(opts.Frontend),
	}
}

// isValidProjectName 校验项目名称是否合法。
// 仅允许字母、数字、下划线和连字符。
func isValidProjectName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// titleCase 将字符串首字母大写。
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// dbTitle 返回数据库的标题形式。
func dbTitle(db string) string {
	switch db {
	case "postgres":
		return "PostgreSQL"
	case "mysql":
		return "MySQL"
	case "sqlite":
		return "SQLite"
	default:
		return titleCase(db)
	}
}

// frontendTitle 返回前端框架的标题形式。
func frontendTitle(f string) string {
	switch f {
	case "react":
		return "React 19 + TypeScript + Vite"
	case "vue":
		return "Vue 3 + TypeScript + Vite"
	case "svelte":
		return "Svelte 5 + TypeScript + Vite"
	default:
		return titleCase(f)
	}
}
