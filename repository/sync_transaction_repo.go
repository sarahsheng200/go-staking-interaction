package repository

import (
	"fmt"
	"gorm.io/gorm"
	"staking-interaction/database"
	"staking-interaction/model"
)

type TxRepository struct {
	db *gorm.DB // 事务对应的 GORM DB 实例（已绑定事务）
}

// GetAccountAssetByAccountId 事务内查询用户资产
func (t *TxRepository) GetAccountAssetByAccountId(accountID int) (*model.AccountAsset, error) {
	var asset model.AccountAsset
	if err := t.db.Where("account_id = ?", accountID).First(&asset).Error; err != nil {
		return nil, fmt.Errorf("tx get asset failed: %w", err)
	}
	return &asset, nil
}

// AddBill 事务内创建账单
func (t *TxRepository) AddBill(bill *model.Bill) error {
	if err := t.db.Create(bill).Error; err != nil {
		return fmt.Errorf("tx add bill failed: %w", err)
	}
	return nil
}

// AddTransactionLog 事务内创建交易日志
func (t *TxRepository) AddTransactionLog(log *model.TransactionLog) error {
	if log == nil {
		return fmt.Errorf("log 不能为 nil")
	}
	if err := t.db.Create(log).Error; err != nil {
		return fmt.Errorf("add trancation log failed: %w", err)
	}
	return nil
}

// UpdateAsset 事务内更新用户资产
func (t *TxRepository) UpdateAsset(asset *model.AccountAsset) error {
	if err := t.db.Save(asset).Error; err != nil {
		return fmt.Errorf("tx update asset failed: %w", err)
	}
	return nil
}

// GetTransactionLog 事务内查询交易记录
func (t *TxRepository) FindTransactionLog(hash string) bool {
	var count int64
	t.db.Where("hash = ?", hash).Select("count(*)").Count(&count)
	return count > 0
}

// WithTransaction 核心事务封装函数
// 参数：fn 是“需要在事务内执行的业务逻辑”（接收 TxRepository，返回 error）
func TxWithTransaction(fn func(txRepo *TxRepository) error) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		txRepo := &TxRepository{db: tx} // 所有操作共享同一个事务

		if err := fn(txRepo); err != nil {
			// GORM 会自动回滚事务
			return fmt.Errorf("transaction business failed: %w", err)
		}

		return nil // GORM 会自动提交事务
	})
}
