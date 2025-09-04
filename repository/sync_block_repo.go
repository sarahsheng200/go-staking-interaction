package repository

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	constant "staking-interaction/common/config"
	"staking-interaction/database"
	"staking-interaction/model"
	"time"
)

type TxRepository struct {
	db    *gorm.DB // 事务对应的 GORM DB 实例（已绑定事务）
	redis *redis.Client
}

// GetAccountAssetByAccountId 事务内查询用户资产
func (t *TxRepository) GetAccountAssetByAccountId(accountID int) (*model.AccountAsset, error) {
	var asset model.AccountAsset
	if err := t.db.Where("account_id = ?", accountID).First(&asset).Error; err != nil {
		return nil, fmt.Errorf("tx get asset failed: %w", err)
	}
	return &asset, nil
}

// GetAssetByAccountIdWithLock 添加行锁的资产查询
func (t *TxRepository) GetAssetByAccountIdWithLock(accountID int) (*model.AccountAsset, error) {
	var asset model.AccountAsset

	// 使用 FOR UPDATE 添加行锁，防止并发修改
	err := t.db.Set("gorm:query_option", "FOR UPDATE").
		Where("account_id = ?", accountID).
		First(&asset).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account asset not found for account_id: %d", accountID)
		}
		return nil, fmt.Errorf("get asset with lock failed: %w", err)
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

// TransactionExists 事务内查询交易记录
func (t *TxRepository) TransactionExists(hash string) (bool, error) {
	var count int64
	err := t.db.Model(&model.TransactionLog{}).Where("hash = ?", hash).Select("count(*)").Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (t *TxRepository) UpdateAssetWithOptimisticLock(
	asset *model.AccountAsset,
	newBalance string,
	tokenType int,
) error {
	// 记录当前版本号
	currentVersion := asset.Version

	// 更新余额和版本号
	switch tokenType {
	case constant.TokenTypeMTK:
		asset.MtkBalance = newBalance
	case constant.TokenTypeBNB:
		asset.BnbBalance = newBalance
	default:
		return fmt.Errorf("unsupported token type: %d", tokenType)
	}

	asset.Version = currentVersion + 1
	asset.UpdatedAt = time.Now()

	// 使用乐观锁更新（WHERE version = current_version）
	result := t.db.Model(asset).
		Where("account_id = ? AND version = ?", asset.AccountID, currentVersion).
		Updates(map[string]interface{}{
			"bnb_balance": asset.BnbBalance,
			"mtk_balance": asset.MtkBalance,
			"version":     asset.Version,
			"updated_at":  asset.UpdatedAt,
		})

	if result.Error != nil {
		return fmt.Errorf("updateasset with optimistic lock failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock failed: asset was modified by another process")
	}

	return nil
}

// TxWithTransaction 核心事务封装函数
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
