package mgredis

import (
	"context"
	"testing"
	"time"
)

// TestOpener 测试opener函数
func TestOpener(t *testing.T) {
	ctx := context.Background()

	t.Run("成功创建连接", func(t *testing.T) {
		// 如果本地没有Redis服务，跳过此测试
		cfg := RedisConfig{
			Addr: "127.0.0.1:6379",
			DB:   0,
		}

		client, err := opener(ctx, cfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}
		defer client.Close()

		// 验证客户端可用
		if err := client.Ping(ctx).Err(); err != nil {
			t.Errorf("Ping失败: %v", err)
		}
	})

	t.Run("配置缺少地址", func(t *testing.T) {
		cfg := RedisConfig{
			// Addr为空
			DB: 0,
		}

		client, err := opener(ctx, cfg)
		if err == nil {
			if client != nil {
				client.Close()
			}
			t.Fatal("预期返回错误，但成功创建了客户端")
		}

		if !IsErrNoAddr(err) {
			t.Errorf("预期ErrNoAddr错误，实际得到: %v", err)
		}
	})

	t.Run("连接失败_错误的地址", func(t *testing.T) {
		cfg := RedisConfig{
			Addr:        "127.0.0.1:9999", // 不存在的端口
			DialTimeout: 1 * time.Second,
		}

		client, err := opener(ctx, cfg)
		if err == nil {
			if client != nil {
				client.Close()
			}
			t.Fatal("预期返回连接错误，但成功创建了客户端")
		}

		if !IsErrPingFailed(err) {
			t.Errorf("预期ErrPingFailed错误，实际得到: %v", err)
		}
	})

	t.Run("配置默认值设置", func(t *testing.T) {
		cfg := RedisConfig{
			Addr: "127.0.0.1:6379",
		}

		// 检查默认值设置
		if err := cfg.CheckAndSetDefaults(); err != nil {
			t.Fatalf("CheckAndSetDefaults失败: %v", err)
		}

		// 验证默认值
		if cfg.PoolSize != 10 {
			t.Errorf("预期PoolSize为10，实际为%d", cfg.PoolSize)
		}
		if cfg.MinIdleConns != 2 {
			t.Errorf("预期MinIdleConns为2，实际为%d", cfg.MinIdleConns)
		}
		if cfg.DialTimeout != 5*time.Second {
			t.Errorf("预期DialTimeout为5s，实际为%v", cfg.DialTimeout)
		}
		if cfg.ReadTimeout != 3*time.Second {
			t.Errorf("预期ReadTimeout为3s，实际为%v", cfg.ReadTimeout)
		}
		if cfg.WriteTimeout != 3*time.Second {
			t.Errorf("预期WriteTimeout为3s，实际为%v", cfg.WriteTimeout)
		}
		if cfg.MaxRetries != 3 {
			t.Errorf("预期MaxRetries为3，实际为%d", cfg.MaxRetries)
		}
		if cfg.PoolTimeout != 4*time.Second {
			t.Errorf("预期PoolTimeout为4s，实际为%v", cfg.PoolTimeout)
		}
		if cfg.IdleTimeout != 5*time.Minute {
			t.Errorf("预期IdleTimeout为5m，实际为%v", cfg.IdleTimeout)
		}
	})

	t.Run("自定义配置不被覆盖", func(t *testing.T) {
		cfg := RedisConfig{
			Addr:         "127.0.0.1:6379",
			PoolSize:     20,
			MinIdleConns: 5,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			MaxRetries:   5,
			PoolTimeout:  8 * time.Second,
			IdleTimeout:  10 * time.Minute,
		}

		// 检查默认值设置
		if err := cfg.CheckAndSetDefaults(); err != nil {
			t.Fatalf("CheckAndSetDefaults失败: %v", err)
		}

		// 验证自定义值不被覆盖
		if cfg.PoolSize != 20 {
			t.Errorf("预期PoolSize为20，实际为%d", cfg.PoolSize)
		}
		if cfg.MinIdleConns != 5 {
			t.Errorf("预期MinIdleConns为5，实际为%d", cfg.MinIdleConns)
		}
		if cfg.DialTimeout != 10*time.Second {
			t.Errorf("预期DialTimeout为10s，实际为%v", cfg.DialTimeout)
		}
	})
}

