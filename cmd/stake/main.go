package main

import (
	"flag"
	"fmt"
	"log"
	"staking-interaction/adapter"
	"staking-interaction/database"
	"staking-interaction/service"
)

func main() {
	amountFlag := flag.Int64("amount", 0, "质押金额（最小单位，如wei）")
	periodFlag := flag.Uint("period", 0, "质押周期（天）")
	flag.Parse()
	// 验证必填参数
	if *amountFlag <= 0 {
		log.Fatal("amount should be greater than 0: -amount")
	}
	if *periodFlag < 0 {
		log.Fatal("period is invalid: -period")
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

	stakeService := service.NewStakeService(clientInfo)
	response, err := stakeService.Stake(*amountFlag, uint8(*periodFlag))
	if err != nil {
		log.Fatal("Stake failed: ", err)
	}
	fmt.Printf("stake success!\n")
	fmt.Printf("Hash: %s \n", response.Hash)
	fmt.Printf("ContractAddress: %s \n", response.ContractAddress)
	fmt.Printf("FromAddress: %s \n", response.FromAddress.Hex())
	fmt.Printf("Method: %s \n", response.Method)
}
