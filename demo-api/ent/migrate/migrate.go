// Package migrate 提供数据库迁移功能。
// 在正式开发中，由 Ent 代码生成器自动生成。此文件为初始占位。
package migrate

import (
	"context"
	"database/sql"
)

// Schema 是数据库迁移管理器。
type Schema struct {
	db *sql.DB
}

// Option 是迁移选项函数类型。
type Option func(*MigrateOptions)

// MigrateOptions 是迁移选项集合。
type MigrateOptions struct {
	DropIndex  bool
	DropColumn bool
}

// WithDropIndex 设置迁移时是否删除不存在的索引。
func WithDropIndex(drop bool) Option {
	return func(o *MigrateOptions) {
		o.DropIndex = drop
	}
}

// WithDropColumn 设置迁移时是否删除不存在的列。
func WithDropColumn(drop bool) Option {
	return func(o *MigrateOptions) {
		o.DropColumn = drop
	}
}

// Create 执行数据库迁移，创建或更新表结构。
// 占位实现，实际行为由 Ent 代码生成后替换。
func (s *Schema) Create(ctx context.Context, opts ...Option) error {
	// 占位：正式生成后此方法由 Ent 自动实现
	return nil
}
