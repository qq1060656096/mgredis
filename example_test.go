package mgredis_test

import (
	"context"
	"fmt"
	"time"

	"github.com/qq1060656096/mgredis"
)

// Example_singleGroup 单组管理示例
func Example_singleGroup() {
	ctx := context.Background()

	// 创建单组管理器
	group := mgredis.New()
	defer group.Close(ctx)

	// 注册Redis客户端
	_, err := group.Register(ctx, "cache", mgredis.RedisConfig{
		Name:         "主缓存",
		Addr:         "127.0.0.1:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	// 获取Redis客户端
	client, err := group.Get(ctx, "cache")
	if err != nil {
		panic(err)
	}

	// 使用Redis客户端
	err = client.Set(ctx, "key", "value", time.Minute).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
}

// Example_multipleGroups 多组管理示例
func Example_multipleGroups() {
	ctx := context.Background()

	// 创建多组管理器
	manager := mgredis.NewManager()
	defer manager.Close(ctx)

	// 添加会话缓存组
	manager.AddGroup("session-cache")
	sessionGroup, err := manager.Group("session-cache")
	if err != nil {
		panic(err)
	}

	// 注册会话Redis实例
	_, err = sessionGroup.Register(ctx, "primary", mgredis.RedisConfig{
		Name:     "会话缓存",
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1,
		PoolSize: 5,
	})
	if err != nil {
		panic(err)
	}

	// 添加限流器组
	manager.AddGroup("rate-limiter")
	rlGroup, err := manager.Group("rate-limiter")
	if err != nil {
		panic(err)
	}

	// 注册限流Redis实例
	_, err = rlGroup.Register(ctx, "primary", mgredis.RedisConfig{
		Name:     "限流器",
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       2,
		PoolSize: 5,
	})
	if err != nil {
		panic(err)
	}

	// 获取不同组的Redis客户端
	sessionRedis, _ := sessionGroup.Get(ctx, "primary")
	rateRedis, _ := rlGroup.Get(ctx, "primary")

	// 使用各自的客户端
	_ = sessionRedis.Set(ctx, "session:123", "user_data", 30*time.Minute).Err()
	_ = rateRedis.Incr(ctx, "rate:api:123").Err()

	// 列出所有组名
	groupNames := manager.ListGroupNames()
	fmt.Println("组名:", groupNames)
}

// Example_mustGet MustGet示例（如果不存在会panic）
func Example_mustGet() {
	ctx := context.Background()

	group := mgredis.New()
	defer group.Close(ctx)

	// 注册客户端
	_, _ = group.Register(ctx, "cache", mgredis.RedisConfig{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})

	// MustGet 如果不存在会panic
	client := group.MustGet(ctx, "cache")
	_ = client.Ping(ctx).Err()
}

// Example_unregister 注销示例
func Example_unregister() {
	ctx := context.Background()

	group := mgredis.New()
	defer group.Close(ctx)

	// 注册客户端
	_, _ = group.Register(ctx, "temp", mgredis.RedisConfig{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})

	// 使用后注销
	client, _ := group.Get(ctx, "temp")
	_ = client.Set(ctx, "key", "value", time.Second).Err()

	// 注销客户端（会关闭连接）
	err := group.Unregister(ctx, "temp")
	if err != nil {
		panic(err)
	}
}
