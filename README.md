# mgredis

基于 `bizutil/registry` 的 Redis 客户端管理器，提供统一的连接管理、惰性初始化和生命周期管理。

## 特性

- 🚀 **惰性初始化** - 首次 Get 时才创建连接
- 📦 **单组/多组管理** - 支持简单场景的单组管理和复杂场景的多组管理
- 🔒 **线程安全** - 底层使用 bizutil/registry 保证并发安全
- ⚙️ **灵活配置** - 支持连接池、超时、重试等完整配置
- 🎯 **类型安全** - 使用泛型提供类型安全的 API
- 🔌 **自动连接测试** - 创建连接时自动 Ping 测试
- 🧹 **优雅关闭** - 统一管理所有连接的生命周期

## 安装

```bash
go get github.com/qq1060656096/mgredis
```

依赖：
- `github.com/qq1060656096/bizutil/registry`
- `github.com/redis/go-redis/v9`

## 快速开始

### 单组管理

适用于简单场景，管理多个命名的 Redis 客户端实例。

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/qq1060656096/mgredis"
)

func main() {
    ctx := context.Background()
    
    // 创建单组管理器
    group := mgredis.New()
    defer group.Close(ctx)
    
    // 注册 Redis 客户端
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
    
    // 获取 Redis 客户端（惰性初始化，实际在此时创建连接）
    client, err := group.Get(ctx, "cache")
    if err != nil {
        panic(err)
    }
    
    // 使用 Redis 客户端
    err = client.Set(ctx, "key", "value", time.Minute).Err()
    if err != nil {
        panic(err)
    }
    
    val, err := client.Get(ctx, "key").Result()
    if err != nil {
        panic(err)
    }
    fmt.Println(val) // 输出: value
}
```

### 多组管理

适用于复杂场景，按业务场景分组管理 Redis 连接。

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/qq1060656096/mgredis"
)

func main() {
    ctx := context.Background()
    
    // 创建多组管理器
    manager := mgredis.NewManager()
    defer manager.Close(ctx)
    
    // 添加会话缓存组
    manager.AddGroup("session-cache")
    sessionGroup, _ := manager.Group("session-cache")
    
    // 注册会话 Redis 实例
    _, err := sessionGroup.Register(ctx, "primary", mgredis.RedisConfig{
        Name:     "会话缓存",
        Addr:     "127.0.0.1:6379",
        DB:       1,
        PoolSize: 5,
    })
    if err != nil {
        panic(err)
    }
    
    // 添加限流器组
    manager.AddGroup("rate-limiter")
    rlGroup, _ := manager.Group("rate-limiter")
    
    // 注册限流 Redis 实例
    _, err = rlGroup.Register(ctx, "primary", mgredis.RedisConfig{
        Name:     "限流器",
        Addr:     "127.0.0.1:6379",
        DB:       2,
        PoolSize: 5,
    })
    if err != nil {
        panic(err)
    }
    
    // 获取不同组的 Redis 客户端
    sessionRedis, _ := sessionGroup.Get(ctx, "primary")
    rateRedis, _ := rlGroup.Get(ctx, "primary")
    
    // 使用各自的客户端
    _ = sessionRedis.Set(ctx, "session:123", "user_data", 30*time.Minute).Err()
    _ = rateRedis.Incr(ctx, "rate:api:123").Err()
    
    // 列出所有组名
    groupNames := manager.ListGroupNames()
    fmt.Println("组名:", groupNames)
}
```

## 配置说明

### RedisConfig 配置项

```go
type RedisConfig struct {
    // Name 资源描述名称，用于日志等
    Name string
    
    // Addr Redis服务器地址，格式: "host:port" (必填)
    Addr string
    
    // Password 密码，为空表示无密码
    Password string
    
    // DB 数据库索引，默认为0
    DB int
    
    // PoolSize 最大连接数，默认为10
    PoolSize int
    
    // MinIdleConns 最小空闲连接数，默认为2
    MinIdleConns int
    
    // DialTimeout 连接超时时间，默认为5秒
    DialTimeout time.Duration
    
    // ReadTimeout 读取超时时间，默认为3秒
    ReadTimeout time.Duration
    
    // WriteTimeout 写入超时时间，默认为3秒
    WriteTimeout time.Duration
    
    // MaxRetries 最大重试次数，默认为3
    MaxRetries int
    
    // PoolTimeout 从连接池获取连接的超时时间，默认为4秒
    PoolTimeout time.Duration
    
    // IdleTimeout 空闲连接超时时间，默认为5分钟
    IdleTimeout time.Duration
}
```

