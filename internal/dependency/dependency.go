// Package dependency 负责初始化生成项目的依赖。
// 支持初始化 Go 模块（go mod init）和 Node.js 项目（package.json），
// 并根据技术栈添加对应的 Go 依赖。
package dependency

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/example/go-scaffold/internal/errors"
	"github.com/example/go-scaffold/internal/log"
)

// DependencyManager 负责初始化生成项目的依赖。
// WorkDir 为生成项目的根目录路径。
type DependencyManager struct {
	WorkDir string // 工作目录（生成项目根目录）
}

// NewDependencyManager 创建依赖管理器实例。
// workDir 为生成项目的根目录路径。
func NewDependencyManager(workDir string) *DependencyManager {
	return &DependencyManager{WorkDir: workDir}
}

// InitGoModule 初始化 Go 模块，执行 go mod init <modulePath>。
// modulePath 为 Go 模块完整路径，如 github.com/example/my-api。
// 若 go.mod 已存在（例如由模板渲染生成）则跳过初始化。
func (m *DependencyManager) InitGoModule(modulePath string) error {
	// 检查 go.mod 是否已存在（模板可能已生成）
	goModPath := filepath.Join(m.WorkDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		log.Info("go.mod 已存在，跳过 go mod init", "dir", m.WorkDir)
		return nil
	}
	log.Info("初始化 Go 模块", "module", modulePath, "dir", m.WorkDir)
	// 执行 go mod init
	cmd := exec.Command("go", "mod", "init", modulePath)
	cmd.Dir = m.WorkDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.NewScaffoldError(
			errors.CodeInternal,
			fmt.Sprintf("初始化 Go 模块失败: %s", modulePath),
			err,
		)
	}
	return nil
}

// AddGoDeps 根据技术栈添加 Go 依赖。
// backend 为后端框架（echo/fiber/gin），orm 为 ORM（ent/sqlc/gorm），
// database 为数据库（postgres/mysql/sqlite）。
func (m *DependencyManager) AddGoDeps(backend, orm, database string) error {
	// 收集依赖列表
	deps := collectGoDeps(backend, orm, database)
	if len(deps) == 0 {
		return nil
	}
	log.Info("添加 Go 依赖", "count", len(deps), "dir", m.WorkDir)
	// 执行 go get 添加依赖
	args := append([]string{"get"}, deps...)
	cmd := exec.Command("go", args...)
	cmd.Dir = m.WorkDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.NewScaffoldError(
			errors.CodeInternal,
			"添加 Go 依赖失败",
			err,
		).WithDetails("依赖: " + strings.Join(deps, ", "))
	}
	// 执行 go mod tidy 整理依赖
	if err := m.tidy(); err != nil {
		log.Warn("go mod tidy 执行失败，可稍后手动执行", "err", err)
	}
	return nil
}

// tidy 执行 go mod tidy 整理依赖。
func (m *DependencyManager) tidy() error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = m.WorkDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// InitNodeProject 初始化前端项目 package.json。
// projectName 为前端项目名称。
// 若 package.json 已存在则跳过初始化。
func (m *DependencyManager) InitNodeProject(projectName string) error {
	pkgPath := filepath.Join(m.WorkDir, "package.json")
	// 若已存在则跳过
	if _, err := os.Stat(pkgPath); err == nil {
		log.Info("package.json 已存在，跳过初始化", "dir", m.WorkDir)
		return nil
	}
	log.Info("初始化 Node.js 项目", "name", projectName, "dir", m.WorkDir)
	// 使用 npm init -y 快速生成 package.json
	cmd := exec.Command("npm", "init", "-y")
	cmd.Dir = m.WorkDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.NewScaffoldError(
			errors.CodeInternal,
			"初始化 Node.js 项目失败",
			err,
		)
	}
	return nil
}

// InstallDeps 执行依赖安装。
// language 支持 "go"（执行 go mod tidy）和 "node"（执行 npm install）。
func (m *DependencyManager) InstallDeps(language string) error {
	switch strings.ToLower(language) {
	case "go":
		log.Info("安装 Go 依赖", "dir", m.WorkDir)
		return m.tidy()
	case "node", "npm":
		log.Info("安装 Node.js 依赖", "dir", m.WorkDir)
		cmd := exec.Command("npm", "install")
		cmd.Dir = m.WorkDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return errors.NewScaffoldError(
				errors.CodeInternal,
				"安装 Node.js 依赖失败",
				err,
			)
		}
		return nil
	default:
		return errors.NewScaffoldError(
			errors.CodeValidation,
			fmt.Sprintf("不支持的依赖类型: %s（仅支持 go/node）", language),
			nil,
		)
	}
}

// collectGoDeps 根据技术栈组合收集需要添加的 Go 依赖。
func collectGoDeps(backend, orm, database string) []string {
	var deps []string
	// 后端框架依赖
	switch strings.ToLower(backend) {
	case "echo":
		deps = append(deps,
			"github.com/labstack/echo/v4@latest",
			"github.com/labstack/echo/v4/middleware@latest",
			"github.com/danielgtaylor/huma/v2@latest",
			"github.com/joho/godotenv@latest",
		)
	case "gin":
		deps = append(deps, "github.com/gin-gonic/gin@latest")
	case "fiber":
		deps = append(deps, "github.com/gofiber/fiber/v2@latest")
	}
	// ORM 依赖
	switch strings.ToLower(orm) {
	case "ent":
		deps = append(deps, "entgo.io/ent@latest")
	case "gorm":
		deps = append(deps, "gorm.io/gorm@latest")
	case "sqlc":
		deps = append(deps, "github.com/sqlc-dev/sqlc@latest")
	}
	// 数据库驱动依赖
	switch strings.ToLower(database) {
	case "postgres":
		deps = append(deps, "github.com/lib/pq@latest", "entgo.io/ent/dialect/sql@latest")
	case "mysql":
		deps = append(deps, "github.com/go-sql-driver/mysql@latest")
	case "sqlite":
		deps = append(deps, "github.com/mattn/go-sqlite3@latest")
	}
	// JWT 依赖（默认添加，便于后续扩展）
	deps = append(deps, "github.com/golang-jwt/jwt/v5@latest", "golang.org/x/crypto@latest")
	return deps
}
