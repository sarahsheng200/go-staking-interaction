package repository

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"staking-interaction/common/config"
	"staking-interaction/database"
	"staking-interaction/model"
	"time"
)

type SwRepo struct {
	db *gorm.DB // 事务对应的 GORM DB 实例（已绑定事务）
}

func (w *SwRepo) GetAssetByAccountIdWithLock(accountID int) (*model.AccountAsset, error) {
	var asset model.AccountAsset

	// 使用 FOR UPDATE 添加行锁，防止并发修改
	err := w.db.Set("gorm:query_option", "FOR UPDATE").
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

func (w *SwRepo) UpdateWithdrawalInfo(withdraw model.Withdrawal) error {
	if err := w.db.Save(withdraw).Error; err != nil {
		return fmt.Errorf("SwRepo UpdateWithdrawalInfo failed: %w", err)
	}
	return nil
}

// AddBill 事务内创建账单
func (w *SwRepo) AddBill(bill *model.Bill) error {
	if err := w.db.Create(bill).Error; err != nil {
		return fmt.Errorf("tx add bill failed: %w", err)
	}
	return nil
}

func (w *SwRepo) UpdateAssetWithOptimisticLock(
	asset *model.AccountAsset,
	newBalance string,
	tokenType int,
) error {
	// 记录当前版本号
	currentVersion := asset.Version

	// 更新余额和版本号
	switch tokenType {
	case config.TokenTypeMTK:
		asset.MtkBalance = newBalance
	case config.TokenTypeBNB:
		asset.BnbBalance = newBalance
	default:
		return fmt.Errorf("unsupported token type: %d", tokenType)
	}

	asset.Version = currentVersion + 1
	asset.UpdatedAt = time.Now()

	// 使用乐观锁更新（WHERE version = current_version）
	result := w.db.Model(asset).
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

func SwWithTransaction(fn func(wd *SwRepo) error) error {
	db := database.DB

	return db.Transaction(func(wd *gorm.DB) error {
		SwRepo := &SwRepo{db: wd}

		if err := fn(SwRepo); err != nil {
			return fmt.Errorf("WithTransaction failed: %w", err)
		}

		return nil
	})
}
