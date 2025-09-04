package main

import (
	"log"
	"os"
	"os/signal"
	"staking-interaction/adapter"
	"staking-interaction/database"
	"staking-interaction/listener"
	"staking-interaction/service"
	"syscall"
)

func main() {
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

	//sync block
	log.Println("Initializing sync block...")
	syncBlock := listener.NewSyncBlockInfo(clientInfo.Client, 5)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("SyncBlockInfo panic: %v", r)
			}
		}()
		log.Println("SyncBlockInfo goroutine started")
		syncBlock.Start()
		log.Println("SyncBlockInfo goroutine ended")
	}()
	defer func() {
		log.Println("Stopping SyncBlockInfo...")
		syncBlock.Stop()
	}()

	// withdraw handler
	log.Println("Initializing withdraw handler...")
	txService := service.NewTransactionService(clientInfo)
	withdrawHd := listener.NewWithdrawHandler(txService)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("WithdrawHandler panic: %v", r)
			}
		}()
		log.Println("WithdrawHandler goroutine started")
		withdrawHd.Start()
		log.Println("WithdrawHandler goroutine ended")
	}()
	defer func() {
		log.Println(" Stopping WithdrawHandler...")
		withdrawHd.Stop()
	}()

	//sync withdraw
	log.Println("Initializing sync withdraw...")
	syncWithdraw := listener.NewSyncWithdrawHandler(clientInfo.Client)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("SyncWithdrawHandler panic: %v", r)
			}
		}()
		log.Println("SyncWithdrawHandler goroutine started")
		syncWithdraw.Start()
		log.Println("SyncWithdrawHandler goroutine ended")
	}()
	defer func() {
		log.Println("Stopping SyncWithdrawHandler...")
		syncWithdraw.Stop()
	}()

	log.Println("SyncWithdrawHandler launched, all services initialized!")

	// 创建系统信号接收器
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("shutdown signal received, exiting...")

}
