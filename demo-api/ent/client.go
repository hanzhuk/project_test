// Package ent 提供 Ent ORM 客户端的基础定义。
// 在正式开发中，应在 schema/ 下定义实体后运行 `go generate ./ent`
// 自动生成完整的客户端代码。此文件为初始占位，保证项目可编译。
package ent

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/example/demo-api/ent/migrate"

	_ "github.com/lib/pq" // PostgreSQL 驱动
)

// Client 是 Ent ORM 的数据库客户端。
// 正式使用时由 Ent 代码生成器自动生成，此处为最小占位实现。
type Client struct {
	db     *sql.DB
	Schema *Schema
}

// Schema 提供数据库 Schema 管理功能。
type Schema struct {
	db *sql.DB
}

// Create 执行数据库迁移，创建或更新表结构。
func (s *Schema) Create(ctx context.Context, opts ...migrate.Option) error {
	// 占位：正式生成后此方法由 Ent 自动实现
	return nil
}

// Open 打开数据库连接并返回 Ent 客户端。
func Open(driverName, dataSourceName string) (*Client, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}
	return &Client{
		db:     db,
		Schema: &Schema{db: db},
	}, nil
}

// Close 关闭数据库连接。
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

