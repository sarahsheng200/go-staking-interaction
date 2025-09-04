package redis

import (
	constant "staking-interaction/common"
	"time"
)

// Redis 配置结构
type RedisConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	Password     string        `json:"password" yaml:"password"`
	Database     int           `json:"database" yaml:"database"`
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`
}

// 默认配置
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:         "127.0.0.1",
		Port:         6379,
		Password:     "", // 没有密码
		Database:     0,  // 默认数据库
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  5 * time.Minute,
	}
}

// 从环境变量或配置文件读取
func LoadRedisConfig() *RedisConfig {
	config := DefaultRedisConfig()

	config.Password = constant.REDIS_PASSWORD

	return config
}
