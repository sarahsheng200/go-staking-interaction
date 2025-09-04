package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"staking-interaction/common/config"
	"time"
)

func NewRedisClient(config *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %v", err)
	}

	fmt.Println("Redis connection success.")
	return client, nil
}

func NewRedisClientWithRetry(config *config.RedisConfig, maxRetries int) (*redis.Client, error) {
	var client *redis.Client
	var err error

	for i := 0; i <= maxRetries; i++ {
		client, err = NewRedisClient(config)
		if err == nil {
			return client, nil
		}
		if i < maxRetries {
			waitTime := time.Duration(i) * 2 * time.Second
			fmt.Printf("Retrying in %v...\n", waitTime)
			time.Sleep(waitTime)
		}
	}
	return nil, fmt.Errorf("redis connection failed: %v, tried %d times", err, maxRetries)
}
