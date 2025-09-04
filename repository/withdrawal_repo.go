package repository

import (
	"fmt"
	"gorm.io/gorm"
	"staking-interaction/database"
	"staking-interaction/model"
)

type WdRepo struct {
	db *gorm.DB // 事务对应的 GORM DB 实例（已绑定事务）
}

func (w *WdRepo) GetAssetByAddress(walletAddress string) (*model.AccountAsset, error) {
	var asset model.AccountAsset
	if err := w.db.Joins("LEFT JOIN account ON account.account_id=account_asset.account_id").
		Where("account.wallet_address = ?", walletAddress).
		First(&asset).Error; err != nil {
		return nil, fmt.Errorf("WdRepo GetBnbBalanceByAddress failed: %w", err)
	}
	return &asset, nil
}

func (w *WdRepo) UpdateWithdrawalInfo(withdraw model.Withdrawal) error {
	if err := w.db.Save(withdraw).Error; err != nil {
		return fmt.Errorf("WdRepo UpdateWithdrawalInfo failed: %w", err)
	}
	return nil
}

func (w *WdRepo) AddBill(bill *model.Bill) error {
	if err := w.db.Create(bill).Error; err != nil {
		return fmt.Errorf("WdRepo AddBill failed: %w", err)
	}
	return nil
}

func (w *WdRepo) UpdateAsset(asset *model.AccountAsset) error {
	return database.DB.Save(asset).Error
}

func (w *WdRepo) GetAccountAsset(accountId int) (*model.AccountAsset, error) {
	asset := &model.AccountAsset{}
	err := database.DB.Model(&model.AccountAsset{}).Where("account_id = ?", accountId).First(asset).Error
	if err != nil {
		return nil, fmt.Errorf("repo: get account asset failed: %w", err)
	}

	return asset, nil
}

func (w *WdRepo) GetAccount(fromAddress string) *model.Account {
	account := &model.Account{}
	database.DB.Model(&model.Account{}).Where("wallet_address = ?", fromAddress).First(account)

	return account
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
