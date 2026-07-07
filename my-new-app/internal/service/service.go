// Package service 实现业务逻辑层。
// 它接收 handler 传递的参数，调用 repository 层访问数据库，
// 并返回处理结果。业务规则和逻辑在此层封装。
package service

import (
	"context"
	"log/slog"

	"github.com/example/my-new-app/ent"
)

// Service 是业务逻辑层，持有 Ent 客户端。
type Service struct {
	Client *ent.Client // Ent 数据库客户端
}

// New 创建 Service 实例。
func New(client *ent.Client) *Service {
	return &Service{Client: client}
}

// HealthCheck 检查服务健康状态。
func (s *Service) HealthCheck(ctx context.Context) error {
	slog.InfoContext(ctx, "执行健康检查")
	// 此处可添加数据库连通性检查
	return nil
}
