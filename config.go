package mgredis

import (
	"time"
)

// RedisConfig Redis客户端配置
type RedisConfig struct {
	// Name 资源描述名称，用于日志等
	Name string `json:"name"`

	// Addr Redis服务器地址，格式: "host:port"
	Addr string `json:"addr"`

	// Password 密码，为空表示无密码
	Password string `json:"password"`

	// DB 数据库索引，默认为0
	DB int `json:"db"`

	// PoolSize 最大连接数，默认为10
	PoolSize int `json:"pool_size"`

	// MinIdleConns 最小空闲连接数，默认为2
	MinIdleConns int `json:"min_idle_conns"`

	// DialTimeout 连接超时时间，默认为5秒
	DialTimeout time.Duration `json:"dial_timeout"`

	// ReadTimeout 读取超时时间，默认为3秒
	ReadTimeout time.Duration `json:"read_timeout"`

	// WriteTimeout 写入超时时间，默认为3秒
	WriteTimeout time.Duration `json:"write_timeout"`

	// MaxRetries 最大重试次数，默认为3
	MaxRetries int `json:"max_retries"`

	// PoolTimeout 从连接池获取连接的超时时间，默认为4秒
	PoolTimeout time.Duration `json:"pool_timeout"`

	// IdleTimeout 空闲连接超时时间，默认为5分钟
	IdleTimeout time.Duration `json:"idle_timeout"`
}

// CheckAndSetDefaults 检查配置并设置默认值
func (cfg *RedisConfig) CheckAndSetDefaults() error {
	if cfg.Addr == "" {
		return ErrNoAddr
	}

	// 设置默认值
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 10
	}

	if cfg.MinIdleConns <= 0 {
		cfg.MinIdleConns = 2
	}

	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 5 * time.Second
	}

	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 3 * time.Second
	}

	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 3 * time.Second
	}

	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}

	if cfg.PoolTimeout <= 0 {
		cfg.PoolTimeout = 4 * time.Second
	}

	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = 5 * time.Minute
	}

	return nil
}
