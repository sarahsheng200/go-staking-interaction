package repository

import (
	"fmt"
	"staking-interaction/database"
	"staking-interaction/model"
)

func GetAccount(fromAddress string) *model.Account {
	account := &model.Account{}
	database.DB.Model(&model.Account{}).Where("wallet_address = ?", fromAddress).First(account)

	return account
}

func GetAccountAsset(accountId int) (*model.AccountAsset, error) {
	asset := &model.AccountAsset{}
	err := database.DB.Model(&model.AccountAsset{}).Where("account_id = ?", accountId).First(asset).Error
	if err != nil {
		return nil, fmt.Errorf("repo: get account asset failed: %w", err)
	}

	return asset, nil
}

func UpdateAsset(asset *model.AccountAsset) error {
	// 注意：需确保asset包含主键ID或唯一索引字段（如AccountID+TokenAddr）
	return database.DB.Save(asset).Error
}
