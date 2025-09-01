package repository

import (
	"fmt"
	"staking-interaction/database"
	"staking-interaction/model"
)

func AddBill(bill *model.Bill) error {
	if bill == nil {
		return fmt.Errorf("bill 不能为 nil")
	}
	err := database.DB.Create(bill).Error
	if err != nil {
		return fmt.Errorf("创建账单失败: %w", err)
	}
	return nil
}
