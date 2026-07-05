// Package ast 提供基于 Go 标准库 go/ast、go/parser 与 go/format 的语法树代码解析与精准修改工具。
// 它可以精准地对已有的 Go 源文件进行导包补全（Import Injection）和函数声明追加。
package ast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// Injector 结构体封装 AST 节点修改逻辑。
type Injector struct {
	Fset *token.FileSet
	Node *ast.File
}

// NewInjectorFromFile 从给定的 Go 源文件加载并解析 AST 语法树。
func NewInjectorFromFile(filePath string) (*Injector, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取 Go 文件失败: %w", err)
	}
	return NewInjectorFromSource(content)
}

// NewInjectorFromSource 从源代码字节流解析 AST 语法树。
func NewInjectorFromSource(source []byte) (*Injector, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("解析 Go AST 失败: %w", err)
	}
	return &Injector{
		Fset: fset,
		Node: node,
	}, nil
}

// AddImport 向 AST 语法树安全追加 import 声明（防重）。
func (ij *Injector) AddImport(importPath string) bool {
	// 校验是否已存在该 import
	for _, imp := range ij.Node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if path == importPath {
			return false // 已存在，无需注入
		}
	}

	// 构造新的 ImportSpec 节点
	newImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("%q", importPath),
		},
	}

	// 查找现有的 import 声明块，若无则插入
	var decl *ast.GenDecl
	for _, d := range ij.Node.Decls {
		if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
			decl = g
			break
		}
	}

	if decl != nil {
		decl.Specs = append(decl.Specs, newImport)
	} else {
		// 创建全新的 import 声明块
		newDecl := &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{newImport},
		}
		// 插入到 package 声明之后
		ij.Node.Decls = append([]ast.Decl{newDecl}, ij.Node.Decls...)
	}

	return true
}

// AddFunctionFromSource 解析一段 Go 函数声明源代码，并插入到 AST 语法树末尾。
func (ij *Injector) AddFunctionFromSource(fnCode string) error {
	// 将片段包裹在 dummy package 中进行解析
	dummySrc := fmt.Sprintf("package dummy\n\n%s", fnCode)
	dummyFset := token.NewFileSet()
	dummyNode, err := parser.ParseFile(dummyFset, "", dummySrc, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("解析欲插入的函数声明失败: %w", err)
	}

	for _, decl := range dummyNode.Decls {
		if fnDecl, ok := decl.(*ast.FuncDecl); ok {
			ij.Node.Decls = append(ij.Node.Decls, fnDecl)
		}
	}

	return nil
}

// Format 重新格式化 AST 语法树并输出完整的 Go 源代码。
func (ij *Injector) Format() ([]byte, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, ij.Fset, ij.Node); err != nil {
		return nil, fmt.Errorf("格式化 AST 输出失败: %w", err)
	}
	return buf.Bytes(), nil
}

// WriteToFile 将修改并格式化后的 AST 代码写回文件。
func (ij *Injector) WriteToFile(filePath string) error {
	formatted, err := ij.Format()
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, formatted, 0o644)
}

// ValidateGoSource 使用 go/parser 验证 Go 代码源码的 AST 语法正确性。
// 若语法有错（如缺少花括号、非法的语句结构），返回解析错误。
func ValidateGoSource(src []byte) error {
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("Go AST 语法校验失败: %w", err)
	}
	return nil
}
