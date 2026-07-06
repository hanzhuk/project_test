package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// osStat 封装 os.Stat，返回文件信息和是否存在。
func osStat(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// titleCaseStr 将字符串首字母大写。
func titleCaseStr(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// dbTitleStr 返回数据库的标题形式。
func dbTitleStr(db string) string {
	switch db {
	case "postgres":
		return "PostgreSQL"
	case "mysql":
		return "MySQL"
	case "sqlite":
		return "SQLite"
	default:
		return titleCaseStr(db)
	}
}

// frontendTitleStr 返回前端框架的标题形式。
func frontendTitleStr(f string) string {
	switch f {
	case "react":
		return "React 19 + TypeScript + Vite"
	case "vue":
		return "Vue 3 + TypeScript + Vite"
	case "svelte":
		return "Svelte 5 + TypeScript + Vite"
	default:
		return titleCaseStr(f)
	}
}

// injectRoutes 将实体路由自动注入到 handler.go 的 RegisterRoutes 函数。
// 若路由已存在则跳过，保证幂等。
func injectRoutes(projectDir, entity string) error {
	handlerPath := filepath.Join(projectDir, "internal", "handler", "handler.go")
	data, err := os.ReadFile(handlerPath)
	if err != nil {
		return err
	}

	entityLower := strings.ToLower(entity)
	entityPlural := entityLower + "s"

	// 已注册则跳过
	if strings.Contains(string(data), fmt.Sprintf("h.Create%s", entity)) {
		return nil
	}

	routes := fmt.Sprintf(
		"\n\t// %s 路由\n\tv1.POST(\"/%s\", h.Create%s)\n\tv1.GET(\"/%s\", h.List%ss)\n\tv1.GET(\"/%s/:id\", h.Get%s)\n\tv1.PUT(\"/%s/:id\", h.Update%s)\n\tv1.DELETE(\"/%s/:id\", h.Delete%s)",
		entity, entityPlural, entity, entityPlural, entity, entityPlural, entity, entityPlural, entity, entityPlural, entity,
	)

	// 将 TODO 注释块替换为实际路由
	lines := strings.Split(string(data), "\n")
	var out []string
	skipTodo := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "// TODO: 在此处注册业务路由") {
			out = append(out, routes)
			skipTodo = true
			continue
		}
		// 跳过 TODO 块内的注释行（形如 // v1.XXX）
		if skipTodo && strings.HasPrefix(trimmed, "// v1.") {
			continue
		}
		skipTodo = false
		out = append(out, line)
	}

	return os.WriteFile(handlerPath, []byte(strings.Join(out, "\n")), 0o644)
}

// fixTsxImports 修复 AI 生成的 TSX 文件常见 import 问题：
// 1. 补全缺失的 React hooks（useState / useEffect 等）
// 2. 将 React.FormEvent 替换为 FormEvent 并加入 import
func fixTsxImports(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	src := string(data)

	// 修复 React.XxxType → XxxType
	reactTypes := []string{"FormEvent", "ChangeEvent", "MouseEvent", "KeyboardEvent", "FocusEvent"}
	for _, t := range reactTypes {
		src = strings.ReplaceAll(src, "React."+t, t)
	}

	// 收集实际用到的 hooks 和类型
	candidates := []string{"useState", "useEffect", "useCallback", "useMemo", "useRef",
		"useContext", "FormEvent", "ChangeEvent", "MouseEvent", "KeyboardEvent"}
	var needed []string
	for _, name := range candidates {
		if strings.Contains(src, name) {
			needed = append(needed, name)
		}
	}
	if len(needed) == 0 {
		return os.WriteFile(filePath, []byte(src), 0o644)
	}

	newImport := fmt.Sprintf("import { %s } from 'react'", strings.Join(needed, ", "))

	// 替换现有的 react import 行
	lines := strings.Split(src, "\n")
	replaced := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import") &&
			(strings.Contains(line, "from 'react'") || strings.Contains(line, `from "react"`)) &&
			!strings.Contains(line, "from 'react-") {
			lines[i] = newImport
			replaced = true
			break
		}
	}
	if !replaced {
		// 在文件顶部插入
		lines = append([]string{newImport}, lines...)
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o644)
}