### 推荐配置

#### 缓存场景
```go
mgredis.RedisConfig{
    Addr:         "127.0.0.1:6379",
    DB:           0,
    PoolSize:     10,
    MinIdleConns: 2,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
}
```

#### 高并发场景
```go
mgredis.RedisConfig{
    Addr:         "127.0.0.1:6379",
    DB:           0,
    PoolSize:     100,
    MinIdleConns: 10,
    DialTimeout:  5 * time.Second,
    ReadTimeout:  1 * time.Second,
    WriteTimeout: 1 * time.Second,
    MaxRetries:   3,
}
```

## API 文档

### Group（单组管理器）

```go
// 创建单组管理器
group := mgredis.New()

// 注册 Redis 客户端（返回 client 和 isNew 标志）
client, isNew, err := group.Register(ctx, "name", config)

// 获取 Redis 客户端（惰性初始化）
client, err := group.Get(ctx, "name")

// 必须获取（失败会 panic）
client := group.MustGet(ctx, "name")

// 注销客户端（会关闭连接）
err := group.Unregister(ctx, "name")

// 列出所有已注册的名称
names := group.ListNames()

// 关闭所有客户端
group.Close(ctx)
```

### Manager（多组管理器）

```go
// 创建多组管理器
manager := mgredis.NewManager()

// 添加组
manager.AddGroup("group-name")

// 获取组
group, err := manager.Group("group-name")

// 获取所有组名
groupNames := manager.ListGroupNames()

// 关闭所有组的所有客户端
manager.Close(ctx)
```

## 错误处理

```go
// 判断错误类型
if mgredis.IsErrNoAddr(err) {
    // 缺少 Redis 地址
}

if mgredis.IsErrPingFailed(err) {
    // 连接测试失败
}

if mgredis.IsErrClientNotFound(err) {
    // 客户端未找到
}
```

## 高级用法

### 主从切换

```go
group := mgredis.New()

// 注册主库
_, _ = group.Register(ctx, "cache", mgredis.RedisConfig{
    Addr: "127.0.0.1:6379",
    DB:   0,
})

// 使用中...

// 主库故障，切换到从库
_ = group.Unregister(ctx, "cache")
_, _ = group.Register(ctx, "cache", mgredis.RedisConfig{
    Addr: "127.0.0.1:6380", // 新地址
    DB:   0,
})
```

### 动态注册

```go
group := mgredis.New()

// 运行时根据需要动态注册
for _, shard := range shards {
    _, err := group.Register(ctx, shard.Name, mgredis.RedisConfig{
        Addr: shard.Addr,
        DB:   shard.DB,
    })
    if err != nil {
        log.Printf("注册分片 %s 失败: %v", shard.Name, err)
    }
}
```

## 注意事项

1. **惰性初始化**：实际的 Redis 连接在首次 `Get` 时创建，而不是 `Register` 时
2. **连接测试**：创建连接时会自动执行 `Ping` 测试，确保连接可用
3. **优雅关闭**：使用 `defer Close(ctx)` 确保程序退出时关闭所有连接
4. **重复注册**：同名重复注册会返回已存在的客户端，不会创建新连接
5. **线程安全**：所有 API 都是线程安全的，可以在多个 goroutine 中并发使用

## 设计原理

`mgredis` 采用与 `mgorm` 相同的设计模式：

- 使用 `bizutil/registry` 的泛型资源管理能力
- 提供 `opener` 函数创建 Redis 客户端
- 提供 `closer` 函数关闭 Redis 客户端
- 支持单组 (`Group`) 和多组 (`Manager`) 两种管理模式
- 惰性初始化减少不必要的连接开销
- 统一的生命周期管理


## 许可证

[Apache License](LICENSE)

## 参考

- [mgorm](https://github.com/qq1060656096/mgorm) - GORM 数据库管理器
- [bizutil/registry](https://github.com/qq1060656096/bizutil) - 通用资源注册管理包
- [go-redis](https://github.com/redis/go-redis) - Redis Go 客户端
