package mgredis

import (
	"context"
	"fmt"
	"time"

	"github.com/qq1060656096/bizutil/registry"
	"github.com/redis/go-redis/v9"
)

// opener 创建Redis客户端连接
func opener(ctx context.Context, cfg RedisConfig) (*redis.Client, error) {
	// 检查并设置默认值
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, err
	}

	// 创建Redis客户端选项
	opts := &redis.Options{
		Addr:            cfg.Addr,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		MaxRetries:      cfg.MaxRetries,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
	}

	// 创建客户端
	client := redis.NewClient(opts)

	// 使用超时上下文进行Ping测试
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx2).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("%w: %v", ErrPingFailed, err)
	}

	return client, nil
}

// closer 关闭Redis客户端连接
func closer(ctx context.Context, client *redis.Client) error {
	if client == nil {
		return nil
	}
	return client.Close()
}

// Group 是单一组管理（key => redis client）
type Group registry.Group[RedisConfig, *redis.Client]

// Manager 是多组管理
type Manager registry.Manager[RedisConfig, *redis.Client]

// New 创建单组Redis客户端管理器
// 用于管理多个命名的Redis客户端实例
// 支持惰性初始化（首次Get时创建）和安全关闭所有资源
func New() Group {
	return registry.NewGroup[RedisConfig, *redis.Client](opener, closer)
}

// NewManager 创建多组Redis客户端管理器
// 用于管理多个组，每个组可以包含多个命名的Redis客户端实例
// 适用于需要按业务场景分组管理Redis连接的复杂场景
func NewManager() Manager {
	return registry.New[RedisConfig, *redis.Client](opener, closer)
}
