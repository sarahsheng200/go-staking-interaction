package repository

import (
	"fmt"
	"staking-interaction/adapter"
	"staking-interaction/model"
)

func AddTransactionLog(log *model.TransactionLog) error {
	if log == nil {
		return fmt.Errorf("log 不能为 nil")
	}
	err := adapter.DB.Create(log).Error
	if err != nil {
		return fmt.Errorf("创建trancation log失败: %w", err)
	}
	return nil
}
