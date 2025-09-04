package common

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type DistributedLock struct {
	redis      *redis.Client
	lockKey    string
	lockVal    string
	expiration time.Duration
}

type LockManager struct {
	redis *redis.Client
}

func NewLockManager(redisClient *redis.Client) *LockManager {
	return &LockManager{redis: redisClient}
}

// GetAssetLock 获取资产锁
func (l *LockManager) GetAssetLock(accountId int, tokenType int, expiration time.Duration) *DistributedLock {
	lockKey := fmt.Sprintf("asset_lock:%d:%d", accountId, tokenType)
	lockVal := generateLockValue()

	return &DistributedLock{
		redis:      l.redis,
		lockKey:    lockKey,
		lockVal:    lockVal,
		expiration: expiration,
	}
}

// GetTransactionLogLock 获取账户锁
func (l *LockManager) GetTransactionLogLock(logId int, expiration time.Duration) *DistributedLock {
	lockKey := fmt.Sprintf("log_lock:%d", logId)
	lockVal := generateLockValue()

	return &DistributedLock{
		redis:      l.redis,
		lockKey:    lockKey,
		lockVal:    lockVal,
		expiration: expiration,
	}
}

// GetWithdrawLock 获取提现锁
func (l *LockManager) GetWithdrawLock(withdrawId int, expiration time.Duration) *DistributedLock {
	lockKey := fmt.Sprintf("withdraw_lock:%d", withdrawId)
	lockVal := generateLockValue()

	return &DistributedLock{
		redis:      l.redis,
		lockKey:    lockKey,
		lockVal:    lockVal,
		expiration: expiration,
	}
}

// TryLock 尝试锁
func (dl *DistributedLock) TryLock(ctx context.Context) (bool, error) {
	res, err := dl.redis.SetNX(ctx, dl.lockKey, dl.lockVal, dl.expiration).Result()
	if err != nil {
		return false, fmt.Errorf("try lock failed: %w", err)
	}
	return res, nil
}

func (dl *DistributedLock) Lock(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		locked, err := dl.TryLock(ctx)
		if err != nil {
			return fmt.Errorf("try lock failed: %w", err)
		}
		if locked {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			continue
		}
	}
	return fmt.Errorf("lock timeout")
}

// Unlock 释放锁（使用 Lua 脚本确保原子性）
func (dl *DistributedLock) Unlock(ctx context.Context) error {
	script := `
        if redis.call("GET", KEYS[1]) == ARGV[1] then
            return redis.call("DEL", KEYS[1])
        else
            return 0
        end
    `

	result, err := dl.redis.Eval(ctx, script, []string{dl.lockKey}, dl.lockVal).Result()
	if err != nil {
		return fmt.Errorf("unlock failed: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock not found or expired")
	}

	return nil
}

// Renew 续期锁
func (dl *DistributedLock) Renew(ctx context.Context) error {
	script := `
        if redis.call("GET", KEYS[1]) == ARGV[1] then
            return redis.call("EXPIRE", KEYS[1], ARGV[2])
        else
            return 0
        end
    `

	result, err := dl.redis.Eval(ctx, script, []string{dl.lockKey}, dl.lockVal, int(dl.expiration.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("renew lock failed: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock not found or not owned")
	}

	return nil
}

func generateLockValue() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
