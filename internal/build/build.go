// Package build 负责生成项目的构建与打包。
// 作为脚手架 CLI 对生成项目的构建代理，默认调用生成项目的 make build；
// --docker 时调用 make docker-build。脚手架自身不直接编译容器镜像。
package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/example/go-scaffold/internal/errors"
	"github.com/example/go-scaffold/internal/log"
)

// Builder 负责生成项目的构建与打包。
// ProjectDir 为生成项目的根目录路径。
type Builder struct {
	ProjectDir string // 生成项目根目录
}

// NewBuilder 创建构建器实例。
// projectDir 为生成项目的根目录路径。
func NewBuilder(projectDir string) *Builder {
	return &Builder{ProjectDir: projectDir}
}

// BuildOptions 定义构建选项。
type BuildOptions struct {
	Docker  bool   // 是否执行 Docker 构建
	Output  string // 二进制文件输出目录，默认 ./bin
	Target  string // 构建目标：server、web、all
}

// Build 执行项目构建。
// 根据 opts.Docker 决定调用 make build 或 make docker-build。
// 构建过程中输出编译日志，失败时返回详细错误信息。
func (b *Builder) Build(opts BuildOptions) error {
	// 校验项目目录存在
	if _, err := os.Stat(b.ProjectDir); err != nil {
		return errors.NewScaffoldError(
			errors.CodeNotFound,
			fmt.Sprintf("项目目录不存在: %s", b.ProjectDir),
			err,
		)
	}
	// 设置默认输出目录
	if opts.Output == "" {
		opts.Output = "./bin"
	}
	// 设置默认目标
	if opts.Target == "" {
		opts.Target = "server"
	}

	if opts.Docker {
		// 执行 Docker 构建
		return b.buildDocker()
	}
	// 执行普通构建
	return b.buildBinary(opts)
}

// buildBinary 调用 make build 或直接 go build 编译二进制。
func (b *Builder) buildBinary(opts BuildOptions) error {
	log.Info("开始构建项目", "dir", b.ProjectDir, "target", opts.Target)

	// 优先尝试使用 Makefile
	makefilePath := filepath.Join(b.ProjectDir, "Makefile")
	if _, err := os.Stat(makefilePath); err == nil {
		// Makefile 存在，调用 make build
		return b.runCommand("make", "build")
	}

	// 无 Makefile 时直接调用 go build
	switch opts.Target {
	case "server", "all":
		// 构建 server 二进制
		if err := b.buildServer(opts.Output); err != nil {
			return err
		}
	case "web":
		// 前端构建
		return b.buildWeb()
	}
	log.Info("项目构建完成", "dir", b.ProjectDir)
	return nil
}

// buildServer 编译后端 server 二进制文件。
func (b *Builder) buildServer(output string) error {
	// 确保输出目录存在
	absOutput := output
	if !filepath.IsAbs(absOutput) {
		absOutput = filepath.Join(b.ProjectDir, output)
	}
	if err := os.MkdirAll(absOutput, 0o755); err != nil {
		return errors.NewScaffoldError(
			errors.CodeInternal,
			"创建输出目录失败",
			err,
		)
	}
	// 使用 -ldflags "-s -w" 优化体积
	binaryName := "server"
	binaryPath := filepath.Join(absOutput, binaryName)
	cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", binaryPath, "./cmd/server")
	cmd.Dir = b.ProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.NewScaffoldError(
			errors.CodeInternal,
			"编译后端二进制失败",
			err,
		)
	}
	log.Info("后端二进制构建完成", "path", binaryPath)
	return nil
}

// buildWeb 构建前端项目。
func (b *Builder) buildWeb() error {
	webDir := filepath.Join(b.ProjectDir, "web")
	if _, err := os.Stat(webDir); err != nil {
		// 无前端目录则跳过
		log.Info("未发现前端目录，跳过前端构建")
		return nil
	}
	return b.runCommandDir("npm", []string{"run", "build"}, webDir)
}

// buildDocker 调用 make docker-build 构建 Docker 镜像。
func (b *Builder) buildDocker() error {
	log.Info("开始构建 Docker 镜像", "dir", b.ProjectDir)
	makefilePath := filepath.Join(b.ProjectDir, "Makefile")
	if _, err := os.Stat(makefilePath); err == nil {
		return b.runCommand("make", "docker-build")
	}
	// 无 Makefile 时直接调用 docker build
	dockerfilePath := filepath.Join(b.ProjectDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err != nil {
		return errors.NewScaffoldError(
			errors.CodeNotFound,
			"项目未包含 Makefile 或 Dockerfile，无法执行 Docker 构建",
			err,
		)
	}
	return b.runCommand("docker", "build", "-t", filepath.Base(b.ProjectDir)+":latest", ".")
}

// runCommand 在项目目录下执行命令。
func (b *Builder) runCommand(name string, args ...string) error {
	return b.runCommandDir(name, args, b.ProjectDir)
}

// runCommandDir 在指定目录下执行命令。
func (b *Builder) runCommandDir(name string, args []string, dir string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.NewScaffoldError(
			errors.CodeInternal,
			fmt.Sprintf("执行命令失败: %s %s", name, fmt.Sprint(args)),
			err,
		)
	}
	return nil
}
