// Package metadata 负责项目元数据（.go-scaffold.json）的读写。
// 元数据记录项目生成时的配置参数，供后续 generate 命令读取技术栈上下文。
// 该元数据仅用于脚手架 CLI 后续读取上下文，不进入生成项目运行时。
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/go-scaffold/internal/errors"
)

// metadataFileName 是元数据文件在生成项目根目录下的文件名。
const metadataFileName = ".go-scaffold.json"

// ProjectMetadata 记录项目生成时的配置参数。
// JSON tag 用于序列化到 .go-scaffold.json 文件。
type ProjectMetadata struct {
	ProjectName   string    `json:"project_name"`            // 项目名称
	Backend       string    `json:"backend"`                 // 后端框架
	ORM           string    `json:"orm"`                     // ORM 框架
	Database      string    `json:"database"`                // 数据库类型
	Frontend      string    `json:"frontend"`                // 前端框架
	ModulePath    string    `json:"module_path"`             // Go 模块路径
	EnableJWT     bool      `json:"enable_jwt"`              // 是否启用 JWT 认证
	EnableDocker  bool      `json:"enable_docker"`           // 是否生成 Docker 配置
	EnableCI      bool      `json:"enable_ci"`               // 是否生成 CI 配置
	GeneratedAt   time.Time `json:"generated_at"`            // 生成时间
}

// LoadMetadata 从项目根目录读取 .go-scaffold.json 元数据文件。
// projectDir 为生成项目的根目录路径。
// 若文件不存在则返回 NotFound 错误。
func LoadMetadata(projectDir string) (*ProjectMetadata, error) {
	// 拼接元数据文件路径
	path := filepath.Join(projectDir, metadataFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 元数据文件不存在，返回分类错误
			return nil, errors.NewScaffoldError(
				errors.CodeNotFound,
				fmt.Sprintf("未找到项目元数据文件 %s，请在已初始化的项目目录下执行该命令", metadataFileName),
				err,
			)
		}
		return nil, errors.NewScaffoldError(
			errors.CodeInternal,
			"读取项目元数据文件失败",
			err,
		)
	}
	// 反序列化 JSON
	meta := &ProjectMetadata{}
	if err := json.Unmarshal(data, meta); err != nil {
		return nil, errors.NewScaffoldError(
			errors.CodeInvalidOutput,
			"项目元数据文件格式错误，无法解析",
			err,
		)
	}
	return meta, nil
}

// SaveMetadata 将项目元数据写入 projectDir 目录下的 .go-scaffold.json。
// projectDir 为生成项目的根目录路径。
func SaveMetadata(projectDir string, meta *ProjectMetadata) error {
	// 序列化为带缩进的 JSON
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return errors.NewScaffoldError(
			errors.CodeInternal,
			"序列化项目元数据失败",
			err,
		)
	}
	// 拼接写入路径
	path := filepath.Join(projectDir, metadataFileName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return errors.NewScaffoldError(
			errors.CodeInternal,
			"写入项目元数据文件失败",
			err,
		)
	}
	return nil
}
