package router

import (
	"github.com/gin-gonic/gin"
	"log"
	airdropService "staking-interaction/service/airdrop"
	stakeService "staking-interaction/service/stake"
	"staking-interaction/service/transaction"
)

func InitRouter() *gin.Engine {
	log.Println("Initializing router")
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	group := router.Group("/v1")

	staking := group.Group("/staking")
	{
		staking.POST("/stake", stakeService.Stake)
		staking.POST("/withdraw", stakeService.Withdraw)
		staking.GET("stake/:address", stakeService.GetAllStakesByFromAddress)
	}

	airdrop := group.Group("/airdropping")
	{
		airdrop.POST("/airdroperc20", airdropService.AirdropERC20)
		airdrop.POST("/generateWallet", airdropService.GenerateMultiWallets)
	}

	transfer := group.Group("/transfer")
	{
		transfer.POST("/transferERC20", transaction.SendErc20)
		transfer.POST("/transferBNB", transaction.SendBNB)
	}

	return router
}
