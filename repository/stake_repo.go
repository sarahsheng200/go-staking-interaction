package repository

import (
	"log"
	"staking-interaction/database"
	"staking-interaction/model"
)

func AddStakeInfo(stake model.Stake) {
	database.DB.Create(&stake)

	log.Printf("Add " + stake.Method + " info to database success, indexId: " + stake.IndexNum)
}

func GetAllStakesById(id string) model.Stake {
	var stake model.Stake
	database.DB.Where("id = ?", id).First(&stake)
	return stake
}
