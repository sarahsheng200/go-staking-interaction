package repository

import (
	"log"
	"staking-interaction/database"
	"staking-interaction/model/stake"
)

func AddStakeInfo(stake stake.Stake) {
	database.DB.Create(&stake)

	log.Printf("Add " + stake.Method + " info to database success, indexId: " + stake.IndexNum)
}

func GetAllStakesByFromAddress(fromAddress string) stake.Stake {
	var stake stake.Stake
	database.DB.Where("from_address = ?", fromAddress).First(&stake)
	return stake
}
