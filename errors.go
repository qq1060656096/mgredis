package mgredis

import "errors"

var (
	// ErrNoAddr 缺少Redis服务器地址
	ErrNoAddr = errors.New("mgredis: missing Addr in RedisConfig")

	// ErrPingFailed Redis连接测试失败
	ErrPingFailed = errors.New("mgredis: ping failed")

	// ErrClientNotFound 未找到指定名称的Redis客户端
	ErrClientNotFound = errors.New("mgredis: redis client not found")
)

// IsErrNoAddr 判断是否为缺少地址错误
func IsErrNoAddr(err error) bool {
	return errors.Is(err, ErrNoAddr)
}

// IsErrPingFailed 判断是否为连接测试失败错误
func IsErrPingFailed(err error) bool {
	return errors.Is(err, ErrPingFailed)
}

// IsErrClientNotFound 判断是否为客户端未找到错误
func IsErrClientNotFound(err error) bool {
	return errors.Is(err, ErrClientNotFound)
}
