package main

import (
	"context"
	"fmt"
	"log"
	_ "math/big"
	"net/http"
	"os"
	"os/signal"
	"staking-interaction/database"
	srouter "staking-interaction/router"
	"staking-interaction/service"
	"syscall"
	"time"
)

const PORT = 8084

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

	router := srouter.InitRouter()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", PORT),
		Handler: router,
	}
	log.Println(fmt.Sprintf("Listening and serving HTTP on Port: %d, Pid: %d", PORT, os.Getpid()))

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server :", err)
		}
	}()
	client := service.InitContracts()
	defer client.Close()
	service.InitStakeContract()
	service.InitAirdropContract()
	//stakeService.ListenToEvents()

	// 创建系统信号接收器
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("shutdown server...")

	//stakeService.CloseListener()
	log.Println("stake event listener closed")

	// 创建5s的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("server shutdown:", err)
	}
	log.Printf("server exiting...")
}
