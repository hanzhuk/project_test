package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/example/go-scaffold/internal/ai"
	"github.com/example/go-scaffold/internal/log"
	"github.com/example/go-scaffold/internal/metadata"
	"github.com/example/go-scaffold/internal/tui"
)

// newGenerateCmd 创建 generate 命令。
// generate 命令基于自然语言描述，调用本地 Ollama 模型生成业务代码。
// 必须在已初始化的项目目录下执行（需存在 .go-scaffold.json 元数据）。
// 采用 TUI 交互式界面收集生成所需信息（描述、类型、实体、字段等）。
func newGenerateCmd() *cobra.Command {
	var (
		model  string
		dryRun bool
		noTUI  bool
	)

	cmd := &cobra.Command{
		Use:   "generate [description] [flags]",
		Short: "基于自然语言生成代码",
		Long: `基于自然语言描述，调用本地 Ollama 模型生成 Go 业务代码。

采用 Tool Calling 协议：模型通过 write_file 工具返回文件写入指令，
脚手架解析后执行写入并通过 go/ast 校验语法。

执行 generate 后进入 TUI 代码生成表单界面，收集描述、代码类型、
实体名称、字段列表等信息；若命令行提供描述则预填充。

必须在已初始化的项目目录下执行（需存在 .go-scaffold.json 元数据），
技术栈上下文从元数据自动读取，无需重复指定。

示例：
  go-scaffold generate                    # 启动 TUI 表单
  go-scaffold generate "user CRUD"        # 预填充描述进入 TUI
  go-scaffold generate "todo CRUD" --no-tui --dry-run  # 跳过 TUI 直接生成`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// 获取命令行描述（预填充用）
			description := ""
			if len(args) > 0 {
				description = args[0]
			}

			// 读取当前目录的项目元数据
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("获取当前目录失败: %w", err)
			}
			meta, err := metadata.LoadMetadata(cwd)
			if err != nil {
				return fmt.Errorf("请在已初始化的项目目录下执行 generate 命令: %w", err)
			}

			// 解析模型名称
			useModel := resolveOllamaModel(model)
			ollamaHost := resolveOllamaHost("")

			// 创建 Ollama 客户端
			client := ai.NewOllamaClient(ollamaHost)

			// 检查 Ollama 服务连通性
			ctx := context.Background()
			if err := client.Ping(ctx); err != nil {
				return fmt.Errorf("Ollama 服务不可用 (%s): %w\n请确认已运行: ollama serve", ollamaHost, err)
			}

			// 构建生成输入
			var input *ai.GenerateInput
			// 决定是否使用 TUI 表单（默认使用，--no-tui 跳过）
			if !noTUI {
				// 启动 TUI 代码生成表单，预填充描述
				log.Info("启动 TUI 代码生成表单")
				result, err := tui.RunGenerateForm(description)
				if err != nil {
					return fmt.Errorf("TUI 表单运行失败: %w", err)
				}
				if result.Cancel {
					fmt.Println("已取消代码生成")
					return nil
				}
				input = result.Input
				// TUI 中点击预览等价于 --dry-run
				if result.IsDryRun {
					dryRun = true
				}
			} else {
				// 非 TUI 模式，从命令行描述构建输入
				input = &ai.GenerateInput{
					Description: description,
					CodeType:    "auto",
				}
				if entity := parseEntityFromDescription(description); entity != "" {
					input.Entity = entity
				}
			}

			// 校验描述非空
			if input.Description == "" {
				return fmt.Errorf("描述不能为空，请输入需要生成的代码需求")
			}

			// 创建代码生成器
			generator := ai.NewCodeGenerator(client, useModel)

			// 带重试生成代码
			log.Info("开始 AI 代码生成",
				"model", useModel,
				"description", input.Description,
				"entity", input.Entity,
				"code_type", input.CodeType,
				"project", meta.ProjectName,
			)
			code, err := generator.GenerateStepByStep(ctx, input, meta, nil)
			if err != nil {
				return err
			}

			// dry-run 模式仅预览不写入
			if dryRun {
				printPreview(code)
				return nil
			}

			// 写入文件
			written, err := writeGeneratedFiles(cwd, code)
			if err != nil {
				return err
			}

			// 自动注册路由到 handler.go
			if input.Entity != "" {
				if err := injectRoutes(cwd, input.Entity); err != nil {
					log.Warn("自动注册路由失败，请手动添加", "err", err)
				}
			}

			// 修复 TSX 文件的 import 问题
			for _, p := range written {
				if strings.HasSuffix(p, ".tsx") {
					fullPath := filepath.Join(cwd, p)
					if err := fixTsxImports(fullPath); err != nil {
						log.Warn("修复 TSX import 失败", "path", p, "err", err)
					}
				}
			}

			// 输出成功信息
			printGenerateSuccess(input, code, written)
			return nil
		},
	}

	// 注册命令 flag
	cmd.Flags().StringVar(&model, "model", "", "指定使用的 Ollama 模型，默认 ornith:9b")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "仅显示生成结果，不写入文件")
	cmd.Flags().BoolVar(&noTUI, "no-tui", false, "跳过 TUI 表单，直接使用命令行描述生成")

	return cmd
}

