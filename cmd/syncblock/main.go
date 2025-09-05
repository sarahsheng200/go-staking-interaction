package main

import (
	"log"
	"os"
	"os/signal"
	"staking-interaction/adapter"
	redisClient "staking-interaction/common/redis"
	"staking-interaction/database"
	"staking-interaction/listener"
	"staking-interaction/service"
	"syscall"
)

func main() {
	// 1. 初始化数据库
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

	// 2. 初始化 Redis 连接
	redis, err := redisClient.NewRedisClientWithRetry()
	if err != nil {
		log.Fatal("Redis connection failed: ", err)
	}
	defer redis.Close()

	// 3. 创建 LockManager 实例
	lockManager := redisClient.NewLockManager(redis)
	log.Println("LockManager initialized successfully")

	// 4. 初始化区块链客户端
	clientInfo, err := adapter.NewInitClient()
	if err != nil {
		log.Fatal("Init client failed: ", err)
	}
	// 5. syncBlock
	log.Println("Initializing sync block...")
	syncBlock := listener.NewSyncBlockInfo(clientInfo.Client, nil, lockManager)
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

	// 6. withdraw handler
	log.Println("Initializing withdraw handler...")
	txService := service.NewTransactionService(clientInfo)
	withdrawHd := listener.NewWithdrawHandler(txService, lockManager)
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

	// 7. sync withdraw
	log.Println("Initializing sync withdraw...")
	syncWithdraw := listener.NewSyncWithdrawHandler(clientInfo.Client, lockManager)
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

	// 8. 等待关闭信号
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("shutdown signal received, exiting...")

}
