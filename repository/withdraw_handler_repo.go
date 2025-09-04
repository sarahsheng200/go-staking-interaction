package repository

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"staking-interaction/database"
	"staking-interaction/model"
)

type WdRepo struct {
	db *gorm.DB // 事务对应的 GORM DB 实例（已绑定事务）
}

// GetAssetByAccountIdWithLock 添加行锁的资产查询
func (w *WdRepo) GetAssetByAccountIdWithLock(accountID int) (*model.AccountAsset, error) {
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

func (w *WdRepo) UpdateWithdrawalInfo(withdraw model.Withdrawal) error {
	if err := w.db.Save(withdraw).Error; err != nil {
		return fmt.Errorf("WdRepo UpdateWithdrawalInfo failed: %w", err)
	}
	return nil
}

func GetWithdrawalInfoByStatus(status int) ([]model.Withdrawal, error) {
	var withdrawalInfo []model.Withdrawal
	if err := database.DB.Where("status=?", status).Find(&withdrawalInfo).Error; err != nil {
		return nil, fmt.Errorf("WdRepo GetWithdrawalInfoByStatus failed: %w", err)
	}
	return withdrawalInfo, nil
}

func WdWithTransaction(fn func(wd *WdRepo) error) error {
	db := database.DB

	return db.Transaction(func(wd *gorm.DB) error {
		wdRepo := &WdRepo{db: wd}

		if err := fn(wdRepo); err != nil {
			return fmt.Errorf("WithTransaction failed: %w", err)
		}

		return nil
	})
}
