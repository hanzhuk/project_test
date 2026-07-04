// Package template 封装脚手架的模板引擎，基于 Go 标准库 text/template。
// 模板是脚手架内部的静态资源，渲染产物才是生成项目的代码。
// 模板文件位于脚手架内部 templates/ 目录下，按后端框架、ORM、数据库、前端分类。
package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	gotemplate "text/template"
)

// TemplateEngine 负责加载和渲染模板文件。
// BaseDir 为模板根目录，位于脚手架内部 templates/；
// FuncMap 存放自定义模板函数。
type TemplateEngine struct {
	BaseDir string               // 模板根目录
	FuncMap gotemplate.FuncMap   // 自定义模板函数
}

// NewTemplateEngine 创建模板引擎实例。
// baseDir 为模板根目录路径。
func NewTemplateEngine(baseDir string) (*TemplateEngine, error) {
	// 校验模板目录是否存在
	info, err := os.Stat(baseDir)
	if err != nil {
		return nil, fmt.Errorf("模板目录不存在: %s: %w", baseDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("模板路径不是目录: %s", baseDir)
	}
	// 初始化引擎并注册内置模板函数
	e := &TemplateEngine{
		BaseDir: baseDir,
		FuncMap: gotemplate.FuncMap{
			// title 将字符串首字母大写
			"title": strings.Title,
			// lower 将字符串转为小写
			"lower": strings.ToLower,
			// upper 将字符串转为大写
			"upper": strings.ToUpper,
			// toLowerSnake 将驼峰转为下划线小写
			"toLowerSnake": toLowerSnake,
			// contains 判断字符串是否包含子串
			"contains": strings.Contains,
		},
	}
	return e, nil
}

// RegisterFunc 注册自定义模板函数，供模板渲染时调用。
func (e *TemplateEngine) RegisterFunc(name string, fn any) error {
	if name == "" {
		return fmt.Errorf("模板函数名不能为空")
	}
	if fn == nil {
		return fmt.Errorf("模板函数不能为空: %s", name)
	}
	e.FuncMap[name] = fn
	return nil
}

// RenderFile 渲染单个模板文件到目标路径。
// tplPath 为相对于 BaseDir 的模板文件路径，dstPath 为输出文件路径。
// data 为注入模板的变量数据。
func (e *TemplateEngine) RenderFile(tplPath, dstPath string, data map[string]any) error {
	// 读取模板文件内容（先尝试相对路径，再尝试绝对路径）
	content, err := e.readTemplate(tplPath)
	if err != nil {
		return fmt.Errorf("读取模板文件失败 %s: %w", tplPath, err)
	}
	// 解析模板，name 使用文件名
	tmpl, err := gotemplate.New(filepath.Base(tplPath)).Funcs(e.FuncMap).Parse(content)
	if err != nil {
		return fmt.Errorf("解析模板失败 %s (行号见错误): %w", tplPath, err)
	}
	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("创建目标目录失败 %s: %w", filepath.Dir(dstPath), err)
	}
	// 创建目标文件
	f, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败 %s: %w", dstPath, err)
	}
	defer f.Close()
	// 执行模板渲染
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("渲染模板失败 %s: %w", tplPath, err)
	}
	return nil
}

// RenderString 渲染字符串模板，返回渲染结果。
// 用于动态生成的模板内容场景（如 AI 生成后的二次处理）。
func (e *TemplateEngine) RenderString(tplContent string, data map[string]any) (string, error) {
	tmpl, err := gotemplate.New("inline").Funcs(e.FuncMap).Parse(tplContent)
	if err != nil {
		return "", fmt.Errorf("解析字符串模板失败: %w", err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染字符串模板失败: %w", err)
	}
	return buf.String(), nil
}

// RenderDir 渲染整个模板目录到目标目录。
// tplDir 为相对于 BaseDir 的模板目录，dstDir 为输出目录。
// 跳过以 .tpl 之外的文件也会被渲染；模板文件名需以 .tpl 结尾才会被当作模板处理，
// 其他文件（如静态资源）直接拷贝。
// .tpl 后缀在输出时会自动去除。
func (e *TemplateEngine) RenderDir(tplDir, dstDir string, data map[string]any) error {
	// 解析模板目录的绝对路径
	srcDir := tplDir
	if !filepath.IsAbs(srcDir) {
		srcDir = filepath.Join(e.BaseDir, tplDir)
	}
	// 遍历模板目录
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 计算相对路径
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		// 跳过根目录
		if rel == "." {
			return nil
		}
		// 目标路径
		dstPath := filepath.Join(dstDir, rel)
		// 去除 .tpl 后缀
		if strings.HasSuffix(rel, ".tpl") {
			dstPath = strings.TrimSuffix(dstPath, ".tpl")
		}
		if d.IsDir() {
			// 创建目录
			return os.MkdirAll(dstPath, 0o755)
		}
		// 若是模板文件则渲染，否则直接拷贝
		if strings.HasSuffix(path, ".tpl") {
			return e.RenderFile(path, dstPath, data)
		}
		// 静态文件直接拷贝
		return copyFile(path, dstPath)
	})
}

// readTemplate 读取模板文件内容。
// path 可以是绝对路径或相对于 BaseDir 的相对路径。
func (e *TemplateEngine) readTemplate(path string) (string, error) {
	// 若为绝对路径直接读取
	if filepath.IsAbs(path) {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	// 尝试相对于 BaseDir 读取
	full := filepath.Join(e.BaseDir, path)
	data, err := os.ReadFile(full)
	if err != nil {
		// 兜底再尝试原始路径
		data, err = os.ReadFile(path)
		if err != nil {
			return "", err
		}
	}
	return string(data), nil
}

// copyFile 拷贝源文件到目标路径。
func copyFile(src, dst string) error {
	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// toLowerSnake 将驼峰命名转为下划线小写命名。
// 例如 "UserName" -> "user_name"。
func toLowerSnake(s string) string {
	var buf strings.Builder
	for i, r := range s {
		// 大写字母前插入下划线（非首字符）
		if i > 0 && r >= 'A' && r <= 'Z' {
			buf.WriteByte('_')
		}
		// 转小写写入
		if r >= 'A' && r <= 'Z' {
			buf.WriteRune(r + 32)
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
