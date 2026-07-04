package command

import (
	"os"
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
