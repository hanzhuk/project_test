// Package config 负责脚手架 CLI 自身的全局配置管理。
// 配置来源优先级为：命令行参数 > 环境变量 > 配置文件。
// 默认配置文件路径为用户主目录下的 ~/.go-scaffold/config.yaml。
// 注意：该配置仅作用于脚手架工具运行时，与生成项目的 .env 配置无关。
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config 定义脚手架 CLI 自身的全局配置。
// 这些字段记录用户偏好和默认选项，供 init/generate 命令使用。
type Config struct {
	DefaultBackend  string `mapstructure:"default_backend"`  // 默认后端框架
	DefaultORM      string `mapstructure:"default_orm"`       // 默认 ORM
	DefaultDatabase string `mapstructure:"default_database"`  // 默认数据库
	DefaultFrontend string `mapstructure:"default_frontend"`  // 默认前端框架
	OllamaHost      string `mapstructure:"ollama_host"`        // 本地 Ollama 地址
	OllamaModel     string `mapstructure:"ollama_model"`       // 默认模型，如 ornith:9b
	LogLevel        string `mapstructure:"log_level"`          // 脚手架日志级别
	ModulePrefix    string `mapstructure:"module_prefix"`       // 生成项目模块前缀
}

// DefaultConfig 返回带有默认值的配置实例。
// 这些默认值对应需求文档中默认技术栈：Echo + Ent + PostgreSQL + React。
func DefaultConfig() *Config {
	return &Config{
		DefaultBackend:  "echo",
		DefaultORM:      "ent",
		DefaultDatabase: "postgres",
		DefaultFrontend: "react",
		OllamaHost:      "http://localhost:11434",
		OllamaModel:     "ornith:9b",
		LogLevel:        "info",
		ModulePrefix:    "github.com/example",
	}
}

// configDir 返回脚手架配置目录路径（~/.go-scaffold）。
func configDir() (string, error) {
	// 获取用户主目录
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".go-scaffold"), nil
}

// LoadConfig 从配置文件、环境变量加载配置。
// paths 为可选的配置文件路径，若不提供则使用默认路径 ~/.go-scaffold/config.yaml。
// 若配置文件不存在，返回默认配置而不报错。
func LoadConfig(paths ...string) (*Config, error) {
	v := viper.New()
	// 设置默认值
	cfg := DefaultConfig()
	v.SetDefault("default_backend", cfg.DefaultBackend)
	v.SetDefault("default_orm", cfg.DefaultORM)
	v.SetDefault("default_database", cfg.DefaultDatabase)
	v.SetDefault("default_frontend", cfg.DefaultFrontend)
	v.SetDefault("ollama_host", cfg.OllamaHost)
	v.SetDefault("ollama_model", cfg.OllamaModel)
	v.SetDefault("log_level", cfg.LogLevel)
	v.SetDefault("module_prefix", cfg.ModulePrefix)

	// 支持环境变量覆盖，前缀为 GO_SCAFFOLD_
	v.SetEnvPrefix("go_scaffold")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 确定配置文件路径
	configPath := ""
	if len(paths) > 0 && paths[0] != "" {
		configPath = paths[0]
	} else {
		// 使用默认路径
		dir, err := configDir()
		if err == nil {
			configPath = filepath.Join(dir, "config.yaml")
		}
	}

	// 若配置文件存在则读取
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			// 配置文件不存在时使用默认值，其他错误才返回
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	// 反序列化到 Config 结构
	result := &Config{}
	if err := v.Unmarshal(result); err != nil {
		return nil, err
	}
	// 保证 ModulePrefix 不为空
	if result.ModulePrefix == "" {
		result.ModulePrefix = cfg.ModulePrefix
	}
	return result, nil
}

// SaveConfig 将配置保存到指定路径的 YAML 文件。
// 若路径为空，则保存到默认路径 ~/.go-scaffold/config.yaml。
func SaveConfig(cfg *Config, path string) error {
	v := viper.New()
	v.Set("default_backend", cfg.DefaultBackend)
	v.Set("default_orm", cfg.DefaultORM)
	v.Set("default_database", cfg.DefaultDatabase)
	v.Set("default_frontend", cfg.DefaultFrontend)
	v.Set("ollama_host", cfg.OllamaHost)
	v.Set("ollama_model", cfg.OllamaModel)
	v.Set("log_level", cfg.LogLevel)
	v.Set("module_prefix", cfg.ModulePrefix)

	// 确定保存路径
	if path == "" {
		dir, err := configDir()
		if err != nil {
			return err
		}
		// 确保目录存在
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		path = filepath.Join(dir, "config.yaml")
	}
	v.SetConfigType("yaml")
	return v.WriteConfigAs(path)
}

// MergeConfig 合并多个配置源。
// 优先级从低到高：base 在最底层，后续 overrides 依次覆盖非空字段。
// 仅当 override 中字符串字段非空时才覆盖 base 的对应字段。
func MergeConfig(base *Config, overrides ...*Config) *Config {
	merged := &Config{
		DefaultBackend:  base.DefaultBackend,
		DefaultORM:      base.DefaultORM,
		DefaultDatabase: base.DefaultDatabase,
		DefaultFrontend: base.DefaultFrontend,
		OllamaHost:      base.OllamaHost,
		OllamaModel:     base.OllamaModel,
		LogLevel:        base.LogLevel,
		ModulePrefix:    base.ModulePrefix,
	}
	// 依次应用覆盖配置
	for _, o := range overrides {
		if o == nil {
			continue
		}
		if o.DefaultBackend != "" {
			merged.DefaultBackend = o.DefaultBackend
		}
		if o.DefaultORM != "" {
			merged.DefaultORM = o.DefaultORM
		}
		if o.DefaultDatabase != "" {
			merged.DefaultDatabase = o.DefaultDatabase
		}
		if o.DefaultFrontend != "" {
			merged.DefaultFrontend = o.DefaultFrontend
		}
		if o.OllamaHost != "" {
			merged.OllamaHost = o.OllamaHost
		}
		if o.OllamaModel != "" {
			merged.OllamaModel = o.OllamaModel
		}
		if o.LogLevel != "" {
			merged.LogLevel = o.LogLevel
		}
		if o.ModulePrefix != "" {
			merged.ModulePrefix = o.ModulePrefix
		}
	}
	return merged
}
