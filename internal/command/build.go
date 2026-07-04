package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/example/go-scaffold/internal/build"
	"github.com/example/go-scaffold/internal/metadata"
)

// newBuildCmd 创建 build 命令。
// build 命令是脚手架 CLI 对生成项目的构建代理：
//   - 默认调用生成项目目录下的 make build；
//   - --docker 时调用 make docker-build；
//   - 脚手架自身不直接编译容器镜像。
func newBuildCmd() *cobra.Command {
	var (
		docker bool
		output string
		target string
	)

	cmd := &cobra.Command{
		Use:   "build [flags]",
		Short: "构建生成的项目",
		Long: `构建生成的项目。

作为脚手架 CLI 对生成项目的构建代理：
  - 默认调用生成项目目录下的 make build 编译二进制；
  - --docker 时调用 make docker-build 构建镜像；
  - 脚手架自身不直接编译容器镜像，也不管理运行时容器。

必须在已初始化的项目目录下执行。

示例：
  go-scaffold build
  go-scaffold build --docker
  go-scaffold build --output ./bin --target server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 获取当前目录作为项目目录
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("获取当前目录失败: %w", err)
			}

			// 校验是否在已初始化的项目目录下
			if _, err := metadata.LoadMetadata(cwd); err != nil {
				return fmt.Errorf("请在已初始化的项目目录下执行 build 命令: %w", err)
			}

			// 创建构建器
			builder := build.NewBuilder(cwd)

			// 构建选项
			opts := build.BuildOptions{
				Docker: docker,
				Output: output,
				Target: target,
			}

			// 执行构建
			fmt.Printf("开始构建项目: %s\n", cwd)
			if docker {
				fmt.Println("构建模式: Docker 镜像")
			} else {
				fmt.Printf("构建模式: 二进制 (目标: %s, 输出: %s)\n", target, output)
			}

			if err := builder.Build(opts); err != nil {
				return fmt.Errorf("构建失败: %w", err)
			}

			fmt.Println("\n✓ 项目构建成功")
			return nil
		},
	}

	// 注册命令 flag
	cmd.Flags().BoolVar(&docker, "docker", false, "是否在生成项目内执行 make docker-build")
	cmd.Flags().StringVar(&output, "output", "./bin", "二进制文件输出目录")
	cmd.Flags().StringVar(&target, "target", "server", "构建目标，可选值：server、web、all")

	return cmd
}
