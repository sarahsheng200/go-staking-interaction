package config

import (
	"time"
)

// Redis 配置结构
type RedisConfig struct {
	Host         string        `json:"host" yaml:"host"`
	Port         int           `json:"port" yaml:"port"`
	Password     string        `json:"password" yaml:"password"`
	Database     int           `json:"database" yaml:"database"`             // 数据库编号 (0-15)
	PoolSize     int           `json:"pool_size" yaml:"pool_size"`           // 连接池大小
	MinIdleConns int           `json:"min_idle_conns" yaml:"min_idle_conns"` // 最小空闲连接数
	DialTimeout  time.Duration `json:"dial_timeout" yaml:"dial_timeout"`     // 建立连接超时
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`     // 读操作超时
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`   // 写操作超时
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`     // 空闲连接超时
}

// 配置
func LoadRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:         REDIS_HOST,
		Port:         REDIS_PORT,
		Password:     "",
		Database:     REDIS_DATABASE, // 默认数据库
		PoolSize:     10,
		MinIdleConns: 5,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  5 * time.Minute,
	}
}
