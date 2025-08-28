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
	toAddressFlag := flag.String("toAddress", "", "stake index")
	amountFlag := flag.String("amount", "", "stake index")
	flag.Parse()

	amount, e := utils.StringToBigInt(*amountFlag)
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

	transactionService := service.NewTransactionService(clientInfo)
	response, err := transactionService.SendErc20(*toAddressFlag, amount)
	if err != nil {
		log.Fatal("Transaction ERC failed: ", err)
	}

	fmt.Printf("Transaction ERC success!\n")
	fmt.Printf("Hash: %s \n", response.Hash)
	fmt.Printf("Symbol: %s \n", response.Symbol)
	fmt.Printf("Decimal: %v \n", response.Decimal)
}
