package main

import (
	"flag"
	"fmt"
	"log"
	"staking-interaction/adapter"
	"staking-interaction/database"
	"staking-interaction/service"
	"staking-interaction/utils"
)

func main() {
	indexFlag := flag.String("index", "", "stake index")
	flag.Parse()

	index, e := utils.StringToBigInt(*indexFlag)
	if e != nil {
		log.Fatal(e)
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
	response, err := stakeService.Withdraw(index)

	if err != nil {
		log.Fatal("Withdraw failed: ", err)
	}
	fmt.Printf("withdraw success!\n")
	fmt.Printf("Hash: %s \n", response.Hash)
	fmt.Printf("ContractAddress: %s \n", response.ContractAddress)
	fmt.Printf("FromAddress: %s \n", response.FromAddress.Hex())
	fmt.Printf("Method: %s \n", response.Method)
}