// parseEntityFromDescription 从描述中尝试解析实体名称。
// 简单策略：取描述中第一个独立英文单词作为实体名。
func parseEntityFromDescription(desc string) string {
	if desc == "" {
		return ""
	}
	// 按空格分割，寻找首字母大写或全小写的英文单词
	words := strings.Fields(desc)
	for _, w := range words {
		// 清理标点
		w = strings.Trim(w, "，。,.!?；;：:、")
		if w == "" {
			continue
		}
		// 检查是否为纯字母单词
		isAlpha := true
		for _, r := range w {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
				isAlpha = false
				break
			}
		}
		if isAlpha && len(w) > 1 {
			// 首字母大写返回
			return strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return ""
}

// writeGeneratedFiles 将生成的代码文件写入磁盘。
// projectDir 为项目根目录，code 为生成的代码。
// 返回已写入的文件路径列表。
func writeGeneratedFiles(projectDir string, code *ai.GeneratedCode) ([]string, error) {
	var written []string
	for _, f := range code.Files {
		// 拼接完整路径
		fullPath := filepath.Join(projectDir, f.Path)
		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return written, fmt.Errorf("创建目录失败 %s: %w", filepath.Dir(fullPath), err)
		}
		// 写入文件
		if err := os.WriteFile(fullPath, []byte(f.Content), 0o644); err != nil {
			return written, fmt.Errorf("写入文件失败 %s: %w", fullPath, err)
		}
		written = append(written, f.Path)
		log.Info("已写入文件", "path", f.Path)
	}
	return written, nil
}

// printPreview 在 dry-run 模式下预览生成的代码。
func printPreview(code *ai.GeneratedCode) {
	fmt.Printf("\n=== 预览模式（--dry-run，未写入文件）===\n")
	fmt.Printf("共生成 %d 个文件:\n\n", len(code.Files))
	for _, f := range code.Files {
		fmt.Printf("--- 文件: %s ---\n", f.Path)
		fmt.Println(f.Content)
		fmt.Println()
	}
}

// printGenerateSuccess 输出生成成功信息。
func printGenerateSuccess(input *ai.GenerateInput, code *ai.GeneratedCode, written []string) {
	fmt.Printf("\n✓ 代码生成成功\n")
	if input.Entity != "" {
		fmt.Printf("  实体: %s\n", input.Entity)
	}
	fmt.Printf("  类型: %s\n", input.CodeType)
	if len(written) > 0 {
		fmt.Printf("  输出: %s\n", written[0])
	}
	fmt.Printf("\n已生成文件:\n")
	for _, p := range written {
		fmt.Printf("  - %s\n", p)
	}
	fmt.Printf("\n后续步骤:\n")
	fmt.Printf("  1. 编译 Ent Schema 客户端 (使用 Ent ORM 时):\n")
	fmt.Printf("     go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema\n")
	fmt.Printf("  2. 整理并下载包依赖:\n")
	fmt.Printf("     go mod tidy\n")
	fmt.Printf("  3. 编译校验:\n")
	fmt.Printf("     go build ./...\n")
}
