package main

import (
	"flag"
	"math/big"
	"os"
	"staking-interaction/adapter"
	"staking-interaction/common/logger"
	"staking-interaction/database"
	"staking-interaction/service"
	"staking-interaction/utils"
)

func main() {
	log := logger.GetLogger()
	log.WithFields(map[string]interface{}{
		"module": "cmd/airdroperc",
	})

	countFlag := flag.Int("count", 0, "airdroperc count")
	batchSizeFlag := flag.Int("batchSize", 0, "batch size of airdroperc")
	amountFlag := flag.String("amount", "", "amount range of airdroperc: 0-amount")

	flag.Parse()

	// 验证必填参数
	if *countFlag <= 0 {
		log.WithFields(map[string]interface{}{
			"action": "validate_input",
			"param":  "count",
			"value":  *countFlag,
			"detail": "count should be greater than 0",
		}).Fatal("Invalid argument: -count")
	}
	if *batchSizeFlag <= 0 {
		log.WithFields(map[string]interface{}{
			"action": "validate_input",
			"param":  "batchSize",
			"value":  *batchSizeFlag,
			"detail": "batchSize should be greater than 0",
		}).Fatal("Invalid argument: -batchSize")
	}
	maxAmount := new(big.Int)
	if _, ok := maxAmount.SetString(*amountFlag, 10); !ok {
		log.WithFields(map[string]interface{}{
			"action": "validate_input",
			"param":  "amount",
			"value":  *amountFlag,
			"detail": "amount format is invalid",
		}).Fatal("Invalid argument: -amount")
	}

	// 连接数据库
	err := database.MysqlConn()
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "init_db",
			"error_code": "DB_CONN_FAIL",
			"detail":     err.Error(),
		}).Fatal("MySQL database connect failed")
		return
	}
	log.WithFields(map[string]interface{}{
		"action": "init_db",
		"detail": "MySQL database connected",
	}).Info("Database connected")
	defer func() {
		err := database.CloseConn()
		if err != nil {
			log.WithFields(map[string]interface{}{
				"action":     "close_db",
				"error_code": "DB_CLOSE_FAIL",
				"detail":     err.Error(),
			}).Error("Close database failed")
		}
	}()

	clientInfo, err := adapter.NewInitClient()
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "init_client",
			"error_code": "CLIENT_INIT_FAIL",
			"detail":     err.Error(),
		}).Fatal("Init client failed")
	}
	defer clientInfo.CloseInitClient()

	amountArray, err := utils.GenerateRandomAmount(*countFlag, maxAmount)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "generate_amount",
			"error_code": "AMOUNT_GEN_FAIL",
			"detail":     err.Error(),
		}).Fatal("Failed to generate random amounts")
	}

	airdropService := service.NewAirdropService(clientInfo, log)
	response, err := airdropService.AirdropERC20(*countFlag, *batchSizeFlag, amountArray)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "airdrop_erc",
			"error_code": "AIRDROP_FAIL",
			"detail":     err.Error(),
		}).Fatal("Airdrop erc failed")
	}

	// 输出结果
	log.WithFields(map[string]interface{}{
		"action":           "airdrop_erc",
		"result":           "success",
		"Msg":              response.Msg,
		"CompletedBatches": response.CompletedBatches,
		"SuccessBatches":   response.SuccessBatches,
		"FailBatches":      response.FailBatches,
		"Data":             response.Data,
	}).Info("Airdrop ERC succeeded")

	if !utils.IsEmptyOrSpaceString(response.Error) {
		log.WithFields(map[string]interface{}{
			"action": "airdrop_erc",
			"result": "fail",
			"detail": response.Error,
		}).Error("Airdrop ERC error")
	}

	os.Exit(0)
}
