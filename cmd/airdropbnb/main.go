package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"staking-interaction/adapter"
	"staking-interaction/database"
	"staking-interaction/service"
	"staking-interaction/utils"
)

func main() {
	countFlag := flag.Int("count", 0, "airdroperc count")
	batchSizeFlag := flag.Int("batchSize", 0, "batch size of airdroperc")
	amountFlag := flag.String("amount", "", "amount range of airdroperc: 0-amount")

	flag.Parse()

	// 验证必填参数
	if *countFlag <= 0 {
		log.Fatal("count should be greater than 0: -count")
	}
	if *batchSizeFlag <= 0 {
		log.Fatal("batchSize should be greater than 0: -batchSize")
	}
	maxAmount := new(big.Int)
	if _, ok := maxAmount.SetString(*amountFlag, 10); !ok {
		log.Fatalf("amount format is invalid: %s", *amountFlag)
	}

	err := database.MysqlConn()
	if err != nil {
		log.Fatal("MySQL database connect failed: ", err)
		return
	}
	defer func() {
		err := database.CloseConn()
		if err != nil {
			log.Fatal("Close database failed: ", err)
		}
	}()

	clientInfo, err := adapter.NewInitClient()
	if err != nil {
		log.Fatal("Init client failed: ", err)
	}
	defer clientInfo.CloseInitClient()

	amountArray, err := utils.GenerateRandomAmount(*countFlag, maxAmount)
	if err != nil {
		log.Fatalf("failed to generate random amounts: %v", err)
	}

	airdropService := service.NewAirdropService(clientInfo)
	response, err := airdropService.AirdropBNB(*countFlag, *batchSizeFlag, amountArray)
	if err != nil {
		log.Fatal("Airdrop ERC20 failed: ", err)
	}

	// 输出结果
	fmt.Printf("airdro bnb success!\n")
	fmt.Printf("Msg: %s\n", response.Msg)
	fmt.Printf("CompletedBatches: %d \n", response.CompletedBatches)
	fmt.Printf("SuccessBatches: %d \n", response.SuccessBatches)
	fmt.Printf("FailBatches: %d \n", response.FailBatches)
	fmt.Printf("Data: %v \n", response.Data)
	if !utils.IsEmptyOrSpaceString(response.Error) {
		fmt.Printf("Error: %s \n", response.Error)
	}

	os.Exit(0)
}
