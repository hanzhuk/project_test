// Package repository 封装数据库访问层。
// 使用 Ent ORM 创建数据库客户端，并提供数据访问操作。
package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/example/my-new-app/ent"
	"github.com/example/my-new-app/ent/migrate"

	"github.com/example/my-new-app/internal/config"

	_ "github.com/lib/pq" // PostgreSQL 驱动
)

// Repository 是数据访问层，持有 Ent 客户端。
type Repository struct {
	Client *ent.Client // Ent 数据库客户端
}

// New 创建数据访问层实例，初始化 Ent 客户端连接数据库。
func New(cfg *config.Config) (*Repository, error) {
	// 打开数据库连接
	client, err := ent.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}
	slog.Info("数据库连接成功", slog.String("db", cfg.DBName))
	return &Repository{Client: client}, nil
}

// Close 关闭数据库连接。
func (r *Repository) Close() {
	if r.Client != nil {
		if err := r.Client.Close(); err != nil {
			slog.Error("关闭数据库连接失败", slog.Any("err", err))
		}
	}
}

// GetClient 返回 Ent 数据库客户端，供 handler 层使用。
func (r *Repository) GetClient() *ent.Client {
	return r.Client
}

// AutoMigrate 执行数据库自动迁移，创建或更新表结构。
func (r *Repository) AutoMigrate(ctx context.Context) error {
	// 执行 Ent 自动迁移
	if err := r.Client.Schema.Create(ctx, migrate.WithDropIndex(true), migrate.WithDropColumn(true)); err != nil {
		return fmt.Errorf("执行数据库迁移失败: %w", err)
	}
	return nil
}