// TestCloser 测试closer函数
func TestCloser(t *testing.T) {
	ctx := context.Background()

	t.Run("关闭正常客户端", func(t *testing.T) {
		cfg := RedisConfig{
			Addr: "127.0.0.1:6379",
			DB:   0,
		}

		client, err := opener(ctx, cfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		// 测试关闭
		err = closer(ctx, client)
		if err != nil {
			t.Errorf("关闭客户端失败: %v", err)
		}
	})

	t.Run("关闭nil客户端", func(t *testing.T) {
		err := closer(ctx, nil)
		if err != nil {
			t.Errorf("关闭nil客户端应该返回nil，实际返回: %v", err)
		}
	})
}

// TestNew 测试New函数
func TestNew(t *testing.T) {
	ctx := context.Background()

	t.Run("创建单组管理器", func(t *testing.T) {
		group := New()
		if group == nil {
			t.Fatal("预期创建成功，但返回nil")
		}
		defer group.Close(ctx)
	})

	t.Run("注册和获取客户端", func(t *testing.T) {
		group := New()
		defer group.Close(ctx)

		cfg := RedisConfig{
			Name: "测试缓存",
			Addr: "127.0.0.1:6379",
			DB:   0,
		}

		// 注册客户端
		_, err := group.Register(ctx, "cache", cfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		// 获取客户端
		client, err := group.Get(ctx, "cache")
		if err != nil {
			t.Fatalf("获取客户端失败: %v", err)
		}

		if client == nil {
			t.Fatal("预期返回客户端，但得到nil")
		}

		// 再次获取应该返回同一个客户端实例
		client2, err := group.Get(ctx, "cache")
		if err != nil {
			t.Fatalf("获取客户端失败: %v", err)
		}

		if client != client2 {
			t.Error("预期获取同一个客户端实例")
		}

		// 验证客户端可用
		if err := client.Ping(ctx).Err(); err != nil {
			t.Errorf("Ping失败: %v", err)
		}
	})

	t.Run("注册失败_无效配置", func(t *testing.T) {
		group := New()
		defer group.Close(ctx)

		cfg := RedisConfig{
			// Addr为空，应该失败
			DB: 0,
		}

		// Register 采用延迟初始化，不会立即验证配置
		_, err := group.Register(ctx, "invalid", cfg)
		if err != nil {
			t.Fatalf("Register不应该失败: %v", err)
		}

		// Get 时才会真正创建连接并验证配置
		_, err = group.Get(ctx, "invalid")
		if err == nil {
			t.Fatal("预期Get时返回错误，但成功了")
		}

		if !IsErrNoAddr(err) {
			t.Errorf("预期ErrNoAddr错误，实际得到: %v", err)
		}
	})

	t.Run("注销客户端", func(t *testing.T) {
		group := New()
		defer group.Close(ctx)

		cfg := RedisConfig{
			Addr: "127.0.0.1:6379",
			DB:   0,
		}

		// 注册客户端
		_, err := group.Register(ctx, "temp", cfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		// 注销客户端
		err = group.Unregister(ctx, "temp")
		if err != nil {
			t.Errorf("注销客户端失败: %v", err)
		}

		// 再次获取应该失败
		_, err = group.Get(ctx, "temp")
		if err == nil {
			t.Error("预期获取已注销的客户端失败，但成功了")
		}
	})

	t.Run("MustGet存在的客户端", func(t *testing.T) {
		group := New()
		defer group.Close(ctx)

		cfg := RedisConfig{
			Addr: "127.0.0.1:6379",
			DB:   0,
		}

		// 注册客户端
		_, err := group.Register(ctx, "cache", cfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		// MustGet应该成功
		client := group.MustGet(ctx, "cache")
		if client == nil {
			t.Fatal("MustGet返回nil")
		}

		// 验证客户端可用
		if err := client.Ping(ctx).Err(); err != nil {
			t.Errorf("Ping失败: %v", err)
		}
	})

	t.Run("关闭所有客户端", func(t *testing.T) {
		group := New()

		cfg := RedisConfig{
			Addr: "127.0.0.1:6379",
			DB:   0,
		}

		// 注册多个客户端
		_, err := group.Register(ctx, "cache1", cfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		cfg.DB = 1
		_, err = group.Register(ctx, "cache2", cfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		// 关闭所有客户端
		errs := group.Close(ctx)
		if len(errs) > 0 {
			t.Errorf("关闭所有客户端失败: %v", errs)
		}
	})
}

// TestNewManager 测试NewManager函数
func TestNewManager(t *testing.T) {
	ctx := context.Background()

	t.Run("创建多组管理器", func(t *testing.T) {
		manager := NewManager()
		if manager == nil {
			t.Fatal("预期创建成功，但返回nil")
		}
		defer manager.Close(ctx)
	})

	t.Run("添加和获取组", func(t *testing.T) {
		manager := NewManager()
		defer manager.Close(ctx)

		// 添加组
		manager.AddGroup("session-cache")
		manager.AddGroup("rate-limiter")

		// 获取组
		group1, err := manager.Group("session-cache")
		if err != nil {
			t.Fatalf("获取组失败: %v", err)
		}
		if group1 == nil {
			t.Fatal("预期返回组，但得到nil")
		}

		group2, err := manager.Group("rate-limiter")
		if err != nil {
			t.Fatalf("获取组失败: %v", err)
		}
		if group2 == nil {
			t.Fatal("预期返回组，但得到nil")
		}
	})

	t.Run("列出组名", func(t *testing.T) {
		manager := NewManager()
		defer manager.Close(ctx)

		// 添加多个组
		manager.AddGroup("group1")
		manager.AddGroup("group2")
		manager.AddGroup("group3")

		// 列出组名
		groupNames := manager.ListGroupNames()
		if len(groupNames) != 3 {
			t.Errorf("预期3个组，实际有%d个", len(groupNames))
		}

		// 验证组名存在
		nameMap := make(map[string]bool)
		for _, name := range groupNames {
			nameMap[name] = true
		}

		if !nameMap["group1"] || !nameMap["group2"] || !nameMap["group3"] {
			t.Errorf("组名不完整: %v", groupNames)
		}
	})

	t.Run("在不同组中注册客户端", func(t *testing.T) {
		manager := NewManager()
		defer manager.Close(ctx)

		// 添加组
		manager.AddGroup("session-cache")
		manager.AddGroup("rate-limiter")

		// 获取组
		sessionGroup, _ := manager.Group("session-cache")
		rlGroup, _ := manager.Group("rate-limiter")

		// 在不同组中注册客户端
		cfg1 := RedisConfig{
			Name: "会话缓存",
			Addr: "127.0.0.1:6379",
			DB:   1,
		}
		_, err := sessionGroup.Register(ctx, "primary", cfg1)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		cfg2 := RedisConfig{
			Name: "限流器",
			Addr: "127.0.0.1:6379",
			DB:   2,
		}
		_, err = rlGroup.Register(ctx, "primary", cfg2)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		// 获取客户端
		sessionRedis, err := sessionGroup.Get(ctx, "primary")
		if err != nil {
			t.Fatalf("获取会话Redis客户端失败: %v", err)
		}

		rateRedis, err := rlGroup.Get(ctx, "primary")
		if err != nil {
			t.Fatalf("获取限流Redis客户端失败: %v", err)
		}

		// 验证两个客户端不同
		if sessionRedis == rateRedis {
			t.Error("预期不同的客户端实例")
		}
	})

	t.Run("关闭所有组", func(t *testing.T) {
		manager := NewManager()

		cfg := RedisConfig{
			Addr: "127.0.0.1:6379",
			DB:   0,
		}

		// 添加多个组并注册客户端
		manager.AddGroup("group1")
		group1, _ := manager.Group("group1")
		_, err := group1.Register(ctx, "client1", cfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		manager.AddGroup("group2")
		group2, _ := manager.Group("group2")
		cfg.DB = 1
		_, err = group2.Register(ctx, "client2", cfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		// 关闭所有组
		errs := manager.Close(ctx)
		if len(errs) > 0 {
			t.Errorf("关闭所有组失败: %v", errs)
		}
	})
}

// TestIntegration 集成测试
func TestIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("单组完整流程", func(t *testing.T) {
		group := New()
		defer group.Close(ctx)

		cfg := RedisConfig{
			Name: "测试缓存",
			Addr: "127.0.0.1:6379",
			DB:   0,
		}

		// 注册客户端
		_, err := group.Register(ctx, "cache", cfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		// 获取客户端
		client, err := group.Get(ctx, "cache")
		if err != nil {
			t.Fatalf("获取客户端失败: %v", err)
		}

		// 使用客户端
		testKey := "test:integration:key"
		testValue := "test-value"

		err = client.Set(ctx, testKey, testValue, time.Minute).Err()
		if err != nil {
			t.Fatalf("Set操作失败: %v", err)
		}

		val, err := client.Get(ctx, testKey).Result()
		if err != nil {
			t.Fatalf("Get操作失败: %v", err)
		}

		if val != testValue {
			t.Errorf("预期值为%s，实际为%s", testValue, val)
		}

		// 清理测试数据
		client.Del(ctx, testKey)
	})

	t.Run("多组完整流程", func(t *testing.T) {
		manager := NewManager()
		defer manager.Close(ctx)

		// 创建会话缓存组
		manager.AddGroup("session-cache")
		sessionGroup, _ := manager.Group("session-cache")

		sessionCfg := RedisConfig{
			Name: "会话缓存",
			Addr: "127.0.0.1:6379",
			DB:   1,
		}

		_, err := sessionGroup.Register(ctx, "primary", sessionCfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		// 获取会话客户端
		sessionClient, err := sessionGroup.Get(ctx, "primary")
		if err != nil {
			t.Fatalf("获取会话客户端失败: %v", err)
		}

		// 创建限流器组
		manager.AddGroup("rate-limiter")
		rlGroup, _ := manager.Group("rate-limiter")

		rlCfg := RedisConfig{
			Name: "限流器",
			Addr: "127.0.0.1:6379",
			DB:   2,
		}

		_, err = rlGroup.Register(ctx, "primary", rlCfg)
		if err != nil {
			t.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
			return
		}

		// 获取限流客户端
		rlClient, err := rlGroup.Get(ctx, "primary")
		if err != nil {
			t.Fatalf("获取限流客户端失败: %v", err)
		}

		// 使用不同的客户端
		sessionKey := "test:session:123"
		rateKey := "test:rate:api:123"

		err = sessionClient.Set(ctx, sessionKey, "user_data", 30*time.Minute).Err()
		if err != nil {
			t.Fatalf("设置会话数据失败: %v", err)
		}

		err = rlClient.Incr(ctx, rateKey).Err()
		if err != nil {
			t.Fatalf("增加限流计数失败: %v", err)
		}

		// 验证数据
		sessionVal, err := sessionClient.Get(ctx, sessionKey).Result()
		if err != nil {
			t.Fatalf("获取会话数据失败: %v", err)
		}
		if sessionVal != "user_data" {
			t.Errorf("会话数据不匹配")
		}

		rateVal, err := rlClient.Get(ctx, rateKey).Int()
		if err != nil {
			t.Fatalf("获取限流计数失败: %v", err)
		}
		if rateVal != 1 {
			t.Errorf("预期限流计数为1，实际为%d", rateVal)
		}

		// 清理测试数据
		sessionClient.Del(ctx, sessionKey)
		rlClient.Del(ctx, rateKey)
	})
}

// BenchmarkOpener 性能测试：创建客户端
func BenchmarkOpener(b *testing.B) {
	ctx := context.Background()
	cfg := RedisConfig{
		Addr: "127.0.0.1:6379",
		DB:   0,
	}

	// 先测试一次看是否可以连接
	client, err := opener(ctx, cfg)
	if err != nil {
		b.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
		return
	}
	client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client, err := opener(ctx, cfg)
		if err != nil {
			b.Fatalf("创建客户端失败: %v", err)
		}
		client.Close()
	}
}

// BenchmarkGroupRegister 性能测试：注册客户端
func BenchmarkGroupRegister(b *testing.B) {
	ctx := context.Background()
	cfg := RedisConfig{
		Addr: "127.0.0.1:6379",
		DB:   0,
	}

	// 先测试一次看是否可以连接
	group := New()
	_, err := group.Register(ctx, "test", cfg)
	if err != nil {
		b.Skipf("跳过测试: 无法连接到Redis服务器 (%v)", err)
		return
	}
	group.Close(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := New()
		_, _ = group.Register(ctx, "cache", cfg)
		group.Close(ctx)
	}
}

// ExampleNew 示例：创建单组管理器
func ExampleNew() {
	ctx := context.Background()

	// 创建单组管理器
	group := New()
	defer group.Close(ctx)

	// 注册Redis客户端
	_, _ = group.Register(ctx, "cache", RedisConfig{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})

	// 获取客户端使用
	client, _ := group.Get(ctx, "cache")
	_ = client.Ping(ctx).Err()
}

// ExampleNewManager 示例：创建多组管理器
func ExampleNewManager() {
	ctx := context.Background()

	// 创建多组管理器
	manager := NewManager()
	defer manager.Close(ctx)

	// 添加组
	manager.AddGroup("session-cache")
	sessionGroup, _ := manager.Group("session-cache")

	// 在组中注册客户端
	_, _ = sessionGroup.Register(ctx, "primary", RedisConfig{
		Addr: "127.0.0.1:6379",
		DB:   1,
	})

	// 使用客户端
	client, _ := sessionGroup.Get(ctx, "primary")
	_ = client.Ping(ctx).Err()
}
